package pssh

import (
	"fmt"
	"io"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"github.com/fyne-io/terminal"
	"golang.org/x/crypto/ssh"
)

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

	// Create Fyne app and window
	a := app.New()
	w := a.NewWindow(title)
	w.Resize(fyne.NewSize(800, 600))

	// Create terminal widget on UI thread
	var t *terminal.Terminal
	fyne.Do(func() {
		t = terminal.New()
	})

	// Start the SSH shell session
	go func() {
		err := session.Run("$SHELL || bash")
		if err != nil {
			fmt.Printf("SSH session ended: %v\n", err)
		}
	}()

	// Connect terminal to SSH session
	go func() {
		defer func() {
			fyne.Do(func() {
				a.Quit()
			})
		}()
		err := t.RunWithConnection(stdinPipe, stdoutPipe)
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

	// Add listener using fyne.Do to ensure it's called on the UI thread
	fyne.Do(func() {
		t.AddListener(configChan)
	})

	// Create terminal widget instance
	termWidget := &TerminalWidget{
		Connection: conn,
		Session:    session,
		Terminal:   t,
		Window:     w,
		App:        a,
		StdinPipe:  stdinPipe,
		StdoutPipe: stdoutPipe,
	}

	// Set window content using fyne.Do to ensure UI thread safety
	fyne.Do(func() {
		w.SetContent(t)
	})

	// Store terminal widget
	tm.mutex.Lock()
	tm.terminals[conn.Config.Host] = termWidget
	tm.mutex.Unlock()

	return termWidget, nil
}

// ShowTerminal displays the terminal window
func (tw *TerminalWidget) ShowTerminal() {
	fyne.Do(func() {
		tw.Window.ShowAndRun()
	})
}

// ShowTerminalWindow displays the terminal window without running (for multiple terminals)
func (tw *TerminalWidget) ShowTerminalWindow() {
	fyne.Do(func() {
		tw.Window.Show()
	})
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
		fyne.Do(func() {
			tw.Window.Close()
		})
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

		// Create terminal on UI thread
		var t *terminal.Terminal
		fyne.Do(func() {
			t = terminal.New()
		})

		// Start SSH shell
		go func(s *ssh.Session, host string) {
			err := s.Run("$SHELL || bash")
			if err != nil {
				fmt.Printf("SSH session ended for %s: %v\n", host, err)
			}
		}(session, conn.Config.Host)

		// Connect terminal to SSH
		go func(term *terminal.Terminal, stdin io.WriteCloser, stdout io.Reader, host string) {
			err := term.RunWithConnection(stdin, stdout)
			if err != nil {
				fmt.Printf("Terminal connection error for %s: %v\n", host, err)
			}
		}(t, stdinPipe, stdoutPipe, conn.Config.Host)

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

		// Add listener using fyne.Do to ensure UI thread safety
		fyne.Do(func() {
			t.AddListener(configChan)
		})

		// Create terminal widget and store it
		termWidget := &TerminalWidget{
			Connection: conn,
			Session:    session,
			Terminal:   t,
			Window:     w,
			App:        a,
			StdinPipe:  stdinPipe,
			StdoutPipe: stdoutPipe,
		}

		tm.mutex.Lock()
		tm.terminals[conn.Config.Host] = termWidget
		tm.mutex.Unlock()

		// Add tab using fyne.Do for UI thread safety
		tabTitle := fmt.Sprintf("%s (%s)", conn.Config.Host, conn.Config.Username)
		fyne.Do(func() {
			tabs.Append(container.NewTabItem(tabTitle, t))
		})
	}

	if len(tabs.Items) == 0 {
		return fmt.Errorf("no successful terminal connections created")
	}

	// Set window content and show using fyne.Do for UI thread safety
	fyne.Do(func() {
		w.SetContent(tabs)
		w.ShowAndRun()
	})

	return nil
}

// SSHMultiTerminal represents a single terminal widget handling multiple SSH sessions
type SSHMultiTerminal struct {
	sessions       []*ssh.Session
	connections    []*SSHConnection
	terminal       *terminal.Terminal
	window         fyne.Window
	app            fyne.App
	stdinWriters   []io.WriteCloser
	activeSessions map[int]bool // Track which sessions are still active
	mutex          sync.RWMutex
}

// multiWriter writes to multiple writers simultaneously
type multiWriter struct {
	writers []io.WriteCloser
	mutex   sync.RWMutex
}

func (mw *multiWriter) Write(p []byte) (n int, err error) {
	mw.mutex.RLock()
	defer mw.mutex.RUnlock()

	for _, w := range mw.writers {
		if w != nil {
			w.Write(p)
		}
	}
	return len(p), nil
}

func (mw *multiWriter) Close() error {
	mw.mutex.Lock()
	defer mw.mutex.Unlock()

	var errs []error
	for _, w := range mw.writers {
		if w != nil {
			if err := w.Close(); err != nil {
				errs = append(errs, err)
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
}

func newMultiReader(readers []io.Reader) *multiReader {
	mr := &multiReader{
		readers: readers,
		output:  make(chan []byte, 100),
		done:    make(chan struct{}),
	}

	// Start goroutines to read from each reader
	for i, reader := range readers {
		go func(r io.Reader, id int) {
			buf := make([]byte, 1024)
			for {
				select {
				case <-mr.done:
					return
				default:
					n, err := r.Read(buf)
					if n > 0 {
						// Add a prefix to identify which session the output comes from
						if len(mr.readers) > 1 {
							prefixed := fmt.Sprintf("[%d] %s", id+1, buf[:n])
							mr.output <- []byte(prefixed)
						} else {
							mr.output <- buf[:n]
						}
					}
					if err != nil {
						if err != io.EOF {
							fmt.Printf("Error reading from session %d: %v\n", id+1, err)
						}
						return
					}
				}
			}
		}(reader, i)
	}

	return mr
}

func (mr *multiReader) Read(p []byte) (n int, err error) {
	select {
	case data := <-mr.output:
		n = copy(p, data)
		if n < len(data) {
			// If buffer is too small, we lose data - could be improved
			return n, nil
		}
		return n, nil
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

// NewSSHMultiTerminal creates a single terminal that handles multiple SSH sessions
func (tm *TerminalManager) NewSSHMultiTerminal(connections []*SSHConnection, title string) (*SSHMultiTerminal, error) {
	fmt.Printf("NewSSHMultiTerminal called with %d connections\n", len(connections))

	if len(connections) == 0 {
		return nil, fmt.Errorf("no connections provided")
	}

	var sessions []*ssh.Session
	var stdinWriters []io.WriteCloser
	var stdoutReaders []io.Reader
	var validConnections []*SSHConnection

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

		sessions = append(sessions, session)
		stdinWriters = append(stdinWriters, stdinPipe)
		stdoutReaders = append(stdoutReaders, stdoutPipe)
		validConnections = append(validConnections, conn)

		fmt.Printf("Successfully created session for %s\n", conn.Config.Host)
	}

	fmt.Printf("Created %d valid sessions out of %d connections\n", len(sessions), len(connections))

	if len(sessions) == 0 {
		return nil, fmt.Errorf("no valid SSH sessions could be created")
	}

	fmt.Printf("Creating Fyne app and window...\n")
	// Create Fyne app and window
	a := app.New()
	w := a.NewWindow(title)
	w.Resize(fyne.NewSize(1000, 700))

	fmt.Printf("Creating terminal widget...\n")
	// Create terminal widget on UI thread
	var t *terminal.Terminal
	fyne.Do(func() {
		t = terminal.New()
	})

	fmt.Printf("Setting up multi-reader/writer...\n")
	// Create multi-writer for stdin (distributes input to all sessions)
	multiStdin := &multiWriter{writers: stdinWriters}

	// Create multi-reader for stdout (combines output from all sessions)
	multiStdout := newMultiReader(stdoutReaders)

	fmt.Printf("Starting %d SSH shell sessions...\n", len(sessions))
	// Track active sessions
	activeSessions := make(map[int]bool)
	for i := range sessions {
		activeSessions[i] = true
	}

	// Start SSH shell sessions with proper interactive shells
	for i, session := range sessions {
		go func(s *ssh.Session, host string, index int) {
			defer func() {
				// Mark session as inactive when it ends
				activeSessions[index] = false
			}()

			fmt.Printf("Starting interactive shell session for %s (session %d)\n", host, index)

			// Request a pseudo-terminal for interactive shell
			err := s.RequestPty("xterm", 80, 24, ssh.TerminalModes{
				ssh.ECHO:          1,     // enable echoing
				ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
				ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
			})
			if err != nil {
				fmt.Printf("Failed to request pty for %s: %v\n", host, err)
				return
			}

			// Start interactive shell
			err = s.Shell()
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

	fmt.Printf("Connecting terminal to SSH sessions...\n")
	// Connect terminal to the multi-reader/writer
	go func() {
		// Give shells a moment to start up
		time.Sleep(100 * time.Millisecond)

		defer func() {
			fmt.Printf("Terminal connection ended, quitting app\n")
			fyne.Do(func() {
				a.Quit()
			})
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

	// Add listener using fyne.Do to ensure UI thread safety
	fyne.Do(func() {
		t.AddListener(configChan)
	})

	fmt.Printf("Creating SSH multi-terminal instance...\n")
	// Create SSH multi-terminal instance
	sshMultiTerm := &SSHMultiTerminal{
		sessions:       sessions,
		connections:    validConnections,
		terminal:       t,
		window:         w,
		app:            a,
		stdinWriters:   stdinWriters,
		activeSessions: activeSessions,
		mutex:          sync.RWMutex{},
	}

	// Set window content using fyne.Do for UI thread safety
	fyne.Do(func() {
		w.SetContent(t)
	})

	fmt.Printf("SSH multi-terminal created successfully with %d sessions\n", len(sessions))
	return sshMultiTerm, nil
}

// ShowTerminal displays the SSH multi-terminal window
func (smt *SSHMultiTerminal) ShowTerminal() {
	fmt.Printf("ShowTerminal called - will block until window closes\n")
	fyne.Do(func() {
		smt.window.ShowAndRun()
	})
	fmt.Printf("ShowTerminal finished - window closed\n")
}

// ShowTerminalWindow displays the terminal window without running
func (smt *SSHMultiTerminal) ShowTerminalWindow() {
	fmt.Printf("ShowTerminalWindow called - non-blocking\n")
	fyne.Do(func() {
		smt.window.Show()
	})
	fmt.Printf("ShowTerminalWindow finished - window should be visible\n")
}

// Close closes all SSH sessions and the terminal
func (smt *SSHMultiTerminal) Close() error {
	smt.mutex.Lock()
	defer smt.mutex.Unlock()

	var errs []error

	// Close all SSH sessions
	for i, session := range smt.sessions {
		if session != nil {
			if err := session.Close(); err != nil {
				errs = append(errs, fmt.Errorf("failed to close SSH session for %s: %v",
					smt.connections[i].Config.Host, err))
			}
		}
	}

	// Close stdin writers
	for _, writer := range smt.stdinWriters {
		if writer != nil {
			if err := writer.Close(); err != nil {
				errs = append(errs, fmt.Errorf("failed to close stdin writer: %v", err))
			}
		}
	}

	// Close window
	if smt.window != nil {
		fyne.Do(func() {
			smt.window.Close()
		})
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing SSH multi-terminal: %v", errs)
	}

	return nil
}

// GetConnectedHosts returns a list of connected host names
func (smt *SSHMultiTerminal) GetConnectedHosts() []string {
	smt.mutex.RLock()
	defer smt.mutex.RUnlock()

	var hosts []string
	for _, conn := range smt.connections {
		hosts = append(hosts, conn.Config.Host)
	}
	return hosts
}

// GetSessionCount returns the number of active SSH sessions
func (smt *SSHMultiTerminal) GetSessionCount() int {
	smt.mutex.RLock()
	defer smt.mutex.RUnlock()

	return len(smt.sessions)
}
