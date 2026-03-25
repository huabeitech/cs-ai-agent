package services

import (
	"cs-agent/internal/models"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
)

var KnowledgeRetrieveLogService = newKnowledgeRetrieveLogService()

func newKnowledgeRetrieveLogService() *knowledgeRetrieveLogService {
	return &knowledgeRetrieveLogService{}
}

type knowledgeRetrieveLogService struct {
}

func (s *knowledgeRetrieveLogService) Get(id int64) *models.KnowledgeRetrieveLog {
	ret := &models.KnowledgeRetrieveLog{}
	if err := sqls.DB().First(ret, "id = ?", id).Error; err != nil {
		return nil
	}
	return ret
}

func (s *knowledgeRetrieveLogService) FindPageByParams(params *params.QueryParams) (list []models.KnowledgeRetrieveLog, paging *sqls.Paging) {
	cnd := &params.Cnd
	cnd.Find(sqls.DB(), &list)
	count := cnd.Count(sqls.DB(), &models.KnowledgeRetrieveLog{})
	paging = &sqls.Paging{
		Page:  cnd.Paging.Page,
		Limit: cnd.Paging.Limit,
		Total: count,
	}
	return
}

func (s *knowledgeRetrieveLogService) FindHitsByRetrieveLogID(retrieveLogID int64) []models.KnowledgeRetrieveHit {
	if retrieveLogID <= 0 {
		return nil
	}
	var list []models.KnowledgeRetrieveHit
	sqls.DB().Where("retrieve_log_id = ?", retrieveLogID).Order("rank_no asc, id asc").Find(&list)
	return list
}
