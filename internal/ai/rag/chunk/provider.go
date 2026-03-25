package chunk

import (
	"context"
	"cs-agent/internal/pkg/enums"
)

type Provider interface {
	Name() string
	Supports(contentType enums.KnowledgeDocumentContentType) bool
	Chunk(ctx context.Context, req *ChunkRequest) ([]ChunkResult, error)
}
