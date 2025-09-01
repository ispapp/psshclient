package ui

import (
	"log"

	"github.com/ispapp/psshclient/internal/data"
	"github.com/ispapp/psshclient/internal/dialogs"
	"github.com/ispapp/psshclient/internal/settings"
	"github.com/ispapp/psshclient/internal/widgets"
	"github.com/ispapp/psshclient/internal/windows"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func NewMainUI(app fyne.App) fyne.Window {
	// Initialize settings first
	if err := settings.Initialize(); err != nil {
		log.Printf("Failed to initialize settings: %v", err)
		// Continue with defaults
	}

	// Initialize global data bindings
	data.Init()

	MainWindow := app.NewWindow("Agent Deployer")
	actionLabel := widget.NewLabel("Action will appear here")

	// Create devices table with SSH functionality
	devicesTable := widgets.CreateDevicesTableWithWindow(MainWindow, app)

	// Create settings tab
	settingsTab := widgets.CreateSettingsTab(MainWindow)

	tabs := container.NewAppTabs(
		container.NewTabItem("Devices", devicesTable),
		container.NewTabItem("Settings", settingsTab),
	)
	winmanager := windows.NewWindowManager(app)
	winmanager.SetMainWindow(MainWindow)
	// Create scan menu
	scanMenu := fyne.NewMenu("Scan",
		fyne.NewMenuItem("Start Fast Scan", func() {
			actionLabel.SetText("Selected: Start Fast Scan")
			dialogs.ShowFastScanDialog(MainWindow)
		}),
		fyne.NewMenuItem("Scan Subnet", func() {
			actionLabel.SetText("Selected: Scan Subnet")
			dialogs.ShowSubnetScanDialog(MainWindow)
		}),
		fyne.NewMenuItem("Neighbor Discovery", func() {
			actionLabel.SetText("Selected: Neighbor Discovery")
			dialogs.ShowNeighborDiscoveryDialog(winmanager)
		}),
		fyne.NewMenuItem("Start Syn Scan", func() {
			actionLabel.SetText("Selected: Start Syn Scan")
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Exit", func() {
			app.Quit()
		}),
	) // Create File menu
	fileMenu := fyne.NewMenu("File",
		fyne.NewMenuItem("Import CSV", func() {
			dialogs.ShowCSVImportDialog(MainWindow)
		}),
		fyne.NewMenuItem("Export CSV", func() {
			dialogs.ShowCSVExportDialog(MainWindow)
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
		scanMenu,
		fileMenu,
		helpMenu,
	)
	MainWindow.SetMainMenu(mainMenu)
	//tabs.Append(container.NewTabItemWithIcon("Home", theme.HomeIcon(), widget.NewLabel("Home tab")))

	tabs.SetTabLocation(container.TabLocation(container.ScrollBoth))

	MainWindow.SetContent(tabs)
	MainWindow.Resize(fyne.NewSize(1000, 800))
	MainWindow.SetPadded(true)
	// Set up cleanup when the main window closes
	MainWindow.SetOnClosed(func() {
		// Save current devices to database before closing
		data.SaveDevicesToDB()
		// Close database connection
		data.Close()
	})

	return MainWindow
}
