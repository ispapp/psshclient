package code

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
)

// KeyboardShortcuts handles VS Code-like keyboard shortcuts
type KeyboardShortcuts struct {
	editor *CodeEditor
}

// NewKeyboardShortcuts creates a new keyboard shortcut handler
func NewKeyboardShortcuts(editor *CodeEditor) *KeyboardShortcuts {
	return &KeyboardShortcuts{editor: editor}
}

// HandleKeyEvent processes keyboard events for shortcuts
func (ks *KeyboardShortcuts) HandleKeyEvent(event *fyne.KeyEvent) bool {
	// Check for modifier key combinations
	if event.Name == fyne.KeyTab {
		ks.handleTab(event)
		return true
	}

	// Handle Ctrl/Cmd combinations
	if ks.isModifierPressed(event) {
		return ks.handleModifierShortcuts(event)
	}

	return false
}

// isModifierPressed checks if Ctrl (Windows/Linux) or Cmd (Mac) is pressed
func (ks *KeyboardShortcuts) isModifierPressed(event *fyne.KeyEvent) bool {
	// Check for desktop key events to access modifier info
	if event.Name == desktop.KeyControlLeft || event.Name == desktop.KeyControlRight {
		return true
	}
	return false
}

// handleModifierShortcuts handles Ctrl/Cmd + key combinations
func (ks *KeyboardShortcuts) handleModifierShortcuts(event *fyne.KeyEvent) bool {
	switch event.Name {
	case fyne.KeyS:
		// Ctrl+S / Cmd+S - Save
		ks.editor.Save()
		return true

	case fyne.KeyZ:
		// Ctrl+Z / Cmd+Z - Undo
		ks.handleUndo()
		return true

	case fyne.KeyY:
		// Ctrl+Y / Cmd+Y - Redo
		ks.handleRedo()
		return true

	case fyne.KeyA:
		// Ctrl+A / Cmd+A - Select All
		ks.handleSelectAll()
		return true

	case fyne.KeyC:
		// Ctrl+C / Cmd+C - Copy
		ks.handleCopy()
		return true

	case fyne.KeyX:
		// Ctrl+X / Cmd+X - Cut
		ks.handleCut()
		return true

	case fyne.KeyV:
		// Ctrl+V / Cmd+V - Paste
		ks.handlePaste()
		return true

	case fyne.KeyF:
		// Ctrl+F / Cmd+F - Find
		ks.handleFind()
		return true

	case fyne.KeyH:
		// Ctrl+H / Cmd+H - Find and Replace
		ks.handleFindReplace()
		return true

	case fyne.KeyG:
		// Ctrl+G / Cmd+G - Go to Line
		ks.handleGoToLine()
		return true

	case fyne.KeyD:
		// Ctrl+D / Cmd+D - Duplicate Line
		ks.handleDuplicateLine()
		return true

	case fyne.KeyL:
		// Ctrl+L / Cmd+L - Select Line
		ks.handleSelectLine()
		return true

	case fyne.KeySlash:
		// Ctrl+/ / Cmd+/ - Toggle Comment
		ks.editor.CommentToggle()
		return true

	case fyne.KeyLeftBracket:
		// Ctrl+[ / Cmd+[ - Decrease Indent
		ks.editor.UnindentSelection()
		return true

	case fyne.KeyRightBracket:
		// Ctrl+] / Cmd+] - Increase Indent
		ks.editor.IndentSelection()
		return true

	case fyne.KeyEqual:
		// Ctrl++ / Cmd++ - Zoom In
		ks.handleZoomIn()
		return true

	case fyne.KeyMinus:
		// Ctrl+- / Cmd+- - Zoom Out
		ks.handleZoomOut()
		return true

	case fyne.Key0:
		// Ctrl+0 / Cmd+0 - Reset Zoom
		ks.handleZoomReset()
		return true

	case fyne.KeyUp:
		// Ctrl+Up / Cmd+Up - Move Line Up
		ks.handleMoveLineUp()
		return true

	case fyne.KeyDown:
		// Ctrl+Down / Cmd+Down - Move Line Down
		ks.handleMoveLineDown()
		return true
	}

	return false
}

// handleTab processes tab key for indentation
func (ks *KeyboardShortcuts) handleTab(event *fyne.KeyEvent) {
	// Check if Shift is pressed for unindent
	if ks.isShiftPressed(event) {
		ks.editor.UnindentSelection()
	} else {
		ks.editor.IndentSelection()
	}
}

// isShiftPressed checks if Shift key is pressed
func (ks *KeyboardShortcuts) isShiftPressed(event *fyne.KeyEvent) bool {
	if event.Name == desktop.KeyShiftLeft || event.Name == desktop.KeyShiftRight {
		return true
	}
	return false
}

// Individual shortcut handlers
func (ks *KeyboardShortcuts) handleUndo() {
	// Implementation for undo
	// This would require maintaining an undo stack
	fmt.Println("Undo triggered")
}

func (ks *KeyboardShortcuts) handleRedo() {
	// Implementation for redo
	fmt.Println("Redo triggered")
}

func (ks *KeyboardShortcuts) handleSelectAll() {
	// Select all text in the editor
	text := ks.editor.GetText()
	if len(text) > 0 {
		// Set selection to entire content
		ks.editor.content.SetText(text)
		// Move cursor to end
		ks.editor.content.CursorColumn = len(text)
	}
}

func (ks *KeyboardShortcuts) handleCopy() {
	// Copy selected text to clipboard
	// This would interface with the system clipboard
	fmt.Println("Copy triggered")
}

func (ks *KeyboardShortcuts) handleCut() {
	// Cut selected text to clipboard
	fmt.Println("Cut triggered")
}

func (ks *KeyboardShortcuts) handlePaste() {
	// Paste from clipboard
	fmt.Println("Paste triggered")
}

func (ks *KeyboardShortcuts) handleFind() {
	// Open find dialog
	fmt.Println("Find triggered")
}

func (ks *KeyboardShortcuts) handleFindReplace() {
	// Open find and replace dialog
	fmt.Println("Find and Replace triggered")
}

func (ks *KeyboardShortcuts) handleGoToLine() {
	// Open go to line dialog
	fmt.Println("Go to Line triggered")
}

func (ks *KeyboardShortcuts) handleDuplicateLine() {
	// Duplicate current line
	cursor := ks.editor.content.CursorRow
	text := ks.editor.GetText()
	lines := splitLinesKeepEmpty(text)

	if cursor < len(lines) {
		currentLine := lines[cursor]
		// Insert duplicate line below current line
		newLines := make([]string, 0, len(lines)+1)
		newLines = append(newLines, lines[:cursor+1]...)
		newLines = append(newLines, currentLine)
		newLines = append(newLines, lines[cursor+1:]...)

		ks.editor.SetText(joinLinesKeepEmpty(newLines))
	}
}

func (ks *KeyboardShortcuts) handleSelectLine() {
	// Select entire current line
	cursor := ks.editor.content.CursorRow
	text := ks.editor.GetText()
	lines := splitLinesKeepEmpty(text)

	if cursor < len(lines) {
		// Calculate start and end positions for the line
		var start int
		for i := 0; i < cursor; i++ {
			start += len(lines[i]) + 1 // +1 for newline
		}
		end := start + len(lines[cursor])

		// Set selection (this would need to be implemented in the editor)
		fmt.Printf("Select line %d from %d to %d\n", cursor+1, start, end)
	}
}

func (ks *KeyboardShortcuts) handleZoomIn() {
	// Increase font size
	currentSize := ks.editor.fontSize
	newSize := currentSize + 2
	if newSize <= 48 { // Max font size
		ks.editor.SetFontSize(newSize)
	}
}

func (ks *KeyboardShortcuts) handleZoomOut() {
	// Decrease font size
	currentSize := ks.editor.fontSize
	newSize := currentSize - 2
	if newSize >= 8 { // Min font size
		ks.editor.SetFontSize(newSize)
	}
}

func (ks *KeyboardShortcuts) handleZoomReset() {
	// Reset to default font size
	ks.editor.SetFontSize(14)
}

func (ks *KeyboardShortcuts) handleMoveLineUp() {
	// Move current line up
	cursor := ks.editor.content.CursorRow
	if cursor > 0 {
		text := ks.editor.GetText()
		lines := splitLinesKeepEmpty(text)

		if cursor < len(lines) {
			// Swap current line with line above
			lines[cursor], lines[cursor-1] = lines[cursor-1], lines[cursor]
			ks.editor.SetText(joinLinesKeepEmpty(lines))

			// Move cursor up with the line
			ks.editor.content.CursorRow = cursor - 1
		}
	}
}

func (ks *KeyboardShortcuts) handleMoveLineDown() {
	// Move current line down
	cursor := ks.editor.content.CursorRow
	text := ks.editor.GetText()
	lines := splitLinesKeepEmpty(text)

	if cursor < len(lines)-1 {
		// Swap current line with line below
		lines[cursor], lines[cursor+1] = lines[cursor+1], lines[cursor]
		ks.editor.SetText(joinLinesKeepEmpty(lines))

		// Move cursor down with the line
		ks.editor.content.CursorRow = cursor + 1
	}
}

// Additional utility functions for text manipulation

// GetSelectedText returns the currently selected text
func (ks *KeyboardShortcuts) GetSelectedText() string {
	// This would need to be implemented based on Fyne's selection handling
	return ks.editor.GetText()
}

// ReplaceSelectedText replaces the selected text with new text
func (ks *KeyboardShortcuts) ReplaceSelectedText(newText string) {
	// This would replace the currently selected text
	// currentText := ks.editor.GetText()
	ks.editor.SetText(newText)
	ks.editor.content.CursorColumn = len(newText)
	ks.editor.content.CursorRow = len(splitLinesKeepEmpty(newText)) - 1
	ks.editor.Refresh()
}

// InsertTextAtCursor inserts text at the current cursor position
func (ks *KeyboardShortcuts) InsertTextAtCursor(text string) {
	currentText := ks.editor.GetText()
	cursor := ks.editor.content.CursorRow
	lines := splitLinesKeepEmpty(currentText)

	if cursor < len(lines) {
		line := lines[cursor]
		col := ks.editor.content.CursorColumn
		if col <= len(line) {
			newLine := line[:col] + text + line[col:]
			lines[cursor] = newLine
			ks.editor.SetText(joinLinesKeepEmpty(lines))

			// Move cursor to end of inserted text
			ks.editor.content.CursorColumn = col + len(text)
		}
	}
}

// Local helpers to replace missing fyne utilities

func splitLinesKeepEmpty(text string) []string {
	return strings.Split(text, "\n")
}

func joinLinesKeepEmpty(lines []string) string {
	return strings.Join(lines, "\n")
}
