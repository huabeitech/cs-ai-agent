package skills

import (
	"strings"
	"testing"

	"cs-agent/internal/models"
	"cs-agent/internal/pkg/enums"
)

func TestBuildRunLogMatchedPlan(t *testing.T) {
	log := BuildRunLog(
		RuntimeContext{
			ConversationID:  11,
			AIAgentID:       22,
			ManualSkillCode: "manual_refund",
			IntentCode:      "refund",
			UserMessage:     "我要退款",
		},
		&ExecutionPlan{
			AIAgent: &models.AIAgent{ID: 22},
			AIConfig: &models.AIConfig{
				ID:        33,
				ModelName: "gpt-test",
				Provider:  enums.AIProviderOpenAI,
			},
			Skill: &models.SkillDefinition{
				ID:   44,
				Code: "refund_skill",
			},
			MatchReason: "llm_route",
		},
		&ExecutionTrace{Status: "ok"},
		nil,
	)

	if log == nil {
		t.Fatalf("expected run log")
	}
	if log.ConversationID != 11 || log.AIAgentID != 22 || log.AIConfigID != 33 {
		t.Fatalf("unexpected ids in run log: %#v", log)
	}
	if !log.Matched || !log.FinalSelected || log.SkillCode != "refund_skill" {
		t.Fatalf("expected matched skill log, got %#v", log)
	}
	if log.MatchReason != "llm_route" {
		t.Fatalf("unexpected match reason: %q", log.MatchReason)
	}
	if !strings.Contains(log.TraceData, `"status":"ok"`) {
		t.Fatalf("expected trace data to contain status, got %q", log.TraceData)
	}
}

func TestBuildRunLogNotMatchedAndError(t *testing.T) {
	log := BuildRunLog(
		RuntimeContext{
			AIAgentID:   22,
			UserMessage: "随便问问",
		},
		nil,
		&ExecutionTrace{Status: "route_error"},
		assertErr("route failed"),
	)

	if log == nil {
		t.Fatalf("expected run log")
	}
	if log.Matched {
		t.Fatalf("expected unmatched log")
	}
	if log.ErrorMessage != "route failed" {
		t.Fatalf("unexpected error message: %q", log.ErrorMessage)
	}

	noMatchLog := BuildRunLog(
		RuntimeContext{AIAgentID: 22, UserMessage: "随便问问"},
		&ExecutionPlan{MatchReason: ""},
		&ExecutionTrace{Status: "not_matched"},
		nil,
	)
	if noMatchLog.MatchReason != "not_matched" {
		t.Fatalf("expected default not_matched reason, got %q", noMatchLog.MatchReason)
	}
}

type assertErr string

func (e assertErr) Error() string {
	return string(e)
}
