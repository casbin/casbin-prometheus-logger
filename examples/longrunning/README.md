# Long-Running Test Case

This example demonstrates a continuous authorization simulation that generates realistic metrics for Prometheus and Grafana monitoring. It simulates real-world authorization patterns based on classic Casbin use cases.

## Features

- **Continuous Operation**: Runs indefinitely to provide ongoing metrics
- **Multiple Authorization Models**:
  - **RBAC** (Role-Based Access Control): User roles with different permission levels
  - **ABAC** (Attribute-Based Access Control): Decisions based on user attributes (age, department, location)
  - **ReBAC** (Relationship-Based Access Control): Ownership, hierarchies, and group memberships
  - **Complex Scenarios**: Combined patterns with conditional access
- **Policy Changes**: Simulates occasional policy updates (add/remove/load/save)
- **Resource Friendly**: Configurable timing to avoid CPU/memory exhaustion
- **Graceful Shutdown**: Handles Ctrl+C and SIGTERM signals properly

## Use Cases

This example is designed for:
- Testing Prometheus + Grafana dashboards with realistic authorization data
- Load testing monitoring infrastructure
- Understanding authorization pattern distributions
- Developing and testing alert rules
- Demonstrating the casbin-prometheus-logger in action

## Running the Example

### Prerequisites

- Go 1.23 or higher
- Prometheus (optional, for metrics collection)
- Grafana (optional, for visualization)

### Basic Usage

```bash
cd examples/longrunning
go run main.go
```

The example will:
1. Start an HTTP server on port 8080
2. Expose metrics at `http://localhost:8080/metrics`
3. Begin continuous authorization simulation
4. Log each authorization decision to the console
5. Run until interrupted with Ctrl+C

### With Prometheus

1. Configure Prometheus to scrape the metrics endpoint:

```yaml
# prometheus.yml
scrape_configs:
  - job_name: 'casbin-longrunning'
    scrape_interval: 5s
    static_configs:
      - targets: ['localhost:8080']
```

2. Start Prometheus:

```bash
prometheus --config.file=prometheus.yml
```

3. Start the example:

```bash
go run main.go
```

4. Visit `http://localhost:9090` to view Prometheus

### With Grafana

After setting up Prometheus:

1. Add Prometheus as a data source in Grafana
2. Create a dashboard with queries like:
   - `rate(casbin_enforce_total[1m])` - Authorization requests per second
   - `histogram_quantile(0.95, rate(casbin_enforce_duration_seconds_bucket[1m]))` - 95th percentile latency
   - `casbin_enforce_total{allowed="true"}` - Allowed requests
   - `casbin_enforce_total{allowed="false"}` - Denied requests
   - `casbin_policy_operations_total` - Policy operations
   - `casbin_policy_rules_count` - Rules affected by operations

## Configuration

Edit constants in `main.go` to customize behavior:

```go
const (
    metricsPort        = ":8080"                  // Metrics endpoint port
    requestInterval    = 100 * time.Millisecond   // Time between requests
    policyChangeChance = 0.01                     // Policy change probability
)
```

## Authorization Scenarios

### RBAC Examples (40% of requests)
- Admin role: Full access (read, write, delete)
- Editor role: Read and write access
- Viewer role: Read-only access
- Multi-tenant domain separation

### ABAC Examples (30% of requests)
- Age-based access control
- Department-based permissions
- Time-based restrictions (office hours)
- Location-based access (office vs. remote)

### ReBAC Examples (20% of requests)
- Resource ownership
- Manager-employee hierarchies
- Group memberships
- Friend/follower relationships
- Parent-child resource inheritance

### Complex Examples (10% of requests)
- Combined RBAC + ABAC + ReBAC
- MFA-required operations
- Temporary/time-limited access
- Cross-domain administration

## Metrics Generated

The example generates the following Prometheus metrics:

- `casbin_enforce_total` - Total authorization requests (labeled by allowed, domain)
- `casbin_enforce_duration_seconds` - Request duration histogram
- `casbin_policy_operations_total` - Policy changes (labeled by operation, success)
- `casbin_policy_operations_duration_seconds` - Policy operation duration
- `casbin_policy_rules_count` - Number of rules in policy operations

## Example Output

```
Starting metrics server on :8080
Visit http://localhost:8080/metrics to see the metrics

=== Starting Long-Running Authorization Simulation ===
This simulation runs continuously to generate metrics for Prometheus/Grafana
Press Ctrl+C to stop...

[ENFORCE] alice read document1 (domain: org1) -> allowed: true, duration: 2.3ms
[ENFORCE] bob write document2 (domain: org1) -> allowed: true, duration: 1.8ms
[ENFORCE] charlie delete document3 (domain: org1) -> allowed: false, duration: 2.1ms
[addPolicy] rules: 3, duration: 15.2ms
[ENFORCE] eve read salary_data (domain: default) -> allowed: false, duration: 4.5ms
...
```

## Stopping the Example

Press `Ctrl+C` to gracefully stop the example. It will:
1. Stop accepting new simulation requests
2. Shut down the HTTP server
3. Clean up resources
4. Exit cleanly

## License

This example is part of the casbin-prometheus-logger project and follows the same Apache 2.0 License.
