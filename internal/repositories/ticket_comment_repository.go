package repositories

import (
	"cs-agent/internal/models"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
	"gorm.io/gorm"
)

var TicketCommentRepository = newTicketCommentRepository()

func newTicketCommentRepository() *ticketCommentRepository {
	return &ticketCommentRepository{}
}

type ticketCommentRepository struct {
}

func (r *ticketCommentRepository) Get(db *gorm.DB, id int64) *models.TicketComment {
	ret := &models.TicketComment{}
	if err := db.First(ret, "id = ?", id).Error; err != nil {
		return nil
	}
	return ret
}

func (r *ticketCommentRepository) Take(db *gorm.DB, where ...interface{}) *models.TicketComment {
	ret := &models.TicketComment{}
	if err := db.Take(ret, where...).Error; err != nil {
		return nil
	}
	return ret
}

func (r *ticketCommentRepository) Find(db *gorm.DB, cnd *sqls.Cnd) (list []models.TicketComment) {
	cnd.Find(db, &list)
	return
}

func (r *ticketCommentRepository) FindOne(db *gorm.DB, cnd *sqls.Cnd) *models.TicketComment {
	ret := &models.TicketComment{}
	if err := cnd.FindOne(db, &ret); err != nil {
		return nil
	}
	return ret
}

func (r *ticketCommentRepository) FindPageByParams(db *gorm.DB, params *params.QueryParams) (list []models.TicketComment, paging *sqls.Paging) {
	return r.FindPageByCnd(db, &params.Cnd)
}

func (r *ticketCommentRepository) FindPageByCnd(db *gorm.DB, cnd *sqls.Cnd) (list []models.TicketComment, paging *sqls.Paging) {
	cnd.Find(db, &list)
	count := cnd.Count(db, &models.TicketComment{})

	paging = &sqls.Paging{
		Page:  cnd.Paging.Page,
		Limit: cnd.Paging.Limit,
		Total: count,
	}
	return
}

func (r *ticketCommentRepository) FindBySql(db *gorm.DB, sqlStr string, paramArr ...interface{}) (list []models.TicketComment) {
	db.Raw(sqlStr, paramArr...).Scan(&list)
	return
}

func (r *ticketCommentRepository) CountBySql(db *gorm.DB, sqlStr string, paramArr ...interface{}) (count int64) {
	db.Raw(sqlStr, paramArr...).Count(&count)
	return
}

func (r *ticketCommentRepository) Count(db *gorm.DB, cnd *sqls.Cnd) int64 {
	return cnd.Count(db, &models.TicketComment{})
}

func (r *ticketCommentRepository) Create(db *gorm.DB, t *models.TicketComment) (err error) {
	err = db.Create(t).Error
	return
}

func (r *ticketCommentRepository) Update(db *gorm.DB, t *models.TicketComment) (err error) {
	err = db.Save(t).Error
	return
}

func (r *ticketCommentRepository) Updates(db *gorm.DB, id int64, columns map[string]interface{}) (err error) {
	err = db.Model(&models.TicketComment{}).Where("id = ?", id).Updates(columns).Error
	return
}

func (r *ticketCommentRepository) UpdateColumn(db *gorm.DB, id int64, name string, value interface{}) (err error) {
	err = db.Model(&models.TicketComment{}).Where("id = ?", id).UpdateColumn(name, value).Error
	return
}

func (r *ticketCommentRepository) Delete(db *gorm.DB, id int64) {
	db.Delete(&models.TicketComment{}, "id = ?", id)
}
