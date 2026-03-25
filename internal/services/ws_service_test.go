package services

import (
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/dto"
	"testing"
)

func TestNormalizeRealtimeTopics(t *testing.T) {
	got := normalizeRealtimeTopics([]string{
		"",
		" conversation:1 ",
		"conversation:1",
		"user:2",
		"user:2",
	})

	if len(got) != 2 {
		t.Fatalf("expected 2 topics, got %d", len(got))
	}
	if got[0] != "conversation:1" {
		t.Fatalf("unexpected first topic: %s", got[0])
	}
	if got[1] != "user:2" {
		t.Fatalf("unexpected second topic: %s", got[1])
	}
}

func TestDefaultTopics(t *testing.T) {
	svc := newWsService()

	visitorClient := &ClientSession{
		Role: realtimeRoleUser,
		Principal: &dto.AuthPrincipal{
			IsVisitor: true,
			VisitorID: "visitor_1",
		},
		Topics: make(map[string]struct{}),
	}
	visitorTopics := svc.defaultTopics(visitorClient)
	if len(visitorTopics) != 1 || visitorTopics[0] != "visitor:visitor_1" {
		t.Fatalf("unexpected visitor topics: %#v", visitorTopics)
	}

	adminClient := &ClientSession{
		Role: realtimeRoleAdmin,
		Principal: &dto.AuthPrincipal{
			UserID: 99,
		},
		Topics: make(map[string]struct{}),
	}
	adminTopics := svc.defaultTopics(adminClient)
	if len(adminTopics) != 2 {
		t.Fatalf("unexpected admin topic count: %#v", adminTopics)
	}
	if adminTopics[0] != "admin:99" || adminTopics[1] != "admin:all" {
		t.Fatalf("unexpected admin topics: %#v", adminTopics)
	}
}

func TestRouteConversationTopics(t *testing.T) {
	svc := newWsService()

	conversation := &models.Conversation{
		ID:                12,
		SourceUserID:      7,
		ExternalUserID:    "visitor_x",
		CurrentAssigneeID: 3,
	}

	got := svc.routeConversationTopics(conversation)
	expected := []string{
		"conversation:12",
		"user:7",
		"visitor:visitor_x",
		"admin:3",
	}

	if len(got) != len(expected) {
		t.Fatalf("unexpected topic count: %#v", got)
	}
	for i, item := range expected {
		if got[i] != item {
			t.Fatalf("unexpected topic at %d: %s", i, got[i])
		}
	}
}

func TestRouteConversationTopicsFallbackToAdminAll(t *testing.T) {
	svc := newWsService()

	conversation := &models.Conversation{
		ID:             18,
		ExternalUserID: "visitor_y",
	}

	got := svc.routeConversationTopics(conversation)
	expected := []string{
		"conversation:18",
		"visitor:visitor_y",
		"admin:all",
	}

	if len(got) != len(expected) {
		t.Fatalf("unexpected topic count: %#v", got)
	}
	for i, item := range expected {
		if got[i] != item {
			t.Fatalf("unexpected topic at %d: %s", i, got[i])
		}
	}
}

func TestRealtimeClientEnqueueBuffer(t *testing.T) {
	client := &ClientSession{
		Send:   make(chan []byte, 1),
		Topics: make(map[string]struct{}),
	}

	if ok := client.enqueue([]byte("first")); !ok {
		t.Fatal("expected first enqueue to succeed")
	}
	if ok := client.enqueue([]byte("second")); ok {
		t.Fatal("expected second enqueue to fail when buffer is full")
	}
}

func TestConnectionManagerSubscribeAndUnregister(t *testing.T) {
	manager := newWsConnectionManager()
	session := &ClientSession{
		ID:     "conn_1",
		Topics: make(map[string]struct{}),
	}

	if count := manager.Register(session, []string{"user:1"}); count != 1 {
		t.Fatalf("expected session count 1, got %d", count)
	}
	applied := manager.Subscribe(session, []string{"conversation:9"})
	if len(applied) != 1 || applied[0] != "conversation:9" {
		t.Fatalf("unexpected subscribed topics: %#v", applied)
	}

	targets := manager.FindByTopics([]string{"user:1", "conversation:9"})
	if len(targets) != 1 || targets[0].ID != "conn_1" {
		t.Fatalf("unexpected targets: %#v", targets)
	}

	if count := manager.Unregister(session); count != 0 {
		t.Fatalf("expected session count 0, got %d", count)
	}
	if len(manager.FindByTopics([]string{"user:1", "conversation:9"})) != 0 {
		t.Fatal("expected topics to be cleaned after unregister")
	}
}
