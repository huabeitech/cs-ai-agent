package services

import (
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/repositories"
	"encoding/json"
	"strings"

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

func (s *channelService) Delete(id int64) {
	repositories.ChannelRepository.Delete(sqls.DB(), id)
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
