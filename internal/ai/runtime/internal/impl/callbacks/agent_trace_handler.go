package callbacks

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"cs-agent/internal/pkg/toolx"

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
		if metadata, ok := h.resolveToolMetadata(item.ToolName); ok && strings.TrimSpace(metadata.ToolCode) == toolx.BuiltinToolSearchToolCode {
			h.collector.AddToolSearchItem(h.buildToolSearchTraceItem(argumentsInJSON, result, err))
		}
		return result, err
	}, nil
}

func (h *RuntimeTraceHandler) resolveToolMetadata(modelToolName string) (ToolMetadata, bool) {
	if h == nil || h.toolMetadataBy == nil {
		return ToolMetadata{}, false
	}
	modelToolName = strings.TrimSpace(modelToolName)
	if modelToolName == "" {
		return ToolMetadata{}, false
	}
	if modelToolName == toolx.BuiltinToolSearchToolName {
		return ToolMetadata{
			ToolCode:   toolx.BuiltinToolSearchToolCode,
			ServerCode: toolx.BuiltinToolCatalogServerCode,
			ToolName:   toolx.BuiltinToolSearchToolName,
		}, true
	}
	metadata, ok := h.toolMetadataBy[modelToolName]
	return metadata, ok
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

func (h *RuntimeTraceHandler) buildToolSearchTraceItem(argumentsInJSON string, result string, runErr error) ToolSearchTraceItem {
	item := ToolSearchTraceItem{
		Action: "search",
		Status: "ok",
	}
	args := parseToolArguments(argumentsInJSON)
	item.Query = strings.TrimSpace(readToolSearchString(args, "regex_pattern"))
	if runErr != nil {
		item.Status = "error"
		item.ErrorMessage = runErr.Error()
		return item
	}
	payload := make(map[string]any)
	if err := json.Unmarshal([]byte(strings.TrimSpace(result)), &payload); err != nil {
		return item
	}
	selectedTools, _ := payload["selectedTools"].([]any)
	item.CandidateToolCodes = h.extractSelectedToolCodes(selectedTools)
	return item
}

func readToolSearchString(data map[string]any, key string) string {
	if len(data) == 0 {
		return ""
	}
	value, ok := data[key]
	if !ok {
		return ""
	}
	text, _ := value.(string)
	return text
}

func (h *RuntimeTraceHandler) extractSelectedToolCodes(items []any) []string {
	if len(items) == 0 {
		return nil
	}
	ret := make([]string, 0, len(items))
	for _, item := range items {
		toolName, ok := item.(string)
		if !ok {
			continue
		}
		toolName = strings.TrimSpace(toolName)
		if toolName == "" {
			continue
		}
		toolCode := toolName
		if metadata, ok := h.resolveToolMetadata(toolName); ok && strings.TrimSpace(metadata.ToolCode) != "" {
			toolCode = strings.TrimSpace(metadata.ToolCode)
		}
		if toolCode == "" {
			continue
		}
		ret = append(ret, toolCode)
	}
	return ret
}
