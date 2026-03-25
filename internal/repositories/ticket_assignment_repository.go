package repositories

import (
	"cs-agent/internal/models"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
	"gorm.io/gorm"
)

var TicketAssignmentRepository = newTicketAssignmentRepository()

func newTicketAssignmentRepository() *ticketAssignmentRepository {
	return &ticketAssignmentRepository{}
}

type ticketAssignmentRepository struct {
}

func (r *ticketAssignmentRepository) Get(db *gorm.DB, id int64) *models.TicketAssignment {
	ret := &models.TicketAssignment{}
	if err := db.First(ret, "id = ?", id).Error; err != nil {
		return nil
	}
	return ret
}

func (r *ticketAssignmentRepository) Take(db *gorm.DB, where ...interface{}) *models.TicketAssignment {
	ret := &models.TicketAssignment{}
	if err := db.Take(ret, where...).Error; err != nil {
		return nil
	}
	return ret
}

func (r *ticketAssignmentRepository) Find(db *gorm.DB, cnd *sqls.Cnd) (list []models.TicketAssignment) {
	cnd.Find(db, &list)
	return
}

func (r *ticketAssignmentRepository) FindOne(db *gorm.DB, cnd *sqls.Cnd) *models.TicketAssignment {
	ret := &models.TicketAssignment{}
	if err := cnd.FindOne(db, &ret); err != nil {
		return nil
	}
	return ret
}

func (r *ticketAssignmentRepository) FindPageByParams(db *gorm.DB, params *params.QueryParams) (list []models.TicketAssignment, paging *sqls.Paging) {
	return r.FindPageByCnd(db, &params.Cnd)
}

func (r *ticketAssignmentRepository) FindPageByCnd(db *gorm.DB, cnd *sqls.Cnd) (list []models.TicketAssignment, paging *sqls.Paging) {
	cnd.Find(db, &list)
	count := cnd.Count(db, &models.TicketAssignment{})

	paging = &sqls.Paging{
		Page:  cnd.Paging.Page,
		Limit: cnd.Paging.Limit,
		Total: count,
	}
	return
}

func (r *ticketAssignmentRepository) FindBySql(db *gorm.DB, sqlStr string, paramArr ...interface{}) (list []models.TicketAssignment) {
	db.Raw(sqlStr, paramArr...).Scan(&list)
	return
}

func (r *ticketAssignmentRepository) CountBySql(db *gorm.DB, sqlStr string, paramArr ...interface{}) (count int64) {
	db.Raw(sqlStr, paramArr...).Count(&count)
	return
}

func (r *ticketAssignmentRepository) Count(db *gorm.DB, cnd *sqls.Cnd) int64 {
	return cnd.Count(db, &models.TicketAssignment{})
}

func (r *ticketAssignmentRepository) Create(db *gorm.DB, t *models.TicketAssignment) (err error) {
	err = db.Create(t).Error
	return
}

func (r *ticketAssignmentRepository) Update(db *gorm.DB, t *models.TicketAssignment) (err error) {
	err = db.Save(t).Error
	return
}

func (r *ticketAssignmentRepository) Updates(db *gorm.DB, id int64, columns map[string]interface{}) (err error) {
	err = db.Model(&models.TicketAssignment{}).Where("id = ?", id).Updates(columns).Error
	return
}

func (r *ticketAssignmentRepository) UpdateColumn(db *gorm.DB, id int64, name string, value interface{}) (err error) {
	err = db.Model(&models.TicketAssignment{}).Where("id = ?", id).UpdateColumn(name, value).Error
	return
}

func (r *ticketAssignmentRepository) Delete(db *gorm.DB, id int64) {
	db.Delete(&models.TicketAssignment{}, "id = ?", id)
}
