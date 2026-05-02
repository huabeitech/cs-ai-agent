package builders

import (
	"testing"
	"time"

	"cs-agent/internal/models"
	"cs-agent/internal/pkg/enums"
)

func TestBuildLightweightTicket(t *testing.T) {
	now := time.Date(2026, 5, 2, 12, 30, 0, 0, time.Local)
	ticket := &models.Ticket{
		ID:                12,
		TicketNo:          "TK202605020001",
		Title:             "登录失败",
		Description:       "客户反馈无法登录",
		Source:            enums.TicketSourceManual,
		Channel:           "web",
		CustomerID:        3,
		ConversationID:    4,
		Status:            enums.TicketStatusPending,
		CurrentAssigneeID: 5,
		AuditFields: models.AuditFields{
			CreateUserID:   1,
			CreateUserName: "admin",
			CreatedAt:      now,
			UpdatedAt:      now,
		},
	}
	ctx := &TicketBuildContext{
		TagsByTicketID: map[int64][]models.Tag{
			12: {{ID: 8, Name: "登录", Status: enums.StatusOk}},
		},
		Users: map[int64]*models.User{
			5: {ID: 5, Username: "agent", Nickname: "客服"},
		},
		Customers: map[int64]*models.Customer{
			3: {ID: 3, Name: "客户"},
		},
	}

	out := BuildTicketWithContext(ticket, ctx)
	if out == nil {
		t.Fatalf("expected ticket response")
	}
	if out.ID != ticket.ID || out.TicketNo != ticket.TicketNo || out.Status != ticket.Status {
		t.Fatalf("unexpected ticket response: %+v", out)
	}
	if out.CurrentAssigneeName != "客服" {
		t.Fatalf("expected assignee name, got %q", out.CurrentAssigneeName)
	}
	if len(out.Tags) != 1 || out.Tags[0].ID != 8 {
		t.Fatalf("expected tag response, got %+v", out.Tags)
	}
	if out.Customer == nil || out.Customer.ID != 3 {
		t.Fatalf("expected customer response, got %+v", out.Customer)
	}
}

func TestBuildTicketWithoutContextLeavesOptionalLookupsEmpty(t *testing.T) {
	now := time.Date(2026, 5, 2, 12, 30, 0, 0, time.Local)
	ticket := &models.Ticket{
		ID:                12,
		TicketNo:          "TK202605020001",
		Title:             "登录失败",
		Description:       "客户反馈无法登录",
		CustomerID:        3,
		CurrentAssigneeID: 5,
		Status:            enums.TicketStatusPending,
		AuditFields: models.AuditFields{
			CreateUserID:   1,
			CreateUserName: "admin",
			CreatedAt:      now,
			UpdatedAt:      now,
		},
	}

	out := BuildTicket(ticket)
	if out == nil {
		t.Fatalf("expected ticket response")
	}
	if out.Tags != nil {
		t.Fatalf("expected tags to stay empty without context, got %+v", out.Tags)
	}
	if out.Customer != nil {
		t.Fatalf("expected customer to stay empty without context, got %+v", out.Customer)
	}
	if out.CurrentAssigneeName != "" {
		t.Fatalf("expected assignee name to stay empty without context, got %q", out.CurrentAssigneeName)
	}
}

func TestBuildTicketProgress(t *testing.T) {
	now := time.Date(2026, 5, 2, 12, 30, 0, 0, time.Local)
	progress := &models.TicketProgress{
		ID:        1,
		TicketID:  2,
		Content:   "已联系客户",
		AuthorID:  3,
		CreatedAt: now,
	}
	ctx := &TicketDetailBuildContext{
		Users: map[int64]*models.User{
			3: {ID: 3, Username: "agent", Nickname: "客服"},
		},
	}

	out := BuildTicketProgressWithContext(progress, ctx)
	if out == nil {
		t.Fatalf("expected progress response")
	}
	if out.ID != progress.ID || out.TicketID != progress.TicketID || out.Content != progress.Content {
		t.Fatalf("unexpected progress response: %+v", out)
	}
	if out.AuthorName != "客服" {
		t.Fatalf("expected author name, got %q", out.AuthorName)
	}
	if out.CreatedAt == "" {
		t.Fatalf("expected createdAt to be formatted")
	}
}
