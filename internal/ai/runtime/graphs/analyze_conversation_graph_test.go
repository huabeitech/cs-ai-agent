package graphs

import (
	"testing"

	"cs-agent/internal/models"
	"cs-agent/internal/pkg/enums"
)

func TestBuildAnalyzeConversationResult_RecommendsHandoffForComplaint(t *testing.T) {
	conversation := &models.Conversation{
		Subject:            "用户投诉扣费异常",
		LastMessageSummary: "用户反馈被重复扣费，并要求人工处理",
	}
	messages := []models.Message{
		{SenderType: enums.IMSenderTypeCustomer, Content: "你们重复扣费了，我要投诉并转人工"},
	}

	got := buildAnalyzeConversationResult(conversation, messages, AnalyzeConversationInput{
		NeedHumanHandoff: true,
	})

	if got.UserIntent != "handoff_request" {
		t.Fatalf("expected handoff_request, got %q", got.UserIntent)
	}
	if got.RiskLevel != "high" {
		t.Fatalf("expected high risk, got %q", got.RiskLevel)
	}
	if got.RecommendedNextAction != "handoff_to_human" {
		t.Fatalf("expected handoff_to_human, got %q", got.RecommendedNextAction)
	}
}

func TestBuildAnalyzeConversationResult_RecommendsPrepareTicket(t *testing.T) {
	conversation := &models.Conversation{
		Subject:            "订单无法支付",
		LastMessageSummary: "用户要求登记问题并尽快处理",
	}
	messages := []models.Message{
		{SenderType: enums.IMSenderTypeCustomer, Content: "麻烦帮我建个工单，订单一直支付失败"},
	}

	got := buildAnalyzeConversationResult(conversation, messages, AnalyzeConversationInput{
		NeedTicket: true,
	})

	if got.UserIntent != "ticket_request" {
		t.Fatalf("expected ticket_request, got %q", got.UserIntent)
	}
	if got.RecommendedNextAction != "prepare_ticket" {
		t.Fatalf("expected prepare_ticket, got %q", got.RecommendedNextAction)
	}
}
