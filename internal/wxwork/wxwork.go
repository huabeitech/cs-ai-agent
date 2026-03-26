package wxwork

import (
	"cs-agent/internal/pkg/config"

	"github.com/silenceper/wechat/v2/work"
	wxconfig "github.com/silenceper/wechat/v2/work/config"
)

var (
	w *work.Work
)

func Init(cfg *config.Config) {
	w = work.NewWork(&wxconfig.Config{
		CorpID:         cfg.WxWork.CorpID,
		CorpSecret:     cfg.WxWork.CorpSecret,
		AgentID:        cfg.WxWork.AgentID,
		Cache:          nil,
		RasPrivateKey:  cfg.WxWork.RsaPrivateKey,
		Token:          cfg.WxWork.Token,
		EncodingAESKey: cfg.WxWork.EncodingAESKey,
	})
}
