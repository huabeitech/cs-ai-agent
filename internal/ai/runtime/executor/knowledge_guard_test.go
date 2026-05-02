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
	agent.FallbackMode = enums.AIAgentFallbackModeSuggestRetry
	decision := buildKnowledgeGuardDecision(agent, &retrievers.KnowledgeRetrieveResult{
		KnowledgeBaseIDs: []int64{1},
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
	decision := buildKnowledgeGuardDecision(agent, &retrievers.KnowledgeRetrieveResult{
		KnowledgeBaseIDs: []int64{1},
	})

	if decision.FallbackReply != "请联系人工客服" {
		t.Fatalf("expected agent fallback message, got %q", decision.FallbackReply)
	}
}

func TestBuildKnowledgeGuardDecisionInjectsStrictInstructionOnHit(t *testing.T) {
	agent := newKnowledgeGuardAgentFixture()
	decision := buildKnowledgeGuardDecision(agent, &retrievers.KnowledgeRetrieveResult{
		KnowledgeBaseIDs: []int64{1},
		Hits: []rag.RetrieveResult{
			{KnowledgeBaseID: 1, Score: 0.88},
		},
		ContextText: "知识库上下文",
		AnswerMode:  enums.KnowledgeAnswerModeStrict,
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

func TestBuildKnowledgeGuardDecisionFallsBackWhenHitHasNoContext(t *testing.T) {
	agent := newKnowledgeGuardAgentFixture()
	agent.FallbackMessage = "我暂时没有找到足够准确的信息。"
	decision := buildKnowledgeGuardDecision(agent, &retrievers.KnowledgeRetrieveResult{
		KnowledgeBaseIDs: []int64{1},
		Hits: []rag.RetrieveResult{
			{KnowledgeBaseID: 1, Score: 0.88},
		},
		AnswerMode: enums.KnowledgeAnswerModeStrict,
	})

	if decision.FallbackReply != "我暂时没有找到足够准确的信息。" {
		t.Fatalf("expected fallback on empty context, got %q", decision.FallbackReply)
	}
	if len(decision.Instructions) != 0 {
		t.Fatalf("expected no instructions on empty context, got %d", len(decision.Instructions))
	}
}

func TestBuildKnowledgeUnavailableDecisionFallsBackWhenAgentHasKnowledge(t *testing.T) {
	agent := newKnowledgeGuardAgentFixture()
	agent.FallbackMessage = "知识库暂时不可用。"
	decision := buildKnowledgeUnavailableDecision(agent, []int64{1})

	if decision.FallbackReply != "知识库暂时不可用。" {
		t.Fatalf("expected fallback when knowledge unavailable, got %q", decision.FallbackReply)
	}
}

func TestBuildKnowledgeUnavailableDecisionSkipsWhenAgentHasNoKnowledge(t *testing.T) {
	decision := buildKnowledgeUnavailableDecision(newKnowledgeGuardAgentFixture(), nil)

	if decision.FallbackReply != "" {
		t.Fatalf("expected no fallback without knowledge bases, got %q", decision.FallbackReply)
	}
}

func TestResolveKnowledgeHumanSupportFallbackUsesAgentMessage(t *testing.T) {
	agent := newKnowledgeGuardAgentFixture()
	agent.FallbackMessage = "我暂时没有找到足够准确的信息。"

	got := resolveKnowledgeHumanSupportFallback(agent)

	want := "我暂时没有找到足够准确的信息。 建议你联系人工客服进一步确认。"
	if got != want {
		t.Fatalf("unexpected fallback: %q", got)
	}
}

func TestResolveKnowledgeHumanSupportFallbackUsesDefault(t *testing.T) {
	agent := newKnowledgeGuardAgentFixture()

	got := resolveKnowledgeHumanSupportFallback(agent)

	want := "当前知识库暂无明确信息。 建议你联系人工客服进一步确认。"
	if got != want {
		t.Fatalf("unexpected fallback: %q", got)
	}
}

func newKnowledgeGuardAgentFixture() models.AIAgent {
	return models.AIAgent{
		FallbackMode: enums.AIAgentFallbackModeNoAnswer,
	}
}
