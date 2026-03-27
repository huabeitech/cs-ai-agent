package services

import (
	"cs-agent/internal/models"
	"cs-agent/internal/repositories"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
)

var SkillDefinitionService = newSkillDefinitionService()

func newSkillDefinitionService() *skillDefinitionService {
	return &skillDefinitionService{}
}

type skillDefinitionService struct {
}

func (s *skillDefinitionService) Get(id int64) *models.SkillDefinition {
	return repositories.SkillDefinitionRepository.Get(sqls.DB(), id)
}

func (s *skillDefinitionService) Take(where ...interface{}) *models.SkillDefinition {
	return repositories.SkillDefinitionRepository.Take(sqls.DB(), where...)
}

func (s *skillDefinitionService) Find(cnd *sqls.Cnd) []models.SkillDefinition {
	return repositories.SkillDefinitionRepository.Find(sqls.DB(), cnd)
}

func (s *skillDefinitionService) FindOne(cnd *sqls.Cnd) *models.SkillDefinition {
	return repositories.SkillDefinitionRepository.FindOne(sqls.DB(), cnd)
}

func (s *skillDefinitionService) FindPageByParams(params *params.QueryParams) (list []models.SkillDefinition, paging *sqls.Paging) {
	return repositories.SkillDefinitionRepository.FindPageByParams(sqls.DB(), params)
}

func (s *skillDefinitionService) FindPageByCnd(cnd *sqls.Cnd) (list []models.SkillDefinition, paging *sqls.Paging) {
	return repositories.SkillDefinitionRepository.FindPageByCnd(sqls.DB(), cnd)
}

func (s *skillDefinitionService) Count(cnd *sqls.Cnd) int64 {
	return repositories.SkillDefinitionRepository.Count(sqls.DB(), cnd)
}

func (s *skillDefinitionService) Create(t *models.SkillDefinition) error {
	return repositories.SkillDefinitionRepository.Create(sqls.DB(), t)
}

func (s *skillDefinitionService) Update(t *models.SkillDefinition) error {
	return repositories.SkillDefinitionRepository.Update(sqls.DB(), t)
}

func (s *skillDefinitionService) Updates(id int64, columns map[string]interface{}) error {
	return repositories.SkillDefinitionRepository.Updates(sqls.DB(), id, columns)
}

func (s *skillDefinitionService) UpdateColumn(id int64, name string, value interface{}) error {
	return repositories.SkillDefinitionRepository.UpdateColumn(sqls.DB(), id, name, value)
}

func (s *skillDefinitionService) Delete(id int64) {
	repositories.SkillDefinitionRepository.Delete(sqls.DB(), id)
}

func (s *skillDefinitionService) NextPriority() int {
	if max := repositories.SkillDefinitionRepository.FindOne(sqls.DB(), sqls.NewCnd().Desc("priority").Desc("id")); max != nil {
		return max.Priority + 1
	}
	return 1
}

func (s *skillDefinitionService) UpdatePriority(ids []int64) error {
	return sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		for i, id := range ids {
			if err := repositories.SkillDefinitionRepository.UpdateColumn(ctx.Tx, id, "priority", i+1); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *skillDefinitionService) GetByCode(code string) *models.SkillDefinition {
	return repositories.SkillDefinitionRepository.GetByCode(sqls.DB(), code)
}
