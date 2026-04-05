package open

import (
	"cs-agent/internal/models"
	"cs-agent/internal/services"
	"strings"

	"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple/web"
)

func requireEnabledChannel(ctx iris.Context) (*models.Channel, *web.JsonResult) {
	channelID := strings.TrimSpace(ctx.GetHeader("X-Channel-Id"))
	if channelID == "" {
		channelID = strings.TrimSpace(ctx.URLParam("channelId"))
	}
	if channelID == "" {
		return nil, web.JsonErrorMsg("channelId不能为空")
	}
	channel := services.ChannelService.GetEnabledWebChannelByChannelID(channelID)
	if channel == nil {
		return nil, web.JsonErrorMsg("接入渠道不存在或已停用")
	}
	return channel, nil
}
