package repositories

import (
	"cs-agent/internal/models"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
	"gorm.io/gorm"
)

var CustomerRepository = newCustomerRepository()

func newCustomerRepository() *customerRepository {
	return &customerRepository{}
}

type customerRepository struct {
}

func (r *customerRepository) Get(db *gorm.DB, id int64) *models.Customer {
	ret := &models.Customer{}
	if err := db.First(ret, "id = ?", id).Error; err != nil {
		return nil
	}
	return ret
}

func (r *customerRepository) Take(db *gorm.DB, where ...interface{}) *models.Customer {
	ret := &models.Customer{}
	if err := db.Take(ret, where...).Error; err != nil {
		return nil
	}
	return ret
}

func (r *customerRepository) Find(db *gorm.DB, cnd *sqls.Cnd) (list []models.Customer) {
	cnd.Find(db, &list)
	return
}

func (r *customerRepository) FindOne(db *gorm.DB, cnd *sqls.Cnd) *models.Customer {
	ret := &models.Customer{}
	if err := cnd.FindOne(db, &ret); err != nil {
		return nil
	}
	return ret
}

func (r *customerRepository) FindPageByParams(db *gorm.DB, params *params.QueryParams) (list []models.Customer, paging *sqls.Paging) {
	return r.FindPageByCnd(db, &params.Cnd)
}

func (r *customerRepository) FindPageByCnd(db *gorm.DB, cnd *sqls.Cnd) (list []models.Customer, paging *sqls.Paging) {
	cnd.Find(db, &list)
	count := cnd.Count(db, &models.Customer{})

	paging = &sqls.Paging{
		Page:  cnd.Paging.Page,
		Limit: cnd.Paging.Limit,
		Total: count,
	}
	return
}

func (r *customerRepository) FindBySql(db *gorm.DB, sqlStr string, paramArr... interface{}) (list []models.Customer) {
	db.Raw(sqlStr, paramArr...).Scan(&list)
	return
}

func (r *customerRepository) CountBySql(db *gorm.DB, sqlStr string, paramArr... interface{}) (count int64) {
	db.Raw(sqlStr, paramArr...).Count(&count)
	return
}

func (r *customerRepository) Count(db *gorm.DB, cnd *sqls.Cnd) int64 {
	return cnd.Count(db, &models.Customer{})
}

func (r *customerRepository) Create(db *gorm.DB, t *models.Customer) (err error) {
	err = db.Create(t).Error
	return
}

func (r *customerRepository) Update(db *gorm.DB, t *models.Customer) (err error) {
	err = db.Save(t).Error
	return
}

func (r *customerRepository) Updates(db *gorm.DB, id int64, columns map[string]interface{}) (err error) {
	err = db.Model(&models.Customer{}).Where("id = ?", id).Updates(columns).Error
	return
}

func (r *customerRepository) UpdateColumn(db *gorm.DB, id int64, name string, value interface{}) (err error) {
	err = db.Model(&models.Customer{}).Where("id = ?", id).UpdateColumn(name, value).Error
	return
}

func (r *customerRepository) Delete(db *gorm.DB, id int64) {
	db.Delete(&models.Customer{}, "id = ?", id)
}

