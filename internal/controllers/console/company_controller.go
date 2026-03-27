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

type CompanyController struct {
	Ctx iris.Context
}

func (c *CompanyController) AnyList() *web.JsonResult {
	if err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionCompanyView); err != nil {
		return web.JsonError(err)
	}
	list, paging := services.CompanyService.FindPageByCnd(params.NewPagedSqlCnd(c.Ctx,
		params.QueryFilter{ParamName: "status"},
		params.QueryFilter{ParamName: "name", Op: params.Like},
		params.QueryFilter{ParamName: "code", Op: params.Like},
	).Desc("id"))
	return web.JsonData(&web.PageResult{Results: builders.BuildCompanyList(list), Page: paging})
}

func (c *CompanyController) GetBy(id int64) *web.JsonResult {
	if err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionCompanyView); err != nil {
		return web.JsonError(err)
	}
	item := services.CompanyService.Get(id)
	if item == nil {
		return web.JsonErrorMsg("公司不存在")
	}
	ret := builders.BuildCompanyResponse(item)
	return web.JsonData(&ret)
}

func (c *CompanyController) PostCreate() *web.JsonResult {
	if err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionCompanyCreate); err != nil {
		return web.JsonError(err)
	}
	req := request.CreateCompanyRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	item, err := services.CompanyService.CreateCompany(req, services.AuthService.GetAuthPrincipal(c.Ctx))
	if err != nil {
		return web.JsonError(err)
	}
	ret := builders.BuildCompanyResponse(item)
	return web.JsonData(&ret)
}

func (c *CompanyController) PostUpdate() *web.JsonResult {
	if err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionCompanyUpdate); err != nil {
		return web.JsonError(err)
	}
	req := request.UpdateCompanyRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	if err := services.CompanyService.UpdateCompany(req, services.AuthService.GetAuthPrincipal(c.Ctx)); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}

func (c *CompanyController) PostDelete() *web.JsonResult {
	if err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionCompanyDelete); err != nil {
		return web.JsonError(err)
	}
	req := request.DeleteCompanyRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	if err := services.CompanyService.DeleteCompany(req.ID); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}

func (c *CompanyController) PostUpdate_status() *web.JsonResult {
	if err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionCompanyUpdate); err != nil {
		return web.JsonError(err)
	}
	req := request.UpdateCompanyStatusRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	if err := services.CompanyService.UpdateStatus(req.ID, req.Status, services.AuthService.GetAuthPrincipal(c.Ctx)); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}
