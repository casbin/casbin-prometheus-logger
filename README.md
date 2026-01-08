# casbin-prometheus-logger

[![Go Report Card](https://goreportcard.com/badge/github.com/casbin/casbin-prometheus-logger)](https://goreportcard.com/report/github.com/casbin/casbin-prometheus-logger)
[![Go](https://github.com/casbin/casbin-prometheus-logger/actions/workflows/ci.yml/badge.svg)](https://github.com/casbin/casbin-prometheus-logger/actions/workflows/ci.yml)
[![Coverage Status](https://codecov.io/gh/casbin/casbin-prometheus-logger/branch/master/graph/badge.svg)](https://codecov.io/gh/casbin/casbin-prometheus-logger)
[![GoDoc](https://godoc.org/github.com/casbin/casbin-prometheus-logger?status.svg)](https://godoc.org/github.com/casbin/casbin-prometheus-logger)
[![Release](https://img.shields.io/github/release/casbin/casbin-prometheus-logger.svg)](https://github.com/casbin/casbin-prometheus-logger/releases/latest)
[![Discord](https://img.shields.io/discord/1022748306096537660?logo=discord&label=discord&color=5865F2)](https://discord.gg/S5UjpzGZjN)

A Prometheus logger implementation for [Casbin](https://github.com/casbin/casbin), providing event-driven metrics collection for authorization events.

## Features

- **Event-Driven Logging**: Implements the Casbin Logger interface with support for event-driven logging
- **Prometheus Metrics**: Exports comprehensive metrics for Casbin operations
- **Customizable Event Types**: Filter which event types to log
- **Custom Callbacks**: Add custom processing for log entries
- **Multiple Registries**: Support for both default and custom Prometheus registries

## Metrics Exported

### Enforce Metrics
- `casbin_enforce_total` - Total number of enforce requests (labeled by `allowed`, `domain`)
- `casbin_enforce_duration_seconds` - Duration of enforce requests (labeled by `allowed`, `domain`)

### Policy Operation Metrics
- `casbin_policy_operations_total` - Total number of policy operations (labeled by `operation`, `success`)
- `casbin_policy_operations_duration_seconds` - Duration of policy operations (labeled by `operation`)
- `casbin_policy_rules_count` - Number of policy rules affected by operations (labeled by `operation`)

## Installation

```bash
go get github.com/casbin/casbin-prometheus-logger
```

## Usage

### Basic Usage

```go
package main

import (
    "net/http"
    
    prometheuslogger "github.com/casbin/casbin-prometheus-logger"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
    // Create logger with default Prometheus registry
    logger := prometheuslogger.NewPrometheusLogger()
    defer logger.Unregister()
    
    // Or create with custom registry
    registry := prometheus.NewRegistry()
    logger := prometheuslogger.NewPrometheusLoggerWithRegistry(registry)
    defer logger.UnregisterFrom(registry)
    
    // Use with Casbin
    // enforcer.SetLogger(logger)
    
    // Expose metrics endpoint
    http.Handle("/metrics", promhttp.Handler())
    http.ListenAndServe(":8080", nil)
}
```

### Configure Event Types

```go
// Only log specific event types
logger.SetEventTypes([]prometheuslogger.EventType{
    prometheuslogger.EventEnforce,
    prometheuslogger.EventAddPolicy,
})
```

### Add Custom Callback

```go
// Add custom processing for log entries
logger.SetLogCallback(func(entry *prometheuslogger.LogEntry) error {
    fmt.Printf("Event: %s, Duration: %v\n", entry.EventType, entry.Duration)
    return nil
})
```

## Event Types

The logger supports the following event types:

- `EventEnforce` - Authorization enforcement requests
- `EventAddPolicy` - Policy addition operations
- `EventRemovePolicy` - Policy removal operations
- `EventLoadPolicy` - Policy loading operations
- `EventSavePolicy` - Policy saving operations

## Example

See the [examples/basic](examples/basic/main.go) directory for a complete working example.

To run the example:

```bash
cd examples/basic
go run main.go
```

Then visit http://localhost:8080/metrics to see the exported metrics.

## Quick Prometheus + Grafana Setup

1. Install Prometheus: follow the official guide at https://prometheus.io/docs/introduction/first_steps/. After starting it, add this service to `prometheus.yml` under `scrape_configs` (e.g., `http://<host>:8080/metrics`).
2. Install Grafana: follow https://grafana.com/docs/grafana/latest/setup-grafana/ and log in via the browser on the default port.
3. Import data in Grafana:
   - In "Connections", add a new Data Source and choose Prometheus.
   - Set the URL to your Prometheus endpoint (e.g., `http://localhost:9090`), then Save & Test.
   - In "Dashboards", click "Import"  or paste a JSON. Pick the Prometheus data source you just added to finish.

## License

This project is licensed under the Apache 2.0 License - see the [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Related Projects

- [Casbin](https://github.com/casbin/casbin) - An authorization library that supports access control models
- [Prometheus](https://prometheus.io/) - Monitoring system and time series database
