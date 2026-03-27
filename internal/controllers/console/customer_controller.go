package console

import (
	"cs-agent/internal/builders"
	"cs-agent/internal/pkg/constants"
	"cs-agent/internal/pkg/dto/request"
	"cs-agent/internal/services"

	"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple/web"
	"github.com/mlogclub/simple/web/params"
)

type CustomerController struct {
	Ctx iris.Context
}

func (c *CustomerController) AnyList() *web.JsonResult {
	if err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionCustomerView); err != nil {
		return web.JsonError(err)
	}
	list, paging := services.CustomerService.FindPageByCnd(params.NewPagedSqlCnd(c.Ctx,
		params.QueryFilter{ParamName: "status"},
		params.QueryFilter{ParamName: "gender"},
		params.QueryFilter{ParamName: "companyId"},
		params.QueryFilter{ParamName: "name", Op: params.Like},
		params.QueryFilter{ParamName: "primaryMobile", Op: params.Like},
		params.QueryFilter{ParamName: "primaryEmail", Op: params.Like},
	).Desc("id"))
	return web.JsonData(&web.PageResult{Results: builders.BuildCustomerList(list), Page: paging})
}

func (c *CustomerController) GetBy(id int64) *web.JsonResult {
	if err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionCustomerView); err != nil {
		return web.JsonError(err)
	}
	item := services.CustomerService.Get(id)
	if item == nil {
		return web.JsonErrorMsg("客户不存在")
	}
	ret := builders.BuildCustomerResponse(item)
	return web.JsonData(&ret)
}

func (c *CustomerController) PostCreate() *web.JsonResult {
	if err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionCustomerCreate); err != nil {
		return web.JsonError(err)
	}
	req := request.CreateCustomerRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	item, err := services.CustomerService.CreateCustomer(req, services.AuthService.GetAuthPrincipal(c.Ctx))
	if err != nil {
		return web.JsonError(err)
	}
	ret := builders.BuildCustomerResponse(item)
	return web.JsonData(&ret)
}

func (c *CustomerController) PostUpdate() *web.JsonResult {
	if err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionCustomerUpdate); err != nil {
		return web.JsonError(err)
	}
	req := request.UpdateCustomerRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	if err := services.CustomerService.UpdateCustomer(req, services.AuthService.GetAuthPrincipal(c.Ctx)); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}

func (c *CustomerController) PostDelete() *web.JsonResult {
	if err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionCustomerDelete); err != nil {
		return web.JsonError(err)
	}
	req := request.DeleteCustomerRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	if err := services.CustomerService.DeleteCustomer(req.ID); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}

func (c *CustomerController) PostUpdate_status() *web.JsonResult {
	if err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionCustomerUpdate); err != nil {
		return web.JsonError(err)
	}
	req := request.UpdateCustomerStatusRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	if err := services.CustomerService.UpdateStatus(req.ID, req.Status, services.AuthService.GetAuthPrincipal(c.Ctx)); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}
