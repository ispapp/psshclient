package main

import (
	"github.com/ispapp/psshclient/internal/ui"
	"github.com/ispapp/psshclient/internal/ui/theme"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

var GlobalApp fyne.App

func main() {
	// Create new Fyne application
	GlobalApp = app.NewWithID("co.ispapp.psshclient")
	// Create main UI with tabbed interface
	_theme := theme.AppTheme{}
	_theme.ApplyTheme(GlobalApp)
	mainUI := ui.NewMainUI(GlobalApp)

	// Show and run the application
	mainUI.ShowAndRun()
}
