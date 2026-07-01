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

var markdownParser = goldmark.New()

type MarkdownRenderer struct{}

func NewMarkdownRenderer() *MarkdownRenderer {
	return &MarkdownRenderer{}
}

func (m *MarkdownRenderer) ToMarkup(md string) []widget.RichTextSegment {
	if md == "" {
		return nil
	}
	return m.toMarkupFrom([]byte(md))
}

func (m *MarkdownRenderer) toMarkupFrom(src []byte) []widget.RichTextSegment {
	reader := gmtext.NewReader(src)
	doc := markdownParser.Parser().Parse(reader)

	var result []widget.RichTextSegment
	for child := doc.FirstChild(); child != nil; child = child.NextSibling() {
		segs := m.renderNode(child, src)
		result = append(result, segs...)
	}

	return result
}

func (m *MarkdownRenderer) AppendMarkup(existing []widget.RichTextSegment, fullText string, newText string) []widget.RichTextSegment {
	_ = existing
	_ = newText
	// Goldmark 语法节点会跨行重组，当前增量拼接无法安全截断旧 segments。
	// 这里退回到整段重解析，优先保证流式渲染结果正确。
	return m.toMarkupFrom([]byte(fullText))
}

func (m *MarkdownRenderer) renderNode(node ast.Node, source []byte) []widget.RichTextSegment {
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

func (m *MarkdownRenderer) renderChildren(node ast.Node, source []byte) []widget.RichTextSegment {
	var result []widget.RichTextSegment
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		segs := m.renderNode(child, source)
		result = append(result, segs...)
	}
	return result
}

func (m *MarkdownRenderer) renderHeading(node *ast.Heading, source []byte) []widget.RichTextSegment {
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

func (m *MarkdownRenderer) renderParagraph(node *ast.Paragraph, source []byte) []widget.RichTextSegment {
	segs := m.renderChildren(node, source)
	segs = append(segs, textSeg("\n", widget.RichTextStyleParagraph))
	return segs
}

func (m *MarkdownRenderer) renderTextBlock(node *ast.TextBlock, source []byte) []widget.RichTextSegment {
	return m.renderChildren(node, source)
}

func (m *MarkdownRenderer) renderList(node *ast.List, source []byte) []widget.RichTextSegment {
	var segs []widget.RichTextSegment
	segs = append(segs, textSeg("\n", widget.RichTextStyleParagraph))
	segs = append(segs, m.renderChildren(node, source)...)
	return segs
}

func (m *MarkdownRenderer) renderListItem(node *ast.ListItem, source []byte) []widget.RichTextSegment {
	var prefix string
	if parent, ok := node.Parent().(*ast.List); ok && parent.IsOrdered() {
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

func (m *MarkdownRenderer) renderCodeBlock(node *ast.CodeBlock, source []byte) []widget.RichTextSegment {
	var sb strings.Builder
	sb.WriteString("\n")
	for line := node.FirstChild(); line != nil; line = line.NextSibling() {
		if t, ok := line.(*ast.Text); ok {
			sb.Write(t.Value(source))
			sb.WriteString("\n")
		}
	}
	return []widget.RichTextSegment{
		textSeg(sb.String(), widget.RichTextStyleCodeBlock),
	}
}

func (m *MarkdownRenderer) renderFencedCodeBlock(node *ast.FencedCodeBlock, source []byte) []widget.RichTextSegment {
	var sb strings.Builder
	sb.WriteString("\n")
	for line := node.FirstChild(); line != nil; line = line.NextSibling() {
		if t, ok := line.(*ast.Text); ok {
			sb.Write(t.Value(source))
			sb.WriteString("\n")
		}
	}
	return []widget.RichTextSegment{
		textSeg(sb.String(), widget.RichTextStyleCodeBlock),
	}
}

func (m *MarkdownRenderer) renderBlockquote(node *ast.Blockquote, source []byte) []widget.RichTextSegment {
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

func (m *MarkdownRenderer) renderText(node *ast.Text, source []byte) []widget.RichTextSegment {
	t := string(node.Value(source))
	if node.SoftLineBreak() {
		t += " "
	}
	return []widget.RichTextSegment{textSeg(t, widget.RichTextStyleInline)}
}

func (m *MarkdownRenderer) renderCodeSpan(node *ast.CodeSpan, source []byte) []widget.RichTextSegment {
	content := m.getChildrenText(node, source)
	return []widget.RichTextSegment{textSeg(content, widget.RichTextStyleCodeInline)}
}

func (m *MarkdownRenderer) renderEmphasis(node *ast.Emphasis, source []byte) []widget.RichTextSegment {
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

func (m *MarkdownRenderer) renderLink(node *ast.Link, source []byte) []widget.RichTextSegment {
	content := m.getChildrenText(node, source)
	u, err := url.Parse(string(node.Destination))
	if err != nil {
		return []widget.RichTextSegment{textSeg(content, widget.RichTextStyleInline)}
	}
	return []widget.RichTextSegment{
		&widget.HyperlinkSegment{
			Text: content,
			URL:  u,
		},
	}
}

func (m *MarkdownRenderer) renderAutoLink(node *ast.AutoLink, source []byte) []widget.RichTextSegment {
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

func (m *MarkdownRenderer) renderRawHTML(node *ast.RawHTML, source []byte) []widget.RichTextSegment {
	if node.Segments == nil {
		return nil
	}
	raw := node.Segments.Value(source)
	return []widget.RichTextSegment{textSeg(string(raw), widget.RichTextStyleInline)}
}

func (m *MarkdownRenderer) getChildrenText(node ast.Node, source []byte) string {
	var sb strings.Builder
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		switch n := child.(type) {
		case *ast.Text:
			sb.Write(n.Value(source))
		case *ast.String:
			sb.Write(n.Value)
		default:
			sb.WriteString(m.getChildrenText(n, source))
		}
	}
	return sb.String()
}

func textSeg(t string, style widget.RichTextStyle) *widget.TextSegment {
	return &widget.TextSegment{Text: t, Style: style}
}
