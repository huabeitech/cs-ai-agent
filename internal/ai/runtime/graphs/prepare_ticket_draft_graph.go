package graphs

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"cs-agent/internal/models"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/services"
)

type PrepareTicketDraftInput struct {
	Title           string `json:"title"`
	Description     string `json:"description"`
	Issue           string `json:"issue"`
	Impact          string `json:"impact"`
	ExpectedOutcome string `json:"expectedOutcome"`
	CurrentAttempt  string `json:"currentAttempt"`
	Priority        int64  `json:"priority"`
	Severity        int    `json:"severity"`
}

type PrepareTicketDraftResult struct {
	Ready             bool     `json:"ready"`
	Title             string   `json:"title"`
	Description       string   `json:"description"`
	Priority          int64    `json:"priority,omitempty"`
	Severity          int      `json:"severity,omitempty"`
	MissingFields     []string `json:"missingFields,omitempty"`
	FollowUpQuestions []string `json:"followUpQuestions,omitempty"`
	ConversationFacts []string `json:"conversationFacts,omitempty"`
}

type PrepareTicketDraftGraph struct {
	conversation models.Conversation
}

func NewPrepareTicketDraftGraph(conversation models.Conversation) *PrepareTicketDraftGraph {
	return &PrepareTicketDraftGraph{conversation: conversation}
}

func (g *PrepareTicketDraftGraph) Run(_ context.Context, argumentsInJSON string) (string, error) {
	input, err := g.parseInput(argumentsInJSON)
	if err != nil {
		return "", err
	}
	messages, _, _ := services.MessageService.FindByConversationIDCursor(g.conversation.ID, 0, 6, "", "")
	result := buildPrepareTicketDraftResult(g.conversation, messages, input)
	buf, err := json.Marshal(result)
	if err != nil {
		return "", err
	}
	return string(buf), nil
}

func (g *PrepareTicketDraftGraph) parseInput(argumentsInJSON string) (PrepareTicketDraftInput, error) {
	var input PrepareTicketDraftInput
	if strings.TrimSpace(argumentsInJSON) == "" {
		return input, nil
	}
	if err := json.Unmarshal([]byte(argumentsInJSON), &input); err != nil {
		return input, fmt.Errorf("invalid prepare ticket draft arguments: %w", err)
	}
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)
	input.Issue = strings.TrimSpace(input.Issue)
	input.Impact = strings.TrimSpace(input.Impact)
	input.ExpectedOutcome = strings.TrimSpace(input.ExpectedOutcome)
	input.CurrentAttempt = strings.TrimSpace(input.CurrentAttempt)
	return input, nil
}

func buildPrepareTicketDraftResult(conversation models.Conversation, messages []models.Message, input PrepareTicketDraftInput) PrepareTicketDraftResult {
	result := PrepareTicketDraftResult{
		Priority:          input.Priority,
		Severity:          input.Severity,
		MissingFields:     make([]string, 0, 2),
		FollowUpQuestions: make([]string, 0, 2),
		ConversationFacts: buildConversationFacts(conversation, messages),
	}
	result.Title = buildDraftTitle(conversation, input)
	result.Description = buildDraftDescription(conversation, messages, input)
	if strings.TrimSpace(result.Title) == "" {
		result.MissingFields = append(result.MissingFields, "title")
		result.FollowUpQuestions = append(result.FollowUpQuestions, "请补充一个简洁的工单标题，明确概括用户遇到的问题。")
	}
	if !hasSufficientIssueContext(input, result.Description) {
		result.MissingFields = append(result.MissingFields, "issue")
		result.FollowUpQuestions = append(result.FollowUpQuestions, "请补充具体问题现象、报错信息或用户诉求，以便整理成工单。")
	}
	result.Ready = result.Title != "" && result.Description != "" && len(result.MissingFields) == 0
	return result
}

func buildDraftTitle(conversation models.Conversation, input PrepareTicketDraftInput) string {
	switch {
	case input.Title != "":
		return limitText(input.Title, 80)
	case input.Issue != "":
		return limitText(input.Issue, 80)
	case strings.TrimSpace(conversation.LastMessageSummary) != "":
		return limitText(conversation.LastMessageSummary, 80)
	default:
		return ""
	}
}

func buildDraftDescription(conversation models.Conversation, messages []models.Message, input PrepareTicketDraftInput) string {
	if input.Description != "" {
		return input.Description
	}
	parts := make([]string, 0, 6)
	if input.Issue != "" {
		parts = append(parts, "问题现象："+input.Issue)
	}
	if input.Impact != "" {
		parts = append(parts, "影响范围："+input.Impact)
	}
	if input.ExpectedOutcome != "" {
		parts = append(parts, "用户诉求："+input.ExpectedOutcome)
	}
	if input.CurrentAttempt != "" {
		parts = append(parts, "已尝试处理："+input.CurrentAttempt)
	}
	if strings.TrimSpace(conversation.LastMessageSummary) != "" {
		parts = append(parts, "会话摘要："+strings.TrimSpace(conversation.LastMessageSummary))
	}
	if recent := buildRecentMessageDigest(messages); recent != "" {
		parts = append(parts, "最近消息："+recent)
	}
	return strings.TrimSpace(strings.Join(parts, "\n"))
}

func hasSufficientIssueContext(input PrepareTicketDraftInput, description string) bool {
	if input.Issue != "" || input.Description != "" {
		return true
	}
	return len([]rune(strings.TrimSpace(description))) >= 30
}

func buildConversationFacts(conversation models.Conversation, messages []models.Message) []string {
	facts := make([]string, 0, 4)
	if strings.TrimSpace(conversation.LastMessageSummary) != "" {
		facts = append(facts, "最近摘要："+strings.TrimSpace(conversation.LastMessageSummary))
	}
	if digest := buildRecentMessageDigest(messages); digest != "" {
		facts = append(facts, "最近消息："+digest)
	}
	return facts
}

func buildRecentMessageDigest(messages []models.Message) string {
	if len(messages) == 0 {
		return ""
	}
	parts := make([]string, 0, len(messages))
	for i := range messages {
		content := strings.TrimSpace(messages[i].Content)
		if content == "" {
			continue
		}
		parts = append(parts, messageSenderLabel(messages[i].SenderType)+"："+limitText(content, 60))
	}
	return strings.Join(parts, " | ")
}

func messageSenderLabel(senderType enums.IMSenderType) string {
	switch senderType {
	case enums.IMSenderTypeCustomer:
		return "用户"
	case enums.IMSenderTypeAgent:
		return "客服"
	case enums.IMSenderTypeAI:
		return "AI"
	default:
		return "消息"
	}
}

func limitText(value string, max int) string {
	value = strings.TrimSpace(value)
	if max <= 0 {
		return value
	}
	runes := []rune(value)
	if len(runes) <= max {
		return value
	}
	return strings.TrimSpace(string(runes[:max])) + "..."
}
