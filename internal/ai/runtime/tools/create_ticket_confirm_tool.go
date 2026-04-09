package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"cs-agent/internal/ai/runtime/registry"
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/dto"
	"cs-agent/internal/pkg/dto/request"
	"cs-agent/internal/services"

	componenttool "github.com/cloudwego/eino/components/tool"
	einotool "github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	einojsonschema "github.com/eino-contrib/jsonschema"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

const (
	CreateTicketConfirmToolCode = "builtin/create_ticket_with_confirmation"
	CreateTicketConfirmToolName = "create_ticket_with_confirmation"
)

type CreateTicketConfirmState struct {
	Request request.CreateTicketFromConversationRequest
}

type CreateTicketConfirmInterruptInfo struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

func init() {
	schema.RegisterName[CreateTicketConfirmState]("cs_agent_create_ticket_confirm_state")
	schema.RegisterName[CreateTicketConfirmInterruptInfo]("cs_agent_create_ticket_confirm_interrupt_info")
}

type CreateTicketConfirmTool struct {
	conversation *models.Conversation
	aiAgent      *models.AIAgent
}

func NewCreateTicketConfirmTool() *CreateTicketConfirmTool {
	return &CreateTicketConfirmTool{}
}

func (t *CreateTicketConfirmTool) Name() string {
	return CreateTicketConfirmToolName
}

func (t *CreateTicketConfirmTool) Code() string {
	return CreateTicketConfirmToolCode
}

func (t *CreateTicketConfirmTool) Enabled(ctx registry.Context) bool {
	return ctx.Conversation != nil && ctx.AIAgent != nil
}

func (t *CreateTicketConfirmTool) Build(ctx registry.Context) (einotool.BaseTool, error) {
	if !t.Enabled(ctx) {
		return nil, nil
	}
	return &CreateTicketConfirmTool{
		conversation: ctx.Conversation,
		aiAgent:      ctx.AIAgent,
	}, nil
}

func (t *CreateTicketConfirmTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: CreateTicketConfirmToolName,
		Desc: "当用户明确希望创建工单、投诉单、报障单，且你已经整理出工单标题和描述后，调用此工具。该工具不会立即创建工单，而是会先向用户发起确认；只有用户确认后才真正创建。不要在信息不足时调用。",
		ParamsOneOf: schema.NewParamsOneOfByJSONSchema(&einojsonschema.Schema{
			Version: einojsonschema.Version,
			Type:    "object",
			Required: []string{
				"title",
				"description",
			},
			Properties: orderedmap.New[string, *einojsonschema.Schema](orderedmap.WithInitialData(
				orderedmap.Pair[string, *einojsonschema.Schema]{
					Key: "title",
					Value: &einojsonschema.Schema{
						Type:        "string",
						Description: "工单标题，简洁概括问题。",
					},
				},
				orderedmap.Pair[string, *einojsonschema.Schema]{
					Key: "description",
					Value: &einojsonschema.Schema{
						Type:        "string",
						Description: "工单描述，清晰整理用户问题、现象和诉求。",
					},
				},
				orderedmap.Pair[string, *einojsonschema.Schema]{
					Key: "priority",
					Value: &einojsonschema.Schema{
						Type:        "integer",
						Description: "工单优先级，可选；未知时可不传。",
					},
				},
				orderedmap.Pair[string, *einojsonschema.Schema]{
					Key: "severity",
					Value: &einojsonschema.Schema{
						Type:        "integer",
						Description: "严重度，可选；1=轻微，2=严重，3=致命。",
					},
				},
			)),
		}),
		Extra: map[string]any{
			"toolCode": CreateTicketConfirmToolCode,
		},
	}, nil
}

func (t *CreateTicketConfirmTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...einotool.Option) (string, error) {
	if t == nil || t.conversation == nil || t.aiAgent == nil {
		return "", fmt.Errorf("ticket confirmation tool not initialized")
	}
	wasInterrupted, hasState, state := componenttool.GetInterruptState[CreateTicketConfirmState](ctx)
	if !wasInterrupted {
		req, err := t.buildCreateRequest(argumentsInJSON)
		if err != nil {
			return "", err
		}
		info := CreateTicketConfirmInterruptInfo{
			Type:    "ticket_creation_confirmation",
			Message: t.buildConfirmationPrompt(req),
		}
		return "", componenttool.StatefulInterrupt(ctx, info, CreateTicketConfirmState{Request: req})
	}
	if !hasState {
		return "", fmt.Errorf("ticket confirmation state missing")
	}
	isResumeTarget, hasData, resumeText := componenttool.GetResumeContext[string](ctx)
	if !isResumeTarget {
		info := CreateTicketConfirmInterruptInfo{
			Type:    "ticket_creation_confirmation",
			Message: t.buildConfirmationPrompt(state.Request),
		}
		return "", componenttool.StatefulInterrupt(ctx, info, state)
	}
	if !hasData {
		info := CreateTicketConfirmInterruptInfo{
			Type:    "ticket_creation_confirmation",
			Message: "请回复“确认”或“取消”。",
		}
		return "", componenttool.StatefulInterrupt(ctx, info, state)
	}
	decision := ParseConfirmationDecision(resumeText)
	switch decision {
	case DecisionConfirm:
		item, err := services.TicketService.CreateFromConversation(state.Request, t.buildAIPrincipal())
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("工单已创建，工单号：%s，标题：%s。", strings.TrimSpace(item.TicketNo), strings.TrimSpace(item.Title)), nil
	case DecisionCancel:
		return "已取消本次工单创建。", nil
	default:
		info := CreateTicketConfirmInterruptInfo{
			Type:    "ticket_creation_confirmation",
			Message: "我需要你的明确确认，请直接回复“确认”或“取消”。",
		}
		return "", componenttool.StatefulInterrupt(ctx, info, state)
	}
}

func (t *CreateTicketConfirmTool) buildCreateRequest(argumentsInJSON string) (request.CreateTicketFromConversationRequest, error) {
	req := request.CreateTicketFromConversationRequest{
		ConversationID:     t.conversation.ID,
		SyncToConversation: true,
	}
	raw := make(map[string]any)
	if strings.TrimSpace(argumentsInJSON) != "" {
		if err := json.Unmarshal([]byte(argumentsInJSON), &raw); err != nil {
			return req, fmt.Errorf("invalid create ticket arguments: %w", err)
		}
	}
	req.Title = strings.TrimSpace(getStringValue(raw, "title"))
	req.Description = strings.TrimSpace(getStringValue(raw, "description"))
	req.Priority = getInt64Value(raw, "priority")
	req.Severity = int(getInt64Value(raw, "severity"))
	if req.Title == "" {
		req.Title = strings.TrimSpace(t.conversation.Subject)
	}
	if req.Description == "" {
		req.Description = strings.TrimSpace(t.conversation.LastMessageSummary)
	}
	if strings.TrimSpace(req.Title) == "" {
		return req, fmt.Errorf("ticket title is required")
	}
	return req, nil
}

func (t *CreateTicketConfirmTool) buildConfirmationPrompt(req request.CreateTicketFromConversationRequest) string {
	return fmt.Sprintf("我准备为你创建工单。\n标题：%s\n描述：%s\n请直接回复“确认”或“取消”。",
		strings.TrimSpace(req.Title), strings.TrimSpace(req.Description))
}

func (t *CreateTicketConfirmTool) buildAIPrincipal() *dto.AuthPrincipal {
	username := "AI"
	if strings.TrimSpace(t.aiAgent.Name) != "" {
		username = strings.TrimSpace(t.aiAgent.Name)
	}
	return &dto.AuthPrincipal{
		UserID:   0,
		Username: username,
		Nickname: username,
	}
}
