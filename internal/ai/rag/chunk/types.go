package chunk

import "cs-agent/internal/pkg/enums"

type ChunkRequest struct {
	KnowledgeBaseID int64
	DocumentID      int64
	DocumentTitle   string
	ContentType     enums.KnowledgeDocumentContentType
	Content         string
	PlainText       string
	Options         ChunkOptions
}

type ChunkOptions struct {
	Provider       string
	TargetTokens   int
	MaxTokens      int
	OverlapTokens  int
	EnableFallback bool
}

type ChunkResult struct {
	ChunkNo     int
	Title       string
	Content     string
	ChunkType   enums.KnowledgeChunkType
	SectionPath string
	CharCount   int
	TokenCount  int
	Metadata    map[string]any
}
