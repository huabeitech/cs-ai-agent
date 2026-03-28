package open

import (
	"cs-agent/internal/builders"
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/config"
	"cs-agent/internal/pkg/dto/request"
	"cs-agent/internal/services"
	"cs-agent/internal/services/storage"
	"strings"

	"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple/web"
	"github.com/mlogclub/simple/web/params"
)

type ImMessageController struct {
	Ctx iris.Context
	Cfg *config.Config
}

func (c *ImMessageController) AnyList() *web.JsonResult {
	if _, rsp := requireEnabledWidgetSite(c.Ctx); rsp != nil {
		return rsp
	}
	externalInfo, err := request.GetExternalInfo(c.Ctx)
	if err != nil {
		return web.JsonError(err)
	}

	conversationID, _ := params.GetInt64(c.Ctx, "conversationId")
	if conversationID <= 0 {
		return web.JsonErrorMsg("conversationId不能为空")
	}
	conversation := services.ConversationService.Get(conversationID)
	if conversation == nil {
		return web.JsonErrorMsg("会话不存在")
	}
	if !services.ConversationService.IsCustomerConversationOwner(conversation, *externalInfo) {
		return web.JsonErrorMsg("无权访问该会话")
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

func (c *ImMessageController) PostSend() *web.JsonResult {
	if _, rsp := requireEnabledWidgetSite(c.Ctx); rsp != nil {
		return rsp
	}
	externalInfo, err := request.GetExternalInfo(c.Ctx)
	if err != nil {
		return web.JsonError(err)
	}

	req := request.SendConversationMessageRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}

	item, err := services.MessageService.SendCustomerMessage(req.ConversationID, req.ClientMsgID, req.MessageType, req.Content, req.Payload, *externalInfo)
	if err != nil {
		return web.JsonError(err)
	}
	services.AIReplyService.TriggerReplyAsync(item.ID)
	return web.JsonData(builders.BuildMessageResponse(item))
}

func (c *ImMessageController) PostRead() *web.JsonResult {
	if _, rsp := requireEnabledWidgetSite(c.Ctx); rsp != nil {
		return rsp
	}
	externalInfo, err := request.GetExternalInfo(c.Ctx)
	if err != nil {
		return web.JsonError(err)
	}

	req := request.ReadConversationRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	if err := services.ConversationService.MarkCustomerConversationReadToMessage(req.ConversationID, req.MessageID, externalInfo); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}

func (c *ImMessageController) PostUpload_image() *web.JsonResult {
	if _, rsp := requireEnabledWidgetSite(c.Ctx); rsp != nil {
		return rsp
	}
	externalInfo, err := request.GetExternalInfo(c.Ctx)
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

	item, err := services.AssetService.UploadFileForExternal(c.Cfg, header, "im-images", *externalInfo)
	if err != nil {
		return web.JsonError(err)
	}
	provider, err := storage.NewProvider(c.Cfg.Storage)
	if err != nil {
		return web.JsonError(err)
	}
	return web.JsonData(builders.BuildAssetResponse(item, provider))
}
