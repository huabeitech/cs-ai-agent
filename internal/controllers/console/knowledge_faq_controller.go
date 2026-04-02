package console

import (
	"cs-agent/internal/builders"
	"cs-agent/internal/pkg/constants"
	"cs-agent/internal/pkg/dto/request"
	"cs-agent/internal/pkg/dto/response"
	"cs-agent/internal/services"

	"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple/web"
	"github.com/mlogclub/simple/web/params"
)

type KnowledgeFAQController struct {
	Ctx iris.Context
}

func (c *KnowledgeFAQController) AnyList() *web.JsonResult {
	if _, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionKnowledgeFAQView); err != nil {
		return web.JsonError(err)
	}

	cnd := params.NewPagedSqlCnd(c.Ctx,
		params.QueryFilter{ParamName: "knowledgeBaseId"},
		params.QueryFilter{ParamName: "question", Op: params.Like},
	).Desc("id")
	list, paging := services.KnowledgeFAQService.FindPageByCnd(cnd)
	results := make([]response.KnowledgeFAQResponse, 0, len(list))
	for _, item := range list {
		results = append(results, builders.BuildKnowledgeFAQ(&item))
	}
	return web.JsonData(&web.PageResult{Results: results, Page: paging})
}

func (c *KnowledgeFAQController) GetBy(id int64) *web.JsonResult {
	if _, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionKnowledgeFAQView); err != nil {
		return web.JsonError(err)
	}

	item := services.KnowledgeFAQService.Get(id)
	if item == nil {
		return web.JsonErrorMsg("FAQ不存在")
	}
	return web.JsonData(builders.BuildKnowledgeFAQ(item))
}

func (c *KnowledgeFAQController) PostCreate() *web.JsonResult {
	operator, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionKnowledgeFAQCreate)
	if err != nil {
		return web.JsonError(err)
	}
	req := request.CreateKnowledgeFAQRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	item, err := services.KnowledgeFAQService.CreateKnowledgeFAQ(req, operator)
	if err != nil {
		return web.JsonError(err)
	}
	return web.JsonData(builders.BuildKnowledgeFAQ(item))
}

func (c *KnowledgeFAQController) PostUpdate() *web.JsonResult {
	operator, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionKnowledgeFAQUpdate)
	if err != nil {
		return web.JsonError(err)
	}
	req := request.UpdateKnowledgeFAQRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	if err := services.KnowledgeFAQService.UpdateKnowledgeFAQ(req, operator); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}

func (c *KnowledgeFAQController) PostDelete() *web.JsonResult {
	if _, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionKnowledgeFAQDelete); err != nil {
		return web.JsonError(err)
	}
	var req struct {
		ID int64 `json:"id"`
	}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	if err := services.KnowledgeFAQService.DeleteKnowledgeFAQ(req.ID); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}
