package runtime

import (
	"testing"

	"cs-agent/internal/models"
)

func TestNormalizeAllowedToolCodes(t *testing.T) {
	ret := normalizeAllowedToolCodes([]string{
		" ",
		"graph/create_ticket_with_confirmation",
		"builtin/create_ticket_with_confirmation",
		"graph/handoff_to_human",
		"graph/handoff_to_human",
	})
	if len(ret) != 2 {
		t.Fatalf("expected 2 tool codes, got %d: %#v", len(ret), ret)
	}
	if ret[0] != "graph/create_ticket_with_confirmation" {
		t.Fatalf("unexpected first tool code: %s", ret[0])
	}
	if ret[1] != "graph/handoff_to_human" {
		t.Fatalf("unexpected second tool code: %s", ret[1])
	}
}

func TestToolCatalogResolveAllowedToolCodes(t *testing.T) {
	catalog := newToolCatalog()
	agent := &models.AIAgent{
		AllowedMCPTools: `[{"toolCode":"graph/create_ticket_with_confirmation"},{"toolCode":"graph/handoff_to_human"}]`,
	}
	skill := &models.SkillDefinition{
		ToolWhitelist: `["builtin/create_ticket_with_confirmation","graph/prepare_ticket_draft"]`,
	}
	ret := catalog.resolveAllowedToolCodes(agent, skill)
	if len(ret) != 1 {
		t.Fatalf("expected 1 tool code, got %d: %#v", len(ret), ret)
	}
	if ret[0] != "graph/create_ticket_with_confirmation" {
		t.Fatalf("unexpected tool code: %s", ret[0])
	}
}

func TestToolCatalogResolveAllowedToolCodesFallsBackWhenSkillEmpty(t *testing.T) {
	catalog := newToolCatalog()
	agent := &models.AIAgent{
		AllowedMCPTools: `[{"toolCode":"graph/create_ticket_with_confirmation"},{"toolCode":"graph/handoff_to_human"}]`,
	}
	ret := catalog.resolveAllowedToolCodes(agent, nil)
	if len(ret) != 2 {
		t.Fatalf("expected 2 tool codes, got %d: %#v", len(ret), ret)
	}
}
