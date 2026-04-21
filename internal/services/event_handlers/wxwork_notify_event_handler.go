package event_handlers

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"cs-agent/internal/events"
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/pkg/eventbus"
	"cs-agent/internal/services"
)

func registerWxWorkNotifyEventHandlers() {
	eventbus.
		Register(eventbus.WithErrorHandler[events.TicketCreatedEvent](handleWxWorkNotifyEventError)).
		Subscribe(handleTicketCreatedNotify)
	eventbus.
		Register(eventbus.WithErrorHandler[events.TicketAssignedEvent](handleWxWorkNotifyEventError)).
		Subscribe(handleTicketAssignedNotify)
	eventbus.
		Register(eventbus.WithErrorHandler[events.ConversationAssignedEvent](handleWxWorkNotifyEventError)).
		Subscribe(handleConversationAssignedNotify)
}

func handleWxWorkNotifyEventError(ctx context.Context, err error) {
	slog.Warn("handle wxwork notify event failed", "error", err)
}

func handleTicketCreatedNotify(ctx context.Context, event events.TicketCreatedEvent) error {
	if event.TicketID <= 0 {
		return nil
	}
	ticket := services.TicketService.Get(event.TicketID)
	if ticket == nil {
		return nil
	}
	return services.WxWorkNotifyService.SendTextToAssigneeOrDefault(ticket.CurrentAssigneeID, "工单创建提醒", buildTicketCreatedNotifyBody(ticket))
}

func handleTicketAssignedNotify(ctx context.Context, event events.TicketAssignedEvent) error {
	if event.TicketID <= 0 || event.ToUserID <= 0 {
		return nil
	}
	ticket := services.TicketService.Get(event.TicketID)
	if ticket == nil {
		return nil
	}
	return services.WxWorkNotifyService.SendTextToAssigneeOrDefault(event.ToUserID, "工单指派提醒", buildTicketAssignedNotifyBody(ticket, event.ToUserID, event.Reason))
}

func handleConversationAssignedNotify(ctx context.Context, event events.ConversationAssignedEvent) error {
	if event.ConversationID <= 0 || event.ToUserID <= 0 {
		return nil
	}
	conversation := services.ConversationService.Get(event.ConversationID)
	if conversation == nil {
		return nil
	}
	return services.WxWorkNotifyService.SendTextToAssigneeOrDefault(event.ToUserID, conversationAssignedNotifyTitle(event.AssignType), buildConversationAssignedNotifyBody(conversation, event.ToUserID, event.Reason, event.AssignType))
}

func conversationAssignedNotifyTitle(assignType string) string {
	switch strings.TrimSpace(assignType) {
	case events.ConversationAssignTypeTransfer:
		return "会话转接提醒"
	case events.ConversationAssignTypeAutoAssign:
		return "会话自动分配提醒"
	default:
		return "会话分配提醒"
	}
}

func buildConversationAssignedNotifyBody(conversation *models.Conversation, assigneeID int64, reason string, assignType string) string {
	if conversation == nil {
		return ""
	}
	reasonLabel := "分配原因"
	if strings.TrimSpace(assignType) == events.ConversationAssignTypeTransfer {
		reasonLabel = "转接原因"
	}
	lines := []string{
		fmt.Sprintf("会话ID: #%d", conversation.ID),
		fmt.Sprintf("会话主题: %s", defaultIfBlank(conversation.Subject, "-")),
		fmt.Sprintf("接入渠道: %s", enums.GetExternalSourceLabel(conversation.ExternalSource)),
		fmt.Sprintf("当前状态: %s", enums.GetIMConversationStatusLabel(conversation.Status)),
		fmt.Sprintf("处理人: %s", resolveNotifyUserLabel(assigneeID)),
	}
	if strings.TrimSpace(reason) != "" {
		lines = append(lines, fmt.Sprintf("%s: %s", reasonLabel, strings.TrimSpace(reason)))
	}
	lines = append(lines, fmt.Sprintf("时间: %s", time.Now().Format("2006-01-02 15:04:05")))
	return strings.Join(lines, "\n")
}

func buildTicketCreatedNotifyBody(ticket *models.Ticket) string {
	if ticket == nil {
		return ""
	}
	lines := []string{
		fmt.Sprintf("工单号: %s", defaultIfBlank(ticket.TicketNo, fmt.Sprintf("#%d", ticket.ID))),
		fmt.Sprintf("工单标题: %s", defaultIfBlank(ticket.Title, "-")),
		fmt.Sprintf("工单来源: %s", defaultIfBlank(string(ticket.Source), "-")),
		fmt.Sprintf("当前状态: %s", enums.GetTicketStatusLabel(ticket.Status)),
	}
	if ticket.CurrentAssigneeID > 0 {
		lines = append(lines, fmt.Sprintf("处理人: %s", resolveNotifyUserLabel(ticket.CurrentAssigneeID)))
	}
	lines = append(lines, fmt.Sprintf("时间: %s", time.Now().Format("2006-01-02 15:04:05")))
	return strings.Join(lines, "\n")
}

func buildTicketAssignedNotifyBody(ticket *models.Ticket, assigneeID int64, reason string) string {
	if ticket == nil {
		return ""
	}
	lines := []string{
		fmt.Sprintf("工单号: %s", defaultIfBlank(ticket.TicketNo, fmt.Sprintf("#%d", ticket.ID))),
		fmt.Sprintf("工单标题: %s", defaultIfBlank(ticket.Title, "-")),
		fmt.Sprintf("当前状态: %s", enums.GetTicketStatusLabel(ticket.Status)),
		fmt.Sprintf("处理人: %s", resolveNotifyUserLabel(assigneeID)),
	}
	if strings.TrimSpace(reason) != "" {
		lines = append(lines, fmt.Sprintf("指派原因: %s", strings.TrimSpace(reason)))
	}
	lines = append(lines, fmt.Sprintf("时间: %s", time.Now().Format("2006-01-02 15:04:05")))
	return strings.Join(lines, "\n")
}

func resolveNotifyUserLabel(userID int64) string {
	if userID <= 0 {
		return "-"
	}
	user := services.UserService.Get(userID)
	if user == nil {
		return fmt.Sprintf("用户#%d", userID)
	}
	if nickname := strings.TrimSpace(user.Nickname); nickname != "" {
		return nickname
	}
	if username := strings.TrimSpace(user.Username); username != "" {
		return username
	}
	return fmt.Sprintf("用户#%d", userID)
}

func defaultIfBlank(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value != "" {
		return value
	}
	return strings.TrimSpace(fallback)
}
