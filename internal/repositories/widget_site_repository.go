package repositories

import (
	"cs-agent/internal/models"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
	"gorm.io/gorm"
)

var WidgetSiteRepository = newWidgetSiteRepository()

func newWidgetSiteRepository() *widgetSiteRepository {
	return &widgetSiteRepository{}
}

type widgetSiteRepository struct {
}

func (r *widgetSiteRepository) Get(db *gorm.DB, id int64) *models.WidgetSite {
	ret := &models.WidgetSite{}
	if err := db.First(ret, "id = ?", id).Error; err != nil {
		return nil
	}
	return ret
}

func (r *widgetSiteRepository) Take(db *gorm.DB, where ...interface{}) *models.WidgetSite {
	ret := &models.WidgetSite{}
	if err := db.Take(ret, where...).Error; err != nil {
		return nil
	}
	return ret
}

func (r *widgetSiteRepository) Find(db *gorm.DB, cnd *sqls.Cnd) (list []models.WidgetSite) {
	cnd.Find(db, &list)
	return
}

func (r *widgetSiteRepository) FindOne(db *gorm.DB, cnd *sqls.Cnd) *models.WidgetSite {
	ret := &models.WidgetSite{}
	if err := cnd.FindOne(db, &ret); err != nil {
		return nil
	}
	return ret
}

func (r *widgetSiteRepository) FindPageByParams(db *gorm.DB, params *params.QueryParams) (list []models.WidgetSite, paging *sqls.Paging) {
	return r.FindPageByCnd(db, &params.Cnd)
}

func (r *widgetSiteRepository) FindPageByCnd(db *gorm.DB, cnd *sqls.Cnd) (list []models.WidgetSite, paging *sqls.Paging) {
	cnd.Find(db, &list)
	count := cnd.Count(db, &models.WidgetSite{})

	paging = &sqls.Paging{
		Page:  cnd.Paging.Page,
		Limit: cnd.Paging.Limit,
		Total: count,
	}
	return
}

func (r *widgetSiteRepository) FindBySql(db *gorm.DB, sqlStr string, paramArr ...interface{}) (list []models.WidgetSite) {
	db.Raw(sqlStr, paramArr...).Scan(&list)
	return
}

func (r *widgetSiteRepository) CountBySql(db *gorm.DB, sqlStr string, paramArr ...interface{}) (count int64) {
	db.Raw(sqlStr, paramArr...).Count(&count)
	return
}

func (r *widgetSiteRepository) Count(db *gorm.DB, cnd *sqls.Cnd) int64 {
	return cnd.Count(db, &models.WidgetSite{})
}

func (r *widgetSiteRepository) Create(db *gorm.DB, t *models.WidgetSite) (err error) {
	err = db.Create(t).Error
	return
}

func (r *widgetSiteRepository) Update(db *gorm.DB, t *models.WidgetSite) (err error) {
	err = db.Save(t).Error
	return
}

func (r *widgetSiteRepository) Updates(db *gorm.DB, id int64, columns map[string]interface{}) (err error) {
	err = db.Model(&models.WidgetSite{}).Where("id = ?", id).Updates(columns).Error
	return
}

func (r *widgetSiteRepository) UpdateColumn(db *gorm.DB, id int64, name string, value interface{}) (err error) {
	err = db.Model(&models.WidgetSite{}).Where("id = ?", id).UpdateColumn(name, value).Error
	return
}

func (r *widgetSiteRepository) Delete(db *gorm.DB, id int64) {
	db.Delete(&models.WidgetSite{}, "id = ?", id)
}
