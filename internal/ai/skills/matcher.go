package skills

import (
	"errors"

	"cs-agent/internal/models"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/repositories"
	"cs-agent/internal/services"

	"github.com/mlogclub/simple/common/strs"
	"github.com/mlogclub/simple/sqls"
)

type intentTriggerConfig struct {
	Intents []string `json:"intents"`
}

// MatchSkill 对单个 SkillDefinition 执行命中判断。
func MatchSkill(ctx RuntimeContext) (*models.SkillDefinition, error) {
	// skills := loadCandidateSkills(ctx)
	// TODO 待实现
	return nil, errors.New("not support")
}

func loadCandidateSkills(ctx RuntimeContext) []models.SkillDefinition {
	if strs.IsNotBlank(ctx.ManualSkillCode) {
		if skill := repositories.SkillDefinitionRepository.GetByCode(sqls.DB(), ctx.ManualSkillCode); skill != nil && skill.Status == enums.StatusOk {
			return []models.SkillDefinition{*skill}
		}
		return nil
	} else {
		return services.SkillDefinitionService.Find(sqls.NewCnd().Eq("status", enums.StatusOk).Desc("priority").Desc("id"))
	}
}
