package pssh

import (
	"fmt"
	"ispappclient/internal/resources"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// ConnectionProgress represents the progress of connecting to multiple devices
type ConnectionProgress struct {
	Total     int
	Connected int
	Failed    int
	Progress  *widget.ProgressBar
	Status    *widget.Label
}

// SSHCredentials holds the SSH login credentials
type SSHCredentials struct {
	Username   string
	Password   string
	PrivateKey []byte
	UseKey     bool
}

// ShowCredentialsDialog displays a dialog to collect SSH credentials
func ShowCredentialsDialog(parent fyne.Window, callback func(SSHCredentials, bool)) {
	usernameEntry := widget.NewEntry()
	usernameEntry.SetPlaceHolder("Username")

	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder("Password")

	keyCheck := widget.NewCheck("Use SSH Key", func(checked bool) {
		passwordEntry.Disable()
		if !checked {
			passwordEntry.Enable()
		}
	})

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Username:", Widget: usernameEntry},
			{Text: "Password:", Widget: passwordEntry},
			{Text: "", Widget: keyCheck},
		},
	}

	dialog.ShowCustomConfirm("SSH Credentials", "Connect", "Cancel", form, func(confirmed bool) {
		if confirmed {
			credentials := SSHCredentials{
				Username: usernameEntry.Text,
				Password: passwordEntry.Text,
				UseKey:   keyCheck.Checked,
			}
			callback(credentials, true)
		} else {
			callback(SSHCredentials{}, false)
		}
	}, parent)
}

// ShowConnectionProgress displays a progress dialog for connecting to multiple devices
func ShowConnectionProgress(parent fyne.Window, deviceCount int) *ConnectionProgress {
	progress := widget.NewProgressBar()
	progress.Max = float64(deviceCount)

	status := widget.NewLabel("Connecting to devices...")

	content := container.NewVBox(
		widget.NewLabel(fmt.Sprintf("Connecting to %d devices", deviceCount)),
		progress,
		status,
	)

	// Create a modal dialog
	d := dialog.NewCustom("Connection Progress", "Cancel", content, parent)
	d.Resize(fyne.NewSize(400, 150))
	d.Show()

	cp := &ConnectionProgress{
		Total:     deviceCount,
		Connected: 0,
		Failed:    0,
		Progress:  progress,
		Status:    status,
	}

	return cp
}

// UpdateProgress updates the connection progress
func (cp *ConnectionProgress) UpdateProgress(connected, failed int, currentHost string) {
	cp.Connected = connected
	cp.Failed = failed

	cp.Progress.SetValue(float64(connected + failed))
	cp.Status.SetText(fmt.Sprintf("Connected: %d, Failed: %d, Current: %s",
		connected, failed, currentHost))
}

// ConnectToMultipleDevices connects to multiple devices and shows progress
func ConnectToMultipleDevices(devices []string, credentials SSHCredentials, parent fyne.Window) ([]*SSHConnection, error) {
	if len(devices) == 0 {
		return nil, fmt.Errorf("no devices provided")
	}

	// Show progress dialog
	progressDialog := ShowConnectionProgress(parent, len(devices))

	// Create SSH manager
	manager := NewSSHManager()

	// Create connection configurations
	var configs []ConnectionConfig
	for _, device := range devices {
		config := NewConnectionConfig(device, 22, credentials.Username, credentials.Password)
		if credentials.UseKey {
			config.PrivateKey = credentials.PrivateKey
		}
		configs = append(configs, config)
	}

	// Connect in parallel
	resultChan := manager.ConnectMultiple(configs)

	var successfulConnections []*SSHConnection
	connected := 0
	failed := 0

	// Process results
	for result := range resultChan {
		if result.Error == nil {
			connected++
			successfulConnections = append(successfulConnections, result.Connection)
		} else {
			failed++
			fmt.Printf("Failed to connect to %s: %v\n", result.Host, result.Error)
		}

		progressDialog.UpdateProgress(connected, failed, result.Host)

		// Small delay to make progress visible
		time.Sleep(100 * time.Millisecond)
	}

	// Close progress dialog after a short delay
	go func() {
		time.Sleep(1 * time.Second)
		progressDialog.Status.SetText(fmt.Sprintf("Completed: %d connected, %d failed", connected, failed))
	}()

	if len(successfulConnections) == 0 {
		return nil, fmt.Errorf("failed to connect to any devices")
	}

	return successfulConnections, nil
}

// OpenMultipleTerminals opens a single multi-device terminal for SSH connections
func OpenMultipleTerminals(connections []*SSHConnection, parentWindow fyne.Window) error {
	if len(connections) == 0 {
		return fmt.Errorf("no connections provided")
	}

	// Create terminal manager
	termManager := NewTerminalManager()

	// Create multi-device terminal in main thread
	var multiTerm *SSHMultiTerminal
	var err error

	// Use Fyne's Do to ensure UI operations happen in main thread
	fyne.Do(func() {
		multiTerm, err = termManager.NewSSHMultiTerminal(connections)
		if err == nil {
			term := multiTerm.GetWidget()
			dialog := dialog.NewCustomWithoutButtons("", term, parentWindow)
			dialog.SetButtons([]fyne.CanvasObject{
				widget.NewButtonWithIcon("Close", theme.CancelIcon(), func() {
					dialog.Hide()
					term.Exit()
				}),
			})
			dialog.SetIcon(resources.ResourceIconPng)
			dialog.Resize(fyne.NewSize(800, 600))
			dialog.Show()
		}
	})

	if err != nil {
		return fmt.Errorf("failed to create multi-device terminal: %v", err)
	}

	return nil
}

// OpenMultipleTabbedTerminals opens multiple terminals in a tabbed interface
func OpenMultipleTabbedTerminals(connections []*SSHConnection) error {
	if len(connections) == 0 {
		return fmt.Errorf("no connections provided")
	}
	// Create terminal manager
	termManager := NewTerminalManager()
	// Create title with device count
	title := fmt.Sprintf("SSH Terminals (%d devices)", len(connections))
	// Create tabbed terminal window
	err := termManager.MultiTerminalWindow(connections, title)
	if err != nil {
		return fmt.Errorf("failed to create tabbed terminal window: %v", err)
	}
	return nil
}

// CreateSSHMenuActions creates menu actions for the devices table
func CreateSSHMenuActions(selectedDevices []string, parent fyne.Window, app fyne.App) []*fyne.MenuItem {
	return []*fyne.MenuItem{
		fyne.NewMenuItem("Open Multi-Device SSH Terminal", func() {
			ShowCredentialsDialog(parent, func(creds SSHCredentials, confirmed bool) {
				if !confirmed {
					return
				}

				connections, err := ConnectToMultipleDevices(selectedDevices, creds, parent)
				if err != nil {
					dialog.ShowError(err, parent)
					return
				}

				// Open multi-device terminal
				err = OpenMultipleTerminals(connections, parent)
				if err != nil {
					dialog.ShowError(err, parent)
				}
			})
		}),
		fyne.NewMenuItem("Open Tabbed SSH Terminals", func() {
			ShowCredentialsDialog(parent, func(creds SSHCredentials, confirmed bool) {
				if !confirmed {
					return
				}

				connections, err := ConnectToMultipleDevices(selectedDevices, creds, parent)
				if err != nil {
					dialog.ShowError(err, parent)
					return
				}

				// Open tabbed terminals
				err = OpenMultipleTabbedTerminals(connections)
				if err != nil {
					dialog.ShowError(err, parent)
				}
			})
		}),
	}
}
