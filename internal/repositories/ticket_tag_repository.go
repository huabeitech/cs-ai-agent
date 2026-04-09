package repositories

import (
	"cs-agent/internal/models"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
	"gorm.io/gorm"
)

var TicketTagRepository = newTicketTagRepository()

func newTicketTagRepository() *ticketTagRepository {
	return &ticketTagRepository{}
}

type ticketTagRepository struct{}

func (r *ticketTagRepository) Get(db *gorm.DB, id int64) *models.TicketTag {
	ret := &models.TicketTag{}
	if err := db.First(ret, "id = ?", id).Error; err != nil {
		return nil
	}
	return ret
}

func (r *ticketTagRepository) Take(db *gorm.DB, where ...interface{}) *models.TicketTag {
	ret := &models.TicketTag{}
	if err := db.Take(ret, where...).Error; err != nil {
		return nil
	}
	return ret
}

func (r *ticketTagRepository) Find(db *gorm.DB, cnd *sqls.Cnd) (list []models.TicketTag) {
	cnd.Find(db, &list)
	return
}

func (r *ticketTagRepository) FindOne(db *gorm.DB, cnd *sqls.Cnd) *models.TicketTag {
	ret := &models.TicketTag{}
	if err := cnd.FindOne(db, &ret); err != nil {
		return nil
	}
	return ret
}

func (r *ticketTagRepository) FindPageByParams(db *gorm.DB, params *params.QueryParams) (list []models.TicketTag, paging *sqls.Paging) {
	return r.FindPageByCnd(db, &params.Cnd)
}

func (r *ticketTagRepository) FindPageByCnd(db *gorm.DB, cnd *sqls.Cnd) (list []models.TicketTag, paging *sqls.Paging) {
	cnd.Find(db, &list)
	count := cnd.Count(db, &models.TicketTag{})
	paging = &sqls.Paging{
		Page:  cnd.Paging.Page,
		Limit: cnd.Paging.Limit,
		Total: count,
	}
	return
}

func (r *ticketTagRepository) Count(db *gorm.DB, cnd *sqls.Cnd) int64 {
	return cnd.Count(db, &models.TicketTag{})
}

func (r *ticketTagRepository) Create(db *gorm.DB, t *models.TicketTag) error {
	return db.Create(t).Error
}

func (r *ticketTagRepository) DeleteByTicketID(db *gorm.DB, ticketID int64) error {
	return db.Where("ticket_id = ?", ticketID).Delete(&models.TicketTag{}).Error
}
