package response

import (
	"cs-agent/internal/pkg/enums"
	"time"
)

type KnowledgeBaseResponse struct {
	ID                    int64        `json:"id"`
	Name                  string       `json:"name"`
	Description           string       `json:"description"`
	Status                enums.Status `json:"status"`
	StatusName            string       `json:"statusName"`
	DefaultTopK           int          `json:"defaultTopK"`
	DefaultScoreThreshold float64      `json:"defaultScoreThreshold"`
	DefaultRerankLimit    int          `json:"defaultRerankLimit"`
	ChunkProvider         string       `json:"chunkProvider"`
	ChunkTargetTokens     int          `json:"chunkTargetTokens"`
	ChunkMaxTokens        int          `json:"chunkMaxTokens"`
	ChunkOverlapTokens    int          `json:"chunkOverlapTokens"`
	AnswerMode            int          `json:"answerMode"`
	AnswerModeName        string       `json:"answerModeName"`
	FallbackMode          int          `json:"fallbackMode"`
	FallbackModeName      string       `json:"fallbackModeName"`
	DocumentCount         int64        `json:"documentCount"`
	Remark                string       `json:"remark"`
	CreatedAt             time.Time    `json:"createdAt"`
	UpdatedAt             time.Time    `json:"updatedAt"`
	CreateUserName        string       `json:"createUserName"`
	UpdateUserName        string       `json:"updateUserName"`
}

type KnowledgeDocumentResponse struct {
	ID                int64                              `json:"id"`
	KnowledgeBaseID   int64                              `json:"knowledgeBaseId"`
	KnowledgeBaseName string                             `json:"knowledgeBaseName,omitempty"`
	Title             string                             `json:"title"`
	ContentType       enums.KnowledgeDocumentContentType `json:"contentType"`
	Content           string                             `json:"content"`
	Status            enums.Status                       `json:"status"`
	StatusName        string                             `json:"statusName"`
	ContentHash       string                             `json:"contentHash"`
	CreatedAt         time.Time                          `json:"createdAt"`
	UpdatedAt         time.Time                          `json:"updatedAt"`
	CreateUserName    string                             `json:"createUserName"`
	UpdateUserName    string                             `json:"updateUserName"`
}

type KnowledgeChunkResponse struct {
	ID              int64        `json:"id"`
	KnowledgeBaseID int64        `json:"knowledgeBaseId"`
	DocumentID      int64        `json:"documentId"`
	DocumentTitle   string       `json:"documentTitle,omitempty"`
	ChunkNo         int          `json:"chunkNo"`
	Title           string       `json:"title"`
	Content         string       `json:"content"`
	ContentHash     string       `json:"contentHash"`
	CharCount       int          `json:"charCount"`
	TokenCount      int          `json:"tokenCount"`
	Status          enums.Status `json:"status"`
	StatusName      string       `json:"statusName"`
	VectorID        string       `json:"vectorId"`
	CreatedAt       time.Time    `json:"createdAt"`
	UpdatedAt       time.Time    `json:"updatedAt"`
}

type KnowledgeSearchResult struct {
	KnowledgeBaseID int64   `json:"knowledgeBaseId"`
	ChunkID         int64   `json:"chunkId"`
	DocumentID      int64   `json:"documentId"`
	DocumentTitle   string  `json:"documentTitle"`
	ChunkNo         int     `json:"chunkNo"`
	Title           string  `json:"title"`
	SectionPath     string  `json:"sectionPath"`
	Content         string  `json:"content"`
	Score           float64 `json:"score"`
	RerankScore     float64 `json:"rerankScore"`
}

type KnowledgeSearchResponse struct {
	Question  string                  `json:"question"`
	Results   []KnowledgeSearchResult `json:"results"`
	HitCount  int                     `json:"hitCount"`
	LatencyMs int64                   `json:"latencyMs"`
}

type KnowledgeAnswerResponse struct {
	Question         string                  `json:"question"`
	RewriteQuestion  string                  `json:"rewriteQuestion,omitempty"`
	Answer           string                  `json:"answer"`
	AnswerStatus     int                     `json:"answerStatus"`
	AnswerStatusName string                  `json:"answerStatusName"`
	Citations        []KnowledgeCitation     `json:"citations"`
	Hits             []KnowledgeSearchResult `json:"hits"`
	HitCount         int                     `json:"hitCount"`
	TopScore         float64                 `json:"topScore"`
	LatencyMs        int64                   `json:"latencyMs"`
	RetrieveMs       int64                   `json:"retrieveMs"`
	GenerateMs       int64                   `json:"generateMs"`
	PromptTokens     int                     `json:"promptTokens"`
	CompletionTokens int                     `json:"completionTokens"`
	ModelName        string                  `json:"modelName"`
	RetrieveLogID    int64                   `json:"retrieveLogId"`
}

type KnowledgeCitation struct {
	DocumentID    int64   `json:"documentId"`
	DocumentTitle string  `json:"documentTitle"`
	ChunkNo       int     `json:"chunkNo"`
	Title         string  `json:"title"`
	SectionPath   string  `json:"sectionPath"`
	Snippet       string  `json:"snippet"`
	Score         float64 `json:"score"`
}

type KnowledgeCompareProviderResult struct {
	Provider           string                  `json:"provider"`
	HitCount           int                     `json:"hitCount"`
	BuildMs            int64                   `json:"buildMs"`
	RetrieveMs         int64                   `json:"retrieveMs"`
	Top1Matched        bool                    `json:"top1Matched"`
	Top3Matched        bool                    `json:"top3Matched"`
	MatchedDocumentIDs []int64                 `json:"matchedDocumentIds"`
	Results            []KnowledgeSearchResult `json:"results"`
}

type KnowledgeCompareResponse struct {
	Question  string                           `json:"question"`
	Providers []KnowledgeCompareProviderResult `json:"providers"`
	LatencyMs int64                            `json:"latencyMs"`
}

type KnowledgeRetrieveLogResponse struct {
	ID                 int64     `json:"id"`
	KnowledgeBaseID    int64     `json:"knowledgeBaseId"`
	KnowledgeBaseName  string    `json:"knowledgeBaseName,omitempty"`
	Channel            string    `json:"channel"`
	ChannelName        string    `json:"channelName"`
	Scene              string    `json:"scene"`
	SceneName          string    `json:"sceneName"`
	SessionID          string    `json:"sessionId"`
	ConversationID     int64     `json:"conversationId"`
	RequestID          string    `json:"requestId"`
	Question           string    `json:"question"`
	RewriteQuestion    string    `json:"rewriteQuestion"`
	Answer             string    `json:"answer"`
	AnswerStatus       int       `json:"answerStatus"`
	AnswerStatusName   string    `json:"answerStatusName"`
	HitCount           int       `json:"hitCount"`
	TopScore           float64   `json:"topScore"`
	ChunkProvider      string    `json:"chunkProvider"`
	ChunkTargetTokens  int       `json:"chunkTargetTokens"`
	ChunkMaxTokens     int       `json:"chunkMaxTokens"`
	ChunkOverlapTokens int       `json:"chunkOverlapTokens"`
	RerankEnabled      bool      `json:"rerankEnabled"`
	RerankLimit        int       `json:"rerankLimit"`
	CitationCount      int       `json:"citationCount"`
	UsedChunkCount     int       `json:"usedChunkCount"`
	LatencyMs          int64     `json:"latencyMs"`
	RetrieveMs         int64     `json:"retrieveMs"`
	GenerateMs         int64     `json:"generateMs"`
	PromptTokens       int       `json:"promptTokens"`
	CompletionTokens   int       `json:"completionTokens"`
	ModelName          string    `json:"modelName"`
	TraceData          string    `json:"traceData"`
	CreatedAt          time.Time `json:"createdAt"`
}

type KnowledgeRetrieveHitResponse struct {
	ID              int64     `json:"id"`
	RetrieveLogID   int64     `json:"retrieveLogId"`
	KnowledgeBaseID int64     `json:"knowledgeBaseId"`
	ChunkID         int64     `json:"chunkId"`
	DocumentID      int64     `json:"documentId"`
	DocumentTitle   string    `json:"documentTitle"`
	ChunkNo         int       `json:"chunkNo"`
	Title           string    `json:"title"`
	SectionPath     string    `json:"sectionPath"`
	ChunkType       string    `json:"chunkType"`
	ChunkTypeName   string    `json:"chunkTypeName"`
	Provider        string    `json:"provider"`
	RankNo          int       `json:"rankNo"`
	Score           float64   `json:"score"`
	RerankScore     float64   `json:"rerankScore"`
	UsedInAnswer    bool      `json:"usedInAnswer"`
	IsCitation      bool      `json:"isCitation"`
	Snippet         string    `json:"snippet"`
	CreatedAt       time.Time `json:"createdAt"`
}

type KnowledgeRetrieveLogDetailResponse struct {
	Log  KnowledgeRetrieveLogResponse   `json:"log"`
	Hits []KnowledgeRetrieveHitResponse `json:"hits"`
}

type KnowledgeFeedbackResponse struct {
	ID               int64     `json:"id"`
	RetrieveLogID    int64     `json:"retrieveLogId"`
	FeedbackType     int       `json:"feedbackType"`
	FeedbackTypeName string    `json:"feedbackTypeName"`
	FeedbackReason   string    `json:"feedbackReason"`
	UserID           int64     `json:"userId"`
	AgentID          int64     `json:"agentId"`
	Remark           string    `json:"remark"`
	CreatedAt        time.Time `json:"createdAt"`
}
