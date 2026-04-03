package wxwork

import (
	"log/slog"
	"strings"
	"sync"

	"github.com/silenceper/wechat/v2/work/kf"
	"github.com/silenceper/wechat/v2/work/kf/syncmsg"
)

const syncMsgLimit = 1000

type SyncMessageHandler func(message syncmsg.Message)

var (
	syncMessageHandlerMu sync.RWMutex
	syncMessageHandlers  = make(map[string]SyncMessageHandler)
)

// RegSyncMessageHandler 注册 sync_msg 拉取后的消息处理器。
func RegSyncMessageHandler(msgType, eventType string, handler SyncMessageHandler) {
	syncMessageHandlerMu.Lock()
	defer syncMessageHandlerMu.Unlock()

	key := buildSyncMessageKey(msgType, eventType)
	if syncMessageHandlers[key] != nil {
		slog.Error("duplicate wxwork sync message handler registration", "key", key)
		return
	}
	syncMessageHandlers[key] = handler
}

// buildSyncMessageKey 生成 sync_msg 消息处理器的注册键。
func buildSyncMessageKey(msgType, eventType string) string {
	msgType = strings.TrimSpace(msgType)
	eventType = strings.TrimSpace(eventType)
	if msgType == callbackMsgTypeEvent {
		return msgType + ":" + eventType
	}
	return msgType
}

func consumeSyncMessages(message kf.CallbackMessage) {
	cli, err := GetWorkCli().GetKF()
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
			dispatchSyncMessage(item)
		}

		if result.HasMore != 1 || strings.TrimSpace(result.NextCursor) == "" {
			return
		}
		cursor = result.NextCursor
	}
}

func dispatchSyncMessage(message syncmsg.Message) {
	handler := getSyncMessageHandler(message)
	if handler == nil {
		return
	}
	handler(message)
}

func getSyncMessageHandler(message syncmsg.Message) SyncMessageHandler {
	syncMessageHandlerMu.RLock()
	defer syncMessageHandlerMu.RUnlock()

	key := buildSyncMessageKey(message.MsgType, message.EventType)
	handler, ok := syncMessageHandlers[key]
	if ok {
		return handler
	}

	slog.Warn("wxwork sync message handler not found",
		"key", key,
		"msg_type", message.MsgType,
		"event_type", message.EventType,
		"msg_id", message.MsgID,
		"open_kfid", message.OpenKFID,
	)
	return nil
}
