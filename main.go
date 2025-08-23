package main

import (
	"ispappclient/internal/ui"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

var GlobalApp fyne.App

func main() {
	// Create new Fyne application
	GlobalApp = app.NewWithID("co.ispapp.psshclient")
	// Create main UI with tabbed interface
	mainUI := ui.NewMainUI(GlobalApp)

	// Show and run the application
	mainUI.ShowAndRun()
}
