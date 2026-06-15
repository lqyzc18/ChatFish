package gui

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

// MarkdownRenderer 将 Markdown 文本转换为 Fyne RichText Markup
type MarkdownRenderer struct {
	parser goldmark.Markdown
}

// NewMarkdownRenderer 创建 Markdown 渲染器
func NewMarkdownRenderer() *MarkdownRenderer {
	return &MarkdownRenderer{
		parser: goldmark.New(),
	}
}

// ToMarkup 将 Markdown 文本转换为 Fyne Markup 列表
func (m *MarkdownRenderer) ToMarkup(text string) []widget.Markup {
	if text == "" {
		return nil
	}

	// 解析 Markdown
	reader := text.NewReader([]byte(text))
	doc := m.parser.Parser().Parse(reader)

	var result []widget.Markup
	// 遍历文档节点
	for child := doc.FirstChild(); child != nil; child = child.NextSibling() {
		markups := m.renderNode(child, text)
		result = append(result, markups...)
	}

	return result
}

// renderNode 渲染单个 AST 节点
func (m *MarkdownRenderer) renderNode(node ast.Node, source string) []widget.Markup {
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
		return []widget.Markup{widget.NewText("\n")}
	case *ast.Text:
		return m.renderText(n, source)
	case *ast.String:
		return []widget.Markup{widget.NewText(string(n.Value))}
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
func (m *MarkdownRenderer) renderChildren(node ast.Node, source string) []widget.Markup {
	var result []widget.Markup
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		markups := m.renderNode(child, source)
		result = append(result, markups...)
	}
	return result
}

// renderHeading 渲染标题
func (m *MarkdownRenderer) renderHeading(node *ast.Heading, source string) []widget.Markup {
	var sb strings.Builder
	sb.WriteString("\n")

	// 根据标题级别添加标记
	for i := 0; i < node.Level; i++ {
		sb.WriteString("#")
	}
	sb.WriteString(" ")

	// 渲染标题内容
	content := m.getChildrenText(node, source)
	sb.WriteString(content)
	sb.WriteString("\n")

	return []widget.Markup{widget.NewText(sb.String())}
}

// renderParagraph 渲染段落
func (m *MarkdownRenderer) renderParagraph(node *ast.Paragraph, source string) []widget.Markup {
	var markups []widget.Markup
	markups = append(markups, m.renderChildren(node, source)...)
	markups = append(markups, widget.NewText("\n"))
	return markups
}

// renderTextBlock 渲染文本块
func (m *MarkdownRenderer) renderTextBlock(node *ast.TextBlock, source string) []widget.Markup {
	return m.renderChildren(node, source)
}

// renderList 渲染列表
func (m *MarkdownRenderer) renderList(node *ast.List, source string) []widget.Markup {
	var markups []widget.Markup
	markups = append(markups, widget.NewText("\n"))
	markups = append(markups, m.renderChildren(node, source)...)
	return markups
}

// renderListItem 渲染列表项
func (m *MarkdownRenderer) renderListItem(node *ast.ListItem, source string) []widget.Markup {
	var sb strings.Builder

	if node.ListFlags&ast.ListItemNumbered != 0 {
		sb.WriteString("  1. ")
	} else {
		sb.WriteString("  • ")
	}

	content := m.getChildrenText(node, source)
	sb.WriteString(content)
	sb.WriteString("\n")

	return []widget.Markup{widget.NewText(sb.String())}
}

// renderCodeBlock 渲染代码块
func (m *MarkdownRenderer) renderCodeBlock(node *ast.CodeBlock, source string) []widget.Markup {
	var sb strings.Builder
	sb.WriteString("\n```\n")

	for line := node.FirstChild(); line != nil; line = line.NextSibling() {
		if text, ok := line.(*ast.Text); ok {
			sb.Write(text.Value)
			sb.WriteString("\n")
		}
	}

	sb.WriteString("```\n")
	return []widget.Markup{widget.NewText(sb.String())}
}

// renderFencedCodeBlock 渲染围栏代码块
func (m *MarkdownRenderer) renderFencedCodeBlock(node *ast.FencedCodeBlock, source string) []widget.Markup {
	var sb strings.Builder
	sb.WriteString("\n```\n")

	if node.Language != nil {
		sb.WriteString("语言: ")
		sb.Write(node.Language)
		sb.WriteString("\n")
	}

	for line := node.FirstChild(); line != nil; line = line.NextSibling() {
		if text, ok := line.(*ast.Text); ok {
			sb.Write(text.Value)
			sb.WriteString("\n")
		}
	}

	sb.WriteString("```\n")
	return []widget.Markup{widget.NewText(sb.String())}
}

// renderBlockquote 渲染引用块
func (m *MarkdownRenderer) renderBlockquote(node *ast.Blockquote, source string) []widget.Markup {
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

	return []widget.Markup{widget.NewText(sb.String())}
}

// renderText 渲染文本节点
func (m *MarkdownRenderer) renderText(node *ast.Text, source string) []widget.Markup {
	text := string(node.Value)
	if node.SoftLineBreak() {
		text += " "
	}
	return []widget.Markup{widget.NewText(text)}
}

// renderCodeSpan 渲染行内代码
func (m *MarkdownRenderer) renderCodeSpan(node *ast.CodeSpan, source string) []widget.Markup {
	content := m.getChildrenText(node, source)
	return []widget.Markup{widget.NewText("`" + content + "`")}
}

// renderEmphasis 渲染强调文本
func (m *MarkdownRenderer) renderEmphasis(node *ast.Emphasis, source string) []widget.Markup {
	content := m.getChildrenText(node, source)

	switch node.Level {
	case 1:
		// 斜体
		return []widget.Markup{widget.NewText("_" + content + "_")}
	case 2:
		// 粗体
		return []widget.Markup{widget.NewText("**" + content + "**")}
	case 3:
		// 粗斜体
		return []widget.Markup{widget.NewText("***" + content + "***")}
	default:
		return []widget.Markup{widget.NewText(content)}
	}
}

// renderLink 渲染链接
func (m *MarkdownRenderer) renderLink(node *ast.Link, source string) []widget.Markup {
	content := m.getChildrenText(node, source)
	return []widget.Markup{widget.NewText(content + " (" + string(node.Destination) + ")")}
}

// renderAutoLink 渲染自动链接
func (m *MarkdownRenderer) renderAutoLink(node *ast.AutoLink, source string) []widget.Markup {
	return []widget.Markup{widget.NewText(string(node.URL(nil)))}
}

// renderRawHTML 渲染原始 HTML
func (m *MarkdownRenderer) renderRawHTML(node *ast.RawHTML, source string) []widget.Markup {
	var sb strings.Builder
	for _, segment := range node.Segments {
		sb.Write(segment.Value(source))
	}
	return []widget.Markup{widget.NewText(sb.String())}
}

// getChildrenText 获取节点所有子节点的文本内容
func (m *MarkdownRenderer) getChildrenText(node ast.Node, source string) string {
	var sb strings.Builder
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		switch n := child.(type) {
		case *ast.Text:
			sb.Write(n.Value)
		case *ast.String:
			sb.Write(n.Value)
		case *ast.CodeSpan:
			sb.WriteString("`")
			sb.WriteString(m.getChildrenText(n, source))
			sb.WriteString("`")
		case *ast.Emphasis:
			prefix := strings.Repeat("*", n.Level)
			sb.WriteString(prefix)
			sb.WriteString(m.getChildrenText(n, source))
			sb.WriteString(prefix)
		case *ast.Link:
			content := m.getChildrenText(n, source)
			sb.WriteString(content)
			sb.WriteString(" (")
			sb.Write(n.Destination)
			sb.WriteString(")")
		default:
			sb.WriteString(m.getChildrenText(child, source))
		}
	}
	return sb.String()
}
