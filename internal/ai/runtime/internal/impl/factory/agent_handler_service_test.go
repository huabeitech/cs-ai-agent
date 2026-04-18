package factory

import (
	"context"
	"testing"

	einocallbacks "cs-agent/internal/ai/runtime/internal/impl/callbacks"
)

func TestAgentHandlerServiceBuildWithCollectorOnly(t *testing.T) {
	collector := einocallbacks.NewRuntimeTraceCollector()
	service := NewAgentHandlerService(nil)

	handlers, err := service.Build(context.Background(), BuildAgentHandlersInput{
		Collector: collector,
		InstructionSummary: einocallbacks.InstructionTraceSummary{
			SectionTitles:     []string{"系统治理规则"},
			HasGovernanceRule: true,
		},
	})
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}
	if len(handlers) != 1 {
		t.Fatalf("expected 1 handler, got %d", len(handlers))
	}
	if !collector.Data.Instruction.HasGovernanceRule {
		t.Fatalf("instruction summary was not written to collector: %#v", collector.Data.Instruction)
	}
	if len(collector.Data.Instruction.SectionTitles) != 1 {
		t.Fatalf("unexpected section titles: %#v", collector.Data.Instruction.SectionTitles)
	}
}

func TestAgentHandlerServiceBuildWithEmptyInput(t *testing.T) {
	service := NewAgentHandlerService(nil)

	handlers, err := service.Build(context.Background(), BuildAgentHandlersInput{})
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}
	if len(handlers) != 0 {
		t.Fatalf("expected no handlers, got %d", len(handlers))
	}
}
