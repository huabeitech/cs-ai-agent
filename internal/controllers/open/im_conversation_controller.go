package open

import (
	"cs-agent/internal/builders"
	"cs-agent/internal/pkg/dto/request"
	"cs-agent/internal/pkg/dto/response"
	"cs-agent/internal/services"

	"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple/web"
	"github.com/mlogclub/simple/web/params"
)

type ImConversationController struct {
	Ctx iris.Context
}

func (c *ImConversationController) AnyList() *web.JsonResult {
	if _, rsp := requireEnabledWidgetSite(c.Ctx); rsp != nil {
		return rsp
	}
	principal, err := services.AuthService.GetImPrincipal(c.Ctx)
	if err != nil {
		return web.JsonError(err)
	}

	cnd := params.NewPagedSqlCnd(c.Ctx,
		params.QueryFilter{ParamName: "status"},
		params.QueryFilter{ParamName: "externalSource"},
		params.QueryFilter{ParamName: "serviceMode"},
	).Desc("last_message_at").Desc("id")
	if principal.IsVisitor {
		cnd = cnd.Eq("external_id", principal.VisitorID)
	} else {
		cnd = cnd.Eq("source_user_id", principal.UserID)
	}

	list, paging := services.ConversationService.FindPageByCnd(cnd)
	results := make([]response.ConversationResponse, 0, len(list))
	for _, item := range list {
		results = append(results, builders.BuildConversationResponse(&item))
	}
	return web.JsonData(&web.PageResult{Results: results, Page: paging})
}

func (c *ImConversationController) GetBy(id int64) *web.JsonResult {
	if _, rsp := requireEnabledWidgetSite(c.Ctx); rsp != nil {
		return rsp
	}
	principal, err := services.AuthService.GetImPrincipal(c.Ctx)
	if err != nil {
		return web.JsonError(err)
	}

	item := services.ConversationService.Get(id)
	if item == nil {
		return web.JsonErrorMsg("会话不存在")
	}
	if !services.ConversationService.IsCustomerConversationOwner(item, principal) {
		return web.JsonErrorMsg("无权访问该会话")
	}

	detail := response.ConversationDetailResponse{
		ConversationResponse: builders.BuildConversationResponse(item),
		Participants:         builders.BuildParticipantResponses(id),
	}
	return web.JsonData(detail)
}

func (c *ImConversationController) PostCreate_or_match() *web.JsonResult {
	site, rsp := requireEnabledWidgetSite(c.Ctx)
	if rsp != nil {
		return rsp
	}
	principal, err := services.AuthService.GetImPrincipal(c.Ctx)
	if err != nil {
		return web.JsonError(err)
	}

	req := request.CreateOrMatchConversationRequest{}
	if c.Ctx.GetContentLength() > 0 {
		if err := params.ReadJSON(c.Ctx, &req); err != nil {
			return web.JsonError(err)
		}
	}

	item, err := services.ConversationService.Create(req.ExternalSource, req.Subject, site.AIAgentID, principal)
	if err != nil {
		return web.JsonError(err)
	}
	return web.JsonData(builders.BuildConversationResponse(item))
}

func (c *ImConversationController) PostClose() *web.JsonResult {
	if _, rsp := requireEnabledWidgetSite(c.Ctx); rsp != nil {
		return rsp
	}
	principal, err := services.AuthService.GetImPrincipal(c.Ctx)
	if err != nil {
		return web.JsonError(err)
	}

	req := request.CloseConversationRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	if err := services.ConversationService.CloseCustomerConversation(req.ConversationID, principal); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}
