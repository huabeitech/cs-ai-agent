package api

import (
	"cs-agent/internal/services"
	"log/slog"

	"github.com/kataras/iris/v12"
)

func HandleAdminWebsocket(ctx iris.Context) {
	principal := services.AuthService.GetAuthPrincipal(ctx)
	if principal == nil {
		if _, err := services.AuthService.Authenticate(ctx); err != nil {
			_ = ctx.StopWithJSON(iris.StatusUnauthorized, map[string]any{
				"message": err.Error(),
			})
			return
		}
		principal = services.AuthService.GetAuthPrincipal(ctx)
	}
	if err := services.WsService.UpgradeAdminConnection(ctx, principal); err != nil {
		slog.Error("upgrade admin websocket failed", "error", err, "path", ctx.Path())
		ctx.StopExecution()
		return
	}
}
