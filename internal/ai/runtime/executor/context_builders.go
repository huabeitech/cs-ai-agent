package executor

import (
	"context"
	"strings"

	"cs-agent/internal/ai/runtime/internal/impl/adapter"
	"cs-agent/internal/ai/runtime/internal/impl/callbacks"
	"cs-agent/internal/pkg/utils"

	"github.com/cloudwego/eino/schema"
)

func buildRunMessages(ctx context.Context, req RunInput, summary *RunResult, collector *callbacks.RuntimeTraceCollector, gate *KnowledgeAnswerabilityGate) []*schema.Message {
	history := adapter.BuildHistoryMessages(req.Conversation.ID, req.UserMessage.ID, 12)
	if summary != nil {
		summary.HistoryMessageCount = len(history.Messages)
	}
	if collector != nil {
		collector.Data.Input.HistoryMessageCount = len(history.Messages)
		collector.Data.Input.KnowledgeBaseIDs = utils.SplitInt64s(req.AIAgent.KnowledgeIDs)
		collector.Data.Input.CurrentUserMessagePreview = preview(req.UserMessage.Content, 120)
	}
	messages := make([]*schema.Message, 0, len(history.Messages)+3)
	messages = append(messages, history.Messages...)
	decision := appendRetrievedContext(ctx, req, summary, collector, gate, &messages)
	if strings.TrimSpace(decision.FallbackReply) != "" {
		if summary != nil {
			summary.ReplyText = decision.FallbackReply
		}
		return messages
	}
	messages = append(messages, schema.UserMessage(strings.TrimSpace(req.UserMessage.Content)))
	return messages
}

func appendRetrievedContext(ctx context.Context, req RunInput, summary *RunResult, collector *callbacks.RuntimeTraceCollector, gate *KnowledgeAnswerabilityGate, messages *[]*schema.Message) knowledgeGuardDecision {
	if messages == nil {
		return knowledgeGuardDecision{}
	}
	if gate == nil {
		gate = NewKnowledgeAnswerabilityGate()
	}
	state, err := gate.Evaluate(ctx, answerabilityGateInput{
		Request:   req,
		Summary:   summary,
		Collector: collector,
		Messages:  append([]*schema.Message(nil), (*messages)...),
	})
	if err != nil || state == nil {
		decision := buildKnowledgeUnavailableDecision(req.AIAgent, utils.SplitInt64s(req.AIAgent.KnowledgeIDs))
		if strings.TrimSpace(decision.FallbackReply) != "" {
			decision.FallbackReply = resolveKnowledgeHumanSupportFallback(req.AIAgent)
		}
		return decision
	}
	if strings.TrimSpace(state.FallbackReply) != "" {
		return knowledgeGuardDecision{FallbackReply: state.FallbackReply}
	}
	if state.SkipGate {
		return knowledgeGuardDecision{}
	}
	*messages = append((*messages)[:0], state.Input.Messages...)
	return state.Decision
}
