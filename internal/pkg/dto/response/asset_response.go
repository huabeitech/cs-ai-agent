package response

import "cs-agent/internal/pkg/enums"

type AssetResponse struct {
	ID             int64               `json:"id"`
	AssetID        string              `json:"assetId"`
	Provider       enums.AssetProvider `json:"provider"`
	StorageKey     string              `json:"storageKey"`
	Filename       string              `json:"filename"`
	FileSize       int64               `json:"fileSize"`
	MimeType       string              `json:"mimeType"`
	Status         enums.AssetStatus   `json:"status"`
	URL            string              `json:"url"`
	CreatedAt      string              `json:"createdAt"`
	UpdatedAt      string              `json:"updatedAt"`
	CreateUserID   int64               `json:"createUserId"`
	CreateUserName string              `json:"createUserName"`
	UpdateUserID   int64               `json:"updateUserId"`
	UpdateUserName string              `json:"updateUserName"`
}
