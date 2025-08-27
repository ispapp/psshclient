package dialogs

import (
	"encoding/csv"
	"fmt"
	"ispappclient/internal/data"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
)

// ShowCSVExportDialog shows a dialog to export devices to a CSV file
func ShowCSVExportDialog(parent fyne.Window) {
	fileDialog := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil {
			dialog.ShowError(err, parent)
			return
		}
		if writer == nil {
			return
		}
		defer writer.Close()

		// Get all devices from the list
		devices := data.GetDevices()
		if len(devices) == 0 {
			dialog.ShowInformation("Export Skipped", "There are no devices to export.", parent)
			return
		}

		// Create a CSV writer
		csvWriter := csv.NewWriter(writer)
		defer csvWriter.Flush()

		// Write header row
		headers := []string{"IP Address", "Hostname", "SSH Port", "Username", "Password", "Status"}
		if err := csvWriter.Write(headers); err != nil {
			dialog.ShowError(fmt.Errorf("failed to write CSV header: %v", err), parent)
			return
		}

		// Write device data
		for _, device := range devices {
			record := []string{
				device.IP,
				device.Hostname,
				strconv.Itoa(device.SSHPort),
				device.Username,
				device.Password,
				device.Status,
			}
			if err := csvWriter.Write(record); err != nil {
				dialog.ShowError(fmt.Errorf("failed to write device record: %v", err), parent)
				return
			}
		}

		dialog.ShowInformation("Export Successful", fmt.Sprintf("Exported %d devices to %s", len(devices), writer.URI().Name()), parent)

	}, parent)

	fileDialog.SetFileName("devices.csv")
	fileDialog.SetFilter(storage.NewExtensionFileFilter([]string{".csv"}))
	fileDialog.Show()
}
