package builders

import (
	"strings"

	"cs-agent/internal/models"
	"cs-agent/internal/pkg/dto/response"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/pkg/utils"
	"cs-agent/internal/services"

	"github.com/mlogclub/simple/sqls"
)

func BuildConversation(item *models.Conversation) response.ConversationResponse {
	agentReadState, customerReadState := services.ConversationReadStateService.GetConversationReadStates(item.ID)
	ret := response.ConversationResponse{
		ID:                        item.ID,
		AIAgentID:                 item.AIAgentID,
		ChannelID:                 item.ChannelID,
		CustomerID:                item.CustomerID,
		ExternalSource:            item.ExternalSource,
		ExternalID:                item.ExternalID,
		Subject:                   item.Subject,
		Status:                    item.Status,
		ServiceMode:               item.ServiceMode,
		Priority:                  item.Priority,
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
		CustomerOnline:            services.WsService.IsGuestOnline(item.ExternalID),
		ClosedAt:                  utils.FormatTimePtr(item.ClosedAt),
		ClosedBy:                  item.ClosedBy,
		CloseReason:               item.CloseReason,
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

func BuildMessages(list []models.Message) []response.MessageResponse {
	if len(list) == 0 {
		return nil
	}
	agentReadState, customerReadState := services.ConversationReadStateService.GetConversationReadStates(list[0].ConversationID)
	aiSenderNames, userSenderNames := collectMessageSenderNameMaps(list)
	agentProfiles := collectAgentProfilesByMessages(list)
	ret := make([]response.MessageResponse, 0, len(list))
	for i := range list {
		ret = append(ret, BuildMessageWithReadStates(&list[i], agentReadState, customerReadState, aiSenderNames, userSenderNames, agentProfiles))
	}
	return ret
}

func BuildMessage(item *models.Message) response.MessageResponse {
	agentReadState, customerReadState := services.ConversationReadStateService.GetConversationReadStates(item.ConversationID)
	return BuildMessageWithReadStates(item, agentReadState, customerReadState, nil, nil, nil)
}

func BuildMessageWithReadStates(item *models.Message, agentReadState, customerReadState *models.ConversationReadState, aiSenderNames, userSenderNames map[int64]string, agentProfiles map[int64]*models.AgentProfile) response.MessageResponse {
	content, payload := utils.BuildRenderableMessage(item)
	ret := response.MessageResponse{
		ID:              item.ID,
		ConversationID:  item.ConversationID,
		ClientMsgID:     item.ClientMsgID,
		SenderType:      item.SenderType,
		SenderID:        item.SenderID,
		MessageType:     item.MessageType,
		Content:         content,
		Payload:         payload,
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
			if aiSenderNames != nil {
				ret.SenderName = aiSenderNames[item.SenderID]
			} else if aiAgent := services.AIAgentService.Get(item.SenderID); aiAgent != nil {
				ret.SenderName = aiAgent.Name
			}
		} else if item.SenderType == enums.IMSenderTypeAgent {
			profile := agentProfiles[item.SenderID]
			if profile != nil {
				if dn := strings.TrimSpace(profile.DisplayName); dn != "" {
					ret.SenderName = dn
				}
				if av := strings.TrimSpace(profile.Avatar); av != "" {
					ret.SenderAvatar = av
				}
			}
			if ret.SenderName == "" {
				if userSenderNames != nil {
					ret.SenderName = userSenderNames[item.SenderID]
				} else if user := services.UserService.Get(item.SenderID); user != nil {
					ret.SenderName = user.Nickname
					if ret.SenderName == "" {
						ret.SenderName = user.Username
					}
				}
			}
		} else if userSenderNames != nil {
			ret.SenderName = userSenderNames[item.SenderID]
		} else if user := services.UserService.Get(item.SenderID); user != nil {
			ret.SenderName = user.Nickname
			if ret.SenderName == "" {
				ret.SenderName = user.Username
			}
		}
	}
	return ret
}

func collectAgentProfilesByMessages(list []models.Message) map[int64]*models.AgentProfile {
	var agentUserIDs []int64
	seen := make(map[int64]struct{})
	for i := range list {
		m := &list[i]
		if m.SenderType != enums.IMSenderTypeAgent || m.SenderID <= 0 {
			continue
		}
		if _, ok := seen[m.SenderID]; ok {
			continue
		}
		seen[m.SenderID] = struct{}{}
		agentUserIDs = append(agentUserIDs, m.SenderID)
	}
	if len(agentUserIDs) == 0 {
		return nil
	}
	profiles := services.AgentProfileService.Find(sqls.NewCnd().In("user_id", agentUserIDs))
	out := make(map[int64]*models.AgentProfile, len(profiles))
	for i := range profiles {
		out[profiles[i].UserID] = &profiles[i]
	}
	return out
}

func collectMessageSenderNameMaps(list []models.Message) (aiNames map[int64]string, userNames map[int64]string) {
	aiNames = make(map[int64]string)
	userNames = make(map[int64]string)
	var aiIDs, userIDs []int64
	seenAI := make(map[int64]struct{})
	seenUser := make(map[int64]struct{})
	for i := range list {
		m := &list[i]
		if m.SenderID <= 0 {
			continue
		}
		if m.SenderType == enums.IMSenderTypeAI {
			if _, ok := seenAI[m.SenderID]; ok {
				continue
			}
			seenAI[m.SenderID] = struct{}{}
			aiIDs = append(aiIDs, m.SenderID)
			continue
		}
		if _, ok := seenUser[m.SenderID]; ok {
			continue
		}
		seenUser[m.SenderID] = struct{}{}
		userIDs = append(userIDs, m.SenderID)
	}
	for _, a := range services.AIAgentService.FindByIds(aiIDs) {
		aiNames[a.ID] = a.Name
	}
	for _, u := range services.UserService.FindByIds(userIDs) {
		name := u.Nickname
		if name == "" {
			name = u.Username
		}
		userNames[u.ID] = name
	}
	return aiNames, userNames
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
