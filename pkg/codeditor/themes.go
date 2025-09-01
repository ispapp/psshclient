package codeditor

import (
	"encoding/json"
	"image/color"
	"os"
)

// GetDefaultTheme returns the default dark theme for the editor
func GetDefaultTheme() *Theme {
	return &Theme{
		Name:       "Default Dark",
		Background: color.NRGBA{R: 30, G: 30, B: 30, A: 255},    // Dark background
		Foreground: color.NRGBA{R: 212, G: 212, B: 212, A: 255}, // Light text
		TokenColors: map[TokenType]color.Color{
			TokenKeyword:    color.NRGBA{R: 86, G: 156, B: 214, A: 255},  // Blue
			TokenString:     color.NRGBA{R: 206, G: 145, B: 120, A: 255}, // Orange
			TokenComment:    color.NRGBA{R: 106, G: 153, B: 85, A: 255},  // Green
			TokenNumber:     color.NRGBA{R: 181, G: 206, B: 168, A: 255}, // Light green
			TokenOperator:   color.NRGBA{R: 212, G: 212, B: 212, A: 255}, // White
			TokenIdentifier: color.NRGBA{R: 156, G: 220, B: 254, A: 255}, // Light blue
			TokenFunction:   color.NRGBA{R: 220, G: 220, B: 170, A: 255}, // Yellow
			TokenType_:      color.NRGBA{R: 78, G: 201, B: 176, A: 255},  // Cyan
			TokenPlain:      color.NRGBA{R: 212, G: 212, B: 212, A: 255}, // White
		},
	}
}

// GetLightTheme returns a light theme for the editor
func GetLightTheme() *Theme {
	return &Theme{
		Name:       "Light",
		Background: color.NRGBA{R: 255, G: 255, B: 255, A: 255}, // White background
		Foreground: color.NRGBA{R: 0, G: 0, B: 0, A: 255},       // Black text
		TokenColors: map[TokenType]color.Color{
			TokenKeyword:    color.NRGBA{R: 0, G: 0, B: 255, A: 255},    // Blue
			TokenString:     color.NRGBA{R: 163, G: 21, B: 21, A: 255},  // Dark red
			TokenComment:    color.NRGBA{R: 0, G: 128, B: 0, A: 255},    // Green
			TokenNumber:     color.NRGBA{R: 9, G: 134, B: 88, A: 255},   // Dark green
			TokenOperator:   color.NRGBA{R: 0, G: 0, B: 0, A: 255},      // Black
			TokenIdentifier: color.NRGBA{R: 0, G: 0, B: 0, A: 255},      // Black
			TokenFunction:   color.NRGBA{R: 121, G: 94, B: 38, A: 255},  // Brown
			TokenType_:      color.NRGBA{R: 43, G: 145, B: 175, A: 255}, // Teal
			TokenPlain:      color.NRGBA{R: 0, G: 0, B: 0, A: 255},      // Black
		},
	}
}

// GetVSCodeDarkTheme returns a VS Code dark theme
func GetVSCodeDarkTheme() *Theme {
	return &Theme{
		Name:       "VS Code Dark",
		Background: color.NRGBA{R: 30, G: 30, B: 30, A: 255},    // VS Code dark background
		Foreground: color.NRGBA{R: 212, G: 212, B: 212, A: 255}, // Light text
		TokenColors: map[TokenType]color.Color{
			TokenKeyword:    color.NRGBA{R: 197, G: 134, B: 192, A: 255}, // Purple
			TokenString:     color.NRGBA{R: 206, G: 145, B: 120, A: 255}, // Orange
			TokenComment:    color.NRGBA{R: 106, G: 153, B: 85, A: 255},  // Green
			TokenNumber:     color.NRGBA{R: 181, G: 206, B: 168, A: 255}, // Light green
			TokenOperator:   color.NRGBA{R: 212, G: 212, B: 212, A: 255}, // White
			TokenIdentifier: color.NRGBA{R: 156, G: 220, B: 254, A: 255}, // Light blue
			TokenFunction:   color.NRGBA{R: 220, G: 220, B: 170, A: 255}, // Yellow
			TokenType_:      color.NRGBA{R: 78, G: 201, B: 176, A: 255},  // Cyan
			TokenPlain:      color.NRGBA{R: 212, G: 212, B: 212, A: 255}, // White
		},
	}
}

// GetMonokaiTheme returns a Monokai theme
func GetMonokaiTheme() *Theme {
	return &Theme{
		Name:       "Monokai",
		Background: color.NRGBA{R: 39, G: 40, B: 34, A: 255},    // Monokai background
		Foreground: color.NRGBA{R: 248, G: 248, B: 242, A: 255}, // Light text
		TokenColors: map[TokenType]color.Color{
			TokenKeyword:    color.NRGBA{R: 249, G: 38, B: 114, A: 255},  // Pink
			TokenString:     color.NRGBA{R: 230, G: 219, B: 116, A: 255}, // Yellow
			TokenComment:    color.NRGBA{R: 117, G: 113, B: 94, A: 255},  // Gray
			TokenNumber:     color.NRGBA{R: 174, G: 129, B: 255, A: 255}, // Purple
			TokenOperator:   color.NRGBA{R: 248, G: 248, B: 242, A: 255}, // White
			TokenIdentifier: color.NRGBA{R: 248, G: 248, B: 242, A: 255}, // White
			TokenFunction:   color.NRGBA{R: 166, G: 226, B: 46, A: 255},  // Green
			TokenType_:      color.NRGBA{R: 102, G: 217, B: 239, A: 255}, // Cyan
			TokenPlain:      color.NRGBA{R: 248, G: 248, B: 242, A: 255}, // White
		},
	}
}

// ThemeDefinition represents a theme definition for JSON loading
type ThemeDefinition struct {
	Name        string            `json:"name"`
	Background  string            `json:"background"`
	Foreground  string            `json:"foreground"`
	TokenColors map[string]string `json:"tokenColors"`
}

// LoadThemeFromJSON loads a theme from a JSON file
func LoadThemeFromJSON(filepath string) (*Theme, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	var themeDef ThemeDefinition
	if err := json.Unmarshal(data, &themeDef); err != nil {
		return nil, err
	}

	theme := &Theme{
		Name:        themeDef.Name,
		Background:  parseHexColor(themeDef.Background),
		Foreground:  parseHexColor(themeDef.Foreground),
		TokenColors: make(map[TokenType]color.Color),
	}

	// Map string token types to TokenType constants
	tokenTypeMap := map[string]TokenType{
		"keyword":    TokenKeyword,
		"string":     TokenString,
		"comment":    TokenComment,
		"number":     TokenNumber,
		"operator":   TokenOperator,
		"identifier": TokenIdentifier,
		"function":   TokenFunction,
		"type":       TokenType_,
		"plain":      TokenPlain,
	}

	for tokenTypeStr, colorStr := range themeDef.TokenColors {
		if tokenType, exists := tokenTypeMap[tokenTypeStr]; exists {
			theme.TokenColors[tokenType] = parseHexColor(colorStr)
		}
	}

	// Set default colors for missing token types
	defaultTheme := GetDefaultTheme()
	for tokenType, defaultColor := range defaultTheme.TokenColors {
		if _, exists := theme.TokenColors[tokenType]; !exists {
			theme.TokenColors[tokenType] = defaultColor
		}
	}

	return theme, nil
}

// SaveThemeToJSON saves a theme to a JSON file
func SaveThemeToJSON(theme *Theme, filepath string) error {
	themeDef := ThemeDefinition{
		Name:        theme.Name,
		Background:  colorToHex(theme.Background),
		Foreground:  colorToHex(theme.Foreground),
		TokenColors: make(map[string]string),
	}

	// Map TokenType constants to string token types
	tokenTypeMap := map[TokenType]string{
		TokenKeyword:    "keyword",
		TokenString:     "string",
		TokenComment:    "comment",
		TokenNumber:     "number",
		TokenOperator:   "operator",
		TokenIdentifier: "identifier",
		TokenFunction:   "function",
		TokenType_:      "type",
		TokenPlain:      "plain",
	}

	for tokenType, color := range theme.TokenColors {
		if tokenTypeStr, exists := tokenTypeMap[tokenType]; exists {
			themeDef.TokenColors[tokenTypeStr] = colorToHex(color)
		}
	}

	data, err := json.MarshalIndent(themeDef, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath, data, 0644)
}

// parseHexColor converts a hex color string to color.Color
func parseHexColor(hexColor string) color.Color {
	if len(hexColor) == 0 {
		return color.NRGBA{R: 0, G: 0, B: 0, A: 255}
	}

	// Remove # prefix if present
	if hexColor[0] == '#' {
		hexColor = hexColor[1:]
	}

	// Handle short format (#RGB -> #RRGGBB)
	if len(hexColor) == 3 {
		hexColor = string([]byte{
			hexColor[0], hexColor[0],
			hexColor[1], hexColor[1],
			hexColor[2], hexColor[2],
		})
	}

	if len(hexColor) != 6 {
		return color.NRGBA{R: 0, G: 0, B: 0, A: 255}
	}

	// Parse hex values
	r := parseHexByte(hexColor[0:2])
	g := parseHexByte(hexColor[2:4])
	b := parseHexByte(hexColor[4:6])

	return color.NRGBA{R: r, G: g, B: b, A: 255}
}

// parseHexByte converts a 2-character hex string to byte
func parseHexByte(hex string) uint8 {
	var result uint8 = 0
	for _, char := range hex {
		result *= 16
		if char >= '0' && char <= '9' {
			result += uint8(char - '0')
		} else if char >= 'a' && char <= 'f' {
			result += uint8(char - 'a' + 10)
		} else if char >= 'A' && char <= 'F' {
			result += uint8(char - 'A' + 10)
		}
	}
	return result
}

// colorToHex converts a color.Color to hex string
func colorToHex(c color.Color) string {
	if c == nil {
		return "#000000"
	}

	r, g, b, _ := c.RGBA()
	// Convert from 16-bit to 8-bit
	r8 := uint8(r >> 8)
	g8 := uint8(g >> 8)
	b8 := uint8(b >> 8)

	return formatHex(r8, g8, b8)
}

// formatHex formats RGB values as hex string
func formatHex(r, g, b uint8) string {
	return "#" + byteToHex(r) + byteToHex(g) + byteToHex(b)
}

// byteToHex converts a byte to 2-character hex string
func byteToHex(b uint8) string {
	const hexChars = "0123456789abcdef"
	return string([]byte{hexChars[b>>4], hexChars[b&0x0f]})
}

// GetBuiltinThemes returns all built-in themes
func GetBuiltinThemes() map[string]*Theme {
	return map[string]*Theme{
		"default":     GetDefaultTheme(),
		"light":       GetLightTheme(),
		"vscode-dark": GetVSCodeDarkTheme(),
		"monokai":     GetMonokaiTheme(),
	}
}
