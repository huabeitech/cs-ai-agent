package graphs

import (
	"testing"

	"cs-agent/internal/models"
)

func TestCreateTicketGraphBuildCreateRequest(t *testing.T) {
	graph := NewCreateTicketGraph(&models.Conversation{
		ID:                 12,
		Subject:            "fallback-title",
		LastMessageSummary: "fallback-description",
	}, &models.AIAgent{Name: "AI"})

	req, err := graph.buildCreateRequest(`{"title":"  test title ","description":" desc ","priority":2,"severity":3}`)
	if err != nil {
		t.Fatalf("buildCreateRequest returned error: %v", err)
	}
	if req.Title != "test title" || req.Description != "desc" {
		t.Fatalf("unexpected request text fields: %#v", req)
	}
	if req.Priority != 2 || req.Severity != 3 {
		t.Fatalf("unexpected request numeric fields: %#v", req)
	}
}

func TestCreateTicketGraphBuildCreateRequestFallbacks(t *testing.T) {
	graph := NewCreateTicketGraph(&models.Conversation{
		ID:                 12,
		Subject:            "fallback-title",
		LastMessageSummary: "fallback-description",
	}, &models.AIAgent{Name: "AI"})

	req, err := graph.buildCreateRequest(`{}`)
	if err != nil {
		t.Fatalf("buildCreateRequest returned error: %v", err)
	}
	if req.Title != "fallback-title" {
		t.Fatalf("unexpected fallback title: %#v", req)
	}
	if req.Description != "fallback-description" {
		t.Fatalf("unexpected fallback description: %#v", req)
	}
}
