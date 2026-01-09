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

## Prometheus + Grafana Setup

This section guides you through setting up Prometheus and Grafana to visualize Casbin metrics.

### 1. Install and Configure Prometheus

1. **Install Prometheus**: Follow the official guide at https://prometheus.io/docs/introduction/first_steps/

2. **Configure Prometheus** to scrape metrics from your application. Edit your `prometheus.yml` configuration file and add a new job under `scrape_configs`:

```yaml
global:
  scrape_interval: 15s # Scrape targets every 15 seconds

scrape_configs:
  - job_name: "casbin-app"
    static_configs:
      - targets: ["localhost:8080"]  # Replace with your app's host:port
```

Replace `localhost:8080` with the actual address where your application exposes the `/metrics` endpoint.

3. **Start Prometheus** and verify it's scraping metrics by visiting `http://localhost:9090/targets` in your browser.

### 2. Install Grafana

Follow the official Grafana installation guide at https://grafana.com/docs/grafana/latest/setup-grafana/installation/ for your platform. After installation, access Grafana via your browser (default: `http://localhost:3000`) and log in with the default credentials (admin/admin).

### 3. Configure Grafana Data Source

1. In Grafana, navigate to **Connections** → **Data Sources**
2. Click **Add data source**
3. Select **Prometheus**
4. Set the URL to your Prometheus endpoint (e.g., `http://localhost:9090`)
5. Click **Save & Test** to verify the connection

### 4. Import Casbin Dashboard

A pre-built Grafana dashboard is available for visualizing Casbin metrics. You can import the dashboard JSON file from [grafana-dashboard.json](grafana-dashboard.json).

**To import the dashboard:**

1. In Grafana, go to **Dashboards** → **Import**
2. Upload the `grafana-dashboard.json` file or paste its contents
3. Select the Prometheus data source you configured in the previous step
4. Click **Import**

### Dashboard Panels

The dashboard includes the following panels organized into two sections:

#### Enforce Metrics
- **Total Enforce Rate** - Overall rate of enforce requests per second
- **Enforce Rate Detail (History)** - Historical view broken down by allowed/denied status and domain
- **Enforce Duration (Latency Distribution)** - Histogram showing p50, p90, p95, and p99 latencies
- **Enforce Duration by Status and Domain** - Average duration broken down by status and domain

#### Policy Metrics
- **Policy Operation Rate** - Current rate of policy operations (Add/Save/Load/Remove)
- **Policy Operations (Success/Failure)** - Pie chart showing success vs failure distribution
- **Policy Operation Rate History** - Historical view of policy operations activity
- **Policy Rules Affected History** - Trend of the number of policy rules affected over time
- **Policy Operation Duration (Latency Distribution)** - Histogram showing p50, p90, and p99 latencies for policy operations
- **Policy Operation Average Duration** - Average duration by operation type

## Examples

### Basic Example

See the [examples/basic](examples/basic/main.go) directory for a complete working example demonstrating basic usage.

To run the basic example:

```bash
cd examples/basic
go run main.go
```

Then visit http://localhost:8080/metrics to see the exported metrics.

### Long-Running Test Case

The repository includes a long-running test case that continuously simulates authorization requests for testing Prometheus and Grafana. This test generates realistic metrics based on classic Casbin patterns.

The test demonstrates:
- **RBAC** (Role-Based Access Control) scenarios (40%)
- **ABAC** (Attribute-Based Access Control) scenarios (30%)
- **ReBAC** (Relationship-Based Access Control) scenarios (20%)
- **Complex** scenarios combining multiple patterns (10%)
- Continuous operation for real-world monitoring
- Graceful shutdown handling

To run the long-running test:

```bash
go test -v -run TestLongRunning -timeout 0
```

The test will run continuously until interrupted with Ctrl+C, generating metrics that can be scraped by Prometheus at `http://localhost:8080/metrics` and visualized in Grafana.

For detailed information about the test scenarios, see the test documentation in [longrunning_test.go](longrunning_test.go).

## License

This project is licensed under the Apache 2.0 License - see the [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Related Projects

- [Casbin](https://github.com/casbin/casbin) - An authorization library that supports access control models
- [Prometheus](https://prometheus.io/) - Monitoring system and time series database
