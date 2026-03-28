package console

import (
	"cs-agent/internal/builders"
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/config"
	"cs-agent/internal/pkg/constants"
	"cs-agent/internal/pkg/dto/request"
	"cs-agent/internal/pkg/dto/response"
	"cs-agent/internal/services"
	"cs-agent/internal/services/storage"
	"strings"

	"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple/common/strs"
	"github.com/mlogclub/simple/web"
	"github.com/mlogclub/simple/web/params"
)

type ConversationController struct {
	Ctx iris.Context
	Cfg *config.Config
}

func (c *ConversationController) AnyList() *web.JsonResult {
	if _, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionConversationView); err != nil {
		return web.JsonError(err)
	}

	cnd := params.NewPagedSqlCnd(c.Ctx,
		params.QueryFilter{ParamName: "status"},
		params.QueryFilter{ParamName: "externalSource"},
		params.QueryFilter{ParamName: "serviceMode"},
		params.QueryFilter{ParamName: "currentAssigneeId"},
		params.QueryFilter{ParamName: "sourceUserId"},
	).Desc("last_message_at").Desc("id")

	if keyword, _ := params.Get(c.Ctx, "keyword"); strs.IsNotBlank(keyword) {
		cnd.Where("subject LIKE ? OR external_id LIKE ? OR last_message_summary LIKE ?", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}
	if tagID, _ := params.GetInt64(c.Ctx, "tagId"); tagID > 0 {
		cnd.Where("id IN (SELECT conversation_id FROM conversation_tag_rels WHERE tag_id = ?)", tagID)
	}

	list, paging := services.ConversationService.FindPageByCnd(cnd)
	results := make([]response.ConversationResponse, 0, len(list))
	for _, item := range list {
		results = append(results, builders.BuildConversationResponse(&item))
	}
	return web.JsonData(&web.PageResult{Results: results, Page: paging})
}

func (c *ConversationController) AnyConversations() *web.JsonResult {
	operator, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionConversationView)
	if err != nil {
		return web.JsonError(err)
	}

	filterValue, _ := params.Get(c.Ctx, "filter")
	keyword, _ := params.Get(c.Ctx, "keyword")
	page, _ := params.GetInt(c.Ctx, "page")
	limit, _ := params.GetInt(c.Ctx, "limit")

	list, paging, err := services.ConversationService.ListConversations(
		operator.UserID,
		request.AgentConversationFilter(strings.TrimSpace(filterValue)),
		keyword,
		page,
		limit,
	)
	if err != nil {
		return web.JsonError(err)
	}

	results := make([]response.ConversationResponse, 0, len(list))
	for _, item := range list {
		results = append(results, builders.BuildConversationResponse(&item))
	}
	return web.JsonData(&web.PageResult{Results: results, Page: paging})
}

func (c *ConversationController) GetBy(id int64) *web.JsonResult {
	if _, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionConversationView); err != nil {
		return web.JsonError(err)
	}

	item := services.ConversationService.Get(id)
	if item == nil {
		return web.JsonErrorMsg("会话不存在")
	}

	detail := response.ConversationDetailResponse{
		ConversationResponse: builders.BuildConversationResponse(item),
		Participants:         builders.BuildParticipantResponses(id),
	}
	return web.JsonData(detail)
}

func (c *ConversationController) AnyMessage_list() *web.JsonResult {
	if _, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionConversationView); err != nil {
		return web.JsonError(err)
	}

	conversationID, _ := params.GetInt64(c.Ctx, "conversationId")
	if conversationID <= 0 {
		return web.JsonErrorMsg("conversationId不能为空")
	}
	if services.ConversationService.Get(conversationID) == nil {
		return web.JsonErrorMsg("会话不存在")
	}

	cnd := params.NewPagedSqlCnd(c.Ctx,
		params.QueryFilter{ParamName: "conversationId"},
		params.QueryFilter{ParamName: "senderType"},
		params.QueryFilter{ParamName: "messageType"},
	).Desc("seq_no")
	list, paging := services.MessageService.FindPageByCnd(cnd)
	results := make([]models.Message, 0, len(list))
	for i := len(list) - 1; i >= 0; i-- {
		results = append(results, list[i])
	}
	return web.JsonData(&web.PageResult{Results: builders.BuildMessageResponses(results), Page: paging})
}

func (c *ConversationController) PostAssign() *web.JsonResult {
	operator, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionConversationAssign)
	if err != nil {
		return web.JsonError(err)
	}

	req := request.AssignConversationRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	if err := services.ConversationService.AssignConversation(req.ConversationID, req.AssigneeID, req.Reason, operator); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}

func (c *ConversationController) PostDispatch() *web.JsonResult {
	operator, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionConversationAssign)
	if err != nil {
		return web.JsonError(err)
	}

	req := request.DispatchConversationRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	if err := services.ConversationService.DispatchConversation(req.ConversationID, operator); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}

func (c *ConversationController) PostTransfer() *web.JsonResult {
	operator, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionConversationTransfer)
	if err != nil {
		return web.JsonError(err)
	}

	req := request.TransferConversationRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	if err := services.ConversationService.TransferConversation(req.ConversationID, req.ToUserID, req.Reason, operator); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}

func (c *ConversationController) PostClose() *web.JsonResult {
	operator, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionConversationClose)
	if err != nil {
		return web.JsonError(err)
	}

	req := request.CloseConversationRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	if err := services.ConversationService.CloseConversation(req.ConversationID, req.CloseReason, operator); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}

func (c *ConversationController) PostSend_message() *web.JsonResult {
	operator, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionConversationSend)
	if err != nil {
		return web.JsonError(err)
	}

	req := request.SendConversationMessageRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	item, err := services.MessageService.SendAgentMessage(req.ConversationID, 0, req.ClientMsgID, req.MessageType, req.Content, req.Payload, operator)
	if err != nil {
		return web.JsonError(err)
	}
	return web.JsonData(builders.BuildMessageResponse(item))
}

func (c *ConversationController) PostRead() *web.JsonResult {
	operator, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionConversationView)
	if err != nil {
		return web.JsonError(err)
	}

	req := request.ReadConversationRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	if err := services.ConversationService.MarkAgentConversationReadToMessage(req.ConversationID, req.MessageID, operator); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}

func (c *ConversationController) PostUpload_image() *web.JsonResult {
	operator, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionConversationSend)
	if err != nil {
		return web.JsonError(err)
	}

	f, header, err := c.Ctx.FormFile("file")
	if err != nil {
		return web.JsonErrorMsg("请选择上传图片")
	}
	_ = f.Close()

	if !strings.HasPrefix(strings.ToLower(header.Header.Get("Content-Type")), "image/") {
		return web.JsonErrorMsg("仅支持上传图片文件")
	}

	item, err := services.AssetService.UploadFile(c.Cfg, header, "im-images", operator)
	if err != nil {
		return web.JsonError(err)
	}
	provider, err := storage.NewProvider(c.Cfg.Storage)
	if err != nil {
		return web.JsonError(err)
	}
	return web.JsonData(builders.BuildAssetResponse(item, provider))
}

func (c *ConversationController) PostAdd_tag() *web.JsonResult {
	operator, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionConversationTag)
	if err != nil {
		return web.JsonError(err)
	}

	req := request.AddConversationTagRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	if err := services.ConversationTagService.AddTag(req, operator); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}

func (c *ConversationController) PostRemove_tag() *web.JsonResult {
	if _, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionConversationTag); err != nil {
		return web.JsonError(err)
	}

	req := request.RemoveConversationTagRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	if err := services.ConversationTagService.RemoveTag(req); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}
