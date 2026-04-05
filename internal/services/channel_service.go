package services

import (
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/dto"
	"cs-agent/internal/pkg/dto/request"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/pkg/errorsx"
	"cs-agent/internal/pkg/utils"
	"cs-agent/internal/repositories"
	"encoding/json"
	"strings"
	"time"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
)

var ChannelService = newChannelService()

func newChannelService() *channelService {
	return &channelService{}
}

type channelService struct {
}

type WxWorkKFChannelConfig struct {
	OpenKfID string `json:"openKfId"`
}

func (s *channelService) Get(id int64) *models.Channel {
	return repositories.ChannelRepository.Get(sqls.DB(), id)
}

func (s *channelService) Take(where ...interface{}) *models.Channel {
	return repositories.ChannelRepository.Take(sqls.DB(), where...)
}

func (s *channelService) Find(cnd *sqls.Cnd) []models.Channel {
	return repositories.ChannelRepository.Find(sqls.DB(), cnd)
}

func (s *channelService) FindOne(cnd *sqls.Cnd) *models.Channel {
	return repositories.ChannelRepository.FindOne(sqls.DB(), cnd)
}

func (s *channelService) FindPageByParams(params *params.QueryParams) (list []models.Channel, paging *sqls.Paging) {
	return repositories.ChannelRepository.FindPageByParams(sqls.DB(), params)
}

func (s *channelService) FindPageByCnd(cnd *sqls.Cnd) (list []models.Channel, paging *sqls.Paging) {
	return repositories.ChannelRepository.FindPageByCnd(sqls.DB(), cnd)
}

func (s *channelService) Count(cnd *sqls.Cnd) int64 {
	return repositories.ChannelRepository.Count(sqls.DB(), cnd)
}

func (s *channelService) Create(t *models.Channel) error {
	return repositories.ChannelRepository.Create(sqls.DB(), t)
}

func (s *channelService) Update(t *models.Channel) error {
	return repositories.ChannelRepository.Update(sqls.DB(), t)
}

func (s *channelService) Updates(id int64, columns map[string]interface{}) error {
	return repositories.ChannelRepository.Updates(sqls.DB(), id, columns)
}

func (s *channelService) UpdateColumn(id int64, name string, value interface{}) error {
	return repositories.ChannelRepository.UpdateColumn(sqls.DB(), id, name, value)
}

func (s *channelService) CreateChannel(req request.CreateChannelRequest, operator *dto.AuthPrincipal) (*models.Channel, error) {
	if operator == nil {
		return nil, errorsx.Unauthorized("未登录或登录已过期")
	}
	item, err := s.buildChannelModel(0, req)
	if err != nil {
		return nil, err
	}
	item.AuditFields = utils.BuildAuditFields(operator)
	if err := repositories.ChannelRepository.Create(sqls.DB(), item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *channelService) UpdateChannel(req request.UpdateChannelRequest, operator *dto.AuthPrincipal) error {
	if operator == nil {
		return errorsx.Unauthorized("未登录或登录已过期")
	}
	current := s.Get(req.ID)
	if current == nil || current.Status == enums.StatusDeleted {
		return errorsx.InvalidParam("接入渠道不存在")
	}
	item, err := s.buildChannelModel(req.ID, req.CreateChannelRequest)
	if err != nil {
		return err
	}
	return repositories.ChannelRepository.Updates(sqls.DB(), req.ID, map[string]any{
		"channel_type":     item.ChannelType,
		"ai_agent_id":      item.AIAgentID,
		"name":             item.Name,
		"app_id":           item.AppID,
		"config_json":      item.ConfigJSON,
		"status":           item.Status,
		"remark":           item.Remark,
		"update_user_id":   operator.UserID,
		"update_user_name": operator.Username,
		"updated_at":       time.Now(),
	})
}

func (s *channelService) UpdateStatus(id int64, status int, operator *dto.AuthPrincipal) error {
	if operator == nil {
		return errorsx.Unauthorized("未登录或登录已过期")
	}
	item := s.Get(id)
	if item == nil || item.Status == enums.StatusDeleted {
		return errorsx.InvalidParam("接入渠道不存在")
	}
	if status != int(enums.StatusOk) && status != int(enums.StatusDisabled) {
		return errorsx.InvalidParam("状态值不合法")
	}
	return s.Updates(id, map[string]any{
		"status":           status,
		"update_user_id":   operator.UserID,
		"update_user_name": operator.Username,
		"updated_at":       time.Now(),
	})
}

func (s *channelService) DeleteChannel(id int64, operator *dto.AuthPrincipal) error {
	if operator == nil {
		return errorsx.Unauthorized("未登录或登录已过期")
	}
	item := s.Get(id)
	if item == nil || item.Status == enums.StatusDeleted {
		return errorsx.InvalidParam("接入渠道不存在")
	}
	return s.Updates(id, map[string]any{
		"status":           enums.StatusDeleted,
		"update_user_id":   operator.UserID,
		"update_user_name": operator.Username,
		"updated_at":       time.Now(),
	})
}

func (s *channelService) ParseWxWorkKFChannelConfig(raw string) (*WxWorkKFChannelConfig, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return &WxWorkKFChannelConfig{}, nil
	}
	cfg := &WxWorkKFChannelConfig{}
	if err := json.Unmarshal([]byte(raw), cfg); err != nil {
		return nil, err
	}
	cfg.OpenKfID = strings.TrimSpace(cfg.OpenKfID)
	return cfg, nil
}

func (s *channelService) GetEnabledWxWorkKFChannelByOpenKfID(openKfID string) *models.Channel {
	openKfID = strings.TrimSpace(openKfID)
	if openKfID == "" {
		return nil
	}
	channels := s.Find(sqls.NewCnd().
		Eq("channel_type", enums.ChannelTypeWxWorkKF).
		Eq("status", enums.StatusOk).
		Asc("id"))
	for i := range channels {
		cfg, err := s.ParseWxWorkKFChannelConfig(channels[i].ConfigJSON)
		if err != nil {
			continue
		}
		if cfg != nil && cfg.OpenKfID == openKfID {
			return &channels[i]
		}
	}
	return nil
}

func (s *channelService) GetEnabledWebChannelByAppID(appID string) *models.Channel {
	appID = strings.TrimSpace(appID)
	if appID == "" {
		return nil
	}
	return s.Take("channel_type = ? AND app_id = ? AND status = ?", enums.ChannelTypeWeb, appID, enums.StatusOk)
}

func (s *channelService) buildChannelModel(id int64, req request.CreateChannelRequest) (*models.Channel, error) {
	channelType := strings.TrimSpace(req.ChannelType)
	if channelType != enums.ChannelTypeWeb && channelType != enums.ChannelTypeWxWorkKF {
		return nil, errorsx.InvalidParam("渠道类型不合法")
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, errorsx.InvalidParam("渠道名称不能为空")
	}
	if req.AIAgentID <= 0 {
		return nil, errorsx.InvalidParam("请选择 AI Agent")
	}
	aiAgent := AIAgentService.Get(req.AIAgentID)
	if aiAgent == nil || aiAgent.Status != enums.StatusOk {
		return nil, errorsx.InvalidParam("AI Agent 不存在或未启用")
	}
	status := enums.Status(req.Status)
	if req.Status == 0 {
		status = enums.StatusOk
	}
	if status != enums.StatusOk && status != enums.StatusDisabled {
		return nil, errorsx.InvalidParam("渠道状态不合法")
	}

	appID := strings.TrimSpace(req.AppID)
	configJSON := strings.TrimSpace(req.ConfigJSON)
	switch channelType {
	case enums.ChannelTypeWeb:
		if appID == "" {
			appID = strings.TrimSpace(appID)
		}
		if appID == "" {
			return nil, errorsx.InvalidParam("web 渠道 AppID 不能为空")
		}
		if exists := s.Take("app_id = ? AND channel_type = ? AND status <> ? AND id <> ?", appID, enums.ChannelTypeWeb, enums.StatusDeleted, id); exists != nil {
			return nil, errorsx.InvalidParam("web 渠道 AppID 已存在")
		}
	case enums.ChannelTypeWxWorkKF:
		appID = ""
		cfg, err := s.ParseWxWorkKFChannelConfig(configJSON)
		if err != nil {
			return nil, errorsx.InvalidParam("企业微信渠道配置不合法")
		}
		if cfg == nil || cfg.OpenKfID == "" {
			return nil, errorsx.InvalidParam("企业微信渠道配置缺少 openKfId")
		}
		if channel := s.GetEnabledWxWorkKFChannelByOpenKfID(cfg.OpenKfID); channel != nil && channel.ID != id {
			return nil, errorsx.InvalidParam("openKfId 已被其他渠道使用")
		}
	}

	return &models.Channel{
		ChannelType: channelType,
		AIAgentID:   req.AIAgentID,
		Name:        name,
		AppID:       appID,
		ConfigJSON:  configJSON,
		Status:      status,
		Remark:      strings.TrimSpace(req.Remark),
	}, nil
}
