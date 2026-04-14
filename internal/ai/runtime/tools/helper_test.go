package tools

import (
	"testing"

	"cs-agent/internal/pkg/toolx"
)

func TestNewRuntimeStaticTool(t *testing.T) {
	items := []string{
		toolx.GraphTriageServiceRequest.Code,
		toolx.GraphAnalyzeConversation.Code,
		toolx.GraphPrepareTicketDraft.Code,
		toolx.GraphCreateTicketConfirm.Code,
		toolx.GraphHandoffConversation.Code,
	}
	for _, item := range items {
		tool := NewRuntimeStaticTool(item)
		if tool == nil {
			t.Fatalf("expected runtime static tool for %s", item)
		}
		if tool.Spec().Code != item {
			t.Fatalf("unexpected tool code for %s: %s", item, tool.Spec().Code)
		}
	}
}

func TestNewRuntimeStaticToolReturnsNilForUnknownTool(t *testing.T) {
	if tool := NewRuntimeStaticTool("builtin/unknown_tool"); tool != nil {
		t.Fatalf("expected nil tool for unknown tool code")
	}
}
