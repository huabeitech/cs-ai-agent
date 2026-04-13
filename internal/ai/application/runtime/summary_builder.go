package runtime

import (
	"strings"

	runtimeexecutor "cs-agent/internal/ai/runtime/executor"
)

func toSummary(summary *runtimeexecutor.RunResult) *Summary {
	if summary == nil {
		return nil
	}
	ret := &Summary{
		RunID:                 summary.RunID,
		Status:                summary.Status,
		ReplyText:             summary.ReplyText,
		PlannedSkillCode:      strings.TrimSpace(summary.SelectedSkillCode),
		PlannedSkillName:      strings.TrimSpace(summary.SelectedSkillName),
		PlanReason:            strings.TrimSpace(summary.SkillRouteReason),
		SkillRouteTrace:       strings.TrimSpace(summary.SkillRouteTrace),
		SkillAllowedToolCodes: append([]string(nil), summary.SkillAllowedToolCodes...),
		ModelName:             summary.ModelName,
		PromptTokens:          summary.PromptTokens,
		CompletionTokens:      summary.CompletionTokens,
		HistoryMessageCount:   summary.HistoryMessageCount,
		RetrieverCount:        summary.RetrieverCount,
		ToolCallCount:         summary.ToolCallCount,
		ToolCodes:             append([]string(nil), summary.ToolCodes...),
		InvokedToolCodes:      append([]string(nil), summary.InvokedToolCodes...),
		CheckPointID:          summary.CheckPointID,
		Interrupted:           summary.Interrupted,
		TraceData:             summary.TraceData,
		ErrorMessage:          summary.ErrorMessage,
	}
	if len(summary.Interrupts) > 0 {
		ret.Interrupts = make([]InterruptContextSummary, 0, len(summary.Interrupts))
		for _, item := range summary.Interrupts {
			ret.Interrupts = append(ret.Interrupts, InterruptContextSummary{
				Type:        item.Type,
				ID:          item.ID,
				InfoPreview: item.InfoPreview,
			})
		}
	}
	return ret
}
