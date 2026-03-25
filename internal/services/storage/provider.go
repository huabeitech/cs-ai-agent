package storage

import (
	"cs-agent/internal/pkg/config"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/pkg/errorsx"
	"mime/multipart"
)

type FileStorageProvider interface {
	Upload(file *multipart.FileHeader, key string) (*StoredFile, error)
	GetURL(key string) (string, error)
	Delete(key string) error
}

func NewProvider(cfg config.StorageConfig) (FileStorageProvider, error) {
	switch enums.AssetProvider(cfg.Default) {
	case "", enums.AssetProviderLocal:
		return NewLocalStorage(cfg.Local), nil
	case enums.AssetProviderOSS:
		return NewOSSStorage(cfg.OSS), nil
	default:
		return nil, errorsx.InvalidParam("不支持的文件存储类型")
	}
}
