package dialogs

import (
	"context"
	"fmt"
	"github.com/ispapp/psshclient/internal/data"
	"github.com/ispapp/psshclient/internal/scanner"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// ShowSubnetScanDialog shows a dialog to input subnet and start scanning
func ShowSubnetScanDialog(parent fyne.Window) {
	// Create input fields
	subnetEntry := widget.NewEntry()
	subnetEntry.SetText("192.168.1.0/24") // Default subnet
	subnetEntry.SetPlaceHolder("Enter subnet (e.g., 192.168.1.0/24)")

	// Create form
	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Subnet:", Widget: subnetEntry},
		},
	}

	// Create dialog
	d := dialog.NewCustomConfirm("Scan Subnet", "Start Scan", "Cancel",
		container.NewVBox(
			widget.NewLabel("Enter the subnet to scan for devices with SSH/Telnet:"),
			form,
		), func(confirmed bool) {
			if confirmed && subnetEntry.Text != "" {
				StartSubnetScan(subnetEntry.Text, parent)
			}
		}, parent)
	d.Resize(fyne.NewSize(400, 200))
	d.Show()
}

// StartSubnetScan starts the subnet scanning process
func StartSubnetScan(subnet string, parent fyne.Window) {
	// Clear previous results and set scanning state in main thread
	fyne.Do(func() {
		data.ClearDevices()
		data.SetScanning(true)
		data.SetScanProgress("Starting scan...")
	})

	// Create progress dialog
	progressBar := widget.NewProgressBarInfinite()
	progressLabel := widget.NewLabel("Starting scan...")

	// Bind progress label to global progress binding
	progressLabel.Bind(data.ScanProgress)

	progressContent := container.NewVBox(
		widget.NewLabel("Scanning subnet: "+subnet),
		progressLabel,
		progressBar,
	)

	progressDialog := dialog.NewCustom("Scanning...", "Cancel", progressContent, parent)
	progressDialog.Resize(fyne.NewSize(400, 150))

	// Create context for cancellation
	ctx, cancel := context.WithCancel(context.Background())

	// Handle cancel button
	progressDialog.SetOnClosed(func() {
		cancel()
		fyne.Do(func() {
			data.SetScanning(false)
		})
	})

	progressDialog.Show()

	// Start scanning in goroutine
	go func() {
		defer func() {
			fyne.Do(func() {
				data.SetScanning(false)
				if progressDialog != nil {
					progressDialog.Hide()
				}
			})
		}()

		// Progress callback function
		progressCallback := func(message string) {
			// Use fyne.Do to ensure UI updates happen in main thread
			fyne.Do(func() {
				data.SetScanProgress(message)
			})
		}

		// Perform the scan
		devices, err := scanner.ScanSubnet(ctx, subnet, progressCallback)

		// Check if scan was cancelled
		select {
		case <-ctx.Done():
			fyne.Do(func() {
				data.SetScanProgress("Scan cancelled")
			})
			return
		default:
		}

		if err != nil {
			fyne.Do(func() {
				data.SetScanProgress("Scan failed: " + err.Error())
				dialog.ShowError(err, parent)
			})
			return
		}

		// Add devices to global list
		for _, device := range devices {
			data.AddDevice(device)
		}

		fyne.Do(func() {
			data.SetScanProgress("Scan completed successfully")
		})

		// Show completion dialog
		go func() {
			time.Sleep(1 * time.Second) // Show completion message briefly
			fyne.Do(func() {
				dialog.ShowInformation("Scan Complete",
					fmt.Sprintf("Found %d devices with SSH/Telnet ports open", len(devices)),
					parent)
			})
		}()
	}()
}

// ShowFastScanDialog shows a dialog to start a fast scan of the local network
func ShowFastScanDialog(parent fyne.Window) {
	content := container.NewVBox(
		widget.NewLabel("This will perform a fast scan of your local network."),
		widget.NewLabel("It will automatically detect your network range and scan for devices with SSH/Telnet ports."),
	)

	d := dialog.NewCustomConfirm("Fast Scan", "Start", "Cancel", content, func(confirmed bool) {
		if confirmed {
			StartFastScan(parent)
		}
	}, parent)

	d.Resize(fyne.NewSize(400, 150))
	d.Show()
}

// StartFastScan starts a fast scan of the local network
func StartFastScan(parent fyne.Window) {
	// For now, we'll default to a common local network range
	// In a more advanced implementation, we could auto-detect the local network
	defaultSubnet := "192.168.1.0/24"
	StartSubnetScan(defaultSubnet, parent)
}
