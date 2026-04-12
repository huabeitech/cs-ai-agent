package callbacks

type ToolTraceItem struct {
	ToolCode      string         `json:"toolCode"`
	ServerCode    string         `json:"serverCode"`
	ToolName      string         `json:"toolName"`
	Arguments     map[string]any `json:"arguments,omitempty"`
	ResultPreview string         `json:"resultPreview,omitempty"`
	LatencyMs     int64          `json:"latencyMs,omitempty"`
	Status        string         `json:"status,omitempty"`
	ErrorMessage  string         `json:"errorMessage,omitempty"`
}

type ToolSearchTraceItem struct {
	Action             string   `json:"action,omitempty"`
	Query              string   `json:"query,omitempty"`
	TargetToolCode     string   `json:"targetToolCode,omitempty"`
	TargetServerCode   string   `json:"targetServerCode,omitempty"`
	TargetToolName     string   `json:"targetToolName,omitempty"`
	CandidateToolCodes []string `json:"candidateToolCodes,omitempty"`
	Status             string   `json:"status,omitempty"`
	ErrorMessage       string   `json:"errorMessage,omitempty"`
}

type GraphToolTraceItem struct {
	ToolCode      string         `json:"toolCode"`
	ToolName      string         `json:"toolName"`
	Arguments     map[string]any `json:"arguments,omitempty"`
	ResultPreview string         `json:"resultPreview,omitempty"`
	LatencyMs     int64          `json:"latencyMs,omitempty"`
	Status        string         `json:"status,omitempty"`
	ErrorMessage  string         `json:"errorMessage,omitempty"`
}

type RetrieverTraceItem struct {
	Query           string  `json:"query,omitempty"`
	KnowledgeBaseID int64   `json:"knowledgeBaseId,omitempty"`
	DocumentID      int64   `json:"documentId,omitempty"`
	DocumentTitle   string  `json:"documentTitle,omitempty"`
	Score           float64 `json:"score,omitempty"`
	LatencyMs       int64   `json:"latencyMs,omitempty"`
}

type InstructionTraceSummary struct {
	SectionTitles     []string
	HasProjectRule    bool
	HasGovernanceRule bool
	HasAgentRule      bool
	HasSkillRule      bool
	HasToolRule       bool
}

type RuntimeTraceData struct {
	Version   string         `json:"version"`
	Status    string         `json:"status"`
	RunID     string         `json:"runId,omitempty"`
	Skill     SkillTraceData `json:"skill,omitempty"`
	Interrupt struct {
		CheckPointID string                  `json:"checkPointId,omitempty"`
		Items        []InterruptTraceContext `json:"items,omitempty"`
	} `json:"interrupt"`
	Model struct {
		Provider string `json:"provider,omitempty"`
		Name     string `json:"name,omitempty"`
	} `json:"model"`
	Instruction struct {
		SectionTitles     []string `json:"sectionTitles,omitempty"`
		HasProjectRule    bool     `json:"hasProjectRule,omitempty"`
		HasGovernanceRule bool     `json:"hasGovernanceRule,omitempty"`
		HasAgentRule      bool     `json:"hasAgentRule,omitempty"`
		HasSkillRule      bool     `json:"hasSkillRule,omitempty"`
		HasToolRule       bool     `json:"hasToolRule,omitempty"`
	} `json:"instruction"`
	Input struct {
		HistoryMessageCount       int      `json:"historyMessageCount,omitempty"`
		KnowledgeBaseIDs          []int64  `json:"knowledgeBaseIds,omitempty"`
		ToolCodes                 []string `json:"toolCodes,omitempty"`
		StaticToolCodes           []string `json:"staticToolCodes,omitempty"`
		DynamicToolCodes          []string `json:"dynamicToolCodes,omitempty"`
		ToolSearchEnabled         bool     `json:"toolSearchEnabled,omitempty"`
		CurrentUserMessagePreview string   `json:"currentUserMessagePreview,omitempty"`
	} `json:"input"`
	Retriever struct {
		Count int                  `json:"count,omitempty"`
		Items []RetrieverTraceItem `json:"items,omitempty"`
	} `json:"retriever"`
	Tools struct {
		Count int             `json:"count,omitempty"`
		Items []ToolTraceItem `json:"items,omitempty"`
	} `json:"tools"`
	ToolSearch struct {
		Count int                   `json:"count,omitempty"`
		Items []ToolSearchTraceItem `json:"items,omitempty"`
	} `json:"toolSearch"`
	GraphTools struct {
		Count int                  `json:"count,omitempty"`
		Items []GraphToolTraceItem `json:"items,omitempty"`
	} `json:"graphTools"`
	Output struct {
		ReplyText    string `json:"replyText,omitempty"`
		FinishReason string `json:"finishReason,omitempty"`
	} `json:"output"`
	Error struct {
		Message string `json:"message,omitempty"`
		Stage   string `json:"stage,omitempty"`
	} `json:"error"`
}

type SkillTraceData struct {
	Code               string   `json:"code,omitempty"`
	Name               string   `json:"name,omitempty"`
	RouteReason        string   `json:"routeReason,omitempty"`
	RouteTrace         string   `json:"routeTrace,omitempty"`
	AllowedToolCodes   []string `json:"allowedToolCodes,omitempty"`
	MiddlewareEnabled  bool     `json:"middlewareEnabled,omitempty"`
	MiddlewareToolName string   `json:"middlewareToolName,omitempty"`
}

type InterruptTraceContext struct {
	Type        string `json:"type,omitempty"`
	ID          string `json:"id"`
	InfoPreview string `json:"infoPreview,omitempty"`
}
