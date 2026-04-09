package factory

import (
	"context"
	"strings"

	einoadapter "cs-agent/internal/ai/runtime/internal/impl/adapter"
	einoagents "cs-agent/internal/ai/runtime/internal/impl/agents"
	einocallbacks "cs-agent/internal/ai/runtime/internal/impl/callbacks"
	"cs-agent/internal/models"

	"github.com/cloudwego/eino/adk"
	einobasetool "github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
)

type AgentFactory struct {
	chatModelFactory *ChatModelFactory
	toolFactory      *ToolFactory
}

func NewAgentFactory() *AgentFactory {
	return &AgentFactory{
		chatModelFactory: NewChatModelFactory(),
		toolFactory:      NewToolFactory(),
	}
}

func (f *AgentFactory) BuildCustomerServiceAgent(ctx context.Context, aiAgent *models.AIAgent, aiConfig *models.AIConfig,
	toolDefinitions []einoadapter.MCPToolDefinition, extraTools []einobasetool.BaseTool, extraToolCodes map[string]string,
	collector *einocallbacks.RuntimeTraceCollector) (*einoagents.CustomerServiceAgent, error) {
	if aiAgent == nil || aiConfig == nil {
		return nil, nil
	}
	chatModel, err := f.chatModelFactory.Build(ctx, aiConfig)
	if err != nil {
		return nil, err
	}
	baseTools, err := f.toolFactory.BuildBaseToolsByDefinitions(ctx, toolDefinitions)
	if err != nil {
		return nil, err
	}
	allTools := make([]einobasetool.BaseTool, 0, len(baseTools)+len(extraTools))
	allTools = append(allTools, extraTools...)
	allTools = append(allTools, baseTools...)
	handlers := make([]adk.ChatModelAgentMiddleware, 0, 1)
	if collector != nil {
		toolMetadataBy := make(map[string]einocallbacks.ToolMetadata, len(toolDefinitions))
		for _, item := range toolDefinitions {
			toolMetadataBy[item.ModelName] = einocallbacks.ToolMetadata{
				ToolCode:   item.ToolCode,
				ServerCode: item.ServerCode,
				ToolName:   item.ToolName,
			}
		}
		handlers = append(handlers, einocallbacks.NewRuntimeTraceHandler(collector, toolMetadataBy))
	}
	inner, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        strings.TrimSpace(aiAgent.Name),
		Description: strings.TrimSpace(aiAgent.Description),
		Instruction: buildAgentInstruction(aiAgent, extraToolCodes),
		Model:       chatModel,
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: allTools,
			},
		},
		Handlers: handlers,
	})
	if err != nil {
		return nil, err
	}
	return &einoagents.CustomerServiceAgent{Inner: inner}, nil
}

func buildAgentInstruction(aiAgent *models.AIAgent, extraToolCodes map[string]string) string {
	baseInstruction := ""
	if aiAgent != nil {
		baseInstruction = strings.TrimSpace(aiAgent.SystemPrompt)
	}
	appendixParts := make([]string, 0, 1)
	if hasToolCode(extraToolCodes, "builtin/create_ticket_with_confirmation") {
		appendixParts = append(appendixParts, strings.TrimSpace(`
你可以在确认信息充分后调用 create_ticket_with_confirmation 工具来创建工单，但必须遵守以下规则：
1. 只有在用户明确表达希望提交工单、投诉、报障、售后处理等诉求时，才考虑调用该工具。
2. 调用前你必须已经整理出清晰的工单标题和问题描述；如果信息不足，先继续追问，不要过早调用。
3. 一旦准备创建工单，必须调用 create_ticket_with_confirmation 工具，禁止直接口头宣称“已经创建工单”。
4. 该工具会先向用户发起确认。用户确认后才会真正创建工单；用户取消则结束本次建单流程。
5. 如果用户只是咨询、抱怨或泛泛表达不满，但没有明确要求建单，优先继续澄清，不要主动创建工单。
`))
	}
	if len(appendixParts) == 0 {
		return baseInstruction
	}
	if baseInstruction == "" {
		return strings.Join(appendixParts, "\n\n")
	}
	return baseInstruction + "\n\n" + strings.Join(appendixParts, "\n\n")
}

func hasToolCode(toolCodes map[string]string, target string) bool {
	target = strings.TrimSpace(target)
	if target == "" {
		return false
	}
	for _, toolCode := range toolCodes {
		if strings.TrimSpace(toolCode) == target {
			return true
		}
	}
	return false
}
