package storage

import (
	"cs-agent/internal/pkg/config"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/pkg/errorsx"
	"io"
)

type FileStorageProvider interface {
	ProviderType() enums.AssetProvider
	Upload(reader io.Reader, key string, info UploadInfo) (*StoredFile, error)
	GetURL(key string) string
	GetSignedURL(key string) string
	Delete(key string) error
	Read(key string) (io.ReadCloser, error)
}

func GetDefault() (FileStorageProvider, error) {
	return NewProvider(config.Current().Storage.Default)
}

func NewProvider(provider enums.AssetProvider) (FileStorageProvider, error) {
	cfg := config.Current().Storage

	switch provider {
	case "", enums.AssetProviderLocal:
		return NewLocalStorage(cfg.Local), nil
	case enums.AssetProviderOSS:
		return NewOSSStorage(cfg.OSS), nil
	default:
		return nil, errorsx.InvalidParam("不支持的文件存储类型")
	}
}
