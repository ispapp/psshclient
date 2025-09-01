package pssh

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"github.com/fyne-io/terminal"
	"golang.org/x/crypto/ssh"
)

var errorPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)error`),
	regexp.MustCompile(`(?i)fail`),
	regexp.MustCompile(`(?i)closed`),
	regexp.MustCompile(`(?i)command not found`),
	regexp.MustCompile(`(?i)permission denied`),
	regexp.MustCompile(`(?i)no such file or directory`),
	regexp.MustCompile(`(?i)unrecognized command`),
	regexp.MustCompile(`(?i)invalid argument`),
}

// TerminalWidget represents a terminal connected to an SSH session
type TerminalWidget struct {
	Connection *SSHConnection
	Session    *ssh.Session
	Terminal   *terminal.Terminal
	Window     fyne.Window
	App        fyne.App
	StdinPipe  io.WriteCloser
	StdoutPipe io.Reader
	mutex      sync.RWMutex
}

// TerminalManager manages multiple terminal widgets
type TerminalManager struct {
	terminals map[string]*TerminalWidget
	mutex     sync.RWMutex
}

// NewTerminalManager creates a new terminal manager
func NewTerminalManager() *TerminalManager {
	return &TerminalManager{
		terminals: make(map[string]*TerminalWidget),
	}
}

// CreateTerminalWidget creates a new terminal widget for an SSH connection
func (tm *TerminalManager) CreateTerminalWidget(conn *SSHConnection, title string) (*TerminalWidget, error) {
	if !conn.IsConnected() {
		return nil, fmt.Errorf("SSH connection is not established")
	}

	// Create a new SSH session
	session, err := conn.CreateSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create SSH session: %v", err)
	}

	// Request a pseudo-terminal with proper settings for interactive shell
	err = session.RequestPty("xterm-256color", 80, 24, ssh.TerminalModes{
		ssh.ECHO:          1,     // enable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	})
	if err != nil {
		session.Close()
		return nil, fmt.Errorf("failed to request pty: %v", err)
	}

	// Get stdin and stdout pipes
	stdinPipe, err := session.StdinPipe()
	if err != nil {
		session.Close()
		return nil, fmt.Errorf("failed to get stdin pipe: %v", err)
	}

	stdoutPipe, err := session.StdoutPipe()
	if err != nil {
		session.Close()
		return nil, fmt.Errorf("failed to get stdout pipe: %v", err)
	}

	// Wrap stdout with banner suppression
	customBanner := fmt.Sprintf("Connected to %s (%s)", conn.Config.Host, conn.Config.Username)
	wrappedStdout := newBanneredReader(stdoutPipe, customBanner)

	// Create Fyne app and window
	a := app.New()
	w := a.NewWindow(title)
	w.Resize(fyne.NewSize(800, 600))

	// Create terminal widget directly (no fyne.Do needed yet)
	t := terminal.New()
	// Force the terminal to initialize its internal structures
	t.Resize(fyne.NewSize(800, 600))

	// Start the SSH shell session
	go func() {
		err := session.Shell()
		if err != nil {
			fmt.Printf("SSH shell failed: %v\n", err)
		} else {
			err = session.Wait()
			if err != nil {
				fmt.Printf("SSH session ended: %v\n", err)
			}
		}
	}()

	// Connect terminal to SSH session
	// Set window content directly
	w.SetContent(t)

	go func() {
		defer func() {
			// Use a goroutine to avoid blocking
			go func() {
				a.Quit()
			}()
		}()
		err := t.RunWithConnection(stdinPipe, wrappedStdout)
		if err != nil {
			fmt.Printf("Terminal connection error: %v\n", err)
		}
	}()

	// Set up dynamic terminal resizing
	configChan := make(chan terminal.Config, 1)
	go func() {
		rows, cols := uint(0), uint(0)
		for {
			config := <-configChan
			if rows == config.Rows && cols == config.Columns {
				continue
			}
			rows, cols = config.Rows, config.Columns
			err := session.WindowChange(int(rows), int(cols))
			if err != nil {
				fmt.Printf("Failed to resize terminal: %v\n", err)
			}
		}
	}()

	// Add listener directly (no fyne.Do needed yet)
	t.AddListener(configChan)

	// Create terminal widget instance
	termWidget := &TerminalWidget{
		Connection: conn,
		Session:    session,
		Terminal:   t,
		Window:     w,
		App:        a,
		StdinPipe:  stdinPipe,
		StdoutPipe: wrappedStdout,
	}

	// Set window content using fyne.Do to ensure UI thread safety - already done above
	// (window content was set before connecting terminal)

	// Store terminal widget
	tm.mutex.Lock()
	tm.terminals[conn.Config.Host] = termWidget
	tm.mutex.Unlock()

	return termWidget, nil
}

// ShowTerminal displays the terminal window
func (tw *TerminalWidget) ShowTerminal() {
	// ShowAndRun() should be called directly, not through fyne.Do()
	tw.Window.ShowAndRun()
}

// ShowTerminalWindow displays the terminal window without running (for multiple terminals)
func (tw *TerminalWidget) ShowTerminalWindow() {
	// Show() should be called directly, not through fyne.Do()
	tw.Window.Show()
}

// Close closes the terminal widget and SSH session
func (tw *TerminalWidget) Close() error {
	tw.mutex.Lock()
	defer tw.mutex.Unlock()

	var errs []error

	if tw.Session != nil {
		if err := tw.Session.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close SSH session: %v", err))
		}
	}

	if tw.Window != nil {
		// Close window directly
		tw.Window.Close()
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing terminal: %v", errs)
	}

	return nil
}

// GetTerminal returns a terminal widget by host
func (tm *TerminalManager) GetTerminal(host string) (*TerminalWidget, bool) {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()
	terminal, exists := tm.terminals[host]
	return terminal, exists
}

// CloseTerminal closes a specific terminal
func (tm *TerminalManager) CloseTerminal(host string) error {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	if terminal, exists := tm.terminals[host]; exists {
		err := terminal.Close()
		delete(tm.terminals, host)
		return err
	}

	return fmt.Errorf("terminal for host %s not found", host)
}

// CloseAllTerminals closes all managed terminals
func (tm *TerminalManager) CloseAllTerminals() error {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	var errs []error
	for host, terminal := range tm.terminals {
		if err := terminal.Close(); err != nil {
			errs = append(errs, fmt.Errorf("error closing terminal for %s: %v", host, err))
		}
		delete(tm.terminals, host)
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing terminals: %v", errs)
	}

	return nil
}

// ListTerminals returns a list of active terminal hosts
func (tm *TerminalManager) ListTerminals() []string {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	var hosts []string
	for host := range tm.terminals {
		hosts = append(hosts, host)
	}
	return hosts
}

// MultiTerminalWindow creates a window with multiple terminals in tabs
func (tm *TerminalManager) MultiTerminalWindow(connections []*SSHConnection, title string) error {
	if len(connections) == 0 {
		return fmt.Errorf("no connections provided")
	}

	// Create main app and window
	a := app.New()
	w := a.NewWindow(title)
	w.Resize(fyne.NewSize(1000, 700))

	// Create tab container for multiple terminals
	tabs := container.NewAppTabs()

	// Create terminals for each connection
	for _, conn := range connections {
		if !conn.IsConnected() {
			continue
		}

		// Create SSH session
		session, err := conn.CreateSession()
		if err != nil {
			fmt.Printf("Failed to create session for %s: %v\n", conn.Config.Host, err)
			continue
		}

		// Request a pseudo-terminal with proper settings for interactive shell
		err = session.RequestPty("xterm-256color", 80, 24, ssh.TerminalModes{
			ssh.ECHO:          1,     // enable echoing
			ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
			ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
		})
		if err != nil {
			session.Close()
			fmt.Printf("Failed to request pty for %s: %v\n", conn.Config.Host, err)
			continue
		}

		// Get pipes
		stdinPipe, err := session.StdinPipe()
		if err != nil {
			session.Close()
			fmt.Printf("Failed to get stdin pipe for %s: %v\n", conn.Config.Host, err)
			continue
		}

		stdoutPipe, err := session.StdoutPipe()
		if err != nil {
			session.Close()
			fmt.Printf("Failed to get stdout pipe for %s: %v\n", conn.Config.Host, err)
			continue
		}

		// Wrap stdout with banner suppression
		customBanner := fmt.Sprintf("Connected to %s (%s)", conn.Config.Host, conn.Config.Username)
		wrappedStdout := newBanneredReader(stdoutPipe, customBanner)

		// Create terminal directly (no fyne.Do needed yet)
		t := terminal.New()
		t.Resize(fyne.NewSize(1000, 700))
		// Start SSH shell
		go func(s *ssh.Session, host string) {
			err := s.Shell()
			if err != nil {
				fmt.Printf("Failed to start shell for %s: %v\n", host, err)
			} else {
				err = s.Wait()
				if err != nil {
					fmt.Printf("SSH session ended for %s: %v\n", host, err)
				}
			}
		}(session, conn.Config.Host)

		// Connect terminal to SSH
		go func(term *terminal.Terminal, stdin io.WriteCloser, stdout io.Reader, host string) {
			err := term.RunWithConnection(stdin, stdout)
			if err != nil {
				fmt.Printf("Terminal connection error for %s: %v\n", host, err)
			}
		}(t, stdinPipe, wrappedStdout, conn.Config.Host)

		// Set up resizing
		configChan := make(chan terminal.Config, 1)
		go func(s *ssh.Session, ch <-chan terminal.Config, host string) {
			rows, cols := uint(0), uint(0)
			for {
				config := <-ch
				if rows == config.Rows && cols == config.Columns {
					continue
				}
				rows, cols = config.Rows, config.Columns
				err := s.WindowChange(int(rows), int(cols))
				if err != nil {
					fmt.Printf("Failed to resize terminal for %s: %v\n", host, err)
				}
			}
		}(session, configChan, conn.Config.Host)

		// Add listener directly (no fyne.Do needed yet)
		t.AddListener(configChan)

		// Create terminal widget and store it
		termWidget := &TerminalWidget{
			Connection: conn,
			Session:    session,
			Terminal:   t,
			Window:     w,
			App:        a,
			StdinPipe:  stdinPipe,
			StdoutPipe: wrappedStdout,
		}

		tm.mutex.Lock()
		tm.terminals[conn.Config.Host] = termWidget
		tm.mutex.Unlock()

		// Add tab directly (no fyne.Do needed yet)
		tabTitle := fmt.Sprintf("%s (%s)", conn.Config.Host, conn.Config.Username)
		tabs.Append(container.NewTabItem(tabTitle, t))
	}

	if len(tabs.Items) == 0 {
		return fmt.Errorf("no successful terminal connections created")
	}

	// Set window content and show directly
	w.SetContent(tabs)
	w.ShowAndRun()

	return nil
}

// SSHMultiTerminal represents a single terminal widget handling multiple SSH sessions
type SSHMultiTerminal struct {
	sessions       []*ssh.Session
	connections    []*SSHConnection
	terminal       *terminal.Terminal
	stdinWriters   []io.WriteCloser
	stdoutReaders  []io.Reader  // Store stdout readers during setup
	activeSessions map[int]bool // Track which sessions are still active
	multiStdin     *multiWriter // Store multi-writer for proper cleanup
	multiStdout    *multiReader // Store multi-reader for proper cleanup
	state          *SSHMultiTerminalState
	closed         bool       // Track if already closed to prevent double cleanup
	closeMutex     sync.Mutex // Protect against concurrent close calls
}

// SSHMultiTerminalState manages shared state for a multi-terminal session.
type SSHMultiTerminalState struct {
	mutex           sync.Mutex
	isTabCompletion bool
}

// multiWriter writes to multiple writers simultaneously
type multiWriter struct {
	writers []io.WriteCloser
	state   *SSHMultiTerminalState
}

func (mw *multiWriter) Write(p []byte) (n int, err error) {
	// Check for TAB key press
	if mw.state != nil && bytes.Equal(p, []byte{'\t'}) {
		mw.state.mutex.Lock()
		mw.state.isTabCompletion = true
		mw.state.mutex.Unlock()

		// Reset the flag after a short delay to allow remote hosts to respond.
		time.AfterFunc(250*time.Millisecond, func() {
			mw.state.mutex.Lock()
			mw.state.isTabCompletion = false
			mw.state.mutex.Unlock()
		})
	}

	for _, w := range mw.writers {
		if w != nil {
			w.Write(p)
		}
	}
	return len(p), nil
}

func (mw *multiWriter) Close() error {
	var errs []error
	for _, w := range mw.writers {
		if w != nil {
			if err := w.Close(); err != nil {
				// Ignore EOF errors as they indicate the pipe is already closed
				if err != io.EOF {
					errs = append(errs, err)
				}
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing writers: %v", errs)
	}
	return nil
}

// multiReader reads from multiple readers and combines their output
type multiReader struct {
	readers []io.Reader
	output  chan []byte
	done    chan struct{}
	mutex   sync.RWMutex
	buffer  bytes.Buffer // Internal buffer for unread data
	state   *SSHMultiTerminalState
}

func newMultiReader(readers []io.Reader, hostnames []string, state *SSHMultiTerminalState) *multiReader {
	// Buffer the output channel to handle multiple concurrent writers
	mr := &multiReader{
		readers: readers,
		output:  make(chan []byte, len(readers)*2),
		done:    make(chan struct{}),
		state:   state,
	}

	var wg sync.WaitGroup
	wg.Add(len(readers))

	// Start goroutines to read from each reader
	for i, reader := range readers {
		go func(r io.Reader, id int, host string) {
			defer wg.Done()
			buf := make([]byte, 8192) // Use a larger buffer for efficiency
			for {
				select {
				case <-mr.done:
					return
				default:
					n, err := r.Read(buf)
					if n > 0 {
						data := buf[:n]

						// Handle TAB completion filtering
						isTab := false
						if mr.state != nil {
							mr.state.mutex.Lock()
							isTab = mr.state.isTabCompletion
							mr.state.mutex.Unlock()
						}

						// If it's a tab completion, only show output from the last reader.
						if isTab && id != len(mr.readers)-1 {
							continue // Skip output from all but the last session
						}

						// Check for error keywords in the output
						for _, pattern := range errorPatterns {
							if pattern.Match(data) {
								// Send notification via the main app thread
								// fyne.CurrentApp().SendNotification(fyne.NewNotification(
								// 	fmt.Sprintf("Alert from %s", host),
								// 	string(data),
								// ))
								break // Send only one notification per data chunk
							}
						}

						// Prefix output with hostname for clarity
						prefixedData := []byte(fmt.Sprintf("[%s]> ", host))
						prefixedData = append(prefixedData, data...)
						mr.output <- prefixedData
					}
					if err != nil {
						if err != io.EOF {
							fmt.Printf("Error reading from session %d (%s): %v\n", id+1, host, err)
						}
						return // End goroutine on error or EOF
					}
				}
			}
		}(reader, i, hostnames[i])
	}

	// Goroutine to close the output channel once all readers have finished
	go func() {
		wg.Wait()
		close(mr.output)
	}()

	return mr
}

func (mr *multiReader) Read(p []byte) (n int, err error) {
	mr.mutex.Lock()
	defer mr.mutex.Unlock()

	// If our internal buffer has data, read from it first.
	if mr.buffer.Len() > 0 {
		return mr.buffer.Read(p)
	}

	// Wait for new data from the channel or for the reader to be closed.
	select {
	case data, ok := <-mr.output:
		if !ok {
			return 0, io.EOF // Channel is closed, no more data.
		}
		// Write new data to our buffer and then read from it.
		mr.buffer.Write(data)
		return mr.buffer.Read(p)
	case <-mr.done:
		return 0, io.EOF
	}
}

func (mr *multiReader) Close() error {
	mr.mutex.Lock()
	defer mr.mutex.Unlock()

	select {
	case <-mr.done:
		// Already closed
	default:
		close(mr.done)
	}
	return nil
}

// banneredReader wraps a reader to suppress banners and add custom messages
type banneredReader struct {
	reader       io.Reader
	customBanner string
	firstRead    bool
}

func newBanneredReader(reader io.Reader, customBanner string) *banneredReader {
	return &banneredReader{
		reader:       reader,
		customBanner: customBanner,
		firstRead:    true,
		// Removed banner patterns since we're doing pass-through now
	}
}

func (br *banneredReader) Read(p []byte) (n int, err error) {
	// For multi-terminal scenarios, we want minimal interference
	// Just read from underlying reader without modification
	return br.reader.Read(p)
}

// NewSSHMultiTerminal creates a single terminal widget that handles multiple SSH sessions
func (tm *TerminalManager) NewSSHMultiTerminal(connections []*SSHConnection) (*SSHMultiTerminal, error) {
	fmt.Printf("NewSSHMultiTerminal called with %d connections\n", len(connections))

	if len(connections) == 0 {
		return nil, fmt.Errorf("no connections provided")
	}

	var sessions []*ssh.Session
	var stdinWriters []io.WriteCloser
	var stdoutReaders []io.Reader
	var validConnections []*SSHConnection
	var hostnames []string

	// Create sessions for all valid connections
	for i, conn := range connections {
		fmt.Printf("Processing connection %d: %s (connected: %v)\n", i, conn.Config.Host, conn.IsConnected())

		if !conn.IsConnected() {
			fmt.Printf("Skipping disconnected host: %s\n", conn.Config.Host)
			continue
		}

		// Create SSH session
		session, err := conn.CreateSession()
		if err != nil {
			fmt.Printf("Failed to create session for %s: %v\n", conn.Config.Host, err)
			continue
		}

		// Request a pseudo-terminal with proper settings for interactive shell
		err = session.RequestPty("xterm-256color", 80, 24, ssh.TerminalModes{
			ssh.ECHO:          1,      // enable echoing
			ssh.TTY_OP_ISPEED: 115200, // input speed = 115.2kbaud
			ssh.TTY_OP_OSPEED: 115200, // output speed = 115.2kbaud
		})
		if err != nil {
			session.Close()
			fmt.Printf("Failed to request pty for %s: %v\n", conn.Config.Host, err)
			continue
		}

		// Get stdin and stdout pipes
		stdinPipe, err := session.StdinPipe()
		if err != nil {
			session.Close()
			fmt.Printf("Failed to get stdin pipe for %s: %v\n", conn.Config.Host, err)
			continue
		}

		stdoutPipe, err := session.StdoutPipe()
		if err != nil {
			session.Close()
			fmt.Printf("Failed to get stdout pipe for %s: %v\n", conn.Config.Host, err)
			continue
		}

		// Wrap stdout with banner suppression
		customBanner := fmt.Sprintf("Connected to %s (%s)", conn.Config.Host, conn.Config.Username)
		wrappedStdout := newBanneredReader(stdoutPipe, customBanner)

		sessions = append(sessions, session)
		stdinWriters = append(stdinWriters, stdinPipe)
		stdoutReaders = append(stdoutReaders, wrappedStdout)
		validConnections = append(validConnections, conn)
		hostnames = append(hostnames, conn.Config.Host)

		fmt.Printf("Successfully created session for %s\n", conn.Config.Host)
	}

	fmt.Printf("Created %d valid sessions out of %d connections\n", len(sessions), len(connections))

	if len(sessions) == 0 {
		return nil, fmt.Errorf("no valid SSH sessions could be created")
	}

	fmt.Printf("Starting %d SSH shell sessions...\n", len(sessions))
	// Track active sessions
	activeSessions := make(map[int]bool)
	for i := range sessions {
		activeSessions[i] = true
	}

	// Start SSH shell sessions - start the shell AFTER creating the terminal connection
	for i, session := range sessions {
		go func(s *ssh.Session, host string, index int) {
			defer func() {
				// Mark session as inactive when it ends
				activeSessions[index] = false
			}()

			fmt.Printf("Starting interactive shell session for %s (session %d)\n", host, index)

			// Start interactive shell
			err := s.Shell()
			if err != nil {
				fmt.Printf("Failed to start shell for %s: %v\n", host, err)
				return
			}

			fmt.Printf("Shell started successfully for %s\n", host)

			// Wait for shell to end
			err = s.Wait()
			if err != nil {
				fmt.Printf("SSH shell session ended for %s: %v\n", host, err)
			} else {
				fmt.Printf("SSH shell session for %s ended normally\n", host)
			}
		}(session, validConnections[i].Config.Host, i)
	}

	// Create terminal widget immediately
	t := terminal.New()
	t.Resize(fyne.NewSize(1000, 700))

	// Create a shared state for the session
	state := &SSHMultiTerminalState{}

	// Create multi-writer for stdin (distributes input to all sessions)
	multiStdin := &multiWriter{writers: stdinWriters, state: state}

	// Create multi-reader for stdout (combines output from all sessions)
	multiStdout := newMultiReader(stdoutReaders, hostnames, state)

	// Connect terminal to the multi-reader/writer in a goroutine
	go func() {
		defer func() {
			fmt.Printf("Terminal connection goroutine ending\n")
		}()

		err := t.RunWithConnection(multiStdin, multiStdout)
		if err != nil {
			fmt.Printf("Terminal connection error: %v\n", err)
		}
	}()

	// Set up dynamic terminal resizing for all sessions
	configChan := make(chan terminal.Config, 1)
	go func() {
		rows, cols := uint(0), uint(0)
		for {
			config := <-configChan
			if rows == config.Rows && cols == config.Columns {
				continue
			}
			rows, cols = config.Rows, config.Columns

			fmt.Printf("Terminal resize requested: %dx%d\n", rows, cols)

			// Resize all sessions with better error handling
			for i, session := range sessions {
				if session == nil {
					fmt.Printf("Skipping resize for %s: session is nil\n", validConnections[i].Config.Host)
					continue
				}

				// Check if session is still active
				if !activeSessions[i] {
					fmt.Printf("Skipping resize for %s: session ended\n", validConnections[i].Config.Host)
					continue
				}

				// Check if connection is still alive before resizing
				if !validConnections[i].IsConnected() {
					fmt.Printf("Skipping resize for %s: connection lost\n", validConnections[i].Config.Host)
					activeSessions[i] = false // Mark as inactive
					continue
				}

				err := session.WindowChange(int(rows), int(cols))
				if err != nil {
					fmt.Printf("Failed to resize terminal for %s: %v (connection may be lost)\n",
						validConnections[i].Config.Host, err)
					// Don't mark as inactive yet - might just be a temporary error
				} else {
					fmt.Printf("Successfully resized terminal for %s to %dx%d\n",
						validConnections[i].Config.Host, rows, cols)
				}
			}
		}
	}()

	// Add listener for terminal resizing
	t.AddListener(configChan)

	// Create SSH multi-terminal instance
	sshMultiTerm := &SSHMultiTerminal{
		sessions:       sessions,
		connections:    validConnections,
		terminal:       t,
		stdinWriters:   stdinWriters,
		stdoutReaders:  stdoutReaders,
		activeSessions: activeSessions,
		multiStdin:     multiStdin,
		multiStdout:    multiStdout,
		state:          state,
	}

	fmt.Printf("SSH multi-terminal widget created successfully with %d sessions\n", len(sessions))
	return sshMultiTerm, nil
}

// GetWidget returns the terminal widget that can be embedded in any container
func (smt *SSHMultiTerminal) GetWidget() *terminal.Terminal {
	return smt.terminal
}

// CreateStandaloneWindow creates a standalone window containing the multi-terminal widget
func (smt *SSHMultiTerminal) CreateStandaloneWindow(title string) (fyne.Window, fyne.App) {
	a := app.New()
	w := a.NewWindow(title)
	w.Resize(fyne.NewSize(1000, 700))

	w.SetContent(smt.terminal)

	// Set up cleanup when window is closed
	w.SetOnClosed(func() {
		fmt.Printf("Standalone window closed, cleaning up sessions\n")
		smt.Close()
		a.Quit()
		fmt.Printf("Standalone window cleanup complete\n")
	})

	return w, a
}

// Example usage functions

// CreateMultiTerminalWindow creates a standalone window with multiple SSH terminals (legacy compatibility)
func (tm *TerminalManager) CreateMultiTerminalWindow(connections []*SSHConnection, title string) error {
	// Create the multi-terminal widget
	multiTerm, err := tm.NewSSHMultiTerminal(connections)
	if err != nil {
		return fmt.Errorf("failed to create multi-terminal: %v", err)
	}

	// Create a standalone window for it
	window, _ := multiTerm.CreateStandaloneWindow(title)

	// Show and run the window (this will block until the window is closed)
	window.ShowAndRun()

	return nil
}

// CreateMultiTerminalWidget creates a multi-terminal widget that can be embedded in any container
func (tm *TerminalManager) CreateMultiTerminalWidget(connections []*SSHConnection) (*terminal.Terminal, *SSHMultiTerminal, error) {
	// Create the multi-terminal widget
	multiTerm, err := tm.NewSSHMultiTerminal(connections)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create multi-terminal: %v", err)
	}

	// Return both the widget and the controller for lifecycle management
	return multiTerm.GetWidget(), multiTerm, nil
}

// Close closes all SSH sessions and the terminal
func (smt *SSHMultiTerminal) Close() error {
	smt.closeMutex.Lock()
	defer smt.closeMutex.Unlock()

	// Prevent double cleanup
	if smt.closed {
		fmt.Printf("SSHMultiTerminal.Close() already called, skipping\n")
		return nil
	}

	fmt.Printf("SSHMultiTerminal.Close() called\n")
	smt.closed = true

	var errs []error

	// Close multi-reader and multi-writer first
	if smt.multiStdout != nil {
		if err := smt.multiStdout.Close(); err != nil {
			// multiReader close should not fail, but handle gracefully
			fmt.Printf("Warning: failed to close multi-reader: %v\n", err)
		}
	}

	if smt.multiStdin != nil {
		if err := smt.multiStdin.Close(); err != nil {
			// Only report significant errors (non-EOF)
			if err != io.EOF && !isConnectionClosed(err) {
				errs = append(errs, fmt.Errorf("failed to close multi-writer: %v", err))
			} else {
				fmt.Printf("Multi-writer already closed or connections ended: %v\n", err)
			}
		}
	}

	// Close all SSH sessions
	for i, session := range smt.sessions {
		if session != nil {
			if err := session.Close(); err != nil {
				// Only report non-EOF errors as significant
				if err != io.EOF && err.Error() != "session is not connected" {
					errs = append(errs, fmt.Errorf("failed to close SSH session for %s: %v",
						smt.connections[i].Config.Host, err))
				} else {
					fmt.Printf("SSH session for %s already closed (%v)\n", smt.connections[i].Config.Host, err)
				}
			}
		}
	}

	// Close stdin writers (ignore EOF errors as they indicate already closed pipes)
	for i, writer := range smt.stdinWriters {
		if writer != nil {
			if err := writer.Close(); err != nil {
				// Ignore EOF errors as they indicate the pipe is already closed
				if err != io.EOF {
					errs = append(errs, fmt.Errorf("failed to close stdin writer for %s: %v",
						smt.connections[i].Config.Host, err))
				} else {
					fmt.Printf("Stdin writer for %s already closed (EOF)\n", smt.connections[i].Config.Host)
				}
			}
		}
	}

	if len(errs) > 0 {
		fmt.Printf("Errors during cleanup: %v\n", errs)
		return fmt.Errorf("errors closing SSH multi-terminal: %v", errs)
	}

	fmt.Printf("SSHMultiTerminal.Close() completed successfully\n")
	return nil
}

// GetConnectedHosts returns a list of connected host names
func (smt *SSHMultiTerminal) GetConnectedHosts() []string {
	var hosts []string
	for _, conn := range smt.connections {
		hosts = append(hosts, conn.Config.Host)
	}
	return hosts
}

// GetSessionCount returns the number of active SSH sessions
func (smt *SSHMultiTerminal) GetSessionCount() int {
	return len(smt.sessions)
}

// isConnectionClosed checks if an error indicates a closed connection
func isConnectionClosed(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return errStr == "session is not connected" ||
		errStr == "connection lost" ||
		errStr == "broken pipe" ||
		errStr == "use of closed network connection" ||
		errStr == "EOF"
}
