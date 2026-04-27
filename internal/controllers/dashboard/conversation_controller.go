package dashboard

import (
	"cs-agent/internal/builders"
	"cs-agent/internal/pkg/constants"
	"cs-agent/internal/pkg/dto/request"
	"cs-agent/internal/pkg/dto/response"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/services"
	"strconv"
	"strings"

	"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple/common/strs"
	"github.com/mlogclub/simple/web"
	"github.com/mlogclub/simple/web/params"
	"github.com/spf13/cast"
)

type ConversationController struct {
	Ctx iris.Context
}

func (c *ConversationController) AnyList() *web.JsonResult {
	if _, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionConversationView); err != nil {
		return web.JsonError(err)
	}

	cnd := params.NewPagedSqlCnd(c.Ctx,
		params.QueryFilter{ParamName: "status"},
		params.QueryFilter{ParamName: "serviceMode"},
		params.QueryFilter{ParamName: "currentAssigneeId"},
	).Desc("last_message_at").Desc("id")

	paging := params.GetPaging(c.Ctx)

	if keyword, _ := params.Get(c.Ctx, "keyword"); strs.IsNotBlank(keyword) {
		keywordLike := "%" + strings.TrimSpace(keyword) + "%"
		cnd.Where("last_message_summary LIKE ? OR customer_id IN (SELECT id FROM t_customer WHERE name LIKE ?)", keywordLike, keywordLike)
	}

	// 标签搜索
	if tagID, _ := params.GetInt64(c.Ctx, "tagId"); tagID > 0 {
		tagIDs := services.TagService.GetSelfAndDescendantIDs(tagID)
		if len(tagIDs) == 0 {
			return web.JsonData(&web.PageResult{
				Results: []response.ConversationResponse{},
				Page:    paging,
			})
		}
		cnd.Where("id IN (SELECT conversation_id FROM conversation_tag_rels WHERE tag_id IN (?))", tagIDs)
	}
	if agentTeamID, _ := params.GetInt64(c.Ctx, "agentTeamId"); agentTeamID > 0 {
		userIDs := services.AgentProfileService.GetUserIDsByTeamID(agentTeamID)
		if len(userIDs) == 0 {
			return web.JsonData(&web.PageResult{
				Results: []response.ConversationResponse{},
				Page:    paging,
			})
		}
		cnd.In("current_assignee_id", userIDs)
	}

	list, paging := services.ConversationService.FindPageByCnd(cnd)
	results := make([]response.ConversationResponse, 0, len(list))
	for _, item := range list {
		results = append(results, builders.BuildConversation(&item))
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
	paging := params.GetPaging(c.Ctx)

	list, paging, err := services.ConversationService.ListConversations(
		operator.UserID,
		request.AgentConversationFilter(strings.TrimSpace(filterValue)),
		keyword,
		paging,
	)
	if err != nil {
		return web.JsonError(err)
	}

	results := make([]response.ConversationResponse, 0, len(list))
	for _, item := range list {
		results = append(results, builders.BuildConversation(&item))
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
		ConversationResponse: builders.BuildConversation(item),
		Participants:         builders.BuildParticipantResponses(id),
	}
	return web.JsonData(detail)
}

func (c *ConversationController) AnyMessage_list() *web.JsonResult {
	if _, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionConversationView); err != nil {
		return web.JsonError(err)
	}

	var (
		conversationID, _ = params.GetInt64(c.Ctx, "conversationId")
		senderType, _     = params.Get(c.Ctx, "senderType")
		messageType, _    = params.Get(c.Ctx, "messageType")
		cursor, _         = params.GetInt64(c.Ctx, "cursor")
		limit, _          = params.GetInt(c.Ctx, "limit")
	)
	if conversation := services.ConversationService.Get(conversationID); conversation == nil {
		return web.JsonErrorMsg("会话不存在")
	}

	list, nextCursor, hasMore := services.MessageService.FindByConversationIDCursor(
		conversationID, cursor, limit, senderType, messageType,
	)
	results := builders.BuildMessages(list)

	return web.JsonCursorData(results, cast.ToString(nextCursor), hasMore)
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
	if err := services.ConversationService.AssignConversation(req, operator); err != nil {
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
	if err := services.ConversationService.AutoAssignConversation(req.ConversationID, operator); err != nil {
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

func (c *ConversationController) PostLink_customer() *web.JsonResult {
	operator, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionConversationLinkCustomer)
	if err != nil {
		return web.JsonError(err)
	}
	req := request.LinkConversationCustomerRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	if err := services.ConversationService.LinkConversationCustomer(req.ConversationID, req.CustomerID, operator); err != nil {
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
	return web.JsonData(builders.BuildMessage(item))
}

func (c *ConversationController) PostRecall_message() *web.JsonResult {
	operator, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionConversationSend)
	if err != nil {
		return web.JsonError(err)
	}

	req := request.RecallConversationMessageRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	item, err := services.MessageService.RecallAgentMessage(req.MessageID, operator)
	if err != nil {
		return web.JsonError(err)
	}
	return web.JsonData(builders.BuildMessage(item))
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

	rawConv := strings.TrimSpace(c.Ctx.FormValue("conversationId"))
	if rawConv == "" {
		return web.JsonErrorMsg("conversationId不能为空")
	}
	conversationID, err := strconv.ParseInt(rawConv, 10, 64)
	if err != nil || conversationID <= 0 {
		return web.JsonErrorMsg("conversationId不能为空")
	}
	if _, err := services.MessageService.ValidateConversationSender(conversationID, enums.IMSenderTypeAgent, operator, nil); err != nil {
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

	item, err := services.AssetService.UploadFile(header, "images", operator)
	if err != nil {
		return web.JsonError(err)
	}
	return web.JsonData(builders.BuildAsset(item))
}

func (c *ConversationController) PostUpload_attachment() *web.JsonResult {
	operator, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionConversationSend)
	if err != nil {
		return web.JsonError(err)
	}

	rawConv := strings.TrimSpace(c.Ctx.FormValue("conversationId"))
	if rawConv == "" {
		return web.JsonErrorMsg("conversationId不能为空")
	}
	conversationID, err := strconv.ParseInt(rawConv, 10, 64)
	if err != nil || conversationID <= 0 {
		return web.JsonErrorMsg("conversationId不能为空")
	}
	if _, err := services.MessageService.ValidateConversationSender(conversationID, enums.IMSenderTypeAgent, operator, nil); err != nil {
		return web.JsonError(err)
	}

	f, header, err := c.Ctx.FormFile("file")
	if err != nil {
		return web.JsonErrorMsg("请选择上传附件")
	}
	_ = f.Close()

	item, err := services.AssetService.UploadFile(header, "attachments", operator)
	if err != nil {
		return web.JsonError(err)
	}
	return web.JsonData(builders.BuildAsset(item))
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
