package instruction

import (
	"strings"

	"cs-agent/internal/ai/runtime/tooling"
	"cs-agent/internal/models"
)

type Service struct {
	assembler                     *Assembler
	governanceInstructionProvider *GovernanceInstructionProvider
	skillInstructionProvider      *SkillInstructionProvider
	toolAppendixProvider          *ToolAppendixProvider
}

func NewService() *Service {
	return &Service{
		assembler:                     NewAssembler(),
		governanceInstructionProvider: NewGovernanceInstructionProvider(),
		skillInstructionProvider:      NewSkillInstructionProvider(),
		toolAppendixProvider:          NewToolAppendixProvider(),
	}
}

func (s *Service) Build(
	aiAgent models.AIAgent,
	selectedSkill *models.SkillDefinition,
	toolDefinitions []tooling.MCPToolDefinition,
	extraToolCodes map[string]string,
) AssemblyResult {
	governanceInstruction := ""
	skillInstruction := ""
	toolAppendices := make([]string, 0)
	if s != nil && s.governanceInstructionProvider != nil {
		governanceInstruction = s.governanceInstructionProvider.Resolve()
	}
	if s != nil && s.skillInstructionProvider != nil {
		skillInstruction = s.skillInstructionProvider.Resolve(selectedSkill)
	}
	if s != nil && s.toolAppendixProvider != nil {
		toolAppendices = s.toolAppendixProvider.Build(toolDefinitions, extraToolCodes)
	}
	assembler := NewAssembler()
	if s != nil && s.assembler != nil {
		assembler = s.assembler
	}
	return assembler.Assemble(AssemblerInput{
		AgentInstruction:      strings.TrimSpace(aiAgent.SystemPrompt),
		GovernanceInstruction: governanceInstruction,
		SkillInstruction:      skillInstruction,
		ToolAppendices:        toolAppendices,
	})
}
