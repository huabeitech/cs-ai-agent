package skills

import (
	"context"
	"fmt"
	"strings"
	"time"

	"cs-agent/internal/ai"
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/pkg/errorsx"
	"cs-agent/internal/pkg/utils"
	"cs-agent/internal/repositories"

	"github.com/mlogclub/simple/common/strs"
	"github.com/mlogclub/simple/sqls"
)

type intentTriggerConfig struct {
	Intents []string `json:"intents"`
}

// MatchSkill 对单个 SkillDefinition 执行命中判断。
func MatchSkill(execCtx context.Context, ctx RuntimeContext, aiAgent *models.AIAgent, aiConfig *models.AIConfig) (*models.SkillDefinition, string, *RouteTrace, error) {
	if strs.IsNotBlank(ctx.ManualSkillCode) {
		skill := repositories.SkillDefinitionRepository.GetByCode(sqls.DB(), ctx.ManualSkillCode)
		if skill == nil || skill.Status != enums.StatusOk {
			return nil, "", nil, errorsx.InvalidParam("Skill 不存在或未启用")
		}
		return skill, "manual_skill_code", &RouteTrace{
			Status:            "manual_selected",
			SelectedSkillCode: skill.Code,
		}, nil
	}

	candidates := loadCandidateSkills(aiAgent)
	trace := &RouteTrace{
		Status:              "started",
		CandidateSkillCodes: make([]string, 0, len(candidates)),
	}
	for _, item := range candidates {
		trace.CandidateSkillCodes = append(trace.CandidateSkillCodes, item.Code)
	}
	if len(candidates) == 0 {
		trace.Status = "no_candidate"
		return nil, "no_enabled_skill_bound", trace, nil
	}

	intentCode := strings.TrimSpace(ctx.IntentCode)
	if intentCode != "" {
		for _, item := range candidates {
			if strings.EqualFold(strings.TrimSpace(item.Code), intentCode) {
				trace.Status = "intent_selected"
				trace.SelectedSkillCode = item.Code
				return &item, "intent_code", trace, nil
			}
		}
	}

	if len(candidates) == 1 {
		trace.Status = "single_candidate"
		trace.SelectedSkillCode = candidates[0].Code
		return &candidates[0], "single_candidate", trace, nil
	}

	selected, routeTrace, err := routeSkillWithLLM(execCtx, aiConfig, ctx.UserMessage, candidates)
	if routeTrace != nil {
		trace.Status = routeTrace.Status
		trace.SelectedSkillCode = routeTrace.SelectedSkillCode
		trace.RawDecision = routeTrace.RawDecision
		trace.LatencyMs = routeTrace.LatencyMs
		trace.Error = routeTrace.Error
	}
	if err != nil {
		if trace.Error == "" {
			trace.Error = err.Error()
		}
		return nil, "route_error", trace, err
	}
	if selected == nil {
		if trace.Status == "started" {
			trace.Status = "not_matched"
		}
		return nil, "route_none", trace, nil
	}
	return selected, "llm_route", trace, nil
}

func loadCandidateSkills(aiAgent *models.AIAgent) []models.SkillDefinition {
	if aiAgent == nil {
		return nil
	}
	skillIDs := utils.SplitInt64s(aiAgent.SkillIDs)
	if len(skillIDs) == 0 {
		return nil
	}
	ret := make([]models.SkillDefinition, 0, len(skillIDs))
	for _, id := range skillIDs {
		skill := repositories.SkillDefinitionRepository.Get(sqls.DB(), id)
		if skill == nil || skill.Status != enums.StatusOk {
			continue
		}
		ret = append(ret, *skill)
	}
	return ret
}

func routeSkillWithLLM(ctx context.Context, aiConfig *models.AIConfig, userMessage string, candidates []models.SkillDefinition) (*models.SkillDefinition, *RouteTrace, error) {
	trace := &RouteTrace{Status: "started"}
	if aiConfig == nil {
		trace.Status = "config_error"
		trace.Error = "ai config is nil"
		return nil, trace, errorsx.InvalidParam("Skill 路由依赖的 AI 配置不可用")
	}
	if len(candidates) == 0 {
		trace.Status = "no_candidate"
		return nil, trace, nil
	}
	userMessage = strings.TrimSpace(userMessage)
	if userMessage == "" {
		trace.Status = "empty_user_message"
		return nil, trace, nil
	}
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
		lines = append(lines, fmt.Sprintf("- skillCode=%s; name=%s; description=%s", strings.TrimSpace(item.Code), strings.TrimSpace(item.Name), strings.TrimSpace(item.Description)))
	}
	lines = append(lines, "")
	lines = append(lines, "请只输出一个 skillCode 或 NONE。")
	return strings.Join(lines, "\n")
}

func normalizeRouteDecision(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	raw = strings.Trim(raw, "`")
	raw = strings.TrimSpace(raw)
	if idx := strings.Index(raw, "\n"); idx >= 0 {
		raw = raw[:idx]
	}
	raw = strings.TrimSpace(raw)
	raw = strings.Trim(raw, "\"'")
	if strings.EqualFold(raw, "NONE") {
		return "NONE"
	}
	return raw
}
