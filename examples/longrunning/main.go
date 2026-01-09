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

package main

import (
	"context"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	prometheuslogger "github.com/casbin/casbin-prometheus-logger"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Configuration for the long-running test
const (
	metricsPort        = ":8080"
	requestInterval    = 100 * time.Millisecond // Time between simulated requests
	policyChangeChance = 0.01                   // 1% chance of policy change per iteration
)

func main() {
	// Create a custom Prometheus registry
	registry := prometheus.NewRegistry()

	// Create a new PrometheusLogger with the custom registry
	logger := prometheuslogger.NewPrometheusLoggerWithRegistry(registry)
	defer logger.UnregisterFrom(registry)

	// Set a custom callback for logging
	err := logger.SetLogCallback(func(entry *prometheuslogger.LogEntry) error {
		if entry.EventType == prometheuslogger.EventEnforce {
			log.Printf("[ENFORCE] %s %s %s (domain: %s) -> allowed: %v, duration: %v",
				entry.Subject, entry.Action, entry.Object, entry.Domain, entry.Allowed, entry.Duration)
		} else {
			log.Printf("[%s] rules: %d, duration: %v", entry.EventType, entry.RuleCount, entry.Duration)
		}
		return nil
	})
	if err != nil {
		log.Fatalf("Failed to set callback: %v", err)
	}

	// Set up HTTP server for metrics
	http.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))

	server := &http.Server{Addr: metricsPort}
	go func() {
		log.Printf("Starting metrics server on %s", metricsPort)
		log.Printf("Visit http://localhost%s/metrics to see the metrics", metricsPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start metrics server: %v", err)
		}
	}()

	// Set up graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("\nReceived shutdown signal, gracefully shutting down...")
		cancel()
		if err := server.Shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down server: %v", err)
		}
	}()

	// Initialize random number generator for simulation
	// Note: Using math/rand (not crypto/rand) is intentional for this simulation
	// as we don't need cryptographic randomness for test data generation
	randSource := rand.NewSource(time.Now().UnixNano())
	rng := rand.New(randSource)

	// Run the continuous authorization simulation
	log.Println("\n=== Starting Long-Running Authorization Simulation ===")
	log.Println("This simulation runs continuously to generate metrics for Prometheus/Grafana")
	log.Println("Press Ctrl+C to stop...")
	log.Println()

	runContinuousSimulation(ctx, logger, rng)

	log.Println("Simulation stopped.")
}

// runContinuousSimulation runs an endless loop of authorization checks
func runContinuousSimulation(ctx context.Context, logger *prometheuslogger.PrometheusLogger, rng *rand.Rand) {
	ticker := time.NewTicker(requestInterval)
	defer ticker.Stop()

	iterationCount := 0

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			iterationCount++

			// Simulate different authorization patterns
			switch iterationCount % 10 {
			case 0, 1, 2, 3:
				// 40% RBAC scenarios
				simulateRBACRequest(logger, rng)
			case 4, 5, 6:
				// 30% ABAC scenarios
				simulateABACRequest(logger, rng)
			case 7, 8:
				// 20% ReBAC scenarios
				simulateReBACRequest(logger, rng)
			case 9:
				// 10% mixed/complex scenarios
				simulateComplexRequest(logger, rng)
			}

			// Occasionally simulate policy changes
			if rng.Float64() < policyChangeChance {
				simulatePolicyChange(logger, rng)
			}
		}
	}
}

// simulateRBACRequest simulates role-based access control scenarios
func simulateRBACRequest(logger *prometheuslogger.PrometheusLogger, rng *rand.Rand) {
	// Classic RBAC: user -> role -> permission
	scenarios := []struct {
		subject string
		role    string
		object  string
		action  string
		domain  string
		allowed bool
	}{
		// Admin role - should have broad access
		{"alice", "admin", "document1", "read", "org1", true},
		{"alice", "admin", "document1", "write", "org1", true},
		{"alice", "admin", "document1", "delete", "org1", true},

		// Editor role - can read and write but not delete
		{"bob", "editor", "document2", "read", "org1", true},
		{"bob", "editor", "document2", "write", "org1", true},
		{"bob", "editor", "document2", "delete", "org1", false},

		// Viewer role - read only
		{"charlie", "viewer", "document3", "read", "org1", true},
		{"charlie", "viewer", "document3", "write", "org1", false},
		{"charlie", "viewer", "document3", "delete", "org1", false},

		// Multi-tenant scenarios
		{"dave", "admin", "document4", "read", "org2", true},
		{"eve", "viewer", "document5", "read", "org2", true},
		{"eve", "viewer", "document6", "read", "org1", false}, // Wrong domain
	}

	scenario := scenarios[rng.Intn(len(scenarios))]

	entry := &prometheuslogger.LogEntry{
		EventType: prometheuslogger.EventEnforce,
		Subject:   scenario.subject,
		Object:    scenario.object,
		Action:    scenario.action,
		Domain:    scenario.domain,
	}

	logger.OnBeforeEvent(entry)
	// Simulate processing time (variable)
	time.Sleep(time.Duration(1+rng.Intn(5)) * time.Millisecond)
	entry.Allowed = scenario.allowed
	logger.OnAfterEvent(entry)
}

// simulateABACRequest simulates attribute-based access control scenarios
func simulateABACRequest(logger *prometheuslogger.PrometheusLogger, rng *rand.Rand) {
	// ABAC: decisions based on attributes (age, department, time, etc.)
	scenarios := []struct {
		subject    string
		attributes map[string]interface{}
		object     string
		action     string
		domain     string
		allowed    bool
	}{
		// Age-based access
		{"alice", map[string]interface{}{"age": 30, "dept": "engineering"}, "confidential_doc", "read", "default", true},
		{"bob", map[string]interface{}{"age": 17, "dept": "intern"}, "confidential_doc", "read", "default", false},

		// Department-based access
		{"charlie", map[string]interface{}{"dept": "hr"}, "salary_data", "read", "default", true},
		{"dave", map[string]interface{}{"dept": "marketing"}, "salary_data", "read", "default", false},

		// Time-based access (office hours)
		{"eve", map[string]interface{}{"current_time": "14:00"}, "time_sensitive_data", "read", "default", true},
		{"frank", map[string]interface{}{"current_time": "22:00"}, "time_sensitive_data", "read", "default", false},

		// Location-based access
		{"grace", map[string]interface{}{"location": "office"}, "internal_tool", "use", "default", true},
		{"henry", map[string]interface{}{"location": "remote"}, "internal_tool", "use", "default", false},
	}

	scenario := scenarios[rng.Intn(len(scenarios))]

	entry := &prometheuslogger.LogEntry{
		EventType: prometheuslogger.EventEnforce,
		Subject:   scenario.subject,
		Object:    scenario.object,
		Action:    scenario.action,
		Domain:    scenario.domain,
	}

	logger.OnBeforeEvent(entry)
	// ABAC typically has slightly longer processing time due to attribute evaluation
	time.Sleep(time.Duration(2+rng.Intn(8)) * time.Millisecond)
	entry.Allowed = scenario.allowed
	logger.OnAfterEvent(entry)
}

// simulateReBACRequest simulates relationship-based access control scenarios
func simulateReBACRequest(logger *prometheuslogger.PrometheusLogger, rng *rand.Rand) {
	// ReBAC: decisions based on relationships (ownership, hierarchy, groups)
	scenarios := []struct {
		subject      string
		relationship string
		object       string
		action       string
		domain       string
		allowed      bool
	}{
		// Ownership relationships
		{"alice", "owner", "project1", "delete", "default", true},
		{"bob", "member", "project1", "delete", "default", false},
		{"bob", "member", "project1", "read", "default", true},

		// Hierarchical relationships
		{"manager_alice", "manager_of", "employee_bob", "view_profile", "default", true},
		{"employee_bob", "reports_to", "manager_alice", "view_profile", "default", false},

		// Group membership
		{"charlie", "member_of", "engineering_team", "access_repo", "default", true},
		{"dave", "member_of", "marketing_team", "access_repo", "default", false},

		// Friend relationships (social network style)
		{"eve", "friend_of", "frank", "view_photos", "default", true},
		{"grace", "follower_of", "frank", "view_photos", "default", false},

		// Parent-child resource relationships
		{"henry", "owner", "folder1", "read", "default", true},
		{"henry", "owner", "folder1/subfolder/file", "read", "default", true}, // Inherited
		{"iris", "viewer", "folder1/subfolder/file", "write", "default", false},
	}

	scenario := scenarios[rng.Intn(len(scenarios))]

	entry := &prometheuslogger.LogEntry{
		EventType: prometheuslogger.EventEnforce,
		Subject:   scenario.subject,
		Object:    scenario.object,
		Action:    scenario.action,
		Domain:    scenario.domain,
	}

	logger.OnBeforeEvent(entry)
	// ReBAC may involve graph traversal, so slightly longer processing
	time.Sleep(time.Duration(3+rng.Intn(10)) * time.Millisecond)
	entry.Allowed = scenario.allowed
	logger.OnAfterEvent(entry)
}

// simulateComplexRequest simulates complex authorization scenarios
func simulateComplexRequest(logger *prometheuslogger.PrometheusLogger, rng *rand.Rand) {
	// Complex scenarios combining RBAC, ABAC, and ReBAC
	scenarios := []struct {
		subject string
		object  string
		action  string
		domain  string
		allowed bool
		desc    string
	}{
		// Role + Attribute + Relationship
		{"alice", "sensitive_project", "deploy", "production", true, "Senior engineer in production domain"},
		{"junior_bob", "sensitive_project", "deploy", "production", false, "Junior engineer, lacks permission"},

		// Cross-domain with roles
		{"admin_charlie", "global_config", "modify", "global", true, "Global admin"},
		{"local_dave", "global_config", "modify", "org1", false, "Local admin, no global access"},

		// Conditional access (e.g., requires MFA)
		{"eve_with_mfa", "financial_data", "transfer", "default", true, "MFA verified"},
		{"frank_no_mfa", "financial_data", "transfer", "default", false, "MFA required but not provided"},

		// Temporary access (time-limited)
		{"grace_contractor", "project_alpha", "read", "default", true, "Active contract"},
		{"henry_ex_contractor", "project_alpha", "read", "default", false, "Contract expired"},
	}

	scenario := scenarios[rng.Intn(len(scenarios))]

	entry := &prometheuslogger.LogEntry{
		EventType: prometheuslogger.EventEnforce,
		Subject:   scenario.subject,
		Object:    scenario.object,
		Action:    scenario.action,
		Domain:    scenario.domain,
	}

	logger.OnBeforeEvent(entry)
	// Complex scenarios may take longer to evaluate
	time.Sleep(time.Duration(5+rng.Intn(15)) * time.Millisecond)
	entry.Allowed = scenario.allowed
	logger.OnAfterEvent(entry)
}

// simulatePolicyChange simulates policy management operations
func simulatePolicyChange(logger *prometheuslogger.PrometheusLogger, rng *rand.Rand) {
	operations := []struct {
		eventType prometheuslogger.EventType
		ruleCount int
	}{
		{prometheuslogger.EventAddPolicy, rng.Intn(5) + 1},
		{prometheuslogger.EventRemovePolicy, rng.Intn(3) + 1},
		{prometheuslogger.EventLoadPolicy, rng.Intn(100) + 50},
		{prometheuslogger.EventSavePolicy, rng.Intn(100) + 50},
	}

	op := operations[rng.Intn(len(operations))]

	entry := &prometheuslogger.LogEntry{
		EventType: op.eventType,
		RuleCount: op.ruleCount,
	}

	logger.OnBeforeEvent(entry)
	// Policy operations typically take longer
	time.Sleep(time.Duration(10+rng.Intn(30)) * time.Millisecond)
	logger.OnAfterEvent(entry)
}
