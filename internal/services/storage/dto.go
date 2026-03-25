package storage

import "cs-agent/internal/pkg/enums"

type StoredFile struct {
	Provider   enums.AssetProvider
	StorageKey string
	Filename   string
	FileSize   int64
	MimeType   string
}
