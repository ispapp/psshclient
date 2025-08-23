package main

import (
	"fmt"
	"time"

	"ispappclient/pkg/pssh"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func main() {
	// Create Fyne app
	myApp := app.New()
	myWindow := myApp.NewWindow("PSSH Demo")
	myWindow.Resize(fyne.NewSize(600, 400))

	// Demo device list (replace with your actual device scanner results)
	devices := []string{
		"192.168.1.1",
		"192.168.1.2",
		"192.168.1.3",
		"10.0.0.1",
	}

	// Create device list widget
	deviceList := widget.NewList(
		func() int { return len(devices) },
		func() fyne.CanvasObject {
			return widget.NewLabel("template")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(devices[i])
		},
	)

	// Track selected devices
	var selectedDevices []string
	deviceList.OnSelected = func(id widget.ListItemID) {
		selectedDevices = []string{devices[id]}
	}

	// Create buttons for SSH operations
	singleWindowBtn := widget.NewButton("SSH (Separate Windows)", func() {
		if len(selectedDevices) == 0 {
			dialog.ShowInformation("No Selection", "Please select a device first", myWindow)
			return
		}
		openSSHTerminals(selectedDevices, false, myWindow, myApp)
	})

	tabbedWindowBtn := widget.NewButton("SSH (Tabbed Window)", func() {
		if len(selectedDevices) == 0 {
			dialog.ShowInformation("No Selection", "Please select a device first", myWindow)
			return
		}
		openSSHTerminals(selectedDevices, true, myWindow, myApp)
	})

	multiSelectBtn := widget.NewButton("Select All for SSH", func() {
		selectedDevices = make([]string, len(devices))
		copy(selectedDevices, devices)
		openSSHTerminals(selectedDevices, true, myWindow, myApp)
	})

	// Test connection button
	testConnBtn := widget.NewButton("Test SSH Connection", func() {
		if len(selectedDevices) == 0 {
			dialog.ShowInformation("No Selection", "Please select a device first", myWindow)
			return
		}
		testConnections(selectedDevices, myWindow)
	})

	// Create layout
	buttons := container.NewVBox(
		widget.NewLabel("SSH Operations:"),
		singleWindowBtn,
		tabbedWindowBtn,
		multiSelectBtn,
		widget.NewSeparator(),
		testConnBtn,
	)

	content := container.NewHSplit(
		container.NewBorder(
			widget.NewLabel("Available Devices:"),
			nil, nil, nil,
			deviceList,
		),
		buttons,
	)

	myWindow.SetContent(content)
	myWindow.ShowAndRun()
}

func openSSHTerminals(devices []string, useTabs bool, parent fyne.Window, app fyne.App) {
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
			dialog.ShowInformation("No Connections", "No successful SSH connections were established", parent)
			return
		}

		// Open terminals
		err = pssh.OpenMultipleTerminals(connections, app)
		if err != nil {
			dialog.ShowError(err, parent)
		}
	})
}

func testConnections(devices []string, parent fyne.Window) {
	progress := widget.NewProgressBar()
	progress.Max = float64(len(devices))

	statusLabel := widget.NewLabel("Testing connections...")

	content := container.NewVBox(
		widget.NewLabel(fmt.Sprintf("Testing %d devices...", len(devices))),
		progress,
		statusLabel,
	)

	d := dialog.NewCustom("Testing Connections", "Close", content, parent)
	d.Resize(fyne.NewSize(400, 150))
	d.Show()

	go func() {
		var results []string
		var resultText string
		for i, device := range devices {
			statusLabel.SetText(fmt.Sprintf("Testing %s...", device))

			err := pssh.TestConnection(device, 22, 3*time.Second)
			if err != nil {
				results = append(results, fmt.Sprintf("❌ %s: %v", device, err))
			} else {
				results = append(results, fmt.Sprintf("✅ %s: SSH port open", device))
			}

			progress.SetValue(float64(i + 1))
			time.Sleep(100 * time.Millisecond) // Small delay for UI update
		}

		// Show results
		for _, result := range results {
			resultText += result + "\n"
		}

		statusLabel.SetText("Test completed!")
		time.Sleep(1 * time.Second)

		// Close the progress dialog and show results in a new dialog
		d.Hide()

		// Show results in a new dialog
		resultText = ""
		for _, result := range results {
			resultText += result + "\n"
		}

		resultLabel := widget.NewRichTextFromMarkdown("## Connection Test Results\n\n```\n" + resultText + "```")
		resultScroll := container.NewScroll(resultLabel)
		resultScroll.SetMinSize(fyne.NewSize(500, 300))

		resultDialog := dialog.NewCustom("Connection Test Results", "Close", resultScroll, parent)
		resultDialog.Resize(fyne.NewSize(600, 400))
		resultDialog.Show()
	}()
}
