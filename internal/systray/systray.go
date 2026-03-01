// Copyright (c) 2026 Kartoza
// SPDX-License-Identifier: MIT

// Package systray provides system tray integration
package systray

import (
	"fmt"
	"os/exec"
	"runtime"
	"sync"

	"fyne.io/systray"

	"github.com/kartoza/gatus-monitor/internal/config"
	"github.com/kartoza/gatus-monitor/internal/gatus"
	"github.com/kartoza/gatus-monitor/internal/icons"
	"github.com/kartoza/gatus-monitor/internal/monitor"
)

// instanceMenuItem holds menu item and associated data for an instance
type instanceMenuItem struct {
	menuItem *systray.MenuItem
	url      string
	name     string
}

// TrayApp represents the system tray application
type TrayApp struct {
	monitor          *monitor.Monitor
	config           *config.Manager
	currentStatus    monitor.OverallStatus
	onSettings       func()
	onQuit           func()
	statusMenuItem   *systray.MenuItem
	instanceItems    []*instanceMenuItem
	instanceStatuses map[string]*gatus.EndpointStatus
	mu               sync.RWMutex
}

// New creates a new system tray application
func New(mon *monitor.Monitor, cfg *config.Manager, onSettings, onQuit func()) *TrayApp {
	return &TrayApp{
		monitor:          mon,
		config:           cfg,
		currentStatus:    monitor.StatusGreen,
		onSettings:       onSettings,
		onQuit:           onQuit,
		instanceItems:    make([]*instanceMenuItem, 0),
		instanceStatuses: make(map[string]*gatus.EndpointStatus),
	}
}

// Run starts the system tray application
func (app *TrayApp) Run() {
	systray.Run(app.onReady, app.onExit)
}

// Quit exits the system tray application
func (app *TrayApp) Quit() {
	systray.Quit()
}

// onReady is called when the system tray is ready
func (app *TrayApp) onReady() {
	// Set initial icon
	app.updateIcon(monitor.StatusGreen)

	// Set tooltip
	systray.SetTooltip("Gatus Monitor - All systems operational")

	// Create menu items
	app.statusMenuItem = systray.AddMenuItem("Status: Green", "Current overall status")
	app.statusMenuItem.Disable()

	systray.AddSeparator()

	// Create instance menu items
	app.createInstanceMenuItems()

	systray.AddSeparator()

	mSettings := systray.AddMenuItem("Settings", "Configure Gatus endpoints")
	mRefresh := systray.AddMenuItem("Refresh Now", "Force refresh all endpoints")

	systray.AddSeparator()

	mQuit := systray.AddMenuItem("Quit", "Exit Gatus Monitor")

	// Handle menu clicks
	go func() {
		for {
			select {
			case <-mSettings.ClickedCh:
				if app.onSettings != nil {
					app.onSettings()
				}
			case <-mRefresh.ClickedCh:
				app.forceRefresh()
			case <-mQuit.ClickedCh:
				if app.onQuit != nil {
					app.onQuit()
				}
				app.Quit()
			}
		}
	}()
}

// createInstanceMenuItems creates menu items for each configured instance
func (app *TrayApp) createInstanceMenuItems() {
	if app.config == nil {
		return
	}

	instances := app.config.Get().Instances
	for _, instance := range instances {
		// Create menu item with gray dot initially (unknown status)
		title := fmt.Sprintf("⚪ %s", instance.Name)
		item := systray.AddMenuItem(title, fmt.Sprintf("Open %s in browser", instance.URL))

		instanceItem := &instanceMenuItem{
			menuItem: item,
			url:      instance.URL,
			name:     instance.Name,
		}
		app.instanceItems = append(app.instanceItems, instanceItem)

		// Handle click in goroutine
		go app.handleInstanceClick(instanceItem)
	}
}

// handleInstanceClick handles clicks on instance menu items
func (app *TrayApp) handleInstanceClick(item *instanceMenuItem) {
	for range item.menuItem.ClickedCh {
		_ = openURL(item.url)
	}
}

// openURL opens a URL in the default browser
func openURL(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default: // linux, freebsd, etc.
		cmd = exec.Command("xdg-open", url)
	}

	return cmd.Start()
}

// onExit is called when the system tray exits
func (app *TrayApp) onExit() {
	// Cleanup if needed
}

// UpdateStatus updates the tray icon and tooltip based on status
func (app *TrayApp) UpdateStatus(status monitor.OverallStatus, details map[string]*gatus.EndpointStatus) {
	app.mu.Lock()
	app.currentStatus = status
	app.instanceStatuses = details
	app.mu.Unlock()

	// Update icon
	app.updateIcon(status)

	// Update tooltip
	summary := app.generateSummary(status, details)
	systray.SetTooltip(summary)

	// Update status menu item
	if app.statusMenuItem != nil {
		statusText := fmt.Sprintf("Status: %s", capitalizeFirst(status.String()))
		app.statusMenuItem.SetTitle(statusText)
	}

	// Update instance menu items with status dots
	app.updateInstanceMenuItems(details)
}

// updateInstanceMenuItems updates the status dots for each instance
func (app *TrayApp) updateInstanceMenuItems(details map[string]*gatus.EndpointStatus) {
	app.mu.RLock()
	defer app.mu.RUnlock()

	for _, item := range app.instanceItems {
		dot := "⚪" // Gray - unknown/pending
		if status, ok := details[item.name]; ok {
			if !status.Reachable {
				dot = "🔴" // Red - unreachable
			} else if status.ErrorCount >= 3 {
				dot = "🔴" // Red - 3+ errors
			} else if status.ErrorCount > 0 {
				dot = "🟠" // Orange - 1-2 errors
			} else {
				dot = "🟢" // Green - healthy
			}
		}
		item.menuItem.SetTitle(fmt.Sprintf("%s %s", dot, item.name))
	}
}

// updateIcon updates the system tray icon
func (app *TrayApp) updateIcon(status monitor.OverallStatus) {
	icon := icons.GetIconForStatus(status.String())
	systray.SetIcon(icon)
}

// generateSummary generates a summary string for the tooltip
func (app *TrayApp) generateSummary(status monitor.OverallStatus, details map[string]*gatus.EndpointStatus) string {
	// Get the configured endpoint count from the monitor
	configuredCount := 0
	if app.monitor != nil {
		configuredCount = app.monitor.GetConfiguredEndpointCount()
	}

	if configuredCount == 0 {
		return "Gatus Monitor - No endpoints configured"
	}

	summary := fmt.Sprintf("Gatus Monitor - %s\n", capitalizeFirst(status.String()))

	totalErrors := 0
	unreachable := 0
	healthy := 0

	for _, detail := range details {
		if !detail.Reachable {
			unreachable++
		} else if detail.ErrorCount > 0 {
			totalErrors += detail.ErrorCount
		} else {
			healthy++
		}
	}

	summary += fmt.Sprintf("Total Endpoints: %d\n", configuredCount)
	summary += fmt.Sprintf("Healthy: %d", healthy)

	if unreachable > 0 {
		summary += fmt.Sprintf(" | Unreachable: %d", unreachable)
	}

	if totalErrors > 0 {
		summary += fmt.Sprintf(" | Errors: %d", totalErrors)
	}

	return summary
}

// forceRefresh forces a refresh of all endpoints
func (app *TrayApp) forceRefresh() {
	// Restart monitoring to force immediate queries
	if app.monitor != nil {
		_ = app.monitor.Restart()
	}
}

// capitalizeFirst capitalizes the first letter of a string
func capitalizeFirst(s string) string {
	if s == "" {
		return ""
	}
	return string(s[0]-32) + s[1:]
}
