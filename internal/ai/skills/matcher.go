package skills

import (
	"context"
	"strings"

	"cs-agent/internal/models"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/pkg/errorsx"

	"github.com/mlogclub/simple/common/strs"
)

type intentTriggerConfig struct {
	Intents []string `json:"intents"`
}

// MatchSkill 对单个 SkillDefinition 执行命中判断。
func MatchSkill(execCtx context.Context, ctx RuntimeContext) (*models.SkillDefinition, string, *RouteTrace, error) {
	loader := newCandidateLoader()
	if strs.IsNotBlank(ctx.ManualSkillCode) {
		skill := loader.findManualSkillDefinition(ctx.ManualSkillCode)
		if skill == nil || skill.Status != enums.StatusOk {
			return nil, "", nil, errorsx.InvalidParam("Skill 不存在或未启用")
		}
		return skill, "manual_skill_code", &RouteTrace{
			Status:            "manual_selected",
			SelectedSkillCode: skill.Code,
		}, nil
	}

	candidates := loader.loadCandidateSkills(ctx.AIAgent)
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

	selected, routeTrace, err := routeSkillWithLLM(execCtx, ctx.AIConfig, ctx.UserMessage, candidates)
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
