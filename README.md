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
- `casbin_enforce_total` - Total number of enforce requests (labeled by `allowed`, `domain`, and optionally `subject`, `object`, `action`)
- `casbin_enforce_duration_seconds` - Duration of enforce requests (labeled by `allowed`, `domain`, and optionally `subject`, `object`, `action`)

### Policy Operation Metrics
- `casbin_policy_operations_total` - Total number of policy operations (labeled by `operation`, `success`)
- `casbin_policy_operations_duration_seconds` - Duration of policy operations (labeled by `operation`)
- `casbin_policy_rules_count` - Number of policy rules affected by operations (labeled by `operation`)

### Policy State Metrics
- `casbin_policy_state_count` - Current number of policy rules by type (labeled by `ptype`)

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

### Configure Optional Enforce Metric Labels

```go
// Create logger with additional labels for enforce metrics
options := &prometheuslogger.PrometheusLoggerOptions{
    EnforceLabels: []string{
        prometheuslogger.EnforceLabelSubject,
        prometheuslogger.EnforceLabelObject,
        prometheuslogger.EnforceLabelAction,
    },
}

registry := prometheus.NewRegistry()
logger := prometheuslogger.NewPrometheusLoggerWithOptions(registry, options)
defer logger.UnregisterFrom(registry)

// Enforce metrics will now include subject, object, and action labels
// in addition to the default allowed and domain labels
```

### Update Policy State Metrics

```go
// Update the current policy state count
// This helps monitor permission growth over time
logger.UpdatePolicyState("p", 100)   // 100 p-type policies
logger.UpdatePolicyState("g", 50)    // 50 g-type role assignments
logger.UpdatePolicyState("g1", 25)   // 25 g1-type role assignments
logger.UpdatePolicyState("g2", 10)   // 10 g2-type role assignments
logger.UpdatePolicyState("g3", 5)    // 5 g3-type role assignments
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

## License

This project is licensed under the Apache 2.0 License - see the [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Related Projects

- [Casbin](https://github.com/casbin/casbin) - An authorization library that supports access control models
- [Prometheus](https://prometheus.io/) - Monitoring system and time series database
