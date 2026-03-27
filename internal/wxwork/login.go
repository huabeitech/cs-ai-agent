package wxwork

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"cs-agent/internal/pkg/dto/response"
	"cs-agent/internal/pkg/errorsx"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"
)

const (
	StateTTL             = 5 * time.Minute
	LoginTicketTTL       = 1 * time.Minute
	defaultLoginNextPath = "/dashboard"
)

var (
	errStateInvalid  = fmt.Errorf("企业微信登录状态无效")
	loginTicketStore sync.Map
)

type statePayload struct {
	Next      string `json:"next"`
	Nonce     string `json:"nonce"`
	ExpiredAt int64  `json:"expiredAt"`
}

type loginTicket struct {
	Response  *response.LoginResponse
	ExpiredAt time.Time
}

func BuildLoginURL(state string) (string, error) {
	if !Enabled() {
		return "", fmt.Errorf("企业微信登录未启用")
	}
	if strings.TrimSpace(wxCfg.OAuthRedirect) == "" {
		return "", fmt.Errorf("企业微信登录回调地址未配置")
	}
	if strings.TrimSpace(wxCfg.AgentID) == "" {
		return "", fmt.Errorf("企业微信 AgentID 未配置")
	}
	return fmt.Sprintf(
		"https://open.weixin.qq.com/connect/oauth2/authorize?appid=%s&redirect_uri=%s&response_type=code&scope=snsapi_privateinfo&agentid=%s&state=%s#wechat_redirect",
		url.QueryEscape(strings.TrimSpace(wxCfg.CorpID)),
		url.QueryEscape(strings.TrimSpace(wxCfg.OAuthRedirect)),
		url.QueryEscape(strings.TrimSpace(wxCfg.AgentID)),
		url.QueryEscape(strings.TrimSpace(state)),
	), nil
}

func BuildQRCodeLoginURL(state string) (string, error) {
	if !Enabled() {
		return "", fmt.Errorf("企业微信登录未启用")
	}
	if strings.TrimSpace(wxCfg.OAuthRedirect) == "" {
		return "", fmt.Errorf("企业微信登录回调地址未配置")
	}
	if strings.TrimSpace(wxCfg.AgentID) == "" {
		return "", fmt.Errorf("企业微信 AgentID 未配置")
	}
	return fmt.Sprintf(
		"https://open.work.weixin.qq.com/wwopen/sso/qrConnect?appid=%s&agentid=%s&redirect_uri=%s&state=%s",
		url.QueryEscape(strings.TrimSpace(wxCfg.CorpID)),
		url.QueryEscape(strings.TrimSpace(wxCfg.AgentID)),
		url.QueryEscape(strings.TrimSpace(wxCfg.OAuthRedirect)),
		url.QueryEscape(strings.TrimSpace(state)),
	), nil
}

func CreateState(next string) (string, error) {
	secret := strings.TrimSpace(StateSecret())
	if secret == "" {
		return "", errorsx.BusinessError(1, "企业微信登录密钥未配置")
	}
	payload := statePayload{
		Next:      sanitizeNextPath(next),
		ExpiredAt: time.Now().Add(StateTTL).Unix(),
	}
	nonce, err := randomToken("ws_")
	if err != nil {
		return "", err
	}
	payload.Nonce = nonce

	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	encoded := base64.RawURLEncoding.EncodeToString(body)
	return encoded + "." + signState(encoded, secret), nil
}

func ParseState(state string) (string, error) {
	secret := strings.TrimSpace(StateSecret())
	if secret == "" {
		return "", errStateInvalid
	}
	parts := strings.Split(strings.TrimSpace(state), ".")
	if len(parts) != 2 {
		return "", errStateInvalid
	}
	if !hmac.Equal([]byte(parts[1]), []byte(signState(parts[0], secret))) {
		return "", errStateInvalid
	}
	body, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return "", errStateInvalid
	}
	payload := statePayload{}
	if err = json.Unmarshal(body, &payload); err != nil {
		return "", errStateInvalid
	}
	if payload.ExpiredAt <= time.Now().Unix() {
		return "", errStateInvalid
	}
	return sanitizeNextPath(payload.Next), nil
}

func IssueLoginTicket(loginResp *response.LoginResponse) (string, error) {
	if loginResp == nil {
		return "", fmt.Errorf("登录结果不能为空")
	}
	ticket, err := randomToken("wlt_")
	if err != nil {
		return "", err
	}
	cleanupExpiredLoginTickets()
	loginTicketStore.Store(ticket, loginTicket{
		Response:  loginResp,
		ExpiredAt: time.Now().Add(LoginTicketTTL),
	})
	return ticket, nil
}

func ConsumeLoginTicket(ticket string) (*response.LoginResponse, error) {
	ticket = strings.TrimSpace(ticket)
	if ticket == "" {
		return nil, errorsx.InvalidParam("ticket 不能为空")
	}
	value, ok := loginTicketStore.LoadAndDelete(ticket)
	if !ok {
		return nil, errorsx.Unauthorized("登录票据无效或已过期")
	}
	record, ok := value.(loginTicket)
	if !ok || record.Response == nil || time.Now().After(record.ExpiredAt) {
		return nil, errorsx.Unauthorized("登录票据无效或已过期")
	}
	return record.Response, nil
}

func signState(content, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(content))
	return hex.EncodeToString(mac.Sum(nil))
}

func cleanupExpiredLoginTickets() {
	now := time.Now()
	loginTicketStore.Range(func(key, value any) bool {
		record, ok := value.(loginTicket)
		if !ok || now.After(record.ExpiredAt) {
			loginTicketStore.Delete(key)
		}
		return true
	})
}

func sanitizeNextPath(next string) string {
	next = strings.TrimSpace(next)
	if next == "" || !strings.HasPrefix(next, "/") || strings.HasPrefix(next, "//") {
		return defaultLoginNextPath
	}
	return next
}

func randomToken(prefix string) (string, error) {
	buf := make([]byte, 24)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return prefix + hex.EncodeToString(buf), nil
}
