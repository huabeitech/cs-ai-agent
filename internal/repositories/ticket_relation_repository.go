package repositories

import (
	"cs-agent/internal/models"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
	"gorm.io/gorm"
)

var TicketRelationRepository = newTicketRelationRepository()

func newTicketRelationRepository() *ticketRelationRepository {
	return &ticketRelationRepository{}
}

type ticketRelationRepository struct {
}

func (r *ticketRelationRepository) Get(db *gorm.DB, id int64) *models.TicketRelation {
	ret := &models.TicketRelation{}
	if err := db.First(ret, "id = ?", id).Error; err != nil {
		return nil
	}
	return ret
}

func (r *ticketRelationRepository) Take(db *gorm.DB, where ...interface{}) *models.TicketRelation {
	ret := &models.TicketRelation{}
	if err := db.Take(ret, where...).Error; err != nil {
		return nil
	}
	return ret
}

func (r *ticketRelationRepository) Find(db *gorm.DB, cnd *sqls.Cnd) (list []models.TicketRelation) {
	cnd.Find(db, &list)
	return
}

func (r *ticketRelationRepository) FindOne(db *gorm.DB, cnd *sqls.Cnd) *models.TicketRelation {
	ret := &models.TicketRelation{}
	if err := cnd.FindOne(db, &ret); err != nil {
		return nil
	}
	return ret
}

func (r *ticketRelationRepository) FindPageByParams(db *gorm.DB, params *params.QueryParams) (list []models.TicketRelation, paging *sqls.Paging) {
	return r.FindPageByCnd(db, &params.Cnd)
}

func (r *ticketRelationRepository) FindPageByCnd(db *gorm.DB, cnd *sqls.Cnd) (list []models.TicketRelation, paging *sqls.Paging) {
	cnd.Find(db, &list)
	count := cnd.Count(db, &models.TicketRelation{})

	paging = &sqls.Paging{
		Page:  cnd.Paging.Page,
		Limit: cnd.Paging.Limit,
		Total: count,
	}
	return
}

func (r *ticketRelationRepository) FindBySql(db *gorm.DB, sqlStr string, paramArr ...interface{}) (list []models.TicketRelation) {
	db.Raw(sqlStr, paramArr...).Scan(&list)
	return
}

func (r *ticketRelationRepository) CountBySql(db *gorm.DB, sqlStr string, paramArr ...interface{}) (count int64) {
	db.Raw(sqlStr, paramArr...).Count(&count)
	return
}

func (r *ticketRelationRepository) Count(db *gorm.DB, cnd *sqls.Cnd) int64 {
	return cnd.Count(db, &models.TicketRelation{})
}

func (r *ticketRelationRepository) Create(db *gorm.DB, t *models.TicketRelation) (err error) {
	err = db.Create(t).Error
	return
}

func (r *ticketRelationRepository) Update(db *gorm.DB, t *models.TicketRelation) (err error) {
	err = db.Save(t).Error
	return
}

func (r *ticketRelationRepository) Updates(db *gorm.DB, id int64, columns map[string]interface{}) (err error) {
	err = db.Model(&models.TicketRelation{}).Where("id = ?", id).Updates(columns).Error
	return
}

func (r *ticketRelationRepository) UpdateColumn(db *gorm.DB, id int64, name string, value interface{}) (err error) {
	err = db.Model(&models.TicketRelation{}).Where("id = ?", id).UpdateColumn(name, value).Error
	return
}

func (r *ticketRelationRepository) Delete(db *gorm.DB, id int64) {
	db.Delete(&models.TicketRelation{}, "id = ?", id)
}

func (r *ticketRelationRepository) DeleteByTicketRelation(db *gorm.DB, ticketID, relatedTicketID int64, relationType string) error {
	return db.Where("ticket_id = ? AND related_ticket_id = ? AND relation_type = ?", ticketID, relatedTicketID, relationType).
		Delete(&models.TicketRelation{}).Error
}
