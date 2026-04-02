package repositories

import (
	"cs-agent/internal/models"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
	"gorm.io/gorm"
)

var TicketWatcherRepository = newTicketWatcherRepository()

func newTicketWatcherRepository() *ticketWatcherRepository {
	return &ticketWatcherRepository{}
}

type ticketWatcherRepository struct {
}

func (r *ticketWatcherRepository) Get(db *gorm.DB, id int64) *models.TicketWatcher {
	ret := &models.TicketWatcher{}
	if err := db.First(ret, "id = ?", id).Error; err != nil {
		return nil
	}
	return ret
}

func (r *ticketWatcherRepository) Take(db *gorm.DB, where ...interface{}) *models.TicketWatcher {
	ret := &models.TicketWatcher{}
	if err := db.Take(ret, where...).Error; err != nil {
		return nil
	}
	return ret
}

func (r *ticketWatcherRepository) Find(db *gorm.DB, cnd *sqls.Cnd) (list []models.TicketWatcher) {
	cnd.Find(db, &list)
	return
}

func (r *ticketWatcherRepository) FindOne(db *gorm.DB, cnd *sqls.Cnd) *models.TicketWatcher {
	ret := &models.TicketWatcher{}
	if err := cnd.FindOne(db, &ret); err != nil {
		return nil
	}
	return ret
}

func (r *ticketWatcherRepository) FindPageByParams(db *gorm.DB, params *params.QueryParams) (list []models.TicketWatcher, paging *sqls.Paging) {
	return r.FindPageByCnd(db, &params.Cnd)
}

func (r *ticketWatcherRepository) FindPageByCnd(db *gorm.DB, cnd *sqls.Cnd) (list []models.TicketWatcher, paging *sqls.Paging) {
	cnd.Find(db, &list)
	count := cnd.Count(db, &models.TicketWatcher{})

	paging = &sqls.Paging{
		Page:  cnd.Paging.Page,
		Limit: cnd.Paging.Limit,
		Total: count,
	}
	return
}

func (r *ticketWatcherRepository) FindBySql(db *gorm.DB, sqlStr string, paramArr... interface{}) (list []models.TicketWatcher) {
	db.Raw(sqlStr, paramArr...).Scan(&list)
	return
}

func (r *ticketWatcherRepository) CountBySql(db *gorm.DB, sqlStr string, paramArr... interface{}) (count int64) {
	db.Raw(sqlStr, paramArr...).Count(&count)
	return
}

func (r *ticketWatcherRepository) Count(db *gorm.DB, cnd *sqls.Cnd) int64 {
	return cnd.Count(db, &models.TicketWatcher{})
}

func (r *ticketWatcherRepository) Create(db *gorm.DB, t *models.TicketWatcher) (err error) {
	err = db.Create(t).Error
	return
}

func (r *ticketWatcherRepository) Update(db *gorm.DB, t *models.TicketWatcher) (err error) {
	err = db.Save(t).Error
	return
}

func (r *ticketWatcherRepository) Updates(db *gorm.DB, id int64, columns map[string]interface{}) (err error) {
	err = db.Model(&models.TicketWatcher{}).Where("id = ?", id).Updates(columns).Error
	return
}

func (r *ticketWatcherRepository) UpdateColumn(db *gorm.DB, id int64, name string, value interface{}) (err error) {
	err = db.Model(&models.TicketWatcher{}).Where("id = ?", id).UpdateColumn(name, value).Error
	return
}

func (r *ticketWatcherRepository) Delete(db *gorm.DB, id int64) {
	db.Delete(&models.TicketWatcher{}, "id = ?", id)
}

