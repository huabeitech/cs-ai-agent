package dashboard

import (
	"cs-agent/internal/builders"
	"cs-agent/internal/pkg/constants"
	"cs-agent/internal/pkg/dto/response"
	"cs-agent/internal/services"

	"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple/web"
	"github.com/mlogclub/simple/web/params"
)

type KnowledgeRetrieveLogController struct {
	Ctx iris.Context
}

func (c *KnowledgeRetrieveLogController) AnyList() *web.JsonResult {
	if _, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionKnowledgeDocumentView); err != nil {
		return web.JsonError(err)
	}

	cnd := params.NewPagedSqlCnd(c.Ctx,
		params.QueryFilter{ParamName: "knowledgeBaseId"},
		params.QueryFilter{ParamName: "question", Op: params.Like},
		params.QueryFilter{ParamName: "channel"},
		params.QueryFilter{ParamName: "scene"},
		params.QueryFilter{ParamName: "chunkProvider"},
	).Desc("id")

	if answerStatus, ok := params.GetInt64(c.Ctx, "answerStatus"); ok && answerStatus > 0 {
		cnd.Where("answer_status = ?", answerStatus)
	}
	if rerankEnabled, ok := params.GetInt64(c.Ctx, "rerankEnabled"); ok {
		cnd.Where("rerank_enabled = ?", rerankEnabled > 0)
	}

	queryParams := params.NewQueryParams(c.Ctx)
	queryParams.Cnd = *cnd
	list, paging := services.KnowledgeRetrieveLogService.FindPageByParams(queryParams)
	results := make([]response.KnowledgeRetrieveLogResponse, 0, len(list))
	for _, item := range list {
		resp := builders.BuildKnowledgeRetrieveLog(&item)
		if knowledgeBase := services.KnowledgeBaseService.Get(item.KnowledgeBaseID); knowledgeBase != nil {
			resp.KnowledgeBaseName = knowledgeBase.Name
		}
		results = append(results, resp)
	}
	return web.JsonData(&web.PageResult{Results: results, Page: paging})
}

func (c *KnowledgeRetrieveLogController) GetBy(id int64) *web.JsonResult {
	if _, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionKnowledgeDocumentView); err != nil {
		return web.JsonError(err)
	}

	logItem := services.KnowledgeRetrieveLogService.Get(id)
	if logItem == nil {
		return web.JsonErrorMsg("检索日志不存在")
	}

	logResp := builders.BuildKnowledgeRetrieveLog(logItem)
	if knowledgeBase := services.KnowledgeBaseService.Get(logItem.KnowledgeBaseID); knowledgeBase != nil {
		logResp.KnowledgeBaseName = knowledgeBase.Name
	}

	hits := services.KnowledgeRetrieveLogService.FindHitsByRetrieveLogID(id)
	hitResults := make([]response.KnowledgeRetrieveHitResponse, 0, len(hits))
	for _, item := range hits {
		hitResults = append(hitResults, builders.BuildKnowledgeRetrieveHitResponse(&item))
	}

	return web.JsonData(response.KnowledgeRetrieveLogDetailResponse{
		Log:  logResp,
		Hits: hitResults,
	})
}
