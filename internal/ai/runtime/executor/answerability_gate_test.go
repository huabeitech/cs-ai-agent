package executor

import (
	"context"
	"errors"
	"strings"
	"testing"

	"cs-agent/internal/ai/rag"
	"cs-agent/internal/ai/runtime/internal/impl/callbacks"
	"cs-agent/internal/ai/runtime/internal/impl/retrievers"
	"cs-agent/internal/models"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

func TestParseAnswerabilityDecisionRejectsMalformedJSON(t *testing.T) {
	_, err := parseAnswerabilityDecision(`{"answerable": true`)

	if err == nil {
		t.Fatal("expected malformed JSON to be rejected")
	}
}

func TestParseAnswerabilityDecisionRejectsAnswerableWithoutSupportingChunkIDs(t *testing.T) {
	_, err := parseAnswerabilityDecision(`{"answerable": true, "reason": "directly supported"}`)

	if err == nil {
		t.Fatal("expected answerable decision without supporting chunks to be rejected")
	}
}

func TestParseAnswerabilityDecisionAcceptsAnswerableWithSupportingChunkIDs(t *testing.T) {
	got, err := parseAnswerabilityDecision("```json\n{\"answerable\": true, \"reason\": \"directly supported\", \"supportingChunkIds\": [\" chunk-1 \", \"chunk-2\"]}\n```")
	if err != nil {
		t.Fatalf("parse decision failed: %v", err)
	}

	if !got.Answerable {
		t.Fatal("expected answerable decision")
	}
	if got.SupportingChunkIDs[0] != "chunk-1" || got.SupportingChunkIDs[1] != "chunk-2" {
		t.Fatalf("unexpected supporting chunks: %#v", got.SupportingChunkIDs)
	}
}

func TestKnowledgeAnswerabilityGateEvaluateFallsBackOnRetrieverError(t *testing.T) {
	collector := callbacks.NewRuntimeTraceCollector()
	gate := newTestKnowledgeAnswerabilityGate(&fakeKnowledgeContextRetriever{
		knowledgeBaseIDs: []int64{1},
		err:              errors.New("vector store unavailable"),
	}, nil)

	state, err := gate.Evaluate(context.Background(), answerabilityGateInput{
		Request:   newAnswerabilityGateRunInput("是否支持退款？", "1"),
		Summary:   &RunResult{},
		Collector: collector,
	})
	if err != nil {
		t.Fatalf("Evaluate returned error: %v", err)
	}

	if !strings.Contains(state.FallbackReply, "建议你联系人工客服进一步确认。") {
		t.Fatalf("expected human-support fallback, got %q", state.FallbackReply)
	}
	if collector.Data.Answerability.Status != answerabilityStatusUnanswerable {
		t.Fatalf("unexpected answerability status: %q", collector.Data.Answerability.Status)
	}
	if collector.Data.Answerability.Reason != "knowledge retrieval failed" {
		t.Fatalf("unexpected reason: %q", collector.Data.Answerability.Reason)
	}
}

func TestKnowledgeAnswerabilityGateEvaluateSkipsWhenNoKnowledgeConfigured(t *testing.T) {
	collector := callbacks.NewRuntimeTraceCollector()
	gate := newTestKnowledgeAnswerabilityGate(&fakeKnowledgeContextRetriever{}, nil)

	state, err := gate.Evaluate(context.Background(), answerabilityGateInput{
		Request:   newAnswerabilityGateRunInput("是否支持退款？", ""),
		Collector: collector,
	})
	if err != nil {
		t.Fatalf("Evaluate returned error: %v", err)
	}

	if !state.SkipGate {
		t.Fatal("expected gate to skip without knowledge")
	}
	if state.FallbackReply != "" {
		t.Fatalf("expected no fallback when gate skips, got %q", state.FallbackReply)
	}
	if collector.Data.Answerability.Status != answerabilityStatusSkipped {
		t.Fatalf("unexpected answerability status: %q", collector.Data.Answerability.Status)
	}
}

func TestKnowledgeAnswerabilityGateEvaluateFallsBackWhenConfiguredRetrieverUnavailable(t *testing.T) {
	collector := callbacks.NewRuntimeTraceCollector()
	gate := newTestKnowledgeAnswerabilityGate(nil, nil)

	state, err := gate.Evaluate(context.Background(), answerabilityGateInput{
		Request:   newAnswerabilityGateRunInput("是否支持退款？", "1"),
		Collector: collector,
	})
	if err != nil {
		t.Fatalf("Evaluate returned error: %v", err)
	}

	if state.SkipGate {
		t.Fatal("expected configured knowledge to fail closed, not skip")
	}
	if !strings.Contains(state.FallbackReply, "建议你联系人工客服进一步确认。") {
		t.Fatalf("expected human-support fallback, got %q", state.FallbackReply)
	}
	if collector.Data.Answerability.Status != answerabilityStatusUnanswerable {
		t.Fatalf("unexpected answerability status: %q", collector.Data.Answerability.Status)
	}
	if collector.Data.Answerability.Reason != "knowledge retriever unavailable" {
		t.Fatalf("unexpected reason: %q", collector.Data.Answerability.Reason)
	}
}

func TestKnowledgeAnswerabilityGateEvaluateFallsBackOnGrayZoneUnanswerableDecision(t *testing.T) {
	collector := callbacks.NewRuntimeTraceCollector()
	gate := newTestKnowledgeAnswerabilityGate(newAnswerabilityRetrieverWithHit(), &fakeAnswerabilityChatModel{
		response: `{"answerable": false, "reason": "retrieved snippets mention refunds but not the requested condition", "missingInfo": ["refund condition"]}`,
	})

	state, err := gate.Evaluate(context.Background(), answerabilityGateInput{
		Request:   newAnswerabilityGateRunInput("满足什么条件可以退款？", "1"),
		Collector: collector,
	})
	if err != nil {
		t.Fatalf("Evaluate returned error: %v", err)
	}

	if !strings.Contains(state.FallbackReply, "建议你联系人工客服进一步确认。") {
		t.Fatalf("expected human-support fallback, got %q", state.FallbackReply)
	}
	if collector.Data.Answerability.Status != answerabilityStatusUnanswerable {
		t.Fatalf("unexpected answerability status: %q", collector.Data.Answerability.Status)
	}
	if collector.Data.Answerability.MissingInfo[0] != "refund condition" {
		t.Fatalf("unexpected missing info: %#v", collector.Data.Answerability.MissingInfo)
	}
}

func TestKnowledgeAnswerabilityGateEvaluateAllowsAnswerableDecisionAndProducesKnowledgeInstruction(t *testing.T) {
	collector := callbacks.NewRuntimeTraceCollector()
	summary := &RunResult{}
	chatModel := &fakeAnswerabilityChatModel{
		response: `{"answerable": true, "reason": "refund condition is directly supported", "supportingChunkIds": ["101"]}`,
	}
	gate := newTestKnowledgeAnswerabilityGate(newAnswerabilityRetrieverWithHit(), chatModel)

	state, err := gate.Evaluate(context.Background(), answerabilityGateInput{
		Request:   newAnswerabilityGateRunInput("满足什么条件可以退款？", "1"),
		Summary:   summary,
		Collector: collector,
	})
	if err != nil {
		t.Fatalf("Evaluate returned error: %v", err)
	}

	if state.FallbackReply != "" {
		t.Fatalf("expected answerable gate to allow, got fallback %q", state.FallbackReply)
	}
	if len(state.Decision.Instructions) != 1 {
		t.Fatalf("expected one knowledge instruction, got %d", len(state.Decision.Instructions))
	}
	if !strings.Contains(state.Decision.Instructions[0].Content, "知识库回答约束") {
		t.Fatalf("unexpected instruction: %q", state.Decision.Instructions[0].Content)
	}
	if summary.RetrieverCount != 1 {
		t.Fatalf("expected retriever count 1, got %d", summary.RetrieverCount)
	}
	if collector.Data.Answerability.Status != answerabilityStatusAnswerable {
		t.Fatalf("unexpected answerability status: %q", collector.Data.Answerability.Status)
	}
	if collector.Data.Answerability.SupportingChunkIDs[0] != "101" {
		t.Fatalf("unexpected supporting chunks: %#v", collector.Data.Answerability.SupportingChunkIDs)
	}
	if len(chatModel.input) == 0 || !strings.Contains(chatModel.input[len(chatModel.input)-1].Content, "chunkId: 101") {
		t.Fatalf("expected grader prompt to include chunk id, got %#v", chatModel.input)
	}
}

func newTestKnowledgeAnswerabilityGate(retriever knowledgeContextRetriever, chatModel model.BaseChatModel) *KnowledgeAnswerabilityGate {
	return &KnowledgeAnswerabilityGate{
		newRetriever: func(aiAgent models.AIAgent) knowledgeContextRetriever {
			return retriever
		},
		newChatModel: func(ctx context.Context, aiConfig models.AIConfig) (model.BaseChatModel, error) {
			return chatModel, nil
		},
	}
}

func newAnswerabilityGateRunInput(question string, knowledgeIDs string) RunInput {
	return RunInput{
		UserMessage: models.Message{Content: question},
		AIAgent: models.AIAgent{
			KnowledgeIDs: knowledgeIDs,
		},
	}
}

func newAnswerabilityRetrieverWithHit() *fakeKnowledgeContextRetriever {
	return &fakeKnowledgeContextRetriever{
		knowledgeBaseIDs: []int64{1},
		result: &retrievers.KnowledgeRetrieveResult{
			KnowledgeBaseIDs: []int64{1},
			Hits: []rag.RetrieveResult{
				{KnowledgeBaseID: 1, DocumentID: 10, ChunkID: 101, Score: 0.93, Content: "购买后七天内且未使用可以退款。"},
			},
			ContextResults: []rag.RetrieveResult{
				{KnowledgeBaseID: 1, DocumentID: 10, ChunkID: 101, Score: 0.93, Content: "购买后七天内且未使用可以退款。"},
			},
			ContextText: "购买后七天内且未使用可以退款。",
		},
	}
}

type fakeKnowledgeContextRetriever struct {
	knowledgeBaseIDs []int64
	result           *retrievers.KnowledgeRetrieveResult
	err              error
	lastOptions      retrievers.KnowledgeRetrieveOptions
	lastQuery        string
}

func (f *fakeKnowledgeContextRetriever) KnowledgeBaseIDs() []int64 {
	return append([]int64(nil), f.knowledgeBaseIDs...)
}

func (f *fakeKnowledgeContextRetriever) RetrieveContextByOptions(ctx context.Context, opts retrievers.KnowledgeRetrieveOptions, query string) (*retrievers.KnowledgeRetrieveResult, error) {
	f.lastOptions = opts
	f.lastQuery = query
	return f.result, f.err
}

type fakeAnswerabilityChatModel struct {
	response string
	err      error
	input    []*schema.Message
}

func (f *fakeAnswerabilityChatModel) Generate(ctx context.Context, input []*schema.Message, opts ...model.Option) (*schema.Message, error) {
	f.input = input
	if f.err != nil {
		return nil, f.err
	}
	return schema.AssistantMessage(f.response, nil), nil
}

func (f *fakeAnswerabilityChatModel) Stream(ctx context.Context, input []*schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	return nil, errors.New("stream is not implemented in fakeAnswerabilityChatModel")
}
