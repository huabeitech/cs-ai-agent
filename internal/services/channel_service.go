package services

import (
	"crypto/rand"
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/dto"
	"cs-agent/internal/pkg/dto/request"
	"cs-agent/internal/pkg/dto/response"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/pkg/errorsx"
	"cs-agent/internal/pkg/irisx"
	"cs-agent/internal/pkg/utils"
	"cs-agent/internal/repositories"
	"cs-agent/internal/wxwork"
	"encoding/base64"
	"encoding/json"
	"strings"
	"time"

	"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple/common/strs"
	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
	"github.com/silenceper/wechat/v2/work/kf"
)

var ChannelService = newChannelService()

func newChannelService() *channelService {
	return &channelService{}
}

type channelService struct {
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
		"channel_id":       item.ChannelID,
		"ai_agent_id":      item.AIAgentID,
		"name":             item.Name,
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

func (s *channelService) ParseWxWorkKFChannelConfig(raw string) (*dto.WxWorkKFChannelConfig, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return &dto.WxWorkKFChannelConfig{}, nil
	}
	cfg := &dto.WxWorkKFChannelConfig{}
	if err := json.Unmarshal([]byte(raw), cfg); err != nil {
		return nil, err
	}
	cfg.OpenKfID = strings.TrimSpace(cfg.OpenKfID)
	return cfg, nil
}

func (s *channelService) ListWxWorkKFAccounts() ([]response.WxWorkKFAccountResponse, error) {
	if !wxwork.Enabled() || wxwork.GetWorkCli() == nil {
		return nil, errorsx.BusinessError(1, "企业微信未启用或配置不完整")
	}
	cli, err := wxwork.GetWorkCli().GetKF()
	if err != nil {
		return nil, err
	}

	const limit = 100
	accounts := make([]response.WxWorkKFAccountResponse, 0)
	for offset := 0; ; offset += limit {
		result, err := cli.AccountPaging(&kf.AccountPagingRequest{
			Offset: offset,
			Limit:  limit,
		})
		if err != nil {
			return nil, err
		}
		for _, item := range result.AccountList {
			openKfID := strings.TrimSpace(item.OpenKFID)
			if openKfID == "" {
				continue
			}
			accounts = append(accounts, response.WxWorkKFAccountResponse{
				OpenKfID:        openKfID,
				Name:            strings.TrimSpace(item.Name),
				Avatar:          strings.TrimSpace(item.Avatar),
				ManagePrivilege: item.ManagePrivilege,
			})
		}
		if len(result.AccountList) < limit {
			break
		}
	}
	return accounts, nil
}

func (s *channelService) ParseWebChannelConfig(raw string) (*dto.WebChannelConfig, error) {
	raw = strings.TrimSpace(raw)
	cfg := &dto.WebChannelConfig{
		Title:      "在线客服",
		Subtitle:   "欢迎咨询",
		ThemeColor: "#2563eb",
		Position:   "right",
		Width:      "380px",
	}
	if raw != "" {
		if err := json.Unmarshal([]byte(raw), cfg); err != nil {
			return nil, err
		}
	}
	cfg.Title = strings.TrimSpace(cfg.Title)
	if cfg.Title == "" {
		cfg.Title = "在线客服"
	}
	cfg.Subtitle = strings.TrimSpace(cfg.Subtitle)
	cfg.ThemeColor = strings.TrimSpace(cfg.ThemeColor)
	if cfg.ThemeColor == "" {
		cfg.ThemeColor = "#2563eb"
	}
	cfg.Position = strings.TrimSpace(cfg.Position)
	if cfg.Position == "" {
		cfg.Position = "right"
	}
	if cfg.Position != "left" && cfg.Position != "right" {
		return nil, errorsx.InvalidParam("Web渠道配置 position 只能为 left 或 right")
	}
	cfg.Width = strings.TrimSpace(cfg.Width)
	if cfg.Width == "" {
		cfg.Width = "380px"
	}
	cfg.UserTokenSecret = strings.TrimSpace(cfg.UserTokenSecret)
	return cfg, nil
}

func (s *channelService) ParseWechatMPChannelConfig(raw string) (*dto.WechatMPChannelConfig, error) {
	raw = strings.TrimSpace(raw)
	cfg := &dto.WechatMPChannelConfig{
		Title:      "公众号客服",
		Subtitle:   "欢迎咨询",
		ThemeColor: "#2563eb",
	}
	if raw != "" {
		if err := json.Unmarshal([]byte(raw), cfg); err != nil {
			return nil, err
		}
	}
	cfg.Title = strings.TrimSpace(cfg.Title)
	if cfg.Title == "" {
		cfg.Title = "公众号客服"
	}
	cfg.Subtitle = strings.TrimSpace(cfg.Subtitle)
	cfg.ThemeColor = strings.TrimSpace(cfg.ThemeColor)
	if cfg.ThemeColor == "" {
		cfg.ThemeColor = "#2563eb"
	}
	cfg.UserTokenSecret = strings.TrimSpace(cfg.UserTokenSecret)
	return cfg, nil
}

func (s *channelService) GetUserTokenSecret(channel *models.Channel) string {
	if channel == nil {
		return ""
	}
	switch channel.ChannelType {
	case enums.ChannelTypeWeb:
		cfg, err := s.ParseWebChannelConfig(channel.ConfigJSON)
		if err != nil {
			return ""
		}
		return strings.TrimSpace(cfg.UserTokenSecret)
	case enums.ChannelTypeWechatMP:
		cfg, err := s.ParseWechatMPChannelConfig(channel.ConfigJSON)
		if err != nil {
			return ""
		}
		return strings.TrimSpace(cfg.UserTokenSecret)
	default:
		return ""
	}
}

func (s *channelService) ResetUserTokenSecret(channelID int64, operator *dto.AuthPrincipal) (string, error) {
	if operator == nil {
		return "", errorsx.Unauthorized("未登录或登录已过期")
	}
	channel := s.Get(channelID)
	if channel == nil || channel.Status == enums.StatusDeleted {
		return "", errorsx.InvalidParam("接入渠道不存在")
	}
	if channel.ChannelType != enums.ChannelTypeWeb && channel.ChannelType != enums.ChannelTypeWechatMP {
		return "", errorsx.InvalidParam("当前渠道不支持用户 JWT Secret")
	}
	secret, err := generateUserTokenSecret()
	if err != nil {
		return "", err
	}
	var configJSON string
	switch channel.ChannelType {
	case enums.ChannelTypeWeb:
		cfg, err := s.ParseWebChannelConfig(channel.ConfigJSON)
		if err != nil {
			return "", err
		}
		cfg.UserTokenSecret = secret
		raw, err := json.Marshal(cfg)
		if err != nil {
			return "", err
		}
		configJSON = string(raw)
	case enums.ChannelTypeWechatMP:
		cfg, err := s.ParseWechatMPChannelConfig(channel.ConfigJSON)
		if err != nil {
			return "", err
		}
		cfg.UserTokenSecret = secret
		raw, err := json.Marshal(cfg)
		if err != nil {
			return "", err
		}
		configJSON = string(raw)
	}
	if err := repositories.ChannelRepository.Updates(sqls.DB(), channelID, map[string]any{
		"config_json":      configJSON,
		"update_user_id":   operator.UserID,
		"update_user_name": operator.Username,
		"updated_at":       time.Now(),
	}); err != nil {
		return "", err
	}
	return secret, nil
}

func generateUserTokenSecret() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
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

func (s *channelService) GetEnabledChannel(ctx iris.Context) *models.Channel {
	channelID := irisx.GetChannelID(ctx)
	channel := repositories.ChannelRepository.GetByChannelID(sqls.DB(), channelID)
	if channel == nil {
		return nil
	}
	if channel.Status != enums.StatusOk {
		return nil
	}
	return channel
}

func (s *channelService) buildChannelModel(id int64, req request.CreateChannelRequest) (*models.Channel, error) {
	channelType := strings.TrimSpace(req.ChannelType)
	if channelType != enums.ChannelTypeWeb && channelType != enums.ChannelTypeWechatMP && channelType != enums.ChannelTypeWxWorkKF {
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

	channelID := ""
	if id > 0 {
		current := s.Get(id)
		if current == nil || current.Status == enums.StatusDeleted {
			return nil, errorsx.InvalidParam("接入渠道不存在")
		}
		channelID = strings.TrimSpace(current.ChannelID)
	}
	configJSON := strings.TrimSpace(req.ConfigJSON)
	switch channelType {
	case enums.ChannelTypeWeb:
		if channelID == "" {
			channelID = strs.UUID()
		}
		if exists := s.Take("channel_id = ? AND status <> ? AND id <> ?", channelID, enums.StatusDeleted, id); exists != nil {
			return nil, errorsx.InvalidParam("渠道标识已存在")
		}
		cfg, err := s.ParseWebChannelConfig(configJSON)
		if err != nil {
			return nil, errorsx.InvalidParam("Web渠道配置不合法")
		}
		if strings.TrimSpace(cfg.UserTokenSecret) == "" {
			secret, err := generateUserTokenSecret()
			if err != nil {
				return nil, err
			}
			cfg.UserTokenSecret = secret
		}
		configBytes, err := json.Marshal(cfg)
		if err != nil {
			return nil, err
		}
		configJSON = string(configBytes)
	case enums.ChannelTypeWechatMP:
		if channelID == "" {
			channelID = strs.UUID()
		}
		if exists := s.Take("channel_id = ? AND status <> ? AND id <> ?", channelID, enums.StatusDeleted, id); exists != nil {
			return nil, errorsx.InvalidParam("渠道标识已存在")
		}
		cfg, err := s.ParseWechatMPChannelConfig(configJSON)
		if err != nil {
			return nil, errorsx.InvalidParam("微信公众号渠道配置不合法")
		}
		if strings.TrimSpace(cfg.UserTokenSecret) == "" {
			secret, err := generateUserTokenSecret()
			if err != nil {
				return nil, err
			}
			cfg.UserTokenSecret = secret
		}
		configBytes, err := json.Marshal(cfg)
		if err != nil {
			return nil, err
		}
		configJSON = string(configBytes)
	case enums.ChannelTypeWxWorkKF:
		if channelID == "" {
			channelID = strs.UUID()
		}
		if exists := s.Take("channel_id = ? AND status <> ? AND id <> ?", channelID, enums.StatusDeleted, id); exists != nil {
			return nil, errorsx.InvalidParam("渠道标识已存在")
		}
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
		ChannelID:   channelID,
		AIAgentID:   req.AIAgentID,
		Name:        name,
		ConfigJSON:  configJSON,
		Status:      status,
		Remark:      strings.TrimSpace(req.Remark),
	}, nil
}
