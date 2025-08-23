# PSSH Package - Project Structure and Usage

## Project Structure

```
pkg/pssh/
├── pssh.go           # Core SSH connection management
├── terminal.go       # Terminal widget integration
├── ui_helpers.go     # UI integration helpers
├── examples.go       # Integration examples
├── pssh_test.go      # Unit tests
└── README.md         # Documentation

cmd/demo/
└── main.go          # Demo application showing package usage
```

## Features Completed ✅

### Core SSH Functionality
- ✅ Parallel SSH connections to multiple hosts
- ✅ Connection management with thread-safe operations
- ✅ Support for password and SSH key authentication
- ✅ Connection testing and validation
- ✅ IPv6 address support

### Terminal Integration
- ✅ Fyne terminal widget integration
- ✅ SSH session piping (stdin/stdout)
- ✅ Dynamic terminal resizing
- ✅ Multiple terminal modes:
  - Individual windows
  - Tabbed interface
- ✅ Terminal lifecycle management

### UI Components
- ✅ Credentials input dialog
- ✅ Connection progress tracking
- ✅ Context menu integration
- ✅ Error handling and user feedback
- ✅ Device selection workflows

### Testing & Documentation
- ✅ Unit tests for core functionality
- ✅ Comprehensive documentation
- ✅ Usage examples
- ✅ Demo application

## Usage Summary

### 1. Import the Package
```go
import "ispappclient/pkg/pssh"
```

### 2. Connect to Multiple Devices
```go
// Create SSH manager
manager := pssh.NewSSHManager()

// Create connection configs
configs := []pssh.ConnectionConfig{
    pssh.NewConnectionConfig("192.168.1.1", 22, "admin", "password"),
    pssh.NewConnectionConfig("192.168.1.2", 22, "admin", "password"),
}

// Connect in parallel
resultChan := manager.ConnectMultiple(configs)
```

### 3. Open Terminal Widgets
```go
// Create terminal manager
termManager := pssh.NewTerminalManager()

// Open tabbed terminals
err := termManager.MultiTerminalWindow(connections, "SSH Terminals")
```

### 4. UI Integration
```go
// Show credentials dialog
pssh.ShowCredentialsDialog(window, func(creds pssh.SSHCredentials, confirmed bool) {
    if confirmed {
        // Connect and open terminals
        connections, _ := pssh.ConnectToMultipleDevices(devices, creds, window)
        pssh.OpenMultipleTerminals(connections, true)
    }
})
```

## Integration with Your Existing App

### Device Table Integration
Add SSH functionality to your device table by:

1. **Adding context menu**:
```go
// In your devices table widget
table.OnSecondaryTap = func(ev *fyne.PointEvent) {
    selectedDevices := getSelectedDevices() // Your implementation
    if len(selectedDevices) > 0 {
        menu := fyne.NewMenu("SSH Options", 
            pssh.CreateSSHMenuActions(selectedDevices, window)...)
        widget.ShowPopUpMenuAtPosition(menu, window.Canvas(), ev.AbsolutePosition)
    }
}
```

2. **Adding SSH buttons to UI**:
```go
sshBtn := widget.NewButton("Open SSH", func() {
    selectedDevices := getSelectedDevices()
    // Use pssh.ShowCredentialsDialog() to get credentials
    // Then pssh.ConnectToMultipleDevices() and pssh.OpenMultipleTerminals()
})
```

### Scanner Integration
Extract device information from your scanner results:

```go
func getSSHDevices(scanResults []scanner.Device) []string {
    var sshHosts []string
    for _, device := range scanResults {
        if device.HasSSH || device.Ports["22"] == "open" {
            sshHosts = append(sshHosts, device.IP)
        }
    }
    return sshHosts
}
```

## Testing

Run the unit tests:
```bash
go test ./pkg/pssh -v
```

Run the demo application:
```bash
go run ./cmd/demo
```

## Dependencies Added

The following dependencies were added to your project:
- `golang.org/x/crypto/ssh` - SSH client functionality
- `github.com/fyne-io/terminal` - Terminal widget for Fyne
- Upgraded Fyne to v2.6.3 for latest features

## Security Notes

- Currently uses `ssh.InsecureIgnoreHostKey()` for simplicity
- For production, implement proper host key verification
- Supports both password and SSH key authentication
- Credentials are handled securely in memory

## Ready to Use! 🚀

The PSSH package is now complete and ready for integration with your device scanner application. You can:

1. Select multiple devices from your scanner results
2. Right-click to show SSH options
3. Enter credentials once for all devices
4. Watch connection progress
5. Access terminals in tabbed or separate windows

The package provides a clean, thread-safe API for managing SSH connections and terminal widgets in your Fyne application.
