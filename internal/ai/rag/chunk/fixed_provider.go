package chunk

import (
	"context"
	"cs-agent/internal/pkg/enums"
)

type fixedProvider struct{}

func NewFixedProvider() Provider {
	return &fixedProvider{}
}

func (p *fixedProvider) Name() string {
	return string(enums.KnowledgeChunkProviderFixed)
}

func (p *fixedProvider) Supports(contentType enums.KnowledgeDocumentContentType) bool {
	return true
}

func (p *fixedProvider) Chunk(ctx context.Context, req *ChunkRequest) ([]ChunkResult, error) {
	text := req.PlainText
	if text == "" {
		text = req.Content
	}
	parts := splitPlainText(text, req.Options)
	results := make([]ChunkResult, 0, len(parts))
	for i, part := range parts {
		results = append(results, ChunkResult{
			ChunkNo:     i,
			Title:       req.DocumentTitle,
			Content:     part,
			ChunkType:   enums.KnowledgeChunkTypeText,
			SectionPath: req.DocumentTitle,
			CharCount:   len([]rune(part)),
			TokenCount:  estimateTokenCount(part),
			Metadata: map[string]any{
				"provider": enums.KnowledgeChunkProviderFixed,
			},
		})
	}
	return results, nil
}
