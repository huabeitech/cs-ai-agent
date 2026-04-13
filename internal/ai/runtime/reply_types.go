package runtime

import "encoding/json"

type aiReplyTraceData struct {
	Status           string          `json:"status"`
	RuntimeLatencyMs int64           `json:"runtimeLatencyMs,omitempty"`
	RecheckMs        int64           `json:"recheckMs,omitempty"`
	CommitMs         int64           `json:"commitMs,omitempty"`
	FinalAction      string          `json:"finalAction,omitempty"`
	ResumeSource     string          `json:"resumeSource,omitempty"`
	ReplySent        bool            `json:"replySent,omitempty"`
	ReplyMessageID   int64           `json:"replyMessageId,omitempty"`
	Runtime          json.RawMessage `json:"runtime,omitempty"`
}

const (
	defaultAIReplyAsyncTimeoutSeconds = 180
	maxAIReplyAsyncTimeoutSeconds     = 600
)
