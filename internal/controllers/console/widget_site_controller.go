package console

import (
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/dto/request"
	"cs-agent/internal/pkg/dto/response"
	"cs-agent/internal/services"

	"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple/web"
	"github.com/mlogclub/simple/web/params"
)

type WidgetSiteController struct {
	Ctx iris.Context
}

func (c *WidgetSiteController) AnyList() *web.JsonResult {
	list, paging := services.WidgetSiteService.FindPageByCnd(params.NewPagedSqlCnd(c.Ctx,
		params.QueryFilter{ParamName: "status"},
		params.QueryFilter{ParamName: "name", Op: params.Like},
		params.QueryFilter{ParamName: "appId", Op: params.Like},
	).Desc("id"))
	results := make([]response.WidgetSiteResponse, 0, len(list))
	for _, item := range list {
		results = append(results, buildWidgetSiteResponse(&item))
	}
	return web.JsonData(&web.PageResult{Results: results, Page: paging})
}

func (c *WidgetSiteController) GetBy(id int64) *web.JsonResult {
	item := services.WidgetSiteService.Get(id)
	if item == nil {
		return web.JsonErrorMsg("widget site not found")
	}
	return web.JsonData(buildWidgetSiteResponse(item))
}

func (c *WidgetSiteController) PostCreate() *web.JsonResult {
	req := request.CreateWidgetSiteRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	item, err := services.WidgetSiteService.CreateSite(req, services.AuthService.GetAuthPrincipal(c.Ctx))
	if err != nil {
		return web.JsonError(err)
	}
	return web.JsonData(buildWidgetSiteResponse(item))
}

func (c *WidgetSiteController) PostUpdate() *web.JsonResult {
	req := request.UpdateWidgetSiteRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	if err := services.WidgetSiteService.UpdateSite(req, services.AuthService.GetAuthPrincipal(c.Ctx)); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}

func (c *WidgetSiteController) PostUpdate_status() *web.JsonResult {
	req := request.UpdateWidgetSiteStatusRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	if err := services.WidgetSiteService.UpdateStatus(req.ID, req.Status, services.AuthService.GetAuthPrincipal(c.Ctx)); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}

func (c *WidgetSiteController) PostDelete() *web.JsonResult {
	req := request.DeleteWidgetSiteRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	if err := services.WidgetSiteService.DeleteSite(req.ID); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}

func buildWidgetSiteResponse(item *models.WidgetSite) response.WidgetSiteResponse {
	ret := response.BuildWidgetSiteResponse(item)
	if aiAgent := services.AIAgentService.Get(item.AIAgentID); aiAgent != nil {
		ret.AIAgentName = aiAgent.Name
	}
	return ret
}
