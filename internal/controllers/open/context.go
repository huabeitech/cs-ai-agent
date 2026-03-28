package open

import (
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/openidentity"

	"github.com/kataras/iris/v12"
)

const (
	ctxKeyOpenImWidgetSite   = "openImWidgetSite"
	ctxKeyOpenImExternalInfo = "openImExternalInfo"
)

// WidgetSiteFromCtx 返回 OpenImContextMiddleware 注入的接入站点（未走中间件时为 nil）。
func WidgetSiteFromCtx(ctx iris.Context) *models.WidgetSite {
	v := ctx.Values().Get(ctxKeyOpenImWidgetSite)
	site, _ := v.(*models.WidgetSite)
	return site
}

// ExternalInfoFromCtx 返回中间件注入的外部访客身份；仅 widget 子路径或未启用中间件时为 nil。
func ExternalInfoFromCtx(ctx iris.Context) *openidentity.ExternalInfo {
	v := ctx.Values().Get(ctxKeyOpenImExternalInfo)
	ext, _ := v.(*openidentity.ExternalInfo)
	return ext
}
