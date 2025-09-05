package theme

import (
	"image/color"

	"github.com/ispapp/psshclient/internal/resources"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"

	_ "embed" // Required for go:embed
)

var FontBytes []byte

var IconBytes []byte

type AppTheme struct{}

func (m *AppTheme) Font(style fyne.TextStyle) fyne.Resource {
	if style.Monospace {
		return theme.DefaultTheme().Font(style) // Use default monospace if needed
	}
	return resources.ResourceFiraSansCondensedRegularTtf
}

func (m *AppTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	if name == theme.IconNameHome {
		return resources.ResourceIconPng
	}
	if name == theme.IconNameFile {
		return resources.ResourceIconPng
	}
	if name == fyne.ThemeIconName(theme.SizeNameWindowButtonIcon) {
		return resources.ResourceIconPng
	}
	return theme.DefaultTheme().Icon(name) // Fallback to default for other icons
}

func (m *AppTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	// Custom color palette: rgb(192, 201, 238), rgb(162, 170, 219), rgb(137, 138, 196)
	switch name {
	case theme.ColorNameBackground:
		if variant == theme.VariantDark {
			return color.RGBA{30, 32, 42, 255} // Dark background to contrast with palette
		}
		return color.RGBA{248, 250, 252, 255} // Light background

	case theme.ColorNameForeground:
		if variant == theme.VariantDark {
			return color.RGBA{243, 251, 246, 255} // Light text for dark theme (general UI text)
		}
		return color.RGBA{243, 251, 246, 255} // Dark text for light theme

	case theme.ColorNamePrimary:
		if variant == theme.VariantDark {
			return color.RGBA{192, 201, 238, 255} // Primary accent color for dark theme
		}
		return color.RGBA{66, 72, 83, 255} // Darker accent for light theme
	}
	// Fallback to default theme for unhandled colors
	return theme.DefaultTheme().Color(name, variant)
}

func (m *AppTheme) Size(name fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(name)
}

func (m *AppTheme) ApplyTheme(a fyne.App) {
	a.Settings().SetTheme(m)
	a.SetIcon(m.Icon(theme.IconNameHome))
	// Set the app to use dark variant by default for the ISP theme
	a.Preferences().SetString("theme", "dark")
}
