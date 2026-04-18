package factory

import (
	"testing"

	runtimeinstruction "cs-agent/internal/ai/runtime/instruction"
)

func TestBuildInstructionTraceSummary(t *testing.T) {
	got := buildInstructionTraceSummary(runtimeinstruction.AssemblySummary{
		SectionTitles:     []string{"系统治理规则", "当前技能上下文"},
		HasGovernanceRule: true,
		HasAgentRule:      true,
		HasSkillRule:      true,
		HasToolRule:       false,
	})

	if len(got.SectionTitles) != 2 {
		t.Fatalf("unexpected section titles: %#v", got.SectionTitles)
	}
	if !got.HasGovernanceRule || !got.HasAgentRule || !got.HasSkillRule {
		t.Fatalf("unexpected summary flags: %#v", got)
	}
	if got.HasToolRule {
		t.Fatalf("expected HasToolRule false, got %#v", got)
	}
}
