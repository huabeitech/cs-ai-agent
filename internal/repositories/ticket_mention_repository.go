package repositories

import (
	"cs-agent/internal/models"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
	"gorm.io/gorm"
)

var TicketMentionRepository = newTicketMentionRepository()

func newTicketMentionRepository() *ticketMentionRepository {
	return &ticketMentionRepository{}
}

type ticketMentionRepository struct {
}

func (r *ticketMentionRepository) Get(db *gorm.DB, id int64) *models.TicketMention {
	ret := &models.TicketMention{}
	if err := db.First(ret, "id = ?", id).Error; err != nil {
		return nil
	}
	return ret
}

func (r *ticketMentionRepository) Take(db *gorm.DB, where ...interface{}) *models.TicketMention {
	ret := &models.TicketMention{}
	if err := db.Take(ret, where...).Error; err != nil {
		return nil
	}
	return ret
}

func (r *ticketMentionRepository) Find(db *gorm.DB, cnd *sqls.Cnd) (list []models.TicketMention) {
	cnd.Find(db, &list)
	return
}

func (r *ticketMentionRepository) FindOne(db *gorm.DB, cnd *sqls.Cnd) *models.TicketMention {
	ret := &models.TicketMention{}
	if err := cnd.FindOne(db, &ret); err != nil {
		return nil
	}
	return ret
}

func (r *ticketMentionRepository) FindPageByParams(db *gorm.DB, params *params.QueryParams) (list []models.TicketMention, paging *sqls.Paging) {
	return r.FindPageByCnd(db, &params.Cnd)
}

func (r *ticketMentionRepository) FindPageByCnd(db *gorm.DB, cnd *sqls.Cnd) (list []models.TicketMention, paging *sqls.Paging) {
	cnd.Find(db, &list)
	count := cnd.Count(db, &models.TicketMention{})
	paging = &sqls.Paging{Page: cnd.Paging.Page, Limit: cnd.Paging.Limit, Total: count}
	return
}

func (r *ticketMentionRepository) Count(db *gorm.DB, cnd *sqls.Cnd) int64 {
	return cnd.Count(db, &models.TicketMention{})
}

func (r *ticketMentionRepository) Create(db *gorm.DB, t *models.TicketMention) error {
	return db.Create(t).Error
}
