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

type graphAnalyzeConversationResult struct {
	RecommendedNextAction string `json:"recommendedNextAction"`
	RiskLevel             string `json:"riskLevel"`
}

type graphTriageAnalysisResult struct {
	RiskLevel string `json:"riskLevel"`
}

type graphTriageTicketDraftResult struct {
	Ready bool `json:"ready"`
}

type graphTriageServiceRequestResult struct {
	RecommendedAction string                        `json:"recommendedAction"`
	Analysis          graphTriageAnalysisResult     `json:"analysis"`
	TicketDraft       *graphTriageTicketDraftResult `json:"ticketDraft"`
}

type toolSearchArguments struct {
	Query        string `json:"query"`
	RegexPattern string `json:"regex_pattern"`
	ToolCode     string `json:"toolCode"`
}

type toolSearchCandidateResult struct {
	ToolCode string `json:"toolCode"`
}

type toolSearchInvokeResult struct {
	SelectedTools []string `json:"selectedTools"`
}

type toolSearchSearchResult struct {
	Candidates []toolSearchCandidateResult `json:"candidates"`
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
		if metadata, ok := h.resolveToolMetadata(item.ToolName); ok && strings.TrimSpace(metadata.SourceType) == toolx.GraphCreateTicketConfirm.ServerCode {
			recommendedAction, riskLevel, ticketDraftReady := parseGraphToolOutcome(item.ToolCode, result)
			h.collector.AddGraphToolItem(GraphToolTraceItem{
				ToolCode:          item.ToolCode,
				ToolName:          item.ToolName,
				Arguments:         item.Arguments,
				ResultPreview:     item.ResultPreview,
				ResultReduced:     item.ResultReduced,
				OriginalChars:     item.OriginalChars,
				KeptChars:         item.KeptChars,
				LatencyMs:         item.LatencyMs,
				Status:            item.Status,
				ErrorMessage:      item.ErrorMessage,
				RecommendedAction: recommendedAction,
				RiskLevel:         riskLevel,
				TicketDraftReady:  ticketDraftReady,
			})
		}
		if metadata, ok := h.resolveToolMetadata(item.ToolName); ok && strings.TrimSpace(metadata.ToolCode) == toolx.BuiltinToolSearch.Code {
			h.collector.AddToolSearchItem(h.buildToolSearchTraceItem(argumentsInJSON, result, err))
		}
		return result, err
	}, nil
}

func parseGraphToolOutcome(toolCode string, result string) (recommendedAction, riskLevel string, ticketDraftReady bool) {
	toolCode = strings.TrimSpace(toolCode)
	if toolCode == "" || strings.TrimSpace(result) == "" {
		return "", "", false
	}
	switch toolCode {
	case toolx.GraphAnalyzeConversation.Code:
		var payload graphAnalyzeConversationResult
		if err := json.Unmarshal([]byte(strings.TrimSpace(result)), &payload); err != nil {
			return "", "", false
		}
		return strings.TrimSpace(payload.RecommendedNextAction), strings.TrimSpace(payload.RiskLevel), false
	case toolx.GraphTriageServiceRequest.Code:
		var payload graphTriageServiceRequestResult
		if err := json.Unmarshal([]byte(strings.TrimSpace(result)), &payload); err != nil {
			return "", "", false
		}
		recommendedAction = strings.TrimSpace(payload.RecommendedAction)
		riskLevel = strings.TrimSpace(payload.Analysis.RiskLevel)
		if payload.TicketDraft != nil {
			ticketDraftReady = payload.TicketDraft.Ready
		}
		return recommendedAction, riskLevel, ticketDraftReady
	default:
		return "", "", false
	}
}

func (h *RuntimeTraceHandler) resolveToolMetadata(modelToolName string) (ToolMetadata, bool) {
	if h == nil || h.toolMetadataBy == nil {
		return ToolMetadata{}, false
	}
	modelToolName = strings.TrimSpace(modelToolName)
	if modelToolName == "" {
		return ToolMetadata{}, false
	}
	if spec, ok := toolx.GetRegisteredToolSpecByName(modelToolName); ok {
		resolved := toolx.ResolveToolMetadata(spec.Code, spec.Name)
		return ToolMetadata{
			ToolCode:   resolved.ToolCode,
			ServerCode: resolved.ServerCode,
			ToolName:   resolved.ToolName,
			SourceType: resolved.SourceType,
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
	var args toolSearchArguments
	if strings.TrimSpace(argumentsInJSON) != "" {
		_ = json.Unmarshal([]byte(argumentsInJSON), &args)
	}
	item.Query = strings.TrimSpace(firstNonBlank(args.Query, args.RegexPattern))
	item.TargetToolCode = strings.TrimSpace(args.ToolCode)
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
	item.CandidateToolCodes = h.extractCandidateToolCodes(result)
	return item
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

func (h *RuntimeTraceHandler) extractCandidateToolCodes(result string) []string {
	result = strings.TrimSpace(result)
	if result == "" {
		return nil
	}
	var invokePayload toolSearchInvokeResult
	if err := json.Unmarshal([]byte(result), &invokePayload); err == nil && len(invokePayload.SelectedTools) > 0 {
		return h.extractSelectedToolCodes(invokePayload.SelectedTools)
	}
	var searchPayload toolSearchSearchResult
	if err := json.Unmarshal([]byte(result), &searchPayload); err == nil && len(searchPayload.Candidates) > 0 {
		return extractCandidateObjectCodes(searchPayload.Candidates)
	}
	return nil
}

func (h *RuntimeTraceHandler) extractSelectedToolCodes(items []string) []string {
	if len(items) == 0 {
		return nil
	}
	ret := make([]string, 0, len(items))
	for _, item := range items {
		toolName := strings.TrimSpace(item)
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

func extractCandidateObjectCodes(items []toolSearchCandidateResult) []string {
	if len(items) == 0 {
		return nil
	}
	ret := make([]string, 0, len(items))
	for _, item := range items {
		toolCode := strings.TrimSpace(item.ToolCode)
		if toolCode == "" {
			continue
		}
		ret = append(ret, toolCode)
	}
	return ret
}
