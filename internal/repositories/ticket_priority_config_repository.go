package repositories

import (
	"cs-agent/internal/models"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
	"gorm.io/gorm"
)

var TicketPriorityConfigRepository = newTicketPriorityConfigRepository()

func newTicketPriorityConfigRepository() *ticketPriorityConfigRepository {
	return &ticketPriorityConfigRepository{}
}

type ticketPriorityConfigRepository struct{}

func (r *ticketPriorityConfigRepository) Get(db *gorm.DB, id int64) *models.TicketPriorityConfig {
	ret := &models.TicketPriorityConfig{}
	if err := db.First(ret, id).Error; err != nil {
		return nil
	}
	return ret
}

func (r *ticketPriorityConfigRepository) Take(db *gorm.DB, where ...interface{}) *models.TicketPriorityConfig {
	ret := &models.TicketPriorityConfig{}
	if err := db.Take(ret, where...).Error; err != nil {
		return nil
	}
	return ret
}

func (r *ticketPriorityConfigRepository) Find(db *gorm.DB, cnd *sqls.Cnd) (list []models.TicketPriorityConfig) {
	cnd.Find(db, &list)
	return
}

func (r *ticketPriorityConfigRepository) FindOne(db *gorm.DB, cnd *sqls.Cnd) *models.TicketPriorityConfig {
	ret := &models.TicketPriorityConfig{}
	if err := cnd.FindOne(db, ret); err != nil {
		return nil
	}
	return ret
}

func (r *ticketPriorityConfigRepository) FindPageByParams(db *gorm.DB, queryParams *params.QueryParams) (list []models.TicketPriorityConfig, paging *sqls.Paging) {
	return r.FindPageByCnd(db, &queryParams.Cnd)
}

func (r *ticketPriorityConfigRepository) FindPageByCnd(db *gorm.DB, cnd *sqls.Cnd) (list []models.TicketPriorityConfig, paging *sqls.Paging) {
	cnd.Find(db, &list)
	count := cnd.Count(db, &models.TicketPriorityConfig{})
	paging = &sqls.Paging{Page: cnd.Paging.Page, Limit: cnd.Paging.Limit, Total: count}
	return
}

func (r *ticketPriorityConfigRepository) Count(db *gorm.DB, cnd *sqls.Cnd) int64 {
	return cnd.Count(db, &models.TicketPriorityConfig{})
}

func (r *ticketPriorityConfigRepository) Create(db *gorm.DB, t *models.TicketPriorityConfig) error {
	return db.Create(t).Error
}

func (r *ticketPriorityConfigRepository) Update(db *gorm.DB, t *models.TicketPriorityConfig) error {
	return db.Save(t).Error
}

func (r *ticketPriorityConfigRepository) Updates(db *gorm.DB, id int64, columns map[string]any) error {
	return db.Model(&models.TicketPriorityConfig{}).Where("id = ?", id).Updates(columns).Error
}

func (r *ticketPriorityConfigRepository) Delete(db *gorm.DB, id int64) error {
	return db.Delete(&models.TicketPriorityConfig{}, "id = ?", id).Error
}
