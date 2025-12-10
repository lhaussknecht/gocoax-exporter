package collector

import (
	"fmt"
	"log"

	"github.com/louispool/gocoax-exporter/config"
	"github.com/prometheus/client_golang/prometheus"
)

// MultiDeviceRegistry manages collectors for multiple goCoax devices
type MultiDeviceRegistry struct {
	collectors []*GoCoaxCollector
}

// NewMultiDeviceRegistry creates a registry with collectors for all configured devices
func NewMultiDeviceRegistry(cfg *config.Config) (*MultiDeviceRegistry, error) {
	if len(cfg.Devices) == 0 {
		return nil, fmt.Errorf("no devices configured")
	}

	registry := &MultiDeviceRegistry{
		collectors: make([]*GoCoaxCollector, 0, len(cfg.Devices)),
	}

	timeout := cfg.GetTimeout()

	// Create a collector for each configured device
	for i, device := range cfg.Devices {
		collector, err := NewGoCoaxCollector(
			device.Name,
			device.Address,
			device.Username,
			device.Password,
			timeout,
		)
		if err != nil {
			log.Printf("Warning: failed to create collector for device %s: %v", device.Name, err)
			continue
		}

		registry.collectors = append(registry.collectors, collector)
		log.Printf("Created collector %d/%d for device: %s (%s)", i+1, len(cfg.Devices), device.Name, device.Address)
	}

	if len(registry.collectors) == 0 {
		return nil, fmt.Errorf("no collectors were successfully created")
	}

	return registry, nil
}

// Register registers all device collectors with the Prometheus registry
func (r *MultiDeviceRegistry) Register(registry *prometheus.Registry) error {
	for i, collector := range r.collectors {
		if err := registry.Register(collector); err != nil {
			return fmt.Errorf("failed to register collector %d: %w", i, err)
		}
	}

	log.Printf("Successfully registered %d device collector(s)", len(r.collectors))
	return nil
}

// Close releases resources for all collectors
func (r *MultiDeviceRegistry) Close() error {
	var lastErr error

	for _, collector := range r.collectors {
		if err := collector.Close(); err != nil {
			lastErr = err
			log.Printf("Error closing collector: %v", err)
		}
	}

	return lastErr
}

// GetCollectorCount returns the number of active collectors
func (r *MultiDeviceRegistry) GetCollectorCount() int {
	return len(r.collectors)
}
