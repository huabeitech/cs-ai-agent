package utils

import (
	"strings"

	"golang.org/x/net/html"
)

type ContentChunkType string

const (
	ContentChunkTypeText  ContentChunkType = "text"
	ContentChunkTypeImage ContentChunkType = "image"
)

type ContentChunk struct {
	Type    ContentChunkType //
	Content string           // text content or image url
}

func SplitHTMLContentChunks(content string) ([]ContentChunk, error) {
	root, err := html.Parse(strings.NewReader("<div>" + strings.TrimSpace(content) + "</div>"))
	if err != nil {
		return nil, err
	}

	var (
		chunks []ContentChunk
		buffer strings.Builder
	)
	flushText := func() {
		text := normalizeContentChunkText(buffer.String())
		buffer.Reset()
		if text == "" {
			return
		}
		chunks = append(chunks, ContentChunk{
			Type:    ContentChunkTypeText,
			Content: text,
		})
	}

	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node == nil {
			return
		}
		switch node.Type {
		case html.TextNode:
			buffer.WriteString(node.Data)
		case html.ElementNode:
			switch node.Data {
			case "br":
				buffer.WriteString("\n")
			case "p", "div", "li", "blockquote":
				buffer.WriteString("\n")
			case "img":
				flushText()
				src := strings.TrimSpace(findContentChunkHTMLAttr(node, "src"))
				if src == "" {
					return
				}
				chunks = append(chunks, ContentChunk{
					Type:    ContentChunkTypeImage,
					Content: src,
				})
				return
			}
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
		if node.Type == html.ElementNode {
			switch node.Data {
			case "p", "div", "li", "blockquote":
				buffer.WriteString("\n")
			}
		}
	}
	walk(root)
	flushText()
	return chunks, nil
}

func normalizeContentChunkText(value string) string {
	value = strings.ReplaceAll(value, "\u00a0", " ")
	lines := strings.Split(value, "\n")
	normalizedLines := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.Join(strings.Fields(strings.TrimSpace(line)), " ")
		if line == "" {
			if len(normalizedLines) > 0 && normalizedLines[len(normalizedLines)-1] != "" {
				normalizedLines = append(normalizedLines, "")
			}
			continue
		}
		normalizedLines = append(normalizedLines, line)
	}
	for len(normalizedLines) > 0 && normalizedLines[0] == "" {
		normalizedLines = normalizedLines[1:]
	}
	for len(normalizedLines) > 0 && normalizedLines[len(normalizedLines)-1] == "" {
		normalizedLines = normalizedLines[:len(normalizedLines)-1]
	}
	return strings.Join(normalizedLines, "\n")
}

func findContentChunkHTMLAttr(node *html.Node, key string) string {
	for _, attr := range node.Attr {
		if strings.EqualFold(attr.Key, key) {
			return attr.Val
		}
	}
	return ""
}
