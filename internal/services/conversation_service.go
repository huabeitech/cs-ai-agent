package services

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"cs-agent/internal/models"
	"cs-agent/internal/pkg/constants"
	"cs-agent/internal/pkg/dto"
	"cs-agent/internal/pkg/dto/request"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/pkg/errorsx"
	"cs-agent/internal/pkg/utils"
	"cs-agent/internal/repositories"
	"slices"
	"strings"
	"time"

	"github.com/mlogclub/simple/common/strs"
	"github.com/mlogclub/simple/sqls"
)

var ConversationService = newConversationService()

func newConversationService() *conversationService {
	return &conversationService{}
}

type conversationService struct {
}

func (s *conversationService) Get(id int64) *models.Conversation {
	return repositories.ConversationRepository.Get(sqls.DB(), id)
}

func (s *conversationService) Find(cnd *sqls.Cnd) []models.Conversation {
	return repositories.ConversationRepository.Find(sqls.DB(), cnd)
}

func (s *conversationService) FindOne(cnd *sqls.Cnd) *models.Conversation {
	return repositories.ConversationRepository.FindOne(sqls.DB(), cnd)
}

func (s *conversationService) FindPageByCnd(cnd *sqls.Cnd) (list []models.Conversation, paging *sqls.Paging) {
	return repositories.ConversationRepository.FindPageByCnd(sqls.DB(), cnd)
}

func (s *conversationService) ListConversations(userID int64, filter request.AgentConversationFilter, keyword string, page, limit int) (list []models.Conversation, paging *sqls.Paging, err error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 100
	}
	if limit > 100 {
		limit = 100
	}

	cnd := sqls.NewCnd().
		Eq("current_assignee_id", userID).
		Page(page, limit)

	if strs.IsNotBlank(keyword) {
		keyword = strings.TrimSpace(keyword)
		cnd.Where("subject LIKE ? OR external_user_id LIKE ? OR last_message_summary LIKE ?", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	switch filter {
	case request.AgentConversationFilterMine:
		cnd.Desc("last_active_at").Desc("id")
	case request.AgentConversationFilterActive:
		cnd.Eq("status", enums.IMConversationStatusActive).Desc("last_active_at").Desc("id")
	case request.AgentConversationFilterPending:
		cnd.Eq("status", enums.IMConversationStatusPending).Asc("last_active_at").Desc("id")
	case request.AgentConversationFilterClosed:
		cnd.Eq("status", enums.IMConversationStatusClosed).Desc("last_active_at").Desc("id")
	default:
		return nil, nil, errorsx.InvalidParam("会话筛选项不合法")
	}

	list, paging = repositories.ConversationRepository.FindPageByCnd(sqls.DB(), cnd)
	return list, paging, nil
}

func (s *conversationService) Updates(id int64, columns map[string]interface{}) error {
	return repositories.ConversationRepository.Updates(sqls.DB(), id, columns)
}

func (s *conversationService) getLatestNotFinished(channelType enums.IMConversationChannel, user *dto.AuthPrincipal) *models.Conversation {
	cnd := sqls.NewCnd()
	if user.IsVisitor {
		cnd.Eq("external_user_id", user.VisitorID)
	} else {
		cnd.Eq("source_user_id", user.UserID)
	}
	cnd.Eq("channel_type", channelType)
	cnd.In("status", []enums.IMConversationStatus{
		enums.IMConversationStatusPending,
		enums.IMConversationStatusActive,
	})
	cnd.Desc("id")
	return s.FindOne(cnd)
}

func (s *conversationService) Create(channelType enums.IMConversationChannel, subject string, aiAgentID int64, operator *dto.AuthPrincipal) (*models.Conversation, error) {
	if operator == nil {
		return nil, errorsx.Unauthorized("未登录或登录已过期")
	}
	if strs.IsBlank(string(channelType)) {
		channelType = enums.IMConversationChannelWebChat
	}
	subject = strs.DefaultIfBlank(strings.TrimSpace(subject), s.buildDefaultSubject(operator))

	// 会话存在，直接返回
	if conversation := s.getLatestNotFinished(channelType, operator); conversation != nil {
		return conversation, nil
	}

	aiAgent := AIAgentService.Get(aiAgentID)
	if aiAgent == nil || aiAgent.Status != enums.StatusOk {
		return nil, errorsx.InvalidParam("AI Agent not found")
	}

	now := time.Now()
	auditFields := utils.BuildAuditFields(operator)
	auditFields.CreatedAt = now
	auditFields.UpdatedAt = now
	conversation := &models.Conversation{
		AIAgentID:         aiAgentID,
		ChannelType:       channelType,
		Subject:           subject,
		Status:            enums.IMConversationStatusPending,
		ServiceMode:       aiAgent.ServiceMode,
		Priority:          0,
		SourceUserID:      operator.UserID,
		ExternalUserID:    "",
		CurrentAssigneeID: 0,
		CurrentTeamID:     0,
		LastMessageAt:     now,
		LastActiveAt:      now,
		AuditFields:       auditFields,
	}
	if operator.IsVisitor {
		conversation.SourceUserID = 0
		conversation.ExternalUserID = operator.VisitorID
	}
	now = conversation.CreatedAt
	if err := sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		if err := ctx.Tx.Create(conversation).Error; err != nil {
			return err
		}
		if err := ConversationParticipantService.EnsureCustomerParticipantTx(ctx, conversation.ID, operator); err != nil {
			return err
		}
		return ConversationEventLogService.CreateEvent(ctx, conversation.ID, enums.IMEventTypeCreate, enums.IMSenderTypeCustomer, operator.UserID, "用户创建会话", "", time.Now())
	}); err != nil {
		return nil, err
	}

	// 推送会话创建事件
	WsService.PublishConversationChanged(conversation, enums.IMRealtimeEventConversationCreated)

	// AI Agent仅人工模式，且有值班客服，尝试自动分配会话
	if conversation.Status == enums.IMConversationStatusPending &&
		aiAgent.ServiceMode == enums.IMConversationServiceModeHumanOnly &&
		len(utils.SplitInt64s(aiAgent.TeamIDs)) > 0 {
		if dispatched, err := ConversationDispatchService.DispatchPendingConversation(conversation, aiAgent); err != nil {
			return nil, err
		} else if dispatched != nil {
			return dispatched, nil
		}
	}
	return s.Get(conversation.ID), nil
}

func (s *conversationService) AssignConversation(conversationID, assigneeID int64, reason string, operator *dto.AuthPrincipal) error {
	if operator == nil {
		return errorsx.Unauthorized("未登录或登录已过期")
	}
	if !s.isAdmin(operator) {
		return errorsx.Forbidden("只有管理员可以分配会话")
	}
	if assigneeID <= 0 {
		return errorsx.InvalidParam("目标客服不能为空")
	}
	targetUser := UserService.Get(assigneeID)
	if targetUser == nil || targetUser.Status != enums.StatusOk {
		return errorsx.InvalidParam("目标客服不存在")
	}
	if err := sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		conversation := repositories.ConversationRepository.Get(ctx.Tx, conversationID)
		if conversation == nil {
			return errorsx.InvalidParam("会话不存在")
		}
		if conversation.Status != enums.IMConversationStatusPending {
			return errorsx.InvalidParam("只有待接入会话允许分配")
		}
		now := time.Now()
		if err := ConversationAssignmentService.FinishActiveAssignmentsTx(ctx, conversationID, now); err != nil {
			return err
		}
		if err := ConversationAssignmentService.CreateAssignmentTx(ctx, conversationID, conversation.CurrentAssigneeID, assigneeID, enums.IMAssignmentTypeAssign, reason, operator, now); err != nil {
			return err
		}
		if err := ctx.Tx.Model(&models.Conversation{}).Where("id = ?", conversationID).Updates(map[string]any{
			"current_assignee_id": assigneeID,
			"status":              enums.IMConversationStatusActive,
			"update_user_id":      operator.UserID,
			"update_user_name":    operator.Username,
			"updated_at":          now,
		}).Error; err != nil {
			return err
		}
		return ConversationEventLogService.CreateEvent(ctx, conversationID, enums.IMEventTypeAssign, enums.IMSenderTypeAgent, operator.UserID, "会话已分配", s.buildEventPayload(map[string]any{
			"fromStatus":     conversation.Status,
			"toStatus":       enums.IMConversationStatusActive,
			"fromAssigneeId": conversation.CurrentAssigneeID,
			"toAssigneeId":   assigneeID,
			"reason":         strings.TrimSpace(reason),
		}), now)
	}); err != nil {
		return err
	}
	if conversation := s.Get(conversationID); conversation != nil {
		WsService.PublishConversationChanged(conversation, enums.IMRealtimeEventConversationAssigned)
	}
	return nil
}

func (s *conversationService) DispatchConversation(conversationID int64, operator *dto.AuthPrincipal) error {
	if operator == nil {
		return errorsx.Unauthorized("未登录或登录已过期")
	}
	if !s.isAdmin(operator) {
		return errorsx.Forbidden("只有管理员可以自动分配会话")
	}

	conversation := s.Get(conversationID)
	if conversation == nil {
		return errorsx.InvalidParam("会话不存在")
	}
	if conversation.Status != enums.IMConversationStatusPending {
		return errorsx.InvalidParam("只有待接入会话允许自动分配")
	}
	if conversation.CurrentAssigneeID > 0 {
		return errorsx.InvalidParam("当前会话已分配客服")
	}

	dispatched, err := ConversationDispatchService.DispatchConversation(conversationID)
	if err != nil {
		return err
	}
	if dispatched == nil {
		return errorsx.InvalidParam("当前暂无可自动分配的值班客服")
	}
	return nil
}

func (s *conversationService) TransferConversation(conversationID, toUserID int64, reason string, operator *dto.AuthPrincipal) error {
	if operator == nil {
		return errorsx.Unauthorized("未登录或登录已过期")
	}
	if !s.isAdmin(operator) {
		return errorsx.Forbidden("只有管理员可以转接会话")
	}
	if toUserID <= 0 {
		return errorsx.InvalidParam("目标客服不能为空")
	}
	targetUser := UserService.Get(toUserID)
	if targetUser == nil || targetUser.Status != enums.StatusOk {
		return errorsx.InvalidParam("目标客服不存在")
	}
	if err := sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		conversation := &models.Conversation{}
		if err := ctx.Tx.First(conversation, "id = ?", conversationID).Error; err != nil {
			return errorsx.InvalidParam("会话不存在")
		}
		if conversation.Status != enums.IMConversationStatusActive {
			return errorsx.InvalidParam("只有处理中会话允许转接")
		}
		if conversation.CurrentAssigneeID <= 0 {
			return errorsx.InvalidParam("当前会话未分配客服")
		}
		if conversation.CurrentAssigneeID == toUserID {
			return errorsx.InvalidParam("目标客服不能与当前指派人相同")
		}
		now := time.Now()
		if err := ConversationAssignmentService.FinishActiveAssignmentsTx(ctx, conversationID, now); err != nil {
			return err
		}
		if err := ConversationAssignmentService.CreateAssignmentTx(ctx, conversationID, conversation.CurrentAssigneeID, toUserID, enums.IMAssignmentTypeTransfer, reason, operator, now); err != nil {
			return err
		}
		if err := ctx.Tx.Model(&models.Conversation{}).Where("id = ?", conversationID).Updates(map[string]any{
			"current_assignee_id": toUserID,
			"status":              enums.IMConversationStatusActive,
			"update_user_id":      operator.UserID,
			"update_user_name":    operator.Username,
			"updated_at":          now,
		}).Error; err != nil {
			return err
		}
		return ConversationEventLogService.CreateEvent(ctx, conversationID, enums.IMEventTypeTransfer, enums.IMSenderTypeAgent, operator.UserID, "会话已转接", s.buildEventPayload(map[string]any{
			"fromStatus":     conversation.Status,
			"toStatus":       enums.IMConversationStatusActive,
			"fromAssigneeId": conversation.CurrentAssigneeID,
			"toAssigneeId":   toUserID,
			"reason":         strings.TrimSpace(reason),
		}), now)
	}); err != nil {
		return err
	}
	if conversation := s.Get(conversationID); conversation != nil {
		WsService.PublishConversationChanged(conversation, enums.IMRealtimeEventConversationTransferred)
	}
	return nil
}

func (s *conversationService) CloseConversation(conversationID int64, closeReason string, operator *dto.AuthPrincipal) error {
	if operator == nil {
		return errorsx.Unauthorized("未登录或登录已过期")
	}
	return s.closeConversation(conversationID, enums.IMSenderTypeAgent, closeReason, operator)
}

func (s *conversationService) CloseCustomerConversation(conversationID int64, operator *dto.AuthPrincipal) error {
	if operator == nil {
		return errorsx.Unauthorized("未登录或登录已过期")
	}
	conversation := s.Get(conversationID)
	if conversation == nil {
		return errorsx.InvalidParam("会话不存在")
	}
	if !s.IsCustomerConversationOwner(conversation, operator) {
		// return errorsx.Forbidden("无权访问该会话")
		return nil
	}
	return s.closeConversation(conversationID, enums.IMSenderTypeCustomer, "", operator)
}

func (s *conversationService) closeConversation(conversationID int64, senderType enums.IMSenderType, closeReason string, operator *dto.AuthPrincipal) error {
	if err := sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		conversation := &models.Conversation{}
		if err := ctx.Tx.First(conversation, "id = ?", conversationID).Error; err != nil {
			return errorsx.InvalidParam("会话不存在")
		}
		if conversation.Status == enums.IMConversationStatusClosed {
			return nil
		}
		if conversation.Status != enums.IMConversationStatusPending && conversation.Status != enums.IMConversationStatusActive {
			return errorsx.InvalidParam("当前状态不允许关闭会话")
		}
		now := time.Now()
		updateUserName := operator.Username
		eventDesc := "会话已关闭"
		closeReason = strings.TrimSpace(closeReason)
		if senderType == enums.IMSenderTypeCustomer {
			if operator.IsVisitor {
				updateUserName = operator.Nickname
				eventDesc = "访客关闭会话"
			} else {
				eventDesc = "客户关闭会话"
			}
		} else {
			if closeReason == "" {
				return errorsx.InvalidParam("关闭原因不能为空")
			}
			if !s.canCloseConversation(conversation, operator) {
				return errorsx.Forbidden("无权关闭该会话")
			}
		}
		if err := ConversationAssignmentService.FinishActiveAssignmentsTx(ctx, conversationID, now); err != nil {
			return err
		}
		if err := ctx.Tx.Model(&models.Conversation{}).Where("id = ?", conversationID).Updates(map[string]any{
			"status":           enums.IMConversationStatusClosed,
			"closed_at":        now,
			"closed_by":        operator.UserID,
			"close_reason":     closeReason,
			"update_user_id":   operator.UserID,
			"update_user_name": updateUserName,
			"updated_at":       now,
		}).Error; err != nil {
			return err
		}
		return ConversationEventLogService.CreateEvent(ctx, conversationID, enums.IMEventTypeClose, senderType, operator.UserID, eventDesc, s.buildEventPayload(map[string]any{
			"fromStatus":     conversation.Status,
			"toStatus":       enums.IMConversationStatusClosed,
			"fromAssigneeId": conversation.CurrentAssigneeID,
			"toAssigneeId":   conversation.CurrentAssigneeID,
			"closeReason":    closeReason,
		}), now)
	}); err != nil {
		return err
	}
	if conversation := s.Get(conversationID); conversation != nil {
		WsService.PublishConversationChanged(conversation, enums.IMRealtimeEventConversationClosed)
	}
	return nil
}

func (s *conversationService) MarkAgentRead(conversationID int64, operator *dto.AuthPrincipal) error {
	if operator == nil {
		return errorsx.Unauthorized("未登录或登录已过期")
	}
	conversation := s.Get(conversationID)
	if conversation == nil {
		return errorsx.InvalidParam("会话不存在")
	}
	changed, err := s.markConversationRead(conversation, enums.IMSenderTypeAgent, operator, 0)
	if err != nil {
		return err
	}
	if changed {
		if updated := s.Get(conversationID); updated != nil {
			WsService.PublishConversationChanged(updated, enums.IMRealtimeEventConversationRead)
		}
	}
	return nil
}

func (s *conversationService) MarkCustomerRead(conversationID int64, operator *dto.AuthPrincipal) error {
	if operator == nil {
		return errorsx.Unauthorized("未登录或登录已过期")
	}
	conversation := s.Get(conversationID)
	if conversation == nil {
		return errorsx.InvalidParam("会话不存在")
	}
	if !s.IsCustomerConversationOwner(conversation, operator) {
		return errorsx.Forbidden("无权访问该会话")
	}
	changed, err := s.markConversationRead(conversation, enums.IMSenderTypeCustomer, operator, 0)
	if err != nil {
		return err
	}
	if changed {
		if updated := s.Get(conversationID); updated != nil {
			WsService.PublishConversationChanged(updated, enums.IMRealtimeEventConversationRead)
		}
	}
	return nil
}

func (s *conversationService) MarkConversationRead(conversationID int64, operatorType string, operator *dto.AuthPrincipal) error {
	switch strings.TrimSpace(operatorType) {
	case string(enums.IMSenderTypeAgent):
		return s.MarkAgentRead(conversationID, operator)
	case string(enums.IMSenderTypeCustomer):
		return s.MarkCustomerRead(conversationID, operator)
	default:
		return errorsx.InvalidParam("不支持的已读操作类型")
	}
}

func (s *conversationService) MarkConversationReadToMessage(conversationID, messageID int64, operatorType enums.IMSenderType, operator *dto.AuthPrincipal) error {
	conversation := s.Get(conversationID)
	if conversation == nil {
		return errorsx.InvalidParam("会话不存在")
	}
	changed := false
	var err error
	switch operatorType {
	case enums.IMSenderTypeAgent:
		if operator == nil {
			return errorsx.Unauthorized("未登录或登录已过期")
		}
		changed, err = s.markConversationRead(conversation, enums.IMSenderTypeAgent, operator, messageID)
		if err != nil {
			return err
		}
	case enums.IMSenderTypeCustomer:
		if operator == nil {
			return errorsx.Unauthorized("未登录或登录已过期")
		}
		if !s.IsCustomerConversationOwner(conversation, operator) {
			return errorsx.Forbidden("无权访问该会话")
		}
		changed, err = s.markConversationRead(conversation, enums.IMSenderTypeCustomer, operator, messageID)
		if err != nil {
			return err
		}
	default:
		return errorsx.InvalidParam("不支持的已读操作类型")
	}
	if changed {
		if updated := s.Get(conversationID); updated != nil {
			WsService.PublishConversationChanged(updated, enums.IMRealtimeEventConversationRead)
		}
	}
	return nil
}

func (s *conversationService) markConversationRead(conversation *models.Conversation, readerType enums.IMSenderType, operator *dto.AuthPrincipal, messageID int64) (bool, error) {
	if conversation == nil {
		return false, errorsx.InvalidParam("会话不存在")
	}
	targetMessage, err := MessageService.GetConversationReadTarget(conversation.ID, messageID)
	if err != nil {
		return false, err
	}
	if targetMessage == nil {
		if readerType == enums.IMSenderTypeAgent && conversation.AgentUnreadCount == 0 {
			return false, nil
		}
		if readerType == enums.IMSenderTypeCustomer && conversation.CustomerUnreadCount == 0 {
			return false, nil
		}
		now := time.Now()
		updates := map[string]any{
			"update_user_id":   operator.UserID,
			"update_user_name": operator.Username,
			"updated_at":       now,
		}
		if readerType == enums.IMSenderTypeAgent {
			updates["agent_unread_count"] = 0
		} else {
			updates["customer_unread_count"] = 0
		}
		return true, s.Updates(conversation.ID, updates)
	}

	currentReadState := ConversationReadStateService.GetByReader(conversation.ID, string(readerType), operator)
	if currentReadState != nil && currentReadState.LastReadSeqNo >= targetMessage.SeqNo {
		if readerType == enums.IMSenderTypeAgent && conversation.AgentUnreadCount == 0 {
			return false, nil
		}
		if readerType == enums.IMSenderTypeCustomer && conversation.CustomerUnreadCount == 0 {
			return false, nil
		}
	}

	err = sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		currentConversation := &models.Conversation{}
		if err := ctx.Tx.First(currentConversation, "id = ?", conversation.ID).Error; err != nil {
			return errorsx.InvalidParam("会话不存在")
		}
		now := time.Now()
		if _, err := ConversationReadStateService.MarkReadTx(ctx, currentConversation, readerType, operator, targetMessage, now); err != nil {
			return err
		}
		agentReadState, customerReadState, err := ConversationReadStateService.GetConversationReadStatesTx(ctx, currentConversation.ID)
		if err != nil {
			return err
		}
		agentUnreadCount, err := s.countUnreadByState(ctx, currentConversation.ID, agentReadState, enums.IMSenderTypeCustomer)
		if err != nil {
			return err
		}
		customerUnreadCount, err := s.countUnreadByState(ctx, currentConversation.ID, customerReadState, enums.IMSenderTypeAgent, enums.IMSenderTypeAI)
		if err != nil {
			return err
		}
		if readerType == enums.IMSenderTypeAgent && currentConversation.AgentUnreadCount == agentUnreadCount && currentReadState != nil && currentReadState.LastReadSeqNo >= targetMessage.SeqNo {
			return nil
		}
		if readerType == enums.IMSenderTypeCustomer && currentConversation.CustomerUnreadCount == customerUnreadCount && currentReadState != nil && currentReadState.LastReadSeqNo >= targetMessage.SeqNo {
			return nil
		}
		updateUserName := operator.Username
		if readerType == enums.IMSenderTypeCustomer && operator.IsVisitor {
			updateUserName = operator.Nickname
		}
		return ctx.Tx.Model(&models.Conversation{}).Where("id = ?", currentConversation.ID).Updates(map[string]any{
			"agent_unread_count":    agentUnreadCount,
			"customer_unread_count": customerUnreadCount,
			"update_user_id":        operator.UserID,
			"update_user_name":      updateUserName,
			"updated_at":            now,
		}).Error
	})
	if err != nil {
		return false, err
	}
	return true, nil
}

func (s *conversationService) countUnreadByState(ctx *sqls.TxContext, conversationID int64, state *models.ConversationReadState, senderTypes ...enums.IMSenderType) (int, error) {
	lastReadSeqNo := int64(0)
	if state != nil {
		lastReadSeqNo = state.LastReadSeqNo
	}
	normalizedSenderTypes := make([]enums.IMSenderType, 0, len(senderTypes))
	for _, senderType := range senderTypes {
		normalizedSenderTypes = append(normalizedSenderTypes, senderType)
	}
	count, err := ConversationReadStateService.CountUnreadMessagesTx(ctx, conversationID, lastReadSeqNo, normalizedSenderTypes...)
	return int(count), err
}

func (s *conversationService) IsCustomerConversationOwner(conversation *models.Conversation, operator *dto.AuthPrincipal) bool {
	if conversation == nil || operator == nil {
		return false
	}
	if operator.IsVisitor {
		return strings.TrimSpace(conversation.ExternalUserID) != "" && conversation.ExternalUserID == operator.VisitorID
	}
	return conversation.SourceUserID > 0 && conversation.SourceUserID == operator.UserID
}

func (s *conversationService) BuildConversationSummary(conversation *models.Conversation) string {
	if conversation == nil {
		return ""
	}
	if strings.TrimSpace(conversation.LastMessageSummary) != "" {
		return conversation.LastMessageSummary
	}
	return strings.TrimSpace(conversation.Subject)
}

func (s *conversationService) canCloseConversation(conversation *models.Conversation, operator *dto.AuthPrincipal) bool {
	if conversation == nil || operator == nil {
		return false
	}
	if s.isAdmin(operator) {
		return true
	}
	return conversation.Status == enums.IMConversationStatusActive && conversation.CurrentAssigneeID > 0 && conversation.CurrentAssigneeID == operator.UserID
}

func (s *conversationService) isAdmin(operator *dto.AuthPrincipal) bool {
	if operator == nil {
		return false
	}
	return slices.Contains(operator.Roles, constants.RoleCodeSuperAdmin) || slices.Contains(operator.Roles, constants.RoleCodeAdmin)
}

func (s *conversationService) buildEventPayload(payload map[string]any) string {
	if len(payload) == 0 {
		return ""
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return ""
	}
	return string(data)
}

func (s *conversationService) buildDefaultSubject(operator *dto.AuthPrincipal) string {
	if operator == nil {
		return "访客unknown"
	}

	if operator.IsVisitor {
		visitorID := strings.TrimSpace(operator.VisitorID)
		if visitorID == "" {
			return "访客unknown"
		}
		return fmt.Sprintf("访客%s", hashUUID(visitorID))
	}

	if operator.Username != "" {
		return operator.Username
	}

	return fmt.Sprintf("用户%d", operator.UserID)
}

func hashUUID(uuid string) string {
	if uuid == "" {
		return "unknown"
	}

	h := md5.Sum([]byte(uuid))
	return hex.EncodeToString(h[:])[:8]
}
