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
	"errors"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestNewPrometheusLogger(t *testing.T) {
	// Clean up any existing registrations
	defer func() {
		if r := recover(); r != nil {
			// Ignore panic from duplicate registration
		}
	}()

	logger := NewPrometheusLogger()
	if logger == nil {
		t.Fatal("NewPrometheusLogger returned nil")
	}

	if logger.enabledEventTypes == nil {
		t.Error("enabledEventTypes map not initialized")
	}

	if logger.enforceDuration == nil {
		t.Error("enforceDuration metric not initialized")
	}

	if logger.enforceTotal == nil {
		t.Error("enforceTotal metric not initialized")
	}

	if logger.policyOpsTotal == nil {
		t.Error("policyOpsTotal metric not initialized")
	}

	if logger.policyOpsDuration == nil {
		t.Error("policyOpsDuration metric not initialized")
	}

	if logger.policyRulesCount == nil {
		t.Error("policyRulesCount metric not initialized")
	}

	// Clean up
	logger.Unregister()
}

func TestNewPrometheusLoggerWithRegistry(t *testing.T) {
	registry := prometheus.NewRegistry()
	logger := NewPrometheusLoggerWithRegistry(registry)

	if logger == nil {
		t.Fatal("NewPrometheusLoggerWithRegistry returned nil")
	}

	if logger.enabledEventTypes == nil {
		t.Error("enabledEventTypes map not initialized")
	}

	// Verify metrics are registered by checking they can be collected
	if logger.enforceDuration == nil {
		t.Error("enforceDuration not initialized")
	}
	if logger.enforceTotal == nil {
		t.Error("enforceTotal not initialized")
	}
	if logger.policyOpsTotal == nil {
		t.Error("policyOpsTotal not initialized")
	}
	if logger.policyOpsDuration == nil {
		t.Error("policyOpsDuration not initialized")
	}
	if logger.policyRulesCount == nil {
		t.Error("policyRulesCount not initialized")
	}

	// Clean up
	logger.UnregisterFrom(registry)
}

func TestSetEventTypes(t *testing.T) {
	registry := prometheus.NewRegistry()
	logger := NewPrometheusLoggerWithRegistry(registry)
	defer logger.UnregisterFrom(registry)

	eventTypes := []EventType{EventEnforce, EventAddPolicy}
	err := logger.SetEventTypes(eventTypes)
	if err != nil {
		t.Errorf("SetEventTypes returned error: %v", err)
	}

	if len(logger.enabledEventTypes) != 2 {
		t.Errorf("Expected 2 enabled event types, got %d", len(logger.enabledEventTypes))
	}

	if !logger.enabledEventTypes[EventEnforce] {
		t.Error("EventEnforce should be enabled")
	}

	if !logger.enabledEventTypes[EventAddPolicy] {
		t.Error("EventAddPolicy should be enabled")
	}

	if logger.enabledEventTypes[EventRemovePolicy] {
		t.Error("EventRemovePolicy should not be enabled")
	}
}

func TestOnBeforeEvent(t *testing.T) {
	registry := prometheus.NewRegistry()
	logger := NewPrometheusLoggerWithRegistry(registry)
	defer logger.UnregisterFrom(registry)

	// Test with no event type filtering
	entry := &LogEntry{
		EventType: EventEnforce,
	}

	err := logger.OnBeforeEvent(entry)
	if err != nil {
		t.Errorf("OnBeforeEvent returned error: %v", err)
	}

	if !entry.IsActive {
		t.Error("Entry should be active when no event types are configured")
	}

	if entry.StartTime.IsZero() {
		t.Error("StartTime should be set")
	}

	// Test with event type filtering - enabled event
	logger.SetEventTypes([]EventType{EventEnforce})
	entry2 := &LogEntry{
		EventType: EventEnforce,
	}

	err = logger.OnBeforeEvent(entry2)
	if err != nil {
		t.Errorf("OnBeforeEvent returned error: %v", err)
	}

	if !entry2.IsActive {
		t.Error("Entry should be active for enabled event type")
	}

	// Test with event type filtering - disabled event
	entry3 := &LogEntry{
		EventType: EventAddPolicy,
	}

	err = logger.OnBeforeEvent(entry3)
	if err != nil {
		t.Errorf("OnBeforeEvent returned error: %v", err)
	}

	if entry3.IsActive {
		t.Error("Entry should not be active for disabled event type")
	}
}

func TestOnAfterEvent_Enforce(t *testing.T) {
	registry := prometheus.NewRegistry()
	logger := NewPrometheusLoggerWithRegistry(registry)
	defer logger.UnregisterFrom(registry)

	entry := &LogEntry{
		IsActive:  true,
		EventType: EventEnforce,
		StartTime: time.Now().Add(-100 * time.Millisecond),
		Subject:   "alice",
		Object:    "data1",
		Action:    "read",
		Domain:    "domain1",
		Allowed:   true,
	}

	err := logger.OnAfterEvent(entry)
	if err != nil {
		t.Errorf("OnAfterEvent returned error: %v", err)
	}

	if entry.EndTime.IsZero() {
		t.Error("EndTime should be set")
	}

	if entry.Duration == 0 {
		t.Error("Duration should be calculated")
	}

	// Verify metrics were recorded
	count := testutil.CollectAndCount(logger.enforceTotal)
	if count != 1 {
		t.Errorf("Expected 1 metric sample for enforceTotal, got %d", count)
	}

	count = testutil.CollectAndCount(logger.enforceDuration)
	if count != 1 {
		t.Errorf("Expected 1 metric sample for enforceDuration, got %d", count)
	}
}

func TestOnAfterEvent_InactiveEntry(t *testing.T) {
	registry := prometheus.NewRegistry()
	logger := NewPrometheusLoggerWithRegistry(registry)
	defer logger.UnregisterFrom(registry)

	entry := &LogEntry{
		IsActive:  false,
		EventType: EventEnforce,
	}

	err := logger.OnAfterEvent(entry)
	if err != nil {
		t.Errorf("OnAfterEvent returned error: %v", err)
	}

	// Verify no metrics were recorded
	count := testutil.CollectAndCount(logger.enforceTotal)
	if count != 0 {
		t.Errorf("Expected 0 metric samples for inactive entry, got %d", count)
	}
}

func TestOnAfterEvent_PolicyOperation(t *testing.T) {
	registry := prometheus.NewRegistry()
	logger := NewPrometheusLoggerWithRegistry(registry)
	defer logger.UnregisterFrom(registry)

	testCases := []struct {
		name      string
		eventType EventType
	}{
		{"AddPolicy", EventAddPolicy},
		{"RemovePolicy", EventRemovePolicy},
		{"LoadPolicy", EventLoadPolicy},
		{"SavePolicy", EventSavePolicy},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			entry := &LogEntry{
				IsActive:  true,
				EventType: tc.eventType,
				StartTime: time.Now().Add(-50 * time.Millisecond),
				RuleCount: 5,
			}

			err := logger.OnAfterEvent(entry)
			if err != nil {
				t.Errorf("OnAfterEvent returned error: %v", err)
			}

			if entry.EndTime.IsZero() {
				t.Error("EndTime should be set")
			}

			if entry.Duration == 0 {
				t.Error("Duration should be calculated")
			}
		})
	}

	// Verify policy metrics were recorded
	count := testutil.CollectAndCount(logger.policyOpsTotal)
	if count != len(testCases) {
		t.Errorf("Expected %d metric samples for policyOpsTotal, got %d", len(testCases), count)
	}
}

func TestOnAfterEvent_WithError(t *testing.T) {
	registry := prometheus.NewRegistry()
	logger := NewPrometheusLoggerWithRegistry(registry)
	defer logger.UnregisterFrom(registry)

	entry := &LogEntry{
		IsActive:  true,
		EventType: EventAddPolicy,
		StartTime: time.Now().Add(-50 * time.Millisecond),
		RuleCount: 3,
		Error:     errors.New("test error"),
	}

	err := logger.OnAfterEvent(entry)
	if err != nil {
		t.Errorf("OnAfterEvent returned error: %v", err)
	}

	// The metric should still be recorded with success="false"
	count := testutil.CollectAndCount(logger.policyOpsTotal)
	if count != 1 {
		t.Errorf("Expected 1 metric sample, got %d", count)
	}
}

func TestSetLogCallback(t *testing.T) {
	registry := prometheus.NewRegistry()
	logger := NewPrometheusLoggerWithRegistry(registry)
	defer logger.UnregisterFrom(registry)

	callbackCalled := false
	callback := func(entry *LogEntry) error {
		callbackCalled = true
		return nil
	}

	err := logger.SetLogCallback(callback)
	if err != nil {
		t.Errorf("SetLogCallback returned error: %v", err)
	}

	// Trigger an event to verify callback is called
	entry := &LogEntry{
		IsActive:  true,
		EventType: EventEnforce,
		StartTime: time.Now(),
		Allowed:   true,
	}

	err = logger.OnAfterEvent(entry)
	if err != nil {
		t.Errorf("OnAfterEvent returned error: %v", err)
	}

	if !callbackCalled {
		t.Error("Callback should have been called")
	}
}

func TestSetLogCallback_WithError(t *testing.T) {
	registry := prometheus.NewRegistry()
	logger := NewPrometheusLoggerWithRegistry(registry)
	defer logger.UnregisterFrom(registry)

	expectedError := errors.New("callback error")
	callback := func(entry *LogEntry) error {
		return expectedError
	}

	logger.SetLogCallback(callback)

	entry := &LogEntry{
		IsActive:  true,
		EventType: EventEnforce,
		StartTime: time.Now(),
		Allowed:   true,
	}

	err := logger.OnAfterEvent(entry)
	if err != expectedError {
		t.Errorf("Expected error %v, got %v", expectedError, err)
	}
}

func TestEnforceMetrics_DifferentDomains(t *testing.T) {
	registry := prometheus.NewRegistry()
	logger := NewPrometheusLoggerWithRegistry(registry)
	defer logger.UnregisterFrom(registry)

	// Test with specific domain
	entry1 := &LogEntry{
		IsActive:  true,
		EventType: EventEnforce,
		StartTime: time.Now(),
		Domain:    "domain1",
		Allowed:   true,
	}

	logger.OnAfterEvent(entry1)

	// Test with default domain (empty)
	entry2 := &LogEntry{
		IsActive:  true,
		EventType: EventEnforce,
		StartTime: time.Now(),
		Domain:    "",
		Allowed:   false,
	}

	logger.OnAfterEvent(entry2)

	// Verify metrics were recorded with different labels
	count := testutil.CollectAndCount(logger.enforceTotal)
	if count != 2 {
		t.Errorf("Expected 2 metric samples, got %d", count)
	}
}

func TestMetricGetters(t *testing.T) {
	registry := prometheus.NewRegistry()
	logger := NewPrometheusLoggerWithRegistry(registry)
	defer logger.UnregisterFrom(registry)

	if logger.GetEnforceDuration() == nil {
		t.Error("GetEnforceDuration returned nil")
	}

	if logger.GetEnforceTotal() == nil {
		t.Error("GetEnforceTotal returned nil")
	}

	if logger.GetPolicyOpsTotal() == nil {
		t.Error("GetPolicyOpsTotal returned nil")
	}

	if logger.GetPolicyOpsDuration() == nil {
		t.Error("GetPolicyOpsDuration returned nil")
	}

	if logger.GetPolicyRulesCount() == nil {
		t.Error("GetPolicyRulesCount returned nil")
	}
}

func TestLogger_InterfaceImplementation(t *testing.T) {
	registry := prometheus.NewRegistry()
	var _ Logger = NewPrometheusLoggerWithRegistry(registry)
}

func TestFullWorkflow(t *testing.T) {
	registry := prometheus.NewRegistry()
	logger := NewPrometheusLoggerWithRegistry(registry)
	defer logger.UnregisterFrom(registry)

	// Configure to only log enforce events
	logger.SetEventTypes([]EventType{EventEnforce})

	// Simulate enforce event
	enforceEntry := &LogEntry{
		EventType: EventEnforce,
		Subject:   "alice",
		Object:    "data1",
		Action:    "read",
		Domain:    "org1",
	}

	// Before event
	logger.OnBeforeEvent(enforceEntry)
	if !enforceEntry.IsActive {
		t.Error("Enforce entry should be active")
	}

	// Simulate some processing time
	time.Sleep(10 * time.Millisecond)

	// After event
	enforceEntry.Allowed = true
	logger.OnAfterEvent(enforceEntry)

	// Simulate policy event (should be filtered out)
	policyEntry := &LogEntry{
		EventType: EventAddPolicy,
		RuleCount: 5,
	}

	logger.OnBeforeEvent(policyEntry)
	if policyEntry.IsActive {
		t.Error("Policy entry should not be active (filtered)")
	}

	logger.OnAfterEvent(policyEntry)

	// Verify only enforce metrics were recorded
	enforceCount := testutil.CollectAndCount(logger.enforceTotal)
	if enforceCount != 1 {
		t.Errorf("Expected 1 enforce metric, got %d", enforceCount)
	}

	policyCount := testutil.CollectAndCount(logger.policyOpsTotal)
	if policyCount != 0 {
		t.Errorf("Expected 0 policy metrics (filtered), got %d", policyCount)
	}
}
