package repositories

import (
	"cs-agent/internal/models"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
	"gorm.io/gorm"
)

var TicketRepository = newTicketRepository()

func newTicketRepository() *ticketRepository {
	return &ticketRepository{}
}

type ticketRepository struct {
}

func (r *ticketRepository) Get(db *gorm.DB, id int64) *models.Ticket {
	ret := &models.Ticket{}
	if err := db.First(ret, "id = ?", id).Error; err != nil {
		return nil
	}
	return ret
}

func (r *ticketRepository) Take(db *gorm.DB, where ...interface{}) *models.Ticket {
	ret := &models.Ticket{}
	if err := db.Take(ret, where...).Error; err != nil {
		return nil
	}
	return ret
}

func (r *ticketRepository) Find(db *gorm.DB, cnd *sqls.Cnd) (list []models.Ticket) {
	cnd.Find(db, &list)
	return
}

func (r *ticketRepository) FindOne(db *gorm.DB, cnd *sqls.Cnd) *models.Ticket {
	ret := &models.Ticket{}
	if err := cnd.FindOne(db, &ret); err != nil {
		return nil
	}
	return ret
}

func (r *ticketRepository) FindPageByParams(db *gorm.DB, params *params.QueryParams) (list []models.Ticket, paging *sqls.Paging) {
	return r.FindPageByCnd(db, &params.Cnd)
}

func (r *ticketRepository) FindPageByCnd(db *gorm.DB, cnd *sqls.Cnd) (list []models.Ticket, paging *sqls.Paging) {
	cnd.Find(db, &list)
	count := cnd.Count(db, &models.Ticket{})

	paging = &sqls.Paging{
		Page:  cnd.Paging.Page,
		Limit: cnd.Paging.Limit,
		Total: count,
	}
	return
}

func (r *ticketRepository) FindBySql(db *gorm.DB, sqlStr string, paramArr ...interface{}) (list []models.Ticket) {
	db.Raw(sqlStr, paramArr...).Scan(&list)
	return
}

func (r *ticketRepository) CountBySql(db *gorm.DB, sqlStr string, paramArr ...interface{}) (count int64) {
	db.Raw(sqlStr, paramArr...).Count(&count)
	return
}

func (r *ticketRepository) Count(db *gorm.DB, cnd *sqls.Cnd) int64 {
	return cnd.Count(db, &models.Ticket{})
}

func (r *ticketRepository) Create(db *gorm.DB, t *models.Ticket) (err error) {
	err = db.Create(t).Error
	return
}

func (r *ticketRepository) Update(db *gorm.DB, t *models.Ticket) (err error) {
	err = db.Save(t).Error
	return
}

func (r *ticketRepository) Updates(db *gorm.DB, id int64, columns map[string]interface{}) (err error) {
	err = db.Model(&models.Ticket{}).Where("id = ?", id).Updates(columns).Error
	return
}

func (r *ticketRepository) UpdateColumn(db *gorm.DB, id int64, name string, value interface{}) (err error) {
	err = db.Model(&models.Ticket{}).Where("id = ?", id).UpdateColumn(name, value).Error
	return
}

func (r *ticketRepository) Delete(db *gorm.DB, id int64) {
	db.Delete(&models.Ticket{}, "id = ?", id)
}
