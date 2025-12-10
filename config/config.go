package config

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	ListenAddress string   `yaml:"listen_address"`
	ScrapeTimeout int      `yaml:"scrape_timeout"` // Timeout in seconds
	Devices       []Device `yaml:"devices"`
}

// Device represents a single goCoax device configuration
type Device struct {
	Name     string `yaml:"name"`
	Address  string `yaml:"address"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

// Load reads configuration from a YAML file and applies defaults
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Apply defaults
	if cfg.ListenAddress == "" {
		cfg.ListenAddress = ":9090"
	}
	if cfg.ScrapeTimeout == 0 {
		cfg.ScrapeTimeout = 10
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Load from environment variables if set
	cfg.loadFromEnv()

	return cfg, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if len(c.Devices) == 0 {
		return fmt.Errorf("at least one device must be configured")
	}

	for i, device := range c.Devices {
		if device.Name == "" {
			return fmt.Errorf("device %d: name is required", i)
		}
		if device.Address == "" {
			return fmt.Errorf("device %d (%s): address is required", i, device.Name)
		}

		// Validate address format (should be host:port or just host)
		if !strings.Contains(device.Address, ":") {
			c.Devices[i].Address = device.Address + ":80" // Default to port 80
		}

		// Try to parse the address
		host, port, err := net.SplitHostPort(c.Devices[i].Address)
		if err != nil {
			return fmt.Errorf("device %d (%s): invalid address format: %w", i, device.Name, err)
		}

		// Validate host
		if host == "" {
			return fmt.Errorf("device %d (%s): empty hostname", i, device.Name)
		}

		// Validate port
		portNum, err := strconv.Atoi(port)
		if err != nil || portNum < 1 || portNum > 65535 {
			return fmt.Errorf("device %d (%s): invalid port number", i, device.Name)
		}

		if device.Username == "" {
			return fmt.Errorf("device %d (%s): username is required", i, device.Name)
		}
		if device.Password == "" {
			return fmt.Errorf("device %d (%s): password is required", i, device.Name)
		}
	}

	if c.ScrapeTimeout < 1 {
		return fmt.Errorf("scrape_timeout must be at least 1 second")
	}

	return nil
}

// loadFromEnv loads configuration overrides from environment variables
func (c *Config) loadFromEnv() {
	if addr := os.Getenv("GOCOAX_LISTEN_ADDRESS"); addr != "" {
		c.ListenAddress = addr
	}

	if timeout := os.Getenv("GOCOAX_SCRAPE_TIMEOUT"); timeout != "" {
		if t, err := strconv.Atoi(timeout); err == nil && t > 0 {
			c.ScrapeTimeout = t
		}
	}

	// Allow overriding device credentials via environment
	for i := range c.Devices {
		prefix := fmt.Sprintf("GOCOAX_DEVICE_%d_", i)

		if name := os.Getenv(prefix + "NAME"); name != "" {
			c.Devices[i].Name = name
		}
		if addr := os.Getenv(prefix + "ADDRESS"); addr != "" {
			c.Devices[i].Address = addr
		}
		if user := os.Getenv(prefix + "USERNAME"); user != "" {
			c.Devices[i].Username = user
		}
		if pass := os.Getenv(prefix + "PASSWORD"); pass != "" {
			c.Devices[i].Password = pass
		}
	}
}

// GetTimeout returns the scrape timeout as a time.Duration
func (c *Config) GetTimeout() time.Duration {
	return time.Duration(c.ScrapeTimeout) * time.Second
}
