# goCoax Prometheus Exporter

A Prometheus exporter for goCoax MoCA (Multimedia over Coax Alliance) network bridges.

## Features

- Multi-device support - monitor multiple goCoax devices from a single exporter instance
- MoCA 1.x, 2.0, and 2.5 protocol support
- PHY rate metrics (NPER, VLPER, GCD) between all nodes in the network
- Node topology and version information
- Network-level metrics

## Status

🚧 **Under Development** 🚧

This project is currently being implemented. Check back soon for updates.

## Metrics

The exporter will provide the following metrics:

- `gocoax_phy_rate_nper_mbps` - Normal Packet Error Rate PHY rate in Mbps
- `gocoax_phy_rate_vlper_mbps` - Very Low Packet Error Rate PHY rate in Mbps (MoCA 2.5)
- `gocoax_phy_rate_gcd_mbps` - Greatest Common Divisor rate in Mbps
- `gocoax_node_info` - Node information with MoCA version labels
- `gocoax_up` - Device availability (1 = up, 0 = down)
- `gocoax_scrape_duration_seconds` - Time taken to scrape device metrics

## Configuration

Configuration will be provided via YAML file. Example:

```yaml
listen_address: ":9090"
scrape_timeout: 10
devices:
  - name: "bridge-1"
    address: "192.168.98.50:80"
    username: "admin"
    password: "secret"
  - name: "bridge-2"
    address: "192.168.98.53:80"
    username: "admin"
    password: "secret"
```

## Installation

### From Source

```bash
go build -o gocoax-exporter
./gocoax-exporter -config config.yaml
```

## Usage

```bash
./gocoax-exporter -config config.yaml
```

The exporter will listen on `:9090` (default) and expose metrics at `/metrics`.

## Prometheus Configuration

Add the following to your `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'gocoax'
    static_configs:
      - targets: ['localhost:9090']
```

## License

MIT License (to be added)

## Development

See [CLAUDE.md](CLAUDE.md) for development guidance.
