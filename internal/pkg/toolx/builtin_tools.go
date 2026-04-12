package toolx

import "strings"

type ToolSpec struct {
	Code         string
	ServerCode   string
	Name         string
	Title        string
	Description  string
	SourceType   string
	AutoInjected bool
	Aliases      []string
}

var (
	BuiltinToolSearch = ToolSpec{
		Code:         "builtin/tool_search",
		ServerCode:   "builtin",
		Name:         "tool_search",
		Title:        "搜索并调用动态工具",
		Description:  "用于搜索当前允许使用的 MCP 工具，并在确认目标 toolCode 后动态调用该工具。适合处理长尾工具，不应替代固定内置流程工具。",
		SourceType:   "builtin",
		AutoInjected: true,
	}
	BuiltinSkill = ToolSpec{
		Code:         "builtin/skill",
		ServerCode:   "builtin",
		Name:         "skill",
		Title:        "加载专项技能说明",
		Description:  "用于加载当前命中的专项技能说明文档。仅在本轮已命中 Skill 时可用，适合将专项处理规则按需注入上下文。",
		SourceType:   "builtin",
		AutoInjected: true,
	}
	GraphTriageServiceRequest = ToolSpec{
		Code:        "graph/triage_service_request",
		ServerCode:  "graph",
		Name:        "triage_service_request",
		Title:       "升级分流判断",
		Description: "Graph Tool。用于综合分析当前对话，判断应继续解答、整理工单草稿还是转人工，并在需要建单时一并整理工单草稿。",
		SourceType:  "graph",
	}
	GraphAnalyzeConversation = ToolSpec{
		Code:        "graph/analyze_conversation",
		ServerCode:  "graph",
		Name:        "analyze_conversation",
		Title:       "分析对话风险与摘要",
		Description: "Graph Tool。用于整理当前对话摘要、识别风险信号，并给出继续解答、建单或转人工的建议。",
		SourceType:  "graph",
	}
	GraphPrepareTicketDraft = ToolSpec{
		Code:        "graph/prepare_ticket_draft",
		ServerCode:  "graph",
		Name:        "prepare_ticket_draft",
		Title:       "整理工单草稿",
		Description: "Graph Tool。用于根据当前会话和已收集信息整理工单草稿，输出建议标题、描述、缺失字段和追问建议。",
		SourceType:  "graph",
	}
	GraphCreateTicketConfirm = ToolSpec{
		Code:        "graph/create_ticket_with_confirmation",
		ServerCode:  "graph",
		Name:        "create_ticket_with_confirmation",
		Title:       "创建工单确认流程",
		Description: "Graph Tool。用于封装建单参数整理、用户确认、真正建单和结果返回的确定性流程。",
		SourceType:  "graph",
		Aliases:     []string{"builtin/create_ticket_with_confirmation"},
	}
	GraphHandoffConversation = ToolSpec{
		Code:        "graph/handoff_to_human",
		ServerCode:  "graph",
		Name:        "handoff_to_human",
		Title:       "转人工确认流程",
		Description: "Graph Tool。用于封装转人工原因整理、用户确认、真正转人工和结果返回的确定性流程。",
		SourceType:  "graph",
	}
	RegisteredToolSpecs = []ToolSpec{
		BuiltinToolSearch,
		BuiltinSkill,
		GraphTriageServiceRequest,
		GraphAnalyzeConversation,
		GraphPrepareTicketDraft,
		GraphCreateTicketConfirm,
		GraphHandoffConversation,
	}
	AgentDirectToolSpecs = []ToolSpec{
		GraphCreateTicketConfirm,
		GraphHandoffConversation,
		BuiltinToolSearch,
	}
)

var (
	toolSpecByCode       = buildToolSpecByCode()
	toolSpecByName       = buildToolSpecByName()
	toolAliasToCanonical = buildToolAliasToCanonical()
)

func buildToolSpecByCode() map[string]ToolSpec {
	ret := make(map[string]ToolSpec, len(RegisteredToolSpecs))
	for _, spec := range RegisteredToolSpecs {
		if strings.TrimSpace(spec.Code) == "" {
			continue
		}
		ret[spec.Code] = spec
	}
	return ret
}

func buildToolAliasToCanonical() map[string]string {
	ret := make(map[string]string)
	for _, spec := range RegisteredToolSpecs {
		for _, alias := range spec.Aliases {
			alias = strings.TrimSpace(alias)
			if alias == "" {
				continue
			}
			ret[alias] = spec.Code
		}
	}
	return ret
}

func buildToolSpecByName() map[string]ToolSpec {
	ret := make(map[string]ToolSpec, len(RegisteredToolSpecs))
	for _, spec := range RegisteredToolSpecs {
		name := strings.TrimSpace(spec.Name)
		if name == "" {
			continue
		}
		ret[name] = spec
	}
	return ret
}

func ListRegisteredToolSpecs() []ToolSpec {
	return append([]ToolSpec(nil), RegisteredToolSpecs...)
}

func GetRegisteredToolSpec(toolCode string) (ToolSpec, bool) {
	toolCode = NormalizeToolCodeAlias(strings.TrimSpace(toolCode))
	spec, ok := toolSpecByCode[toolCode]
	return spec, ok
}

func GetRegisteredToolSpecByName(name string) (ToolSpec, bool) {
	name = strings.TrimSpace(name)
	spec, ok := toolSpecByName[name]
	return spec, ok
}

func GetRegisteredToolTitle(toolCode string) string {
	spec, ok := GetRegisteredToolSpec(toolCode)
	if !ok {
		return ""
	}
	return spec.Title
}

func GetRegisteredToolDescription(toolCode string) string {
	spec, ok := GetRegisteredToolSpec(toolCode)
	if !ok {
		return ""
	}
	return spec.Description
}

func GetRegisteredToolIdentity(toolCode string) (serverCode, toolName string, ok bool) {
	spec, ok := GetRegisteredToolSpec(toolCode)
	if !ok {
		return "", "", false
	}
	return spec.ServerCode, spec.Name, true
}

func ResolveToolSourceType(toolCode string) string {
	if spec, ok := GetRegisteredToolSpec(toolCode); ok {
		return spec.SourceType
	}
	toolCode = strings.TrimSpace(toolCode)
	switch {
	case strings.HasPrefix(toolCode, "graph/"):
		return "graph"
	case strings.HasPrefix(toolCode, "builtin/"):
		return "builtin"
	default:
		return "mcp"
	}
}

func IsAutoInjectedToolCode(toolCode string) bool {
	spec, ok := GetRegisteredToolSpec(toolCode)
	return ok && spec.AutoInjected
}

func ListAgentDirectToolSpecs() []ToolSpec {
	return append([]ToolSpec(nil), AgentDirectToolSpecs...)
}

func IsAgentDirectToolCode(toolCode string) bool {
	toolCode = NormalizeToolCodeAlias(strings.TrimSpace(toolCode))
	for _, spec := range AgentDirectToolSpecs {
		if spec.Code == toolCode {
			return true
		}
	}
	return false
}

func NormalizeToolCodeAlias(toolCode string) string {
	toolCode = strings.TrimSpace(toolCode)
	if canonical, ok := toolAliasToCanonical[toolCode]; ok {
		return canonical
	}
	return toolCode
}
