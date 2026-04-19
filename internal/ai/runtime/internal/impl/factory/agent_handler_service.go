package factory

import (
	"context"
	"strings"

	einocallbacks "cs-agent/internal/ai/runtime/internal/impl/callbacks"
	"cs-agent/internal/ai/runtime/registry"
	runtimetooling "cs-agent/internal/ai/runtime/tooling"
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
	AIAgent                    models.AIAgent
	InstructionToolDefinitions []runtimetooling.MCPToolDefinition
	DynamicToolDefinitions     []runtimetooling.MCPToolDefinition
	DynamicTools               []einobasetool.BaseTool
	StaticToolMetadata         map[string]registry.ToolMetadata
	Collector                  *einocallbacks.RuntimeTraceCollector
	InstructionSummary         einocallbacks.InstructionTraceSummary
}

func NewAgentHandlerService(skillMiddleware *SkillMiddlewareService) *AgentHandlerService {
	if skillMiddleware == nil {
		panic("skill middleware is required")
	}
	return &AgentHandlerService{skillMiddleware: skillMiddleware}
}

func (s *AgentHandlerService) Build(ctx context.Context, input BuildAgentHandlersInput) ([]adk.ChatModelAgentMiddleware, error) {
	handlers := make([]adk.ChatModelAgentMiddleware, 0, 4)
	skillMetadataByCode := buildRuntimeSkillMetadataMap(input.AIAgent)
	toolMetadataBy := buildRuntimeTraceToolMetadata(input.DynamicToolDefinitions, input.StaticToolMetadata, len(skillMetadataByCode) > 0)
	traceSkillMetadata := make(map[string]einocallbacks.SkillMetadata, len(skillMetadataByCode))
	for code, item := range skillMetadataByCode {
		traceSkillMetadata[code] = einocallbacks.SkillMetadata{
			Code:             item.Code,
			Name:             item.Name,
			Description:      item.Description,
			AllowedToolCodes: append([]string(nil), item.AllowedToolCodes...),
		}
	}
	if input.Collector != nil {
		if len(skillMetadataByCode) > 0 {
			input.Collector.SetSkillMiddleware(true, toolx.BuiltinSkill.Name)
		}
		input.Collector.SetVisibleSkills(traceSkillMetadata)
		input.Collector.SetInstructionSummary(input.InstructionSummary)
		handlers = append(handlers, einocallbacks.NewRuntimeTraceHandler(input.Collector, toolMetadataBy, traceSkillMetadata))
	}
	if len(input.DynamicTools) > 0 {
		toolSearchHandler, err := einotoolsearch.New(ctx, &einotoolsearch.Config{
			DynamicTools: input.DynamicTools,
		})
		if err != nil {
			return nil, err
		}
		handlers = append(handlers, toolSearchHandler)
	}
	if len(skillMetadataByCode) > 0 {
		skillHandler, err := s.skillMiddleware.Build(ctx, input.AIAgent, input.InstructionToolDefinitions)
		if err != nil {
			return nil, err
		}
		handlers = append(handlers, skillHandler)
		handlers = append(handlers, NewRuntimeToolFilterMiddleware(
			input.Collector,
			toolMetadataBy,
			traceSkillMetadata,
			dynamicToolModelNames(input.DynamicToolDefinitions),
		))
	}
	return handlers, nil
}

func dynamicToolModelNames(definitions []runtimetooling.MCPToolDefinition) []string {
	ret := make([]string, 0, len(definitions))
	for _, item := range definitions {
		modelName := strings.TrimSpace(item.ModelName)
		if modelName == "" {
			modelName = strings.TrimSpace(runtimetooling.BuildModelToolName(item))
		}
		if modelName == "" {
			continue
		}
		ret = append(ret, modelName)
	}
	return ret
}
