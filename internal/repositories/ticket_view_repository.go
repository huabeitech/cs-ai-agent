package repositories

import (
	"cs-agent/internal/models"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
	"gorm.io/gorm"
)

var TicketViewRepository = newTicketViewRepository()

func newTicketViewRepository() *ticketViewRepository {
	return &ticketViewRepository{}
}

type ticketViewRepository struct {
}

func (r *ticketViewRepository) Get(db *gorm.DB, id int64) *models.TicketView {
	ret := &models.TicketView{}
	if err := db.First(ret, "id = ?", id).Error; err != nil {
		return nil
	}
	return ret
}

func (r *ticketViewRepository) Take(db *gorm.DB, where ...interface{}) *models.TicketView {
	ret := &models.TicketView{}
	if err := db.Take(ret, where...).Error; err != nil {
		return nil
	}
	return ret
}

func (r *ticketViewRepository) Find(db *gorm.DB, cnd *sqls.Cnd) (list []models.TicketView) {
	cnd.Find(db, &list)
	return
}

func (r *ticketViewRepository) FindOne(db *gorm.DB, cnd *sqls.Cnd) *models.TicketView {
	ret := &models.TicketView{}
	if err := cnd.FindOne(db, &ret); err != nil {
		return nil
	}
	return ret
}

func (r *ticketViewRepository) FindPageByParams(db *gorm.DB, params *params.QueryParams) (list []models.TicketView, paging *sqls.Paging) {
	return r.FindPageByCnd(db, &params.Cnd)
}

func (r *ticketViewRepository) FindPageByCnd(db *gorm.DB, cnd *sqls.Cnd) (list []models.TicketView, paging *sqls.Paging) {
	cnd.Find(db, &list)
	count := cnd.Count(db, &models.TicketView{})

	paging = &sqls.Paging{
		Page:  cnd.Paging.Page,
		Limit: cnd.Paging.Limit,
		Total: count,
	}
	return
}

func (r *ticketViewRepository) Count(db *gorm.DB, cnd *sqls.Cnd) int64 {
	return cnd.Count(db, &models.TicketView{})
}

func (r *ticketViewRepository) Create(db *gorm.DB, t *models.TicketView) error {
	return db.Create(t).Error
}

func (r *ticketViewRepository) Update(db *gorm.DB, t *models.TicketView) error {
	return db.Save(t).Error
}

func (r *ticketViewRepository) Updates(db *gorm.DB, id int64, columns map[string]interface{}) error {
	return db.Model(&models.TicketView{}).Where("id = ?", id).Updates(columns).Error
}

func (r *ticketViewRepository) Delete(db *gorm.DB, id int64) {
	db.Delete(&models.TicketView{}, "id = ?", id)
}
