package storage

import (
	"cs-agent/internal/pkg/config"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/pkg/errorsx"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/mlogclub/simple/common/strs"
)

type OSSStorage struct {
	cfg config.OSSStorageConfig
}

func NewOSSStorage(cfg config.OSSStorageConfig) *OSSStorage {
	return &OSSStorage{cfg: cfg}
}

func (s *OSSStorage) ProviderType() enums.AssetProvider {
	return enums.AssetProviderOSS
}

func (s *OSSStorage) Upload(reader io.Reader, key string, info UploadInfo) (*StoredFile, error) {
	bucket, err := s.getBucket()
	if err != nil {
		return nil, err
	}

	key = normalizeOSSKey(key)
	options := []oss.Option{}
	if mimeType := strings.TrimSpace(info.MimeType); mimeType != "" {
		options = append(options, oss.ContentType(mimeType))
	}
	if info.FileSize > 0 {
		options = append(options, oss.ContentLength(info.FileSize))
	}

	if err := bucket.PutObject(key, reader, options...); err != nil {
		return nil, err
	}

	return &StoredFile{
		Provider:   enums.AssetProviderOSS,
		StorageKey: key,
		URL:        s.GetURL(key),
		Filename:   info.Filename,
		FileSize:   info.FileSize,
		MimeType:   info.MimeType,
	}, nil
}

func (s *OSSStorage) GetURL(key string) string {
	key = normalizeOSSKey(key)
	if key == "" {
		return ""
	}

	if baseURL := strings.TrimRight(strings.TrimSpace(s.cfg.BaseURL), "/"); baseURL != "" {
		return baseURL + "/" + key
	}

	if s.cfg.Private {
		bucket, err := s.getBucket()
		if err != nil {
			return ""
		}
		signedURL, err := bucket.SignURL(key, oss.HTTPGet, s.signedURLExpire())
		if err == nil {
			return signedURL
		}
	}

	return s.objectURL(key)
}

func (s *OSSStorage) GetSignedURL(key string) string {
	key = normalizeOSSKey(key)
	if strs.IsBlank(key) {
		return ""
	}
	if !s.cfg.Private {
		return s.GetURL(key)
	}

	bucket, err := s.getBucket()
	if err != nil {
		return ""
	}
	signedURL, err := bucket.SignURL(key, oss.HTTPGet, s.signedURLExpire())
	if err != nil {
		return ""
	}
	return signedURL
}

func (s *OSSStorage) Delete(key string) error {
	bucket, err := s.getBucket()
	if err != nil {
		return err
	}
	return bucket.DeleteObject(normalizeOSSKey(key))
}

func (s *OSSStorage) Read(key string) (io.ReadCloser, error) {
	bucket, err := s.getBucket()
	if err != nil {
		return nil, err
	}
	return bucket.GetObject(normalizeOSSKey(key))
}

func (s *OSSStorage) getBucket() (*oss.Bucket, error) {
	if err := s.validate(); err != nil {
		return nil, err
	}

	client, err := oss.New(
		s.endpoint(),
		strings.TrimSpace(s.cfg.AccessKeyID),
		strings.TrimSpace(s.cfg.AccessKeySecret),
	)
	if err != nil {
		return nil, err
	}
	return client.Bucket(strings.TrimSpace(s.cfg.Bucket))
}

func (s *OSSStorage) validate() error {
	if strings.TrimSpace(s.cfg.Endpoint) == "" {
		return errorsx.InvalidParam("OSS endpoint 未配置")
	}
	if strings.TrimSpace(s.cfg.Bucket) == "" {
		return errorsx.InvalidParam("OSS bucket 未配置")
	}
	if strings.TrimSpace(s.cfg.AccessKeyID) == "" {
		return errorsx.InvalidParam("OSS accessKeyId 未配置")
	}
	if strings.TrimSpace(s.cfg.AccessKeySecret) == "" {
		return errorsx.InvalidParam("OSS accessKeySecret 未配置")
	}
	return nil
}

func (s *OSSStorage) endpoint() string {
	endpoint := strings.TrimSpace(s.cfg.Endpoint)
	if endpoint == "" {
		return ""
	}
	if strings.Contains(endpoint, "://") {
		return endpoint
	}
	return "https://" + endpoint
}

func (s *OSSStorage) objectURL(key string) string {
	u, err := url.Parse(s.endpoint())
	if err != nil || u.Host == "" {
		return ""
	}
	return fmt.Sprintf("%s://%s.%s/%s", u.Scheme, strings.TrimSpace(s.cfg.Bucket), u.Host, key)
}

func (s *OSSStorage) signedURLExpire() int64 {
	if s.cfg.SignedURLExpire > 0 {
		return int64(s.cfg.SignedURLExpire)
	}
	return 600
}

func normalizeOSSKey(key string) string {
	return strings.TrimLeft(strings.TrimSpace(key), "/")
}
