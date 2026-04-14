package factory

import (
	"context"

	einoadapter "cs-agent/internal/ai/runtime/internal/impl/adapter"
	einocallbacks "cs-agent/internal/ai/runtime/internal/impl/callbacks"
	"cs-agent/internal/ai/runtime/registry"
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/toolx"

	"github.com/cloudwego/eino/adk"
	einotoolsearch "github.com/cloudwego/eino/adk/middlewares/dynamictool/toolsearch"
	einobasetool "github.com/cloudwego/eino/components/tool"
)

type AgentHandlerService struct {
	skillMiddleware *SkillMiddlewareService
}

type BuildAgentHandlersInput struct {
	SelectedSkill              *models.SkillDefinition
	InstructionToolDefinitions []einoadapter.MCPToolDefinition
	DynamicToolDefinitions     []einoadapter.MCPToolDefinition
	DynamicTools               []einobasetool.BaseTool
	StaticToolMetadata         map[string]registry.ToolMetadata
	Collector                  *einocallbacks.RuntimeTraceCollector
	InstructionSummary         einocallbacks.InstructionTraceSummary
}

func NewAgentHandlerService(skillMiddleware *SkillMiddlewareService) *AgentHandlerService {
	if skillMiddleware == nil {
		skillMiddleware = NewSkillMiddlewareService()
	}
	return &AgentHandlerService{skillMiddleware: skillMiddleware}
}

func (s *AgentHandlerService) Build(ctx context.Context, input BuildAgentHandlersInput) ([]adk.ChatModelAgentMiddleware, error) {
	handlers := make([]adk.ChatModelAgentMiddleware, 0, 3)
	if len(input.DynamicTools) > 0 {
		toolSearchHandler, err := einotoolsearch.New(ctx, &einotoolsearch.Config{
			DynamicTools: input.DynamicTools,
		})
		if err != nil {
			return nil, err
		}
		handlers = append(handlers, toolSearchHandler)
	}
	if input.SelectedSkill != nil {
		skillHandler, err := s.skillMiddleware.Build(ctx, input.SelectedSkill, input.InstructionToolDefinitions)
		if err != nil {
			return nil, err
		}
		handlers = append(handlers, skillHandler)
	}
	if input.Collector != nil {
		toolMetadataBy := buildRuntimeTraceToolMetadata(input.DynamicToolDefinitions, input.StaticToolMetadata, input.SelectedSkill)
		if input.SelectedSkill != nil {
			input.Collector.SetSkillMiddleware(true, toolx.BuiltinSkill.Name)
		}
		input.Collector.SetInstructionSummary(input.InstructionSummary)
		handlers = append(handlers, einocallbacks.NewRuntimeTraceHandler(input.Collector, toolMetadataBy))
	}
	return handlers, nil
}
