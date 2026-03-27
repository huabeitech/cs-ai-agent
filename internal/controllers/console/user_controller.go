package console

import (
	"cs-agent/internal/builders"
	"cs-agent/internal/pkg/constants"
	"cs-agent/internal/pkg/dto/request"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/services"

	"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple/web"
	"github.com/mlogclub/simple/web/params"
)

type UserController struct {
	Ctx iris.Context
}

func (c *UserController) AnyList() *web.JsonResult {
	if _, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionUserView); err != nil {
		return web.JsonError(err)
	}

	cnd := params.NewPagedSqlCnd(c.Ctx,
		params.QueryFilter{ParamName: "status"},
		params.QueryFilter{ParamName: "username", Op: params.Like},
		params.QueryFilter{ParamName: "nickname", Op: params.Like},
	).Desc("id")
	cnd.Where("status <> ?", enums.StatusDeleted)
	list, paging := services.UserService.FindPageByCnd(cnd)
	results := builders.BuildUserList(list, builders.UserBuildOptions{
		Roles:       true,
		Permissions: false,
	})
	return web.JsonData(&web.PageResult{Results: results, Page: paging})
}

func (c *UserController) AnyList_all() *web.JsonResult {
	if _, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionUserView); err != nil {
		return web.JsonError(err)
	}

	cnd := params.NewSqlCnd(c.Ctx,
		params.QueryFilter{ParamName: "status"},
		params.QueryFilter{ParamName: "username", Op: params.Like},
		params.QueryFilter{ParamName: "nickname", Op: params.Like},
	).Desc("id")
	cnd.Where("status <> ?", enums.StatusDeleted)

	list := services.UserService.Find(cnd)
	results := builders.BuildUserList(list, builders.UserBuildOptions{
		Roles:       true,
		Permissions: false,
	})
	return web.JsonData(results)
}

func (c *UserController) GetBy(id int64) *web.JsonResult {
	if _, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionUserView); err != nil {
		return web.JsonError(err)
	}

	item := services.UserService.Get(id)
	if item == nil {
		return web.JsonErrorMsg("用户不存在")
	}
	return web.JsonData(builders.BuildUserResponse(item, builders.UserBuildOptions{
		Roles:       true,
		Permissions: true,
	}))
}

func (c *UserController) PostCreate() *web.JsonResult {
	operator, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionUserCreate)
	if err != nil {
		return web.JsonError(err)
	}

	req := request.CreateUserRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	user, err := services.UserService.CreateUser(req, operator)
	if err != nil {
		return web.JsonError(err)
	}
	return web.JsonData(builders.BuildUserResponse(user, builders.UserBuildOptions{
		Roles:       true,
		Permissions: true,
	}))
}

func (c *UserController) PostUpdate() *web.JsonResult {
	operator, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionUserUpdate)
	if err != nil {
		return web.JsonError(err)
	}

	req := request.UpdateUserRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	if err := services.UserService.UpdateUser(req, operator); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}

func (c *UserController) PostDelete() *web.JsonResult {
	operator, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionUserDelete)
	if err != nil {
		return web.JsonError(err)
	}

	req := request.DeleteUserRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	if err := services.UserService.DeleteUser(req.ID, operator); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}

func (c *UserController) PostUpdate_status() *web.JsonResult {
	operator, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionUserUpdate)
	if err != nil {
		return web.JsonError(err)
	}

	req := request.UpdateUserStatusRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	if err := services.UserService.UpdateStatus(req.ID, req.Status, operator); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}

func (c *UserController) PostReset_password() *web.JsonResult {
	operator, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionUserUpdate)
	if err != nil {
		return web.JsonError(err)
	}

	var req struct {
		UserID int64 `json:"userId"`
	}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	password, err := services.UserService.ResetPassword(req.UserID, operator)
	if err != nil {
		return web.JsonError(err)
	}
	return web.JsonData(map[string]any{
		"password": password,
	})
}

func (c *UserController) PostChange_password() *web.JsonResult {
	principal := services.AuthService.GetAuthPrincipal(c.Ctx)
	if principal == nil {
		if _, err := services.AuthService.Authenticate(c.Ctx); err != nil {
			return web.JsonError(err)
		}
		principal = services.AuthService.GetAuthPrincipal(c.Ctx)
	}
	if principal == nil {
		return web.JsonErrorMsg("未登录或登录已过期")
	}

	req := request.ChangePasswordRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	if err := services.UserService.ChangeOwnPassword(req.Password, principal); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}

func (c *UserController) PostAssign_role() *web.JsonResult {
	operator, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionUserAssignRole)
	if err != nil {
		return web.JsonError(err)
	}

	req := request.AssignRoleRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	if err := services.UserService.AssignRoles(req.UserID, req.RoleIDs, operator); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}
