package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"chatfish/internal/config"
)

type SettingsView struct {
	container   *fyne.Container
	apiKeyEntry *widget.Entry
	saveBtn     *widget.Button
	cancelBtn   *widget.Button
	cfg         *config.Config
	onSave      func(*config.Config)
	onCancel    func()
}

func NewSettingsView(cfg *config.Config, onSave func(*config.Config), onCancel func(), _ func(string)) *SettingsView {
	sv := &SettingsView{
		cfg:      cfg,
		onSave:   onSave,
		onCancel: onCancel,
	}
	sv.init()
	return sv
}

func (sv *SettingsView) init() {
	sv.apiKeyEntry = widget.NewPasswordEntry()
	sv.apiKeyEntry.SetPlaceHolder("输入您的 MiniMax API Key")
	if sv.cfg != nil {
		sv.apiKeyEntry.SetText(sv.cfg.APIKey)
	}

	sv.saveBtn = widget.NewButton("保存", sv.onSaveClick)
	sv.saveBtn.Importance = widget.HighImportance

	sv.cancelBtn = widget.NewButton("取消", func() {
		if sv.onCancel != nil {
			sv.onCancel()
		}
	})

	title := canvas.NewText("API Key 设置", primaryColor)
	title.TextSize = 18
	title.TextStyle = fyne.TextStyle{Bold: true}

	desc := widget.NewLabel("请输入您的 MiniMax API Key 以使用 AI 对话功能")

	form := widget.NewForm(
		widget.NewFormItem("API Key", sv.apiKeyEntry),
	)

	buttons := container.NewHBox(
		layout.NewSpacer(),
		sv.cancelBtn,
		sv.saveBtn,
	)

	sv.container = container.NewVBox(
		title,
		widget.NewSeparator(),
		desc,
		widget.NewLabel(""),
		form,
		widget.NewLabel(""),
		widget.NewSeparator(),
		buttons,
	)
}

func (sv *SettingsView) Widget() fyne.CanvasObject {
	return container.NewPadded(sv.container)
}

func (sv *SettingsView) onSaveClick() {
	if sv.cfg == nil {
		sv.cfg = &config.Config{}
	}

	sv.cfg.APIKey = sv.apiKeyEntry.Text

	if sv.onSave != nil {
		sv.onSave(sv.cfg)
	}

	if sv.onCancel != nil {
		sv.onCancel()
	}
}
