package code

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

// EditorSettings represents the configuration for the code editor
type EditorSettings struct {
	// Appearance
	Theme      string  `json:"theme"`
	FontSize   float32 `json:"fontSize"`
	FontFamily string  `json:"fontFamily"`
	LineHeight float32 `json:"lineHeight"`

	// Editor behavior
	TabSize         int  `json:"tabSize"`
	InsertSpaces    bool `json:"insertSpaces"`
	AutoIndent      bool `json:"autoIndent"`
	WordWrap        bool `json:"wordWrap"`
	ShowLineNumbers bool `json:"showLineNumbers"`
	ShowWhitespace  bool `json:"showWhitespace"`

	// Advanced features
	AutoSave       bool `json:"autoSave"`
	AutoSaveDelay  int  `json:"autoSaveDelay"` // in seconds
	FormatOnSave   bool `json:"formatOnSave"`
	TrimWhitespace bool `json:"trimWhitespace"`

	// Language-specific settings
	LanguageSettings map[string]*LanguageConfig `json:"languageSettings"`

	// Shortcuts (customizable)
	KeyBindings map[string]string `json:"keyBindings"`
}

// LanguageConfig represents language-specific configuration
type LanguageConfig struct {
	TabSize      int    `json:"tabSize"`
	InsertSpaces bool   `json:"insertSpaces"`
	Formatter    string `json:"formatter"`
	Linter       string `json:"linter"`
	FilePattern  string `json:"filePattern"`
}

// DefaultSettings returns the default editor settings
func DefaultSettings() *EditorSettings {
	return &EditorSettings{
		Theme:           "Dark",
		FontSize:        14,
		FontFamily:      "monospace",
		LineHeight:      1.4,
		TabSize:         4,
		InsertSpaces:    true,
		AutoIndent:      true,
		WordWrap:        false,
		ShowLineNumbers: true,
		ShowWhitespace:  false,
		AutoSave:        false,
		AutoSaveDelay:   30,
		FormatOnSave:    true,
		TrimWhitespace:  true,
		LanguageSettings: map[string]*LanguageConfig{
			"go": {
				TabSize:      8,
				InsertSpaces: false, // Go uses tabs
				Formatter:    "gofmt",
				Linter:       "golint",
				FilePattern:  "*.go",
			},
			"python": {
				TabSize:      4,
				InsertSpaces: true,
				Formatter:    "autopep8",
				Linter:       "pylint",
				FilePattern:  "*.py",
			},
			"javascript": {
				TabSize:      2,
				InsertSpaces: true,
				Formatter:    "prettier",
				Linter:       "eslint",
				FilePattern:  "*.js",
			},
			"typescript": {
				TabSize:      2,
				InsertSpaces: true,
				Formatter:    "prettier",
				Linter:       "tslint",
				FilePattern:  "*.ts",
			},
			"json": {
				TabSize:      2,
				InsertSpaces: true,
				Formatter:    "json",
				Linter:       "jsonlint",
				FilePattern:  "*.json",
			},
			"yaml": {
				TabSize:      2,
				InsertSpaces: true,
				Formatter:    "yaml",
				Linter:       "yamllint",
				FilePattern:  "*.yml,*.yaml",
			},
			"html": {
				TabSize:      2,
				InsertSpaces: true,
				Formatter:    "prettier",
				Linter:       "htmlhint",
				FilePattern:  "*.html",
			},
			"css": {
				TabSize:      2,
				InsertSpaces: true,
				Formatter:    "prettier",
				Linter:       "csslint",
				FilePattern:  "*.css",
			},
			"sql": {
				TabSize:      2,
				InsertSpaces: true,
				Formatter:    "sql-formatter",
				Linter:       "sqlint",
				FilePattern:  "*.sql",
			},
		},
		KeyBindings: map[string]string{
			"save":          "Ctrl+S",
			"undo":          "Ctrl+Z",
			"redo":          "Ctrl+Y",
			"copy":          "Ctrl+C",
			"cut":           "Ctrl+X",
			"paste":         "Ctrl+V",
			"selectAll":     "Ctrl+A",
			"find":          "Ctrl+F",
			"replace":       "Ctrl+H",
			"gotoLine":      "Ctrl+G",
			"comment":       "Ctrl+/",
			"indent":        "Ctrl+]",
			"unindent":      "Ctrl+[",
			"duplicateLine": "Ctrl+D",
			"moveLinesUp":   "Ctrl+Up",
			"moveLinesDown": "Ctrl+Down",
			"zoomIn":        "Ctrl+=",
			"zoomOut":       "Ctrl+-",
			"zoomReset":     "Ctrl+0",
		},
	}
}

// SettingsManager handles loading and saving editor settings
type SettingsManager struct {
	settingsPath string
	settings     *EditorSettings
}

// NewSettingsManager creates a new settings manager
func NewSettingsManager() *SettingsManager {
	homeDir, _ := os.UserHomeDir()
	settingsPath := filepath.Join(homeDir, ".ispapp-editor", "settings.json")

	sm := &SettingsManager{
		settingsPath: settingsPath,
		settings:     DefaultSettings(),
	}

	// Try to load existing settings
	sm.Load()

	return sm
}

// Load loads settings from the configuration file
func (sm *SettingsManager) Load() error {
	if _, err := os.Stat(sm.settingsPath); os.IsNotExist(err) {
		// Settings file doesn't exist, use defaults and save them
		return sm.Save()
	}

	data, err := ioutil.ReadFile(sm.settingsPath)
	if err != nil {
		return fmt.Errorf("failed to read settings file: %v", err)
	}

	var settings EditorSettings
	if err := json.Unmarshal(data, &settings); err != nil {
		return fmt.Errorf("failed to parse settings file: %v", err)
	}

	sm.settings = &settings
	return nil
}

// Save saves the current settings to the configuration file
func (sm *SettingsManager) Save() error {
	// Ensure the settings directory exists
	settingsDir := filepath.Dir(sm.settingsPath)
	if err := os.MkdirAll(settingsDir, 0755); err != nil {
		return fmt.Errorf("failed to create settings directory: %v", err)
	}

	data, err := json.MarshalIndent(sm.settings, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %v", err)
	}

	if err := ioutil.WriteFile(sm.settingsPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write settings file: %v", err)
	}

	return nil
}

// GetSettings returns the current settings
func (sm *SettingsManager) GetSettings() *EditorSettings {
	return sm.settings
}

// UpdateSettings updates the settings and saves them
func (sm *SettingsManager) UpdateSettings(settings *EditorSettings) error {
	sm.settings = settings
	return sm.Save()
}

// GetLanguageConfig returns the configuration for a specific language
func (sm *SettingsManager) GetLanguageConfig(language string) *LanguageConfig {
	if config, exists := sm.settings.LanguageSettings[language]; exists {
		return config
	}

	// Return default config for unknown languages
	return &LanguageConfig{
		TabSize:      sm.settings.TabSize,
		InsertSpaces: sm.settings.InsertSpaces,
		Formatter:    "generic",
		Linter:       "",
		FilePattern:  "*",
	}
}

// SetLanguageConfig sets the configuration for a specific language
func (sm *SettingsManager) SetLanguageConfig(language string, config *LanguageConfig) {
	if sm.settings.LanguageSettings == nil {
		sm.settings.LanguageSettings = make(map[string]*LanguageConfig)
	}
	sm.settings.LanguageSettings[language] = config
}

// ApplySettingsToEditor applies the settings to a code editor instance
func (sm *SettingsManager) ApplySettingsToEditor(editor *CodeEditor) {
	settings := sm.settings

	// Apply theme
	for _, theme := range AvailableThemes {
		if theme.Name == settings.Theme {
			editor.SetTheme(theme)
			break
		}
	}

	// Apply basic settings
	editor.SetFontSize(settings.FontSize)
	editor.SetTabSize(settings.TabSize)
	editor.SetShowLineNumbers(settings.ShowLineNumbers)
	editor.SetAutoIndent(settings.AutoIndent)
	editor.SetWordWrap(settings.WordWrap)

	// Apply language-specific settings if available
	if editor.language != "" {
		langConfig := sm.GetLanguageConfig(editor.language)
		editor.SetTabSize(langConfig.TabSize)
	}
}

// ExportSettings exports settings to a JSON string
func (sm *SettingsManager) ExportSettings() (string, error) {
	data, err := json.MarshalIndent(sm.settings, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to export settings: %v", err)
	}
	return string(data), nil
}

// ImportSettings imports settings from a JSON string
func (sm *SettingsManager) ImportSettings(jsonData string) error {
	var settings EditorSettings
	if err := json.Unmarshal([]byte(jsonData), &settings); err != nil {
		return fmt.Errorf("failed to import settings: %v", err)
	}

	sm.settings = &settings
	return sm.Save()
}

// ResetToDefaults resets all settings to default values
func (sm *SettingsManager) ResetToDefaults() error {
	sm.settings = DefaultSettings()
	return sm.Save()
}

// Utility functions for specific setting categories

// GetThemeNames returns a list of available theme names
func GetThemeNames() []string {
	names := make([]string, len(AvailableThemes))
	for i, theme := range AvailableThemes {
		names[i] = theme.Name
	}
	return names
}

// GetSupportedLanguages returns a list of supported programming languages
func GetSupportedLanguages() []string {
	return []string{
		"go", "python", "javascript", "typescript", "java", "c", "cpp", "rust",
		"html", "css", "scss", "less", "json", "yaml", "xml", "markdown",
		"sql", "shell", "bash", "powershell", "dockerfile", "makefile",
		"php", "ruby", "perl", "lua", "r", "matlab", "kotlin", "swift",
		"dart", "scala", "clojure", "haskell", "erlang", "elixir",
	}
}

// ValidateSettings validates the settings for correctness
func (sm *SettingsManager) ValidateSettings() []string {
	var errors []string

	// Validate font size
	if sm.settings.FontSize < 8 || sm.settings.FontSize > 48 {
		errors = append(errors, "Font size must be between 8 and 48")
	}

	// Validate tab size
	if sm.settings.TabSize < 1 || sm.settings.TabSize > 16 {
		errors = append(errors, "Tab size must be between 1 and 16")
	}

	// Validate line height
	if sm.settings.LineHeight < 1.0 || sm.settings.LineHeight > 3.0 {
		errors = append(errors, "Line height must be between 1.0 and 3.0")
	}

	// Validate theme
	validTheme := false
	for _, theme := range AvailableThemes {
		if theme.Name == sm.settings.Theme {
			validTheme = true
			break
		}
	}
	if !validTheme {
		errors = append(errors, fmt.Sprintf("Unknown theme: %s", sm.settings.Theme))
	}

	// Validate auto-save delay
	if sm.settings.AutoSaveDelay < 1 || sm.settings.AutoSaveDelay > 300 {
		errors = append(errors, "Auto-save delay must be between 1 and 300 seconds")
	}

	return errors
}
