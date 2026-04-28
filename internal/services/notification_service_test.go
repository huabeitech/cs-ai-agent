package services_test

import (
	"testing"

	"cs-agent/internal/models"
	"cs-agent/internal/pkg/dto/request"
	"cs-agent/internal/services"

	"github.com/glebarez/sqlite"
	"github.com/mlogclub/simple/sqls"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func TestNotificationServiceCreateAndUnreadCount(t *testing.T) {
	setupNotificationTestDB(t)

	item, err := services.NotificationService.Create(request.CreateNotificationRequest{
		RecipientUserID:  101,
		Title:            "工单指派提醒",
		Content:          "工单 TK-1 已指派给你",
		NotificationType: "ticket_assigned",
		BizType:          "ticket",
		BizID:            1,
		ActionURL:        "/dashboard/tickets/1",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if item.ID == 0 {
		t.Fatalf("expected notification id to be assigned")
	}
	if item.RecipientUserID != 101 || item.ReadAt != nil {
		t.Fatalf("unexpected notification: %+v", item)
	}
	if got := services.NotificationService.CountUnread(101); got != 1 {
		t.Fatalf("expected unread count 1, got %d", got)
	}
	if got := services.NotificationService.CountUnread(102); got != 0 {
		t.Fatalf("expected unread count 0 for another user, got %d", got)
	}
}

func TestNotificationServiceMarkReadRequiresOwner(t *testing.T) {
	setupNotificationTestDB(t)

	item, err := services.NotificationService.Create(request.CreateNotificationRequest{
		RecipientUserID:  201,
		Title:            "会话分配提醒",
		Content:          "会话 #9 已分配给你",
		NotificationType: "conversation_assigned",
		BizType:          "conversation",
		BizID:            9,
		ActionURL:        "/dashboard/conversations?conversationId=9",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if err := services.NotificationService.MarkRead(item.ID, 202); err == nil {
		t.Fatalf("expected foreign user mark read to fail")
	}
	if got := services.NotificationService.CountUnread(201); got != 1 {
		t.Fatalf("expected notification to remain unread, got %d", got)
	}
	if err := services.NotificationService.MarkRead(item.ID, 201); err != nil {
		t.Fatalf("MarkRead() owner error = %v", err)
	}
	if got := services.NotificationService.CountUnread(201); got != 0 {
		t.Fatalf("expected unread count 0 after mark read, got %d", got)
	}
}

func TestNotificationServiceMarkAllReadOnlyCurrentUser(t *testing.T) {
	setupNotificationTestDB(t)

	for _, userID := range []int64{301, 301, 302} {
		if _, err := services.NotificationService.Create(request.CreateNotificationRequest{
			RecipientUserID:  userID,
			Title:            "工单指派提醒",
			Content:          "工单已指派给你",
			NotificationType: "ticket_assigned",
			BizType:          "ticket",
			BizID:            userID,
			ActionURL:        "/dashboard/tickets/1",
		}); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	}

	if err := services.NotificationService.MarkAllRead(301); err != nil {
		t.Fatalf("MarkAllRead() error = %v", err)
	}
	if got := services.NotificationService.CountUnread(301); got != 0 {
		t.Fatalf("expected user 301 unread count 0, got %d", got)
	}
	if got := services.NotificationService.CountUnread(302); got != 1 {
		t.Fatalf("expected user 302 unread count 1, got %d", got)
	}
}

func setupNotificationTestDB(t *testing.T) *gorm.DB {
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
	if err := db.AutoMigrate(&models.Notification{}); err != nil {
		t.Fatalf("auto migrate error = %v", err)
	}
	sqls.SetDB(db)
	return db
}
