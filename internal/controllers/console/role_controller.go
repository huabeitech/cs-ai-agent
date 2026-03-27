package console

import (
	"cs-agent/internal/pkg/constants"
	"cs-agent/internal/pkg/dto/request"
	"cs-agent/internal/pkg/dto/response"
	"cs-agent/internal/services"

	"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web"
	"github.com/mlogclub/simple/web/params"
)

type RoleController struct {
	Ctx iris.Context
}

func (c *RoleController) AnyList() *web.JsonResult {
	if _, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionRoleView); err != nil {
		return web.JsonError(err)
	}
	cnd := params.NewPagedSqlCnd(c.Ctx,
		params.QueryFilter{ParamName: "status"},
		params.QueryFilter{ParamName: "code", Op: params.Like},
	).Asc("sort_no").Desc("id")
	list, paging := services.RoleService.FindPageByCnd(cnd)
	results := make([]response.RoleResponse, 0, len(list))
	for _, item := range list {
		results = append(results, response.RoleResponse{
			ID:       item.ID,
			Name:     item.Name,
			Code:     item.Code,
			Status:   item.Status,
			IsSystem: item.IsSystem,
			SortNo:   item.SortNo,
		})
	}
	return web.JsonData(&web.PageResult{Results: results, Page: paging})
}

func (c *RoleController) GetList_all() *web.JsonResult {
	if _, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionRoleView); err != nil {
		return web.JsonError(err)
	}

	list := services.RoleService.Find(sqls.NewCnd().Asc("sort_no").Desc("id"))
	results := make([]response.RoleResponse, 0, len(list))
	for _, item := range list {
		results = append(results, response.RoleResponse{
			ID:       item.ID,
			Name:     item.Name,
			Code:     item.Code,
			Status:   item.Status,
			IsSystem: item.IsSystem,
			SortNo:   item.SortNo,
		})
	}
	return web.JsonData(results)
}

func (c *RoleController) GetBy(id int64) *web.JsonResult {
	if _, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionRoleView); err != nil {
		return web.JsonError(err)
	}

	item := services.RoleService.Get(id)
	if item == nil {
		return web.JsonErrorMsg("角色不存在")
	}

	permissionCodes := make([]string, 0)
	list := services.RolePermissionService.Find(sqls.NewCnd().Eq("role_id", item.ID))
	for _, relation := range list {
		permission := services.PermissionService.Get(relation.PermissionID)
		if permission != nil {
			permissionCodes = append(permissionCodes, permission.Code)
		}
	}
	return web.JsonData(&response.RoleResponse{
		ID:          item.ID,
		Name:        item.Name,
		Code:        item.Code,
		Status:      item.Status,
		IsSystem:    item.IsSystem,
		SortNo:      item.SortNo,
		Permissions: permissionCodes,
	})
}

func (c *RoleController) PostCreate() *web.JsonResult {
	operator, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionRoleCreate)
	if err != nil {
		return web.JsonError(err)
	}

	req := request.CreateRoleRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	role, err := services.RoleService.CreateRole(req, operator)
	if err != nil {
		return web.JsonError(err)
	}
	return web.JsonData(&response.RoleResponse{
		ID:       role.ID,
		Name:     role.Name,
		Code:     role.Code,
		Status:   role.Status,
		IsSystem: role.IsSystem,
		SortNo:   role.SortNo,
	})
}

func (c *RoleController) PostUpdate() *web.JsonResult {
	operator, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionRoleUpdate)
	if err != nil {
		return web.JsonError(err)
	}

	req := request.UpdateRoleRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	if err := services.RoleService.UpdateRole(req, operator); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}

func (c *RoleController) PostDelete() *web.JsonResult {
	if _, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionRoleDelete); err != nil {
		return web.JsonError(err)
	}

	req := request.DeleteRoleRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	if err := services.RoleService.DeleteRole(req.ID); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}

func (c *RoleController) PostUpdate_status() *web.JsonResult {
	operator, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionRoleUpdate)
	if err != nil {
		return web.JsonError(err)
	}

	req := request.UpdateRoleStatusRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	if err := services.RoleService.UpdateStatus(req.ID, req.Status, operator); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}

func (c *RoleController) PostAssign_permission() *web.JsonResult {
	operator, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionRoleAssignPermission)
	if err != nil {
		return web.JsonError(err)
	}

	req := request.AssignPermissionRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	if err := services.RoleService.AssignPermissions(req.RoleID, req.PermissionIDs, operator); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}

func (c *RoleController) PostUpdate_sort() *web.JsonResult {
	var ids []int64
	if err := c.Ctx.ReadJSON(&ids); err != nil {
		return web.JsonError(err)
	}
	if err := services.RoleService.UpdateSort(ids); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}
