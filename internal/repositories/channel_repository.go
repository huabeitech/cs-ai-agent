package repositories

import (
	"cs-agent/internal/models"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
	"gorm.io/gorm"
)

var ChannelRepository = newChannelRepository()

func newChannelRepository() *channelRepository {
	return &channelRepository{}
}

type channelRepository struct {
}

func (r *channelRepository) Get(db *gorm.DB, id int64) *models.Channel {
	ret := &models.Channel{}
	if err := db.First(ret, "id = ?", id).Error; err != nil {
		return nil
	}
	return ret
}

func (r *channelRepository) Take(db *gorm.DB, where ...interface{}) *models.Channel {
	ret := &models.Channel{}
	if err := db.Take(ret, where...).Error; err != nil {
		return nil
	}
	return ret
}

func (r *channelRepository) Find(db *gorm.DB, cnd *sqls.Cnd) (list []models.Channel) {
	cnd.Find(db, &list)
	return
}

func (r *channelRepository) FindOne(db *gorm.DB, cnd *sqls.Cnd) *models.Channel {
	ret := &models.Channel{}
	if err := cnd.FindOne(db, &ret); err != nil {
		return nil
	}
	return ret
}

func (r *channelRepository) FindPageByParams(db *gorm.DB, params *params.QueryParams) (list []models.Channel, paging *sqls.Paging) {
	return r.FindPageByCnd(db, &params.Cnd)
}

func (r *channelRepository) FindPageByCnd(db *gorm.DB, cnd *sqls.Cnd) (list []models.Channel, paging *sqls.Paging) {
	cnd.Find(db, &list)
	count := cnd.Count(db, &models.Channel{})

	paging = &sqls.Paging{
		Page:  cnd.Paging.Page,
		Limit: cnd.Paging.Limit,
		Total: count,
	}
	return
}

func (r *channelRepository) FindBySql(db *gorm.DB, sqlStr string, paramArr... interface{}) (list []models.Channel) {
	db.Raw(sqlStr, paramArr...).Scan(&list)
	return
}

func (r *channelRepository) CountBySql(db *gorm.DB, sqlStr string, paramArr... interface{}) (count int64) {
	db.Raw(sqlStr, paramArr...).Count(&count)
	return
}

func (r *channelRepository) Count(db *gorm.DB, cnd *sqls.Cnd) int64 {
	return cnd.Count(db, &models.Channel{})
}

func (r *channelRepository) Create(db *gorm.DB, t *models.Channel) (err error) {
	err = db.Create(t).Error
	return
}

func (r *channelRepository) Update(db *gorm.DB, t *models.Channel) (err error) {
	err = db.Save(t).Error
	return
}

func (r *channelRepository) Updates(db *gorm.DB, id int64, columns map[string]interface{}) (err error) {
	err = db.Model(&models.Channel{}).Where("id = ?", id).Updates(columns).Error
	return
}

func (r *channelRepository) UpdateColumn(db *gorm.DB, id int64, name string, value interface{}) (err error) {
	err = db.Model(&models.Channel{}).Where("id = ?", id).UpdateColumn(name, value).Error
	return
}

func (r *channelRepository) Delete(db *gorm.DB, id int64) {
	db.Delete(&models.Channel{}, "id = ?", id)
}

