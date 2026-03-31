package skills

import (
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/pkg/errorsx"
	"cs-agent/internal/repositories"

	"github.com/mlogclub/simple/common/strs"
	"github.com/mlogclub/simple/sqls"
)

type intentTriggerConfig struct {
	Intents []string `json:"intents"`
}

// MatchSkill 对单个 SkillDefinition 执行命中判断。
func MatchSkill(ctx RuntimeContext) (*models.SkillDefinition, error) {
	if strs.IsNotBlank(ctx.ManualSkillCode) {
		skill := repositories.SkillDefinitionRepository.GetByCode(sqls.DB(), ctx.ManualSkillCode)
		if skill == nil || skill.Status != enums.StatusOk {
			return nil, errorsx.InvalidParam("Skill 不存在或未启用")
		}
		return skill, nil
	}
	return nil, nil
}

func loadCandidateSkills(ctx RuntimeContext) []models.SkillDefinition {
	if strs.IsNotBlank(ctx.ManualSkillCode) {
		if skill := repositories.SkillDefinitionRepository.GetByCode(sqls.DB(), ctx.ManualSkillCode); skill != nil && skill.Status == enums.StatusOk {
			return []models.SkillDefinition{*skill}
		}
		return nil
	} else {
		return repositories.SkillDefinitionRepository.Find(sqls.DB(), sqls.NewCnd().Eq("status", enums.StatusOk).Desc("priority").Desc("id"))
	}
}
