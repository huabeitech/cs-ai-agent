package wxwork

import (
	"cs-agent/internal/pkg/config"
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

func Init() {
	w = nil
	wxCfg = config.WxWorkConfig{}
	cfg := config.Current()
	if !cfg.WxWork.Enabled {
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
		RasPrivateKey:  wxCfg.RSAPrivateKey,
		Token:          wxCfg.Token,
		EncodingAESKey: wxCfg.EncodingAESKey,
		Cache:          cache.NewMemory(),
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

func GetWorkCli() *work.Work {
	return w
}
