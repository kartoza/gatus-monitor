// Copyright (c) 2026 Kartoza
// SPDX-License-Identifier: MIT

// Package config manages application configuration
package config

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/kartoza/gatus-monitor/internal/storage"
)

const (
	// DefaultQueryInterval is the default time between queries in seconds
	DefaultQueryInterval = 60

	// MinQueryInterval is the minimum allowed query interval
	MinQueryInterval = 10

	// MaxQueryInterval is the maximum allowed query interval
	MaxQueryInterval = 3600
)

// GatusInstance represents a single Gatus monitoring instance
type GatusInstance struct {
	Name     string `json:"name"`      // Friendly name for this instance
	URL      string `json:"url"`       // URL to the Gatus landing page
	IconURL  string `json:"icon_url"`  // URL to the icon (favicon from monitored site)
	IconData []byte `json:"icon_data"` // Cached icon data
}

// Config represents the application configuration
type Config struct {
	QueryInterval int             `json:"query_interval"` // Query interval in seconds
	GatusURLs     []string        `json:"gatus_urls"`     // Legacy: List of Gatus base URLs (deprecated)
	Instances     []GatusInstance `json:"instances"`      // Gatus instances with metadata
}

// Manager handles configuration persistence and validation
type Manager struct {
	storage *storage.Storage
	config  *Config
}

// NewManager creates a new configuration manager
func NewManager() (*Manager, error) {
	store, err := storage.New()
	if err != nil {
		return nil, fmt.Errorf("failed to create storage: %w", err)
	}

	manager := &Manager{
		storage: store,
		config:  DefaultConfig(),
	}

	// Load existing config if available
	if err := manager.Load(); err != nil {
		// If load fails, use default config
		// This is not a fatal error
		manager.config = DefaultConfig()
	}

	return manager, nil
}

// DefaultConfig returns a configuration with default values
func DefaultConfig() *Config {
	return &Config{
		QueryInterval: DefaultQueryInterval,
		GatusURLs:     []string{},
		Instances:     []GatusInstance{},
	}
}

// GetAPIEndpoint returns the API endpoint URL for a Gatus instance
func (gi *GatusInstance) GetAPIEndpoint() string {
	return gi.URL + "/api/v1/endpoints/statuses"
}

// Get returns the current configuration
func (m *Manager) Get() *Config {
	return m.config
}

// Update updates the configuration and persists it
func (m *Manager) Update(config *Config) error {
	if err := Validate(config); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	m.config = config

	if err := m.Save(); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	return nil
}

// Load loads the configuration from disk
func (m *Manager) Load() error {
	data, err := m.storage.ReadConfig()
	if err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}

	// No config file exists, use defaults
	if data == nil {
		m.config = DefaultConfig()
		return nil
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	// Migrate legacy GatusURLs to Instances
	if len(config.GatusURLs) > 0 && len(config.Instances) == 0 {
		for i, urlStr := range config.GatusURLs {
			config.Instances = append(config.Instances, GatusInstance{
				Name: fmt.Sprintf("Gatus %d", i+1),
				URL:  urlStr,
			})
		}
		// Clear legacy field after migration
		config.GatusURLs = []string{}
	}

	// Validate loaded config
	if err := Validate(&config); err != nil {
		return fmt.Errorf("invalid config file: %w", err)
	}

	m.config = &config
	return nil
}

// Save saves the configuration to disk
func (m *Manager) Save() error {
	data, err := json.MarshalIndent(m.config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := m.storage.WriteConfig(data); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// GetQueryInterval returns the query interval as a time.Duration
func (m *Manager) GetQueryInterval() time.Duration {
	return time.Duration(m.config.QueryInterval) * time.Second
}

// GetConfigPath returns the path to the configuration file
func (m *Manager) GetConfigPath() string {
	return m.storage.GetConfigPath()
}

// Validate validates a configuration
func Validate(config *Config) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	// Validate query interval
	if config.QueryInterval < MinQueryInterval {
		return fmt.Errorf("query interval must be at least %d seconds", MinQueryInterval)
	}

	if config.QueryInterval > MaxQueryInterval {
		return fmt.Errorf("query interval must be at most %d seconds", MaxQueryInterval)
	}

	// Validate instances
	for i, instance := range config.Instances {
		if instance.Name == "" {
			return fmt.Errorf("instance at index %d must have a name", i)
		}

		if instance.URL == "" {
			return fmt.Errorf("instance at index %d must have a URL", i)
		}

		parsedURL, err := url.Parse(instance.URL)
		if err != nil {
			return fmt.Errorf("instance %q URL is invalid: %w", instance.Name, err)
		}

		// Ensure scheme is http or https
		if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
			return fmt.Errorf("instance %q URL must use http or https scheme", instance.Name)
		}

		// Ensure host is present
		if parsedURL.Host == "" {
			return fmt.Errorf("instance %q URL must have a host", instance.Name)
		}
	}

	return nil
}

// AddInstance adds a Gatus instance to the configuration
func (m *Manager) AddInstance(instance GatusInstance) error {
	// Check if instance with same name or URL already exists
	for _, existing := range m.config.Instances {
		if existing.Name == instance.Name {
			return fmt.Errorf("instance with name %q already exists", instance.Name)
		}
		if existing.URL == instance.URL {
			return fmt.Errorf("instance with URL %q already exists", instance.URL)
		}
	}

	// Validate by creating a temporary config
	tempConfig := &Config{
		QueryInterval: m.config.QueryInterval,
		Instances:     append(m.config.Instances, instance),
	}

	if err := Validate(tempConfig); err != nil {
		return err
	}

	// Update and save
	return m.Update(tempConfig)
}

// RemoveInstance removes a Gatus instance from the configuration
func (m *Manager) RemoveInstance(name string) error {
	newInstances := make([]GatusInstance, 0, len(m.config.Instances))
	found := false

	for _, existing := range m.config.Instances {
		if existing.Name == name {
			found = true
			continue
		}
		newInstances = append(newInstances, existing)
	}

	if !found {
		return fmt.Errorf("instance %q not found", name)
	}

	tempConfig := &Config{
		QueryInterval: m.config.QueryInterval,
		Instances:     newInstances,
	}

	return m.Update(tempConfig)
}

// UpdateInstance updates an existing instance
func (m *Manager) UpdateInstance(oldName string, newInstance GatusInstance) error {
	newInstances := make([]GatusInstance, 0, len(m.config.Instances))
	found := false

	for _, existing := range m.config.Instances {
		if existing.Name == oldName {
			newInstances = append(newInstances, newInstance)
			found = true
		} else {
			newInstances = append(newInstances, existing)
		}
	}

	if !found {
		return fmt.Errorf("instance %q not found", oldName)
	}

	tempConfig := &Config{
		QueryInterval: m.config.QueryInterval,
		Instances:     newInstances,
	}

	return m.Update(tempConfig)
}

// SetQueryInterval sets the query interval
func (m *Manager) SetQueryInterval(seconds int) error {
	tempConfig := &Config{
		QueryInterval: seconds,
		Instances:     m.config.Instances,
	}

	return m.Update(tempConfig)
}

// NewTestManager creates a Manager with the given config for testing purposes.
// This should only be used in tests.
func NewTestManager(cfg *Config) *Manager {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	return &Manager{
		storage: nil,
		config:  cfg,
	}
}
