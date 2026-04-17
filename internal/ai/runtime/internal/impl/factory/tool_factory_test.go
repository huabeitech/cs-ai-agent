package factory

import (
	"testing"

	"cs-agent/internal/models"
)

func TestBuildMCPToolsSkipsGraphAndBuiltinTools(t *testing.T) {
	aiAgent := models.AIAgent{
		AllowedMCPTools: `[
			{"toolCode":"graph/create_ticket_with_confirmation","serverCode":"graph","toolName":"create_ticket_with_confirmation"},
			{"toolCode":"builtin/tool_search","serverCode":"builtin","toolName":"tool_search"},
			{"toolCode":"system/list_agents","serverCode":"system","toolName":"list_agents"}
		]`,
	}

	got, err := NewToolFactory().BuildMCPTools(aiAgent)
	if err != nil {
		t.Fatalf("BuildMCPTools returned error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 dynamic mcp tool, got %d: %#v", len(got), got)
	}
	if got[0].ToolCode != "system/list_agents" {
		t.Fatalf("unexpected tool code: %#v", got[0])
	}
	if got[0].ServerCode != "system" || got[0].ToolName != "list_agents" {
		t.Fatalf("unexpected tool identity: %#v", got[0])
	}
}
