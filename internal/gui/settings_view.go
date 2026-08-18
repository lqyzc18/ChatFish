package gui

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"chatfish/internal/config"
)

// SettingsView 是设置页面的视图组件，提供 API 配置的表单界面。
type SettingsView struct {
	container    *fyne.Container
	apiKeyEntry  *widget.Entry
	baseURLEntry *widget.Entry
	modelEntry   *widget.Entry
	saveBtn      *widget.Button
	cancelBtn    *widget.Button
	cfg          *config.Config
	onSave       func(*config.Config) error
	onCancel     func()
}

// NewSettingsView 创建设置视图，cfg 为当前配置，onSave 为保存回调，onCancel 为取消回调。
func NewSettingsView(cfg *config.Config, onSave func(*config.Config) error, onCancel func()) *SettingsView {
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
	sv.apiKeyEntry.SetPlaceHolder("输入您的 API Key")
	// API Key 必填校验：为空时阻止保存
	sv.apiKeyEntry.Validator = func(s string) error {
		if strings.TrimSpace(s) == "" {
			return fmt.Errorf("API Key 不能为空")
		}
		return nil
	}
	if sv.cfg != nil {
		sv.apiKeyEntry.SetText(sv.cfg.APIKey)
	}

	sv.baseURLEntry = widget.NewEntry()
	sv.baseURLEntry.SetPlaceHolder("https://api.minimaxi.com/v1 (默认)")
	if sv.cfg != nil && sv.cfg.BaseURL != "" {
		sv.baseURLEntry.SetText(sv.cfg.BaseURL)
	}

	sv.modelEntry = widget.NewEntry()
	sv.modelEntry.SetPlaceHolder("MiniMax-M3 (默认)")
	if sv.cfg != nil && sv.cfg.Model != "" {
		sv.modelEntry.SetText(sv.cfg.Model)
	}

	sv.saveBtn = widget.NewButton("保存", sv.onSaveClick)
	sv.saveBtn.Importance = widget.HighImportance
	// API Key 为空时初始禁用保存按钮
	sv.saveBtn.Disable()

	// 校验状态变化时自动控制保存按钮
	sv.apiKeyEntry.SetOnValidationChanged(func(err error) {
		if err != nil {
			sv.saveBtn.Disable()
		} else {
			sv.saveBtn.Enable()
		}
	})
	if err := sv.apiKeyEntry.Validate(); err == nil {
		sv.saveBtn.Enable()
	}

	sv.cancelBtn = widget.NewButton("取消", func() {
		if sv.onCancel != nil {
			sv.onCancel()
		}
	})

	title := canvas.NewText("API 设置", primaryColor)
	title.TextSize = 18
	title.TextStyle = fyne.TextStyle{Bold: true}

	desc := widget.NewLabel("配置 AI 对话服务参数")

	apiKeyItem := widget.NewFormItem("API Key", sv.apiKeyEntry)
	apiKeyItem.Required = true

	form := widget.NewForm(
		apiKeyItem,
		widget.NewFormItem("Base URL", sv.baseURLEntry),
		widget.NewFormItem("Model", sv.modelEntry),
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

// Widget 返回设置视图的根容器。
func (sv *SettingsView) Widget() fyne.CanvasObject {
	return container.NewPadded(sv.container)
}

func (sv *SettingsView) onSaveClick() {
	// 先触发校验，校验不通过直接拦截
	if err := sv.apiKeyEntry.Validate(); err != nil {
		return
	}

	if sv.cfg == nil {
		sv.cfg = &config.Config{}
	}

	sv.cfg.APIKey = sv.apiKeyEntry.Text
	sv.cfg.BaseURL = sv.baseURLEntry.Text
	sv.cfg.Model = sv.modelEntry.Text

	if sv.onSave != nil {
		if err := sv.onSave(sv.cfg); err != nil {
			return
		}
	}

	if sv.onCancel != nil {
		sv.onCancel()
	}
}
