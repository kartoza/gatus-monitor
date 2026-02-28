package monitor

import (
	"testing"
	"time"

	"github.com/kartoza/gatus-monitor/internal/config"
	"github.com/kartoza/gatus-monitor/internal/gatus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOverallStatus_String(t *testing.T) {
	tests := []struct {
		status   OverallStatus
		expected string
	}{
		{StatusGreen, "green"},
		{StatusOrange, "orange"},
		{StatusRed, "red"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.status.String())
		})
	}
}

func TestNew(t *testing.T) {
	cfg := &config.Manager{}
	callback := func(status OverallStatus, details map[string]*gatus.EndpointStatus) {}

	monitor := New(cfg, callback)

	require.NotNil(t, monitor)
	assert.NotNil(t, monitor.clients)
	assert.NotNil(t, monitor.statuses)
	assert.Equal(t, StatusGreen, monitor.overallStatus)
}

func TestGetOverallStatus(t *testing.T) {
	cfg := &config.Manager{}
	monitor := New(cfg, nil)

	status := monitor.GetOverallStatus()
	assert.Equal(t, StatusGreen, status)
}

func TestGetEndpointStatuses(t *testing.T) {
	cfg := &config.Manager{}
	monitor := New(cfg, nil)

	// Add some test statuses
	monitor.statuses["url1"] = &gatus.EndpointStatus{
		URL:        "url1",
		ErrorCount: 1,
		Reachable:  true,
	}

	statuses := monitor.GetEndpointStatuses()
	assert.Len(t, statuses, 1)
	assert.Equal(t, "url1", statuses["url1"].URL)
	assert.Equal(t, 1, statuses["url1"].ErrorCount)
}

func TestUpdateOverallStatus_Green(t *testing.T) {
	cfg := &config.Manager{}

	statusChanged := false
	var receivedStatus OverallStatus

	callback := func(status OverallStatus, details map[string]*gatus.EndpointStatus) {
		statusChanged = true
		receivedStatus = status
	}

	monitor := New(cfg, callback)

	// Add endpoints with no errors
	monitor.statuses["url1"] = &gatus.EndpointStatus{
		URL:        "url1",
		ErrorCount: 0,
		Reachable:  true,
	}
	monitor.statuses["url2"] = &gatus.EndpointStatus{
		URL:        "url2",
		ErrorCount: 0,
		Reachable:  true,
	}

	monitor.updateOverallStatus()

	// Give callback time to execute
	time.Sleep(10 * time.Millisecond)

	assert.Equal(t, StatusGreen, monitor.GetOverallStatus())
	assert.False(t, statusChanged) // Status didn't change from initial green
}

func TestUpdateOverallStatus_Orange(t *testing.T) {
	cfg := &config.Manager{}

	statusChanged := false
	var receivedStatus OverallStatus

	callback := func(status OverallStatus, details map[string]*gatus.EndpointStatus) {
		statusChanged = true
		receivedStatus = status
	}

	monitor := New(cfg, callback)

	// Add endpoints with 1-2 errors
	monitor.statuses["url1"] = &gatus.EndpointStatus{
		URL:        "url1",
		ErrorCount: 1,
		Reachable:  true,
	}
	monitor.statuses["url2"] = &gatus.EndpointStatus{
		URL:        "url2",
		ErrorCount: 0,
		Reachable:  true,
	}

	monitor.updateOverallStatus()

	// Give callback time to execute
	time.Sleep(10 * time.Millisecond)

	assert.Equal(t, StatusOrange, monitor.GetOverallStatus())
	assert.True(t, statusChanged)
	assert.Equal(t, StatusOrange, receivedStatus)
}

func TestUpdateOverallStatus_Red(t *testing.T) {
	cfg := &config.Manager{}

	statusChanged := false
	var receivedStatus OverallStatus

	callback := func(status OverallStatus, details map[string]*gatus.EndpointStatus) {
		statusChanged = true
		receivedStatus = status
	}

	monitor := New(cfg, callback)

	// Add endpoints with 3+ errors
	monitor.statuses["url1"] = &gatus.EndpointStatus{
		URL:        "url1",
		ErrorCount: 3,
		Reachable:  true,
	}

	monitor.updateOverallStatus()

	// Give callback time to execute
	time.Sleep(10 * time.Millisecond)

	assert.Equal(t, StatusRed, monitor.GetOverallStatus())
	assert.True(t, statusChanged)
	assert.Equal(t, StatusRed, receivedStatus)
}

func TestUpdateOverallStatus_Unreachable(t *testing.T) {
	cfg := &config.Manager{}

	statusChanged := false
	var receivedStatus OverallStatus

	callback := func(status OverallStatus, details map[string]*gatus.EndpointStatus) {
		statusChanged = true
		receivedStatus = status
	}

	monitor := New(cfg, callback)

	// Add unreachable endpoint
	monitor.statuses["url1"] = &gatus.EndpointStatus{
		URL:        "url1",
		ErrorCount: 0,
		Reachable:  false,
	}

	monitor.updateOverallStatus()

	// Give callback time to execute
	time.Sleep(10 * time.Millisecond)

	assert.Equal(t, StatusRed, monitor.GetOverallStatus())
	assert.True(t, statusChanged)
	assert.Equal(t, StatusRed, receivedStatus)
}

func TestGetStatusSummary(t *testing.T) {
	tests := []struct {
		name     string
		statuses map[string]*gatus.EndpointStatus
		contains []string
	}{
		{
			name:     "no endpoints",
			statuses: map[string]*gatus.EndpointStatus{},
			contains: []string{"No endpoints configured"},
		},
		{
			name: "all green",
			statuses: map[string]*gatus.EndpointStatus{
				"url1": {URL: "url1", ErrorCount: 0, Reachable: true},
				"url2": {URL: "url2", ErrorCount: 0, Reachable: true},
			},
			contains: []string{"green", "Endpoints: 2"},
		},
		{
			name: "with errors",
			statuses: map[string]*gatus.EndpointStatus{
				"url1": {URL: "url1", ErrorCount: 2, Reachable: true},
			},
			contains: []string{"Errors: 2"},
		},
		{
			name: "with unreachable",
			statuses: map[string]*gatus.EndpointStatus{
				"url1": {URL: "url1", ErrorCount: 0, Reachable: false},
			},
			contains: []string{"Unreachable: 1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Manager{}
			monitor := New(cfg, nil)
			monitor.statuses = tt.statuses

			// Update overall status to reflect the test data
			monitor.updateOverallStatus()

			summary := monitor.GetStatusSummary()

			for _, substr := range tt.contains {
				assert.Contains(t, summary, substr)
			}
		})
	}
}

func TestUpdateOverallStatus_NoCallback(t *testing.T) {
	cfg := &config.Manager{}
	monitor := New(cfg, nil) // No callback

	monitor.statuses["url1"] = &gatus.EndpointStatus{
		URL:        "url1",
		ErrorCount: 3,
		Reachable:  true,
	}

	// Should not panic
	monitor.updateOverallStatus()

	assert.Equal(t, StatusRed, monitor.GetOverallStatus())
}

func TestStatusChangePriority(t *testing.T) {
	tests := []struct {
		name           string
		statuses       map[string]*gatus.EndpointStatus
		expectedStatus OverallStatus
	}{
		{
			name: "red overrides orange and green",
			statuses: map[string]*gatus.EndpointStatus{
				"url1": {URL: "url1", ErrorCount: 0, Reachable: true},
				"url2": {URL: "url2", ErrorCount: 1, Reachable: true},
				"url3": {URL: "url3", ErrorCount: 3, Reachable: true},
			},
			expectedStatus: StatusRed,
		},
		{
			name: "orange overrides green",
			statuses: map[string]*gatus.EndpointStatus{
				"url1": {URL: "url1", ErrorCount: 0, Reachable: true},
				"url2": {URL: "url2", ErrorCount: 2, Reachable: true},
			},
			expectedStatus: StatusOrange,
		},
		{
			name: "all green",
			statuses: map[string]*gatus.EndpointStatus{
				"url1": {URL: "url1", ErrorCount: 0, Reachable: true},
				"url2": {URL: "url2", ErrorCount: 0, Reachable: true},
			},
			expectedStatus: StatusGreen,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Manager{}
			monitor := New(cfg, nil)
			monitor.statuses = tt.statuses

			monitor.updateOverallStatus()

			assert.Equal(t, tt.expectedStatus, monitor.GetOverallStatus())
		})
	}
}
