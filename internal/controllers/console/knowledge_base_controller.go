package console

import (
	"context"
	"log/slog"

	"cs-agent/internal/ai/rag"
	"cs-agent/internal/builders"
	"cs-agent/internal/pkg/constants"
	"cs-agent/internal/pkg/dto/request"
	"cs-agent/internal/pkg/dto/response"
	"cs-agent/internal/repositories"
	"cs-agent/internal/services"

	"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web"
	"github.com/mlogclub/simple/web/params"
)

type KnowledgeBaseController struct {
	Ctx iris.Context
}

func (c *KnowledgeBaseController) AnyList() *web.JsonResult {
	if _, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionKnowledgeBaseView); err != nil {
		return web.JsonError(err)
	}

	cnd := params.NewPagedSqlCnd(c.Ctx,
		params.QueryFilter{ParamName: "status"},
		params.QueryFilter{ParamName: "name", Op: params.Like},
	).Asc("sort_no").Desc("id")
	list, paging := services.KnowledgeBaseService.FindPageByCnd(cnd)
	results := make([]response.KnowledgeBaseResponse, 0, len(list))
	for _, item := range list {
		docCount := repositories.KnowledgeDocumentRepository.CountByKnowledgeBaseID(sqls.DB(), item.ID)
		faqCount := repositories.KnowledgeFAQRepository.CountByKnowledgeBaseID(sqls.DB(), item.ID)
		resp := builders.BuildKnowledgeBase(&item)
		resp.DocumentCount = docCount
		resp.FAQCount = faqCount
		results = append(results, resp)
	}
	return web.JsonData(&web.PageResult{Results: results, Page: paging})
}

func (c *KnowledgeBaseController) AnyList_all() *web.JsonResult {
	if _, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionKnowledgeBaseView); err != nil {
		return web.JsonError(err)
	}

	list := services.KnowledgeBaseService.Find(params.NewSqlCnd(c.Ctx,
		params.QueryFilter{ParamName: "status"},
	).Asc("sort_no").Desc("id"))
	results := make([]response.KnowledgeBaseResponse, 0, len(list))
	for _, item := range list {
		resp := builders.BuildKnowledgeBase(&item)
		results = append(results, resp)
	}
	return web.JsonData(results)
}

func (c *KnowledgeBaseController) GetBy(id int64) *web.JsonResult {
	if _, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionKnowledgeBaseView); err != nil {
		return web.JsonError(err)
	}

	item := services.KnowledgeBaseService.Get(id)
	if item == nil {
		return web.JsonErrorMsg("知识库不存在")
	}
	resp := builders.BuildKnowledgeBase(item)
	resp.DocumentCount = repositories.KnowledgeDocumentRepository.CountByKnowledgeBaseID(sqls.DB(), item.ID)
	resp.FAQCount = repositories.KnowledgeFAQRepository.CountByKnowledgeBaseID(sqls.DB(), item.ID)
	return web.JsonData(resp)
}

func (c *KnowledgeBaseController) PostCreate() *web.JsonResult {
	user, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionKnowledgeBaseCreate)
	if err != nil {
		return web.JsonError(err)
	}

	req := request.CreateKnowledgeBaseRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	item, err := services.KnowledgeBaseService.CreateKnowledgeBase(req, user)
	if err != nil {
		return web.JsonError(err)
	}
	return web.JsonData(builders.BuildKnowledgeBase(item))
}

func (c *KnowledgeBaseController) PostUpdate() *web.JsonResult {
	user, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionKnowledgeBaseUpdate)
	if err != nil {
		return web.JsonError(err)
	}

	req := request.UpdateKnowledgeBaseRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	if err := services.KnowledgeBaseService.UpdateKnowledgeBase(req, user); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}

func (c *KnowledgeBaseController) PostDelete() *web.JsonResult {
	if _, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionKnowledgeBaseDelete); err != nil {
		return web.JsonError(err)
	}

	var req struct {
		ID int64 `json:"id"`
	}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	if err := services.KnowledgeBaseService.DeleteKnowledgeBase(req.ID); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}

func (c *KnowledgeBaseController) PostUpdate_sort() *web.JsonResult {
	var ids []int64
	if err := c.Ctx.ReadJSON(&ids); err != nil {
		return web.JsonError(err)
	}
	if err := services.KnowledgeBaseService.UpdateSort(ids); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}

func (c *KnowledgeBaseController) PostRebuild_index() *web.JsonResult {
	if _, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionKnowledgeBaseUpdate); err != nil {
		return web.JsonError(err)
	}

	var req struct {
		ID int64 `json:"id"`
	}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}

	knowledgeBase := services.KnowledgeBaseService.Get(req.ID)
	if knowledgeBase == nil {
		return web.JsonErrorMsg("知识库不存在")
	}

	go func() {
		ctx := context.Background()
		if err := rag.Index.RebuildKnowledgeBaseIndex(ctx, req.ID); err != nil {
			slog.Error("Failed to rebuild knowledge base index", "knowledge_base_id", req.ID, "error", err)
		}
	}()

	return web.JsonSuccess()
}
