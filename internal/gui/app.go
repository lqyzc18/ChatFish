// Package gui 负责桌面 GUI 端的窗口、控件与事件编排。
//
// 本文件的关键点是：把 chat/service 的回调“绑定到当前会话”，
// 通过 chatVersion/minRequestID 过滤掉被取消或过期请求的迟到回调，
// 避免 UI 出现错乱（例如：清空对话后旧回复又被渲染出来）。
package gui

import (
	"strings"
	"sync/atomic"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"chatfish/internal/chat"
	"chatfish/internal/config"
)

const (
	appTitle     = "ChatFish"
	windowWidth  = 900
	windowHeight = 700
)

// App 是 ChatFish 应用的顶层结构，管理窗口生命周期和各组件的协调。
type App struct {
	fyneApp        fyne.App
	mainWindow     fyne.Window
	settingsWindow fyne.Window
	chatView       *ChatView
	settings       *SettingsView
	chatSvc        *chat.Service
	cfg            *config.Config
	chatVersion    atomic.Uint64
	// minRequestID 表示“最小可接受的 requestID”。
	// 在用户点击清空/取消时会递增，从而让旧请求回调在 UI 过滤阶段被丢弃。
	minRequestID atomic.Uint64
}

// Run 启动 ChatFish 应用，初始化 GUI 并阻塞在主事件循环中。
func Run() {
	a := app.NewWithID("com.chatfish.app")
	a.Settings().SetTheme(&customLightTheme{})

	w := a.NewWindow(appTitle)
	w.Resize(fyne.NewSize(windowWidth, windowHeight))
	w.CenterOnScreen()
	w.SetMaster()

	app := &App{
		fyneApp:    a,
		mainWindow: w,
	}
	app.init()
	w.ShowAndRun()
}

func (a *App) init() {
	cfg, err := config.Load()
	if err != nil {
		// 配置加载失败时提示用户，使用空配置继续运行
		cfg = &config.Config{}
		dialog.ShowError(err, a.mainWindow)
	}
	a.cfg = cfg

	maxBubbleWidth := float32(windowWidth) * bubbleMaxWidthPercent
	a.chatView = NewChatView(a.onSendMessage, maxBubbleWidth)
	a.settings = NewSettingsView(a.cfg, a.onSettingsSave, a.closeSettingsWindow)

	toolbar := a.createToolbar()

	titleLabel := canvas.NewText("  ChatFish", primaryColor)
	titleLabel.TextSize = 18
	titleLabel.TextStyle = fyne.TextStyle{Bold: true}

	header := container.NewHBox(titleLabel, layout.NewSpacer(), toolbar)
	separator := canvas.NewRectangle(separatorColor)
	separator.SetMinSize(fyne.NewSize(0, 2))

	content := container.NewBorder(
		container.NewVBox(header, separator),
		nil, nil, nil,
		a.chatView.Widget(),
	)

	a.mainWindow.SetContent(content)

	if a.cfg.APIKey != "" {
		if err := a.initChatService(); err != nil {
			a.chatView.ShowError("初始化 AI 服务失败: " + err.Error())
		}
	}

	// 注册 Ctrl+Enter 发送快捷键
	ctrlEnter := &desktop.CustomShortcut{KeyName: fyne.KeyReturn, Modifier: fyne.KeyModifierControl}
	a.mainWindow.Canvas().AddShortcut(ctrlEnter, func(shortcut fyne.Shortcut) {
		a.chatView.sendBtn.Tapped(&fyne.PointEvent{})
	})
}

func (a *App) createToolbar() *widget.Toolbar {
	return widget.NewToolbar(
		widget.NewToolbarAction(clearIcon(), func() {
			if a.chatSvc != nil {
				// 将 minRequestID 提高到“刚取消那轮之后”，用于忽略旧请求的延迟回调渲染。
				a.minRequestID.Store(a.chatSvc.Cancel() + 1)
				a.chatSvc.ClearHistory()
			}
			a.chatView.Clear()
		}),
		widget.NewToolbarAction(themeSettingsIcon(), func() {
			a.showSettings()
		}),
		widget.NewToolbarAction(helpIcon(), func() {
			a.showHelp()
		}),
	)
}

func (a *App) initChatService() error {
	version := a.chatVersion.Load() + 1
	if strings.TrimSpace(a.cfg.APIKey) == "" {
		oldService := a.chatSvc
		a.chatSvc = nil
		a.chatVersion.Store(version)
		a.minRequestID.Store(0)
		if oldService != nil {
			oldService.Cancel()
		}
		return nil
	}

	svc, err := chat.NewService(a.cfg.APIKey, a.cfg.BaseURL, a.cfg.Model, chat.WithGUIOutput(chat.GUIStreamCallbacks{
		OnStart: func(requestID uint64) {
			// 这些回调可能来自模型流读取协程，因此必须通过 runForCurrentRequest + fyne.Do 保证 UI 安全。
			a.runForCurrentRequest(version, requestID, func() { a.chatView.AddAIMessageStart() })
		},
		OnChunk: func(requestID uint64, text string) {
			a.runForCurrentRequest(version, requestID, func() { a.chatView.AddAIMessageChunk(text) })
		},
		OnFinish: func(requestID uint64) {
			a.runForCurrentRequest(version, requestID, func() { a.chatView.AddAIMessageEnd() })
		},
		OnError: func(requestID uint64, err error) {
			a.runForCurrentRequest(version, requestID, func() {
				a.chatView.ShowError("流式响应异常: " + err.Error())
			})
		},
	}))
	if err != nil {
		return err
	}

	oldService := a.chatSvc
	a.chatSvc = svc
	a.chatVersion.Store(version)
	a.minRequestID.Store(0)
	if oldService != nil {
		oldService.Cancel()
	}
	return nil
}

func (a *App) runForCurrentRequest(version, requestID uint64, action func()) {
	fyne.Do(func() {
		// version 用于识别“配置/模型切换导致的服务重建”，minRequestID 用于识别“清空/取消后的过期请求”。
		if a.chatVersion.Load() == version && requestID >= a.minRequestID.Load() {
			action()
		}
	})
}

func (a *App) onSendMessage(text string) {
	svc := a.chatSvc
	if svc == nil {
		a.chatView.ShowError("请先在设置中配置 API Key")
		return
	}
	version := a.chatVersion.Load()
	a.chatView.AddUserMessage(text)
	a.chatView.SetLoading(true)
	go func() {
		// svc.Chat() 会在内部读取流并通过 GUIStreamCallbacks 推送增量内容。
		// 这里只负责：等待 Chat() 结束后在“仍是当前版本且仍然是该 svc”时恢复 UI。
		if err := svc.Chat(text); err != nil {
			fyne.Do(func() {
				if a.chatVersion.Load() == version && a.chatSvc == svc {
					a.chatView.SetLoading(false)
				}
			})
		}
	}()
}

func (a *App) onSettingsSave(cfg *config.Config) error {
	a.cfg = cfg
	if err := config.Save(cfg); err != nil {
		a.chatView.ShowError("保存配置失败: " + err.Error())
		return err
	}
	if err := a.initChatService(); err != nil {
		a.chatView.ShowError("初始化 AI 服务失败: " + err.Error())
		return err
	}
	return nil
}

func (a *App) closeSettingsWindow() {
	if a.settingsWindow != nil {
		a.settingsWindow.Close()
	}
}

func (a *App) showSettings() {
	a.mainWindow.Hide()
	a.settingsWindow = a.fyneApp.NewWindow("设置")
	a.settingsWindow.Resize(fyne.NewSize(420, 250))
	a.settingsWindow.CenterOnScreen()
	a.settingsWindow.SetContent(a.settings.Widget())
	a.settingsWindow.SetOnClosed(func() {
		a.mainWindow.Show()
		a.settingsWindow = nil
	})
	a.settingsWindow.Show()
}

func (a *App) showHelp() {
	a.mainWindow.Hide()
	helpWindow := a.fyneApp.NewWindow("使用帮助")
	helpWindow.Resize(fyne.NewSize(400, 350))
	helpWindow.CenterOnScreen()

	title := widget.NewLabelWithStyle("ChatFish 使用帮助", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	content1 := widget.NewLabelWithStyle("基本功能", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	items1 := widget.NewLabel("  • 发送消息：在输入框中输入消息，按回车或点击发送按钮\n  • 清除对话：点击工具栏的清除按钮\n  • 设置：点击工具栏的设置按钮配置 API Key")
	content2 := widget.NewLabelWithStyle("快捷操作", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	items2 := widget.NewLabel("  • 点击发送按钮或按 Enter 后换行再点发送：发送消息\n  • Enter：换行")
	closeBtn := widget.NewButton("关闭", func() { helpWindow.Close() })

	box := container.NewVBox(title, widget.NewSeparator(), content1, items1, widget.NewSeparator(), content2, items2, widget.NewSeparator(), closeBtn)
	helpWindow.SetContent(container.NewPadded(box))
	helpWindow.SetOnClosed(func() { a.mainWindow.Show() })
	helpWindow.Show()
}
