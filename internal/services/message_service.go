package services

import (
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/config"
	"cs-agent/internal/pkg/dto"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/pkg/errorsx"
	"cs-agent/internal/pkg/openidentity"
	"cs-agent/internal/repositories"
	"slices"
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

// FindByConversationIDCursor 按 id 游标分页：cursor=0 取最新 limit 条；cursor>0 取 id<cursor 的更旧消息。
// 返回的 list 已按 id 升序（时间正序）。nextCursor 为下一页请求传入的游标（本批最小 id）；hasMore 表示可能还有更旧消息。
func (s *messageService) FindByConversationIDCursor(conversationID int64, cursor int64, limit int, senderType, messageType string) (list []models.Message, nextCursor int64, hasMore bool) {
	if limit > 100 {
		limit = 100
	} else if limit <= 0 {
		limit = 20
	}
	cnd := sqls.NewCnd().Eq("conversation_id", conversationID).Limit(limit).Desc("id")
	if cursor > 0 {
		cnd.Lt("id", cursor)
	}
	if strs.IsNotBlank(senderType) {
		cnd.Eq("sender_type", senderType)
	}
	if strs.IsNotBlank(messageType) {
		cnd.Eq("message_type", messageType)
	}
	list = s.Find(cnd)
	nextCursor = cursor
	hasMore = false
	if len(list) > 0 {
		nextCursor = list[len(list)-1].ID
		hasMore = len(list) == limit
	}
	slices.Reverse(list)
	return list, nextCursor, hasMore
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

// FindPageByCndForImListAscending 与 FindPageByCnd 相同分页条件，将结果按 seq 升序排列（开放 IM 时间正序展示）。
func (s *messageService) FindPageByCndForImListAscending(cnd *sqls.Cnd) (list []models.Message, paging *sqls.Paging) {
	list, paging = s.FindPageByCnd(cnd)
	if len(list) <= 1 {
		return list, paging
	}
	for i, j := 0, len(list)-1; i < j; i, j = i+1, j-1 {
		list[i], list[j] = list[j], list[i]
	}
	return list, paging
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

func (s *messageService) SendMessage(cfg *config.Config, conversationID int64, senderType enums.IMSenderType, reqSenderID int64, clientMsgID string, messageType enums.IMMessageType, content, payload string, operator *dto.AuthPrincipal, external *openidentity.ExternalInfo) (*models.Message, error) {
	switch senderType {
	case enums.IMSenderTypeAgent:
		return s.sendMessage(cfg, conversationID, enums.IMSenderTypeAgent, reqSenderID, clientMsgID, messageType, content, payload, operator, nil)
	case enums.IMSenderTypeAI:
		return s.sendMessage(cfg, conversationID, enums.IMSenderTypeAI, reqSenderID, clientMsgID, messageType, content, payload, operator, nil)
	case enums.IMSenderTypeCustomer:
		return s.sendMessage(cfg, conversationID, enums.IMSenderTypeCustomer, 0, clientMsgID, messageType, content, payload, nil, external)
	default:
		return nil, errorsx.InvalidParam("不支持的发送人类型")
	}
}

func (s *messageService) SendAgentMessage(cfg *config.Config, conversationID int64, reqSenderID int64, clientMsgID string, messageType enums.IMMessageType, content, payload string, operator *dto.AuthPrincipal) (*models.Message, error) {
	return s.sendMessage(cfg, conversationID, enums.IMSenderTypeAgent, reqSenderID, clientMsgID, messageType, content, payload, operator, nil)
}

func (s *messageService) SendAIMessage(cfg *config.Config, conversationID int64, aiAgentID int64, clientMsgID string, messageType enums.IMMessageType, content, payload string, operator *dto.AuthPrincipal) (*models.Message, error) {
	return s.sendMessage(cfg, conversationID, enums.IMSenderTypeAI, aiAgentID, clientMsgID, messageType, content, payload, operator, nil)
}

func (s *messageService) SendCustomerMessage(cfg *config.Config, conversationID int64, clientMsgID string, messageType enums.IMMessageType, content, payload string, external openidentity.ExternalInfo) (*models.Message, error) {
	ext := external
	return s.sendMessage(cfg, conversationID, enums.IMSenderTypeCustomer, 0, clientMsgID, messageType, content, payload, nil, &ext)
}

func (s *messageService) sendMessage(cfg *config.Config, conversationID int64, senderType enums.IMSenderType, reqSenderID int64, clientMsgID string, messageType enums.IMMessageType, content, payload string, operator *dto.AuthPrincipal, external *openidentity.ExternalInfo) (*models.Message, error) {
	if senderType == enums.IMSenderTypeCustomer {
		if external == nil || strings.TrimSpace(external.ExternalID) == "" {
			return nil, errorsx.Unauthorized("外部用户标识不能为空")
		}
	} else if operator == nil {
		return nil, errorsx.Unauthorized("未登录或登录已过期")
	}
	if conversationID <= 0 {
		return nil, errorsx.InvalidParam("会话不能为空")
	}
	if strs.IsBlank(string(messageType)) {
		messageType = enums.IMMessageTypeText
	}
	conversation, err := s.ValidateConversationSender(conversationID, senderType, operator, external)
	if err != nil {
		return nil, err
	}

	var summary string
	content, payload, summary, err = s.normalizeMessageContent(cfg, conversationID, messageType, content, payload)
	if err != nil {
		return nil, err
	}
	clientMsgID = strings.TrimSpace(clientMsgID)
	if content == "" && payload == "" {
		return nil, errorsx.InvalidParam("消息内容不能为空")
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
		auditUserID := int64(0)
		auditUserName := ""
		if operator != nil {
			auditUserID = operator.UserID
			auditUserName = operator.Username
		}
		if senderType == enums.IMSenderTypeCustomer && external != nil {
			auditUserID = 0
			auditUserName = displayExternalName(external)
		}
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
				CreateUserID:   auditUserID,
				CreateUserName: auditUserName,
				UpdatedAt:      now,
				UpdateUserID:   auditUserID,
				UpdateUserName: auditUserName,
			},
		}
		switch senderType {
		case enums.IMSenderTypeAgent:
			if message.SenderID == 0 {
				message.SenderID = operator.UserID
			}
		case enums.IMSenderTypeAI:
			if message.SenderID == 0 {
				message.SenderID = reqSenderID
			}
		default:
			message.SenderID = 0
		}
		if err := ctx.Tx.Create(message).Error; err != nil {
			return err
		}

		readStateType := senderType
		if senderType == enums.IMSenderTypeAI {
			readStateType = enums.IMSenderTypeAgent
		}
		if readStateType == enums.IMSenderTypeAgent {
			if _, err := ConversationReadStateService.MarkAgentRead(ctx, conversation, operator, message, now); err != nil {
				return err
			}
		} else {
			if _, err := ConversationReadStateService.MarkCustomerRead(ctx, conversation, external, message, now); err != nil {
				return err
			}
		}
		agentReadState, customerReadState := ConversationReadStateService.getConversationReadStates(ctx.Tx, conversationID)
		agentUnreadCount, err := ConversationReadStateService.CountUnreadMessages(ctx, conversationID, readSeqNo(agentReadState), enums.IMSenderTypeCustomer)
		if err != nil {
			return err
		}
		customerUnreadCount, err := ConversationReadStateService.CountUnreadMessages(ctx, conversationID, readSeqNo(customerReadState), enums.IMSenderTypeAgent, enums.IMSenderTypeAI)
		if err != nil {
			return err
		}

		updateUserID := int64(0)
		updateUserName := ""
		if operator != nil {
			updateUserID = operator.UserID
			updateUserName = operator.Username
		}
		if senderType == enums.IMSenderTypeCustomer && external != nil {
			updateUserID = 0
			updateUserName = displayExternalName(external)
		}
		eventOperatorID := int64(0)
		if operator != nil {
			eventOperatorID = operator.UserID
		}
		eventOperatorType := senderType
		eventContent := enums.GetIMSenderTypeLabel(senderType) + "发送消息"
		if err := repositories.ConversationRepository.Updates(ctx.Tx, conversationID, map[string]any{
			"last_message_id":       message.ID,
			"last_message_at":       now,
			"last_active_at":        now,
			"last_message_summary":  limitText(summary, 255),
			"update_user_id":        updateUserID,
			"update_user_name":      updateUserName,
			"updated_at":            now,
			"agent_unread_count":    agentUnreadCount,
			"customer_unread_count": customerUnreadCount,
		}); err != nil {
			return err
		}
		if err := ConversationEventLogService.CreateEvent(ctx, conversationID, enums.IMEventTypeMessageSend, eventOperatorType, eventOperatorID, eventContent, ""); err != nil {
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
		if senderType == enums.IMSenderTypeCustomer && ret != nil {
			AIReplyService.TriggerReplyAsync(ret.ID)
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
	case enums.IMMessageTypeAttachment:
		return "[附件]"
	case enums.IMMessageTypeHTML:
		return buildHTMLSummary(content)
	case "":
		return ""
	default:
		return "[" + string(messageType) + "]"
	}
}

func (s *messageService) normalizeMessageContent(cfg *config.Config, conversationID int64, messageType enums.IMMessageType, content, payload string) (string, string, string, error) {
	switch messageType {
	case enums.IMMessageTypeHTML:
		sanitized := sanitizeMessageHTML(content)
		summary := buildHTMLSummary(sanitized)
		if summary == "" {
			return "", "", "", errorsx.InvalidParam("消息内容不能为空")
		}
		return sanitized, "", summary, nil
	case enums.IMMessageTypeImage, enums.IMMessageTypeAttachment:
		assetPayload, err := parseIMMessageAssetPayload(payload)
		if err != nil {
			return "", "", "", err
		}
		asset := AssetService.GetByAssetID(assetPayload.AssetID)
		if err := validateConversationAsset(asset, conversationID, messageType); err != nil {
			return "", "", "", err
		}
		canonicalPayload, err := buildIMMessageAssetPayload(cfg, asset)
		if err != nil {
			return "", "", "", err
		}
		summary := "[附件]"
		if messageType == enums.IMMessageTypeImage {
			summary = "[图片]"
		}
		content = strings.TrimSpace(asset.Filename)
		return content, canonicalPayload, summary + suffixFilenameForSummary(asset.Filename), nil
	default:
		content = strings.TrimSpace(content)
		if content == "" && strings.TrimSpace(payload) == "" {
			return "", "", "", errorsx.InvalidParam("消息内容不能为空")
		}
		return content, strings.TrimSpace(payload), buildMessageSummary(messageType, content), nil
	}
}

func (s *messageService) ValidateConversationSender(conversationID int64, senderType enums.IMSenderType, operator *dto.AuthPrincipal, external *openidentity.ExternalInfo) (*models.Conversation, error) {
	conversation := ConversationService.Get(conversationID)
	if conversation == nil {
		return nil, errorsx.InvalidParam("会话不存在")
	}
	if conversation.Status == enums.IMConversationStatusClosed {
		return nil, errorsx.InvalidParam("会话已关闭")
	}
	switch senderType {
	case enums.IMSenderTypeAgent:
		if operator == nil {
			return nil, errorsx.Unauthorized("未登录或登录已过期")
		}
		if conversation.Status != enums.IMConversationStatusActive || conversation.CurrentAssigneeID == 0 {
			return nil, errorsx.InvalidParam("会话未分配客服，暂不允许发送消息")
		}
		if conversation.CurrentAssigneeID != operator.UserID {
			return nil, errorsx.Forbidden("当前会话已分配给其他客服")
		}
	case enums.IMSenderTypeAI:
		if operator == nil {
			return nil, errorsx.Unauthorized("未登录或登录已过期")
		}
		if conversation.CurrentAssigneeID != 0 {
			return nil, errorsx.Forbidden("当前会话已由人工客服接管")
		}
	case enums.IMSenderTypeCustomer:
		if external == nil || !ConversationService.IsCustomerConversationOwner(conversation, *external) {
			return nil, errorsx.Forbidden("无权访问该会话")
		}
	default:
		return nil, errorsx.InvalidParam("不支持的发送人类型")
	}
	return conversation, nil
}

func suffixFilenameForSummary(filename string) string {
	filename = strings.TrimSpace(filename)
	if filename == "" {
		return ""
	}
	return " " + filename
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
