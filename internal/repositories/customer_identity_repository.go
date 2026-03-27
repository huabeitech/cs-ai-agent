package repositories

import (
	"cs-agent/internal/models"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
	"gorm.io/gorm"
)

var CustomerIdentityRepository = newCustomerIdentityRepository()

func newCustomerIdentityRepository() *customerIdentityRepository {
	return &customerIdentityRepository{}
}

type customerIdentityRepository struct {
}

func (r *customerIdentityRepository) Get(db *gorm.DB, id int64) *models.CustomerIdentity {
	ret := &models.CustomerIdentity{}
	if err := db.First(ret, "id = ?", id).Error; err != nil {
		return nil
	}
	return ret
}

func (r *customerIdentityRepository) Take(db *gorm.DB, where ...interface{}) *models.CustomerIdentity {
	ret := &models.CustomerIdentity{}
	if err := db.Take(ret, where...).Error; err != nil {
		return nil
	}
	return ret
}

func (r *customerIdentityRepository) Find(db *gorm.DB, cnd *sqls.Cnd) (list []models.CustomerIdentity) {
	cnd.Find(db, &list)
	return
}

func (r *customerIdentityRepository) FindOne(db *gorm.DB, cnd *sqls.Cnd) *models.CustomerIdentity {
	ret := &models.CustomerIdentity{}
	if err := cnd.FindOne(db, &ret); err != nil {
		return nil
	}
	return ret
}

func (r *customerIdentityRepository) FindPageByParams(db *gorm.DB, params *params.QueryParams) (list []models.CustomerIdentity, paging *sqls.Paging) {
	return r.FindPageByCnd(db, &params.Cnd)
}

func (r *customerIdentityRepository) FindPageByCnd(db *gorm.DB, cnd *sqls.Cnd) (list []models.CustomerIdentity, paging *sqls.Paging) {
	cnd.Find(db, &list)
	count := cnd.Count(db, &models.CustomerIdentity{})

	paging = &sqls.Paging{
		Page:  cnd.Paging.Page,
		Limit: cnd.Paging.Limit,
		Total: count,
	}
	return
}

func (r *customerIdentityRepository) FindBySql(db *gorm.DB, sqlStr string, paramArr... interface{}) (list []models.CustomerIdentity) {
	db.Raw(sqlStr, paramArr...).Scan(&list)
	return
}

func (r *customerIdentityRepository) CountBySql(db *gorm.DB, sqlStr string, paramArr... interface{}) (count int64) {
	db.Raw(sqlStr, paramArr...).Count(&count)
	return
}

func (r *customerIdentityRepository) Count(db *gorm.DB, cnd *sqls.Cnd) int64 {
	return cnd.Count(db, &models.CustomerIdentity{})
}

func (r *customerIdentityRepository) Create(db *gorm.DB, t *models.CustomerIdentity) (err error) {
	err = db.Create(t).Error
	return
}

func (r *customerIdentityRepository) Update(db *gorm.DB, t *models.CustomerIdentity) (err error) {
	err = db.Save(t).Error
	return
}

func (r *customerIdentityRepository) Updates(db *gorm.DB, id int64, columns map[string]interface{}) (err error) {
	err = db.Model(&models.CustomerIdentity{}).Where("id = ?", id).Updates(columns).Error
	return
}

func (r *customerIdentityRepository) UpdateColumn(db *gorm.DB, id int64, name string, value interface{}) (err error) {
	err = db.Model(&models.CustomerIdentity{}).Where("id = ?", id).UpdateColumn(name, value).Error
	return
}

func (r *customerIdentityRepository) Delete(db *gorm.DB, id int64) {
	db.Delete(&models.CustomerIdentity{}, "id = ?", id)
}

