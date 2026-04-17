package utils

import (
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/config"
	"cs-agent/internal/pkg/enums"
	"strings"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/mlogclub/simple/sqls"
	"gorm.io/gorm"
)

func TestBuildIMMessageAssetPayloadForResponseAddsSignedURL(t *testing.T) {
	config.SetCurrent(&config.Config{
		Storage: config.StorageConfig{
			Default: enums.AssetProviderLocal,
			Local: config.LocalStorageConfig{
				BaseURL: "https://files.example.com",
			},
		},
	})

	payload := `{"assetId":"asset_1","provider":"local","storageKey":"attachments/demo.png","filename":"demo.png"}`
	got := buildIMMessageAssetPayloadForResponse(payload)

	if !strings.Contains(got, `"provider":"local"`) {
		t.Fatalf("expected provider in payload, got: %s", got)
	}
	if !strings.Contains(got, `"storageKey":"attachments/demo.png"`) {
		t.Fatalf("expected storageKey in payload, got: %s", got)
	}
	if !strings.Contains(got, `"url":"https://files.example.com/attachments/demo.png"`) {
		t.Fatalf("expected signed url in payload, got: %s", got)
	}
}

func TestSanitizeMessageHTMLStripsStoredSrcForManagedImages(t *testing.T) {
	html := `<p><img src="https://files.example.com/demo.png" data-provider="local" data-storage-key="attachments/demo.png" alt="demo"></p>`

	got := SanitizeMessageHTML(html)

	if strings.Contains(got, `src=`) {
		t.Fatalf("expected src removed from stored html, got: %s", got)
	}
	if !strings.Contains(got, `data-provider="local"`) {
		t.Fatalf("expected data-provider kept, got: %s", got)
	}
	if !strings.Contains(got, `data-storage-key="attachments/demo.png"`) {
		t.Fatalf("expected data-storage-key kept, got: %s", got)
	}
}

func TestBuildMessageHTMLForResponseAddsSignedURL(t *testing.T) {
	config.SetCurrent(&config.Config{
		Storage: config.StorageConfig{
			Default: enums.AssetProviderLocal,
			Local: config.LocalStorageConfig{
				BaseURL: "https://files.example.com",
			},
		},
	})

	html := `<p><img data-provider="local" data-storage-key="attachments/demo.png" alt="demo"></p>`
	got := BuildMessageHTMLForResponse(html)

	if !strings.Contains(got, `src="https://files.example.com/attachments/demo.png"`) {
		t.Fatalf("expected signed src in response html, got: %s", got)
	}
}

func TestBuildRenderableMessageTransformsPayloadAndHTML(t *testing.T) {
	config.SetCurrent(&config.Config{
		Storage: config.StorageConfig{
			Default: enums.AssetProviderLocal,
			Local: config.LocalStorageConfig{
				BaseURL: "https://files.example.com",
			},
		},
	})

	image := &models.Message{
		MessageType: enums.IMMessageTypeImage,
		Payload:     `{"assetId":"asset_1","provider":"local","storageKey":"attachments/demo.png","filename":"demo.png"}`,
	}
	_, imagePayload := BuildRenderableMessage(image)
	if !strings.Contains(imagePayload, `"url":"https://files.example.com/attachments/demo.png"`) {
		t.Fatalf("expected image payload signed url, got: %s", imagePayload)
	}

	htmlMsg := &models.Message{
		MessageType: enums.IMMessageTypeHTML,
		Content:     `<p><img data-provider="local" data-storage-key="attachments/demo.png"></p>`,
	}
	htmlContent, _ := BuildRenderableMessage(htmlMsg)
	if !strings.Contains(htmlContent, `src="https://files.example.com/attachments/demo.png"`) {
		t.Fatalf("expected html content signed src, got: %s", htmlContent)
	}
}

func TestNormalizeMessageHTMLAssetsKeepsValidAttrsAndRemovesSrc(t *testing.T) {
	setupMessageTestDB(t)
	config.SetCurrent(&config.Config{
		Storage: config.StorageConfig{
			Default: enums.AssetProviderLocal,
			Local: config.LocalStorageConfig{
				BaseURL: "https://files.example.com",
			},
		},
	})
	createTestAsset(t, &models.Asset{
		AssetID:    "asset_local_1",
		Provider:   enums.AssetProviderLocal,
		StorageKey: "images/demo.png",
		Filename:   "demo.png",
		FileSize:   123,
		MimeType:   "image/png",
		Status:     enums.AssetStatusSuccess,
	})

	got, err := NormalizeMessageHTMLAssets(`<p><img src="https://files.example.com/images/demo.png" data-asset-id="asset_local_1" data-provider="local" data-storage-key="images/demo.png" alt="demo"></p>`)
	if err != nil {
		t.Fatalf("expected normalization success, got error: %v", err)
	}

	if !strings.Contains(got, `data-asset-id="asset_local_1"`) {
		t.Fatalf("expected data-asset-id added, got: %s", got)
	}
	if !strings.Contains(got, `data-provider="local"`) {
		t.Fatalf("expected data-provider added, got: %s", got)
	}
	if !strings.Contains(got, `data-storage-key="images/demo.png"`) {
		t.Fatalf("expected data-storage-key added, got: %s", got)
	}
	if strings.Contains(got, `src=`) {
		t.Fatalf("expected src removed after asset binding, got: %s", got)
	}
}

func TestNormalizeMessageHTMLAssetsRejectsMissingAssetMetadata(t *testing.T) {
	setupMessageTestDB(t)
	config.SetCurrent(&config.Config{
		Storage: config.StorageConfig{
			Default: enums.AssetProviderLocal,
			Local: config.LocalStorageConfig{
				BaseURL: "https://files.example.com",
			},
		},
	})

	_, err := NormalizeMessageHTMLAssets(`<p><img src="https://unknown.example.com/demo.png" alt="demo"></p>`)
	if err == nil {
		t.Fatalf("expected missing image asset metadata rejected")
	}
}

func TestNormalizeMessageHTMLAssetsRejectsIncompleteAttrs(t *testing.T) {
	setupMessageTestDB(t)
	config.SetCurrent(&config.Config{
		Storage: config.StorageConfig{
			Default: enums.AssetProviderLocal,
			Local: config.LocalStorageConfig{
				BaseURL: "https://files.example.com",
			},
		},
	})

	_, err := NormalizeMessageHTMLAssets(`<p><img data-asset-id="asset1" data-provider="local" alt="demo"></p>`)
	if err == nil {
		t.Fatalf("expected incomplete asset attrs rejected")
	}
}

func TestNormalizeMessageHTMLAssetsRejectsMismatchedAttrs(t *testing.T) {
	setupMessageTestDB(t)
	config.SetCurrent(&config.Config{
		Storage: config.StorageConfig{
			Default: enums.AssetProviderLocal,
			Local: config.LocalStorageConfig{
				BaseURL: "https://files.example.com",
			},
		},
	})
	createTestAsset(t, &models.Asset{
		AssetID:    "asset_local_2",
		Provider:   enums.AssetProviderLocal,
		StorageKey: "images/real.png",
		Filename:   "real.png",
		FileSize:   456,
		MimeType:   "image/png",
		Status:     enums.AssetStatusSuccess,
	})

	_, err := NormalizeMessageHTMLAssets(`<p><img data-asset-id="asset_local_2" data-provider="local" data-storage-key="images/wrong.png" alt="demo"></p>`)
	if err == nil {
		t.Fatalf("expected mismatched asset attrs rejected")
	}
}

func setupMessageTestDB(t *testing.T) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}
	if err := db.AutoMigrate(&models.Asset{}); err != nil {
		t.Fatalf("auto migrate asset failed: %v", err)
	}
	sqls.SetDB(db)
}

func createTestAsset(t *testing.T, item *models.Asset) {
	t.Helper()
	now := time.Now()
	if item.CreatedAt.IsZero() {
		item.CreatedAt = now
	}
	if item.UpdatedAt.IsZero() {
		item.UpdatedAt = now
	}
	if err := sqls.DB().Create(item).Error; err != nil {
		t.Fatalf("create asset failed: %v", err)
	}
}
