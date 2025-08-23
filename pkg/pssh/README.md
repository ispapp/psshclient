# PSSH - Parallel SSH Client Package

A Go package for managing multiple SSH connections with Fyne GUI terminal widgets.

## Features

- **Parallel SSH Connections**: Connect to multiple hosts simultaneously
- **Terminal Widgets**: Integrated Fyne terminal widgets for SSH sessions
- **Connection Management**: Centralized SSH connection management
- **Progress Tracking**: Visual progress for bulk connections
- **Multiple Terminal Modes**: 
  - Tabbed terminals in a single window
  - Separate terminal windows
- **Credential Management**: Support for password and SSH key authentication
- **Dynamic Resizing**: Automatic terminal resize support
- **Error Handling**: Comprehensive error handling and reporting

## Installation

The package requires the following dependencies:

```bash
go get golang.org/x/crypto/ssh
go get github.com/fyne-io/terminal
go get fyne.io/fyne/v2
```

## Usage

### Basic SSH Connection

```go
import "ispappclient/pkg/pssh"

// Create a single connection
config := pssh.NewConnectionConfig("192.168.1.1", 22, "admin", "password")
conn := &pssh.SSHConnection{Config: config}
err := conn.Connect()
if err != nil {
    log.Fatal(err)
}
defer conn.Close()
```

### Multiple Parallel Connections

```go
// Create SSH manager
manager := pssh.NewSSHManager()

// Create multiple connection configs
configs := []pssh.ConnectionConfig{
    pssh.NewConnectionConfig("192.168.1.1", 22, "admin", "password"),
    pssh.NewConnectionConfig("192.168.1.2", 22, "admin", "password"),
    pssh.NewConnectionConfig("192.168.1.3", 22, "admin", "password"),
}

// Connect in parallel
resultChan := manager.ConnectMultiple(configs)

// Process results
for result := range resultChan {
    if result.Error != nil {
        fmt.Printf("Failed to connect to %s: %v\n", result.Host, result.Error)
        continue
    }
    fmt.Printf("Successfully connected to %s\n", result.Host)
}

// Close all connections when done
defer manager.CloseAll()
```

### Terminal Widgets

```go
// Create terminal manager
termManager := pssh.NewTerminalManager()

// Create terminal for a single connection
termWidget, err := termManager.CreateTerminalWidget(conn, "SSH Terminal")
if err != nil {
    log.Fatal(err)
}

// Show terminal (blocking)
termWidget.ShowTerminal()

// Or for multiple terminals in tabs
connections := []*pssh.SSHConnection{conn1, conn2, conn3}
err = termManager.MultiTerminalWindow(connections, "Multiple SSH Terminals")
if err != nil {
    log.Fatal(err)
}
```

### UI Integration

```go
// Show credentials dialog
pssh.ShowCredentialsDialog(parentWindow, func(creds pssh.SSHCredentials, confirmed bool) {
    if !confirmed {
        return
    }
    
    // Connect to multiple devices
    deviceIPs := []string{"192.168.1.1", "192.168.1.2", "192.168.1.3"}
    connections, err := pssh.ConnectToMultipleDevices(deviceIPs, creds, parentWindow)
    if err != nil {
        dialog.ShowError(err, parentWindow)
        return
    }
    
    // Open terminals
    err = pssh.OpenMultipleTerminals(connections, true) // true for tabbed mode
    if err != nil {
        dialog.ShowError(err, parentWindow)
    }
})
```

### Context Menu Integration

```go
// Add to your device table
selectedDevices := []string{"192.168.1.1", "192.168.1.2"}
menuActions := pssh.CreateSSHMenuActions(selectedDevices, parentWindow)

menu := fyne.NewMenu("SSH Options", menuActions...)
// Show menu at cursor position
```

## API Reference

### Core Types

#### `ConnectionConfig`
- `Host`: Target hostname or IP address
- `Port`: SSH port (default: 22)
- `Username`: SSH username
- `Password`: SSH password
- `Timeout`: Connection timeout
- `PrivateKey`: SSH private key (optional)

#### `SSHConnection`
- `Connect()`: Establish SSH connection
- `CreateSession()`: Create new SSH session
- `Close()`: Close connection
- `IsConnected()`: Check connection status

#### `SSHManager`
- `ConnectMultiple()`: Connect to multiple hosts in parallel
- `GetConnection()`: Get connection by host
- `CloseAll()`: Close all managed connections

#### `TerminalManager`
- `CreateTerminalWidget()`: Create terminal widget for SSH connection
- `MultiTerminalWindow()`: Create tabbed terminal window
- `CloseAllTerminals()`: Close all terminal widgets

### UI Helpers

#### `ShowCredentialsDialog()`
Shows a dialog to collect SSH credentials from user.

#### `ShowConnectionProgress()`
Displays connection progress for multiple devices.

#### `ConnectToMultipleDevices()`
High-level function to connect to multiple devices with progress tracking.

#### `OpenMultipleTerminals()`
Opens terminal windows for multiple SSH connections.

## Error Handling

The package provides comprehensive error handling:

- Connection timeouts
- Authentication failures
- Network errors
- Session creation failures
- Terminal widget errors

All functions return detailed error information for proper debugging.

## Security Notes

- The package uses `ssh.InsecureIgnoreHostKey()` for simplicity
- In production, implement proper host key verification
- Support for both password and SSH key authentication
- Credentials are handled securely in memory

## Thread Safety

- All managers are thread-safe with proper mutex locking
- Concurrent connections are safely managed
- Terminal widgets can be created and destroyed concurrently

## Integration with Existing Projects

See `examples.go` for detailed integration examples with:
- Device scanner results
- Fyne UI components
- Context menus
- Progress dialogs
- Error handling patterns
