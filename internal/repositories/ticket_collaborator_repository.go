package repositories

import (
	"cs-agent/internal/models"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
	"gorm.io/gorm"
)

var TicketCollaboratorRepository = newTicketCollaboratorRepository()

func newTicketCollaboratorRepository() *ticketCollaboratorRepository {
	return &ticketCollaboratorRepository{}
}

type ticketCollaboratorRepository struct {
}

func (r *ticketCollaboratorRepository) Get(db *gorm.DB, id int64) *models.TicketCollaborator {
	ret := &models.TicketCollaborator{}
	if err := db.First(ret, "id = ?", id).Error; err != nil {
		return nil
	}
	return ret
}

func (r *ticketCollaboratorRepository) Take(db *gorm.DB, where ...interface{}) *models.TicketCollaborator {
	ret := &models.TicketCollaborator{}
	if err := db.Take(ret, where...).Error; err != nil {
		return nil
	}
	return ret
}

func (r *ticketCollaboratorRepository) Find(db *gorm.DB, cnd *sqls.Cnd) (list []models.TicketCollaborator) {
	cnd.Find(db, &list)
	return
}

func (r *ticketCollaboratorRepository) FindOne(db *gorm.DB, cnd *sqls.Cnd) *models.TicketCollaborator {
	ret := &models.TicketCollaborator{}
	if err := cnd.FindOne(db, &ret); err != nil {
		return nil
	}
	return ret
}

func (r *ticketCollaboratorRepository) FindPageByParams(db *gorm.DB, params *params.QueryParams) (list []models.TicketCollaborator, paging *sqls.Paging) {
	return r.FindPageByCnd(db, &params.Cnd)
}

func (r *ticketCollaboratorRepository) FindPageByCnd(db *gorm.DB, cnd *sqls.Cnd) (list []models.TicketCollaborator, paging *sqls.Paging) {
	cnd.Find(db, &list)
	count := cnd.Count(db, &models.TicketCollaborator{})

	paging = &sqls.Paging{
		Page:  cnd.Paging.Page,
		Limit: cnd.Paging.Limit,
		Total: count,
	}
	return
}

func (r *ticketCollaboratorRepository) Count(db *gorm.DB, cnd *sqls.Cnd) int64 {
	return cnd.Count(db, &models.TicketCollaborator{})
}

func (r *ticketCollaboratorRepository) Create(db *gorm.DB, t *models.TicketCollaborator) error {
	return db.Create(t).Error
}

func (r *ticketCollaboratorRepository) Delete(db *gorm.DB, id int64) {
	db.Delete(&models.TicketCollaborator{}, "id = ?", id)
}
