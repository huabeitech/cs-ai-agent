package repositories

import (
	"cs-agent/internal/models"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
	"gorm.io/gorm"
)

var LoginCredentialLogRepository = newLoginCredentialLogRepository()

func newLoginCredentialLogRepository() *loginCredentialLogRepository {
	return &loginCredentialLogRepository{}
}

type loginCredentialLogRepository struct {
}

func (r *loginCredentialLogRepository) Get(db *gorm.DB, id int64) *models.LoginCredentialLog {
	ret := &models.LoginCredentialLog{}
	if err := db.First(ret, "id = ?", id).Error; err != nil {
		return nil
	}
	return ret
}

func (r *loginCredentialLogRepository) Take(db *gorm.DB, where ...interface{}) *models.LoginCredentialLog {
	ret := &models.LoginCredentialLog{}
	if err := db.Take(ret, where...).Error; err != nil {
		return nil
	}
	return ret
}

func (r *loginCredentialLogRepository) Find(db *gorm.DB, cnd *sqls.Cnd) (list []models.LoginCredentialLog) {
	cnd.Find(db, &list)
	return
}

func (r *loginCredentialLogRepository) FindOne(db *gorm.DB, cnd *sqls.Cnd) *models.LoginCredentialLog {
	ret := &models.LoginCredentialLog{}
	if err := cnd.FindOne(db, &ret); err != nil {
		return nil
	}
	return ret
}

func (r *loginCredentialLogRepository) FindPageByParams(db *gorm.DB, params *params.QueryParams) (list []models.LoginCredentialLog, paging *sqls.Paging) {
	return r.FindPageByCnd(db, &params.Cnd)
}

func (r *loginCredentialLogRepository) FindPageByCnd(db *gorm.DB, cnd *sqls.Cnd) (list []models.LoginCredentialLog, paging *sqls.Paging) {
	cnd.Find(db, &list)
	count := cnd.Count(db, &models.LoginCredentialLog{})

	paging = &sqls.Paging{
		Page:  cnd.Paging.Page,
		Limit: cnd.Paging.Limit,
		Total: count,
	}
	return
}

func (r *loginCredentialLogRepository) FindBySql(db *gorm.DB, sqlStr string, paramArr ...interface{}) (list []models.LoginCredentialLog) {
	db.Raw(sqlStr, paramArr...).Scan(&list)
	return
}

func (r *loginCredentialLogRepository) CountBySql(db *gorm.DB, sqlStr string, paramArr ...interface{}) (count int64) {
	db.Raw(sqlStr, paramArr...).Count(&count)
	return
}

func (r *loginCredentialLogRepository) Count(db *gorm.DB, cnd *sqls.Cnd) int64 {
	return cnd.Count(db, &models.LoginCredentialLog{})
}

func (r *loginCredentialLogRepository) Create(db *gorm.DB, t *models.LoginCredentialLog) (err error) {
	err = db.Create(t).Error
	return
}

func (r *loginCredentialLogRepository) Update(db *gorm.DB, t *models.LoginCredentialLog) (err error) {
	err = db.Save(t).Error
	return
}

func (r *loginCredentialLogRepository) Updates(db *gorm.DB, id int64, columns map[string]interface{}) (err error) {
	err = db.Model(&models.LoginCredentialLog{}).Where("id = ?", id).Updates(columns).Error
	return
}

func (r *loginCredentialLogRepository) UpdateColumn(db *gorm.DB, id int64, name string, value interface{}) (err error) {
	err = db.Model(&models.LoginCredentialLog{}).Where("id = ?", id).UpdateColumn(name, value).Error
	return
}

func (r *loginCredentialLogRepository) Delete(db *gorm.DB, id int64) {
	db.Delete(&models.LoginCredentialLog{}, "id = ?", id)
}
