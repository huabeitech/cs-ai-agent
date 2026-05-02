package executor

import (
	"context"
	"time"

	"cs-agent/internal/ai/runtime/internal/impl/callbacks"
	"cs-agent/internal/ai/runtime/internal/impl/factory"
	"cs-agent/internal/ai/runtime/internal/impl/retrievers"
	"cs-agent/internal/models"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

const (
	answerabilityNodeRetrieve = "retrieve_knowledge"
	answerabilityNodeGrade    = "grade_answerability"
	answerabilityNodeAllow    = "allow_agent"
	answerabilityNodeFallback = "fallback"

	answerabilityStatusSkipped      = "skipped"
	answerabilityStatusAnswerable   = "answerable"
	answerabilityStatusUnanswerable = "unanswerable"
)

type knowledgeContextRetriever interface {
	KnowledgeBaseIDs() []int64
	RetrieveContextByOptions(ctx context.Context, opts retrievers.KnowledgeRetrieveOptions, query string) (*retrievers.KnowledgeRetrieveResult, error)
}

type answerabilityRetrieverFactory func(aiAgent models.AIAgent) knowledgeContextRetriever

type answerabilityChatModelFactory func(ctx context.Context, aiConfig models.AIConfig) (model.BaseChatModel, error)

type KnowledgeAnswerabilityGate struct {
	newRetriever answerabilityRetrieverFactory
	newChatModel answerabilityChatModelFactory
	now          func() time.Time
}

type answerabilityGateInput struct {
	Request   RunInput
	Summary   *RunResult
	Collector *callbacks.RuntimeTraceCollector
	Messages  []*schema.Message
}

type answerabilityGateState struct {
	Input          answerabilityGateInput
	KnowledgeIDs   []int64
	RetrieveResult *retrievers.KnowledgeRetrieveResult
	Decision       knowledgeGuardDecision
	Grade          answerabilityDecision
	SkipGate       bool
	FallbackReply  string
	ErrorMessage   string
}

type answerabilityDecision struct {
	Answerable         bool     `json:"answerable"`
	Reason             string   `json:"reason"`
	SupportingChunkIDs []string `json:"supportingChunkIds"`
	MissingInfo        []string `json:"missingInfo"`
}

func NewKnowledgeAnswerabilityGate() *KnowledgeAnswerabilityGate {
	chatModelFactory := factory.NewChatModelFactory()
	return &KnowledgeAnswerabilityGate{
		newRetriever: func(aiAgent models.AIAgent) knowledgeContextRetriever {
			return retrievers.NewKnowledgeRetriever(aiAgent)
		},
		newChatModel: func(ctx context.Context, aiConfig models.AIConfig) (model.BaseChatModel, error) {
			return chatModelFactory.Build(ctx, aiConfig)
		},
		now: time.Now,
	}
}
