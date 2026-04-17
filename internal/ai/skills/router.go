package skills

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"cs-agent/internal/ai"
	"cs-agent/internal/models"
)

func routeSkillWithLLM(ctx context.Context, aiConfig models.AIConfig, userMessage string, candidates []models.SkillDefinition) (*models.SkillDefinition, *RouteTrace, error) {
	trace := &RouteTrace{Status: "started"}
	if len(candidates) == 0 {
		trace.Status = "no_candidate"
		return nil, trace, nil
	}
	userMessage = strings.TrimSpace(userMessage)
	if userMessage == "" {
		trace.Status = "empty_user_message"
		return nil, trace, nil
	}
	// TODO 系统体是指为什么不是配置在AIAgent中的？
	systemPrompt := "你是客服技能路由器。你只能在候选 Skill 中选择一个最合适的 skillCode，或者返回 NONE。只有当用户问题与 Skill 的职责边界明确匹配时才选择；如果不明确、信息不足、多个 Skill 都不够确定，就返回 NONE。输出只能是 skillCode 或 NONE，不能输出其他内容。"
	userPrompt := buildSkillRoutePrompt(userMessage, candidates)
	startedAt := time.Now()
	result, err := ai.LLM.ChatWithConfig(ctx, aiConfig, systemPrompt, userPrompt)
	trace.LatencyMs = time.Since(startedAt).Milliseconds()
	if err != nil {
		trace.Status = "route_error"
		trace.Error = err.Error()
		return nil, trace, err
	}
	decision := normalizeRouteDecision(result.Content)
	trace.RawDecision = strings.TrimSpace(result.Content)
	if decision == "" || decision == "NONE" {
		trace.Status = "not_matched"
		return nil, trace, nil
	}
	for _, item := range candidates {
		if strings.EqualFold(item.Code, decision) {
			trace.Status = "llm_selected"
			trace.SelectedSkillCode = item.Code
			return &item, trace, nil
		}
	}
	trace.Status = "invalid_decision"
	trace.Error = fmt.Sprintf("invalid route decision: %s", decision)
	return nil, trace, nil
}

func buildSkillRoutePrompt(userMessage string, candidates []models.SkillDefinition) string {
	lines := make([]string, 0, len(candidates)+4)
	lines = append(lines, "用户问题：")
	lines = append(lines, strings.TrimSpace(userMessage))
	lines = append(lines, "")
	lines = append(lines, "候选 Skills：")
	for _, item := range candidates {
		line := fmt.Sprintf("- skillCode=%s; name=%s; description=%s", strings.TrimSpace(item.Code), strings.TrimSpace(item.Name), strings.TrimSpace(item.Description))
		if examples := parseSkillExamples(item.Examples); len(examples) > 0 {
			line += "; examples=" + strings.Join(examples, " | ")
		}
		lines = append(lines, line)
	}
	lines = append(lines, "")
	lines = append(lines, "请只输出一个 skillCode 或 NONE。")
	return strings.Join(lines, "\n")
}

func parseSkillExamples(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	var items []string
	if err := json.Unmarshal([]byte(raw), &items); err != nil {
		return nil
	}
	ret := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		ret = append(ret, item)
		if len(ret) >= 3 {
			break
		}
	}
	return ret
}

func normalizeRouteDecision(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if idx := strings.Index(raw, "\n"); idx >= 0 {
		raw = raw[:idx]
	}
	raw = strings.TrimSpace(raw)
	raw = strings.Trim(raw, "`")
	raw = strings.TrimSpace(raw)
	raw = strings.Trim(raw, "\"'")
	if strings.EqualFold(raw, "NONE") {
		return "NONE"
	}
	return raw
}
