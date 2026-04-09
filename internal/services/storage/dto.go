package storage

import (
	"cs-agent/internal/pkg/dto"
	"cs-agent/internal/pkg/enums"
)

type UploadInfo struct {
	Prefix    string
	Filename  string
	FileSize  int64
	MimeType  string
	Principal *dto.AuthPrincipal
}

type StoredFile struct {
	Provider   enums.AssetProvider
	StorageKey string
	URL        string
	Filename   string
	FileSize   int64
	MimeType   string
}
