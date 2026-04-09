package repositories

import (
	"cs-agent/internal/models"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
	"gorm.io/gorm"
)

var ConversationParticipantRepository = newConversationParticipantRepository()

func newConversationParticipantRepository() *conversationParticipantRepository {
	return &conversationParticipantRepository{}
}

type conversationParticipantRepository struct {
}

func (r *conversationParticipantRepository) Get(db *gorm.DB, id int64) *models.ConversationParticipant {
	ret := &models.ConversationParticipant{}
	if err := db.First(ret, "id = ?", id).Error; err != nil {
		return nil
	}
	return ret
}

func (r *conversationParticipantRepository) Take(db *gorm.DB, where ...interface{}) *models.ConversationParticipant {
	ret := &models.ConversationParticipant{}
	if err := db.Take(ret, where...).Error; err != nil {
		return nil
	}
	return ret
}

func (r *conversationParticipantRepository) Find(db *gorm.DB, cnd *sqls.Cnd) (list []models.ConversationParticipant) {
	cnd.Find(db, &list)
	return
}

func (r *conversationParticipantRepository) FindOne(db *gorm.DB, cnd *sqls.Cnd) *models.ConversationParticipant {
	ret := &models.ConversationParticipant{}
	if err := cnd.FindOne(db, &ret); err != nil {
		return nil
	}
	return ret
}

func (r *conversationParticipantRepository) FindPageByParams(db *gorm.DB, params *params.QueryParams) (list []models.ConversationParticipant, paging *sqls.Paging) {
	return r.FindPageByCnd(db, &params.Cnd)
}

func (r *conversationParticipantRepository) FindPageByCnd(db *gorm.DB, cnd *sqls.Cnd) (list []models.ConversationParticipant, paging *sqls.Paging) {
	cnd.Find(db, &list)
	count := cnd.Count(db, &models.ConversationParticipant{})

	paging = &sqls.Paging{
		Page:  cnd.Paging.Page,
		Limit: cnd.Paging.Limit,
		Total: count,
	}
	return
}

func (r *conversationParticipantRepository) FindBySql(db *gorm.DB, sqlStr string, paramArr ...interface{}) (list []models.ConversationParticipant) {
	db.Raw(sqlStr, paramArr...).Scan(&list)
	return
}

func (r *conversationParticipantRepository) CountBySql(db *gorm.DB, sqlStr string, paramArr ...interface{}) (count int64) {
	db.Raw(sqlStr, paramArr...).Count(&count)
	return
}

func (r *conversationParticipantRepository) Count(db *gorm.DB, cnd *sqls.Cnd) int64 {
	return cnd.Count(db, &models.ConversationParticipant{})
}

func (r *conversationParticipantRepository) Create(db *gorm.DB, t *models.ConversationParticipant) (err error) {
	err = db.Create(t).Error
	return
}

func (r *conversationParticipantRepository) Update(db *gorm.DB, t *models.ConversationParticipant) (err error) {
	err = db.Save(t).Error
	return
}

func (r *conversationParticipantRepository) Updates(db *gorm.DB, id int64, columns map[string]interface{}) (err error) {
	err = db.Model(&models.ConversationParticipant{}).Where("id = ?", id).Updates(columns).Error
	return
}

func (r *conversationParticipantRepository) UpdateColumn(db *gorm.DB, id int64, name string, value interface{}) (err error) {
	err = db.Model(&models.ConversationParticipant{}).Where("id = ?", id).UpdateColumn(name, value).Error
	return
}

func (r *conversationParticipantRepository) Delete(db *gorm.DB, id int64) {
	db.Delete(&models.ConversationParticipant{}, "id = ?", id)
}
