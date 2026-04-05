package services

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"cs-agent/internal/models"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/repositories"
	"cs-agent/internal/wxwork"

	"github.com/mlogclub/simple/sqls"
	"github.com/silenceper/wechat/v2/work/kf/sendmsg"
)

const (
	wxWorkKFOutboxBatchSize = 20
	wxWorkKFOutboxMaxRetry  = 6
)

var WxWorkKFOutboundService = newWxWorkKFOutboundService()

func newWxWorkKFOutboundService() *wxWorkKFOutboundService {
	return &wxWorkKFOutboundService{}
}

type wxWorkKFOutboundService struct {
}

func (s *wxWorkKFOutboundService) DispatchPendingOutbox(limit int) (int, error) {
	if !wxwork.Enabled() {
		return 0, nil
	}
	if limit <= 0 {
		limit = wxWorkKFOutboxBatchSize
	}

	items := ChannelMessageOutboxService.ListPending(enums.ChannelTypeWxWorkKF, limit)
	if len(items) == 0 {
		return 0, nil
	}

	successCount := 0
	for i := range items {
		if err := s.processOutbox(items[i].ID); err != nil {
			slog.Warn("process wxwork kf outbox failed",
				"outbox_id", items[i].ID,
				"message_id", items[i].MessageID,
				"error", err,
			)
			continue
		}
		successCount++
	}
	return successCount, nil
}

func (s *wxWorkKFOutboundService) processOutbox(outboxID int64) error {
	outbox := ChannelMessageOutboxService.Get(outboxID)
	if outbox == nil {
		return nil
	}
	if outbox.ChannelType != enums.ChannelTypeWxWorkKF {
		return nil
	}
	if outbox.SendStatus == string(enums.ChannelMessageOutboxStatusSent) {
		return nil
	}
	if outbox.NextRetryAt != nil && outbox.NextRetryAt.After(time.Now()) {
		return nil
	}

	now := time.Now()
	if err := ChannelMessageOutboxService.Updates(outbox.ID, map[string]any{
		"send_status":      string(enums.ChannelMessageOutboxStatusSending),
		"updated_at":       now,
		"update_user_id":   outbox.UpdateUserID,
		"update_user_name": outbox.UpdateUserName,
	}); err != nil {
		return err
	}

	message := MessageService.Get(outbox.MessageID)
	if message == nil {
		return s.markOutboxFailed(outbox, "平台消息不存在")
	}
	conversation := ConversationService.Get(outbox.ConversationID)
	if conversation == nil {
		return s.markOutboxFailed(outbox, "平台会话不存在")
	}
	mapping := WxWorkKFConversationService.Take("conversation_id = ?", conversation.ID)
	if mapping == nil {
		return s.markOutboxFailed(outbox, "企业微信会话映射不存在")
	}
	if mapping.ChannelID <= 0 {
		return s.markOutboxFailed(outbox, "企业微信会话映射缺少渠道ID")
	}
	channel := ChannelService.Get(mapping.ChannelID)
	if channel == nil || channel.Status != enums.StatusOk || channel.ChannelType != enums.ChannelTypeWxWorkKF {
		return s.markOutboxFailed(outbox, "企业微信接入渠道不存在、未启用或类型不匹配")
	}
	if strings.TrimSpace(mapping.OpenKfID) == "" || strings.TrimSpace(mapping.ExternalUserID) == "" {
		return s.markOutboxFailed(outbox, "企业微信会话映射缺少发送必要参数")
	}
	if message.MessageType != enums.IMMessageTypeText {
		return s.markOutboxFailed(outbox, "当前仅支持企业微信文本消息下行")
	}

	wxMsgID, sendErr := s.sendTextMessage(mapping, message)
	if sendErr != nil {
		return s.markOutboxFailed(outbox, sendErr.Error())
	}

	return sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		now = time.Now()
		if err := repositories.ChannelMessageOutboxRepository.Updates(ctx.Tx, outbox.ID, map[string]any{
			"send_status":      string(enums.ChannelMessageOutboxStatusSent),
			"sent_at":          now,
			"last_error":       "",
			"updated_at":       now,
			"update_user_id":   outbox.UpdateUserID,
			"update_user_name": outbox.UpdateUserName,
		}); err != nil {
			return err
		}
		if existing := repositories.WxWorkKFMessageRefRepository.Take(ctx.Tx, "message_id = ? AND direction = ?", message.ID, string(enums.WxWorkKFMessageDirectionOut)); existing == nil {
			if err := repositories.WxWorkKFMessageRefRepository.Create(ctx.Tx, &models.WxWorkKFMessageRef{
				ConversationID: conversation.ID,
				MessageID:      message.ID,
				WxMsgID:        strings.TrimSpace(wxMsgID),
				Direction:      string(enums.WxWorkKFMessageDirectionOut),
				Origin:         0,
				OpenKfID:       mapping.OpenKfID,
				ExternalUserID: mapping.ExternalUserID,
				SendStatus:     string(enums.WxWorkKFMessageSendStatusSent),
				RawPayload:     strings.TrimSpace(outbox.Payload),
				Status:         enums.StatusOk,
				AuditFields: models.AuditFields{
					CreatedAt:      now,
					CreateUserID:   outbox.UpdateUserID,
					CreateUserName: outbox.UpdateUserName,
					UpdatedAt:      now,
					UpdateUserID:   outbox.UpdateUserID,
					UpdateUserName: outbox.UpdateUserName,
				},
			}); err != nil {
				return err
			}
		}
		return ConversationEventLogService.CreateEvent(ctx, conversation.ID, enums.IMEventTypeWxWorkKFOutbound, enums.IMSenderTypeAgent, message.SenderID, "企业微信文本消息发送成功", "")
	})
}

func (s *wxWorkKFOutboundService) sendTextMessage(mapping *models.WxWorkKFConversation, message *models.Message) (string, error) {
	cli, err := wxwork.GetWorkCli().GetKF()
	if err != nil {
		return "", err
	}

	req := sendmsg.Text{}
	req.Message.ToUser = strings.TrimSpace(mapping.ExternalUserID)
	req.Message.OpenKFID = strings.TrimSpace(mapping.OpenKfID)
	req.Message.MsgID = s.buildOutboundClientMsgID(message.ID)
	req.MsgType = "text"
	req.Text.Content = strings.TrimSpace(message.Content)

	resp, err := cli.SendMsg(req)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(resp.MsgID) == "" {
		return "", fmt.Errorf("企业微信返回的消息ID为空")
	}
	return strings.TrimSpace(resp.MsgID), nil
}

func (s *wxWorkKFOutboundService) markOutboxFailed(outbox *models.ChannelMessageOutbox, errMsg string) error {
	if outbox == nil {
		return nil
	}
	now := time.Now()
	retryCount := outbox.RetryCount + 1
	nextRetryAt := s.nextRetryAt(retryCount)
	status := string(enums.ChannelMessageOutboxStatusFailed)
	if retryCount >= wxWorkKFOutboxMaxRetry {
		nextRetryAt = nil
	}
	return ChannelMessageOutboxService.Updates(outbox.ID, map[string]any{
		"send_status":      status,
		"retry_count":      retryCount,
		"next_retry_at":    nextRetryAt,
		"last_error":       strings.TrimSpace(errMsg),
		"updated_at":       now,
		"update_user_id":   outbox.UpdateUserID,
		"update_user_name": outbox.UpdateUserName,
	})
}

func (s *wxWorkKFOutboundService) nextRetryAt(retryCount int) *time.Time {
	delay := time.Minute
	switch {
	case retryCount <= 1:
		delay = 30 * time.Second
	case retryCount == 2:
		delay = time.Minute
	case retryCount == 3:
		delay = 2 * time.Minute
	default:
		delay = 5 * time.Minute
	}
	t := time.Now().Add(delay)
	return &t
}

func (s *wxWorkKFOutboundService) buildOutboundClientMsgID(messageID int64) string {
	return fmt.Sprintf("outbox_wxwork_kf_%d", messageID)
}

type wxWorkKFOutboundPayload struct {
	ConversationID int64               `json:"conversationId"`
	MessageID      int64               `json:"messageId"`
	MessageType    enums.IMMessageType `json:"messageType"`
	Content        string              `json:"content"`
	Payload        string              `json:"payload"`
	SenderID       int64               `json:"senderId"`
}

func (s *wxWorkKFOutboundService) parseOutboxPayload(raw string) (*wxWorkKFOutboundPayload, error) {
	payload := &wxWorkKFOutboundPayload{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(raw)), payload); err != nil {
		return nil, err
	}
	return payload, nil
}
