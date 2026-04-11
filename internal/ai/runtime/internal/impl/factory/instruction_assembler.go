package factory

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type InstructionAssembler struct {
	governanceInstruction string
}

type InstructionAssemblerInput struct {
	AgentInstruction   string
	SkillInstruction   string
	ToolAppendices     []string
	ProjectRoot        string
	ProjectInstruction string
}

// InstructionAssemblySummary 描述 instruction 各组成部分的来源摘要。
type InstructionAssemblySummary struct {
	SectionTitles     []string
	HasProjectRule    bool
	HasGovernanceRule bool
	HasAgentRule      bool
	HasSkillRule      bool
	HasToolRule       bool
}

// InstructionAssemblyResult 为 instruction 装配结果。
type InstructionAssemblyResult struct {
	Text    string
	Summary InstructionAssemblySummary
}

var (
	projectInstructionOnce sync.Once
	projectInstructionText string
)

func NewInstructionAssembler() *InstructionAssembler {
	return &InstructionAssembler{
		governanceInstruction: strings.TrimSpace(`
你正在一个有明确工程约束的客服系统中工作。
执行时必须严格遵守当前注入的项目规则、Agent 规则和技能规则。
如果存在工具白名单限制，只能调用当前允许的工具；信息不足时优先追问，不要伪造事实或跳过必要确认。
`),
	}
}

func (a *InstructionAssembler) Build(input InstructionAssemblerInput) string {
	return a.Assemble(input).Text
}

// Assemble 构建最终 instruction 文本及其来源摘要。
func (a *InstructionAssembler) Assemble(input InstructionAssemblerInput) InstructionAssemblyResult {
	parts := make([]string, 0, 5)
	summary := InstructionAssemblySummary{SectionTitles: make([]string, 0, 5)}
	projectInstruction := strings.TrimSpace(input.ProjectInstruction)
	if projectInstruction == "" {
		projectInstruction = loadProjectInstruction(input.ProjectRoot)
	}
	if projectInstruction != "" {
		parts = append(parts, buildInstructionSection("项目级规则", projectInstruction))
		summary.HasProjectRule = true
		summary.SectionTitles = append(summary.SectionTitles, "项目级规则")
	}
	if a != nil && strings.TrimSpace(a.governanceInstruction) != "" {
		parts = append(parts, buildInstructionSection("系统治理规则", a.governanceInstruction))
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
	return InstructionAssemblyResult{
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

func loadProjectInstruction(projectRoot string) string {
	// TODO 这里还要读取工程目录中的AGENTS.md吗？
	projectInstructionOnce.Do(func() {
		candidates := []string{
			"AGENTS.md",
		}
		projectRoot = strings.TrimSpace(projectRoot)
		if projectRoot != "" {
			candidates = append([]string{filepath.Join(projectRoot, "AGENTS.md")}, candidates...)
		}
		for _, candidate := range candidates {
			candidate = strings.TrimSpace(candidate)
			if candidate == "" {
				continue
			}
			data, err := os.ReadFile(candidate)
			if err != nil {
				continue
			}
			projectInstructionText = strings.TrimSpace(string(data))
			if projectInstructionText != "" {
				return
			}
		}
	})
	return projectInstructionText
}
