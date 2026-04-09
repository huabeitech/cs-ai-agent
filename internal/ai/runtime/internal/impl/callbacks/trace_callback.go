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

type RetrieverTraceItem struct {
	Query           string  `json:"query,omitempty"`
	KnowledgeBaseID int64   `json:"knowledgeBaseId,omitempty"`
	DocumentID      int64   `json:"documentId,omitempty"`
	DocumentTitle   string  `json:"documentTitle,omitempty"`
	Score           float64 `json:"score,omitempty"`
	LatencyMs       int64   `json:"latencyMs,omitempty"`
}

type RuntimeTraceData struct {
	Version   string `json:"version"`
	Status    string `json:"status"`
	RunID     string `json:"runId,omitempty"`
	Interrupt struct {
		CheckPointID string                  `json:"checkPointId,omitempty"`
		Items        []InterruptTraceContext `json:"items,omitempty"`
	} `json:"interrupt"`
	Model struct {
		Provider string `json:"provider,omitempty"`
		Name     string `json:"name,omitempty"`
	} `json:"model"`
	Input struct {
		HistoryMessageCount       int      `json:"historyMessageCount,omitempty"`
		KnowledgeBaseIDs          []int64  `json:"knowledgeBaseIds,omitempty"`
		ToolCodes                 []string `json:"toolCodes,omitempty"`
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
	Output struct {
		ReplyText    string `json:"replyText,omitempty"`
		FinishReason string `json:"finishReason,omitempty"`
	} `json:"output"`
	Error struct {
		Message string `json:"message,omitempty"`
		Stage   string `json:"stage,omitempty"`
	} `json:"error"`
}

type InterruptTraceContext struct {
	Type        string `json:"type,omitempty"`
	ID          string `json:"id"`
	InfoPreview string `json:"infoPreview,omitempty"`
}
