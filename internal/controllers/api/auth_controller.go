package api

import (
	"cs-agent/internal/pkg/config"
	"cs-agent/internal/pkg/dto/request"
	"cs-agent/internal/services"
	"net/url"
	"strings"

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
	b.Handle("GET", "/wxwork/login", "GetWxWorkLogin")
	b.Handle("GET", "/wxwork/callback", "GetWxWorkCallback")
	b.Handle("POST", "/wxwork/exchange", "PostWxWorkExchange")
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

func (c *AuthController) GetWxwork_login() {
	loginURL, err := services.AuthService.BuildWxWorkLoginURL(c.Ctx.URLParam("next"))
	if err != nil {
		c.redirectWxWorkError(err.Error())
		return
	}
	c.Ctx.Redirect(loginURL, iris.StatusFound)
}

func (c *AuthController) GetWxwork_callback() {
	ticket, next, err := services.AuthService.LoginByWxWork(
		c.Ctx.URLParam("code"),
		c.Ctx.URLParam("state"),
		c.Cfg.Auth,
		c.Ctx.RemoteAddr(),
		c.Ctx.GetHeader("User-Agent"),
	)
	if err != nil {
		c.redirectWxWorkError(err.Error())
		return
	}
	c.Ctx.Redirect("/login/wxwork/callback?ticket="+url.QueryEscape(ticket)+"&next="+url.QueryEscape(next), iris.StatusFound)
}

func (c *AuthController) PostWxwork_exchange() *web.JsonResult {
	req := request.WxWorkExchangeRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	ret, err := services.AuthService.ExchangeWxWorkLoginTicket(req.Ticket)
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

func (c *AuthController) redirectWxWorkError(message string) {
	if idx := strings.Index(message, ": "); idx >= 0 {
		message = message[idx+2:]
	}
	c.Ctx.Redirect("/login?wxworkError="+url.QueryEscape(message), iris.StatusFound)
}
