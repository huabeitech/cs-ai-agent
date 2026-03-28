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

func (c *ImConversationController) GetBy(id int64) *web.JsonResult {
	if _, rsp := requireEnabledWidgetSite(c.Ctx); rsp != nil {
		return rsp
	}
	externalSourceID, err := request.GetExternalInfo(c.Ctx)
	if err != nil {
		return web.JsonError(err)
	}

	item := services.ConversationService.Get(id)
	if item == nil {
		return web.JsonErrorMsg("会话不存在")
	}
	if !services.ConversationService.IsCustomerConversationOwner(item, *externalSourceID) {
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
	externalSourceID, err := request.GetExternalInfo(c.Ctx)
	if err != nil {
		return web.JsonError(err)
	}

	// TODO 这里不需要post body
	req := request.CreateOrMatchConversationRequest{}
	if c.Ctx.GetContentLength() > 0 {
		if err := params.ReadJSON(c.Ctx, &req); err != nil {
			return web.JsonError(err)
		}
	}

	item, err := services.ConversationService.Create(*externalSourceID, req.Subject, site.AIAgentID)
	if err != nil {
		return web.JsonError(err)
	}
	return web.JsonData(builders.BuildConversationResponse(item))
}

func (c *ImConversationController) PostClose() *web.JsonResult {
	if _, rsp := requireEnabledWidgetSite(c.Ctx); rsp != nil {
		return rsp
	}
	externalSourceID, err := request.GetExternalInfo(c.Ctx)
	if err != nil {
		return web.JsonError(err)
	}

	req := request.CloseConversationRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	if err := services.ConversationService.CloseCustomerConversation(req.ConversationID, *externalSourceID); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}
