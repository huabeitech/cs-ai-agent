package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"cs-agent/internal/events"
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/pkg/errorsx"
	"cs-agent/internal/pkg/eventbus"
	"cs-agent/internal/repositories"

	"github.com/mlogclub/simple/sqls"
)

var ConversationHumanDispatchService = newConversationHumanDispatchService()

const (
	HandoffWaitingMessage  = "已为你转接人工客服，请稍候。"
	HandoffOffHoursMessage = "当前暂不在人工客服服务时间内，你可以先继续描述问题，我会尽力协助；服务时间开始后也可以再次转人工。"
)

type HandoffDecisionType string

const (
	HandoffDecisionAssigned   HandoffDecisionType = "assigned"
	HandoffDecisionTeamPool   HandoffDecisionType = "team_pool"
	HandoffDecisionGlobalPool HandoffDecisionType = "global_pool"
	HandoffDecisionOffHours   HandoffDecisionType = "off_hours"
)

type HandoffDecisionResult struct {
	Decision   HandoffDecisionType
	TeamID     int64
	AssigneeID int64
	Message    string
}

type conversationHumanDispatchService struct{}

func newConversationHumanDispatchService() *conversationHumanDispatchService {
	return &conversationHumanDispatchService{}
}

func (s *conversationHumanDispatchService) HandoffByAI(conversationID int64, aiAgent models.AIAgent, reason string) (*HandoffDecisionResult, error) {
	conversation := ConversationService.Get(conversationID)
	if conversation == nil {
		return nil, errorsx.InvalidParam("会话不存在")
	}
	teamIDs := orderedPositiveIDs(aiAgent.TeamIDs)
	activeTeamIDs := ConversationDispatchService.findActiveScheduleTeamIDs(teamIDs, time.Now())
	if len(activeTeamIDs) == 0 {
		if err := s.createEvent(conversationID, enums.IMEventTypeTransfer, enums.IMSenderTypeAI, aiAgent.ID, "转人工失败：非服务时间", strings.TrimSpace(reason)); err != nil {
			return nil, err
		}
		if err := s.sendAIText(conversationID, aiAgent.ID, HandoffOffHoursMessage); err != nil {
			return nil, err
		}
		return &HandoffDecisionResult{Decision: HandoffDecisionOffHours, Message: HandoffOffHoursMessage}, nil
	}

	if err := s.markHandoff(conversationID, aiAgent, reason); err != nil {
		return nil, err
	}
	return s.dispatchAfterHandoff(conversationID, aiAgent.ID, activeTeamIDs, strings.TrimSpace(reason), true)
}

func (s *conversationHumanDispatchService) ApplyHumanOnlyCreate(conversationID int64, aiAgent models.AIAgent) (*HandoffDecisionResult, error) {
	teamIDs := orderedPositiveIDs(aiAgent.TeamIDs)
	activeTeamIDs := ConversationDispatchService.findActiveScheduleTeamIDs(teamIDs, time.Now())
	if len(activeTeamIDs) == 0 {
		if err := s.moveToGlobalPool(conversationID, aiAgent.Name); err != nil {
			return nil, err
		}
		if err := s.sendAIText(conversationID, aiAgent.ID, HandoffWaitingMessage); err != nil {
			return nil, err
		}
		return &HandoffDecisionResult{Decision: HandoffDecisionGlobalPool, Message: HandoffWaitingMessage}, nil
	}
	return s.dispatchAfterHandoff(conversationID, aiAgent.ID, activeTeamIDs, "仅人工模式新会话", false)
}

func (s *conversationHumanDispatchService) DispatchPendingConversation(conversationID int64, aiAgent models.AIAgent) (*HandoffDecisionResult, error) {
	conversation := ConversationService.Get(conversationID)
	if conversation == nil {
		return nil, errorsx.InvalidParam("会话不存在")
	}
	if conversation.Status != enums.IMConversationStatusPending || conversation.CurrentAssigneeID > 0 {
		return nil, errorsx.InvalidParam("只有待接入未分配会话允许自动分配")
	}
	activeTeamIDs := ConversationDispatchService.findActiveScheduleTeamIDs(orderedPositiveIDs(aiAgent.TeamIDs), time.Now())
	if len(activeTeamIDs) == 0 {
		return &HandoffDecisionResult{Decision: HandoffDecisionOffHours}, nil
	}
	candidates, _, err := ConversationDispatchService.pickDispatchCandidates(activeTeamIDs, time.Now())
	if err != nil {
		return nil, err
	}
	if len(candidates) > 0 {
		dispatched, err := ConversationDispatchService.tryAssignConversation(conversationID, candidates[0].profile, "自动分配")
		if err != nil {
			return nil, err
		}
		if dispatched != nil {
			return &HandoffDecisionResult{
				Decision:   HandoffDecisionAssigned,
				TeamID:     dispatched.CurrentTeamID,
				AssigneeID: dispatched.CurrentAssigneeID,
			}, nil
		}
	}
	teamID := activeTeamIDs[0]
	if err := s.moveToTeamPool(conversationID, teamID, "手动触发自动分配"); err != nil {
		return nil, err
	}
	return &HandoffDecisionResult{Decision: HandoffDecisionTeamPool, TeamID: teamID}, nil
}

func (s *conversationHumanDispatchService) dispatchAfterHandoff(conversationID, aiAgentID int64, activeTeamIDs []int64, reason string, publishAssignEvent bool) (*HandoffDecisionResult, error) {
	if err := s.sendAIText(conversationID, aiAgentID, HandoffWaitingMessage); err != nil {
		return nil, err
	}

	candidates, _, err := ConversationDispatchService.pickDispatchCandidates(activeTeamIDs, time.Now())
	if err != nil {
		return nil, err
	}
	if len(candidates) > 0 {
		dispatched, err := ConversationDispatchService.tryAssignConversation(conversationID, candidates[0].profile, "自动分配")
		if err != nil {
			return nil, err
		}
		if dispatched != nil {
			if publishAssignEvent {
				eventbus.PublishAsync(context.Background(), events.ConversationAssignedEvent{
					ConversationID: dispatched.ID,
					ToUserID:       dispatched.CurrentAssigneeID,
					OperatorID:     systemDispatchPrincipal().UserID,
					Reason:         "自动分配",
					AssignType:     events.ConversationAssignTypeAutoAssign,
				})
			}
			return &HandoffDecisionResult{
				Decision:   HandoffDecisionAssigned,
				TeamID:     dispatched.CurrentTeamID,
				AssigneeID: dispatched.CurrentAssigneeID,
				Message:    HandoffWaitingMessage,
			}, nil
		}
	}

	teamID := activeTeamIDs[0]
	if err := s.moveToTeamPool(conversationID, teamID, reason); err != nil {
		return nil, err
	}
	return &HandoffDecisionResult{Decision: HandoffDecisionTeamPool, TeamID: teamID, Message: HandoffWaitingMessage}, nil
}

func (s *conversationHumanDispatchService) markHandoff(conversationID int64, aiAgent models.AIAgent, reason string) error {
	now := time.Now()
	trimmedReason := strings.TrimSpace(reason)
	return sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		if err := repositories.ConversationRepository.Updates(ctx.Tx, conversationID, map[string]any{
			"handoff_at":          now,
			"handoff_reason":      trimmedReason,
			"status":              enums.IMConversationStatusPending,
			"current_team_id":     0,
			"current_assignee_id": 0,
			"update_user_id":      0,
			"update_user_name":    aiAgent.Name,
			"updated_at":          now,
		}); err != nil {
			return err
		}
		return ConversationEventLogService.CreateEvent(ctx, conversationID, enums.IMEventTypeTransfer, enums.IMSenderTypeAI, aiAgent.ID, "AI转人工", trimmedReason)
	})
}

func (s *conversationHumanDispatchService) moveToTeamPool(conversationID, teamID int64, reason string) error {
	now := time.Now()
	return sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		conversation := repositories.ConversationRepository.Get(ctx.Tx, conversationID)
		if conversation == nil {
			return errorsx.InvalidParam("会话不存在")
		}
		if err := ConversationAssignmentService.FinishActiveAssignments(ctx, conversationID, now); err != nil {
			return err
		}
		if err := repositories.ConversationRepository.Updates(ctx.Tx, conversationID, map[string]any{
			"status":              enums.IMConversationStatusPending,
			"current_team_id":     teamID,
			"current_assignee_id": 0,
			"update_user_id":      0,
			"update_user_name":    "system",
			"updated_at":          now,
		}); err != nil {
			return err
		}
		return ConversationEventLogService.CreateEvent(ctx, conversationID, enums.IMEventTypeTransfer, enums.IMSenderTypeSystem, 0, "会话进入客服组待接入", ConversationService.buildEventPayload(map[string]any{
			"fromStatus":     conversation.Status,
			"toStatus":       enums.IMConversationStatusPending,
			"fromAssigneeId": conversation.CurrentAssigneeID,
			"toAssigneeId":   int64(0),
			"toTeamId":       teamID,
			"reason":         strings.TrimSpace(reason),
			"decision":       string(HandoffDecisionTeamPool),
		}))
	})
}

func (s *conversationHumanDispatchService) moveToGlobalPool(conversationID int64, operatorName string) error {
	now := time.Now()
	return sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		conversation := repositories.ConversationRepository.Get(ctx.Tx, conversationID)
		if conversation == nil {
			return errorsx.InvalidParam("会话不存在")
		}
		if err := repositories.ConversationRepository.Updates(ctx.Tx, conversationID, map[string]any{
			"status":              enums.IMConversationStatusPending,
			"current_team_id":     0,
			"current_assignee_id": 0,
			"update_user_id":      0,
			"update_user_name":    operatorName,
			"updated_at":          now,
		}); err != nil {
			return err
		}
		return ConversationEventLogService.CreateEvent(ctx, conversationID, enums.IMEventTypeTransfer, enums.IMSenderTypeSystem, 0, "会话进入全局待接入", ConversationService.buildEventPayload(map[string]any{
			"fromStatus": conversation.Status,
			"toStatus":   enums.IMConversationStatusPending,
			"decision":   string(HandoffDecisionGlobalPool),
		}))
	})
}

func (s *conversationHumanDispatchService) createEvent(conversationID int64, eventType enums.IMEventType, senderType enums.IMSenderType, senderID int64, content, payload string) error {
	return sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		return ConversationEventLogService.CreateEvent(ctx, conversationID, eventType, senderType, senderID, content, payload)
	})
}

func (s *conversationHumanDispatchService) sendAIText(conversationID, aiAgentID int64, content string) error {
	_, err := MessageService.SendAIServiceNotice(conversationID, aiAgentID, content)
	return err
}

func orderedPositiveIDs(value string) []int64 {
	return uniquePositiveInt64sFromStrings(strings.Split(value, ","))
}

func uniquePositiveInt64sFromStrings(values []string) []int64 {
	seen := make(map[int64]struct{}, len(values))
	ret := make([]int64, 0, len(values))
	for _, value := range values {
		var id int64
		_, _ = fmt.Sscan(strings.TrimSpace(value), &id)
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		ret = append(ret, id)
	}
	return ret
}
