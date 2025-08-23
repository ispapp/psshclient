package widgets

import (
	"fmt"
	"ispappclient/internal/data"
	"ispappclient/internal/scanner"
	"ispappclient/pkg/pssh"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// CreateDevicesTable creates a table widget to display discovered devices
func CreateDevicesTable() *fyne.Container {
	return CreateDevicesTableWithWindow(nil, nil)
}

// CreateDevicesTableWithWindow creates a table widget with SSH functionality
func CreateDevicesTableWithWindow(parentWindow fyne.Window, app fyne.App) *fyne.Container {
	// Create table headers
	headers := []string{"Select", "IP Address", "Hostname", "SSH (22)", "Telnet (23)", "Status"}

	// Track selected devices
	selectedDevices := make(map[int]bool)
	var selectionMutex sync.Mutex

	// Create table widget
	table := widget.NewTable(
		func() (int, int) {
			// Return rows, columns
			deviceCount := data.DeviceList.Length()
			return deviceCount + 1, len(headers) // +1 for header row
		},
		func() fyne.CanvasObject {
			// Create cell template - use different templates for different columns
			return widget.NewLabel("template")
		},
		func(id widget.TableCellID, obj fyne.CanvasObject) {
			// Update cell content
			if id.Row == 0 {
				// Header row
				label := obj.(*widget.Label)
				if id.Col < len(headers) {
					label.SetText(headers[id.Col])
					label.TextStyle.Bold = true
				}
			} else {
				// Data row
				deviceIndex := id.Row - 1
				if deviceIndex < data.DeviceList.Length() {
					if deviceObj, err := data.DeviceList.GetValue(deviceIndex); err == nil {
						if device, ok := deviceObj.(scanner.Device); ok {
							label := obj.(*widget.Label)

							switch id.Col {
							case 0: // Selection checkbox
								selectionMutex.Lock()
								if selectedDevices[deviceIndex] {
									label.SetText("☑")
								} else {
									label.SetText("☐")
								}
								selectionMutex.Unlock()
							case 1: // IP Address
								label.SetText(device.IP)
							case 2: // Hostname
								label.SetText(device.Hostname)
							case 3: // SSH (22)
								if device.Port22 {
									label.SetText("✓ Open")
								} else {
									label.SetText("✗ Closed")
								}
							case 4: // Telnet (23)
								if device.Port23 {
									label.SetText("✓ Open")
								} else {
									label.SetText("✗ Closed")
								}
							case 5: // Status
								label.SetText(device.Status)
							}
							label.TextStyle.Bold = false
						}
					}
				}
			}
		})

	// Handle cell taps for selection
	table.OnSelected = func(id widget.TableCellID) {
		if id.Row > 0 { // Skip header row
			deviceIndex := id.Row - 1
			if id.Col == 0 { // Selection column
				selectionMutex.Lock()
				selectedDevices[deviceIndex] = !selectedDevices[deviceIndex]
				selectionMutex.Unlock()
				table.Refresh()
			}
		}
	}

	// Set column widths
	table.SetColumnWidth(0, 60)  // Select checkbox
	table.SetColumnWidth(1, 120) // IP Address
	table.SetColumnWidth(2, 150) // Hostname
	table.SetColumnWidth(3, 80)  // SSH
	table.SetColumnWidth(4, 80)  // Telnet
	table.SetColumnWidth(5, 80)  // Status

	// Listen for changes to the device list
	data.DeviceList.AddListener(binding.NewDataListener(func() {
		table.Refresh()
	}))

	// Create container with table and status
	statusLabel := widget.NewLabel("No devices scanned yet")

	// Create SSH control buttons
	var sshControls *fyne.Container
	if parentWindow != nil && app != nil {
		sshControls = createSSHControls(selectedDevices, parentWindow, app)
	}

	// Update status label when device list changes
	data.DeviceList.AddListener(binding.NewDataListener(func() {
		count := data.DeviceList.Length()
		sshCount := getSSHDeviceCount()
		if count == 0 {
			statusLabel.SetText("No devices found")
		} else {
			statusLabel.SetText(fmt.Sprintf("Found %d device(s), %d with SSH", count, sshCount))
		}
	}))

	// Bind scanning status
	scanningLabel := widget.NewLabel("")
	scanningLabel.Bind(data.ScanProgress)

	// Create top section with status and SSH controls
	var topSection *fyne.Container
	if sshControls != nil {
		topSection = container.NewVBox(
			container.NewHBox(statusLabel, widget.NewSeparator(), sshControls),
			scanningLabel,
		)
	} else {
		topSection = container.NewVBox(statusLabel, scanningLabel)
	}

	content := container.NewBorder(
		topSection, // Top
		nil,        // Bottom
		nil,        // Left
		nil,        // Right
		table,      // Center
	)

	return content
}

// Helper functions for SSH device management

// getSelectedSSHDevices returns the IP addresses of selected devices that have SSH enabled
func getSelectedSSHDevices(selectedDevices map[int]bool) []string {
	var sshDevices []string

	for deviceIndex, selected := range selectedDevices {
		if !selected {
			continue
		}

		if deviceIndex < data.DeviceList.Length() {
			if deviceObj, err := data.DeviceList.GetValue(deviceIndex); err == nil {
				if device, ok := deviceObj.(scanner.Device); ok && device.Port22 {
					sshDevices = append(sshDevices, device.IP)
				}
			}
		}
	}

	return sshDevices
}

// getSSHDeviceCount returns the total number of devices with SSH enabled
func getSSHDeviceCount() int {
	count := 0
	deviceCount := data.DeviceList.Length()

	for i := 0; i < deviceCount; i++ {
		if deviceObj, err := data.DeviceList.GetValue(i); err == nil {
			if device, ok := deviceObj.(scanner.Device); ok && device.Port22 {
				count++
			}
		}
	}

	return count
}

// createSSHControls creates SSH control buttons
func createSSHControls(selectedDevices map[int]bool, parentWindow fyne.Window, app fyne.App) *fyne.Container {
	// Multi-Device SSH Terminal button
	sshTerminalBtn := widget.NewButton("Multi-Device SSH Terminal", func() {
		selectedSSHDevices := getSelectedSSHDevices(selectedDevices)
		if len(selectedSSHDevices) == 0 {
			dialog.ShowInformation("No SSH Devices",
				"Please select devices with SSH support (port 22 open)", parentWindow)
			return
		}
		openSSHTerminals(selectedSSHDevices, parentWindow, app)
	})

	// Select All SSH button
	selectAllSSHBtn := widget.NewButton("Select All SSH", func() {
		// Clear current selection
		for k := range selectedDevices {
			delete(selectedDevices, k)
		}

		// Select all devices with SSH
		deviceCount := data.DeviceList.Length()
		selectedCount := 0
		for i := 0; i < deviceCount; i++ {
			if deviceObj, err := data.DeviceList.GetValue(i); err == nil {
				if device, ok := deviceObj.(scanner.Device); ok && device.Port22 {
					selectedDevices[i] = true
					selectedCount++
				}
			}
		}

		// Show feedback to user
		if selectedCount > 0 {
			dialog.ShowInformation("Selection Updated",
				fmt.Sprintf("Selected %d devices with SSH support. Click on a device row to refresh the display.", selectedCount),
				parentWindow)
		} else {
			dialog.ShowInformation("No SSH Devices",
				"No devices with SSH support found", parentWindow)
		}
	})

	// Clear Selection button
	clearSelectionBtn := widget.NewButton("Clear Selection", func() {
		// Clear all selections
		for k := range selectedDevices {
			delete(selectedDevices, k)
		}
		dialog.ShowInformation("Selection Cleared",
			"All device selections have been cleared. Click on a device row to refresh the display.",
			parentWindow)
	})

	return container.NewHBox(
		selectAllSSHBtn,
		clearSelectionBtn,
		widget.NewSeparator(),
		sshTerminalBtn,
	)
}

// openSSHTerminals handles opening SSH terminals for selected devices
func openSSHTerminals(devices []string, parent fyne.Window, app fyne.App) {
	pssh.ShowCredentialsDialog(parent, func(creds pssh.SSHCredentials, confirmed bool) {
		if !confirmed {
			return
		}

		// Connect to devices with progress tracking
		connections, err := pssh.ConnectToMultipleDevices(devices, creds, parent)
		if err != nil {
			dialog.ShowError(err, parent)
			return
		}

		if len(connections) == 0 {
			dialog.ShowInformation("No Connections",
				"No successful SSH connections were established", parent)
			return
		}

		// Open multi-device terminal
		err = pssh.OpenMultipleTerminals(connections, app)
		if err != nil {
			dialog.ShowError(err, parent)
		}
	})
}
