# casbin-prometheus-logger

[![Go Report Card](https://goreportcard.com/badge/github.com/casbin/casbin-prometheus-logger)](https://goreportcard.com/report/github.com/casbin/casbin-prometheus-logger)
[![Go](https://github.com/casbin/casbin-prometheus-logger/actions/workflows/ci.yml/badge.svg)](https://github.com/casbin/casbin-prometheus-logger/actions/workflows/ci.yml)
[![Coverage Status](https://codecov.io/gh/casbin/casbin-prometheus-logger/branch/master/graph/badge.svg)](https://codecov.io/gh/casbin/casbin-prometheus-logger)
[![GoDoc](https://godoc.org/github.com/casbin/casbin-prometheus-logger?status.svg)](https://godoc.org/github.com/casbin/casbin-prometheus-logger)
[![Release](https://img.shields.io/github/release/casbin/casbin-prometheus-logger.svg)](https://github.com/casbin/casbin-prometheus-logger/releases/latest)
[![Discord](https://img.shields.io/discord/1022748306096537660?logo=discord&label=discord&color=5865F2)](https://discord.gg/S5UjpzGZjN)

Prometheus metrics for [Casbin](https://github.com/casbin/casbin) authorization enforcement. Track which policies are being enforced, how often access is granted or denied, and how long enforcement takes.

## Why?

When running Casbin in production, you want visibility into:
- Authorization patterns (who's accessing what, and are they allowed?)
- Performance (is policy enforcement slowing down your app?)
- Policy changes (how many rules are being added/removed?)

This logger plugs into Casbin and exports all that data as Prometheus metrics.

## Features

- Track enforcement requests (allowed/denied, by domain)
- Monitor policy operations (add, remove, load, save)
- Filter specific event types
- Add custom callbacks for log entries
- Works with custom Prometheus registries

## Metrics

**Enforcement:**
- `casbin_enforce_total` - Count of requests, labeled by `allowed` (true/false) and `domain`
- `casbin_enforce_duration_seconds` - Request duration histogram

**Policy Operations:**
- `casbin_policy_operations_total` - Count of operations (add/remove/load/save), labeled by `operation` and `success`
- `casbin_policy_operations_duration_seconds` - Operation duration histogram
- `casbin_policy_rules_count` - Number of rules affected per operation

## Installation

```bash
go get github.com/casbin/casbin-prometheus-logger
```

## Usage

### Basic Setup

```go
package main

import (
    "net/http"
    
    prometheuslogger "github.com/casbin/casbin-prometheus-logger"
    "github.com/casbin/casbin/v2"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
    // Create logger
    logger := prometheuslogger.NewPrometheusLogger()
    defer logger.Unregister()
    
    // Attach to Casbin enforcer
    enforcer, err := casbin.NewEnforcer("model.conf", "policy.csv")
    if err != nil {
        panic(err)
    }
    enforcer.SetLogger(logger)
    
    // Now your enforcement calls automatically generate metrics
    enforcer.Enforce("alice", "data1", "read")
    
    // Expose metrics
    http.Handle("/metrics", promhttp.Handler())
    http.ListenAndServe(":8080", nil)
}
```

### With Custom Registry

```go
// Useful when you need separate metrics per service or integrate with existing monitoring
registry := prometheus.NewRegistry()
logger := prometheuslogger.NewPrometheusLoggerWithRegistry(registry)
defer logger.UnregisterFrom(registry)
```

### Filter Events

Only track specific operations:

```go
logger.SetEventTypes([]prometheuslogger.EventType{
    prometheuslogger.EventEnforce,
    prometheuslogger.EventAddPolicy,
})
```

### Custom Processing

Hook into the logging pipeline:

```go
logger.SetLogCallback(func(entry *prometheuslogger.LogEntry) error {
    // Your custom logic here
    fmt.Printf("Event: %s, Duration: %v\n", entry.EventType, entry.Duration)
    return nil
})
```

## Event Types

- `EventEnforce` - Authorization checks
- `EventAddPolicy` - Adding rules
- `EventRemovePolicy` - Removing rules
- `EventLoadPolicy` - Loading policy from storage
- `EventSavePolicy` - Saving policy to storage

## Example

Check out [examples/basic/main.go](examples/basic/main.go) for a complete working example.

```bash
cd examples/basic
go run main.go
# Visit http://localhost:8080/metrics
```

## License

This project is licensed under the Apache 2.0 License - see the [LICENSE](LICENSE) file for details.

## Contributing

Found a bug or have an idea? Open an issue or send a PR. For major changes, open an issue first to discuss what you'd like to change.

## Related Projects

- [Casbin](https://github.com/casbin/casbin) - Authorization library
- [Prometheus](https://prometheus.io/) - Monitoring and alerting
