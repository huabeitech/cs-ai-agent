package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"text/template"

	"cs-agent/internal/ai"
	"cs-agent/internal/ai/mcps"
	"cs-agent/internal/pkg/errorsx"
)

type toolExecutor struct{}

func newToolExecutor() *toolExecutor {
	return &toolExecutor{}
}

func (e *toolExecutor) Execute(ctx context.Context, turnCtx TurnContext, question string, tool *MCPTool, planReason string) (*TurnResult, error) {
	if tool == nil {
		return nil, errorsx.InvalidParam("Tool 不存在")
	}
	if turnCtx.AIConfig == nil {
		return nil, errorsx.InvalidParam("AI 配置不可用")
	}
	arguments, err := buildDirectToolArguments(tool.Arguments, turnCtx, question)
	if err != nil {
		return nil, err
	}
	toolResult, err := mcps.Runtime.CallTool(ctx, tool.ServerCode, tool.ToolName, arguments)
	if err != nil {
		return &TurnResult{
			Action:          ActionFallback,
			Question:        question,
			Reason:          "Tool执行失败",
			PlannedAction:   ActionTool,
			PlannedToolCode: tool.Code(),
			PlanReason:      planReason,
		}, err
	}
	toolSummary := buildToolSummary(toolResult)
	if strings.TrimSpace(toolSummary) == "" {
		return &TurnResult{
			Action:          ActionFallback,
			Question:        question,
			Reason:          "Tool未返回有效结果",
			PlannedAction:   ActionTool,
			PlannedToolCode: tool.Code(),
			PlanReason:      planReason,
		}, nil
	}

	systemPrompt := strings.TrimSpace(turnCtx.AIAgent.SystemPrompt)
	if systemPrompt == "" {
		systemPrompt = "你是客服助手。请依据工具结果准确回答用户问题，不要编造工具结果中不存在的事实。"
	}
	userPrompt := fmt.Sprintf("用户问题：%s\n\n工具：%s\n工具结果：\n%s", strings.TrimSpace(question), tool.Code(), toolSummary)
	result, err := ai.LLM.ChatWithConfig(ctx, turnCtx.AIConfig, systemPrompt, userPrompt)
	if err != nil {
		return &TurnResult{
			Action:          ActionFallback,
			Question:        question,
			Reason:          "Tool结果整理失败",
			PlannedAction:   ActionTool,
			PlannedToolCode: tool.Code(),
			PlanReason:      planReason,
		}, err
	}
	return &TurnResult{
		Action:          ActionReply,
		Question:        question,
		ReplyText:       strings.TrimSpace(result.Content),
		Reason:          "direct_tool",
		PlannedAction:   ActionTool,
		PlannedToolCode: tool.Code(),
		PlanReason:      planReason,
	}, nil
}

func buildDirectToolArguments(templateArgs map[string]string, turnCtx TurnContext, question string) (map[string]any, error) {
	if len(templateArgs) == 0 {
		return map[string]any{}, nil
	}
	data := map[string]any{
		"userMessage":     strings.TrimSpace(question),
		"conversationId":  conversationID(turnCtx.Conversation),
		"aiAgentId":       turnCtx.AIAgent.ID,
		"manualSkillCode": strings.TrimSpace(turnCtx.ManualSkillCode),
		"intentCode":      strings.TrimSpace(turnCtx.IntentCode),
	}
	ret := make(map[string]any, len(templateArgs))
	for key, value := range templateArgs {
		rendered, err := renderDirectToolTemplate(value, data)
		if err != nil {
			return nil, errorsx.InvalidParam("Direct Tool arguments 模板不合法")
		}
		ret[key] = rendered
	}
	return ret, nil
}

func renderDirectToolTemplate(raw string, data map[string]any) (string, error) {
	tpl, err := template.New("direct_tool_arg").Option("missingkey=zero").Parse(raw)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return strings.TrimSpace(buf.String()), nil
}

func buildToolSummary(result *mcps.ToolCallResult) string {
	if result == nil {
		return ""
	}
	lines := make([]string, 0, len(result.Content)+2)
	if result.StructuredContent != nil {
		if data, err := json.Marshal(result.StructuredContent); err == nil {
			lines = append(lines, string(data))
		}
	}
	for _, item := range result.Content {
		if strings.TrimSpace(item.Text) != "" {
			lines = append(lines, strings.TrimSpace(item.Text))
			continue
		}
		if item.Data != nil {
			if data, err := json.Marshal(item.Data); err == nil {
				lines = append(lines, string(data))
			}
		}
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}
