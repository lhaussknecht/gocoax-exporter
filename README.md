# goCoax Prometheus Exporter

A Prometheus exporter for goCoax MoCA (Multimedia over Coax Alliance) network bridges that exposes detailed PHY layer performance metrics.

## Features

- **Multi-device support** - Monitor multiple goCoax devices from a single exporter instance
- **MoCA 1.x, 2.0, and 2.5 protocol support** - Full support for all MoCA versions with version-specific calculations
- **PHY rate metrics** - NPER, VLPER, and GCD rates between all nodes in the network
- **Node topology** - Automatic discovery of network nodes and MoCA versions
- **Resilient operation** - Retry logic with exponential backoff for transient failures
- **Graceful degradation** - Continue operation even if some devices or nodes fail
- **Web interface** - Built-in landing page showing exporter status and configuration

## Status

✅ **Production Ready** - Fully implemented and tested

## Metrics

The exporter provides the following metrics:

### PHY Rate Metrics

- **`gocoax_phy_rate_nper_mbps`** - Normal Packet Error Rate PHY rate in Mbps between nodes
  - Labels: `device`, `from_node`, `to_node`
  - Represents the data rate for normal reliability transmissions

- **`gocoax_phy_rate_vlper_mbps`** - Very Low Packet Error Rate PHY rate in Mbps (MoCA 2.5)
  - Labels: `device`, `from_node`, `to_node`
  - Only available on MoCA 2.5 networks, represents the data rate for high-reliability transmissions

- **`gocoax_phy_rate_gcd_mbps`** - Greatest Common Divisor rate in Mbps
  - Labels: `device`, `node`
  - Represents the broadcast/multicast rate for a node

### Node and Device Metrics

- **`gocoax_node_info`** - Node information (value always 1)
  - Labels: `device`, `node`, `moca_version` (e.g., "2.5", "2.0", "1.1"), `is_nc` ("true"/"false")
  - Provides topology and version information for each node

- **`gocoax_up`** - Device availability indicator (1 = up, 0 = down)
  - Labels: `device`
  - Indicates whether the device is reachable and responding

- **`gocoax_scrape_duration_seconds`** - Time taken to scrape device metrics
  - Labels: `device`
  - Useful for monitoring exporter performance

- **`gocoax_scrape_errors_total`** - Total number of scrape errors
  - Labels: `device`
  - Counter metric for tracking failures

### Example Metrics Output

```
# HELP gocoax_phy_rate_nper_mbps Normal Packet Error Rate PHY rate in Mbps between nodes
# TYPE gocoax_phy_rate_nper_mbps gauge
gocoax_phy_rate_nper_mbps{device="bridge-50",from_node="0",to_node="1"} 2983
gocoax_phy_rate_nper_mbps{device="bridge-50",from_node="1",to_node="0"} 3488

# HELP gocoax_node_info Node information with MoCA version
# TYPE gocoax_node_info gauge
gocoax_node_info{device="bridge-50",is_nc="true",moca_version="2.5",node="0"} 1
gocoax_node_info{device="bridge-50",is_nc="false",moca_version="2.5",node="1"} 1

# HELP gocoax_up Device is reachable and responding
# TYPE gocoax_up gauge
gocoax_up{device="bridge-50"} 1
```

## Configuration

Configuration is provided via a YAML file. See `examples/config.yaml.example` for a complete example.

### Configuration File

Create a `config.yaml` file:

```yaml
# Address and port for the exporter to listen on (default: ":9090")
listen_address: ":9090"

# Timeout for scraping device metrics in seconds (default: 10)
scrape_timeout: 10

# List of goCoax devices to monitor
devices:
  - name: "bridge-50"              # Friendly name for labels
    address: "192.168.98.50:80"    # Device IP and port (port defaults to 80 if omitted)
    username: "admin"               # HTTP Basic Auth username
    password: "your-password"       # HTTP Basic Auth password

  - name: "bridge-53"
    address: "192.168.98.53:80"
    username: "admin"
    password: "your-password"
```

### Environment Variables

All configuration options can be overridden with environment variables:

- `GOCOAX_LISTEN_ADDRESS` - Override listen address
- `GOCOAX_SCRAPE_TIMEOUT` - Override scrape timeout (seconds)
- `GOCOAX_DEVICE_0_NAME` - Override first device name
- `GOCOAX_DEVICE_0_ADDRESS` - Override first device address
- `GOCOAX_DEVICE_0_USERNAME` - Override first device username
- `GOCOAX_DEVICE_0_PASSWORD` - Override first device password

(Repeat for `DEVICE_1_`, `DEVICE_2_`, etc.)

## Installation

### From Source

Requirements:
- Go 1.21 or later

```bash
git clone https://github.com/louispool/gocoax-exporter
cd gocoax-exporter
go build -o gocoax-exporter
```

### Binary Release

Download pre-built binaries from the [Releases](https://github.com/louispool/gocoax-exporter/releases) page (coming soon).

## Usage

### Running the Exporter

```bash
# With config file
./gocoax-exporter -config config.yaml

# Show version
./gocoax-exporter -version

# Get help
./gocoax-exporter -help
```

### Command-line Flags

- `-config` - Path to configuration file (default: `config.yaml`)
- `-version` - Show version and exit

### Endpoints

Once running, the exporter provides the following HTTP endpoints:

- **`http://localhost:9090/metrics`** - Prometheus metrics endpoint
- **`http://localhost:9090/health`** - Health check endpoint (returns `OK`)
- **`http://localhost:9090/`** - Landing page with status information

## Prometheus Configuration

Add the goCoax exporter to your `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'gocoax'
    static_configs:
      - targets: ['localhost:9090']
    scrape_interval: 30s  # Adjust based on your needs
    scrape_timeout: 10s
```

### Example PromQL Queries

```promql
# Total PHY rate capacity from node 0
sum(gocoax_phy_rate_nper_mbps{from_node="0"})

# Average PHY rate across all links
avg(gocoax_phy_rate_nper_mbps)

# Nodes by MoCA version
count by (moca_version) (gocoax_node_info)

# Devices that are down
gocoax_up == 0

# Slow scrapes (taking longer than 5 seconds)
gocoax_scrape_duration_seconds > 5
```

## Understanding MoCA PHY Rates

### NPER vs VLPER

- **NPER (Normal Packet Error Rate)**: Standard reliability mode for most traffic
- **VLPER (Very Low Packet Error Rate)**: High-reliability mode available in MoCA 2.5, offering lower error rates at potentially lower data rates
- **GCD (Greatest Common Divisor)**: The multicast/broadcast rate that all nodes can receive

### Asymmetric Rates

PHY rates are directional and often asymmetric. A link from node 0 to node 1 may have a different rate than from node 1 to node 0 due to:
- Different RF conditions in each direction
- Asymmetric interference
- Node-specific capabilities

## Troubleshooting

### Exporter won't start

- Check that the configuration file path is correct
- Verify YAML syntax with `yamllint config.yaml`
- Ensure device addresses are reachable: `ping 192.168.98.50`

### No metrics for a device

- Check `gocoax_up` metric - if 0, the device is unreachable
- Verify credentials are correct
- Check device is responding: `curl -u admin:password http://192.168.98.50/`
- Look at exporter logs for specific error messages

### Metrics seem incorrect

- PHY rates are calculated using the MoCA specification formulas
- Rates vary based on network conditions and are expected to change
- Compare with the device's web interface PHY Rates page
- Ensure you're comparing the correct direction (from_node -> to_node)

## Development

### Building from Source

```bash
# Build
go build -o gocoax-exporter

# Run tests
go test ./...

# Run with race detection
go run -race main.go -config config.yaml
```

### Project Structure

```
gocoax-exporter/
├── main.go              # HTTP server and application entry point
├── client/              # goCoax device API client
│   └── client.go
├── collector/           # Prometheus collector implementation
│   ├── collector.go     # Main collector logic
│   ├── phyrate.go       # PHY rate calculation engine
│   └── registry.go      # Multi-device registry
├── config/              # Configuration management
│   └── config.go
└── examples/            # Example files and reference data
    ├── config.yaml.example
    └── PHY Rates.html   # Reference web interface
```

See [CLAUDE.md](CLAUDE.md) for detailed development guidance.

### Running Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run with verbose output
go test -v ./...
```

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass (`go test ./...`)
6. Commit your changes
7. Push to the branch
8. Open a Pull Request

## License

MIT License

Copyright (c) 2025

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.

## Acknowledgments

- Built with [Prometheus Go client library](https://github.com/prometheus/client_golang)
- MoCA specification formulas derived from the example device web interface
- Developed with assistance from [Claude Code](https://claude.com/claude-code)
