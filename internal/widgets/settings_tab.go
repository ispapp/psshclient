package widgets

import (
	"fmt"
	"ispappclient/internal/settings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
)

// CreateSettingsTab creates the settings tab content
func CreateSettingsTab(parentWindow fyne.Window) *container.Scroll {
	// Network Settings
	sshPortEntry := widget.NewEntry()
	sshPortEntry.SetText(settings.Current.GetDefaultSSHPortString())

	telnetPortEntry := widget.NewEntry()
	telnetPortEntry.SetText(settings.Current.GetDefaultTelnetPortString())

	timeoutEntry := widget.NewEntry()
	timeoutEntry.SetText(settings.Current.GetConnectionTimeoutString())

	// SSH Default Credentials
	usernameEntry := widget.NewEntry()
	usernameEntry.SetText(settings.Current.DefaultSSHUsername)

	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetText(settings.Current.DefaultSSHPassword)

	// Database Settings
	dbPathEntry := widget.NewEntry()
	dbPathEntry.SetText(settings.Current.DatabasePath)

	dbPathBtn := widget.NewButton("Browse", func() {
		fileDialog := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
			if err != nil {
				dialog.ShowError(err, parentWindow)
				return
			}
			if writer == nil {
				return // User cancelled
			}
			defer writer.Close()

			dbPathEntry.SetText(writer.URI().Path())
		}, parentWindow)

		fileDialog.SetFilter(storage.NewExtensionFileFilter([]string{".db"}))
		fileDialog.SetFileName("devices.db")
		fileDialog.Show()
	})

	autoSaveCheck := widget.NewCheck("Auto-save devices to database", func(checked bool) {
		settings.Current.AutoSaveDevices = checked
	})
	autoSaveCheck.SetChecked(settings.Current.AutoSaveDevices)

	cleanupDaysEntry := widget.NewEntry()
	cleanupDaysEntry.SetText(settings.Current.GetCleanupOldDaysString())

	// Terminal Settings
	termRowsEntry := widget.NewEntry()
	termRowsEntry.SetText(settings.Current.GetTerminalRowsString())

	termColsEntry := widget.NewEntry()
	termColsEntry.SetText(settings.Current.GetTerminalColsString())

	termFontEntry := widget.NewEntry()
	termFontEntry.SetText(settings.Current.TerminalFont)

	termFontSizeEntry := widget.NewEntry()
	termFontSizeEntry.SetText(settings.Current.GetTerminalFontSizeString())

	// Scanning Settings
	scanTimeoutEntry := widget.NewEntry()
	scanTimeoutEntry.SetText(settings.Current.GetScanTimeoutString())

	maxScansEntry := widget.NewEntry()
	maxScansEntry.SetText(settings.Current.GetMaxConcurrentScansString())

	// Save and Reset buttons
	saveBtn := widget.NewButton("Save Settings", func() {
		// Validate and save all settings
		var errors []string

		// Update settings from UI
		if err := settings.Current.SetDefaultSSHPortString(sshPortEntry.Text); err != nil {
			errors = append(errors, "Invalid SSH port: "+err.Error())
		}

		if err := settings.Current.SetDefaultTelnetPortString(telnetPortEntry.Text); err != nil {
			errors = append(errors, "Invalid Telnet port: "+err.Error())
		}

		if err := settings.Current.SetConnectionTimeoutString(timeoutEntry.Text); err != nil {
			errors = append(errors, "Invalid connection timeout: "+err.Error())
		}

		settings.Current.DefaultSSHUsername = usernameEntry.Text
		settings.Current.DefaultSSHPassword = passwordEntry.Text
		settings.Current.DatabasePath = dbPathEntry.Text

		if err := settings.Current.SetCleanupOldDaysString(cleanupDaysEntry.Text); err != nil {
			errors = append(errors, "Invalid cleanup days: "+err.Error())
		}

		if err := settings.Current.SetTerminalRowsString(termRowsEntry.Text); err != nil {
			errors = append(errors, "Invalid terminal rows: "+err.Error())
		}

		if err := settings.Current.SetTerminalColsString(termColsEntry.Text); err != nil {
			errors = append(errors, "Invalid terminal columns: "+err.Error())
		}

		settings.Current.TerminalFont = termFontEntry.Text

		if err := settings.Current.SetTerminalFontSizeString(termFontSizeEntry.Text); err != nil {
			errors = append(errors, "Invalid terminal font size: "+err.Error())
		}

		if err := settings.Current.SetScanTimeoutString(scanTimeoutEntry.Text); err != nil {
			errors = append(errors, "Invalid scan timeout: "+err.Error())
		}

		if err := settings.Current.SetMaxConcurrentScansString(maxScansEntry.Text); err != nil {
			errors = append(errors, "Invalid max concurrent scans: "+err.Error())
		}

		// Validate settings
		validationErrors := settings.Current.Validate()
		errors = append(errors, validationErrors...)

		if len(errors) > 0 {
			errorMsg := "Please fix the following errors:\n\n"
			for _, err := range errors {
				errorMsg += "• " + err + "\n"
			}
			dialog.ShowError(fmt.Errorf(errorMsg), parentWindow)
			return
		}

		// Save settings
		if err := settings.Save(); err != nil {
			dialog.ShowError(fmt.Errorf("Failed to save settings: %v", err), parentWindow)
			return
		}

		dialog.ShowInformation("Settings Saved", "All settings have been saved successfully!", parentWindow)
	})

	resetBtn := widget.NewButton("Reset to Defaults", func() {
		dialog.ShowConfirm("Reset Settings",
			"Are you sure you want to reset all settings to their default values?",
			func(confirmed bool) {
				if confirmed {
					// Reset to defaults
					defaults := settings.DefaultSettings()
					*settings.Current = *defaults

					// Update UI
					sshPortEntry.SetText(settings.Current.GetDefaultSSHPortString())
					telnetPortEntry.SetText(settings.Current.GetDefaultTelnetPortString())
					timeoutEntry.SetText(settings.Current.GetConnectionTimeoutString())
					usernameEntry.SetText(settings.Current.DefaultSSHUsername)
					passwordEntry.SetText(settings.Current.DefaultSSHPassword)
					dbPathEntry.SetText(settings.Current.DatabasePath)
					autoSaveCheck.SetChecked(settings.Current.AutoSaveDevices)
					cleanupDaysEntry.SetText(settings.Current.GetCleanupOldDaysString())
					termRowsEntry.SetText(settings.Current.GetTerminalRowsString())
					termColsEntry.SetText(settings.Current.GetTerminalColsString())
					termFontEntry.SetText(settings.Current.TerminalFont)
					termFontSizeEntry.SetText(settings.Current.GetTerminalFontSizeString())
					scanTimeoutEntry.SetText(settings.Current.GetScanTimeoutString())
					maxScansEntry.SetText(settings.Current.GetMaxConcurrentScansString())

					dialog.ShowInformation("Settings Reset", "All settings have been reset to default values.", parentWindow)
				}
			}, parentWindow)
	})

	// Create form sections
	networkSection := container.NewVBox(
		widget.NewCard("Network Settings", "", container.NewGridWithColumns(2,
			widget.NewLabel("Default SSH Port:"), sshPortEntry,
			widget.NewLabel("Default Telnet Port:"), telnetPortEntry,
			widget.NewLabel("Connection Timeout (seconds):"), timeoutEntry,
		)),
	)

	sshSection := container.NewVBox(
		widget.NewCard("SSH Default Credentials", "", container.NewGridWithColumns(2,
			widget.NewLabel("Default Username:"), usernameEntry,
			widget.NewLabel("Default Password:"), passwordEntry,
		)),
	)

	databaseSection := container.NewVBox(
		widget.NewCard("Database Settings", "", container.NewVBox(
			container.NewGridWithColumns(2,
				widget.NewLabel("Database Path:"), container.NewBorder(nil, nil, nil, dbPathBtn, dbPathEntry),
				widget.NewLabel("Cleanup old devices after (days):"), cleanupDaysEntry,
			),
			autoSaveCheck,
		)),
	)

	terminalSection := container.NewVBox(
		widget.NewCard("Terminal Settings", "", container.NewGridWithColumns(2,
			widget.NewLabel("Default Rows:"), termRowsEntry,
			widget.NewLabel("Default Columns:"), termColsEntry,
			widget.NewLabel("Font Family:"), termFontEntry,
			widget.NewLabel("Font Size:"), termFontSizeEntry,
		)),
	)

	scanningSection := container.NewVBox(
		widget.NewCard("Scanning Settings", "", container.NewGridWithColumns(2,
			widget.NewLabel("Scan Timeout (seconds):"), scanTimeoutEntry,
			widget.NewLabel("Max Concurrent Scans:"), maxScansEntry,
		)),
	)

	buttonsSection := container.NewHBox(
		saveBtn,
		resetBtn,
	)

	// Create scrollable content
	content := container.NewVBox(
		widget.NewLabel("Application Settings"),
		widget.NewSeparator(),
		networkSection,
		sshSection,
		databaseSection,
		terminalSection,
		scanningSection,
		widget.NewSeparator(),
		buttonsSection,
	)

	return container.NewScroll(content)
}
