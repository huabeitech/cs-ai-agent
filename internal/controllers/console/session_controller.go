package console

import (
	"cs-agent/internal/pkg/constants"
	"cs-agent/internal/pkg/dto/request"
	"cs-agent/internal/pkg/dto/response"
	"cs-agent/internal/pkg/utils"
	"cs-agent/internal/services"
	"time"

	"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple/web"
	"github.com/mlogclub/simple/web/params"
)

type SessionController struct {
	Ctx iris.Context
}

func (c *SessionController) AnyList() *web.JsonResult {
	if err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionSessionView); err != nil {
		return web.JsonError(err)
	}

	cnd := params.NewPagedSqlCnd(c.Ctx,
		params.QueryFilter{ParamName: "userId"},
		params.QueryFilter{ParamName: "tokenType"},
		params.QueryFilter{ParamName: "clientType"},
	).Desc("id")
	list, paging := services.LoginSessionService.FindPageByCnd(cnd)
	results := make([]response.SessionResponse, 0, len(list))
	for _, item := range list {
		username := ""
		if user := services.UserService.Get(item.UserID); user != nil {
			username = user.Username
		}
		results = append(results, response.SessionResponse{
			ID:         item.ID,
			UserID:     item.UserID,
			Username:   username,
			TokenType:  item.TokenType,
			ClientType: item.ClientType,
			ClientIP:   item.ClientIP,
			UserAgent:  item.UserAgent,
			ExpiredAt:  item.ExpiredAt.Format(time.DateTime),
			RevokedAt:  utils.FormatTimePtr(item.RevokedAt),
			LastSeenAt: utils.FormatTimePtr(item.LastSeenAt),
		})
	}
	return web.JsonData(&web.PageResult{Results: results, Page: paging})
}

func (c *SessionController) PostRevoke() *web.JsonResult {
	if err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionSessionRevoke); err != nil {
		return web.JsonError(err)
	}

	req := request.RevokeSessionRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	principal := services.AuthService.GetAuthPrincipal(c.Ctx)
	if err := services.LoginSessionService.Revoke(req.ID, principal.UserID, principal.Username); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}

func (c *SessionController) PostRevokeByUser() *web.JsonResult {
	if err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionSessionRevoke); err != nil {
		return web.JsonError(err)
	}

	req := request.RevokeUserSessionsRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	principal := services.AuthService.GetAuthPrincipal(c.Ctx)
	if err := services.LoginSessionService.RevokeByUser(req.UserID, principal.UserID, principal.Username); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}
