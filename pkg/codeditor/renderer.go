package codeditor

import (
	"fmt"
	"image/color"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
)

// scriptEditorRenderer implements the renderer for the script editor widget
type scriptEditorRenderer struct {
	editor      *ScriptEditor
	container   *container.Scroll
	textObjects []fyne.CanvasObject
	lineNumbers *canvas.Text
	background  *canvas.Rectangle
	cursor      *canvas.Rectangle
}

// newScriptEditorRenderer creates a new renderer for the script editor
func newScriptEditorRenderer(editor *ScriptEditor) *scriptEditorRenderer {
	background := canvas.NewRectangle(editor.theme.Background)

	cursor := canvas.NewRectangle(editor.theme.Foreground)
	cursor.Resize(fyne.NewSize(2, editor.fontSize+4))

	lineNumbers := canvas.NewText("", editor.theme.TokenColors[TokenComment])
	lineNumbers.TextStyle = fyne.TextStyle{Monospace: true}
	lineNumbers.TextSize = editor.fontSize

	content := container.NewMax()
	scroll := container.NewScroll(content)

	renderer := &scriptEditorRenderer{
		editor:      editor,
		container:   scroll,
		background:  background,
		cursor:      cursor,
		lineNumbers: lineNumbers,
	}

	renderer.refresh()
	return renderer
}

// Layout positions all the visual elements
func (r *scriptEditorRenderer) Layout(size fyne.Size) {
	r.background.Resize(size)
	r.container.Resize(size)

	// Position line numbers if enabled
	if r.editor.showLines {
		r.lineNumbers.Move(fyne.NewPos(5, 5))
	}

	// Update cursor position
	r.updateCursorPosition()
}

// MinSize returns the minimum size for the widget
func (r *scriptEditorRenderer) MinSize() fyne.Size {
	return fyne.NewSize(200, 100)
}

// Refresh updates the visual representation
func (r *scriptEditorRenderer) Refresh() {
	r.refresh()
}

// Objects returns all canvas objects that make up the widget
func (r *scriptEditorRenderer) Objects() []fyne.CanvasObject {
	objects := []fyne.CanvasObject{r.background, r.container}

	if r.editor.showLines {
		objects = append(objects, r.lineNumbers)
	}

	objects = append(objects, r.cursor)
	return objects
}

// Destroy cleans up the renderer
func (r *scriptEditorRenderer) Destroy() {
	// Clean up resources if needed
}

// BackgroundColor returns the background color for the widget
func (r *scriptEditorRenderer) BackgroundColor() color.Color {
	return r.editor.theme.Background
}

// refresh rebuilds the text objects with syntax highlighting
func (r *scriptEditorRenderer) refresh() {
	// Clear existing text objects
	r.textObjects = []fyne.CanvasObject{}

	// Tokenize the content
	tokens := r.tokenize(r.editor.content, r.editor.language)

	// Create line numbers if enabled
	if r.editor.showLines {
		r.updateLineNumbers()
	}

	// Create text objects for each token
	var x, y float32 = 0, 0
	lineNumberWidth := float32(0)

	if r.editor.showLines {
		lineNumberWidth = 40 // Space for line numbers
		x = lineNumberWidth
	}

	lineHeight := r.editor.fontSize + 4
	charWidth := r.editor.fontSize * 0.6 // Approximate character width for monospace

	for _, token := range tokens {
		color := r.editor.theme.TokenColors[token.Type]
		if color == nil {
			color = r.editor.theme.Foreground
		}

		// Handle newlines
		if strings.Contains(token.Value, "\n") {
			lines := strings.Split(token.Value, "\n")
			for i, line := range lines {
				if i > 0 {
					y += lineHeight
					x = lineNumberWidth
				}

				if line != "" {
					text := canvas.NewText(line, color)
					text.TextStyle = fyne.TextStyle{Monospace: true}
					text.TextSize = r.editor.fontSize
					text.Move(fyne.NewPos(x, y))
					r.textObjects = append(r.textObjects, text)
					x += float32(len(line)) * charWidth
				}
			}
		} else {
			text := canvas.NewText(token.Value, color)
			text.TextStyle = fyne.TextStyle{Monospace: true}
			text.TextSize = r.editor.fontSize
			text.Move(fyne.NewPos(x, y))
			r.textObjects = append(r.textObjects, text)
			x += float32(len(token.Value)) * charWidth
		}
	}

	// Update the container content
	content := container.NewMax(r.textObjects...)
	r.container.Content = content

	// Update cursor position
	r.updateCursorPosition()
}

// updateLineNumbers creates line number text
func (r *scriptEditorRenderer) updateLineNumbers() {
	lines := strings.Split(r.editor.content, "\n")
	lineText := ""

	for i := range lines {
		lineText += fmt.Sprintf("%3d\n", i+1)
	}

	r.lineNumbers.Text = strings.TrimSuffix(lineText, "\n")
	r.lineNumbers.Refresh()
}

// updateCursorPosition positions the cursor at the current cursor location
func (r *scriptEditorRenderer) updateCursorPosition() {
	if r.editor.cursor > len(r.editor.content) {
		return
	}

	// Calculate cursor position
	beforeCursor := r.editor.content[:r.editor.cursor]
	lines := strings.Split(beforeCursor, "\n")

	lineNumber := len(lines) - 1
	columnNumber := len(lines[len(lines)-1])

	lineHeight := r.editor.fontSize + 4
	charWidth := r.editor.fontSize * 0.6

	x := float32(columnNumber) * charWidth
	y := float32(lineNumber) * lineHeight

	if r.editor.showLines {
		x += 40 // Add line number width
	}

	r.cursor.Move(fyne.NewPos(x, y))
	r.cursor.Resize(fyne.NewSize(2, lineHeight))
}

// tokenize breaks down the content into tokens for syntax highlighting
func (r *scriptEditorRenderer) tokenize(content, language string) []Token {
	lexer := NewLexer(language)
	return lexer.Tokenize(content)
}
