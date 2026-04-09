package callbacks

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	einotool "github.com/cloudwego/eino/components/tool"

	"github.com/cloudwego/eino/adk"
)

type ToolMetadata struct {
	ToolCode   string
	ServerCode string
	ToolName   string
}

type RuntimeTraceHandler struct {
	*adk.BaseChatModelAgentMiddleware
	collector      *RuntimeTraceCollector
	toolMetadataBy map[string]ToolMetadata
}

func NewRuntimeTraceHandler(collector *RuntimeTraceCollector, toolMetadataBy map[string]ToolMetadata) *RuntimeTraceHandler {
	return &RuntimeTraceHandler{
		BaseChatModelAgentMiddleware: &adk.BaseChatModelAgentMiddleware{},
		collector:                    collector,
		toolMetadataBy:               toolMetadataBy,
	}
}

func (h *RuntimeTraceHandler) WrapInvokableToolCall(_ context.Context, endpoint adk.InvokableToolCallEndpoint, tCtx *adk.ToolContext) (adk.InvokableToolCallEndpoint, error) {
	return func(ctx context.Context, argumentsInJSON string, opts ...einotool.Option) (string, error) {
		startedAt := time.Now()
		result, err := endpoint(ctx, argumentsInJSON, opts...)
		item := ToolTraceItem{
			ResultPreview: previewToolText(result, 300),
			LatencyMs:     time.Since(startedAt).Milliseconds(),
			Status:        "ok",
		}
		if tCtx != nil {
			item.ToolName = strings.TrimSpace(tCtx.Name)
			if metadata, ok := h.toolMetadataBy[item.ToolName]; ok {
				item.ToolCode = metadata.ToolCode
				item.ServerCode = metadata.ServerCode
				item.ToolName = metadata.ToolName
			}
		}
		if arguments := parseToolArguments(argumentsInJSON); len(arguments) > 0 {
			item.Arguments = arguments
		}
		if err != nil {
			item.Status = "error"
			item.ErrorMessage = err.Error()
		}
		h.collector.AddToolItem(item)
		return result, err
	}, nil
}

func parseToolArguments(argumentsInJSON string) map[string]any {
	argumentsInJSON = strings.TrimSpace(argumentsInJSON)
	if argumentsInJSON == "" {
		return nil
	}
	ret := make(map[string]any)
	if err := json.Unmarshal([]byte(argumentsInJSON), &ret); err != nil {
		return nil
	}
	return ret
}

func previewToolText(text string, limit int) string {
	if limit <= 0 {
		return ""
	}
	text = strings.TrimSpace(text)
	runes := []rune(text)
	if len(runes) <= limit {
		return text
	}
	return string(runes[:limit]) + "..."
}
