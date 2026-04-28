package event_handlers

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"cs-agent/internal/events"
	"cs-agent/internal/pkg/dto/request"
	"cs-agent/internal/pkg/eventbus"
	"cs-agent/internal/services"

	"github.com/mlogclub/simple/common/strs"
)

func init() {
	eventbus.
		Register[events.TicketAssignedEvent]().
		Subscribe(handleTicketAssignedInAppNotification)
	eventbus.
		Register[events.ConversationAssignedEvent]().
		Subscribe(handleConversationAssignedInAppNotification)
}

func handleTicketAssignedInAppNotification(ctx context.Context, event events.TicketAssignedEvent) error {
	if event.TicketID <= 0 || event.ToUserID <= 0 {
		return nil
	}
	ticket := services.TicketService.Get(event.TicketID)
	if ticket == nil {
		return nil
	}
	content := fmt.Sprintf("工单 %s 已指派给你", strs.DefaultIfBlank(ticket.TicketNo, fmt.Sprintf("#%d", ticket.ID)))
	if title := strings.TrimSpace(ticket.Title); title != "" {
		content = content + "\n" + title
	}
	if reason := strings.TrimSpace(event.Reason); reason != "" {
		content = content + "\n指派原因: " + reason
	}
	_, err := services.NotificationService.CreateAndPush(request.CreateNotificationRequest{
		RecipientUserID:  event.ToUserID,
		Title:            "工单指派提醒",
		Content:          content,
		NotificationType: "ticket_assigned",
		BizType:          "ticket",
		BizID:            ticket.ID,
		ActionURL:        fmt.Sprintf("/dashboard/tickets/%d", ticket.ID),
	})
	if err != nil {
		slog.Error("create ticket assigned in-app notification failed", "error", err, "ticketId", event.TicketID, "toUserId", event.ToUserID)
	}
	return nil
}

func handleConversationAssignedInAppNotification(ctx context.Context, event events.ConversationAssignedEvent) error {
	if event.ConversationID <= 0 || event.ToUserID <= 0 {
		return nil
	}
	conversation := services.ConversationService.Get(event.ConversationID)
	if conversation == nil {
		return nil
	}
	content := fmt.Sprintf("会话 #%d 已分配给你", conversation.ID)
	if summary := strings.TrimSpace(services.ConversationService.BuildConversationSummary(conversation)); summary != "" {
		content = content + "\n" + summary
	}
	if reason := strings.TrimSpace(event.Reason); reason != "" {
		content = content + "\n分配原因: " + reason
	}
	_, err := services.NotificationService.CreateAndPush(request.CreateNotificationRequest{
		RecipientUserID:  event.ToUserID,
		Title:            conversationAssignedNotifyTitle(event.AssignType),
		Content:          content,
		NotificationType: "conversation_assigned",
		BizType:          "conversation",
		BizID:            conversation.ID,
		ActionURL:        fmt.Sprintf("/dashboard/conversations?conversationId=%d", conversation.ID),
	})
	if err != nil {
		slog.Error("create conversation assigned in-app notification failed", "error", err, "conversationId", event.ConversationID, "toUserId", event.ToUserID)
	}
	return nil
}
