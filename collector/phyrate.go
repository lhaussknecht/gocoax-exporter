package collector

import (
	"fmt"
)

// MoCA protocol constants
const (
	LDPC_LEN_100MHZ = 3900
	LDPC_LEN_50MHZ  = 1200
	FFT_LEN_100MHZ  = 512
	FFT_LEN_50MHZ   = 256
	MAX_NUM_NODES   = 16
)

// PHYRateMatrix stores PHY rates between nodes
type PHYRateMatrix struct {
	NPER  map[int]map[int]int // [fromNode][toNode] = rate in Mbps
	VLPER map[int]map[int]int // [fromNode][toNode] = rate in Mbps
	GCD   map[int]int          // [node] = rate in Mbps (self-to-self)
}

// NewPHYRateMatrix creates a new empty PHY rate matrix
func NewPHYRateMatrix() *PHYRateMatrix {
	return &PHYRateMatrix{
		NPER:  make(map[int]map[int]int),
		VLPER: make(map[int]map[int]int),
		GCD:   make(map[int]int),
	}
}

// FMRPayloadParser parses FMR payload data to extract PHY rate parameters
type FMRPayloadParser struct {
	fmrData        []uint32
	readIndex      int
	alignmentFlag  bool
	entryNodeID    int
	entryMocaVer   int
	ncMocaVer      int
	mocaNetVer     int
	nodeBitMask    int
	nodeVersions   map[int]int // [nodeID] = mocaVersion
}

// NewFMRPayloadParser creates a new FMR payload parser
func NewFMRPayloadParser(fmrData []uint32, entryNodeID, entryMocaVer, ncMocaVer, mocaNetVer, nodeBitMask int, nodeVersions map[int]int) *FMRPayloadParser {
	return &FMRPayloadParser{
		fmrData:       fmrData,
		readIndex:     10, // FMR payload starts at index 10
		alignmentFlag: true,
		entryNodeID:   entryNodeID,
		entryMocaVer:  entryMocaVer,
		ncMocaVer:     ncMocaVer,
		mocaNetVer:    mocaNetVer,
		nodeBitMask:   nodeBitMask,
		nodeVersions:  nodeVersions,
	}
}

// parseFMREntry parses a single FMR entry for a destination node
func (p *FMRPayloadParser) parseFMREntry(destNodeID int) (gapNper, gapVLper, ofdmbNper, ofdmbVLper int, err error) {
	// Check if destination node is present
	if (p.nodeBitMask & (1 << destNodeID)) == 0 {
		// Node not present, skip data based on entry node's version
		return p.skipAbsentNode()
	}

	// Determine FMR payload version for this entry
	destMocaVer := p.nodeVersions[destNodeID]
	var fmrPayloadVer int

	if p.ncMocaVer < 0x20 {
		// If NC is MoCA 1.x, version is min of entry payload version and dest version
		entryPayloadVer := min(p.entryMocaVer, p.ncMocaVer)
		fmrPayloadVer = min(entryPayloadVer, destMocaVer)
	} else {
		// If NC is MoCA 2.x, version is entry node's version
		fmrPayloadVer = p.entryMocaVer
	}

	// Parse based on version
	if fmrPayloadVer == 0x20 || fmrPayloadVer == 0x25 {
		return p.parseMoCA2x()
	}

	return p.parseMoCA1x()
}

// parseMoCA2x parses MoCA 2.0/2.5 FMR entry (6 bytes per entry)
func (p *FMRPayloadParser) parseMoCA2x() (gapNper, gapVLper, ofdmbNper, ofdmbVLper int, err error) {
	if p.readIndex >= len(p.fmrData)-1 {
		return 0, 0, 0, 0, fmt.Errorf("insufficient FMR data at index %d", p.readIndex)
	}

	if p.alignmentFlag {
		// Aligned read: data is at current and next index
		gapNper = int((p.fmrData[p.readIndex] >> 24) & 0xFF)
		gapVLper = int((p.fmrData[p.readIndex] >> 16) & 0xFF)
		ofdmbNper = int(p.fmrData[p.readIndex] & 0xFFFF)
		ofdmbVLper = int((p.fmrData[p.readIndex+1] >> 16) & 0xFFFF)
		p.readIndex++
	} else {
		// Unaligned read: data spans current and next index
		gapNper = int((p.fmrData[p.readIndex] >> 8) & 0xFF)
		gapVLper = int(p.fmrData[p.readIndex] & 0xFF)
		ofdmbNper = int((p.fmrData[p.readIndex+1] >> 16) & 0xFFFF)
		ofdmbVLper = int(p.fmrData[p.readIndex+1] & 0xFFFF)
		p.readIndex += 2
	}

	p.alignmentFlag = !p.alignmentFlag
	return gapNper, gapVLper, ofdmbNper, ofdmbVLper, nil
}

// parseMoCA1x parses MoCA 1.x FMR entry (2 bytes per entry)
func (p *FMRPayloadParser) parseMoCA1x() (gapNper, gapVLper, ofdmbNper, ofdmbVLper int, err error) {
	if p.readIndex >= len(p.fmrData) {
		return 0, 0, 0, 0, fmt.Errorf("insufficient FMR data at index %d", p.readIndex)
	}

	gapVLper = 0
	ofdmbVLper = 0

	if p.alignmentFlag {
		// Aligned read: upper 16 bits
		gapNper = int((p.fmrData[p.readIndex] & 0xF8000000) >> 27)
		ofdmbNper = int((p.fmrData[p.readIndex] & 0x07FF0000) >> 16)
	} else {
		// Unaligned read: lower 16 bits
		gapNper = int((p.fmrData[p.readIndex] & 0x0000F800) >> 11)
		ofdmbNper = int(p.fmrData[p.readIndex] & 0x000007FF)
		p.readIndex++
	}

	p.alignmentFlag = !p.alignmentFlag
	return gapNper, gapVLper, ofdmbNper, ofdmbVLper, nil
}

// skipAbsentNode skips FMR data for an absent node
func (p *FMRPayloadParser) skipAbsentNode() (gapNper, gapVLper, ofdmbNper, ofdmbVLper int, err error) {
	fmrPayloadVer := p.entryMocaVer

	if fmrPayloadVer >= 0x20 {
		// MoCA 2.x: 6 bytes per entry
		if p.alignmentFlag {
			p.readIndex++
		} else {
			p.readIndex += 2
		}
	} else {
		// MoCA 1.x: 2 bytes per entry
		if !p.alignmentFlag {
			p.readIndex++
		}
	}

	p.alignmentFlag = !p.alignmentFlag
	return 0, 0, 0, 0, nil
}

// parseGCDForMixedMode parses the GCD field for mixed-mode networks (MoCA 2.x node in 1.x network)
func (p *FMRPayloadParser) parseGCDForMixedMode() (gapGcd, ofdmbGcd int, err error) {
	// GCD data is at fixed offset 34 for mixed-mode networks
	const gcdOffset = 34

	if gcdOffset >= len(p.fmrData) {
		return 0, 0, fmt.Errorf("insufficient FMR data for GCD at offset %d", gcdOffset)
	}

	gapGcd = int((p.fmrData[gcdOffset] >> 24) & 0xFF)
	ofdmbGcd = int((p.fmrData[gcdOffset] >> 8) & 0xFFFF)

	return gapGcd, ofdmbGcd, nil
}

// CalculateNPERRate calculates the NPER (Normal Packet Error Rate) PHY rate
func CalculateNPERRate(gapNper, ofdmbNper int, fmrPayloadVer int, gapVLper int) int {
	if gapNper == 0 {
		return 0
	}

	// Special case: if VLPER is 0 and version is 2.0, use 50MHz formula
	if gapVLper == 0 && fmrPayloadVer == 0x20 {
		return (LDPC_LEN_50MHZ * ofdmbNper) / ((FFT_LEN_50MHZ + (gapNper*2 + 10)) * 26)
	}

	// Default: use 100MHz formula
	return (LDPC_LEN_100MHZ * ofdmbNper) / ((FFT_LEN_100MHZ + ((gapNper + 10) * 2)) * 46)
}

// CalculateVLPERRate calculates the VLPER (Very Low Packet Error Rate) PHY rate
func CalculateVLPERRate(gapVLper, ofdmbVLper int) int {
	if gapVLper == 0 {
		return 0
	}

	return (LDPC_LEN_100MHZ * ofdmbVLper) / ((FFT_LEN_100MHZ + ((gapVLper + 10) * 2)) * 46)
}

// CalculateGCDRate calculates the GCD (Greatest Common Divisor) rate for a node
func CalculateGCDRate(gapGcd, ofdmbGcd int, mocaNodeVer int) int {
	if mocaNodeVer >= 0x20 {
		// MoCA 2.x: use 100MHz formula
		return (LDPC_LEN_100MHZ * ofdmbGcd) / ((FFT_LEN_100MHZ + ((gapGcd + 10) * 2)) * 46)
	}

	// MoCA 1.x: use 50MHz formula
	return (LDPC_LEN_50MHZ * ofdmbGcd) / ((FFT_LEN_50MHZ + (gapGcd*2 + 10)) * 26)
}

// CalculatePHYRates processes FMR data for a node and calculates all PHY rates
func CalculatePHYRates(
	entryNodeID int,
	fmrData []uint32,
	entryMocaVer int,
	ncMocaVer int,
	mocaNetVer int,
	nodeBitMask int,
	nodeVersions map[int]int,
) (*PHYRateMatrix, error) {
	matrix := NewPHYRateMatrix()
	matrix.NPER[entryNodeID] = make(map[int]int)
	matrix.VLPER[entryNodeID] = make(map[int]int)

	// Determine entry node payload version
	entryPayloadVer := min(entryMocaVer, ncMocaVer)

	parser := NewFMRPayloadParser(fmrData, entryNodeID, entryPayloadVer, ncMocaVer, mocaNetVer, nodeBitMask, nodeVersions)

	// Parse FMR data for each possible destination node
	for destNodeID := 0; destNodeID < MAX_NUM_NODES; destNodeID++ {
		gapNper, gapVLper, ofdmbNper, ofdmbVLper, err := parser.parseFMREntry(destNodeID)
		if err != nil {
			return nil, fmt.Errorf("failed to parse FMR entry for node %d->%d: %w", entryNodeID, destNodeID, err)
		}

		// Skip absent nodes
		if (nodeBitMask & (1 << destNodeID)) == 0 {
			continue
		}

		// Determine FMR payload version for calculation
		destMocaVer := nodeVersions[destNodeID]
		var fmrPayloadVer int
		if ncMocaVer < 0x20 {
			fmrPayloadVer = min(entryPayloadVer, destMocaVer)
		} else {
			fmrPayloadVer = entryMocaVer
		}

		// Calculate rates
		rateNper := CalculateNPERRate(gapNper, ofdmbNper, fmrPayloadVer, gapVLper)
		rateVLper := CalculateVLPERRate(gapVLper, ofdmbVLper)

		matrix.NPER[entryNodeID][destNodeID] = rateNper
		matrix.VLPER[entryNodeID][destNodeID] = rateVLper

		// Calculate GCD for self-to-self
		if entryNodeID == destNodeID {
			gcdRate := CalculateGCDRate(gapNper, ofdmbNper, entryMocaVer)
			matrix.GCD[entryNodeID] = gcdRate
		}
	}

	// Handle mixed-mode network GCD (2.x node in 1.x network)
	if mocaNetVer < 0x20 && ncMocaVer >= 0x20 && entryMocaVer >= 0x20 {
		gapGcd, ofdmbGcd, err := parser.parseGCDForMixedMode()
		if err == nil {
			matrix.GCD[entryNodeID] = (LDPC_LEN_50MHZ * ofdmbGcd) / ((FFT_LEN_50MHZ + (gapGcd*2 + 10)) * 26)
		}
	}

	return matrix, nil
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
