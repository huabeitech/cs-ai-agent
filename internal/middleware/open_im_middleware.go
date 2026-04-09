package middleware

import (
	"cs-agent/internal/pkg/irisx"
	"cs-agent/internal/pkg/openidentity"
	"cs-agent/internal/services"
	"strings"

	"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple/web"
)

// OpenImContextMiddleware 校验 X-Channel-Id / channelId 对应启用 web 渠道；除 /api/open/im/widget 外解析并缓存外部访客身份。
func OpenImContextMiddleware(ctx iris.Context) {
	channelID := strings.TrimSpace(ctx.GetHeader("X-Channel-Id"))
	if channelID == "" {
		channelID = strings.TrimSpace(ctx.URLParam("channelId"))
	}
	if channelID == "" {
		ctx.StopExecution()
		_ = ctx.JSON(web.JsonErrorMsg("channelId不能为空"))
		return
	}
	channel := services.ChannelService.GetEnabledWebChannelByChannelID(channelID)
	if channel == nil {
		ctx.StopExecution()
		_ = ctx.JSON(web.JsonErrorMsg("接入渠道不存在或已停用"))
		return
	}
	irisx.SetOpenImChannel(ctx, channel)

	path := ctx.Path()
	if strings.Contains(path, "/open/im/widget") {
		ctx.Next()
		return
	}

	ext, err := openidentity.GetExternalInfo(ctx)
	if err != nil {
		ctx.StopExecution()
		_ = ctx.JSON(web.JsonError(err))
		return
	}
	irisx.SetOpenImExternalInfo(ctx, ext)
	ctx.Next()
}
