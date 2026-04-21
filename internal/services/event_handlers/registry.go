package event_handlers

import "sync"

var registerOnce sync.Once

func Register() {
	registerOnce.Do(func() {
		registerWxWorkNotifyEventHandlers()
	})
}
