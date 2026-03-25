package repositories

import (
	"cs-agent/internal/models"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
	"gorm.io/gorm"
)

var TicketAttachmentRepository = newTicketAttachmentRepository()

func newTicketAttachmentRepository() *ticketAttachmentRepository {
	return &ticketAttachmentRepository{}
}

type ticketAttachmentRepository struct {
}

func (r *ticketAttachmentRepository) Get(db *gorm.DB, id int64) *models.TicketAttachment {
	ret := &models.TicketAttachment{}
	if err := db.First(ret, "id = ?", id).Error; err != nil {
		return nil
	}
	return ret
}

func (r *ticketAttachmentRepository) Take(db *gorm.DB, where ...interface{}) *models.TicketAttachment {
	ret := &models.TicketAttachment{}
	if err := db.Take(ret, where...).Error; err != nil {
		return nil
	}
	return ret
}

func (r *ticketAttachmentRepository) Find(db *gorm.DB, cnd *sqls.Cnd) (list []models.TicketAttachment) {
	cnd.Find(db, &list)
	return
}

func (r *ticketAttachmentRepository) FindOne(db *gorm.DB, cnd *sqls.Cnd) *models.TicketAttachment {
	ret := &models.TicketAttachment{}
	if err := cnd.FindOne(db, &ret); err != nil {
		return nil
	}
	return ret
}

func (r *ticketAttachmentRepository) FindPageByParams(db *gorm.DB, params *params.QueryParams) (list []models.TicketAttachment, paging *sqls.Paging) {
	return r.FindPageByCnd(db, &params.Cnd)
}

func (r *ticketAttachmentRepository) FindPageByCnd(db *gorm.DB, cnd *sqls.Cnd) (list []models.TicketAttachment, paging *sqls.Paging) {
	cnd.Find(db, &list)
	count := cnd.Count(db, &models.TicketAttachment{})

	paging = &sqls.Paging{
		Page:  cnd.Paging.Page,
		Limit: cnd.Paging.Limit,
		Total: count,
	}
	return
}

func (r *ticketAttachmentRepository) FindBySql(db *gorm.DB, sqlStr string, paramArr ...interface{}) (list []models.TicketAttachment) {
	db.Raw(sqlStr, paramArr...).Scan(&list)
	return
}

func (r *ticketAttachmentRepository) CountBySql(db *gorm.DB, sqlStr string, paramArr ...interface{}) (count int64) {
	db.Raw(sqlStr, paramArr...).Count(&count)
	return
}

func (r *ticketAttachmentRepository) Count(db *gorm.DB, cnd *sqls.Cnd) int64 {
	return cnd.Count(db, &models.TicketAttachment{})
}

func (r *ticketAttachmentRepository) Create(db *gorm.DB, t *models.TicketAttachment) (err error) {
	err = db.Create(t).Error
	return
}

func (r *ticketAttachmentRepository) Update(db *gorm.DB, t *models.TicketAttachment) (err error) {
	err = db.Save(t).Error
	return
}

func (r *ticketAttachmentRepository) Updates(db *gorm.DB, id int64, columns map[string]interface{}) (err error) {
	err = db.Model(&models.TicketAttachment{}).Where("id = ?", id).Updates(columns).Error
	return
}

func (r *ticketAttachmentRepository) UpdateColumn(db *gorm.DB, id int64, name string, value interface{}) (err error) {
	err = db.Model(&models.TicketAttachment{}).Where("id = ?", id).UpdateColumn(name, value).Error
	return
}

func (r *ticketAttachmentRepository) Delete(db *gorm.DB, id int64) {
	db.Delete(&models.TicketAttachment{}, "id = ?", id)
}
