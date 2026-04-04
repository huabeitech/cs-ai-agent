package wx_callback_handlers

import (
	"cs-agent/internal/wxwork"
	"log/slog"
	"strings"

	"github.com/silenceper/wechat/v2/work/kf"
)

const syncMsgLimit = 1000

func init() {
	wxwork.RegHandler("event", "kf_msg_or_event", kf_msg_or_event_handler)
}

func kf_msg_or_event_handler(message kf.CallbackMessage) {
	cli, err := wxwork.GetWorkCli().GetKF()
	if err != nil {
		slog.Error("get wxwork kf client failed", "error", err)
		return
	}

	cursor := ""
	for {
		result, syncErr := cli.SyncMsg(kf.SyncMsgOptions{
			Cursor:   cursor,
			Token:    message.Token,
			Limit:    syncMsgLimit,
			OpenKfID: message.OpenKfID,
		})
		if syncErr != nil {
			slog.Error("sync wxwork messages failed",
				"open_kfid", message.OpenKfID,
				"cursor", cursor,
				"error", syncErr,
			)
			return
		}

		for _, item := range result.MsgList {
			slog.Info("consume wxwork message",
				"open_kfid", message.OpenKfID,
				"msg_id", item.MsgID,
				"msg_type", item.MsgType,
				"event", item.EventType,
			)

			// TODO 存储微信消息记录，存储的时候需要去重
			// TODO 消息转成客服会话，然后将客服会话消息转给客服控制台
		}

		if result.HasMore != 1 || strings.TrimSpace(result.NextCursor) == "" {
			return
		}
		cursor = result.NextCursor
	}
}
