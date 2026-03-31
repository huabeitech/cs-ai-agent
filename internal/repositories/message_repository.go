package repositories

import (
	"cs-agent/internal/models"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
	"gorm.io/gorm"
)

var MessageRepository = newMessageRepository()

func newMessageRepository() *messageRepository {
	return &messageRepository{}
}

type messageRepository struct {
}

func (r *messageRepository) Get(db *gorm.DB, id int64) *models.Message {
	ret := &models.Message{}
	if err := db.First(ret, "id = ?", id).Error; err != nil {
		return nil
	}
	return ret
}

func (r *messageRepository) Take(db *gorm.DB, where ...interface{}) *models.Message {
	ret := &models.Message{}
	if err := db.Take(ret, where...).Error; err != nil {
		return nil
	}
	return ret
}

func (r *messageRepository) Find(db *gorm.DB, cnd *sqls.Cnd) (list []models.Message) {
	cnd.Find(db, &list)
	return
}

func (r *messageRepository) FindLastUnrecalledByConversationID(db *gorm.DB, conversationID int64) *models.Message {
	ret := &models.Message{}
	if err := db.
		Where("conversation_id = ? AND recalled_at IS NULL AND send_status <> ?", conversationID, 6).
		Order("seq_no DESC").
		Order("id DESC").
		Limit(1).
		Take(ret).Error; err != nil {
		return nil
	}
	return ret
}

func (r *messageRepository) FindOne(db *gorm.DB, cnd *sqls.Cnd) *models.Message {
	ret := &models.Message{}
	if err := cnd.FindOne(db, &ret); err != nil {
		return nil
	}
	return ret
}

func (r *messageRepository) FindPageByParams(db *gorm.DB, params *params.QueryParams) (list []models.Message, paging *sqls.Paging) {
	return r.FindPageByCnd(db, &params.Cnd)
}

func (r *messageRepository) FindPageByCnd(db *gorm.DB, cnd *sqls.Cnd) (list []models.Message, paging *sqls.Paging) {
	cnd.Find(db, &list)
	count := cnd.Count(db, &models.Message{})

	paging = &sqls.Paging{
		Page:  cnd.Paging.Page,
		Limit: cnd.Paging.Limit,
		Total: count,
	}
	return
}

func (r *messageRepository) FindBySql(db *gorm.DB, sqlStr string, paramArr ...interface{}) (list []models.Message) {
	db.Raw(sqlStr, paramArr...).Scan(&list)
	return
}

func (r *messageRepository) CountBySql(db *gorm.DB, sqlStr string, paramArr ...interface{}) (count int64) {
	db.Raw(sqlStr, paramArr...).Count(&count)
	return
}

func (r *messageRepository) Count(db *gorm.DB, cnd *sqls.Cnd) int64 {
	return cnd.Count(db, &models.Message{})
}

func (r *messageRepository) Create(db *gorm.DB, t *models.Message) (err error) {
	err = db.Create(t).Error
	return
}

func (r *messageRepository) Update(db *gorm.DB, t *models.Message) (err error) {
	err = db.Save(t).Error
	return
}

func (r *messageRepository) Updates(db *gorm.DB, id int64, columns map[string]interface{}) (err error) {
	err = db.Model(&models.Message{}).Where("id = ?", id).Updates(columns).Error
	return
}

func (r *messageRepository) UpdateColumn(db *gorm.DB, id int64, name string, value interface{}) (err error) {
	err = db.Model(&models.Message{}).Where("id = ?", id).UpdateColumn(name, value).Error
	return
}

func (r *messageRepository) Delete(db *gorm.DB, id int64) {
	db.Delete(&models.Message{}, "id = ?", id)
}
