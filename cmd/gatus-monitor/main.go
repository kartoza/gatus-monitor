// Gatus Monitor - Cross-platform system tray app for monitoring Gatus endpoints
package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"fyne.io/fyne/v2/app"

	"github.com/kartoza/gatus-monitor/internal/config"
	"github.com/kartoza/gatus-monitor/internal/gatus"
	"github.com/kartoza/gatus-monitor/internal/monitor"
	"github.com/kartoza/gatus-monitor/internal/systray"
	"github.com/kartoza/gatus-monitor/internal/ui"
)

const (
	appName    = "Gatus Monitor"
	appVersion = "0.1.0"
)

type application struct {
	config         *config.Manager
	monitor        *monitor.Monitor
	tray           *systray.TrayApp
	settingsWindow *ui.SettingsWindow
}

func main() {
	// Print version
	fmt.Printf("%s v%s\n", appName, appVersion)

	// Initialize application
	app, err := newApplication()
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	// Handle graceful shutdown
	go app.handleShutdown()

	// Start monitoring
	if err := app.monitor.Start(); err != nil {
		log.Fatalf("Failed to start monitoring: %v", err)
	}

	// Run system tray (blocking)
	app.tray.Run()
}

// newApplication creates and initializes the application
func newApplication() (*application, error) {
	// Load configuration
	cfg, err := config.NewManager()
	if err != nil {
		return nil, fmt.Errorf("failed to create config manager: %w", err)
	}

	// Create Fyne app for settings window
	fyneApp := app.NewWithID("com.kartoza.gatus-monitor")

	app := &application{
		config: cfg,
	}

	// Create monitor with status callback
	app.monitor = monitor.New(cfg, app.onStatusChange)

	// Create system tray
	app.tray = systray.New(app.monitor, app.showSettings, app.quit)

	// Create settings window with shared Fyne app
	app.settingsWindow = ui.NewSettingsWindowWithApp(fyneApp, cfg, app.onConfigChanged)

	return app, nil
}

// onStatusChange is called when the overall monitoring status changes
func (app *application) onStatusChange(status monitor.OverallStatus, details map[string]*gatus.EndpointStatus) {
	log.Printf("Status changed to: %s", status.String())

	// Update system tray
	app.tray.UpdateStatus(status, details)
}

// showSettings displays the settings window
func (app *application) showSettings() {
	// Reload current config before showing
	app.settingsWindow.Reload()
	app.settingsWindow.Show()
}

// onConfigChanged is called when the configuration is updated
func (app *application) onConfigChanged() {
	log.Println("Configuration changed, restarting monitor...")

	// Restart monitoring with new configuration
	if err := app.monitor.Restart(); err != nil {
		log.Printf("Failed to restart monitor: %v", err)
	}
}

// quit handles application shutdown
func (app *application) quit() {
	log.Println("Shutting down...")

	// Stop monitoring
	app.monitor.Stop()

	os.Exit(0)
}

// handleShutdown handles graceful shutdown on signals
func (app *application) handleShutdown() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan

	log.Println("Received shutdown signal")
	app.quit()
}
