package executor

import (
	"context"
	"strings"

	"cs-agent/internal/ai/runtime/internal/impl/adapter"
	"cs-agent/internal/ai/runtime/internal/impl/callbacks"
	"cs-agent/internal/ai/runtime/internal/impl/retrievers"
	"cs-agent/internal/pkg/utils"

	"github.com/cloudwego/eino/schema"
)

func buildRunMessages(ctx context.Context, req RunInput, summary *RunResult, collector *callbacks.RuntimeTraceCollector) []*schema.Message {
	if req.Conversation == nil || req.UserMessage == nil {
		return nil
	}
	history := adapter.BuildHistoryMessages(req.Conversation.ID, req.UserMessage.ID, 12)
	if summary != nil {
		summary.HistoryMessageCount = len(history.Messages)
	}
	if collector != nil {
		collector.Data.Input.HistoryMessageCount = len(history.Messages)
		if req.AIAgent != nil {
			collector.Data.Input.KnowledgeBaseIDs = utils.SplitInt64s(req.AIAgent.KnowledgeIDs)
		}
		collector.Data.Input.CurrentUserMessagePreview = preview(req.UserMessage.Content, 120)
	}
	messages := make([]*schema.Message, 0, len(history.Messages)+3)
	messages = append(messages, history.Messages...)
	appendRetrievedContext(ctx, req, summary, collector, &messages)
	messages = append(messages, schema.UserMessage(strings.TrimSpace(req.UserMessage.Content)))
	return messages
}

func appendRetrievedContext(ctx context.Context, req RunInput, summary *RunResult, collector *callbacks.RuntimeTraceCollector, messages *[]*schema.Message) {
	if req.AIAgent == nil || req.UserMessage == nil || messages == nil {
		return
	}
	retriever := retrievers.NewKnowledgeRetriever(req.AIAgent)
	retrieveOptions := retrievers.DefaultKnowledgeRetrieveOptions()
	retrieveOptions.QueryPreview = preview(req.UserMessage.Content, 120)
	retrieveResult, retrieveErr := retriever.RetrieveContextByOptions(ctx, retrieveOptions, strings.TrimSpace(req.UserMessage.Content))
	if retrieveErr != nil || retrieveResult == nil {
		return
	}
	if summary != nil {
		summary.RetrieverCount = len(retrieveResult.Hits)
	}
	if collector != nil {
		collector.SetRetrieverSummary(retrieveResult.TraceSummary)
		collector.Data.Retriever.Items = append(collector.Data.Retriever.Items, retrieveResult.TraceItems...)
	}
	if strings.TrimSpace(retrieveResult.ContextText) != "" {
		*messages = append(*messages, schema.SystemMessage(retrieveResult.ContextText))
	}
}
