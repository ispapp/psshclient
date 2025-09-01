package dialogs

import (
	"fmt"
	"strconv"
	"time"

	"github.com/ispapp/psshclient/internal/data"
	"github.com/ispapp/psshclient/internal/scanner"
	"github.com/ispapp/psshclient/internal/settings"
	"github.com/ispapp/psshclient/internal/windows"
	"github.com/ispapp/psshclient/pkg/goneighbors"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// ShowNeighborDiscoveryDialog shows the neighbor discovery dialog with minimal design
func ShowNeighborDiscoveryDialog(wins *windows.WindowManager) {
	fyne.Do(func() {
		// Create interval input
		intervalEntry := widget.NewEntry()
		intervalEntry.SetText("5")
		intervalEntry.SetPlaceHolder("Scan interval in seconds")

		// Start/Stop button
		startStopBtn := widget.NewButton("Start Discovery", nil)

		// Status label
		statusLabel := widget.NewLabel("Ready to discover neighbors")

		// Create results table with 6 columns (added checkbox column)
		resultsTable := widget.NewTable(
			func() (int, int) { return 0, 6 },
			func() fyne.CanvasObject {
				// Create different widgets based on column
				return widget.NewLabel("")
			},
			func(id widget.TableCellID, object fyne.CanvasObject) {
				// Will be updated when results are available
			},
		)

		// Set column widths
		resultsTable.SetColumnWidth(0, 50)  // Select checkbox
		resultsTable.SetColumnWidth(1, 170) // IP Address
		resultsTable.SetColumnWidth(2, 140) // MAC Address
		resultsTable.SetColumnWidth(3, 150) // Identity
		resultsTable.SetColumnWidth(4, 120) // Platform
		resultsTable.SetColumnWidth(5, 80)  // Protocol

		// Headers
		headers := []string{"Select", "IP Address", "MAC Address", "Identity", "Platform", "Protocol"}

		var discoveredNeighbors []goneighbors.Neighbor
		var selectedNeighbors = make(map[int]bool) // Track selected neighbors by index
		var _scanner *goneighbors.NeighborScanner
		var isScanning bool

		// Update table function
		updateTable := func(neighbors []goneighbors.Neighbor) {
			// Maintain fixed positions by not replacing the entire list
			// Instead, add new neighbors to the end and preserve existing ones
			existingIPs := make(map[string]int) // IP -> index mapping
			for i, existing := range discoveredNeighbors {
				existingIPs[existing.IPAddress] = i
			}

			// Add new neighbors that aren't already in the list
			for _, neighbor := range neighbors {
				if _, exists := existingIPs[neighbor.IPAddress]; !exists {
					discoveredNeighbors = append(discoveredNeighbors, neighbor)
				} else {
					// Update existing neighbor with new data while preserving position
					index := existingIPs[neighbor.IPAddress]
					discoveredNeighbors[index] = neighbor
				}
			}

			resultsTable.Length = func() (int, int) {
				return len(discoveredNeighbors) + 1, 6 // +1 for header, 6 columns
			}

			resultsTable.CreateCell = func() fyne.CanvasObject {
				return widget.NewLabel("")
			}

			resultsTable.UpdateCell = func(id widget.TableCellID, object fyne.CanvasObject) {
				label := object.(*widget.Label)

				if id.Row == 0 {
					// Header row
					label.SetText(headers[id.Col])
					label.TextStyle = fyne.TextStyle{Bold: true}
					return
				}

				// Data rows
				if id.Row-1 < len(discoveredNeighbors) {
					neighbor := discoveredNeighbors[id.Row-1]
					neighborIndex := id.Row - 1

					switch id.Col {
					case 0:
						// Select column - show checkbox status
						if selectedNeighbors[neighborIndex] {
							label.SetText("☑")
						} else {
							label.SetText("☐")
						}
					case 1:
						label.SetText(neighbor.IPAddress)
					case 2:
						label.SetText(neighbor.MACAddress)
					case 3:
						label.SetText(neighbor.Identity)
					case 4:
						label.SetText(neighbor.Platform)
					case 5:
						label.SetText(string(neighbor.Protocol))
					}
					label.TextStyle = fyne.TextStyle{}
				}
			}

			resultsTable.Refresh()
		}

		// Initially show empty table with headers
		updateTable([]goneighbors.Neighbor{})

		// Start/Stop discovery function
		toggleDiscovery := func() {
			if !isScanning {
				// Start discovery
				intervalStr := intervalEntry.Text
				interval, err := strconv.Atoi(intervalStr)
				if err != nil || interval <= 0 {
					dialog.ShowError(fmt.Errorf("invalid interval: %s", intervalStr), wins.GetMainWindow())
					return
				}

				_scanner = goneighbors.NewNeighborScanner()
				isScanning = true
				startStopBtn.SetText("Stop Discovery")
				statusLabel.SetText("Discovering neighbors...")

				go func() {
					// Start continuous discovery
					neighborChan := _scanner.GetUpdateChannel()

					// Start the discovery process with a single long-running session
					go func() {
						if !isScanning {
							return
						}

						// Start a single discovery session that runs until stopped
						_, err := _scanner.StartDiscoveryWithoutSSHCheck(time.Duration(24*60*60) * time.Second) // Run for 24 hours max
						if err != nil && isScanning {
							fyne.Do(func() {
								statusLabel.SetText(fmt.Sprintf("Error: %v", err))
							})
						}
					}()

					// Update UI periodically
					ticker := time.NewTicker(time.Duration(interval) * time.Second)
					defer ticker.Stop()

					for {
						if !isScanning {
							break
						}

						select {
						case <-ticker.C:
							// Update UI with current neighbors
							neighbors := _scanner.GetNeighbors()
							fyne.Do(func() {
								updateTable(neighbors)
								statusLabel.SetText(fmt.Sprintf("Found %d neighbors - scanning continues...", len(neighbors)))
							})

						case <-neighborChan:
							// Real-time update when new neighbor is discovered
							neighbors := _scanner.GetNeighbors()
							fyne.Do(func() {
								updateTable(neighbors)
								statusLabel.SetText(fmt.Sprintf("Found %d neighbors - scanning...", len(neighbors)))
							})
						}
					}
				}()
			} else {
				// Stop discovery
				isScanning = false
				if _scanner != nil {
					_scanner.Stop()
				}
				startStopBtn.SetText("Start Discovery")
				statusLabel.SetText(fmt.Sprintf("Discovery stopped - %d neighbors found", len(discoveredNeighbors)))
			}
		}

		startStopBtn.OnTapped = toggleDiscovery

		// Declare buttons first
		var addSelectedBtn *widget.Button
		var selectAllBtn *widget.Button
		var deselectAllBtn *widget.Button

		// Function to update button text with selection count
		updateButtonText := func() {
			selectedCount := 0
			for _, selected := range selectedNeighbors {
				if selected {
					selectedCount++
				}
			}

			if selectedCount > 0 {
				addSelectedBtn.SetText(fmt.Sprintf("Add Selected (%d) to Devices", selectedCount))
			} else {
				addSelectedBtn.SetText("Add Selected to Devices")
			}
		}

		// Handle table cell selection for checkbox toggling
		resultsTable.OnSelected = func(id widget.TableCellID) {
			if id.Row > 0 && id.Col == 0 && id.Row-1 < len(discoveredNeighbors) {
				// Toggle selection for checkbox column
				neighborIndex := id.Row - 1
				selectedNeighbors[neighborIndex] = !selectedNeighbors[neighborIndex]
				resultsTable.Refresh()
				updateButtonText()
			}
		}

		// Add selected devices button
		_ = settings.RefreshCurrent()
		addSelectedBtn = widget.NewButton("Add Selected to Devices", func() {
			addedCount := 0
			for index, selected := range selectedNeighbors {
				if selected && index < len(discoveredNeighbors) {
					neighbor := discoveredNeighbors[index]
					if neighbor.IPAddress != "" {
						device := scanner.Device{
							IP:        neighbor.IPAddress,
							Hostname:  neighbor.Identity,
							SSHStatus: true, // Assume devices have SSH
							// TELNETStatus: true,                                      // Assume devices have Telnet
							SSHPort:   settings.Current.DefaultSSHPort, // Default SSH port
							Status:    string(neighbor.Protocol),
							Username:  settings.Current.DefaultSSHUsername,
							Password:  settings.Current.DefaultSSHPassword,
							Connected: false,
						}

						data.AddDevice(device)
						addedCount++
					}
				}
			}

			if addedCount > 0 {
				statusLabel.SetText(fmt.Sprintf("Added %d device(s) to device list", addedCount))
				// Clear selections after adding
				selectedNeighbors = make(map[int]bool)
				resultsTable.Refresh()
				updateButtonText()
			} else {
				statusLabel.SetText("No devices selected to add")
			}
		})

		// Select all / Deselect all buttons
		selectAllBtn = widget.NewButton("Select All", func() {
			for i := range discoveredNeighbors {
				selectedNeighbors[i] = true
			}
			resultsTable.Refresh()
			updateButtonText()
		})

		deselectAllBtn = widget.NewButton("Deselect All", func() {
			selectedNeighbors = make(map[int]bool)
			resultsTable.Refresh()
			updateButtonText()
		})

		// Create minimal form layout
		topContainer := container.NewBorder(
			nil, nil,
			widget.NewLabel("Interval:"),
			startStopBtn,
			intervalEntry,
		)

		// Create button row for device management
		buttonRow := container.NewHBox(
			selectAllBtn,
			deselectAllBtn,
			widget.NewSeparator(),
			addSelectedBtn,
		)

		content := container.NewBorder(
			container.NewVBox(topContainer, statusLabel, buttonRow),
			nil, nil, nil,
			resultsTable,
		)

		// Create and show dialog
		win, err := wins.NewWindow("Neighbor Discovery", "neighbor_discovery") // Register window
		if err != nil {
			dialog.ShowError(err, wins.GetMainWindow())
			return
		}
		win.Window.SetContent(content)

		win.Window.Resize(fyne.NewSize(800, 600))

		// Clean up when dialog closes
		win.Window.SetOnClosed(func() {
			if isScanning && _scanner != nil {
				isScanning = false
				_scanner.Stop()
			}
		})
		win.Window.Show()
		win.Window.RequestFocus()
	})
}
