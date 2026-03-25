package chunk

import (
	"context"
	"cs-agent/internal/pkg/enums"
	"fmt"
)

type Registry struct {
	providers map[string]Provider
}

func NewRegistry() *Registry {
	return &Registry{
		providers: make(map[string]Provider),
	}
}

func NewDefaultRegistry() *Registry {
	r := NewRegistry()
	r.Register(NewFixedProvider())
	r.Register(NewStructuredProvider())
	r.Register(NewFAQProvider())
	return r
}

func (r *Registry) Register(p Provider) {
	if p == nil {
		return
	}
	r.providers[p.Name()] = p
}

func (r *Registry) Get(name string) Provider {
	if name == "" {
		return nil
	}
	return r.providers[name]
}

func (r *Registry) Resolve(name string, contentType enums.KnowledgeDocumentContentType) Provider {
	if p := r.Get(name); p != nil && p.Supports(contentType) {
		return p
	}
	if p := r.Get(string(enums.KnowledgeChunkProviderStructured)); p != nil && p.Supports(contentType) {
		return p
	}
	return r.Get(string(enums.KnowledgeChunkProviderFixed))
}

func (r *Registry) Chunk(ctx context.Context, req *ChunkRequest) ([]ChunkResult, error) {
	if req == nil {
		return nil, fmt.Errorf("chunk request is nil")
	}
	provider := r.Resolve(req.Options.Provider, req.ContentType)
	if provider == nil {
		return nil, fmt.Errorf("chunk provider not found")
	}
	return provider.Chunk(ctx, req)
}
