package graphs

import (
	"testing"

	"cs-agent/internal/models"
	"cs-agent/internal/pkg/enums"
)

func TestBuildPrepareTicketDraftResult_UsesConversationFallbacks(t *testing.T) {
	conversation := models.Conversation{
		Subject:            "企业微信登录异常",
		LastMessageSummary: "用户反馈企业微信扫码后页面空白，无法进入工作台",
	}
	messages := []models.Message{
		{SenderType: enums.IMSenderTypeCustomer, Content: "扫码登录后一直白屏"},
		{SenderType: enums.IMSenderTypeAI, Content: "请问是否有报错提示"},
	}

	got := buildPrepareTicketDraftResult(conversation, messages, PrepareTicketDraftInput{
		Impact:          "无法进入后台处理客户消息",
		ExpectedOutcome: "恢复正常登录",
	})

	if got.Title == "" {
		t.Fatalf("expected draft title to be generated")
	}
	if got.Description == "" {
		t.Fatalf("expected draft description to be generated")
	}
	if !got.Ready {
		t.Fatalf("expected conversation summary and recent messages to be enough, got %#v", got)
	}
	if len(got.ConversationFacts) == 0 {
		t.Fatalf("expected conversation facts to be populated")
	}
}

func TestBuildPrepareTicketDraftResult_ReadyWithExplicitIssue(t *testing.T) {
	conversation := models.Conversation{
		Subject:            "订单支付失败",
		LastMessageSummary: "用户反馈连续支付失败",
	}

	got := buildPrepareTicketDraftResult(conversation, nil, PrepareTicketDraftInput{
		Issue:           "用户连续三次支付订单失败，页面提示网络异常。",
		ExpectedOutcome: "希望尽快恢复支付并完成下单。",
		CurrentAttempt:  "已尝试切换网络和刷新页面，问题仍存在。",
	})

	if !got.Ready {
		t.Fatalf("expected draft to be ready, got %#v", got)
	}
	if got.Title == "" || got.Description == "" {
		t.Fatalf("expected title and description to be populated, got %#v", got)
	}
}
