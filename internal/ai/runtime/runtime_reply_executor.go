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

type runtimeReplyRunInput struct {
	Conversation models.Conversation
	Message      models.Message
	AIAgent      models.AIAgent
	Trace        *aiReplyTraceData
}

type runtimeReplyResumeInput struct {
	Conversation     models.Conversation
	Message          models.Message
	AIAgent          models.AIAgent
	PendingInterrupt *models.ConversationInterrupt
	Trace            *aiReplyTraceData
}

func newRuntimeReplyExecutor() *runtimeReplyExecutor {
	return &runtimeReplyExecutor{}
}

func (e *runtimeReplyExecutor) Run(ctx context.Context, input runtimeReplyRunInput) (*applicationruntime.Summary, error) {
	aiConfig := svc.AIConfigService.Get(input.AIAgent.AIConfigID)
	if aiConfig == nil {
		return nil, fmt.Errorf("ai config is nil")
	}
	runtimeStartedAt := time.Now()
	summary, err := Service.Run(ctx, applicationruntime.Request{
		Conversation: input.Conversation,
		UserMessage:  input.Message,
		AIAgent:      input.AIAgent,
		AIConfig:     *aiConfig,
	})
	if input.Trace != nil {
		input.Trace.RuntimeLatencyMs = time.Since(runtimeStartedAt).Milliseconds()
		e.fillTraceFromSummary(input.Trace, summary, err)
	}
	return summary, err
}

func (e *runtimeReplyExecutor) ResumePendingInterrupt(ctx context.Context, input runtimeReplyResumeInput) (*applicationruntime.Summary, error) {
	if input.PendingInterrupt == nil {
		return nil, fmt.Errorf("pending interrupt is required")
	}
	aiConfig := svc.AIConfigService.Get(input.AIAgent.AIConfigID)
	if aiConfig == nil {
		return nil, fmt.Errorf("ai config is nil")
	}
	runtimeStartedAt := time.Now()
	if input.Trace != nil {
		input.Trace.ResumeSource = "pending_interrupt"
	}
	summary, err := Service.Resume(ctx, applicationruntime.ResumeRequest{
		Conversation: input.Conversation,
		AIAgent:      input.AIAgent,
		AIConfig:     *aiConfig,
		CheckPointID: strings.TrimSpace(input.PendingInterrupt.CheckPointID),
		ResumeData: map[string]string{
			strings.TrimSpace(input.PendingInterrupt.InterruptID): strings.TrimSpace(input.Message.Content),
		},
	})
	if input.Trace != nil {
		input.Trace.RuntimeLatencyMs = time.Since(runtimeStartedAt).Milliseconds()
		e.fillTraceFromSummary(input.Trace, summary, err)
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
