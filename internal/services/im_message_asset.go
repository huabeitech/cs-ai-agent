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
	AssetID    string              `json:"assetId"`
	Provider   enums.AssetProvider `json:"provider,omitempty"`
	StorageKey string              `json:"storageKey,omitempty"`
	Filename   string              `json:"filename,omitempty"`
	FileSize   int64               `json:"fileSize,omitempty"`
	MimeType   string              `json:"mimeType,omitempty"`
	URL        string              `json:"url,omitempty"`
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
	ret.Provider = enums.AssetProvider(strings.TrimSpace(string(ret.Provider)))
	ret.StorageKey = strings.TrimSpace(ret.StorageKey)
	if ret.AssetID == "" {
		return nil, errorsx.InvalidParam("附件消息缺少 assetId")
	}
	return ret, nil
}

func buildIMMessageAssetPayload(asset *models.Asset) (string, error) {
	if asset == nil {
		return "", errorsx.InvalidParam("附件不存在")
	}
	payload, err := json.Marshal(imMessageAssetPayload{
		AssetID:    asset.AssetID,
		Provider:   asset.Provider,
		StorageKey: asset.StorageKey,
		Filename:   asset.Filename,
		FileSize:   asset.FileSize,
		MimeType:   asset.MimeType,
	})
	if err != nil {
		return "", err
	}
	return string(payload), nil
}

func buildIMMessageAssetPayloadForResponse(payload string) string {
	assetPayload, err := parseIMMessageAssetPayload(payload)
	if err != nil {
		return strings.TrimSpace(payload)
	}
	assetPayload = hydrateIMMessageAssetPayload(assetPayload)
	if assetPayload.Provider != "" && assetPayload.StorageKey != "" {
		if provider, err := storage.NewProvider(assetPayload.Provider); err == nil {
			assetPayload.URL = provider.GetSignedURL(assetPayload.StorageKey)
		}
	}
	data, err := json.Marshal(assetPayload)
	if err != nil {
		return strings.TrimSpace(payload)
	}
	return string(data)
}

func hydrateIMMessageAssetPayload(payload *imMessageAssetPayload) *imMessageAssetPayload {
	if payload == nil {
		return nil
	}
	if payload.Provider != "" && payload.StorageKey != "" {
		return payload
	}
	if payload.AssetID == "" {
		return payload
	}
	asset := AssetService.GetByAssetID(payload.AssetID)
	if asset == nil {
		return payload
	}
	if payload.Provider == "" {
		payload.Provider = asset.Provider
	}
	if payload.StorageKey == "" {
		payload.StorageKey = strings.TrimSpace(asset.StorageKey)
	}
	if payload.Filename == "" {
		payload.Filename = strings.TrimSpace(asset.Filename)
	}
	if payload.FileSize <= 0 {
		payload.FileSize = asset.FileSize
	}
	if payload.MimeType == "" {
		payload.MimeType = strings.TrimSpace(asset.MimeType)
	}
	return payload
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
