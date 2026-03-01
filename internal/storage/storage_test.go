// Copyright (c) 2026 Kartoza
// SPDX-License-Identifier: MIT

package storage

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	storage, err := New()
	require.NoError(t, err)
	require.NotNil(t, storage)

	assert.NotEmpty(t, storage.configDir)
	assert.NotEmpty(t, storage.configFile)
	assert.Contains(t, storage.configFile, "gatus-monitor")
	assert.Contains(t, storage.configFile, "config.json")
}

func TestGetConfigPath(t *testing.T) {
	storage, err := New()
	require.NoError(t, err)

	path := storage.GetConfigPath()
	assert.NotEmpty(t, path)
	assert.Contains(t, path, "config.json")
}

func TestEnsureConfigDir(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()
	storage := &Storage{
		configDir:  filepath.Join(tmpDir, "test-config"),
		configFile: filepath.Join(tmpDir, "test-config", "config.json"),
	}

	err := storage.EnsureConfigDir()
	require.NoError(t, err)

	// Verify directory was created
	info, err := os.Stat(storage.configDir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())

	// Verify permissions (Unix-like systems)
	if runtime.GOOS != "windows" {
		assert.Equal(t, os.FileMode(0700), info.Mode().Perm())
	}
}

func TestWriteAndReadConfig(t *testing.T) {
	tmpDir := t.TempDir()
	storage := &Storage{
		configDir:  tmpDir,
		configFile: filepath.Join(tmpDir, "config.json"),
	}

	testData := []byte(`{"test": "data"}`)

	// Write config
	err := storage.WriteConfig(testData)
	require.NoError(t, err)

	// Read config
	data, err := storage.ReadConfig()
	require.NoError(t, err)
	assert.Equal(t, testData, data)

	// Verify file permissions (Unix-like systems)
	if runtime.GOOS != "windows" {
		info, err := os.Stat(storage.configFile)
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0600), info.Mode().Perm())
	}
}

func TestReadConfig_NonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	storage := &Storage{
		configDir:  tmpDir,
		configFile: filepath.Join(tmpDir, "nonexistent.json"),
	}

	data, err := storage.ReadConfig()
	require.NoError(t, err)
	assert.Nil(t, data)
}

func TestConfigExists(t *testing.T) {
	tmpDir := t.TempDir()
	storage := &Storage{
		configDir:  tmpDir,
		configFile: filepath.Join(tmpDir, "config.json"),
	}

	// Should not exist initially
	assert.False(t, storage.ConfigExists())

	// Create the file
	err := storage.WriteConfig([]byte("test"))
	require.NoError(t, err)

	// Should exist now
	assert.True(t, storage.ConfigExists())
}

func TestGetConfigDir_PlatformSpecific(t *testing.T) {
	configDir, err := getConfigDir()
	require.NoError(t, err)
	assert.NotEmpty(t, configDir)

	switch runtime.GOOS {
	case "windows":
		assert.Contains(t, configDir, "gatus-monitor")
	case "darwin":
		assert.Contains(t, configDir, "Library")
		assert.Contains(t, configDir, "Application Support")
		assert.Contains(t, configDir, "gatus-monitor")
	default: // Linux and others
		assert.Contains(t, configDir, "gatus-monitor")
	}
}
