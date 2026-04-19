package runtime

import (
	"context"
	"fmt"
	"strings"

	applicationruntime "cs-agent/internal/ai/application/runtime"
	"cs-agent/internal/ai/runtime/graphs"
	svc "cs-agent/internal/services"
)

type replyInterruptService struct{}

func newReplyInterruptService() *replyInterruptService {
	return &replyInterruptService{}
}

func (s *replyInterruptService) ResumePendingInterrupt(ctx context.Context, owner *aiReplyService, replyCtx aiReplyContext) error {
	if replyCtx.PendingInterrupt == nil {
		return fmt.Errorf("pending interrupt is required")
	}
	summary, err := owner.executor.ResumePendingInterrupt(ctx, runtimeReplyResumeInput{
		Conversation:     replyCtx.Conversation,
		Message:          replyCtx.Message,
		AIAgent:          replyCtx.AIAgent,
		PendingInterrupt: replyCtx.PendingInterrupt,
		Trace:            replyCtx.Trace,
	})
	replyCtx.setSummary(summary)
	if err != nil {
		if isCheckpointMissingError(err) {
			summary = expiredInterruptSummary()
			replyCtx.setSummary(summary)
			replyCtx.Trace.Status = "interrupt_expired"
			replyCtx.Trace.FinalAction = "expired"
			replyMessage, expireErr := owner.commit.CommitAIReply(replyCommitInput{
				Conversation: replyCtx.Conversation,
				Message:      replyCtx.Message,
				AIAgent:      replyCtx.AIAgent,
				ReplyText:    summary.ReplyText,
				Trace:        replyCtx.Trace,
				ClientPrefix: "ai_interrupt_expired",
			})
			if expireErr != nil {
				return expireErr
			}
			lastResumeMessageID := int64(0)
			if replyMessage != nil {
				lastResumeMessageID = replyMessage.ID
			}
			if expireMarkErr := svc.ConversationInterruptService.MarkExpired(replyCtx.PendingInterrupt.ID, lastResumeMessageID); expireMarkErr != nil {
				return expireMarkErr
			}
			return nil
		}
		return err
	}
	if summary != nil && summary.Interrupted {
		return s.HandleInterruptedResume(owner, replyCtx, summary)
	}
	if summary != nil && strings.TrimSpace(summary.ReplyText) != "" {
		replyMessage, err := owner.commit.CommitAIReply(replyCommitInput{
			Conversation: replyCtx.Conversation,
			Message:      replyCtx.Message,
			AIAgent:      replyCtx.AIAgent,
			ReplyText:    summary.ReplyText,
			Trace:        replyCtx.Trace,
			ClientPrefix: "ai_resume",
		})
		if err != nil {
			return err
		}
		replyMessageID := int64(0)
		if replyMessage != nil {
			replyMessageID = replyMessage.ID
		}
		if graphs.IsCancellationReply(summary.ReplyText) {
			return svc.ConversationInterruptService.MarkCancelled(replyCtx.PendingInterrupt.ID, replyMessageID)
		}
		return svc.ConversationInterruptService.MarkResolved(replyCtx.PendingInterrupt.ID, replyMessageID)
	}
	return svc.ConversationInterruptService.MarkResolved(replyCtx.PendingInterrupt.ID, 0)
}

func (s *replyInterruptService) HandleInterruptedSummary(owner *aiReplyService, replyCtx aiReplyContext, summary *applicationruntime.Summary) error {
	pending := buildConversationInterrupt(replyCtx.Conversation, replyCtx.Message, replyCtx.AIAgent, summary)
	if err := svc.ConversationInterruptService.CreateOrUpdatePending(pending); err != nil {
		return err
	}
	pending = svc.ConversationInterruptService.GetByCheckPointID(summary.CheckPointID)
	replyText := resolveInterruptPrompt(summary)
	replyMessage, err := owner.commit.CommitAIReply(replyCommitInput{
		Conversation: replyCtx.Conversation,
		Message:      replyCtx.Message,
		AIAgent:      replyCtx.AIAgent,
		ReplyText:    replyText,
		Trace:        replyCtx.Trace,
		ClientPrefix: "ai_interrupt",
	})
	if err != nil {
		return err
	}
	if replyMessage != nil && pending != nil {
		return svc.ConversationInterruptService.MarkPendingAgain(pending.ID, pending.InterruptID, replyText, replyMessage.ID)
	}
	return nil
}

func (s *replyInterruptService) HandleInterruptedResume(owner *aiReplyService, replyCtx aiReplyContext, summary *applicationruntime.Summary) error {
	if replyCtx.PendingInterrupt == nil {
		return fmt.Errorf("pending interrupt is required")
	}
	replyText := resolveInterruptPrompt(summary)
	replyMessage, err := owner.commit.CommitAIReply(replyCommitInput{
		Conversation: replyCtx.Conversation,
		Message:      replyCtx.Message,
		AIAgent:      replyCtx.AIAgent,
		ReplyText:    replyText,
		Trace:        replyCtx.Trace,
		ClientPrefix: "ai_interrupt_resume",
	})
	if err != nil {
		return err
	}
	if replyMessage != nil {
		return svc.ConversationInterruptService.MarkPendingAgain(replyCtx.PendingInterrupt.ID, firstInterruptID(summary), replyText, replyMessage.ID)
	}
	return nil
}
