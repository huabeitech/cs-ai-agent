package services

import (
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/dto"
	"cs-agent/internal/pkg/dto/request"
	"cs-agent/internal/pkg/enums"
	"testing"
	"time"

	"github.com/mlogclub/simple/sqls"
)

func TestBuildDefaultSubject(t *testing.T) {
	tests := []struct {
		name     string
		operator *dto.AuthPrincipal
		expected string
	}{
		{
			name: "访客-正常ID",
			operator: &dto.AuthPrincipal{
				IsVisitor: true,
				VisitorID: "visitor_550e8400-e29b-41d4-a716-446655440000",
			},
			expected: "访客a3f5b2c1",
		},
		{
			name: "访客-空ID",
			operator: &dto.AuthPrincipal{
				IsVisitor: true,
				VisitorID: "",
			},
			expected: "访客unknown",
		},
		{
			name: "登录用户-有用户名",
			operator: &dto.AuthPrincipal{
				IsVisitor: false,
				UserID:    1001,
				Username:  "张三",
			},
			expected: "张三",
		},
		{
			name: "登录用户-无用户名",
			operator: &dto.AuthPrincipal{
				IsVisitor: false,
				UserID:    1002,
				Username:  "",
			},
			expected: "用户1002",
		},
		{
			name:     "operator为nil",
			operator: nil,
			expected: "访客unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newConversationService()
			result := svc.buildDefaultSubject(tt.operator)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestHashUUID(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"visitor_550e8400-e29b-41d4-a716-446655440000", "a3f5b2c1"},
		{"", "unknown"},
		{"abc", "90015098"},
	}

	for _, tt := range tests {
		result := hashUUID(tt.input)
		if result != tt.expected {
			t.Errorf("input=%s, expected %s, got %s", tt.input, tt.expected, result)
		}
	}
}

func TestFindAgentConversationPageAppliesFilterAndSort(t *testing.T) {
	setupIMServiceTestDB(t)

	now := time.Now()
	seedConversationListTestItem(t, &models.Conversation{
		ID:                101,
		Subject:           "mine active older",
		Status:            enums.IMConversationStatusActive,
		CurrentAssigneeID: 9001,
		LastMessageAt:     timePtr(now.Add(-2 * time.Hour)),
	})
	seedConversationListTestItem(t, &models.Conversation{
		ID:                102,
		Subject:           "mine active newer",
		Status:            enums.IMConversationStatusActive,
		CurrentAssigneeID: 9001,
		LastMessageAt:     timePtr(now.Add(-1 * time.Hour)),
	})
	seedConversationListTestItem(t, &models.Conversation{
		ID:                103,
		Subject:           "mine pending first",
		Status:            enums.IMConversationStatusPending,
		CurrentAssigneeID: 9001,
	})
	seedConversationListTestItem(t, &models.Conversation{
		ID:                104,
		Subject:           "mine closed",
		Status:            enums.IMConversationStatusClosed,
		CurrentAssigneeID: 9001,
		LastMessageAt:     timePtr(now.Add(-30 * time.Minute)),
	})
	seedConversationListTestItem(t, &models.Conversation{
		ID:                99,
		Subject:           "mine pending earlier",
		Status:            enums.IMConversationStatusPending,
		CurrentAssigneeID: 9001,
	})
	seedConversationListTestItem(t, &models.Conversation{
		ID:                105,
		Subject:           "other active",
		Status:            enums.IMConversationStatusActive,
		CurrentAssigneeID: 9002,
		LastMessageAt:     timePtr(now),
	})

	operator := &dto.AuthPrincipal{UserID: 9001, Username: "agent_9001"}

	mineList, _, err := ConversationService.FindAgentConversationPage(operator, request.AgentConversationFilterMine, "", 1, 100)
	if err != nil {
		t.Fatalf("find mine conversations failed: %v", err)
	}
	if len(mineList) != 5 {
		t.Fatalf("expected 5 mine conversations, got %d", len(mineList))
	}
	if mineList[0].ID != 104 || mineList[1].ID != 102 || mineList[2].ID != 101 {
		t.Fatalf("unexpected mine order: got ids [%d %d %d]", mineList[0].ID, mineList[1].ID, mineList[2].ID)
	}

	activeList, _, err := ConversationService.FindAgentConversationPage(operator, request.AgentConversationFilterActive, "", 1, 100)
	if err != nil {
		t.Fatalf("find active conversations failed: %v", err)
	}
	if len(activeList) != 2 || activeList[0].ID != 102 || activeList[1].ID != 101 {
		t.Fatalf("unexpected active result order")
	}

	pendingList, _, err := ConversationService.FindAgentConversationPage(operator, request.AgentConversationFilterPending, "", 1, 100)
	if err != nil {
		t.Fatalf("find pending conversations failed: %v", err)
	}
	if len(pendingList) != 2 || pendingList[0].ID != 99 || pendingList[1].ID != 103 {
		t.Fatalf("unexpected pending result order")
	}

	closedList, _, err := ConversationService.FindAgentConversationPage(operator, request.AgentConversationFilterClosed, "", 1, 100)
	if err != nil {
		t.Fatalf("find closed conversations failed: %v", err)
	}
	if len(closedList) != 1 || closedList[0].ID != 104 {
		t.Fatalf("unexpected closed result")
	}
}

func TestFindAgentConversationPageLimitsToHundred(t *testing.T) {
	setupIMServiceTestDB(t)

	for i := 0; i < 101; i++ {
		id := int64(2000 + i)
		seedConversationListTestItem(t, &models.Conversation{
			ID:                id,
			Subject:           "limit test",
			Status:            enums.IMConversationStatusActive,
			CurrentAssigneeID: 9101,
			LastMessageAt:     timePtr(time.Now().Add(time.Duration(i) * time.Minute)),
		})
	}

	list, paging, err := ConversationService.FindAgentConversationPage(
		&dto.AuthPrincipal{UserID: 9101, Username: "agent_9101"},
		request.AgentConversationFilterMine,
		"",
		1,
		999,
	)
	if err != nil {
		t.Fatalf("find mine conversations failed: %v", err)
	}
	if len(list) != 100 {
		t.Fatalf("expected 100 conversations, got %d", len(list))
	}
	if paging == nil || paging.Limit != 100 {
		t.Fatalf("expected paging limit 100, got %+v", paging)
	}
}

func seedConversationListTestItem(t *testing.T, item *models.Conversation) {
	t.Helper()

	now := time.Now()
	if item.ChannelType == "" {
		item.ChannelType = enums.IMConversationChannelWebChat
	}
	if item.Subject == "" {
		item.Subject = "test conversation"
	}
	item.AuditFields.CreatedAt = now
	item.AuditFields.UpdatedAt = now
	item.AuditFields.CreateUserName = "tester"
	item.AuditFields.UpdateUserName = "tester"

	if err := sqls.DB().Create(item).Error; err != nil {
		t.Fatalf("seed conversation failed: %v", err)
	}
}

func timePtr(v time.Time) *time.Time {
	return &v
}
