package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"cs-agent/internal/ai/mcps"
	"cs-agent/internal/ai/runtime/registry"
	"cs-agent/internal/pkg/toolx"

	einotool "github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	einojsonschema "github.com/eino-contrib/jsonschema"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

const (
	ToolSearchToolCode = toolx.BuiltinToolSearchToolCode
	ToolSearchToolName = toolx.BuiltinToolSearchToolName
)

type ToolSearchTool struct {
	allowedToolCodes []string
}

func NewToolSearchTool() *ToolSearchTool {
	return &ToolSearchTool{}
}

func (t *ToolSearchTool) Name() string {
	return ToolSearchToolName
}

func (t *ToolSearchTool) Code() string {
	return ToolSearchToolCode
}

func (t *ToolSearchTool) Enabled(ctx registry.Context) bool {
	return len(filterAllowedMCPToolCodes(ctx.AllowedToolCodes)) > 0
}

func (t *ToolSearchTool) Build(ctx registry.Context) (einotool.BaseTool, error) {
	if !t.Enabled(ctx) {
		return nil, nil
	}
	return &ToolSearchTool{
		allowedToolCodes: filterAllowedMCPToolCodes(ctx.AllowedToolCodes),
	}, nil
}

func (t *ToolSearchTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: ToolSearchToolName,
		Desc: "当你需要使用当前会话允许的长尾 MCP 工具时，先调用本工具搜索合适的 toolCode；确认目标后，可再次调用本工具并传入 toolCode 与 arguments 代理执行。不要用它替代明确固定的内置流程工具。",
		ParamsOneOf: schema.NewParamsOneOfByJSONSchema(&einojsonschema.Schema{
			Version: einojsonschema.Version,
			Type:    "object",
			Properties: orderedmap.New[string, *einojsonschema.Schema](orderedmap.WithInitialData(
				orderedmap.Pair[string, *einojsonschema.Schema]{
					Key: "query",
					Value: &einojsonschema.Schema{
						Type:        "string",
						Description: "要搜索的工具意图、能力或关键词；当只想列出候选工具时使用。",
					},
				},
				orderedmap.Pair[string, *einojsonschema.Schema]{
					Key: "toolCode",
					Value: &einojsonschema.Schema{
						Type:        "string",
						Description: "已确定目标后要调用的 MCP toolCode，例如 mcp_server/tool_name。",
					},
				},
				orderedmap.Pair[string, *einojsonschema.Schema]{
					Key: "arguments",
					Value: &einojsonschema.Schema{
						Type:                 "object",
						Description:          "调用目标工具时传入的参数对象。",
						AdditionalProperties: &einojsonschema.Schema{},
					},
				},
			)),
		}),
		Extra: map[string]any{
			"toolCode": ToolSearchToolCode,
		},
	}, nil
}

func (t *ToolSearchTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...einotool.Option) (string, error) {
	if t == nil {
		return "", fmt.Errorf("tool search tool is nil")
	}
	req, err := parseToolSearchRequest(argumentsInJSON)
	if err != nil {
		return "", err
	}
	if req.ToolCode != "" {
		return t.invokeTargetTool(ctx, req.ToolCode, req.Arguments)
	}
	return t.searchCandidates(ctx, req.Query)
}

type toolSearchRequest struct {
	Query     string
	ToolCode  string
	Arguments map[string]any
}

type toolSearchCandidate struct {
	ToolCode    string `json:"toolCode"`
	ServerCode  string `json:"serverCode"`
	ToolName    string `json:"toolName"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
}

func parseToolSearchRequest(argumentsInJSON string) (*toolSearchRequest, error) {
	argumentsInJSON = strings.TrimSpace(argumentsInJSON)
	if argumentsInJSON == "" {
		return &toolSearchRequest{}, nil
	}
	raw := make(map[string]any)
	if err := json.Unmarshal([]byte(argumentsInJSON), &raw); err != nil {
		return nil, fmt.Errorf("invalid tool_search arguments: %w", err)
	}
	req := &toolSearchRequest{
		Query:    strings.TrimSpace(getStringValue(raw, "query")),
		ToolCode: strings.TrimSpace(getStringValue(raw, "toolCode")),
	}
	if value, ok := raw["arguments"]; ok {
		args, ok := value.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("tool_search arguments must be an object")
		}
		req.Arguments = args
	}
	return req, nil
}

func (t *ToolSearchTool) searchCandidates(ctx context.Context, query string) (string, error) {
	candidates, err := t.loadAllowedCandidates(ctx)
	if err != nil {
		return "", err
	}
	matched := filterCandidatesByQuery(candidates, query)
	if len(matched) == 0 {
		return "未找到匹配的动态工具，请换个关键词，或继续向用户追问后再搜索。", nil
	}
	if len(matched) > 8 {
		matched = matched[:8]
	}
	buf, err := json.Marshal(map[string]any{
		"query":      strings.TrimSpace(query),
		"total":      len(matched),
		"candidates": matched,
	})
	if err != nil {
		return "", err
	}
	return string(buf), nil
}

func (t *ToolSearchTool) invokeTargetTool(ctx context.Context, toolCode string, arguments map[string]any) (string, error) {
	toolCode = strings.TrimSpace(toolCode)
	serverCode, toolName := toolx.SplitMCPToolCode(toolCode)
	if serverCode == "" || toolName == "" {
		return "", fmt.Errorf("tool_search 只支持调用 MCP toolCode")
	}
	if !containsToolCode(t.allowedToolCodes, toolCode) {
		return "", fmt.Errorf("目标工具未被当前会话授权")
	}
	result, err := mcps.Runtime.CallTool(ctx, serverCode, toolName, cloneArguments(arguments))
	if err != nil {
		return "", err
	}
	return buildToolCallResultSummary(result), nil
}

func (t *ToolSearchTool) loadAllowedCandidates(ctx context.Context) ([]toolSearchCandidate, error) {
	serverToToolCodes := make(map[string]map[string]struct{})
	for _, toolCode := range t.allowedToolCodes {
		serverCode, toolName := toolx.SplitMCPToolCode(toolCode)
		if serverCode == "" || toolName == "" {
			continue
		}
		if _, ok := serverToToolCodes[serverCode]; !ok {
			serverToToolCodes[serverCode] = make(map[string]struct{})
		}
		serverToToolCodes[serverCode][toolCode] = struct{}{}
	}
	serverCodes := make([]string, 0, len(serverToToolCodes))
	for serverCode := range serverToToolCodes {
		serverCodes = append(serverCodes, serverCode)
	}
	slices.Sort(serverCodes)
	ret := make([]toolSearchCandidate, 0)
	for _, serverCode := range serverCodes {
		tools, err := mcps.Runtime.ListTools(ctx, serverCode)
		if err != nil {
			return nil, err
		}
		allowed := serverToToolCodes[serverCode]
		for _, item := range tools {
			toolCode := toolx.BuildMCPToolCode(serverCode, item.Name)
			if _, ok := allowed[toolCode]; !ok {
				continue
			}
			ret = append(ret, toolSearchCandidate{
				ToolCode:    toolCode,
				ServerCode:  serverCode,
				ToolName:    strings.TrimSpace(item.Name),
				Title:       strings.TrimSpace(item.Title),
				Description: strings.TrimSpace(item.Description),
			})
		}
	}
	return ret, nil
}

func filterAllowedMCPToolCodes(input []string) []string {
	if len(input) == 0 {
		return nil
	}
	ret := make([]string, 0, len(input))
	for _, item := range input {
		item = strings.TrimSpace(item)
		serverCode, toolName := toolx.SplitMCPToolCode(item)
		if serverCode == "" || toolName == "" {
			continue
		}
		ret = append(ret, item)
	}
	return ret
}

func containsToolCode(items []string, target string) bool {
	target = strings.TrimSpace(target)
	if target == "" {
		return false
	}
	for _, item := range items {
		if strings.TrimSpace(item) == target {
			return true
		}
	}
	return false
}

func filterCandidatesByQuery(candidates []toolSearchCandidate, query string) []toolSearchCandidate {
	query = strings.TrimSpace(strings.ToLower(query))
	if query == "" {
		return candidates
	}
	ret := make([]toolSearchCandidate, 0, len(candidates))
	for _, item := range candidates {
		searchText := strings.ToLower(strings.Join([]string{
			item.ToolCode,
			item.ServerCode,
			item.ToolName,
			item.Title,
			item.Description,
		}, "\n"))
		if strings.Contains(searchText, query) {
			ret = append(ret, item)
		}
	}
	return ret
}

func cloneArguments(input map[string]any) map[string]any {
	if len(input) == 0 {
		return map[string]any{}
	}
	ret := make(map[string]any, len(input))
	for key, value := range input {
		ret[key] = value
	}
	return ret
}

func buildToolCallResultSummary(result *mcps.ToolCallResult) string {
	if result == nil {
		return ""
	}
	lines := make([]string, 0, len(result.Content)+2)
	if result.IsError {
		lines = append(lines, "tool returned an error")
	}
	if result.StructuredContent != nil {
		if data, err := json.Marshal(result.StructuredContent); err == nil {
			lines = append(lines, string(data))
		}
	}
	for _, item := range result.Content {
		switch item.Type {
		case "text":
			if text := strings.TrimSpace(item.Text); text != "" {
				lines = append(lines, text)
			}
		default:
			if item.Data == nil {
				continue
			}
			if data, err := json.Marshal(item.Data); err == nil {
				lines = append(lines, string(data))
			}
		}
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}
