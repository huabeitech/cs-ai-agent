package repositories

import (
	"cs-agent/internal/models"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
	"gorm.io/gorm"
)

var ConversationAssignmentRepository = newConversationAssignmentRepository()

func newConversationAssignmentRepository() *conversationAssignmentRepository {
	return &conversationAssignmentRepository{}
}

type conversationAssignmentRepository struct {
}

func (r *conversationAssignmentRepository) Get(db *gorm.DB, id int64) *models.ConversationAssignment {
	ret := &models.ConversationAssignment{}
	if err := db.First(ret, "id = ?", id).Error; err != nil {
		return nil
	}
	return ret
}

func (r *conversationAssignmentRepository) Take(db *gorm.DB, where ...interface{}) *models.ConversationAssignment {
	ret := &models.ConversationAssignment{}
	if err := db.Take(ret, where...).Error; err != nil {
		return nil
	}
	return ret
}

func (r *conversationAssignmentRepository) Find(db *gorm.DB, cnd *sqls.Cnd) (list []models.ConversationAssignment) {
	cnd.Find(db, &list)
	return
}

func (r *conversationAssignmentRepository) FindOne(db *gorm.DB, cnd *sqls.Cnd) *models.ConversationAssignment {
	ret := &models.ConversationAssignment{}
	if err := cnd.FindOne(db, &ret); err != nil {
		return nil
	}
	return ret
}

func (r *conversationAssignmentRepository) FindPageByParams(db *gorm.DB, params *params.QueryParams) (list []models.ConversationAssignment, paging *sqls.Paging) {
	return r.FindPageByCnd(db, &params.Cnd)
}

func (r *conversationAssignmentRepository) FindPageByCnd(db *gorm.DB, cnd *sqls.Cnd) (list []models.ConversationAssignment, paging *sqls.Paging) {
	cnd.Find(db, &list)
	count := cnd.Count(db, &models.ConversationAssignment{})

	paging = &sqls.Paging{
		Page:  cnd.Paging.Page,
		Limit: cnd.Paging.Limit,
		Total: count,
	}
	return
}

func (r *conversationAssignmentRepository) FindBySql(db *gorm.DB, sqlStr string, paramArr ...interface{}) (list []models.ConversationAssignment) {
	db.Raw(sqlStr, paramArr...).Scan(&list)
	return
}

func (r *conversationAssignmentRepository) CountBySql(db *gorm.DB, sqlStr string, paramArr ...interface{}) (count int64) {
	db.Raw(sqlStr, paramArr...).Count(&count)
	return
}

func (r *conversationAssignmentRepository) Count(db *gorm.DB, cnd *sqls.Cnd) int64 {
	return cnd.Count(db, &models.ConversationAssignment{})
}

func (r *conversationAssignmentRepository) Create(db *gorm.DB, t *models.ConversationAssignment) (err error) {
	err = db.Create(t).Error
	return
}

func (r *conversationAssignmentRepository) Update(db *gorm.DB, t *models.ConversationAssignment) (err error) {
	err = db.Save(t).Error
	return
}

func (r *conversationAssignmentRepository) Updates(db *gorm.DB, id int64, columns map[string]interface{}) (err error) {
	err = db.Model(&models.ConversationAssignment{}).Where("id = ?", id).Updates(columns).Error
	return
}

func (r *conversationAssignmentRepository) UpdateColumn(db *gorm.DB, id int64, name string, value interface{}) (err error) {
	err = db.Model(&models.ConversationAssignment{}).Where("id = ?", id).UpdateColumn(name, value).Error
	return
}

func (r *conversationAssignmentRepository) Delete(db *gorm.DB, id int64) {
	db.Delete(&models.ConversationAssignment{}, "id = ?", id)
}
