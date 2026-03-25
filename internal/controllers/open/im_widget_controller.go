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
	site, rsp := requireEnabledWidgetSite(c.Ctx)
	if rsp != nil {
		return rsp
	}

	ret := response.WidgetConfigResponse{
		AppID: site.AppID,
	}
	return web.JsonData(ret)
}
