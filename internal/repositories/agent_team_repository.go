package repositories

import (
	"cs-agent/internal/models"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
	"gorm.io/gorm"
)

var AgentTeamRepository = newAgentTeamRepository()

func newAgentTeamRepository() *agentTeamRepository {
	return &agentTeamRepository{}
}

type agentTeamRepository struct {
}

func (r *agentTeamRepository) Get(db *gorm.DB, id int64) *models.AgentTeam {
	ret := &models.AgentTeam{}
	if err := db.First(ret, "id = ?", id).Error; err != nil {
		return nil
	}
	return ret
}

func (r *agentTeamRepository) Take(db *gorm.DB, where ...interface{}) *models.AgentTeam {
	ret := &models.AgentTeam{}
	if err := db.Take(ret, where...).Error; err != nil {
		return nil
	}
	return ret
}

func (r *agentTeamRepository) Find(db *gorm.DB, cnd *sqls.Cnd) (list []models.AgentTeam) {
	cnd.Find(db, &list)
	return
}

func (r *agentTeamRepository) FindOne(db *gorm.DB, cnd *sqls.Cnd) *models.AgentTeam {
	ret := &models.AgentTeam{}
	if err := cnd.FindOne(db, &ret); err != nil {
		return nil
	}
	return ret
}

func (r *agentTeamRepository) FindPageByParams(db *gorm.DB, params *params.QueryParams) (list []models.AgentTeam, paging *sqls.Paging) {
	return r.FindPageByCnd(db, &params.Cnd)
}

func (r *agentTeamRepository) FindPageByCnd(db *gorm.DB, cnd *sqls.Cnd) (list []models.AgentTeam, paging *sqls.Paging) {
	cnd.Find(db, &list)
	count := cnd.Count(db, &models.AgentTeam{})

	paging = &sqls.Paging{
		Page:  cnd.Paging.Page,
		Limit: cnd.Paging.Limit,
		Total: count,
	}
	return
}

func (r *agentTeamRepository) FindBySql(db *gorm.DB, sqlStr string, paramArr ...interface{}) (list []models.AgentTeam) {
	db.Raw(sqlStr, paramArr...).Scan(&list)
	return
}

func (r *agentTeamRepository) CountBySql(db *gorm.DB, sqlStr string, paramArr ...interface{}) (count int64) {
	db.Raw(sqlStr, paramArr...).Count(&count)
	return
}

func (r *agentTeamRepository) Count(db *gorm.DB, cnd *sqls.Cnd) int64 {
	return cnd.Count(db, &models.AgentTeam{})
}

func (r *agentTeamRepository) Create(db *gorm.DB, t *models.AgentTeam) (err error) {
	err = db.Create(t).Error
	return
}

func (r *agentTeamRepository) Update(db *gorm.DB, t *models.AgentTeam) (err error) {
	err = db.Save(t).Error
	return
}

func (r *agentTeamRepository) Updates(db *gorm.DB, id int64, columns map[string]interface{}) (err error) {
	err = db.Model(&models.AgentTeam{}).Where("id = ?", id).Updates(columns).Error
	return
}

func (r *agentTeamRepository) UpdateColumn(db *gorm.DB, id int64, name string, value interface{}) (err error) {
	err = db.Model(&models.AgentTeam{}).Where("id = ?", id).UpdateColumn(name, value).Error
	return
}

func (r *agentTeamRepository) Delete(db *gorm.DB, id int64) {
	db.Delete(&models.AgentTeam{}, "id = ?", id)
}

func (r *agentTeamRepository) FindByIds(db *gorm.DB, ids []int64) []models.AgentTeam {
	if len(ids) == 0 {
		return []models.AgentTeam{}
	}
	var list []models.AgentTeam
	db.Where("id IN ?", ids).Find(&list)
	return list
}
