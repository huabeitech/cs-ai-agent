package rag

import (
	"cs-agent/internal/pkg/enums"
	"fmt"
	"strings"

	"github.com/spf13/cast"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	htmlparser "golang.org/x/net/html"
)

var plainTextMarkdown = goldmark.New(
	goldmark.WithExtensions(extension.GFM),
	goldmark.WithParserOptions(
		parser.WithAutoHeadingID(),
	),
	goldmark.WithRendererOptions(
		html.WithHardWraps(),
		html.WithXHTML(),
	),
)

func ExtractPlainText(content string, contentType enums.KnowledgeDocumentContentType) string {
	switch contentType {
	case enums.KnowledgeDocumentContentTypeMarkdown:
		return ExtractPlainTextFromMarkdown(content)
	case enums.KnowledgeDocumentContentTypeHTML:
		return ExtractPlainTextFromHTML(content)
	default:
		return normalizeWhitespace(content)
	}
}

func ExtractPlainTextFromMarkdown(content string) string {
	content = normalizeWhitespace(content)
	if content == "" {
		return ""
	}

	var buf strings.Builder
	if err := plainTextMarkdown.Convert([]byte(content), &buf); err != nil {
		return normalizeWhitespace(content)
	}
	return ExtractPlainTextFromHTML(buf.String())
}

func ExtractPlainTextFromHTML(content string) string {
	content = strings.TrimSpace(content)
	if content == "" {
		return ""
	}

	parent := &htmlparser.Node{
		Type: htmlparser.ElementNode,
		Data: "div",
	}
	nodes, err := htmlparser.ParseFragment(strings.NewReader(content), parent)
	if err != nil {
		return normalizeWhitespace(content)
	}

	var builder strings.Builder
	for _, node := range nodes {
		writeHTMLNodeText(&builder, node)
	}
	return normalizeWhitespace(builder.String())
}

func ExtractStringPayload(payload map[string]interface{}, key string) string {
	if payload == nil {
		return ""
	}
	value, ok := payload[key]
	if !ok || value == nil {
		return ""
	}
	if str, ok := value.(string); ok {
		return str
	}
	return normalizeWhitespace(fmt.Sprintf("%v", value))
}

func ExtractInt64Payload(payload map[string]interface{}, key string) int64 {
	if payload == nil {
		return 0
	}
	if value, ok := payload[key]; ok {
		return cast.ToInt64(value)
	}
	return 0
}

func ExtractIntPayload(payload map[string]interface{}, key string) int {
	if payload == nil {
		return 0
	}
	if value, ok := payload[key]; ok {
		return cast.ToInt(value)
	}
	return 0
}

func writeHTMLNodeText(builder *strings.Builder, node *htmlparser.Node) {
	if node == nil {
		return
	}

	switch node.Type {
	case htmlparser.TextNode:
		builder.WriteString(node.Data)
	case htmlparser.ElementNode:
		if shouldSeparateHTMLText(node.Data) {
			builder.WriteByte(' ')
		}
	}

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		writeHTMLNodeText(builder, child)
	}

	if node.Type == htmlparser.ElementNode && shouldSeparateHTMLText(node.Data) {
		builder.WriteByte(' ')
	}
}

func shouldSeparateHTMLText(tag string) bool {
	switch tag {
	case "p", "div", "br", "li", "ul", "ol", "blockquote", "pre", "table", "tr", "td", "th", "h1", "h2", "h3", "h4", "h5", "h6":
		return true
	default:
		return false
	}
}

func normalizeWhitespace(content string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(content)), " ")
}
