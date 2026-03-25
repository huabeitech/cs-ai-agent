package repositories

import (
	"cs-agent/internal/models"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
	"gorm.io/gorm"
)

var ConversationEventLogRepository = newConversationEventLogRepository()

func newConversationEventLogRepository() *conversationEventLogRepository {
	return &conversationEventLogRepository{}
}

type conversationEventLogRepository struct {
}

func (r *conversationEventLogRepository) Get(db *gorm.DB, id int64) *models.ConversationEventLog {
	ret := &models.ConversationEventLog{}
	if err := db.First(ret, "id = ?", id).Error; err != nil {
		return nil
	}
	return ret
}

func (r *conversationEventLogRepository) Take(db *gorm.DB, where ...interface{}) *models.ConversationEventLog {
	ret := &models.ConversationEventLog{}
	if err := db.Take(ret, where...).Error; err != nil {
		return nil
	}
	return ret
}

func (r *conversationEventLogRepository) Find(db *gorm.DB, cnd *sqls.Cnd) (list []models.ConversationEventLog) {
	cnd.Find(db, &list)
	return
}

func (r *conversationEventLogRepository) FindOne(db *gorm.DB, cnd *sqls.Cnd) *models.ConversationEventLog {
	ret := &models.ConversationEventLog{}
	if err := cnd.FindOne(db, &ret); err != nil {
		return nil
	}
	return ret
}

func (r *conversationEventLogRepository) FindPageByParams(db *gorm.DB, params *params.QueryParams) (list []models.ConversationEventLog, paging *sqls.Paging) {
	return r.FindPageByCnd(db, &params.Cnd)
}

func (r *conversationEventLogRepository) FindPageByCnd(db *gorm.DB, cnd *sqls.Cnd) (list []models.ConversationEventLog, paging *sqls.Paging) {
	cnd.Find(db, &list)
	count := cnd.Count(db, &models.ConversationEventLog{})

	paging = &sqls.Paging{
		Page:  cnd.Paging.Page,
		Limit: cnd.Paging.Limit,
		Total: count,
	}
	return
}

func (r *conversationEventLogRepository) FindBySql(db *gorm.DB, sqlStr string, paramArr ...interface{}) (list []models.ConversationEventLog) {
	db.Raw(sqlStr, paramArr...).Scan(&list)
	return
}

func (r *conversationEventLogRepository) CountBySql(db *gorm.DB, sqlStr string, paramArr ...interface{}) (count int64) {
	db.Raw(sqlStr, paramArr...).Count(&count)
	return
}

func (r *conversationEventLogRepository) Count(db *gorm.DB, cnd *sqls.Cnd) int64 {
	return cnd.Count(db, &models.ConversationEventLog{})
}

func (r *conversationEventLogRepository) Create(db *gorm.DB, t *models.ConversationEventLog) (err error) {
	err = db.Create(t).Error
	return
}

func (r *conversationEventLogRepository) Update(db *gorm.DB, t *models.ConversationEventLog) (err error) {
	err = db.Save(t).Error
	return
}

func (r *conversationEventLogRepository) Updates(db *gorm.DB, id int64, columns map[string]interface{}) (err error) {
	err = db.Model(&models.ConversationEventLog{}).Where("id = ?", id).Updates(columns).Error
	return
}

func (r *conversationEventLogRepository) UpdateColumn(db *gorm.DB, id int64, name string, value interface{}) (err error) {
	err = db.Model(&models.ConversationEventLog{}).Where("id = ?", id).UpdateColumn(name, value).Error
	return
}

func (r *conversationEventLogRepository) Delete(db *gorm.DB, id int64) {
	db.Delete(&models.ConversationEventLog{}, "id = ?", id)
}
