package repositories

import (
	"log/slog"

	"cs-agent/internal/models"
	"cs-agent/internal/pkg/enums"

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

type CompanyCustomerCount struct {
	CompanyID int64 `gorm:"column:company_id"`
	Count     int64 `gorm:"column:cnt"`
}

func (r *customerRepository) CountByCompanyIDs(db *gorm.DB, companyIDs []int64, excludeStatus int) map[int64]int64 {
	ret := make(map[int64]int64)
	if len(companyIDs) == 0 {
		return ret
	}

	rows := make([]CompanyCustomerCount, 0, len(companyIDs))
	db.Model(&models.Customer{}).
		Select("company_id, count(1) as cnt").
		Where("company_id in ?", companyIDs).
		Where("status <> ?", excludeStatus).
		Group("company_id").
		Scan(&rows)

	for _, row := range rows {
		ret[row.CompanyID] = row.Count
	}
	return ret
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

// FindPageByCndForCustomerList 客户列表：主表 t_customer 左连 t_customer_contact，便于按任意联系方式检索（与 auth_service 连表风格一致，表名写死 t_ 前缀）。
func (r *customerRepository) FindPageByCndForCustomerList(db *gorm.DB, cnd *sqls.Cnd) (list []models.Customer, paging *sqls.Paging) {
	deleted := int(enums.StatusDeleted)
	join := "LEFT JOIN t_customer_contact AS cc ON cc.customer_id = c.id AND cc.status <> ?"

	listBase := db.Table("t_customer AS c").Select("DISTINCT c.*").Joins(join, deleted)
	if err := cnd.Build(listBase).Find(&list).Error; err != nil {
		slog.Error(err.Error(), slog.Any("error", err))
	}

	countBase := db.Table("t_customer AS c").Joins(join, deleted)
	countQ := cnd.Build(countBase)
	var count int64
	if err := countQ.Distinct("c.id").Count(&count).Error; err != nil {
		slog.Error(err.Error(), slog.Any("error", err))
	}

	paging = &sqls.Paging{
		Page:  cnd.Paging.Page,
		Limit: cnd.Paging.Limit,
		Total: count,
	}
	return
}

func (r *customerRepository) FindBySql(db *gorm.DB, sqlStr string, paramArr ...interface{}) (list []models.Customer) {
	db.Raw(sqlStr, paramArr...).Scan(&list)
	return
}

func (r *customerRepository) CountBySql(db *gorm.DB, sqlStr string, paramArr ...interface{}) (count int64) {
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
