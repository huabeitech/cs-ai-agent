package repositories

import (
	"cs-agent/internal/models"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
	"gorm.io/gorm"
)

var QuickReplyRepository = newQuickReplyRepository()

func newQuickReplyRepository() *quickReplyRepository {
	return &quickReplyRepository{}
}

type quickReplyRepository struct {
}

func (r *quickReplyRepository) Get(db *gorm.DB, id int64) *models.QuickReply {
	ret := &models.QuickReply{}
	if err := db.First(ret, "id = ?", id).Error; err != nil {
		return nil
	}
	return ret
}

func (r *quickReplyRepository) Take(db *gorm.DB, where ...interface{}) *models.QuickReply {
	ret := &models.QuickReply{}
	if err := db.Take(ret, where...).Error; err != nil {
		return nil
	}
	return ret
}

func (r *quickReplyRepository) Find(db *gorm.DB, cnd *sqls.Cnd) (list []models.QuickReply) {
	cnd.Find(db, &list)
	return
}

func (r *quickReplyRepository) FindOne(db *gorm.DB, cnd *sqls.Cnd) *models.QuickReply {
	ret := &models.QuickReply{}
	if err := cnd.FindOne(db, &ret); err != nil {
		return nil
	}
	return ret
}

func (r *quickReplyRepository) FindPageByParams(db *gorm.DB, params *params.QueryParams) (list []models.QuickReply, paging *sqls.Paging) {
	return r.FindPageByCnd(db, &params.Cnd)
}

func (r *quickReplyRepository) FindPageByCnd(db *gorm.DB, cnd *sqls.Cnd) (list []models.QuickReply, paging *sqls.Paging) {
	cnd.Find(db, &list)
	count := cnd.Count(db, &models.QuickReply{})

	paging = &sqls.Paging{
		Page:  cnd.Paging.Page,
		Limit: cnd.Paging.Limit,
		Total: count,
	}
	return
}

func (r *quickReplyRepository) FindBySql(db *gorm.DB, sqlStr string, paramArr ...interface{}) (list []models.QuickReply) {
	db.Raw(sqlStr, paramArr...).Scan(&list)
	return
}

func (r *quickReplyRepository) CountBySql(db *gorm.DB, sqlStr string, paramArr ...interface{}) (count int64) {
	db.Raw(sqlStr, paramArr...).Count(&count)
	return
}

func (r *quickReplyRepository) Count(db *gorm.DB, cnd *sqls.Cnd) int64 {
	return cnd.Count(db, &models.QuickReply{})
}

func (r *quickReplyRepository) Create(db *gorm.DB, t *models.QuickReply) (err error) {
	err = db.Create(t).Error
	return
}

func (r *quickReplyRepository) Update(db *gorm.DB, t *models.QuickReply) (err error) {
	err = db.Save(t).Error
	return
}

func (r *quickReplyRepository) Updates(db *gorm.DB, id int64, columns map[string]interface{}) (err error) {
	err = db.Model(&models.QuickReply{}).Where("id = ?", id).Updates(columns).Error
	return
}

func (r *quickReplyRepository) UpdateColumn(db *gorm.DB, id int64, name string, value interface{}) (err error) {
	err = db.Model(&models.QuickReply{}).Where("id = ?", id).UpdateColumn(name, value).Error
	return
}

func (r *quickReplyRepository) Delete(db *gorm.DB, id int64) {
	db.Delete(&models.QuickReply{}, "id = ?", id)
}
