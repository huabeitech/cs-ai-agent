package runtime

import (
	"strings"
	"testing"

	"cs-agent/internal/pkg/toolx"
)

func TestSummaryPrimaryToolCodePrefersToolSearchTarget(t *testing.T) {
	summary := &Summary{
		InvokedToolCodes: []string{toolx.BuiltinToolSearch.Code},
		TraceData: `{
			"toolSearch": {
				"items": [
					{"targetToolCode":"mcp/server/tool_a"}
				]
			}
		}`,
	}

	if got := summaryPrimaryToolCode(summary); got != "mcp/server/tool_a" {
		t.Fatalf("unexpected primary tool code: %q", got)
	}
}

func TestToRunLogFinalAction(t *testing.T) {
	if got := toRunLogFinalAction(&Summary{PlannedSkillCode: "refund", ReplyText: "ok"}); got != "skill" {
		t.Fatalf("expected skill final action, got %q", got)
	}

	graphSummary := &Summary{
		ReplyText: "ok",
		TraceData: `{
			"graphTools": {
				"items": [
					{"toolCode":"` + toolx.GraphAnalyzeConversation.Code + `"}
				]
			}
		}`,
	}
	if got := toRunLogFinalAction(graphSummary); got != "graph" {
		t.Fatalf("expected graph final action, got %q", got)
	}

	if got := toRunLogFinalAction(&Summary{Status: "fallback"}); got != "fallback" {
		t.Fatalf("expected fallback final action, got %q", got)
	}
}

func TestExtractInterruptMessageAndCheckpointError(t *testing.T) {
	if got := extractInterruptMessage(`{"message":"请补充订单号"}`); got != "请补充订单号" {
		t.Fatalf("unexpected interrupt message: %q", got)
	}
	if got := extractInterruptMessage("not-json"); got != "" {
		t.Fatalf("expected empty message for invalid json, got %q", got)
	}

	err := fakeErr("Failed to load from checkpoint: record does not exist")
	if !isCheckpointMissingError(err) {
		t.Fatalf("expected checkpoint missing error to be detected")
	}
	if isCheckpointMissingError(fakeErr("other error")) {
		t.Fatalf("expected unrelated error to be ignored")
	}
}

func TestGraphPlanReason(t *testing.T) {
	summary := &Summary{
		TraceData: `{
			"graphTools": {
				"items": [
					{
						"toolCode":"` + toolx.GraphTriageServiceRequest.Code + `",
						"recommendedAction":"create_ticket",
						"ticketDraftReady": true
					}
				]
			}
		}`,
	}
	got := graphPlanReason(summary)
	if !strings.Contains(got, "create_ticket") || !strings.Contains(got, "ready ticket draft") {
		t.Fatalf("unexpected graph plan reason: %q", got)
	}
}

type fakeErr string

func (e fakeErr) Error() string {
	return string(e)
}
