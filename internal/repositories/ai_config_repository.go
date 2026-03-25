package repositories

import (
	"cs-agent/internal/models"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
	"gorm.io/gorm"
)

var AIConfigRepository = newAIConfigRepository()

func newAIConfigRepository() *aIConfigRepository {
	return &aIConfigRepository{}
}

type aIConfigRepository struct {
}

func (r *aIConfigRepository) Get(db *gorm.DB, id int64) *models.AIConfig {
	ret := &models.AIConfig{}
	if err := db.First(ret, "id = ?", id).Error; err != nil {
		return nil
	}
	return ret
}

func (r *aIConfigRepository) Take(db *gorm.DB, where ...interface{}) *models.AIConfig {
	ret := &models.AIConfig{}
	if err := db.Take(ret, where...).Error; err != nil {
		return nil
	}
	return ret
}

func (r *aIConfigRepository) Find(db *gorm.DB, cnd *sqls.Cnd) (list []models.AIConfig) {
	cnd.Find(db, &list)
	return
}

func (r *aIConfigRepository) FindOne(db *gorm.DB, cnd *sqls.Cnd) *models.AIConfig {
	ret := &models.AIConfig{}
	if err := cnd.FindOne(db, &ret); err != nil {
		return nil
	}
	return ret
}

func (r *aIConfigRepository) FindPageByParams(db *gorm.DB, params *params.QueryParams) (list []models.AIConfig, paging *sqls.Paging) {
	return r.FindPageByCnd(db, &params.Cnd)
}

func (r *aIConfigRepository) FindPageByCnd(db *gorm.DB, cnd *sqls.Cnd) (list []models.AIConfig, paging *sqls.Paging) {
	cnd.Find(db, &list)
	count := cnd.Count(db, &models.AIConfig{})

	paging = &sqls.Paging{
		Page:  cnd.Paging.Page,
		Limit: cnd.Paging.Limit,
		Total: count,
	}
	return
}

func (r *aIConfigRepository) FindBySql(db *gorm.DB, sqlStr string, paramArr... interface{}) (list []models.AIConfig) {
	db.Raw(sqlStr, paramArr...).Scan(&list)
	return
}

func (r *aIConfigRepository) CountBySql(db *gorm.DB, sqlStr string, paramArr... interface{}) (count int64) {
	db.Raw(sqlStr, paramArr...).Count(&count)
	return
}

func (r *aIConfigRepository) Count(db *gorm.DB, cnd *sqls.Cnd) int64 {
	return cnd.Count(db, &models.AIConfig{})
}

func (r *aIConfigRepository) Create(db *gorm.DB, t *models.AIConfig) (err error) {
	err = db.Create(t).Error
	return
}

func (r *aIConfigRepository) Update(db *gorm.DB, t *models.AIConfig) (err error) {
	err = db.Save(t).Error
	return
}

func (r *aIConfigRepository) Updates(db *gorm.DB, id int64, columns map[string]interface{}) (err error) {
	err = db.Model(&models.AIConfig{}).Where("id = ?", id).Updates(columns).Error
	return
}

func (r *aIConfigRepository) UpdateColumn(db *gorm.DB, id int64, name string, value interface{}) (err error) {
	err = db.Model(&models.AIConfig{}).Where("id = ?", id).UpdateColumn(name, value).Error
	return
}

func (r *aIConfigRepository) Delete(db *gorm.DB, id int64) {
	db.Delete(&models.AIConfig{}, "id = ?", id)
}

