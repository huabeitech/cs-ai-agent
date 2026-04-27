package graphs

import (
	"testing"

	"cs-agent/internal/models"
	"cs-agent/internal/pkg/enums"
)

func TestTriageServiceRequestResult_PrepareTicket(t *testing.T) {
	conversation := models.Conversation{
		LastMessageSummary: "用户要求建单跟进支付失败问题",
	}
	messages := []models.Message{
		{SenderType: enums.IMSenderTypeCustomer, Content: "帮我建个工单，支付一直失败"},
	}

	analysis := buildAnalyzeConversationResult(conversation, messages, AnalyzeConversationInput{
		NeedTicket: true,
	})
	if analysis.RecommendedNextAction != "prepare_ticket" {
		t.Fatalf("expected prepare_ticket, got %q", analysis.RecommendedNextAction)
	}

	draft := buildPrepareTicketDraftResult(conversation, messages, PrepareTicketDraftInput{})
	if draft.Title == "" || draft.Description == "" {
		t.Fatalf("expected draft to be populated, got %#v", draft)
	}
}

func TestTriageServiceRequestResult_Handoff(t *testing.T) {
	conversation := models.Conversation{
		LastMessageSummary: "用户要求人工处理扣费投诉",
	}
	messages := []models.Message{
		{SenderType: enums.IMSenderTypeCustomer, Content: "我要投诉并转人工，你们重复扣费了"},
	}

	analysis := buildAnalyzeConversationResult(conversation, messages, AnalyzeConversationInput{
		NeedHumanHandoff: true,
	})
	if analysis.RecommendedNextAction != "handoff_to_human" {
		t.Fatalf("expected handoff_to_human, got %q", analysis.RecommendedNextAction)
	}
}
