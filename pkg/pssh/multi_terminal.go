package pssh

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// MultiDeviceTerminal represents a terminal that can send commands to multiple devices
type MultiDeviceTerminal struct {
	connections map[string]*SSHConnection
	output      *widget.RichText
	input       *widget.Entry
	window      fyne.Window
	app         fyne.App
	content     string // Store content manually
	mutex       sync.RWMutex
}

// CommandResult represents the result of executing a command on a device
type CommandResult struct {
	Hostname string
	Output   string
	Error    error
}

// NewMultiDeviceTerminal creates a new multi-device terminal
func NewMultiDeviceTerminal(connections []*SSHConnection, title string, parentApp fyne.App) (*MultiDeviceTerminal, error) {
	if len(connections) == 0 {
		return nil, fmt.Errorf("no connections provided")
	}

	// Create Fyne window using provided app
	w := parentApp.NewWindow(title)
	w.Resize(fyne.NewSize(1000, 700))

	// Create output widget
	output := widget.NewRichText()
	output.Wrapping = fyne.TextWrapWord

	// Create input widget
	input := widget.NewEntry()
	input.SetPlaceHolder("Enter command to execute on all selected devices...")

	// Create multi-device terminal
	mdt := &MultiDeviceTerminal{
		connections: make(map[string]*SSHConnection),
		output:      output,
		input:       input,
		window:      w,
		app:         parentApp,
	}

	// Store connections with hostname as key
	for _, conn := range connections {
		hostname := conn.Config.Host
		mdt.connections[hostname] = conn
	}

	// Set up input handling
	input.OnSubmitted = func(command string) {
		if strings.TrimSpace(command) != "" {
			mdt.executeCommand(command)
			input.SetText("")
		}
	}

	// Create scroll container for output
	outputScroll := container.NewScroll(output)
	outputScroll.SetMinSize(fyne.NewSize(900, 500))

	// Create layout
	content := container.NewBorder(
		widget.NewCard("Multi-Device SSH Terminal",
			fmt.Sprintf("Connected to %d devices", len(connections)),
			widget.NewLabel("")),
		container.NewBorder(nil, nil,
			widget.NewLabel("Command: "),
			widget.NewButton("Send", func() {
				if cmd := input.Text; strings.TrimSpace(cmd) != "" {
					mdt.executeCommand(cmd)
					input.SetText("")
				}
			}),
			input),
		nil, nil,
		outputScroll,
	)

	w.SetContent(content)

	// Set window close handler
	w.SetOnClosed(func() {
		// Close all SSH connections
		for _, conn := range mdt.connections {
			if conn != nil {
				conn.Close()
			}
		}
	})

	// Add welcome message
	deviceList := make([]string, 0, len(mdt.connections))
	for hostname := range mdt.connections {
		deviceList = append(deviceList, hostname)
	}

	welcomeMsg := fmt.Sprintf("Multi-Device SSH Terminal started.\nConnected to %d devices: %s\n\nType commands to execute on all devices simultaneously.\nSpecial commands:\n  - 'clear': Clear terminal output\n  - 'exit': Close terminal\n\n",
		len(deviceList), strings.Join(deviceList, ", "))

	mdt.appendOutput("SYSTEM", welcomeMsg, "green")

	return mdt, nil
}

// Show displays the terminal window
func (mdt *MultiDeviceTerminal) Show() {
	mdt.window.Show()
}

// executeCommand runs a command on all connected devices
func (mdt *MultiDeviceTerminal) executeCommand(command string) {
	command = strings.TrimSpace(command)

	// Handle special commands
	switch command {
	case "clear":
		mdt.mutex.Lock()
		mdt.content = ""
		mdt.output.ParseMarkdown("")
		mdt.mutex.Unlock()
		return
	case "exit":
		mdt.window.Close()
		return
	}

	// Show command being executed
	mdt.appendOutput("INPUT", fmt.Sprintf("Executing: %s", command), "blue")

	// Execute command on all devices in parallel
	results := make(chan CommandResult, len(mdt.connections))

	for hostname, conn := range mdt.connections {
		go func(h string, c *SSHConnection) {
			result := mdt.executeOnDevice(h, c, command)
			results <- result
		}(hostname, conn)
	}

	// Collect results
	for i := 0; i < len(mdt.connections); i++ {
		result := <-results
		if result.Error != nil {
			mdt.appendOutput(result.Hostname, fmt.Sprintf("Error: %v", result.Error), "red")
		} else {
			mdt.appendOutput(result.Hostname, result.Output, "black")
		}
	}

	mdt.appendOutput("SYSTEM", "Command completed on all devices.\n", "green")
}

// executeOnDevice runs a command on a specific device
func (mdt *MultiDeviceTerminal) executeOnDevice(hostname string, conn *SSHConnection, command string) CommandResult {
	if conn == nil || conn.Client == nil {
		return CommandResult{
			Hostname: hostname,
			Output:   "",
			Error:    fmt.Errorf("no active connection"),
		}
	}

	// Create session
	session, err := conn.Client.NewSession()
	if err != nil {
		return CommandResult{
			Hostname: hostname,
			Output:   "",
			Error:    fmt.Errorf("failed to create session: %v", err),
		}
	}
	defer session.Close()

	// Execute command
	output, err := session.CombinedOutput(command)

	return CommandResult{
		Hostname: hostname,
		Output:   string(output),
		Error:    err,
	}
}

// appendOutput adds formatted output to the terminal
func (mdt *MultiDeviceTerminal) appendOutput(hostname, text, color string) {
	mdt.mutex.Lock()
	defer mdt.mutex.Unlock()

	timestamp := time.Now().Format("15:04:05")

	var prefix string
	switch hostname {
	case "SYSTEM":
		prefix = fmt.Sprintf("[%s] [SYSTEM] ", timestamp)
	case "INPUT":
		prefix = fmt.Sprintf("[%s] [INPUT] ", timestamp)
	default:
		prefix = fmt.Sprintf("[%s] [%s] ", timestamp, hostname)
	}

	// Format the output based on color
	var formattedText string
	switch color {
	case "red":
		formattedText = fmt.Sprintf("**%s%s**\n", prefix, text)
	case "green":
		formattedText = fmt.Sprintf("*%s%s*\n", prefix, text)
	case "blue":
		formattedText = fmt.Sprintf("__%s%s__\n", prefix, text)
	default:
		formattedText = fmt.Sprintf("%s%s\n", prefix, text)
	}

	mdt.content += formattedText
	mdt.output.ParseMarkdown(mdt.content)

	// Auto-scroll to bottom by updating the scroll container
	// This is a workaround since we can't directly scroll RichText
	go func() {
		time.Sleep(10 * time.Millisecond)
		// The scroll will automatically go to bottom when content is updated
	}()
}
