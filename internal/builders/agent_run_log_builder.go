package builders

import (
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/dto/response"
	"encoding/json"
	"strings"
)

func BuildAgentRunLog(item *models.AgentRunLog) response.AgentRunLogResponse {
	if item == nil {
		return response.AgentRunLogResponse{}
	}
	hitlStatus, hitlStatusName, hitlSummary := resolveAgentRunLogHITL(item)
	recommendedAction, riskLevel, ticketDraftReady := resolveGraphOutcome(item.GraphToolTrace)
	return response.AgentRunLogResponse{
		ID:                item.ID,
		ConversationID:    item.ConversationID,
		MessageID:         item.MessageID,
		AIAgentID:         item.AIAgentID,
		AIConfigID:        item.AIConfigID,
		UserMessage:       item.UserMessage,
		PlannedAction:     item.PlannedAction,
		PlannedSkillCode:  item.PlannedSkillCode,
		PlannedSkillName:  item.PlannedSkillName,
		SkillRouteTrace:   item.SkillRouteTrace,
		ToolSearchTrace:   item.ToolSearchTrace,
		GraphToolTrace:    item.GraphToolTrace,
		GraphToolCode:     item.GraphToolCode,
		RecommendedAction: recommendedAction,
		RiskLevel:         riskLevel,
		TicketDraftReady:  ticketDraftReady,
		HandoffReason:     item.HandoffReason,
		PlannedToolCode:   item.PlannedToolCode,
		PlanReason:        item.PlanReason,
		InterruptType:     item.InterruptType,
		ResumeSource:      item.ResumeSource,
		HitlStatus:        hitlStatus,
		HitlStatusName:    hitlStatusName,
		HitlSummary:       hitlSummary,
		FinalAction:       item.FinalAction,
		FinalStatus:       item.FinalStatus,
		ReplyText:         item.ReplyText,
		ErrorMessage:      item.ErrorMessage,
		LatencyMs:         item.LatencyMs,
		TraceData:         item.TraceData,
		CreatedAt:         item.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

func resolveGraphOutcome(raw string) (recommendedAction, riskLevel string, ticketDraftReady bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", "", false
	}
	var payload struct {
		Items []struct {
			RecommendedAction string `json:"recommendedAction"`
			RiskLevel         string `json:"riskLevel"`
			TicketDraftReady  bool   `json:"ticketDraftReady"`
		} `json:"items"`
	}
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return "", "", false
	}
	for _, item := range payload.Items {
		if strings.TrimSpace(item.RecommendedAction) != "" || strings.TrimSpace(item.RiskLevel) != "" || item.TicketDraftReady {
			return strings.TrimSpace(item.RecommendedAction), strings.TrimSpace(item.RiskLevel), item.TicketDraftReady
		}
	}
	return "", "", false
}

func resolveAgentRunLogHITL(item *models.AgentRunLog) (status, statusName, summary string) {
	if item == nil {
		return "", "", ""
	}
	replyText := strings.TrimSpace(item.ReplyText)
	switch {
	case strings.TrimSpace(item.FinalStatus) == "interrupted":
		return "pending", "等待确认", "Graph Tool 已发起确认，正在等待用户回复。"
	case strings.TrimSpace(item.FinalStatus) == "expired":
		return "expired", "已过期", "确认 checkpoint 已失效，需要重新发起。"
	case strings.Contains(replyText, "已取消本次工单创建") || strings.Contains(replyText, "已取消本次转人工"):
		return "cancelled", "已取消", "用户已明确取消，本次确认流程已终止。"
	case strings.TrimSpace(item.ResumeSource) != "":
		return "confirmed", "已确认", "用户确认后已恢复执行，并完成后续流程。"
	case strings.TrimSpace(item.InterruptType) != "":
		return "triggered", "已触发", "本次运行涉及确认式 HITL 流程。"
	default:
		return "", "", ""
	}
}
