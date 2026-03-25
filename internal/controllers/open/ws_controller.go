package open

import (
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/errorsx"
	"cs-agent/internal/services"
	"log/slog"

	"github.com/kataras/iris/v12"
)

func HandleImWebsocket(ctx iris.Context) {
	site, err := resolveEnabledWidgetSiteForWS(ctx)
	if err != nil {
		_ = ctx.StopWithJSON(iris.StatusBadRequest, map[string]any{
			"message": err.Error(),
		})
		return
	}

	principal := services.AuthService.GetAuthPrincipal(ctx)
	if principal == nil {
		if _, err := services.AuthService.GetImPrincipal(ctx); err != nil {
			_ = ctx.StopWithJSON(iris.StatusUnauthorized, map[string]any{
				"message": err.Error(),
			})
			return
		}
		principal = services.AuthService.GetAuthPrincipal(ctx)
	}
	if err := services.WsService.UpgradeUserConnection(ctx, principal); err != nil {
		slog.Error("upgrade open im websocket failed", "error", err, "path", ctx.Path(), "appId", site.AppID)
		ctx.StopExecution()
		return
	}
}

func resolveEnabledWidgetSiteForWS(ctx iris.Context) (*models.WidgetSite, error) {
	site, rsp := requireEnabledWidgetSite(ctx)
	if rsp == nil {
		return site, nil
	}
	return nil, errorsx.InvalidParam("接入站点不存在或已停用")
}
