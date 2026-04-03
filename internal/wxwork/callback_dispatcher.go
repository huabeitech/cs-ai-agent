package wxwork

import (
	"log/slog"
	"strings"
	"sync"

	"github.com/panjf2000/ants/v2"
	"github.com/silenceper/wechat/v2/work/kf"
	"github.com/silenceper/wechat/v2/work/kf/syncmsg"
)

const (
	callbackWorkerCount       = 16
	callbackMaxBlockingTasks  = 10000
	callbackMsgTypeEvent      = "event"
	callbackEventKFMsgOrEvent = "kf_msg_or_event"
	syncMsgLimit              = 1000
)

type CallbackHandler func(message kf.CallbackMessage)

type SyncMessageHandler func(message syncmsg.Message)

var (
	callbackHandlerMu sync.RWMutex
	callbackHandlers  map[string]CallbackHandler

	syncMessageHandlerMu sync.RWMutex
	syncMessageHandlers  map[string]SyncMessageHandler

	callbackPool *ants.PoolWithFunc
)

func init() {
	var err error

	callbackHandlers = make(map[string]CallbackHandler)
	syncMessageHandlers = make(map[string]SyncMessageHandler)
	callbackPool, err = ants.NewPoolWithFunc(
		callbackWorkerCount,
		dispatchCallbackMessage,
		ants.WithMaxBlockingTasks(callbackMaxBlockingTasks),
	)
	if err != nil {
		slog.Error("init wxwork callback pool failed", "error", err)
	}

	RegCallbackHandler(callbackMsgTypeEvent, callbackEventKFMsgOrEvent, consumeSyncMessages)
}

// ConsumeCallback 异步消费企业微信客服回调。
func ConsumeCallback(message kf.CallbackMessage) error {
	if callbackPool == nil {
		slog.Error("wxwork callback pool is not initialized",
			"msg_type", message.MsgType,
			"event", message.Event,
			"open_kfid", message.OpenKfID,
		)
		return nil
	}

	if err := callbackPool.Invoke(message); err != nil {
		slog.Error("invoke wxwork callback failed",
			"msg_type", message.MsgType,
			"event", message.Event,
			"open_kfid", message.OpenKfID,
			"error", err,
		)
		return err
	}
	return nil
}

// RegCallbackHandler 注册企业微信客服回调处理器。
func RegCallbackHandler(msgType, event string, handler CallbackHandler) {
	callbackHandlerMu.Lock()
	defer callbackHandlerMu.Unlock()

	key := BuildCallbackKey(msgType, event)
	if callbackHandlers[key] != nil {
		slog.Error("duplicate wxwork callback handler registration", "key", key)
		return
	}
	callbackHandlers[key] = handler
}

// RegSyncMessageHandler 注册 sync_msg 拉取后的消息处理器。
func RegSyncMessageHandler(msgType, eventType string, handler SyncMessageHandler) {
	syncMessageHandlerMu.Lock()
	defer syncMessageHandlerMu.Unlock()

	key := BuildSyncMessageKey(msgType, eventType)
	if syncMessageHandlers[key] != nil {
		slog.Error("duplicate wxwork sync message handler registration", "key", key)
		return
	}
	syncMessageHandlers[key] = handler
}

// BuildCallbackKey 生成客服回调处理器的注册键。
func BuildCallbackKey(msgType, event string) string {
	msgType = strings.TrimSpace(msgType)
	event = strings.TrimSpace(event)
	if msgType == callbackMsgTypeEvent {
		return msgType + ":" + event
	}
	return msgType
}

// BuildSyncMessageKey 生成 sync_msg 消息处理器的注册键。
func BuildSyncMessageKey(msgType, eventType string) string {
	msgType = strings.TrimSpace(msgType)
	eventType = strings.TrimSpace(eventType)
	if msgType == callbackMsgTypeEvent {
		return msgType + ":" + eventType
	}
	return msgType
}

func dispatchCallbackMessage(value any) {
	message, ok := value.(kf.CallbackMessage)
	if !ok {
		slog.Error("invalid wxwork callback message type")
		return
	}

	handler := getCallbackHandler(message)
	if handler == nil {
		return
	}
	handler(message)
}

func getCallbackHandler(message kf.CallbackMessage) CallbackHandler {
	callbackHandlerMu.RLock()
	defer callbackHandlerMu.RUnlock()

	key := BuildCallbackKey(message.MsgType, message.Event)
	handler, ok := callbackHandlers[key]
	if ok {
		return handler
	}

	slog.Warn("wxwork callback handler not found",
		"key", key,
		"msg_type", message.MsgType,
		"event", message.Event,
	)
	return nil
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

	key := BuildSyncMessageKey(message.MsgType, message.EventType)
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
