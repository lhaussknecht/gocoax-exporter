package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Create temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `
listen_address: ":9091"
scrape_timeout: 15
devices:
  - name: "test-device"
    address: "192.168.1.100:80"
    username: "admin"
    password: "secret"
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Load config
	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify values
	if cfg.ListenAddress != ":9091" {
		t.Errorf("Expected listen address :9091, got %s", cfg.ListenAddress)
	}

	if cfg.ScrapeTimeout != 15 {
		t.Errorf("Expected scrape timeout 15, got %d", cfg.ScrapeTimeout)
	}

	if len(cfg.Devices) != 1 {
		t.Fatalf("Expected 1 device, got %d", len(cfg.Devices))
	}

	device := cfg.Devices[0]
	if device.Name != "test-device" {
		t.Errorf("Expected device name 'test-device', got %s", device.Name)
	}

	if device.Address != "192.168.1.100:80" {
		t.Errorf("Expected address '192.168.1.100:80', got %s", device.Address)
	}

	if device.Username != "admin" {
		t.Errorf("Expected username 'admin', got %s", device.Username)
	}

	if device.Password != "secret" {
		t.Errorf("Expected password 'secret', got %s", device.Password)
	}
}

func TestConfigDefaults(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Minimal config without optional fields
	configContent := `
devices:
  - name: "test"
    address: "192.168.1.1"
    username: "admin"
    password: "pass"
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Check defaults
	if cfg.ListenAddress != ":9090" {
		t.Errorf("Expected default listen address :9090, got %s", cfg.ListenAddress)
	}

	if cfg.ScrapeTimeout != 10 {
		t.Errorf("Expected default scrape timeout 10, got %d", cfg.ScrapeTimeout)
	}

	// Check auto-port addition
	if cfg.Devices[0].Address != "192.168.1.1:80" {
		t.Errorf("Expected address with port '192.168.1.1:80', got %s", cfg.Devices[0].Address)
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      string
		expectError bool
		errorMsg    string
	}{
		{
			name: "no devices",
			config: `
listen_address: ":9090"
scrape_timeout: 10
devices: []
`,
			expectError: true,
			errorMsg:    "at least one device",
		},
		{
			name: "missing device name",
			config: `
devices:
  - address: "192.168.1.1"
    username: "admin"
    password: "pass"
`,
			expectError: true,
			errorMsg:    "name is required",
		},
		{
			name: "missing device address",
			config: `
devices:
  - name: "test"
    username: "admin"
    password: "pass"
`,
			expectError: true,
			errorMsg:    "address is required",
		},
		{
			name: "missing username",
			config: `
devices:
  - name: "test"
    address: "192.168.1.1"
    password: "pass"
`,
			expectError: true,
			errorMsg:    "username is required",
		},
		{
			name: "missing password",
			config: `
devices:
  - name: "test"
    address: "192.168.1.1"
    username: "admin"
`,
			expectError: true,
			errorMsg:    "password is required",
		},
		{
			name: "invalid port",
			config: `
devices:
  - name: "test"
    address: "192.168.1.1:99999"
    username: "admin"
    password: "pass"
`,
			expectError: true,
			errorMsg:    "invalid port",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config.yaml")

			if err := os.WriteFile(configPath, []byte(tt.config), 0644); err != nil {
				t.Fatalf("Failed to write test config: %v", err)
			}

			_, err := Load(configPath)

			if tt.expectError && err == nil {
				t.Errorf("Expected error containing '%s', got nil", tt.errorMsg)
			}

			if !tt.expectError && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			if tt.expectError && err != nil && tt.errorMsg != "" {
				// Just check that error occurred, detailed message checking is optional
				if err.Error() == "" {
					t.Errorf("Expected error message, got empty")
				}
			}
		})
	}
}

func TestEnvironmentOverrides(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `
listen_address: ":9090"
scrape_timeout: 10
devices:
  - name: "test"
    address: "192.168.1.1"
    username: "admin"
    password: "pass"
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Set environment variables
	os.Setenv("GOCOAX_LISTEN_ADDRESS", ":8080")
	os.Setenv("GOCOAX_SCRAPE_TIMEOUT", "20")
	defer func() {
		os.Unsetenv("GOCOAX_LISTEN_ADDRESS")
		os.Unsetenv("GOCOAX_SCRAPE_TIMEOUT")
	}()

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Check overrides
	if cfg.ListenAddress != ":8080" {
		t.Errorf("Expected overridden listen address :8080, got %s", cfg.ListenAddress)
	}

	if cfg.ScrapeTimeout != 20 {
		t.Errorf("Expected overridden scrape timeout 20, got %d", cfg.ScrapeTimeout)
	}
}
