package builders

import (
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/dto/response"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/pkg/utils"
	"cs-agent/internal/services"

	"github.com/mlogclub/simple/sqls"
)

func BuildConversationResponse(item *models.Conversation) response.ConversationResponse {
	agentReadState, customerReadState := services.ConversationReadStateService.GetConversationReadStates(item.ID)
	ret := response.ConversationResponse{
		ID:                        item.ID,
		AIAgentID:                 item.AIAgentID,
		CustomerID:                item.CustomerID,
		ChannelType:               item.ChannelType,
		Subject:                   item.Subject,
		Status:                    item.Status,
		ServiceMode:               item.ServiceMode,
		Priority:                  item.Priority,
		SourceUserID:              item.SourceUserID,
		ExternalUserID:            item.ExternalUserID,
		CurrentAssigneeID:         item.CurrentAssigneeID,
		LastMessageID:             item.LastMessageID,
		LastMessageAt:             utils.FormatTime(item.LastMessageAt),
		LastActiveAt:              utils.FormatTime(item.LastActiveAt),
		LastMessageSummary:        item.LastMessageSummary,
		CustomerUnreadCount:       item.CustomerUnreadCount,
		AgentUnreadCount:          item.AgentUnreadCount,
		CustomerLastReadMessageID: readStateMessageID(customerReadState),
		CustomerLastReadSeqNo:     readStateSeqNo(customerReadState),
		CustomerLastReadAt:        readStateAt(customerReadState),
		AgentLastReadMessageID:    readStateMessageID(agentReadState),
		AgentLastReadSeqNo:        readStateSeqNo(agentReadState),
		AgentLastReadAt:           readStateAt(agentReadState),
		ClosedAt:                  utils.FormatTimePtr(item.ClosedAt),
		ClosedBy:                  item.ClosedBy,
		CloseReason:               item.CloseReason,
		Tags:                      buildConversationTagResponses(item.ID),
	}
	if item.CurrentAssigneeID > 0 {
		if user := services.UserService.Get(item.CurrentAssigneeID); user != nil {
			ret.CurrentAssigneeName = user.Nickname
			if ret.CurrentAssigneeName == "" {
				ret.CurrentAssigneeName = user.Username
			}
		}
	}
	if item.ClosedBy > 0 {
		if user := services.UserService.Get(item.ClosedBy); user != nil {
			ret.ClosedByName = user.Nickname
			if ret.ClosedByName == "" {
				ret.ClosedByName = user.Username
			}
		}
	}
	return ret
}

func buildConversationTagResponses(conversationID int64) []response.ConversationTagResponse {
	relations := services.ConversationTagService.Find(sqls.NewCnd().Eq("conversation_id", conversationID))
	if len(relations) == 0 {
		return nil
	}
	ret := make([]response.ConversationTagResponse, 0, len(relations))
	for _, relation := range relations {
		// TODO 循环查库
		tag := services.TagService.Get(relation.TagID)
		if tag == nil {
			continue
		}
		ret = append(ret, response.ConversationTagResponse{
			ID:   tag.ID,
			Name: tag.Name,
		})
	}
	return ret
}

func BuildParticipantResponses(conversationID int64) []response.ConversationParticipantResponse {
	list := services.ConversationParticipantService.Find(sqls.NewCnd().Eq("conversation_id", conversationID).Asc("id"))
	if len(list) == 0 {
		return nil
	}
	ret := make([]response.ConversationParticipantResponse, 0, len(list))
	for _, item := range list {
		ret = append(ret, response.ConversationParticipantResponse{
			ID:                    item.ID,
			ParticipantType:       item.ParticipantType,
			ParticipantID:         item.ParticipantID,
			ExternalParticipantID: item.ExternalParticipantID,
			JoinedAt:              utils.FormatTimePtr(item.JoinedAt),
			LeftAt:                utils.FormatTimePtr(item.LeftAt),
			Status:                item.Status,
		})
	}
	return ret
}

func BuildMessageResponses(list []models.Message) []response.MessageResponse {
	if len(list) == 0 {
		return nil
	}
	agentReadState, customerReadState := services.ConversationReadStateService.GetConversationReadStates(list[0].ConversationID)
	ret := make([]response.MessageResponse, 0, len(list))
	for i := range list {
		ret = append(ret, BuildMessageResponseWithReadStates(&list[i], agentReadState, customerReadState))
	}
	return ret
}

func BuildMessageResponse(item *models.Message) response.MessageResponse {
	agentReadState, customerReadState := services.ConversationReadStateService.GetConversationReadStates(item.ConversationID)
	return BuildMessageResponseWithReadStates(item, agentReadState, customerReadState)
}

func BuildMessageResponseWithReadStates(item *models.Message, agentReadState, customerReadState *models.ConversationReadState) response.MessageResponse {
	ret := response.MessageResponse{
		ID:              item.ID,
		ConversationID:  item.ConversationID,
		ClientMsgID:     item.ClientMsgID,
		SenderType:      item.SenderType,
		SenderID:        item.SenderID,
		MessageType:     item.MessageType,
		Content:         item.Content,
		Payload:         item.Payload,
		SeqNo:           item.SeqNo,
		SendStatus:      item.SendStatus,
		SentAt:          utils.FormatTimePtr(item.SentAt),
		DeliveredAt:     utils.FormatTimePtr(item.DeliveredAt),
		ReadAt:          utils.FormatTimePtr(item.ReadAt),
		CustomerRead:    isMessageRead(item, customerReadState),
		CustomerReadAt:  readMessageAt(item, customerReadState),
		AgentRead:       isMessageRead(item, agentReadState),
		AgentReadAt:     readMessageAt(item, agentReadState),
		RecalledAt:      utils.FormatTimePtr(item.RecalledAt),
		QuotedMessageID: item.QuotedMessageID,
	}
	if item.SenderID > 0 {
		if item.SenderType == enums.IMSenderTypeAI {
			if aiAgent := services.AIAgentService.Get(item.SenderID); aiAgent != nil {
				ret.SenderName = aiAgent.Name
			}
		} else if user := services.UserService.Get(item.SenderID); user != nil {
			ret.SenderName = user.Nickname
			if ret.SenderName == "" {
				ret.SenderName = user.Username
			}
		}
	}
	return ret
}

func isMessageRead(item *models.Message, state *models.ConversationReadState) bool {
	return item != nil && state != nil && state.LastReadSeqNo >= item.SeqNo
}

func readMessageAt(item *models.Message, state *models.ConversationReadState) string {
	if !isMessageRead(item, state) {
		return ""
	}
	return utils.FormatTimePtr(state.LastReadAt)
}

func readStateMessageID(state *models.ConversationReadState) int64 {
	if state == nil {
		return 0
	}
	return state.LastReadMessageID
}

func readStateSeqNo(state *models.ConversationReadState) int64 {
	if state == nil {
		return 0
	}
	return state.LastReadSeqNo
}

func readStateAt(state *models.ConversationReadState) string {
	if state == nil {
		return ""
	}
	return utils.FormatTimePtr(state.LastReadAt)
}
