package services

import (
	"testing"

	"cs-agent/internal/pkg/dto/request"
)

func TestBuildKnowledgeBaseModelUsesLowerDefaultScoreThreshold(t *testing.T) {
	item, err := KnowledgeBaseService.buildKnowledgeBaseModel(request.CreateKnowledgeBaseRequest{})
	if err != nil {
		t.Fatalf("build knowledge base model failed: %v", err)
	}
	if item.DefaultScoreThreshold != 0.2 {
		t.Fatalf("expected default score threshold 0.2, got %v", item.DefaultScoreThreshold)
	}
}
