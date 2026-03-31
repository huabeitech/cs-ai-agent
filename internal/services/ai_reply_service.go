package services

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"cs-agent/internal/ai/agent"
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/dto"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/pkg/errorsx"
	"cs-agent/internal/repositories"

	"github.com/mlogclub/simple/sqls"
)

var AIReplyService = newAIReplyService()

func newAIReplyService() *aiReplyService {
	return &aiReplyService{}
}

type aiReplyService struct{}

const (
	defaultAIReplyAsyncTimeoutSeconds = 180
	maxAIReplyAsyncTimeoutSeconds     = 600
)

func (s *aiReplyService) TriggerReplyAsync(messageID int64) {
	if messageID <= 0 {
		return
	}
	go func() {
		startedAt := time.Now()
		timeout := s.resolveReplyTimeout(messageID)
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		if err := s.TriggerReply(ctx, messageID); err != nil {
			slog.Error("failed to trigger ai reply",
				"message_id", messageID,
				"timeout_ms", timeout.Milliseconds(),
				"elapsed_ms", time.Since(startedAt).Milliseconds(),
				"error", err)
		}
	}()
}

func (s *aiReplyService) resolveReplyTimeout(messageID int64) time.Duration {
	if messageID <= 0 {
		return time.Duration(defaultAIReplyAsyncTimeoutSeconds) * time.Second
	}
	message := MessageService.Get(messageID)
	if message == nil {
		return time.Duration(defaultAIReplyAsyncTimeoutSeconds) * time.Second
	}
	conversation := ConversationService.Get(message.ConversationID)
	if conversation == nil || conversation.AIAgentID <= 0 {
		return time.Duration(defaultAIReplyAsyncTimeoutSeconds) * time.Second
	}
	aiAgent := AIAgentService.Get(conversation.AIAgentID)
	return s.buildReplyTimeout(aiAgent)
}

func (s *aiReplyService) buildReplyTimeout(aiAgent *models.AIAgent) time.Duration {
	timeoutSeconds := defaultAIReplyAsyncTimeoutSeconds
	if aiAgent != nil && aiAgent.ReplyTimeoutSeconds > 0 {
		timeoutSeconds = aiAgent.ReplyTimeoutSeconds
	}
	if timeoutSeconds > maxAIReplyAsyncTimeoutSeconds {
		timeoutSeconds = maxAIReplyAsyncTimeoutSeconds
	}
	return time.Duration(timeoutSeconds) * time.Second
}

func (s *aiReplyService) TriggerReply(ctx context.Context, messageID int64) error {
	message := MessageService.Get(messageID)
	if message == nil {
		return errorsx.InvalidParam("消息不存在")
	}
	if message.SenderType != enums.IMSenderTypeCustomer {
		return nil
	}

	conversation := ConversationService.Get(message.ConversationID)
	if conversation == nil {
		return errorsx.InvalidParam("会话不存在")
	}
	if conversation.LastMessageID != message.ID {
		return nil
	}
	if conversation.AIAgentID <= 0 {
		return nil
	}
	if conversation.HandoffAt != nil || conversation.CurrentAssigneeID > 0 {
		return nil
	}

	aiAgent := AIAgentService.Get(conversation.AIAgentID)
	if aiAgent == nil || aiAgent.Status != enums.StatusOk {
		return nil
	}
	if aiAgent.ServiceMode == enums.IMConversationServiceModeHumanOnly {
		return nil
	}

	question := strings.TrimSpace(message.Content)
	if question == "" {
		question = strings.TrimSpace(conversation.LastMessageSummary)
	}
	if question == "" {
		return nil
	}
	if s.shouldHandoffByQuestion(question, aiAgent) {
		return s.handoffConversation(conversation, aiAgent, "用户主动要求人工")
	}
	if aiAgent.ServiceMode != enums.IMConversationServiceModeAIOnly &&
		aiAgent.MaxAIReplyRounds > 0 &&
		conversation.AIReplyRounds >= aiAgent.MaxAIReplyRounds {
		return s.handoffConversation(conversation, aiAgent, "达到AI最大回复轮次")
	}

	aiConfig := AIConfigService.Get(aiAgent.AIConfigID)
	if aiConfig == nil || aiConfig.Status != enums.StatusOk {
		return errorsx.InvalidParam("AI Agent 关联的 AI 配置不可用")
	}
	runtime := agent.NewRuntime()
	result, err := runtime.RunConversationTurn(ctx, agent.TurnContext{
		Message:      message,
		Conversation: conversation,
		AIAgent:      aiAgent,
		AIConfig:     aiConfig,
	})
	if err != nil {
		slog.Warn("ai reply generation failed, fallback instead",
			"conversation_id", conversation.ID,
			"message_id", message.ID,
			"ai_agent_id", aiAgent.ID,
			"error", err)
		if result != nil && strings.TrimSpace(result.Reason) != "" {
			return s.handleFallback(conversation, aiAgent, result.Reason)
		}
		return s.handleFallback(conversation, aiAgent, "AI生成失败")
	}
	if result == nil || result.Action == agent.ActionNoop {
		return nil
	}
	switch result.Action {
	case agent.ActionReply:
		if strings.TrimSpace(result.ReplyText) == "" {
			return s.handleFallback(conversation, aiAgent, "AI回复为空")
		}
		if _, err := MessageService.SendAIMessage(nil, conversation.ID, aiAgent.ID, fmt.Sprintf("ai_%d", message.ID), enums.IMMessageTypeText, result.ReplyText, "", s.buildAIPrincipal(aiAgent)); err != nil {
			return err
		}
		return ConversationService.Updates(conversation.ID, map[string]any{
			"ai_reply_rounds": conversation.AIReplyRounds + 1,
		})
	case agent.ActionFallback:
		return s.handleFallback(conversation, aiAgent, result.Reason)
	default:
		return nil
	}
}

func (s *aiReplyService) buildAIPrincipal(aiAgent *models.AIAgent) *dto.AuthPrincipal {
	username := "AI"
	if aiAgent != nil && strings.TrimSpace(aiAgent.Name) != "" {
		username = aiAgent.Name
	}
	return &dto.AuthPrincipal{
		UserID:   0,
		Username: username,
		Nickname: username,
	}
}

func (s *aiReplyService) shouldHandoffByQuestion(question string, aiAgent *models.AIAgent) bool {
	if aiAgent == nil || aiAgent.ServiceMode == enums.IMConversationServiceModeAIOnly {
		return false
	}
	normalized := strings.ReplaceAll(strings.ToLower(strings.TrimSpace(question)), " ", "")
	if normalized == "" {
		return false
	}
	keywords := []string{"人工", "转人工", "人工客服", "真人", "客服"}
	for _, keyword := range keywords {
		if strings.Contains(normalized, keyword) {
			return true
		}
	}
	return false
}

func (s *aiReplyService) handleFallback(conversation *models.Conversation, aiAgent *models.AIAgent, reason string) error {
	if conversation == nil || aiAgent == nil {
		return nil
	}
	switch enums.AIAgentFallbackMode(aiAgent.FallbackMode) {
	case enums.AIAgentFallbackModeHandoff: // 转人工
		return s.handoffConversation(conversation, aiAgent, reason)
	case enums.AIAgentFallbackModeGuideRephrase: // 引导补充信息或换个问法
		_, err := MessageService.SendAIMessage(nil, conversation.ID, aiAgent.ID, fmt.Sprintf("ai_fallback_%d", conversation.LastMessageID), enums.IMMessageTypeText, s.buildFallbackMessage(aiAgent), "", s.buildAIPrincipal(aiAgent))
		return err
	default: // 直接声明无答案
		_, err := MessageService.SendAIMessage(nil, conversation.ID, aiAgent.ID, fmt.Sprintf("ai_fallback_%d", conversation.LastMessageID), enums.IMMessageTypeText, s.buildFallbackMessage(aiAgent), "", s.buildAIPrincipal(aiAgent))
		return err
	}
}

func (s *aiReplyService) buildFallbackMessage(aiAgent *models.AIAgent) string {
	if aiAgent != nil && strings.TrimSpace(aiAgent.FallbackMessage) != "" {
		return strings.TrimSpace(aiAgent.FallbackMessage)
	}
	if aiAgent != nil && enums.AIAgentFallbackMode(aiAgent.FallbackMode) == enums.AIAgentFallbackModeGuideRephrase {
		return "我暂时没有找到足够准确的信息。你可以补充订单号、产品名或更具体的问题，我再继续帮你查。"
	}
	return "我暂时没有找到明确答案。"
}

func (s *aiReplyService) handoffConversation(conversation *models.Conversation, aiAgent *models.AIAgent, reason string) error {
	if conversation == nil || aiAgent == nil {
		return nil
	}
	now := time.Now()
	if err := sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		if err := repositories.ConversationRepository.Updates(ctx.Tx, conversation.ID, map[string]any{
			"handoff_at":          now,
			"handoff_reason":      strings.TrimSpace(reason),
			"status":              enums.IMConversationStatusPending,
			"current_team_id":     0,
			"current_assignee_id": 0,
			"update_user_id":      0,
			"update_user_name":    aiAgent.Name,
			"updated_at":          now,
		}); err != nil {
			return err
		}
		return ConversationEventLogService.CreateEvent(ctx, conversation.ID, enums.IMEventTypeTransfer, enums.IMSenderTypeAI, aiAgent.ID, "AI转人工", strings.TrimSpace(reason))
	}); err != nil {
		return err
	}
	if _, err := MessageService.SendAIMessage(nil, conversation.ID, aiAgent.ID, fmt.Sprintf("ai_handoff_%d", conversation.LastMessageID), enums.IMMessageTypeText, "已为你转接人工客服，请稍候。", "", s.buildAIPrincipal(aiAgent)); err != nil {
		return err
	}
	if _, err := ConversationDispatchService.DispatchConversation(conversation.ID); err != nil {
		slog.Warn("auto dispatch conversation after ai handoff failed",
			"conversation_id", conversation.ID,
			"ai_agent_id", aiAgent.ID,
			"error", err)
	}
	return nil
}
