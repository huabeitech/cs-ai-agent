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
		PlannedSkillName: item.PlannedSkillName,
		SkillRouteTrace:  item.SkillRouteTrace,
		ToolSearchTrace:  item.ToolSearchTrace,
		GraphToolTrace:   item.GraphToolTrace,
		GraphToolCode:    item.GraphToolCode,
		HandoffReason:    item.HandoffReason,
		PlannedToolCode:  item.PlannedToolCode,
		PlanReason:       item.PlanReason,
		InterruptType:    item.InterruptType,
		ResumeSource:     item.ResumeSource,
		FinalAction:      item.FinalAction,
		FinalStatus:      item.FinalStatus,
		ReplyText:        item.ReplyText,
		ErrorMessage:     item.ErrorMessage,
		LatencyMs:        item.LatencyMs,
		TraceData:        item.TraceData,
		CreatedAt:        item.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}
