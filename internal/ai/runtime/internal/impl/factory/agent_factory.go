package factory

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	einoadapter "cs-agent/internal/ai/runtime/internal/impl/adapter"
	einoagents "cs-agent/internal/ai/runtime/internal/impl/agents"
	einocallbacks "cs-agent/internal/ai/runtime/internal/impl/callbacks"
	"cs-agent/internal/ai/runtime/registry"
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/toolx"

	"github.com/cloudwego/eino/adk"
	einotoolsearch "github.com/cloudwego/eino/adk/middlewares/dynamictool/toolsearch"
	einoskill "github.com/cloudwego/eino/adk/middlewares/skill"
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
	// StaticToolMetadata 为固定工具的 modelName -> metadata 映射，用于 trace 和运行日志归因。
	StaticToolMetadata map[string]registry.ToolMetadata
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
	handlers := make([]adk.ChatModelAgentMiddleware, 0, 2)
	if len(dynamicTools) > 0 {
		toolSearchHandler, toolSearchErr := einotoolsearch.New(ctx, &einotoolsearch.Config{
			DynamicTools: dynamicTools,
		})
		if toolSearchErr != nil {
			return nil, toolSearchErr
		}
		handlers = append(handlers, toolSearchHandler)
	}
	if input.SelectedSkill != nil {
		skillHandler, skillErr := f.buildSelectedSkillMiddleware(ctx, input.SelectedSkill, input.InstructionToolDefinitions)
		if skillErr != nil {
			return nil, skillErr
		}
		handlers = append(handlers, skillHandler)
	}
	if input.Collector != nil {
		toolMetadataBy := buildRuntimeTraceToolMetadata(input.DynamicMCPToolDefinitions, input.StaticToolMetadata, input.SelectedSkill)
		if input.SelectedSkill != nil {
			input.Collector.SetSkillMiddleware(true, toolx.BuiltinSkill.Name)
		}
		handlers = append(handlers, einocallbacks.NewRuntimeTraceHandler(input.Collector, toolMetadataBy))
	}
	instructionResult := assembleAgentInstruction(input.AIAgent, input.SelectedSkill, input.InstructionToolDefinitions, input.StaticToolCodes)
	if input.Collector != nil {
		input.Collector.SetInstructionSummary(einocallbacks.InstructionTraceSummary{
			SectionTitles:     append([]string(nil), instructionResult.Summary.SectionTitles...),
			HasProjectRule:    instructionResult.Summary.HasProjectRule,
			HasGovernanceRule: instructionResult.Summary.HasGovernanceRule,
			HasAgentRule:      instructionResult.Summary.HasAgentRule,
			HasSkillRule:      instructionResult.Summary.HasSkillRule,
			HasToolRule:       instructionResult.Summary.HasToolRule,
		})
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

func (f *AgentFactory) buildSelectedSkillMiddleware(ctx context.Context, selectedSkill *models.SkillDefinition, toolDefinitions []einoadapter.MCPToolDefinition) (adk.ChatModelAgentMiddleware, error) {
	backend, err := newSelectedSkillBackend(selectedSkill, toolDefinitions)
	if err != nil {
		return nil, err
	}
	toolName := toolx.BuiltinSkill.Name
	return einoskill.NewMiddleware(ctx, &einoskill.Config{
		Backend:       backend,
		SkillToolName: &toolName,
		UseChinese:    true,
	})
}

func assembleAgentInstruction(aiAgent *models.AIAgent, selectedSkill *models.SkillDefinition, toolDefinitions []einoadapter.MCPToolDefinition, extraToolCodes map[string]string) InstructionAssemblyResult {
	baseInstruction := ""
	if aiAgent != nil {
		baseInstruction = strings.TrimSpace(aiAgent.SystemPrompt)
	}
	appendixParts := buildInstructionAppendices(selectedSkill, toolDefinitions, extraToolCodes)
	return NewInstructionAssembler().Assemble(InstructionAssemblerInput{
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

func buildSelectedSkillActivationInstruction(skill *models.SkillDefinition) string {
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
	lines = append(lines, "", "执行要求：", "- 本轮优先处理该技能范围内的问题。", fmt.Sprintf("- 需要专项处理细节时，优先调用 %s 工具加载该技能说明后再继续。", toolx.BuiltinSkill.Name), "- 如果关键信息不足，先向用户追问。", "- 不得调用当前技能未授权的工具。")
	return strings.TrimSpace(strings.Join(lines, "\n"))
}

func buildSelectedSkillDocument(skill *models.SkillDefinition, toolDefinitions []einoadapter.MCPToolDefinition) string {
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
	if content := strings.TrimSpace(skill.Instruction); content != "" {
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
