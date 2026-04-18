package factory

import (
	"context"
	"strings"

	runtimeinstruction "cs-agent/internal/ai/runtime/instruction"
	einoagents "cs-agent/internal/ai/runtime/internal/impl/agents"
	einocallbacks "cs-agent/internal/ai/runtime/internal/impl/callbacks"
	"cs-agent/internal/ai/runtime/registry"
	runtimetooling "cs-agent/internal/ai/runtime/tooling"
	"cs-agent/internal/models"

	"github.com/cloudwego/eino/adk"
	einobasetool "github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
)

type AgentFactory struct {
	chatModelFactory   *ChatModelFactory
	toolFactory        *ToolFactory
	instructionService *runtimeinstruction.Service
	handlerService     *AgentHandlerService
}

// BuildCustomerServiceAgentInput 定义客服 Agent 的装配输入。
//
// 之所以收敛成单一输入对象，而不是继续堆叠函数参数，是为了避免：
// 1. 调用点无法看懂每个位置参数的语义；
// 2. instruction 用工具、动态工具、中间件工具之间职责混淆；
// 3. 后续扩展装配项时继续拉长函数签名。
type BuildCustomerServiceAgentInput struct {
	// AIAgent 为当前运行的业务 Agent 配置，提供名称、描述、系统提示词等基础信息。
	AIAgent models.AIAgent
	// AIConfig 为模型配置，决定底层使用哪个 ChatModel。
	AIConfig models.AIConfig
	// SelectedSkill 为当前命中的技能；为空表示本次运行未命中专项技能。
	SelectedSkill *models.SkillDefinition
	// InstructionToolDefinitions 用于生成 instruction 中的工具说明。
	// 它描述“当前允许模型理解和使用的 MCP 工具范围”。
	InstructionToolDefinitions []runtimetooling.MCPToolDefinition
	// DynamicMCPToolDefinitions 用于接入 Eino tool_search middleware 的动态工具集合。
	// 这些工具默认不直接挂在 ToolsNode 上，而是经 tool_search 选择后再暴露给模型。
	DynamicMCPToolDefinitions []runtimetooling.MCPToolDefinition
	// StaticTools 为当前运行时直接挂载到 ToolsNode 的固定工具，例如 Graph Tool。
	StaticTools []einobasetool.BaseTool
	// StaticToolCodes 为固定工具的 modelName -> toolCode 映射，用于 trace 和运行日志归因。
	StaticToolCodes map[string]string
	// StaticToolMetadata 为固定工具的 modelName -> metadata 映射，用于 trace 和运行日志归因。
	StaticToolMetadata map[string]registry.ToolMetadata
	// Collector 用于收集运行链路中的 tool trace、graph trace 等调试信息。
	Collector *einocallbacks.RuntimeTraceCollector
}

func NewAgentFactory() *AgentFactory {
	return &AgentFactory{
		chatModelFactory:   NewChatModelFactory(),
		toolFactory:        NewToolFactory(),
		instructionService: runtimeinstruction.NewService(nil, nil, nil, nil),
		handlerService:     NewAgentHandlerService(nil),
	}
}

// BuildCustomerServiceAgent 根据装配输入构建客服 ChatModelAgent。
func (f *AgentFactory) BuildCustomerServiceAgent(ctx context.Context, input BuildCustomerServiceAgentInput) (*einoagents.CustomerServiceAgent, error) {
	chatModel, err := f.chatModelFactory.Build(ctx, input.AIConfig)
	if err != nil {
		return nil, err
	}
	dynamicTools, err := f.toolFactory.BuildBaseToolsByDefinitions(ctx, input.DynamicMCPToolDefinitions)
	if err != nil {
		return nil, err
	}
	allTools := make([]einobasetool.BaseTool, 0, len(input.StaticTools))
	allTools = append(allTools, input.StaticTools...)
	instructionResult := f.instructionService.Build(input.AIAgent, input.SelectedSkill, input.InstructionToolDefinitions, input.StaticToolCodes)
	handlers := make([]adk.ChatModelAgentMiddleware, 0, 3)
	if f.handlerService != nil {
		builtHandlers, err := f.handlerService.Build(ctx, BuildAgentHandlersInput{
			SelectedSkill:              input.SelectedSkill,
			InstructionToolDefinitions: input.InstructionToolDefinitions,
			DynamicToolDefinitions:     input.DynamicMCPToolDefinitions,
			DynamicTools:               dynamicTools,
			StaticToolMetadata:         input.StaticToolMetadata,
			Collector:                  input.Collector,
			InstructionSummary:         buildInstructionTraceSummary(instructionResult.Summary),
		})
		if err != nil {
			return nil, err
		}
		handlers = append(handlers, builtHandlers...)
	}
	inner, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        strings.TrimSpace(input.AIAgent.Name),
		Description: strings.TrimSpace(input.AIAgent.Description),
		Instruction: instructionResult.Text,
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
