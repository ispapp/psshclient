package main

import (
	"log"

	"github.com/ispapp/psshclient/pkg/code"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// chage it to main to run as standalone app
func TEST() {
	// Create new Fyne application
	myApp := app.New()
	myWindow := myApp.NewWindow("Code Editor Demo")
	myWindow.Resize(fyne.NewSize(1000, 700))

	// Create code editor with dark theme and Go syntax highlighting
	editor := code.NewCodeEditor(code.DarkTheme, "go")

	// Set some sample Go code
	sampleCode := `package main

import (
	"fmt"
	"strings"
)

// Example function demonstrating Go syntax highlighting
func main() {
	message := "Hello, World!"
	fmt.Println(message)
	
	// String manipulation example
	words := strings.Split(message, " ")
	for i, word := range words {
		fmt.Printf("Word %d: %s\n", i, word)
	}
	
	// Number operations
	numbers := []int{1, 2, 3, 4, 5}
	sum := 0
	for _, num := range numbers {
		sum += num
	}
	fmt.Printf("Sum: %d\n", sum)
}

// Example struct
type Person struct {
	Name string
	Age  int
}

// Example method
func (p Person) Greet() string {
	return fmt.Sprintf("Hello, I'm %s and I'm %d years old", p.Name, p.Age)
}`

	editor.SetText(sampleCode)

	// Create theme selector
	themeSelect := widget.NewSelect([]string{"Dark", "Light", "Monokai", "Solarized Dark"}, func(value string) {
		var selectedTheme *code.Theme
		switch value {
		case "Dark":
			selectedTheme = code.DarkTheme
		case "Light":
			selectedTheme = code.LightTheme
		case "Monokai":
			selectedTheme = code.MonokaiTheme
		case "Solarized Dark":
			selectedTheme = code.SolarizedDarkTheme
		default:
			selectedTheme = code.DarkTheme
		}
		editor.SetTheme(selectedTheme)
	})
	themeSelect.SetSelected("Dark")

	// Create language selector
	languageSelect := widget.NewSelect([]string{"go", "javascript", "python", "bash", "yaml", "json"}, func(value string) {
		editor.SetLanguage(value)

		// Set sample code based on language
		var sampleCode string
		switch value {
		case "go":
			sampleCode = `package main

import "fmt"

func main() {
	fmt.Println("Hello, Go!")
}`
		case "javascript":
			sampleCode = `function hello() {
	console.log("Hello, JavaScript!");
	const numbers = [1, 2, 3, 4, 5];
	const sum = numbers.reduce((a, b) => a + b, 0);
	console.log("Sum:", sum);
}`
		case "python":
			sampleCode = `def hello():
	print("Hello, Python!")
	numbers = [1, 2, 3, 4, 5]
	total = sum(numbers)
	print(f"Sum: {total}")

if __name__ == "__main__":
	hello()`
		case "bash":
			sampleCode = `#!/bin/bash

echo "Hello, Bash!"
numbers=(1 2 3 4 5)
sum=0
for num in "${numbers[@]}"; do
	sum=$((sum + num))
done
echo "Sum: $sum"`
		case "yaml":
			sampleCode = `# YAML Configuration Example
app:
  name: "Code Editor Demo"
  version: "1.0.0"
  
database:
  host: "localhost"
  port: 5432
  name: "demo_db"
  
features:
  - syntax_highlighting
  - themes
  - line_numbers`
		case "json":
			sampleCode = `{
  "app": {
    "name": "Code Editor Demo",
    "version": "1.0.0"
  },
  "database": {
    "host": "localhost",
    "port": 5432,
    "name": "demo_db"
  },
  "features": [
    "syntax_highlighting",
    "themes",
    "line_numbers"
  ]
}`
		default:
			sampleCode = "// Select a language to see sample code"
		}
		editor.SetText(sampleCode)
	})
	languageSelect.SetSelected("go")

	// Create control buttons
	saveBtn := widget.NewButton("Save", func() {
		text := editor.GetText()
		log.Printf("Saving content: %d characters\n", len(text))
		// In a real application, you would save to file here
	})

	clearBtn := widget.NewButton("Clear", func() {
		editor.SetText("")
	})

	// Add mode label and toggle button for edit/view mode
	modeLabel := widget.NewLabel("Mode: Edit")
	var toggleBtn *widget.Button
	toggleBtn = widget.NewButton("View Mode (Syntax Highlight)", func() {
		editor.ToggleMode()
		// Update labels immediately
		if !editor.IsEditMode() {
			toggleBtn.SetText("Edit Mode")
			modeLabel.SetText("Mode: View (Highlighted)")
		} else {
			toggleBtn.SetText("View Mode (Syntax Highlight)")
			modeLabel.SetText("Mode: Edit")
		}
	})

	// Create toolbar
	toolbar := container.NewHBox(
		widget.NewLabel("Theme:"),
		themeSelect,
		widget.NewSeparator(),
		widget.NewLabel("Language:"),
		languageSelect,
		widget.NewSeparator(),
		modeLabel,
		toggleBtn,
		widget.NewSeparator(),
		saveBtn,
		clearBtn,
	)

	// Create main layout
	content := container.NewBorder(
		toolbar, // top
		nil,     // bottom
		nil,     // left
		nil,     // right
		editor,  // center
	)

	myWindow.SetContent(content)
	myWindow.ShowAndRun()
}
