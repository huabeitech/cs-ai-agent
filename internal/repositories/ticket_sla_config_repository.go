package repositories

import (
	"cs-agent/internal/models"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
	"gorm.io/gorm"
)

var TicketSLAConfigRepository = newTicketSLAConfigRepository()

func newTicketSLAConfigRepository() *ticketSLAConfigRepository {
	return &ticketSLAConfigRepository{}
}

type ticketSLAConfigRepository struct{}

func (r *ticketSLAConfigRepository) Get(db *gorm.DB, id int64) *models.TicketSLAConfig {
	ret := &models.TicketSLAConfig{}
	if err := db.First(ret, "id = ?", id).Error; err != nil {
		return nil
	}
	return ret
}

func (r *ticketSLAConfigRepository) Take(db *gorm.DB, where ...interface{}) *models.TicketSLAConfig {
	ret := &models.TicketSLAConfig{}
	if err := db.Take(ret, where...).Error; err != nil {
		return nil
	}
	return ret
}

func (r *ticketSLAConfigRepository) Find(db *gorm.DB, cnd *sqls.Cnd) (list []models.TicketSLAConfig) {
	cnd.Find(db, &list)
	return
}

func (r *ticketSLAConfigRepository) FindOne(db *gorm.DB, cnd *sqls.Cnd) *models.TicketSLAConfig {
	ret := &models.TicketSLAConfig{}
	if err := cnd.FindOne(db, &ret); err != nil {
		return nil
	}
	return ret
}

func (r *ticketSLAConfigRepository) FindPageByParams(db *gorm.DB, params *params.QueryParams) (list []models.TicketSLAConfig, paging *sqls.Paging) {
	return r.FindPageByCnd(db, &params.Cnd)
}

func (r *ticketSLAConfigRepository) FindPageByCnd(db *gorm.DB, cnd *sqls.Cnd) (list []models.TicketSLAConfig, paging *sqls.Paging) {
	cnd.Find(db, &list)
	count := cnd.Count(db, &models.TicketSLAConfig{})
	paging = &sqls.Paging{Page: cnd.Paging.Page, Limit: cnd.Paging.Limit, Total: count}
	return
}

func (r *ticketSLAConfigRepository) Count(db *gorm.DB, cnd *sqls.Cnd) int64 {
	return cnd.Count(db, &models.TicketSLAConfig{})
}

func (r *ticketSLAConfigRepository) Create(db *gorm.DB, t *models.TicketSLAConfig) error {
	return db.Create(t).Error
}

func (r *ticketSLAConfigRepository) Update(db *gorm.DB, t *models.TicketSLAConfig) error {
	return db.Save(t).Error
}

func (r *ticketSLAConfigRepository) Updates(db *gorm.DB, id int64, columns map[string]interface{}) error {
	return db.Model(&models.TicketSLAConfig{}).Where("id = ?", id).Updates(columns).Error
}

func (r *ticketSLAConfigRepository) UpdateColumn(db *gorm.DB, id int64, name string, value interface{}) error {
	return db.Model(&models.TicketSLAConfig{}).Where("id = ?", id).UpdateColumn(name, value).Error
}

func (r *ticketSLAConfigRepository) Delete(db *gorm.DB, id int64) error {
	return db.Delete(&models.TicketSLAConfig{}, "id = ?", id).Error
}
