package runtime

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"cs-agent/internal/models"
	"cs-agent/internal/pkg/enums"
	svc "cs-agent/internal/services"
)

func (s *aiReplyService) resolveReplyTimeout(aiAgent models.AIAgent) time.Duration {
	if aiAgent.ReplyTimeoutSeconds <= 0 {
		return time.Duration(defaultAIReplyAsyncTimeoutSeconds) * time.Second
	}
	if aiAgent.ReplyTimeoutSeconds > maxAIReplyAsyncTimeoutSeconds {
		return time.Duration(maxAIReplyAsyncTimeoutSeconds) * time.Second
	}
	return time.Duration(aiAgent.ReplyTimeoutSeconds) * time.Second
}

func (s *aiReplyService) TriggerReplyAsync(conversation models.Conversation, message models.Message) {
	go func() {
		aiAgent := svc.AIAgentService.Get(conversation.AIAgentID)
		if aiAgent == nil || aiAgent.Status != enums.StatusOk {
			return
		}
		startedAt := time.Now()
		timeout := s.resolveReplyTimeout(*aiAgent)
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		if err := s.TriggerReply(ctx, conversation, message, *aiAgent); err != nil {
			slog.Error("failed to trigger ai reply",
				"message_id", message.ID,
				"timeout_ms", timeout.Milliseconds(),
				"elapsed_ms", time.Since(startedAt).Milliseconds(),
				"error", err)
		}
	}()
}

func (s *aiReplyService) TriggerReply(ctx context.Context, conversation models.Conversation, message models.Message, aiAgent models.AIAgent) (retErr error) {
	startedAt := time.Now()
	trace := &aiReplyTraceData{Status: "started"}
	var summary *Summary
	if err := ctx.Err(); err != nil {
		return err
	}
	if s.eligibility != nil && !s.eligibility.CanReply(conversation, message, aiAgent) {
		return nil
	}
	defer func() {
		s.runlog.Write(startedAt, message, conversation, aiAgent, message.Content, retErr, trace, summary)
	}()
	if pendingInterrupt := svc.ConversationInterruptService.FindLatestPendingByConversationID(conversation.ID); pendingInterrupt != nil {
		return s.interrupts.ResumePendingInterrupt(ctx, s, conversation, message, aiAgent, pendingInterrupt, trace, &summary)
	}
	var err error
	summary, err = s.executor.Run(ctx, conversation, message, aiAgent, trace)
	if err != nil {
		return err
	}
	if summary != nil && summary.Interrupted {
		return s.interrupts.HandleInterruptedSummary(s, conversation, message, aiAgent, summary, trace)
	}
	if summary != nil && strings.TrimSpace(summary.ReplyText) != "" {
		replyMessage, err := s.commit.SendAIReply(conversation, message, aiAgent, summary.ReplyText, trace, "ai_reply")
		if err != nil {
			return err
		}
		if err := s.commit.IncrementAIReplyRounds(conversation.ID, conversation.AIReplyRounds+1, aiAgent.Name); err != nil {
			return err
		}
		trace.ReplySent = replyMessage != nil
	}
	return nil
}
