package graphs

import (
	"testing"

	"cs-agent/internal/models"
)

func TestHandoffGraphBuildReason(t *testing.T) {
	graph := NewHandoffGraph(&models.Conversation{ID: 1}, &models.AIAgent{Name: "AI"})

	reason, err := graph.buildReason(`{"reason":"  用户需要人工确认  "}`)
	if err != nil {
		t.Fatalf("buildReason returned error: %v", err)
	}
	if reason != "用户需要人工确认" {
		t.Fatalf("unexpected reason: %q", reason)
	}
}

func TestHandoffGraphBuildReasonFallback(t *testing.T) {
	graph := NewHandoffGraph(&models.Conversation{ID: 1}, &models.AIAgent{Name: "AI"})

	reason, err := graph.buildReason(`{}`)
	if err != nil {
		t.Fatalf("buildReason returned error: %v", err)
	}
	if reason != "用户需要转人工支持" {
		t.Fatalf("unexpected fallback reason: %q", reason)
	}
}
