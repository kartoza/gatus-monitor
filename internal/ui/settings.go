// Package ui provides the settings user interface
package ui

import (
	"fmt"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/kartoza/gatus-monitor/internal/config"
)

// SettingsWindow manages the settings UI
type SettingsWindow struct {
	app             fyne.App
	window          fyne.Window
	config          *config.Manager
	onConfigChanged func()
	urlList         *widget.List
	urls            []string
	intervalEntry   *widget.Entry
}

// NewSettingsWindow creates a new settings window (creates its own Fyne app)
func NewSettingsWindow(cfg *config.Manager, onConfigChanged func()) *SettingsWindow {
	sw := &SettingsWindow{
		config:          cfg,
		onConfigChanged: onConfigChanged,
		urls:            make([]string, len(cfg.Get().GatusURLs)),
	}
	copy(sw.urls, cfg.Get().GatusURLs)

	return sw
}

// NewSettingsWindowWithApp creates a new settings window with an existing Fyne app
func NewSettingsWindowWithApp(fyneApp fyne.App, cfg *config.Manager, onConfigChanged func()) *SettingsWindow {
	sw := &SettingsWindow{
		app:             fyneApp,
		config:          cfg,
		onConfigChanged: onConfigChanged,
		urls:            make([]string, len(cfg.Get().GatusURLs)),
	}
	copy(sw.urls, cfg.Get().GatusURLs)

	return sw
}

// ensureWindow ensures the window is created (lazy initialization)
func (sw *SettingsWindow) ensureWindow() {
	if sw.window != nil {
		return
	}

	if sw.app == nil {
		sw.app = app.New()
	}

	sw.window = sw.app.NewWindow("Gatus Monitor Settings")
	sw.buildUI()
}

// Show displays the settings window
func (sw *SettingsWindow) Show() {
	// Run in goroutine to avoid blocking and handle threading
	go func() {
		sw.ensureWindow()
		if sw.window != nil {
			sw.window.Show()
			sw.window.RequestFocus()
		}
	}()
}

// Hide hides the settings window
func (sw *SettingsWindow) Hide() {
	if sw.window != nil {
		sw.window.Hide()
	}
}

// buildUI constructs the settings UI
func (sw *SettingsWindow) buildUI() {
	// Query interval section
	intervalLabel := widget.NewLabel("Query Interval (seconds):")
	sw.intervalEntry = widget.NewEntry()
	sw.intervalEntry.SetText(strconv.Itoa(sw.config.Get().QueryInterval))
	sw.intervalEntry.SetPlaceHolder("60")

	intervalHelp := widget.NewLabel("Minimum: 10, Maximum: 3600, Default: 60")
	intervalHelp.TextStyle = fyne.TextStyle{Italic: true}

	intervalContainer := container.NewVBox(
		intervalLabel,
		sw.intervalEntry,
		intervalHelp,
	)

	// URL list section
	urlLabel := widget.NewLabel("Gatus Endpoint URLs:")

	sw.urlList = widget.NewList(
		func() int {
			return len(sw.urls)
		},
		func() fyne.CanvasObject {
			return container.NewBorder(nil, nil, nil,
				widget.NewButton("Remove", func() {}),
				widget.NewLabel(""),
			)
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			border := item.(*fyne.Container)
			label := border.Objects[0].(*widget.Label)
			button := border.Objects[1].(*widget.Button)

			label.SetText(sw.urls[id])
			button.OnTapped = func() {
				sw.removeURL(id)
			}
		},
	)

	addButton := widget.NewButton("Add URL", sw.showAddURLDialog)

	urlContainer := container.NewBorder(
		container.NewVBox(urlLabel, addButton),
		nil, nil, nil,
		sw.urlList,
	)

	// Save and Cancel buttons
	saveButton := widget.NewButton("Save", sw.save)
	cancelButton := widget.NewButton("Cancel", func() {
		sw.Hide()
	})

	buttons := container.NewHBox(
		saveButton,
		cancelButton,
	)

	// Main content
	content := container.NewVBox(
		intervalContainer,
		widget.NewSeparator(),
		urlContainer,
		widget.NewSeparator(),
		buttons,
	)

	sw.window.SetContent(content)
	sw.window.Resize(fyne.NewSize(600, 400))
}

// showAddURLDialog shows a dialog to add a new URL
func (sw *SettingsWindow) showAddURLDialog() {
	urlEntry := widget.NewEntry()
	urlEntry.SetPlaceHolder("https://status.example.com")

	dialog.ShowForm("Add Gatus URL", "Add", "Cancel",
		[]*widget.FormItem{
			widget.NewFormItem("URL", urlEntry),
		},
		func(ok bool) {
			if !ok {
				return
			}

			url := urlEntry.Text
			if url == "" {
				dialog.ShowError(fmt.Errorf("URL cannot be empty"), sw.window)
				return
			}

			// Validate URL
			tempConfig := &config.Config{
				QueryInterval: sw.config.Get().QueryInterval,
				GatusURLs:     append(sw.urls, url),
			}

			if err := config.Validate(tempConfig); err != nil {
				dialog.ShowError(err, sw.window)
				return
			}

			// Add URL to list
			sw.urls = append(sw.urls, url)
			sw.urlList.Refresh()
		},
		sw.window,
	)
}

// removeURL removes a URL from the list
func (sw *SettingsWindow) removeURL(index int) {
	if index < 0 || index >= len(sw.urls) {
		return
	}

	// Remove URL
	sw.urls = append(sw.urls[:index], sw.urls[index+1:]...)
	sw.urlList.Refresh()
}

// save saves the configuration
func (sw *SettingsWindow) save() {
	// Parse interval
	intervalStr := sw.intervalEntry.Text
	interval, err := strconv.Atoi(intervalStr)
	if err != nil {
		dialog.ShowError(fmt.Errorf("invalid interval: %w", err), sw.window)
		return
	}

	// Create new config
	newConfig := &config.Config{
		QueryInterval: interval,
		GatusURLs:     sw.urls,
	}

	// Validate
	if err := config.Validate(newConfig); err != nil {
		dialog.ShowError(err, sw.window)
		return
	}

	// Update config
	if err := sw.config.Update(newConfig); err != nil {
		dialog.ShowError(fmt.Errorf("failed to save configuration: %w", err), sw.window)
		return
	}

	// Notify about config change
	if sw.onConfigChanged != nil {
		sw.onConfigChanged()
	}

	// Show success message
	dialog.ShowInformation("Success", "Configuration saved successfully", sw.window)

	// Close window
	sw.Hide()
}

// Reload reloads the configuration from the manager
func (sw *SettingsWindow) Reload() {
	cfg := sw.config.Get()
	sw.urls = make([]string, len(cfg.GatusURLs))
	copy(sw.urls, cfg.GatusURLs)

	// Only update UI elements if window has been created
	if sw.window != nil && sw.intervalEntry != nil {
		sw.intervalEntry.SetText(strconv.Itoa(cfg.QueryInterval))
		sw.urlList.Refresh()
	}
}
