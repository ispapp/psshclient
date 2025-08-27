package pssh

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strings"
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

// NewConnectionConfigForMikroTik creates a connection config optimized for MikroTik devices
func NewConnectionConfigForMikroTik(host, username, password string) ConnectionConfig {
	return ConnectionConfig{
		Host:     host,
		Port:     22,
		Username: username,
		Password: password,
		Timeout:  15 * time.Second, // MikroTik devices might be slower to respond
	}
}

// NewConnectionConfigWithKey creates a connection config with SSH key authentication
func NewConnectionConfigWithKey(host, username string, privateKey []byte) ConnectionConfig {
	return ConnectionConfig{
		Host:       host,
		Port:       22,
		Username:   username,
		PrivateKey: privateKey,
		Timeout:    30 * time.Second,
	}
}

// NewSSHConnection creates a new SSH connection with the given configuration
func NewSSHConnection(config ConnectionConfig) *SSHConnection {
	return &SSHConnection{
		Config:    config,
		Connected: false,
	}
}

// Connect establishes an SSH connection using enhanced authentication
func (conn *SSHConnection) Connect() error {
	return conn.ConnectWithAllMethods()
}

// ConnectWithAllMethods attempts connection with multiple authentication methods
func (conn *SSHConnection) ConnectWithAllMethods() error {
	conn.mutex.Lock()
	defer conn.mutex.Unlock()

	// Create SSH client configuration
	config := &ssh.ClientConfig{
		User:            conn.Config.Username,
		Timeout:         conn.Config.Timeout,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // Note: Use proper host key verification in production
	}

	// Build authentication methods in order of preference
	var authMethods []ssh.AuthMethod

	// 1. Private key authentication (if provided)
	if len(conn.Config.PrivateKey) > 0 {
		signer, err := ssh.ParsePrivateKey(conn.Config.PrivateKey)
		if err != nil {
			fmt.Printf("Warning: Failed to parse private key: %v\n", err)
		} else {
			authMethods = append(authMethods, ssh.PublicKeys(signer))
		}
	}

	// 2. Password authentication (if provided)
	if conn.Config.Password != "" {
		authMethods = append(authMethods, ssh.Password(conn.Config.Password))
	}

	// 3. Keyboard-interactive authentication (for systems that require it)
	if conn.Config.Password != "" {
		authMethods = append(authMethods, ssh.KeyboardInteractive(func(user, instruction string, questions []string, echos []bool) ([]string, error) {
			answers := make([]string, len(questions))
			for i := range answers {
				answers[i] = conn.Config.Password
			}
			return answers, nil
		}))
	}

	// 4. Try common default keys if no specific key provided
	if len(conn.Config.PrivateKey) == 0 {
		// Try to load common SSH keys from default locations
		commonKeys := []string{
			"~/.ssh/id_rsa",
			"~/.ssh/id_ed25519",
			"~/.ssh/id_ecdsa",
		}

		for _, keyPath := range commonKeys {
			if signer, err := loadPrivateKeyFromFile(keyPath); err == nil {
				authMethods = append(authMethods, ssh.PublicKeys(signer))
			}
		}
	}

	if len(authMethods) == 0 {
		conn.Error = fmt.Errorf("no authentication methods available")
		return conn.Error
	}

	config.Auth = authMethods

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

// ProbeSSHAuthMethods tests what authentication methods are supported by an SSH server
func ProbeSSHAuthMethods(host string, port int) ([]string, error) {
	config := &ssh.ClientConfig{
		User:            "test", // Dummy user to probe auth methods
		Timeout:         5 * time.Second,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth:            []ssh.AuthMethod{}, // No auth methods to trigger method listing
	}

	address := net.JoinHostPort(host, fmt.Sprintf("%d", port))
	_, err := ssh.Dial("tcp", address, config)

	if err != nil {
		// Parse the error to extract supported methods
		errStr := err.Error()
		if strings.Contains(errStr, "no supported methods remain") {
			// Extract methods from error message
			if strings.Contains(errStr, "attempted methods") {
				start := strings.Index(errStr, "[")
				end := strings.Index(errStr, "]")
				if start != -1 && end != -1 {
					methodsStr := errStr[start+1 : end]
					methods := strings.Split(methodsStr, " ")
					return methods, nil
				}
			}
		}
		return nil, fmt.Errorf("failed to probe SSH server: %v", err)
	}

	return []string{"none"}, nil // Shouldn't happen with dummy auth
}

// loadPrivateKeyFromFile loads a private key from a file path
func loadPrivateKeyFromFile(keyPath string) (ssh.Signer, error) {
	// Expand ~ to home directory
	if strings.HasPrefix(keyPath, "~") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		keyPath = filepath.Join(homeDir, keyPath[1:])
	}

	// Check if file exists
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		return nil, err
	}

	// Read the private key file
	keyBytes, err := ioutil.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}

	// Parse the private key
	signer, err := ssh.ParsePrivateKey(keyBytes)
	if err != nil {
		return nil, err
	}

	return signer, nil
}

// RunCommand runs a command on the remote server and returns its output
func (conn *SSHConnection) RunCommand(command string) (string, error) {
	conn.mutex.Lock()
	defer conn.mutex.Unlock()

	if !conn.Connected || conn.Client == nil {
		return "", fmt.Errorf("not connected")
	}

	// Create a new session for this command
	session, err := conn.Client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	// Run the command and capture the output
	output, err := session.CombinedOutput(command)
	if err != nil {
		return "", fmt.Errorf("failed to run command: %w", err)
	}

	return string(output), nil
}
