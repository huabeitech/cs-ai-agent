package registry_test

import (
	"context"
	"testing"

	"cs-agent/internal/ai/runtime/registry"
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/toolx"

	einotool "github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

type stubTool struct {
	name string
	code string
}

func (t stubTool) Spec() toolx.ToolSpec {
	return toolx.ToolSpec{
		Code:       t.code,
		Name:       t.name,
		ServerCode: toolx.GraphCreateTicketConfirm.ServerCode,
		SourceType: toolx.GraphCreateTicketConfirm.SourceType,
	}
}

func (t stubTool) Name() string { return t.name }
func (t stubTool) Code() string { return t.code }
func (t stubTool) Enabled(registry.Context) bool {
	return true
}
func (t stubTool) Build(registry.Context) (einotool.BaseTool, error) {
	return stubBaseTool{name: t.name}, nil
}

type stubBaseTool struct {
	name string
}

func (t stubBaseTool) Info(context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{Name: t.name}, nil
}

func TestResolveBuildsStaticToolMetadata(t *testing.T) {
	r := registry.NewRegistry(stubTool{
		name: toolx.GraphCreateTicketConfirm.Name,
		code: toolx.GraphCreateTicketConfirm.Code,
	})
	toolSet, err := r.Resolve(registry.Context{
		Conversation: models.Conversation{ID: 1},
		AIAgent:      models.AIAgent{ID: 1},
	})
	if err != nil {
		t.Fatalf("resolve returned error: %v", err)
	}
	if toolSet == nil {
		t.Fatalf("expected tool set")
	}
	if len(toolSet.StaticToolMetadata) != 1 {
		t.Fatalf("expected 1 metadata item, got %d", len(toolSet.StaticToolMetadata))
	}
	item, ok := toolSet.StaticToolMetadata[toolx.GraphCreateTicketConfirm.Name]
	if !ok {
		t.Fatalf("missing metadata for %s", toolx.GraphCreateTicketConfirm.Name)
	}
	if item.ToolCode != toolx.GraphCreateTicketConfirm.Code {
		t.Fatalf("unexpected tool code: %s", item.ToolCode)
	}
	if item.ServerCode != toolx.GraphCreateTicketConfirm.ServerCode {
		t.Fatalf("unexpected server code: %s", item.ServerCode)
	}
	if item.ToolName != toolx.GraphCreateTicketConfirm.Name {
		t.Fatalf("unexpected tool name: %s", item.ToolName)
	}
	if item.SourceType != toolx.GraphCreateTicketConfirm.SourceType {
		t.Fatalf("unexpected source type: %s", item.SourceType)
	}
}
