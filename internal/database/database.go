package database

import (
	"database/sql"
	"fmt"
	"github.com/ispapp/psshclient/internal/scanner"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	conn *sql.DB
}

// New creates a new database connection
func New(dbPath string) (*DB, error) {
	fmt.Printf("Starting database initialization...\n")
	fmt.Printf("Database path: %s\n", dbPath)

	// Create app data directory if it doesn't exist
	dbDir := filepath.Dir(dbPath)
	fmt.Printf("Database directory: %s\n", dbDir)

	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %v", err)
	}
	fmt.Printf("Database directory created/verified\n")

	// Open database connection
	fmt.Printf("Opening SQLite database...\n")
	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}
	fmt.Printf("Database connection opened\n")

	// Test the connection
	fmt.Printf("Testing database connection...\n")
	if err := conn.Ping(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}
	fmt.Printf("Database connection tested successfully\n")

	db := &DB{conn: conn}

	// Initialize database schema
	fmt.Printf("Initializing database schema...\n")
	if err := db.initSchema(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to initialize database schema: %v", err)
	}
	fmt.Printf("Database schema initialized\n")

	// Run migrations
	fmt.Printf("Running database migrations...\n")
	if err := db.runMigrations(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to run database migrations: %v", err)
	}
	fmt.Printf("Database migrations completed\n")

	fmt.Printf("Database initialization completed successfully\n")
	return db, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.conn.Close()
}

// initSchema creates the necessary tables if they don't exist
func (db *DB) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS devices (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		ip TEXT UNIQUE NOT NULL,
		hostname TEXT NOT NULL DEFAULT '',
		port22 BOOLEAN NOT NULL DEFAULT 0,
		port23 BOOLEAN NOT NULL DEFAULT 0,
		ssh_port INTEGER NOT NULL DEFAULT 22,
		status TEXT NOT NULL DEFAULT '',
		username TEXT NOT NULL DEFAULT '',
		password TEXT NOT NULL DEFAULT '',
		connected BOOLEAN NOT NULL DEFAULT 0,
		last_seen DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_devices_ip ON devices(ip);
	CREATE INDEX IF NOT EXISTS idx_devices_last_seen ON devices(last_seen);

	CREATE TABLE IF NOT EXISTS app_settings (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		key TEXT UNIQUE NOT NULL,
		value TEXT NOT NULL,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_app_settings_key ON app_settings(key);
	`

	_, err := db.conn.Exec(schema)
	return err
}

// SaveDevice saves or updates a device in the database
func (db *DB) SaveDevice(device scanner.Device) error {
	query := `
	INSERT INTO devices (ip, hostname, port22, port23, ssh_port, status, username, password, connected, last_seen, updated_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	ON CONFLICT(ip) DO UPDATE SET
		hostname = excluded.hostname,
		port22 = excluded.port22,
		port23 = excluded.port23,
		ssh_port = excluded.ssh_port,
		status = excluded.status,
		username = excluded.username,
		password = excluded.password,
		connected = excluded.connected,
		last_seen = CURRENT_TIMESTAMP,
		updated_at = CURRENT_TIMESTAMP
	`

	_, err := db.conn.Exec(query, device.IP, device.Hostname, device.Port22, device.Port23, device.SSHPort,
		device.Status, device.Username, device.Password, device.Connected)
	return err
}

// SaveDevices saves multiple devices in a transaction
func (db *DB) SaveDevices(devices []scanner.Device) error {
	tx, err := db.conn.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
	INSERT INTO devices (ip, hostname, port22, port23, ssh_port, status, username, password, connected, last_seen, updated_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	ON CONFLICT(ip) DO UPDATE SET
		hostname = excluded.hostname,
		port22 = excluded.port22,
		port23 = excluded.port23,
		ssh_port = excluded.ssh_port,
		status = excluded.status,
		username = excluded.username,
		password = excluded.password,
		connected = excluded.connected,
		last_seen = CURRENT_TIMESTAMP,
		updated_at = CURRENT_TIMESTAMP
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %v", err)
	}
	defer stmt.Close()

	for _, device := range devices {
		_, err := stmt.Exec(device.IP, device.Hostname, device.Port22, device.Port23, device.SSHPort,
			device.Status, device.Username, device.Password, device.Connected)
		if err != nil {
			return fmt.Errorf("failed to save device %s: %v", device.IP, err)
		}
	}

	return tx.Commit()
}

// LoadDevices loads all devices from the database
func (db *DB) LoadDevices() ([]scanner.Device, error) {
	query := `
	SELECT ip, hostname, port22, port23, ssh_port, status, username, password, connected
	FROM devices
	ORDER BY last_seen DESC, ip ASC
	`

	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query devices: %v", err)
	}
	defer rows.Close()

	var devices []scanner.Device
	for rows.Next() {
		var device scanner.Device
		err := rows.Scan(&device.IP, &device.Hostname, &device.Port22, &device.Port23, &device.SSHPort,
			&device.Status, &device.Username, &device.Password, &device.Connected)
		if err != nil {
			return nil, fmt.Errorf("failed to scan device row: %v", err)
		}
		devices = append(devices, device)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating device rows: %v", err)
	}

	return devices, nil
}

// LoadRecentDevices loads devices seen within the last specified duration
func (db *DB) LoadRecentDevices(since time.Duration) ([]scanner.Device, error) {
	query := `
	SELECT ip, hostname, port22, port23, ssh_port, status, username, password, connected
	FROM devices
	WHERE last_seen > datetime('now', '-' || ? || ' seconds')
	ORDER BY last_seen DESC, ip ASC
	`

	rows, err := db.conn.Query(query, int(since.Seconds()))
	if err != nil {
		return nil, fmt.Errorf("failed to query recent devices: %v", err)
	}
	defer rows.Close()

	var devices []scanner.Device
	for rows.Next() {
		var device scanner.Device
		err := rows.Scan(&device.IP, &device.Hostname, &device.Port22, &device.Port23, &device.SSHPort,
			&device.Status, &device.Username, &device.Password, &device.Connected)
		if err != nil {
			return nil, fmt.Errorf("failed to scan device row: %v", err)
		}
		devices = append(devices, device)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating device rows: %v", err)
	}

	return devices, nil
}

// DeleteOldDevices removes devices not seen for the specified duration
func (db *DB) DeleteOldDevices(olderThan time.Duration) error {
	query := `
	DELETE FROM devices
	WHERE last_seen < datetime('now', '-' || ? || ' seconds')
	`

	result, err := db.conn.Exec(query, int(olderThan.Seconds()))
	if err != nil {
		return fmt.Errorf("failed to delete old devices: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %v", err)
	}

	fmt.Printf("Deleted %d old devices\n", rowsAffected)
	return nil
}

// SaveSetting saves an application setting
func (db *DB) SaveSetting(key, value string) error {
	query := `
	INSERT INTO app_settings (key, value, updated_at)
	VALUES (?, ?, CURRENT_TIMESTAMP)
	ON CONFLICT(key) DO UPDATE SET
		value = excluded.value,
		updated_at = CURRENT_TIMESTAMP
	`

	_, err := db.conn.Exec(query, key, value)
	return err
}

// LoadSetting loads an application setting
func (db *DB) LoadSetting(key string) (string, error) {
	query := `SELECT value FROM app_settings WHERE key = ?`

	var value string
	err := db.conn.QueryRow(query, key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil // Setting doesn't exist, return empty string
	}
	if err != nil {
		return "", fmt.Errorf("failed to load setting %s: %v", key, err)
	}

	return value, nil
}

// GetDeviceCount returns the total number of devices in the database
func (db *DB) GetDeviceCount() (int, error) {
	query := `SELECT COUNT(*) FROM devices`

	var count int
	err := db.conn.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get device count: %v", err)
	}

	return count, nil
}

// GetSSHDeviceCount returns the number of devices with SSH enabled
func (db *DB) GetSSHDeviceCount() (int, error) {
	query := `SELECT COUNT(*) FROM devices WHERE port22 = 1`

	var count int
	err := db.conn.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get SSH device count: %v", err)
	}

	return count, nil
}

// UpdateDeviceConnection updates only the connection status of a device
func (db *DB) UpdateDeviceConnection(ip string, connected bool) error {
	query := `
	UPDATE devices 
	SET connected = ?, updated_at = CURRENT_TIMESTAMP
	WHERE ip = ?
	`

	result, err := db.conn.Exec(query, connected, ip)
	if err != nil {
		return fmt.Errorf("failed to update device connection: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("device with IP %s not found", ip)
	}

	return nil
}

// UpdateDeviceCredentials updates username and password for a device
func (db *DB) UpdateDeviceCredentials(ip, username, password string) error {
	query := `
	UPDATE devices 
	SET username = ?, password = ?, updated_at = CURRENT_TIMESTAMP
	WHERE ip = ?
	`

	result, err := db.conn.Exec(query, username, password, ip)
	if err != nil {
		return fmt.Errorf("failed to update device credentials: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("device with IP %s not found", ip)
	}

	return nil
}

// migrationVersion returns the current database schema version
func (db *DB) migrationVersion() (int, error) {
	var version int
	err := db.conn.QueryRow("PRAGMA user_version").Scan(&version)
	if err != nil {
		return 0, fmt.Errorf("failed to get database version: %v", err)
	}
	return version, nil
}

// setMigrationVersion sets the database schema version
func (db *DB) setMigrationVersion(version int) error {
	query := fmt.Sprintf("PRAGMA user_version = %d", version)
	_, err := db.conn.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to set database version: %v", err)
	}
	return nil
}

// runMigrations runs database migrations if needed
func (db *DB) runMigrations() error {
	currentVersion, err := db.migrationVersion()
	if err != nil {
		return err
	}

	// Current target version
	targetVersion := 1

	if currentVersion >= targetVersion {
		return nil // No migration needed
	}

	fmt.Printf("Running database migrations from version %d to %d\n", currentVersion, targetVersion)

	// Migration scripts
	migrations := []string{
		// Version 1: Initial schema (already created in initSchema)
		"",
	}

	for i := currentVersion; i < targetVersion; i++ {
		if i < len(migrations) && migrations[i] != "" {
			fmt.Printf("Running migration %d\n", i+1)
			_, err := db.conn.Exec(migrations[i])
			if err != nil {
				return fmt.Errorf("migration %d failed: %v", i+1, err)
			}
		}
	}

	// Update version
	return db.setMigrationVersion(targetVersion)
}
