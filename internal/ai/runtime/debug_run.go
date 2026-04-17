package runtime

import (
	"context"
	"strings"

	applicationruntime "cs-agent/internal/ai/application/runtime"
	"cs-agent/internal/ai/runtime/graphs"
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/dto/request"
	"cs-agent/internal/pkg/dto/response"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/pkg/errorsx"
	svc "cs-agent/internal/services"
)

func init() {
	svc.SkillDebugRunHook = DebugRunSkill
	svc.SkillDebugResumeHook = DebugResumeSkill
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
	summary, err := Service.Run(ctx, applicationruntime.Request{
		Conversation:    conversation,
		UserMessage:     message,
		AIAgent:         *aiAgent,
		AIConfig:        *aiConfig,
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

func DebugResumeSkill(ctx context.Context, req request.SkillDebugResumeRequest) (*response.SkillDebugRunResponse, error) {
	aiAgent := svc.AIAgentService.Get(req.AIAgentID)
	if aiAgent == nil || aiAgent.Status != enums.StatusOk {
		return nil, errorsx.InvalidParam("AI Agent不存在或未启用")
	}
	aiConfig := svc.AIConfigService.Get(aiAgent.AIConfigID)
	if aiConfig == nil {
		return nil, errorsx.InvalidParam("AI Agent关联的AI配置不存在")
	}
	pendingInterrupt := svc.ConversationInterruptService.GetByCheckPointID(strings.TrimSpace(req.CheckPointID))
	if pendingInterrupt == nil {
		return nil, errorsx.InvalidParam("CheckPoint 不存在")
	}
	if pendingInterrupt.AIAgentID > 0 && pendingInterrupt.AIAgentID != req.AIAgentID {
		return nil, errorsx.InvalidParam("CheckPoint 与 AI Agent 不匹配")
	}
	conversationID := req.ConversationID
	if conversationID <= 0 {
		conversationID = pendingInterrupt.ConversationID
	}
	if conversationID <= 0 {
		return nil, errorsx.InvalidParam("会话不存在")
	}
	conversation := svc.ConversationService.Get(conversationID)
	if conversation == nil {
		return nil, errorsx.InvalidParam("会话不存在")
	}
	if conversation.AIAgentID > 0 && conversation.AIAgentID != req.AIAgentID {
		return nil, errorsx.InvalidParam("会话与 AI Agent 不匹配")
	}
	resumeText := strings.TrimSpace(req.UserMessage)
	summary, err := Service.Resume(ctx, applicationruntime.ResumeRequest{
		Conversation: conversation,
		AIAgent:      *aiAgent,
		AIConfig:     *aiConfig,
		CheckPointID: strings.TrimSpace(req.CheckPointID),
		ResumeData: map[string]string{
			strings.TrimSpace(pendingInterrupt.InterruptID): resumeText,
		},
	})
	if err != nil {
		if isCheckpointMissingError(err) {
			summary = &applicationruntime.Summary{
				Status:    "expired",
				ReplyText: graphs.ConfirmationExpiredReply,
			}
			if pendingInterrupt.ID > 0 {
				_ = svc.ConversationInterruptService.MarkExpired(pendingInterrupt.ID, 0)
			}
			return buildSkillDebugResumeResponse(req, summary, conversationID), nil
		}
		return buildSkillDebugResumeResponse(req, summary, conversationID), err
	}
	if pendingInterrupt.ID > 0 {
		if summary != nil && summary.Interrupted {
			_ = svc.ConversationInterruptService.MarkPendingAgain(pendingInterrupt.ID, firstInterruptID(summary), resolveInterruptPrompt(summary), 0)
		} else if summary != nil && graphs.IsCancellationReply(summary.ReplyText) {
			_ = svc.ConversationInterruptService.MarkCancelled(pendingInterrupt.ID, 0)
		} else {
			_ = svc.ConversationInterruptService.MarkResolved(pendingInterrupt.ID, 0)
		}
	}
	return buildSkillDebugResumeResponse(req, summary, conversationID), nil
}

func buildSkillDebugRunResponse(req request.SkillDebugRunRequest, summary *applicationruntime.Summary, skill *models.SkillDefinition) *response.SkillDebugRunResponse {
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
	resp.ToolWhitelist = append([]string(nil), summary.SkillAllowedToolCodes...)
	resp.ExposedToolCodes = append([]string(nil), summary.ToolCodes...)
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

func buildSkillDebugResumeResponse(req request.SkillDebugResumeRequest, summary *applicationruntime.Summary, conversationID int64) *response.SkillDebugRunResponse {
	resp := &response.SkillDebugRunResponse{
		ConversationID: conversationID,
		AIAgentID:      req.AIAgentID,
	}
	if summary == nil {
		return resp
	}
	resp.SkillCode = strings.TrimSpace(summary.PlannedSkillCode)
	resp.SkillName = strings.TrimSpace(summary.PlannedSkillName)
	resp.ReplyText = summary.ReplyText
	resp.PlanReason = summary.PlanReason
	resp.SkillRouteTrace = summary.SkillRouteTrace
	resp.ToolWhitelist = append([]string(nil), summary.SkillAllowedToolCodes...)
	resp.ExposedToolCodes = append([]string(nil), summary.ToolCodes...)
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
