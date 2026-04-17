package runtime

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	applicationruntime "cs-agent/internal/ai/application/runtime"
	"cs-agent/internal/ai/runtime/graphs"
	"cs-agent/internal/models"
	svc "cs-agent/internal/services"
)

type runtimeReplyExecutor struct{}

func newRuntimeReplyExecutor() *runtimeReplyExecutor {
	return &runtimeReplyExecutor{}
}

func (e *runtimeReplyExecutor) Run(ctx context.Context, conversation models.Conversation, message models.Message, aiAgent models.AIAgent, trace *aiReplyTraceData) (*applicationruntime.Summary, error) {
	aiConfig := svc.AIConfigService.Get(aiAgent.AIConfigID)
	if aiConfig == nil {
		return nil, fmt.Errorf("ai config is nil")
	}
	runtimeStartedAt := time.Now()
	summary, err := Service.Run(ctx, applicationruntime.Request{
		Conversation: conversation,
		UserMessage:  message,
		AIAgent:      aiAgent,
		AIConfig:     *aiConfig,
	})
	if trace != nil {
		trace.RuntimeLatencyMs = time.Since(runtimeStartedAt).Milliseconds()
		e.fillTraceFromSummary(trace, summary, err)
	}
	return summary, err
}

func (e *runtimeReplyExecutor) ResumePendingInterrupt(ctx context.Context, conversation models.Conversation, message models.Message, aiAgent models.AIAgent, pendingInterrupt *models.ConversationInterrupt, trace *aiReplyTraceData) (*applicationruntime.Summary, error) {
	if pendingInterrupt == nil {
		return nil, nil
	}
	aiConfig := svc.AIConfigService.Get(aiAgent.AIConfigID)
	if aiConfig == nil {
		return nil, fmt.Errorf("ai config is nil")
	}
	runtimeStartedAt := time.Now()
	if trace != nil {
		trace.ResumeSource = "pending_interrupt"
	}
	summary, err := Service.Resume(ctx, applicationruntime.ResumeRequest{
		Conversation: conversation,
		AIAgent:      aiAgent,
		AIConfig:     *aiConfig,
		CheckPointID: strings.TrimSpace(pendingInterrupt.CheckPointID),
		ResumeData: map[string]string{
			strings.TrimSpace(pendingInterrupt.InterruptID): strings.TrimSpace(message.Content),
		},
	})
	if trace != nil {
		trace.RuntimeLatencyMs = time.Since(runtimeStartedAt).Milliseconds()
		e.fillTraceFromSummary(trace, summary, err)
	}
	return summary, err
}

func (e *runtimeReplyExecutor) fillTraceFromSummary(trace *aiReplyTraceData, summary *applicationruntime.Summary, runErr error) {
	if trace == nil {
		return
	}
	if runErr != nil {
		trace.Status = "runtime_error"
		trace.FinalAction = "error"
		if summary != nil {
			trace.Runtime = json.RawMessage(summary.TraceData)
		}
		return
	}
	trace.Status = "runtime_prepared"
	trace.FinalAction = toRunLogFinalAction(summary)
	if summary != nil && strings.TrimSpace(summary.TraceData) != "" {
		trace.Runtime = json.RawMessage(summary.TraceData)
	}
}

func expiredInterruptSummary() *applicationruntime.Summary {
	return &applicationruntime.Summary{
		Status:    "expired",
		ReplyText: graphs.ConfirmationExpiredReply,
	}
}
