package dashboard

import (
	"cs-agent/internal/builders"
	"cs-agent/internal/pkg/constants"
	"cs-agent/internal/pkg/dto/request"
	"cs-agent/internal/pkg/dto/response"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/services"

	"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple/web"
	"github.com/mlogclub/simple/web/params"
)

type KnowledgeDocumentController struct {
	Ctx iris.Context
}

func (c *KnowledgeDocumentController) AnyList() *web.JsonResult {
	if _, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionKnowledgeDocumentView); err != nil {
		return web.JsonError(err)
	}

	cnd := params.NewPagedSqlCnd(c.Ctx,
		params.QueryFilter{ParamName: "knowledgeBaseId"},
		params.QueryFilter{ParamName: "title", Op: params.Like},
	).Desc("id")

	if status, ok := params.GetInt64(c.Ctx, "status"); ok {
		cnd.Where("status = ?", status)
	} else {
		cnd.Where("status != ?", enums.StatusDeleted)
	}
	if indexStatus, ok := params.Get(c.Ctx, "indexStatus"); ok {
		if !enums.IsValidKnowledgeDocumentIndexStatus(indexStatus) {
			return web.JsonErrorMsg("indexStatus参数不合法")
		}
		cnd.Where("index_status = ?", indexStatus)
	}

	list, paging := services.KnowledgeDocumentService.FindPageListByCnd(cnd)
	results := make([]response.KnowledgeDocumentListResponse, 0, len(list))
	for _, item := range list {
		results = append(results, builders.BuildKnowledgeDocumentList(&item))
	}
	return web.JsonData(&web.PageResult{Results: results, Page: paging})
}

func (c *KnowledgeDocumentController) GetBy(id int64) *web.JsonResult {
	if _, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionKnowledgeDocumentView); err != nil {
		return web.JsonError(err)
	}

	item := services.KnowledgeDocumentService.Get(id)
	if item == nil {
		return web.JsonErrorMsg("文档不存在")
	}
	return web.JsonData(builders.BuildKnowledgeDocument(item))
}

func (c *KnowledgeDocumentController) PostCreate() *web.JsonResult {
	operator, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionKnowledgeDocumentCreate)
	if err != nil {
		return web.JsonError(err)
	}

	req := request.CreateKnowledgeDocumentRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	item, err := services.KnowledgeDocumentService.CreateKnowledgeDocument(req, operator)
	if err != nil {
		return web.JsonError(err)
	}
	return web.JsonData(builders.BuildKnowledgeDocument(item))
}

func (c *KnowledgeDocumentController) PostUpdate() *web.JsonResult {
	operator, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionKnowledgeDocumentUpdate)
	if err != nil {
		return web.JsonError(err)
	}

	req := request.UpdateKnowledgeDocumentRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	if err := services.KnowledgeDocumentService.UpdateKnowledgeDocument(req, operator); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}

func (c *KnowledgeDocumentController) PostDelete() *web.JsonResult {
	if _, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionKnowledgeDocumentDelete); err != nil {
		return web.JsonError(err)
	}

	var req struct {
		ID int64 `json:"id"`
	}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}

	if err := services.KnowledgeDocumentService.DeleteKnowledgeDocument(req.ID); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}
