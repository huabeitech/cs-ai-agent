package open

import (
	"cs-agent/internal/pkg/openidentity"
	"cs-agent/internal/services"
	"strings"

	"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple/web"
)

// OpenImContextMiddleware 校验 X-Widget-App-Id / appId 对应启用 web 渠道；除 /api/open/im/widget 外解析并缓存外部访客身份。
func OpenImContextMiddleware(ctx iris.Context) {
	appID := strings.TrimSpace(ctx.GetHeader("X-Widget-App-Id"))
	if appID == "" {
		appID = strings.TrimSpace(ctx.URLParam("appId"))
	}
	if appID == "" {
		ctx.StopExecution()
		_ = ctx.JSON(web.JsonErrorMsg("appId不能为空"))
		return
	}
	channel := services.ChannelService.GetEnabledWebChannelByAppID(appID)
	if channel == nil {
		ctx.StopExecution()
		_ = ctx.JSON(web.JsonErrorMsg("接入渠道不存在或已停用"))
		return
	}
	ctx.Values().Set(ctxKeyOpenImChannel, channel)

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
	ctx.Values().Set(ctxKeyOpenImExternalInfo, ext)
	ctx.Next()
}
