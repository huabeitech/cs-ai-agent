package runtime

import (
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/enums"

	"github.com/mlogclub/simple/common/strs"
)

type replyEligibility struct{}

func newReplyEligibility() *replyEligibility {
	return &replyEligibility{}
}

func (e *replyEligibility) CanReply(conversation models.Conversation, message models.Message, aiAgent models.AIAgent) bool {
	if message.SenderType != enums.IMSenderTypeCustomer {
		return false
	}
	if conversation.HandoffAt != nil || conversation.CurrentAssigneeID > 0 {
		return false
	}
	if aiAgent.ServiceMode == enums.IMConversationServiceModeHumanOnly {
		return false
	}
	if strs.IsBlank(message.Content) {
		return false
	}
	return true
}
