package repositories

import (
	"cs-agent/internal/models"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
	"gorm.io/gorm"
)

var TicketEventLogRepository = newTicketEventLogRepository()

func newTicketEventLogRepository() *ticketEventLogRepository {
	return &ticketEventLogRepository{}
}

type ticketEventLogRepository struct {
}

func (r *ticketEventLogRepository) Get(db *gorm.DB, id int64) *models.TicketEventLog {
	ret := &models.TicketEventLog{}
	if err := db.First(ret, "id = ?", id).Error; err != nil {
		return nil
	}
	return ret
}

func (r *ticketEventLogRepository) Take(db *gorm.DB, where ...interface{}) *models.TicketEventLog {
	ret := &models.TicketEventLog{}
	if err := db.Take(ret, where...).Error; err != nil {
		return nil
	}
	return ret
}

func (r *ticketEventLogRepository) Find(db *gorm.DB, cnd *sqls.Cnd) (list []models.TicketEventLog) {
	cnd.Find(db, &list)
	return
}

func (r *ticketEventLogRepository) FindOne(db *gorm.DB, cnd *sqls.Cnd) *models.TicketEventLog {
	ret := &models.TicketEventLog{}
	if err := cnd.FindOne(db, &ret); err != nil {
		return nil
	}
	return ret
}

func (r *ticketEventLogRepository) FindPageByParams(db *gorm.DB, params *params.QueryParams) (list []models.TicketEventLog, paging *sqls.Paging) {
	return r.FindPageByCnd(db, &params.Cnd)
}

func (r *ticketEventLogRepository) FindPageByCnd(db *gorm.DB, cnd *sqls.Cnd) (list []models.TicketEventLog, paging *sqls.Paging) {
	cnd.Find(db, &list)
	count := cnd.Count(db, &models.TicketEventLog{})

	paging = &sqls.Paging{
		Page:  cnd.Paging.Page,
		Limit: cnd.Paging.Limit,
		Total: count,
	}
	return
}

func (r *ticketEventLogRepository) FindBySql(db *gorm.DB, sqlStr string, paramArr... interface{}) (list []models.TicketEventLog) {
	db.Raw(sqlStr, paramArr...).Scan(&list)
	return
}

func (r *ticketEventLogRepository) CountBySql(db *gorm.DB, sqlStr string, paramArr... interface{}) (count int64) {
	db.Raw(sqlStr, paramArr...).Count(&count)
	return
}

func (r *ticketEventLogRepository) Count(db *gorm.DB, cnd *sqls.Cnd) int64 {
	return cnd.Count(db, &models.TicketEventLog{})
}

func (r *ticketEventLogRepository) Create(db *gorm.DB, t *models.TicketEventLog) (err error) {
	err = db.Create(t).Error
	return
}

func (r *ticketEventLogRepository) Update(db *gorm.DB, t *models.TicketEventLog) (err error) {
	err = db.Save(t).Error
	return
}

func (r *ticketEventLogRepository) Updates(db *gorm.DB, id int64, columns map[string]interface{}) (err error) {
	err = db.Model(&models.TicketEventLog{}).Where("id = ?", id).Updates(columns).Error
	return
}

func (r *ticketEventLogRepository) UpdateColumn(db *gorm.DB, id int64, name string, value interface{}) (err error) {
	err = db.Model(&models.TicketEventLog{}).Where("id = ?", id).UpdateColumn(name, value).Error
	return
}

func (r *ticketEventLogRepository) Delete(db *gorm.DB, id int64) {
	db.Delete(&models.TicketEventLog{}, "id = ?", id)
}

