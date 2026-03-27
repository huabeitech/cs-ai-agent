package services

import (
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/dto"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/pkg/errorsx"
	"cs-agent/internal/repositories"
	"strings"
	"time"

	"github.com/microcosm-cc/bluemonday"
	"github.com/mlogclub/simple/common/strs"
	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
	"golang.org/x/net/html"
)

var MessageService = newMessageService()

func newMessageService() *messageService {
	return &messageService{}
}

type messageService struct {
}

func (s *messageService) Get(id int64) *models.Message {
	return repositories.MessageRepository.Get(sqls.DB(), id)
}

func (s *messageService) Take(where ...interface{}) *models.Message {
	return repositories.MessageRepository.Take(sqls.DB(), where...)
}

func (s *messageService) Find(cnd *sqls.Cnd) []models.Message {
	return repositories.MessageRepository.Find(sqls.DB(), cnd)
}

func (s *messageService) FindOne(cnd *sqls.Cnd) *models.Message {
	return repositories.MessageRepository.FindOne(sqls.DB(), cnd)
}

func (s *messageService) FindPageByParams(params *params.QueryParams) (list []models.Message, paging *sqls.Paging) {
	return repositories.MessageRepository.FindPageByParams(sqls.DB(), params)
}

func (s *messageService) FindPageByCnd(cnd *sqls.Cnd) (list []models.Message, paging *sqls.Paging) {
	return repositories.MessageRepository.FindPageByCnd(sqls.DB(), cnd)
}

func (s *messageService) Count(cnd *sqls.Cnd) int64 {
	return repositories.MessageRepository.Count(sqls.DB(), cnd)
}

func (s *messageService) Create(t *models.Message) error {
	return repositories.MessageRepository.Create(sqls.DB(), t)
}

func (s *messageService) Update(t *models.Message) error {
	return repositories.MessageRepository.Update(sqls.DB(), t)
}

func (s *messageService) Updates(id int64, columns map[string]interface{}) error {
	return repositories.MessageRepository.Updates(sqls.DB(), id, columns)
}

func (s *messageService) UpdateColumn(id int64, name string, value interface{}) error {
	return repositories.MessageRepository.UpdateColumn(sqls.DB(), id, name, value)
}

func (s *messageService) Delete(id int64) {
	repositories.MessageRepository.Delete(sqls.DB(), id)
}

func (s *messageService) GetConversationReadTarget(conversationID, messageID int64) (*models.Message, error) {
	if messageID > 0 {
		message := s.Get(messageID)
		if message == nil || message.ConversationID != conversationID {
			return nil, errorsx.InvalidParam("消息不存在")
		}
		return message, nil
	}
	return s.FindOne(sqls.NewCnd().Eq("conversation_id", conversationID).Desc("seq_no").Desc("id")), nil
}

func (s *messageService) SendMessage(conversationID int64, senderType enums.IMSenderType, reqSenderID int64, clientMsgID string, messageType enums.IMMessageType, content, payload string, operator *dto.AuthPrincipal) (*models.Message, error) {
	switch senderType {
	case enums.IMSenderTypeAgent:
		return s.sendMessage(conversationID, enums.IMSenderTypeAgent, reqSenderID, clientMsgID, messageType, content, payload, operator)
	case enums.IMSenderTypeAI:
		return s.sendMessage(conversationID, enums.IMSenderTypeAI, reqSenderID, clientMsgID, messageType, content, payload, operator)
	case enums.IMSenderTypeCustomer:
		return s.sendMessage(conversationID, enums.IMSenderTypeCustomer, 0, clientMsgID, messageType, content, payload, operator)
	default:
		return nil, errorsx.InvalidParam("不支持的发送人类型")
	}
}

func (s *messageService) SendAgentMessage(conversationID int64, reqSenderID int64, clientMsgID string, messageType enums.IMMessageType, content, payload string, operator *dto.AuthPrincipal) (*models.Message, error) {
	return s.sendMessage(conversationID, enums.IMSenderTypeAgent, reqSenderID, clientMsgID, messageType, content, payload, operator)
}

func (s *messageService) SendAIMessage(conversationID int64, aiAgentID int64, clientMsgID string, messageType enums.IMMessageType, content, payload string, operator *dto.AuthPrincipal) (*models.Message, error) {
	return s.sendMessage(conversationID, enums.IMSenderTypeAI, aiAgentID, clientMsgID, messageType, content, payload, operator)
}

func (s *messageService) SendCustomerMessage(conversationID int64, clientMsgID string, messageType enums.IMMessageType, content, payload string, operator *dto.AuthPrincipal) (*models.Message, error) {
	return s.sendMessage(conversationID, enums.IMSenderTypeCustomer, 0, clientMsgID, messageType, content, payload, operator)
}

func (s *messageService) sendMessage(conversationID int64, senderType enums.IMSenderType, reqSenderID int64, clientMsgID string, messageType enums.IMMessageType, content, payload string, operator *dto.AuthPrincipal) (*models.Message, error) {
	if operator == nil {
		return nil, errorsx.Unauthorized("未登录或登录已过期")
	}
	if conversationID <= 0 {
		return nil, errorsx.InvalidParam("会话不能为空")
	}
	if strs.IsBlank(string(messageType)) {
		messageType = enums.IMMessageTypeText
	}
	var summary string
	var err error
	content, summary, err = normalizeMessageContent(messageType, content, payload)
	if err != nil {
		return nil, err
	}
	payload = strings.TrimSpace(payload)
	clientMsgID = strings.TrimSpace(clientMsgID)
	if content == "" && payload == "" {
		return nil, errorsx.InvalidParam("消息内容不能为空")
	}

	conversation := ConversationService.Get(conversationID)
	if conversation == nil {
		return nil, errorsx.InvalidParam("会话不存在")
	}

	if conversation.Status == enums.IMConversationStatusClosed {
		return nil, errorsx.InvalidParam("会话已关闭")
	}

	switch senderType {
	case enums.IMSenderTypeAgent:
		if conversation.Status != enums.IMConversationStatusActive {
			return nil, errorsx.InvalidParam("会话未分配客服，暂不允许发送消息")
		}
		if conversation.CurrentAssigneeID == 0 {
			return nil, errorsx.InvalidParam("会话未分配客服，暂不允许发送消息")
		}
		if conversation.CurrentAssigneeID != 0 && conversation.CurrentAssigneeID != operator.UserID {
			return nil, errorsx.Forbidden("当前会话已分配给其他客服")
		}
	case enums.IMSenderTypeAI:
		if conversation.CurrentAssigneeID != 0 {
			return nil, errorsx.Forbidden("当前会话已由人工客服接管")
		}
	case enums.IMSenderTypeCustomer:
		if !ConversationService.IsCustomerConversationOwner(conversation, operator) {
			return nil, errorsx.Forbidden("无权访问该会话")
		}
	default:
		return nil, errorsx.InvalidParam("不支持的发送人类型")
	}

	var ret *models.Message
	err = sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		if clientMsgID != "" {
			existing := &models.Message{}
			if err := ctx.Tx.Where("conversation_id = ? AND client_msg_id = ?", conversationID, clientMsgID).First(existing).Error; err == nil {
				ret = existing
				return nil
			}
		}

		var last models.Message
		nextSeq := int64(1)
		if err := ctx.Tx.Where("conversation_id = ?", conversationID).Order("seq_no DESC").Limit(1).Take(&last).Error; err == nil {
			nextSeq = last.SeqNo + 1
		}

		now := time.Now()
		message := &models.Message{
			ConversationID: conversationID,
			ClientMsgID:    clientMsgID,
			SenderType:     senderType,
			SenderID:       reqSenderID,
			MessageType:    messageType,
			Content:        content,
			Payload:        payload,
			SeqNo:          nextSeq,
			SendStatus:     int(enums.IMMessageStatusSent),
			SentAt:         &now,
			AuditFields: models.AuditFields{
				CreatedAt:      now,
				CreateUserID:   operator.UserID,
				CreateUserName: operator.Username,
				UpdatedAt:      now,
				UpdateUserID:   operator.UserID,
				UpdateUserName: operator.Username,
			},
		}
		if senderType == enums.IMSenderTypeAgent {
			if message.SenderID == 0 {
				message.SenderID = operator.UserID
			}
		} else if senderType == enums.IMSenderTypeAI {
			if message.SenderID == 0 {
				message.SenderID = reqSenderID
			}
		} else if operator.IsVisitor {
			message.SenderID = 0
			message.AuditFields.CreateUserName = operator.Nickname
			message.AuditFields.UpdateUserName = operator.Nickname
		} else {
			message.SenderID = operator.UserID
		}
		if err := ctx.Tx.Create(message).Error; err != nil {
			return err
		}

		readStateType := senderType
		if senderType == enums.IMSenderTypeAI {
			readStateType = enums.IMSenderTypeAgent
		}
		if _, err := ConversationReadStateService.MarkReadTx(ctx, conversation, readStateType, operator, message, now); err != nil {
			return err
		}
		agentReadState, customerReadState, err := ConversationReadStateService.GetConversationReadStatesTx(ctx, conversationID)
		if err != nil {
			return err
		}
		agentUnreadCount, err := ConversationReadStateService.CountUnreadMessagesTx(ctx, conversationID, readSeqNo(agentReadState), enums.IMSenderTypeCustomer)
		if err != nil {
			return err
		}
		customerUnreadCount, err := ConversationReadStateService.CountUnreadMessagesTx(ctx, conversationID, readSeqNo(customerReadState), enums.IMSenderTypeAgent, enums.IMSenderTypeAI)
		if err != nil {
			return err
		}

		updateUserName := operator.Username
		if senderType == enums.IMSenderTypeCustomer && operator.IsVisitor {
			updateUserName = operator.Nickname
		}
		eventOperatorType := senderType
		eventContent := enums.GetIMSenderTypeLabel(senderType) + "发送消息"
		if err := ctx.Tx.Model(&models.Conversation{}).Where("id = ?", conversationID).Updates(map[string]any{
			"last_message_id":       message.ID,
			"last_message_at":       now,
			"last_active_at":        now,
			"last_message_summary":  limitText(summary, 255),
			"update_user_id":        operator.UserID,
			"update_user_name":      updateUserName,
			"updated_at":            now,
			"agent_unread_count":    agentUnreadCount,
			"customer_unread_count": customerUnreadCount,
		}).Error; err != nil {
			return err
		}
		if err := ConversationEventLogService.CreateEvent(ctx, conversationID, enums.IMEventTypeMessageSend, eventOperatorType, operator.UserID, eventContent, "", now); err != nil {
			return err
		}
		ret = message
		return nil
	})
	if err == nil {
		if conversation := ConversationService.Get(conversationID); conversation != nil {
			WsService.PublishMessageCreated(conversation, ret)
			WsService.PublishConversationChanged(conversation, enums.IMRealtimeEventConversationUpdated)
		}
	}
	return ret, err
}

func limitText(value string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	value = strings.TrimSpace(value)
	runes := []rune(value)
	if len(runes) <= maxLen {
		return value
	}
	return string(runes[:maxLen])
}

func buildMessageSummary(messageType enums.IMMessageType, content string) string {
	content = strings.TrimSpace(content)
	if content != "" {
		return content
	}
	switch messageType {
	case enums.IMMessageTypeImage:
		return "[图片]"
	case enums.IMMessageTypeHTML:
		return buildHTMLSummary(content)
	case "":
		return ""
	default:
		return "[" + string(messageType) + "]"
	}
}

func normalizeMessageContent(messageType enums.IMMessageType, content, payload string) (string, string, error) {
	switch messageType {
	case enums.IMMessageTypeHTML:
		sanitized := sanitizeMessageHTML(content)
		summary := buildHTMLSummary(sanitized)
		if summary == "" {
			return "", "", errorsx.InvalidParam("消息内容不能为空")
		}
		return sanitized, summary, nil
	default:
		content = strings.TrimSpace(content)
		if content == "" && strings.TrimSpace(payload) == "" {
			return "", "", errorsx.InvalidParam("消息内容不能为空")
		}
		return content, buildMessageSummary(messageType, content), nil
	}
}

func readSeqNo(state *models.ConversationReadState) int64 {
	if state == nil {
		return 0
	}
	return state.LastReadSeqNo
}

func sanitizeMessageHTML(content string) string {
	policy := bluemonday.UGCPolicy()
	policy.AllowElements("img")
	policy.AllowAttrs("src", "alt", "title").OnElements("img")
	policy.AllowURLSchemes("http", "https")
	policy.AllowStandardURLs()
	policy.AllowElements("p", "br")
	return strings.TrimSpace(policy.Sanitize(content))
}

func buildHTMLSummary(content string) string {
	if strings.TrimSpace(content) == "" {
		return ""
	}
	doc, err := html.Parse(strings.NewReader("<div>" + content + "</div>"))
	if err != nil {
		return strings.TrimSpace(content)
	}
	parts := make([]string, 0, 8)
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node == nil {
			return
		}
		if node.Type == html.TextNode {
			text := strings.TrimSpace(node.Data)
			if text != "" {
				parts = append(parts, text)
			}
		}
		if node.Type == html.ElementNode && node.Data == "img" {
			parts = append(parts, "[图片]")
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(doc)
	return strings.TrimSpace(strings.Join(parts, " "))
}
