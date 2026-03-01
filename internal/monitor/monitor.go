// Copyright (c) 2026 Kartoza
// SPDX-License-Identifier: MIT

// Package monitor coordinates Gatus endpoint monitoring
package monitor

import (
	"context"
	"fmt"
	"sync"

	"github.com/kartoza/gatus-monitor/internal/config"
	"github.com/kartoza/gatus-monitor/internal/gatus"
	"github.com/kartoza/gatus-monitor/internal/scheduler"
)

// OverallStatus represents the overall health status
type OverallStatus int

const (
	// StatusGreen indicates no errors across all endpoints
	StatusGreen OverallStatus = iota
	// StatusOrange indicates 1-2 errors on any endpoint
	StatusOrange
	// StatusRed indicates 3+ errors on any endpoint
	StatusRed
)

// String returns the string representation of the status
func (s OverallStatus) String() string {
	switch s {
	case StatusGreen:
		return "green"
	case StatusOrange:
		return "orange"
	case StatusRed:
		return "red"
	default:
		return "unknown"
	}
}

// StatusCallback is called when the overall status changes
type StatusCallback func(status OverallStatus, details map[string]*gatus.EndpointStatus)

// Monitor coordinates monitoring of multiple Gatus endpoints
type Monitor struct {
	config         *config.Manager
	scheduler      *scheduler.Scheduler
	clients        map[string]*gatus.Client
	statuses       map[string]*gatus.EndpointStatus
	overallStatus  OverallStatus
	statusCallback StatusCallback
	mu             sync.RWMutex
}

// New creates a new monitor
func New(cfg *config.Manager, callback StatusCallback) *Monitor {
	return &Monitor{
		config:         cfg,
		scheduler:      nil,
		clients:        make(map[string]*gatus.Client),
		statuses:       make(map[string]*gatus.EndpointStatus),
		overallStatus:  StatusGreen,
		statusCallback: callback,
	}
}

// Start begins monitoring all configured endpoints
func (m *Monitor) Start() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	cfg := m.config.Get()

	// Create scheduler with configured interval
	m.scheduler = scheduler.New(m.config.GetQueryInterval())

	// Create clients and tasks for each instance
	for _, instance := range cfg.Instances {
		// Use the base URL - the client will append the API path
		client := gatus.NewClient(instance.URL)
		m.clients[instance.Name] = client

		// Add task for this instance
		m.scheduler.AddTask(instance.Name, m.queryEndpoint)
	}

	// Start scheduler
	m.scheduler.Start()

	return nil
}

// Stop stops monitoring
func (m *Monitor) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.scheduler != nil {
		m.scheduler.Stop()
	}
}

// Restart restarts monitoring with current configuration
func (m *Monitor) Restart() error {
	m.Stop()
	return m.Start()
}

// UpdateConfig updates the configuration and restarts monitoring
func (m *Monitor) UpdateConfig(cfg *config.Config) error {
	if err := m.config.Update(cfg); err != nil {
		return fmt.Errorf("failed to update config: %w", err)
	}

	return m.Restart()
}

// queryEndpoint queries a single Gatus endpoint
func (m *Monitor) queryEndpoint(ctx context.Context, name string) error {
	m.mu.RLock()
	client, exists := m.clients[name]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("client not found for name: %s", name)
	}

	// Query the endpoint
	status, err := client.GetStatus(ctx)
	if err != nil {
		// Store error status
		m.mu.Lock()
		m.statuses[name] = status
		m.mu.Unlock()

		// Update overall status
		m.updateOverallStatus()
		return err
	}

	// Store successful status
	m.mu.Lock()
	m.statuses[name] = status
	m.mu.Unlock()

	// Update overall status
	m.updateOverallStatus()

	return nil
}

// updateOverallStatus calculates and updates the overall status
func (m *Monitor) updateOverallStatus() {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Determine maximum error count across all endpoints
	maxErrors := 0
	for _, status := range m.statuses {
		if !status.Reachable {
			// Unreachable endpoint counts as high error
			maxErrors = 3
			break
		}
		if status.ErrorCount > maxErrors {
			maxErrors = status.ErrorCount
		}
	}

	// Determine overall status based on error count
	var newStatus OverallStatus
	if maxErrors == 0 {
		newStatus = StatusGreen
	} else if maxErrors >= 1 && maxErrors <= 2 {
		newStatus = StatusOrange
	} else {
		newStatus = StatusRed
	}

	// Create a copy of statuses for callback
	statusCopy := make(map[string]*gatus.EndpointStatus)
	for k, v := range m.statuses {
		// Create a copy of the status
		statusCopy[k] = &gatus.EndpointStatus{
			URL:         v.URL,
			ErrorCount:  v.ErrorCount,
			LastChecked: v.LastChecked,
			LastError:   v.LastError,
			Reachable:   v.Reachable,
		}
	}

	// Update current status
	m.overallStatus = newStatus

	// Always notify callback so individual endpoint statuses can be updated
	go m.notifyStatusChange(newStatus, statusCopy)
}

// notifyStatusChange calls the status callback
func (m *Monitor) notifyStatusChange(status OverallStatus, details map[string]*gatus.EndpointStatus) {
	if m.statusCallback != nil {
		m.statusCallback(status, details)
	}
}

// GetOverallStatus returns the current overall status
func (m *Monitor) GetOverallStatus() OverallStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.overallStatus
}

// GetEndpointStatuses returns the current status of all endpoints
func (m *Monitor) GetEndpointStatuses() map[string]*gatus.EndpointStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	statusCopy := make(map[string]*gatus.EndpointStatus)
	for k, v := range m.statuses {
		statusCopy[k] = &gatus.EndpointStatus{
			URL:         v.URL,
			ErrorCount:  v.ErrorCount,
			LastChecked: v.LastChecked,
			LastError:   v.LastError,
			Reachable:   v.Reachable,
		}
	}

	return statusCopy
}

// GetStatusSummary returns a human-readable summary of the current status
func (m *Monitor) GetStatusSummary() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	configuredCount := len(m.config.Get().Instances)
	if configuredCount == 0 {
		return "No endpoints configured"
	}

	totalErrors := 0
	unreachable := 0

	for _, status := range m.statuses {
		if !status.Reachable {
			unreachable++
		} else {
			totalErrors += status.ErrorCount
		}
	}

	summary := fmt.Sprintf("Overall: %s | Endpoints: %d",
		m.overallStatus.String(), configuredCount)

	if unreachable > 0 {
		summary += fmt.Sprintf(" | Unreachable: %d", unreachable)
	}

	if totalErrors > 0 {
		summary += fmt.Sprintf(" | Errors: %d", totalErrors)
	}

	return summary
}

// GetConfiguredEndpointCount returns the number of configured endpoints
func (m *Monitor) GetConfiguredEndpointCount() int {
	return len(m.config.Get().Instances)
}
