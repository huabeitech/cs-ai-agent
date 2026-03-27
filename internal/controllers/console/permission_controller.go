package console

import (
	"cs-agent/internal/pkg/config"
	"cs-agent/internal/pkg/constants"
	"cs-agent/internal/pkg/dto/response"
	"cs-agent/internal/services"

	"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple/common/strs"
	"github.com/mlogclub/simple/web"
	"github.com/mlogclub/simple/web/params"
)

type PermissionController struct {
	Ctx iris.Context
	Cfg *config.Config
}

func (c *PermissionController) AnyList() *web.JsonResult {
	if _, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionPermissionView); err != nil {
		return web.JsonError(err)
	}

	cnd := params.NewPagedSqlCnd(c.Ctx,
		params.QueryFilter{ParamName: "groupName"},
		params.QueryFilter{ParamName: "type"},
		params.QueryFilter{ParamName: "status"},
	).Desc("id")

	if keyword, _ := params.Get(c.Ctx, "keyword"); strs.IsNotBlank(keyword) {
		cnd.Where("(name LIKE ? OR code LIKE ?)", "%"+keyword+"%", "%"+keyword+"%")
	}

	list, paging := services.PermissionService.FindPageByCnd(cnd)
	results := make([]response.PermissionResponse, 0, len(list))
	for _, item := range list {
		results = append(results, response.PermissionResponse{
			ID:        item.ID,
			Name:      item.Name,
			Code:      item.Code,
			Type:      item.Type,
			GroupName: item.GroupName,
			Method:    item.Method,
			ApiPath:   item.APIPath,
			Status:    item.Status,
			SortNo:    item.SortNo,
		})
	}
	return web.JsonData(&web.PageResult{Results: results, Page: paging})
}

func (c *PermissionController) GetBy(id int64) *web.JsonResult {
	if _, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionPermissionView); err != nil {
		return web.JsonError(err)
	}

	item := services.PermissionService.Get(id)
	if item == nil {
		return web.JsonErrorMsg("权限不存在")
	}
	return web.JsonData(&response.PermissionResponse{
		ID:        item.ID,
		Name:      item.Name,
		Code:      item.Code,
		Type:      item.Type,
		GroupName: item.GroupName,
		Method:    item.Method,
		ApiPath:   item.APIPath,
		Status:    item.Status,
		SortNo:    item.SortNo,
	})
}
