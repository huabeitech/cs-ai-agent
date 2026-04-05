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
	if ChannelFromCtx(c.Ctx) == nil {
		return web.JsonErrorMsg("接入渠道未初始化")
	}
	external := ExternalInfoFromCtx(c.Ctx)
	if external == nil {
		return web.JsonErrorMsg("外部身份未初始化")
	}

	item := services.ConversationService.Get(id)
	if item == nil {
		return web.JsonErrorMsg("会话不存在")
	}
	if !services.ConversationService.IsCustomerConversationOwner(item, *external) {
		return web.JsonErrorMsg("无权访问该会话")
	}

	detail := response.ConversationDetailResponse{
		ConversationResponse: builders.BuildConversation(item),
		Participants:         builders.BuildParticipantResponses(id),
	}
	return web.JsonData(detail)
}

func (c *ImConversationController) PostCreate_or_match() *web.JsonResult {
	channel := ChannelFromCtx(c.Ctx)
	if channel == nil {
		return web.JsonErrorMsg("接入渠道未初始化")
	}
	external := ExternalInfoFromCtx(c.Ctx)
	if external == nil {
		return web.JsonErrorMsg("外部身份未初始化")
	}

	item, err := services.ConversationService.Create(*external, channel.AIAgentID)
	if err != nil {
		return web.JsonError(err)
	}
	return web.JsonData(builders.BuildConversation(item))
}

func (c *ImConversationController) PostClose() *web.JsonResult {
	if ChannelFromCtx(c.Ctx) == nil {
		return web.JsonErrorMsg("接入渠道未初始化")
	}
	external := ExternalInfoFromCtx(c.Ctx)
	if external == nil {
		return web.JsonErrorMsg("外部身份未初始化")
	}

	req := request.CloseConversationRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	if err := services.ConversationService.CloseCustomerConversation(req.ConversationID, *external); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}
