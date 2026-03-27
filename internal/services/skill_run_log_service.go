package services

import (
	"cs-agent/internal/models"
	"cs-agent/internal/repositories"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
)

var SkillRunLogService = newSkillRunLogService()

func newSkillRunLogService() *skillRunLogService {
	return &skillRunLogService{}
}

type skillRunLogService struct {
}

func (s *skillRunLogService) Get(id int64) *models.SkillRunLog {
	return repositories.SkillRunLogRepository.Get(sqls.DB(), id)
}

func (s *skillRunLogService) Take(where ...interface{}) *models.SkillRunLog {
	return repositories.SkillRunLogRepository.Take(sqls.DB(), where...)
}

func (s *skillRunLogService) Find(cnd *sqls.Cnd) []models.SkillRunLog {
	return repositories.SkillRunLogRepository.Find(sqls.DB(), cnd)
}

func (s *skillRunLogService) FindOne(cnd *sqls.Cnd) *models.SkillRunLog {
	return repositories.SkillRunLogRepository.FindOne(sqls.DB(), cnd)
}

func (s *skillRunLogService) FindPageByParams(params *params.QueryParams) (list []models.SkillRunLog, paging *sqls.Paging) {
	return repositories.SkillRunLogRepository.FindPageByParams(sqls.DB(), params)
}

func (s *skillRunLogService) FindPageByCnd(cnd *sqls.Cnd) (list []models.SkillRunLog, paging *sqls.Paging) {
	return repositories.SkillRunLogRepository.FindPageByCnd(sqls.DB(), cnd)
}

func (s *skillRunLogService) Count(cnd *sqls.Cnd) int64 {
	return repositories.SkillRunLogRepository.Count(sqls.DB(), cnd)
}

func (s *skillRunLogService) Create(t *models.SkillRunLog) error {
	return repositories.SkillRunLogRepository.Create(sqls.DB(), t)
}

func (s *skillRunLogService) Update(t *models.SkillRunLog) error {
	return repositories.SkillRunLogRepository.Update(sqls.DB(), t)
}

func (s *skillRunLogService) Updates(id int64, columns map[string]interface{}) error {
	return repositories.SkillRunLogRepository.Updates(sqls.DB(), id, columns)
}

func (s *skillRunLogService) UpdateColumn(id int64, name string, value interface{}) error {
	return repositories.SkillRunLogRepository.UpdateColumn(sqls.DB(), id, name, value)
}

func (s *skillRunLogService) Delete(id int64) {
	repositories.SkillRunLogRepository.Delete(sqls.DB(), id)
}
