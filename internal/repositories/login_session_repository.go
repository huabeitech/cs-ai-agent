package repositories

import (
	"cs-agent/internal/models"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
	"gorm.io/gorm"
)

var LoginSessionRepository = newLoginSessionRepository()

func newLoginSessionRepository() *loginSessionRepository {
	return &loginSessionRepository{}
}

type loginSessionRepository struct {
}

func (r *loginSessionRepository) Get(db *gorm.DB, id int64) *models.LoginSession {
	ret := &models.LoginSession{}
	if err := db.First(ret, "id = ?", id).Error; err != nil {
		return nil
	}
	return ret
}

func (r *loginSessionRepository) Take(db *gorm.DB, where ...interface{}) *models.LoginSession {
	ret := &models.LoginSession{}
	if err := db.Take(ret, where...).Error; err != nil {
		return nil
	}
	return ret
}

func (r *loginSessionRepository) Find(db *gorm.DB, cnd *sqls.Cnd) (list []models.LoginSession) {
	cnd.Find(db, &list)
	return
}

func (r *loginSessionRepository) FindOne(db *gorm.DB, cnd *sqls.Cnd) *models.LoginSession {
	ret := &models.LoginSession{}
	if err := cnd.FindOne(db, &ret); err != nil {
		return nil
	}
	return ret
}

func (r *loginSessionRepository) FindPageByParams(db *gorm.DB, params *params.QueryParams) (list []models.LoginSession, paging *sqls.Paging) {
	return r.FindPageByCnd(db, &params.Cnd)
}

func (r *loginSessionRepository) FindPageByCnd(db *gorm.DB, cnd *sqls.Cnd) (list []models.LoginSession, paging *sqls.Paging) {
	cnd.Find(db, &list)
	count := cnd.Count(db, &models.LoginSession{})

	paging = &sqls.Paging{
		Page:  cnd.Paging.Page,
		Limit: cnd.Paging.Limit,
		Total: count,
	}
	return
}

func (r *loginSessionRepository) FindBySql(db *gorm.DB, sqlStr string, paramArr ...interface{}) (list []models.LoginSession) {
	db.Raw(sqlStr, paramArr...).Scan(&list)
	return
}

func (r *loginSessionRepository) CountBySql(db *gorm.DB, sqlStr string, paramArr ...interface{}) (count int64) {
	db.Raw(sqlStr, paramArr...).Count(&count)
	return
}

func (r *loginSessionRepository) Count(db *gorm.DB, cnd *sqls.Cnd) int64 {
	return cnd.Count(db, &models.LoginSession{})
}

func (r *loginSessionRepository) Create(db *gorm.DB, t *models.LoginSession) (err error) {
	err = db.Create(t).Error
	return
}

func (r *loginSessionRepository) Update(db *gorm.DB, t *models.LoginSession) (err error) {
	err = db.Save(t).Error
	return
}

func (r *loginSessionRepository) Updates(db *gorm.DB, id int64, columns map[string]interface{}) (err error) {
	err = db.Model(&models.LoginSession{}).Where("id = ?", id).Updates(columns).Error
	return
}

func (r *loginSessionRepository) UpdateColumn(db *gorm.DB, id int64, name string, value interface{}) (err error) {
	err = db.Model(&models.LoginSession{}).Where("id = ?", id).UpdateColumn(name, value).Error
	return
}

func (r *loginSessionRepository) Delete(db *gorm.DB, id int64) {
	db.Delete(&models.LoginSession{}, "id = ?", id)
}
