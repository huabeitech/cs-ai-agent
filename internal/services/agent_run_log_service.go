package services

import (
	"cs-agent/internal/models"
	"cs-agent/internal/repositories"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
)

var AgentRunLogService = newAgentRunLogService()

func newAgentRunLogService() *agentRunLogService {
	return &agentRunLogService{}
}

type agentRunLogService struct{}

func (s *agentRunLogService) Get(id int64) *models.AgentRunLog {
	return repositories.AgentRunLogRepository.Get(sqls.DB(), id)
}

func (s *agentRunLogService) Take(where ...interface{}) *models.AgentRunLog {
	return repositories.AgentRunLogRepository.Take(sqls.DB(), where...)
}

func (s *agentRunLogService) Find(cnd *sqls.Cnd) []models.AgentRunLog {
	return repositories.AgentRunLogRepository.Find(sqls.DB(), cnd)
}

func (s *agentRunLogService) FindOne(cnd *sqls.Cnd) *models.AgentRunLog {
	return repositories.AgentRunLogRepository.FindOne(sqls.DB(), cnd)
}

func (s *agentRunLogService) FindPageByParams(params *params.QueryParams) (list []models.AgentRunLog, paging *sqls.Paging) {
	return repositories.AgentRunLogRepository.FindPageByParams(sqls.DB(), params)
}

func (s *agentRunLogService) FindPageByCnd(cnd *sqls.Cnd) (list []models.AgentRunLog, paging *sqls.Paging) {
	return repositories.AgentRunLogRepository.FindPageByCnd(sqls.DB(), cnd)
}

func (s *agentRunLogService) Count(cnd *sqls.Cnd) int64 {
	return repositories.AgentRunLogRepository.Count(sqls.DB(), cnd)
}

func (s *agentRunLogService) Create(t *models.AgentRunLog) error {
	return repositories.AgentRunLogRepository.Create(sqls.DB(), t)
}

func (s *agentRunLogService) Update(t *models.AgentRunLog) error {
	return repositories.AgentRunLogRepository.Update(sqls.DB(), t)
}

func (s *agentRunLogService) Updates(id int64, columns map[string]interface{}) error {
	return repositories.AgentRunLogRepository.Updates(sqls.DB(), id, columns)
}

func (s *agentRunLogService) UpdateColumn(id int64, name string, value interface{}) error {
	return repositories.AgentRunLogRepository.UpdateColumn(sqls.DB(), id, name, value)
}

func (s *agentRunLogService) Delete(id int64) {
	repositories.AgentRunLogRepository.Delete(sqls.DB(), id)
}
