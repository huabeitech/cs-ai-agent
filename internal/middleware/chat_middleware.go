package middleware

import (
	"cs-agent/internal/pkg/irisx"
	"cs-agent/internal/pkg/openidentity"
	"cs-agent/internal/services"

	"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple/web"
)

func ExternalInfoMiddleware(ctx iris.Context) {
	channel := services.ChannelService.GetEnabledChannel(ctx)
	var userTokenSecret string
	if channel != nil {
		userTokenSecret = services.ChannelService.GetUserTokenSecret(channel)
	}
	ext, err := openidentity.GetExternalInfoWithUserTokenSecret(ctx, userTokenSecret)
	if err != nil {
		ctx.StopExecution()
		_ = ctx.JSON(web.JsonError(err))
		return
	}
	irisx.SetExternalInfo(ctx, ext)
	ctx.Next()
}
