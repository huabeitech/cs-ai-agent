package wxwork

import (
	"cs-agent/internal/pkg/config"
	"fmt"
	"net/url"
	"strings"

	"github.com/silenceper/wechat/v2/cache"
	"github.com/silenceper/wechat/v2/work"
	"github.com/silenceper/wechat/v2/work/addresslist"
	wxconfig "github.com/silenceper/wechat/v2/work/config"
	"github.com/silenceper/wechat/v2/work/oauth"
)

var (
	w     *work.Work
	wxCfg config.WxWorkConfig
)

type LoginUser struct {
	CorpID         string                       `json:"corpId"`
	UserID         string                       `json:"userId"`
	OpenID         string                       `json:"openId,omitempty"`
	ExternalUserID string                       `json:"externalUserId,omitempty"`
	UserTicket     string                       `json:"userTicket,omitempty"`
	Name           string                       `json:"name,omitempty"`
	Avatar         string                       `json:"avatar,omitempty"`
	Mobile         string                       `json:"mobile,omitempty"`
	Email          string                       `json:"email,omitempty"`
	BizMail        string                       `json:"bizMail,omitempty"`
	UserInfo       *oauth.GetUserInfoResponse   `json:"userInfo,omitempty"`
	UserDetail     *oauth.GetUserDetailResponse `json:"userDetail,omitempty"`
	UserProfile    *addresslist.UserGetResponse `json:"userProfile,omitempty"`
}

func Init(cfg *config.Config) {
	w = nil
	wxCfg = config.WxWorkConfig{}
	if cfg == nil || !cfg.WxWork.Enabled {
		return
	}
	wxCfg = cfg.WxWork
	if strings.TrimSpace(wxCfg.CorpID) == "" || strings.TrimSpace(wxCfg.CorpSecret) == "" {
		return
	}
	w = work.NewWork(&wxconfig.Config{
		CorpID:         wxCfg.CorpID,
		CorpSecret:     wxCfg.CorpSecret,
		AgentID:        wxCfg.AgentID,
		Cache:          cache.NewMemory(),
		RasPrivateKey:  wxCfg.RsaPrivateKey,
		Token:          wxCfg.Token,
		EncodingAESKey: wxCfg.EncodingAESKey,
	})
}

func Enabled() bool {
	return w != nil && wxCfg.Enabled
}

func StateSecret() string {
	if strings.TrimSpace(wxCfg.StateSecret) != "" {
		return strings.TrimSpace(wxCfg.StateSecret)
	}
	return strings.TrimSpace(wxCfg.CorpSecret)
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

func GetLoginUser(code string) (*LoginUser, error) {
	if !Enabled() {
		return nil, fmt.Errorf("企业微信登录未启用")
	}
	code = strings.TrimSpace(code)
	if code == "" {
		return nil, fmt.Errorf("微信授权 code 不能为空")
	}

	oauthClient := w.GetOauth()
	userInfo, err := oauthClient.GetUserInfo(code)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(userInfo.UserID) == "" {
		return nil, fmt.Errorf("当前登录身份不是企业内部成员")
	}

	ret := &LoginUser{
		CorpID:         wxCfg.CorpID,
		UserID:         strings.TrimSpace(userInfo.UserID),
		OpenID:         strings.TrimSpace(userInfo.OpenID),
		ExternalUserID: strings.TrimSpace(userInfo.ExternalUserID),
		UserTicket:     strings.TrimSpace(userInfo.UserTicket),
		UserInfo:       userInfo,
	}

	if ret.UserTicket != "" {
		if detail, detailErr := oauthClient.GetUserDetail(&oauth.GetUserDetailRequest{UserTicket: ret.UserTicket}); detailErr == nil {
			ret.UserDetail = detail
			ret.Avatar = strings.TrimSpace(detail.Avatar)
			ret.Mobile = strings.TrimSpace(detail.Mobile)
			ret.Email = strings.TrimSpace(detail.Email)
			ret.BizMail = strings.TrimSpace(detail.BizMail)
		}
	}

	if profile, profileErr := w.GetAddressList().UserGet(ret.UserID); profileErr == nil {
		ret.UserProfile = profile
		if strings.TrimSpace(profile.Name) != "" {
			ret.Name = strings.TrimSpace(profile.Name)
		}
		if strings.TrimSpace(profile.Avatar) != "" {
			ret.Avatar = strings.TrimSpace(profile.Avatar)
		}
		if ret.Mobile == "" && strings.TrimSpace(profile.Mobile) != "" {
			ret.Mobile = strings.TrimSpace(profile.Mobile)
		}
		if ret.Email == "" && strings.TrimSpace(profile.Email) != "" {
			ret.Email = strings.TrimSpace(profile.Email)
		}
		if ret.BizMail == "" && strings.TrimSpace(profile.BizMail) != "" {
			ret.BizMail = strings.TrimSpace(profile.BizMail)
		}
	}

	if ret.Name == "" {
		ret.Name = ret.UserID
	}
	return ret, nil
}
