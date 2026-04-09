package rag

import (
	"testing"

	"cs-agent/internal/models"
)

func TestResolveKnowledgeBaseSearchOptionsUsesKnowledgeBaseDefaults(t *testing.T) {
	topK, scoreThreshold := resolveKnowledgeBaseSearchOptions(RetrieveRequest{}, &models.KnowledgeBase{
		DefaultTopK:           6,
		DefaultScoreThreshold: 0.42,
	})

	if topK != 6 {
		t.Fatalf("expected topK 6, got %d", topK)
	}
	if scoreThreshold != float32(0.42) {
		t.Fatalf("expected score threshold 0.42, got %v", scoreThreshold)
	}
}

func TestResolveKnowledgeBaseSearchOptionsRequestOverridesKnowledgeBaseDefaults(t *testing.T) {
	topK, scoreThreshold := resolveKnowledgeBaseSearchOptions(RetrieveRequest{
		TopK:           9,
		ScoreThreshold: 0.55,
	}, &models.KnowledgeBase{
		DefaultTopK:           6,
		DefaultScoreThreshold: 0.42,
	})

	if topK != 9 {
		t.Fatalf("expected request topK 9, got %d", topK)
	}
	if scoreThreshold != float32(0.55) {
		t.Fatalf("expected request score threshold 0.55, got %v", scoreThreshold)
	}
}

func TestResolveKnowledgeBaseSearchOptionsUsesSystemDefaults(t *testing.T) {
	topK, scoreThreshold := resolveKnowledgeBaseSearchOptions(RetrieveRequest{}, nil)

	if topK != 8 {
		t.Fatalf("expected fallback topK 8, got %d", topK)
	}
	if scoreThreshold != float32(0.3) {
		t.Fatalf("expected fallback score threshold 0.3, got %v", scoreThreshold)
	}
}
