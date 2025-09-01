// Package FyneCodeEditorWidget provides a powerful, customizable code editor widget for Fyne applications.
//
// This package implements a script editor with syntax highlighting, VS Code-style themes,
// and comprehensive editing capabilities. It supports multiple programming languages
// and allows for extensive customization of appearance and behavior.
//
// Key Features:
//   - Syntax highlighting for Go, Python, JavaScript, Java, C/C++, and more
//   - VS Code-style color themes with JSON support
//   - Customizable font size, tab size, and line numbers
//   - Full keyboard navigation and text editing
//   - Extensible architecture for adding new languages and themes
//
// Basic Usage:
//
//	editor := FyneCodeEditorWidget.NewScriptEditor()
//	editor.SetLanguage("go")
//	editor.SetTheme(FyneCodeEditorWidget.GetDefaultTheme())
//	editor.SetContent("package main\n\nfunc main() {\n\t// Your code here\n}")
//
// The widget follows Fyne's standard architecture with separate Widget and WidgetRenderer
// interfaces. The lexer provides tokenization for syntax highlighting, while the theme
// system allows for comprehensive color customization.
//
// For complete examples and documentation, see the README.md file and example.go.
package codeditor
