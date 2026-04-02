package console

import (
	"context"

	"cs-agent/internal/ai/rag"
	"cs-agent/internal/pkg/constants"
	"cs-agent/internal/pkg/dto/request"
	"cs-agent/internal/services"

	"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple/web"
	"github.com/mlogclub/simple/web/params"
)

type KnowledgeRetrieveController struct {
	Ctx iris.Context
}

func (c *KnowledgeRetrieveController) PostDebugSearch() *web.JsonResult {
	if _, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionKnowledgeDocumentView); err != nil {
		return web.JsonError(err)
	}

	req := request.KnowledgeSearchRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}

	resp, err := rag.Answer.DebugSearch(context.Background(), req)
	if err != nil {
		return web.JsonError(err)
	}
	return web.JsonData(resp)
}

func (c *KnowledgeRetrieveController) PostDebugAnswer() *web.JsonResult {
	operator, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionKnowledgeDocumentView)
	if err != nil {
		return web.JsonError(err)
	}

	req := request.KnowledgeAnswerRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}

	resp, err := rag.Answer.DebugAnswer(context.Background(), req, operator)
	if err != nil {
		return web.JsonError(err)
	}
	return web.JsonData(resp)
}

func (c *KnowledgeRetrieveController) PostBuild() *web.JsonResult {
	if _, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionKnowledgeDocumentUpdate); err != nil {
		return web.JsonError(err)
	}

	req := struct {
		DocumentID int64 `json:"documentId"`
	}{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	if req.DocumentID <= 0 {
		return web.JsonErrorMsg("documentId不能为空")
	}
	if err := rag.Answer.BuildDocumentIndex(context.Background(), req.DocumentID); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}
