package gui

import (
	"strings"

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
				a.chatSvc.Cancel()
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
	// 取消旧服务正在进行的请求
	if a.chatSvc != nil {
		a.chatSvc.Cancel()
		a.chatSvc = nil
	}

	if strings.TrimSpace(a.cfg.APIKey) == "" {
		return nil
	}
	svc, err := chat.NewService(a.cfg.APIKey, a.cfg.BaseURL, a.cfg.Model, chat.WithGUIOutput(chat.GUIStreamCallbacks{
		OnStart:  func() { fyne.Do(func() { a.chatView.AddAIMessageStart() }) },
		OnChunk:  func(text string) { fyne.Do(func() { a.chatView.AddAIMessageChunk(text) }) },
		OnFinish: func() { fyne.Do(func() { a.chatView.AddAIMessageEnd() }) },
		OnError:  func(err error) { fyne.Do(func() { a.chatView.ShowError("流式响应异常: " + err.Error()) }) },
	}))
	if err != nil {
		return err
	}
	a.chatSvc = svc
	return nil
}

func (a *App) onSendMessage(text string) {
	svc := a.chatSvc
	if svc == nil {
		a.chatView.ShowError("请先在设置中配置 API Key")
		return
	}
	a.chatView.AddUserMessage(text)
	a.chatView.SetLoading(true)
	go func() {
		if err := svc.Chat(text); err != nil {
			fyne.Do(func() {
				a.chatView.SetLoading(false)
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
