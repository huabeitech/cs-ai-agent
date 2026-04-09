package storage

import (
	"cs-agent/internal/pkg/config"
	"cs-agent/internal/pkg/enums"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type LocalStorage struct {
	cfg config.LocalStorageConfig
}

func NewLocalStorage(cfg config.LocalStorageConfig) *LocalStorage {
	return &LocalStorage{cfg: cfg}
}

func (s *LocalStorage) ProviderType() enums.AssetProvider {
	return enums.AssetProviderLocal
}

func (s *LocalStorage) Upload(reader io.Reader, key string, info UploadInfo) (*StoredFile, error) {
	fullPath := filepath.Join(s.cfg.Root, filepath.FromSlash(key))
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return nil, err
	}

	dst, err := os.Create(fullPath)
	if err != nil {
		return nil, err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, reader); err != nil {
		return nil, err
	}

	return &StoredFile{
		Provider:   enums.AssetProviderLocal,
		StorageKey: key,
		URL:        strings.TrimRight(s.cfg.BaseURL, "/") + "/" + strings.TrimLeft(key, "/"),
		Filename:   info.Filename,
		FileSize:   info.FileSize,
		MimeType:   info.MimeType,
	}, nil
}

func (s *LocalStorage) GetURL(key string) string {
	baseURL := strings.TrimRight(strings.TrimSpace(s.cfg.BaseURL), "/")
	return strings.TrimRight(baseURL, "/") + "/" + strings.TrimLeft(key, "/")
}

func (s *LocalStorage) GetSignedURL(key string) string {
	return s.GetURL(key)
}

func (s *LocalStorage) Delete(key string) error {
	fullPath := filepath.Join(s.cfg.Root, filepath.FromSlash(key))
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return nil
	}
	return os.Remove(fullPath)
}

func (s *LocalStorage) Read(key string) (io.ReadCloser, error) {
	fullPath := filepath.Join(s.cfg.Root, filepath.FromSlash(strings.TrimSpace(key)))
	return os.Open(fullPath)
}
