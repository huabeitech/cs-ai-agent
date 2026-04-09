package repositories

import (
	"cs-agent/internal/models"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
	"gorm.io/gorm"
)

var ConversationRepository = newConversationRepository()

func newConversationRepository() *conversationRepository {
	return &conversationRepository{}
}

type conversationRepository struct {
}

func (r *conversationRepository) Get(db *gorm.DB, id int64) *models.Conversation {
	ret := &models.Conversation{}
	if err := db.First(ret, "id = ?", id).Error; err != nil {
		return nil
	}
	return ret
}

func (r *conversationRepository) Take(db *gorm.DB, where ...interface{}) *models.Conversation {
	ret := &models.Conversation{}
	if err := db.Take(ret, where...).Error; err != nil {
		return nil
	}
	return ret
}

func (r *conversationRepository) Find(db *gorm.DB, cnd *sqls.Cnd) (list []models.Conversation) {
	cnd.Find(db, &list)
	return
}

func (r *conversationRepository) FindOne(db *gorm.DB, cnd *sqls.Cnd) *models.Conversation {
	ret := &models.Conversation{}
	if err := cnd.FindOne(db, &ret); err != nil {
		return nil
	}
	return ret
}

func (r *conversationRepository) FindPageByParams(db *gorm.DB, params *params.QueryParams) (list []models.Conversation, paging *sqls.Paging) {
	return r.FindPageByCnd(db, &params.Cnd)
}

func (r *conversationRepository) FindPageByCnd(db *gorm.DB, cnd *sqls.Cnd) (list []models.Conversation, paging *sqls.Paging) {
	cnd.Find(db, &list)
	count := cnd.Count(db, &models.Conversation{})

	paging = &sqls.Paging{
		Page:  cnd.Paging.Page,
		Limit: cnd.Paging.Limit,
		Total: count,
	}
	return
}

func (r *conversationRepository) FindBySql(db *gorm.DB, sqlStr string, paramArr ...interface{}) (list []models.Conversation) {
	db.Raw(sqlStr, paramArr...).Scan(&list)
	return
}

func (r *conversationRepository) CountBySql(db *gorm.DB, sqlStr string, paramArr ...interface{}) (count int64) {
	db.Raw(sqlStr, paramArr...).Count(&count)
	return
}

func (r *conversationRepository) Count(db *gorm.DB, cnd *sqls.Cnd) int64 {
	return cnd.Count(db, &models.Conversation{})
}

func (r *conversationRepository) Create(db *gorm.DB, t *models.Conversation) (err error) {
	err = db.Create(t).Error
	return
}

func (r *conversationRepository) Update(db *gorm.DB, t *models.Conversation) (err error) {
	err = db.Save(t).Error
	return
}

func (r *conversationRepository) Updates(db *gorm.DB, id int64, columns map[string]interface{}) (err error) {
	err = db.Model(&models.Conversation{}).Where("id = ?", id).Updates(columns).Error
	return
}

func (r *conversationRepository) UpdateColumn(db *gorm.DB, id int64, name string, value interface{}) (err error) {
	err = db.Model(&models.Conversation{}).Where("id = ?", id).UpdateColumn(name, value).Error
	return
}

func (r *conversationRepository) Delete(db *gorm.DB, id int64) {
	db.Delete(&models.Conversation{}, "id = ?", id)
}
