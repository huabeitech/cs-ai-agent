package wxwork

import (
	"testing"

	"github.com/silenceper/wechat/v2/work/kf"
	"github.com/silenceper/wechat/v2/work/kf/syncmsg"
)

func TestBuildCallbackKey(t *testing.T) {
	if got := buildCallbackKey("event", "kf_msg_or_event"); got != "event:kf_msg_or_event" {
		t.Fatalf("unexpected callback key: %s", got)
	}
	if got := buildCallbackKey("text", "ignored"); got != "text" {
		t.Fatalf("unexpected callback key for non-event: %s", got)
	}
}

func TestBuildSyncMessageKey(t *testing.T) {
	if got := buildSyncMessageKey("event", "enter_session"); got != "event:enter_session" {
		t.Fatalf("unexpected sync message key: %s", got)
	}
	if got := buildSyncMessageKey("text", "ignored"); got != "text" {
		t.Fatalf("unexpected sync message key for non-event: %s", got)
	}
}

func TestGetCallbackHandler(t *testing.T) {
	original := callbackHandlers
	callbackHandlers = map[string]CallbackHandler{}
	defer func() {
		callbackHandlers = original
	}()

	called := false
	handler := func(message kf.CallbackMessage) {
		called = message.Event == "kf_msg_or_event"
	}
	RegCallbackHandler("event", "kf_msg_or_event", handler)

	got := getCallbackHandler(kf.CallbackMessage{MsgType: "event", Event: "kf_msg_or_event"})
	if got == nil {
		t.Fatalf("expected callback handler")
	}
	got(kf.CallbackMessage{MsgType: "event", Event: "kf_msg_or_event"})
	if !called {
		t.Fatalf("expected callback handler to be called")
	}
}

func TestGetSyncMessageHandler(t *testing.T) {
	original := syncMessageHandlers
	syncMessageHandlers = map[string]SyncMessageHandler{}
	defer func() {
		syncMessageHandlers = original
	}()

	called := false
	handler := func(message syncmsg.Message) {
		called = message.EventType == "enter_session"
	}
	RegSyncMessageHandler("event", "enter_session", handler)

	got := getSyncMessageHandler(syncmsg.Message{MsgType: "event", EventType: "enter_session"})
	if got == nil {
		t.Fatalf("expected sync message handler")
	}
	got(syncmsg.Message{MsgType: "event", EventType: "enter_session"})
	if !called {
		t.Fatalf("expected sync message handler to be called")
	}
}
