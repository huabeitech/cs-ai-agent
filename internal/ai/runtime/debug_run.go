package runtime

import (
	"context"
	"strings"

	"cs-agent/internal/models"
	"cs-agent/internal/pkg/dto/request"
	"cs-agent/internal/pkg/dto/response"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/pkg/errorsx"
	svc "cs-agent/internal/services"
)

func init() {
	svc.SkillDebugRunHook = DebugRunSkill
}

func DebugRunSkill(ctx context.Context, req request.SkillDebugRunRequest) (*response.SkillDebugRunResponse, error) {
	aiAgent := svc.AIAgentService.Get(req.AIAgentID)
	if aiAgent == nil || aiAgent.Status != enums.StatusOk {
		return nil, errorsx.InvalidParam("AI Agent不存在或未启用")
	}
	aiConfig := svc.AIConfigService.Get(aiAgent.AIConfigID)
	if aiConfig == nil {
		return nil, errorsx.InvalidParam("AI Agent关联的AI配置不存在")
	}
	conversation := &models.Conversation{ID: req.ConversationID, AIAgentID: req.AIAgentID}
	if req.ConversationID > 0 {
		conversation = svc.ConversationService.Get(req.ConversationID)
		if conversation == nil {
			return nil, errorsx.InvalidParam("会话不存在")
		}
	}
	message := &models.Message{
		ConversationID: req.ConversationID,
		SenderType:     enums.IMSenderTypeCustomer,
		MessageType:    enums.IMMessageTypeText,
		Content:        strings.TrimSpace(req.UserMessage),
	}
	summary, err := Service.Run(ctx, Request{
		Conversation:    conversation,
		UserMessage:     message,
		AIAgent:         aiAgent,
		AIConfig:        aiConfig,
		ManualSkillCode: strings.TrimSpace(req.SkillCode),
	})
	if err != nil {
		return buildSkillDebugRunResponse(req, summary, nil), err
	}
	selectedSkill := svc.SkillDefinitionService.GetByCode(strings.TrimSpace(req.SkillCode))
	if summary == nil || strings.TrimSpace(summary.PlannedSkillCode) == "" {
		return nil, errorsx.InvalidParam("Skill 未命中")
	}
	return buildSkillDebugRunResponse(req, summary, selectedSkill), nil
}

func buildSkillDebugRunResponse(req request.SkillDebugRunRequest, summary *Summary, skill *models.SkillDefinition) *response.SkillDebugRunResponse {
	resp := &response.SkillDebugRunResponse{
		ConversationID: req.ConversationID,
		AIAgentID:      req.AIAgentID,
	}
	if skill != nil {
		resp.SkillCode = skill.Code
		resp.SkillName = skill.Name
	}
	if summary == nil {
		return resp
	}
	if resp.SkillCode == "" {
		resp.SkillCode = strings.TrimSpace(summary.PlannedSkillCode)
	}
	resp.ReplyText = summary.ReplyText
	resp.PlanReason = summary.PlanReason
	resp.SkillRouteTrace = summary.SkillRouteTrace
	resp.SkillAllowedToolCodes = append([]string(nil), summary.SkillAllowedToolCodes...)
	resp.ToolCodes = append([]string(nil), summary.ToolCodes...)
	resp.InvokedToolCodes = append([]string(nil), summary.InvokedToolCodes...)
	resp.ToolSearchTrace = extractToolSearchTrace(summary)
	resp.GraphToolTrace = extractGraphToolTrace(summary)
	resp.GraphToolCode = firstGraphToolCode(summary)
	resp.InterruptType = firstInterruptType(summary)
	resp.CheckPointID = summary.CheckPointID
	resp.Interrupted = summary.Interrupted
	resp.TraceData = summary.TraceData
	resp.ErrorMessage = summary.ErrorMessage
	return resp
}
