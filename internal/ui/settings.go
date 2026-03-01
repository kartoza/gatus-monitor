// Copyright (c) 2026 Kartoza
// SPDX-License-Identifier: MIT

// Package ui provides the settings user interface
package ui

import (
	"fmt"
	"log"
	"net/url"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/kartoza/gatus-monitor/internal/config"
	"github.com/kartoza/gatus-monitor/internal/icons"
)

// SettingsWindow manages the settings UI
type SettingsWindow struct {
	app             fyne.App
	window          fyne.Window
	config          *config.Manager
	onConfigChanged func()
	instanceList    *widget.List
	instances       []config.GatusInstance
	intervalEntry   *widget.Entry

	// Edit panel fields
	editMode      bool
	editIndex     int
	editNameEntry *widget.Entry
	editURLEntry  *widget.Entry
	editIconEntry *widget.Entry
	editPanel     *fyne.Container
	listPanel     *fyne.Container
}

// NewSettingsWindow creates a new settings window (creates its own Fyne app)
func NewSettingsWindow(cfg *config.Manager, onConfigChanged func()) *SettingsWindow {
	sw := &SettingsWindow{
		config:          cfg,
		onConfigChanged: onConfigChanged,
		instances:       make([]config.GatusInstance, len(cfg.Get().Instances)),
	}
	copy(sw.instances, cfg.Get().Instances)

	return sw
}

// NewSettingsWindowWithApp creates a new settings window with an existing Fyne app
func NewSettingsWindowWithApp(fyneApp fyne.App, cfg *config.Manager, onConfigChanged func()) *SettingsWindow {
	sw := &SettingsWindow{
		app:             fyneApp,
		config:          cfg,
		onConfigChanged: onConfigChanged,
		instances:       make([]config.GatusInstance, len(cfg.Get().Instances)),
	}
	copy(sw.instances, cfg.Get().Instances)

	return sw
}

// Show displays the settings window
func (sw *SettingsWindow) Show() {
	// Use fyne.Do() to ensure we're on the correct thread
	fyne.Do(func() {
		sw.showWindow()
	})
}

// showWindow creates and shows the window (must be called via fyne.Do())
func (sw *SettingsWindow) showWindow() {
	if sw.window != nil {
		sw.window.Show()
		sw.window.RequestFocus()
		return
	}

	if sw.app == nil {
		sw.app = app.New()
	}

	sw.window = sw.app.NewWindow("Gatus Monitor Settings")
	sw.buildUI()
	sw.window.Resize(fyne.NewSize(700, 500))
	sw.window.Show()
	sw.window.RequestFocus()
}

// buildUI constructs the settings UI
func (sw *SettingsWindow) buildUI() {
	// Header section (pinned to top)
	titleLabel := widget.NewLabel("Gatus Monitor Settings")
	titleLabel.TextStyle = fyne.TextStyle{Bold: true}

	header := container.NewVBox(
		titleLabel,
		widget.NewSeparator(),
	)

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
		widget.NewSeparator(),
	)

	// Instances list section (left panel)
	instanceLabel := widget.NewLabel("Gatus Instances:")
	addButton := widget.NewButton("Add New Instance", func() {
		sw.showEditPanel(-1)
	})

	sw.instanceList = widget.NewList(
		func() int {
			return len(sw.instances)
		},
		func() fyne.CanvasObject {
			return container.NewBorder(nil, nil, nil,
				container.NewHBox(
					widget.NewButton("Edit", func() {}),
					widget.NewButton("Remove", func() {}),
				),
				widget.NewLabel(""),
			)
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			border := item.(*fyne.Container)
			label := border.Objects[0].(*widget.Label)
			buttonBox := border.Objects[1].(*fyne.Container)
			editButton := buttonBox.Objects[0].(*widget.Button)
			removeButton := buttonBox.Objects[1].(*widget.Button)

			instance := sw.instances[id]
			label.SetText(fmt.Sprintf("%s - %s", instance.Name, instance.URL))

			editButton.OnTapped = func() {
				sw.showEditPanel(id)
			}
			removeButton.OnTapped = func() {
				sw.removeInstance(id)
			}
		},
	)

	sw.listPanel = container.NewBorder(
		container.NewVBox(instanceLabel, addButton),
		nil, nil, nil,
		sw.instanceList,
	)

	// Edit panel (right panel, initially hidden)
	sw.editPanel = sw.buildEditPanel()

	// Instruction label above the split pane
	instructionLabel := widget.NewLabel("Select an instance from the list to edit, or click 'Add New Instance' to create one.")
	instructionLabel.Wrapping = fyne.TextWrapWord
	instructionLabel.TextStyle = fyne.TextStyle{Italic: true}

	// Instances section with list on left, edit panel on right
	instancesContainer := container.NewHSplit(
		sw.listPanel,
		sw.editPanel,
	)
	instancesContainer.SetOffset(0.5)

	// Body content - use Border layout so instancesContainer expands to fill space
	bodyContent := container.NewBorder(
		container.NewVBox(intervalContainer, instructionLabel),
		nil, nil, nil,
		instancesContainer,
	)

	// Footer section (pinned to bottom)
	saveButton := widget.NewButton("Save", sw.save)
	cancelButton := widget.NewButton("Cancel", func() {
		sw.Hide()
	})

	// Create clickable links for footer
	kartozaLink := widget.NewHyperlink("Kartoza", parseURL("https://kartoza.com"))
	donateLink := widget.NewHyperlink("Donate!", parseURL("https://github.com/sponsors/kartoza"))
	githubLink := widget.NewHyperlink("GitHub", parseURL("https://github.com/kartoza/gatus-monitor"))

	// Attribution footer
	footerAttribution := container.NewHBox(
		widget.NewLabel("Made with"),
		widget.NewLabel("💗"),
		widget.NewLabel("by"),
		kartozaLink,
		widget.NewLabel("|"),
		donateLink,
		widget.NewLabel("|"),
		githubLink,
	)

	footer := container.NewVBox(
		widget.NewSeparator(),
		container.NewHBox(
			saveButton,
			cancelButton,
		),
		container.NewCenter(footerAttribution),
	)

	// Main layout: header at top, footer at bottom, body in center (fills available space)
	content := container.NewBorder(
		header,
		footer,
		nil, nil,
		bodyContent,
	)

	sw.window.SetContent(content)
}

// buildEditPanel creates the edit panel for instances
func (sw *SettingsWindow) buildEditPanel() *fyne.Container {
	sw.editNameEntry = widget.NewEntry()
	sw.editNameEntry.SetPlaceHolder("My Gatus Server")

	sw.editURLEntry = widget.NewEntry()
	sw.editURLEntry.SetPlaceHolder("https://status.example.com")

	sw.editIconEntry = widget.NewEntry()
	sw.editIconEntry.SetPlaceHolder("https://example.com/favicon.ico (optional)")

	// Initially empty - instructions are now above the split pane
	return container.NewVBox()
}

// showEditPanel shows the edit panel for adding or editing an instance
func (sw *SettingsWindow) showEditPanel(index int) {
	sw.editMode = true
	sw.editIndex = index

	if index >= 0 && index < len(sw.instances) {
		// Edit existing instance
		instance := sw.instances[index]
		sw.editNameEntry.SetText(instance.Name)
		sw.editURLEntry.SetText(instance.URL)
		sw.editIconEntry.SetText(instance.IconURL)
	} else {
		// Add new instance
		sw.editNameEntry.SetText("")
		sw.editURLEntry.SetText("")
		sw.editIconEntry.SetText("")
	}

	// Rebuild edit panel with form
	saveInstanceButton := widget.NewButton("Save Instance", sw.saveInstance)
	cancelEditButton := widget.NewButton("Cancel", sw.hideEditPanel)

	form := container.NewVBox(
		widget.NewLabel("Instance Details:"),
		widget.NewSeparator(),
		widget.NewLabel("Friendly Name:"),
		sw.editNameEntry,
		widget.NewLabel("Gatus URL:"),
		sw.editURLEntry,
		widget.NewLabel("Icon URL (optional):"),
		sw.editIconEntry,
		widget.NewLabel("Leave blank to auto-detect icon"),
		widget.NewSeparator(),
		container.NewHBox(saveInstanceButton, cancelEditButton),
	)

	sw.editPanel.Objects = []fyne.CanvasObject{form}
	sw.editPanel.Refresh()
}

// hideEditPanel hides the edit panel
func (sw *SettingsWindow) hideEditPanel() {
	sw.editMode = false
	sw.editIndex = -1

	// Clear the edit panel - instructions are above the split pane
	sw.editPanel.Objects = []fyne.CanvasObject{}
	sw.editPanel.Refresh()
}

// saveInstance saves the current instance being edited
func (sw *SettingsWindow) saveInstance() {
	name := sw.editNameEntry.Text
	url := sw.editURLEntry.Text
	iconURL := sw.editIconEntry.Text

	if name == "" {
		sw.showError("Name cannot be empty")
		return
	}

	if url == "" {
		sw.showError("URL cannot be empty")
		return
	}

	// Fetch icon data if needed
	var iconData []byte
	var err error

	if sw.editIndex >= 0 && sw.editIndex < len(sw.instances) {
		// Editing existing instance - only fetch if URL or IconURL changed
		instance := sw.instances[sw.editIndex]
		iconData = instance.IconData
		if url != instance.URL || iconURL != instance.IconURL {
			iconData, err = icons.FetchIcon(url, iconURL)
			if err != nil {
				log.Printf("Warning: Failed to fetch icon for %s: %v", name, err)
			}
		}
	} else {
		// New instance - fetch icon
		iconData, err = icons.FetchIcon(url, iconURL)
		if err != nil {
			log.Printf("Warning: Failed to fetch icon for %s: %v", name, err)
		}
	}

	// Create instance
	newInstance := config.GatusInstance{
		Name:     name,
		URL:      url,
		IconURL:  iconURL,
		IconData: iconData,
	}

	// Update or add instance
	if sw.editIndex >= 0 && sw.editIndex < len(sw.instances) {
		sw.instances[sw.editIndex] = newInstance
	} else {
		sw.instances = append(sw.instances, newInstance)
	}

	// Validate
	tempConfig := &config.Config{
		QueryInterval: sw.config.Get().QueryInterval,
		Instances:     sw.instances,
	}

	if err := config.Validate(tempConfig); err != nil {
		sw.showError(err.Error())
		return
	}

	// Refresh list and hide edit panel
	sw.instanceList.Refresh()
	sw.hideEditPanel()
}

// showError displays an error message inline
func (sw *SettingsWindow) showError(message string) {
	errorLabel := widget.NewLabel(message)
	errorLabel.Wrapping = fyne.TextWrapWord
	errorLabel.TextStyle = fyne.TextStyle{Bold: true}

	okButton := widget.NewButton("OK", func() {
		sw.showEditPanel(sw.editIndex)
	})

	sw.editPanel.Objects = []fyne.CanvasObject{
		container.NewVBox(
			widget.NewLabel("Error"),
			widget.NewSeparator(),
			errorLabel,
			widget.NewSeparator(),
			okButton,
		),
	}
	sw.editPanel.Refresh()
}

// removeInstance removes an instance from the list
func (sw *SettingsWindow) removeInstance(index int) {
	if index < 0 || index >= len(sw.instances) {
		return
	}

	// Remove instance
	sw.instances = append(sw.instances[:index], sw.instances[index+1:]...)
	sw.instanceList.Refresh()
}

// save saves the configuration
func (sw *SettingsWindow) save() {
	// Parse interval
	intervalStr := sw.intervalEntry.Text
	interval, err := strconv.Atoi(intervalStr)
	if err != nil {
		log.Printf("Invalid interval: %v", err)
		return
	}

	// Create new config
	newConfig := &config.Config{
		QueryInterval: interval,
		Instances:     sw.instances,
	}

	// Validate
	if err := config.Validate(newConfig); err != nil {
		log.Printf("Invalid config: %v", err)
		return
	}

	// Update config
	if err := sw.config.Update(newConfig); err != nil {
		log.Printf("Failed to save config: %v", err)
		return
	}

	// Notify about config change
	if sw.onConfigChanged != nil {
		sw.onConfigChanged()
	}

	log.Println("Configuration saved successfully")

	// Close window
	sw.Hide()
}

// Hide hides the settings window
func (sw *SettingsWindow) Hide() {
	fyne.Do(func() {
		if sw.window != nil {
			sw.window.Hide()
		}
	})
}

// Reload reloads the configuration from the manager
func (sw *SettingsWindow) Reload() {
	cfg := sw.config.Get()
	sw.instances = make([]config.GatusInstance, len(cfg.Instances))
	copy(sw.instances, cfg.Instances)

	// Only update UI elements if window has been created
	fyne.Do(func() {
		if sw.window != nil && sw.intervalEntry != nil {
			sw.intervalEntry.SetText(strconv.Itoa(cfg.QueryInterval))
			sw.instanceList.Refresh()
		}
	})
}

// parseURL parses a URL string and returns a *url.URL, logging any errors
func parseURL(urlStr string) *url.URL {
	u, err := url.Parse(urlStr)
	if err != nil {
		log.Printf("Failed to parse URL %s: %v", urlStr, err)
		return nil
	}
	return u
}
