package repositories

import (
	"cs-agent/internal/models"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
	"gorm.io/gorm"
)

var CustomerContactRepository = newCustomerContactRepository()

func newCustomerContactRepository() *customerContactRepository {
	return &customerContactRepository{}
}

type customerContactRepository struct {
}

func (r *customerContactRepository) Get(db *gorm.DB, id int64) *models.CustomerContact {
	ret := &models.CustomerContact{}
	if err := db.First(ret, "id = ?", id).Error; err != nil {
		return nil
	}
	return ret
}

func (r *customerContactRepository) Take(db *gorm.DB, where ...interface{}) *models.CustomerContact {
	ret := &models.CustomerContact{}
	if err := db.Take(ret, where...).Error; err != nil {
		return nil
	}
	return ret
}

func (r *customerContactRepository) Find(db *gorm.DB, cnd *sqls.Cnd) (list []models.CustomerContact) {
	cnd.Find(db, &list)
	return
}

func (r *customerContactRepository) FindOne(db *gorm.DB, cnd *sqls.Cnd) *models.CustomerContact {
	ret := &models.CustomerContact{}
	if err := cnd.FindOne(db, &ret); err != nil {
		return nil
	}
	return ret
}

func (r *customerContactRepository) FindPageByParams(db *gorm.DB, params *params.QueryParams) (list []models.CustomerContact, paging *sqls.Paging) {
	return r.FindPageByCnd(db, &params.Cnd)
}

func (r *customerContactRepository) FindPageByCnd(db *gorm.DB, cnd *sqls.Cnd) (list []models.CustomerContact, paging *sqls.Paging) {
	cnd.Find(db, &list)
	count := cnd.Count(db, &models.CustomerContact{})

	paging = &sqls.Paging{
		Page:  cnd.Paging.Page,
		Limit: cnd.Paging.Limit,
		Total: count,
	}
	return
}

func (r *customerContactRepository) FindBySql(db *gorm.DB, sqlStr string, paramArr... interface{}) (list []models.CustomerContact) {
	db.Raw(sqlStr, paramArr...).Scan(&list)
	return
}

func (r *customerContactRepository) CountBySql(db *gorm.DB, sqlStr string, paramArr... interface{}) (count int64) {
	db.Raw(sqlStr, paramArr...).Count(&count)
	return
}

func (r *customerContactRepository) Count(db *gorm.DB, cnd *sqls.Cnd) int64 {
	return cnd.Count(db, &models.CustomerContact{})
}

func (r *customerContactRepository) Create(db *gorm.DB, t *models.CustomerContact) (err error) {
	err = db.Create(t).Error
	return
}

func (r *customerContactRepository) Update(db *gorm.DB, t *models.CustomerContact) (err error) {
	err = db.Save(t).Error
	return
}

func (r *customerContactRepository) Updates(db *gorm.DB, id int64, columns map[string]interface{}) (err error) {
	err = db.Model(&models.CustomerContact{}).Where("id = ?", id).Updates(columns).Error
	return
}

func (r *customerContactRepository) UpdateColumn(db *gorm.DB, id int64, name string, value interface{}) (err error) {
	err = db.Model(&models.CustomerContact{}).Where("id = ?", id).UpdateColumn(name, value).Error
	return
}

func (r *customerContactRepository) Delete(db *gorm.DB, id int64) {
	db.Delete(&models.CustomerContact{}, "id = ?", id)
}

