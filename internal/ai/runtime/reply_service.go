package runtime

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"cs-agent/internal/models"
	"cs-agent/internal/pkg/dto"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/pkg/toolx"
	"cs-agent/internal/repositories"
	svc "cs-agent/internal/services"

	"github.com/mlogclub/simple/common/strs"
	"github.com/mlogclub/simple/sqls"
)

var AIReplyService = newAIReplyService()

func init() {
	svc.TriggerAIReplyAsyncHook = AIReplyService.TriggerReplyAsync
}

func newAIReplyService() *aiReplyService {
	return &aiReplyService{}
}

type aiReplyService struct{}

type aiReplyTraceData struct {
	Status           string          `json:"status"`
	RuntimeLatencyMs int64           `json:"runtimeLatencyMs,omitempty"`
	RecheckMs        int64           `json:"recheckMs,omitempty"`
	CommitMs         int64           `json:"commitMs,omitempty"`
	FinalAction      string          `json:"finalAction,omitempty"`
	ResumeSource     string          `json:"resumeSource,omitempty"`
	ReplySent        bool            `json:"replySent,omitempty"`
	ReplyMessageID   int64           `json:"replyMessageId,omitempty"`
	Runtime          json.RawMessage `json:"runtime,omitempty"`
}

const (
	defaultAIReplyAsyncTimeoutSeconds = 180
	maxAIReplyAsyncTimeoutSeconds     = 600
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
	if message.SenderType != enums.IMSenderTypeCustomer {
		return nil
	}
	if conversation.HandoffAt != nil || conversation.CurrentAssigneeID > 0 {
		return nil
	}
	if aiAgent.ServiceMode == enums.IMConversationServiceModeHumanOnly {
		return nil
	}
	if strs.IsBlank(message.Content) {
		return nil
	}
	defer func() {
		s.writeRunLog(startedAt, message, conversation, aiAgent, message.Content, retErr, trace, summary)
	}()
	if pendingInterrupt := svc.ConversationInterruptService.FindLatestPendingByConversationID(conversation.ID); pendingInterrupt != nil {
		return s.resumePendingInterrupt(ctx, conversation, message, aiAgent, pendingInterrupt, trace, &summary)
	}
	if s.shouldHandoffByQuestion(message.Content, aiAgent) {
		return s.handoffConversation(conversation, aiAgent, "用户主动要求人工")
	}
	if aiAgent.ServiceMode != enums.IMConversationServiceModeAIOnly &&
		aiAgent.MaxAIReplyRounds > 0 &&
		conversation.AIReplyRounds >= aiAgent.MaxAIReplyRounds {
		return s.handoffConversation(conversation, aiAgent, "达到AI最大回复轮次")
	}
	aiConfig := svc.AIConfigService.Get(aiAgent.AIConfigID)
	if aiConfig == nil {
		return fmt.Errorf("ai config is nil")
	}

	runtimeStartedAt := time.Now()
	var err error
	summary, err = Service.Run(ctx, Request{
		Conversation: &conversation,
		UserMessage:  &message,
		AIAgent:      &aiAgent,
		AIConfig:     aiConfig,
	})
	trace.RuntimeLatencyMs = time.Since(runtimeStartedAt).Milliseconds()
	if err != nil {
		trace.Status = "runtime_error"
		trace.FinalAction = "error"
		if summary != nil {
			trace.Runtime = json.RawMessage(summary.TraceData)
		}
		return err
	}
	trace.Status = "runtime_prepared"
	trace.FinalAction = toRunLogFinalAction(summary)
	if summary != nil && strings.TrimSpace(summary.TraceData) != "" {
		trace.Runtime = json.RawMessage(summary.TraceData)
	}
	if summary != nil && summary.Interrupted {
		return s.handleInterruptedSummary(conversation, message, aiAgent, summary, trace)
	}
	if summary != nil && strings.TrimSpace(summary.ReplyText) != "" {
		replyMessage, err := s.sendAIReply(conversation, message, aiAgent, summary.ReplyText, trace, "ai_reply")
		if err != nil {
			return err
		}
		if err := s.incrementAIReplyRounds(conversation.ID, conversation.AIReplyRounds+1, aiAgent.Name); err != nil {
			return err
		}
		trace.ReplySent = replyMessage != nil
	}
	return nil
}

func (s *aiReplyService) resumePendingInterrupt(ctx context.Context, conversation models.Conversation, message models.Message, aiAgent models.AIAgent,
	pendingInterrupt *models.ConversationInterrupt, trace *aiReplyTraceData, summaryRef **Summary) error {
	if pendingInterrupt == nil {
		return nil
	}
	aiConfig := svc.AIConfigService.Get(aiAgent.AIConfigID)
	if aiConfig == nil {
		return fmt.Errorf("ai config is nil")
	}
	runtimeStartedAt := time.Now()
	trace.ResumeSource = "pending_interrupt"
	summary, err := Service.Resume(ctx, ResumeRequest{
		Conversation: &conversation,
		AIAgent:      &aiAgent,
		AIConfig:     aiConfig,
		CheckPointID: strings.TrimSpace(pendingInterrupt.CheckPointID),
		ResumeData: map[string]any{
			strings.TrimSpace(pendingInterrupt.InterruptID): strings.TrimSpace(message.Content),
		},
	})
	trace.RuntimeLatencyMs = time.Since(runtimeStartedAt).Milliseconds()
	*summaryRef = summary
	if err != nil {
		if isCheckpointMissingError(err) {
			summary = &Summary{
				Status:    "expired",
				ReplyText: "本次确认已失效，请重新发起。",
			}
			*summaryRef = summary
			trace.Status = "interrupt_expired"
			trace.FinalAction = "expired"
			replyMessage, expireErr := s.sendAIReply(conversation, message, aiAgent, summary.ReplyText, trace, "ai_interrupt_expired")
			if expireErr != nil {
				return expireErr
			}
			if err := s.incrementAIReplyRounds(conversation.ID, conversation.AIReplyRounds+1, aiAgent.Name); err != nil {
				return err
			}
			lastResumeMessageID := int64(0)
			if replyMessage != nil {
				lastResumeMessageID = replyMessage.ID
			}
			if expireMarkErr := svc.ConversationInterruptService.MarkExpired(pendingInterrupt.ID, lastResumeMessageID); expireMarkErr != nil {
				return expireMarkErr
			}
			return nil
		}
		trace.Status = "runtime_error"
		trace.FinalAction = "error"
		if summary != nil {
			trace.Runtime = json.RawMessage(summary.TraceData)
		}
		return err
	}
	trace.Status = "runtime_prepared"
	trace.FinalAction = toRunLogFinalAction(summary)
	if summary != nil && strings.TrimSpace(summary.TraceData) != "" {
		trace.Runtime = json.RawMessage(summary.TraceData)
	}
	if summary != nil && summary.Interrupted {
		return s.handleInterruptedResume(conversation, message, aiAgent, pendingInterrupt, summary, trace)
	}
	if summary != nil && strings.TrimSpace(summary.ReplyText) != "" {
		replyMessage, err := s.sendAIReply(conversation, message, aiAgent, summary.ReplyText, trace, "ai_resume")
		if err != nil {
			return err
		}
		if err := s.incrementAIReplyRounds(conversation.ID, conversation.AIReplyRounds+1, aiAgent.Name); err != nil {
			return err
		}
		replyMessageID := int64(0)
		if replyMessage != nil {
			replyMessageID = replyMessage.ID
		}
		if isCancellationReply(summary.ReplyText) {
			return svc.ConversationInterruptService.MarkCancelled(pendingInterrupt.ID, replyMessageID)
		}
		return svc.ConversationInterruptService.MarkResolved(pendingInterrupt.ID, replyMessageID)
	}
	return svc.ConversationInterruptService.MarkResolved(pendingInterrupt.ID, 0)
}

func (s *aiReplyService) handleInterruptedSummary(conversation models.Conversation, message models.Message, aiAgent models.AIAgent,
	summary *Summary, trace *aiReplyTraceData) error {
	pending := buildConversationInterrupt(conversation, message, aiAgent, summary)
	if err := svc.ConversationInterruptService.CreateOrUpdatePending(pending); err != nil {
		return err
	}
	pending = svc.ConversationInterruptService.GetByCheckPointID(summary.CheckPointID)
	replyText := resolveInterruptPrompt(summary)
	replyMessage, err := s.sendAIReply(conversation, message, aiAgent, replyText, trace, "ai_interrupt")
	if err != nil {
		return err
	}
	if err := s.incrementAIReplyRounds(conversation.ID, conversation.AIReplyRounds+1, aiAgent.Name); err != nil {
		return err
	}
	if replyMessage != nil && pending != nil {
		return svc.ConversationInterruptService.MarkPendingAgain(pending.ID, pending.InterruptID, replyText, replyMessage.ID)
	}
	return nil
}

func (s *aiReplyService) handleInterruptedResume(conversation models.Conversation, message models.Message, aiAgent models.AIAgent,
	pendingInterrupt *models.ConversationInterrupt, summary *Summary, trace *aiReplyTraceData) error {
	if pendingInterrupt == nil {
		return nil
	}
	replyText := resolveInterruptPrompt(summary)
	replyMessage, err := s.sendAIReply(conversation, message, aiAgent, replyText, trace, "ai_interrupt_resume")
	if err != nil {
		return err
	}
	if err := s.incrementAIReplyRounds(conversation.ID, conversation.AIReplyRounds+1, aiAgent.Name); err != nil {
		return err
	}
	if replyMessage != nil {
		return svc.ConversationInterruptService.MarkPendingAgain(pendingInterrupt.ID, firstInterruptID(summary), replyText, replyMessage.ID)
	}
	return nil
}

func (s *aiReplyService) sendAIReply(conversation models.Conversation, message models.Message, aiAgent models.AIAgent,
	replyText string, trace *aiReplyTraceData, clientPrefix string) (*models.Message, error) {
	replyText = strings.TrimSpace(replyText)
	if replyText == "" {
		return nil, nil
	}
	commitStartedAt := time.Now()
	replyMessage, err := svc.MessageService.SendAIMessage(conversation.ID, aiAgent.ID,
		fmt.Sprintf("%s_%d", strings.TrimSpace(clientPrefix), message.ID), enums.IMMessageTypeText, replyText, "", s.buildAIPrincipal(aiAgent))
	if trace != nil {
		trace.CommitMs = time.Since(commitStartedAt).Milliseconds()
		trace.ReplySent = err == nil && replyMessage != nil
		if replyMessage != nil {
			trace.ReplyMessageID = replyMessage.ID
		}
	}
	return replyMessage, err
}

func (s *aiReplyService) shouldHandoffByQuestion(question string, aiAgent models.AIAgent) bool {
	if aiAgent.ServiceMode == enums.IMConversationServiceModeAIOnly {
		return false
	}
	normalized := strings.ReplaceAll(strings.ToLower(strings.TrimSpace(question)), " ", "")
	if normalized == "" {
		return false
	}
	keywords := []string{"转人工", "人工客服"}
	for _, keyword := range keywords {
		if strings.Contains(normalized, keyword) {
			return true
		}
	}
	return false
}

func (s *aiReplyService) writeRunLog(startedAt time.Time, message models.Message, conversation models.Conversation, aiAgent models.AIAgent,
	question string, runErr error, trace *aiReplyTraceData, summary *Summary) {
	errorMessage := ""
	if runErr != nil {
		errorMessage = runErr.Error()
	} else if summary != nil && strings.TrimSpace(summary.ErrorMessage) != "" {
		errorMessage = strings.TrimSpace(summary.ErrorMessage)
	}
	traceData := buildAIReplyTraceData(trace)
	plannedAction, plannedToolCode, planReason := buildRunLogPlan(summary)
	logItem := &models.AgentRunLog{
		ConversationID:   conversation.ID,
		MessageID:        message.ID,
		AIAgentID:        aiAgent.ID,
		AIConfigID:       aiAgent.AIConfigID,
		UserMessage:      strings.TrimSpace(question),
		PlannedAction:    plannedAction,
		PlannedSkillCode: strings.TrimSpace(summaryPlannedSkillCode(summary)),
		PlannedSkillName: strings.TrimSpace(summaryPlannedSkillName(summary)),
		SkillRouteTrace:  strings.TrimSpace(summarySkillRouteTrace(summary)),
		ToolSearchTrace:  extractToolSearchTrace(summary),
		PlannedToolCode:  plannedToolCode,
		PlanReason:       planReason,
		InterruptType:    firstInterruptType(summary),
		ResumeSource:     runLogResumeSource(trace),
		FinalAction:      toRunLogFinalAction(summary),
		FinalStatus:      runLogFinalStatus(summary),
		ReplyText:        buildRunLogReplyText(summary),
		ErrorMessage:     errorMessage,
		LatencyMs:        time.Since(startedAt).Milliseconds(),
		TraceData:        traceData,
		CreatedAt:        time.Now(),
	}
	if err := svc.AgentRunLogService.Create(logItem); err != nil {
		slog.Warn("create agent run log failed",
			"message_id", message.ID,
			"conversation_id", logItem.ConversationID,
			"ai_agent_id", aiAgent.ID,
			"error", err)
	}
}

func buildAIReplyTraceData(trace *aiReplyTraceData) string {
	if trace == nil {
		return ""
	}
	data, err := json.Marshal(trace)
	if err != nil {
		return ""
	}
	return string(data)
}

func buildRunLogPlan(summary *Summary) (plannedAction, plannedToolCode, planReason string) {
	if summary == nil {
		return "", "", ""
	}
	if skillCode := strings.TrimSpace(summaryPlannedSkillCode(summary)); skillCode != "" {
		reason := strings.TrimSpace(summary.PlanReason)
		if reason == "" {
			reason = "skill_selected"
		}
		return "skill", "", reason
	}
	if strings.TrimSpace(summary.Status) == "expired" {
		return "interrupt", "", "pending interrupt checkpoint expired"
	}
	if summary.Interrupted {
		return "tool", summaryPrimaryToolCode(summary), "agent interrupted and is waiting for user confirmation"
	}
	if len(summary.InvokedToolCodes) > 0 {
		toolCode := summaryPrimaryToolCode(summary)
		reason := "agent invoked MCP tool"
		if toolCode != "" && toolCode != firstInvokedToolCode(summary) {
			reason = "agent invoked dynamic tool via tool_search"
		}
		return "tool", toolCode, reason
	}
	if strings.TrimSpace(summary.ReplyText) != "" {
		return "reply", "", "agent replied directly"
	}
	if strings.TrimSpace(summary.ErrorMessage) != "" {
		return "error", "", "runtime execution failed"
	}
	return "fallback", "", "runtime produced empty reply"
}

func toRunLogFinalAction(summary *Summary) string {
	if summary == nil {
		return ""
	}
	if skillCode := strings.TrimSpace(summaryPlannedSkillCode(summary)); skillCode != "" && strings.TrimSpace(summary.ReplyText) != "" {
		return "skill"
	}
	switch strings.TrimSpace(summary.Status) {
	case "completed":
		return "reply"
	case "fallback":
		return "fallback"
	case "error":
		return "error"
	case "interrupted":
		return "interrupted"
	case "expired":
		return "expired"
	default:
		return strings.TrimSpace(summary.Status)
	}
}

func buildRunLogReplyText(summary *Summary) string {
	if summary == nil {
		return ""
	}
	return strings.TrimSpace(summary.ReplyText)
}

func summaryPlannedSkillCode(summary *Summary) string {
	if summary == nil {
		return ""
	}
	return strings.TrimSpace(summary.PlannedSkillCode)
}

func summaryPlannedSkillName(summary *Summary) string {
	if summary == nil {
		return ""
	}
	return strings.TrimSpace(summary.PlannedSkillName)
}

func summarySkillRouteTrace(summary *Summary) string {
	if summary == nil {
		return ""
	}
	return strings.TrimSpace(summary.SkillRouteTrace)
}

func runLogResumeSource(trace *aiReplyTraceData) string {
	if trace == nil {
		return ""
	}
	return strings.TrimSpace(trace.ResumeSource)
}

func runLogFinalStatus(summary *Summary) string {
	if summary == nil {
		return ""
	}
	return strings.TrimSpace(summary.Status)
}

func (s *aiReplyService) incrementAIReplyRounds(conversationID int64, nextRounds int, aiAgentName string) error {
	return repositories.ConversationRepository.Updates(sqls.DB(), conversationID, map[string]any{
		"ai_reply_rounds":  nextRounds,
		"update_user_id":   0,
		"update_user_name": strings.TrimSpace(aiAgentName),
		"updated_at":       time.Now(),
	})
}

func buildConversationInterrupt(conversation models.Conversation, message models.Message, aiAgent models.AIAgent, summary *Summary) *models.ConversationInterrupt {
	if summary == nil {
		return nil
	}
	now := time.Now()
	item := svc.ConversationInterruptService.GetByCheckPointID(summary.CheckPointID)
	if item == nil {
		item = &models.ConversationInterrupt{
			CheckPointID: summary.CheckPointID,
			CreatedAt:    now,
		}
	}
	item.ConversationID = conversation.ID
	item.AIAgentID = aiAgent.ID
	item.SourceMessageID = message.ID
	item.InterruptID = firstInterruptID(summary)
	item.InterruptType = firstInterruptType(summary)
	item.Status = "pending"
	item.PromptText = resolveInterruptPrompt(summary)
	item.UpdatedAt = now
	return item
}

func resolveInterruptPrompt(summary *Summary) string {
	if summary == nil || len(summary.Interrupts) == 0 {
		return "请继续补充信息后再试。"
	}
	if prompt := extractInterruptMessage(summary.Interrupts[0].InfoPreview); prompt != "" {
		return prompt
	}
	if prompt := strings.TrimSpace(summary.Interrupts[0].InfoPreview); prompt != "" {
		return prompt
	}
	return "请继续补充信息后再试。"
}

func extractInterruptMessage(infoPreview string) string {
	infoPreview = strings.TrimSpace(infoPreview)
	if infoPreview == "" {
		return ""
	}
	payload := make(map[string]any)
	if err := json.Unmarshal([]byte(infoPreview), &payload); err != nil {
		return ""
	}
	if message, ok := payload["message"].(string); ok {
		return strings.TrimSpace(message)
	}
	return ""
}

func firstInterruptID(summary *Summary) string {
	if summary == nil || len(summary.Interrupts) == 0 {
		return ""
	}
	return strings.TrimSpace(summary.Interrupts[0].ID)
}

func firstInterruptType(summary *Summary) string {
	if summary == nil || len(summary.Interrupts) == 0 {
		return ""
	}
	return strings.TrimSpace(summary.Interrupts[0].Type)
}

func firstInvokedToolCode(summary *Summary) string {
	if summary == nil {
		return ""
	}
	if len(summary.InvokedToolCodes) > 0 {
		return strings.TrimSpace(summary.InvokedToolCodes[0])
	}
	return ""
}

func summaryPrimaryToolCode(summary *Summary) string {
	if summary == nil {
		return ""
	}
	toolCode := firstInvokedToolCode(summary)
	if toolCode != toolx.BuiltinToolSearchToolCode {
		return toolCode
	}
	if targetToolCode := firstToolSearchTargetToolCode(summary); targetToolCode != "" {
		return targetToolCode
	}
	return toolCode
}

func extractToolSearchTrace(summary *Summary) string {
	if summary == nil {
		return ""
	}
	trace := parseRuntimeTraceData(summary.TraceData)
	if len(trace.ToolSearch.Items) == 0 || len(trace.ToolSearch.Raw) == 0 {
		return ""
	}
	return string(trace.ToolSearch.Raw)
}

func firstToolSearchTargetToolCode(summary *Summary) string {
	trace := parseRuntimeTraceData(summary.TraceData)
	for _, item := range trace.ToolSearch.Items {
		toolCode := strings.TrimSpace(item.TargetToolCode)
		if toolCode != "" {
			return toolCode
		}
		if len(item.CandidateToolCodes) == 1 {
			toolCode = strings.TrimSpace(item.CandidateToolCodes[0])
			if toolCode != "" {
				return toolCode
			}
		}
	}
	return ""
}

type runtimeTraceProjection struct {
	ToolSearch struct {
		Raw   json.RawMessage `json:"-"`
		Items []struct {
			TargetToolCode     string   `json:"targetToolCode"`
			CandidateToolCodes []string `json:"candidateToolCodes"`
		} `json:"items"`
	} `json:"toolSearch"`
}

func parseRuntimeTraceData(raw string) runtimeTraceProjection {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return runtimeTraceProjection{}
	}
	var payload map[string]json.RawMessage
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return runtimeTraceProjection{}
	}
	var trace runtimeTraceProjection
	if toolSearchRaw, ok := payload["toolSearch"]; ok && len(toolSearchRaw) > 0 {
		trace.ToolSearch.Raw = append(json.RawMessage(nil), toolSearchRaw...)
		_ = json.Unmarshal(toolSearchRaw, &trace.ToolSearch)
	}
	return trace
}

func isCancellationReply(replyText string) bool {
	replyText = strings.TrimSpace(replyText)
	return strings.Contains(replyText, "已取消本次工单创建")
}

func isCheckpointMissingError(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(strings.TrimSpace(err.Error()))
	return strings.Contains(message, "failed to load from checkpoint") && strings.Contains(message, "not exist")
}

func (s *aiReplyService) handoffConversation(conversation models.Conversation, aiAgent models.AIAgent, reason string) error {
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
		return svc.ConversationEventLogService.CreateEvent(ctx, conversation.ID, enums.IMEventTypeTransfer, enums.IMSenderTypeAI, aiAgent.ID, "AI转人工", strings.TrimSpace(reason))
	}); err != nil {
		return err
	}
	if _, err := svc.MessageService.SendAIMessage(conversation.ID, aiAgent.ID, fmt.Sprintf("ai_handoff_%d", conversation.LastMessageID), enums.IMMessageTypeText, "已为你转接人工客服，请稍候。", "", s.buildAIPrincipal(aiAgent)); err != nil {
		return err
	}
	if _, err := svc.ConversationDispatchService.DispatchConversation(conversation.ID); err != nil {
		slog.Warn("auto dispatch conversation after ai handoff failed",
			"conversation_id", conversation.ID,
			"ai_agent_id", aiAgent.ID,
			"error", err)
	}
	return nil
}

func (s *aiReplyService) buildAIPrincipal(aiAgent models.AIAgent) *dto.AuthPrincipal {
	username := "AI"
	if strings.TrimSpace(aiAgent.Name) != "" {
		username = aiAgent.Name
	}
	return &dto.AuthPrincipal{
		UserID:   0,
		Username: username,
		Nickname: username,
	}
}
