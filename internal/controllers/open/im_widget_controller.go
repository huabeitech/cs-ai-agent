package open

import (
	"cs-agent/internal/pkg/dto/response"

	"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple/web"
)

type ImWidgetController struct {
	Ctx iris.Context
}

func (c *ImWidgetController) AnyConfig() *web.JsonResult {
	site := WidgetSiteFromCtx(c.Ctx)
	if site == nil {
		return web.JsonErrorMsg("接入站点未初始化")
	}

	ret := response.WidgetConfigResponse{
		AppID: site.AppID,
	}
	return web.JsonData(ret)
}
