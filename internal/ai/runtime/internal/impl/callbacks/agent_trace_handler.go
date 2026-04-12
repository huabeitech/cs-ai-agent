package callbacks

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	impladapter "cs-agent/internal/ai/runtime/internal/impl/adapter"
	"cs-agent/internal/pkg/toolx"

	einotool "github.com/cloudwego/eino/components/tool"

	"github.com/cloudwego/eino/adk"
)

type ToolMetadata struct {
	ToolCode   string
	ServerCode string
	ToolName   string
	SourceType string
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
		reductionInfo := impladapter.ParseReductionInfo(result)
		item.ResultReduced = reductionInfo.Reduced
		item.OriginalChars = reductionInfo.OriginalChars
		item.KeptChars = reductionInfo.KeptChars
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
		if metadata, ok := h.resolveToolMetadata(item.ToolName); ok && strings.TrimSpace(metadata.SourceType) == toolx.GraphToolCatalogServerCode {
			h.collector.AddGraphToolItem(GraphToolTraceItem{
				ToolCode:      item.ToolCode,
				ToolName:      item.ToolName,
				Arguments:     item.Arguments,
				ResultPreview: item.ResultPreview,
				ResultReduced: item.ResultReduced,
				OriginalChars: item.OriginalChars,
				KeptChars:     item.KeptChars,
				LatencyMs:     item.LatencyMs,
				Status:        item.Status,
				ErrorMessage:  item.ErrorMessage,
			})
		}
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
			SourceType: toolx.BuiltinToolCatalogServerCode,
		}, true
	}
	if modelToolName == toolx.BuiltinSkillToolName {
		return ToolMetadata{
			ToolCode:   toolx.BuiltinSkillToolCode,
			ServerCode: toolx.BuiltinToolCatalogServerCode,
			ToolName:   toolx.BuiltinSkillToolName,
			SourceType: toolx.BuiltinToolCatalogServerCode,
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
	item := ToolSearchTraceItem{Status: "ok"}
	args := parseToolArguments(argumentsInJSON)
	item.Query = strings.TrimSpace(firstNonBlank(
		readToolSearchString(args, "query"),
		readToolSearchString(args, "regex_pattern"),
	))
	item.TargetToolCode = strings.TrimSpace(readToolSearchString(args, "toolCode"))
	item.TargetServerCode, item.TargetToolName = toolx.SplitMCPToolCode(item.TargetToolCode)
	if item.TargetToolCode != "" {
		item.Action = "invoke"
	} else {
		item.Action = "search"
	}
	if runErr != nil {
		item.Status = "error"
		item.ErrorMessage = runErr.Error()
		return item
	}
	payload := make(map[string]any)
	if err := json.Unmarshal([]byte(strings.TrimSpace(result)), &payload); err != nil {
		return item
	}
	item.CandidateToolCodes = h.extractCandidateToolCodes(payload)
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

func firstNonBlank(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}

func (h *RuntimeTraceHandler) extractCandidateToolCodes(payload map[string]any) []string {
	if len(payload) == 0 {
		return nil
	}
	if items, ok := payload["selectedTools"].([]any); ok {
		return h.extractSelectedToolCodes(items)
	}
	if items, ok := payload["candidates"].([]any); ok {
		return h.extractCandidateObjects(items)
	}
	return nil
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

func (h *RuntimeTraceHandler) extractCandidateObjects(items []any) []string {
	if len(items) == 0 {
		return nil
	}
	ret := make([]string, 0, len(items))
	for _, item := range items {
		obj, ok := item.(map[string]any)
		if !ok {
			continue
		}
		toolCode := strings.TrimSpace(readToolSearchString(obj, "toolCode"))
		if toolCode == "" {
			continue
		}
		ret = append(ret, toolCode)
	}
	return ret
}
