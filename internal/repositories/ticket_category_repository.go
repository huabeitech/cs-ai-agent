package repositories

import (
	"cs-agent/internal/models"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
	"gorm.io/gorm"
)

var TicketCategoryRepository = newTicketCategoryRepository()

func newTicketCategoryRepository() *ticketCategoryRepository {
	return &ticketCategoryRepository{}
}

type ticketCategoryRepository struct{}

func (r *ticketCategoryRepository) Get(db *gorm.DB, id int64) *models.TicketCategory {
	ret := &models.TicketCategory{}
	if err := db.First(ret, "id = ?", id).Error; err != nil {
		return nil
	}
	return ret
}

func (r *ticketCategoryRepository) Take(db *gorm.DB, where ...interface{}) *models.TicketCategory {
	ret := &models.TicketCategory{}
	if err := db.Take(ret, where...).Error; err != nil {
		return nil
	}
	return ret
}

func (r *ticketCategoryRepository) Find(db *gorm.DB, cnd *sqls.Cnd) (list []models.TicketCategory) {
	cnd.Find(db, &list)
	return
}

func (r *ticketCategoryRepository) FindOne(db *gorm.DB, cnd *sqls.Cnd) *models.TicketCategory {
	ret := &models.TicketCategory{}
	if err := cnd.FindOne(db, &ret); err != nil {
		return nil
	}
	return ret
}

func (r *ticketCategoryRepository) FindPageByParams(db *gorm.DB, params *params.QueryParams) (list []models.TicketCategory, paging *sqls.Paging) {
	return r.FindPageByCnd(db, &params.Cnd)
}

func (r *ticketCategoryRepository) FindPageByCnd(db *gorm.DB, cnd *sqls.Cnd) (list []models.TicketCategory, paging *sqls.Paging) {
	cnd.Find(db, &list)
	count := cnd.Count(db, &models.TicketCategory{})
	paging = &sqls.Paging{Page: cnd.Paging.Page, Limit: cnd.Paging.Limit, Total: count}
	return
}

func (r *ticketCategoryRepository) Count(db *gorm.DB, cnd *sqls.Cnd) int64 {
	return cnd.Count(db, &models.TicketCategory{})
}

func (r *ticketCategoryRepository) Create(db *gorm.DB, t *models.TicketCategory) error {
	return db.Create(t).Error
}

func (r *ticketCategoryRepository) Update(db *gorm.DB, t *models.TicketCategory) error {
	return db.Save(t).Error
}

func (r *ticketCategoryRepository) Updates(db *gorm.DB, id int64, columns map[string]interface{}) error {
	return db.Model(&models.TicketCategory{}).Where("id = ?", id).Updates(columns).Error
}

func (r *ticketCategoryRepository) UpdateColumn(db *gorm.DB, id int64, name string, value interface{}) error {
	return db.Model(&models.TicketCategory{}).Where("id = ?", id).UpdateColumn(name, value).Error
}

func (r *ticketCategoryRepository) Delete(db *gorm.DB, id int64) {
	db.Delete(&models.TicketCategory{}, "id = ?", id)
}
