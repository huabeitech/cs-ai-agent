package services

import (
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/config"
	"cs-agent/internal/pkg/dto"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/pkg/errorsx"
	"cs-agent/internal/pkg/utils"
	"cs-agent/internal/repositories"
	"cs-agent/internal/services/storage"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mlogclub/simple/common/strs"
	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
)

var AssetService = newAssetService()

func newAssetService() *assetService {
	return &assetService{}
}

type assetService struct {
}

func (s *assetService) Get(id int64) *models.Asset {
	return repositories.AssetRepository.Get(sqls.DB(), id)
}

func (s *assetService) Take(where ...interface{}) *models.Asset {
	return repositories.AssetRepository.Take(sqls.DB(), where...)
}

func (s *assetService) Find(cnd *sqls.Cnd) []models.Asset {
	return repositories.AssetRepository.Find(sqls.DB(), cnd)
}

func (s *assetService) FindOne(cnd *sqls.Cnd) *models.Asset {
	return repositories.AssetRepository.FindOne(sqls.DB(), cnd)
}

func (s *assetService) FindPageByParams(params *params.QueryParams) (list []models.Asset, paging *sqls.Paging) {
	return repositories.AssetRepository.FindPageByParams(sqls.DB(), params)
}

func (s *assetService) FindPageByCnd(cnd *sqls.Cnd) (list []models.Asset, paging *sqls.Paging) {
	return repositories.AssetRepository.FindPageByCnd(sqls.DB(), cnd)
}

func (s *assetService) Count(cnd *sqls.Cnd) int64 {
	return repositories.AssetRepository.Count(sqls.DB(), cnd)
}

func (s *assetService) Create(t *models.Asset) error {
	return repositories.AssetRepository.Create(sqls.DB(), t)
}

func (s *assetService) Update(t *models.Asset) error {
	return repositories.AssetRepository.Update(sqls.DB(), t)
}

func (s *assetService) Updates(id int64, columns map[string]interface{}) error {
	return repositories.AssetRepository.Updates(sqls.DB(), id, columns)
}

func (s *assetService) UpdateColumn(id int64, name string, value interface{}) error {
	return repositories.AssetRepository.UpdateColumn(sqls.DB(), id, name, value)
}

func (s *assetService) Delete(id int64) {
	repositories.AssetRepository.Delete(sqls.DB(), id)
}

func (s *assetService) UploadFile(cfg *config.Config, file *multipart.FileHeader, prefix string, principal *dto.AuthPrincipal) (*models.Asset, error) {
	if principal == nil {
		return nil, errorsx.Unauthorized("未登录或登录已过期")
	}
	if cfg == nil {
		return nil, errorsx.InvalidParam("系统配置不存在")
	}
	if file == nil {
		return nil, errorsx.InvalidParam("请选择上传文件")
	}
	if file.Size > cfg.Storage.MaxUploadSizeBytes() {
		return nil, errorsx.InvalidParam("上传文件超过大小限制")
	}
	prefix = normalizeAssetPrefix(prefix)
	if prefix == "" {
		return nil, errorsx.InvalidParam("上传前缀不能为空")
	}

	provider, err := storage.NewProvider(cfg.Storage)
	if err != nil {
		return nil, err
	}

	assetID, key := s.generateStorageKey(prefix, file.Filename)
	item := &models.Asset{
		AssetID:     assetID,
		Provider:    cfg.Storage.Default,
		StorageKey:  key,
		Filename:    file.Filename,
		FileSize:    file.Size,
		MimeType:    file.Header.Get("Content-Type"),
		Status:      enums.AssetStatusPending,
		AuditFields: utils.BuildAuditFields(principal),
	}
	if strs.IsBlank(string(item.Provider)) {
		item.Provider = enums.AssetProviderLocal
	}
	if err := repositories.AssetRepository.Create(sqls.DB(), item); err != nil {
		return nil, err
	}

	storedFile, err := provider.Upload(file, key)
	if err != nil {
		_ = s.markAssetStatus(item.ID, enums.AssetStatusFailed, principal)
		return nil, err
	}

	item.Provider = storedFile.Provider
	item.Filename = storedFile.Filename
	item.FileSize = storedFile.FileSize
	item.MimeType = storedFile.MimeType
	item.Status = enums.AssetStatusSuccess
	_ = repositories.AssetRepository.Update(sqls.DB(), item) // 这个更新很简单，默认认为他一定能成功

	return item, nil
}

func (s *assetService) DeleteAsset(id int64, principal *dto.AuthPrincipal) error {
	if principal == nil {
		return errorsx.Unauthorized("未登录或登录已过期")
	}
	item := s.Get(id)
	if item == nil {
		return errorsx.InvalidParam("文件不存在")
	}
	return repositories.AssetRepository.Updates(sqls.DB(), id, map[string]any{
		"status":           enums.AssetStatusDeleted,
		"update_user_id":   principal.UserID,
		"update_user_name": principal.Username,
		"updated_at":       time.Now(),
	})
}

func (s *assetService) markAssetStatus(id int64, status enums.AssetStatus, principal *dto.AuthPrincipal) error {
	updates := map[string]any{
		"status":     status,
		"updated_at": time.Now(),
	}
	if principal != nil {
		updates["update_user_id"] = principal.UserID
		updates["update_user_name"] = principal.Username
	}
	return repositories.AssetRepository.Updates(sqls.DB(), id, updates)
}

func (s *assetService) generateStorageKey(prefix, filename string) (string, string) {
	assetID := strings.ReplaceAll(uuid.NewString(), "-", "")
	ext := strings.ToLower(filepath.Ext(strings.TrimSpace(filename)))
	datePath := time.Now().Format("2006/01/02")
	env := currentAssetEnv()
	return assetID, strings.TrimLeft(filepath.ToSlash(filepath.Join(env, prefix, datePath, assetID+ext)), "/")
}

func normalizeAssetPrefix(prefix string) string {
	prefix = strings.TrimSpace(prefix)
	prefix = strings.Trim(prefix, "/")
	prefix = strings.ReplaceAll(prefix, "..", "")
	prefix = filepath.ToSlash(prefix)
	for strings.Contains(prefix, "//") {
		prefix = strings.ReplaceAll(prefix, "//", "/")
	}
	return strings.Trim(prefix, "/")
}

func currentAssetEnv() string {
	for _, key := range []string{"APP_ENV", "GO_ENV"} {
		if value := strings.TrimSpace(os.Getenv(key)); value != "" {
			return value
		}
	}
	return string(enums.AssetProviderLocal)
}
