package skills

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"text/template"
	"time"

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

func executeByPlan(ctx context.Context, plan *ExecutionPlan, runtimeCtx RuntimeContext) (string, *ExecutionTrace, error) {
	if plan == nil || plan.Skill == nil {
		return "", nil, nil
	}
	trace := &ExecutionTrace{
		Status:        "started",
		ExecutionMode: string(plan.Skill.ExecutionMode),
	}
	switch plan.Skill.ExecutionMode {
	case "", enums.SkillExecutionModePromptOnly:
		replyText, err := executePromptOnly(ctx, plan, runtimeCtx, trace)
		return replyText, trace, err
	case enums.SkillExecutionModeMCPTool:
		replyText, err := executeMCPTool(ctx, plan, runtimeCtx, trace)
		return replyText, trace, err
	default:
		trace.Status = "invalid_execution_mode"
		return "", trace, errorsx.InvalidParam("Skill执行模式不支持")
	}
}

func executePromptOnly(ctx context.Context, plan *ExecutionPlan, runtimeCtx RuntimeContext, trace *ExecutionTrace) (string, error) {
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
	promptTrace := &PromptTrace{Status: "started"}
	if trace != nil {
		trace.Prompt = promptTrace
	}
	startedAt := time.Now()
	result, err := ai.LLM.ChatWithConfig(ctx, plan.AIConfig, systemPrompt, userPrompt)
	promptTrace.LatencyMs = time.Since(startedAt).Milliseconds()
	if err != nil {
		promptTrace.Status = "error"
		promptTrace.Error = err.Error()
		if trace != nil {
			trace.Status = "error"
		}
		return "", err
	}
	promptTrace.Status = "ok"
	promptTrace.ModelName = result.ModelName
	promptTrace.PromptTokens = result.PromptTokens
	promptTrace.CompletionTokens = result.CompletionTokens
	if trace != nil {
		trace.Status = "ok"
	}
	return strings.TrimSpace(result.Content), nil
}

func executeMCPTool(ctx context.Context, plan *ExecutionPlan, runtimeCtx RuntimeContext, trace *ExecutionTrace) (string, error) {
	cfg, err := parseMCPToolExecutionConfig(plan.Skill.ExecutionConfig)
	if err != nil {
		if trace != nil {
			trace.Status = "config_error"
		}
		return "", err
	}
	arguments, err := buildToolArguments(cfg.Arguments, runtimeCtx)
	if err != nil {
		if trace != nil {
			trace.Status = "argument_error"
		}
		return "", err
	}
	mcpTrace := &MCPExecutionTrace{
		Status:     "started",
		ServerCode: cfg.ServerCode,
		ToolName:   cfg.ToolName,
		Arguments:  arguments,
	}
	if trace != nil {
		trace.MCP = mcpTrace
	}
	toolStartedAt := time.Now()
	toolResult, err := mcps.Runtime.CallTool(ctx, cfg.ServerCode, cfg.ToolName, arguments)
	mcpTrace.LatencyMs = time.Since(toolStartedAt).Milliseconds()
	if err != nil {
		mcpTrace.Status = "error"
		mcpTrace.Error = err.Error()
		if trace != nil {
			trace.Status = "error"
		}
		return "", err
	}
	mcpTrace.Status = "ok"
	mcpTrace.IsError = toolResult.IsError
	mcpTrace.ContentItemCount = len(toolResult.Content)
	mcpTrace.HasStructuredContent = toolResult.StructuredContent != nil
	toolSummary := buildToolSummary(toolResult)
	mcpTrace.ResultPreview = truncateTraceText(toolSummary, 500)
	if strings.TrimSpace(toolSummary) == "" {
		if trace != nil {
			trace.Status = "empty_tool_result"
		}
		return "", errorsx.InvalidParam("MCP工具未返回有效结果")
	}
	systemPrompt := strings.TrimSpace(plan.Skill.Prompt)
	if systemPrompt == "" {
		systemPrompt = "你是客服技能助手。请依据工具结果准确回答用户问题，不要编造工具结果中不存在的事实。"
	}
	userPrompt := fmt.Sprintf("用户问题：%s\n\n工具结果：\n%s", strings.TrimSpace(runtimeCtx.UserMessage), toolSummary)
	summaryTrace := &PromptTrace{Status: "started"}
	mcpTrace.SummaryPrompt = summaryTrace
	summaryStartedAt := time.Now()
	result, err := ai.LLM.ChatWithConfig(ctx, plan.AIConfig, systemPrompt, userPrompt)
	summaryTrace.LatencyMs = time.Since(summaryStartedAt).Milliseconds()
	if err != nil {
		summaryTrace.Status = "error"
		summaryTrace.Error = err.Error()
		if trace != nil {
			trace.Status = "error"
		}
		return "", err
	}
	summaryTrace.Status = "ok"
	summaryTrace.ModelName = result.ModelName
	summaryTrace.PromptTokens = result.PromptTokens
	summaryTrace.CompletionTokens = result.CompletionTokens
	if trace != nil {
		trace.Status = "ok"
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

func truncateTraceText(raw string, limit int) string {
	raw = strings.TrimSpace(raw)
	if raw == "" || limit <= 0 {
		return raw
	}
	runes := []rune(raw)
	if len(runes) <= limit {
		return raw
	}
	return strings.TrimSpace(string(runes[:limit])) + "..."
}
