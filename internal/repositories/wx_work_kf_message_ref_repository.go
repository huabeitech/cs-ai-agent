package repositories

import (
	"cs-agent/internal/models"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
	"gorm.io/gorm"
)

var WxWorkKFMessageRefRepository = newWxWorkKFMessageRefRepository()

func newWxWorkKFMessageRefRepository() *wxWorkKFMessageRefRepository {
	return &wxWorkKFMessageRefRepository{}
}

type wxWorkKFMessageRefRepository struct {
}

func (r *wxWorkKFMessageRefRepository) Get(db *gorm.DB, id int64) *models.WxWorkKFMessageRef {
	ret := &models.WxWorkKFMessageRef{}
	if err := db.First(ret, "id = ?", id).Error; err != nil {
		return nil
	}
	return ret
}

func (r *wxWorkKFMessageRefRepository) Take(db *gorm.DB, where ...interface{}) *models.WxWorkKFMessageRef {
	ret := &models.WxWorkKFMessageRef{}
	if err := db.Take(ret, where...).Error; err != nil {
		return nil
	}
	return ret
}

func (r *wxWorkKFMessageRefRepository) Find(db *gorm.DB, cnd *sqls.Cnd) (list []models.WxWorkKFMessageRef) {
	cnd.Find(db, &list)
	return
}

func (r *wxWorkKFMessageRefRepository) FindOne(db *gorm.DB, cnd *sqls.Cnd) *models.WxWorkKFMessageRef {
	ret := &models.WxWorkKFMessageRef{}
	if err := cnd.FindOne(db, &ret); err != nil {
		return nil
	}
	return ret
}

func (r *wxWorkKFMessageRefRepository) FindPageByParams(db *gorm.DB, params *params.QueryParams) (list []models.WxWorkKFMessageRef, paging *sqls.Paging) {
	return r.FindPageByCnd(db, &params.Cnd)
}

func (r *wxWorkKFMessageRefRepository) FindPageByCnd(db *gorm.DB, cnd *sqls.Cnd) (list []models.WxWorkKFMessageRef, paging *sqls.Paging) {
	cnd.Find(db, &list)
	count := cnd.Count(db, &models.WxWorkKFMessageRef{})

	paging = &sqls.Paging{
		Page:  cnd.Paging.Page,
		Limit: cnd.Paging.Limit,
		Total: count,
	}
	return
}

func (r *wxWorkKFMessageRefRepository) FindBySql(db *gorm.DB, sqlStr string, paramArr... interface{}) (list []models.WxWorkKFMessageRef) {
	db.Raw(sqlStr, paramArr...).Scan(&list)
	return
}

func (r *wxWorkKFMessageRefRepository) CountBySql(db *gorm.DB, sqlStr string, paramArr... interface{}) (count int64) {
	db.Raw(sqlStr, paramArr...).Count(&count)
	return
}

func (r *wxWorkKFMessageRefRepository) Count(db *gorm.DB, cnd *sqls.Cnd) int64 {
	return cnd.Count(db, &models.WxWorkKFMessageRef{})
}

func (r *wxWorkKFMessageRefRepository) Create(db *gorm.DB, t *models.WxWorkKFMessageRef) (err error) {
	err = db.Create(t).Error
	return
}

func (r *wxWorkKFMessageRefRepository) Update(db *gorm.DB, t *models.WxWorkKFMessageRef) (err error) {
	err = db.Save(t).Error
	return
}

func (r *wxWorkKFMessageRefRepository) Updates(db *gorm.DB, id int64, columns map[string]interface{}) (err error) {
	err = db.Model(&models.WxWorkKFMessageRef{}).Where("id = ?", id).Updates(columns).Error
	return
}

func (r *wxWorkKFMessageRefRepository) UpdateColumn(db *gorm.DB, id int64, name string, value interface{}) (err error) {
	err = db.Model(&models.WxWorkKFMessageRef{}).Where("id = ?", id).UpdateColumn(name, value).Error
	return
}

func (r *wxWorkKFMessageRefRepository) Delete(db *gorm.DB, id int64) {
	db.Delete(&models.WxWorkKFMessageRef{}, "id = ?", id)
}

