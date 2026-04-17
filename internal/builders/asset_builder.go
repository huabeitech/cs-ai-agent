package builders

import (
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/dto/response"
	"cs-agent/internal/services/storage"
	"log/slog"
)

func BuildAsset(item *models.Asset) response.AssetResponse {
	ret := response.AssetResponse{
		ID:             item.ID,
		AssetID:        item.AssetID,
		Provider:       item.Provider,
		Filename:       item.Filename,
		FileSize:       item.FileSize,
		MimeType:       item.MimeType,
		StorageKey:     item.StorageKey,
		Status:         item.Status,
		CreatedAt:      item.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:      item.UpdatedAt.Format("2006-01-02 15:04:05"),
		CreateUserID:   item.CreateUserID,
		CreateUserName: item.CreateUserName,
		UpdateUserID:   item.UpdateUserID,
		UpdateUserName: item.UpdateUserName,
	}

	if provider, err := storage.GetProvider(item.Provider); err != nil {
		slog.Error("get storage provider failed", "provider", item.Provider, "error", err)
	} else {
		ret.URL = provider.GetSignedURL(item.StorageKey)
	}

	return ret
}
