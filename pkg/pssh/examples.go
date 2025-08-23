package pssh

// Example integration with the existing devices table
// This file shows how to integrate the pssh package with your existing application

/*
Example usage in your devices table widget:

1. Import the pssh package:
   import "ispappclient/pkg/pssh"

2. Add SSH functionality to your devices table:

func CreateDevicesTableWithSSH() *fyne.Container {
    table := CreateDevicesTable() // Your existing function

    // Add context menu for SSH operations
    table.OnSecondaryTap = func(ev *fyne.PointEvent) {
        // Get selected devices (you'll need to implement device selection)
        selectedDevices := getSelectedDevices() // Implement this function

        if len(selectedDevices) > 0 {
            menu := fyne.NewMenu("SSH Options",
                pssh.CreateSSHMenuActions(selectedDevices, yourWindow)...)
            widget.ShowPopUpMenuAtPosition(menu, yourWindow.Canvas(), ev.AbsolutePosition)
        }
    }

    return table
}

3. Example of connecting to multiple devices:

func connectToSelectedDevices(devices []scanner.Device, window fyne.Window) {
    // Extract IP addresses from devices
    var ips []string
    for _, device := range devices {
        if device.HasSSH { // Check if SSH is available
            ips = append(ips, device.IP)
        }
    }

    if len(ips) == 0 {
        dialog.ShowInformation("No SSH Devices",
            "No devices with SSH support selected", window)
        return
    }

    // Show credentials dialog
    pssh.ShowCredentialsDialog(window, func(creds pssh.SSHCredentials, confirmed bool) {
        if !confirmed {
            return
        }

        // Connect to devices
        connections, err := pssh.ConnectToMultipleDevices(ips, creds, window)
        if err != nil {
            dialog.ShowError(err, window)
            return
        }

        // Open terminals in tabs
        err = pssh.OpenMultipleTerminals(connections, true)
        if err != nil {
            dialog.ShowError(err, window)
        }
    })
}

4. Integration with your main UI:

// In your main_ui.go file, add SSH buttons:
func createSSHControls(selectedDevicesFunc func() []scanner.Device, window fyne.Window) *fyne.Container {
    sshSingleBtn := widget.NewButton("SSH (Separate Windows)", func() {
        devices := selectedDevicesFunc()
        if len(devices) == 0 {
            dialog.ShowInformation("No Selection", "Please select devices first", window)
            return
        }
        connectToSelectedDevices(devices, window)
    })

    sshTabsBtn := widget.NewButton("SSH (Tabbed)", func() {
        devices := selectedDevicesFunc()
        if len(devices) == 0 {
            dialog.ShowInformation("No Selection", "Please select devices first", window)
            return
        }
        connectToSelectedDevices(devices, window)
    })

    return container.NewHBox(sshSingleBtn, sshTabsBtn)
}

5. Example of programmatic SSH connection (without UI):

func connectDirectly() {
    // Create SSH manager
    manager := pssh.NewSSHManager()

    // Create connections
    configs := []pssh.ConnectionConfig{
        pssh.NewConnectionConfig("192.168.1.1", 22, "admin", "password"),
        pssh.NewConnectionConfig("192.168.1.2", 22, "admin", "password"),
    }

    // Connect in parallel
    resultChan := manager.ConnectMultiple(configs)

    var connections []*pssh.SSHConnection
    for result := range resultChan {
        if result.Error != nil {
            fmt.Printf("Failed to connect to %s: %v\n", result.Host, result.Error)
            continue
        }
        connections = append(connections, result.Connection)
        fmt.Printf("Successfully connected to %s\n", result.Host)
    }

    // Create terminal manager and open terminals
    termManager := pssh.NewTerminalManager()
    err := termManager.MultiTerminalWindow(connections, "SSH Terminals")
    if err != nil {
        fmt.Printf("Failed to open terminals: %v\n", err)
    }
}
*/
