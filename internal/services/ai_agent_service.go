package services

import (
	"slices"
	"strings"
	"time"

	"cs-agent/internal/models"
	"cs-agent/internal/pkg/dto"
	"cs-agent/internal/pkg/dto/request"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/pkg/errorsx"
	"cs-agent/internal/pkg/utils"
	"cs-agent/internal/repositories"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
)

var AIAgentService = newAIAgentService()

func newAIAgentService() *aIAgentService {
	return &aIAgentService{}
}

type aIAgentService struct {
}

func (s *aIAgentService) Get(id int64) *models.AIAgent {
	return repositories.AIAgentRepository.Get(sqls.DB(), id)
}

func (s *aIAgentService) Take(where ...interface{}) *models.AIAgent {
	return repositories.AIAgentRepository.Take(sqls.DB(), where...)
}

func (s *aIAgentService) Find(cnd *sqls.Cnd) []models.AIAgent {
	return repositories.AIAgentRepository.Find(sqls.DB(), cnd)
}

func (s *aIAgentService) FindOne(cnd *sqls.Cnd) *models.AIAgent {
	return repositories.AIAgentRepository.FindOne(sqls.DB(), cnd)
}

func (s *aIAgentService) FindPageByParams(params *params.QueryParams) (list []models.AIAgent, paging *sqls.Paging) {
	return repositories.AIAgentRepository.FindPageByParams(sqls.DB(), params)
}

func (s *aIAgentService) FindPageByCnd(cnd *sqls.Cnd) (list []models.AIAgent, paging *sqls.Paging) {
	return repositories.AIAgentRepository.FindPageByCnd(sqls.DB(), cnd)
}

func (s *aIAgentService) Count(cnd *sqls.Cnd) int64 {
	return repositories.AIAgentRepository.Count(sqls.DB(), cnd)
}

func (s *aIAgentService) CreateAIAgent(req request.CreateAIAgentRequest, operator *dto.AuthPrincipal) (*models.AIAgent, error) {
	if operator == nil {
		return nil, errorsx.Unauthorized("未登录或登录已过期")
	}
	item, err := s.buildAIAgentModel(0, req)
	if err != nil {
		return nil, err
	}
	item.Status = enums.StatusOk
	item.SortNo = 0
	item.AuditFields = utils.BuildAuditFields(operator)
	if err := repositories.AIAgentRepository.Create(sqls.DB(), item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *aIAgentService) UpdateAIAgent(req request.UpdateAIAgentRequest, operator *dto.AuthPrincipal) error {
	if operator == nil {
		return errorsx.Unauthorized("未登录或登录已过期")
	}
	if s.Get(req.ID) == nil {
		return errorsx.InvalidParam("AI Agent 不存在")
	}
	item, err := s.buildAIAgentModel(req.ID, req.CreateAIAgentRequest)
	if err != nil {
		return err
	}
	return repositories.AIAgentRepository.Updates(sqls.DB(), req.ID, map[string]any{
		"name":                item.Name,
		"description":         item.Description,
		"ai_config_id":        item.AIConfigID,
		"service_mode":        item.ServiceMode,
		"system_prompt":       item.SystemPrompt,
		"welcome_message":     item.WelcomeMessage,
		"team_ids":            item.TeamIDs,
		"handoff_mode":        item.HandoffMode,
		"max_ai_reply_rounds": item.MaxAIReplyRounds,
		"fallback_mode":       item.FallbackMode,
		"knowledge_ids":       item.KnowledgeIDs,
		"remark":              item.Remark,
		"update_user_id":      operator.UserID,
		"update_user_name":    operator.Username,
		"updated_at":          time.Now(),
	})
}

func (s *aIAgentService) DeleteAIAgent(id int64, operator *dto.AuthPrincipal) error {
	current := s.Get(id)
	if current == nil {
		return errorsx.InvalidParam("AI Agent 不存在")
	}
	if WidgetSiteService.Take("ai_agent_id = ?", id) != nil {
		return errorsx.Forbidden("已有接入站点绑定该 AI Agent，无法删除")
	}
	return repositories.AIAgentRepository.Updates(sqls.DB(), id, map[string]any{
		"status":           enums.StatusDeleted,
		"update_user_id":   operator.UserID,
		"update_user_name": operator.Username,
		"updated_at":       time.Now(),
	})
}

func (s *aIAgentService) buildAIAgentModel(id int64, req request.CreateAIAgentRequest) (*models.AIAgent, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, errorsx.InvalidParam("AI Agent 名称不能为空")
	}
	if exists := s.Take("name = ? AND id <> ?", name, id); exists != nil {
		return nil, errorsx.InvalidParam("AI Agent 名称已存在")
	}
	if req.AIConfigID <= 0 {
		return nil, errorsx.InvalidParam("AI 配置不能为空")
	}
	aiConfig := AIConfigService.Get(req.AIConfigID)
	if aiConfig == nil {
		return nil, errorsx.InvalidParam("AI 配置不存在")
	}
	if aiConfig.Status != enums.StatusOk {
		return nil, errorsx.InvalidParam("AI 配置未启用")
	}
	if !slices.Contains(enums.IMConversationServiceModeValues, req.ServiceMode) {
		return nil, errorsx.InvalidParam("服务模式不合法")
	}
	teamIDs, err := s.normalizeTeamIDs(req.TeamIDs)
	if err != nil {
		return nil, err
	}

	if !slices.Contains(enums.AIAgentHandoffModeValues, enums.AIAgentHandoffMode(req.HandoffMode)) {
		return nil, errorsx.InvalidParam("转人工模式不合法")
	}
	if enums.AIAgentHandoffMode(req.HandoffMode) == enums.AIAgentHandoffModeDefaultTeamPool && len(teamIDs) == 0 {
		return nil, errorsx.InvalidParam("默认客服组待接入池模式必须至少选择一个客服组")
	}

	if !slices.Contains(enums.AIAgentFallbackModeValues, enums.AIAgentFallbackMode(req.FallbackMode)) {
		return nil, errorsx.InvalidParam("兜底模式不合法")
	}

	knowledgeIDs, err := s.normalizeKnowledgeIDs(req.KnowledgeIDs)
	if err != nil {
		return nil, err
	}
	if len(knowledgeIDs) == 0 {
		return nil, errorsx.InvalidParam("请至少选择一个知识库")
	}
	return &models.AIAgent{
		Name:             name,
		Description:      strings.TrimSpace(req.Description),
		AIConfigID:       req.AIConfigID,
		ServiceMode:      req.ServiceMode,
		SystemPrompt:     strings.TrimSpace(req.SystemPrompt),
		WelcomeMessage:   strings.TrimSpace(req.WelcomeMessage),
		TeamIDs:          utils.JoinInt64s(teamIDs),
		HandoffMode:      req.HandoffMode,
		MaxAIReplyRounds: req.MaxAIReplyRounds,
		FallbackMode:     req.FallbackMode,
		KnowledgeIDs:     utils.JoinInt64s(knowledgeIDs),
		Remark:           strings.TrimSpace(req.Remark),
	}, nil
}

func (s *aIAgentService) normalizeTeamIDs(input []int64) ([]int64, error) {
	ret := make([]int64, 0, len(input))
	seen := make(map[int64]struct{})
	for _, id := range input {
		if id <= 0 {
			continue
		}
		if _, exists := seen[id]; exists {
			continue
		}
		team := AgentTeamService.Get(id)
		if team == nil {
			return nil, errorsx.InvalidParam("客服组不存在")
		}
		if team.Status != enums.StatusOk {
			return nil, errorsx.InvalidParam("客服组未启用")
		}
		seen[id] = struct{}{}
		ret = append(ret, id)
	}
	slices.Sort(ret)
	return ret, nil
}

func (s *aIAgentService) normalizeKnowledgeIDs(input []int64) ([]int64, error) {
	ret := make([]int64, 0, len(input))
	seen := make(map[int64]struct{})
	for _, id := range input {
		if id <= 0 {
			continue
		}
		if _, exists := seen[id]; exists {
			continue
		}
		kb := KnowledgeBaseService.Get(id)
		if kb == nil {
			return nil, errorsx.InvalidParam("知识库不存在")
		}
		if kb.Status != enums.StatusOk {
			return nil, errorsx.InvalidParam("知识库未启用")
		}
		seen[id] = struct{}{}
		ret = append(ret, id)
	}
	return ret, nil
}

func (s *aIAgentService) UpdateSort(ids []int64) error {
	return sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		for i, id := range ids {
			if err := repositories.AIAgentRepository.UpdateColumn(ctx.Tx, id, "sort_no", i+1); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *aIAgentService) UpdateStatus(id int64, status int, operator *dto.AuthPrincipal) error {
	if operator == nil {
		return errorsx.Unauthorized("未登录或登录已过期")
	}
	current := s.Get(id)
	if current == nil {
		return errorsx.InvalidParam("AI Agent 不存在")
	}
	if status != int(enums.StatusOk) && status != int(enums.StatusDisabled) {
		return errorsx.InvalidParam("状态值不合法")
	}

	return repositories.AIAgentRepository.Updates(sqls.DB(), id, map[string]any{
		"status":           status,
		"update_user_id":   operator.UserID,
		"update_user_name": operator.Username,
		"updated_at":       time.Now(),
	})
}
