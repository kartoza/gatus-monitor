// Copyright (c) 2026 Kartoza
// SPDX-License-Identifier: MIT

// Package gatus provides a client for the Gatus API
package gatus

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const (
	// APIPath is the path to the Gatus API endpoint
	APIPath = "/api/v1/endpoints/statuses"

	// RequestTimeout is the maximum time to wait for a response
	RequestTimeout = 30 * time.Second
)

// Client is a Gatus API client
type Client struct {
	httpClient *http.Client
	baseURL    string
}

// EndpointStatus represents the status of a Gatus endpoint
type EndpointStatus struct {
	URL         string
	ErrorCount  int
	LastChecked time.Time
	LastError   error
	Reachable   bool
}

// NewClient creates a new Gatus API client
func NewClient(baseURL string) *Client {
	// Ensure base URL doesn't end with /
	baseURL = strings.TrimRight(baseURL, "/")

	return &Client{
		httpClient: &http.Client{
			Timeout: RequestTimeout,
		},
		baseURL: baseURL,
	}
}

// GetStatus queries the Gatus API and returns the endpoint status
func (c *Client) GetStatus(ctx context.Context) (*EndpointStatus, error) {
	status := &EndpointStatus{
		URL:         c.baseURL,
		LastChecked: time.Now(),
		Reachable:   false,
	}

	url := c.baseURL + APIPath

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		status.LastError = fmt.Errorf("failed to create request: %w", err)
		return status, status.LastError
	}

	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		status.LastError = fmt.Errorf("failed to make request: %w", err)
		return status, status.LastError
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		status.LastError = fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		return status, status.LastError
	}

	// The API returns an array of endpoints directly, not wrapped in an object
	var endpoints []Endpoint
	if err := json.NewDecoder(resp.Body).Decode(&endpoints); err != nil {
		status.LastError = fmt.Errorf("failed to decode response: %w", err)
		return status, status.LastError
	}

	// Count errors in the response
	errorCount := countErrors(endpoints)
	status.ErrorCount = errorCount
	status.Reachable = true
	status.LastError = nil

	return status, nil
}

// countErrors counts the number of failed endpoints in the API response
func countErrors(endpoints []Endpoint) int {
	if endpoints == nil {
		return 0
	}

	errorCount := 0
	for _, endpoint := range endpoints {
		// Check if endpoint has any results
		if len(endpoint.Results) == 0 {
			continue
		}

		// Check the most recent result (results are ordered most recent first)
		latestResult := endpoint.Results[0]

		// Count as error if not successful
		if !latestResult.Success {
			errorCount++
		}
	}

	return errorCount
}

// Endpoint represents a monitored endpoint in Gatus
type Endpoint struct {
	Name    string   `json:"name"`
	Group   string   `json:"group"`
	Results []Result `json:"results"`
}

// Result represents a health check result
type Result struct {
	Success   bool      `json:"success"`
	Timestamp time.Time `json:"timestamp"`
	Errors    []string  `json:"errors"`
}
