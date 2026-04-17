package graphs

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"cs-agent/internal/models"
	"cs-agent/internal/pkg/dto"
	"cs-agent/internal/pkg/dto/request"
	"cs-agent/internal/services"

	componenttool "github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

type CreateTicketGraphState struct {
	Request request.CreateTicketFromConversationRequest
}

type CreateTicketGraphInterruptInfo struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

type createTicketGraphArgs struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Priority    int64  `json:"priority"`
	Severity    int    `json:"severity"`
}

func init() {
	schema.RegisterName[CreateTicketGraphState]("cs_agent_create_ticket_graph_state")
	schema.RegisterName[CreateTicketGraphInterruptInfo]("cs_agent_create_ticket_graph_interrupt_info")
}

type CreateTicketGraph struct {
	conversation *models.Conversation
	aiAgent      models.AIAgent
}

func NewCreateTicketGraph(conversation *models.Conversation, aiAgent models.AIAgent) *CreateTicketGraph {
	return &CreateTicketGraph{
		conversation: conversation,
		aiAgent:      aiAgent,
	}
}

func (g *CreateTicketGraph) Run(ctx context.Context, argumentsInJSON string) (string, error) {
	if g == nil || g.conversation == nil {
		return "", fmt.Errorf("create ticket graph not initialized")
	}
	wasInterrupted, hasState, state := componenttool.GetInterruptState[CreateTicketGraphState](ctx)
	if !wasInterrupted {
		req, err := g.buildCreateRequest(argumentsInJSON)
		if err != nil {
			return "", err
		}
		info := CreateTicketGraphInterruptInfo{
			Type:    InterruptTypeTicketCreationConfirmation,
			Message: g.buildConfirmationPrompt(req),
		}
		return "", componenttool.StatefulInterrupt(ctx, info, CreateTicketGraphState{Request: req})
	}
	if !hasState {
		return "", fmt.Errorf("create ticket graph state missing")
	}
	isResumeTarget, hasData, resumeText := componenttool.GetResumeContext[string](ctx)
	if !isResumeTarget {
		info := CreateTicketGraphInterruptInfo{
			Type:    InterruptTypeTicketCreationConfirmation,
			Message: g.buildConfirmationPrompt(state.Request),
		}
		return "", componenttool.StatefulInterrupt(ctx, info, state)
	}
	if !hasData {
		info := CreateTicketGraphInterruptInfo{
			Type:    InterruptTypeTicketCreationConfirmation,
			Message: ConfirmOrCancelPrompt,
		}
		return "", componenttool.StatefulInterrupt(ctx, info, state)
	}
	decision := ParseConfirmationDecision(resumeText)
	switch decision {
	case ConfirmationDecisionConfirm:
		item, err := services.TicketService.CreateFromConversation(state.Request, g.buildAIPrincipal())
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("工单已创建，工单号：%s，标题：%s。", strings.TrimSpace(item.TicketNo), strings.TrimSpace(item.Title)), nil
	case ConfirmationDecisionCancel:
		return CancelCreateTicketReply, nil
	default:
		info := CreateTicketGraphInterruptInfo{
			Type:    InterruptTypeTicketCreationConfirmation,
			Message: NeedExplicitConfirmationPrompt,
		}
		return "", componenttool.StatefulInterrupt(ctx, info, state)
	}
}

func (g *CreateTicketGraph) buildCreateRequest(argumentsInJSON string) (request.CreateTicketFromConversationRequest, error) {
	req := request.CreateTicketFromConversationRequest{
		ConversationID:     g.conversation.ID,
		SyncToConversation: true,
	}
	var args createTicketGraphArgs
	if strings.TrimSpace(argumentsInJSON) != "" {
		if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
			return req, fmt.Errorf("invalid create ticket arguments: %w", err)
		}
	}
	req.Title = strings.TrimSpace(args.Title)
	req.Description = strings.TrimSpace(args.Description)
	req.Priority = args.Priority
	req.Severity = args.Severity
	if req.Title == "" {
		req.Title = strings.TrimSpace(g.conversation.Subject)
	}
	if req.Description == "" {
		req.Description = strings.TrimSpace(g.conversation.LastMessageSummary)
	}
	if strings.TrimSpace(req.Title) == "" {
		return req, fmt.Errorf("ticket title is required")
	}
	return req, nil
}

func (g *CreateTicketGraph) buildConfirmationPrompt(req request.CreateTicketFromConversationRequest) string {
	return fmt.Sprintf("我准备为你创建工单。\n标题：%s\n描述：%s\n请直接回复“确认”或“取消”。",
		strings.TrimSpace(req.Title), strings.TrimSpace(req.Description))
}

func (g *CreateTicketGraph) buildAIPrincipal() *dto.AuthPrincipal {
	username := "AI"
	if strings.TrimSpace(g.aiAgent.Name) != "" {
		username = strings.TrimSpace(g.aiAgent.Name)
	}
	return &dto.AuthPrincipal{
		UserID:   0,
		Username: username,
		Nickname: username,
	}
}
