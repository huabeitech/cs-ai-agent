package dashboard

import (
	"context"
	"strings"
	"time"

	"cs-agent/internal/builders"
	"cs-agent/internal/pkg/constants"
	"cs-agent/internal/pkg/dto/request"
	"cs-agent/internal/pkg/dto/response"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/services"

	"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple/web"
	"github.com/mlogclub/simple/web/params"
)

type SkillDefinitionController struct {
	Ctx iris.Context
}

func (c *SkillDefinitionController) AnyList() *web.JsonResult {
	if _, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionSkillDefinitionView); err != nil {
		return web.JsonError(err)
	}

	cnd := params.NewPagedSqlCnd(c.Ctx,
		params.QueryFilter{ParamName: "status"},
		params.QueryFilter{ParamName: "name", Op: params.Like},
		params.QueryFilter{ParamName: "code", Op: params.Like},
	).Desc("id")
	if _, ok := params.Get(c.Ctx, "status"); !ok {
		cnd.Where("status <> ?", enums.StatusDeleted)
	}
	list, paging := services.SkillDefinitionService.FindPageByCnd(cnd)
	results := make([]response.SkillDefinitionResponse, 0, len(list))
	for _, item := range list {
		results = append(results, builders.BuildSkillDefinitionResponse(&item))
	}
	return web.JsonData(&web.PageResult{Results: results, Page: paging})
}

func (c *SkillDefinitionController) GetList_all() *web.JsonResult {
	if _, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionSkillDefinitionView); err != nil {
		return web.JsonError(err)
	}

	cnd := params.NewSqlCnd(c.Ctx,
		params.QueryFilter{ParamName: "status"},
	).Desc("id")
	if status, ok := params.Get(c.Ctx, "status"); !ok || strings.TrimSpace(status) == "" {
		cnd.Where("status <> ?", enums.StatusDeleted)
	}
	list := services.SkillDefinitionService.Find(cnd)
	results := make([]response.SkillDefinitionResponse, 0, len(list))
	for _, item := range list {
		results = append(results, builders.BuildSkillDefinitionResponse(&item))
	}
	return web.JsonData(results)
}

func (c *SkillDefinitionController) GetBy(id int64) *web.JsonResult {
	if _, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionSkillDefinitionView); err != nil {
		return web.JsonError(err)
	}

	item := services.SkillDefinitionService.Get(id)
	if item == nil {
		return web.JsonErrorMsg("Skill 不存在")
	}
	return web.JsonData(builders.BuildSkillDefinitionResponse(item))
}

func (c *SkillDefinitionController) PostCreate() *web.JsonResult {
	operator, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionSkillDefinitionCreate)
	if err != nil {
		return web.JsonError(err)
	}

	req := request.CreateSkillDefinitionRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	item, err := services.SkillDefinitionService.CreateSkillDefinition(req, operator)
	if err != nil {
		return web.JsonError(err)
	}
	return web.JsonData(builders.BuildSkillDefinitionResponse(item))
}

func (c *SkillDefinitionController) PostUpdate() *web.JsonResult {
	operator, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionSkillDefinitionUpdate)
	if err != nil {
		return web.JsonError(err)
	}

	req := request.UpdateSkillDefinitionRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	if err := services.SkillDefinitionService.UpdateSkillDefinition(req, operator); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}

func (c *SkillDefinitionController) PostUpdate_status() *web.JsonResult {
	operator, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionSkillDefinitionUpdate)
	if err != nil {
		return web.JsonError(err)
	}

	req := request.UpdateSkillDefinitionStatusRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	if req.ID <= 0 {
		return web.JsonErrorMsg("Skill ID 不合法")
	}
	if !enums.IsValidStatus(req.Status) {
		return web.JsonErrorMsg("状态值不合法")
	}
	item := services.SkillDefinitionService.Get(req.ID)
	if item == nil {
		return web.JsonErrorMsg("Skill 不存在")
	}
	if item.Status == enums.StatusDeleted {
		return web.JsonErrorMsg("已删除的 Skill 不能直接修改状态，请先恢复")
	}
	if req.Status == int(enums.StatusDeleted) {
		return web.JsonErrorMsg("请使用删除接口处理删除状态")
	}

	if err := services.SkillDefinitionService.Updates(req.ID, map[string]any{
		"status":           req.Status,
		"update_user_id":   operator.UserID,
		"update_user_name": operator.Username,
		"updated_at":       time.Now(),
	}); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}

func (c *SkillDefinitionController) PostDelete() *web.JsonResult {
	operator, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionSkillDefinitionDelete)
	if err != nil {
		return web.JsonError(err)
	}

	req := request.DeleteSkillDefinitionRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	if req.ID <= 0 {
		return web.JsonErrorMsg("Skill ID 不合法")
	}
	if services.SkillDefinitionService.Get(req.ID) == nil {
		return web.JsonErrorMsg("Skill 不存在")
	}
	if err := services.SkillDefinitionService.Updates(req.ID, map[string]any{
		"status":           enums.StatusDeleted,
		"update_user_id":   operator.UserID,
		"update_user_name": operator.Username,
		"updated_at":       time.Now(),
	}); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}

func (c *SkillDefinitionController) PostRestore() *web.JsonResult {
	operator, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionSkillDefinitionDelete)
	if err != nil {
		return web.JsonError(err)
	}

	req := request.RestoreSkillDefinitionRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	if req.ID <= 0 {
		return web.JsonErrorMsg("Skill ID 不合法")
	}

	item := services.SkillDefinitionService.Get(req.ID)
	if item == nil {
		return web.JsonErrorMsg("Skill 不存在")
	}
	if item.Status != enums.StatusDeleted {
		return web.JsonErrorMsg("仅已删除的 Skill 支持恢复")
	}

	if err := services.SkillDefinitionService.Updates(req.ID, map[string]any{
		"status":           enums.StatusDisabled,
		"update_user_id":   operator.UserID,
		"update_user_name": operator.Username,
		"updated_at":       time.Now(),
	}); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}

func (c *SkillDefinitionController) PostDebug_run() *web.JsonResult {
	if _, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionSkillDefinitionView); err != nil {
		return web.JsonError(err)
	}

	req := request.SkillDebugRunRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	resp, err := services.SkillRuntimeService.DebugRun(context.Background(), req)
	if err != nil {
		return web.JsonError(err)
	}
	return web.JsonData(resp)
}

func (c *SkillDefinitionController) PostDebug_resume() *web.JsonResult {
	if _, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionSkillDefinitionView); err != nil {
		return web.JsonError(err)
	}

	req := request.SkillDebugResumeRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	resp, err := services.SkillRuntimeService.DebugResume(context.Background(), req)
	if err != nil {
		return web.JsonError(err)
	}
	return web.JsonData(resp)
}
