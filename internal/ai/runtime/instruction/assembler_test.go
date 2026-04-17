package instruction

import (
	"strings"
	"testing"
)

func TestAssemblerRespectsProvidedSources(t *testing.T) {
	result := NewAssembler().Assemble(AssemblerInput{
		ProjectInstruction:    "project-rule",
		GovernanceInstruction: "governance-rule",
		AgentInstruction:      "agent-rule",
		SkillInstruction:      "skill-rule",
		ToolAppendices:        []string{"tool-rule-1", "tool-rule-2"},
	})
	if !strings.Contains(result.Text, "项目级规则：\nproject-rule") {
		t.Fatalf("missing project instruction: %s", result.Text)
	}
	if !strings.Contains(result.Text, "系统治理规则：\ngovernance-rule") {
		t.Fatalf("missing governance instruction: %s", result.Text)
	}
	if !strings.Contains(result.Text, "当前技能上下文：\nskill-rule") {
		t.Fatalf("missing skill instruction: %s", result.Text)
	}
	if !strings.Contains(result.Text, "工具补充规则：\ntool-rule-1") {
		t.Fatalf("missing tool appendix: %s", result.Text)
	}
	if !result.Summary.HasProjectRule || !result.Summary.HasGovernanceRule || !result.Summary.HasAgentRule || !result.Summary.HasSkillRule || !result.Summary.HasToolRule {
		t.Fatalf("unexpected summary: %#v", result.Summary)
	}
}

func TestAssemblerUsesDefaultGovernanceInstruction(t *testing.T) {
	result := NewAssembler().Assemble(AssemblerInput{})
	expectedSnippets := []string{
		"禁止承诺未经系统确认的处理时效、完成时间、回访时间或联系时间。",
		"禁止代表人工团队、技术团队、售后团队承诺后续动作",
		"不能自行补充内部处理流程、SLA 或跟进安排。",
	}
	for _, snippet := range expectedSnippets {
		if !strings.Contains(result.Text, snippet) {
			t.Fatalf("missing governance snippet %q in assembled text: %s", snippet, result.Text)
		}
	}
	if !result.Summary.HasGovernanceRule {
		t.Fatalf("expected governance rule summary, got %#v", result.Summary)
	}
}
