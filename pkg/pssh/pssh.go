package pssh

import (
	"fmt"
	"net"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

// ConnectionConfig holds SSH connection configuration
type ConnectionConfig struct {
	Host       string
	Port       int
	Username   string
	Password   string
	Timeout    time.Duration
	PrivateKey []byte // Optional: SSH private key for key-based auth
}

// SSHConnection represents an active SSH connection
type SSHConnection struct {
	Config    ConnectionConfig
	Client    *ssh.Client
	Session   *ssh.Session
	Connected bool
	Error     error
	mutex     sync.RWMutex
}

// ConnectionResult holds the result of a connection attempt
type ConnectionResult struct {
	Host       string
	Connection *SSHConnection
	Error      error
}

// SSHManager manages multiple SSH connections
type SSHManager struct {
	connections map[string]*SSHConnection
	mutex       sync.RWMutex
}

// NewSSHManager creates a new SSH manager
func NewSSHManager() *SSHManager {
	return &SSHManager{
		connections: make(map[string]*SSHConnection),
	}
}

// NewConnectionConfig creates a new connection configuration
func NewConnectionConfig(host string, port int, username, password string) ConnectionConfig {
	return ConnectionConfig{
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
		Timeout:  30 * time.Second,
	}
}

// Connect establishes an SSH connection
func (conn *SSHConnection) Connect() error {
	conn.mutex.Lock()
	defer conn.mutex.Unlock()

	// Create SSH client configuration
	config := &ssh.ClientConfig{
		User:            conn.Config.Username,
		Timeout:         conn.Config.Timeout,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // Note: Use proper host key verification in production
	}

	// Set authentication method
	if len(conn.Config.PrivateKey) > 0 {
		// Use private key authentication
		signer, err := ssh.ParsePrivateKey(conn.Config.PrivateKey)
		if err != nil {
			conn.Error = fmt.Errorf("failed to parse private key: %v", err)
			return conn.Error
		}
		config.Auth = []ssh.AuthMethod{ssh.PublicKeys(signer)}
	} else {
		// Use password authentication
		config.Auth = []ssh.AuthMethod{ssh.Password(conn.Config.Password)}
	}

	// Connect to SSH server
	address := net.JoinHostPort(conn.Config.Host, fmt.Sprintf("%d", conn.Config.Port))
	client, err := ssh.Dial("tcp", address, config)
	if err != nil {
		conn.Error = fmt.Errorf("failed to connect to %s: %v", address, err)
		conn.Connected = false
		return conn.Error
	}

	conn.Client = client
	conn.Connected = true
	conn.Error = nil
	return nil
}

// CreateSession creates a new SSH session
func (conn *SSHConnection) CreateSession() (*ssh.Session, error) {
	conn.mutex.RLock()
	defer conn.mutex.RUnlock()

	if !conn.Connected || conn.Client == nil {
		return nil, fmt.Errorf("connection not established")
	}

	session, err := conn.Client.NewSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %v", err)
	}

	conn.Session = session
	return session, nil
}

// Close closes the SSH connection
func (conn *SSHConnection) Close() error {
	conn.mutex.Lock()
	defer conn.mutex.Unlock()

	var errs []error

	if conn.Session != nil {
		if err := conn.Session.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close session: %v", err))
		}
		conn.Session = nil
	}

	if conn.Client != nil {
		if err := conn.Client.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close client: %v", err))
		}
		conn.Client = nil
	}

	conn.Connected = false

	if len(errs) > 0 {
		return fmt.Errorf("errors closing connection: %v", errs)
	}

	return nil
}

// IsConnected returns whether the connection is active
func (conn *SSHConnection) IsConnected() bool {
	conn.mutex.RLock()
	defer conn.mutex.RUnlock()
	return conn.Connected
}

// ConnectMultiple connects to multiple hosts in parallel
func (manager *SSHManager) ConnectMultiple(configs []ConnectionConfig) <-chan ConnectionResult {
	resultChan := make(chan ConnectionResult, len(configs))

	var wg sync.WaitGroup
	for _, config := range configs {
		wg.Add(1)
		go func(cfg ConnectionConfig) {
			defer wg.Done()

			conn := &SSHConnection{Config: cfg}
			err := conn.Connect()

			result := ConnectionResult{
				Host:       cfg.Host,
				Connection: conn,
				Error:      err,
			}

			if err == nil {
				manager.mutex.Lock()
				manager.connections[cfg.Host] = conn
				manager.mutex.Unlock()
			}

			resultChan <- result
		}(config)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	return resultChan
}

// GetConnection returns a connection by host
func (manager *SSHManager) GetConnection(host string) (*SSHConnection, bool) {
	manager.mutex.RLock()
	defer manager.mutex.RUnlock()
	conn, exists := manager.connections[host]
	return conn, exists
}

// CloseAll closes all managed connections
func (manager *SSHManager) CloseAll() error {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()

	var errs []error
	for host, conn := range manager.connections {
		if err := conn.Close(); err != nil {
			errs = append(errs, fmt.Errorf("error closing connection to %s: %v", host, err))
		}
		delete(manager.connections, host)
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing connections: %v", errs)
	}

	return nil
}

// ListConnections returns a list of connected hosts
func (manager *SSHManager) ListConnections() []string {
	manager.mutex.RLock()
	defer manager.mutex.RUnlock()

	var hosts []string
	for host, conn := range manager.connections {
		if conn.IsConnected() {
			hosts = append(hosts, host)
		}
	}
	return hosts
}

// TestConnection tests if a host is reachable on the specified port
func TestConnection(host string, port int, timeout time.Duration) error {
	address := net.JoinHostPort(host, fmt.Sprintf("%d", port))
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %v", address, err)
	}
	defer conn.Close()
	return nil
}
