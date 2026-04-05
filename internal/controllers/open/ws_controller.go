package open

import (
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/errorsx"
	"cs-agent/internal/pkg/openidentity"
	"cs-agent/internal/services"
	"log/slog"

	"github.com/kataras/iris/v12"
)

func HandleImWebsocket(ctx iris.Context) {
	channel, err := resolveEnabledChannelForWS(ctx)
	if err != nil {
		_ = ctx.StopWithJSON(iris.StatusBadRequest, map[string]any{
			"message": err.Error(),
		})
		return
	}

	// 与 Open IM HTTP 一致：优先站内 AuthPrincipal；否则使用外部访客身份（Header/query，见 openidentity）。
	// 二者不应在业务上同时作为「客户身份」使用；本入口在 principal 非空时不再解析 external，避免语义冲突。
	principal := services.AuthService.GetAuthPrincipal(ctx)
	var external *openidentity.ExternalInfo
	if principal == nil {
		ext, err := openidentity.GetExternalInfo(ctx)
		if err != nil {
			_ = ctx.StopWithJSON(iris.StatusUnauthorized, map[string]any{
				"message": err.Error(),
			})
			return
		}
		external = ext
	}
	if err := services.WsService.UpgradeUserConnection(ctx, principal, external); err != nil {
		slog.Error("upgrade open im websocket failed", "error", err, "path", ctx.Path(), "channelId", channel.ChannelID, "channel_id", channel.ID)
		ctx.StopExecution()
		return
	}
}

func resolveEnabledChannelForWS(ctx iris.Context) (*models.Channel, error) {
	channel, rsp := requireEnabledChannel(ctx)
	if rsp == nil {
		return channel, nil
	}
	return nil, errorsx.InvalidParam("接入渠道不存在或已停用")
}
