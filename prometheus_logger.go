// Copyright 2026 The casbin Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package prometheuslogger

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// PrometheusLogger is a logger that exports metrics to Prometheus.
type PrometheusLogger struct {
	enabledEventTypes map[EventType]bool
	callback          func(entry *LogEntry) error

	// Configuration
	enforceLabels []string // Optional labels for enforce metrics (e.g., "subject", "object", "action")

	// Prometheus metrics
	enforceDuration    *prometheus.HistogramVec
	enforceTotal       *prometheus.CounterVec
	policyOpsTotal     *prometheus.CounterVec
	policyOpsDuration  *prometheus.HistogramVec
	policyRulesCount   *prometheus.GaugeVec
	policyStateCount   *prometheus.GaugeVec // Current count of policies by type
}

// NewPrometheusLogger creates a new PrometheusLogger with default metrics.
func NewPrometheusLogger() *PrometheusLogger {
	return NewPrometheusLoggerWithOptions(nil, nil)
}

// NewPrometheusLoggerWithRegistry creates a new PrometheusLogger with a custom registry.
func NewPrometheusLoggerWithRegistry(registry *prometheus.Registry) *PrometheusLogger {
	return NewPrometheusLoggerWithOptions(registry, nil)
}

// PrometheusLoggerOptions provides configuration options for the logger.
type PrometheusLoggerOptions struct {
	// EnforceLabels specifies optional labels for enforce metrics.
	// Valid values: "subject", "object", "action"
	// By default, only "allowed" and "domain" labels are used.
	EnforceLabels []string
}

// NewPrometheusLoggerWithOptions creates a new PrometheusLogger with custom options.
// If registry is nil, the default Prometheus registry is used.
// If options is nil, default options are used.
func NewPrometheusLoggerWithOptions(registry *prometheus.Registry, options *PrometheusLoggerOptions) *PrometheusLogger {
	if options == nil {
		options = &PrometheusLoggerOptions{}
	}

	// Build enforce label list: always include "allowed" and "domain"
	enforceLabels := []string{"allowed", "domain"}
	for _, label := range options.EnforceLabels {
		if label == "subject" || label == "object" || label == "action" {
			enforceLabels = append(enforceLabels, label)
		}
	}

	logger := &PrometheusLogger{
		enabledEventTypes: make(map[EventType]bool),
		enforceLabels:     enforceLabels,
		enforceDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "casbin_enforce_duration_seconds",
				Help:    "Duration of enforce requests in seconds",
				Buckets: prometheus.DefBuckets,
			},
			enforceLabels,
		),
		enforceTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "casbin_enforce_total",
				Help: "Total number of enforce requests",
			},
			enforceLabels,
		),
		policyOpsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "casbin_policy_operations_total",
				Help: "Total number of policy operations",
			},
			[]string{"operation", "success"},
		),
		policyOpsDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "casbin_policy_operations_duration_seconds",
				Help:    "Duration of policy operations in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"operation"},
		),
		policyRulesCount: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "casbin_policy_rules_count",
				Help: "Number of policy rules affected by operations",
			},
			[]string{"operation"},
		),
		policyStateCount: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "casbin_policy_state_count",
				Help: "Current number of policy rules by type",
			},
			[]string{"ptype"},
		),
	}

	// Register all metrics with the provided registry or default
	if registry != nil {
		registry.MustRegister(
			logger.enforceDuration,
			logger.enforceTotal,
			logger.policyOpsTotal,
			logger.policyOpsDuration,
			logger.policyRulesCount,
			logger.policyStateCount,
		)
	} else {
		prometheus.MustRegister(
			logger.enforceDuration,
			logger.enforceTotal,
			logger.policyOpsTotal,
			logger.policyOpsDuration,
			logger.policyRulesCount,
			logger.policyStateCount,
		)
	}

	return logger
}

// SetEventTypes configures which event types should be logged.
func (p *PrometheusLogger) SetEventTypes(eventTypes []EventType) error {
	p.enabledEventTypes = make(map[EventType]bool)
	for _, eventType := range eventTypes {
		p.enabledEventTypes[eventType] = true
	}
	return nil
}

// OnBeforeEvent is called before an event occurs.
func (p *PrometheusLogger) OnBeforeEvent(entry *LogEntry) error {
	if len(p.enabledEventTypes) > 0 && !p.enabledEventTypes[entry.EventType] {
		entry.IsActive = false
		return nil
	}

	entry.IsActive = true
	entry.StartTime = time.Now()
	return nil
}

// OnAfterEvent is called after an event completes and records metrics.
func (p *PrometheusLogger) OnAfterEvent(entry *LogEntry) error {
	if !entry.IsActive {
		return nil
	}

	entry.EndTime = time.Now()
	entry.Duration = entry.EndTime.Sub(entry.StartTime)

	// Record metrics based on event type
	switch entry.EventType {
	case EventEnforce:
		p.recordEnforceMetrics(entry)
	case EventAddPolicy, EventRemovePolicy, EventLoadPolicy, EventSavePolicy:
		p.recordPolicyMetrics(entry)
	}

	// Call custom callback if set
	if p.callback != nil {
		return p.callback(entry)
	}

	return nil
}

// SetLogCallback sets a custom callback function for log entries.
func (p *PrometheusLogger) SetLogCallback(callback func(entry *LogEntry) error) error {
	p.callback = callback
	return nil
}

// recordEnforceMetrics records metrics for enforce events.
func (p *PrometheusLogger) recordEnforceMetrics(entry *LogEntry) {
	domain := entry.Domain
	if domain == "" {
		domain = "default"
	}

	allowed := "false"
	if entry.Allowed {
		allowed = "true"
	}

	// Build label values based on configured labels
	labelValues := make([]string, len(p.enforceLabels))
	for i, label := range p.enforceLabels {
		switch label {
		case "allowed":
			labelValues[i] = allowed
		case "domain":
			labelValues[i] = domain
		case "subject":
			labelValues[i] = entry.Subject
		case "object":
			labelValues[i] = entry.Object
		case "action":
			labelValues[i] = entry.Action
		}
	}

	p.enforceDuration.WithLabelValues(labelValues...).Observe(entry.Duration.Seconds())
	p.enforceTotal.WithLabelValues(labelValues...).Inc()
}

// recordPolicyMetrics records metrics for policy operation events.
func (p *PrometheusLogger) recordPolicyMetrics(entry *LogEntry) {
	operation := string(entry.EventType)
	success := "true"
	if entry.Error != nil {
		success = "false"
	}

	p.policyOpsTotal.WithLabelValues(operation, success).Inc()
	p.policyOpsDuration.WithLabelValues(operation).Observe(entry.Duration.Seconds())

	if entry.RuleCount > 0 {
		p.policyRulesCount.WithLabelValues(operation).Set(float64(entry.RuleCount))
	}
}

// UpdatePolicyState updates the current policy state count for a given policy type.
// ptype should be one of: "p", "g", "g1", "g2", "g3", etc.
// count is the current number of policies of that type.
func (p *PrometheusLogger) UpdatePolicyState(ptype string, count int) {
	p.policyStateCount.WithLabelValues(ptype).Set(float64(count))
}

// Unregister unregisters all metrics from the default Prometheus registry.
// This is useful for testing or when you need to recreate the logger.
func (p *PrometheusLogger) Unregister() {
	prometheus.Unregister(p.enforceDuration)
	prometheus.Unregister(p.enforceTotal)
	prometheus.Unregister(p.policyOpsTotal)
	prometheus.Unregister(p.policyOpsDuration)
	prometheus.Unregister(p.policyRulesCount)
	prometheus.Unregister(p.policyStateCount)
}

// UnregisterFrom unregisters all metrics from a specific Prometheus registry.
func (p *PrometheusLogger) UnregisterFrom(registry *prometheus.Registry) bool {
	result := true
	result = registry.Unregister(p.enforceDuration) && result
	result = registry.Unregister(p.enforceTotal) && result
	result = registry.Unregister(p.policyOpsTotal) && result
	result = registry.Unregister(p.policyOpsDuration) && result
	result = registry.Unregister(p.policyRulesCount) && result
	result = registry.Unregister(p.policyStateCount) && result
	return result
}

// GetEnforceDuration returns the enforce duration histogram metric.
func (p *PrometheusLogger) GetEnforceDuration() *prometheus.HistogramVec {
	return p.enforceDuration
}

// GetEnforceTotal returns the enforce total counter metric.
func (p *PrometheusLogger) GetEnforceTotal() *prometheus.CounterVec {
	return p.enforceTotal
}

// GetPolicyOpsTotal returns the policy operations total counter metric.
func (p *PrometheusLogger) GetPolicyOpsTotal() *prometheus.CounterVec {
	return p.policyOpsTotal
}

// GetPolicyOpsDuration returns the policy operations duration histogram metric.
func (p *PrometheusLogger) GetPolicyOpsDuration() *prometheus.HistogramVec {
	return p.policyOpsDuration
}

// GetPolicyRulesCount returns the policy rules count gauge metric.
func (p *PrometheusLogger) GetPolicyRulesCount() *prometheus.GaugeVec {
	return p.policyRulesCount
}

// GetPolicyStateCount returns the policy state count gauge metric.
func (p *PrometheusLogger) GetPolicyStateCount() *prometheus.GaugeVec {
	return p.policyStateCount
}
