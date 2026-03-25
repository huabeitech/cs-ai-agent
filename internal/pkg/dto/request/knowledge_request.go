package request

import "cs-agent/internal/pkg/enums"

type CreateKnowledgeBaseRequest struct {
	Name                  string  `json:"name"`
	Description           string  `json:"description"`
	DefaultTopK           int     `json:"defaultTopK"`
	DefaultScoreThreshold float64 `json:"defaultScoreThreshold"`
	DefaultRerankLimit    int     `json:"defaultRerankLimit"`
	ChunkProvider         string  `json:"chunkProvider"`
	ChunkTargetTokens     int     `json:"chunkTargetTokens"`
	ChunkMaxTokens        int     `json:"chunkMaxTokens"`
	ChunkOverlapTokens    int     `json:"chunkOverlapTokens"`
	AnswerMode            int     `json:"answerMode"`
	FallbackMode          int     `json:"fallbackMode"`
	Remark                string  `json:"remark"`
}

type UpdateKnowledgeBaseRequest struct {
	ID int64 `json:"id"`
	CreateKnowledgeBaseRequest
}

type CreateKnowledgeDocumentRequest struct {
	KnowledgeBaseID int64                              `json:"knowledgeBaseId"`
	Title           string                             `json:"title"`
	ContentType     enums.KnowledgeDocumentContentType `json:"contentType"`
	Content         string                             `json:"content"`
}

type UpdateKnowledgeDocumentRequest struct {
	ID int64 `json:"id"`
	CreateKnowledgeDocumentRequest
}

type KnowledgeSearchRequest struct {
	KnowledgeBaseID int64   `json:"knowledgeBaseId"`
	Question        string  `json:"question"`
	TopK            int     `json:"topK"`
	ScoreThreshold  float64 `json:"scoreThreshold"`
	RerankLimit     int     `json:"rerankLimit"`
	Channel         string  `json:"channel"`
	Scene           string  `json:"scene"`
	SessionID       string  `json:"sessionId"`
	ConversationID  int64   `json:"conversationId"`
}

type KnowledgeAnswerRequest struct {
	KnowledgeBaseID int64   `json:"knowledgeBaseId"`
	Question        string  `json:"question"`
	TopK            int     `json:"topK"`
	ScoreThreshold  float64 `json:"scoreThreshold"`
	RerankLimit     int     `json:"rerankLimit"`
	Channel         string  `json:"channel"`
	Scene           string  `json:"scene"`
	SessionID       string  `json:"sessionId"`
	ConversationID  int64   `json:"conversationId"`
	AnswerMode      int     `json:"answerMode"`
	FallbackMode    int     `json:"fallbackMode"`
}

type KnowledgeCompareRequest struct {
	KnowledgeBaseID int64    `json:"knowledgeBaseId"`
	Question        string   `json:"question"`
	Providers       []string `json:"providers"`
	ExpectedDocIDs  []int64  `json:"expectedDocIds"`
	TopK            int      `json:"topK"`
	ScoreThreshold  float64  `json:"scoreThreshold"`
}

type CreateKnowledgeFeedbackRequest struct {
	RetrieveLogID  int64  `json:"retrieveLogId"`
	FeedbackType   int    `json:"feedbackType"`
	FeedbackReason string `json:"feedbackReason"`
	Remark         string `json:"remark"`
}
