package services

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"cs-agent/internal/ai"
	"cs-agent/internal/ai/rag"
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/dto"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/pkg/errorsx"
	"cs-agent/internal/pkg/utils"
	"cs-agent/internal/repositories"

	"github.com/mlogclub/simple/common/strs"
	"github.com/mlogclub/simple/sqls"
)

var AIReplyService = newAIReplyService()

func newAIReplyService() *aiReplyService {
	return &aiReplyService{}
}

type aiReplyService struct{}

type aiKnowledgeCandidate struct {
	knowledgeBase *models.KnowledgeBase
	results       []rag.RetrieveResult
	topScore      float32
}

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

	question := s.buildQuestion(message, conversation)
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

	candidate, err := s.pickKnowledgeCandidate(ctx, aiAgent, question)
	if err != nil {
		return err
	}
	if candidate == nil || len(candidate.results) == 0 {
		return s.handleFallback(conversation, aiAgent, "知识库未命中")
	}

	aiConfig := AIConfigService.Get(aiAgent.AIConfigID)
	if aiConfig == nil || aiConfig.Status != enums.StatusOk {
		return errorsx.InvalidParam("AI Agent 关联的 AI 配置不可用")
	}

	replyText, err := s.generateReply(ctx, aiConfig, aiAgent, candidate, question)
	if err != nil {
		slog.Warn("ai reply generation failed, fallback instead",
			"conversation_id", conversation.ID,
			"message_id", message.ID,
			"ai_agent_id", aiAgent.ID,
			"knowledge_base_id", candidate.knowledgeBase.ID,
			"error", err)
		return s.handleFallback(conversation, aiAgent, "AI生成失败")
	}
	if strings.TrimSpace(replyText) == "" {
		return s.handleFallback(conversation, aiAgent, "AI回复为空")
	}

	if _, err := MessageService.SendAIMessage(nil, conversation.ID, aiAgent.ID, fmt.Sprintf("ai_%d", message.ID), enums.IMMessageTypeText, replyText, "", s.buildAIPrincipal(aiAgent)); err != nil {
		return err
	}
	return ConversationService.Updates(conversation.ID, map[string]any{
		"ai_reply_rounds": conversation.AIReplyRounds + 1,
	})
}

func (s *aiReplyService) buildQuestion(message *models.Message, conversation *models.Conversation) string {
	if message == nil {
		return ""
	}
	question := strings.TrimSpace(message.Content)
	if question != "" {
		return question
	}
	if conversation != nil {
		return strings.TrimSpace(conversation.LastMessageSummary)
	}
	return ""
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

func (s *aiReplyService) pickKnowledgeCandidate(ctx context.Context, aiAgent *models.AIAgent, question string) (*aiKnowledgeCandidate, error) {
	knowledgeIDs := utils.SplitInt64s(aiAgent.KnowledgeIDs)
	results, err := rag.Retrieve.Retrieve(ctx, rag.RetrieveRequest{
		KnowledgeBaseIDs: knowledgeIDs,
		Query:            question,
		TopK:             8,
		ScoreThreshold:   0.3,
	})
	if err != nil {
		slog.Warn("knowledge retrieve failed",
			"knowledge_base_ids", utils.JoinInt64s(knowledgeIDs),
			"conversation_ai_agent_id", aiAgent.ID,
			"error", err)
		return nil, nil
	}
	if len(results) == 0 {
		return nil, nil
	}

	var knowledgeBase *models.KnowledgeBase
	if results[0].KnowledgeBaseID > 0 {
		knowledgeBase = KnowledgeBaseService.Get(results[0].KnowledgeBaseID)
	}

	return &aiKnowledgeCandidate{
		knowledgeBase: knowledgeBase,
		results:       results,
		topScore:      results[0].Score,
	}, nil
}

func (s *aiReplyService) generateReply(ctx context.Context, aiConfig *models.AIConfig, aiAgent *models.AIAgent, candidate *aiKnowledgeCandidate, question string) (string, error) {
	if aiConfig == nil || aiAgent == nil || candidate == nil || candidate.knowledgeBase == nil {
		return "", nil
	}
	contextText := rag.Retrieve.BuildContext(ctx, candidate.results, 4000)
	if strings.TrimSpace(contextText) == "" {
		return "", nil
	}

	systemPrompt := s.buildSystemPrompt(aiAgent, candidate.knowledgeBase)
	userPrompt := fmt.Sprintf("用户问题：%s\n\n参考资料：\n%s", question, contextText)
	result, err := ai.LLM.ChatWithConfig(ctx, aiConfig, systemPrompt, userPrompt)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(result.Content), nil
}

func (s *aiReplyService) buildSystemPrompt(aiAgent *models.AIAgent, knowledgeBase *models.KnowledgeBase) string {
	if aiAgent != nil && strs.IsNotBlank(aiAgent.SystemPrompt) {
		return strings.TrimSpace(aiAgent.SystemPrompt)
	}

	basePrompt := "你是严格的客服知识库助手。只能依据提供的知识片段回答；如果资料不足，请明确说明知识库暂无明确信息。"
	if knowledgeBase != nil && enums.KnowledgeAnswerMode(knowledgeBase.AnswerMode) == enums.KnowledgeAnswerModeAssist {
		basePrompt = "你是客服知识库助手。请优先依据提供的知识片段回答，可以做轻度归纳，但不要编造未提供的事实。"
	}
	return basePrompt
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
