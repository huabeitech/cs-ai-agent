package services

import (
	"context"
	"fmt"
	"strings"

	"cs-agent/internal/pkg/dto/request"
	"cs-agent/internal/pkg/dto/response"
	"cs-agent/internal/pkg/errorsx"
)

var SkillRuntimeService = newSkillRuntimeService()
var SkillDebugRunHook func(ctx context.Context, req request.SkillDebugRunRequest) (*response.SkillDebugRunResponse, error)

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
	if SkillDebugRunHook == nil {
		return nil, fmt.Errorf("skill debug runner is not initialized")
	}
	return SkillDebugRunHook(ctx, req)
}
