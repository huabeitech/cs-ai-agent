package services

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"cs-agent/internal/models"
	"cs-agent/internal/pkg/config"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/pkg/utils"
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

type wxWorkKFOutboundChunk struct {
	MessageType enums.IMMessageType
	Content     string
	AssetID     string
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
	chunks, buildErr := s.buildOutboundChunks(config.Current(), message)
	if buildErr != nil {
		return s.markOutboxFailed(outbox, buildErr.Error())
	}
	if len(chunks) == 0 {
		return s.markOutboxFailed(outbox, "当前消息无法转换为企业微信下行消息")
	}

	wxMsgIDs := make([]string, 0, len(chunks))
	for i := range chunks {
		wxMsgID, sendErr := s.sendOutboundChunk(mapping, message, chunks[i], i)
		if sendErr != nil {
			return s.markOutboxFailed(outbox, sendErr.Error())
		}
		wxMsgIDs = append(wxMsgIDs, wxMsgID)
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
			for i := range wxMsgIDs {
				rawPayload := strings.TrimSpace(outbox.Payload)
				if len(chunks) > i {
					if payload, err := json.Marshal(map[string]any{
						"messageId":    message.ID,
						"chunkIndex":   i,
						"chunkType":    chunks[i].MessageType,
						"chunkText":    strings.TrimSpace(chunks[i].Content),
						"chunkAssetId": strings.TrimSpace(chunks[i].AssetID),
					}); err == nil {
						rawPayload = string(payload)
					}
				}
				if err := repositories.WxWorkKFMessageRefRepository.Create(ctx.Tx, &models.WxWorkKFMessageRef{
					ConversationID: conversation.ID,
					MessageID:      message.ID,
					WxMsgID:        strings.TrimSpace(wxMsgIDs[i]),
					Direction:      string(enums.WxWorkKFMessageDirectionOut),
					Origin:         0,
					OpenKfID:       mapping.OpenKfID,
					ExternalUserID: mapping.ExternalUserID,
					SendStatus:     string(enums.WxWorkKFMessageSendStatusSent),
					RawPayload:     rawPayload,
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
		}
		return ConversationEventLogService.CreateEvent(ctx, conversation.ID, enums.IMEventTypeWxWorkKFOutbound, message.SenderType, message.SenderID, fmt.Sprintf("企业微信消息发送成功，共%d条", len(wxMsgIDs)), "")
	})
}

func (s *wxWorkKFOutboundService) sendOutboundChunk(mapping *models.WxWorkKFConversation, message *models.Message, chunk wxWorkKFOutboundChunk, chunkIndex int) (string, error) {
	switch chunk.MessageType {
	case enums.IMMessageTypeText:
		return s.sendTextMessage(mapping, message, chunk.Content, chunkIndex)
	case enums.IMMessageTypeImage:
		return s.sendImageMessage(mapping, message, chunk, chunkIndex)
	default:
		return "", fmt.Errorf("不支持的企业微信下行消息类型: %s", chunk.MessageType)
	}
}

func (s *wxWorkKFOutboundService) sendTextMessage(mapping *models.WxWorkKFConversation, message *models.Message, content string, chunkIndex int) (string, error) {
	cli, err := wxwork.GetWorkCli().GetKF()
	if err != nil {
		return "", err
	}

	req := sendmsg.Text{}
	req.Message.ToUser = strings.TrimSpace(mapping.ExternalUserID)
	req.Message.OpenKFID = strings.TrimSpace(mapping.OpenKfID)
	req.Message.MsgID = s.buildOutboundClientMsgID(message.ID, chunkIndex)
	req.MsgType = "text"
	req.Text.Content = strings.TrimSpace(content)

	resp, err := cli.SendMsg(req)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(resp.MsgID) == "" {
		return "", fmt.Errorf("企业微信返回的消息ID为空")
	}
	return strings.TrimSpace(resp.MsgID), nil
}

func (s *wxWorkKFOutboundService) sendImageMessage(mapping *models.WxWorkKFConversation, message *models.Message, chunk wxWorkKFOutboundChunk, chunkIndex int) (string, error) {
	cfg := config.Current()
	if cfg == nil {
		return "", fmt.Errorf("系统配置不存在")
	}
	if strings.TrimSpace(chunk.AssetID) == "" {
		return "", fmt.Errorf("图片消息缺少 assetId")
	}

	asset := AssetService.GetByAssetID(chunk.AssetID)
	if asset == nil {
		return "", fmt.Errorf("图片资源不存在")
	}
	fileReader, err := s.openAssetReader(cfg, asset)
	if err != nil {
		return "", err
	}
	defer fileReader.Close()

	materialCli := wxwork.GetWorkCli().GetMaterial()
	uploadResp, err := materialCli.UploadTempFileFromReader(asset.Filename, "image", fileReader)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(uploadResp.MediaID) == "" {
		return "", fmt.Errorf("企业微信返回的图片 media_id 为空")
	}

	kfCli, err := wxwork.GetWorkCli().GetKF()
	if err != nil {
		return "", err
	}
	req := sendmsg.Image{}
	req.Message.ToUser = strings.TrimSpace(mapping.ExternalUserID)
	req.Message.OpenKFID = strings.TrimSpace(mapping.OpenKfID)
	req.Message.MsgID = s.buildOutboundClientMsgID(message.ID, chunkIndex)
	req.MsgType = "image"
	req.Image.MediaID = strings.TrimSpace(uploadResp.MediaID)

	resp, err := kfCli.SendMsg(req)
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

func (s *wxWorkKFOutboundService) buildOutboundClientMsgID(messageID int64, chunkIndex int) string {
	return fmt.Sprintf("outbox_wxwork_kf_%d_%d", messageID, chunkIndex)
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

func (s *wxWorkKFOutboundService) buildOutboundChunks(cfg *config.Config, message *models.Message) ([]wxWorkKFOutboundChunk, error) {
	if message == nil {
		return nil, fmt.Errorf("平台消息不存在")
	}
	switch message.MessageType {
	case enums.IMMessageTypeText:
		content := strings.TrimSpace(message.Content)
		if content == "" {
			return nil, fmt.Errorf("文本消息内容为空")
		}
		return []wxWorkKFOutboundChunk{{MessageType: enums.IMMessageTypeText, Content: content}}, nil
	case enums.IMMessageTypeHTML:
		return s.buildHTMLChunks(cfg, message.Content)
	default:
		return nil, fmt.Errorf("当前暂不支持企业微信下行消息类型: %s", message.MessageType)
	}
}

func (s *wxWorkKFOutboundService) buildHTMLChunks(cfg *config.Config, content string) ([]wxWorkKFOutboundChunk, error) {
	contentChunks, err := utils.SplitHTMLContentChunks(content)
	if err != nil {
		return nil, err
	}
	chunks := make([]wxWorkKFOutboundChunk, 0, len(contentChunks))
	for _, chunk := range contentChunks {
		switch chunk.Type {
		case utils.ContentChunkTypeText:
			text := strings.TrimSpace(chunk.Content)
			if text == "" {
				continue
			}
			chunks = append(chunks, wxWorkKFOutboundChunk{
				MessageType: enums.IMMessageTypeText,
				Content:     text,
			})
		case utils.ContentChunkTypeImage:
			assetID, resolveErr := s.resolveAssetIDFromImageSrc(cfg, chunk.Content)
			if resolveErr != nil {
				chunks = append(chunks, wxWorkKFOutboundChunk{
					MessageType: enums.IMMessageTypeText,
					Content:     "[图片]",
				})
				continue
			}
			chunks = append(chunks, wxWorkKFOutboundChunk{
				MessageType: enums.IMMessageTypeImage,
				AssetID:     assetID,
			})
		}
	}
	if len(chunks) == 0 {
		return nil, fmt.Errorf("HTML 消息内容为空")
	}
	return chunks, nil
}

func (s *wxWorkKFOutboundService) resolveAssetIDFromImageSrc(cfg *config.Config, src string) (string, error) {
	if cfg == nil {
		return "", fmt.Errorf("系统配置不存在")
	}
	storageKey, err := resolveStorageKeyFromAssetURL(strings.TrimSpace(cfg.Storage.Local.BaseURL), src)
	if err != nil {
		return "", err
	}
	asset := AssetService.Take("storage_key = ?", storageKey)
	if asset == nil {
		return "", fmt.Errorf("未找到图片资源")
	}
	return strings.TrimSpace(asset.AssetID), nil
}

func resolveStorageKeyFromAssetURL(baseURL, rawURL string) (string, error) {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	rawURL = strings.TrimSpace(rawURL)
	if baseURL == "" || rawURL == "" {
		return "", fmt.Errorf("图片URL不合法")
	}
	if strings.HasPrefix(rawURL, baseURL+"/") {
		return strings.TrimLeft(strings.TrimPrefix(rawURL, baseURL), "/"), nil
	}

	baseParsed, baseErr := url.Parse(baseURL)
	rawParsed, rawErr := url.Parse(rawURL)
	if baseErr != nil || rawErr != nil {
		return "", fmt.Errorf("图片URL不合法")
	}
	if !strings.EqualFold(baseParsed.Host, rawParsed.Host) {
		return "", fmt.Errorf("图片URL不属于当前存储域名")
	}
	basePath := strings.TrimRight(baseParsed.Path, "/")
	rawPath := strings.TrimLeft(rawParsed.Path, "/")
	if basePath == "" {
		return rawPath, nil
	}
	basePath = strings.TrimLeft(basePath, "/")
	if !strings.HasPrefix(rawPath, basePath+"/") {
		return "", fmt.Errorf("图片URL不属于当前存储目录")
	}
	return strings.TrimLeft(strings.TrimPrefix(rawPath, basePath), "/"), nil
}

func (s *wxWorkKFOutboundService) openAssetReader(cfg *config.Config, asset *models.Asset) (io.ReadCloser, error) {
	if cfg == nil || asset == nil {
		return nil, fmt.Errorf("图片资源不存在")
	}
	switch asset.Provider {
	case "", enums.AssetProviderLocal:
		fullPath := filepath.Join(cfg.Storage.Local.Root, filepath.FromSlash(strings.TrimSpace(asset.StorageKey)))
		return os.Open(fullPath)
	default:
		return nil, fmt.Errorf("当前暂不支持 %s 存储的企业微信图片下发", asset.Provider)
	}
}
