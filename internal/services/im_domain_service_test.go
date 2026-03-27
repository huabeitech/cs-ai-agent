package services

import (
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/constants"
	"cs-agent/internal/pkg/dto"
	"cs-agent/internal/pkg/enums"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	"github.com/mlogclub/simple/sqls"
)

func setupIMServiceTestDB(t *testing.T) {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "im-service-test.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "t_",
			SingularTable: true,
		},
	})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}
	sqls.SetDB(db)
	if err := sqls.DB().AutoMigrate(models.Models...); err != nil {
		t.Fatalf("auto migrate failed: %v", err)
	}

	t.Cleanup(func() {
		sqlDB, err := db.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
		_ = os.Remove(dbPath)
	})
}

func seedAgentUser(t *testing.T, id int64, username string) {
	t.Helper()

	now := time.Now()
	user := &models.User{
		ID:           id,
		Username:     username,
		Nickname:     username,
		Password:     "hashed",
		Status:       enums.StatusOk,
		PasswordSalt: "",
		AuditFields: models.AuditFields{
			CreatedAt:      now,
			CreateUserID:   id,
			CreateUserName: username,
			UpdatedAt:      now,
			UpdateUserID:   id,
			UpdateUserName: username,
		},
	}
	if err := sqls.DB().Create(user).Error; err != nil {
		t.Fatalf("seed user failed: %v", err)
	}
	if err := sqls.DB().Model(user).Update("status", int(enums.StatusOk)).Error; err != nil {
		t.Fatalf("update user status failed: %v", err)
	}
}

func seedEnabledAIAgent(t *testing.T, id int64, name string) *models.AIAgent {
	t.Helper()

	now := time.Now()
	item := &models.AIAgent{
		ID:               id,
		Name:             name,
		Status:           enums.StatusOk,
		AIConfigID:       0,
		ServiceMode:      enums.IMConversationServiceModeAIFirst,
		HandoffMode:      enums.AIAgentHandoffModeWaitPool,
		MaxAIReplyRounds: 2,
		FallbackMode:     enums.AIAgentFallbackModeGuideRephrase,
		KnowledgeIDs:     "0",
		AuditFields: models.AuditFields{
			CreatedAt:      now,
			CreateUserID:   0,
			CreateUserName: "system",
			UpdatedAt:      now,
			UpdateUserID:   0,
			UpdateUserName: "system",
		},
	}
	if err := sqls.DB().Create(item).Error; err != nil {
		t.Fatalf("seed ai agent failed: %v", err)
	}
	if err := sqls.DB().Model(item).Update("status", int(enums.StatusOk)).Error; err != nil {
		t.Fatalf("update ai agent status failed: %v", err)
	}
	item.Status = enums.StatusOk
	return item
}

func TestCreateOrMatchConversationWithAIAgentPersistsAIAgentID(t *testing.T) {
	setupIMServiceTestDB(t)

	seedEnabledAIAgent(t, 301, "AI接待")
	customer := &dto.AuthPrincipal{UserID: 1011, Username: "customer_ai_agent"}
	item, err := ConversationService.Create(
		enums.IMConversationChannelWebChat,
		"AI接待测试",
		301,
		customer,
	)
	if err != nil {
		t.Fatalf("create conversation with ai agent failed: %v", err)
	}
	if item.AIAgentID != 301 {
		t.Fatalf("expected ai agent id 301, got %d", item.AIAgentID)
	}

	stored := ConversationService.Get(item.ID)
	if stored == nil {
		t.Fatalf("expected stored conversation, got nil")
	}
	if stored.AIAgentID != 301 {
		t.Fatalf("expected stored ai agent id 301, got %d", stored.AIAgentID)
	}
}

func TestCreateConversationInitializesLastMessageAt(t *testing.T) {
	setupIMServiceTestDB(t)

	seedEnabledAIAgent(t, 303, "初始化最后消息时间")
	customer := &dto.AuthPrincipal{UserID: 1014, Username: "customer_last_message_at"}
	item, err := ConversationService.Create(
		enums.IMConversationChannelWebChat,
		"初始化最后消息时间测试",
		303,
		customer,
	)
	if err != nil {
		t.Fatalf("create conversation failed: %v", err)
	}
	if !item.LastMessageAt.Equal(item.CreatedAt) {
		t.Fatalf("expected last message at %v to equal created at %v", item.LastMessageAt, item.CreatedAt)
	}
	if !item.LastActiveTime.Equal(item.CreatedAt) {
		t.Fatalf("expected last active time %v to equal created at %v", item.LastActiveTime, item.CreatedAt)
	}

	stored := ConversationService.Get(item.ID)
	if stored == nil {
		t.Fatalf("expected stored conversation, got nil")
	}
	if !stored.LastMessageAt.Equal(stored.CreatedAt) {
		t.Fatalf("expected stored last message at %v to equal created at %v", stored.LastMessageAt, stored.CreatedAt)
	}
	if !stored.LastActiveTime.Equal(stored.CreatedAt) {
		t.Fatalf("expected stored last active time %v to equal created at %v", stored.LastActiveTime, stored.CreatedAt)
	}
}

func TestCreateOrMatchConversationWithAIAgentFillsExistingConversationAIAgentID(t *testing.T) {
	setupIMServiceTestDB(t)

	customer := &dto.AuthPrincipal{UserID: 1012, Username: "customer_ai_agent_fill"}
	item, err := ConversationService.Create(
		enums.IMConversationChannelWebChat,
		"AI接待补齐测试",
		0,
		customer,
	)
	if err != nil {
		t.Fatalf("create conversation failed: %v", err)
	}
	if item.AIAgentID != 0 {
		t.Fatalf("expected initial ai agent id 0, got %d", item.AIAgentID)
	}

	matched, err := ConversationService.Create(
		enums.IMConversationChannelWebChat,
		"AI接待补齐测试",
		302,
		customer,
	)
	if err != nil {
		t.Fatalf("match conversation with ai agent failed: %v", err)
	}
	if matched.ID != item.ID {
		t.Fatalf("expected matched conversation %d, got %d", item.ID, matched.ID)
	}
	if matched.AIAgentID != 302 {
		t.Fatalf("expected matched ai agent id 302, got %d", matched.AIAgentID)
	}

	stored := ConversationService.Get(item.ID)
	if stored == nil {
		t.Fatalf("expected stored conversation, got nil")
	}
	if stored.AIAgentID != 302 {
		t.Fatalf("expected stored ai agent id 302, got %d", stored.AIAgentID)
	}
}

func TestSendCustomerMessageUpdatesConversationSummary(t *testing.T) {
	setupIMServiceTestDB(t)

	seedEnabledAIAgent(t, 304, "发送消息测试AI")
	customer := &dto.AuthPrincipal{UserID: 1002, Username: "customer_2"}
	conversation, err := ConversationService.Create(enums.IMConversationChannelWebChat, "物流问题", 304, customer)
	if err != nil {
		t.Fatalf("create conversation failed: %v", err)
	}

	message, err := MessageService.SendMessage(conversation.ID, enums.IMSenderTypeCustomer, 0, "client_1", enums.IMMessageTypeText, "你好，想查询物流", "", customer)
	if err != nil {
		t.Fatalf("send message failed: %v", err)
	}
	if message.SeqNo != 1 {
		t.Fatalf("expected seq 1, got %d", message.SeqNo)
	}

	duplicated, err := MessageService.SendCustomerMessage(conversation.ID, "client_1", enums.IMMessageTypeText, "你好，想查询物流", "", customer)
	if err != nil {
		t.Fatalf("repeat send failed: %v", err)
	}
	if duplicated.ID != message.ID {
		t.Fatalf("expected deduplicated message %d, got %d", message.ID, duplicated.ID)
	}

	updated := ConversationService.Get(conversation.ID)
	if updated.LastMessageID != message.ID {
		t.Fatalf("expected last message id %d, got %d", message.ID, updated.LastMessageID)
	}
	if updated.LastActiveTime.Before(conversation.CreatedAt) {
		t.Fatalf("expected last active time %v after created at %v", updated.LastActiveTime, conversation.CreatedAt)
	}
	if updated.LastMessageSummary != "你好，想查询物流" {
		t.Fatalf("unexpected last message summary: %s", updated.LastMessageSummary)
	}
	if updated.AgentUnreadCount != 1 {
		t.Fatalf("expected agent unread 1, got %d", updated.AgentUnreadCount)
	}
}

func TestSendAIMessageUpdatesConversationSummaryAndUnreadCount(t *testing.T) {
	setupIMServiceTestDB(t)

	agent := seedEnabledAIAgent(t, 301, "售前AI")

	customer := &dto.AuthPrincipal{UserID: 1013, Username: "customer_ai_message"}
	conversation, err := ConversationService.Create(
		enums.IMConversationChannelWebChat,
		"AI消息测试",
		agent.ID,
		customer,
	)
	if err != nil {
		t.Fatalf("create conversation failed: %v", err)
	}

	operator := &dto.AuthPrincipal{Username: "售前AI", Nickname: "售前AI"}
	message, err := MessageService.SendAIMessage(conversation.ID, agent.ID, "ai_message_1", enums.IMMessageTypeText, "你好，我来帮你处理。", "", operator)
	if err != nil {
		t.Fatalf("send ai message failed: %v", err)
	}
	if message.SenderType != enums.IMSenderTypeAI {
		t.Fatalf("expected sender type ai, got %s", message.SenderType)
	}
	if message.SenderID != agent.ID {
		t.Fatalf("expected sender id %d, got %d", agent.ID, message.SenderID)
	}

	updated := ConversationService.Get(conversation.ID)
	if updated.LastMessageID != message.ID {
		t.Fatalf("expected last message id %d, got %d", message.ID, updated.LastMessageID)
	}
	if updated.CustomerUnreadCount != 1 {
		t.Fatalf("expected customer unread 1, got %d", updated.CustomerUnreadCount)
	}
	if updated.AgentUnreadCount != 0 {
		t.Fatalf("expected agent unread 0, got %d", updated.AgentUnreadCount)
	}

}

func TestPendingConversationKeepsPendingWhenCustomerAndAISendMessages(t *testing.T) {
	setupIMServiceTestDB(t)

	agent := seedEnabledAIAgent(t, 401, "状态测试AI")

	customer := &dto.AuthPrincipal{UserID: 1201, Username: "customer_pending_status"}
	conversation, err := ConversationService.Create(
		enums.IMConversationChannelWebChat,
		"状态保持测试",
		agent.ID,
		customer,
	)
	if err != nil {
		t.Fatalf("create conversation failed: %v", err)
	}
	if conversation.Status != enums.IMConversationStatusPending {
		t.Fatalf("expected initial status pending, got %d", conversation.Status)
	}

	if _, err := MessageService.SendCustomerMessage(conversation.ID, "pending_customer_1", enums.IMMessageTypeText, "客户先发消息", "", customer); err != nil {
		t.Fatalf("send customer message failed: %v", err)
	}

	updated := ConversationService.Get(conversation.ID)
	if updated.Status != enums.IMConversationStatusPending {
		t.Fatalf("expected status pending after customer message, got %d", updated.Status)
	}
	if updated.ClosedAt != nil {
		t.Fatalf("expected closedAt nil after customer message")
	}

	aiOperator := &dto.AuthPrincipal{Username: "状态测试AI", Nickname: "状态测试AI"}
	if _, err := MessageService.SendAIMessage(conversation.ID, agent.ID, "pending_ai_1", enums.IMMessageTypeText, "AI继续接待", "", aiOperator); err != nil {
		t.Fatalf("send ai message failed: %v", err)
	}

	updated = ConversationService.Get(conversation.ID)
	if updated.Status != enums.IMConversationStatusPending {
		t.Fatalf("expected status pending after ai message, got %d", updated.Status)
	}
	if updated.ClosedAt != nil {
		t.Fatalf("expected closedAt nil after ai message")
	}
}

func TestSendCustomerImageMessageUpdatesConversationSummary(t *testing.T) {
	setupIMServiceTestDB(t)

	customer := &dto.AuthPrincipal{UserID: 1004, Username: "customer_4"}
	conversation, err := ConversationService.Create(enums.IMConversationChannelWebChat, "图片咨询", 0, customer)
	if err != nil {
		t.Fatalf("create conversation failed: %v", err)
	}

	message, err := MessageService.SendCustomerMessage(
		conversation.ID,
		"client_image_1",
		enums.IMMessageTypeImage,
		"",
		`{"url":"https://example.com/test.png","filename":"test.png","mimeType":"image/png"}`,
		customer,
	)
	if err != nil {
		t.Fatalf("send image message failed: %v", err)
	}
	if message.MessageType != enums.IMMessageTypeImage {
		t.Fatalf("expected message type image, got %s", message.MessageType)
	}

	updated := ConversationService.Get(conversation.ID)
	if updated.LastMessageID != message.ID {
		t.Fatalf("expected last message id %d, got %d", message.ID, updated.LastMessageID)
	}
	if updated.LastMessageSummary != "[图片]" {
		t.Fatalf("unexpected last message summary: %s", updated.LastMessageSummary)
	}
	if updated.AgentUnreadCount != 1 {
		t.Fatalf("expected agent unread 1, got %d", updated.AgentUnreadCount)
	}
}

func TestSendCustomerHTMLMessageUpdatesConversationSummary(t *testing.T) {
	setupIMServiceTestDB(t)

	customer := &dto.AuthPrincipal{UserID: 1006, Username: "customer_6"}
	conversation, err := ConversationService.Create(enums.IMConversationChannelWebChat, "HTML咨询", 0, customer)
	if err != nil {
		t.Fatalf("create conversation failed: %v", err)
	}

	message, err := MessageService.SendCustomerMessage(
		conversation.ID,
		"client_html_1",
		enums.IMMessageTypeHTML,
		`<p>请看下面</p><p><img src="https://example.com/test.png" alt="test"></p><p>这是原图</p>`,
		"",
		customer,
	)
	if err != nil {
		t.Fatalf("send html message failed: %v", err)
	}
	if message.MessageType != enums.IMMessageTypeHTML {
		t.Fatalf("expected html message type, got %s", message.MessageType)
	}

	updated := ConversationService.Get(conversation.ID)
	if updated.LastMessageSummary != "请看下面 [图片] 这是原图" {
		t.Fatalf("unexpected last message summary: %s", updated.LastMessageSummary)
	}
}

func TestConversationAssignTransferAndClose(t *testing.T) {
	setupIMServiceTestDB(t)

	agentModel := seedEnabledAIAgent(t, 404, "流程测试AI")
	seedAgentUser(t, 2001, "agent_1")
	seedAgentUser(t, 2002, "agent_2")

	customer := &dto.AuthPrincipal{UserID: 1003, Username: "customer_3"}
	admin := &dto.AuthPrincipal{UserID: 9001, Username: "admin_1", Roles: []string{constants.RoleCodeAdmin}}
	agentOne := &dto.AuthPrincipal{UserID: 2001, Username: "agent_1"}
	agentTwo := &dto.AuthPrincipal{UserID: 2002, Username: "agent_2"}

	conversation, err := ConversationService.Create(enums.IMConversationChannelWebChat, "售后问题", agentModel.ID, customer)
	if err != nil {
		t.Fatalf("create conversation failed: %v", err)
	}

	if err := ConversationService.AssignConversation(conversation.ID, agentOne.UserID, "管理员分配", admin); err != nil {
		t.Fatalf("assign conversation failed: %v", err)
	}
	if err := ConversationService.TransferConversation(conversation.ID, agentTwo.UserID, "升级处理", admin); err != nil {
		t.Fatalf("transfer conversation failed: %v", err)
	}
	if err := ConversationService.CloseConversation(conversation.ID, "处理完成", agentTwo); err != nil {
		t.Fatalf("close conversation failed: %v", err)
	}

	updated := ConversationService.Get(conversation.ID)
	if updated.CurrentAssigneeID != agentTwo.UserID {
		t.Fatalf("expected assignee %d, got %d", agentTwo.UserID, updated.CurrentAssigneeID)
	}
	if updated.Status != enums.IMConversationStatusClosed {
		t.Fatalf("expected closed status 3, got %d", updated.Status)
	}
	if updated.ClosedAt == nil {
		t.Fatalf("expected closedAt to be set")
	}
	if updated.ClosedBy != agentTwo.UserID {
		t.Fatalf("expected closedBy %d, got %d", agentTwo.UserID, updated.ClosedBy)
	}
	if updated.CloseReason != "处理完成" {
		t.Fatalf("expected close reason to be recorded, got %s", updated.CloseReason)
	}

	assignments := ConversationAssignmentService.Find(sqls.NewCnd().Eq("conversation_id", conversation.ID).Asc("id"))
	if len(assignments) != 2 {
		t.Fatalf("expected 2 assignments, got %d", len(assignments))
	}
	if assignments[0].Status != enums.IMAssignmentStatusInactive {
		t.Fatalf("expected first assignment finished, got status %d", assignments[0].Status)
	}
	if assignments[1].ToUserID != agentTwo.UserID || assignments[1].Status != enums.IMAssignmentStatusInactive {
		t.Fatalf("unexpected second assignment: %#v", assignments[1])
	}

	events := ConversationEventLogService.Find(sqls.NewCnd().Eq("conversation_id", conversation.ID).Asc("id"))
	if len(events) != 4 {
		t.Fatalf("expected 4 events, got %d", len(events))
	}
	if events[1].EventType != enums.IMEventTypeAssign {
		t.Fatalf("expected second event to be assign, got %s", events[1].EventType)
	}
}

func TestCustomerCanCloseOwnConversation(t *testing.T) {
	setupIMServiceTestDB(t)

	customer := &dto.AuthPrincipal{IsVisitor: true, VisitorID: "visitor_close_1", Username: "访客", Nickname: "访客"}
	conversation, err := ConversationService.Create(enums.IMConversationChannelWebChat, "关闭测试", 0, customer)
	if err != nil {
		t.Fatalf("create conversation failed: %v", err)
	}

	if err := ConversationService.CloseCustomerConversation(conversation.ID, customer); err != nil {
		t.Fatalf("customer close conversation failed: %v", err)
	}

	updated := ConversationService.Get(conversation.ID)
	if updated.Status != enums.IMConversationStatusClosed {
		t.Fatalf("expected closed status 3, got %d", updated.Status)
	}
	if updated.ClosedAt == nil {
		t.Fatalf("expected closedAt to be set")
	}

	events := ConversationEventLogService.Find(sqls.NewCnd().Eq("conversation_id", conversation.ID).Asc("id"))
	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}
	if events[1].EventType != enums.IMEventTypeClose {
		t.Fatalf("expected close event, got %s", events[1].EventType)
	}
	if events[1].OperatorType != enums.IMSenderTypeCustomer {
		t.Fatalf("expected customer operator type, got %s", events[1].OperatorType)
	}
}

func TestClosedConversationRejectsAllMessageSenders(t *testing.T) {
	setupIMServiceTestDB(t)

	agentModel := seedEnabledAIAgent(t, 402, "关闭态AI")
	seedAgentUser(t, 4001, "agent_closed_1")

	customer := &dto.AuthPrincipal{UserID: 1202, Username: "customer_closed_status"}
	agent := &dto.AuthPrincipal{UserID: 4001, Username: "agent_closed_1"}
	admin := &dto.AuthPrincipal{UserID: 9002, Username: "admin_2", Roles: []string{constants.RoleCodeAdmin}}
	aiOperator := &dto.AuthPrincipal{Username: "关闭态AI", Nickname: "关闭态AI"}

	conversation, err := ConversationService.Create(
		enums.IMConversationChannelWebChat,
		"关闭后禁止发消息",
		agentModel.ID,
		customer,
	)
	if err != nil {
		t.Fatalf("create conversation failed: %v", err)
	}
	if err := ConversationService.AssignConversation(conversation.ID, agent.UserID, "管理员分配", admin); err != nil {
		t.Fatalf("assign conversation failed: %v", err)
	}
	if err := ConversationService.CloseConversation(conversation.ID, "人工关闭", agent); err != nil {
		t.Fatalf("close conversation failed: %v", err)
	}

	if _, err := MessageService.SendCustomerMessage(conversation.ID, "closed_customer_1", enums.IMMessageTypeText, "客户继续发消息", "", customer); err == nil || !strings.Contains(err.Error(), "会话已关闭") {
		t.Fatalf("expected customer send to fail with closed error, got %v", err)
	}
	if _, err := MessageService.SendAgentMessage(conversation.ID, 0, "closed_agent_1", enums.IMMessageTypeText, "客服继续发消息", "", agent); err == nil || !strings.Contains(err.Error(), "会话已关闭") {
		t.Fatalf("expected agent send to fail with closed error, got %v", err)
	}
	if _, err := MessageService.SendAIMessage(conversation.ID, agentModel.ID, "closed_ai_1", enums.IMMessageTypeText, "AI继续发消息", "", aiOperator); err == nil || !strings.Contains(err.Error(), "会话已关闭") {
		t.Fatalf("expected ai send to fail with closed error, got %v", err)
	}
}

func TestCreateAfterClosedConversationCreatesNewConversation(t *testing.T) {
	setupIMServiceTestDB(t)

	agent := seedEnabledAIAgent(t, 403, "重建会话AI")

	customer := &dto.AuthPrincipal{UserID: 1203, Username: "customer_recreate"}
	firstConversation, err := ConversationService.Create(
		enums.IMConversationChannelWebChat,
		"关闭后重建",
		agent.ID,
		customer,
	)
	if err != nil {
		t.Fatalf("create first conversation failed: %v", err)
	}
	if err := ConversationService.CloseCustomerConversation(firstConversation.ID, customer); err != nil {
		t.Fatalf("close customer conversation failed: %v", err)
	}

	secondConversation, err := ConversationService.Create(
		enums.IMConversationChannelWebChat,
		"关闭后重建",
		agent.ID,
		customer,
	)
	if err != nil {
		t.Fatalf("create second conversation failed: %v", err)
	}
	if secondConversation.ID == firstConversation.ID {
		t.Fatalf("expected a new conversation after close, got same id %d", secondConversation.ID)
	}
	if secondConversation.Status != enums.IMConversationStatusPending {
		t.Fatalf("expected new conversation status pending, got %d", secondConversation.Status)
	}
}

func TestConversationReadStateTracksMessageReadDimension(t *testing.T) {
	setupIMServiceTestDB(t)

	seedAgentUser(t, 3001, "agent_read_1")
	seedEnabledAIAgent(t, 405, "已读测试AI")

	customer := &dto.AuthPrincipal{UserID: 1101, Username: "customer_read_1"}
	agent := &dto.AuthPrincipal{UserID: 3001, Username: "agent_read_1"}
	admin := &dto.AuthPrincipal{UserID: 9003, Username: "admin_3", Roles: []string{constants.RoleCodeAdmin}}

	conversation, err := ConversationService.Create(enums.IMConversationChannelWebChat, "已读测试", 405, customer)
	if err != nil {
		t.Fatalf("create conversation failed: %v", err)
	}
	if err := ConversationService.AssignConversation(conversation.ID, agent.UserID, "管理员分配", admin); err != nil {
		t.Fatalf("assign conversation failed: %v", err)
	}

	customerMessage, err := MessageService.SendCustomerMessage(conversation.ID, "read_customer_1", enums.IMMessageTypeText, "客户消息", "", customer)
	if err != nil {
		t.Fatalf("send customer message failed: %v", err)
	}
	agentMessage, err := MessageService.SendAgentMessage(conversation.ID, 0, "read_agent_1", enums.IMMessageTypeText, "客服回复", "", agent)
	if err != nil {
		t.Fatalf("send agent message failed: %v", err)
	}

	if err := ConversationService.MarkConversationReadToMessage(conversation.ID, customerMessage.ID, enums.IMSenderTypeAgent, agent); err != nil {
		t.Fatalf("mark agent read failed: %v", err)
	}
	if err := ConversationService.MarkConversationReadToMessage(conversation.ID, agentMessage.ID, enums.IMSenderTypeCustomer, customer); err != nil {
		t.Fatalf("mark customer read failed: %v", err)
	}

	updated := ConversationService.Get(conversation.ID)
	if updated.AgentUnreadCount != 0 {
		t.Fatalf("expected agent unread 0, got %d", updated.AgentUnreadCount)
	}
	if updated.CustomerUnreadCount != 0 {
		t.Fatalf("expected customer unread 0, got %d", updated.CustomerUnreadCount)
	}

	agentReadState, customerReadState := ConversationReadStateService.GetConversationReadStates(conversation.ID)
	if agentReadState == nil || agentReadState.LastReadMessageID != agentMessage.ID {
		t.Fatalf("expected agent read cursor on latest agent message")
	}
	if customerReadState == nil || customerReadState.LastReadMessageID != agentMessage.ID {
		t.Fatalf("expected customer read cursor on agent message")
	}
	if agentReadState.LastReadSeqNo < agentMessage.SeqNo {
		t.Fatalf("expected agent read cursor to reach latest message")
	}
	if customerReadState.LastReadSeqNo != agentMessage.SeqNo {
		t.Fatalf("expected customer read seq %d, got %d", agentMessage.SeqNo, customerReadState.LastReadSeqNo)
	}
}

func TestSendMessageRefreshesUnreadCountsByReadCursor(t *testing.T) {
	setupIMServiceTestDB(t)

	seedAgentUser(t, 3002, "agent_read_2")
	seedEnabledAIAgent(t, 406, "未读测试AI")

	customer := &dto.AuthPrincipal{UserID: 1102, Username: "customer_read_2"}
	agent := &dto.AuthPrincipal{UserID: 3002, Username: "agent_read_2"}
	admin := &dto.AuthPrincipal{UserID: 9004, Username: "admin_4", Roles: []string{constants.RoleCodeAdmin}}

	conversation, err := ConversationService.Create(enums.IMConversationChannelWebChat, "未读数测试", 406, customer)
	if err != nil {
		t.Fatalf("create conversation failed: %v", err)
	}
	if err := ConversationService.AssignConversation(conversation.ID, agent.UserID, "管理员分配", admin); err != nil {
		t.Fatalf("assign conversation failed: %v", err)
	}
	if _, err := MessageService.SendCustomerMessage(conversation.ID, "read_cursor_customer_1", enums.IMMessageTypeText, "第一条", "", customer); err != nil {
		t.Fatalf("send first customer message failed: %v", err)
	}
	if _, err := MessageService.SendCustomerMessage(conversation.ID, "read_cursor_customer_2", enums.IMMessageTypeText, "第二条", "", customer); err != nil {
		t.Fatalf("send second customer message failed: %v", err)
	}

	beforeRead := ConversationService.Get(conversation.ID)
	if beforeRead.AgentUnreadCount != 2 {
		t.Fatalf("expected agent unread 2 before read, got %d", beforeRead.AgentUnreadCount)
	}

	firstMessage := MessageService.FindOne(sqls.NewCnd().Eq("conversation_id", conversation.ID).Asc("seq_no"))
	if firstMessage == nil {
		t.Fatalf("expected first message")
	}
	if err := ConversationService.MarkConversationReadToMessage(conversation.ID, firstMessage.ID, enums.IMSenderTypeAgent, agent); err != nil {
		t.Fatalf("mark agent read to first message failed: %v", err)
	}

	afterPartialRead := ConversationService.Get(conversation.ID)
	if afterPartialRead.AgentUnreadCount != 1 {
		t.Fatalf("expected agent unread 1 after partial read, got %d", afterPartialRead.AgentUnreadCount)
	}

	if _, err := MessageService.SendAgentMessage(conversation.ID, 0, "read_cursor_agent_1", enums.IMMessageTypeText, "已处理", "", agent); err != nil {
		t.Fatalf("send agent message failed: %v", err)
	}

	afterReply := ConversationService.Get(conversation.ID)
	if afterReply.AgentUnreadCount != 0 {
		t.Fatalf("expected agent unread 0 after agent reply, got %d", afterReply.AgentUnreadCount)
	}
	if afterReply.CustomerUnreadCount != 1 {
		t.Fatalf("expected customer unread 1 after agent reply, got %d", afterReply.CustomerUnreadCount)
	}
}

func TestPendingConversationRejectsAgentSendBeforeAssign(t *testing.T) {
	setupIMServiceTestDB(t)

	seedAgentUser(t, 3003, "agent_assign_gate")
	seedEnabledAIAgent(t, 407, "分配门禁AI")

	customer := &dto.AuthPrincipal{UserID: 1103, Username: "customer_assign_gate"}
	agent := &dto.AuthPrincipal{UserID: 3003, Username: "agent_assign_gate"}

	conversation, err := ConversationService.Create(enums.IMConversationChannelWebChat, "待分配禁止回复", 407, customer)
	if err != nil {
		t.Fatalf("create conversation failed: %v", err)
	}

	if _, err := MessageService.SendAgentMessage(conversation.ID, 0, "agent_gate_1", enums.IMMessageTypeText, "客服直接回复", "", agent); err == nil || !strings.Contains(err.Error(), "会话未分配客服") {
		t.Fatalf("expected pending conversation agent send to fail, got %v", err)
	}
}
