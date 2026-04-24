package open

import (
	"cs-agent/internal/pkg/dto/response"
	"cs-agent/internal/pkg/irisx"
	"cs-agent/internal/services"

	"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple/web"
)

type ImWidgetController struct {
	Ctx iris.Context
}

func (c *ImWidgetController) AnyConfig() *web.JsonResult {
	channel := irisx.GetChannel(c.Ctx)
	if channel == nil {
		return web.JsonErrorMsg("接入渠道未初始化")
	}
	cfg, err := services.ChannelService.ParseWebChannelConfig(channel.ConfigJSON)
	if err != nil {
		return web.JsonErrorMsg("Web渠道配置不合法")
	}

	ret := response.WidgetConfigResponse{
		ChannelID:  channel.ChannelID,
		Title:      cfg.Title,
		Subtitle:   cfg.Subtitle,
		ThemeColor: cfg.ThemeColor,
		Position:   cfg.Position,
		Width:      cfg.Width,
	}
	return web.JsonData(ret)
}
