package pssh

import (
	"testing"
	"time"
)

func TestNewSSHManager(t *testing.T) {
	manager := NewSSHManager()
	if manager == nil {
		t.Fatal("NewSSHManager returned nil")
	}
	if manager.connections == nil {
		t.Fatal("SSH manager connections map is nil")
	}
}

func TestNewTerminalManager(t *testing.T) {
	termManager := NewTerminalManager()
	if termManager == nil {
		t.Fatal("NewTerminalManager returned nil")
	}
	if termManager.terminals == nil {
		t.Fatal("Terminal manager terminals map is nil")
	}
}

func TestNewConnectionConfig(t *testing.T) {
	config := NewConnectionConfig("192.168.1.1", 22, "admin", "password")

	if config.Host != "192.168.1.1" {
		t.Errorf("Expected host 192.168.1.1, got %s", config.Host)
	}
	if config.Port != 22 {
		t.Errorf("Expected port 22, got %d", config.Port)
	}
	if config.Username != "admin" {
		t.Errorf("Expected username admin, got %s", config.Username)
	}
	if config.Password != "password" {
		t.Errorf("Expected password password, got %s", config.Password)
	}
	if config.Timeout != 30*time.Second {
		t.Errorf("Expected timeout 30s, got %v", config.Timeout)
	}
}

func TestSSHConnection(t *testing.T) {
	config := NewConnectionConfig("invalid-host", 22, "admin", "password")
	conn := &SSHConnection{Config: config}

	// This should fail since it's an invalid host
	err := conn.Connect()
	if err == nil {
		t.Error("Expected connection to fail for invalid host")
	}

	// Test IsConnected
	if conn.IsConnected() {
		t.Error("Connection should not be marked as connected after failed attempt")
	}
}

func TestTestConnection(t *testing.T) {
	// Test with invalid host - should fail
	err := TestConnection("invalid-host-that-does-not-exist", 22, 1*time.Second)
	if err == nil {
		t.Error("Expected connection test to fail for invalid host")
	}

	// Test with localhost on an unlikely port - should fail
	err = TestConnection("127.0.0.1", 99999, 1*time.Second)
	if err == nil {
		t.Error("Expected connection test to fail for closed port")
	}
}
