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
	primaryColor   = color.NRGBA{R: 66, G: 133, B: 244, A: 255}
	userBgColor    = color.NRGBA{R: 66, G: 133, B: 244, A: 255}
	aiBgColor      = color.NRGBA{R: 241, G: 243, B: 244, A: 255}
	userTextColor  = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	aiTextColor    = color.NRGBA{R: 32, G: 33, B: 36, A: 255}
	separatorColor = color.NRGBA{R: 218, G: 220, B: 224, A: 255}
	errorTextColor = color.NRGBA{R: 211, G: 47, B: 47, A: 255}
)

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
	// 关键：强制设置一个最小宽度，防止 HBox 将其压缩为 1 字符宽
	bg.SetMinSize(fyne.NewSize(maxWidth*0.8, 0))

	bb := &bubbleBox{
		bg:   bg,
		isAI: isAI,
	}

	if isAI {
		// AI 消息使用 RichText 支持 Markdown
		richText := widget.NewRichText()
		richText.Wrapping = fyne.TextWrapWord
		bb.richText = richText
	} else {
		// 用户消息使用普通 Label
		label := widget.NewLabel(text)
		label.Wrapping = fyne.TextWrapWord
		bb.label = label
	}

	return bb
}

func (b *bubbleBox) Container() fyne.CanvasObject {
	if b.isAI && b.richText != nil {
		padded := container.NewPadded(b.richText)
		return container.NewStack(b.bg, padded)
	}
	padded := container.NewPadded(b.label)
	return container.NewStack(b.bg, padded)
}

func (b *bubbleBox) SetText(text string) {
	if b.isAI && b.richText != nil {
		// AI 消息使用 Markdown 渲染
		renderer := NewMarkdownRenderer()
		b.richText.Segments = renderer.ToMarkup(text)
		b.richText.Refresh()
	} else {
		b.label.SetText(text)
	}
}

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
	loadingLabel  *widget.Label
}

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

	cv.loadingLabel = widget.NewLabel("")
	cv.loadingLabel.Hide()

	inputContainer := container.NewBorder(nil, nil, nil, cv.sendBtn, cv.input)

	separator := canvas.NewRectangle(separatorColor)
	separator.SetMinSize(fyne.NewSize(0, 1))

	cv.container = container.NewBorder(nil, container.NewVBox(cv.loadingLabel, separator, inputContainer), nil, nil, cv.scroll)
}

func (cv *ChatView) Widget() fyne.CanvasObject {
	return cv.container
}

func (cv *ChatView) createBubbleRow(bubbleCanvas fyne.CanvasObject, alignLeft bool) *fyne.Container {
	if alignLeft {
		return container.NewHBox(bubbleCanvas, layout.NewSpacer())
	}
	return container.NewHBox(layout.NewSpacer(), bubbleCanvas)
}

func (cv *ChatView) AddUserMessage(text string) {
	bubble := newBubbleBox(text, userBgColor, userTextColor, cv.maxWidth, false)
	row := cv.createBubbleRow(bubble.Container(), false)
	cv.messageList.Add(row)
	cv.scroll.ScrollToBottom()
}

func (cv *ChatView) AddAIMessageStart() {
	cv.bubbleMu.Lock()
	defer cv.bubbleMu.Unlock()

	cv.currentBubble = newBubbleBox(thinkingPlaceholder, aiBgColor, aiTextColor, cv.maxWidth, true)
	bubbleCanvas := cv.currentBubble.Container()

	roleLabel := canvas.NewText("AI", primaryColor)
	roleLabel.TextStyle = fyne.TextStyle{Bold: true}
	roleLabel.TextSize = 12

	contentCol := container.NewVBox(roleLabel, bubbleCanvas)
	actualRow := cv.createBubbleRow(contentCol, true)
	cv.messageList.Add(actualRow)
	cv.scroll.ScrollToBottom()
}

func (cv *ChatView) AddAIMessageChunk(text string) {
	cv.bubbleMu.Lock()
	defer cv.bubbleMu.Unlock()

	if cv.currentBubble != nil {
		if cv.currentBubble.accumulatedText == "" {
			// 第一个 chunk，替换占位符
			cv.currentBubble.accumulatedText = text
		} else {
			cv.currentBubble.accumulatedText += text
		}
		cv.currentBubble.SetText(cv.currentBubble.accumulatedText)
		cv.scroll.ScrollToBottom()
	}
}

func (cv *ChatView) AddAIMessageEnd() {
	cv.bubbleMu.Lock()
	defer cv.bubbleMu.Unlock()
	cv.currentBubble = nil
	cv.isLoading.Store(false)
	cv.loadingLabel.Hide()
	cv.sendBtn.Enable()
}

func (cv *ChatView) SetLoading(loading bool) {
	cv.isLoading.Store(loading)
	if loading {
		cv.loadingLabel.SetText("正在思考...")
		cv.loadingLabel.Show()
		cv.sendBtn.Disable()
	} else {
		cv.loadingLabel.Hide()
		cv.sendBtn.Enable()
	}
}

func (cv *ChatView) AddAIMessage(text string) {
	bubble := newBubbleBox(text, aiBgColor, aiTextColor, cv.maxWidth, true)
	roleLabel := canvas.NewText("AI", primaryColor)
	roleLabel.TextStyle = fyne.TextStyle{Bold: true}
	roleLabel.TextSize = 12

	bubbleCanvas := bubble.Container()
	contentCol := container.NewVBox(roleLabel, bubbleCanvas)
	row := cv.createBubbleRow(contentCol, true)

	cv.messageList.Add(row)
	cv.scroll.ScrollToBottom()
}

func (cv *ChatView) ShowError(text string) {
	errLabel := canvas.NewText("错误: "+text, errorTextColor)
	errLabel.TextSize = 13
	cv.messageList.Add(errLabel)
	cv.scroll.ScrollToBottom()
}

func (cv *ChatView) Clear() {
	cv.bubbleMu.Lock()
	defer cv.bubbleMu.Unlock()
	cv.messageList.Objects = nil
	cv.messageList.Refresh()
	cv.currentBubble = nil
}
