# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This project aims to create a Prometheus exporter for goCoax MoCA (Multimedia over Coax Alliance) devices. goCoax manufactures coaxial network bridges that use MoCA technology to create high-speed networks over existing coax cable infrastructure.

## Target Device API

The goCoax devices expose a web interface with JSON endpoints that provide network status and PHY (physical layer) rate information. Based on the example HTML file in `examples/`:

- The device web interface is accessible at `http://192.168.98.50/` (default IP)
- Key endpoints include:
  - `/ms/0/0x15` - Local device information (node ID, MoCA version, network info)
  - `/ms/0/0x16` - Network node information (takes node ID as parameter)
  - `/ms/0/0x1D` - FMR (Frame Management Request) information for PHY rates

- Metrics available include:
  - PHY rates between nodes (unicast NPER and VLPER rates in Mbps)
  - MoCA protocol version (1.x, 2.0, 2.5)
  - Node bitmask (active nodes on the network)
  - Network coordinator (NC) information

## MoCA Technology Context

MoCA networks support multiple protocol versions (1.x, 2.0, 2.5) with different capabilities:
- MoCA 1.x: 50MHz bandwidth, basic PHY rates
- MoCA 2.0: 100MHz bandwidth, NPER (Normal Packet Error Rate) support
- MoCA 2.5: Enhanced features, both NPER and VLPER (Very Low Packet Error Rate) modes

PHY rates are calculated using the FMR payload data with version-specific formulas (see lines 193-234 in the example HTML).

## Future Development Structure

When implementing this exporter, the typical structure should include:

- **Scraper/Collector**: HTTP client to fetch data from goCoax device endpoints
- **Parser**: Parse device responses and calculate metrics (PHY rate formulas are complex and version-dependent)
- **Exporter**: Expose metrics in Prometheus format (typically on `:9090` or similar port)
- **Configuration**: Support for device IP, port, authentication (if required), and scrape interval

## Important Implementation Notes

- The device API appears to use POST requests with form data (e.g., `data` and `data2` parameters)
- Network topology is dynamic - use the node bitmask to determine active nodes
- PHY rates are directional (from node X to node Y may differ from Y to X)
- Handle mixed-version networks where different nodes run different MoCA versions
- The example shows the device returns data in JSON format (via `doFormGetJSON` functions)
