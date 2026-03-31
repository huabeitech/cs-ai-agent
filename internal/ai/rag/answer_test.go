package rag

import (
	"testing"

	"cs-agent/internal/models"
	"cs-agent/internal/pkg/dto/response"
	"cs-agent/internal/pkg/enums"
)

func TestBuildFallbackAnswer(t *testing.T) {
	tests := []struct {
		name     string
		mode     enums.KnowledgeFallbackMode
		expected string
	}{
		{
			name:     "no answer",
			mode:     enums.KnowledgeFallbackModeNoAnswer,
			expected: "当前知识库暂无明确信息。",
		},
		{
			name:     "suggest retry",
			mode:     enums.KnowledgeFallbackModeSuggestRetry,
			expected: "当前知识库里没有找到足够明确的信息，你可以换个更具体的问法再试一次。",
		},
		{
			name:     "transfer human",
			mode:     enums.KnowledgeFallbackModeTransferHuman,
			expected: "当前知识库里没有找到足够明确的信息，建议转人工进一步处理。",
		},
	}

	for _, tt := range tests {
		if got := buildFallbackAnswer(tt.mode); got != tt.expected {
			t.Fatalf("%s: expected %q, got %q", tt.name, tt.expected, got)
		}
	}
}

func TestGetAnswerStatusName(t *testing.T) {
	if got := getAnswerStatusName(enums.KnowledgeAnswerStatusNoAnswer); got != "无答案" {
		t.Fatalf("expected no-answer label, got %q", got)
	}
	if got := getAnswerStatusName(enums.KnowledgeAnswerStatusFallback); got != "兜底" {
		t.Fatalf("expected fallback label, got %q", got)
	}
}

func TestResolveRerankLimit(t *testing.T) {
	tests := []struct {
		name         string
		requestLimit int
		defaultLimit int
		expected     int
	}{
		{
			name:         "request overrides default",
			requestLimit: 3,
			defaultLimit: 5,
			expected:     3,
		},
		{
			name:         "default used when request missing",
			requestLimit: 0,
			defaultLimit: 5,
			expected:     5,
		},
		{
			name:         "zero when both missing",
			requestLimit: 0,
			defaultLimit: 0,
			expected:     0,
		},
	}

	for _, tt := range tests {
		if got := resolveRerankLimit(tt.requestLimit, tt.defaultLimit); got != tt.expected {
			t.Fatalf("%s: expected %d, got %d", tt.name, tt.expected, got)
		}
	}
}

func TestResolveDefaultRerankLimit(t *testing.T) {
	items := []models.KnowledgeBase{
		{ID: 11, DefaultRerankLimit: 3},
		{ID: 22, DefaultRerankLimit: 7},
		{ID: 33, DefaultRerankLimit: 5},
	}

	if got := resolveDefaultRerankLimit(items); got != 7 {
		t.Fatalf("expected max rerank limit 7, got %d", got)
	}
}

func TestBuildKnowledgeCitations(t *testing.T) {
	hits := []response.KnowledgeSearchResult{
		{
			DocumentID:    11,
			DocumentTitle: "退款手册",
			ChunkNo:       0,
			Title:         "退款说明",
			SectionPath:   "售后 > 退款说明",
			Content:       "退款申请提交后，预计1-3个工作日到账。",
			Score:         0.91,
		},
		{
			DocumentID:    11,
			DocumentTitle: "退款手册",
			ChunkNo:       0,
			Title:         "退款说明",
			SectionPath:   "售后 > 退款说明",
			Content:       "重复内容",
			Score:         0.89,
		},
	}

	citations := buildKnowledgeCitations(hits, 3)
	if len(citations) != 1 {
		t.Fatalf("expected 1 citation, got %d", len(citations))
	}
	if citations[0].DocumentID != 11 {
		t.Fatalf("expected document id 11, got %d", citations[0].DocumentID)
	}
	if citations[0].SectionPath != "售后 > 退款说明" {
		t.Fatalf("unexpected section path: %q", citations[0].SectionPath)
	}
}
