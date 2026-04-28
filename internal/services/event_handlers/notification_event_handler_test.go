package event_handlers

import (
	"context"
	"testing"
	"time"

	"cs-agent/internal/events"
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/repositories"

	"github.com/glebarez/sqlite"
	"github.com/mlogclub/simple/sqls"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func TestTicketAssignedInAppNotification(t *testing.T) {
	setupNotificationEventHandlerTestDB(t)

	ticket := &models.Ticket{
		TicketNo:          "TK202604280001",
		Title:             "退款处理",
		Source:            enums.TicketSourceManual,
		Status:            enums.TicketStatusOpen,
		CurrentAssigneeID: 11,
		AuditFields: models.AuditFields{
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}
	if err := repositories.TicketRepository.Create(sqls.DB(), ticket); err != nil {
		t.Fatalf("create ticket error = %v", err)
	}

	if err := handleTicketAssignedInAppNotification(context.Background(), events.TicketAssignedEvent{
		TicketID:   ticket.ID,
		FromUserID: 0,
		ToUserID:   11,
		OperatorID: 1,
		Reason:     "需要人工跟进",
	}); err != nil {
		t.Fatalf("handler error = %v", err)
	}

	list := repositories.NotificationRepository.Find(sqls.DB(), sqls.NewCnd().Eq("recipient_user_id", 11))
	if len(list) != 1 {
		t.Fatalf("expected 1 notification, got %d", len(list))
	}
	got := list[0]
	if got.NotificationType != "ticket_assigned" || got.BizType != "ticket" || got.BizID != ticket.ID {
		t.Fatalf("unexpected notification: %+v", got)
	}
	if got.ActionURL != "/dashboard/tickets/1" {
		t.Fatalf("unexpected action url: %q", got.ActionURL)
	}
}

func TestConversationAssignedInAppNotification(t *testing.T) {
	setupNotificationEventHandlerTestDB(t)

	conversation := &models.Conversation{
		CustomerName:      "张三",
		Status:            enums.IMConversationStatusActive,
		CurrentAssigneeID: 22,
		AuditFields: models.AuditFields{
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}
	if err := repositories.ConversationRepository.Create(sqls.DB(), conversation); err != nil {
		t.Fatalf("create conversation error = %v", err)
	}

	if err := handleConversationAssignedInAppNotification(context.Background(), events.ConversationAssignedEvent{
		ConversationID: conversation.ID,
		FromUserID:     0,
		ToUserID:       22,
		OperatorID:     1,
		Reason:         "自动分配",
		AssignType:     events.ConversationAssignTypeAutoAssign,
	}); err != nil {
		t.Fatalf("handler error = %v", err)
	}

	list := repositories.NotificationRepository.Find(sqls.DB(), sqls.NewCnd().Eq("recipient_user_id", 22))
	if len(list) != 1 {
		t.Fatalf("expected 1 notification, got %d", len(list))
	}
	got := list[0]
	if got.NotificationType != "conversation_assigned" || got.BizType != "conversation" || got.BizID != conversation.ID {
		t.Fatalf("unexpected notification: %+v", got)
	}
	if got.ActionURL != "/dashboard/conversations?conversationId=1" {
		t.Fatalf("unexpected action url: %q", got.ActionURL)
	}
}

func setupNotificationEventHandlerTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "t_",
			SingularTable: true,
		},
	})
	if err != nil {
		t.Fatalf("open sqlite error = %v", err)
	}
	t.Cleanup(func() {
		sqlDB, err := db.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	})
	if err := db.AutoMigrate(&models.Notification{}, &models.Ticket{}, &models.Conversation{}); err != nil {
		t.Fatalf("auto migrate error = %v", err)
	}
	sqls.SetDB(db)
	return db
}
