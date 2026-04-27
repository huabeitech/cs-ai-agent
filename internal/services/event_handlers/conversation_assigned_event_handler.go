package event_handlers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"cs-agent/internal/events"
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/pkg/eventbus"
	"cs-agent/internal/services"

	"github.com/mlogclub/simple/common/strs"
)

func init() {
	eventbus.
		Register[events.ConversationAssignedEvent]().
		Subscribe(handleConversationAssignedNotify)
}

func handleConversationAssignedNotify(ctx context.Context, event events.ConversationAssignedEvent) error {
	if event.ConversationID <= 0 || event.ToUserID <= 0 {
		return nil
	}
	conversation := services.ConversationService.Get(event.ConversationID)
	if conversation == nil {
		return nil
	}
	return services.WxWorkNotifyService.SendTextToAssigneeOrDefault(event.ToUserID,
		conversationAssignedNotifyTitle(event.AssignType),
		buildConversationAssignedNotifyBody(conversation, event.ToUserID, event.Reason, event.AssignType))
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
		fmt.Sprintf("会话摘要: %s", strs.DefaultIfBlank(services.ConversationService.BuildConversationSummary(conversation), "-")),
		fmt.Sprintf("接入渠道: %s", resolveConversationChannelLabel(conversation)),
		fmt.Sprintf("当前状态: %s", enums.GetIMConversationStatusLabel(conversation.Status)),
		fmt.Sprintf("处理人: %s", resolveNotifyUserLabel(assigneeID)),
	}
	if strings.TrimSpace(reason) != "" {
		lines = append(lines, fmt.Sprintf("%s: %s", reasonLabel, strings.TrimSpace(reason)))
	}
	lines = append(lines, fmt.Sprintf("时间: %s", time.Now().Format("2006-01-02 15:04:05")))
	return strings.Join(lines, "\n")
}

func resolveConversationChannelLabel(conversation *models.Conversation) string {
	if conversation == nil || conversation.ChannelID <= 0 {
		return "-"
	}
	if channel := services.ChannelService.Get(conversation.ChannelID); channel != nil {
		return strs.DefaultIfBlank(channel.Name, channel.ChannelType)
	}
	return "-"
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
