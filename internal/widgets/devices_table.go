package widgets

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/ispapp/psshclient/internal/data"
	"github.com/ispapp/psshclient/internal/scanner"
	"github.com/ispapp/psshclient/internal/settings"
	"github.com/ispapp/psshclient/pkg/pssh"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// CreateDevicesTable creates a table widget to display discovered devices
func CreateDevicesTable() *fyne.Container {
	return CreateDevicesTableWithWindow(nil, nil)
}

// CreateDevicesTableWithWindow creates a table widget with SSH functionality
func CreateDevicesTableWithWindow(parentWindow fyne.Window, app fyne.App) *fyne.Container {
	// Create table headers
	headers := []string{"Select", "IP Address", "Hostname", "SSH", "SSH Port", "Username", "Password", "Status", "Actions"}

	// Track selected devices and SSH manager
	selectedDevices := make(map[int]bool)
	sshManager := pssh.NewSSHManager()
	var selectionMutex sync.Mutex

	// Create table widget
	table := widget.NewTable(
		func() (int, int) {
			// Return rows, columns
			deviceCount := data.DeviceList.Length()
			return deviceCount + 1, len(headers) // +1 for header row
		},
		func() fyne.CanvasObject {
			// Create cell template - using label as base template
			return widget.NewLabel("template")
		},
		func(id widget.TableCellID, obj fyne.CanvasObject) {
			// Clear and rebuild cell content
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

							case 3: // SSH Status
								if device.Port22 {
									if device.Connected {
										label.SetText("✓ Connected")
									} else {
										label.SetText("✓ Available")
									}
								} else {
									label.SetText("✗ Closed")
								}
							case 4: // SSH Port
								if device.Port22 {
									if device.SSHPort == 0 {
										device.SSHPort = settings.Current.DefaultSSHPort
									}
									label.SetText(fmt.Sprintf("%d", device.SSHPort))
								} else {
									label.SetText("-")
								}

							case 5: // Username
								if device.Port22 {
									label.SetText(device.Username)
								} else {
									label.SetText("-")
								}

							case 6: // Password
								if device.Port22 {
									if device.Password != "" {
										label.SetText("●●●●●●")
									} else {
										label.SetText("")
									}
								} else {
									label.SetText("-")
								}

							case 7: // Overall Status
								label.SetText(device.Status)

							case 8: // Actions
								if device.Port22 {
									if device.Connected {
										label.SetText("🔌 Disconnect")
									} else {
										label.SetText("🔗 Connect")
									}
								} else {
									label.SetText("-")
								}
							}
							label.TextStyle.Bold = false
						}
					}
				}
			}
		})

	// Handle cell taps for actions
	table.OnSelected = func(id widget.TableCellID) {
		if id.Row > 0 { // Skip header row
			deviceIndex := id.Row - 1
			switch id.Col {
			case 0: // Selection column
				selectionMutex.Lock()
				selectedDevices[deviceIndex] = !selectedDevices[deviceIndex]
				selectionMutex.Unlock()
				table.Refresh()
			case 4: // SSH Port column - show entry dialog
				if deviceIndex < data.DeviceList.Length() {
					if deviceObj, err := data.DeviceList.GetValue(deviceIndex); err == nil {
						if device, ok := deviceObj.(scanner.Device); ok && device.Port22 {
							showSSHPortDialog(deviceIndex, device.SSHPort, parentWindow, table)
						}
					}
				}
			case 5: // Username column - show entry dialog
				if deviceIndex < data.DeviceList.Length() {
					if deviceObj, err := data.DeviceList.GetValue(deviceIndex); err == nil {
						if device, ok := deviceObj.(scanner.Device); ok && device.Port22 {
							showUsernameDialog(deviceIndex, device.Username, parentWindow, table)
						}
					}
				}
			case 6: // Password column - show entry dialog
				if deviceIndex < data.DeviceList.Length() {
					if deviceObj, err := data.DeviceList.GetValue(deviceIndex); err == nil {
						if device, ok := deviceObj.(scanner.Device); ok && device.Port22 {
							showPasswordDialog(deviceIndex, device.Password, parentWindow, table)
						}
					}
				}
			case 8: // Actions column - connect/disconnect
				if deviceIndex < data.DeviceList.Length() {
					if deviceObj, err := data.DeviceList.GetValue(deviceIndex); err == nil {
						if device, ok := deviceObj.(scanner.Device); ok && device.Port22 {
							connectToDevice(deviceIndex, sshManager, parentWindow, table)
						}
					}
				}
			}
		}
	}

	// Set column widths for better layout
	table.SetColumnWidth(0, 60)  // Select checkbox
	table.SetColumnWidth(1, 120) // IP Address
	table.SetColumnWidth(2, 150) // Hostname
	table.SetColumnWidth(3, 100) // SSH Status
	table.SetColumnWidth(4, 80)  // SSH Port
	table.SetColumnWidth(5, 100) // Username
	table.SetColumnWidth(6, 100) // Password
	table.SetColumnWidth(7, 80)  // Status
	table.SetColumnWidth(8, 100) // Actions

	// Listen for changes to the device list
	data.DeviceList.AddListener(binding.NewDataListener(func() {
		table.Refresh()
	}))

	// Create container with table and status
	statusLabel := widget.NewLabel("No devices scanned yet")

	// Create SSH control buttons
	var sshControls *fyne.Container
	if parentWindow != nil && app != nil {
		sshControls = createSSHControls(selectedDevices, sshManager, parentWindow, app)
	}

	// Update status label when device list changes
	data.DeviceList.AddListener(binding.NewDataListener(func() {
		count := data.DeviceList.Length()
		sshCount := getSSHDeviceCount()

		// Get database status
		dbStatus := ""
		if data.DB != nil {
			if dbCount, err := data.DB.GetDeviceCount(); err == nil {
				dbStatus = fmt.Sprintf(" (DB: %d)", dbCount)
			}
		}

		if count == 0 {
			statusLabel.SetText("No devices found" + dbStatus)
		} else {
			statusLabel.SetText(fmt.Sprintf("Found %d device(s), %d with SSH%s", count, sshCount, dbStatus))
		}
	}))

	// Bind scanning status
	scanningLabel := widget.NewLabel("")
	scanningLabel.Bind(data.ScanProgress)

	// Create top section with status and SSH controls
	var topSection *fyne.Container
	if sshControls != nil {
		topSection = container.NewVBox(
			sshControls,
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
func createSSHControls(selectedDevices map[int]bool, sshManager *pssh.SSHManager, parentWindow fyne.Window, app fyne.App) *fyne.Container {
	// Multi-Device SSH Terminal button using new terminal widget
	sshTerminalBtn := widget.NewButtonWithIcon("Terminal", theme.ComputerIcon(), func() {
		var connections []*pssh.SSHConnection

		// Get connected devices that are selected
		for deviceIndex, selected := range selectedDevices {
			if !selected {
				continue
			}

			if deviceIndex < data.DeviceList.Length() {
				if deviceObj, err := data.DeviceList.GetValue(deviceIndex); err == nil {
					if device, ok := deviceObj.(scanner.Device); ok && device.Connected {
						// Get the SSH connection from the shared manager
						if conn, exists := sshManager.GetConnection(device.IP); exists {
							fmt.Printf("Adding connection for %s to terminal\n", device.IP)
							connections = append(connections, conn)
						} else {
							fmt.Printf("Connection not found in manager for %s\n", device.IP)
						}
					} else {
						fmt.Printf("Device %s is not connected or not found\n", device.IP)
					}
				}
			}
		}

		fmt.Printf("Total connections found: %d\n", len(connections))

		if len(connections) == 0 {
			dialog.ShowInformation("No SSH Connections",
				"Please connect to devices first using the Connect button in each row", parentWindow)
			return
		}

		// Create new SSH multi-terminal
		fmt.Printf("Creating terminal manager...\n")
		err := pssh.OpenMultipleTerminals(connections, parentWindow)
		if err != nil {
			fmt.Printf("Failed to create multi-terminal: %v\n", err)
			dialog.ShowError(err, parentWindow)
			return
		}
	})

	// Multi-Device Script Runner button
	runScriptBtn := widget.NewButtonWithIcon("Run Script", theme.MediaPlayIcon(), func() {
		var connections []*pssh.SSHConnection

		// Get connected devices that are selected
		for deviceIndex, selected := range selectedDevices {
			if !selected {
				continue
			}

			if deviceIndex < data.DeviceList.Length() {
				if deviceObj, err := data.DeviceList.GetValue(deviceIndex); err == nil {
					if device, ok := deviceObj.(scanner.Device); ok && device.Connected {
						// Get the SSH connection from the shared manager
						if conn, exists := sshManager.GetConnection(device.IP); exists {
							connections = append(connections, conn)
						}
					}
				}
			}
		}

		if len(connections) == 0 {
			dialog.ShowInformation("No SSH Connections",
				"Please connect to and select devices first.", parentWindow)
			return
		}

		showScriptDialog(connections, parentWindow, app)
	})

	// Select All SSH button
	selectAllSSHBtn := widget.NewButtonWithIcon("Select All", theme.ConfirmIcon(), func() {
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
				fmt.Sprintf("Selected %d devices with SSH support.", selectedCount),
				parentWindow)
		} else {
			dialog.ShowInformation("No SSH Devices",
				"No devices with SSH support found", parentWindow)
		}
	})

	// Clear Selection button
	clearSelectionBtn := widget.NewButtonWithIcon("Clear", theme.ContentClearIcon(), func() {
		// Clear all selections
		for k := range selectedDevices {
			delete(selectedDevices, k)
		}
		dialog.ShowInformation("Selection Cleared",
			"All device selections have been cleared.",
			parentWindow)
	})

	// Load Recent button - loads devices from last cleanup setting
	loadRecentBtn := widget.NewButtonWithIcon("Recent", theme.HistoryIcon(), func() {
		data.LoadRecentDevicesFromDB(settings.Current.GetCleanupDuration())
		dialog.ShowInformation("Devices Loaded",
			fmt.Sprintf("Recent devices from the last %d days have been loaded from the database.", settings.Current.CleanupOldDays),
			parentWindow)
	})

	// Load All button - loads all devices from database
	loadAllBtn := widget.NewButtonWithIcon("Load All", theme.FolderOpenIcon(), func() {
		data.LoadDevicesFromDB()
		dialog.ShowInformation("Devices Loaded",
			"All saved devices have been loaded from the database.",
			parentWindow)
	})

	// Save Current button - saves current device list to database
	saveCurrentBtn := widget.NewButtonWithIcon("Save", theme.DocumentSaveIcon(), func() {
		data.SaveDevicesToDB()
		dialog.ShowInformation("Devices Saved",
			"Current device list has been saved to the database.",
			parentWindow)
	})

	// Clear Database button - clears all devices from database
	clearDBBtn := widget.NewButtonWithIcon("Clear", theme.DeleteIcon(), func() {
		dialog.ShowConfirm("Clear Database",
			"Are you sure you want to clear all devices from the database? This cannot be undone.",
			func(confirmed bool) {
				if confirmed {
					data.ClearDevicesAndDB()
					dialog.ShowInformation("Database Cleared",
						"All devices have been removed from the database.",
						parentWindow)
				}
			}, parentWindow)
	})

	// Create database controls section
	dbControls := container.NewHBox(
		loadRecentBtn,
		loadAllBtn,
		saveCurrentBtn,
		clearDBBtn,
	)

	// Create SSH controls section
	sshControls := container.NewHBox(
		selectAllSSHBtn,
		clearSelectionBtn,
		widget.NewSeparator(),
		sshTerminalBtn,
		runScriptBtn,
	)

	// Combine both sections with a separator
	parrentwidth := parentWindow.Canvas().Size().Width
	return container.NewGridWrap(
		fyne.NewSize(parrentwidth, 24),
		container.NewHBox(widget.NewSeparator(), widget.NewIcon(theme.StorageIcon()), dbControls, widget.NewIcon(theme.ComputerIcon()), sshControls),
	)
}

// showScriptDialog shows a dialog to run a script on multiple devices
func showScriptDialog(connections []*pssh.SSHConnection, parent fyne.Window, app fyne.App) {
	scriptInput := widget.NewMultiLineEntry()
	scriptInput.SetPlaceHolder("Enter script to run on all selected devices...")
	scriptInput.SetMinRowsVisible(10)
	scriptInput.Wrapping = fyne.TextWrapOff

	outputBox := container.NewVBox()
	outputScroll := container.NewVScroll(outputBox)
	outputScroll.SetMinSize(fyne.NewSize(600, 300))

	var runBtn *widget.Button
	runBtn = widget.NewButton("Run Script", func() {
		script := scriptInput.Text
		if script == "" {
			return
		}

		runBtn.Disable()
		outputBox.RemoveAll()
		progress := widget.NewProgressBarInfinite()
		outputBox.Add(progress)

		go func() {
			var results []string
			var mu sync.Mutex
			var wg sync.WaitGroup

			for _, conn := range connections {
				wg.Add(1)
				go func(c *pssh.SSHConnection) {
					defer wg.Done()
					output, err := c.RunCommand(script)
					var resultText string
					if err != nil {
						resultText = fmt.Sprintf("--- ERROR on %s ---\n%s\n", c.Config.Host, err.Error())
					} else {
						resultText = fmt.Sprintf("--- Output from %s ---\n%s\n", c.Config.Host, output)
					}
					mu.Lock()
					results = append(results, resultText)
					mu.Unlock()
				}(conn)
			}
			wg.Wait()

			// UI updates can be done directly, but need to be refreshed
			runBtn.Enable()
			outputBox.RemoveAll()
			var fullOutput string
			for _, res := range results {
				fullOutput += res + "\n"
			}
			// Use a single label with wrapped text for the full output
			outputLabel := widget.NewLabel(fullOutput)
			outputLabel.Wrapping = fyne.TextWrapWord
			outputBox.Add(outputLabel)
			outputBox.Refresh()
		}()
	})

	content := container.NewVBox(
		widget.NewLabel("Enter script:"),
		scriptInput,
		runBtn,
		widget.NewLabel("Output:"),
		outputScroll,
	)

	d := dialog.NewCustom("Run Script", "Close", content, parent)
	d.Resize(fyne.NewSize(700, 500))
	d.Show()
}

// showUsernameDialog shows a dialog to enter username for a device
func showUsernameDialog(deviceIndex int, currentUsername string, parent fyne.Window, table *widget.Table) {
	entry := widget.NewEntry()
	entry.SetText(currentUsername)
	entry.SetPlaceHolder("Enter username")

	dialog.ShowCustomConfirm("Enter Username", "OK", "Cancel", entry, func(confirmed bool) {
		if confirmed {
			updateDeviceField(deviceIndex, "username", entry.Text)
			table.Refresh()
		}
	}, parent)
}

// showSSHPortDialog shows a dialog to enter the SSH port for a device
func showSSHPortDialog(deviceIndex int, currentPort int, parent fyne.Window, table *widget.Table) {
	entry := widget.NewEntry()
	if currentPort == 0 {
		currentPort = settings.Current.DefaultSSHPort
	}
	entry.SetText(fmt.Sprintf("%d", currentPort))
	entry.SetPlaceHolder("Enter SSH port")

	dialog.ShowCustomConfirm("Enter SSH Port", "OK", "Cancel", entry, func(confirmed bool) {
		if confirmed {
			updateDeviceField(deviceIndex, "sshport", entry.Text)
			table.Refresh()
		}
	}, parent)
}

// showPasswordDialog shows a dialog to enter password for a device
func showPasswordDialog(deviceIndex int, currentPassword string, parent fyne.Window, table *widget.Table) {
	entry := widget.NewPasswordEntry()
	entry.SetText(currentPassword)
	entry.SetPlaceHolder("Enter password")

	dialog.ShowCustomConfirm("Enter Password", "OK", "Cancel", entry, func(confirmed bool) {
		if confirmed {
			updateDeviceField(deviceIndex, "password", entry.Text)
			table.Refresh()
		}
	}, parent)
}

// updateDeviceField updates a specific field of a device in the device list
func updateDeviceField(deviceIndex int, field, value string) {
	if deviceIndex < data.DeviceList.Length() {
		if deviceObj, err := data.DeviceList.GetValue(deviceIndex); err == nil {
			if device, ok := deviceObj.(scanner.Device); ok {
				switch field {
				case "username":
					device.Username = value
					data.UpdateDevice(deviceIndex, device)
				case "password":
					device.Password = value
					data.UpdateDevice(deviceIndex, device)
				case "sshport":
					if port, err := strconv.Atoi(value); err == nil {
						device.SSHPort = port
						data.UpdateDevice(deviceIndex, device)
					}
				}
			}
		}
	}
}

// connectToDevice handles SSH connection/disconnection for a device
func connectToDevice(deviceIndex int, sshManager *pssh.SSHManager, parentWindow fyne.Window, table *widget.Table) {
	if deviceIndex < data.DeviceList.Length() {
		if deviceObj, err := data.DeviceList.GetValue(deviceIndex); err == nil {
			if device, ok := deviceObj.(scanner.Device); ok {
				if device.Connected {
					// Disconnect
					fmt.Printf("Disconnecting from %s...\n", device.IP)
					if conn, exists := sshManager.GetConnection(device.IP); exists {
						err := conn.Close()
						if err != nil {
							fmt.Printf("Error closing connection to %s: %v\n", device.IP, err)
						} else {
							fmt.Printf("Successfully disconnected from %s\n", device.IP)
						}
					} else {
						fmt.Printf("Connection not found in manager for %s\n", device.IP)
					}
					device.Connected = false
					device.Status = "Disconnected"
					data.UpdateDevice(deviceIndex, device)
				} else {
					// Connect
					if device.Username == "" || device.Password == "" {
						dialog.ShowError(fmt.Errorf("username and password are required"), parentWindow)
						return
					}

					sshPort := device.SSHPort
					if sshPort == 0 {
						sshPort = settings.Current.DefaultSSHPort
					}

					config := pssh.ConnectionConfig{
						Host:     device.IP,
						Port:     sshPort,
						Username: device.Username,
						Password: device.Password,
						Timeout:  settings.Current.GetConnectionTimeout(),
					}

					// Use the manager to connect (which stores the connection)
					fmt.Printf("Attempting to connect to %s...\n", device.IP)
					resultChan := sshManager.ConnectMultiple([]pssh.ConnectionConfig{config})

					// Process the result
					for result := range resultChan {
						if result.Error != nil {
							fmt.Printf("Connection failed for %s: %v\n", device.IP, result.Error)
							dialog.ShowError(fmt.Errorf("failed to connect to %s: %v", device.IP, result.Error), parentWindow)
							return
						}

						if result.Connection.Connected {
							fmt.Printf("Successfully connected to %s\n", device.IP)
							device.Connected = true
							device.Status = "Connected"
							data.UpdateDevice(deviceIndex, device)
						} else {
							fmt.Printf("Connection to %s reported as not connected\n", device.IP)
						}
					}
				}
				table.Refresh()
			}
		}
	}
}
