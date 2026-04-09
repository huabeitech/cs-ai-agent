package services

import (
	"context"
	"strings"

	"cs-agent/internal/ai/skills"
	"cs-agent/internal/pkg/dto/request"
	"cs-agent/internal/pkg/dto/response"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/pkg/errorsx"
)

var SkillRuntimeService = newSkillRuntimeService()

func newSkillRuntimeService() *skillRuntimeService {
	return &skillRuntimeService{}
}

type skillRuntimeService struct{}

func (s *skillRuntimeService) DebugRun(ctx context.Context, req request.SkillDebugRunRequest) (*response.SkillDebugRunResponse, error) {
	if req.AIAgentID <= 0 {
		return nil, errorsx.InvalidParam("aiAgentId不能为空")
	}
	if strings.TrimSpace(req.SkillCode) == "" {
		return nil, errorsx.InvalidParam("skillCode不能为空")
	}
	if strings.TrimSpace(req.UserMessage) == "" {
		return nil, errorsx.InvalidParam("userMessage不能为空")
	}

	aiAgent := AIAgentService.Get(req.AIAgentID)
	if aiAgent == nil || aiAgent.Status != enums.StatusOk {
		return nil, errorsx.InvalidParam("AI Agent不存在或未启用")
	}
	if AIConfigService.Get(aiAgent.AIConfigID) == nil {
		return nil, errorsx.InvalidParam("AI Agent关联的AI配置不存在")
	}
	if req.ConversationID > 0 {
		conversation := ConversationService.Get(req.ConversationID)
		if conversation == nil {
			return nil, errorsx.InvalidParam("会话不存在")
		}
	}

	result, err := skills.Execute(ctx, skills.RuntimeContext{
		AIAgentID:       req.AIAgentID,
		UserMessage:     strings.TrimSpace(req.UserMessage),
		ConversationID:  req.ConversationID,
		ManualSkillCode: strings.TrimSpace(req.SkillCode),
	})
	if err != nil {
		return nil, err
	}
	if result == nil || result.Plan == nil || result.Plan.Skill == nil {
		return nil, errorsx.InvalidParam("Skill 未命中")
	}
	return buildSkillDebugRunResponse(req, result), nil
}

func buildSkillDebugRunResponse(req request.SkillDebugRunRequest, result *skills.ExecutionResult) *response.SkillDebugRunResponse {
	resp := &response.SkillDebugRunResponse{
		ConversationID: req.ConversationID,
		AIAgentID:      req.AIAgentID,
		ReplyText:      result.ReplyText,
	}
	if result.Plan != nil && result.Plan.Skill != nil {
		resp.SkillCode = result.Plan.Skill.Code
		resp.SkillName = result.Plan.Skill.Name
	}
	if result.RunLog != nil {
		resp.RunLogID = result.RunLog.ID
		resp.TraceData = result.RunLog.TraceData
	}
	return resp
}
