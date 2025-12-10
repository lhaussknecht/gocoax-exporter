package collector

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/louispool/gocoax-exporter/client"
	"github.com/prometheus/client_golang/prometheus"
)

// GoCoaxCollector collects metrics from a single goCoax device
type GoCoaxCollector struct {
	client     *client.Client
	deviceName string
	timeout    time.Duration

	// Metric descriptors
	phyRateNPER      *prometheus.Desc
	phyRateVLPER     *prometheus.Desc
	phyRateGCD       *prometheus.Desc
	nodeInfo         *prometheus.Desc
	up               *prometheus.Desc
	scrapeDuration   *prometheus.Desc
	scrapeErrors     *prometheus.Desc
}

// NewGoCoaxCollector creates a new collector for a goCoax device
func NewGoCoaxCollector(deviceName, address, username, password string, timeout time.Duration) (*GoCoaxCollector, error) {
	c, err := client.NewClient(address, username, password, timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	return &GoCoaxCollector{
		client:     c,
		deviceName: deviceName,
		timeout:    timeout,

		phyRateNPER: prometheus.NewDesc(
			"gocoax_phy_rate_nper_mbps",
			"Normal Packet Error Rate PHY rate in Mbps between nodes",
			[]string{"device", "from_node", "to_node"},
			nil,
		),
		phyRateVLPER: prometheus.NewDesc(
			"gocoax_phy_rate_vlper_mbps",
			"Very Low Packet Error Rate PHY rate in Mbps between nodes (MoCA 2.5)",
			[]string{"device", "from_node", "to_node"},
			nil,
		),
		phyRateGCD: prometheus.NewDesc(
			"gocoax_phy_rate_gcd_mbps",
			"Greatest Common Divisor rate in Mbps for node",
			[]string{"device", "node"},
			nil,
		),
		nodeInfo: prometheus.NewDesc(
			"gocoax_node_info",
			"Node information with MoCA version",
			[]string{"device", "node", "moca_version", "is_nc"},
			nil,
		),
		up: prometheus.NewDesc(
			"gocoax_up",
			"Device is reachable and responding (1=up, 0=down)",
			[]string{"device"},
			nil,
		),
		scrapeDuration: prometheus.NewDesc(
			"gocoax_scrape_duration_seconds",
			"Time taken to scrape device metrics",
			[]string{"device"},
			nil,
		),
		scrapeErrors: prometheus.NewDesc(
			"gocoax_scrape_errors_total",
			"Total number of scrape errors",
			[]string{"device"},
			nil,
		),
	}, nil
}

// Describe implements prometheus.Collector
func (c *GoCoaxCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.phyRateNPER
	ch <- c.phyRateVLPER
	ch <- c.phyRateGCD
	ch <- c.nodeInfo
	ch <- c.up
	ch <- c.scrapeDuration
	ch <- c.scrapeErrors
}

// Collect implements prometheus.Collector
func (c *GoCoaxCollector) Collect(ch chan<- prometheus.Metric) {
	startTime := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	if err := c.collectMetrics(ctx, ch); err != nil {
		log.Printf("Error collecting metrics for device %s: %v", c.deviceName, err)
		// Report device as down
		ch <- prometheus.MustNewConstMetric(c.up, prometheus.GaugeValue, 0, c.deviceName)
		ch <- prometheus.MustNewConstMetric(c.scrapeErrors, prometheus.CounterValue, 1, c.deviceName)
	} else {
		// Report device as up
		ch <- prometheus.MustNewConstMetric(c.up, prometheus.GaugeValue, 1, c.deviceName)
		ch <- prometheus.MustNewConstMetric(c.scrapeErrors, prometheus.CounterValue, 0, c.deviceName)
	}

	duration := time.Since(startTime).Seconds()
	ch <- prometheus.MustNewConstMetric(c.scrapeDuration, prometheus.GaugeValue, duration, c.deviceName)
}

// collectMetrics performs the actual metric collection
func (c *GoCoaxCollector) collectMetrics(ctx context.Context, ch chan<- prometheus.Metric) error {
	// Step 1: Get local device information
	localInfo, err := c.client.GetLocalInfo(ctx)
	if err != nil {
		return fmt.Errorf("failed to get local info: %w", err)
	}

	nodeBitMask := localInfo.NodeBitMask
	mocaNetVer := localInfo.MocaNetVersion
	ncNodeID := localInfo.NCNodeID

	// Step 2: Get information for each active node
	nodeVersions := make(map[int]int)
	activeNodes := []int{}

	for nodeID := 0; nodeID < MAX_NUM_NODES; nodeID++ {
		if (nodeBitMask & (1 << nodeID)) == 0 {
			continue
		}

		nodeInfo, err := c.client.GetNetworkNodeInfo(ctx, nodeID)
		if err != nil {
			log.Printf("Warning: failed to get info for node %d: %v", nodeID, err)
			continue
		}

		nodeVersions[nodeID] = nodeInfo.MocaVersion
		activeNodes = append(activeNodes, nodeID)

		// Emit node info metric
		mocaVerStr := formatMocaVersion(nodeInfo.MocaVersion)
		isNC := "false"
		if nodeID == ncNodeID {
			isNC = "true"
		}
		ch <- prometheus.MustNewConstMetric(
			c.nodeInfo,
			prometheus.GaugeValue,
			1,
			c.deviceName,
			strconv.Itoa(nodeID),
			mocaVerStr,
			isNC,
		)
	}

	// Get NC MoCA version
	ncMocaVer := nodeVersions[ncNodeID]

	// Step 3: Get FMR info and calculate PHY rates for each active node
	for _, nodeID := range activeNodes {
		nodeMocaVer := nodeVersions[nodeID]

		// Determine version parameter for FMR request
		mocaVer := min(ncMocaVer, nodeMocaVer)
		var versionParam int
		if mocaVer < 0x20 {
			versionParam = 1 // MoCA 1.x
		} else {
			versionParam = 2 // MoCA 2.x
		}

		// Request FMR info for this node
		nodeMask := 1 << nodeID
		fmrInfo, err := c.client.GetFMRInfo(ctx, nodeMask, versionParam)
		if err != nil {
			log.Printf("Warning: failed to get FMR info for node %d: %v", nodeID, err)
			continue
		}

		// Calculate PHY rates
		matrix, err := CalculatePHYRates(
			nodeID,
			fmrInfo.Data,
			nodeMocaVer,
			ncMocaVer,
			mocaNetVer,
			nodeBitMask,
			nodeVersions,
		)
		if err != nil {
			log.Printf("Warning: failed to calculate PHY rates for node %d: %v", nodeID, err)
			continue
		}

		// Emit NPER metrics
		if nperRates, ok := matrix.NPER[nodeID]; ok {
			for destNode, rate := range nperRates {
				ch <- prometheus.MustNewConstMetric(
					c.phyRateNPER,
					prometheus.GaugeValue,
					float64(rate),
					c.deviceName,
					strconv.Itoa(nodeID),
					strconv.Itoa(destNode),
				)
			}
		}

		// Emit VLPER metrics
		if vlperRates, ok := matrix.VLPER[nodeID]; ok {
			for destNode, rate := range vlperRates {
				// Only emit if rate is non-zero (VLPER only exists for MoCA 2.5)
				if rate > 0 {
					ch <- prometheus.MustNewConstMetric(
						c.phyRateVLPER,
						prometheus.GaugeValue,
						float64(rate),
						c.deviceName,
						strconv.Itoa(nodeID),
						strconv.Itoa(destNode),
					)
				}
			}
		}

		// Emit GCD metric
		if gcdRate, ok := matrix.GCD[nodeID]; ok {
			ch <- prometheus.MustNewConstMetric(
				c.phyRateGCD,
				prometheus.GaugeValue,
				float64(gcdRate),
				c.deviceName,
				strconv.Itoa(nodeID),
			)
		}
	}

	return nil
}

// formatMocaVersion formats a MoCA version code into a readable string
func formatMocaVersion(version int) string {
	major := (version & 0xF0) >> 4
	minor := version & 0x0F

	if major == 1 {
		return fmt.Sprintf("1.%d", minor)
	} else if major == 2 {
		if minor == 5 {
			return "2.5"
		}
		return fmt.Sprintf("2.%d", minor)
	}

	return fmt.Sprintf("%d.%d", major, minor)
}

// Close releases resources held by the collector
func (c *GoCoaxCollector) Close() error {
	return c.client.Close()
}
