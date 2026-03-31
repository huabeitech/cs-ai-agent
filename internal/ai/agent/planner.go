package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"cs-agent/internal/ai"
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/pkg/utils"
	"cs-agent/internal/repositories"

	"github.com/mlogclub/simple/sqls"
)

type planner struct{}

type plannerResult struct {
	Action    string `json:"action"`
	SkillCode string `json:"skillCode"`
	Reason    string `json:"reason"`
}

func newPlanner() *planner {
	return &planner{}
}

func (p *planner) Plan(ctx context.Context, turnCtx TurnContext, question string) (*Plan, error) {
	skills := p.loadCandidateSkills(turnCtx)
	if len(skills) == 0 || turnCtx.AIConfig == nil {
		return &Plan{
			Action: ActionReply,
			Reason: "fallback_to_rag",
		}, nil
	}

	systemPrompt := buildPlannerSystemPrompt()
	userPrompt := buildPlannerUserPrompt(turnCtx, question, skills)
	result, err := ai.LLM.ChatWithConfig(ctx, turnCtx.AIConfig, systemPrompt, userPrompt)
	if err != nil {
		return &Plan{
			Action: ActionReply,
			Reason: "planner_error_fallback_to_rag",
		}, nil
	}
	parsed, ok := parsePlannerResult(result.Content)
	if !ok {
		return &Plan{
			Action: ActionReply,
			Reason: "planner_invalid_fallback_to_rag",
		}, nil
	}
	return p.normalizePlan(parsed, skills), nil
}

func (p *planner) loadCandidateSkills(turnCtx TurnContext) []models.SkillDefinition {
	if turnCtx.AIAgent == nil {
		return nil
	}
	skillIDs := utils.SplitInt64s(turnCtx.AIAgent.SkillIDs)
	if len(skillIDs) == 0 {
		return nil
	}
	allSkills := repositories.SkillDefinitionRepository.Find(sqls.DB(), sqls.NewCnd().
		In("id", skillIDs).
		Eq("status", enums.StatusOk).
		Asc("priority").
		Desc("id"))
	if len(allSkills) == 0 {
		return nil
	}
	ordered := make([]models.SkillDefinition, 0, len(allSkills))
	for _, id := range skillIDs {
		index := slices.IndexFunc(allSkills, func(skill models.SkillDefinition) bool {
			return skill.ID == id
		})
		if index >= 0 {
			ordered = append(ordered, allSkills[index])
		}
	}
	return ordered
}

func (p *planner) normalizePlan(parsed plannerResult, skills []models.SkillDefinition) *Plan {
	action := Action(strings.TrimSpace(strings.ToLower(parsed.Action)))
	switch action {
	case ActionSkill:
		skillCode := strings.TrimSpace(parsed.SkillCode)
		if skillCode != "" && hasSkill(skills, skillCode) {
			return &Plan{
				Action:    ActionSkill,
				SkillCode: skillCode,
				Reason:    strings.TrimSpace(parsed.Reason),
			}
		}
	case ActionHandoff:
		return &Plan{
			Action: ActionHandoff,
			Reason: normalizeReason(parsed.Reason, "planner_handoff"),
		}
	case ActionRAG, ActionReply:
		return &Plan{
			Action: ActionRAG,
			Reason: normalizeReason(parsed.Reason, "planner_rag"),
		}
	}
	return &Plan{
		Action: ActionRAG,
		Reason: "planner_default_rag",
	}
}

func buildPlannerSystemPrompt() string {
	return strings.TrimSpace(`你是客服对话路由器。你的任务不是直接回答用户，而是在 rag、skill、handoff 三种动作中选择一个。

输出必须是 JSON，字段如下：
{"action":"rag|skill|handoff","skillCode":"","reason":""}

规则：
1. 如果问题适合由某个已提供的 skill 处理，返回 action=skill，并填写准确的 skillCode。
2. 如果问题明显需要人工介入、投诉升级、超出机器人能力边界，返回 action=handoff。
3. 其他情况返回 action=rag。
4. 只能从给定 skill 列表中选择，不允许编造 skillCode。
5. 不要输出 markdown，不要输出解释，只输出 JSON。`)
}

func buildPlannerUserPrompt(turnCtx TurnContext, question string, skills []models.SkillDefinition) string {
	var builder strings.Builder
	builder.WriteString("AI Agent：\n")
	if turnCtx.AIAgent != nil {
		builder.WriteString(fmt.Sprintf("- 名称：%s\n", strings.TrimSpace(turnCtx.AIAgent.Name)))
		builder.WriteString(fmt.Sprintf("- 描述：%s\n", strings.TrimSpace(turnCtx.AIAgent.Description)))
	}
	builder.WriteString("\n可用 Skills：\n")
	for _, skill := range skills {
		builder.WriteString(fmt.Sprintf("- code=%s; name=%s; description=%s; executionMode=%s\n",
			skill.Code,
			skill.Name,
			strings.TrimSpace(skill.Description),
			skill.ExecutionMode,
		))
	}
	builder.WriteString("\n用户问题：\n")
	builder.WriteString(strings.TrimSpace(question))
	return builder.String()
}

func parsePlannerResult(raw string) (plannerResult, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return plannerResult{}, false
	}
	if start := strings.Index(raw, "{"); start >= 0 {
		if end := strings.LastIndex(raw, "}"); end > start {
			raw = raw[start : end+1]
		}
	}
	var result plannerResult
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return plannerResult{}, false
	}
	if strings.TrimSpace(result.Action) == "" {
		return plannerResult{}, false
	}
	return result, true
}

func hasSkill(skills []models.SkillDefinition, skillCode string) bool {
	skillCode = strings.TrimSpace(skillCode)
	for _, skill := range skills {
		if skill.Code == skillCode {
			return true
		}
	}
	return false
}

func normalizeReason(reason string, fallback string) string {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return fallback
	}
	return reason
}
