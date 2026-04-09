package irisx

import (
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/openidentity"

	"github.com/kataras/iris/v12"
)

const (
	ctxKeyOpenImChannel      = "openImChannel"
	ctxKeyOpenImExternalInfo = "openImExternalInfo"
)

func SetOpenImChannel(ctx iris.Context, channel *models.Channel) {
	ctx.Values().Set(ctxKeyOpenImChannel, channel)
}

// GetChannel 返回 OpenImContextMiddleware 注入的接入渠道（未走中间件时为 nil）。
func GetChannel(ctx iris.Context) *models.Channel {
	v := ctx.Values().Get(ctxKeyOpenImChannel)
	channel, _ := v.(*models.Channel)
	return channel
}

func SetOpenImExternalInfo(ctx iris.Context, ext *openidentity.ExternalInfo) {
	ctx.Values().Set(ctxKeyOpenImExternalInfo, ext)
}

// GetExternalInfo 返回中间件注入的外部访客身份；仅 widget 子路径或未启用中间件时为 nil。
func GetExternalInfo(ctx iris.Context) *openidentity.ExternalInfo {
	v := ctx.Values().Get(ctxKeyOpenImExternalInfo)
	ext, _ := v.(*openidentity.ExternalInfo)
	return ext
}
