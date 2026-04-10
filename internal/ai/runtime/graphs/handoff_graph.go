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
			Type:    "handoff_confirmation",
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
			Type:    "handoff_confirmation",
			Message: g.buildConfirmationPrompt(state.Reason),
		}
		return "", componenttool.StatefulInterrupt(ctx, info, state)
	}
	if !hasData {
		info := HandoffGraphInterruptInfo{
			Type:    "handoff_confirmation",
			Message: "请回复“确认”或“取消”。",
		}
		return "", componenttool.StatefulInterrupt(ctx, info, state)
	}
	switch parseHandoffDecision(resumeText) {
	case graphDecisionConfirm:
		if err := services.ConversationService.HandoffByAI(g.conversation.ID, g.aiAgent, state.Reason); err != nil {
			return "", err
		}
		return "已为你转接人工客服，请稍候。", nil
	case graphDecisionCancel:
		return "已取消本次转人工。", nil
	default:
		info := HandoffGraphInterruptInfo{
			Type:    "handoff_confirmation",
			Message: "我需要你的明确确认，请直接回复“确认”或“取消”。",
		}
		return "", componenttool.StatefulInterrupt(ctx, info, state)
	}
}

func (g *HandoffGraph) buildReason(argumentsInJSON string) (string, error) {
	reason := "用户需要转人工支持"
	raw := make(map[string]any)
	if strings.TrimSpace(argumentsInJSON) != "" {
		if err := json.Unmarshal([]byte(argumentsInJSON), &raw); err != nil {
			return "", fmt.Errorf("invalid handoff arguments: %w", err)
		}
	}
	if parsed := strings.TrimSpace(graphGetStringValue(raw, "reason")); parsed != "" {
		reason = parsed
	}
	return reason, nil
}

func (g *HandoffGraph) buildConfirmationPrompt(reason string) string {
	return fmt.Sprintf("我准备为你转接人工客服。\n原因：%s\n请直接回复“确认”或“取消”。", strings.TrimSpace(reason))
}

type graphDecision string

const (
	graphDecisionConfirm graphDecision = "confirm"
	graphDecisionCancel  graphDecision = "cancel"
)

func parseHandoffDecision(value string) graphDecision {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return ""
	}
	confirmWords := []string{"确认", "是", "好的", "可以", "ok", "yes", "继续", "同意"}
	for _, item := range confirmWords {
		if strings.Contains(value, item) {
			return graphDecisionConfirm
		}
	}
	cancelWords := []string{"取消", "不用", "不需要", "算了", "no"}
	for _, item := range cancelWords {
		if strings.Contains(value, item) {
			return graphDecisionCancel
		}
	}
	return ""
}

func graphGetStringValue(data map[string]any, key string) string {
	if len(data) == 0 {
		return ""
	}
	value, ok := data[key]
	if !ok || value == nil {
		return ""
	}
	text, _ := value.(string)
	return text
}
