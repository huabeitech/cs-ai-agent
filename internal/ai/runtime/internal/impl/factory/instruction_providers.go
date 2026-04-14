package factory

import (
	"strings"

	einoadapter "cs-agent/internal/ai/runtime/internal/impl/adapter"
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/toolx"
)

type ProjectInstructionProvider struct{}

func NewProjectInstructionProvider() *ProjectInstructionProvider {
	return &ProjectInstructionProvider{}
}

func (p *ProjectInstructionProvider) Resolve() string {
	return strings.TrimSpace(DefaultProjectInstruction)
}

type ToolAppendixProvider struct{}

func NewToolAppendixProvider() *ToolAppendixProvider {
	return &ToolAppendixProvider{}
}

func (p *ToolAppendixProvider) Build(selectedSkill *models.SkillDefinition, toolDefinitions []einoadapter.MCPToolDefinition, extraToolCodes map[string]string) []string {
	appendixParts := make([]string, 0, 2)
	if skillInstruction := buildSelectedSkillActivationInstruction(selectedSkill); skillInstruction != "" {
		appendixParts = append(appendixParts, skillInstruction)
	}
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
