package storage

import (
	"mime"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mlogclub/simple/common/strs"
)

func GenerateStorageKey(info UploadInfo) (assetID string, storageKey string) {
	assetID = strs.UUID()
	var (
		env      = currentAssetEnv()
		prefix   = normalizeAssetPrefix(info.Prefix)
		datePath = time.Now().Format("2006/01/02")
		ext      = getExt(info)
	)

	storageKey = filepath.Join(
		env,
		prefix,
		datePath,
		assetID+ext,
	)
	storageKey = strings.TrimLeft(filepath.ToSlash(storageKey), "/")
	return
}

func getExt(info UploadInfo) string {
	ext := strings.ToLower(filepath.Ext(strings.TrimSpace(info.Filename)))
	if ext == "" {
		ext = getExtByMimeType(info.MimeType)
	}
	return ext
}

func getExtByMimeType(mimeType string) string {
	if strs.IsBlank(mimeType) {
		return ""
	}

	mediaType, _, _ := mime.ParseMediaType(mimeType)
	if mediaType == "" {
		return ""
	}

	// 处理一些非标准的 MIME 类型
	switch mediaType {
	case "image/jfif":
		return ".jpg"
	case "image/pjpeg":
		return ".jpg"
	case "image/jpeg":
		return ".jpg"
	default:
		exts, _ := mime.ExtensionsByType(mediaType)
		if len(exts) > 0 {
			return exts[0]
		}
	}
	return ""
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
	return ""
}
