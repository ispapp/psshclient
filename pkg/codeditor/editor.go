package codeditor

import (
	"image/color"
	"strings"
	"unicode"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

// TokenType represents different types of syntax tokens
type TokenType string

const (
	TokenKeyword    TokenType = "keyword"
	TokenString     TokenType = "string"
	TokenComment    TokenType = "comment"
	TokenNumber     TokenType = "number"
	TokenOperator   TokenType = "operator"
	TokenIdentifier TokenType = "identifier"
	TokenFunction   TokenType = "function"
	TokenType_      TokenType = "type"
	TokenPlain      TokenType = "plain"
)

// Token represents a syntax token with its type and value
type Token struct {
	Type  TokenType
	Value string
	Start int
	End   int
}

// Theme represents a color theme for syntax highlighting
type Theme struct {
	Name        string
	Background  color.Color
	Foreground  color.Color
	TokenColors map[TokenType]color.Color
}

// ScriptEditor is a custom Fyne widget for code editing with syntax highlighting
type ScriptEditor struct {
	widget.BaseWidget
	content   string
	language  string
	theme     *Theme
	fontSize  float32
	tabSize   int
	showLines bool
	cursor    int
	selection [2]int // start, end
	onChange  func(string)
}

// NewScriptEditor creates a new script editor widget
func NewScriptEditor() *ScriptEditor {
	editor := &ScriptEditor{
		language:  "go",
		theme:     GetDefaultTheme(),
		fontSize:  12.0,
		tabSize:   4,
		showLines: true,
		cursor:    0,
		selection: [2]int{-1, -1},
	}
	editor.ExtendBaseWidget(editor)
	return editor
}

// CreateRenderer creates the custom renderer for the script editor
func (e *ScriptEditor) CreateRenderer() fyne.WidgetRenderer {
	return newScriptEditorRenderer(e)
}

// SetContent sets the content of the editor
func (e *ScriptEditor) SetContent(content string) {
	e.content = content
	e.Refresh()
}

// GetContent returns the current content of the editor
func (e *ScriptEditor) GetContent() string {
	return e.content
}

// SetLanguage sets the programming language for syntax highlighting
func (e *ScriptEditor) SetLanguage(language string) {
	e.language = language
	e.Refresh()
}

// GetLanguage returns the current programming language
func (e *ScriptEditor) GetLanguage() string {
	return e.language
}

// SetTheme sets the color theme for the editor
func (e *ScriptEditor) SetTheme(theme *Theme) {
	e.theme = theme
	e.Refresh()
}

// GetTheme returns the current theme
func (e *ScriptEditor) GetTheme() *Theme {
	return e.theme
}

// SetFontSize sets the font size for the editor
func (e *ScriptEditor) SetFontSize(size float32) {
	e.fontSize = size
	e.Refresh()
}

// GetFontSize returns the current font size
func (e *ScriptEditor) GetFontSize() float32 {
	return e.fontSize
}

// SetTabSize sets the tab size for indentation
func (e *ScriptEditor) SetTabSize(size int) {
	e.tabSize = size
	e.Refresh()
}

// GetTabSize returns the current tab size
func (e *ScriptEditor) GetTabSize() int {
	return e.tabSize
}

// SetShowLineNumbers enables or disables line number display
func (e *ScriptEditor) SetShowLineNumbers(show bool) {
	e.showLines = show
	e.Refresh()
}

// GetShowLineNumbers returns whether line numbers are shown
func (e *ScriptEditor) GetShowLineNumbers() bool {
	return e.showLines
}

// SetOnChanged sets the callback function for content changes
func (e *ScriptEditor) SetOnChanged(callback func(string)) {
	e.onChange = callback
}

// insertText inserts text at the current cursor position
func (e *ScriptEditor) insertText(text string) {
	if e.cursor > len(e.content) {
		e.cursor = len(e.content)
	}

	newContent := e.content[:e.cursor] + text + e.content[e.cursor:]
	e.content = newContent
	e.cursor += len(text)

	if e.onChange != nil {
		e.onChange(e.content)
	}
	e.Refresh()
}

// deleteText deletes text before the cursor
func (e *ScriptEditor) deleteText(count int) {
	if e.cursor < count {
		count = e.cursor
	}

	newContent := e.content[:e.cursor-count] + e.content[e.cursor:]
	e.content = newContent
	e.cursor -= count

	if e.onChange != nil {
		e.onChange(e.content)
	}
	e.Refresh()
}

// moveCursor moves the cursor by the specified offset
func (e *ScriptEditor) moveCursor(offset int) {
	newPos := e.cursor + offset
	if newPos < 0 {
		newPos = 0
	} else if newPos > len(e.content) {
		newPos = len(e.content)
	}
	e.cursor = newPos
	e.Refresh()
}

// TypedRune handles typed characters
func (e *ScriptEditor) TypedRune(r rune) {
	if unicode.IsPrint(r) {
		e.insertText(string(r))
	}
}

// TypedKey handles special key presses
func (e *ScriptEditor) TypedKey(key *fyne.KeyEvent) {
	switch key.Name {
	case fyne.KeyBackspace:
		if e.cursor > 0 {
			e.deleteText(1)
		}
	case fyne.KeyDelete:
		if e.cursor < len(e.content) {
			e.content = e.content[:e.cursor] + e.content[e.cursor+1:]
			if e.onChange != nil {
				e.onChange(e.content)
			}
			e.Refresh()
		}
	case fyne.KeyReturn:
		e.insertText("\n")
	case fyne.KeyTab:
		e.insertText(strings.Repeat(" ", e.tabSize))
	case fyne.KeyLeft:
		e.moveCursor(-1)
	case fyne.KeyRight:
		e.moveCursor(1)
	case fyne.KeyUp:
		e.moveCursorUp()
	case fyne.KeyDown:
		e.moveCursorDown()
	case fyne.KeyHome:
		e.moveCursorToLineStart()
	case fyne.KeyEnd:
		e.moveCursorToLineEnd()
	}
}

// moveCursorUp moves cursor to the previous line
func (e *ScriptEditor) moveCursorUp() {
	lines := strings.Split(e.content[:e.cursor], "\n")
	if len(lines) > 1 {
		currentLinePos := len(lines[len(lines)-1])
		prevLineLen := 0
		if len(lines) > 1 {
			prevLineLen = len(lines[len(lines)-2])
		}

		newPos := e.cursor - currentLinePos - 1
		if currentLinePos > prevLineLen {
			newPos -= (currentLinePos - prevLineLen)
		}

		if newPos >= 0 {
			e.cursor = newPos
			e.Refresh()
		}
	}
}

// moveCursorDown moves cursor to the next line
func (e *ScriptEditor) moveCursorDown() {
	lines := strings.Split(e.content, "\n")
	currentLine := strings.Count(e.content[:e.cursor], "\n")

	if currentLine < len(lines)-1 {
		currentLineStart := strings.LastIndex(e.content[:e.cursor], "\n") + 1
		currentLinePos := e.cursor - currentLineStart

		nextLineStart := strings.Index(e.content[e.cursor:], "\n")
		if nextLineStart != -1 {
			nextLineStart += e.cursor + 1
			nextLineEnd := strings.Index(e.content[nextLineStart:], "\n")
			if nextLineEnd == -1 {
				nextLineEnd = len(e.content) - nextLineStart
			}

			newPos := nextLineStart + currentLinePos
			if currentLinePos > nextLineEnd {
				newPos = nextLineStart + nextLineEnd
			}

			if newPos <= len(e.content) {
				e.cursor = newPos
				e.Refresh()
			}
		}
	}
}

// moveCursorToLineStart moves cursor to the beginning of the current line
func (e *ScriptEditor) moveCursorToLineStart() {
	lineStart := strings.LastIndex(e.content[:e.cursor], "\n") + 1
	e.cursor = lineStart
	e.Refresh()
}

// moveCursorToLineEnd moves cursor to the end of the current line
func (e *ScriptEditor) moveCursorToLineEnd() {
	lineEnd := strings.Index(e.content[e.cursor:], "\n")
	if lineEnd == -1 {
		e.cursor = len(e.content)
	} else {
		e.cursor += lineEnd
	}
	e.Refresh()
}

// Focusable makes the editor focusable
func (e *ScriptEditor) Focusable() bool {
	return true
}
