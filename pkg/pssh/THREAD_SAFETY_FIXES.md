# SSH Terminal Thread Safety Fixes

## Problem
The SSH terminal was producing infinite "Fyne call thread" errors:
```
*** Error in Fyne call thread, this should have been called in fyne.Do[AndWait] ***
From: /Users/kimo/go/pkg/mod/github.com/fyne-io/terminal@v0.0.0-20250805210206-f3224d514e14/internal/widget/termgrid.go:58
```

## Root Cause
The fyne-io/terminal widget was trying to update its UI from background goroutines that weren't on the main UI thread. This happens when:
1. The terminal receives data from SSH and tries to refresh the display
2. Terminal operations are called before the widget is fully initialized
3. Multiple concurrent operations try to access the terminal UI simultaneously

## Solution Applied

### 1. **Synchronous Terminal Creation**
- Used channels to ensure terminal widgets are fully created before use
- Added `Resize()` call during creation to force internal structure initialization
- Wait for terminal creation to complete before proceeding

```go
var t *terminal.Terminal
done := make(chan struct{})
fyne.Do(func() {
    t = terminal.New()
    t.Resize(fyne.NewSize(1000, 700)) // Force initialization
    close(done)
})
<-done // Wait for completion
```

### 2. **Proper Window Content Setup**
- Set window content BEFORE connecting terminal to SSH
- Use synchronous channel-based waiting instead of time delays
- Ensure UI is ready before data starts flowing

```go
contentReady := make(chan struct{})
fyne.Do(func() {
    w.SetContent(t)
    close(contentReady)
})
<-contentReady
```

### 3. **Listener Registration Safety**
- Added synchronous waiting for listener registration
- Ensures resize handlers are properly set up before use

```go
listenerReady := make(chan struct{})
fyne.Do(func() {
    t.AddListener(configChan)
    close(listenerReady)
})
<-listenerReady
```

### 4. **Simplified Connection Flow**
- Removed nested `fyne.Do` calls that were causing race conditions
- Direct `RunWithConnection` call after proper initialization
- Better error handling and cleanup

## Key Changes Made

### Files Modified:
- `terminal.go`: Complete rewrite of SSH terminal connection logic

### Functions Fixed:
1. **`NewSSHMultiTerminal()`**: Added synchronous terminal creation and setup
2. **`CreateTerminalWidget()`**: Fixed single terminal widget creation
3. **`MultiTerminalWindow()`**: Fixed tabbed terminal creation

### Patterns Applied:
- **Synchronous Initialization**: Use channels to wait for UI setup completion
- **Sequential Setup**: Window content → Terminal connection → Listener registration
- **Clean Error Handling**: Proper cleanup on connection failures

## Expected Results

✅ **No more "Fyne call thread" errors**  
✅ **Stable terminal display and interaction**  
✅ **Proper SSH session handling**  
✅ **Clean banner suppression working**  
✅ **Visible input/output in terminal**  

## Testing the Fix

1. Run your SSH multi-terminal application
2. Connect to one or more SSH servers
3. Verify no error messages about Fyne call threads
4. Terminal should display properly without flickering or crashes
5. Input should be visible and commands should execute normally

## Technical Notes

The key insight was that the fyne-io/terminal widget has internal goroutines that handle data reading and UI updates. These need to be properly synchronized with the main UI thread. By ensuring:

1. Terminal is fully initialized before data connection
2. Window content is set before terminal starts processing
3. All setup operations complete before data flows

We eliminate the race conditions that caused the thread safety errors.

This approach follows the fyne-io/terminal best practices and ensures stable operation even with multiple SSH connections.
