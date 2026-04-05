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
	channel := ChannelFromCtx(c.Ctx)
	if channel == nil {
		return web.JsonErrorMsg("接入渠道未初始化")
	}

	ret := response.WidgetConfigResponse{
		AppID: channel.AppID,
	}
	return web.JsonData(ret)
}
