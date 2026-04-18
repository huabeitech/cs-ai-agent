package instruction

import (
	"strings"
	"testing"
)

func TestAssemblerRespectsProvidedSources(t *testing.T) {
	result := NewAssembler().Assemble(AssemblerInput{
		GovernanceInstruction: "governance-rule",
		AgentInstruction:      "agent-rule",
		SkillInstruction:      "skill-rule",
		ToolAppendices:        []string{"tool-rule-1", "tool-rule-2"},
	})
	if !strings.Contains(result.Text, "系统治理规则：\ngovernance-rule") {
		t.Fatalf("missing governance instruction: %s", result.Text)
	}
	if !strings.Contains(result.Text, "当前技能上下文：\nskill-rule") {
		t.Fatalf("missing skill instruction: %s", result.Text)
	}
	if !strings.Contains(result.Text, "工具补充规则：\ntool-rule-1") {
		t.Fatalf("missing tool appendix: %s", result.Text)
	}
	if !result.Summary.HasGovernanceRule || !result.Summary.HasAgentRule || !result.Summary.HasSkillRule || !result.Summary.HasToolRule {
		t.Fatalf("unexpected summary: %#v", result.Summary)
	}
}

func TestAssemblerDoesNotInjectGovernanceInstructionWhenInputIsEmpty(t *testing.T) {
	result := NewAssembler().Assemble(AssemblerInput{})
	if result.Text != "" {
		t.Fatalf("expected empty assembled text, got: %s", result.Text)
	}
	if result.Summary.HasGovernanceRule {
		t.Fatalf("expected no governance rule summary, got %#v", result.Summary)
	}
}
