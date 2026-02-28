// Package storage provides persistent file storage with platform-specific paths
package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// Storage handles persistent file operations
type Storage struct {
	configDir  string
	configFile string
}

// New creates a new Storage instance with platform-specific paths
func New() (*Storage, error) {
	configDir, err := getConfigDir()
	if err != nil {
		return nil, fmt.Errorf("failed to determine config directory: %w", err)
	}

	configFile := filepath.Join(configDir, "config.json")

	return &Storage{
		configDir:  configDir,
		configFile: configFile,
	}, nil
}

// GetConfigPath returns the full path to the configuration file
func (s *Storage) GetConfigPath() string {
	return s.configFile
}

// EnsureConfigDir ensures the configuration directory exists with proper permissions
func (s *Storage) EnsureConfigDir() error {
	if err := os.MkdirAll(s.configDir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	return nil
}

// ReadConfig reads the configuration file and returns its contents
func (s *Storage) ReadConfig() ([]byte, error) {
	data, err := os.ReadFile(s.configFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // File doesn't exist yet, not an error
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	return data, nil
}

// WriteConfig writes data to the configuration file with secure permissions
func (s *Storage) WriteConfig(data []byte) error {
	if err := s.EnsureConfigDir(); err != nil {
		return err
	}

	// Write to temporary file first
	tmpFile := s.configFile + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tmpFile, s.configFile); err != nil {
		os.Remove(tmpFile) // Clean up temp file
		return fmt.Errorf("failed to rename config file: %w", err)
	}

	return nil
}

// ConfigExists returns true if the configuration file exists
func (s *Storage) ConfigExists() bool {
	_, err := os.Stat(s.configFile)
	return err == nil
}

// getConfigDir returns the platform-specific configuration directory
func getConfigDir() (string, error) {
	var baseDir string

	switch runtime.GOOS {
	case "windows":
		// Windows: %APPDATA%\gatus-monitor
		baseDir = os.Getenv("APPDATA")
		if baseDir == "" {
			return "", fmt.Errorf("APPDATA environment variable not set")
		}
		return filepath.Join(baseDir, "gatus-monitor"), nil

	case "darwin":
		// macOS: ~/Library/Application Support/gatus-monitor
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %w", err)
		}
		return filepath.Join(homeDir, "Library", "Application Support", "gatus-monitor"), nil

	default:
		// Linux and others: ~/.config/gatus-monitor (XDG Base Directory)
		configHome := os.Getenv("XDG_CONFIG_HOME")
		if configHome == "" {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return "", fmt.Errorf("failed to get home directory: %w", err)
			}
			configHome = filepath.Join(homeDir, ".config")
		}
		return filepath.Join(configHome, "gatus-monitor"), nil
	}
}
