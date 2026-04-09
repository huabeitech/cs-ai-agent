package factory

import (
	"context"
	"encoding/json"
	"fmt"
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
	selectedSkill *models.SkillDefinition, toolDefinitions []einoadapter.MCPToolDefinition, extraTools []einobasetool.BaseTool, extraToolCodes map[string]string,
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
		Instruction: buildAgentInstruction(aiAgent, selectedSkill, toolDefinitions, extraToolCodes),
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

func buildAgentInstruction(aiAgent *models.AIAgent, selectedSkill *models.SkillDefinition, toolDefinitions []einoadapter.MCPToolDefinition, extraToolCodes map[string]string) string {
	baseInstruction := ""
	if aiAgent != nil {
		baseInstruction = strings.TrimSpace(aiAgent.SystemPrompt)
	}
	appendixParts := make([]string, 0, 2)
	if skillInstruction := buildSelectedSkillInstruction(selectedSkill, toolDefinitions); skillInstruction != "" {
		appendixParts = append(appendixParts, skillInstruction)
	}
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
