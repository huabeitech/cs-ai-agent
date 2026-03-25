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

var AgentTeamService = newAgentTeamService()

func newAgentTeamService() *agentTeamService {
	return &agentTeamService{}
}

type agentTeamService struct {
}

func (s *agentTeamService) Get(id int64) *models.AgentTeam {
	return repositories.AgentTeamRepository.Get(sqls.DB(), id)
}

func (s *agentTeamService) Take(where ...interface{}) *models.AgentTeam {
	return repositories.AgentTeamRepository.Take(sqls.DB(), where...)
}

func (s *agentTeamService) Find(cnd *sqls.Cnd) []models.AgentTeam {
	return repositories.AgentTeamRepository.Find(sqls.DB(), cnd)
}

func (s *agentTeamService) FindOne(cnd *sqls.Cnd) *models.AgentTeam {
	return repositories.AgentTeamRepository.FindOne(sqls.DB(), cnd)
}

func (s *agentTeamService) FindPageByParams(params *params.QueryParams) (list []models.AgentTeam, paging *sqls.Paging) {
	return repositories.AgentTeamRepository.FindPageByParams(sqls.DB(), params)
}

func (s *agentTeamService) FindPageByCnd(cnd *sqls.Cnd) (list []models.AgentTeam, paging *sqls.Paging) {
	return repositories.AgentTeamRepository.FindPageByCnd(sqls.DB(), cnd)
}

func (s *agentTeamService) Count(cnd *sqls.Cnd) int64 {
	return repositories.AgentTeamRepository.Count(sqls.DB(), cnd)
}

func (s *agentTeamService) FindByIds(ids []int64) []models.AgentTeam {
	return repositories.AgentTeamRepository.FindByIds(sqls.DB(), ids)
}

func (s *agentTeamService) Create(t *models.AgentTeam) error {
	return repositories.AgentTeamRepository.Create(sqls.DB(), t)
}

func (s *agentTeamService) Update(t *models.AgentTeam) error {
	return repositories.AgentTeamRepository.Update(sqls.DB(), t)
}

func (s *agentTeamService) Updates(id int64, columns map[string]interface{}) error {
	return repositories.AgentTeamRepository.Updates(sqls.DB(), id, columns)
}

func (s *agentTeamService) UpdateColumn(id int64, name string, value interface{}) error {
	return repositories.AgentTeamRepository.UpdateColumn(sqls.DB(), id, name, value)
}

func (s *agentTeamService) Delete(id int64) {
	repositories.AgentTeamRepository.Delete(sqls.DB(), id)
}

func (s *agentTeamService) CreateAgentTeam(req request.CreateAgentTeamRequest, operator *dto.AuthPrincipal) (*models.AgentTeam, error) {
	if operator == nil {
		return nil, errorsx.Unauthorized("未登录或登录已过期")
	}
	item, err := s.buildTeamModel(0, req.Name, req.LeaderUserID, req.Status, req.Description, req.Remark)
	if err != nil {
		return nil, err
	}
	item.AuditFields = utils.BuildAuditFields(operator)
	if err := sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		return ctx.Tx.Create(item).Error
	}); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *agentTeamService) UpdateAgentTeam(req request.UpdateAgentTeamRequest, operator *dto.AuthPrincipal) error {
	if operator == nil {
		return errorsx.Unauthorized("未登录或登录已过期")
	}
	current := s.Get(req.ID)
	if current == nil || current.Status == enums.StatusDeleted {
		return errorsx.InvalidParam("客服组不存在")
	}
	item, err := s.buildTeamModel(req.ID, req.Name, req.LeaderUserID, req.Status, req.Description, req.Remark)
	if err != nil {
		return err
	}
	now := time.Now()
	return sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		return ctx.Tx.Model(&models.AgentTeam{}).Where("id = ?", req.ID).Updates(map[string]any{
			"name":             item.Name,
			"leader_user_id":   item.LeaderUserID,
			"status":           item.Status,
			"description":      item.Description,
			"remark":           item.Remark,
			"update_user_id":   operator.UserID,
			"update_user_name": operator.Username,
			"updated_at":       now,
		}).Error
	})
}

func (s *agentTeamService) DeleteAgentTeam(id int64, operator *dto.AuthPrincipal) error {
	if operator == nil {
		return errorsx.Unauthorized("未登录或登录已过期")
	}
	current := s.Get(id)
	if current == nil || current.Status == enums.StatusDeleted {
		return errorsx.InvalidParam("客服组不存在")
	}
	if AgentProfileService.Take("team_id = ?", id) != nil {
		return errorsx.Forbidden("客服组下仍有关联客服档案，无法删除")
	}
	if AgentTeamScheduleService.Take("team_id = ?", id) != nil {
		return errorsx.Forbidden("客服组下仍有关联组排班，无法删除")
	}
	if AIAgentService.Take(
		"(team_ids = ? OR team_ids LIKE ? OR team_ids LIKE ? OR team_ids LIKE ?) AND status <> ?",
		utils.JoinInt64s([]int64{id}),
		utils.JoinInt64s([]int64{id})+",%",
		"%,"+utils.JoinInt64s([]int64{id}),
		"%,"+utils.JoinInt64s([]int64{id})+",%",
		enums.StatusDeleted,
	) != nil {
		return errorsx.Forbidden("客服组下仍有关联 AI Agent，无法删除")
	}
	now := time.Now()
	return sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		return ctx.Tx.Model(&models.AgentTeam{}).Where("id = ?", id).Updates(map[string]any{
			"status":           enums.StatusDeleted,
			"update_user_id":   operator.UserID,
			"update_user_name": operator.Username,
			"updated_at":       now,
		}).Error
	})
}

func (s *agentTeamService) buildTeamModel(id int64, name string, leaderUserID int64, status int, description, remark string) (*models.AgentTeam, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errorsx.InvalidParam("客服组名称不能为空")
	}
	if exists := s.Take("name = ? AND status <> ? AND id <> ?", name, enums.StatusDeleted, id); exists != nil {
		return nil, errorsx.InvalidParam("客服组名称已存在")
	}
	if leaderUserID > 0 && UserService.Get(leaderUserID) == nil {
		return nil, errorsx.InvalidParam("组长用户不存在")
	}
	if status != 0 && status != 1 {
		return nil, errorsx.InvalidParam("客服组状态不合法")
	}
	return &models.AgentTeam{
		Name:         name,
		LeaderUserID: leaderUserID,
		Status:       enums.Status(status),
		Description:  strings.TrimSpace(description),
		Remark:       strings.TrimSpace(remark),
	}, nil
}
