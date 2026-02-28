// Package systray provides system tray integration
package systray

import (
	"fmt"
	"sync"

	"fyne.io/systray"

	"github.com/kartoza/gatus-monitor/internal/gatus"
	"github.com/kartoza/gatus-monitor/internal/icons"
	"github.com/kartoza/gatus-monitor/internal/monitor"
)

// TrayApp represents the system tray application
type TrayApp struct {
	monitor         *monitor.Monitor
	currentStatus   monitor.OverallStatus
	onSettings      func()
	onQuit          func()
	statusMenuItem  *systray.MenuItem
	mu              sync.RWMutex
}

// New creates a new system tray application
func New(mon *monitor.Monitor, onSettings, onQuit func()) *TrayApp {
	return &TrayApp{
		monitor:       mon,
		currentStatus: monitor.StatusGreen,
		onSettings:    onSettings,
		onQuit:        onQuit,
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

// onExit is called when the system tray exits
func (app *TrayApp) onExit() {
	// Cleanup if needed
}

// UpdateStatus updates the tray icon and tooltip based on status
func (app *TrayApp) UpdateStatus(status monitor.OverallStatus, details map[string]*gatus.EndpointStatus) {
	app.mu.Lock()
	app.currentStatus = status
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
}

// updateIcon updates the system tray icon
func (app *TrayApp) updateIcon(status monitor.OverallStatus) {
	icon := icons.GetIconForStatus(status.String())
	systray.SetIcon(icon)
}

// generateSummary generates a summary string for the tooltip
func (app *TrayApp) generateSummary(status monitor.OverallStatus, details map[string]*gatus.EndpointStatus) string {
	if len(details) == 0 {
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

	summary += fmt.Sprintf("Total Endpoints: %d\n", len(details))
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
		app.monitor.Restart()
	}
}

// capitalizeFirst capitalizes the first letter of a string
func capitalizeFirst(s string) string {
	if s == "" {
		return ""
	}
	return string(s[0]-32) + s[1:]
}
