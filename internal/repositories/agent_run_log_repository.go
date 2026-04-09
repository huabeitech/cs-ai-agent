package repositories

import (
	"cs-agent/internal/models"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
	"gorm.io/gorm"
)

var AgentRunLogRepository = newAgentRunLogRepository()

func newAgentRunLogRepository() *agentRunLogRepository {
	return &agentRunLogRepository{}
}

type agentRunLogRepository struct{}

func (r *agentRunLogRepository) Get(db *gorm.DB, id int64) *models.AgentRunLog {
	ret := &models.AgentRunLog{}
	if err := db.First(ret, "id = ?", id).Error; err != nil {
		return nil
	}
	return ret
}

func (r *agentRunLogRepository) Take(db *gorm.DB, where ...interface{}) *models.AgentRunLog {
	ret := &models.AgentRunLog{}
	if err := db.Take(ret, where...).Error; err != nil {
		return nil
	}
	return ret
}

func (r *agentRunLogRepository) Find(db *gorm.DB, cnd *sqls.Cnd) (list []models.AgentRunLog) {
	cnd.Find(db, &list)
	return
}

func (r *agentRunLogRepository) FindOne(db *gorm.DB, cnd *sqls.Cnd) *models.AgentRunLog {
	ret := &models.AgentRunLog{}
	if err := cnd.FindOne(db, &ret); err != nil {
		return nil
	}
	return ret
}

func (r *agentRunLogRepository) FindPageByParams(db *gorm.DB, params *params.QueryParams) (list []models.AgentRunLog, paging *sqls.Paging) {
	return r.FindPageByCnd(db, &params.Cnd)
}

func (r *agentRunLogRepository) FindPageByCnd(db *gorm.DB, cnd *sqls.Cnd) (list []models.AgentRunLog, paging *sqls.Paging) {
	cnd.Find(db, &list)
	count := cnd.Count(db, &models.AgentRunLog{})

	paging = &sqls.Paging{
		Page:  cnd.Paging.Page,
		Limit: cnd.Paging.Limit,
		Total: count,
	}
	return
}

func (r *agentRunLogRepository) Count(db *gorm.DB, cnd *sqls.Cnd) int64 {
	return cnd.Count(db, &models.AgentRunLog{})
}

func (r *agentRunLogRepository) Create(db *gorm.DB, t *models.AgentRunLog) (err error) {
	err = db.Create(t).Error
	return
}

func (r *agentRunLogRepository) Update(db *gorm.DB, t *models.AgentRunLog) (err error) {
	err = db.Save(t).Error
	return
}

func (r *agentRunLogRepository) Updates(db *gorm.DB, id int64, columns map[string]interface{}) (err error) {
	err = db.Model(&models.AgentRunLog{}).Where("id = ?", id).Updates(columns).Error
	return
}

func (r *agentRunLogRepository) UpdateColumn(db *gorm.DB, id int64, name string, value interface{}) (err error) {
	err = db.Model(&models.AgentRunLog{}).Where("id = ?", id).UpdateColumn(name, value).Error
	return
}

func (r *agentRunLogRepository) Delete(db *gorm.DB, id int64) {
	db.Delete(&models.AgentRunLog{}, "id = ?", id)
}
