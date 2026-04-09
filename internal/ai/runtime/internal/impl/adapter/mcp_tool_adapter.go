package adapter

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"regexp"
	"strings"

	"cs-agent/internal/ai/mcps"

	einojsonschema "github.com/eino-contrib/jsonschema"

	einotool "github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

var toolNameSanitizer = regexp.MustCompile(`[^a-zA-Z0-9_]`)

type MCPToolDefinition struct {
	ToolCode    string
	ServerCode  string
	ToolName    string
	ModelName   string
	Title       string
	Description string
	FixedArgs   map[string]string
}

type MCPTool struct {
	definition MCPToolDefinition
	info       *schema.ToolInfo
}

func NewMCPTool(definition MCPToolDefinition, metadata *mcps.ToolInfo) *MCPTool {
	return &MCPTool{
		definition: definition,
		info:       buildToolInfo(definition, metadata),
	}
}

var _ einotool.InvokableTool = (*MCPTool)(nil)

func (t *MCPTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	if t == nil || t.info == nil {
		return nil, nil
	}
	return t.info, nil
}

func (t *MCPTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...einotool.Option) (string, error) {
	if t == nil {
		return "", fmt.Errorf("mcp tool is nil")
	}
	arguments, err := parseArguments(argumentsInJSON)
	if err != nil {
		return "", err
	}
	arguments = mergeFixedArguments(arguments, t.definition.FixedArgs)
	result, err := mcps.Runtime.CallTool(ctx, t.definition.ServerCode, t.definition.ToolName, arguments)
	if err != nil {
		return "", err
	}
	return buildToolResultSummary(result), nil
}

func buildToolInfo(definition MCPToolDefinition, metadata *mcps.ToolInfo) *schema.ToolInfo {
	desc := strings.TrimSpace(definition.Description)
	if desc == "" && metadata != nil {
		desc = strings.TrimSpace(metadata.Description)
	}
	title := strings.TrimSpace(definition.Title)
	if title == "" && metadata != nil {
		title = strings.TrimSpace(metadata.Title)
	}
	if title != "" && desc != "" {
		desc = title + "\n\n" + desc
	} else if title != "" {
		desc = title
	}
	if desc == "" {
		desc = "Call MCP tool " + strings.TrimSpace(definition.ToolCode)
	}
	info := &schema.ToolInfo{
		Name: BuildModelToolName(definition),
		Desc: desc,
		Extra: map[string]any{
			"toolCode":   definition.ToolCode,
			"serverCode": definition.ServerCode,
			"toolName":   definition.ToolName,
		},
	}
	if js := buildParamsSchema(metadata); js != nil {
		info.ParamsOneOf = schema.NewParamsOneOfByJSONSchema(js)
	}
	return info
}

func buildParamsSchema(metadata *mcps.ToolInfo) *einojsonschema.Schema {
	if metadata == nil || metadata.InputSchema == nil {
		return genericObjectSchema()
	}
	raw, err := json.Marshal(metadata.InputSchema)
	if err != nil || len(raw) == 0 {
		return genericObjectSchema()
	}
	js := &einojsonschema.Schema{}
	if err := json.Unmarshal(raw, js); err != nil {
		return genericObjectSchema()
	}
	return js
}

func genericObjectSchema() *einojsonschema.Schema {
	return &einojsonschema.Schema{
		Version:              einojsonschema.Version,
		Type:                 "object",
		AdditionalProperties: &einojsonschema.Schema{},
	}
}

func parseArguments(argumentsInJSON string) (map[string]any, error) {
	argumentsInJSON = strings.TrimSpace(argumentsInJSON)
	if argumentsInJSON == "" {
		return map[string]any{}, nil
	}
	args := make(map[string]any)
	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return nil, fmt.Errorf("invalid tool arguments: %w", err)
	}
	return args, nil
}

func mergeFixedArguments(arguments map[string]any, fixedArgs map[string]string) map[string]any {
	if len(arguments) == 0 && len(fixedArgs) == 0 {
		return map[string]any{}
	}
	ret := make(map[string]any, len(arguments)+len(fixedArgs))
	for key, value := range arguments {
		ret[key] = value
	}
	for key, value := range fixedArgs {
		ret[key] = strings.TrimSpace(value)
	}
	return ret
}

func buildToolResultSummary(result *mcps.ToolCallResult) string {
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

func BuildModelToolName(definition MCPToolDefinition) string {
	if strings.TrimSpace(definition.ModelName) != "" {
		return strings.TrimSpace(definition.ModelName)
	}
	base := "mcp_" + strings.TrimSpace(definition.ServerCode) + "_" + strings.TrimSpace(definition.ToolName)
	base = toolNameSanitizer.ReplaceAllString(base, "_")
	base = strings.Trim(base, "_")
	if base == "" {
		base = "mcp_tool"
	}
	checksum := crc32.ChecksumIEEE([]byte(definition.ToolCode))
	return fmt.Sprintf("%s_%08x", base, checksum)
}
