package callbacks

import (
	"testing"

	"cs-agent/internal/pkg/toolx"
)

func TestParseGraphToolOutcome(t *testing.T) {
	action, risk, ready := parseGraphToolOutcome(toolx.GraphAnalyzeConversation.Code, `{"recommendedNextAction":"handoff_to_human","riskLevel":"high"}`)
	if action != "handoff_to_human" || risk != "high" || ready {
		t.Fatalf("unexpected analyze graph outcome: %q %q %v", action, risk, ready)
	}

	action, risk, ready = parseGraphToolOutcome(toolx.GraphTriageServiceRequest.Code, `{"recommendedAction":"prepare_ticket","analysis":{"riskLevel":"medium"},"ticketDraft":{"ready":true}}`)
	if action != "prepare_ticket" || risk != "medium" || !ready {
		t.Fatalf("unexpected triage graph outcome: %q %q %v", action, risk, ready)
	}
}

func TestExtractCandidateToolCodes(t *testing.T) {
	handler := &RuntimeTraceHandler{
		toolMetadataBy: map[string]ToolMetadata{
			"tool_search": {ToolCode: toolx.BuiltinToolSearch.Code, ToolName: toolx.BuiltinToolSearch.Name},
			"foo_model":   {ToolCode: "mcp/server/foo", ToolName: "foo"},
		},
	}

	got := handler.extractCandidateToolCodes(`{"selectedTools":["foo_model"]}`)
	if len(got) != 1 || got[0] != "mcp/server/foo" {
		t.Fatalf("unexpected selectedTools codes: %#v", got)
	}

	got = handler.extractCandidateToolCodes(`{"candidates":[{"toolCode":"mcp/server/bar"}]}`)
	if len(got) != 1 || got[0] != "mcp/server/bar" {
		t.Fatalf("unexpected candidate codes: %#v", got)
	}
}
