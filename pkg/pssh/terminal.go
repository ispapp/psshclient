package pssh

import (
	"fmt"
	"io"
	"sync"

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

	// Create terminal widget
	t := terminal.New()

	// Start the SSH shell session
	go func() {
		err := session.Run("$SHELL || bash")
		if err != nil {
			fmt.Printf("SSH session ended: %v\n", err)
		}
	}()

	// Connect terminal to SSH session
	go func() {
		defer a.Quit()
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
	t.AddListener(configChan)

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

	// Set window content
	w.SetContent(t)

	// Store terminal widget
	tm.mutex.Lock()
	tm.terminals[conn.Config.Host] = termWidget
	tm.mutex.Unlock()

	return termWidget, nil
}

// ShowTerminal displays the terminal window
func (tw *TerminalWidget) ShowTerminal() {
	tw.Window.ShowAndRun()
}

// ShowTerminalWindow displays the terminal window without running (for multiple terminals)
func (tw *TerminalWidget) ShowTerminalWindow() {
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

		// Create terminal
		t := terminal.New()

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
		t.AddListener(configChan)

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

		// Add tab
		tabTitle := fmt.Sprintf("%s (%s)", conn.Config.Host, conn.Config.Username)
		tabs.Append(container.NewTabItem(tabTitle, t))
	}

	if len(tabs.Items) == 0 {
		return fmt.Errorf("no successful terminal connections created")
	}

	// Set window content and show
	w.SetContent(tabs)
	w.ShowAndRun()

	return nil
}
