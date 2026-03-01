// Copyright (c) 2026 Kartoza
// SPDX-License-Identifier: MIT

package gatus

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	client := NewClient("https://status.example.com")
	assert.NotNil(t, client)
	assert.Equal(t, "https://status.example.com", client.baseURL)
	assert.NotNil(t, client.httpClient)
}

func TestNewClient_TrimSlash(t *testing.T) {
	client := NewClient("https://status.example.com/")
	assert.Equal(t, "https://status.example.com", client.baseURL)
}

func TestGetStatus_Success(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, APIPath, r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Accept"))

		// API returns array of endpoints directly
		endpoints := []Endpoint{
			{
				Name:  "service-1",
				Group: "production",
				Results: []Result{
					{
						Success:   true,
						Timestamp: time.Now(),
						Errors:    []string{},
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(endpoints)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	status, err := client.GetStatus(context.Background())

	require.NoError(t, err)
	assert.Equal(t, server.URL, status.URL)
	assert.Equal(t, 0, status.ErrorCount)
	assert.True(t, status.Reachable)
	assert.Nil(t, status.LastError)
	assert.False(t, status.LastChecked.IsZero())
}

func TestGetStatus_WithErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// API returns array of endpoints directly
		endpoints := []Endpoint{
			{
				Name:  "service-1",
				Group: "production",
				Results: []Result{
					{
						Success:   false,
						Timestamp: time.Now(),
						Errors:    []string{"connection refused"},
					},
				},
			},
			{
				Name:  "service-2",
				Group: "production",
				Results: []Result{
					{
						Success:   false,
						Timestamp: time.Now(),
						Errors:    []string{"timeout"},
					},
				},
			},
			{
				Name:  "service-3",
				Group: "production",
				Results: []Result{
					{
						Success:   true,
						Timestamp: time.Now(),
						Errors:    []string{},
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(endpoints)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	status, err := client.GetStatus(context.Background())

	require.NoError(t, err)
	assert.Equal(t, 2, status.ErrorCount)
	assert.True(t, status.Reachable)
}

func TestGetStatus_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	status, err := client.GetStatus(context.Background())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected status code")
	assert.False(t, status.Reachable)
	assert.NotNil(t, status.LastError)
}

func TestGetStatus_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	client := NewClient(server.URL)
	status, err := client.GetStatus(context.Background())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode")
	assert.False(t, status.Reachable)
}

func TestGetStatus_NetworkError(t *testing.T) {
	client := NewClient("http://localhost:0")
	status, err := client.GetStatus(context.Background())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to make request")
	assert.False(t, status.Reachable)
}

func TestGetStatus_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	status, err := client.GetStatus(ctx)

	require.Error(t, err)
	assert.False(t, status.Reachable)
}

func TestCountErrors_NilResponse(t *testing.T) {
	count := countErrors(nil)
	assert.Equal(t, 0, count)
}

func TestCountErrors_EmptyEndpoints(t *testing.T) {
	endpoints := []Endpoint{}
	count := countErrors(endpoints)
	assert.Equal(t, 0, count)
}

func TestCountErrors_NoResults(t *testing.T) {
	endpoints := []Endpoint{
		{
			Name:    "service-1",
			Group:   "production",
			Results: []Result{},
		},
	}
	count := countErrors(endpoints)
	assert.Equal(t, 0, count)
}

func TestCountErrors_MixedResults(t *testing.T) {
	endpoints := []Endpoint{
		{
			Name:  "service-1",
			Group: "production",
			Results: []Result{
				{Success: false, Errors: []string{"error"}},
			},
		},
		{
			Name:  "service-2",
			Group: "production",
			Results: []Result{
				{Success: true, Errors: []string{}},
			},
		},
		{
			Name:  "service-3",
			Group: "production",
			Results: []Result{
				{Success: true, Errors: []string{"warning"}}, // Has errors despite success
			},
		},
	}
	count := countErrors(endpoints)
	// Only counts failures based on Success flag, not presence of errors
	assert.Equal(t, 1, count) // Only service-1 (not success)
}
