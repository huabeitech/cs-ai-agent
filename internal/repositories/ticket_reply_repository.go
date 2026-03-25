package repositories

import (
	"cs-agent/internal/models"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
	"gorm.io/gorm"
)

var TicketReplyRepository = newTicketReplyRepository()

func newTicketReplyRepository() *ticketReplyRepository {
	return &ticketReplyRepository{}
}

type ticketReplyRepository struct {
}

func (r *ticketReplyRepository) Get(db *gorm.DB, id int64) *models.TicketReply {
	ret := &models.TicketReply{}
	if err := db.First(ret, "id = ?", id).Error; err != nil {
		return nil
	}
	return ret
}

func (r *ticketReplyRepository) Take(db *gorm.DB, where ...interface{}) *models.TicketReply {
	ret := &models.TicketReply{}
	if err := db.Take(ret, where...).Error; err != nil {
		return nil
	}
	return ret
}

func (r *ticketReplyRepository) Find(db *gorm.DB, cnd *sqls.Cnd) (list []models.TicketReply) {
	cnd.Find(db, &list)
	return
}

func (r *ticketReplyRepository) FindOne(db *gorm.DB, cnd *sqls.Cnd) *models.TicketReply {
	ret := &models.TicketReply{}
	if err := cnd.FindOne(db, &ret); err != nil {
		return nil
	}
	return ret
}

func (r *ticketReplyRepository) FindPageByParams(db *gorm.DB, params *params.QueryParams) (list []models.TicketReply, paging *sqls.Paging) {
	return r.FindPageByCnd(db, &params.Cnd)
}

func (r *ticketReplyRepository) FindPageByCnd(db *gorm.DB, cnd *sqls.Cnd) (list []models.TicketReply, paging *sqls.Paging) {
	cnd.Find(db, &list)
	count := cnd.Count(db, &models.TicketReply{})

	paging = &sqls.Paging{
		Page:  cnd.Paging.Page,
		Limit: cnd.Paging.Limit,
		Total: count,
	}
	return
}

func (r *ticketReplyRepository) FindBySql(db *gorm.DB, sqlStr string, paramArr ...interface{}) (list []models.TicketReply) {
	db.Raw(sqlStr, paramArr...).Scan(&list)
	return
}

func (r *ticketReplyRepository) CountBySql(db *gorm.DB, sqlStr string, paramArr ...interface{}) (count int64) {
	db.Raw(sqlStr, paramArr...).Count(&count)
	return
}

func (r *ticketReplyRepository) Count(db *gorm.DB, cnd *sqls.Cnd) int64 {
	return cnd.Count(db, &models.TicketReply{})
}

func (r *ticketReplyRepository) Create(db *gorm.DB, t *models.TicketReply) (err error) {
	err = db.Create(t).Error
	return
}

func (r *ticketReplyRepository) Update(db *gorm.DB, t *models.TicketReply) (err error) {
	err = db.Save(t).Error
	return
}

func (r *ticketReplyRepository) Updates(db *gorm.DB, id int64, columns map[string]interface{}) (err error) {
	err = db.Model(&models.TicketReply{}).Where("id = ?", id).Updates(columns).Error
	return
}

func (r *ticketReplyRepository) UpdateColumn(db *gorm.DB, id int64, name string, value interface{}) (err error) {
	err = db.Model(&models.TicketReply{}).Where("id = ?", id).UpdateColumn(name, value).Error
	return
}

func (r *ticketReplyRepository) Delete(db *gorm.DB, id int64) {
	db.Delete(&models.TicketReply{}, "id = ?", id)
}
