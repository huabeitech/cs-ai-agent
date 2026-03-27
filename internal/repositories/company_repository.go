package repositories

import (
	"cs-agent/internal/models"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
	"gorm.io/gorm"
)

var CompanyRepository = newCompanyRepository()

func newCompanyRepository() *companyRepository {
	return &companyRepository{}
}

type companyRepository struct {
}

func (r *companyRepository) GetByName(db *gorm.DB, name string) *models.Company {
	ret := &models.Company{}
	if err := db.First(ret, "name = ?", name).Error; err != nil {
		return nil
	}
	return ret
}

func (r *companyRepository) Get(db *gorm.DB, id int64) *models.Company {
	ret := &models.Company{}
	if err := db.First(ret, "id = ?", id).Error; err != nil {
		return nil
	}
	return ret
}

func (r *companyRepository) Take(db *gorm.DB, where ...interface{}) *models.Company {
	ret := &models.Company{}
	if err := db.Take(ret, where...).Error; err != nil {
		return nil
	}
	return ret
}

func (r *companyRepository) Find(db *gorm.DB, cnd *sqls.Cnd) (list []models.Company) {
	cnd.Find(db, &list)
	return
}

func (r *companyRepository) FindOne(db *gorm.DB, cnd *sqls.Cnd) *models.Company {
	ret := &models.Company{}
	if err := cnd.FindOne(db, &ret); err != nil {
		return nil
	}
	return ret
}

func (r *companyRepository) FindPageByParams(db *gorm.DB, params *params.QueryParams) (list []models.Company, paging *sqls.Paging) {
	return r.FindPageByCnd(db, &params.Cnd)
}

func (r *companyRepository) FindPageByCnd(db *gorm.DB, cnd *sqls.Cnd) (list []models.Company, paging *sqls.Paging) {
	cnd.Find(db, &list)
	count := cnd.Count(db, &models.Company{})

	paging = &sqls.Paging{
		Page:  cnd.Paging.Page,
		Limit: cnd.Paging.Limit,
		Total: count,
	}
	return
}

func (r *companyRepository) FindBySql(db *gorm.DB, sqlStr string, paramArr ...interface{}) (list []models.Company) {
	db.Raw(sqlStr, paramArr...).Scan(&list)
	return
}

func (r *companyRepository) CountBySql(db *gorm.DB, sqlStr string, paramArr ...interface{}) (count int64) {
	db.Raw(sqlStr, paramArr...).Count(&count)
	return
}

func (r *companyRepository) Count(db *gorm.DB, cnd *sqls.Cnd) int64 {
	return cnd.Count(db, &models.Company{})
}

func (r *companyRepository) Create(db *gorm.DB, t *models.Company) (err error) {
	err = db.Create(t).Error
	return
}

func (r *companyRepository) Update(db *gorm.DB, t *models.Company) (err error) {
	err = db.Save(t).Error
	return
}

func (r *companyRepository) Updates(db *gorm.DB, id int64, columns map[string]interface{}) (err error) {
	err = db.Model(&models.Company{}).Where("id = ?", id).Updates(columns).Error
	return
}

func (r *companyRepository) UpdateColumn(db *gorm.DB, id int64, name string, value interface{}) (err error) {
	err = db.Model(&models.Company{}).Where("id = ?", id).UpdateColumn(name, value).Error
	return
}

func (r *companyRepository) Delete(db *gorm.DB, id int64) {
	db.Delete(&models.Company{}, "id = ?", id)
}
