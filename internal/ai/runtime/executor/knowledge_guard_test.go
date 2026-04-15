package executor

import (
	"strings"
	"testing"

	"cs-agent/internal/ai/rag"
	"cs-agent/internal/ai/runtime/internal/impl/retrievers"
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/enums"
)

func TestBuildKnowledgeGuardDecisionFallsBackWhenKnowledgeMisses(t *testing.T) {
	agent := newKnowledgeGuardAgentFixture()
	decision := buildKnowledgeGuardDecision(&agent, &retrievers.KnowledgeRetrieveResult{
		KnowledgeBaseIDs: []int64{1},
		FallbackMode:     enums.KnowledgeFallbackModeSuggestRetry,
	})

	if decision.FallbackReply != "当前知识库里没有找到足够明确的信息，你可以换个更具体的问法再试一次。" {
		t.Fatalf("unexpected fallback reply: %q", decision.FallbackReply)
	}
	if len(decision.Instructions) != 0 {
		t.Fatalf("expected no instructions on miss, got %d", len(decision.Instructions))
	}
}

func TestBuildKnowledgeGuardDecisionUsesAgentFallbackMessage(t *testing.T) {
	agent := newKnowledgeGuardAgentFixture()
	agent.FallbackMessage = "请联系人工客服"
	decision := buildKnowledgeGuardDecision(&agent, &retrievers.KnowledgeRetrieveResult{
		KnowledgeBaseIDs: []int64{1},
		FallbackMode:     enums.KnowledgeFallbackModeNoAnswer,
	})

	if decision.FallbackReply != "请联系人工客服" {
		t.Fatalf("expected agent fallback message, got %q", decision.FallbackReply)
	}
}

func TestBuildKnowledgeGuardDecisionInjectsStrictInstructionOnHit(t *testing.T) {
	agent := newKnowledgeGuardAgentFixture()
	decision := buildKnowledgeGuardDecision(&agent, &retrievers.KnowledgeRetrieveResult{
		KnowledgeBaseIDs: []int64{1},
		Hits: []rag.RetrieveResult{
			{KnowledgeBaseID: 1, Score: 0.88},
		},
		AnswerMode:   enums.KnowledgeAnswerModeStrict,
		FallbackMode: enums.KnowledgeFallbackModeNoAnswer,
	})

	if decision.FallbackReply != "" {
		t.Fatalf("expected no fallback reply on hit, got %q", decision.FallbackReply)
	}
	if len(decision.Instructions) != 1 {
		t.Fatalf("expected one instruction, got %d", len(decision.Instructions))
	}
	content := decision.Instructions[0].Content
	if !strings.Contains(content, "只能依据后续提供的知识片段回答") {
		t.Fatalf("unexpected strict instruction: %q", content)
	}
	if !strings.Contains(content, "当前知识库暂无明确信息。") {
		t.Fatalf("expected fallback text in instruction, got %q", content)
	}
}

func newKnowledgeGuardAgentFixture() models.AIAgent {
	return models.AIAgent{}
}
