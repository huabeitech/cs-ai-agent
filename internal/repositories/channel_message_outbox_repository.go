package repositories

import (
	"cs-agent/internal/models"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
	"gorm.io/gorm"
)

var ChannelMessageOutboxRepository = newChannelMessageOutboxRepository()

func newChannelMessageOutboxRepository() *channelMessageOutboxRepository {
	return &channelMessageOutboxRepository{}
}

type channelMessageOutboxRepository struct {
}

func (r *channelMessageOutboxRepository) Get(db *gorm.DB, id int64) *models.ChannelMessageOutbox {
	ret := &models.ChannelMessageOutbox{}
	if err := db.First(ret, "id = ?", id).Error; err != nil {
		return nil
	}
	return ret
}

func (r *channelMessageOutboxRepository) Take(db *gorm.DB, where ...interface{}) *models.ChannelMessageOutbox {
	ret := &models.ChannelMessageOutbox{}
	if err := db.Take(ret, where...).Error; err != nil {
		return nil
	}
	return ret
}

func (r *channelMessageOutboxRepository) Find(db *gorm.DB, cnd *sqls.Cnd) (list []models.ChannelMessageOutbox) {
	cnd.Find(db, &list)
	return
}

func (r *channelMessageOutboxRepository) FindOne(db *gorm.DB, cnd *sqls.Cnd) *models.ChannelMessageOutbox {
	ret := &models.ChannelMessageOutbox{}
	if err := cnd.FindOne(db, &ret); err != nil {
		return nil
	}
	return ret
}

func (r *channelMessageOutboxRepository) FindPageByParams(db *gorm.DB, params *params.QueryParams) (list []models.ChannelMessageOutbox, paging *sqls.Paging) {
	return r.FindPageByCnd(db, &params.Cnd)
}

func (r *channelMessageOutboxRepository) FindPageByCnd(db *gorm.DB, cnd *sqls.Cnd) (list []models.ChannelMessageOutbox, paging *sqls.Paging) {
	cnd.Find(db, &list)
	count := cnd.Count(db, &models.ChannelMessageOutbox{})

	paging = &sqls.Paging{
		Page:  cnd.Paging.Page,
		Limit: cnd.Paging.Limit,
		Total: count,
	}
	return
}

func (r *channelMessageOutboxRepository) FindBySql(db *gorm.DB, sqlStr string, paramArr... interface{}) (list []models.ChannelMessageOutbox) {
	db.Raw(sqlStr, paramArr...).Scan(&list)
	return
}

func (r *channelMessageOutboxRepository) CountBySql(db *gorm.DB, sqlStr string, paramArr... interface{}) (count int64) {
	db.Raw(sqlStr, paramArr...).Count(&count)
	return
}

func (r *channelMessageOutboxRepository) Count(db *gorm.DB, cnd *sqls.Cnd) int64 {
	return cnd.Count(db, &models.ChannelMessageOutbox{})
}

func (r *channelMessageOutboxRepository) Create(db *gorm.DB, t *models.ChannelMessageOutbox) (err error) {
	err = db.Create(t).Error
	return
}

func (r *channelMessageOutboxRepository) Update(db *gorm.DB, t *models.ChannelMessageOutbox) (err error) {
	err = db.Save(t).Error
	return
}

func (r *channelMessageOutboxRepository) Updates(db *gorm.DB, id int64, columns map[string]interface{}) (err error) {
	err = db.Model(&models.ChannelMessageOutbox{}).Where("id = ?", id).Updates(columns).Error
	return
}

func (r *channelMessageOutboxRepository) UpdateColumn(db *gorm.DB, id int64, name string, value interface{}) (err error) {
	err = db.Model(&models.ChannelMessageOutbox{}).Where("id = ?", id).UpdateColumn(name, value).Error
	return
}

func (r *channelMessageOutboxRepository) Delete(db *gorm.DB, id int64) {
	db.Delete(&models.ChannelMessageOutbox{}, "id = ?", id)
}

