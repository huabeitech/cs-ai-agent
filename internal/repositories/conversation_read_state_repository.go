package repositories

import (
	"cs-agent/internal/models"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
	"gorm.io/gorm"
)

var ConversationReadStateRepository = newConversationReadStateRepository()

func newConversationReadStateRepository() *conversationReadStateRepository {
	return &conversationReadStateRepository{}
}

type conversationReadStateRepository struct {
}

func (r *conversationReadStateRepository) Get(db *gorm.DB, id int64) *models.ConversationReadState {
	ret := &models.ConversationReadState{}
	if err := db.First(ret, "id = ?", id).Error; err != nil {
		return nil
	}
	return ret
}

func (r *conversationReadStateRepository) Take(db *gorm.DB, where ...interface{}) *models.ConversationReadState {
	ret := &models.ConversationReadState{}
	if err := db.Take(ret, where...).Error; err != nil {
		return nil
	}
	return ret
}

func (r *conversationReadStateRepository) Find(db *gorm.DB, cnd *sqls.Cnd) (list []models.ConversationReadState) {
	cnd.Find(db, &list)
	return
}

func (r *conversationReadStateRepository) FindOne(db *gorm.DB, cnd *sqls.Cnd) *models.ConversationReadState {
	ret := &models.ConversationReadState{}
	if err := cnd.FindOne(db, &ret); err != nil {
		return nil
	}
	return ret
}

func (r *conversationReadStateRepository) FindPageByParams(db *gorm.DB, params *params.QueryParams) (list []models.ConversationReadState, paging *sqls.Paging) {
	return r.FindPageByCnd(db, &params.Cnd)
}

func (r *conversationReadStateRepository) FindPageByCnd(db *gorm.DB, cnd *sqls.Cnd) (list []models.ConversationReadState, paging *sqls.Paging) {
	cnd.Find(db, &list)
	count := cnd.Count(db, &models.ConversationReadState{})

	paging = &sqls.Paging{
		Page:  cnd.Paging.Page,
		Limit: cnd.Paging.Limit,
		Total: count,
	}
	return
}

func (r *conversationReadStateRepository) Count(db *gorm.DB, cnd *sqls.Cnd) int64 {
	return cnd.Count(db, &models.ConversationReadState{})
}

func (r *conversationReadStateRepository) Create(db *gorm.DB, t *models.ConversationReadState) (err error) {
	err = db.Create(t).Error
	return
}

func (r *conversationReadStateRepository) Update(db *gorm.DB, t *models.ConversationReadState) (err error) {
	err = db.Save(t).Error
	return
}

func (r *conversationReadStateRepository) Updates(db *gorm.DB, id int64, columns map[string]any) (err error) {
	err = db.Model(&models.ConversationReadState{}).Where("id = ?", id).Updates(columns).Error
	return
}

func (r *conversationReadStateRepository) UpdateColumn(db *gorm.DB, id int64, name string, value any) (err error) {
	err = db.Model(&models.ConversationReadState{}).Where("id = ?", id).UpdateColumn(name, value).Error
	return
}

func (r *conversationReadStateRepository) Delete(db *gorm.DB, id int64) {
	db.Delete(&models.ConversationReadState{}, "id = ?", id)
}
