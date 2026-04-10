package console

import (
	"encoding/json"
	"strings"

	"cs-agent/internal/models"
	"cs-agent/internal/pkg/dto/request"
	"cs-agent/internal/pkg/dto/response"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/pkg/toolx"
	"cs-agent/internal/pkg/utils"
	"cs-agent/internal/services"

	"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web"
	"github.com/mlogclub/simple/web/params"
)

type AIAgentController struct {
	Ctx iris.Context
}

func (c *AIAgentController) AnyList() *web.JsonResult {
	cnd := params.NewPagedSqlCnd(c.Ctx,
		params.QueryFilter{ParamName: "status"},
		params.QueryFilter{ParamName: "name", Op: params.Like},
		params.QueryFilter{ParamName: "code", Op: params.Like},
	).Desc("sort_no").Desc("id")
	list, paging := services.AIAgentService.FindPageByCnd(cnd)
	results := make([]response.AIAgentResponse, 0, len(list))
	for _, item := range list {
		results = append(results, buildAIAgentResponse(&item))
	}
	return web.JsonData(&web.PageResult{Results: results, Page: paging})
}

func (c *AIAgentController) GetList_all() *web.JsonResult {
	list := services.AIAgentService.Find(sqls.NewCnd().Where("status = ?", enums.StatusOk).Desc("sort_no").Desc("id"))
	results := make([]response.AIAgentResponse, 0, len(list))
	for _, item := range list {
		results = append(results, buildAIAgentResponse(&item))
	}
	return web.JsonData(results)
}

func (c *AIAgentController) GetBy(id int64) *web.JsonResult {
	item := services.AIAgentService.Get(id)
	if item == nil {
		return web.JsonErrorMsg("AI Agent 不存在")
	}
	return web.JsonData(buildAIAgentResponse(item))
}

func (c *AIAgentController) PostCreate() *web.JsonResult {
	req := request.CreateAIAgentRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	item, err := services.AIAgentService.CreateAIAgent(req, services.AuthService.GetAuthPrincipal(c.Ctx))
	if err != nil {
		return web.JsonError(err)
	}
	return web.JsonData(buildAIAgentResponse(item))
}

func (c *AIAgentController) PostUpdate() *web.JsonResult {
	req := request.UpdateAIAgentRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	if err := services.AIAgentService.UpdateAIAgent(req, services.AuthService.GetAuthPrincipal(c.Ctx)); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}

func (c *AIAgentController) PostDelete() *web.JsonResult {
	req := request.DeleteAIAgentRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	if err := services.AIAgentService.DeleteAIAgent(req.ID, services.AuthService.GetAuthPrincipal(c.Ctx)); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}

func (c *AIAgentController) PostUpdate_sort() *web.JsonResult {
	var ids []int64
	if err := c.Ctx.ReadJSON(&ids); err != nil {
		return web.JsonError(err)
	}
	if err := services.AIAgentService.UpdateSort(ids); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}

func (c *AIAgentController) PostUpdate_status() *web.JsonResult {
	req := request.UpdateAIAgentStatusRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	if err := services.AIAgentService.UpdateStatus(req.ID, req.Status, services.AuthService.GetAuthPrincipal(c.Ctx)); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}

func buildAIAgentResponse(item *models.AIAgent) response.AIAgentResponse {
	ret := response.AIAgentResponse{
		ID:                  item.ID,
		Name:                item.Name,
		Description:         item.Description,
		Status:              item.Status,
		StatusName:          enums.GetStatusLabel(item.Status),
		AIConfigID:          item.AIConfigID,
		ServiceMode:         item.ServiceMode,
		ServiceModeName:     enums.GetIMConversationServiceModeLabel(item.ServiceMode),
		SystemPrompt:        item.SystemPrompt,
		WelcomeMessage:      item.WelcomeMessage,
		ReplyTimeoutSeconds: item.ReplyTimeoutSeconds,
		HandoffMode:         item.HandoffMode,
		HandoffModeName:     enums.GetAIAgentHandoffModeLabel(item.HandoffMode),
		MaxAIReplyRounds:    item.MaxAIReplyRounds,
		FallbackMode:        item.FallbackMode,
		FallbackModeName:    enums.GetAIAgentFallbackModeLabel(item.FallbackMode),
		FallbackMessage:     item.FallbackMessage,
		KnowledgeIDs:        utils.SplitInt64s(item.KnowledgeIDs),
		SkillIDs:            utils.SplitInt64s(item.SkillIDs),
		KnowledgeBaseNames:  make([]string, 0),
		Skills:              make([]response.AIAgentSkillResponse, 0),
		Teams:               make([]response.AIAgentTeamResponse, 0),
		DirectTools:         make([]response.AIAgentMCPToolResponse, 0),
		SortNo:              item.SortNo,
		Remark:              item.Remark,
		CreatedAt:           item.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:           item.UpdatedAt.Format("2006-01-02 15:04:05"),
		CreateUserName:      item.CreateUserName,
		UpdateUserName:      item.UpdateUserName,
	}
	if aiConfig := services.AIConfigService.Get(item.AIConfigID); aiConfig != nil {
		ret.AIConfigName = aiConfig.Name
	}
	for _, id := range utils.SplitInt64s(item.TeamIDs) {
		if team := services.AgentTeamService.Get(id); team != nil {
			ret.Teams = append(ret.Teams, response.AIAgentTeamResponse{
				ID:   team.ID,
				Name: team.Name,
			})
		}
	}
	for _, id := range ret.KnowledgeIDs {
		if knowledgeBase := services.KnowledgeBaseService.Get(id); knowledgeBase != nil {
			ret.KnowledgeBaseNames = append(ret.KnowledgeBaseNames, knowledgeBase.Name)
		}
	}
	for _, id := range ret.SkillIDs {
		if skill := services.SkillDefinitionService.Get(id); skill != nil {
			ret.Skills = append(ret.Skills, response.AIAgentSkillResponse{
				ID:   skill.ID,
				Code: skill.Code,
				Name: skill.Name,
			})
		}
	}
	if raw := strings.TrimSpace(item.AllowedMCPTools); raw != "" {
		var directTools []request.AIAgentMCPToolRequest
		if err := json.Unmarshal([]byte(raw), &directTools); err == nil {
			for _, tool := range directTools {
				toolCode := strings.TrimSpace(tool.ToolCode)
				if toolCode == "" {
					toolCode = toolx.BuildMCPToolCode(tool.ServerCode, tool.ToolName)
				}
				toolCode = toolx.NormalizeToolCodeAlias(toolCode)
				if toolx.IsAutoInjectedToolCode(toolCode) {
					continue
				}
				serverCode := strings.TrimSpace(tool.ServerCode)
				toolName := strings.TrimSpace(tool.ToolName)
				if toolCode == toolx.BuiltinToolSearchToolCode {
					serverCode = toolx.BuiltinToolCatalogServerCode
					toolName = toolx.BuiltinToolSearchToolName
				} else if toolCode == toolx.GraphCreateTicketConfirmToolCode {
					serverCode = toolx.GraphToolCatalogServerCode
					toolName = toolx.GraphCreateTicketConfirmToolName
				} else if toolCode == toolx.GraphHandoffConversationToolCode {
					serverCode = toolx.GraphToolCatalogServerCode
					toolName = toolx.GraphHandoffConversationToolName
				} else if parsedServerCode, parsedToolName := toolx.SplitMCPToolCode(toolCode); parsedServerCode != "" && parsedToolName != "" {
					serverCode = parsedServerCode
					toolName = parsedToolName
				}
				title := strings.TrimSpace(tool.Title)
				if title == "" {
					switch toolCode {
					case toolx.BuiltinToolSearchToolCode:
						title = toolx.BuiltinToolSearchToolTitle
					case toolx.GraphCreateTicketConfirmToolCode:
						title = toolx.GraphCreateTicketConfirmToolTitle
					case toolx.GraphHandoffConversationToolCode:
						title = toolx.GraphHandoffConversationToolTitle
					}
				}
				description := strings.TrimSpace(tool.Description)
				if description == "" {
					switch toolCode {
					case toolx.BuiltinToolSearchToolCode:
						description = toolx.BuiltinToolSearchToolDescription
					case toolx.GraphCreateTicketConfirmToolCode:
						description = toolx.GraphCreateTicketConfirmToolDescription
					case toolx.GraphHandoffConversationToolCode:
						description = toolx.GraphHandoffConversationToolDescription
					}
				}
				ret.DirectTools = append(ret.DirectTools, response.AIAgentMCPToolResponse{
					ToolCode:    toolCode,
					ServerCode:  serverCode,
					ToolName:    toolName,
					Title:       title,
					Description: description,
					Arguments:   tool.Arguments,
				})
			}
		}
	}
	return ret
}
