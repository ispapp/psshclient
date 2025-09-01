package settings

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// AppSettings holds all application settings
type AppSettings struct {
	// Network Settings
	DefaultSSHPort    int `json:"default_ssh_port"`
	DefaultTelnetPort int `json:"default_telnet_port"`
	ConnectionTimeout int `json:"connection_timeout_seconds"`

	// SSH Default Credentials
	DefaultSSHUsername string `json:"default_ssh_username"`
	DefaultSSHPassword string `json:"default_ssh_password"`

	// Database Settings
	DatabasePath    string `json:"database_path"`
	AutoSaveDevices bool   `json:"auto_save_devices"`
	CleanupOldDays  int    `json:"cleanup_old_days"`

	// Terminal Settings
	TerminalRows     int    `json:"terminal_rows"`
	TerminalCols     int    `json:"terminal_cols"`
	TerminalFont     string `json:"terminal_font"`
	TerminalFontSize int    `json:"terminal_font_size"`

	// Scanning Settings
	ScanTimeout        int   `json:"scan_timeout_seconds"`
	MaxConcurrentScans int   `json:"max_concurrent_scans"`
	DefaultScanPorts   []int `json:"default_scan_ports"`

	// UI Settings
	WindowWidth  int    `json:"window_width"`
	WindowHeight int    `json:"window_height"`
	Theme        string `json:"theme"`
}

// DefaultSettings returns the default application settings
func DefaultSettings() *AppSettings {
	homeDir, _ := os.UserHomeDir()
	defaultDBPath := filepath.Join(homeDir, ".ispappclient", "devices.db")

	return &AppSettings{
		// Network Settings
		DefaultSSHPort:    22,
		DefaultTelnetPort: 23,
		ConnectionTimeout: 30,

		// SSH Default Credentials
		DefaultSSHUsername: "admin",
		DefaultSSHPassword: "",

		// Database Settings
		DatabasePath:    defaultDBPath,
		AutoSaveDevices: true,
		CleanupOldDays:  30,

		// Terminal Settings
		TerminalRows:     24,
		TerminalCols:     80,
		TerminalFont:     "monospace",
		TerminalFontSize: 12,

		// Scanning Settings
		ScanTimeout:        10,
		MaxConcurrentScans: 50,
		DefaultScanPorts:   []int{22, 23, 80, 443},

		// UI Settings
		WindowWidth:  800,
		WindowHeight: 600,
		Theme:        "default",
	}
}

// Global settings instance
var Current *AppSettings

// GetCurrent returns the current settings instance
func RefreshCurrent() *AppSettings {
	if Current == nil {
		err := Load()
		if err != nil {
			Current = DefaultSettings()
		}
	}
	return Current
}

// Initialize loads settings from file or creates default settings
func Initialize() error {
	Current = DefaultSettings()

	settingsPath := getSettingsPath()
	if _, err := os.Stat(settingsPath); os.IsNotExist(err) {
		// Settings file doesn't exist, create it with defaults
		return Save()
	}

	// Load existing settings
	return Load()
}

// Load reads settings from the settings file
func Load() error {
	settingsPath := getSettingsPath()

	data, err := os.ReadFile(settingsPath)
	if err != nil {
		return fmt.Errorf("failed to read settings file: %v", err)
	}

	err = json.Unmarshal(data, Current)
	if err != nil {
		return fmt.Errorf("failed to parse settings file: %v", err)
	}

	return nil
}

// Save writes current settings to the settings file
func Save() error {
	settingsPath := getSettingsPath()

	// Ensure directory exists
	settingsDir := filepath.Dir(settingsPath)
	if err := os.MkdirAll(settingsDir, 0755); err != nil {
		return fmt.Errorf("failed to create settings directory: %v", err)
	}

	data, err := json.MarshalIndent(Current, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %v", err)
	}

	err = os.WriteFile(settingsPath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write settings file: %v", err)
	}

	return nil
}

// getSettingsPath returns the path to the settings file
func getSettingsPath() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".ispappclient", "settings.json")
}

// GetConnectionTimeout returns connection timeout as time.Duration
func (s *AppSettings) GetConnectionTimeout() time.Duration {
	return time.Duration(s.ConnectionTimeout) * time.Second
}

// GetScanTimeout returns scan timeout as time.Duration
func (s *AppSettings) GetScanTimeout() time.Duration {
	return time.Duration(s.ScanTimeout) * time.Second
}

// GetCleanupDuration returns cleanup duration as time.Duration
func (s *AppSettings) GetCleanupDuration() time.Duration {
	return time.Duration(s.CleanupOldDays) * 24 * time.Hour
}

// Validation functions
func (s *AppSettings) Validate() []string {
	var errors []string

	if s.DefaultSSHPort <= 0 || s.DefaultSSHPort > 65535 {
		errors = append(errors, "SSH port must be between 1 and 65535")
	}

	if s.DefaultTelnetPort <= 0 || s.DefaultTelnetPort > 65535 {
		errors = append(errors, "Telnet port must be between 1 and 65535")
	}

	if s.ConnectionTimeout <= 0 {
		errors = append(errors, "Connection timeout must be greater than 0")
	}

	if s.TerminalRows <= 0 {
		errors = append(errors, "Terminal rows must be greater than 0")
	}

	if s.TerminalCols <= 0 {
		errors = append(errors, "Terminal columns must be greater than 0")
	}

	if s.ScanTimeout <= 0 {
		errors = append(errors, "Scan timeout must be greater than 0")
	}

	if s.MaxConcurrentScans <= 0 {
		errors = append(errors, "Max concurrent scans must be greater than 0")
	}

	return errors
}

// Helper functions to convert settings to/from strings for UI
func (s *AppSettings) GetDefaultSSHPortString() string {
	return strconv.Itoa(s.DefaultSSHPort)
}

func (s *AppSettings) SetDefaultSSHPortString(value string) error {
	port, err := strconv.Atoi(value)
	if err != nil {
		return err
	}
	s.DefaultSSHPort = port
	return nil
}

func (s *AppSettings) GetDefaultTelnetPortString() string {
	return strconv.Itoa(s.DefaultTelnetPort)
}

func (s *AppSettings) SetDefaultTelnetPortString(value string) error {
	port, err := strconv.Atoi(value)
	if err != nil {
		return err
	}
	s.DefaultTelnetPort = port
	return nil
}

func (s *AppSettings) GetConnectionTimeoutString() string {
	return strconv.Itoa(s.ConnectionTimeout)
}

func (s *AppSettings) SetConnectionTimeoutString(value string) error {
	timeout, err := strconv.Atoi(value)
	if err != nil {
		return err
	}
	s.ConnectionTimeout = timeout
	return nil
}

func (s *AppSettings) GetTerminalRowsString() string {
	return strconv.Itoa(s.TerminalRows)
}

func (s *AppSettings) SetTerminalRowsString(value string) error {
	rows, err := strconv.Atoi(value)
	if err != nil {
		return err
	}
	s.TerminalRows = rows
	return nil
}

func (s *AppSettings) GetTerminalColsString() string {
	return strconv.Itoa(s.TerminalCols)
}

func (s *AppSettings) SetTerminalColsString(value string) error {
	cols, err := strconv.Atoi(value)
	if err != nil {
		return err
	}
	s.TerminalCols = cols
	return nil
}

func (s *AppSettings) GetTerminalFontSizeString() string {
	return strconv.Itoa(s.TerminalFontSize)
}

func (s *AppSettings) SetTerminalFontSizeString(value string) error {
	size, err := strconv.Atoi(value)
	if err != nil {
		return err
	}
	s.TerminalFontSize = size
	return nil
}

func (s *AppSettings) GetScanTimeoutString() string {
	return strconv.Itoa(s.ScanTimeout)
}

func (s *AppSettings) SetScanTimeoutString(value string) error {
	timeout, err := strconv.Atoi(value)
	if err != nil {
		return err
	}
	s.ScanTimeout = timeout
	return nil
}

func (s *AppSettings) GetMaxConcurrentScansString() string {
	return strconv.Itoa(s.MaxConcurrentScans)
}

func (s *AppSettings) SetMaxConcurrentScansString(value string) error {
	scans, err := strconv.Atoi(value)
	if err != nil {
		return err
	}
	s.MaxConcurrentScans = scans
	return nil
}

func (s *AppSettings) GetCleanupOldDaysString() string {
	return strconv.Itoa(s.CleanupOldDays)
}

func (s *AppSettings) SetCleanupOldDaysString(value string) error {
	days, err := strconv.Atoi(value)
	if err != nil {
		return err
	}
	s.CleanupOldDays = days
	return nil
}
