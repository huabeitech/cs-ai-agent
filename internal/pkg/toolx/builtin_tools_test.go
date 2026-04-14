package toolx

import "testing"

func TestResolveToolMetadata(t *testing.T) {
	item := ResolveToolMetadata("builtin/create_ticket_with_confirmation", "")
	if item.ToolCode != GraphCreateTicketConfirm.Code {
		t.Fatalf("unexpected tool code: %s", item.ToolCode)
	}
	if item.ServerCode != GraphCreateTicketConfirm.ServerCode {
		t.Fatalf("unexpected server code: %s", item.ServerCode)
	}
	if item.ToolName != GraphCreateTicketConfirm.Name {
		t.Fatalf("unexpected tool name: %s", item.ToolName)
	}
	if item.SourceType != GraphCreateTicketConfirm.SourceType {
		t.Fatalf("unexpected source type: %s", item.SourceType)
	}
}

func TestResolveToolMetadataFallsBackToName(t *testing.T) {
	item := ResolveToolMetadata("mcp/demo_tool", "demo_tool")
	if item.ToolCode != "mcp/demo_tool" {
		t.Fatalf("unexpected tool code: %s", item.ToolCode)
	}
	if item.ServerCode != "" {
		t.Fatalf("unexpected server code: %s", item.ServerCode)
	}
	if item.ToolName != "demo_tool" {
		t.Fatalf("unexpected tool name: %s", item.ToolName)
	}
	if item.SourceType != "mcp" {
		t.Fatalf("unexpected source type: %s", item.SourceType)
	}
}
