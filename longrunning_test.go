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
	"fmt"
	"math/rand"
	"net/http"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// TestLongRunning simulates real-world authorization scenarios indefinitely.
// This test is designed to run continuously for testing Prometheus and Grafana integration.
// It generates realistic traffic patterns based on classic RBAC, ABAC, and ReBAC scenarios.
//
// To run this test:
//   go test -v -run TestLongRunning -timeout 0
//
// The test will:
// - Start a Prometheus metrics endpoint on http://localhost:8080/metrics
// - Continuously generate authorization events at a controlled rate
// - Simulate realistic allow/deny patterns
// - Generate policy operation events periodically
//
// Press Ctrl+C to stop the test.
func TestLongRunning(t *testing.T) {
	// Skip in normal test runs
	if testing.Short() {
		t.Skip("Skipping long-running test in short mode")
	}

	// Create a custom registry for this test
	registry := prometheus.NewRegistry()
	logger := NewPrometheusLoggerWithRegistry(registry)
	defer logger.UnregisterFrom(registry)

	// Create a new ServeMux to avoid global handler conflicts
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
	
	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	// Start HTTP server for metrics
	go func() {
		t.Logf("Starting metrics server on :8080")
		t.Logf("Visit http://localhost:8080/metrics to see the metrics")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			t.Logf("Metrics server stopped: %v", err)
		}
	}()

	// Wait for server to start
	time.Sleep(500 * time.Millisecond)

	t.Log("Starting long-running test. Press Ctrl+C to stop...")
	t.Log("Generating realistic authorization events based on RBAC, ABAC, and ReBAC patterns")

	// Create a local random number generator
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Run the simulation loop
	runSimulation(t, logger, rng)
}

// runSimulation runs the continuous simulation of authorization events
func runSimulation(t *testing.T, logger *PrometheusLogger, rng *rand.Rand) {
	// Track iterations for periodic events
	iteration := 0

	// Main simulation loop
	for {
		iteration++

		// Perform a batch of enforce checks (10-20 per iteration)
		batchSize := 10 + rng.Intn(11) // 10-20 requests
		for i := 0; i < batchSize; i++ {
			// Randomly select a scenario type
			scenarioType := rng.Intn(3)
			switch scenarioType {
			case 0:
				simulateRBACEnforce(logger, rng)
			case 1:
				simulateABACEnforce(logger, rng)
			case 2:
				simulateReBACEnforce(logger, rng)
			}
		}

		// Periodically simulate policy operations (every ~50 iterations)
		if iteration%50 == 0 {
			simulatePolicyOperation(logger, rng)
		}

		// Log progress every 100 iterations
		if iteration%100 == 0 {
			t.Logf("Completed %d iterations", iteration)
		}

		// Sleep to control request rate (100-300ms between batches)
		// This results in approximately 50-150 requests per second
		sleepTime := 100 + rng.Intn(200)
		time.Sleep(time.Duration(sleepTime) * time.Millisecond)
	}
}

// simulateRBACEnforce simulates Role-Based Access Control scenarios
// Based on classic RBAC model: users -> roles -> permissions
func simulateRBACEnforce(logger *PrometheusLogger, rng *rand.Rand) {
	// RBAC users and their typical roles
	users := []string{"alice", "bob", "charlie", "david", "eve"}
	resources := []string{"project1", "project2", "database", "config", "logs"}
	actions := []string{"read", "write", "delete", "admin"}
	domains := []string{"", "domain1", "domain2"} // Empty string for default domain

	// Select random parameters
	user := users[rng.Intn(len(users))]
	resource := resources[rng.Intn(len(resources))]
	action := actions[rng.Intn(len(actions))]
	domain := domains[rng.Intn(len(domains))]

	// Determine if access should be allowed based on realistic RBAC rules
	allowed := determineRBACAccess(user, resource, action, rng)

	// Create and log the enforce event
	entry := &LogEntry{
		EventType: EventEnforce,
		Subject:   user,
		Object:    resource,
		Action:    action,
		Domain:    domain,
	}

	logger.OnBeforeEvent(entry)
	// Simulate processing time (1-10ms)
	time.Sleep(time.Duration(1+rng.Intn(10)) * time.Millisecond)
	entry.Allowed = allowed
	logger.OnAfterEvent(entry)
}

// determineRBACAccess simulates RBAC policy evaluation
func determineRBACAccess(user, resource, action string, rng *rand.Rand) bool {
	// Simulate realistic RBAC rules
	// Alice is admin - can do anything
	if user == "alice" {
		return true
	}

	// Bob is developer - can read/write but not delete
	if user == "bob" {
		return action == "read" || action == "write"
	}

	// Charlie is viewer - can only read
	if user == "charlie" {
		return action == "read"
	}

	// David has limited access - can read project1 only
	if user == "david" {
		return action == "read" && resource == "project1"
	}

	// Eve is guest - very limited access
	if user == "eve" {
		// Only 30% chance of access (restricted)
		return rng.Float32() < 0.3
	}

	return false
}

// simulateABACEnforce simulates Attribute-Based Access Control scenarios
// Based on classic ABAC model: decisions based on attributes of subject, resource, action, and environment
func simulateABACEnforce(logger *PrometheusLogger, rng *rand.Rand) {
	// ABAC subjects with attributes
	subjects := []struct {
		name       string
		department string
		clearance  int
	}{
		{"alice", "engineering", 3},
		{"bob", "engineering", 2},
		{"charlie", "sales", 1},
		{"david", "hr", 2},
		{"eve", "guest", 0},
	}

	// ABAC resources with attributes
	resources := []struct {
		name              string
		classification    int
		ownerDepartment   string
	}{
		{"data1", 1, "engineering"},
		{"data2", 2, "engineering"},
		{"data3", 3, "engineering"},
		{"hr_records", 2, "hr"},
		{"sales_data", 1, "sales"},
	}

	actions := []string{"read", "write", "delete"}

	// Select random parameters
	subject := subjects[rng.Intn(len(subjects))]
	resource := resources[rng.Intn(len(resources))]
	action := actions[rng.Intn(len(actions))]

	// Determine if access should be allowed based on ABAC rules
	allowed := determineABACAccess(subject.clearance, subject.department, resource.classification, resource.ownerDepartment, action, rng)

	// Create and log the enforce event
	entry := &LogEntry{
		EventType: EventEnforce,
		Subject:   subject.name,
		Object:    resource.name,
		Action:    action,
		Domain:    subject.department,
	}

	logger.OnBeforeEvent(entry)
	// Simulate processing time (2-15ms for ABAC complexity)
	time.Sleep(time.Duration(2+rng.Intn(14)) * time.Millisecond)
	entry.Allowed = allowed
	logger.OnAfterEvent(entry)
}

// determineABACAccess simulates ABAC policy evaluation based on attributes
func determineABACAccess(clearance int, department string, classification int, ownerDept string, action string, rng *rand.Rand) bool {
	// Rule 1: Clearance level must be >= resource classification
	if clearance < classification {
		return false
	}

	// Rule 2: Can only write/delete resources in your own department
	if (action == "write" || action == "delete") && department != ownerDept {
		return false
	}

	// Rule 3: Guest department has very limited access
	if department == "guest" {
		return rng.Float32() < 0.2
	}

	// Rule 4: HR can access HR records regardless
	if department == "hr" && ownerDept == "hr" {
		return true
	}

	return true
}

// simulateReBACEnforce simulates Relationship-Based Access Control scenarios
// Based on classic ReBAC model: decisions based on relationships between entities
func simulateReBACEnforce(logger *PrometheusLogger, rng *rand.Rand) {
	// ReBAC subjects with relationships
	subjects := []struct {
		name         string
		organization string
	}{
		{"alice", "org1"},
		{"bob", "org1"},
		{"charlie", "org2"},
		{"david", "org2"},
		{"eve", "org3"},
	}

	// ReBAC resources with ownership relationships
	resources := []struct {
		name         string
		owner        string
		organization string
		isPublic     bool
	}{
		{"doc1", "alice", "org1", false},
		{"doc2", "bob", "org1", false},
		{"doc3", "charlie", "org2", false},
		{"public_doc", "alice", "org1", true},
		{"shared_doc", "david", "org2", true},
	}

	actions := []string{"read", "write", "share", "delete"}

	// Select random parameters
	subject := subjects[rng.Intn(len(subjects))]
	resource := resources[rng.Intn(len(resources))]
	action := actions[rng.Intn(len(actions))]

	// Determine if access should be allowed based on ReBAC rules
	allowed := determineReBACAccess(subject.name, subject.organization, resource.owner, resource.organization, resource.isPublic, action, rng)

	// Create and log the enforce event
	entry := &LogEntry{
		EventType: EventEnforce,
		Subject:   subject.name,
		Object:    resource.name,
		Action:    action,
		Domain:    subject.organization,
	}

	logger.OnBeforeEvent(entry)
	// Simulate processing time (2-12ms for relationship resolution)
	time.Sleep(time.Duration(2+rng.Intn(11)) * time.Millisecond)
	entry.Allowed = allowed
	logger.OnAfterEvent(entry)
}

// determineReBACAccess simulates ReBAC policy evaluation based on relationships
func determineReBACAccess(subject, subjectOrg, resourceOwner, resourceOrg string, isPublic bool, action string, rng *rand.Rand) bool {
	// Rule 1: Owner can do anything
	if subject == resourceOwner {
		return true
	}

	// Rule 2: Public resources can be read by anyone
	if isPublic && action == "read" {
		return true
	}

	// Rule 3: Same organization members can read
	if subjectOrg == resourceOrg && action == "read" {
		return true
	}

	// Rule 4: Same organization members can share
	if subjectOrg == resourceOrg && action == "share" {
		return rng.Float32() < 0.7 // 70% success rate
	}

	// Rule 5: Cross-organization access is mostly denied
	if subjectOrg != resourceOrg {
		return rng.Float32() < 0.1 // 10% success rate
	}

	return false
}

// simulatePolicyOperation simulates periodic policy operations
func simulatePolicyOperation(logger *PrometheusLogger, rng *rand.Rand) {
	operations := []EventType{
		EventAddPolicy,
		EventRemovePolicy,
		EventLoadPolicy,
		EventSavePolicy,
	}

	operation := operations[rng.Intn(len(operations))]
	ruleCount := 1 + rng.Intn(10) // 1-10 rules affected

	entry := &LogEntry{
		EventType: operation,
		RuleCount: ruleCount,
	}

	logger.OnBeforeEvent(entry)
	// Simulate processing time (5-30ms for policy operations)
	time.Sleep(time.Duration(5+rng.Intn(26)) * time.Millisecond)

	// Simulate occasional errors (5% failure rate)
	if rng.Float32() < 0.05 {
		entry.Error = fmt.Errorf("simulated policy operation error")
	}

	logger.OnAfterEvent(entry)
}
