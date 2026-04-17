package graphs

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"cs-agent/internal/models"
	"cs-agent/internal/services"
)

type AnalyzeConversationInput struct {
	Goal              string `json:"goal"`
	ObservedIssue     string `json:"observedIssue"`
	NeedTicket        bool   `json:"needTicket"`
	NeedHumanHandoff  bool   `json:"needHumanHandoff"`
	NeedQualityCheck  bool   `json:"needQualityCheck"`
	AdditionalContext string `json:"additionalContext"`
}

type AnalyzeConversationResult struct {
	Summary               string   `json:"summary"`
	UserIntent            string   `json:"userIntent"`
	RiskLevel             string   `json:"riskLevel"`
	RiskSignals           []string `json:"riskSignals,omitempty"`
	RecommendedNextAction string   `json:"recommendedNextAction"`
	RecommendedQuestions  []string `json:"recommendedQuestions,omitempty"`
	ConversationFacts     []string `json:"conversationFacts,omitempty"`
}

type AnalyzeConversationGraph struct {
	conversation models.Conversation
}

func NewAnalyzeConversationGraph(conversation models.Conversation) *AnalyzeConversationGraph {
	return &AnalyzeConversationGraph{conversation: conversation}
}

func (g *AnalyzeConversationGraph) Run(_ context.Context, argumentsInJSON string) (string, error) {
	input, err := g.parseInput(argumentsInJSON)
	if err != nil {
		return "", err
	}
	messages, _, _ := services.MessageService.FindByConversationIDCursor(g.conversation.ID, 0, 8, "", "")
	result := buildAnalyzeConversationResult(g.conversation, messages, input)
	buf, err := json.Marshal(result)
	if err != nil {
		return "", err
	}
	return string(buf), nil
}

func (g *AnalyzeConversationGraph) parseInput(argumentsInJSON string) (AnalyzeConversationInput, error) {
	var input AnalyzeConversationInput
	if strings.TrimSpace(argumentsInJSON) == "" {
		return input, nil
	}
	if err := json.Unmarshal([]byte(argumentsInJSON), &input); err != nil {
		return input, fmt.Errorf("invalid analyze conversation arguments: %w", err)
	}
	input.Goal = strings.TrimSpace(input.Goal)
	input.ObservedIssue = strings.TrimSpace(input.ObservedIssue)
	input.AdditionalContext = strings.TrimSpace(input.AdditionalContext)
	return input, nil
}

func buildAnalyzeConversationResult(conversation models.Conversation, messages []models.Message, input AnalyzeConversationInput) AnalyzeConversationResult {
	joined := strings.ToLower(buildConversationCorpus(conversation, messages, input))
	signals := collectRiskSignals(joined, input)
	intent := detectUserIntent(joined, input)
	recommendedAction := recommendNextAction(intent, signals, input)
	result := AnalyzeConversationResult{
		Summary:               buildConversationSummary(conversation, messages, input),
		UserIntent:            intent,
		RiskLevel:             deriveRiskLevel(signals),
		RiskSignals:           signals,
		RecommendedNextAction: recommendedAction,
		RecommendedQuestions:  recommendQuestions(intent, signals, input),
		ConversationFacts:     buildConversationAnalysisFacts(conversation, messages),
	}
	return result
}

func buildConversationSummary(conversation models.Conversation, messages []models.Message, input AnalyzeConversationInput) string {
	parts := make([]string, 0, 4)
	if strings.TrimSpace(conversation.Subject) != "" {
		parts = append(parts, "会话主题："+strings.TrimSpace(conversation.Subject))
	}
	if input.ObservedIssue != "" {
		parts = append(parts, "当前问题："+input.ObservedIssue)
	} else if strings.TrimSpace(conversation.LastMessageSummary) != "" {
		parts = append(parts, "当前问题："+strings.TrimSpace(conversation.LastMessageSummary))
	}
	if digest := buildRecentMessageDigest(messages); digest != "" {
		parts = append(parts, "最近对话："+digest)
	}
	if input.AdditionalContext != "" {
		parts = append(parts, "补充信息："+input.AdditionalContext)
	}
	return strings.TrimSpace(strings.Join(parts, "\n"))
}

func buildConversationCorpus(conversation models.Conversation, messages []models.Message, input AnalyzeConversationInput) string {
	parts := make([]string, 0, len(messages)+4)
	parts = append(parts, strings.TrimSpace(conversation.Subject))
	parts = append(parts, strings.TrimSpace(conversation.LastMessageSummary))
	parts = append(parts, input.Goal, input.ObservedIssue, input.AdditionalContext)
	for i := range messages {
		parts = append(parts, strings.TrimSpace(messages[i].Content))
	}
	return strings.Join(parts, "\n")
}

func collectRiskSignals(joined string, input AnalyzeConversationInput) []string {
	signals := make([]string, 0, 6)
	add := func(signal string) {
		for _, item := range signals {
			if item == signal {
				return
			}
		}
		signals = append(signals, signal)
	}
	if containsAny(joined, "投诉", "举报", "差评", "曝光", "媒体", "起诉", "律师", "12315") {
		add("complaint_escalation")
	}
	if containsAny(joined, "退款", "赔偿", "损失", "扣款", "重复扣费", "金额") {
		add("financial_risk")
	}
	if containsAny(joined, "生气", "愤怒", "垃圾", "太差", "一直没人", "再不处理") {
		add("negative_sentiment")
	}
	if containsAny(joined, "人工", "转人工", "真人", "客服") || input.NeedHumanHandoff {
		add("handoff_requested")
	}
	if containsAny(joined, "工单", "报障", "售后", "登记", "记录问题") || input.NeedTicket {
		add("ticket_expected")
	}
	if input.NeedQualityCheck {
		add("quality_review_requested")
	}
	return signals
}

func detectUserIntent(joined string, input AnalyzeConversationInput) string {
	switch {
	case input.NeedHumanHandoff || containsAny(joined, "人工", "转人工", "真人"):
		return "handoff_request"
	case input.NeedTicket || containsAny(joined, "工单", "报障", "售后", "登记问题"):
		return "ticket_request"
	case containsAny(joined, "投诉", "举报", "差评", "赔偿"):
		return "complaint"
	default:
		return "general_support"
	}
}

func deriveRiskLevel(signals []string) string {
	if len(signals) == 0 {
		return "low"
	}
	if containsSignal(signals, "complaint_escalation") || containsSignal(signals, "financial_risk") {
		return "high"
	}
	if containsSignal(signals, "handoff_requested") || containsSignal(signals, "negative_sentiment") {
		return "medium"
	}
	return "low"
}

func recommendNextAction(intent string, signals []string, input AnalyzeConversationInput) string {
	switch {
	case input.NeedQualityCheck:
		return "quality_review"
	case containsSignal(signals, "handoff_requested") || intent == "handoff_request":
		return "handoff_to_human"
	case containsSignal(signals, "ticket_expected") || intent == "ticket_request":
		return "prepare_ticket"
	case containsSignal(signals, "complaint_escalation"):
		return "handoff_to_human"
	default:
		return "continue_answering"
	}
}

func recommendQuestions(intent string, signals []string, input AnalyzeConversationInput) []string {
	questions := make([]string, 0, 3)
	if containsSignal(signals, "ticket_expected") && strings.TrimSpace(input.ObservedIssue) == "" {
		questions = append(questions, "请进一步确认用户遇到的具体问题现象、报错信息和期望处理结果。")
	}
	if containsSignal(signals, "handoff_requested") {
		questions = append(questions, "请确认用户是否明确要求人工客服，以及当前问题为何需要人工继续处理。")
	}
	if intent == "complaint" {
		questions = append(questions, "请确认投诉点、影响范围和用户当前最希望解决的事项。")
	}
	return questions
}

func buildConversationAnalysisFacts(conversation models.Conversation, messages []models.Message) []string {
	facts := make([]string, 0, 4)
	if strings.TrimSpace(conversation.Subject) != "" {
		facts = append(facts, "会话主题："+strings.TrimSpace(conversation.Subject))
	}
	if strings.TrimSpace(conversation.LastMessageSummary) != "" {
		facts = append(facts, "最近摘要："+strings.TrimSpace(conversation.LastMessageSummary))
	}
	if digest := buildRecentMessageDigest(messages); digest != "" {
		facts = append(facts, "最近消息："+digest)
	}
	return facts
}

func containsAny(value string, keywords ...string) bool {
	for _, keyword := range keywords {
		if keyword != "" && strings.Contains(value, strings.ToLower(keyword)) {
			return true
		}
	}
	return false
}

func containsSignal(signals []string, target string) bool {
	for _, item := range signals {
		if item == target {
			return true
		}
	}
	return false
}
