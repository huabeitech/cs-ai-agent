package instruction

import (
	"strings"

	runtimetooling "cs-agent/internal/ai/runtime/tooling"
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/toolx"
)

type ToolAppendixProvider struct{}

func NewToolAppendixProvider() *ToolAppendixProvider {
	return &ToolAppendixProvider{}
}

type GovernanceInstructionProvider struct{}

func NewGovernanceInstructionProvider() *GovernanceInstructionProvider {
	return &GovernanceInstructionProvider{}
}

func (p *GovernanceInstructionProvider) Resolve() string {
	return strings.TrimSpace(defaultGovernanceInstruction)
}

type SkillInstructionProvider struct{}

func NewSkillInstructionProvider() *SkillInstructionProvider {
	return &SkillInstructionProvider{}
}

func (p *SkillInstructionProvider) Resolve(selectedSkill *models.SkillDefinition) string {
	return BuildSelectedSkillActivationInstruction(selectedSkill)
}

func (p *ToolAppendixProvider) Build(toolDefinitions []runtimetooling.MCPToolDefinition, extraToolCodes map[string]string) []string {
	appendixParts := make([]string, 0, 1)
	toolCodes := make([]string, 0, len(toolDefinitions)+len(extraToolCodes))
	for _, item := range toolDefinitions {
		toolCodes = append(toolCodes, item.ToolCode)
	}
	for _, item := range extraToolCodes {
		toolCodes = append(toolCodes, item)
	}
	appendixParts = append(appendixParts, toolx.BuildToolAppendicesForCodes(len(toolDefinitions) > 0, toolCodes)...)
	return appendixParts
}
