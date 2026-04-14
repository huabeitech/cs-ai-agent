package runtime

import (
	"encoding/json"
	"strings"
	"time"

	"cs-agent/internal/models"
	svc "cs-agent/internal/services"
)

type interruptMessagePreview struct {
	Message string `json:"message"`
}

func buildConversationInterrupt(conversation models.Conversation, message models.Message, aiAgent models.AIAgent, summary *Summary) *models.ConversationInterrupt {
	if summary == nil {
		return nil
	}
	now := time.Now()
	item := svc.ConversationInterruptService.GetByCheckPointID(summary.CheckPointID)
	if item == nil {
		item = &models.ConversationInterrupt{
			CheckPointID: summary.CheckPointID,
			CreatedAt:    now,
		}
	}
	item.ConversationID = conversation.ID
	item.AIAgentID = aiAgent.ID
	item.SourceMessageID = message.ID
	item.InterruptID = firstInterruptID(summary)
	item.InterruptType = firstInterruptType(summary)
	item.Status = "pending"
	item.PromptText = resolveInterruptPrompt(summary)
	item.UpdatedAt = now
	return item
}

func resolveInterruptPrompt(summary *Summary) string {
	if summary == nil || len(summary.Interrupts) == 0 {
		return "请继续补充信息后再试。"
	}
	if prompt := extractInterruptMessage(summary.Interrupts[0].InfoPreview); prompt != "" {
		return prompt
	}
	if prompt := strings.TrimSpace(summary.Interrupts[0].InfoPreview); prompt != "" {
		return prompt
	}
	return "请继续补充信息后再试。"
}

func extractInterruptMessage(infoPreview string) string {
	infoPreview = strings.TrimSpace(infoPreview)
	if infoPreview == "" {
		return ""
	}
	var payload interruptMessagePreview
	if err := json.Unmarshal([]byte(infoPreview), &payload); err != nil {
		return ""
	}
	return strings.TrimSpace(payload.Message)
}

func firstInterruptID(summary *Summary) string {
	if summary == nil || len(summary.Interrupts) == 0 {
		return ""
	}
	return strings.TrimSpace(summary.Interrupts[0].ID)
}

func firstInterruptType(summary *Summary) string {
	if summary == nil || len(summary.Interrupts) == 0 {
		return ""
	}
	return strings.TrimSpace(summary.Interrupts[0].Type)
}

func isCheckpointMissingError(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(strings.TrimSpace(err.Error()))
	return strings.Contains(message, "failed to load from checkpoint") && strings.Contains(message, "not exist")
}
