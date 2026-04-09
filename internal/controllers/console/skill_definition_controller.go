package console

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"cs-agent/internal/builders"
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/constants"
	"cs-agent/internal/pkg/dto/request"
	"cs-agent/internal/pkg/dto/response"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/pkg/errorsx"
	"cs-agent/internal/pkg/utils"
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
	).Asc("priority").Desc("id")
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

	list := services.SkillDefinitionService.Find(params.NewSqlCnd(c.Ctx,
		params.QueryFilter{ParamName: "status"},
	).Asc("priority").Desc("id"))
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
	if err := validateSkillDefinitionRequest(req); err != nil {
		return web.JsonError(err)
	}
	if services.SkillDefinitionService.Take("code = ?", strings.TrimSpace(req.Code)) != nil {
		return web.JsonErrorMsg("Skill 编码已存在")
	}

	item := &models.SkillDefinition{
		Code:             strings.TrimSpace(req.Code),
		Name:             strings.TrimSpace(req.Name),
		Description:      strings.TrimSpace(req.Description),
		Content:          strings.TrimSpace(req.Content),
		Examples:         mustMarshalJSONStringArray(req.Examples),
		AllowedToolCodes: mustMarshalJSONStringArray(req.AllowedToolCodes),
		Priority:         normalizeSkillPriority(req.Priority),
		Status:           enums.StatusOk,
		Remark:           strings.TrimSpace(req.Remark),
		AuditFields:      utils.BuildAuditFields(operator),
	}
	if item.Priority <= 0 {
		item.Priority = services.SkillDefinitionService.NextPriority()
	}
	if err := services.SkillDefinitionService.Create(item); err != nil {
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
	if req.ID <= 0 {
		return web.JsonErrorMsg("Skill ID 不合法")
	}
	if err := validateSkillDefinitionRequest(req.CreateSkillDefinitionRequest); err != nil {
		return web.JsonError(err)
	}

	item := services.SkillDefinitionService.Get(req.ID)
	if item == nil {
		return web.JsonErrorMsg("Skill 不存在")
	}
	exists := services.SkillDefinitionService.Take("code = ? AND id <> ?", strings.TrimSpace(req.Code), req.ID)
	if exists != nil {
		return web.JsonErrorMsg("Skill 编码已存在")
	}

	if err := services.SkillDefinitionService.Updates(req.ID, map[string]any{
		"code":               strings.TrimSpace(req.Code),
		"name":               strings.TrimSpace(req.Name),
		"description":        strings.TrimSpace(req.Description),
		"content":            strings.TrimSpace(req.Content),
		"examples":           mustMarshalJSONStringArray(req.Examples),
		"allowed_tool_codes": mustMarshalJSONStringArray(req.AllowedToolCodes),
		"priority":           resolveSkillPriorityForUpdate(req.Priority, item.Priority),
		"remark":             strings.TrimSpace(req.Remark),
		"update_user_id":     operator.UserID,
		"update_user_name":   operator.Username,
		"updated_at":         time.Now(),
	}); err != nil {
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
	if services.SkillDefinitionService.Get(req.ID) == nil {
		return web.JsonErrorMsg("Skill 不存在")
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

func (c *SkillDefinitionController) PostUpdate_priority() *web.JsonResult {
	if _, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionSkillDefinitionUpdate); err != nil {
		return web.JsonError(err)
	}

	var ids []int64
	if err := c.Ctx.ReadJSON(&ids); err != nil {
		return web.JsonError(err)
	}
	if err := services.SkillDefinitionService.UpdatePriority(ids); err != nil {
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

func validateSkillDefinitionRequest(req request.CreateSkillDefinitionRequest) error {
	code := strings.TrimSpace(req.Code)
	name := strings.TrimSpace(req.Name)
	content := strings.TrimSpace(req.Content)
	if code == "" {
		return errorsx.InvalidParam("Skill 编码不能为空")
	}
	if name == "" {
		return errorsx.InvalidParam("Skill 名称不能为空")
	}
	if content == "" {
		return errorsx.InvalidParam("Content 不能为空")
	}
	if _, err := normalizeJSONStringArray(req.Examples); err != nil {
		return err
	}
	allowedToolCodes, err := normalizeJSONStringArray(req.AllowedToolCodes)
	if err != nil {
		return err
	}
	for _, toolCode := range allowedToolCodes {
		if err := services.ToolCatalogService.ValidateMCPToolCode(toolCode); err != nil {
			return err
		}
	}
	return nil
}

func normalizeSkillPriority(priority int) int {
	if priority < 0 {
		return 0
	}
	return priority
}

func resolveSkillPriorityForUpdate(input, current int) int {
	input = normalizeSkillPriority(input)
	if input <= 0 {
		return current
	}
	return input
}

func normalizeJSONStringArray(input []string) ([]string, error) {
	ret := make([]string, 0, len(input))
	seen := make(map[string]struct{})
	for _, item := range input {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if _, exists := seen[item]; exists {
			continue
		}
		seen[item] = struct{}{}
		ret = append(ret, item)
	}
	return ret, nil
}

func mustMarshalJSONStringArray(input []string) string {
	items, _ := normalizeJSONStringArray(input)
	if len(items) == 0 {
		return "[]"
	}
	buf, err := json.Marshal(items)
	if err != nil {
		return "[]"
	}
	return string(buf)
}
