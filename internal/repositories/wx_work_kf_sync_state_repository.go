package repositories

import (
	"cs-agent/internal/models"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
	"gorm.io/gorm"
)

var WxWorkKFSyncStateRepository = newWxWorkKFSyncStateRepository()

func newWxWorkKFSyncStateRepository() *wxWorkKFSyncStateRepository {
	return &wxWorkKFSyncStateRepository{}
}

type wxWorkKFSyncStateRepository struct {
}

func (r *wxWorkKFSyncStateRepository) Get(db *gorm.DB, id int64) *models.WxWorkKFSyncState {
	ret := &models.WxWorkKFSyncState{}
	if err := db.First(ret, "id = ?", id).Error; err != nil {
		return nil
	}
	return ret
}

func (r *wxWorkKFSyncStateRepository) Take(db *gorm.DB, where ...interface{}) *models.WxWorkKFSyncState {
	ret := &models.WxWorkKFSyncState{}
	if err := db.Take(ret, where...).Error; err != nil {
		return nil
	}
	return ret
}

func (r *wxWorkKFSyncStateRepository) Find(db *gorm.DB, cnd *sqls.Cnd) (list []models.WxWorkKFSyncState) {
	cnd.Find(db, &list)
	return
}

func (r *wxWorkKFSyncStateRepository) FindOne(db *gorm.DB, cnd *sqls.Cnd) *models.WxWorkKFSyncState {
	ret := &models.WxWorkKFSyncState{}
	if err := cnd.FindOne(db, &ret); err != nil {
		return nil
	}
	return ret
}

func (r *wxWorkKFSyncStateRepository) FindPageByParams(db *gorm.DB, params *params.QueryParams) (list []models.WxWorkKFSyncState, paging *sqls.Paging) {
	return r.FindPageByCnd(db, &params.Cnd)
}

func (r *wxWorkKFSyncStateRepository) FindPageByCnd(db *gorm.DB, cnd *sqls.Cnd) (list []models.WxWorkKFSyncState, paging *sqls.Paging) {
	cnd.Find(db, &list)
	count := cnd.Count(db, &models.WxWorkKFSyncState{})

	paging = &sqls.Paging{
		Page:  cnd.Paging.Page,
		Limit: cnd.Paging.Limit,
		Total: count,
	}
	return
}

func (r *wxWorkKFSyncStateRepository) FindBySql(db *gorm.DB, sqlStr string, paramArr... interface{}) (list []models.WxWorkKFSyncState) {
	db.Raw(sqlStr, paramArr...).Scan(&list)
	return
}

func (r *wxWorkKFSyncStateRepository) CountBySql(db *gorm.DB, sqlStr string, paramArr... interface{}) (count int64) {
	db.Raw(sqlStr, paramArr...).Count(&count)
	return
}

func (r *wxWorkKFSyncStateRepository) Count(db *gorm.DB, cnd *sqls.Cnd) int64 {
	return cnd.Count(db, &models.WxWorkKFSyncState{})
}

func (r *wxWorkKFSyncStateRepository) Create(db *gorm.DB, t *models.WxWorkKFSyncState) (err error) {
	err = db.Create(t).Error
	return
}

func (r *wxWorkKFSyncStateRepository) Update(db *gorm.DB, t *models.WxWorkKFSyncState) (err error) {
	err = db.Save(t).Error
	return
}

func (r *wxWorkKFSyncStateRepository) Updates(db *gorm.DB, id int64, columns map[string]interface{}) (err error) {
	err = db.Model(&models.WxWorkKFSyncState{}).Where("id = ?", id).Updates(columns).Error
	return
}

func (r *wxWorkKFSyncStateRepository) UpdateColumn(db *gorm.DB, id int64, name string, value interface{}) (err error) {
	err = db.Model(&models.WxWorkKFSyncState{}).Where("id = ?", id).UpdateColumn(name, value).Error
	return
}

func (r *wxWorkKFSyncStateRepository) Delete(db *gorm.DB, id int64) {
	db.Delete(&models.WxWorkKFSyncState{}, "id = ?", id)
}

