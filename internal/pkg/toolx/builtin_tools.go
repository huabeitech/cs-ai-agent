package toolx

import "strings"

type ToolSpec struct {
	Code          string
	ServerCode    string
	Name          string
	Title         string
	Description   string
	SourceType    string
	AutoInjected  bool
	DirectAccess  bool
	RuntimeStatic bool
	Aliases       []string
	Appendix      string
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
		DirectAccess: true,
		Appendix: strings.TrimSpace(`
当你需要使用长尾 MCP 能力时，优先使用 tool_search 工具，并遵守以下规则：
1. 先调用 tool_search 搜索需要的动态工具，再继续使用已选中的真实工具。
2. 不要假设所有长尾工具一开始就可见；只有被 tool_search 选中的工具，后续模型调用才会暴露出来。
3. 如果当前已有固定内置工具可以完成任务，优先使用固定工具，不要滥用 tool_search。
`),
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
		Code:          "graph/triage_service_request",
		ServerCode:    "graph",
		Name:          "triage_service_request",
		Title:         "升级分流判断",
		Description:   "Graph Tool。用于综合分析当前对话，判断应继续解答、整理工单草稿还是转人工，并在需要建单时一并整理工单草稿。",
		SourceType:    "graph",
		RuntimeStatic: true,
		Appendix: strings.TrimSpace(`
当你需要判断“继续解答 / 建单 / 转人工”这类复杂升级路径时，优先先调用 triage_service_request 这个 Graph Tool，并遵守以下规则：
1. 该工具会综合当前对话输出 recommendedAction，并在需要建单时附带 ticketDraft。
2. 如果 recommendedAction=continue_answering，则优先继续澄清或解答，不要直接升级。
3. 如果 recommendedAction=prepare_ticket，则优先使用 ticketDraft 或继续补充缺失字段，再调用 create_ticket_with_confirmation。
4. 如果 recommendedAction=handoff_to_human，则确认理由充分后再调用 handoff_to_human。
5. 当升级路径不明确时，优先使用该工具，而不是直接凭主 prompt 做复杂分流判断。
`),
	}
	GraphAnalyzeConversation = ToolSpec{
		Code:          "graph/analyze_conversation",
		ServerCode:    "graph",
		Name:          "analyze_conversation",
		Title:         "分析对话风险与摘要",
		Description:   "Graph Tool。用于整理当前对话摘要、识别风险信号，并给出继续解答、建单或转人工的建议。",
		SourceType:    "graph",
		RuntimeStatic: true,
		Appendix: strings.TrimSpace(`
当对话可能涉及投诉升级、退款赔偿、明显负面情绪、是否要建单、是否要转人工等复杂判断时，优先调用 analyze_conversation 这个 Graph Tool，并遵守以下规则：
1. 该工具用于输出结构化摘要、风险信号和下一步建议，不代表实际已经建单或转人工。
2. 如果工具建议为 handoff_to_human，应先确认是否满足转人工条件，再考虑调用 handoff_to_human。
3. 如果工具建议为 prepare_ticket，应优先调用 prepare_ticket_draft 或继续补充信息，而不是直接建单。
4. 如果工具建议为 continue_answering，优先继续澄清和解答，不要过早升级动作。
`),
	}
	GraphPrepareTicketDraft = ToolSpec{
		Code:          "graph/prepare_ticket_draft",
		ServerCode:    "graph",
		Name:          "prepare_ticket_draft",
		Title:         "整理工单草稿",
		Description:   "Graph Tool。用于根据当前会话和已收集信息整理工单草稿，输出建议标题、描述、缺失字段和追问建议。",
		SourceType:    "graph",
		RuntimeStatic: true,
		Appendix: strings.TrimSpace(`
当用户已经表达了建单、投诉、报障、售后处理等诉求，但工单标题、描述或问题整理还比较散乱时，优先调用 prepare_ticket_draft 这个 Graph Tool，并遵守以下规则：
1. 该工具用于整理工单草稿，会返回建议标题、建议描述、缺失字段和追问建议。
2. 如果工具返回 ready=false，优先根据 missingFields 和 followUpQuestions 继续追问，不要直接创建工单。
3. 如果工具返回 ready=true，再结合结果考虑调用 create_ticket_with_confirmation。
4. 该工具用于“整理草稿”，不代表已经创建工单。
`),
	}
	GraphCreateTicketConfirm = ToolSpec{
		Code:          "graph/create_ticket_with_confirmation",
		ServerCode:    "graph",
		Name:          "create_ticket_with_confirmation",
		Title:         "创建工单确认流程",
		Description:   "Graph Tool。用于封装建单参数整理、用户确认、真正建单和结果返回的确定性流程。",
		SourceType:    "graph",
		DirectAccess:  true,
		RuntimeStatic: true,
		Aliases:       []string{"builtin/create_ticket_with_confirmation"},
		Appendix: strings.TrimSpace(`
你可以在确认信息充分后调用 create_ticket_with_confirmation 这个 Graph Tool 来创建工单，但必须遵守以下规则：
1. 只有在用户明确表达希望提交工单、投诉、报障、售后处理等诉求时，才考虑调用该工具。
2. 调用前你必须已经整理出清晰的工单标题和问题描述；如果信息还比较散乱，优先先调用 prepare_ticket_draft 或继续追问，不要过早调用。
3. 一旦准备创建工单，必须调用 create_ticket_with_confirmation 工具，禁止直接口头宣称“已经创建工单”。
4. 该 Graph Tool 会先向用户发起确认。用户确认后才会真正创建工单；用户取消则结束本次建单流程。
5. 如果用户只是咨询、抱怨或泛泛表达不满，但没有明确要求建单，优先继续澄清，不要主动创建工单。
`),
	}
	GraphHandoffConversation = ToolSpec{
		Code:          "graph/handoff_to_human",
		ServerCode:    "graph",
		Name:          "handoff_to_human",
		Title:         "转人工确认流程",
		Description:   "Graph Tool。用于封装转人工原因整理、用户确认、真正转人工和结果返回的确定性流程。",
		SourceType:    "graph",
		DirectAccess:  true,
		RuntimeStatic: true,
		Appendix: strings.TrimSpace(`
你可以在确认需要人工介入后调用 handoff_to_human 这个 Graph Tool 来转人工，但必须遵守以下规则：
1. 只有在用户明确要求人工客服，或你已经判断该问题必须由人工继续处理时，才调用该工具。
2. 调用前先尽量整理清楚转人工原因；如果理由含糊，先追问或澄清，不要直接转人工。
3. 一旦决定转人工，必须调用 handoff_to_human 工具，禁止只在回复里口头说“我帮你转人工了”。
4. 该 Graph Tool 会先向用户发起确认。用户确认后才会真正转人工；用户取消则结束本次转人工流程。
5. 如果问题仍可由当前对话继续解决，优先继续解答，不要过早转人工。
`),
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
	ret := make([]ToolSpec, 0, len(RegisteredToolSpecs))
	for _, spec := range RegisteredToolSpecs {
		if !spec.DirectAccess {
			continue
		}
		ret = append(ret, spec)
	}
	return ret
}

func ListRuntimeStaticToolSpecs() []ToolSpec {
	ret := make([]ToolSpec, 0, len(RegisteredToolSpecs))
	for _, spec := range RegisteredToolSpecs {
		if !spec.RuntimeStatic {
			continue
		}
		ret = append(ret, spec)
	}
	return ret
}

func IsAgentDirectToolCode(toolCode string) bool {
	toolCode = NormalizeToolCodeAlias(strings.TrimSpace(toolCode))
	spec, ok := GetRegisteredToolSpec(toolCode)
	return ok && spec.DirectAccess
}

func NormalizeToolCodeAlias(toolCode string) string {
	toolCode = strings.TrimSpace(toolCode)
	if canonical, ok := toolAliasToCanonical[toolCode]; ok {
		return canonical
	}
	return toolCode
}

func BuildToolAppendices(hasDynamicMCPTools bool, toolCodes map[string]string) []string {
	return BuildToolAppendicesForCodes(hasDynamicMCPTools, toolCodesFromMap(toolCodes))
}

func BuildToolAppendicesForCodes(hasDynamicMCPTools bool, toolCodes []string) []string {
	ret := make([]string, 0, len(toolCodes)+1)
	normalizedToolCodes := NormalizeToolCodes(toolCodes)
	if hasDynamicMCPTools && strings.TrimSpace(BuiltinToolSearch.Appendix) != "" {
		ret = append(ret, BuiltinToolSearch.Appendix)
	}
	for _, spec := range RegisteredToolSpecs {
		if strings.TrimSpace(spec.Appendix) == "" {
			continue
		}
		if spec.Code == BuiltinToolSearch.Code {
			continue
		}
		if containsNormalizedToolCode(normalizedToolCodes, spec.Code) {
			ret = append(ret, spec.Appendix)
		}
	}
	return ret
}

func BuildToolMetadata(toolCode string) (serverCode, toolName, sourceType string, ok bool) {
	spec, ok := GetRegisteredToolSpec(toolCode)
	if !ok {
		return "", "", ResolveToolSourceType(toolCode), false
	}
	return spec.ServerCode, spec.Name, spec.SourceType, true
}

func IsAlwaysAllowedToolCode(toolCode string) bool {
	return NormalizeToolCodeAlias(strings.TrimSpace(toolCode)) == GraphHandoffConversation.Code
}

func IsImpliedAllowedToolCode(toolCode string, allowedToolCodes map[string]struct{}) bool {
	toolCode = NormalizeToolCodeAlias(strings.TrimSpace(toolCode))
	if toolCode == "" || len(allowedToolCodes) == 0 {
		return false
	}
	switch toolCode {
	case GraphTriageServiceRequest.Code, GraphAnalyzeConversation.Code:
		if _, ok := allowedToolCodes[GraphCreateTicketConfirm.Code]; ok {
			return true
		}
		if _, ok := allowedToolCodes[GraphHandoffConversation.Code]; ok {
			return true
		}
	case GraphPrepareTicketDraft.Code:
		_, ok := allowedToolCodes[GraphCreateTicketConfirm.Code]
		return ok
	}
	return false
}

func hasToolCode(toolCodes map[string]string, target string) bool {
	return containsNormalizedToolCode(toolCodesFromMap(toolCodes), target)
}

func containsNormalizedToolCode(toolCodes []string, target string) bool {
	target = strings.TrimSpace(target)
	if target == "" || len(toolCodes) == 0 {
		return false
	}
	for _, item := range NormalizeToolCodes(toolCodes) {
		if item == target {
			return true
		}
	}
	return false
}

func toolCodesFromMap(toolCodes map[string]string) []string {
	if len(toolCodes) == 0 {
		return nil
	}
	ret := make([]string, 0, len(toolCodes))
	for _, item := range toolCodes {
		ret = append(ret, item)
	}
	return ret
}

func NormalizeToolCodes(items []string) []string {
	if len(items) == 0 {
		return nil
	}
	ret := make([]string, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		item = NormalizeToolCodeAlias(strings.TrimSpace(item))
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		ret = append(ret, item)
	}
	return ret
}

func IntersectToolCodes(left []string, right []string) []string {
	left = NormalizeToolCodes(left)
	right = NormalizeToolCodes(right)
	switch {
	case len(left) == 0:
		return right
	case len(right) == 0:
		return left
	}
	rightSet := make(map[string]struct{}, len(right))
	for _, item := range right {
		rightSet[item] = struct{}{}
	}
	ret := make([]string, 0, len(left))
	for _, item := range left {
		if _, ok := rightSet[item]; ok {
			ret = append(ret, item)
		}
	}
	return ret
}
