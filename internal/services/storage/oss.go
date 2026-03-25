package storage

import (
	"cs-agent/internal/pkg/config"
	"errors"
	"mime/multipart"
)

type OSSStorage struct {
	cfg config.OSSStorageConfig
}

func NewOSSStorage(cfg config.OSSStorageConfig) *OSSStorage {
	return &OSSStorage{cfg: cfg}
}

func (o *OSSStorage) Upload(file *multipart.FileHeader, key string) (*StoredFile, error) {
	return nil, errors.New("暂不支持 OSS 文件存储")
}

func (o *OSSStorage) GetURL(key string) (string, error) {
	return "", errors.New("暂不支持 OSS 文件存储")
}

func (o *OSSStorage) Delete(key string) error {
	return errors.New("暂不支持 OSS 文件存储")
}
