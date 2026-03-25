package builders

import (
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/dto/response"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/services/storage"
)

func BuildAssetResponse(item *models.Asset, provider storage.FileStorageProvider) response.AssetResponse {
	ret := response.AssetResponse{
		ID:             item.ID,
		AssetID:        item.AssetID,
		Provider:       item.Provider,
		StorageKey:     item.StorageKey,
		Filename:       item.Filename,
		FileSize:       item.FileSize,
		MimeType:       item.MimeType,
		Status:         item.Status,
		CreatedAt:      item.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:      item.UpdatedAt.Format("2006-01-02 15:04:05"),
		CreateUserID:   item.CreateUserID,
		CreateUserName: item.CreateUserName,
		UpdateUserID:   item.UpdateUserID,
		UpdateUserName: item.UpdateUserName,
	}
	if provider != nil && item.Status == enums.AssetStatusSuccess {
		if url, err := provider.GetURL(item.StorageKey); err == nil {
			ret.URL = url
		}
	}
	if ret.CreatedAt == "0001-01-01 00:00:00" {
		ret.CreatedAt = ""
	}
	if ret.UpdatedAt == "0001-01-01 00:00:00" {
		ret.UpdatedAt = ""
	}
	return ret
}
