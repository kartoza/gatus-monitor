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

// Config represents the application configuration
type Config struct {
	QueryInterval int      `json:"query_interval"` // Query interval in seconds
	GatusURLs     []string `json:"gatus_urls"`     // List of Gatus base URLs
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
	}
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

	// Validate URLs
	for i, urlStr := range config.GatusURLs {
		if urlStr == "" {
			return fmt.Errorf("URL at index %d is empty", i)
		}

		parsedURL, err := url.Parse(urlStr)
		if err != nil {
			return fmt.Errorf("URL at index %d is invalid: %w", i, err)
		}

		// Ensure scheme is http or https
		if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
			return fmt.Errorf("URL at index %d must use http or https scheme", i)
		}

		// Ensure host is present
		if parsedURL.Host == "" {
			return fmt.Errorf("URL at index %d must have a host", i)
		}
	}

	return nil
}

// AddGatusURL adds a Gatus URL to the configuration
func (m *Manager) AddGatusURL(urlStr string) error {
	// Check if URL already exists
	for _, existing := range m.config.GatusURLs {
		if existing == urlStr {
			return fmt.Errorf("URL already exists")
		}
	}

	// Validate URL by creating a temporary config
	tempConfig := &Config{
		QueryInterval: m.config.QueryInterval,
		GatusURLs:     append(m.config.GatusURLs, urlStr),
	}

	if err := Validate(tempConfig); err != nil {
		return err
	}

	// Update and save
	return m.Update(tempConfig)
}

// RemoveGatusURL removes a Gatus URL from the configuration
func (m *Manager) RemoveGatusURL(urlStr string) error {
	newURLs := make([]string, 0, len(m.config.GatusURLs))
	found := false

	for _, existing := range m.config.GatusURLs {
		if existing == urlStr {
			found = true
			continue
		}
		newURLs = append(newURLs, existing)
	}

	if !found {
		return fmt.Errorf("URL not found")
	}

	tempConfig := &Config{
		QueryInterval: m.config.QueryInterval,
		GatusURLs:     newURLs,
	}

	return m.Update(tempConfig)
}

// SetQueryInterval sets the query interval
func (m *Manager) SetQueryInterval(seconds int) error {
	tempConfig := &Config{
		QueryInterval: seconds,
		GatusURLs:     m.config.GatusURLs,
	}

	return m.Update(tempConfig)
}
