package services

import (
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/pkg/errorsx"
	"cs-agent/internal/services/storage"
	"encoding/json"
	"strings"
)

type imMessageAssetPayload struct {
	AssetID  string `json:"assetId"`
	Filename string `json:"filename,omitempty"`
	FileSize int64  `json:"fileSize,omitempty"`
	MimeType string `json:"mimeType,omitempty"`
	URL      string `json:"url,omitempty"`
}

func parseIMMessageAssetPayload(payload string) (*imMessageAssetPayload, error) {
	payload = strings.TrimSpace(payload)
	if payload == "" {
		return nil, errorsx.InvalidParam("附件消息缺少 payload")
	}
	ret := &imMessageAssetPayload{}
	if err := json.Unmarshal([]byte(payload), ret); err != nil {
		return nil, errorsx.InvalidParam("附件消息 payload 格式错误")
	}
	ret.AssetID = strings.TrimSpace(ret.AssetID)
	if ret.AssetID == "" {
		return nil, errorsx.InvalidParam("附件消息缺少 assetId")
	}
	return ret, nil
}

func buildIMMessageAssetPayload(asset *models.Asset) (string, error) {
	if asset == nil {
		return "", errorsx.InvalidParam("附件不存在")
	}
	provider, err := storage.NewProvider(asset.Provider)
	if err != nil {
		return "", err
	}
	payload, err := json.Marshal(imMessageAssetPayload{
		AssetID:  asset.AssetID,
		Filename: asset.Filename,
		FileSize: asset.FileSize,
		MimeType: asset.MimeType,
		URL:      provider.GetURL(asset.StorageKey),
	})
	if err != nil {
		return "", err
	}
	return string(payload), nil
}

func validateConversationAsset(asset *models.Asset, conversationID int64, messageType enums.IMMessageType) error {
	if asset == nil {
		return errorsx.InvalidParam("附件不存在")
	}
	if asset.Status != enums.AssetStatusSuccess {
		return errorsx.InvalidParam("附件尚未上传完成")
	}
	return nil
}
