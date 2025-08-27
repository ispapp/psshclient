package dialogs

import (
	"encoding/csv"
	"fmt"
	"io"
	"github.com/ispapp/psshclient/internal/data"
	"github.com/ispapp/psshclient/internal/scanner"
	"github.com/ispapp/psshclient/internal/settings"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
)

// CSVDevice represents a device from CSV with validation info
type CSVDevice struct {
	Device       scanner.Device
	Valid        bool
	Error        string
	Row          int
	OriginalPort string // Store the original port from CSV
}

// ShowCSVImportDialog shows a dialog for importing devices from CSV
func ShowCSVImportDialog(parent fyne.Window) {
	// File picker button
	var fileLabel *widget.Label
	var previewTable *widget.Table
	var importButton *widget.Button
	var csvDevices []CSVDevice

	fileLabel = widget.NewLabel("No file selected")

	filePickerBtn := widget.NewButton("Select CSV File", func() {
		fileDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil {
				dialog.ShowError(err, parent)
				return
			}
			if reader == nil {
				return // User cancelled
			}
			defer reader.Close()

			// Read and parse CSV
			devices, err := parseCSVFile(reader)
			if err != nil {
				dialog.ShowError(fmt.Errorf("failed to parse CSV: %v", err), parent)
				return
			}

			// Update UI
			fileLabel.SetText(fmt.Sprintf("File: %s (%d devices)", reader.URI().Name(), len(devices)))
			csvDevices = devices
			updatePreviewTable(previewTable, devices)

			// Enable import button if we have valid devices
			validCount := 0
			for _, dev := range devices {
				if dev.Valid {
					validCount++
				}
			}
			importButton.Enable()
			importButton.SetText(fmt.Sprintf("Import %d Valid Devices", validCount))
		}, parent)

		// Set file filter for CSV files
		fileDialog.SetFilter(storage.NewExtensionFileFilter([]string{".csv"}))
		fileDialog.Show()
	})

	// Preview table
	previewTable = widget.NewTable(
		func() (int, int) {
			if len(csvDevices) == 0 {
				return 1, 7 // Header only
			}
			return len(csvDevices) + 1, 7 // +1 for header
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("template")
		},
		func(id widget.TableCellID, obj fyne.CanvasObject) {
			label := obj.(*widget.Label)

			if id.Row == 0 {
				// Header row
				headers := []string{"Status", "IP", "Username", "Port", "Service", "Error", "Row"}
				if id.Col < len(headers) {
					label.SetText(headers[id.Col])
					label.TextStyle.Bold = true
				}
			} else {
				// Data row
				deviceIndex := id.Row - 1
				if deviceIndex < len(csvDevices) {
					device := csvDevices[deviceIndex]

					switch id.Col {
					case 0: // Status
						if device.Valid {
							label.SetText("✓ Valid")
						} else {
							label.SetText("✗ Invalid")
						}
					case 1: // IP
						label.SetText(device.Device.IP)
					case 2: // Username
						label.SetText(device.Device.Username)
					case 3: // Port
						label.SetText(device.OriginalPort)
					case 4: // Service
						if device.Device.Port22 {
							label.SetText("SSH")
						} else {
							label.SetText("-")
						}
					case 5: // Error
						label.SetText(device.Error)
					case 6: // Row
						label.SetText(fmt.Sprintf("%d", device.Row))
					}
					label.TextStyle.Bold = false
				}
			}
		})

	// Set column widths
	previewTable.SetColumnWidth(0, 80)  // Status
	previewTable.SetColumnWidth(1, 120) // IP
	previewTable.SetColumnWidth(2, 100) // Username
	previewTable.SetColumnWidth(3, 60)  // Port
	previewTable.SetColumnWidth(4, 80)  // Service
	previewTable.SetColumnWidth(5, 150) // Error
	previewTable.SetColumnWidth(6, 50)  // Row

	// Import button
	importButton = widget.NewButton("Import Devices", func() {
		importedCount := 0
		errorCount := 0

		for _, csvDevice := range csvDevices {
			if csvDevice.Valid {
				data.AddDevice(csvDevice.Device)
				importedCount++
			} else {
				errorCount++
			}
		}

		message := fmt.Sprintf("Import completed!\n\nImported: %d devices\nSkipped (errors): %d devices",
			importedCount, errorCount)

		dialog.ShowInformation("Import Complete", message, parent)
	})
	importButton.Disable()

	// Instructions
	instructions := widget.NewLabel(`CSV Format Expected:
IP,Username,Password,Port,Service,Status

Example:
10.10.50.2,admin,tnkbrBa9ezTaB,6684,ssh,import
10.100.0.126,admin,tnkbrBa9ezTaB,6684,ssh,import

Notes:
- IP address is required
- Username and password are required for SSH connections
- Port can be any number (22 for SSH is recommended)
- Service should be "ssh" for SSH devices
- Status column is ignored during import`)
	instructions.Wrapping = fyne.TextWrapWord

	// Create scroll container for the preview table
	previewScroll := container.NewScroll(previewTable)
	previewScroll.SetMinSize(fyne.NewSize(600, 300))

	// Layout
	content := container.NewVBox(
		widget.NewLabel("Import Devices from CSV"),
		widget.NewSeparator(),
		instructions,
		widget.NewSeparator(),
		filePickerBtn,
		fileLabel,
		widget.NewLabel("Preview:"),
		previewScroll,
		widget.NewSeparator(),
		container.NewHBox(
			importButton,
			widget.NewButton("Cancel", func() {
				// Dialog will close automatically
			}),
		),
	)

	// Create and show dialog
	d := dialog.NewCustom("Import CSV", "Close", content, parent)
	d.Resize(fyne.NewSize(700, 600))
	d.Show()
}

// parseCSVFile parses a CSV file and returns validated devices
func parseCSVFile(reader io.Reader) ([]CSVDevice, error) {
	csvReader := csv.NewReader(reader)
	csvReader.FieldsPerRecord = -1 // Allow variable number of fields

	var devices []CSVDevice
	rowNum := 0

	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error reading CSV at row %d: %v", rowNum+1, err)
		}

		rowNum++

		// Skip empty rows
		if len(record) == 0 || (len(record) == 1 && strings.TrimSpace(record[0]) == "") {
			continue
		}

		// Validate and create device
		csvDevice := validateCSVRecord(record, rowNum)
		devices = append(devices, csvDevice)
	}

	return devices, nil
}

// validateCSVRecord validates a CSV record and creates a CSVDevice
func validateCSVRecord(record []string, rowNum int) CSVDevice {
	csvDevice := CSVDevice{
		Row: rowNum,
	}

	// Check minimum fields
	if len(record) < 3 {
		csvDevice.Valid = false
		csvDevice.Error = "Not enough fields (need at least IP, Username, Password)"
		return csvDevice
	}

	// Extract fields
	ip := strings.TrimSpace(record[0])
	username := strings.TrimSpace(record[1])
	password := strings.TrimSpace(record[2])

	var port int = 22 // Default SSH port
	var service string = "ssh"
	var originalPort string = "22"

	// Use settings default port if available
	if settings.Current != nil {
		port = settings.Current.DefaultSSHPort
		originalPort = settings.Current.GetDefaultSSHPortString()
	}

	// Parse optional port field
	if len(record) > 3 && strings.TrimSpace(record[3]) != "" {
		originalPort = strings.TrimSpace(record[3])
		if p, err := strconv.Atoi(originalPort); err == nil {
			port = p
		}
	}

	// Parse optional service field
	if len(record) > 4 && strings.TrimSpace(record[4]) != "" {
		service = strings.ToLower(strings.TrimSpace(record[4]))
	}

	// Validate IP address
	if ip == "" {
		csvDevice.Valid = false
		csvDevice.Error = "Empty IP address"
		return csvDevice
	}

	// Basic IP validation (simple check)
	ipParts := strings.Split(ip, ".")
	if len(ipParts) != 4 {
		csvDevice.Valid = false
		csvDevice.Error = "Invalid IP address format"
		return csvDevice
	}

	for _, part := range ipParts {
		if num, err := strconv.Atoi(part); err != nil || num < 0 || num > 255 {
			csvDevice.Valid = false
			csvDevice.Error = "Invalid IP address range"
			return csvDevice
		}
	}

	// Validate username (use default if empty)
	if username == "" {
		if settings.Current != nil && settings.Current.DefaultSSHUsername != "" {
			username = settings.Current.DefaultSSHUsername
		} else {
			csvDevice.Valid = false
			csvDevice.Error = "Empty username and no default set"
			return csvDevice
		}
	}

	// Validate password (use default if empty)
	if password == "" {
		if settings.Current != nil && settings.Current.DefaultSSHPassword != "" {
			password = settings.Current.DefaultSSHPassword
		} else {
			csvDevice.Valid = false
			csvDevice.Error = "Empty password and no default set"
			return csvDevice
		}
	}

	// Validate port
	if port <= 0 || port > 65535 {
		csvDevice.Valid = false
		csvDevice.Error = fmt.Sprintf("Invalid port number: %d", port)
		return csvDevice
	}

	// Create device
	device := scanner.Device{
		IP:       ip,
		Hostname: ip, // Use IP as hostname initially
		Username: username,
		Password: password,
		SSHPort:  port,
		Status:   "Imported",
	}

	// Set port flags based on service
	if service == "ssh" {
		device.Port22 = true // Mark as SSH capable regardless of actual port
	}
	if service == "telnet" {
		device.Port23 = true
	}

	csvDevice.Device = device
	csvDevice.Valid = true
	csvDevice.Error = "OK"
	csvDevice.OriginalPort = originalPort

	return csvDevice
}

// updatePreviewTable updates the preview table with new data
func updatePreviewTable(table *widget.Table, devices []CSVDevice) {
	if table != nil {
		table.Refresh()
	}
}
