package gui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type customLightTheme struct{}

func (t *customLightTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNamePrimary:
		return primaryColor
	case theme.ColorNameBackground:
		return backgroundColor
	case theme.ColorNameButton:
		return primaryColor
	case theme.ColorNameDisabledButton:
		return disabledColor
	case theme.ColorNameDisabled:
		return disabledColor
	case theme.ColorNameForeground:
		return aiTextColor
	case theme.ColorNamePlaceHolder:
		return placeholderColor
	case theme.ColorNameInputBackground:
		return inputBgColor
	case theme.ColorNameHover:
		return hoverColor
	case theme.ColorNameFocus:
		return focusColor
	case theme.ColorNameSeparator:
		return separatorColor
	case theme.ColorNameShadow:
		return shadowColor
	case theme.ColorNameOverlayBackground:
		return backgroundColor
	case theme.ColorNameHeaderBackground:
		return backgroundColor
	case theme.ColorNameMenuBackground:
		return backgroundColor
	case theme.ColorNameWarning:
		return color.NRGBA{R: 255, G: 152, B: 0, A: 255}
	case theme.ColorNameError:
		return errorTextColor
	case theme.ColorNameSuccess:
		return color.NRGBA{R: 76, G: 175, B: 80, A: 255}
	}
	return theme.LightTheme().Color(name, variant)
}

func (t *customLightTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.LightTheme().Font(style)
}

func (t *customLightTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.LightTheme().Icon(name)
}

func (t *customLightTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNameText:
		return 14
	case theme.SizeNameHeadingText:
		return 20
	case theme.SizeNameSubHeadingText:
		return 16
	case theme.SizeNameCaptionText:
		return 11
	case theme.SizeNameInputBorder:
		return 1
	case theme.SizeNamePadding:
		return 8
	case theme.SizeNameInnerPadding:
		return 4
	}
	return theme.LightTheme().Size(name)
}
