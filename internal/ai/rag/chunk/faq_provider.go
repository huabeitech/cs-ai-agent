package chunk

import (
	"context"
	"cs-agent/internal/pkg/enums"
	"regexp"
	"strings"
)

var faqQuestionPattern = regexp.MustCompile(`^(q[:：]\s*|问[:：]\s*|问题[:：]\s*)`)
var faqAnswerPattern = regexp.MustCompile(`^(a[:：]\s*|答[:：]\s*|答案[:：]\s*)`)

type faqProvider struct{}

func NewFAQProvider() Provider {
	return &faqProvider{}
}

func (p *faqProvider) Name() string {
	return string(enums.KnowledgeChunkProviderFAQ)
}

func (p *faqProvider) Supports(contentType enums.KnowledgeDocumentContentType) bool {
	switch contentType {
	case enums.KnowledgeDocumentContentTypeHTML, enums.KnowledgeDocumentContentTypeMarkdown:
		return true
	default:
		return false
	}
}

func (p *faqProvider) Chunk(ctx context.Context, req *ChunkRequest) ([]ChunkResult, error) {
	text := req.PlainText
	if text == "" {
		text = req.Content
	}
	items := parseFAQItems(text)
	if len(items) == 0 {
		return NewStructuredProvider().Chunk(ctx, req)
	}

	results := make([]ChunkResult, 0, len(items))
	chunkNo := 0
	for _, item := range items {
		content := strings.TrimSpace("问题：" + item.Question + "\n回答：" + item.Answer)
		for _, part := range splitPlainText(content, req.Options) {
			part = normalizeText(part)
			if part == "" {
				continue
			}
			results = append(results, ChunkResult{
				ChunkNo:     chunkNo,
				Title:       firstNonEmpty(item.Question, req.DocumentTitle),
				Content:     part,
				ChunkType:   enums.KnowledgeChunkTypeFAQ,
				SectionPath: firstNonEmpty(item.Question, req.DocumentTitle),
				CharCount:   len([]rune(part)),
				TokenCount:  estimateTokenCount(part),
				Metadata: map[string]any{
					"provider": string(enums.KnowledgeChunkProviderFAQ),
					"question": item.Question,
				},
			})
			chunkNo++
		}
	}
	if len(results) == 0 {
		return NewStructuredProvider().Chunk(ctx, req)
	}
	return results, nil
}

type faqItem struct {
	Question string
	Answer   string
}

func parseFAQItems(text string) []faqItem {
	lines := strings.Split(text, "\n")
	items := make([]faqItem, 0)
	current := faqItem{}
	inAnswer := false

	flush := func() {
		current.Question = normalizeFAQPrefix(current.Question, faqQuestionPattern)
		current.Answer = normalizeFAQPrefix(current.Answer, faqAnswerPattern)
		current.Question = normalizeText(current.Question)
		current.Answer = normalizeText(current.Answer)
		if current.Question != "" && current.Answer != "" {
			items = append(items, current)
		}
		current = faqItem{}
		inAnswer = false
	}

	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}
		switch {
		case faqQuestionPattern.MatchString(strings.ToLower(line)):
			if current.Question != "" && current.Answer != "" {
				flush()
			}
			current.Question = line
			inAnswer = false
		case faqAnswerPattern.MatchString(strings.ToLower(line)):
			if current.Question == "" {
				continue
			}
			if current.Answer != "" {
				current.Answer += "\n"
			}
			current.Answer += line
			inAnswer = true
		default:
			if current.Question == "" {
				continue
			}
			if inAnswer {
				current.Answer += "\n" + line
			} else {
				current.Question += " " + line
			}
		}
	}
	flush()
	return items
}

func normalizeFAQPrefix(text string, pattern *regexp.Regexp) string {
	return strings.TrimSpace(pattern.ReplaceAllString(text, ""))
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}
