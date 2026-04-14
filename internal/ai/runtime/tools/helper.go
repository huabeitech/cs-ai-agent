package tools

import (
	"strings"

	"cs-agent/internal/ai/runtime/registry"
	"cs-agent/internal/pkg/toolx"
)

type Decision string

const (
	DecisionConfirm Decision = "confirm"
	DecisionCancel  Decision = "cancel"
)

func ParseConfirmationDecision(value string) Decision {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return ""
	}
	confirmWords := []string{"确认", "是", "好的", "可以", "ok", "yes", "继续", "同意"}
	for _, item := range confirmWords {
		if strings.Contains(value, item) {
			return DecisionConfirm
		}
	}
	cancelWords := []string{"取消", "不用", "不需要", "算了", "no"}
	for _, item := range cancelWords {
		if strings.Contains(value, item) {
			return DecisionCancel
		}
	}
	return ""
}

func NewRuntimeStaticTool(toolCode string) registry.Tool {
	switch toolx.NormalizeToolCodeAlias(strings.TrimSpace(toolCode)) {
	case toolx.GraphTriageServiceRequest.Code:
		return NewTriageServiceRequestTool()
	case toolx.GraphAnalyzeConversation.Code:
		return NewAnalyzeConversationTool()
	case toolx.GraphPrepareTicketDraft.Code:
		return NewPrepareTicketDraftTool()
	case toolx.GraphCreateTicketConfirm.Code:
		return NewCreateTicketGraphTool()
	case toolx.GraphHandoffConversation.Code:
		return NewHandoffGraphTool()
	default:
		return nil
	}
}
