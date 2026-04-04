package repositories

import (
	"cs-agent/internal/models"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
	"gorm.io/gorm"
)

var TicketResolutionCodeRepository = newTicketResolutionCodeRepository()

func newTicketResolutionCodeRepository() *ticketResolutionCodeRepository {
	return &ticketResolutionCodeRepository{}
}

type ticketResolutionCodeRepository struct{}

func (r *ticketResolutionCodeRepository) Get(db *gorm.DB, id int64) *models.TicketResolutionCode {
	ret := &models.TicketResolutionCode{}
	if err := db.First(ret, "id = ?", id).Error; err != nil {
		return nil
	}
	return ret
}

func (r *ticketResolutionCodeRepository) Take(db *gorm.DB, where ...interface{}) *models.TicketResolutionCode {
	ret := &models.TicketResolutionCode{}
	if err := db.Take(ret, where...).Error; err != nil {
		return nil
	}
	return ret
}

func (r *ticketResolutionCodeRepository) Find(db *gorm.DB, cnd *sqls.Cnd) (list []models.TicketResolutionCode) {
	cnd.Find(db, &list)
	return
}

func (r *ticketResolutionCodeRepository) FindOne(db *gorm.DB, cnd *sqls.Cnd) *models.TicketResolutionCode {
	ret := &models.TicketResolutionCode{}
	if err := cnd.FindOne(db, &ret); err != nil {
		return nil
	}
	return ret
}

func (r *ticketResolutionCodeRepository) FindPageByParams(db *gorm.DB, params *params.QueryParams) (list []models.TicketResolutionCode, paging *sqls.Paging) {
	return r.FindPageByCnd(db, &params.Cnd)
}

func (r *ticketResolutionCodeRepository) FindPageByCnd(db *gorm.DB, cnd *sqls.Cnd) (list []models.TicketResolutionCode, paging *sqls.Paging) {
	cnd.Find(db, &list)
	count := cnd.Count(db, &models.TicketResolutionCode{})
	paging = &sqls.Paging{Page: cnd.Paging.Page, Limit: cnd.Paging.Limit, Total: count}
	return
}

func (r *ticketResolutionCodeRepository) Count(db *gorm.DB, cnd *sqls.Cnd) int64 {
	return cnd.Count(db, &models.TicketResolutionCode{})
}

func (r *ticketResolutionCodeRepository) Create(db *gorm.DB, t *models.TicketResolutionCode) error {
	return db.Create(t).Error
}

func (r *ticketResolutionCodeRepository) Update(db *gorm.DB, t *models.TicketResolutionCode) error {
	return db.Save(t).Error
}

func (r *ticketResolutionCodeRepository) Updates(db *gorm.DB, id int64, columns map[string]interface{}) error {
	return db.Model(&models.TicketResolutionCode{}).Where("id = ?", id).Updates(columns).Error
}

func (r *ticketResolutionCodeRepository) UpdateColumn(db *gorm.DB, id int64, name string, value interface{}) error {
	return db.Model(&models.TicketResolutionCode{}).Where("id = ?", id).UpdateColumn(name, value).Error
}

func (r *ticketResolutionCodeRepository) Delete(db *gorm.DB, id int64) {
	db.Delete(&models.TicketResolutionCode{}, "id = ?", id)
}
