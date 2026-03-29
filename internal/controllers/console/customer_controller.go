package console

import (
	"cs-agent/internal/builders"
	"cs-agent/internal/pkg/constants"
	"cs-agent/internal/pkg/dto"
	"cs-agent/internal/pkg/dto/request"
	"cs-agent/internal/services"

	"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple/web"
	"github.com/mlogclub/simple/web/params"
)

type CustomerController struct {
	Ctx iris.Context
}

// PostList POST /list — 客户分页列表，参数见 request.CustomerListRequest。
func (c *CustomerController) PostList() *web.JsonResult {
	if _, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionCustomerView); err != nil {
		return web.JsonError(err)
	}
	var req request.CustomerListRequest
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	list, paging := services.CustomerService.ListCustomers(req)
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
	ret := builders.BuildCustomer(item)
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
	ret := builders.BuildCustomer(item)
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
	ret := builders.BuildCustomer(item)
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
