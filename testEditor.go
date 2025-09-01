package main

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	codeditor "github.com/ispapp/psshclient/pkg/codeditor"
)

// Example demonstrates how to use the FyneCodeEditorWidget
func Example() {
	// Create a new app and window
	myApp := app.New()
	myWindow := myApp.NewWindow("Code Editor Example")
	myWindow.Resize(fyne.NewSize(800, 600))

	// Create a new script editor
	editor := codeditor.NewScriptEditor()

	// Set some example Go code
	exampleCode := `package main

import (
	"fmt"
	"strings"
)

// Hello represents a greeting
type Hello struct {
	Name string
}

// SayHello prints a greeting message
func (h *Hello) SayHello() {
	fmt.Printf("Hello, %s!\n", h.Name)
}

func main() {
	// Create a new Hello instance
	h := Hello{Name: "World"}
	h.SayHello()
	
	// Example with strings
	text := "This is a test string"
	words := strings.Split(text, " ")
	
	for i, word := range words {
		fmt.Printf("Word %d: %s\n", i+1, word)
	}
	
	// Numbers and operations
	x := 42
	y := 3.14
	result := float64(x) * y
	
	fmt.Printf("Result: %.2f\n", result)
}`

	editor.SetContent(exampleCode)
	editor.SetLanguage("go")

	// Set up a callback for content changes
	editor.SetOnChanged(func(content string) {
		// Handle content changes here
		// For example, save to file or validate syntax
	})

	// Create theme selection buttons
	themeButtons := container.NewHBox(
		widget.NewButton("Default Dark", func() {
			editor.SetTheme(GetDefaultTheme())
		}),
		widget.NewButton("Light", func() {
			editor.SetTheme(GetLightTheme())
		}),
		widget.NewButton("VS Code Dark", func() {
			editor.SetTheme(GetVSCodeDarkTheme())
		}),
		widget.NewButton("Monokai", func() {
			editor.SetTheme(GetMonokaiTheme())
		}),
	)

	// Create language selection
	languageSelect := widget.NewSelect([]string{"go", "python", "javascript", "java", "c", "cpp"}, func(selected string) {
		editor.SetLanguage(selected)
	})
	languageSelect.SetSelected("go")

	// Create font size controls
	fontSizeLabel := widget.NewLabel("Font Size: 12")
	fontSizeSlider := widget.NewSlider(8, 24)
	fontSizeSlider.SetValue(12)
	fontSizeSlider.OnChanged = func(value float64) {
		editor.SetFontSize(float32(value))
		fontSizeLabel.SetText(fmt.Sprintf("Font Size: %.0f", value))
	}

	// Create line numbers toggle
	lineNumbersCheck := widget.NewCheck("Show Line Numbers", func(checked bool) {
		editor.SetShowLineNumbers(checked)
	})
	lineNumbersCheck.SetChecked(true)

	// Create tab size controls
	tabSizeLabel := widget.NewLabel("Tab Size: 4")
	tabSizeSlider := widget.NewSlider(2, 8)
	tabSizeSlider.SetValue(4)
	tabSizeSlider.OnChanged = func(value float64) {
		editor.SetTabSize(int(value))
		tabSizeLabel.SetText(fmt.Sprintf("Tab Size: %.0f", value))
	}

	// Create controls panel
	controls := container.NewVBox(
		widget.NewLabel("Themes:"),
		themeButtons,
		widget.NewSeparator(),
		container.NewHBox(widget.NewLabel("Language:"), languageSelect),
		widget.NewSeparator(),
		fontSizeLabel,
		fontSizeSlider,
		widget.NewSeparator(),
		lineNumbersCheck,
		widget.NewSeparator(),
		tabSizeLabel,
		tabSizeSlider,
	)

	// Create file operations
	newButton := widget.NewButton("New", func() {
		editor.SetContent("")
	})

	loadButton := widget.NewButton("Load Example", func() {
		// Load different examples based on language
		switch editor.GetLanguage() {
		case "python":
			pythonCode := `# Python Example
import os
import sys
from typing import List, Dict

class Calculator:
    """A simple calculator class"""
    
    def __init__(self, name: str):
        self.name = name
        self.history: List[str] = []
    
    def add(self, a: float, b: float) -> float:
        """Add two numbers"""
        result = a + b
        self.history.append(f"{a} + {b} = {result}")
        return result
    
    def multiply(self, a: float, b: float) -> float:
        """Multiply two numbers"""
        result = a * b
        self.history.append(f"{a} * {b} = {result}")
        return result

def main():
    calc = Calculator("My Calculator")
    
    # Perform some calculations
    result1 = calc.add(10, 5)
    result2 = calc.multiply(3, 7)
    
    print(f"Calculator: {calc.name}")
    print("History:")
    for entry in calc.history:
        print(f"  {entry}")

if __name__ == "__main__":
    main()`
			editor.SetContent(pythonCode)
		case "javascript":
			jsCode := `// JavaScript Example
class Calculator {
    constructor(name) {
        this.name = name;
        this.history = [];
    }
    
    add(a, b) {
        const result = a + b;
        this.history.push("" + a + " + " + b + " = " + result);
        return result;
    }
    
    multiply(a, b) {
        const result = a * b;
        this.history.push("" + a + " * " + b + " = " + result);
        return result;
    }
    
    getHistory() {
        return this.history;
    }
}

// Async function example
async function fetchData(url) {
    try {
        const response = await fetch(url);
        const data = await response.json();
        return data;
    } catch (error) {
        console.error('Error fetching data:', error);
        throw error;
    }
}

// Main function
function main() {
    const calc = new Calculator("My Calculator");
    
    // Perform calculations
    const result1 = calc.add(10, 5);
    const result2 = calc.multiply(3, 7);
    
    console.log("Calculator: " + calc.name);
    console.log('History:');
    calc.getHistory().forEach(entry => {
        console.log("  " + entry);
    });
}

main();`
			editor.SetContent(jsCode)
		default:
			editor.SetContent(exampleCode)
		}
	})

	fileOps := container.NewHBox(newButton, loadButton)

	// Layout everything
	topControls := container.NewVBox(fileOps, controls)
	content := container.NewBorder(topControls, nil, nil, nil, editor)

	myWindow.SetContent(content)
	myWindow.ShowAndRun()
}

// CreateCodeEditor is a convenience function to create a new code editor with default settings
func CreateCodeEditor(language string, theme *Theme) *ScriptEditor {
	editor := NewScriptEditor()

	if language != "" {
		editor.SetLanguage(language)
	}

	if theme != nil {
		editor.SetTheme(theme)
	} else {
		editor.SetTheme(GetDefaultTheme())
	}

	return editor
}

// CreateCodeEditorWithContent creates a new code editor with initial content
func CreateCodeEditorWithContent(content, language string, theme *Theme) *ScriptEditor {
	editor := CreateCodeEditor(language, theme)
	editor.SetContent(content)
	return editor
}
