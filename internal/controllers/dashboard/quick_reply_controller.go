package dashboard

import (
	"cs-agent/internal/pkg/constants"
	"cs-agent/internal/pkg/dto/request"
	"cs-agent/internal/pkg/dto/response"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/services"

	"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web"
	"github.com/mlogclub/simple/web/params"
)

type QuickReplyController struct {
	Ctx iris.Context
}

func (c *QuickReplyController) AnyList() *web.JsonResult {
	if _, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionQuickReplyView); err != nil {
		return web.JsonError(err)
	}

	cnd := params.NewPagedSqlCnd(c.Ctx,
		params.QueryFilter{ParamName: "status"},
		params.QueryFilter{ParamName: "groupName"},
		params.QueryFilter{ParamName: "title", Op: params.Like},
	).Asc("sort_no").Desc("id")

	list, paging := services.QuickReplyService.FindPageByCnd(cnd)
	results := make([]response.QuickReplyResponse, 0, len(list))
	for _, item := range list {
		results = append(results, response.QuickReplyResponse{
			ID:        item.ID,
			GroupName: item.GroupName,
			Title:     item.Title,
			Content:   item.Content,
			Status:    item.Status,
			SortNo:    item.SortNo,
		})
	}
	return web.JsonData(&web.PageResult{Results: results, Page: paging})
}

func (c *QuickReplyController) GetList_all() *web.JsonResult {
	if _, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionQuickReplyView); err != nil {
		return web.JsonError(err)
	}
	list := services.QuickReplyService.Find(sqls.NewCnd().Eq("status", enums.StatusOk).Asc("sort_no").Desc("id"))
	results := make([]response.QuickReplyResponse, 0, len(list))
	for _, item := range list {
		results = append(results, response.QuickReplyResponse{
			ID:        item.ID,
			GroupName: item.GroupName,
			Title:     item.Title,
			Content:   item.Content,
			Status:    item.Status,
			SortNo:    item.SortNo,
		})
	}
	return web.JsonData(results)
}

func (c *QuickReplyController) PostCreate() *web.JsonResult {
	user, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionQuickReplyCreate)
	if err != nil {
		return web.JsonError(err)
	}

	req := request.CreateQuickReplyRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	item, err := services.QuickReplyService.CreateQuickReply(req, user)
	if err != nil {
		return web.JsonError(err)
	}
	return web.JsonData(&response.QuickReplyResponse{
		ID:        item.ID,
		GroupName: item.GroupName,
		Title:     item.Title,
		Content:   item.Content,
		Status:    item.Status,
		SortNo:    item.SortNo,
	})
}

func (c *QuickReplyController) PostUpdate() *web.JsonResult {
	user, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionQuickReplyUpdate)
	if err != nil {
		return web.JsonError(err)
	}

	req := request.UpdateQuickReplyRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	if err := services.QuickReplyService.UpdateQuickReply(req, user); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}

func (c *QuickReplyController) PostDelete() *web.JsonResult {
	if _, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionQuickReplyDelete); err != nil {
		return web.JsonError(err)
	}

	req := request.DeleteQuickReplyRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	if err := services.QuickReplyService.DeleteQuickReply(req.ID); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}
