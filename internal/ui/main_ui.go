package ui

import (
	"ispappclient/internal/data"
	"ispappclient/internal/dialogs"
	"ispappclient/internal/widgets"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

var (
	proto    = "tcp"
	fastscan = true
	syn      = false
)

func NewMainUI(app fyne.App) fyne.Window {
	// Initialize global data bindings
	data.Init()

	MainWindow := app.NewWindow("Agent Deployer")
	actionLabel := widget.NewLabel("Action will appear here")

	// Create devices table with SSH functionality
	devicesTable := widgets.CreateDevicesTableWithWindow(MainWindow, app)

	tabs := container.NewAppTabs(
		container.NewTabItem("Devices", devicesTable),
		container.NewTabItem("Settings", container.NewVBox(
			widget.NewLabel("Device Actions"),
			actionLabel,
			widget.NewLabel("# SETTINGS FORM (EDIT/SAVE) #"),
		)),
	)
	// Create file menu
	fileMenu := fyne.NewMenu("Scan",
		fyne.NewMenuItem("Start Fast Scan", func() {
			actionLabel.SetText("Selected: Start Fast Scan")
			dialogs.ShowFastScanDialog(MainWindow)
		}),
		fyne.NewMenuItem("Scan Subnet", func() {
			actionLabel.SetText("Selected: Scan Subnet")
			dialogs.ShowSubnetScanDialog(MainWindow)
		}),
		fyne.NewMenuItem("Start Syn Scan", func() {
			actionLabel.SetText("Selected: Start Syn Scan")
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Exit", func() {
			app.Quit()
		}),
	)

	// Create edit menu
	editMenu := fyne.NewMenu("Edit",
		fyne.NewMenuItem("Cut", func() {
			actionLabel.SetText("Selected: Cut")
		}),
		fyne.NewMenuItem("Copy", func() {
			actionLabel.SetText("Selected: Copy")
		}),
		fyne.NewMenuItem("Paste", func() {
			actionLabel.SetText("Selected: Paste")
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Select All SSH Devices", func() {
			actionLabel.SetText("Selected: Select All SSH Devices")
			dialog.ShowInformation("SSH Selection",
				"Use the 'Select All SSH' button in the Devices tab to select all devices with SSH support",
				MainWindow)
		}),
		fyne.NewMenuItem("Open SSH Terminal", func() {
			actionLabel.SetText("Selected: Open SSH Terminal")
			dialog.ShowInformation("SSH Terminal",
				"Select devices in the Devices tab and use the SSH buttons to open terminals",
				MainWindow)
		}),
	)

	// Create help menu
	helpMenu := fyne.NewMenu("Help",
		fyne.NewMenuItem("About", func() {
			dialog.ShowInformation("About", "visit ispapp.co for more information", MainWindow)
		}),
	)

	// Create the main menu
	mainMenu := fyne.NewMainMenu(
		fileMenu,
		editMenu,
		helpMenu,
	)
	MainWindow.SetMainMenu(mainMenu)
	//tabs.Append(container.NewTabItemWithIcon("Home", theme.HomeIcon(), widget.NewLabel("Home tab")))

	tabs.SetTabLocation(container.TabLocation(container.ScrollBoth))

	MainWindow.SetContent(tabs)
	MainWindow.Resize(fyne.NewSize(800, 600))
	return MainWindow
}
