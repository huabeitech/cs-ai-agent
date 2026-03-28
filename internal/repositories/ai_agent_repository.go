package repositories

import (
	"cs-agent/internal/models"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
	"gorm.io/gorm"
)

var AIAgentRepository = newAIAgentRepository()

func newAIAgentRepository() *aIAgentRepository {
	return &aIAgentRepository{}
}

type aIAgentRepository struct {
}

func (r *aIAgentRepository) Get(db *gorm.DB, id int64) *models.AIAgent {
	ret := &models.AIAgent{}
	if err := db.First(ret, "id = ?", id).Error; err != nil {
		return nil
	}
	return ret
}

func (r *aIAgentRepository) Take(db *gorm.DB, where ...interface{}) *models.AIAgent {
	ret := &models.AIAgent{}
	if err := db.Take(ret, where...).Error; err != nil {
		return nil
	}
	return ret
}

func (r *aIAgentRepository) Find(db *gorm.DB, cnd *sqls.Cnd) (list []models.AIAgent) {
	cnd.Find(db, &list)
	return
}

func (r *aIAgentRepository) FindOne(db *gorm.DB, cnd *sqls.Cnd) *models.AIAgent {
	ret := &models.AIAgent{}
	if err := cnd.FindOne(db, &ret); err != nil {
		return nil
	}
	return ret
}

func (r *aIAgentRepository) FindPageByParams(db *gorm.DB, params *params.QueryParams) (list []models.AIAgent, paging *sqls.Paging) {
	return r.FindPageByCnd(db, &params.Cnd)
}

func (r *aIAgentRepository) FindPageByCnd(db *gorm.DB, cnd *sqls.Cnd) (list []models.AIAgent, paging *sqls.Paging) {
	cnd.Find(db, &list)
	count := cnd.Count(db, &models.AIAgent{})

	paging = &sqls.Paging{
		Page:  cnd.Paging.Page,
		Limit: cnd.Paging.Limit,
		Total: count,
	}
	return
}

func (r *aIAgentRepository) FindBySql(db *gorm.DB, sqlStr string, paramArr ...interface{}) (list []models.AIAgent) {
	db.Raw(sqlStr, paramArr...).Scan(&list)
	return
}

func (r *aIAgentRepository) CountBySql(db *gorm.DB, sqlStr string, paramArr ...interface{}) (count int64) {
	db.Raw(sqlStr, paramArr...).Count(&count)
	return
}

func (r *aIAgentRepository) Count(db *gorm.DB, cnd *sqls.Cnd) int64 {
	return cnd.Count(db, &models.AIAgent{})
}

func (r *aIAgentRepository) Create(db *gorm.DB, t *models.AIAgent) (err error) {
	err = db.Create(t).Error
	return
}

func (r *aIAgentRepository) Update(db *gorm.DB, t *models.AIAgent) (err error) {
	err = db.Save(t).Error
	return
}

func (r *aIAgentRepository) Updates(db *gorm.DB, id int64, columns map[string]interface{}) (err error) {
	err = db.Model(&models.AIAgent{}).Where("id = ?", id).Updates(columns).Error
	return
}

func (r *aIAgentRepository) UpdateColumn(db *gorm.DB, id int64, name string, value interface{}) (err error) {
	err = db.Model(&models.AIAgent{}).Where("id = ?", id).UpdateColumn(name, value).Error
	return
}

func (r *aIAgentRepository) Delete(db *gorm.DB, id int64) {
	db.Delete(&models.AIAgent{}, "id = ?", id)
}

func (r *aIAgentRepository) FindByIds(db *gorm.DB, ids []int64) []models.AIAgent {
	if len(ids) == 0 {
		return []models.AIAgent{}
	}
	var list []models.AIAgent
	db.Where("id IN ?", ids).Find(&list)
	return list
}
