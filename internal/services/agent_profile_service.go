package services

import (
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/dto"
	"cs-agent/internal/pkg/dto/request"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/pkg/errorsx"
	"cs-agent/internal/pkg/utils"
	"cs-agent/internal/repositories"
	"strings"
	"time"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
)

var AgentProfileService = newAgentProfileService()

func newAgentProfileService() *agentProfileService {
	return &agentProfileService{}
}

type agentProfileService struct {
}

func (s *agentProfileService) Get(id int64) *models.AgentProfile {
	return repositories.AgentProfileRepository.Get(sqls.DB(), id)
}

func (s *agentProfileService) Take(where ...interface{}) *models.AgentProfile {
	return repositories.AgentProfileRepository.Take(sqls.DB(), where...)
}

func (s *agentProfileService) Find(cnd *sqls.Cnd) []models.AgentProfile {
	return repositories.AgentProfileRepository.Find(sqls.DB(), cnd)
}

func (s *agentProfileService) FindOne(cnd *sqls.Cnd) *models.AgentProfile {
	return repositories.AgentProfileRepository.FindOne(sqls.DB(), cnd)
}

func (s *agentProfileService) FindPageByParams(params *params.QueryParams) (list []models.AgentProfile, paging *sqls.Paging) {
	return repositories.AgentProfileRepository.FindPageByParams(sqls.DB(), params)
}

func (s *agentProfileService) FindPageByCnd(cnd *sqls.Cnd) (list []models.AgentProfile, paging *sqls.Paging) {
	return repositories.AgentProfileRepository.FindPageByCnd(sqls.DB(), cnd)
}

func (s *agentProfileService) Count(cnd *sqls.Cnd) int64 {
	return repositories.AgentProfileRepository.Count(sqls.DB(), cnd)
}

func (s *agentProfileService) GetByUserID(userID int64) *models.AgentProfile {
	return repositories.AgentProfileRepository.FindOne(sqls.DB(), sqls.NewCnd().Eq("user_id", userID))
}

// GetDispatchAgents 获取可用于分配会话的客服
func (s *agentProfileService) GetDispatchAgents(teamIds []int64) []models.AgentProfile {
	return AgentProfileService.Find(sqls.NewCnd().
		In("team_id", teamIds).
		Eq("status", enums.StatusOk).
		Eq("auto_assign_enabled", true).
		Eq("service_status", enums.ServiceStatusIdle))
}

func (s *agentProfileService) CreateAgentProfile(req request.CreateAgentProfileRequest, operator *dto.AuthPrincipal) (*models.AgentProfile, error) {
	if operator == nil {
		return nil, errorsx.Unauthorized("未登录或登录已过期")
	}
	item, err := s.buildProfileModel(0, req)
	if err != nil {
		return nil, err
	}
	item.AuditFields = utils.BuildAuditFields(operator)
	if err := repositories.AgentProfileRepository.Create(sqls.DB(), item); err != nil {
		return nil, err
	}
	s.dispatchPendingConversationsIfEligible(item)
	return item, nil
}

func (s *agentProfileService) UpdateAgentProfile(req request.UpdateAgentProfileRequest, operator *dto.AuthPrincipal) error {
	if operator == nil {
		return errorsx.Unauthorized("未登录或登录已过期")
	}
	current := s.Get(req.ID)
	if current == nil {
		return errorsx.InvalidParam("客服档案不存在")
	}
	item, err := s.buildProfileModel(req.ID, req.CreateAgentProfileRequest)
	if err != nil {
		return err
	}
	if err := repositories.AgentProfileRepository.Updates(sqls.DB(), req.ID, map[string]any{
		"user_id":                 item.UserID,
		"team_id":                 item.TeamID,
		"agent_code":              item.AgentCode,
		"display_name":            item.DisplayName,
		"avatar":                  item.Avatar,
		"service_status":          item.ServiceStatus,
		"max_concurrent_count":    item.MaxConcurrentCount,
		"priority_level":          item.PriorityLevel,
		"auto_assign_enabled":     item.AutoAssignEnabled,
		"receive_offline_message": item.ReceiveOfflineMessage,
		"remark":                  item.Remark,
		"update_user_id":          operator.UserID,
		"update_user_name":        operator.Username,
		"updated_at":              time.Now(),
	}); err != nil {
		return err
	}
	s.dispatchPendingConversationsIfEligible(item)
	return nil
}

func (s *agentProfileService) DeleteAgentProfile(id int64) error {
	current := s.Get(id)
	if current == nil {
		return errorsx.InvalidParam("客服档案不存在")
	}
	repositories.AgentProfileRepository.Delete(sqls.DB(), id)
	return nil
}

func (s *agentProfileService) buildProfileModel(id int64, req request.CreateAgentProfileRequest) (*models.AgentProfile, error) {
	if req.UserID <= 0 {
		return nil, errorsx.InvalidParam("请选择关联用户")
	}
	if UserService.Get(req.UserID) == nil {
		return nil, errorsx.InvalidParam("关联用户不存在")
	}
	if req.TeamID <= 0 {
		return nil, errorsx.InvalidParam("请选择所属客服组")
	}
	if AgentTeamService.Get(req.TeamID) == nil {
		return nil, errorsx.InvalidParam("所属客服组不存在")
	}
	req.AgentCode = strings.TrimSpace(req.AgentCode)
	req.DisplayName = strings.TrimSpace(req.DisplayName)
	if req.AgentCode == "" || req.DisplayName == "" {
		return nil, errorsx.InvalidParam("客服工号和展示名不能为空")
	}
	if exists := s.Take("user_id = ? AND id <> ?", req.UserID, id); exists != nil {
		return nil, errorsx.InvalidParam("该用户已存在客服档案")
	}
	if exists := s.Take("agent_code = ? AND id <> ?", req.AgentCode, id); exists != nil {
		return nil, errorsx.InvalidParam("客服工号已存在")
	}
	if !enums.IsValidServiceStatus(req.ServiceStatus) {
		return nil, errorsx.InvalidParam("客服状态不合法")
	}
	if req.MaxConcurrentCount < 0 {
		return nil, errorsx.InvalidParam("最大并发接待数不能小于 0")
	}
	return &models.AgentProfile{
		UserID:                req.UserID,
		TeamID:                req.TeamID,
		AgentCode:             req.AgentCode,
		DisplayName:           req.DisplayName,
		Avatar:                strings.TrimSpace(req.Avatar),
		ServiceStatus:         req.ServiceStatus,
		MaxConcurrentCount:    req.MaxConcurrentCount,
		PriorityLevel:         req.PriorityLevel,
		AutoAssignEnabled:     req.AutoAssignEnabled,
		ReceiveOfflineMessage: req.ReceiveOfflineMessage,
		Remark:                strings.TrimSpace(req.Remark),
	}, nil
}

func (s *agentProfileService) dispatchPendingConversationsIfEligible(item *models.AgentProfile) {
	if item == nil {
		return
	}
	if item.Status != enums.StatusOk {
		return
	}
	if !item.AutoAssignEnabled || item.MaxConcurrentCount <= 0 {
		return
	}
	if item.ServiceStatus != enums.ServiceStatusIdle {
		return
	}
	_, _ = ConversationDispatchService.DispatchPendingConversations(0)
}
