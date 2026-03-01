// Copyright (c) 2026 Kartoza
// SPDX-License-Identifier: MIT

// Package icons provides icon fetching and caching functionality
package icons

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/html"
)

const (
	// MaxIconSize is the maximum size of an icon we'll download (1MB)
	MaxIconSize = 1024 * 1024
	// FetchTimeout is how long to wait for icon downloads
	FetchTimeout = 10 * time.Second
)

// FetchIcon attempts to fetch an icon for a given URL
// It tries multiple strategies:
// 1. If iconURL is provided, use that directly
// 2. Try to find favicon link in the HTML
// 3. Try common favicon locations
func FetchIcon(baseURL string, iconURL string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), FetchTimeout)
	defer cancel()

	// If explicit icon URL provided, try that first
	if iconURL != "" {
		data, err := downloadIcon(ctx, iconURL)
		if err == nil {
			return data, nil
		}
		// If explicit URL fails, fall through to auto-detection
	}

	// Parse base URL
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}

	// Try to find favicon in HTML
	if data, err := findFaviconInHTML(ctx, baseURL); err == nil {
		return data, nil
	}

	// Try common favicon locations
	commonPaths := []string{
		"/favicon.ico",
		"/apple-touch-icon.png",
		"/favicon.png",
	}

	baseURLStr := fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host)
	for _, path := range commonPaths {
		iconURLStr := baseURLStr + path
		if data, err := downloadIcon(ctx, iconURLStr); err == nil {
			return data, nil
		}
	}

	return nil, fmt.Errorf("no icon found for %s", baseURL)
}

// findFaviconInHTML fetches the HTML page and looks for favicon links
func findFaviconInHTML(ctx context.Context, pageURL string) ([]byte, error) {
	// Fetch the HTML page
	req, err := http.NewRequestWithContext(ctx, "GET", pageURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	// Parse HTML
	doc, err := html.Parse(resp.Body)
	if err != nil {
		return nil, err
	}

	// Find favicon link
	faviconURL := findFaviconLink(doc)
	if faviconURL == "" {
		return nil, fmt.Errorf("no favicon link found")
	}

	// Resolve relative URLs
	parsedPage, _ := url.Parse(pageURL)
	parsedIcon, err := url.Parse(faviconURL)
	if err != nil {
		return nil, err
	}

	absoluteIconURL := parsedPage.ResolveReference(parsedIcon).String()

	// Download the icon
	return downloadIcon(ctx, absoluteIconURL)
}

// findFaviconLink recursively searches HTML for favicon link
func findFaviconLink(n *html.Node) string {
	if n.Type == html.ElementNode && n.Data == "link" {
		var rel, href string
		for _, attr := range n.Attr {
			if attr.Key == "rel" {
				rel = attr.Val
			}
			if attr.Key == "href" {
				href = attr.Val
			}
		}

		// Check for favicon rel types
		relLower := strings.ToLower(rel)
		if strings.Contains(relLower, "icon") {
			return href
		}
	}

	// Recursively search children
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if result := findFaviconLink(c); result != "" {
			return result
		}
	}

	return ""
}

// downloadIcon downloads an icon from a URL
func downloadIcon(ctx context.Context, iconURL string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", iconURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	// Limit read size
	limitedReader := io.LimitReader(resp.Body, MaxIconSize)
	data, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("icon is empty")
	}

	return data, nil
}
