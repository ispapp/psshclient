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
	return theme.DefaultTheme().Color(name, variant)
}

func (m *AppTheme) Size(name fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(name)
}

func (m *AppTheme) ApplyTheme(a fyne.App) {
	a.Settings().SetTheme(m)
	a.SetIcon(m.Icon(theme.IconNameHome))
}
