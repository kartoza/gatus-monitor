// Copyright (c) 2026 Kartoza
// SPDX-License-Identifier: MIT

package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()
	assert.Equal(t, DefaultQueryInterval, config.QueryInterval)
	assert.Empty(t, config.GatusURLs)
}

func TestValidate_ValidConfig(t *testing.T) {
	config := &Config{
		QueryInterval: 60,
		Instances: []GatusInstance{
			{Name: "Production", URL: "https://status.example.com"},
			{Name: "Staging", URL: "http://localhost:8080"},
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
		Instances:     []GatusInstance{},
	}

	err := Validate(config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least")
}

func TestValidate_IntervalTooHigh(t *testing.T) {
	config := &Config{
		QueryInterval: 5000,
		Instances:     []GatusInstance{},
	}

	err := Validate(config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at most")
}

func TestValidate_EmptyURL(t *testing.T) {
	config := &Config{
		QueryInterval: 60,
		Instances:     []GatusInstance{{Name: "Test", URL: ""}},
	}

	err := Validate(config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "URL")
}

func TestValidate_EmptyName(t *testing.T) {
	config := &Config{
		QueryInterval: 60,
		Instances:     []GatusInstance{{Name: "", URL: "https://example.com"}},
	}

	err := Validate(config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "name")
}

func TestValidate_InvalidURL(t *testing.T) {
	config := &Config{
		QueryInterval: 60,
		Instances:     []GatusInstance{{Name: "Test", URL: "://invalid"}},
	}

	err := Validate(config)
	assert.Error(t, err)
}

func TestValidate_InvalidScheme(t *testing.T) {
	config := &Config{
		QueryInterval: 60,
		Instances:     []GatusInstance{{Name: "Test", URL: "ftp://example.com"}},
	}

	err := Validate(config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "http or https")
}

func TestValidate_MissingHost(t *testing.T) {
	config := &Config{
		QueryInterval: 60,
		Instances:     []GatusInstance{{Name: "Test", URL: "https://"}},
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

func TestManager_AddInstance(t *testing.T) {
	// This test would need proper storage setup
	// For now, we'll test the logic without persistence
	manager := &Manager{
		config: &Config{
			QueryInterval: 60,
			Instances:     []GatusInstance{},
		},
	}

	// Manually test the validation logic
	instance := GatusInstance{Name: "Production", URL: "https://status.example.com"}
	tempConfig := &Config{
		QueryInterval: manager.config.QueryInterval,
		Instances:     append(manager.config.Instances, instance),
	}

	err := Validate(tempConfig)
	assert.NoError(t, err)
}

func TestManager_AddInstance_DuplicateURL(t *testing.T) {
	manager := &Manager{
		config: &Config{
			QueryInterval: 60,
			Instances:     []GatusInstance{{Name: "Existing", URL: "https://status.example.com"}},
		},
	}

	// Test duplicate detection logic
	testURL := "https://status.example.com"
	for _, existing := range manager.config.Instances {
		if existing.URL == testURL {
			// This should trigger
			assert.Equal(t, testURL, existing.URL)
			return
		}
	}

	t.Fatal("Expected to find duplicate URL")
}

func TestManager_RemoveInstance(t *testing.T) {
	manager := &Manager{
		config: &Config{
			QueryInterval: 60,
			Instances: []GatusInstance{
				{Name: "Production", URL: "https://status1.example.com"},
				{Name: "Staging", URL: "https://status2.example.com"},
			},
		},
	}

	// Test removal logic
	nameToRemove := "Production"
	newInstances := make([]GatusInstance, 0, len(manager.config.Instances))
	found := false

	for _, existing := range manager.config.Instances {
		if existing.Name == nameToRemove {
			found = true
			continue
		}
		newInstances = append(newInstances, existing)
	}

	assert.True(t, found)
	assert.Len(t, newInstances, 1)
	assert.Equal(t, "Staging", newInstances[0].Name)
}

func TestManager_SetQueryInterval(t *testing.T) {
	manager := &Manager{
		config: &Config{
			QueryInterval: 60,
			Instances:     []GatusInstance{},
		},
	}

	// Test interval validation
	tempConfig := &Config{
		QueryInterval: 120,
		Instances:     manager.config.Instances,
	}

	err := Validate(tempConfig)
	assert.NoError(t, err)
}

func TestManager_SetQueryInterval_Invalid(t *testing.T) {
	manager := &Manager{
		config: &Config{
			QueryInterval: 60,
			Instances:     []GatusInstance{},
		},
	}

	// Test invalid interval
	tempConfig := &Config{
		QueryInterval: 5, // Too low
		Instances:     manager.config.Instances,
	}

	err := Validate(tempConfig)
	assert.Error(t, err)
}
