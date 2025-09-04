package code

import (
	"fmt"
	"image/color"
	"log"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
)

// Theme represents a color theme for the code editor
type Theme struct {
	Name           string
	Background     color.Color
	Foreground     color.Color
	Selection      color.Color
	LineNumber     color.Color
	CurrentLine    color.Color
	Comment        color.Color
	Keyword        color.Color
	String         color.Color
	Number         color.Color
	Function       color.Color
	Operator       color.Color
	Error          color.Color
	Warning        color.Color
	ScrollBarThumb color.Color
	ScrollBarTrack color.Color
}

// Predefined themes
var (
	DarkTheme = &Theme{
		Name:           "Dark",
		Background:     color.RGBA{30, 30, 30, 255},
		Foreground:     color.RGBA{212, 212, 212, 255},
		Selection:      color.RGBA{38, 79, 120, 255},
		LineNumber:     color.RGBA{133, 133, 133, 255},
		CurrentLine:    color.RGBA{40, 40, 40, 255},
		Comment:        color.RGBA{106, 153, 85, 255},
		Keyword:        color.RGBA{86, 156, 214, 255},
		String:         color.RGBA{206, 145, 120, 255},
		Number:         color.RGBA{181, 206, 168, 255},
		Function:       color.RGBA{220, 220, 170, 255},
		Operator:       color.RGBA{212, 212, 212, 255},
		Error:          color.RGBA{244, 71, 71, 255},
		Warning:        color.RGBA{255, 193, 7, 255},
		ScrollBarThumb: color.RGBA{79, 79, 79, 255},
		ScrollBarTrack: color.RGBA{51, 51, 51, 255},
	}

	LightTheme = &Theme{
		Name:           "Light",
		Background:     color.RGBA{255, 255, 255, 255},
		Foreground:     color.RGBA{0, 0, 0, 255},
		Selection:      color.RGBA{173, 214, 255, 255},
		LineNumber:     color.RGBA{149, 149, 149, 255},
		CurrentLine:    color.RGBA{245, 245, 245, 255},
		Comment:        color.RGBA{0, 128, 0, 255},
		Keyword:        color.RGBA{0, 0, 255, 255},
		String:         color.RGBA{163, 21, 21, 255},
		Number:         color.RGBA{9, 134, 88, 255},
		Function:       color.RGBA{121, 94, 38, 255},
		Operator:       color.RGBA{0, 0, 0, 255},
		Error:          color.RGBA{255, 0, 0, 255},
		Warning:        color.RGBA{255, 140, 0, 255},
		ScrollBarThumb: color.RGBA{196, 196, 196, 255},
		ScrollBarTrack: color.RGBA{240, 240, 240, 255},
	}

	MonokaiTheme = &Theme{
		Name:           "Monokai",
		Background:     color.RGBA{39, 40, 34, 255},
		Foreground:     color.RGBA{248, 248, 242, 255},
		Selection:      color.RGBA{73, 72, 62, 255},
		LineNumber:     color.RGBA{144, 144, 144, 255},
		CurrentLine:    color.RGBA{49, 50, 44, 255},
		Comment:        color.RGBA{117, 113, 94, 255},
		Keyword:        color.RGBA{249, 38, 114, 255},
		String:         color.RGBA{230, 219, 116, 255},
		Number:         color.RGBA{174, 129, 255, 255},
		Function:       color.RGBA{166, 226, 46, 255},
		Operator:       color.RGBA{249, 38, 114, 255},
		Error:          color.RGBA{244, 71, 71, 255},
		Warning:        color.RGBA{255, 193, 7, 255},
		ScrollBarThumb: color.RGBA{79, 79, 79, 255},
		ScrollBarTrack: color.RGBA{51, 51, 51, 255},
	}

	SolarizedDarkTheme = &Theme{
		Name:           "Solarized Dark",
		Background:     color.RGBA{0, 43, 54, 255},
		Foreground:     color.RGBA{131, 148, 150, 255},
		Selection:      color.RGBA{7, 54, 66, 255},
		LineNumber:     color.RGBA{88, 110, 117, 255},
		CurrentLine:    color.RGBA{7, 54, 66, 255},
		Comment:        color.RGBA{88, 110, 117, 255},
		Keyword:        color.RGBA{38, 139, 210, 255},
		String:         color.RGBA{42, 161, 152, 255},
		Number:         color.RGBA{211, 54, 130, 255},
		Function:       color.RGBA{181, 137, 0, 255},
		Operator:       color.RGBA{131, 148, 150, 255},
		Error:          color.RGBA{220, 50, 47, 255},
		Warning:        color.RGBA{203, 75, 22, 255},
		ScrollBarThumb: color.RGBA{79, 79, 79, 255},
		ScrollBarTrack: color.RGBA{51, 51, 51, 255},
	}
)

// Available themes
var AvailableThemes = []*Theme{
	DarkTheme,
	LightTheme,
	MonokaiTheme,
	SolarizedDarkTheme,
}

// CodeEditor represents the main code editor widget
type CodeEditor struct {
	widget.BaseWidget

	content     *widget.Entry
	richContent *widget.RichText
	lineNumbers *widget.Label
	scrollX     *container.Scroll
	scrollY     *container.Scroll
	container   *fyne.Container

	theme     *Theme
	language  string
	lexer     chroma.Lexer
	style     *chroma.Style
	formatter chroma.Formatter

	// Editor state
	// Future: Add selection tracking for advanced features
	// selection struct {
	//     start, end int
	// }
	isEditMode      bool   // true for editing (Entry), false for viewing (RichText)
	lastHighlighted string // cache to avoid rehighlighting same content

	// Settings
	tabSize         int
	showLineNumbers bool
	autoIndent      bool
	wordWrap        bool
	fontSize        float32

	// Callbacks
	onTextChanged func(string)
	onSave        func(string)
}

// NewCodeEditor creates a new code editor with the specified theme and language
func NewCodeEditor(theme *Theme, language string) *CodeEditor {
	editor := &CodeEditor{
		theme:           theme,
		language:        language,
		tabSize:         4,
		showLineNumbers: true,
		autoIndent:      true,
		wordWrap:        false,
		fontSize:        14,
	}

	editor.ExtendBaseWidget(editor)
	editor.setupSyntaxHighlighting()
	editor.createUI()
	editor.setupKeyboardShortcuts()

	return editor
}

// SetTheme changes the editor theme
func (e *CodeEditor) SetTheme(newTheme *Theme) {
	if newTheme != nil && (e.theme == nil || e.theme.Name != newTheme.Name) {
		log.Printf("Changing theme from %s to %s",
			func() string {
				if e.theme != nil {
					return e.theme.Name
				} else {
					return "none"
				}
			}(),
			newTheme.Name)
		e.theme = newTheme
		e.setupSyntaxHighlighting()
		e.Refresh()
	}
}

// SetLanguage changes the syntax highlighting language
func (e *CodeEditor) SetLanguage(language string) {
	e.language = language
	e.setupSyntaxHighlighting()
	e.highlightSyntax()
}

// SetText sets the editor content
func (e *CodeEditor) SetText(text string) {
	e.content.SetText(text)
	e.updateLineNumbers()
	e.highlightSyntax()
}

// GetText returns the editor content
func (e *CodeEditor) GetText() string {
	return e.content.Text
}

// SetOnTextChanged sets the callback for text changes
func (e *CodeEditor) SetOnTextChanged(callback func(string)) {
	e.onTextChanged = callback
}

// SetOnSave sets the callback for save action
func (e *CodeEditor) SetOnSave(callback func(string)) {
	e.onSave = callback
}

// SetTabSize sets the tab size
func (e *CodeEditor) SetTabSize(size int) {
	e.tabSize = size
}

// SetShowLineNumbers toggles line numbers display
func (e *CodeEditor) SetShowLineNumbers(show bool) {
	e.showLineNumbers = show
	if show {
		e.container = container.NewBorder(
			nil, nil,
			e.lineNumbers, nil,
			e.scrollY,
		)
	} else {
		e.container.Objects = []fyne.CanvasObject{e.scrollY}
	}
	e.Refresh()
}

// SetAutoIndent toggles auto-indentation
func (e *CodeEditor) SetAutoIndent(auto bool) {
	e.autoIndent = auto
}

// SetWordWrap toggles word wrapping
func (e *CodeEditor) SetWordWrap(wrap bool) {
	e.wordWrap = wrap
	// Update the entry's wrapping mode
	if wrap {
		e.content.Wrapping = fyne.TextWrapWord
	} else {
		e.content.Wrapping = fyne.TextWrapOff
	}
	e.Refresh()
}

// SetFontSize sets the font size
func (e *CodeEditor) SetFontSize(size float32) {
	e.fontSize = size
	// Apply font size to content
	e.content.TextStyle = fyne.TextStyle{}
	e.Refresh()
}

// Save triggers the save callback if set
func (e *CodeEditor) Save() {
	if e.onSave != nil {
		e.formatCode()
		e.onSave(e.content.Text)
	}
}

// setupSyntaxHighlighting initializes the syntax highlighter
func (e *CodeEditor) setupSyntaxHighlighting() {
	// Get lexer for the language
	e.lexer = lexers.Get(e.language)
	if e.lexer == nil {
		e.lexer = lexers.Fallback
	}
	e.lexer = chroma.Coalesce(e.lexer)

	// Set up style based on theme
	styleName := "github"
	if e.theme.Name == "Dark" || e.theme.Name == "Monokai" || e.theme.Name == "Solarized Dark" {
		styleName = "monokai"
	}

	e.style = styles.Get(styleName)
	if e.style == nil {
		e.style = styles.Fallback
	}

	// Create formatter
	e.formatter = formatters.Get(e.language)
	if e.formatter == nil {
		e.formatter = formatters.Fallback
	}

	// Remove the unnecessary nil check as `err` is not used or assigned
	log.Printf("Syntax highlighting setup completed")
}

// createUI builds the editor interface
func (e *CodeEditor) createUI() {
	// Create the main text entry for editing
	e.content = widget.NewMultiLineEntry()
	e.content.Wrapping = fyne.TextWrapOff
	e.content.OnChanged = func(text string) {
		e.updateLineNumbers()
		if e.onTextChanged != nil {
			e.onTextChanged(text)
		}
		// Clear the highlight cache when text changes
		e.lastHighlighted = ""
	}

	// Apply theme colors
	e.content.TextStyle = fyne.TextStyle{Monospace: true}

	// Create RichText widget for syntax highlighting display
	e.richContent = widget.NewRichText()

	// Create line numbers
	e.lineNumbers = widget.NewLabel("1")
	e.lineNumbers.TextStyle = fyne.TextStyle{Monospace: true}
	e.lineNumbers.Alignment = fyne.TextAlignTrailing

	// Create scrollable containers
	e.scrollX = container.NewHScroll(e.content)
	e.scrollY = container.NewVScroll(e.scrollX)

	// Start in edit mode
	e.isEditMode = true

	// Create the main container
	e.container = container.NewBorder(
		nil, nil,
		e.lineNumbers, nil,
		e.scrollY,
	)

	e.updateLineNumbers()
}

// setupKeyboardShortcuts configures VS Code-like shortcuts
func (e *CodeEditor) setupKeyboardShortcuts() {
	// This would be implemented with Fyne's key event handling
	// For now, we'll set up the basic structure
}

// updateLineNumbers updates the line number display
func (e *CodeEditor) updateLineNumbers() {
	if !e.showLineNumbers {
		return
	}

	lines := strings.Split(e.content.Text, "\n")
	lineCount := len(lines)

	var numbers []string
	for i := 1; i <= lineCount; i++ {
		numbers = append(numbers, fmt.Sprintf("%4d", i))
	}

	e.lineNumbers.SetText(strings.Join(numbers, "\n"))
}

// highlightSyntax applies syntax highlighting to the text
func (e *CodeEditor) highlightSyntax() {
	if e.lexer == nil || e.formatter == nil {
		return
	}

	text := e.content.Text
	if text == "" {
		return
	}

	// Tokenize the text using Chroma
	iterator, err := e.lexer.Tokenise(nil, text)
	if err != nil {
		log.Printf("Error tokenizing text: %v", err)
		e.applyThemeColors()
		return
	}

	// Convert tokens to segments with colors
	e.applyTokenizedHighlighting(iterator)
}

// applyTokenizedHighlighting applies syntax highlighting based on tokens
func (e *CodeEditor) applyTokenizedHighlighting(iterator chroma.Iterator) {
	// Apply basic theme colors to the entry widget
	// Since Fyne's Entry doesn't support rich text directly,
	// we'll apply the theme colors to the widget itself
	e.applyThemeColors()

	// In a more advanced implementation, you could:
	// 1. Convert the tokenized text to a RichText widget
	// 2. Use a custom renderer that supports syntax highlighting
	// 3. Create overlays for different token types

	// For now, we'll log the token types for debugging
	if log.Default() != nil {
		tokenCount := 0
		for token := iterator(); token != chroma.EOF; token = iterator() {
			tokenCount++
			if tokenCount > 10 { // Limit logging to prevent spam
				break
			}
		}
	}
}

// applyThemeColors applies the current theme colors to the editor
func (e *CodeEditor) applyThemeColors() {
	if e.theme == nil {
		return
	}

	// Apply monospace font to content entry
	e.content.TextStyle = fyne.TextStyle{Monospace: true}

	// Apply theme to line numbers
	if e.lineNumbers != nil {
		e.lineNumbers.TextStyle = fyne.TextStyle{Monospace: true}
	}

	// Log theme application only once per theme change, not on every highlight
	// The actual visual theming is limited by Fyne's Entry widget capabilities
}

// createRichTextFromTokens creates a RichText widget from tokenized content
// This is an advanced feature for future implementation - uncomment when needed
// TODO: Enable when implementing full syntax highlighting with RichText
/*
func (e *CodeEditor) createRichTextFromTokens(iterator chroma.Iterator) *widget.RichText {
	richText := widget.NewRichText()

	var segments []widget.RichTextSegment
	currentText := ""

	for token := iterator(); token != chroma.EOF; token = iterator() {
		tokenType := token.Type
		tokenValue := token.Value

		// Convert Chroma token types to colors
		colorName := e.getColorForTokenType(tokenType)

		if currentText != "" {
			// Add accumulated text as a segment
			segments = append(segments, &widget.TextSegment{
				Text: currentText,
				Style: widget.RichTextStyle{},
			})
			currentText = ""
		}

		// Add the token as a colored segment
		if colorName != "" {
			segments = append(segments, &widget.TextSegment{
				Text: tokenValue,
				Style: widget.RichTextStyle{
					ColorName: fyne.ThemeColorName(colorName),
				},
			})
		} else {
			segments = append(segments, &widget.TextSegment{
				Text: tokenValue,
				Style: widget.RichTextStyle{},
			})
		}
	}

	// Add any remaining text
	if currentText != "" {
		segments = append(segments, &widget.TextSegment{
			Text: currentText,
			Style: widget.RichTextStyle{},
		})
	}

	richText.Segments = segments
	return richText
}

// getColorForTokenType maps Chroma token types to theme colors
// TODO: Enable when implementing full syntax highlighting with RichText
func (e *CodeEditor) getColorForTokenType(tokenType chroma.TokenType) string {
	if e.theme == nil {
		return ""
	}

	// Map token types to Fyne theme color names
	switch {
	case tokenType.InCategory(chroma.Keyword):
		return "primary" // Map to keyword color
	case tokenType.InCategory(chroma.String):
		return "success" // Map to string color
	case tokenType.InCategory(chroma.Comment):
		return "disabled" // Map to comment color
	case tokenType.InCategory(chroma.Number):
		return "warning" // Map to number color
	case tokenType.InCategory(chroma.Name):
		if tokenType == chroma.NameFunction {
			return "primary" // Map to function color
		}
		return ""
	case tokenType.InCategory(chroma.Operator):
		return ""
	case tokenType.InCategory(chroma.Error):
		return "error"
	default:
		return ""
	}
}
*/ // formatCode formats the code using the lexer (format on save)
func (e *CodeEditor) formatCode() {
	text := e.content.Text

	// Basic formatting for common languages
	switch e.language {
	case "go":
		e.formatGo(text)
	case "python":
		e.formatPython(text)
	case "javascript", "typescript":
		e.formatJavaScript(text)
	case "json":
		e.formatJSON(text)
	default:
		e.formatGeneric(text)
	}
}

// formatGo applies Go-specific formatting
func (e *CodeEditor) formatGo(text string) {
	// Basic Go formatting - in a real implementation,
	// you'd use go/format or similar
	lines := strings.Split(text, "\n")
	var formatted []string
	indentLevel := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			formatted = append(formatted, "")
			continue
		}

		// Decrease indent for closing braces
		if strings.HasPrefix(trimmed, "}") {
			indentLevel--
		}

		// Apply indentation
		indent := strings.Repeat("\t", indentLevel)
		formatted = append(formatted, indent+trimmed)

		// Increase indent for opening braces
		if strings.HasSuffix(trimmed, "{") {
			indentLevel++
		}
	}

	e.content.SetText(strings.Join(formatted, "\n"))
}

// formatPython applies Python-specific formatting
func (e *CodeEditor) formatPython(text string) {
	// Basic Python formatting
	lines := strings.Split(text, "\n")
	var formatted []string
	indentLevel := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			formatted = append(formatted, "")
			continue
		}

		// Apply indentation (4 spaces for Python)
		indent := strings.Repeat("    ", indentLevel)
		formatted = append(formatted, indent+trimmed)

		// Increase indent for colon at end
		if strings.HasSuffix(trimmed, ":") {
			indentLevel++
		}

		// Handle dedentation keywords
		if isDeIndentKeyword(trimmed) {
			indentLevel--
			if indentLevel < 0 {
				indentLevel = 0
			}
		}
	}

	e.content.SetText(strings.Join(formatted, "\n"))
}

// formatJavaScript applies JavaScript-specific formatting
func (e *CodeEditor) formatJavaScript(text string) {
	// Similar to Go formatting but with different conventions
	e.formatGo(text) // Reuse Go formatting for basic structure
}

// formatJSON applies JSON formatting
func (e *CodeEditor) formatJSON(text string) {
	// Basic JSON formatting
	lines := strings.Split(text, "\n")
	var formatted []string
	indentLevel := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		// Decrease indent for closing brackets
		if strings.HasPrefix(trimmed, "}") || strings.HasPrefix(trimmed, "]") {
			indentLevel--
		}

		// Apply indentation (2 spaces for JSON)
		indent := strings.Repeat("  ", indentLevel)
		formatted = append(formatted, indent+trimmed)

		// Increase indent for opening brackets
		if strings.HasSuffix(trimmed, "{") || strings.HasSuffix(trimmed, "[") {
			indentLevel++
		}
	}

	e.content.SetText(strings.Join(formatted, "\n"))
}

// formatGeneric applies generic formatting
func (e *CodeEditor) formatGeneric(text string) {
	// Basic cleanup - remove extra whitespace, normalize line endings
	lines := strings.Split(text, "\n")
	var formatted []string

	for _, line := range lines {
		// Remove trailing whitespace but preserve leading whitespace
		formatted = append(formatted, strings.TrimRight(line, " \t"))
	}

	e.content.SetText(strings.Join(formatted, "\n"))
}

// isDeIndentKeyword checks if a line contains a Python dedent keyword
func isDeIndentKeyword(line string) bool {
	keywords := []string{"else:", "elif", "except:", "finally:", "return", "break", "continue"}
	for _, keyword := range keywords {
		if strings.HasPrefix(line, keyword) {
			return true
		}
	}
	return false
}

// CreateRenderer creates the widget renderer
func (e *CodeEditor) CreateRenderer() fyne.WidgetRenderer {
	return &codeEditorRenderer{
		editor:    e,
		container: e.container,
	}
}

// codeEditorRenderer implements the renderer for the code editor
type codeEditorRenderer struct {
	editor    *CodeEditor
	container *fyne.Container
}

func (r *codeEditorRenderer) Layout(size fyne.Size) {
	r.container.Resize(size)
}

func (r *codeEditorRenderer) MinSize() fyne.Size {
	return fyne.NewSize(400, 300)
}

func (r *codeEditorRenderer) Refresh() {
	r.container.Refresh()
}

func (r *codeEditorRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.container}
}

func (r *codeEditorRenderer) Destroy() {
	// Cleanup if needed
}

// Additional utility functions for advanced features

// IndentSelection indents the selected text or current line
func (e *CodeEditor) IndentSelection() {
	// Implementation for indenting selected text
	cursor := e.content.CursorRow
	text := e.content.Text
	lines := strings.Split(text, "\n")

	if cursor < len(lines) {
		if e.language == "python" {
			lines[cursor] = "    " + lines[cursor]
		} else {
			lines[cursor] = "\t" + lines[cursor]
		}
		e.content.SetText(strings.Join(lines, "\n"))
	}
}

// UnindentSelection unindents the selected text or current line
func (e *CodeEditor) UnindentSelection() {
	cursor := e.content.CursorRow
	text := e.content.Text
	lines := strings.Split(text, "\n")

	if cursor < len(lines) {
		line := lines[cursor]
		if strings.HasPrefix(line, "\t") {
			lines[cursor] = line[1:]
		} else if strings.HasPrefix(line, "    ") {
			lines[cursor] = line[4:]
		} else if strings.HasPrefix(line, "  ") {
			lines[cursor] = line[2:]
		}
		e.content.SetText(strings.Join(lines, "\n"))
	}
}

// CommentToggle toggles line comments for the current line or selection
func (e *CodeEditor) CommentToggle() {
	cursor := e.content.CursorRow
	text := e.content.Text
	lines := strings.Split(text, "\n")

	if cursor < len(lines) {
		line := lines[cursor]
		commentPrefix := getCommentPrefix(e.language)

		if strings.Contains(line, commentPrefix) {
			// Uncomment
			lines[cursor] = strings.Replace(line, commentPrefix+" ", "", 1)
			lines[cursor] = strings.Replace(lines[cursor], commentPrefix, "", 1)
		} else {
			// Comment
			// Find the first non-whitespace character
			trimmed := strings.TrimLeft(line, " \t")
			if trimmed != "" {
				whitespace := line[:len(line)-len(trimmed)]
				lines[cursor] = whitespace + commentPrefix + " " + trimmed
			}
		}
		e.content.SetText(strings.Join(lines, "\n"))
	}
}

// getCommentPrefix returns the comment prefix for a language
func getCommentPrefix(language string) string {
	switch language {
	case "go", "javascript", "typescript", "java", "c", "cpp", "rust":
		return "//"
	case "python", "ruby", "shell", "bash":
		return "#"
	case "sql":
		return "--"
	case "html", "xml":
		return "<!--"
	default:
		return "//"
	}
}

// FindAndReplace provides find and replace functionality
func (e *CodeEditor) FindAndReplace(find, replace string, replaceAll bool) int {
	text := e.content.Text
	var newText string
	count := 0

	if replaceAll {
		newText = strings.ReplaceAll(text, find, replace)
		count = strings.Count(text, find)
	} else {
		newText = strings.Replace(text, find, replace, 1)
		if strings.Contains(text, find) {
			count = 1
		}
	}

	e.content.SetText(newText)
	return count
}

// GoToLine moves the cursor to a specific line
func (e *CodeEditor) GoToLine(lineNumber int) {
	lines := strings.Split(e.content.Text, "\n")
	if lineNumber > 0 && lineNumber <= len(lines) {
		// Calculate the character position for the line
		var pos int
		for i := 0; i < lineNumber-1; i++ {
			pos += len(lines[i]) + 1 // +1 for newline
		}
		e.content.CursorColumn = 0
		e.content.CursorRow = lineNumber - 1
	}
}

// GetCurrentLineNumber returns the current line number
func (e *CodeEditor) GetCurrentLineNumber() int {
	return e.content.CursorRow + 1
}

// GetTotalLines returns the total number of lines
func (e *CodeEditor) GetTotalLines() int {
	return len(strings.Split(e.content.Text, "\n"))
}

// SetReadOnly sets the editor to read-only mode
func (e *CodeEditor) SetReadOnly(readonly bool) {
	if readonly {
		e.setViewMode()
	} else {
		e.setEditMode()
	}
}

// ToggleMode switches between edit and view modes
func (e *CodeEditor) ToggleMode() {
	if e.isEditMode {
		e.setViewMode()
	} else {
		e.setEditMode()
	}
}

// IsEditMode returns true if the editor is in edit mode
func (e *CodeEditor) IsEditMode() bool {
	return e.isEditMode
}

// setEditMode switches to editing mode (Entry widget)
func (e *CodeEditor) setEditMode() {
	if e.isEditMode {
		return
	}

	e.isEditMode = true

	// Sync content from RichText to Entry if needed
	if e.richContent != nil && e.content != nil {
		// The content should already be in sync, but ensure it
		e.updateMainContainer()
	}
}

// setViewMode switches to view mode (RichText widget with syntax highlighting)
func (e *CodeEditor) setViewMode() {
	if !e.isEditMode {
		return
	}

	e.isEditMode = false

	// Update RichText with highlighted content
	e.updateRichTextHighlighting()
	e.updateMainContainer()
}

// updateMainContainer updates the main container based on current mode
func (e *CodeEditor) updateMainContainer() {
	if e.container == nil {
		return
	}

	var mainWidget fyne.CanvasObject

	if e.isEditMode {
		// Use Entry for editing
		mainWidget = e.scrollY
	} else {
		// Use RichText for viewing with syntax highlighting
		if e.richContent != nil {
			richScroll := container.NewVScroll(e.richContent)
			mainWidget = richScroll
		} else {
			mainWidget = e.scrollY
		}
	}

	// Update container
	if e.showLineNumbers && e.lineNumbers != nil {
		e.container.Objects = []fyne.CanvasObject{
			container.NewBorder(nil, nil, e.lineNumbers, nil, mainWidget),
		}
	} else {
		e.container.Objects = []fyne.CanvasObject{mainWidget}
	}

	e.container.Refresh()
}

// updateRichTextHighlighting creates syntax-highlighted RichText content
func (e *CodeEditor) updateRichTextHighlighting() {
	text := e.content.Text

	// Avoid re-highlighting the same content
	if text == e.lastHighlighted && e.richContent != nil {
		return
	}

	if e.richContent == nil {
		e.richContent = widget.NewRichText()
	}

	// Generate highlighted segments
	segments := e.generateHighlightedSegments(text)
	e.richContent.Segments = segments
	e.lastHighlighted = text
}

// generateHighlightedSegments creates RichText segments with syntax highlighting
func (e *CodeEditor) generateHighlightedSegments(text string) []widget.RichTextSegment {
	if text == "" || e.lexer == nil {
		return []widget.RichTextSegment{
			&widget.TextSegment{
				Text:  text,
				Style: widget.RichTextStyle{TextStyle: fyne.TextStyle{Monospace: true}},
			},
		}
	}

	// Tokenize the text
	iterator, err := e.lexer.Tokenise(nil, text)
	if err != nil {
		log.Printf("Error tokenizing text: %v", err)
		return []widget.RichTextSegment{
			&widget.TextSegment{
				Text:  text,
				Style: widget.RichTextStyle{TextStyle: fyne.TextStyle{Monospace: true}},
			},
		}
	}

	var segments []widget.RichTextSegment

	for token := iterator(); token != chroma.EOF; token = iterator() {
		tokenType := token.Type
		tokenValue := token.Value

		// Get color for token type
		colorName := e.getTokenColor(tokenType)

		// Create segment
		segment := &widget.TextSegment{
			Text: tokenValue,
			Style: widget.RichTextStyle{
				TextStyle: fyne.TextStyle{Monospace: true},
			},
		}

		// Apply color if available
		if colorName != "" {
			segment.Style.ColorName = fyne.ThemeColorName(colorName)
		}

		segments = append(segments, segment)
	}

	return segments
}

// getTokenColor maps token types to Fyne theme color names
func (e *CodeEditor) getTokenColor(tokenType chroma.TokenType) string {
	switch {
	case tokenType.InCategory(chroma.Keyword):
		return "primary"
	case tokenType.InCategory(chroma.String):
		return "success"
	case tokenType.InCategory(chroma.Comment):
		return "disabled"
	case tokenType.InCategory(chroma.Number):
		return "warning"
	case tokenType.InCategory(chroma.Name):
		if tokenType == chroma.NameFunction {
			return "primary"
		}
		return ""
	case tokenType.InCategory(chroma.Error):
		return "error"
	default:
		return ""
	}
}

// Legacy interface implementations for compatibility
type CodeWidget interface {
	fyne.CanvasObject
	CreateRenderer() CodeWidgetRenderer
}

type CodeWidgetRenderer interface {
	Layout(fyne.Size)
	MinSize() fyne.Size
	Refresh()
	Objects() []fyne.CanvasObject
	Destroy()
}
