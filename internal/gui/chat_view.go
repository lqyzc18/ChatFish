package gui

import (
	"image/color"
	"sync"
	"sync/atomic"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

const (
	bubbleMaxWidthPercent = 0.7
	bubbleMinPadding      = 10
	bubbleCornerRadius    = 12
	thinkingPlaceholder   = "正在思考..."
)

var (
	primaryColor     = color.NRGBA{R: 66, G: 133, B: 244, A: 255}
	userBgColor      = color.NRGBA{R: 66, G: 133, B: 244, A: 255}
	aiBgColor        = color.NRGBA{R: 241, G: 243, B: 244, A: 255}
	userTextColor    = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	aiTextColor      = color.NRGBA{R: 32, G: 33, B: 36, A: 255}
	separatorColor   = color.NRGBA{R: 218, G: 220, B: 224, A: 255}
	errorTextColor   = color.NRGBA{R: 211, G: 47, B: 47, A: 255}
	disabledColor    = color.NRGBA{R: 189, G: 189, B: 189, A: 255}
	placeholderColor = color.NRGBA{R: 154, G: 160, B: 166, A: 255}
	inputBgColor     = color.NRGBA{R: 245, G: 245, B: 245, A: 255}
	hoverColor       = color.NRGBA{R: 232, G: 240, B: 254, A: 255}
	focusColor       = color.NRGBA{R: 210, G: 227, B: 252, A: 255}
	shadowColor      = color.NRGBA{R: 0, G: 0, B: 0, A: 30}
	backgroundColor  = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
)

// bubbleBox 封装单条消息的气泡组件，支持普通文本和 Markdown 渲染。
type bubbleBox struct {
	bg              *canvas.Rectangle
	label           *widget.Label
	richText        *widget.RichText
	isAI            bool
	accumulatedText string
}

func newBubbleBox(text string, bgColor color.Color, textColor color.Color, maxWidth float32, isAI bool) *bubbleBox {
	bg := canvas.NewRectangle(bgColor)
	bg.CornerRadius = bubbleCornerRadius
	// 强制设置最小宽度，防止 HBox 将气泡压缩为 1 字符宽
	bg.SetMinSize(fyne.NewSize(maxWidth*0.8, 0))

	bb := &bubbleBox{
		bg:   bg,
		isAI: isAI,
	}

	if isAI {
		richText := widget.NewRichText()
		richText.Wrapping = fyne.TextWrapWord
		bb.richText = richText
	} else {
		label := widget.NewLabel(text)
		label.Wrapping = fyne.TextWrapWord
		bb.label = label
	}

	return bb
}

// Container 返回气泡的 Fyne CanvasObject 容器。
func (b *bubbleBox) Container() fyne.CanvasObject {
	if b.isAI && b.richText != nil {
		padded := container.NewPadded(b.richText)
		return container.NewStack(b.bg, padded)
	}
	padded := container.NewPadded(b.label)
	return container.NewStack(b.bg, padded)
}

// SetText 更新气泡内容，AI 消息通过 Fyne 内置 Markdown 渲染器转换。
func (b *bubbleBox) SetText(text string) {
	if b.isAI && b.richText != nil {
		b.richText.ParseMarkdown(text)
	} else {
		b.label.SetText(text)
	}
}

// ChatView 是聊天界面的核心组件，包含消息列表、输入框和发送按钮。
type ChatView struct {
	container     *fyne.Container
	scroll        *container.Scroll
	messageList   *fyne.Container
	input         *widget.Entry
	sendBtn       *widget.Button
	onSend        func(string)
	currentBubble *bubbleBox
	bubbleMu      sync.Mutex
	maxWidth      float32
	isLoading     atomic.Bool
	activity      *widget.Activity
}

// NewChatView 创建聊天视图，onSend 为发送消息的回调。
func NewChatView(onSend func(string), maxBubbleWidth float32) *ChatView {
	cv := &ChatView{
		onSend:   onSend,
		maxWidth: maxBubbleWidth,
	}
	cv.init()
	return cv
}

func (cv *ChatView) init() {
	cv.messageList = container.NewVBox()
	cv.scroll = container.NewVScroll(cv.messageList)

	cv.input = widget.NewMultiLineEntry()
	cv.input.SetPlaceHolder("输入消息...")
	cv.input.Wrapping = fyne.TextWrapWord

	cv.sendBtn = widget.NewButton("发送", func() {
		text := cv.input.Text
		if text != "" && !cv.isLoading.Load() {
			cv.onSend(text)
			cv.input.SetText("")
		}
	})
	cv.sendBtn.Importance = widget.HighImportance

	cv.activity = widget.NewActivity()
	cv.activity.Hide()

	inputContainer := container.NewBorder(nil, nil, nil, cv.sendBtn, cv.input)

	separator := canvas.NewRectangle(separatorColor)
	separator.SetMinSize(fyne.NewSize(0, 1))

	cv.container = container.NewBorder(nil, container.NewHBox(cv.activity, separator, inputContainer), nil, nil, cv.scroll)
}

// Widget 返回聊天视图的根容器。
func (cv *ChatView) Widget() fyne.CanvasObject {
	return cv.container
}

func (cv *ChatView) createBubbleRow(bubbleCanvas fyne.CanvasObject, alignLeft bool) *fyne.Container {
	if alignLeft {
		return container.NewHBox(bubbleCanvas, layout.NewSpacer())
	}
	return container.NewHBox(layout.NewSpacer(), bubbleCanvas)
}

// AddUserMessage 向消息列表添加一条用户消息气泡。
func (cv *ChatView) AddUserMessage(text string) {
	bubble := newBubbleBox(text, userBgColor, userTextColor, cv.maxWidth, false)
	row := cv.createBubbleRow(bubble.Container(), false)
	cv.messageList.Add(row)
	cv.scroll.ScrollToBottom()
}

// AddAIMessageStart 创建一个新的 AI 消息气泡并显示"正在思考..."占位符。
func (cv *ChatView) AddAIMessageStart() {
	cv.bubbleMu.Lock()
	defer cv.bubbleMu.Unlock()

	cv.currentBubble = newBubbleBox(thinkingPlaceholder, aiBgColor, aiTextColor, cv.maxWidth, true)
	cv.currentBubble.richText.ParseMarkdown(thinkingPlaceholder)
	bubbleCanvas := cv.currentBubble.Container()

	roleLabel := canvas.NewText("AI", primaryColor)
	roleLabel.TextStyle = fyne.TextStyle{Bold: true}
	roleLabel.TextSize = 12

	contentCol := container.NewVBox(roleLabel, bubbleCanvas)
	actualRow := cv.createBubbleRow(contentCol, true)
	cv.messageList.Add(actualRow)
	cv.scroll.ScrollToBottom()
}

// AddAIMessageChunk 向当前 AI 消息气泡追加一个文本片段。
func (cv *ChatView) AddAIMessageChunk(text string) {
	cv.bubbleMu.Lock()
	defer cv.bubbleMu.Unlock()

	if cv.currentBubble != nil {
		cv.currentBubble.accumulatedText += text
		cv.currentBubble.richText.ParseMarkdown(cv.currentBubble.accumulatedText)
		cv.scroll.ScrollToBottom()
	}
}

// AddAIMessageEnd 标记当前 AI 消息接收完成，恢复输入状态。
func (cv *ChatView) AddAIMessageEnd() {
	cv.bubbleMu.Lock()
	defer cv.bubbleMu.Unlock()
	cv.currentBubble = nil
	cv.isLoading.Store(false)
	cv.activity.Hide()
	cv.sendBtn.Enable()
}

// SetLoading 设置加载状态，控制发送按钮和加载动画的显隐。
func (cv *ChatView) SetLoading(loading bool) {
	cv.isLoading.Store(loading)
	if loading {
		cv.activity.Start()
		cv.activity.Show()
		cv.sendBtn.Disable()
	} else {
		cv.activity.Stop()
		cv.activity.Hide()
		cv.sendBtn.Enable()
	}
}

// ShowError 在消息列表底部显示一条错误提示。
func (cv *ChatView) ShowError(text string) {
	errLabel := canvas.NewText("错误: "+text, errorTextColor)
	errLabel.TextSize = 13
	cv.messageList.Add(errLabel)
	cv.scroll.ScrollToBottom()
}

// Clear 清空消息列表并重置当前气泡状态。
func (cv *ChatView) Clear() {
	cv.bubbleMu.Lock()
	defer cv.bubbleMu.Unlock()
	cv.messageList.Objects = nil
	cv.messageList.Refresh()
	cv.currentBubble = nil
	cv.isLoading.Store(false)
	cv.activity.Stop()
	cv.activity.Hide()
	cv.sendBtn.Enable()
}
