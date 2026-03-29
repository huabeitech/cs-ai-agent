package console

import (
	"cs-agent/internal/builders"
	"cs-agent/internal/pkg/constants"
	"cs-agent/internal/pkg/dto"
	"cs-agent/internal/pkg/dto/request"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/services"

	"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple/web"
	"github.com/mlogclub/simple/web/params"
)

type CustomerController struct {
	Ctx iris.Context
}

func (c *CustomerController) AnyList() *web.JsonResult {
	if _, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionCustomerView); err != nil {
		return web.JsonError(err)
	}
	cnd := params.NewPagedSqlCnd(c.Ctx,
		params.QueryFilter{ParamName: "status"},
		params.QueryFilter{ParamName: "gender"},
		params.QueryFilter{ParamName: "companyId"},
		params.QueryFilter{ParamName: "name", Op: params.Like},
		params.QueryFilter{ParamName: "primaryMobile", Op: params.Like},
		params.QueryFilter{ParamName: "primaryEmail", Op: params.Like},
	).Desc("id")
	// 默认不返回已删除
	cnd.Where("status <> ?", enums.StatusDeleted)

	list, paging := services.CustomerService.FindPageByCnd(cnd)
	return web.JsonData(&web.PageResult{Results: builders.BuildCustomerList(list), Page: paging})
}

func (c *CustomerController) GetBy(id int64) *web.JsonResult {
	if _, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionCustomerView); err != nil {
		return web.JsonError(err)
	}
	item := services.CustomerService.Get(id)
	if item == nil {
		return web.JsonErrorMsg("客户不存在")
	}
	ret := builders.BuildCustomerResponse(item)
	return web.JsonData(&ret)
}

// PostSave_profile POST /save_profile — 客户主信息与联系方式在同一事务中保存。
func (c *CustomerController) PostSave_profile() *web.JsonResult {
	req := request.SaveCustomerProfileRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	createMode := req.ID == nil || *req.ID <= 0
	var user *dto.AuthPrincipal
	var err error
	if createMode {
		user, err = services.AuthService.RequirePermission(c.Ctx, constants.PermissionCustomerCreate)
	} else {
		user, err = services.AuthService.RequirePermission(c.Ctx, constants.PermissionCustomerUpdate)
	}
	if err != nil {
		return web.JsonError(err)
	}
	item, err := services.CustomerService.SaveCustomerProfile(req, user)
	if err != nil {
		return web.JsonError(err)
	}
	ret := builders.BuildCustomerResponse(item)
	return web.JsonData(&ret)
}

func (c *CustomerController) PostCreate() *web.JsonResult {
	user, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionCustomerCreate)
	if err != nil {
		return web.JsonError(err)
	}
	req := request.CreateCustomerRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	item, err := services.CustomerService.CreateCustomer(req, user)
	if err != nil {
		return web.JsonError(err)
	}
	ret := builders.BuildCustomerResponse(item)
	return web.JsonData(&ret)
}

func (c *CustomerController) PostUpdate() *web.JsonResult {
	user, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionCustomerUpdate)
	if err != nil {
		return web.JsonError(err)
	}
	req := request.UpdateCustomerRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	if err := services.CustomerService.UpdateCustomer(req, user); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}

func (c *CustomerController) PostDelete() *web.JsonResult {
	user, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionCustomerDelete)
	if err != nil {
		return web.JsonError(err)
	}
	req := request.DeleteCustomerRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	if err := services.CustomerService.DeleteCustomer(req.ID, *user); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}

func (c *CustomerController) PostUpdate_status() *web.JsonResult {
	user, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionCustomerUpdate)
	if err != nil {
		return web.JsonError(err)
	}
	req := request.UpdateCustomerStatusRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	if err := services.CustomerService.UpdateStatus(req.ID, req.Status, user); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}
