package data

import (
	"fmt"
	"ispappclient/internal/database"
	"ispappclient/internal/scanner"
	"ispappclient/internal/settings"
	"log"
	"os"
	"path/filepath"
	"time"

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

	// Database instance
	DB *database.DB
)

// Initialize all global bindings and database
func Init() {
	fmt.Printf("Initializing data bindings...\n")

	DeviceList = binding.NewUntypedList()
	ScanProgress = binding.NewString()
	IsScanning = binding.NewBool()

	fmt.Printf("Data bindings initialized\n")

	// Initialize database in the background to avoid blocking UI
	go InitDatabase()

	fmt.Printf("Data initialization completed\n")
}

// InitDatabase initializes the database after the main app has started
func InitDatabase() {
	if DB != nil {
		fmt.Printf("Database already initialized\n")
		return
	}

	fmt.Printf("Starting database initialization...\n")

	var err error
	var dbPath string

	// Use settings database path if available, otherwise use default
	if settings.Current != nil {
		dbPath = settings.Current.DatabasePath
	} else {
		homeDir, _ := os.UserHomeDir()
		dbPath = filepath.Join(homeDir, ".ispappclient", "devices.db")
	}

	DB, err = database.New(dbPath)
	if err != nil {
		log.Printf("Failed to initialize database: %v", err)
		fmt.Printf("Database initialization failed, continuing without database\n")
		return
	}

	fmt.Printf("Database initialized successfully\n")

	// Load devices from database on startup
	LoadDevicesFromDB()

	// Clean up old devices (using settings or default 30 days)
	go func() {
		cleanupDuration := 30 * 24 * time.Hour
		if settings.Current != nil {
			cleanupDuration = settings.Current.GetCleanupDuration()
		}

		if err := DB.DeleteOldDevices(cleanupDuration); err != nil {
			log.Printf("Failed to clean up old devices: %v", err)
		}
	}()
}

// AddDevice adds a new device to the global device list and saves to database
func AddDevice(device scanner.Device) {
	DeviceList.Append(device)

	// Save to database if available
	if DB != nil {
		if err := DB.SaveDevice(device); err != nil {
			log.Printf("Failed to save device %s to database: %v", device.IP, err)
		}
	}
}

// ClearDevices removes all devices from the list (does not clear database)
func ClearDevices() {
	DeviceList.Set([]interface{}{})
}

// ClearDevicesAndDB removes all devices from list and database
func ClearDevicesAndDB() {
	DeviceList.Set([]interface{}{})

	// Also clear from database if available
	if DB != nil {
		// Delete all devices older than 0 seconds (all devices)
		if err := DB.DeleteOldDevices(0); err != nil {
			log.Printf("Failed to clear devices from database: %v", err)
		}
	}
}

// UpdateDevice updates a device both in memory and database
func UpdateDevice(index int, device scanner.Device) {
	if index >= 0 && index < DeviceList.Length() {
		DeviceList.SetValue(index, device)

		// Save to database if available
		if DB != nil {
			if err := DB.SaveDevice(device); err != nil {
				log.Printf("Failed to update device %s in database: %v", device.IP, err)
			}
		}
	}
}

// UpdateDeviceCredentials updates device credentials both in memory and database
func UpdateDeviceCredentials(index int, username, password string) {
	if index >= 0 && index < DeviceList.Length() {
		if deviceObj, err := DeviceList.GetValue(index); err == nil {
			if device, ok := deviceObj.(scanner.Device); ok {
				device.Username = username
				device.Password = password
				DeviceList.SetValue(index, device)

				// Update in database if available
				if DB != nil {
					if err := DB.UpdateDeviceCredentials(device.IP, username, password); err != nil {
						log.Printf("Failed to update credentials for %s in database: %v", device.IP, err)
					}
				}
			}
		}
	}
}

// UpdateDeviceConnection updates device connection status both in memory and database
func UpdateDeviceConnection(index int, connected bool) {
	if index >= 0 && index < DeviceList.Length() {
		if deviceObj, err := DeviceList.GetValue(index); err == nil {
			if device, ok := deviceObj.(scanner.Device); ok {
				device.Connected = connected
				if connected {
					device.Status = "Connected"
				} else {
					device.Status = "Disconnected"
				}
				DeviceList.SetValue(index, device)

				// Update in database if available
				if DB != nil {
					if err := DB.UpdateDeviceConnection(device.IP, connected); err != nil {
						log.Printf("Failed to update connection status for %s in database: %v", device.IP, err)
					}
				}
			}
		}
	}
}

// SaveDevicesToDB saves all current devices to database
func SaveDevicesToDB() {
	if DB == nil {
		return
	}

	devices := GetDevices()
	if err := DB.SaveDevices(devices); err != nil {
		log.Printf("Failed to save devices to database: %v", err)
	} else {
		fmt.Printf("Saved %d devices to database\n", len(devices))
	}
}

// LoadDevicesFromDB loads devices from database and adds them to the device list
func LoadDevicesFromDB() {
	if DB == nil {
		return
	}

	devices, err := DB.LoadDevices()
	if err != nil {
		log.Printf("Failed to load devices from database: %v", err)
		return
	}

	// Clear current list and add loaded devices
	var deviceInterfaces []interface{}
	for _, device := range devices {
		deviceInterfaces = append(deviceInterfaces, device)
	}

	DeviceList.Set(deviceInterfaces)
	fmt.Printf("Loaded %d devices from database\n", len(devices))
}

// LoadRecentDevicesFromDB loads only recently seen devices from database
func LoadRecentDevicesFromDB(since time.Duration) {
	if DB == nil {
		return
	}

	devices, err := DB.LoadRecentDevices(since)
	if err != nil {
		log.Printf("Failed to load recent devices from database: %v", err)
		return
	}

	// Clear current list and add loaded devices
	var deviceInterfaces []interface{}
	for _, device := range devices {
		deviceInterfaces = append(deviceInterfaces, device)
	}

	DeviceList.Set(deviceInterfaces)
	fmt.Printf("Loaded %d recent devices from database\n", len(devices))
}

// GetDeviceByIP finds a device by IP address
func GetDeviceByIP(ip string) (scanner.Device, int, bool) {
	length := DeviceList.Length()
	for i := 0; i < length; i++ {
		if item, err := DeviceList.GetValue(i); err == nil {
			if device, ok := item.(scanner.Device); ok && device.IP == ip {
				return device, i, true
			}
		}
	}
	return scanner.Device{}, -1, false
}

// SaveSetting saves an application setting to database
func SaveSetting(key, value string) {
	if DB != nil {
		if err := DB.SaveSetting(key, value); err != nil {
			log.Printf("Failed to save setting %s: %v", key, err)
		}
	}
}

// LoadSetting loads an application setting from database
func LoadSetting(key string) string {
	if DB == nil {
		return ""
	}

	value, err := DB.LoadSetting(key)
	if err != nil {
		log.Printf("Failed to load setting %s: %v", key, err)
		return ""
	}

	return value
}

// Close closes the database connection
func Close() {
	if DB != nil {
		if err := DB.Close(); err != nil {
			log.Printf("Failed to close database: %v", err)
		}
	}
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
