package dto

import "cs-agent/internal/pkg/enums"

type AuthPrincipal struct {
	UserID      int64
	Username    string
	Nickname    string
	Avatar      string
	Status      enums.Status
	Roles       []string
	Permissions []string
}

type WxWorkKFChannelConfig struct {
	OpenKfID string `json:"openKfId"`
}

type WebChannelConfig struct {
	Title      string `json:"title"`
	Subtitle   string `json:"subtitle"`
	ThemeColor string `json:"themeColor"`
	Position   string `json:"position"`
	Width      string `json:"width"`
}

type WechatMPChannelConfig struct {
	Title      string `json:"title"`
	Subtitle   string `json:"subtitle"`
	ThemeColor string `json:"themeColor"`
}
