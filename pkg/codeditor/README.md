# codeditor

A powerful, customizable code editor widget for Fyne applications with syntax highlighting and VS Code-style theme support.

## Features

- **Syntax Highlighting**: Support for multiple programming languages including Go, Python, JavaScript, Java, C/C++
- **VS Code-Style Themes**: Built-in themes (Dark, Light, VS Code Dark, Monokai) with JSON theme loading support
- **Customizable**: Font size, tab size, line numbers, and more
- **Keyboard Navigation**: Full cursor movement, text editing, and selection support
- **Extensible**: Easy to add new languages and themes

## Installation

```bash
go get github.com/ispapp/psshclient/pkg/codeditor
```

## Quick Start

```go
package main

import (
    "fyne.io/fyne/v2/app"
    editor "github.com/ispapp/psshclient/pkg/codeditor"
)

func main() {
    myApp := app.New()
    myWindow := myApp.NewWindow("Code Editor")
    
    // Create a new code editor
    codeEditor := editor.NewScriptEditor()
    codeEditor.SetLanguage("go")
    codeEditor.SetContent(`package main

import "fmt"

func main() {
    fmt.Println("Hello, World!")
}`)
    
    myWindow.SetContent(codeEditor)
    myWindow.ShowAndRun()
}
```

## API Reference

### Creating an Editor

```go
// Create a new editor with default settings
editor := NewScriptEditor()

// Create with specific language and theme
editor := CreateCodeEditor("python", GetVSCodeDarkTheme())

// Create with initial content
editor := CreateCodeEditorWithContent(code, "go", GetDefaultTheme())
```

### Content Management

```go
// Set and get content
editor.SetContent("your code here")
content := editor.GetContent()

// Listen for changes
editor.SetOnChanged(func(content string) {
    // Handle content changes
    fmt.Println("Content changed:", len(content), "characters")
})
```

### Language Support

```go
// Set programming language for syntax highlighting
editor.SetLanguage("go")        // Go
editor.SetLanguage("python")    // Python
editor.SetLanguage("javascript") // JavaScript
editor.SetLanguage("java")      // Java
editor.SetLanguage("c")         // C
editor.SetLanguage("cpp")       // C++

// Get current language
language := editor.GetLanguage()
```

### Theme Customization

```go
// Use built-in themes
editor.SetTheme(GetDefaultTheme())    // Dark theme
editor.SetTheme(GetLightTheme())      // Light theme
editor.SetTheme(GetVSCodeDarkTheme()) // VS Code Dark
editor.SetTheme(GetMonokaiTheme())    // Monokai

// Load theme from JSON file
theme, err := LoadThemeFromJSON("my-theme.json")
if err == nil {
    editor.SetTheme(theme)
}
```

### Appearance Settings

```go
// Font size
editor.SetFontSize(14.0)
size := editor.GetFontSize()

// Tab size
editor.SetTabSize(4)
tabSize := editor.GetTabSize()

// Line numbers
editor.SetShowLineNumbers(true)
showLines := editor.GetShowLineNumbers()
```

## Theme Format

Themes can be defined in JSON format:

```json
{
  "name": "My Custom Theme",
  "background": "#1e1e1e",
  "foreground": "#d4d4d4",
  "tokenColors": {
    "keyword": "#569cd6",
    "string": "#ce9178",
    "comment": "#6a9955",
    "number": "#b5cea8",
    "operator": "#d4d4d4",
    "identifier": "#9cdcfe",
    "function": "#dcdcaa",
    "type": "#4ec9b0",
    "plain": "#d4d4d4"
  }
}
```

### Saving Themes

```go
// Save a theme to JSON file
theme := GetMonokaiTheme()
err := SaveThemeToJSON(theme, "monokai-theme.json")
```

## Supported Languages

- **Go**: Full syntax highlighting with keywords, types, functions, and more
- **Python**: Support for Python 3 syntax including decorators, f-strings, and async/await
- **JavaScript**: ES6+ support with classes, arrow functions, template literals
- **Java**: Complete Java syntax highlighting
- **C/C++**: Support for both C and C++ with preprocessor directives
- **Generic**: Basic highlighting for unknown languages

## Keyboard Controls

- **Arrow Keys**: Navigate cursor
- **Home/End**: Move to line start/end
- **Backspace/Delete**: Delete characters
- **Tab**: Insert spaces (configurable tab size)
- **Enter**: Insert new line

## Examples

See the `example.go` file for a complete demonstration including:
- Theme switching
- Language selection
- Font size adjustment
- Line number toggle
- Different code samples

## Extending the Editor

### Adding New Languages

1. Add language-specific rules in `lexer.go`
2. Create a new `initialize[Language]Rules()` method
3. Add the language to the switch statement in `initializeRules()`

### Creating Custom Themes

1. Create a new `Theme` struct with your colors
2. Use `SaveThemeToJSON()` to export as JSON
3. Load with `LoadThemeFromJSON()` for reuse

## Architecture

The widget follows Fyne's architecture patterns:

- **ScriptEditor**: Main widget implementing `fyne.Widget`
- **scriptEditorRenderer**: Custom renderer implementing `fyne.WidgetRenderer`
- **Lexer**: Tokenizes code for syntax highlighting
- **Theme**: Defines colors for different token types

## Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Submit a pull request

## License

This project is licensed under the same license as the parent project.
