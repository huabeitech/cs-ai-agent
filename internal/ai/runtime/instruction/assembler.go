package instruction

import "strings"

type Assembler struct{}

type AssemblerInput struct {
	AgentInstruction      string
	GovernanceInstruction string
	SkillInstruction      string
	ToolAppendices        []string
}

type AssemblySummary struct {
	SectionTitles     []string
	HasGovernanceRule bool
	HasAgentRule      bool
	HasSkillRule      bool
	HasToolRule       bool
}

type AssemblyResult struct {
	Text    string
	Summary AssemblySummary
}

func NewAssembler() *Assembler {
	return &Assembler{}
}

func (a *Assembler) Build(input AssemblerInput) string {
	return a.Assemble(input).Text
}

func (a *Assembler) Assemble(input AssemblerInput) AssemblyResult {
	parts := make([]string, 0, 4)
	summary := AssemblySummary{SectionTitles: make([]string, 0, 4)}
	governanceInstruction := strings.TrimSpace(input.GovernanceInstruction)
	if governanceInstruction != "" {
		parts = append(parts, buildInstructionSection("系统治理规则", governanceInstruction))
		summary.HasGovernanceRule = true
		summary.SectionTitles = append(summary.SectionTitles, "系统治理规则")
	}
	if agentInstruction := strings.TrimSpace(input.AgentInstruction); agentInstruction != "" {
		parts = append(parts, buildInstructionSection("Agent 规则", agentInstruction))
		summary.HasAgentRule = true
		summary.SectionTitles = append(summary.SectionTitles, "Agent 规则")
	}
	if skillInstruction := strings.TrimSpace(input.SkillInstruction); skillInstruction != "" {
		parts = append(parts, buildInstructionSection("当前技能上下文", skillInstruction))
		summary.HasSkillRule = true
		summary.SectionTitles = append(summary.SectionTitles, "当前技能上下文")
	}
	if appendix := buildToolAppendix(input.ToolAppendices); appendix != "" {
		parts = append(parts, buildInstructionSection("工具补充规则", appendix))
		summary.HasToolRule = true
		summary.SectionTitles = append(summary.SectionTitles, "工具补充规则")
	}
	return AssemblyResult{
		Text:    strings.TrimSpace(strings.Join(parts, "\n\n")),
		Summary: summary,
	}
}

func buildInstructionSection(title, body string) string {
	title = strings.TrimSpace(title)
	body = strings.TrimSpace(body)
	if body == "" {
		return ""
	}
	if title == "" {
		return body
	}
	return title + "：\n" + body
}

func buildToolAppendix(input []string) string {
	if len(input) == 0 {
		return ""
	}
	parts := make([]string, 0, len(input))
	for _, item := range input {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		parts = append(parts, item)
	}
	return strings.TrimSpace(strings.Join(parts, "\n\n"))
}
