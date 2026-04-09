package toolx

import (
	"encoding/json"
	"strings"

	"cs-agent/internal/pkg/dto/request"
	"cs-agent/internal/pkg/errorsx"
)

func BuildMCPToolCode(serverCode, toolName string) string {
	serverCode = strings.TrimSpace(serverCode)
	toolName = strings.TrimSpace(toolName)
	if serverCode == "" || toolName == "" {
		return ""
	}
	return serverCode + "/" + toolName
}

func SplitMCPToolCode(toolCode string) (string, string) {
	toolCode = strings.TrimSpace(toolCode)
	if toolCode == "" {
		return "", ""
	}
	idx := strings.Index(toolCode, "/")
	if idx <= 0 || idx >= len(toolCode)-1 {
		return "", ""
	}
	return strings.TrimSpace(toolCode[:idx]), strings.TrimSpace(toolCode[idx+1:])
}

func NormalizeMCPToolRequest(item request.AIAgentMCPToolRequest) (request.AIAgentMCPToolRequest, error) {
	toolCode := strings.TrimSpace(item.ToolCode)
	serverCode := strings.TrimSpace(item.ServerCode)
	toolName := strings.TrimSpace(item.ToolName)
	if toolCode != "" {
		parsedServerCode, parsedToolName := SplitMCPToolCode(toolCode)
		if parsedServerCode == "" || parsedToolName == "" {
			return request.AIAgentMCPToolRequest{}, errorsx.InvalidParam("Direct Tool 的 toolCode 格式不合法")
		}
		if serverCode != "" && !strings.EqualFold(serverCode, parsedServerCode) {
			return request.AIAgentMCPToolRequest{}, errorsx.InvalidParam("Direct Tool 的 toolCode 与 serverCode 不一致")
		}
		if toolName != "" && !strings.EqualFold(toolName, parsedToolName) {
			return request.AIAgentMCPToolRequest{}, errorsx.InvalidParam("Direct Tool 的 toolCode 与 toolName 不一致")
		}
		serverCode = parsedServerCode
		toolName = parsedToolName
	} else {
		toolCode = BuildMCPToolCode(serverCode, toolName)
	}
	if toolCode == "" || serverCode == "" || toolName == "" {
		return request.AIAgentMCPToolRequest{}, errorsx.InvalidParam("Direct Tool 的 toolCode、serverCode 和 toolName 不能为空")
	}
	ret := request.AIAgentMCPToolRequest{
		ToolCode:    toolCode,
		ServerCode:  serverCode,
		ToolName:    toolName,
		Title:       strings.TrimSpace(item.Title),
		Description: strings.TrimSpace(item.Description),
	}
	if len(item.Arguments) > 0 {
		ret.Arguments = make(map[string]string, len(item.Arguments))
		for key, value := range item.Arguments {
			key = strings.TrimSpace(key)
			value = strings.TrimSpace(value)
			if key == "" || value == "" {
				continue
			}
			ret.Arguments[key] = value
		}
	}
	return ret, nil
}

func ParseAgentMCPToolsJSON(raw string) ([]request.AIAgentMCPToolRequest, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	var ret []request.AIAgentMCPToolRequest
	if err := json.Unmarshal([]byte(raw), &ret); err != nil {
		return nil, err
	}
	return ret, nil
}
