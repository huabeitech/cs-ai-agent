package middleware

import (
	"cs-agent/internal/services"

	"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple/web"
)

func AuthMiddleware(ctx iris.Context) {
	if !authenticateRequest(ctx) {
		return
	}
	ctx.Next()
}

func authenticateRequest(ctx iris.Context) bool {
	if _, err := services.AuthService.Authenticate(ctx); err != nil {
		_ = ctx.JSON(web.JsonError(err))
		ctx.StopExecution()
		return false
	}
	return true
}
