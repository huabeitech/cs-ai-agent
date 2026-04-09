package open

import (
	"cs-agent/internal/builders"
	"cs-agent/internal/pkg/dto/request"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/pkg/irisx"
	"cs-agent/internal/services"
	"strconv"
	"strings"

	"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple/web"
	"github.com/mlogclub/simple/web/params"
	"github.com/spf13/cast"
)

type ImMessageController struct {
	Ctx iris.Context
}

func (c *ImMessageController) AnyList() *web.JsonResult {
	if irisx.GetChannel(c.Ctx) == nil {
		return web.JsonErrorMsg("接入渠道未初始化")
	}
	external := irisx.GetExternalInfo(c.Ctx)
	if external == nil {
		return web.JsonErrorMsg("外部身份未初始化")
	}

	conversationID, _ := params.GetInt64(c.Ctx, "conversationId")
	if conversationID <= 0 {
		return web.JsonErrorMsg("conversationId不能为空")
	}
	conversation := services.ConversationService.Get(conversationID)
	if conversation == nil {
		return web.JsonErrorMsg("会话不存在")
	}
	if !services.ConversationService.IsCustomerConversationOwner(conversation, *external) {
		return web.JsonErrorMsg("无权访问该会话")
	}

	var (
		senderType, _  = params.Get(c.Ctx, "senderType")
		messageType, _ = params.Get(c.Ctx, "messageType")
		cursor, _      = params.GetInt64(c.Ctx, "cursor")
		limit, _       = params.GetInt(c.Ctx, "limit")
	)
	list, nextCursor, hasMore := services.MessageService.FindByConversationIDCursor(
		conversationID, cursor, limit, senderType, messageType,
	)
	results := builders.BuildMessages(list)
	return web.JsonCursorData(results, cast.ToString(nextCursor), hasMore)
}

func (c *ImMessageController) PostSend() *web.JsonResult {
	if irisx.GetChannel(c.Ctx) == nil {
		return web.JsonErrorMsg("接入渠道未初始化")
	}
	external := irisx.GetExternalInfo(c.Ctx)
	if external == nil {
		return web.JsonErrorMsg("外部身份未初始化")
	}

	req := request.SendConversationMessageRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}

	item, err := services.MessageService.SendCustomerMessage(req.ConversationID, req.ClientMsgID, req.MessageType, req.Content, req.Payload, *external)
	if err != nil {
		return web.JsonError(err)
	}
	return web.JsonData(builders.BuildMessage(item))
}

func (c *ImMessageController) PostRead() *web.JsonResult {
	if irisx.GetChannel(c.Ctx) == nil {
		return web.JsonErrorMsg("接入渠道未初始化")
	}
	external := irisx.GetExternalInfo(c.Ctx)
	if external == nil {
		return web.JsonErrorMsg("外部身份未初始化")
	}

	req := request.ReadConversationRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	if err := services.ConversationService.MarkCustomerConversationReadToMessage(req.ConversationID, req.MessageID, external); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}

func (c *ImMessageController) PostUpload_image() *web.JsonResult {
	if irisx.GetChannel(c.Ctx) == nil {
		return web.JsonErrorMsg("接入渠道未初始化")
	}
	external := irisx.GetExternalInfo(c.Ctx)
	if external == nil {
		return web.JsonErrorMsg("外部身份未初始化")
	}

	rawConv := strings.TrimSpace(c.Ctx.FormValue("conversationId"))
	if rawConv == "" {
		return web.JsonErrorMsg("conversationId不能为空")
	}
	conversationID, err := strconv.ParseInt(rawConv, 10, 64)
	if err != nil || conversationID <= 0 {
		return web.JsonErrorMsg("conversationId不能为空")
	}
	conversation := services.ConversationService.Get(conversationID)
	if conversation == nil {
		return web.JsonErrorMsg("会话不存在")
	}
	if !services.ConversationService.IsCustomerConversationOwner(conversation, *external) {
		return web.JsonErrorMsg("无权访问该会话")
	}
	if _, err := services.MessageService.ValidateConversationSender(conversationID, enums.IMSenderTypeCustomer, nil, external); err != nil {
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

	item, err := services.AssetService.UploadFile(header, "images", nil)
	if err != nil {
		return web.JsonError(err)
	}
	return web.JsonData(builders.BuildAsset(item))
}

func (c *ImMessageController) PostUpload_attachment() *web.JsonResult {
	if irisx.GetChannel(c.Ctx) == nil {
		return web.JsonErrorMsg("接入渠道未初始化")
	}
	external := irisx.GetExternalInfo(c.Ctx)
	if external == nil {
		return web.JsonErrorMsg("外部身份未初始化")
	}

	rawConv := strings.TrimSpace(c.Ctx.FormValue("conversationId"))
	if rawConv == "" {
		return web.JsonErrorMsg("conversationId不能为空")
	}
	conversationID, err := strconv.ParseInt(rawConv, 10, 64)
	if err != nil || conversationID <= 0 {
		return web.JsonErrorMsg("conversationId不能为空")
	}
	if _, err := services.MessageService.ValidateConversationSender(conversationID, enums.IMSenderTypeCustomer, nil, external); err != nil {
		return web.JsonError(err)
	}

	f, header, err := c.Ctx.FormFile("file")
	if err != nil {
		return web.JsonErrorMsg("请选择上传附件")
	}
	_ = f.Close()

	item, err := services.AssetService.UploadFile(header, "attachments", nil)
	if err != nil {
		return web.JsonError(err)
	}
	return web.JsonData(builders.BuildAsset(item))
}
