package code

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// Test what text styling options are available in Fyne
func TestFyneTextStyling(t *testing.T) {
	// Create a test app
	testApp := app.New()
	defer testApp.Quit()

	t.Run("Entry Widget Styling", func(t *testing.T) {
		entry := widget.NewEntry()
		entry.SetText("func main() {\n    fmt.Println(\"Hello\")\n}")

		// Test different text styles
		entry.TextStyle = fyne.TextStyle{
			Bold:      false,
			Italic:    false,
			Monospace: true,
		}

		t.Logf("Entry supports: Bold, Italic, Monospace")
		t.Logf("Entry does NOT support: Color, Background color")
	})

	t.Run("RichText Widget Capabilities", func(t *testing.T) {
		richText := widget.NewRichText()

		// Test creating segments with different styles
		segments := []widget.RichTextSegment{
			&widget.TextSegment{
				Text: "func ",
				Style: widget.RichTextStyle{
					ColorName: theme.ColorNamePrimary,
					TextStyle: fyne.TextStyle{Monospace: true},
				},
			},
			&widget.TextSegment{
				Text: "main",
				Style: widget.RichTextStyle{
					ColorName: theme.ColorNameSuccess,
					TextStyle: fyne.TextStyle{Monospace: true},
				},
			},
			&widget.TextSegment{
				Text: "() {\n    fmt.Println(",
				Style: widget.RichTextStyle{
					TextStyle: fyne.TextStyle{Monospace: true},
				},
			},
			&widget.TextSegment{
				Text: "\"Hello\"",
				Style: widget.RichTextStyle{
					ColorName: theme.ColorNameWarning,
					TextStyle: fyne.TextStyle{Monospace: true},
				},
			},
			&widget.TextSegment{
				Text: ")\n}",
				Style: widget.RichTextStyle{
					TextStyle: fyne.TextStyle{Monospace: true},
				},
			},
		}

		richText.Segments = segments

		t.Logf("RichText supports: Multiple colors, styles per segment")
		t.Logf("RichText limitation: Not editable like Entry")
	})

	t.Run("Custom Theme Colors", func(t *testing.T) {
		// Test available theme colors
		availableColors := []fyne.ThemeColorName{
			theme.ColorNamePrimary,
			theme.ColorNameBackground,
			theme.ColorNameForeground,
			theme.ColorNameSuccess,
			theme.ColorNameWarning,
			theme.ColorNameError,
			theme.ColorNameDisabled,
		}

		for _, colorName := range availableColors {
			t.Logf("Color available: %s", colorName)
		}
	})

	t.Run("Container Overlay Test", func(t *testing.T) {
		// Test if we can overlay colored labels on top of entry
		entry := widget.NewEntry()
		entry.SetText("func main() { fmt.Println(\"Hello\") }")

		// Create colored overlays
		keywordLabel := widget.NewLabel("func")
		keywordLabel.TextStyle = fyne.TextStyle{Monospace: true}

		// This would require precise positioning
		overlay := container.NewWithoutLayout(entry, keywordLabel)

		t.Logf("Overlay approach: Possible but complex positioning required")
		_ = overlay
	})
}

// Test the current editor highlighting functionality
func TestCodeEditorHighlighting(t *testing.T) {
	editor := NewCodeEditor(DarkTheme, "go")

	// Test setting text and highlighting
	sampleCode := `func main() {
    fmt.Println("Hello, World!")
}`

	editor.SetText(sampleCode)

	// Check if lexer was set up correctly
	if editor.lexer == nil {
		t.Error("Lexer should be initialized")
	}

	if editor.formatter == nil {
		t.Error("Formatter should be initialized")
	}

	// Test theme switching
	originalTheme := editor.theme.Name
	editor.SetTheme(LightTheme)

	if editor.theme.Name == originalTheme {
		t.Error("Theme should have changed")
	}

	t.Logf("Editor theme changed from %s to %s", originalTheme, editor.theme.Name)
}

// Benchmark highlighting performance
func BenchmarkHighlighting(b *testing.B) {
	editor := NewCodeEditor(DarkTheme, "go")

	sampleCode := `package main

import (
    "fmt"
    "strings"
)

func main() {
    message := "Hello, World!"
    fmt.Println(message)
    
    words := strings.Split(message, " ")
    for i, word := range words {
        fmt.Printf("Word %d: %s\n", i, word)
    }
}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		editor.SetText(sampleCode)
	}
}
