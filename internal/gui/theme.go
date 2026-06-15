package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

func themeSettingsIcon() fyne.Resource {
	return theme.SettingsIcon()
}

func clearIcon() fyne.Resource {
	return theme.DeleteIcon()
}

func helpIcon() fyne.Resource {
	return theme.HelpIcon()
}
