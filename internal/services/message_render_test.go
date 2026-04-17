package services

import (
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/config"
	"cs-agent/internal/pkg/enums"
	"strings"
	"testing"
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

	got := sanitizeMessageHTML(html)

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
	got := buildMessageHTMLForResponse(html)

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
