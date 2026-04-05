package open

import (
	"cs-agent/internal/models"
	"cs-agent/internal/services"
	"strings"

	"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple/web"
)

func requireEnabledChannel(ctx iris.Context) (*models.Channel, *web.JsonResult) {
	appID := strings.TrimSpace(ctx.GetHeader("X-Widget-App-Id"))
	if appID == "" {
		appID = strings.TrimSpace(ctx.URLParam("appId"))
	}
	if appID == "" {
		return nil, web.JsonErrorMsg("appId不能为空")
	}
	channel := services.ChannelService.GetEnabledWebChannelByAppID(appID)
	if channel == nil {
		return nil, web.JsonErrorMsg("接入渠道不存在或已停用")
	}
	return channel, nil
}
