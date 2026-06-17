package gui

import (
	"fmt"
	"net/url"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	gmtext "github.com/yuin/goldmark/text"
)

// MarkdownRenderer 将 Markdown 文本转换为 Fyne RichTextSegment
type MarkdownRenderer struct {
	parser goldmark.Markdown
}

// NewMarkdownRenderer 创建 Markdown 渲染器
func NewMarkdownRenderer() *MarkdownRenderer {
	return &MarkdownRenderer{
		parser: goldmark.New(),
	}
}

// ToMarkup 将 Markdown 文本转换为 Fyne RichTextSegment 列表
func (m *MarkdownRenderer) ToMarkup(md string) []widget.RichTextSegment {
	if md == "" {
		return nil
	}

	reader := gmtext.NewReader([]byte(md))
	doc := m.parser.Parser().Parse(reader)

	var result []widget.RichTextSegment
	for child := doc.FirstChild(); child != nil; child = child.NextSibling() {
		segs := m.renderNode(child, md)
		result = append(result, segs...)
	}

	return result
}

// renderNode 渲染单个 AST 节点
func (m *MarkdownRenderer) renderNode(node ast.Node, source string) []widget.RichTextSegment {
	switch n := node.(type) {
	case *ast.Document:
		return m.renderChildren(n, source)
	case *ast.Heading:
		return m.renderHeading(n, source)
	case *ast.Paragraph:
		return m.renderParagraph(n, source)
	case *ast.TextBlock:
		return m.renderTextBlock(n, source)
	case *ast.List:
		return m.renderList(n, source)
	case *ast.ListItem:
		return m.renderListItem(n, source)
	case *ast.CodeBlock:
		return m.renderCodeBlock(n, source)
	case *ast.FencedCodeBlock:
		return m.renderFencedCodeBlock(n, source)
	case *ast.Blockquote:
		return m.renderBlockquote(n, source)
	case *ast.ThematicBreak:
		return []widget.RichTextSegment{textSeg("\n", widget.RichTextStyleParagraph)}
	case *ast.Text:
		return m.renderText(n, source)
	case *ast.String:
		return []widget.RichTextSegment{textSeg(string(n.Value), widget.RichTextStyleInline)}
	case *ast.CodeSpan:
		return m.renderCodeSpan(n, source)
	case *ast.Emphasis:
		return m.renderEmphasis(n, source)
	case *ast.Link:
		return m.renderLink(n, source)
	case *ast.AutoLink:
		return m.renderAutoLink(n, source)
	case *ast.RawHTML:
		return m.renderRawHTML(n, source)
	default:
		return m.renderChildren(node, source)
	}
}

// renderChildren 渲染子节点
func (m *MarkdownRenderer) renderChildren(node ast.Node, source string) []widget.RichTextSegment {
	var result []widget.RichTextSegment
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		segs := m.renderNode(child, source)
		result = append(result, segs...)
	}
	return result
}

// renderHeading 渲染标题
func (m *MarkdownRenderer) renderHeading(node *ast.Heading, source string) []widget.RichTextSegment {
	content := m.getChildrenText(node, source)

	style := widget.RichTextStyleHeading
	switch node.Level {
	case 1:
		style.SizeName = fyne.ThemeSizeName("heading")
	case 2:
		style.SizeName = fyne.ThemeSizeName("subHeading")
	default:
		style.SizeName = fyne.ThemeSizeName("caption")
		style.TextStyle = fyne.TextStyle{Bold: true}
	}

	return []widget.RichTextSegment{
		textSeg("\n", widget.RichTextStyleParagraph),
		textSeg(content, style),
		textSeg("\n", widget.RichTextStyleParagraph),
	}
}

// renderParagraph 渲染段落
func (m *MarkdownRenderer) renderParagraph(node *ast.Paragraph, source string) []widget.RichTextSegment {
	segs := m.renderChildren(node, source)
	segs = append(segs, textSeg("\n", widget.RichTextStyleParagraph))
	return segs
}

// renderTextBlock 渲染文本块
func (m *MarkdownRenderer) renderTextBlock(node *ast.TextBlock, source string) []widget.RichTextSegment {
	return m.renderChildren(node, source)
}

// renderList 渲染列表
func (m *MarkdownRenderer) renderList(node *ast.List, source string) []widget.RichTextSegment {
	var segs []widget.RichTextSegment
	segs = append(segs, textSeg("\n", widget.RichTextStyleParagraph))
	segs = append(segs, m.renderChildren(node, source)...)
	return segs
}

// renderListItem 渲染列表项
func (m *MarkdownRenderer) renderListItem(node *ast.ListItem, source string) []widget.RichTextSegment {
	var prefix string
	if parent, ok := node.Parent().(*ast.List); ok && parent.IsOrdered() {
		// 通过遍历兄弟节点计算实际序号
		idx := 1
		for sibling := node.PreviousSibling(); sibling != nil; sibling = sibling.PreviousSibling() {
			idx++
		}
		prefix = fmt.Sprintf("  %d. ", idx)
	} else {
		prefix = "  • "
	}
	content := m.getChildrenText(node, source)
	return []widget.RichTextSegment{
		textSeg(prefix+content+"\n", widget.RichTextStyleParagraph),
	}
}

// renderCodeBlock 渲染代码块
func (m *MarkdownRenderer) renderCodeBlock(node *ast.CodeBlock, source string) []widget.RichTextSegment {
	var sb strings.Builder
	sb.WriteString("\n")
	src := []byte(source)
	for line := node.FirstChild(); line != nil; line = line.NextSibling() {
		if t, ok := line.(*ast.Text); ok {
			sb.Write(t.Value(src))
			sb.WriteString("\n")
		}
	}
	return []widget.RichTextSegment{
		textSeg(sb.String(), widget.RichTextStyleCodeBlock),
	}
}

// renderFencedCodeBlock 渲染围栏代码块
func (m *MarkdownRenderer) renderFencedCodeBlock(node *ast.FencedCodeBlock, source string) []widget.RichTextSegment {
	var sb strings.Builder
	sb.WriteString("\n")
	src := []byte(source)
	for line := node.FirstChild(); line != nil; line = line.NextSibling() {
		if t, ok := line.(*ast.Text); ok {
			sb.Write(t.Value(src))
			sb.WriteString("\n")
		}
	}
	return []widget.RichTextSegment{
		textSeg(sb.String(), widget.RichTextStyleCodeBlock),
	}
}

// renderBlockquote 渲染引用块
func (m *MarkdownRenderer) renderBlockquote(node *ast.Blockquote, source string) []widget.RichTextSegment {
	var sb strings.Builder
	sb.WriteString("\n")
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		content := m.getChildrenText(child, source)
		lines := strings.Split(content, "\n")
		for _, line := range lines {
			if strings.TrimSpace(line) != "" {
				sb.WriteString("> ")
				sb.WriteString(strings.TrimSpace(line))
				sb.WriteString("\n")
			}
		}
	}
	return []widget.RichTextSegment{
		textSeg(sb.String(), widget.RichTextStyleBlockquote),
	}
}

// renderText 渲染文本节点
func (m *MarkdownRenderer) renderText(node *ast.Text, source string) []widget.RichTextSegment {
	t := string(node.Value([]byte(source)))
	if node.SoftLineBreak() {
		t += " "
	}
	return []widget.RichTextSegment{textSeg(t, widget.RichTextStyleInline)}
}

// renderCodeSpan 渲染行内代码
func (m *MarkdownRenderer) renderCodeSpan(node *ast.CodeSpan, source string) []widget.RichTextSegment {
	content := m.getChildrenText(node, source)
	return []widget.RichTextSegment{textSeg(content, widget.RichTextStyleCodeInline)}
}

// renderEmphasis 渲染强调文本
func (m *MarkdownRenderer) renderEmphasis(node *ast.Emphasis, source string) []widget.RichTextSegment {
	content := m.getChildrenText(node, source)
	switch node.Level {
	case 1:
		return []widget.RichTextSegment{textSeg(content, widget.RichTextStyleEmphasis)}
	case 2:
		return []widget.RichTextSegment{textSeg(content, widget.RichTextStyleStrong)}
	default:
		return []widget.RichTextSegment{textSeg(content, widget.RichTextStyleInline)}
	}
}

// renderLink 渲染链接
func (m *MarkdownRenderer) renderLink(node *ast.Link, source string) []widget.RichTextSegment {
	content := m.getChildrenText(node, source)
	u, err := url.Parse(string(node.Destination))
	if err != nil {
		// URL 解析失败时回退为纯文本
		return []widget.RichTextSegment{textSeg(content, widget.RichTextStyleInline)}
	}
	return []widget.RichTextSegment{
		&widget.HyperlinkSegment{
			Text: content,
			URL:  u,
		},
	}
}

// renderAutoLink 渲染自动链接
func (m *MarkdownRenderer) renderAutoLink(node *ast.AutoLink, source string) []widget.RichTextSegment {
	raw := string(node.URL(nil))
	u, err := url.Parse(raw)
	if err != nil {
		return []widget.RichTextSegment{textSeg(raw, widget.RichTextStyleInline)}
	}
	return []widget.RichTextSegment{
		&widget.HyperlinkSegment{
			Text: raw,
			URL:  u,
		},
	}
}

// renderRawHTML 渲染原始 HTML
func (m *MarkdownRenderer) renderRawHTML(node *ast.RawHTML, source string) []widget.RichTextSegment {
	if node.Segments == nil {
		return nil
	}
	raw := node.Segments.Value([]byte(source))
	return []widget.RichTextSegment{textSeg(string(raw), widget.RichTextStyleInline)}
}

// getChildrenText 获取节点所有子节点的文本内容
func (m *MarkdownRenderer) getChildrenText(node ast.Node, source string) string {
	var sb strings.Builder
	src := []byte(source)
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		switch n := child.(type) {
		case *ast.Text:
			sb.Write(n.Value(src))
		case *ast.String:
			sb.Write(n.Value)
		case *ast.CodeSpan:
			sb.WriteString(m.getChildrenText(n, source))
		case *ast.Emphasis:
			sb.WriteString(m.getChildrenText(n, source))
		case *ast.Link:
			sb.WriteString(m.getChildrenText(n, source))
		default:
			sb.WriteString(m.getChildrenText(child, source))
		}
	}
	return sb.String()
}

// textSeg 创建文本段的辅助函数
func textSeg(t string, style widget.RichTextStyle) *widget.TextSegment {
	return &widget.TextSegment{Text: t, Style: style}
}
