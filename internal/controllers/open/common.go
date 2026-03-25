package open

import (
	"cs-agent/internal/models"
	"cs-agent/internal/services"
	"strings"

	"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple/web"
)

func requireEnabledWidgetSite(ctx iris.Context) (*models.WidgetSite, *web.JsonResult) {
	appID := strings.TrimSpace(ctx.GetHeader("X-Widget-App-Id"))
	if appID == "" {
		appID = strings.TrimSpace(ctx.URLParam("appId"))
	}
	if appID == "" {
		return nil, web.JsonErrorMsg("appId不能为空")
	}
	site := services.WidgetSiteService.FindEnabledByAppID(appID)
	if site == nil {
		return nil, web.JsonErrorMsg("接入站点不存在或已停用")
	}
	return site, nil
}
