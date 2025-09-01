package windows

import (
	"fmt"
	"sync"
	"time"

	"fyne.io/fyne/v2"
)

// WindowInfo holds information about a managed window
type WindowInfo struct {
	ID         string
	Title      string
	Window     fyne.Window
	CreatedAt  time.Time
	LastActive time.Time
	IsActive   bool
	WindowType string
	UserData   interface{} // Additional data that can be attached to the window
}

// WindowManager manages multiple application windows
type WindowManager struct {
	mu            sync.RWMutex
	windows       map[string]*WindowInfo
	app           fyne.App
	mainWindow    fyne.Window
	nextID        int
	onWindowClose func(windowID string) // Callback for when a window is closed
}

var WinManager WindowManager

// NewWindowManager creates a new window manager instance
func NewWindowManager(app fyne.App) *WindowManager {
	WinManager = WindowManager{
		windows: make(map[string]*WindowInfo),
		app:     app,
		nextID:  1,
	}
	return &WinManager
}

// SetMainWindow sets the main application window
func (wm *WindowManager) SetMainWindow(window fyne.Window) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	wm.mainWindow = window
}

// GetMainWindow returns the main application window
func (wm *WindowManager) GetMainWindow() fyne.Window {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	return wm.mainWindow
}

// SetOnWindowClose sets a callback function that will be called when a window is closed
func (wm *WindowManager) SetOnWindowClose(callback func(windowID string)) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	wm.onWindowClose = callback
}

// NewWindow creates and registers a new window
func (wm *WindowManager) NewWindow(title string, windowType string) (*WindowInfo, error) {
	if wm.app == nil {
		return nil, fmt.Errorf("no app instance available")
	}

	wm.mu.Lock()
	defer wm.mu.Unlock()

	// Generate unique window ID
	windowID := fmt.Sprintf("window-%d", wm.nextID)
	wm.nextID++

	// Create new window
	window := wm.app.NewWindow(title)

	// Create window info
	info := &WindowInfo{
		ID:         windowID,
		Title:      title,
		Window:     window,
		CreatedAt:  time.Now(),
		LastActive: time.Now(),
		IsActive:   false,
		WindowType: windowType,
	}

	// Set up window close callback
	window.SetCloseIntercept(func() {
		wm.closeWindow(windowID)
	})

	// Store window info
	wm.windows[windowID] = info

	return info, nil
}

// CloseWindow closes a window by ID
func (wm *WindowManager) CloseWindow(windowID string) error {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	return wm.closeWindow(windowID)
}

// closeWindow internal method to close a window (must be called with lock held)
func (wm *WindowManager) closeWindow(windowID string) error {
	info, exists := wm.windows[windowID]
	if !exists {
		return fmt.Errorf("window with ID %s not found", windowID)
	}

	// Close the actual window
	info.Window.Close()

	// Remove from our tracking
	delete(wm.windows, windowID)

	// Call the close callback if set
	if wm.onWindowClose != nil {
		go wm.onWindowClose(windowID)
	}

	return nil
}

// GetWindow returns window info by ID
func (wm *WindowManager) GetWindow(windowID string) (*WindowInfo, bool) {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	info, exists := wm.windows[windowID]
	if exists {
		// Update last active time
		info.LastActive = time.Now()
	}
	return info, exists
}

// ListWindows returns a list of all managed windows
func (wm *WindowManager) ListWindows() []*WindowInfo {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	windows := make([]*WindowInfo, 0, len(wm.windows))
	for _, info := range wm.windows {
		windows = append(windows, info)
	}

	return windows
}

// CountWindows returns the total number of managed windows
func (wm *WindowManager) CountWindows() int {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	return len(wm.windows)
}

// CountWindowsByType returns the number of windows of a specific type
func (wm *WindowManager) CountWindowsByType(windowType string) int {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	count := 0
	for _, info := range wm.windows {
		if info.WindowType == windowType {
			count++
		}
	}

	return count
}

// GetWindowsByType returns all windows of a specific type
func (wm *WindowManager) GetWindowsByType(windowType string) []*WindowInfo {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	windows := make([]*WindowInfo, 0)
	for _, info := range wm.windows {
		if info.WindowType == windowType {
			windows = append(windows, info)
		}
	}

	return windows
}

// CloseAllWindows closes all managed windows except the main window
func (wm *WindowManager) CloseAllWindows() error {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	var errors []string

	for windowID, info := range wm.windows {
		if info.Window != wm.mainWindow {
			if err := wm.closeWindow(windowID); err != nil {
				errors = append(errors, err.Error())
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors closing windows: %v", errors)
	}

	return nil
}

// CloseWindowsByType closes all windows of a specific type
func (wm *WindowManager) CloseWindowsByType(windowType string) error {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	var errors []string
	var windowsToClose []string

	// Collect windows to close
	for windowID, info := range wm.windows {
		if info.WindowType == windowType {
			windowsToClose = append(windowsToClose, windowID)
		}
	}

	// Close collected windows
	for _, windowID := range windowsToClose {
		if err := wm.closeWindow(windowID); err != nil {
			errors = append(errors, err.Error())
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors closing windows: %v", errors)
	}

	return nil
}

// ShowWindow brings a window to the front and makes it active
func (wm *WindowManager) ShowWindow(windowID string) error {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	info, exists := wm.windows[windowID]
	if !exists {
		return fmt.Errorf("window with ID %s not found", windowID)
	}

	info.Window.Show()
	info.Window.RequestFocus()
	info.IsActive = true
	info.LastActive = time.Now()

	return nil
}

// UpdateWindowTitle updates the title of a window
func (wm *WindowManager) UpdateWindowTitle(windowID string, newTitle string) error {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	info, exists := wm.windows[windowID]
	if !exists {
		return fmt.Errorf("window with ID %s not found", windowID)
	}

	info.Title = newTitle
	info.Window.SetTitle(newTitle)

	return nil
}

// GetWindowStats returns statistics about managed windows
func (wm *WindowManager) GetWindowStats() map[string]interface{} {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	stats := make(map[string]interface{})
	stats["total_windows"] = len(wm.windows)

	// Count by type
	typeCount := make(map[string]int)
	for _, info := range wm.windows {
		typeCount[info.WindowType]++
	}
	stats["windows_by_type"] = typeCount

	// Calculate average age
	if len(wm.windows) > 0 {
		totalAge := time.Duration(0)
		for _, info := range wm.windows {
			totalAge += time.Since(info.CreatedAt)
		}
		stats["average_age_seconds"] = totalAge.Seconds() / float64(len(wm.windows))
	}

	return stats
}
