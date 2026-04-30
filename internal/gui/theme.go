package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type ThemeManager struct {
	app fyne.App
}

func NewThemeManager(app fyne.App) *ThemeManager {
	return &ThemeManager{
		app: app,
	}
}

func (tm *ThemeManager) ApplySavedTheme() {
	tm.app.Settings().SetTheme(theme.LightTheme())
}

func (tm *ThemeManager) SetTheme(_ string) {
	tm.app.Settings().SetTheme(theme.LightTheme())
}

func (tm *ThemeManager) CurrentTheme() string {
	return "light"
}

func themeSettingsIcon() fyne.Resource {
	return theme.SettingsIcon()
}

func themeToggleIcon() fyne.Resource {
	return theme.ViewRefreshIcon()
}

func clearIcon() fyne.Resource {
	return theme.DeleteIcon()
}

func helpIcon() fyne.Resource {
	return theme.HelpIcon()
}
