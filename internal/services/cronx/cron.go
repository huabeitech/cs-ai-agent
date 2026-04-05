package cronx

import (
	"cs-agent/internal/services"
	"fmt"
	"log/slog"

	"github.com/robfig/cron/v3"
)

func Init() {
	c := cron.New()

	addFunc(c, "0 4 ? * *", func() {
		fmt.Println("cron test")
	})

	addFunc(c, "@every 30s", func() {
		if _, err := services.ConversationDispatchService.DispatchPendingConversations(0); err != nil {
			slog.Warn("dispatch pending conversations loop failed", "error", err)
		}
	})

	addFunc(c, "@every 15s", func() {
		count, err := services.WxWorkKFOutboundService.DispatchPendingOutbox(20)
		if err != nil {
			slog.Warn("dispatch wxwork kf outbox failed", "error", err)
			return
		}
		if count > 0 {
			slog.Info("wxwork kf outbox dispatched", "count", count)
		}
	})

	addFunc(c, "@every 1m", func() {
		count, err := services.TicketService.ScanAndMarkBreachedSLAs(200)
		if err != nil {
			slog.Warn("scan breached ticket slas failed", "error", err)
			return
		}
		if count > 0 {
			slog.Info("ticket sla breached scan completed", "breachedCount", count)
		}
	})

	c.Start()
}

func addFunc(c *cron.Cron, sepc string, cmd func()) {
	if _, err := c.AddFunc(sepc, cmd); err != nil {
		slog.Error("add cron func error", slog.Any("err", err))
	}
}
