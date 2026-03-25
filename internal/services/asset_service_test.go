package services

import (
	"bytes"
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/config"
	"cs-agent/internal/pkg/dto"
	"cs-agent/internal/pkg/enums"
	"io"
	"mime/multipart"
	"net/textproto"
	"os"
	"path/filepath"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	"github.com/mlogclub/simple/sqls"
)

func setupAssetServiceTestDB(t *testing.T) {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "asset-service-test.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "t_",
			SingularTable: true,
		},
	})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}
	sqls.SetDB(db)
	if err := sqls.DB().AutoMigrate(models.Models...); err != nil {
		t.Fatalf("auto migrate failed: %v", err)
	}
	t.Cleanup(func() {
		sqlDB, err := db.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
		_ = os.Remove(dbPath)
	})
}

func buildTestFileHeader(t *testing.T, filename string, body []byte) *multipart.FileHeader {
	t.Helper()
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", `form-data; name="file"; filename="`+filename+`"`)
	h.Set("Content-Type", "text/plain")
	part, err := writer.CreatePart(h)
	if err != nil {
		t.Fatalf("create multipart part failed: %v", err)
	}
	if _, err := io.Copy(part, bytes.NewReader(body)); err != nil {
		t.Fatalf("write multipart part failed: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer failed: %v", err)
	}
	reader := multipart.NewReader(bytes.NewReader(buf.Bytes()), writer.Boundary())
	form, err := reader.ReadForm(int64(len(buf.Bytes())))
	if err != nil {
		t.Fatalf("read form failed: %v", err)
	}
	files := form.File["file"]
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}
	return files[0]
}

func TestAssetUploadFileSuccess(t *testing.T) {
	setupAssetServiceTestDB(t)

	rootDir := t.TempDir()
	cfg := &config.Config{
		Storage: config.StorageConfig{
			Default:         enums.AssetProviderLocal,
			MaxUploadSizeMB: 5,
			Local: config.LocalStorageConfig{
				Root:    rootDir,
				BaseURL: "/storage",
			},
		},
	}
	principal := &dto.AuthPrincipal{UserID: 1001, Username: "tester"}
	header := buildTestFileHeader(t, "demo.txt", []byte("hello asset"))

	item, err := AssetService.UploadFile(cfg, header, "ticket-files", principal)
	if err != nil {
		t.Fatalf("upload file failed: %v", err)
	}
	if item.ID <= 0 {
		t.Fatalf("expected persisted id, got %d", item.ID)
	}
	if item.Status != enums.AssetStatusSuccess {
		t.Fatalf("expected status success, got %d", item.Status)
	}
	if item.CreateUserID != principal.UserID {
		t.Fatalf("expected create user %d, got %d", principal.UserID, item.CreateUserID)
	}

	saved := AssetService.Get(item.ID)
	if saved == nil {
		t.Fatalf("expected saved asset")
	}
	if saved.Status != enums.AssetStatusSuccess {
		t.Fatalf("expected saved status success, got %d", saved.Status)
	}
	fullPath := filepath.Join(rootDir, filepath.FromSlash(saved.StorageKey))
	if _, err := os.Stat(fullPath); err != nil {
		t.Fatalf("expected uploaded file on disk: %v", err)
	}
}

func TestAssetUploadFileRejectsOversize(t *testing.T) {
	setupAssetServiceTestDB(t)

	cfg := &config.Config{
		Storage: config.StorageConfig{
			Default:         enums.AssetProviderLocal,
			MaxUploadSizeMB: 1,
			Local: config.LocalStorageConfig{
				Root:    t.TempDir(),
				BaseURL: "/storage",
			},
		},
	}
	principal := &dto.AuthPrincipal{UserID: 1002, Username: "tester"}
	header := buildTestFileHeader(t, "large.txt", bytes.Repeat([]byte("a"), (2<<20)))

	if _, err := AssetService.UploadFile(cfg, header, "ticket-files", principal); err == nil {
		t.Fatalf("expected oversize upload error")
	}
}
