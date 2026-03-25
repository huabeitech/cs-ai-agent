package repositories

import (
	"cs-agent/internal/models"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
	"gorm.io/gorm"
)

var AgentProfileRepository = newAgentProfileRepository()

func newAgentProfileRepository() *agentProfileRepository {
	return &agentProfileRepository{}
}

type agentProfileRepository struct {
}

func (r *agentProfileRepository) Get(db *gorm.DB, id int64) *models.AgentProfile {
	ret := &models.AgentProfile{}
	if err := db.First(ret, "id = ?", id).Error; err != nil {
		return nil
	}
	return ret
}

func (r *agentProfileRepository) Take(db *gorm.DB, where ...interface{}) *models.AgentProfile {
	ret := &models.AgentProfile{}
	if err := db.Take(ret, where...).Error; err != nil {
		return nil
	}
	return ret
}

func (r *agentProfileRepository) Find(db *gorm.DB, cnd *sqls.Cnd) (list []models.AgentProfile) {
	cnd.Find(db, &list)
	return
}

func (r *agentProfileRepository) FindOne(db *gorm.DB, cnd *sqls.Cnd) *models.AgentProfile {
	ret := &models.AgentProfile{}
	if err := cnd.FindOne(db, &ret); err != nil {
		return nil
	}
	return ret
}

func (r *agentProfileRepository) FindPageByParams(db *gorm.DB, params *params.QueryParams) (list []models.AgentProfile, paging *sqls.Paging) {
	return r.FindPageByCnd(db, &params.Cnd)
}

func (r *agentProfileRepository) FindPageByCnd(db *gorm.DB, cnd *sqls.Cnd) (list []models.AgentProfile, paging *sqls.Paging) {
	cnd.Find(db, &list)
	count := cnd.Count(db, &models.AgentProfile{})

	paging = &sqls.Paging{
		Page:  cnd.Paging.Page,
		Limit: cnd.Paging.Limit,
		Total: count,
	}
	return
}

func (r *agentProfileRepository) FindBySql(db *gorm.DB, sqlStr string, paramArr ...interface{}) (list []models.AgentProfile) {
	db.Raw(sqlStr, paramArr...).Scan(&list)
	return
}

func (r *agentProfileRepository) CountBySql(db *gorm.DB, sqlStr string, paramArr ...interface{}) (count int64) {
	db.Raw(sqlStr, paramArr...).Count(&count)
	return
}

func (r *agentProfileRepository) Count(db *gorm.DB, cnd *sqls.Cnd) int64 {
	return cnd.Count(db, &models.AgentProfile{})
}

func (r *agentProfileRepository) Create(db *gorm.DB, t *models.AgentProfile) (err error) {
	err = db.Create(t).Error
	return
}

func (r *agentProfileRepository) Update(db *gorm.DB, t *models.AgentProfile) (err error) {
	err = db.Save(t).Error
	return
}

func (r *agentProfileRepository) Updates(db *gorm.DB, id int64, columns map[string]interface{}) (err error) {
	err = db.Model(&models.AgentProfile{}).Where("id = ?", id).Updates(columns).Error
	return
}

func (r *agentProfileRepository) UpdateColumn(db *gorm.DB, id int64, name string, value interface{}) (err error) {
	err = db.Model(&models.AgentProfile{}).Where("id = ?", id).UpdateColumn(name, value).Error
	return
}

func (r *agentProfileRepository) Delete(db *gorm.DB, id int64) {
	db.Delete(&models.AgentProfile{}, "id = ?", id)
}
