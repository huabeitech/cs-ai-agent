package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"cs-agent/internal/ai/runtime/internal/impl/callbacks"
	"cs-agent/internal/ai/runtime/internal/impl/factory"
	"cs-agent/internal/ai/runtime/internal/impl/retrievers"
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/utils"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/compose"
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

func (g *KnowledgeAnswerabilityGate) withDefaults() *KnowledgeAnswerabilityGate {
	if g == nil {
		return NewKnowledgeAnswerabilityGate()
	}
	ret := *g
	defaults := NewKnowledgeAnswerabilityGate()
	if ret.newRetriever == nil {
		ret.newRetriever = defaults.newRetriever
	}
	if ret.newChatModel == nil {
		ret.newChatModel = defaults.newChatModel
	}
	if ret.now == nil {
		ret.now = time.Now
	}
	return &ret
}

func (g *KnowledgeAnswerabilityGate) Evaluate(ctx context.Context, input answerabilityGateInput) (*answerabilityGateState, error) {
	gate := g.withDefaults()
	graph := compose.NewGraph[*answerabilityGateState, *answerabilityGateState]()
	if err := graph.AddLambdaNode(answerabilityNodeRetrieve, compose.InvokableLambda(gate.retrieveKnowledge)); err != nil {
		return nil, err
	}
	if err := graph.AddLambdaNode(answerabilityNodeGrade, compose.InvokableLambda(gate.gradeAnswerability)); err != nil {
		return nil, err
	}
	if err := graph.AddLambdaNode(answerabilityNodeAllow, compose.InvokableLambda(allowAnswerabilityPassThrough)); err != nil {
		return nil, err
	}
	if err := graph.AddLambdaNode(answerabilityNodeFallback, compose.InvokableLambda(fallbackAnswerabilityPassThrough)); err != nil {
		return nil, err
	}
	if err := graph.AddEdge(compose.START, answerabilityNodeRetrieve); err != nil {
		return nil, err
	}
	if err := graph.AddEdge(answerabilityNodeRetrieve, answerabilityNodeGrade); err != nil {
		return nil, err
	}
	if err := graph.AddBranch(answerabilityNodeGrade, compose.NewGraphBranch(routeAnswerabilityGate, map[string]bool{
		answerabilityNodeAllow:    true,
		answerabilityNodeFallback: true,
	})); err != nil {
		return nil, err
	}
	if err := graph.AddEdge(answerabilityNodeAllow, compose.END); err != nil {
		return nil, err
	}
	if err := graph.AddEdge(answerabilityNodeFallback, compose.END); err != nil {
		return nil, err
	}
	runnable, err := graph.Compile(ctx)
	if err != nil {
		return nil, err
	}
	return runnable.Invoke(ctx, &answerabilityGateState{Input: input})
}

func routeAnswerabilityGate(ctx context.Context, state *answerabilityGateState) (string, error) {
	if state == nil {
		return answerabilityNodeFallback, nil
	}
	if state.SkipGate || strings.TrimSpace(state.FallbackReply) == "" {
		return answerabilityNodeAllow, nil
	}
	return answerabilityNodeFallback, nil
}

func allowAnswerabilityPassThrough(ctx context.Context, state *answerabilityGateState) (*answerabilityGateState, error) {
	if state == nil {
		return &answerabilityGateState{}, nil
	}
	if len(state.Decision.Instructions) > 0 {
		state.Input.Messages = append(state.Input.Messages, state.Decision.Instructions...)
	}
	if state.RetrieveResult != nil {
		if contextText := strings.TrimSpace(state.RetrieveResult.ContextText); contextText != "" {
			state.Input.Messages = append(state.Input.Messages, schema.SystemMessage(contextText))
		}
	}
	return state, nil
}

func fallbackAnswerabilityPassThrough(ctx context.Context, state *answerabilityGateState) (*answerabilityGateState, error) {
	if state == nil {
		return &answerabilityGateState{}, nil
	}
	return state, nil
}

func (g *KnowledgeAnswerabilityGate) retrieveKnowledge(ctx context.Context, state *answerabilityGateState) (*answerabilityGateState, error) {
	if state == nil {
		state = &answerabilityGateState{}
	}
	gate := g.withDefaults()
	req := state.Input.Request
	configuredKnowledgeIDs := utils.SplitInt64s(req.AIAgent.KnowledgeIDs)
	retriever := gate.newRetriever(req.AIAgent)
	if retriever == nil {
		state.KnowledgeIDs = append([]int64(nil), configuredKnowledgeIDs...)
		if len(configuredKnowledgeIDs) == 0 {
			state.SkipGate = true
			state.recordAnswerability(answerabilityStatusSkipped, "no knowledge configured", nil)
			return state, nil
		}
		state.FallbackReply = resolveKnowledgeHumanSupportFallback(req.AIAgent)
		state.recordAnswerability(answerabilityStatusUnanswerable, "knowledge retriever unavailable", nil)
		return state, nil
	}
	knowledgeIDs := retriever.KnowledgeBaseIDs()
	state.KnowledgeIDs = append([]int64(nil), knowledgeIDs...)
	if len(knowledgeIDs) == 0 {
		state.SkipGate = true
		state.recordAnswerability(answerabilityStatusSkipped, "no knowledge configured", nil)
		return state, nil
	}
	query := strings.TrimSpace(req.UserMessage.Content)
	if query == "" {
		state.FallbackReply = resolveKnowledgeHumanSupportFallback(req.AIAgent)
		state.recordAnswerability(answerabilityStatusUnanswerable, "empty user question", nil)
		return state, nil
	}
	retrieveOptions := retrievers.DefaultKnowledgeRetrieveOptions()
	retrieveOptions.QueryPreview = preview(req.UserMessage.Content, 120)
	result, err := retriever.RetrieveContextByOptions(ctx, retrieveOptions, query)
	if err != nil {
		state.FallbackReply = resolveKnowledgeHumanSupportFallback(req.AIAgent)
		state.ErrorMessage = err.Error()
		state.recordAnswerability(answerabilityStatusUnanswerable, "knowledge retrieval failed", err)
		return state, nil
	}
	state.RetrieveResult = result
	if state.Input.Summary != nil && result != nil {
		state.Input.Summary.RetrieverCount = len(result.Hits)
	}
	if state.Input.Collector != nil && result != nil {
		state.Input.Collector.SetRetrieverSummary(result.TraceSummary)
		state.Input.Collector.AddRetrieverItems(result.TraceItems)
	}
	if result == nil || len(result.Hits) == 0 || strings.TrimSpace(result.ContextText) == "" {
		state.FallbackReply = resolveKnowledgeHumanSupportFallback(req.AIAgent)
		state.recordAnswerability(answerabilityStatusUnanswerable, "no retrieved context", nil)
		return state, nil
	}
	return state, nil
}

func (g *KnowledgeAnswerabilityGate) gradeAnswerability(ctx context.Context, state *answerabilityGateState) (*answerabilityGateState, error) {
	if state == nil {
		return &answerabilityGateState{}, nil
	}
	if state.SkipGate || strings.TrimSpace(state.FallbackReply) != "" {
		return state, nil
	}
	gate := g.withDefaults()
	started := gate.now()
	req := state.Input.Request
	modelInstance, err := gate.newChatModel(ctx, req.AIConfig)
	if err != nil {
		state.FallbackReply = resolveKnowledgeHumanSupportFallback(req.AIAgent)
		state.ErrorMessage = err.Error()
		state.recordAnswerabilityWithLatency(answerabilityStatusUnanswerable, "answerability model factory failed", err, started)
		return state, nil
	}
	if modelInstance == nil {
		err = fmt.Errorf("answerability model is nil")
		state.FallbackReply = resolveKnowledgeHumanSupportFallback(req.AIAgent)
		state.ErrorMessage = err.Error()
		state.recordAnswerabilityWithLatency(answerabilityStatusUnanswerable, "answerability model factory failed", err, started)
		return state, nil
	}
	messages, err := buildAnswerabilityMessages(ctx, req.UserMessage.Content, state.RetrieveResult)
	if err != nil {
		state.FallbackReply = resolveKnowledgeHumanSupportFallback(req.AIAgent)
		state.ErrorMessage = err.Error()
		state.recordAnswerabilityWithLatency(answerabilityStatusUnanswerable, "answerability prompt failed", err, started)
		return state, nil
	}
	response, err := modelInstance.Generate(ctx, messages)
	if err != nil {
		state.FallbackReply = resolveKnowledgeHumanSupportFallback(req.AIAgent)
		state.ErrorMessage = err.Error()
		state.recordAnswerabilityWithLatency(answerabilityStatusUnanswerable, "answerability model generate failed", err, started)
		return state, nil
	}
	if response == nil {
		err = fmt.Errorf("answerability model returned empty response")
		state.FallbackReply = resolveKnowledgeHumanSupportFallback(req.AIAgent)
		state.ErrorMessage = err.Error()
		state.recordAnswerabilityWithLatency(answerabilityStatusUnanswerable, "answerability model generate failed", err, started)
		return state, nil
	}
	decision, err := parseAnswerabilityDecision(response.Content)
	if err != nil {
		state.FallbackReply = resolveKnowledgeHumanSupportFallback(req.AIAgent)
		state.ErrorMessage = err.Error()
		state.recordAnswerabilityWithLatency(answerabilityStatusUnanswerable, "answerability decision parse failed", err, started)
		return state, nil
	}
	state.Grade = decision
	if !decision.Answerable {
		state.FallbackReply = resolveKnowledgeHumanSupportFallback(req.AIAgent)
		state.recordAnswerabilityWithLatency(answerabilityStatusUnanswerable, decision.Reason, nil, started)
		return state, nil
	}
	state.Decision = buildKnowledgeGuardDecision(req.AIAgent, state.RetrieveResult)
	if strings.TrimSpace(state.Decision.FallbackReply) != "" {
		state.FallbackReply = resolveKnowledgeHumanSupportFallback(req.AIAgent)
		state.recordAnswerabilityWithLatency(answerabilityStatusUnanswerable, "knowledge guard rejected retrieved context", nil, started)
		return state, nil
	}
	state.recordAnswerabilityWithLatency(answerabilityStatusAnswerable, decision.Reason, nil, started)
	return state, nil
}

func buildAnswerabilityMessages(ctx context.Context, question string, result *retrievers.KnowledgeRetrieveResult) ([]*schema.Message, error) {
	contextText := buildAnswerabilityContext(result)
	template := prompt.FromMessages(schema.FString,
		schema.SystemMessage(strings.TrimSpace(`你是一个知识库可回答性判定器。
你只判断“已召回知识片段”是否直接支持回答用户问题，不要使用模型常识补充。
如果问题中的具体对象、条件、步骤、承诺或限制不能被片段直接支持，判定为不可回答。
只输出 JSON，不要输出 Markdown、解释或多余文本。
JSON 字段必须包含：
- answerable: boolean
- reason: string
- supportingChunkIds: string array，answerable 为 true 时必须至少包含一个直接支持的 chunk id
- missingInfo: string array，answerable 为 false 时列出缺失信息`)),
		schema.UserMessage(strings.TrimSpace(`用户问题：
{question}

已召回知识片段：
{context}

请基于上述片段判定是否可以直接回答用户问题。`)),
	)
	return template.Format(ctx, map[string]any{
		"question": strings.TrimSpace(question),
		"context":  contextText,
	})
}

func buildAnswerabilityContext(result *retrievers.KnowledgeRetrieveResult) string {
	if result == nil {
		return ""
	}
	items := result.ContextResults
	if len(items) == 0 {
		items = result.Hits
	}
	if len(items) == 0 {
		return strings.TrimSpace(result.ContextText)
	}
	var builder strings.Builder
	for idx, item := range items {
		if idx > 0 {
			builder.WriteString("\n\n")
		}
		builder.WriteString(fmt.Sprintf("snippet %d\nknowledgeBaseId: %d\ndocumentId: %d\nchunkId: %d\nscore: %.4f\ncontent:\n%s",
			idx+1,
			item.KnowledgeBaseID,
			item.DocumentID,
			item.ChunkID,
			item.Score,
			strings.TrimSpace(item.Content),
		))
	}
	return strings.TrimSpace(builder.String())
}

func parseAnswerabilityDecision(raw string) (answerabilityDecision, error) {
	text := trimMarkdownFence(raw)
	if text == "" {
		return answerabilityDecision{}, fmt.Errorf("answerability decision is empty")
	}
	var decision answerabilityDecision
	if err := json.Unmarshal([]byte(text), &decision); err != nil {
		return answerabilityDecision{}, fmt.Errorf("parse answerability decision: %w", err)
	}
	decision.Reason = strings.TrimSpace(decision.Reason)
	decision.SupportingChunkIDs = trimStringSlice(decision.SupportingChunkIDs)
	decision.MissingInfo = trimStringSlice(decision.MissingInfo)
	if decision.Answerable && len(decision.SupportingChunkIDs) == 0 {
		return answerabilityDecision{}, fmt.Errorf("answerable decision requires supportingChunkIds")
	}
	return decision, nil
}

func trimMarkdownFence(raw string) string {
	text := strings.TrimSpace(raw)
	if !strings.HasPrefix(text, "```") {
		return text
	}
	lines := strings.Split(text, "\n")
	if len(lines) == 0 {
		return text
	}
	if strings.HasPrefix(strings.TrimSpace(lines[0]), "```") {
		lines = lines[1:]
	}
	if len(lines) > 0 && strings.HasPrefix(strings.TrimSpace(lines[len(lines)-1]), "```") {
		lines = lines[:len(lines)-1]
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}

func trimStringSlice(items []string) []string {
	if len(items) == 0 {
		return nil
	}
	ret := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		ret = append(ret, item)
	}
	return ret
}

func (s *answerabilityGateState) recordAnswerability(status string, reason string, err error) {
	s.recordAnswerabilityWithLatency(status, reason, err, time.Time{})
}

func (s *answerabilityGateState) recordAnswerabilityWithLatency(status string, reason string, err error, started time.Time) {
	if s == nil || s.Input.Collector == nil {
		return
	}
	errorMessage := strings.TrimSpace(s.ErrorMessage)
	if err != nil {
		errorMessage = err.Error()
	}
	data := callbacks.AnswerabilityTraceData{
		Status:             status,
		Reason:             strings.TrimSpace(reason),
		SupportingChunkIDs: append([]string(nil), s.Grade.SupportingChunkIDs...),
		MissingInfo:        append([]string(nil), s.Grade.MissingInfo...),
		ErrorMessage:       errorMessage,
	}
	if !started.IsZero() {
		data.LatencyMs = time.Since(started).Milliseconds()
	}
	s.Input.Collector.SetAnswerability(data)
}
