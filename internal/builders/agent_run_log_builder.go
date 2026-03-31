package builders

import (
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/dto/response"
)

func BuildAgentRunLog(item *models.AgentRunLog) response.AgentRunLogResponse {
	if item == nil {
		return response.AgentRunLogResponse{}
	}
	return response.AgentRunLogResponse{
		ID:               item.ID,
		ConversationID:   item.ConversationID,
		MessageID:        item.MessageID,
		AIAgentID:        item.AIAgentID,
		AIConfigID:       item.AIConfigID,
		UserMessage:      item.UserMessage,
		PlannedAction:    item.PlannedAction,
		PlannedSkillCode: item.PlannedSkillCode,
		PlanReason:       item.PlanReason,
		FinalAction:      item.FinalAction,
		ReplyText:        item.ReplyText,
		ErrorMessage:     item.ErrorMessage,
		LatencyMs:        item.LatencyMs,
		CreatedAt:        item.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}
