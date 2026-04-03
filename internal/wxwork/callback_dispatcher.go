package wxwork

import (
	"log/slog"
	"strings"
	"sync"

	"github.com/panjf2000/ants/v2"
	"github.com/silenceper/wechat/v2/work/kf"
)

const (
	callbackWorkerCount       = 16
	callbackMaxBlockingTasks  = 10000
	callbackMsgTypeEvent      = "event"
	callbackEventKFMsgOrEvent = "kf_msg_or_event"
)

type CallbackHandler func(message kf.CallbackMessage)

var (
	callbackHandlerMu sync.RWMutex
	callbackHandlers  map[string]CallbackHandler
	callbackPool      *ants.PoolWithFunc
)

func init() {
	var err error

	callbackHandlers = make(map[string]CallbackHandler)
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

	key := buildCallbackKey(msgType, event)
	if callbackHandlers[key] != nil {
		slog.Error("duplicate wxwork callback handler registration", "key", key)
		return
	}
	callbackHandlers[key] = handler
}

// buildCallbackKey 生成客服回调处理器的注册键。
func buildCallbackKey(msgType, event string) string {
	msgType = strings.TrimSpace(msgType)
	event = strings.TrimSpace(event)
	if msgType == callbackMsgTypeEvent {
		return msgType + ":" + event
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

	key := buildCallbackKey(message.MsgType, message.Event)
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
