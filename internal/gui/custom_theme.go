package gui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

var (
	customPrimary    = color.NRGBA{R: 66, G: 133, B: 244, A: 255}
	customBackground = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	customButton     = color.NRGBA{R: 66, G: 133, B: 244, A: 255}
	customDisabled   = color.NRGBA{R: 189, G: 189, B: 189, A: 255}
	customText       = color.NRGBA{R: 32, G: 33, B: 36, A: 255}
	customPlaceholder = color.NRGBA{R: 154, G: 160, B: 166, A: 255}
	customInputBg    = color.NRGBA{R: 245, G: 245, B: 245, A: 255}
	customHover      = color.NRGBA{R: 232, G: 240, B: 254, A: 255}
	customFocus      = color.NRGBA{R: 210, G: 227, B: 252, A: 255}
	customSeparator  = color.NRGBA{R: 218, G: 220, B: 224, A: 255}
	customShadow     = color.NRGBA{R: 0, G: 0, B: 0, A: 30}
)

type customLightTheme struct{}

func (t *customLightTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNamePrimary:
		return customPrimary
	case theme.ColorNameBackground:
		return customBackground
	case theme.ColorNameButton:
		return customButton
	case theme.ColorNameDisabledButton:
		return customDisabled
	case theme.ColorNameDisabled:
		return customDisabled
	case theme.ColorNameForeground:
		return customText
	case theme.ColorNamePlaceHolder:
		return customPlaceholder
	case theme.ColorNameInputBackground:
		return customInputBg
	case theme.ColorNameHover:
		return customHover
	case theme.ColorNameFocus:
		return customFocus
	case theme.ColorNameSeparator:
		return customSeparator
	case theme.ColorNameShadow:
		return customShadow
	case theme.ColorNameOverlayBackground:
		return customBackground
	case theme.ColorNameHeaderBackground:
		return customBackground
	case theme.ColorNameMenuBackground:
		return customBackground
	case theme.ColorNameWarning:
		return color.NRGBA{R: 255, G: 152, B: 0, A: 255}
	case theme.ColorNameError:
		return color.NRGBA{R: 211, G: 47, B: 47, A: 255}
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
