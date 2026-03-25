package storage

import (
	"cs-agent/internal/pkg/config"
	"cs-agent/internal/pkg/enums"
	"io"
	"mime/multipart"
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

func (l *LocalStorage) Upload(file *multipart.FileHeader, key string) (*StoredFile, error) {
	src, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()

	fullPath := filepath.Join(l.cfg.Root, filepath.FromSlash(key))
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return nil, err
	}

	dst, err := os.Create(fullPath)
	if err != nil {
		return nil, err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return nil, err
	}

	return &StoredFile{
		Provider:   enums.AssetProviderLocal,
		StorageKey: key,
		Filename:   file.Filename,
		FileSize:   file.Size,
		MimeType:   file.Header.Get("Content-Type"),
	}, nil
}

func (l *LocalStorage) GetURL(key string) (string, error) {
	baseURL := strings.TrimRight(strings.TrimSpace(l.cfg.BaseURL), "/")
	return baseURL + "/" + strings.TrimLeft(key, "/"), nil
}

func (l *LocalStorage) Delete(key string) error {
	fullPath := filepath.Join(l.cfg.Root, filepath.FromSlash(key))
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return nil
	}
	return os.Remove(fullPath)
}
