package codeditor

import (
	"testing"

	"fyne.io/fyne/v2/test"
)

func TestNewScriptEditor(t *testing.T) {
	editor := NewScriptEditor()

	if editor == nil {
		t.Fatal("NewScriptEditor returned nil")
	}

	if editor.GetLanguage() != "go" {
		t.Errorf("Expected default language 'go', got '%s'", editor.GetLanguage())
	}

	if editor.GetContent() != "" {
		t.Errorf("Expected empty content, got '%s'", editor.GetContent())
	}

	if editor.GetFontSize() != 12.0 {
		t.Errorf("Expected default font size 12.0, got %f", editor.GetFontSize())
	}

	if editor.GetTabSize() != 4 {
		t.Errorf("Expected default tab size 4, got %d", editor.GetTabSize())
	}

	if !editor.GetShowLineNumbers() {
		t.Error("Expected line numbers to be shown by default")
	}
}

func TestSettersAndGetters(t *testing.T) {
	// Initialize a test app to avoid Fyne app errors
	_ = test.NewApp()

	editor := NewScriptEditor()

	// Test content
	testContent := "package main\n\nfunc main() {}"
	editor.SetContent(testContent)
	if editor.GetContent() != testContent {
		t.Errorf("Content not set correctly. Expected '%s', got '%s'", testContent, editor.GetContent())
	}

	// Test language
	editor.SetLanguage("python")
	if editor.GetLanguage() != "python" {
		t.Errorf("Language not set correctly. Expected 'python', got '%s'", editor.GetLanguage())
	}

	// Test font size
	editor.SetFontSize(16.0)
	if editor.GetFontSize() != 16.0 {
		t.Errorf("Font size not set correctly. Expected 16.0, got %f", editor.GetFontSize())
	}

	// Test tab size
	editor.SetTabSize(8)
	if editor.GetTabSize() != 8 {
		t.Errorf("Tab size not set correctly. Expected 8, got %d", editor.GetTabSize())
	}

	// Test line numbers
	editor.SetShowLineNumbers(false)
	if editor.GetShowLineNumbers() {
		t.Error("Line numbers should be hidden")
	}
}

func TestThemes(t *testing.T) {
	editor := NewScriptEditor()

	// Test default theme
	defaultTheme := GetDefaultTheme()
	if defaultTheme == nil {
		t.Fatal("GetDefaultTheme returned nil")
	}

	editor.SetTheme(defaultTheme)
	if editor.GetTheme() != defaultTheme {
		t.Error("Theme not set correctly")
	}

	// Test other built-in themes
	themes := GetBuiltinThemes()
	if len(themes) == 0 {
		t.Error("No built-in themes found")
	}

	for name, theme := range themes {
		if theme == nil {
			t.Errorf("Built-in theme '%s' is nil", name)
		}
		if theme.Name == "" {
			t.Errorf("Built-in theme '%s' has empty name", name)
		}
		if theme.TokenColors == nil {
			t.Errorf("Built-in theme '%s' has nil TokenColors", name)
		}
	}
}

func TestLexer(t *testing.T) {
	lexer := NewLexer("go")
	if lexer == nil {
		t.Fatal("NewLexer returned nil")
	}

	// Test tokenization of simple Go code
	code := `package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
}`

	tokens := lexer.Tokenize(code)
	if len(tokens) == 0 {
		t.Error("No tokens generated from Go code")
	}

	// Check for some expected token types
	foundKeyword := false
	foundString := false
	foundIdentifier := false

	for _, token := range tokens {
		switch token.Type {
		case TokenKeyword:
			foundKeyword = true
		case TokenString:
			foundString = true
		case TokenIdentifier:
			foundIdentifier = true
		}
	}

	if !foundKeyword {
		t.Error("No keyword tokens found in Go code")
	}
	if !foundString {
		t.Error("No string tokens found in Go code")
	}
	if !foundIdentifier {
		t.Error("No identifier tokens found in Go code")
	}
}

func TestCreateCodeEditor(t *testing.T) {
	// Test with language and theme
	theme := GetLightTheme()
	editor := CreateCodeEditor("python", theme)

	if editor.GetLanguage() != "python" {
		t.Errorf("Expected language 'python', got '%s'", editor.GetLanguage())
	}

	if editor.GetTheme() != theme {
		t.Error("Theme not set correctly")
	}

	// Test with content
	content := "print('Hello, World!')"
	editor2 := CreateCodeEditorWithContent(content, "python", theme)

	if editor2.GetContent() != content {
		t.Errorf("Content not set correctly. Expected '%s', got '%s'", content, editor2.GetContent())
	}
}

func TestHexColorParsing(t *testing.T) {
	// Test parseHexColor function
	testCases := []struct {
		input    string
		expected [3]uint8 // R, G, B
	}{
		{"#ff0000", [3]uint8{255, 0, 0}},     // Red
		{"#00ff00", [3]uint8{0, 255, 0}},     // Green
		{"#0000ff", [3]uint8{0, 0, 255}},     // Blue
		{"#ffffff", [3]uint8{255, 255, 255}}, // White
		{"#000000", [3]uint8{0, 0, 0}},       // Black
		{"#f00", [3]uint8{255, 0, 0}},        // Short form red
		{"fff", [3]uint8{255, 255, 255}},     // Short form white without #
	}

	for _, tc := range testCases {
		color := parseHexColor(tc.input)
		r, g, b, _ := color.RGBA()
		// Convert from 16-bit to 8-bit
		r8, g8, b8 := uint8(r>>8), uint8(g>>8), uint8(b>>8)

		if r8 != tc.expected[0] || g8 != tc.expected[1] || b8 != tc.expected[2] {
			t.Errorf("parseHexColor(%s) = RGB(%d,%d,%d), expected RGB(%d,%d,%d)",
				tc.input, r8, g8, b8, tc.expected[0], tc.expected[1], tc.expected[2])
		}
	}
}

func TestColorToHex(t *testing.T) {
	testCases := []struct {
		r, g, b  uint8
		expected string
	}{
		{255, 0, 0, "#ff0000"},     // Red
		{0, 255, 0, "#00ff00"},     // Green
		{0, 0, 255, "#0000ff"},     // Blue
		{255, 255, 255, "#ffffff"}, // White
		{0, 0, 0, "#000000"},       // Black
	}

	for _, tc := range testCases {
		color := parseHexColor(tc.expected)
		hex := colorToHex(color)

		if hex != tc.expected {
			t.Errorf("colorToHex(RGB(%d,%d,%d)) = %s, expected %s",
				tc.r, tc.g, tc.b, hex, tc.expected)
		}
	}
}
