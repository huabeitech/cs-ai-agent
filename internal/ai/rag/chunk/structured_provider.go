package chunk

import (
	"context"
	"cs-agent/internal/pkg/enums"
	"strings"

	"github.com/gomarkdown/markdown"
	"golang.org/x/net/html"
)

type structuredProvider struct{}

type contentBlock struct {
	Type        string
	Level       int
	Text        string
	Title       string
	SectionPath string
}

func NewStructuredProvider() Provider {
	return &structuredProvider{}
}

func (p *structuredProvider) Name() string {
	return string(enums.KnowledgeChunkProviderStructured)
}

func (p *structuredProvider) Supports(contentType enums.KnowledgeDocumentContentType) bool {
	switch contentType {
	case enums.KnowledgeDocumentContentTypeHTML, enums.KnowledgeDocumentContentTypeMarkdown:
		return true
	default:
		return false
	}
}

func (p *structuredProvider) Chunk(ctx context.Context, req *ChunkRequest) ([]ChunkResult, error) {
	content := req.Content
	if req.ContentType == enums.KnowledgeDocumentContentTypeMarkdown {
		content = string(markdown.ToHTML([]byte(content), nil, nil))
	}

	blocks := parseStructuredBlocks(content, req.DocumentTitle)
	if len(blocks) == 0 {
		return NewFixedProvider().Chunk(ctx, req)
	}

	results := make([]ChunkResult, 0)
	chunkNo := 0
	for _, block := range blocks {
		parts := splitPlainText(block.Text, req.Options)
		for _, part := range parts {
			if part == "" {
				continue
			}
			results = append(results, ChunkResult{
				ChunkNo:     chunkNo,
				Title:       block.Title,
				Content:     part,
				ChunkType:   mapBlockType(block.Type),
				SectionPath: block.SectionPath,
				CharCount:   len([]rune(part)),
				TokenCount:  estimateTokenCount(part),
				Metadata: map[string]any{
					"provider":     enums.KnowledgeChunkProviderStructured,
					"blockType":    block.Type,
					"sectionPath":  block.SectionPath,
					"sectionTitle": block.Title,
				},
			})
			chunkNo++
		}
	}
	if len(results) == 0 {
		return NewFixedProvider().Chunk(ctx, req)
	}
	return results, nil
}

func parseStructuredBlocks(content string, documentTitle string) []contentBlock {
	content = strings.TrimSpace(content)
	if content == "" {
		return nil
	}

	parent := &html.Node{Type: html.ElementNode, Data: "div"}
	nodes, err := html.ParseFragment(strings.NewReader(content), parent)
	if err != nil {
		return nil
	}

	var blocks []contentBlock
	headings := make([]string, 0)
	var walk func(node *html.Node)
	walk = func(node *html.Node) {
		if node == nil {
			return
		}
		if node.Type == html.ElementNode {
			switch node.Data {
			case "h1", "h2", "h3", "h4", "h5", "h6":
				title := normalizeText(nodeText(node))
				if title != "" {
					level := int(node.Data[1] - '0')
					if level <= 0 {
						level = 1
					}
					headings = updateHeadingPath(headings, level, title)
				}
				return
			case "p":
				appendBlock(&blocks, "paragraph", normalizeText(nodeText(node)), currentTitle(headings, documentTitle), strings.Join(headings, " > "))
				return
			case "ul", "ol":
				appendBlock(&blocks, "list", normalizeText(listText(node)), currentTitle(headings, documentTitle), strings.Join(headings, " > "))
				return
			case "table":
				appendBlock(&blocks, "table", normalizeText(tableText(node)), currentTitle(headings, documentTitle), strings.Join(headings, " > "))
				return
			case "pre", "code":
				appendBlock(&blocks, "code", normalizeText(nodeText(node)), currentTitle(headings, documentTitle), strings.Join(headings, " > "))
				return
			}
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	for _, node := range nodes {
		walk(node)
	}
	return blocks
}

func appendBlock(blocks *[]contentBlock, blockType string, text string, title string, sectionPath string) {
	text = normalizeText(text)
	if text == "" {
		return
	}
	if sectionPath == "" {
		sectionPath = title
	}
	*blocks = append(*blocks, contentBlock{
		Type:        blockType,
		Text:        text,
		Title:       title,
		SectionPath: sectionPath,
	})
}

func updateHeadingPath(headings []string, level int, title string) []string {
	if level <= 0 {
		level = 1
	}
	if len(headings) >= level {
		headings = headings[:level-1]
	}
	headings = append(headings, title)
	return headings
}

func currentTitle(headings []string, documentTitle string) string {
	if len(headings) == 0 {
		return documentTitle
	}
	return headings[len(headings)-1]
}

func mapBlockType(blockType string) enums.KnowledgeChunkType {
	switch blockType {
	case "table":
		return enums.KnowledgeChunkTypeTable
	case "code":
		return enums.KnowledgeChunkTypeCode
	default:
		return enums.KnowledgeChunkTypeText
	}
}

func nodeText(node *html.Node) string {
	if node == nil {
		return ""
	}
	var builder strings.Builder
	writeNodeText(&builder, node)
	return builder.String()
}

func writeNodeText(builder *strings.Builder, node *html.Node) {
	if node == nil {
		return
	}
	switch node.Type {
	case html.TextNode:
		builder.WriteString(node.Data)
	case html.ElementNode:
		if shouldSeparate(node.Data) {
			builder.WriteByte(' ')
		}
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		writeNodeText(builder, child)
	}
	if node.Type == html.ElementNode && shouldSeparate(node.Data) {
		builder.WriteByte(' ')
	}
}

func shouldSeparate(tag string) bool {
	switch tag {
	case "p", "div", "br", "li", "ul", "ol", "blockquote", "pre", "table", "tr", "td", "th", "h1", "h2", "h3", "h4", "h5", "h6":
		return true
	default:
		return false
	}
}

func listText(node *html.Node) string {
	items := make([]string, 0)
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == html.ElementNode && child.Data == "li" {
			item := normalizeText(nodeText(child))
			if item != "" {
				items = append(items, item)
			}
		}
	}
	return strings.Join(items, " ")
}

func tableText(node *html.Node) string {
	rows := make([]string, 0)
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n == nil {
			return
		}
		if n.Type == html.ElementNode && n.Data == "tr" {
			cells := make([]string, 0)
			for child := n.FirstChild; child != nil; child = child.NextSibling {
				if child.Type == html.ElementNode && (child.Data == "td" || child.Data == "th") {
					cell := normalizeText(nodeText(child))
					if cell != "" {
						cells = append(cells, cell)
					}
				}
			}
			if len(cells) > 0 {
				rows = append(rows, strings.Join(cells, " | "))
			}
			return
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(node)
	return strings.Join(rows, " ")
}
