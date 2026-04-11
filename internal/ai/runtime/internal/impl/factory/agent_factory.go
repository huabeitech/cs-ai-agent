package factory

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	einoadapter "cs-agent/internal/ai/runtime/internal/impl/adapter"
	einoagents "cs-agent/internal/ai/runtime/internal/impl/agents"
	einocallbacks "cs-agent/internal/ai/runtime/internal/impl/callbacks"
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/toolx"

	"github.com/cloudwego/eino/adk"
	einotoolsearch "github.com/cloudwego/eino/adk/middlewares/dynamictool/toolsearch"
	einobasetool "github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
)

type AgentFactory struct {
	chatModelFactory     *ChatModelFactory
	toolFactory          *ToolFactory
	instructionAssembler *InstructionAssembler
}

// BuildCustomerServiceAgentInput 定义客服 Agent 的装配输入。
//
// 之所以收敛成单一输入对象，而不是继续堆叠函数参数，是为了避免：
// 1. 调用点无法看懂每个位置参数的语义；
// 2. instruction 用工具、动态工具、中间件工具之间职责混淆；
// 3. 后续扩展装配项时继续拉长函数签名。
type BuildCustomerServiceAgentInput struct {
	// AIAgent 为当前运行的业务 Agent 配置，提供名称、描述、系统提示词等基础信息。
	AIAgent *models.AIAgent
	// AIConfig 为模型配置，决定底层使用哪个 ChatModel。
	AIConfig *models.AIConfig
	// SelectedSkill 为当前命中的技能；为空表示本次运行未命中专项技能。
	SelectedSkill *models.SkillDefinition
	// InstructionToolDefinitions 用于生成 instruction 中的工具说明。
	// 它描述“当前允许模型理解和使用的 MCP 工具范围”。
	InstructionToolDefinitions []einoadapter.MCPToolDefinition
	// DynamicMCPToolDefinitions 用于接入 Eino tool_search middleware 的动态工具集合。
	// 这些工具默认不直接挂在 ToolsNode 上，而是经 tool_search 选择后再暴露给模型。
	DynamicMCPToolDefinitions []einoadapter.MCPToolDefinition
	// StaticTools 为当前运行时直接挂载到 ToolsNode 的固定工具，例如 Graph Tool。
	StaticTools []einobasetool.BaseTool
	// StaticToolCodes 为固定工具的 modelName -> toolCode 映射，用于 trace 和运行日志归因。
	StaticToolCodes map[string]string
	// Collector 用于收集运行链路中的 tool trace、graph trace 等调试信息。
	Collector *einocallbacks.RuntimeTraceCollector
}

func NewAgentFactory() *AgentFactory {
	return &AgentFactory{
		chatModelFactory:     NewChatModelFactory(),
		toolFactory:          NewToolFactory(),
		instructionAssembler: NewInstructionAssembler(),
	}
}

// BuildCustomerServiceAgent 根据装配输入构建客服 ChatModelAgent。
func (f *AgentFactory) BuildCustomerServiceAgent(ctx context.Context, input BuildCustomerServiceAgentInput) (*einoagents.CustomerServiceAgent, error) {
	if input.AIAgent == nil || input.AIConfig == nil {
		return nil, nil
	}
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
	handlers := make([]adk.ChatModelAgentMiddleware, 0, 1)
	if len(dynamicTools) > 0 {
		toolSearchHandler, toolSearchErr := einotoolsearch.New(ctx, &einotoolsearch.Config{
			DynamicTools: dynamicTools,
		})
		if toolSearchErr != nil {
			return nil, toolSearchErr
		}
		handlers = append(handlers, toolSearchHandler)
	}
	if input.Collector != nil {
		toolMetadataBy := make(map[string]einocallbacks.ToolMetadata, len(input.DynamicMCPToolDefinitions)+len(input.StaticToolCodes))
		for _, item := range input.DynamicMCPToolDefinitions {
			toolMetadataBy[item.ModelName] = einocallbacks.ToolMetadata{
				ToolCode:   item.ToolCode,
				ServerCode: item.ServerCode,
				ToolName:   item.ToolName,
				SourceType: "mcp",
			}
		}
		for modelName, toolCode := range input.StaticToolCodes {
			modelName = strings.TrimSpace(modelName)
			toolCode = strings.TrimSpace(toolCode)
			if modelName == "" || toolCode == "" {
				continue
			}
			serverCode, toolName := "", ""
			if toolCode == toolx.BuiltinToolSearchToolCode {
				serverCode = toolx.BuiltinToolCatalogServerCode
				toolName = toolx.BuiltinToolSearchToolName
			} else if toolCode == toolx.GraphCreateTicketConfirmToolCode {
				serverCode = toolx.GraphToolCatalogServerCode
				toolName = toolx.GraphCreateTicketConfirmToolName
			} else if toolCode == toolx.GraphHandoffConversationToolCode {
				serverCode = toolx.GraphToolCatalogServerCode
				toolName = toolx.GraphHandoffConversationToolName
			}
			toolMetadataBy[modelName] = einocallbacks.ToolMetadata{
				ToolCode:   toolCode,
				ServerCode: serverCode,
				ToolName:   toolName,
				SourceType: resolveToolSourceType(toolCode),
			}
		}
		handlers = append(handlers, einocallbacks.NewRuntimeTraceHandler(input.Collector, toolMetadataBy))
	}
	inner, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        strings.TrimSpace(input.AIAgent.Name),
		Description: strings.TrimSpace(input.AIAgent.Description),
		Instruction: buildAgentInstruction(input.AIAgent, input.SelectedSkill, input.InstructionToolDefinitions, input.StaticToolCodes),
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

func resolveToolSourceType(toolCode string) string {
	toolCode = strings.TrimSpace(toolCode)
	switch {
	case toolCode == toolx.BuiltinToolSearchToolCode:
		return toolx.BuiltinToolCatalogServerCode
	case strings.HasPrefix(toolCode, toolx.GraphToolCatalogServerCode+"/"):
		return toolx.GraphToolCatalogServerCode
	case strings.HasPrefix(toolCode, toolx.BuiltinToolCatalogServerCode+"/"):
		return toolx.BuiltinToolCatalogServerCode
	default:
		return "mcp"
	}
}

func buildAgentInstruction(aiAgent *models.AIAgent, selectedSkill *models.SkillDefinition, toolDefinitions []einoadapter.MCPToolDefinition, extraToolCodes map[string]string) string {
	baseInstruction := ""
	if aiAgent != nil {
		baseInstruction = strings.TrimSpace(aiAgent.SystemPrompt)
	}
	appendixParts := make([]string, 0, 2)
	if skillInstruction := buildSelectedSkillInstruction(selectedSkill, toolDefinitions); skillInstruction != "" {
		appendixParts = append(appendixParts, skillInstruction)
	}
	if len(toolDefinitions) > 0 {
		appendixParts = append(appendixParts, strings.TrimSpace(`
当你需要使用长尾 MCP 能力时，优先使用 tool_search 工具，并遵守以下规则：
1. 先调用 tool_search 搜索需要的动态工具，再继续使用已选中的真实工具。
2. 不要假设所有长尾工具一开始就可见；只有被 tool_search 选中的工具，后续模型调用才会暴露出来。
3. 如果当前已有固定内置工具可以完成任务，优先使用固定工具，不要滥用 tool_search。
`))
	}
	if hasToolCode(extraToolCodes, toolx.GraphCreateTicketConfirmToolCode) {
		appendixParts = append(appendixParts, strings.TrimSpace(`
你可以在确认信息充分后调用 create_ticket_with_confirmation 这个 Graph Tool 来创建工单，但必须遵守以下规则：
1. 只有在用户明确表达希望提交工单、投诉、报障、售后处理等诉求时，才考虑调用该工具。
2. 调用前你必须已经整理出清晰的工单标题和问题描述；如果信息不足，先继续追问，不要过早调用。
3. 一旦准备创建工单，必须调用 create_ticket_with_confirmation 工具，禁止直接口头宣称“已经创建工单”。
4. 该 Graph Tool 会先向用户发起确认。用户确认后才会真正创建工单；用户取消则结束本次建单流程。
5. 如果用户只是咨询、抱怨或泛泛表达不满，但没有明确要求建单，优先继续澄清，不要主动创建工单。
`))
	}
	if hasToolCode(extraToolCodes, toolx.GraphHandoffConversationToolCode) {
		appendixParts = append(appendixParts, strings.TrimSpace(`
你可以在确认需要人工介入后调用 handoff_to_human 这个 Graph Tool 来转人工，但必须遵守以下规则：
1. 只有在用户明确要求人工客服，或你已经判断该问题必须由人工继续处理时，才调用该工具。
2. 调用前先尽量整理清楚转人工原因；如果理由含糊，先追问或澄清，不要直接转人工。
3. 一旦决定转人工，必须调用 handoff_to_human 工具，禁止只在回复里口头说“我帮你转人工了”。
4. 该 Graph Tool 会先向用户发起确认。用户确认后才会真正转人工；用户取消则结束本次转人工流程。
5. 如果问题仍可由当前对话继续解决，优先继续解答，不要过早转人工。
`))
	}
	projectRoot, _ := os.Getwd()
	return NewInstructionAssembler().Build(InstructionAssemblerInput{
		ProjectRoot:      projectRoot,
		AgentInstruction: baseInstruction,
		SkillInstruction: firstAppendixPart(appendixParts),
		ToolAppendices:   remainingAppendixParts(appendixParts),
	})
}

func firstAppendixPart(parts []string) string {
	if len(parts) == 0 {
		return ""
	}
	return strings.TrimSpace(parts[0])
}

func remainingAppendixParts(parts []string) []string {
	if len(parts) <= 1 {
		return nil
	}
	ret := make([]string, 0, len(parts)-1)
	for _, item := range parts[1:] {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		ret = append(ret, item)
	}
	return ret
}

func buildSelectedSkillInstruction(skill *models.SkillDefinition, toolDefinitions []einoadapter.MCPToolDefinition) string {
	if skill == nil {
		return ""
	}
	lines := []string{
		"当前命中的专项技能：",
		fmt.Sprintf("- code: %s", strings.TrimSpace(skill.Code)),
		fmt.Sprintf("- name: %s", strings.TrimSpace(skill.Name)),
	}
	if desc := strings.TrimSpace(skill.Description); desc != "" {
		lines = append(lines, fmt.Sprintf("- description: %s", desc))
	}
	if content := strings.TrimSpace(skill.Content); content != "" {
		lines = append(lines, "", "技能说明：", content)
	}
	if examples := parseJSONStringArray(skill.Examples); len(examples) > 0 {
		lines = append(lines, "", "典型示例问法：")
		for _, item := range examples {
			lines = append(lines, "- "+item)
		}
	}
	if len(toolDefinitions) > 0 {
		lines = append(lines, "", "当前技能允许使用的工具：")
		for _, item := range toolDefinitions {
			if strings.TrimSpace(item.ToolCode) == "" {
				continue
			}
			line := "- " + strings.TrimSpace(item.ToolCode)
			if title := strings.TrimSpace(item.Title); title != "" {
				line += " | " + title
			}
			lines = append(lines, line)
		}
	}
	lines = append(lines, "", "执行要求：", "- 优先遵循该技能说明完成任务。", "- 如果关键信息不足，先向用户追问。", "- 不得调用当前技能未授权的工具。")
	return strings.TrimSpace(strings.Join(lines, "\n"))
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

func parseJSONStringArray(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	var ret []string
	if err := json.Unmarshal([]byte(raw), &ret); err != nil {
		return nil
	}
	out := make([]string, 0, len(ret))
	for _, item := range ret {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		out = append(out, item)
	}
	return out
}
