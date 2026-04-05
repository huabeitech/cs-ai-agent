package services

import (
	"encoding/json"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"cs-agent/internal/models"
	"cs-agent/internal/pkg/config"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/pkg/errorsx"
	"cs-agent/internal/pkg/openidentity"
	"cs-agent/internal/wxwork"

	"github.com/mlogclub/simple/sqls"
	"github.com/silenceper/wechat/v2/work/kf"
	"github.com/silenceper/wechat/v2/work/kf/syncmsg"
)

const (
	wxWorkKFDefaultAIAgentConfigKey = "wxwork_kf.default_ai_agent_id"
	wxWorkKFSystemOperatorName      = "wxwork_kf"
	wxWorkKFSyncMsgLimit            = 1000
)

var WxWorkKFInboundService = newWxWorkKFInboundService()

func newWxWorkKFInboundService() *wxWorkKFInboundService {
	return &wxWorkKFInboundService{}
}

type wxWorkKFInboundService struct {
}

func (s *wxWorkKFInboundService) SyncCallbackMessages(message kf.CallbackMessage) error {
	cli, err := wxwork.GetWorkCli().GetKF()
	if err != nil {
		return err
	}

	state := WxWorkKFSyncStateService.Take("open_kf_id = ?", message.OpenKfID)
	cursor := ""
	if state != nil {
		cursor = strings.TrimSpace(state.NextCursor)
	}

	for {
		result, syncErr := cli.SyncMsg(kf.SyncMsgOptions{
			Cursor:   cursor,
			Token:    message.Token,
			Limit:    wxWorkKFSyncMsgLimit,
			OpenKfID: message.OpenKfID,
		})
		if syncErr != nil {
			return syncErr
		}

		for _, item := range result.MsgList {
			if err := s.consumeSyncMessage(item); err != nil {
				slog.Error("consume wxwork kf sync message failed",
					"open_kfid", item.OpenKFID,
					"external_userid", item.ExternalUserID,
					"msg_id", item.MsgID,
					"msg_type", item.MsgType,
					"event", item.EventType,
					"error", err,
				)
			}
		}

		if err := s.saveNextCursor(message.OpenKfID, result.NextCursor); err != nil {
			return err
		}

		if result.HasMore != 1 || strings.TrimSpace(result.NextCursor) == "" {
			return nil
		}
		cursor = result.NextCursor
	}
}

func (s *wxWorkKFInboundService) consumeSyncMessage(item syncmsg.Message) error {
	msgID := strings.TrimSpace(item.MsgID)
	if msgID == "" {
		return errorsx.InvalidParam("企业微信消息ID不能为空")
	}
	if WxWorkKFMessageRefService.Take("wx_msg_id = ?", msgID) != nil {
		return nil
	}

	switch strings.TrimSpace(item.MsgType) {
	case "text":
		return s.handleTextMessage(item)
	case "image":
		return s.handleImageMessage(item)
	case "file":
		return s.handleFileMessage(item)
	case "event":
		return s.handleEventMessage(item)
	default:
		return s.handleUnsupportedMessage(item)
	}
}

func (s *wxWorkKFInboundService) handleTextMessage(item syncmsg.Message) error {
	payload := syncmsg.Text{}
	if err := json.Unmarshal(item.OriginData, &payload); err != nil {
		return err
	}

	conversation, err := s.ensureConversation(payload.BaseMessage, map[string]any{
		"msgType": payload.MsgType,
		"menuId":  payload.Text.MenuID,
	})
	if err != nil {
		return err
	}
	message, err := MessageService.SendCustomerMessage(
		config.Current(),
		conversation.ID,
		s.buildInboundClientMsgID(item.MsgID),
		enums.IMMessageTypeText,
		strings.TrimSpace(payload.Text.Content),
		"",
		s.buildExternalInfo(payload.ExternalUserID),
	)
	if err != nil {
		return err
	}
	return s.createMessageRef(conversation.ID, message.ID, item, enums.WxWorkKFMessageDirectionIn, enums.WxWorkKFMessageSendStatusReceived)
}

func (s *wxWorkKFInboundService) handleImageMessage(item syncmsg.Message) error {
	payload := syncmsg.Image{}
	if err := json.Unmarshal(item.OriginData, &payload); err != nil {
		return err
	}
	conversation, err := s.ensureConversation(payload.BaseMessage, map[string]any{
		"msgType": payload.MsgType,
		"mediaId": payload.Image.MediaID,
	})
	if err != nil {
		return err
	}
	normalizedPayload, _ := json.Marshal(map[string]any{
		"provider":   enums.ExternalSourceWxWorkKF,
		"msgId":      payload.MsgID,
		"msgType":    payload.MsgType,
		"mediaId":    strings.TrimSpace(payload.Image.MediaID),
		"openKfId":   strings.TrimSpace(payload.OpenKFID),
		"externalId": strings.TrimSpace(payload.ExternalUserID),
		"sendTime":   payload.SendTime,
		"origin":     payload.Origin,
		"rawPayload": json.RawMessage(item.OriginData),
	})
	message, err := MessageService.SendCustomerMessage(
		config.Current(),
		conversation.ID,
		s.buildInboundClientMsgID(item.MsgID),
		enums.IMMessageTypeImage,
		"[图片]",
		string(normalizedPayload),
		s.buildExternalInfo(payload.ExternalUserID),
	)
	if err != nil {
		return err
	}
	return s.createMessageRef(conversation.ID, message.ID, item, enums.WxWorkKFMessageDirectionIn, enums.WxWorkKFMessageSendStatusReceived)
}

func (s *wxWorkKFInboundService) handleFileMessage(item syncmsg.Message) error {
	payload := syncmsg.File{}
	if err := json.Unmarshal(item.OriginData, &payload); err != nil {
		return err
	}
	conversation, err := s.ensureConversation(payload.BaseMessage, map[string]any{
		"msgType": payload.MsgType,
		"mediaId": payload.File.MediaID,
	})
	if err != nil {
		return err
	}
	normalizedPayload, _ := json.Marshal(map[string]any{
		"provider":   enums.ExternalSourceWxWorkKF,
		"msgId":      payload.MsgID,
		"msgType":    payload.MsgType,
		"mediaId":    strings.TrimSpace(payload.File.MediaID),
		"openKfId":   strings.TrimSpace(payload.OpenKFID),
		"externalId": strings.TrimSpace(payload.ExternalUserID),
		"sendTime":   payload.SendTime,
		"origin":     payload.Origin,
		"rawPayload": json.RawMessage(item.OriginData),
	})
	message, err := MessageService.SendCustomerMessage(
		config.Current(),
		conversation.ID,
		s.buildInboundClientMsgID(item.MsgID),
		enums.IMMessageTypeAttachment,
		"[附件]",
		string(normalizedPayload),
		s.buildExternalInfo(payload.ExternalUserID),
	)
	if err != nil {
		return err
	}
	return s.createMessageRef(conversation.ID, message.ID, item, enums.WxWorkKFMessageDirectionIn, enums.WxWorkKFMessageSendStatusReceived)
}

func (s *wxWorkKFInboundService) handleUnsupportedMessage(item syncmsg.Message) error {
	base, err := s.parseBaseMessage(item.OriginData)
	if err != nil {
		return err
	}
	conversation, convErr := s.ensureConversation(base, map[string]any{"msgType": item.MsgType})
	if convErr != nil {
		return convErr
	}
	content := s.buildUnsupportedContent(item.MsgType)
	message, err := MessageService.SendCustomerMessage(
		config.Current(),
		conversation.ID,
		s.buildInboundClientMsgID(item.MsgID),
		enums.IMMessageTypeText,
		content,
		string(item.OriginData),
		s.buildExternalInfo(base.ExternalUserID),
	)
	if err != nil {
		return err
	}
	return s.createMessageRef(conversation.ID, message.ID, item, enums.WxWorkKFMessageDirectionIn, enums.WxWorkKFMessageSendStatusReceived)
}

func (s *wxWorkKFInboundService) handleEventMessage(item syncmsg.Message) error {
	switch strings.TrimSpace(item.EventType) {
	case enums.WxWorkKFEventTypeEnterSession:
		return s.handleEnterSessionEvent(item)
	case enums.WxWorkKFEventTypeSessionStatusChange:
		return s.handleSessionStatusChangeEvent(item)
	case enums.WxWorkKFEventTypeMsgSendFail:
		return s.handleMsgSendFailEvent(item)
	default:
		return s.recordOrphanEvent(item, "收到未处理的企业微信事件")
	}
}

func (s *wxWorkKFInboundService) handleEnterSessionEvent(item syncmsg.Message) error {
	payload := syncmsg.EnterSessionEvent{}
	if err := json.Unmarshal(item.OriginData, &payload); err != nil {
		return err
	}
	base := s.normalizeEventBaseMessage(payload.BaseMessage, payload.Event.OpenKFID, payload.Event.ExternalUserID)
	conversation, err := s.ensureConversation(base, map[string]any{
		"scene":       payload.Event.Scene,
		"sceneParam":  payload.Event.SceneParam,
		"welcomeCode": payload.Event.WelcomeCode,
	})
	if err != nil {
		return err
	}
	if err := s.createMessageRef(conversation.ID, 0, item, enums.WxWorkKFMessageDirectionIn, enums.WxWorkKFMessageSendStatusReceived); err != nil {
		return err
	}
	return s.appendConversationEvent(conversation.ID, "微信客户进入会话", string(item.OriginData))
}

func (s *wxWorkKFInboundService) handleSessionStatusChangeEvent(item syncmsg.Message) error {
	payload := syncmsg.SessionStatusChangeEvent{}
	if err := json.Unmarshal(item.OriginData, &payload); err != nil {
		return err
	}
	base := s.normalizeEventBaseMessage(payload.BaseMessage, payload.Event.OpenKFID, payload.Event.ExternalUserID)
	base.ReceptionistUserID = payload.Event.NewReceptionistUserID
	conversation, err := s.ensureConversation(base, map[string]any{
		"changeType": payload.Event.ChangeType,
		"msgCode":    payload.Event.MsgCode,
	})
	if err != nil {
		return err
	}
	sessionStatus := enums.WxWorkKFSessionStatusActive
	switch payload.Event.ChangeType {
	case 2:
		sessionStatus = enums.WxWorkKFSessionStatusTransfer
	case 3:
		sessionStatus = enums.WxWorkKFSessionStatusClosed
	}
	if err := s.upsertConversationMapping(conversation.ID, payload.Event.OpenKFID, payload.Event.ExternalUserID, payload.Event.NewReceptionistUserID, sessionStatus, payload.SendTime, payload.MsgID, string(item.OriginData)); err != nil {
		return err
	}
	if err := s.createMessageRef(conversation.ID, 0, item, enums.WxWorkKFMessageDirectionIn, enums.WxWorkKFMessageSendStatusReceived); err != nil {
		return err
	}
	return s.appendConversationEvent(conversation.ID, "微信会话状态变更", string(item.OriginData))
}

func (s *wxWorkKFInboundService) handleMsgSendFailEvent(item syncmsg.Message) error {
	payload := syncmsg.MsgSendFailEvent{}
	if err := json.Unmarshal(item.OriginData, &payload); err != nil {
		return err
	}
	base := s.normalizeEventBaseMessage(payload.BaseMessage, payload.Event.OpenKFID, payload.Event.ExternalUserID)
	conversation, err := s.ensureConversation(base, map[string]any{
		"failMsgId": payload.Event.FailMsgID,
		"failType":  payload.Event.FailType,
	})
	if err != nil {
		return err
	}
	if ref := WxWorkKFMessageRefService.Take("wx_msg_id = ?", payload.Event.FailMsgID); ref != nil {
		_ = WxWorkKFMessageRefService.Updates(ref.ID, map[string]any{
			"send_status": enums.WxWorkKFMessageSendStatusFailed,
			"fail_reason": string(item.OriginData),
			"updated_at":  time.Now(),
		})
	}
	if err := s.createMessageRef(conversation.ID, 0, item, enums.WxWorkKFMessageDirectionIn, enums.WxWorkKFMessageSendStatusReceived); err != nil {
		return err
	}
	return s.appendConversationEvent(conversation.ID, "微信消息发送失败事件", string(item.OriginData))
}

func (s *wxWorkKFInboundService) recordOrphanEvent(item syncmsg.Message, content string) error {
	base, err := s.parseBaseMessage(item.OriginData)
	if err != nil {
		return err
	}
	conversation, convErr := s.ensureConversation(base, map[string]any{"eventType": item.EventType})
	if convErr != nil {
		return convErr
	}
	if err := s.createMessageRef(conversation.ID, 0, item, enums.WxWorkKFMessageDirectionIn, enums.WxWorkKFMessageSendStatusReceived); err != nil {
		return err
	}
	return s.appendConversationEvent(conversation.ID, content, string(item.OriginData))
}

func (s *wxWorkKFInboundService) ensureConversation(base syncmsg.BaseMessage, profile map[string]any) (*models.Conversation, error) {
	externalID := strings.TrimSpace(base.ExternalUserID)
	if externalID == "" {
		return nil, errorsx.InvalidParam("企业微信客户ID不能为空")
	}

	external := s.buildExternalInfo(externalID)
	conversation := ConversationService.FindOne(sqls.NewCnd().
		Eq("external_source", external.ExternalSource).
		Eq("external_id", external.ExternalID).
		In("status", []enums.IMConversationStatus{
			enums.IMConversationStatusPending,
			enums.IMConversationStatusActive,
		}).
		Desc("id"))

	if conversation == nil {
		aiAgentID, err := s.getDefaultAIAgentID()
		if err != nil {
			return nil, err
		}
		conversation, err = ConversationService.Create(external, aiAgentID)
		if err != nil {
			return nil, err
		}
	}

	if err := s.upsertConversationMapping(
		conversation.ID,
		base.OpenKFID,
		base.ExternalUserID,
		base.ReceptionistUserID,
		enums.WxWorkKFSessionStatusActive,
		base.SendTime,
		base.MsgID,
		s.mustMarshal(profile),
	); err != nil {
		return nil, err
	}
	return conversation, nil
}

func (s *wxWorkKFInboundService) upsertConversationMapping(conversationID int64, openKfID, externalUserID, servicerUserID string, sessionStatus enums.WxWorkKFSessionStatus, sendTime uint64, lastMsgID, rawProfile string) error {
	now := time.Now()
	lastMsgTime := s.parseSendTime(sendTime)
	existing := WxWorkKFConversationService.Take("conversation_id = ?", conversationID)
	if existing != nil {
		updates := map[string]any{
			"open_kf_id":       strings.TrimSpace(openKfID),
			"external_user_id": strings.TrimSpace(externalUserID),
			"servicer_user_id": strings.TrimSpace(servicerUserID),
			"session_status":   string(sessionStatus),
			"last_wx_msg_id":   strings.TrimSpace(lastMsgID),
			"updated_at":       now,
			"status":           enums.StatusOk,
		}
		if lastMsgTime != nil {
			updates["last_wx_msg_time"] = *lastMsgTime
		}
		if strings.TrimSpace(rawProfile) != "" {
			updates["raw_profile"] = rawProfile
		}
		return WxWorkKFConversationService.Updates(existing.ID, updates)
	}

	return WxWorkKFConversationService.Create(&models.WxWorkKFConversation{
		ConversationID: conversationID,
		OpenKfID:       strings.TrimSpace(openKfID),
		ExternalUserID: strings.TrimSpace(externalUserID),
		ServicerUserID: strings.TrimSpace(servicerUserID),
		SessionStatus:  string(sessionStatus),
		LastWxMsgID:    strings.TrimSpace(lastMsgID),
		LastWxMsgTime:  lastMsgTime,
		RawProfile:     strings.TrimSpace(rawProfile),
		Status:         enums.StatusOk,
		AuditFields: models.AuditFields{
			CreatedAt:      now,
			CreateUserID:   0,
			CreateUserName: wxWorkKFSystemOperatorName,
			UpdatedAt:      now,
			UpdateUserID:   0,
			UpdateUserName: wxWorkKFSystemOperatorName,
		},
	})
}

func (s *wxWorkKFInboundService) createMessageRef(conversationID, messageID int64, item syncmsg.Message, direction enums.WxWorkKFMessageDirection, sendStatus enums.WxWorkKFMessageSendStatus) error {
	if WxWorkKFMessageRefService.Take("wx_msg_id = ?", item.MsgID) != nil {
		return nil
	}
	now := time.Now()
	return WxWorkKFMessageRefService.Create(&models.WxWorkKFMessageRef{
		ConversationID: conversationID,
		MessageID:      messageID,
		WxMsgID:        strings.TrimSpace(item.MsgID),
		Direction:      string(direction),
		Origin:         int(item.Origin),
		OpenKfID:       strings.TrimSpace(item.OpenKFID),
		ExternalUserID: strings.TrimSpace(item.ExternalUserID),
		SendStatus:     string(sendStatus),
		RawPayload:     string(item.OriginData),
		Status:         enums.StatusOk,
		AuditFields: models.AuditFields{
			CreatedAt:      now,
			CreateUserID:   0,
			CreateUserName: wxWorkKFSystemOperatorName,
			UpdatedAt:      now,
			UpdateUserID:   0,
			UpdateUserName: wxWorkKFSystemOperatorName,
		},
	})
}

func (s *wxWorkKFInboundService) saveNextCursor(openKfID, nextCursor string) error {
	openKfID = strings.TrimSpace(openKfID)
	if openKfID == "" {
		return errorsx.InvalidParam("openKfID不能为空")
	}
	now := time.Now()
	state := WxWorkKFSyncStateService.Take("open_kf_id = ?", openKfID)
	if state != nil {
		return WxWorkKFSyncStateService.Updates(state.ID, map[string]any{
			"next_cursor":  strings.TrimSpace(nextCursor),
			"last_sync_at": now,
			"updated_at":   now,
			"status":       enums.StatusOk,
		})
	}
	return WxWorkKFSyncStateService.Create(&models.WxWorkKFSyncState{
		OpenKfID:   openKfID,
		NextCursor: strings.TrimSpace(nextCursor),
		LastSyncAt: &now,
		Status:     enums.StatusOk,
		AuditFields: models.AuditFields{
			CreatedAt:      now,
			CreateUserID:   0,
			CreateUserName: wxWorkKFSystemOperatorName,
			UpdatedAt:      now,
			UpdateUserID:   0,
			UpdateUserName: wxWorkKFSystemOperatorName,
		},
	})
}

func (s *wxWorkKFInboundService) appendConversationEvent(conversationID int64, content, payload string) error {
	return ConversationEventLogService.Create(&models.ConversationEventLog{
		ConversationID: conversationID,
		EventType:      enums.IMEventTypeWxWorkKFEvent,
		OperatorType:   enums.IMSenderTypeSystem,
		OperatorID:     0,
		Content:        strings.TrimSpace(content),
		Payload:        strings.TrimSpace(payload),
		CreatedAt:      time.Now(),
	})
}

func (s *wxWorkKFInboundService) getDefaultAIAgentID() (int64, error) {
	item := SystemConfigService.Take("config_key = ? AND status = ?", wxWorkKFDefaultAIAgentConfigKey, enums.StatusOk)
	if item == nil {
		return 0, errorsx.InvalidParam("未配置企业微信客服默认AI Agent")
	}
	aiAgentID, err := strconv.ParseInt(strings.TrimSpace(item.ConfigValue), 10, 64)
	if err != nil || aiAgentID <= 0 {
		return 0, errorsx.InvalidParam("企业微信客服默认AI Agent配置不合法")
	}
	agent := AIAgentService.Get(aiAgentID)
	if agent == nil || agent.Status != enums.StatusOk {
		return 0, errorsx.InvalidParam("企业微信客服默认AI Agent不存在或已禁用")
	}
	return aiAgentID, nil
}

func (s *wxWorkKFInboundService) buildExternalInfo(externalUserID string) openidentity.ExternalInfo {
	return openidentity.ExternalInfo{
		ExternalSource: enums.ExternalSourceWxWorkKF,
		ExternalID:     strings.TrimSpace(externalUserID),
		ExternalName:   strings.TrimSpace(externalUserID),
	}
}

func (s *wxWorkKFInboundService) buildInboundClientMsgID(wxMsgID string) string {
	return "wxwork_kf:" + strings.TrimSpace(wxMsgID)
}

func (s *wxWorkKFInboundService) buildUnsupportedContent(msgType string) string {
	switch strings.TrimSpace(msgType) {
	case "voice":
		return "[语音]"
	case "video":
		return "[视频]"
	case "location":
		return "[位置]"
	case "link":
		return "[链接]"
	case "business_card":
		return "[名片]"
	case "miniprogram":
		return "[小程序]"
	default:
		return "[" + strings.TrimSpace(msgType) + "]"
	}
}

func (s *wxWorkKFInboundService) parseSendTime(sendTime uint64) *time.Time {
	if sendTime == 0 {
		return nil
	}
	t := time.Unix(int64(sendTime), 0)
	return &t
}

func (s *wxWorkKFInboundService) mustMarshal(value any) string {
	if value == nil {
		return ""
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return ""
	}
	return string(raw)
}

func (s *wxWorkKFInboundService) parseBaseMessage(raw []byte) (syncmsg.BaseMessage, error) {
	base := syncmsg.BaseMessage{}
	err := json.Unmarshal(raw, &base)
	return base, err
}

func (s *wxWorkKFInboundService) normalizeEventBaseMessage(base syncmsg.BaseMessage, openKfID, externalUserID string) syncmsg.BaseMessage {
	base.OpenKFID = strings.TrimSpace(openKfID)
	base.ExternalUserID = strings.TrimSpace(externalUserID)
	return base
}
