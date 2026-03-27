package services

import (
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/dto"
	"cs-agent/internal/pkg/dto/request"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/pkg/errorsx"
	"cs-agent/internal/pkg/utils"
	"cs-agent/internal/repositories"
	"slices"
	"strings"
	"time"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
)

var AgentTeamScheduleService = newAgentTeamScheduleService()

func newAgentTeamScheduleService() *agentTeamScheduleService {
	return &agentTeamScheduleService{}
}

type agentTeamScheduleService struct {
}

func (s *agentTeamScheduleService) Get(id int64) *models.AgentTeamSchedule {
	return repositories.AgentTeamScheduleRepository.Get(sqls.DB(), id)
}

func (s *agentTeamScheduleService) Take(where ...interface{}) *models.AgentTeamSchedule {
	return repositories.AgentTeamScheduleRepository.Take(sqls.DB(), where...)
}

func (s *agentTeamScheduleService) Find(cnd *sqls.Cnd) []models.AgentTeamSchedule {
	return repositories.AgentTeamScheduleRepository.Find(sqls.DB(), cnd)
}

func (s *agentTeamScheduleService) FindOne(cnd *sqls.Cnd) *models.AgentTeamSchedule {
	return repositories.AgentTeamScheduleRepository.FindOne(sqls.DB(), cnd)
}

func (s *agentTeamScheduleService) FindPageByParams(params *params.QueryParams) (list []models.AgentTeamSchedule, paging *sqls.Paging) {
	return repositories.AgentTeamScheduleRepository.FindPageByParams(sqls.DB(), params)
}

func (s *agentTeamScheduleService) FindPageByCnd(cnd *sqls.Cnd) (list []models.AgentTeamSchedule, paging *sqls.Paging) {
	return repositories.AgentTeamScheduleRepository.FindPageByCnd(sqls.DB(), cnd)
}

func (s *agentTeamScheduleService) Count(cnd *sqls.Cnd) int64 {
	return repositories.AgentTeamScheduleRepository.Count(sqls.DB(), cnd)
}

func (s *agentTeamScheduleService) Create(t *models.AgentTeamSchedule) error {
	return repositories.AgentTeamScheduleRepository.Create(sqls.DB(), t)
}

func (s *agentTeamScheduleService) Update(t *models.AgentTeamSchedule) error {
	return repositories.AgentTeamScheduleRepository.Update(sqls.DB(), t)
}

func (s *agentTeamScheduleService) Updates(id int64, columns map[string]interface{}) error {
	return repositories.AgentTeamScheduleRepository.Updates(sqls.DB(), id, columns)
}

func (s *agentTeamScheduleService) UpdateColumn(id int64, name string, value interface{}) error {
	return repositories.AgentTeamScheduleRepository.UpdateColumn(sqls.DB(), id, name, value)
}

func (s *agentTeamScheduleService) Delete(id int64) {
	repositories.AgentTeamScheduleRepository.Delete(sqls.DB(), id)
}

func (s *agentTeamScheduleService) CreateAgentTeamSchedule(req request.CreateAgentTeamScheduleRequest, operator *dto.AuthPrincipal) (*models.AgentTeamSchedule, error) {
	if operator == nil {
		return nil, errorsx.Unauthorized("未登录或登录已过期")
	}
	item, err := s.buildScheduleModel(0, req.TeamID, req.StartAt, req.EndAt, req.SourceType, req.Remark)
	if err != nil {
		return nil, err
	}
	item.AuditFields = utils.BuildAuditFields(operator)
	if err := sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		return ctx.Tx.Create(item).Error
	}); err != nil {
		return nil, err
	}
	s.dispatchPendingConversationsIfActive(item)
	return item, nil
}

func (s *agentTeamScheduleService) UpdateAgentTeamSchedule(req request.UpdateAgentTeamScheduleRequest, operator *dto.AuthPrincipal) error {
	if operator == nil {
		return errorsx.Unauthorized("未登录或登录已过期")
	}
	if s.Get(req.ID) == nil {
		return errorsx.InvalidParam("客服组排班不存在")
	}
	item, err := s.buildScheduleModel(req.ID, req.TeamID, req.StartAt, req.EndAt, req.SourceType, req.Remark)
	if err != nil {
		return err
	}
	now := time.Now()
	if err := sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		return repositories.AgentTeamScheduleRepository.Updates(ctx.Tx, req.ID, map[string]any{
			"team_id":          item.TeamID,
			"start_at":         item.StartAt,
			"end_at":           item.EndAt,
			"source_type":      item.SourceType,
			"remark":           item.Remark,
			"update_user_id":   operator.UserID,
			"update_user_name": operator.Username,
			"updated_at":       now,
		})
	}); err != nil {
		return err
	}
	s.dispatchPendingConversationsIfActive(item)
	return nil
}

func (s *agentTeamScheduleService) DeleteAgentTeamSchedule(id int64) error {
	if s.Get(id) == nil {
		return errorsx.InvalidParam("客服组排班不存在")
	}
	return sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		return ctx.Tx.Delete(&models.AgentTeamSchedule{}, "id = ?", id).Error
	})
}

func (s *agentTeamScheduleService) buildScheduleModel(id, teamID int64, startAt, endAt, sourceType, remark string) (*models.AgentTeamSchedule, error) {
	if teamID <= 0 {
		return nil, errorsx.InvalidParam("请选择客服组")
	}
	team := AgentTeamService.Get(teamID)
	if team == nil {
		return nil, errorsx.InvalidParam("客服组不存在")
	}
	if !slices.Contains(enums.StatusValues, team.Status) {
		return nil, errorsx.InvalidParam("客服组状态不合法")
	}
	sourceType = strings.TrimSpace(sourceType)
	if sourceType == "" {
		return nil, errorsx.InvalidParam("排班来源不能为空")
	}
	startAtValue, err := parseRequiredDateTime(startAt, "开始时间格式错误")
	if err != nil {
		return nil, err
	}
	endAtValue, err := parseRequiredDateTime(endAt, "结束时间格式错误")
	if err != nil {
		return nil, err
	}
	if !endAtValue.After(startAtValue) {
		return nil, errorsx.InvalidParam("结束时间必须晚于开始时间")
	}
	var count int64
	sqls.DB().Model(&models.AgentTeamSchedule{}).
		Where("team_id = ? AND id <> ? AND start_at < ? AND end_at > ?", teamID, id, endAtValue, startAtValue).
		Count(&count)
	if count > 0 {
		return nil, errorsx.InvalidParam("该客服组在所选时间段已存在排班")
	}
	return &models.AgentTeamSchedule{
		TeamID:     teamID,
		StartAt:    startAtValue,
		EndAt:      endAtValue,
		SourceType: sourceType,
		Remark:     strings.TrimSpace(remark),
	}, nil
}

func parseRequiredDateTime(value, message string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, errorsx.InvalidParam(message)
	}
	ret, err := parseDateTimeValue(value)
	if err != nil {
		return time.Time{}, errorsx.InvalidParam(message + "，请使用 yyyy-MM-dd HH:mm:ss 或 RFC3339")
	}
	return ret, nil
}

func parseDateTimeValue(value string) (time.Time, error) {
	layouts := []string{
		time.DateTime,
		time.RFC3339,
		"2006-01-02T15:04",
		"2006-01-02T15:04:05",
	}
	for _, layout := range layouts {
		if ret, err := time.ParseInLocation(layout, value, time.Local); err == nil {
			return ret, nil
		}
	}
	return time.Time{}, errorsx.InvalidParam("时间格式错误")
}

func (s *agentTeamScheduleService) dispatchPendingConversationsIfActive(item *models.AgentTeamSchedule) {
	if item == nil {
		return
	}
	if item.Status != enums.StatusOk {
		return
	}
	now := time.Now()
	if item.StartAt.After(now) || !item.EndAt.After(now) {
		return
	}
	_, _ = ConversationDispatchService.DispatchPendingConversations(0)
}
