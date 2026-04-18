package utils

import (
	"bytes"
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/repositories"
	"cs-agent/internal/services/storage"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/microcosm-cc/bluemonday"
	"github.com/mlogclub/simple/sqls"
	"golang.org/x/net/html"
)

type imMessageAssetPayload struct {
	AssetID    string              `json:"assetId"`
	Provider   enums.AssetProvider `json:"provider,omitempty"`
	StorageKey string              `json:"storageKey,omitempty"`
	Filename   string              `json:"filename,omitempty"`
	FileSize   int64               `json:"fileSize,omitempty"`
	MimeType   string              `json:"mimeType,omitempty"`
	URL        string              `json:"url,omitempty"`
}

func SanitizeMessageHTML(content string) string {
	policy := bluemonday.UGCPolicy()
	policy.AllowElements("img")
	policy.AllowAttrs("src", "alt", "title", "data-asset-id", "data-provider", "data-storage-key").OnElements("img")
	policy.AllowURLSchemes("http", "https")
	policy.AllowStandardURLs()
	policy.AllowElements("p", "br")
	return stripHTMLImageSrcIfBound(strings.TrimSpace(policy.Sanitize(content)))
}

func NormalizeMessageHTMLAssets(content string) (string, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return "", nil
	}
	doc, err := html.Parse(strings.NewReader("<div>" + content + "</div>"))
	if err != nil {
		return content, nil
	}
	var walkErr error
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node == nil || walkErr != nil {
			return
		}
		if node.Type == html.ElementNode && node.Data == "img" {
			asset, err := normalizeHTMLImageAsset(node)
			if err != nil {
				walkErr = err
				return
			}
			if asset != nil {
				setHTMLAttr(node, "data-asset-id", strings.TrimSpace(asset.AssetID))
				setHTMLAttr(node, "data-provider", strings.TrimSpace(string(asset.Provider)))
				setHTMLAttr(node, "data-storage-key", strings.TrimSpace(asset.StorageKey))
				removeHTMLAttr(node, "src")
			}
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(doc)
	if walkErr != nil {
		return "", walkErr
	}
	return renderHTMLFragment(doc), nil
}

func BuildHTMLSummary(content string) string {
	if strings.TrimSpace(content) == "" {
		return ""
	}
	doc, err := html.Parse(strings.NewReader("<div>" + content + "</div>"))
	if err != nil {
		return strings.TrimSpace(content)
	}
	parts := make([]string, 0, 8)
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node == nil {
			return
		}
		if node.Type == html.TextNode {
			text := strings.TrimSpace(node.Data)
			if text != "" {
				parts = append(parts, text)
			}
		}
		if node.Type == html.ElementNode && node.Data == "img" {
			parts = append(parts, "[图片]")
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(doc)
	return strings.TrimSpace(strings.Join(parts, " "))
}

func BuildRuntimeMessageText(messageType enums.IMMessageType, content string) string {
	content = strings.TrimSpace(content)
	switch messageType {
	case enums.IMMessageTypeHTML:
		return BuildHTMLSummary(content)
	case enums.IMMessageTypeImage:
		if content != "" {
			return "[图片] " + content
		}
		return "[图片]"
	case enums.IMMessageTypeAttachment:
		if content != "" {
			return "[附件] " + content
		}
		return "[附件]"
	default:
		return content
	}
}

func BuildRenderableMessage(item *models.Message) (content, payload string) {
	if item == nil {
		return "", ""
	}
	if item.RecalledAt != nil {
		return "该消息已撤回", ""
	}
	if item.SendStatus == enums.IMMessageStatusRecalled {
		return "该消息已撤回", ""
	}

	content = item.Content
	payload = item.Payload
	switch item.MessageType {
	case enums.IMMessageTypeImage, enums.IMMessageTypeAttachment:
		payload = buildIMMessageAssetPayloadForResponse(item.Payload)
	case enums.IMMessageTypeHTML:
		content = BuildMessageHTMLForResponse(item.Content)
	}
	return content, payload
}

func BuildMessageHTMLForResponse(content string) string {
	content = strings.TrimSpace(content)
	if content == "" {
		return ""
	}
	doc, err := html.Parse(strings.NewReader("<div>" + content + "</div>"))
	if err != nil {
		return content
	}
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node == nil {
			return
		}
		if node.Type == html.ElementNode && node.Data == "img" {
			provider := enums.AssetProvider(strings.TrimSpace(findHTMLAttr(node, "data-provider")))
			storageKey := strings.TrimSpace(findHTMLAttr(node, "data-storage-key"))
			if provider != "" && storageKey != "" {
				if storageProvider, err := storage.NewProvider(provider); err == nil {
					setHTMLAttr(node, "src", storageProvider.GetSignedURL(storageKey))
				}
			}
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(doc)
	return renderHTMLFragment(doc)
}

func stripHTMLImageSrcIfBound(content string) string {
	content = strings.TrimSpace(content)
	if content == "" {
		return ""
	}
	doc, err := html.Parse(strings.NewReader("<div>" + content + "</div>"))
	if err != nil {
		return content
	}
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node == nil {
			return
		}
		if node.Type == html.ElementNode && node.Data == "img" {
			provider := strings.TrimSpace(findHTMLAttr(node, "data-provider"))
			storageKey := strings.TrimSpace(findHTMLAttr(node, "data-storage-key"))
			if provider != "" && storageKey != "" {
				removeHTMLAttr(node, "src")
			}
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(doc)
	return renderHTMLFragment(doc)
}

func renderHTMLFragment(doc *html.Node) string {
	if doc == nil {
		return ""
	}
	root := findHTMLRoot(doc)
	if root == nil {
		return ""
	}
	var buf bytes.Buffer
	for child := root.FirstChild; child != nil; child = child.NextSibling {
		if err := html.Render(&buf, child); err != nil {
			return ""
		}
	}
	return strings.TrimSpace(buf.String())
}

func findHTMLRoot(doc *html.Node) *html.Node {
	var walk func(*html.Node) *html.Node
	walk = func(node *html.Node) *html.Node {
		if node == nil {
			return nil
		}
		if node.Type == html.ElementNode && node.Data == "div" {
			return node
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			if found := walk(child); found != nil {
				return found
			}
		}
		return nil
	}
	return walk(doc)
}

func findHTMLAttr(node *html.Node, key string) string {
	if node == nil {
		return ""
	}
	for _, attr := range node.Attr {
		if attr.Key == key {
			return attr.Val
		}
	}
	return ""
}

func setHTMLAttr(node *html.Node, key, value string) {
	if node == nil {
		return
	}
	for i := range node.Attr {
		if node.Attr[i].Key == key {
			node.Attr[i].Val = value
			return
		}
	}
	node.Attr = append(node.Attr, html.Attribute{Key: key, Val: value})
}

func removeHTMLAttr(node *html.Node, key string) {
	if node == nil {
		return
	}
	dst := node.Attr[:0]
	for _, attr := range node.Attr {
		if attr.Key != key {
			dst = append(dst, attr)
		}
	}
	node.Attr = dst
}

func buildIMMessageAssetPayloadForResponse(payload string) string {
	assetPayload, err := parseIMMessageAssetPayload(payload)
	if err != nil {
		return strings.TrimSpace(payload)
	}
	assetPayload = hydrateIMMessageAssetPayload(assetPayload)
	if assetPayload.Provider != "" && assetPayload.StorageKey != "" {
		if provider, err := storage.NewProvider(assetPayload.Provider); err == nil {
			assetPayload.URL = provider.GetSignedURL(assetPayload.StorageKey)
		}
	}
	data, err := json.Marshal(assetPayload)
	if err != nil {
		return strings.TrimSpace(payload)
	}
	return string(data)
}

func parseIMMessageAssetPayload(payload string) (*imMessageAssetPayload, error) {
	payload = strings.TrimSpace(payload)
	if payload == "" {
		return nil, nil
	}
	ret := &imMessageAssetPayload{}
	if err := json.Unmarshal([]byte(payload), ret); err != nil {
		return nil, err
	}
	ret.AssetID = strings.TrimSpace(ret.AssetID)
	ret.Provider = enums.AssetProvider(strings.TrimSpace(string(ret.Provider)))
	ret.StorageKey = strings.TrimSpace(ret.StorageKey)
	return ret, nil
}

func hydrateIMMessageAssetPayload(payload *imMessageAssetPayload) *imMessageAssetPayload {
	if payload == nil {
		return nil
	}
	if payload.Provider != "" && payload.StorageKey != "" {
		return payload
	}
	if payload.AssetID == "" {
		return payload
	}
	asset := repositories.AssetRepository.GetByAssetID(sqls.DB(), payload.AssetID)
	if asset == nil {
		return payload
	}
	if payload.Provider == "" {
		payload.Provider = asset.Provider
	}
	if payload.StorageKey == "" {
		payload.StorageKey = strings.TrimSpace(asset.StorageKey)
	}
	if payload.Filename == "" {
		payload.Filename = strings.TrimSpace(asset.Filename)
	}
	if payload.FileSize <= 0 {
		payload.FileSize = asset.FileSize
	}
	if payload.MimeType == "" {
		payload.MimeType = strings.TrimSpace(asset.MimeType)
	}
	return payload
}

func normalizeHTMLImageAsset(node *html.Node) (*models.Asset, error) {
	if node == nil {
		return nil, nil
	}
	assetID := strings.TrimSpace(findHTMLAttr(node, "data-asset-id"))
	provider := enums.AssetProvider(strings.TrimSpace(findHTMLAttr(node, "data-provider")))
	storageKey := strings.TrimSpace(findHTMLAttr(node, "data-storage-key"))

	hasAssetID := assetID != ""
	hasProvider := provider != ""
	hasStorageKey := storageKey != ""
	if hasAssetID || hasProvider || hasStorageKey {
		if !(hasAssetID && hasProvider && hasStorageKey) {
			return nil, fmt.Errorf("html message image asset attributes are incomplete")
		}
		asset := repositories.AssetRepository.GetByAssetID(sqls.DB(), assetID)
		if asset == nil {
			return nil, fmt.Errorf("html message image asset not found")
		}
		if asset.Provider != provider || strings.TrimSpace(asset.StorageKey) != storageKey {
			return nil, fmt.Errorf("html message image asset attributes mismatch")
		}
		return asset, nil
	}
	return nil, fmt.Errorf("html message image must include asset metadata")
}
