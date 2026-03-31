package console

import (
	"cs-agent/internal/builders"
	"cs-agent/internal/pkg/config"
	"cs-agent/internal/pkg/constants"
	"cs-agent/internal/pkg/dto/request"
	"cs-agent/internal/pkg/dto/response"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/services"
	"cs-agent/internal/services/storage"
	"strings"

	"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple/web"
	"github.com/mlogclub/simple/web/params"
)

type AssetController struct {
	Ctx iris.Context
	Cfg *config.Config
}

func (c *AssetController) AnyList() *web.JsonResult {
	if _, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionAssetView); err != nil {
		return web.JsonError(err)
	}
	provider, err := storage.NewProvider(c.Cfg.Storage)
	if err != nil {
		return web.JsonError(err)
	}

	cnd := params.NewPagedSqlCnd(c.Ctx,
		params.QueryFilter{ParamName: "provider"},
		params.QueryFilter{ParamName: "status"},
		params.QueryFilter{ParamName: "createUserId"},
		params.QueryFilter{ParamName: "filename", Op: params.Like},
	).Desc("id")
	if strings.TrimSpace(c.Ctx.URLParam("status")) == "" {
		cnd = cnd.Eq("status", enums.AssetStatusSuccess)
	}

	list, paging := services.AssetService.FindPageByCnd(cnd)
	results := make([]response.AssetResponse, 0, len(list))
	for _, item := range list {
		results = append(results, builders.BuildAsset(&item, provider))
	}
	return web.JsonData(&web.PageResult{Results: results, Page: paging})
}

func (c *AssetController) GetBy(id int64) *web.JsonResult {
	if _, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionAssetView); err != nil {
		return web.JsonError(err)
	}
	item := services.AssetService.Get(id)
	if item == nil {
		return web.JsonErrorMsg("文件不存在")
	}
	provider, err := storage.NewProvider(c.Cfg.Storage)
	if err != nil {
		return web.JsonError(err)
	}
	return web.JsonData(builders.BuildAsset(item, provider))
}

func (c *AssetController) PostCreate() *web.JsonResult {
	operator, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionAssetCreate)
	if err != nil {
		return web.JsonError(err)
	}

	req := request.CreateAssetRequest{}
	if err := params.ReadForm(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	f, header, err := c.Ctx.FormFile("file")
	if err != nil {
		return web.JsonErrorMsg("请选择上传文件")
	}
	_ = f.Close()

	item, err := services.AssetService.UploadFile(c.Cfg, header, req.Prefix, operator)
	if err != nil {
		return web.JsonError(err)
	}
	provider, err := storage.NewProvider(c.Cfg.Storage)
	if err != nil {
		return web.JsonError(err)
	}
	return web.JsonData(builders.BuildAsset(item, provider))
}

func (c *AssetController) PostDelete() *web.JsonResult {
	operator, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionAssetDelete)
	if err != nil {
		return web.JsonError(err)
	}

	req := request.DeleteAssetRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	if err := services.AssetService.DeleteAsset(req.ID, operator); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}
