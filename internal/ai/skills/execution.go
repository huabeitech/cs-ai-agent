package skills

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"text/template"

	"cs-agent/internal/ai"
	"cs-agent/internal/ai/mcps"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/pkg/errorsx"
)

type mcpToolExecutionConfig struct {
	ServerCode string            `json:"serverCode"`
	ToolName   string            `json:"toolName"`
	Arguments  map[string]string `json:"arguments"`
}

func executeByPlan(ctx context.Context, plan *ExecutionPlan, runtimeCtx RuntimeContext) (string, error) {
	if plan == nil || plan.Skill == nil {
		return "", nil
	}
	switch plan.Skill.ExecutionMode {
	case "", enums.SkillExecutionModePromptOnly:
		return executePromptOnly(ctx, plan, runtimeCtx)
	case enums.SkillExecutionModeMCPTool:
		return executeMCPTool(ctx, plan, runtimeCtx)
	default:
		return "", errorsx.InvalidParam("Skill执行模式不支持")
	}
}

func executePromptOnly(ctx context.Context, plan *ExecutionPlan, runtimeCtx RuntimeContext) (string, error) {
	if plan == nil || plan.Skill == nil {
		return "", nil
	}
	if plan.AIConfig == nil {
		return "", errorsx.InvalidParam("Skill 关联的 AI 配置不可用")
	}
	systemPrompt := strings.TrimSpace(plan.Skill.Prompt)
	if systemPrompt == "" {
		return "", errorsx.InvalidParam("Skill Prompt 不能为空")
	}
	userPrompt := strings.TrimSpace(runtimeCtx.UserMessage)
	if userPrompt == "" {
		return "", errorsx.InvalidParam("用户消息不能为空")
	}
	result, err := ai.LLM.ChatWithConfig(ctx, plan.AIConfig, systemPrompt, userPrompt)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(result.Content), nil
}

func executeMCPTool(ctx context.Context, plan *ExecutionPlan, runtimeCtx RuntimeContext) (string, error) {
	cfg, err := parseMCPToolExecutionConfig(plan.Skill.ExecutionConfig)
	if err != nil {
		return "", err
	}
	arguments, err := buildToolArguments(cfg.Arguments, runtimeCtx)
	if err != nil {
		return "", err
	}
	toolResult, err := mcps.Runtime.CallTool(ctx, cfg.ServerCode, cfg.ToolName, arguments)
	if err != nil {
		return "", err
	}
	toolSummary := buildToolSummary(toolResult)
	if strings.TrimSpace(toolSummary) == "" {
		return "", errorsx.InvalidParam("MCP工具未返回有效结果")
	}
	systemPrompt := strings.TrimSpace(plan.Skill.Prompt)
	if systemPrompt == "" {
		systemPrompt = "你是客服技能助手。请依据工具结果准确回答用户问题，不要编造工具结果中不存在的事实。"
	}
	userPrompt := fmt.Sprintf("用户问题：%s\n\n工具结果：\n%s", strings.TrimSpace(runtimeCtx.UserMessage), toolSummary)
	result, err := ai.LLM.ChatWithConfig(ctx, plan.AIConfig, systemPrompt, userPrompt)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(result.Content), nil
}

func parseMCPToolExecutionConfig(raw string) (*mcpToolExecutionConfig, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, errorsx.InvalidParam("ExecutionConfig不能为空")
	}
	cfg := &mcpToolExecutionConfig{}
	if err := json.Unmarshal([]byte(raw), cfg); err != nil {
		return nil, errorsx.InvalidParam("ExecutionConfig格式不合法")
	}
	if strings.TrimSpace(cfg.ServerCode) == "" {
		return nil, errorsx.InvalidParam("ExecutionConfig.serverCode不能为空")
	}
	if strings.TrimSpace(cfg.ToolName) == "" {
		return nil, errorsx.InvalidParam("ExecutionConfig.toolName不能为空")
	}
	return cfg, nil
}

func buildToolArguments(templateArgs map[string]string, runtimeCtx RuntimeContext) (map[string]any, error) {
	if len(templateArgs) == 0 {
		return map[string]any{
			"query": strings.TrimSpace(runtimeCtx.UserMessage),
		}, nil
	}
	data := map[string]any{
		"userMessage":     strings.TrimSpace(runtimeCtx.UserMessage),
		"conversationId":  runtimeCtx.ConversationID,
		"aiAgentId":       runtimeCtx.AIAgentID,
		"manualSkillCode": strings.TrimSpace(runtimeCtx.ManualSkillCode),
		"intentCode":      strings.TrimSpace(runtimeCtx.IntentCode),
	}
	ret := make(map[string]any, len(templateArgs))
	for key, value := range templateArgs {
		rendered, err := renderTemplate(value, data)
		if err != nil {
			return nil, errorsx.InvalidParam("ExecutionConfig.arguments模板不合法")
		}
		ret[key] = rendered
	}
	return ret, nil
}

func renderTemplate(raw string, data map[string]any) (string, error) {
	tpl, err := template.New("skill_arg").Option("missingkey=zero").Parse(raw)
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
