package api

import (
	"cs-agent/internal/pkg/config"
	"cs-agent/internal/pkg/dto/request"
	"cs-agent/internal/services"

	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"github.com/mlogclub/simple/web"
	"github.com/mlogclub/simple/web/params"
)

type AuthController struct {
	Ctx iris.Context
	Cfg *config.Config
}

func (c *AuthController) BeforeActivation(b mvc.BeforeActivation) {
	b.Handle("POST", "/login", "PostLogin")
	b.Handle("POST", "/refresh-token", "PostRefreshToken")
	b.Handle("POST", "/logout", "PostLogout")
	b.Handle("GET", "/profile", "GetProfile")
}

func (c *AuthController) PostLogin() *web.JsonResult {
	req := request.LoginRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}

	ret, err := services.AuthService.Login(req, c.Cfg.Auth, c.Ctx.RemoteAddr(), c.Ctx.GetHeader("User-Agent"))
	if err != nil {
		return web.JsonError(err)
	}
	return web.JsonData(ret)
}

func (c *AuthController) PostRefreshToken() *web.JsonResult {
	req := request.RefreshTokenRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}

	ret, err := services.AuthService.RefreshToken(req.RefreshToken, c.Cfg.Auth, c.Ctx.RemoteAddr(), c.Ctx.GetHeader("User-Agent"))
	if err != nil {
		return web.JsonError(err)
	}
	return web.JsonData(ret)
}

func (c *AuthController) PostLogout() *web.JsonResult {
	req := request.LogoutRequest{}
	if c.Ctx.GetContentLength() > 0 {
		if err := params.ReadJSON(c.Ctx, &req); err != nil {
			return web.JsonError(err)
		}
	}

	if err := services.AuthService.Logout(c.Ctx.GetHeader("Authorization"), req.RefreshToken); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}

func (c *AuthController) GetProfile() *web.JsonResult {
	ret, err := services.AuthService.CurrentProfile(c.Ctx)
	if err != nil {
		return web.JsonError(err)
	}
	return web.JsonData(ret)
}
