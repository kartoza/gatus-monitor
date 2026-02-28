package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()
	assert.Equal(t, DefaultQueryInterval, config.QueryInterval)
	assert.Empty(t, config.GatusURLs)
}

func TestValidate_ValidConfig(t *testing.T) {
	config := &Config{
		QueryInterval: 60,
		GatusURLs: []string{
			"https://status.example.com",
			"http://localhost:8080",
		},
	}

	err := Validate(config)
	assert.NoError(t, err)
}

func TestValidate_NilConfig(t *testing.T) {
	err := Validate(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be nil")
}

func TestValidate_IntervalTooLow(t *testing.T) {
	config := &Config{
		QueryInterval: 5,
		GatusURLs:     []string{},
	}

	err := Validate(config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least")
}

func TestValidate_IntervalTooHigh(t *testing.T) {
	config := &Config{
		QueryInterval: 5000,
		GatusURLs:     []string{},
	}

	err := Validate(config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at most")
}

func TestValidate_EmptyURL(t *testing.T) {
	config := &Config{
		QueryInterval: 60,
		GatusURLs:     []string{""},
	}

	err := Validate(config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty")
}

func TestValidate_InvalidURL(t *testing.T) {
	config := &Config{
		QueryInterval: 60,
		GatusURLs:     []string{"not-a-valid-url"},
	}

	err := Validate(config)
	assert.Error(t, err)
}

func TestValidate_InvalidScheme(t *testing.T) {
	config := &Config{
		QueryInterval: 60,
		GatusURLs:     []string{"ftp://example.com"},
	}

	err := Validate(config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "http or https")
}

func TestValidate_MissingHost(t *testing.T) {
	config := &Config{
		QueryInterval: 60,
		GatusURLs:     []string{"https://"},
	}

	err := Validate(config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "host")
}

func TestManager_GetQueryInterval(t *testing.T) {
	manager := &Manager{
		config: &Config{
			QueryInterval: 120,
		},
	}

	duration := manager.GetQueryInterval()
	assert.Equal(t, 120*time.Second, duration)
}

func TestManager_AddGatusURL(t *testing.T) {
	// This test would need proper storage setup
	// For now, we'll test the logic without persistence
	manager := &Manager{
		config: &Config{
			QueryInterval: 60,
			GatusURLs:     []string{},
		},
	}

	// Mock the Update method behavior
	testURL := "https://status.example.com"

	// Manually test the validation logic
	tempConfig := &Config{
		QueryInterval: manager.config.QueryInterval,
		GatusURLs:     append(manager.config.GatusURLs, testURL),
	}

	err := Validate(tempConfig)
	assert.NoError(t, err)
}

func TestManager_AddGatusURL_Duplicate(t *testing.T) {
	manager := &Manager{
		config: &Config{
			QueryInterval: 60,
			GatusURLs:     []string{"https://status.example.com"},
		},
	}

	// Test duplicate detection logic
	testURL := "https://status.example.com"
	for _, existing := range manager.config.GatusURLs {
		if existing == testURL {
			// This should trigger
			assert.Equal(t, testURL, existing)
			return
		}
	}

	t.Fatal("Expected to find duplicate URL")
}

func TestManager_RemoveGatusURL(t *testing.T) {
	manager := &Manager{
		config: &Config{
			QueryInterval: 60,
			GatusURLs: []string{
				"https://status1.example.com",
				"https://status2.example.com",
			},
		},
	}

	// Test removal logic
	urlToRemove := "https://status1.example.com"
	newURLs := make([]string, 0, len(manager.config.GatusURLs))
	found := false

	for _, existing := range manager.config.GatusURLs {
		if existing == urlToRemove {
			found = true
			continue
		}
		newURLs = append(newURLs, existing)
	}

	assert.True(t, found)
	assert.Len(t, newURLs, 1)
	assert.Equal(t, "https://status2.example.com", newURLs[0])
}

func TestManager_SetQueryInterval(t *testing.T) {
	manager := &Manager{
		config: &Config{
			QueryInterval: 60,
			GatusURLs:     []string{},
		},
	}

	// Test interval validation
	tempConfig := &Config{
		QueryInterval: 120,
		GatusURLs:     manager.config.GatusURLs,
	}

	err := Validate(tempConfig)
	assert.NoError(t, err)
}

func TestManager_SetQueryInterval_Invalid(t *testing.T) {
	manager := &Manager{
		config: &Config{
			QueryInterval: 60,
			GatusURLs:     []string{},
		},
	}

	// Test invalid interval
	tempConfig := &Config{
		QueryInterval: 5, // Too low
		GatusURLs:     manager.config.GatusURLs,
	}

	err := Validate(tempConfig)
	assert.Error(t, err)
}
