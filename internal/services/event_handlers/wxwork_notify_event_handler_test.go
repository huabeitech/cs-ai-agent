package event_handlers

import (
	"strings"
	"testing"

	"cs-agent/internal/events"
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/enums"
)

func TestWxWorkNotifyBuildTicketCreatedNotifyBody(t *testing.T) {
	body := buildTicketCreatedNotifyBody(&models.Ticket{
		ID:       12,
		TicketNo: "T-12",
		Title:    "登录失败",
		Source:   enums.TicketSourceManual,
		Status:   enums.TicketStatusNew,
	})

	for _, want := range []string{"工单号: T-12", "工单标题: 登录失败", "当前状态:"} {
		if !strings.Contains(body, want) {
			t.Fatalf("expected body to contain %q, got %q", want, body)
		}
	}
}

func TestWxWorkNotifyBuildConversationAssignedNotifyBody(t *testing.T) {
	body := buildConversationAssignedNotifyBody(&models.Conversation{
		ID:      7,
		Subject: "售后咨询",
		Status:  enums.IMConversationStatusActive,
	}, 0, "客户等待中", events.ConversationAssignTypeTransfer)

	for _, want := range []string{"会话ID: #7", "会话主题: 售后咨询", "转接原因: 客户等待中"} {
		if !strings.Contains(body, want) {
			t.Fatalf("expected body to contain %q, got %q", want, body)
		}
	}
}
