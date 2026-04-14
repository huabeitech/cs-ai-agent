package graphs

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"cs-agent/internal/models"
	"cs-agent/internal/services"

	componenttool "github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

type HandoffGraphState struct {
	Reason string `json:"reason"`
}

type HandoffGraphInterruptInfo struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

type handoffGraphArgs struct {
	Reason string `json:"reason"`
}

func init() {
	schema.RegisterName[HandoffGraphState]("cs_agent_handoff_graph_state")
	schema.RegisterName[HandoffGraphInterruptInfo]("cs_agent_handoff_graph_interrupt_info")
}

type HandoffGraph struct {
	conversation *models.Conversation
	aiAgent      *models.AIAgent
}

func NewHandoffGraph(conversation *models.Conversation, aiAgent *models.AIAgent) *HandoffGraph {
	return &HandoffGraph{
		conversation: conversation,
		aiAgent:      aiAgent,
	}
}

func (g *HandoffGraph) Run(ctx context.Context, argumentsInJSON string) (string, error) {
	if g == nil || g.conversation == nil || g.aiAgent == nil {
		return "", fmt.Errorf("handoff graph not initialized")
	}
	wasInterrupted, hasState, state := componenttool.GetInterruptState[HandoffGraphState](ctx)
	if !wasInterrupted {
		reason, err := g.buildReason(argumentsInJSON)
		if err != nil {
			return "", err
		}
		info := HandoffGraphInterruptInfo{
			Type:    InterruptTypeHandoffConfirmation,
			Message: g.buildConfirmationPrompt(reason),
		}
		return "", componenttool.StatefulInterrupt(ctx, info, HandoffGraphState{Reason: reason})
	}
	if !hasState {
		return "", fmt.Errorf("handoff graph state missing")
	}
	isResumeTarget, hasData, resumeText := componenttool.GetResumeContext[string](ctx)
	if !isResumeTarget {
		info := HandoffGraphInterruptInfo{
			Type:    InterruptTypeHandoffConfirmation,
			Message: g.buildConfirmationPrompt(state.Reason),
		}
		return "", componenttool.StatefulInterrupt(ctx, info, state)
	}
	if !hasData {
		info := HandoffGraphInterruptInfo{
			Type:    InterruptTypeHandoffConfirmation,
			Message: ConfirmOrCancelPrompt,
		}
		return "", componenttool.StatefulInterrupt(ctx, info, state)
	}
	switch parseHandoffDecision(resumeText) {
	case ConfirmationDecisionConfirm:
		if err := services.ConversationService.HandoffByAI(g.conversation.ID, g.aiAgent, state.Reason); err != nil {
			return "", err
		}
		return "已为你转接人工客服，请稍候。", nil
	case ConfirmationDecisionCancel:
		return CancelHandoffReply, nil
	default:
		info := HandoffGraphInterruptInfo{
			Type:    InterruptTypeHandoffConfirmation,
			Message: NeedExplicitConfirmationPrompt,
		}
		return "", componenttool.StatefulInterrupt(ctx, info, state)
	}
}

func (g *HandoffGraph) buildReason(argumentsInJSON string) (string, error) {
	reason := "用户需要转人工支持"
	var args handoffGraphArgs
	if strings.TrimSpace(argumentsInJSON) != "" {
		if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
			return "", fmt.Errorf("invalid handoff arguments: %w", err)
		}
	}
	if parsed := strings.TrimSpace(args.Reason); parsed != "" {
		reason = parsed
	}
	return reason, nil
}

func (g *HandoffGraph) buildConfirmationPrompt(reason string) string {
	return fmt.Sprintf("我准备为你转接人工客服。\n原因：%s\n请直接回复“确认”或“取消”。", strings.TrimSpace(reason))
}

func parseHandoffDecision(value string) ConfirmationDecision {
	return ParseConfirmationDecision(value)
}
