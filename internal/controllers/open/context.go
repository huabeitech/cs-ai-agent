package open

import (
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/openidentity"

	"github.com/kataras/iris/v12"
)

const (
	ctxKeyOpenImChannel      = "openImChannel"
	ctxKeyOpenImExternalInfo = "openImExternalInfo"
)

// ChannelFromCtx 返回 OpenImContextMiddleware 注入的接入渠道（未走中间件时为 nil）。
func ChannelFromCtx(ctx iris.Context) *models.Channel {
	v := ctx.Values().Get(ctxKeyOpenImChannel)
	channel, _ := v.(*models.Channel)
	return channel
}

// ExternalInfoFromCtx 返回中间件注入的外部访客身份；仅 widget 子路径或未启用中间件时为 nil。
func ExternalInfoFromCtx(ctx iris.Context) *openidentity.ExternalInfo {
	v := ctx.Values().Get(ctxKeyOpenImExternalInfo)
	ext, _ := v.(*openidentity.ExternalInfo)
	return ext
}
