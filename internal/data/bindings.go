package data

import (
	"ispappclient/internal/scanner"

	"fyne.io/fyne/v2/data/binding"
)

// Global data bindings
var (
	// DeviceList holds the list of discovered devices
	DeviceList binding.UntypedList

	// ScanProgress holds the current scan progress message
	ScanProgress binding.String

	// IsScanning indicates if a scan is currently running
	IsScanning binding.Bool
)

// Initialize all global bindings
func Init() {
	DeviceList = binding.NewUntypedList()
	ScanProgress = binding.NewString()
	IsScanning = binding.NewBool()
}

// AddDevice adds a new device to the global device list
func AddDevice(device scanner.Device) {
	DeviceList.Append(device)
}

// ClearDevices removes all devices from the list
func ClearDevices() {
	DeviceList.Set([]interface{}{})
}

// GetDevices returns all devices as a slice
func GetDevices() []scanner.Device {
	var devices []scanner.Device
	length := DeviceList.Length()

	for i := 0; i < length; i++ {
		if item, err := DeviceList.GetValue(i); err == nil {
			if device, ok := item.(scanner.Device); ok {
				devices = append(devices, device)
			}
		}
	}

	return devices
}

// SetScanProgress updates the scan progress message
func SetScanProgress(message string) {
	ScanProgress.Set(message)
}

// SetScanning updates the scanning status
func SetScanning(scanning bool) {
	IsScanning.Set(scanning)
}
