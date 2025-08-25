# SSH Authentication Fix

## Problem
```
Connection failed for 192.168.1.170: failed to connect to 192.168.1.170:22: ssh: handshake failed: ssh: unable to authenticate, attempted methods [none password], no supported methods remain
```

## Root Cause
The SSH server at 192.168.1.170 doesn't support the authentication methods your client is trying to use. This commonly happens with:

1. **MikroTik devices** - Often require specific authentication methods
2. **Servers with password auth disabled** - Only accept SSH keys
3. **Systems requiring keyboard-interactive auth** - Need interactive prompts

## Solution Applied

### 1. **Enhanced Authentication Methods**
Updated `ConnectWithAllMethods()` to try multiple authentication methods in order:

```go
// 1. SSH Key authentication (if provided)
// 2. Password authentication  
// 3. Keyboard-interactive authentication
// 4. Common default SSH keys (~/.ssh/id_rsa, etc.)
```

### 2. **Added Helper Functions**
- `NewSSHConnection()` - Create connection with config
- `NewConnectionConfigForMikroTik()` - Optimized for MikroTik devices
- `ProbeSSHAuthMethods()` - Discover supported auth methods
- `loadPrivateKeyFromFile()` - Load SSH keys from filesystem

### 3. **Test Program**
Created `test_ssh_connection.go` to diagnose connection issues:

```bash
go run test_ssh_connection.go
```

## How to Fix Your Connection

### Option 1: Use Enhanced Connection (Recommended)
```go
// Replace your current connection code with:
config := pssh.NewConnectionConfigForMikroTik("192.168.1.170", "admin", "your_password")
conn := pssh.NewSSHConnection(config)
err := conn.Connect() // Now uses enhanced auth methods
```

### Option 2: Diagnose First
```go
// Test what auth methods are supported
methods, err := pssh.ProbeSSHAuthMethods("192.168.1.170", 22)
fmt.Printf("Supported methods: %v\n", methods)
```

### Option 3: Try Common MikroTik Credentials
Many MikroTik devices use these defaults:
- Username: `admin`, Password: `` (empty)
- Username: `admin`, Password: `admin`
- Username: `admin`, Password: `password`

## Common MikroTik SSH Issues

### 1. **SSH Service Disabled**
```bash
# Enable SSH on MikroTik via Winbox/WebFig:
/ip service enable ssh
```

### 2. **Wrong Default Credentials**
```bash
# Reset to defaults if needed:
# On RouterOS: /system reset-configuration
```

### 3. **Firewall Blocking SSH**
```bash
# Allow SSH through firewall:
/ip firewall filter add action=accept chain=input dst-port=22 protocol=tcp
```

### 4. **SSH Key Required**
Some MikroTik configurations require SSH keys:
```bash
# Generate SSH key:
ssh-keygen -t rsa -b 2048 -f ~/.ssh/mikrotik_key

# Add public key to MikroTik:
# /user ssh-keys import public-key-file=mikrotik_key.pub user=admin
```

## Testing Steps

1. **Basic Connectivity**
```bash
telnet 192.168.1.170 22
# Should show SSH banner
```

2. **Manual SSH Test**
```bash
ssh -v admin@192.168.1.170
# Shows detailed connection info
```

3. **Try Different Auth Methods**
```bash
# Password auth
ssh -o PreferredAuthentications=password admin@192.168.1.170

# Key auth  
ssh -o PreferredAuthentications=publickey -i ~/.ssh/id_rsa admin@192.168.1.170

# Keyboard-interactive
ssh -o PreferredAuthentications=keyboard-interactive admin@192.168.1.170
```

## Updated Code Usage

### Single Connection
```go
config := pssh.NewConnectionConfigForMikroTik("192.168.1.170", "admin", "")
conn := pssh.NewSSHConnection(config)
if err := conn.Connect(); err != nil {
    log.Printf("Connection failed: %v", err)
} else {
    log.Println("Connected successfully!")
    defer conn.Close()
}
```

### Multiple Connections (Your Use Case)
```go
// The existing ConnectToMultipleDevices() function now automatically
// uses the enhanced authentication methods
devices := []string{"192.168.1.170", "192.168.1.171"}
credentials := pssh.SSHCredentials{
    Username: "admin",
    Password: "", // Try empty password first for MikroTik
}

connections, err := pssh.ConnectToMultipleDevices(devices, credentials, window)
```

## Next Steps

1. **Run the test program** to diagnose the specific issue
2. **Try common MikroTik credentials** listed above
3. **Check MikroTik device configuration** for SSH settings
4. **Use the enhanced connection methods** in your application

The enhanced authentication should automatically handle most common SSH authentication scenarios, including MikroTik devices.
