package instruction

import (
	"strings"

	runtimetooling "cs-agent/internal/ai/runtime/tooling"
	"cs-agent/internal/models"
)

type Service struct {
	assembler                     *Assembler
	projectInstructionProvider    *ProjectInstructionProvider
	governanceInstructionProvider *GovernanceInstructionProvider
	skillInstructionProvider      *SkillInstructionProvider
	toolAppendixProvider          *ToolAppendixProvider
}

func NewService(
	assembler *Assembler,
	projectProvider *ProjectInstructionProvider,
	governanceProvider *GovernanceInstructionProvider,
	skillProvider *SkillInstructionProvider,
	toolProvider *ToolAppendixProvider,
) *Service {
	if assembler == nil {
		assembler = NewAssembler()
	}
	if projectProvider == nil {
		projectProvider = NewProjectInstructionProvider()
	}
	if governanceProvider == nil {
		governanceProvider = NewGovernanceInstructionProvider()
	}
	if skillProvider == nil {
		skillProvider = NewSkillInstructionProvider()
	}
	if toolProvider == nil {
		toolProvider = NewToolAppendixProvider()
	}
	return &Service{
		assembler:                     assembler,
		projectInstructionProvider:    projectProvider,
		governanceInstructionProvider: governanceProvider,
		skillInstructionProvider:      skillProvider,
		toolAppendixProvider:          toolProvider,
	}
}

func (s *Service) Build(
	aiAgent models.AIAgent,
	selectedSkill *models.SkillDefinition,
	toolDefinitions []runtimetooling.MCPToolDefinition,
	extraToolCodes map[string]string,
) AssemblyResult {
	projectInstruction := ""
	governanceInstruction := ""
	skillInstruction := ""
	toolAppendices := make([]string, 0)
	if s != nil && s.projectInstructionProvider != nil {
		projectInstruction = s.projectInstructionProvider.Resolve()
	}
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
		ProjectInstruction:    projectInstruction,
	})
}
