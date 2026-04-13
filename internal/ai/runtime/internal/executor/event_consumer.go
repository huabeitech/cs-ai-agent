package executor

import (
	"strings"

	"cs-agent/internal/ai/runtime/internal/impl/callbacks"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"
)

func consumeAgentEvents(events *adk.AsyncIterator[*adk.AgentEvent], summary *RunResult, collector *callbacks.RuntimeTraceCollector, toolDefsByModelName map[string]string) {
	if summary == nil {
		return
	}
	if collector == nil {
		collector = callbacks.NewRuntimeTraceCollector()
	}
	for {
		event, ok := events.Next()
		if !ok {
			break
		}
		if event == nil {
			continue
		}
		if event.Action != nil && event.Action.Interrupted != nil {
			summary.Status = "interrupted"
			summary.Interrupted = true
			summary.Interrupts = buildInterruptSummaries(event)
		}
		if event.Err != nil {
			errMsg := strings.TrimSpace(event.Err.Error())
			if errMsg != "" {
				summary.Status = "error"
				summary.ErrorMessage = errMsg
			}
		}
		if event.Output == nil || event.Output.MessageOutput == nil {
			continue
		}
		messageOutput := event.Output.MessageOutput
		switch messageOutput.Role {
		case schema.Assistant:
			replyText := strings.TrimSpace(messageOutput.Message.Content)
			if replyText != "" {
				summary.ReplyText = replyText
			}
		case schema.Tool:
			toolName := strings.TrimSpace(messageOutput.ToolName)
			if toolName == "" {
				continue
			}
			toolCode := toolName
			if mappedCode, ok := toolDefsByModelName[toolName]; ok && strings.TrimSpace(mappedCode) != "" {
				toolCode = strings.TrimSpace(mappedCode)
			}
			summary.InvokedToolCodes = appendIfMissing(summary.InvokedToolCodes, toolCode)
		}
	}
	if summary.Status == "started" {
		switch {
		case strings.TrimSpace(summary.ErrorMessage) != "":
			summary.Status = "error"
		case summary.Interrupted:
			summary.Status = "interrupted"
		case strings.TrimSpace(summary.ReplyText) != "":
			summary.Status = "completed"
		default:
			summary.Status = "fallback"
		}
	}
	summary.ToolCallCount = len(summary.InvokedToolCodes)
}

func buildInterruptSummaries(event *adk.AgentEvent) []InterruptContextSummary {
	if event == nil || event.Action == nil || event.Action.Interrupted == nil {
		return nil
	}
	interrupts := event.Action.Interrupted.InterruptContexts
	result := make([]InterruptContextSummary, 0, len(interrupts))
	for _, item := range interrupts {
		if item == nil {
			continue
		}
		result = append(result, InterruptContextSummary{
			ID:          strings.TrimSpace(item.ID),
			InfoPreview: previewInterruptInfo(item.Info),
		})
	}
	return result
}
