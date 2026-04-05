package wx_callback_handlers

import (
	"cs-agent/internal/services"
	"cs-agent/internal/wxwork"
	"log/slog"

	"github.com/silenceper/wechat/v2/work/kf"
)

func init() {
	wxwork.RegHandler("event", "kf_msg_or_event", kf_msg_or_event_handler)
}

func kf_msg_or_event_handler(message kf.CallbackMessage) {
	if err := services.WxWorkKFInboundService.SyncCallbackMessages(message); err != nil {
		slog.Error("sync wxwork callback messages failed",
			"open_kfid", message.OpenKfID,
			"event", message.Event,
			"error", err,
		)
	}
}
