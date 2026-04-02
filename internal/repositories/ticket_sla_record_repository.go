package repositories

import (
	"cs-agent/internal/models"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
	"gorm.io/gorm"
)

var TicketSLARecordRepository = newTicketSLARecordRepository()

func newTicketSLARecordRepository() *ticketSLARecordRepository {
	return &ticketSLARecordRepository{}
}

type ticketSLARecordRepository struct {
}

func (r *ticketSLARecordRepository) Get(db *gorm.DB, id int64) *models.TicketSLARecord {
	ret := &models.TicketSLARecord{}
	if err := db.First(ret, "id = ?", id).Error; err != nil {
		return nil
	}
	return ret
}

func (r *ticketSLARecordRepository) Take(db *gorm.DB, where ...interface{}) *models.TicketSLARecord {
	ret := &models.TicketSLARecord{}
	if err := db.Take(ret, where...).Error; err != nil {
		return nil
	}
	return ret
}

func (r *ticketSLARecordRepository) Find(db *gorm.DB, cnd *sqls.Cnd) (list []models.TicketSLARecord) {
	cnd.Find(db, &list)
	return
}

func (r *ticketSLARecordRepository) FindOne(db *gorm.DB, cnd *sqls.Cnd) *models.TicketSLARecord {
	ret := &models.TicketSLARecord{}
	if err := cnd.FindOne(db, &ret); err != nil {
		return nil
	}
	return ret
}

func (r *ticketSLARecordRepository) FindPageByParams(db *gorm.DB, params *params.QueryParams) (list []models.TicketSLARecord, paging *sqls.Paging) {
	return r.FindPageByCnd(db, &params.Cnd)
}

func (r *ticketSLARecordRepository) FindPageByCnd(db *gorm.DB, cnd *sqls.Cnd) (list []models.TicketSLARecord, paging *sqls.Paging) {
	cnd.Find(db, &list)
	count := cnd.Count(db, &models.TicketSLARecord{})

	paging = &sqls.Paging{
		Page:  cnd.Paging.Page,
		Limit: cnd.Paging.Limit,
		Total: count,
	}
	return
}

func (r *ticketSLARecordRepository) FindBySql(db *gorm.DB, sqlStr string, paramArr... interface{}) (list []models.TicketSLARecord) {
	db.Raw(sqlStr, paramArr...).Scan(&list)
	return
}

func (r *ticketSLARecordRepository) CountBySql(db *gorm.DB, sqlStr string, paramArr... interface{}) (count int64) {
	db.Raw(sqlStr, paramArr...).Count(&count)
	return
}

func (r *ticketSLARecordRepository) Count(db *gorm.DB, cnd *sqls.Cnd) int64 {
	return cnd.Count(db, &models.TicketSLARecord{})
}

func (r *ticketSLARecordRepository) Create(db *gorm.DB, t *models.TicketSLARecord) (err error) {
	err = db.Create(t).Error
	return
}

func (r *ticketSLARecordRepository) Update(db *gorm.DB, t *models.TicketSLARecord) (err error) {
	err = db.Save(t).Error
	return
}

func (r *ticketSLARecordRepository) Updates(db *gorm.DB, id int64, columns map[string]interface{}) (err error) {
	err = db.Model(&models.TicketSLARecord{}).Where("id = ?", id).Updates(columns).Error
	return
}

func (r *ticketSLARecordRepository) UpdateColumn(db *gorm.DB, id int64, name string, value interface{}) (err error) {
	err = db.Model(&models.TicketSLARecord{}).Where("id = ?", id).UpdateColumn(name, value).Error
	return
}

func (r *ticketSLARecordRepository) Delete(db *gorm.DB, id int64) {
	db.Delete(&models.TicketSLARecord{}, "id = ?", id)
}

