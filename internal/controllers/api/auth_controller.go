package api

import (
	"cs-agent/internal/pkg/config"
	"cs-agent/internal/pkg/dto/request"
	"cs-agent/internal/services"
	"net/url"
	"strings"

	"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple/web"
	"github.com/mlogclub/simple/web/params"
)

type AuthController struct {
	Ctx iris.Context
}

func (c *AuthController) PostLogin() *web.JsonResult {
	cfg := config.Current()
	req := request.LoginRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}

	ret, err := services.AuthService.Login(req, cfg.Auth, c.Ctx.RemoteAddr(), c.Ctx.GetHeader("User-Agent"))
	if err != nil {
		return web.JsonError(err)
	}
	return web.JsonData(ret)
}

func (c *AuthController) PostRefresh_token() *web.JsonResult {
	cfg := config.Current()
	req := request.RefreshTokenRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}

	ret, err := services.AuthService.RefreshToken(req.RefreshToken, cfg.Auth, c.Ctx.RemoteAddr(), c.Ctx.GetHeader("User-Agent"))
	if err != nil {
		return web.JsonError(err)
	}
	return web.JsonData(ret)
}

func (c *AuthController) GetWxwork_login() {
	loginURL, err := services.WxWorkLoginService.BuildWxWorkLoginURL(c.Ctx.URLParam("next"))
	if err != nil {
		c.redirectWxWorkError(err.Error())
		return
	}
	c.Ctx.Redirect(loginURL, iris.StatusFound)
}

func (c *AuthController) GetWxwork_qr_login() {
	loginURL, err := services.WxWorkLoginService.BuildWxWorkQRCodeLoginURL(c.Ctx.URLParam("next"))
	if err != nil {
		c.redirectWxWorkError(err.Error())
		return
	}
	c.Ctx.Redirect(loginURL, iris.StatusFound)
}

func (c *AuthController) GetWxwork_callback() {
	cfg := config.Current()
	ticket, next, err := services.WxWorkLoginService.LoginByWxWork(
		c.Ctx.URLParam("code"),
		c.Ctx.URLParam("state"),
		cfg.Auth,
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
	ret, err := services.WxWorkLoginService.ExchangeWxWorkLoginTicket(req.Ticket)
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
