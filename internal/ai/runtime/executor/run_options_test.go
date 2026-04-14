package executor

import "testing"

func TestBuildResumeTargets(t *testing.T) {
	targets := buildResumeTargets(map[string]string{
		" interrupt-1 ": "确认",
		"":              "ignored",
		"   ":           "ignored",
		"interrupt-2":   "取消",
	})

	if len(targets) != 2 {
		t.Fatalf("expected 2 resume targets, got %d", len(targets))
	}
	if got := targets["interrupt-1"]; got != "确认" {
		t.Fatalf("unexpected target data for interrupt-1: %#v", got)
	}
	if got := targets["interrupt-2"]; got != "取消" {
		t.Fatalf("unexpected target data for interrupt-2: %#v", got)
	}
}

func TestBuildResumeTargetsEmpty(t *testing.T) {
	if got := buildResumeTargets(nil); got != nil {
		t.Fatalf("expected nil targets for nil input, got %#v", got)
	}
	if got := buildResumeTargets(map[string]string{
		"":    "ignored",
		"   ": "ignored",
	}); got != nil {
		t.Fatalf("expected nil targets for blank keys, got %#v", got)
	}
}
