package repositories

import (
	"cs-agent/internal/models"

	"github.com/mlogclub/simple/sqls"
	"gorm.io/gorm"
)

var ConversationInterruptRepository = newConversationInterruptRepository()

func newConversationInterruptRepository() *conversationInterruptRepository {
	return &conversationInterruptRepository{}
}

type conversationInterruptRepository struct{}

func (r *conversationInterruptRepository) Get(db *gorm.DB, id int64) *models.ConversationInterrupt {
	ret := &models.ConversationInterrupt{}
	if err := db.First(ret, "id = ?", id).Error; err != nil {
		return nil
	}
	return ret
}

func (r *conversationInterruptRepository) GetByCheckPointID(db *gorm.DB, checkPointID string) *models.ConversationInterrupt {
	ret := &models.ConversationInterrupt{}
	if err := db.Where("check_point_id = ?", checkPointID).First(ret).Error; err != nil {
		return nil
	}
	return ret
}

func (r *conversationInterruptRepository) FindLatestPendingByConversationID(db *gorm.DB, conversationID int64) *models.ConversationInterrupt {
	ret := &models.ConversationInterrupt{}
	if err := db.Where("conversation_id = ? AND status = ?", conversationID, "pending").Order("id DESC").First(ret).Error; err != nil {
		return nil
	}
	return ret
}

func (r *conversationInterruptRepository) Find(db *gorm.DB, cnd *sqls.Cnd) (list []models.ConversationInterrupt) {
	cnd.Find(db, &list)
	return
}

func (r *conversationInterruptRepository) FindOne(db *gorm.DB, cnd *sqls.Cnd) *models.ConversationInterrupt {
	ret := &models.ConversationInterrupt{}
	if err := cnd.FindOne(db, &ret); err != nil {
		return nil
	}
	return ret
}

func (r *conversationInterruptRepository) Create(db *gorm.DB, t *models.ConversationInterrupt) error {
	return db.Create(t).Error
}

func (r *conversationInterruptRepository) Update(db *gorm.DB, t *models.ConversationInterrupt) error {
	return db.Save(t).Error
}

func (r *conversationInterruptRepository) Updates(db *gorm.DB, id int64, columns map[string]any) error {
	return db.Model(&models.ConversationInterrupt{}).Where("id = ?", id).Updates(columns).Error
}

func (r *conversationInterruptRepository) UpsertByCheckPointID(db *gorm.DB, item *models.ConversationInterrupt) error {
	current := r.GetByCheckPointID(db, item.CheckPointID)
	if current == nil {
		return r.Create(db, item)
	}
	columns := map[string]any{
		"conversation_id":        item.ConversationID,
		"ai_agent_id":            item.AIAgentID,
		"source_message_id":      item.SourceMessageID,
		"last_resume_message_id": item.LastResumeMessageID,
		"interrupt_id":           item.InterruptID,
		"interrupt_type":         item.InterruptType,
		"status":                 item.Status,
		"prompt_text":            item.PromptText,
		"request_data":           item.RequestData,
		"check_point_data":       item.CheckPointData,
		"resume_count":           item.ResumeCount,
		"expires_at":             item.ExpiresAt,
		"updated_at":             item.UpdatedAt,
	}
	return r.Updates(db, current.ID, columns)
}
