package adapter

import "cs-agent/internal/models"

type ConversationSnapshot struct {
	ID                int64
	AIAgentID         int64
	LastMessageID     int64
	CurrentAssigneeID int64
}

func BuildConversationSnapshot(item *models.Conversation) *ConversationSnapshot {
	if item == nil {
		return nil
	}
	return &ConversationSnapshot{
		ID:                item.ID,
		AIAgentID:         item.AIAgentID,
		LastMessageID:     item.LastMessageID,
		CurrentAssigneeID: item.CurrentAssigneeID,
	}
}
