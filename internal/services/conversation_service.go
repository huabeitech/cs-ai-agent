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
	"cs-agent/internal/pkg/openidentity"
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
	if id <= 0 {
		return nil
	}
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

func (s *conversationService) ListConversations(userID int64, filter request.AgentConversationFilter, keyword string, paging *sqls.Paging) ([]models.Conversation, *sqls.Paging, error) {
	cnd := sqls.NewCnd().Page(paging.Page, paging.Limit)

	if strs.IsNotBlank(keyword) {
		keyword = strings.TrimSpace(keyword)
		cnd.Where("subject LIKE ? OR external_id LIKE ? OR last_message_summary LIKE ?", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	switch filter {
	case request.AgentConversationFilterMine:
		cnd.Eq("current_assignee_id", userID).Desc("last_active_at").Desc("id")
	case request.AgentConversationFilterActive:
		cnd.Eq("current_assignee_id", userID).Eq("status", enums.IMConversationStatusActive).Desc("last_active_at").Desc("id")
	case request.AgentConversationFilterPending:
		cnd.Eq("current_assignee_id", 0).Eq("status", enums.IMConversationStatusPending).Asc("last_active_at").Desc("id")
	case request.AgentConversationFilterClosed:
		cnd.Eq("current_assignee_id", userID).Eq("status", enums.IMConversationStatusClosed).Desc("last_active_at").Desc("id")
	default:
		return nil, nil, errorsx.InvalidParam("会话筛选项不合法")
	}

	list, paging := repositories.ConversationRepository.FindPageByCnd(sqls.DB(), cnd)
	return list, paging, nil
}

func (s *conversationService) Updates(id int64, columns map[string]interface{}) error {
	return repositories.ConversationRepository.Updates(sqls.DB(), id, columns)
}

func (s *conversationService) getLatestNotFinished(externalInfo openidentity.ExternalInfo) *models.Conversation {
	cnd := sqls.NewCnd()
	cnd.Eq("external_id", externalInfo.ExternalID)
	cnd.Eq("external_source", externalInfo.ExternalSource)
	cnd.In("status", []enums.IMConversationStatus{
		enums.IMConversationStatusPending,
		enums.IMConversationStatusActive,
	})
	cnd.Desc("id")
	return s.FindOne(cnd)
}

func (s *conversationService) Create(externalInfo openidentity.ExternalInfo, aiAgentID int64) (*models.Conversation, error) {
	subject := s.buildDefaultSubject(externalInfo)

	// 会话存在，直接返回
	if conversation := s.getLatestNotFinished(externalInfo); conversation != nil {
		return conversation, nil
	}

	aiAgent := AIAgentService.Get(aiAgentID)
	if aiAgent == nil || aiAgent.Status != enums.StatusOk {
		return nil, errorsx.InvalidParam("AI Agent not found")
	}

	conversation := &models.Conversation{
		AIAgentID:         aiAgentID,
		ExternalSource:    externalInfo.ExternalSource,
		Subject:           subject,
		Status:            enums.IMConversationStatusPending,
		ServiceMode:       aiAgent.ServiceMode,
		Priority:          0,
		ExternalID:        externalInfo.ExternalID,
		CurrentAssigneeID: 0,
		CurrentTeamID:     0,
		LastMessageAt:     time.Now(),
		LastActiveAt:      time.Now(),
		AuditFields:       utils.BuildAuditFields(nil),
	}
	if identity := repositories.CustomerIdentityRepository.GetBy(sqls.DB(), externalInfo.ExternalSource, externalInfo.ExternalID); identity != nil {
		conversation.CustomerID = identity.CustomerID
	}
	if err := sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		if err := ctx.Tx.Create(conversation).Error; err != nil {
			return err
		}
		if err := ConversationParticipantService.CreateCustomerParticipant(ctx, conversation.ID, externalInfo); err != nil {
			return err
		}
		return ConversationEventLogService.CreateEvent(ctx, conversation.ID, enums.IMEventTypeCreate, enums.IMSenderTypeCustomer, 0, "用户创建会话", "")
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

func (s *conversationService) AssignConversation(req request.AssignConversationRequest, operator *dto.AuthPrincipal) error {
	if operator == nil {
		return errorsx.Unauthorized("未登录或登录已过期")
	}
	targetProfile := AgentProfileService.GetByUserID(req.AssigneeID)
	if targetProfile == nil || targetProfile.Status != enums.StatusOk {
		return errorsx.InvalidParam("目标客服不存在")
	}
	if err := sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		conversation := repositories.ConversationRepository.Get(ctx.Tx, req.ConversationID)
		if conversation == nil {
			return errorsx.InvalidParam("会话不存在")
		}
		if conversation.Status != enums.IMConversationStatusPending {
			return errorsx.InvalidParam("只有待接入会话允许分配")
		}
		now := time.Now()
		if err := ConversationAssignmentService.FinishActiveAssignments(ctx, req.ConversationID, now); err != nil {
			return err
		}
		if err := ConversationAssignmentService.CreateAssignment(ctx, req.ConversationID, conversation.CurrentAssigneeID, req.AssigneeID, enums.IMAssignmentTypeAssign, req.Reason, operator, now); err != nil {
			return err
		}
		if err := repositories.ConversationRepository.Updates(ctx.Tx, req.ConversationID, map[string]any{
			"current_assignee_id": req.AssigneeID,
			"status":              enums.IMConversationStatusActive,
			"update_user_id":      operator.UserID,
			"update_user_name":    operator.Username,
			"updated_at":          now,
		}); err != nil {
			return err
		}
		return ConversationEventLogService.CreateEvent(ctx, req.ConversationID, enums.IMEventTypeAssign, enums.IMSenderTypeAgent, operator.UserID, "会话已分配", s.buildEventPayload(map[string]any{
			"fromStatus":     conversation.Status,
			"toStatus":       enums.IMConversationStatusActive,
			"fromAssigneeId": conversation.CurrentAssigneeID,
			"toAssigneeId":   req.AssigneeID,
			"reason":         strings.TrimSpace(req.Reason),
		}))
	}); err != nil {
		return err
	}
	if conversation := s.Get(req.ConversationID); conversation != nil {
		WsService.PublishConversationChanged(conversation, enums.IMRealtimeEventConversationAssigned)
	}
	return nil
}

func (s *conversationService) AutoAssignConversation(conversationID int64, operator *dto.AuthPrincipal) error {
	if operator == nil {
		return errorsx.Unauthorized("未登录或登录已过期")
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
	if toUserID <= 0 {
		return errorsx.InvalidParam("目标客服不能为空")
	}
	targetProfile := AgentProfileService.GetByUserID(toUserID)
	if targetProfile == nil || targetProfile.Status != enums.StatusOk {
		return errorsx.InvalidParam("目标客服不存在")
	}
	if err := sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		conversation := repositories.ConversationRepository.Get(ctx.Tx, conversationID)
		if conversation == nil {
			return errorsx.InvalidParam("会话不存在")
		}
		if !s.canTransferConversation(conversation, operator) {
			return errorsx.Forbidden("无权转接该会话")
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
		if err := ConversationAssignmentService.FinishActiveAssignments(ctx, conversationID, now); err != nil {
			return err
		}
		if err := ConversationAssignmentService.CreateAssignment(ctx, conversationID, conversation.CurrentAssigneeID, toUserID, enums.IMAssignmentTypeTransfer, reason, operator, now); err != nil {
			return err
		}
		if err := repositories.ConversationRepository.Updates(ctx.Tx, conversationID, map[string]any{
			"current_assignee_id": toUserID,
			"status":              enums.IMConversationStatusActive,
			"update_user_id":      operator.UserID,
			"update_user_name":    operator.Username,
			"updated_at":          now,
		}); err != nil {
			return err
		}
		return ConversationEventLogService.CreateEvent(ctx, conversationID, enums.IMEventTypeTransfer, enums.IMSenderTypeAgent, operator.UserID, "会话已转接", s.buildEventPayload(map[string]any{
			"fromStatus":     conversation.Status,
			"toStatus":       enums.IMConversationStatusActive,
			"fromAssigneeId": conversation.CurrentAssigneeID,
			"toAssigneeId":   toUserID,
			"reason":         strings.TrimSpace(reason),
		}))
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

func (s *conversationService) CloseCustomerConversation(conversationID int64, externalInfo openidentity.ExternalInfo) error {
	conversation := s.Get(conversationID)
	if conversation == nil {
		return errorsx.InvalidParam("会话不存在")
	}
	if !s.IsCustomerConversationOwner(conversation, externalInfo) {
		return errorsx.Forbidden("无权访问该会话")
	}
	return s.closeConversation(conversationID, enums.IMSenderTypeCustomer, "", nil)
}

func (s *conversationService) closeConversation(conversationID int64, senderType enums.IMSenderType, closeReason string, operator *dto.AuthPrincipal) error {
	if err := sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		conversation := repositories.ConversationRepository.Get(ctx.Tx, conversationID)
		if conversation == nil {
			return errorsx.InvalidParam("会话不存在")
		}
		if conversation.Status == enums.IMConversationStatusClosed {
			return nil
		}
		if conversation.Status != enums.IMConversationStatusPending && conversation.Status != enums.IMConversationStatusActive {
			return errorsx.InvalidParam("当前状态不允许关闭会话")
		}
		var (
			now          = time.Now()
			eventDesc    = "会话已关闭"
			operatorID   int64
			operatorName string
		)
		closeReason = strings.TrimSpace(closeReason)
		if senderType == enums.IMSenderTypeCustomer {
			eventDesc = "客户关闭会话"
		} else {
			if operator == nil {
				return errorsx.InvalidParam("无权限操作")
			}
			if closeReason == "" {
				return errorsx.InvalidParam("关闭原因不能为空")
			}
			if !s.canCloseConversation(conversation, operator) {
				return errorsx.Forbidden("无权关闭该会话")
			}
			operatorID = operator.UserID
			operatorName = operator.Nickname
		}
		if err := ConversationAssignmentService.FinishActiveAssignments(ctx, conversationID, now); err != nil {
			return err
		}
		if err := repositories.ConversationRepository.Updates(ctx.Tx, conversationID, map[string]any{
			"status":           enums.IMConversationStatusClosed,
			"closed_at":        now,
			"closed_by":        operatorID,
			"close_reason":     closeReason,
			"update_user_id":   operatorID,
			"update_user_name": operatorName,
			"updated_at":       now,
		}); err != nil {
			return err
		}
		return ConversationEventLogService.CreateEvent(ctx, conversationID, enums.IMEventTypeClose, senderType, operatorID, eventDesc, s.buildEventPayload(map[string]any{
			"fromStatus":     conversation.Status,
			"toStatus":       enums.IMConversationStatusClosed,
			"fromAssigneeId": conversation.CurrentAssigneeID,
			"toAssigneeId":   conversation.CurrentAssigneeID,
			"closeReason":    closeReason,
		}))
	}); err != nil {
		return err
	}
	if conversation := s.Get(conversationID); conversation != nil {
		WsService.PublishConversationChanged(conversation, enums.IMRealtimeEventConversationClosed)
	}
	return nil
}

// MarkAgentConversationReadToMessage 控制台客服将会话已读推进到指定消息。
func (s *conversationService) MarkAgentConversationReadToMessage(conversationID, messageID int64, operator *dto.AuthPrincipal) error {
	if operator == nil {
		return errorsx.Unauthorized("未登录或登录已过期")
	}
	conversation := s.Get(conversationID)
	if conversation == nil {
		return errorsx.InvalidParam("会话不存在")
	}
	changed, err := s.markConversationReadWithActor(conversation, messageID, agentConversationReadActor{operator: operator})
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

// MarkCustomerConversationReadToMessage IM 客户将会话已读推进到指定消息（需为会话归属外部身份）。
func (s *conversationService) MarkCustomerConversationReadToMessage(conversationID, messageID int64, external *openidentity.ExternalInfo) error {
	if external == nil || strings.TrimSpace(external.ExternalID) == "" {
		return errorsx.Unauthorized("外部用户标识不能为空")
	}
	conversation := s.Get(conversationID)
	if conversation == nil {
		return errorsx.InvalidParam("会话不存在")
	}
	if !s.IsCustomerConversationOwner(conversation, *external) {
		return errorsx.Forbidden("无权访问该会话")
	}
	changed, err := s.markConversationReadWithActor(conversation, messageID, customerConversationReadActor{external: external})
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

func displayExternalName(ext *openidentity.ExternalInfo) string {
	if ext == nil {
		return ""
	}
	if n := strings.TrimSpace(ext.ExternalName); n != "" {
		return n
	}
	return strings.TrimSpace(ext.ExternalID)
}

// conversationReadActor 抽象「读者身份」，供 markConversationReadWithActor 共用（包内私有）。
type conversationReadActor interface {
	isAgentSide() bool
	getReadState(conversationID int64) *models.ConversationReadState
	markReadTx(ctx *sqls.TxContext, conversation *models.Conversation, targetMessage *models.Message, now time.Time) error
	conversationUpdateAudit() (userID int64, userName string)
}

type agentConversationReadActor struct {
	operator *dto.AuthPrincipal
}

func (a agentConversationReadActor) isAgentSide() bool { return true }

func (a agentConversationReadActor) getReadState(conversationID int64) *models.ConversationReadState {
	return ConversationReadStateService.GetByAgentReader(conversationID, a.operator)
}

func (a agentConversationReadActor) markReadTx(ctx *sqls.TxContext, conversation *models.Conversation, targetMessage *models.Message, now time.Time) error {
	_, err := ConversationReadStateService.MarkAgentRead(ctx, conversation, a.operator, targetMessage, now)
	return err
}

func (a agentConversationReadActor) conversationUpdateAudit() (int64, string) {
	if a.operator == nil {
		return 0, ""
	}
	return a.operator.UserID, a.operator.Username
}

type customerConversationReadActor struct {
	external *openidentity.ExternalInfo
}

func (a customerConversationReadActor) isAgentSide() bool { return false }

func (a customerConversationReadActor) getReadState(conversationID int64) *models.ConversationReadState {
	return ConversationReadStateService.GetByCustomerReader(conversationID, a.external)
}

func (a customerConversationReadActor) markReadTx(ctx *sqls.TxContext, conversation *models.Conversation, targetMessage *models.Message, now time.Time) error {
	_, err := ConversationReadStateService.MarkCustomerRead(ctx, conversation, a.external, targetMessage, now)
	return err
}

func (a customerConversationReadActor) conversationUpdateAudit() (int64, string) {
	return 0, displayExternalName(a.external)
}

func (s *conversationService) markConversationReadWithActor(conversation *models.Conversation, messageID int64, actor conversationReadActor) (bool, error) {
	if conversation == nil {
		return false, errorsx.InvalidParam("会话不存在")
	}
	targetMessage, err := MessageService.GetConversationReadTarget(conversation.ID, messageID)
	if err != nil {
		return false, err
	}
	if targetMessage == nil {
		if actor.isAgentSide() && conversation.AgentUnreadCount == 0 {
			return false, nil
		}
		if !actor.isAgentSide() && conversation.CustomerUnreadCount == 0 {
			return false, nil
		}
		now := time.Now()
		updateUserID, updateUserName := actor.conversationUpdateAudit()
		updates := map[string]any{
			"update_user_id":   updateUserID,
			"update_user_name": updateUserName,
			"updated_at":       now,
		}
		if actor.isAgentSide() {
			updates["agent_unread_count"] = 0
		} else {
			updates["customer_unread_count"] = 0
		}
		return true, s.Updates(conversation.ID, updates)
	}

	currentReadState := actor.getReadState(conversation.ID)
	if currentReadState != nil && currentReadState.LastReadSeqNo >= targetMessage.SeqNo {
		if actor.isAgentSide() && conversation.AgentUnreadCount == 0 {
			return false, nil
		}
		if !actor.isAgentSide() && conversation.CustomerUnreadCount == 0 {
			return false, nil
		}
	}

	err = sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		currentConversation := repositories.ConversationRepository.Get(ctx.Tx, conversation.ID)
		if currentConversation == nil {
			return errorsx.InvalidParam("会话不存在")
		}
		now := time.Now()
		if err := actor.markReadTx(ctx, currentConversation, targetMessage, now); err != nil {
			return err
		}
		agentReadState, customerReadState := ConversationReadStateService.getConversationReadStates(ctx.Tx, currentConversation.ID)
		agentUnreadCount, err := s.countUnreadByState(ctx, currentConversation.ID, agentReadState, enums.IMSenderTypeCustomer)
		if err != nil {
			return err
		}
		customerUnreadCount, err := s.countUnreadByState(ctx, currentConversation.ID, customerReadState, enums.IMSenderTypeAgent, enums.IMSenderTypeAI)
		if err != nil {
			return err
		}
		if actor.isAgentSide() && currentConversation.AgentUnreadCount == agentUnreadCount && currentReadState != nil && currentReadState.LastReadSeqNo >= targetMessage.SeqNo {
			return nil
		}
		if !actor.isAgentSide() && currentConversation.CustomerUnreadCount == customerUnreadCount && currentReadState != nil && currentReadState.LastReadSeqNo >= targetMessage.SeqNo {
			return nil
		}
		updateUserID, updateUserName := actor.conversationUpdateAudit()
		return repositories.ConversationRepository.Updates(ctx.Tx, currentConversation.ID, map[string]any{
			"agent_unread_count":    agentUnreadCount,
			"customer_unread_count": customerUnreadCount,
			"update_user_id":        updateUserID,
			"update_user_name":      updateUserName,
			"updated_at":            now,
		})
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
	count, err := ConversationReadStateService.CountUnreadMessages(ctx, conversationID, lastReadSeqNo, normalizedSenderTypes...)
	return int(count), err
}

func (s *conversationService) IsCustomerConversationOwner(conversation *models.Conversation, externalInfo openidentity.ExternalInfo) bool {
	if conversation == nil {
		return false
	}
	extID := strings.TrimSpace(externalInfo.ExternalID)
	if extID == "" || strings.TrimSpace(conversation.ExternalID) == "" {
		return false
	}
	if conversation.ExternalID != extID {
		return false
	}
	reqSrc := strings.TrimSpace(string(externalInfo.ExternalSource))
	convSrc := strings.TrimSpace(string(conversation.ExternalSource))
	if convSrc != "" {
		if reqSrc == "" || reqSrc != convSrc {
			return false
		}
	}
	return true
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

func (s *conversationService) canTransferConversation(conversation *models.Conversation, operator *dto.AuthPrincipal) bool {
	if conversation == nil || operator == nil {
		return false
	}
	if s.isAdmin(operator) {
		return true
	}
	return conversation.Status == enums.IMConversationStatusActive &&
		conversation.CurrentAssigneeID > 0 &&
		conversation.CurrentAssigneeID == operator.UserID
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

// LinkConversationCustomer 将会话绑定到指定客户；若会话带外部访客标识则维护 CustomerIdentity（与创建会话时逻辑一致）。
func (s *conversationService) LinkConversationCustomer(conversationID, customerID int64, operator *dto.AuthPrincipal) error {
	if operator == nil {
		return errorsx.Unauthorized("未登录或登录已过期")
	}
	if conversationID <= 0 || customerID <= 0 {
		return errorsx.InvalidParam("参数不合法")
	}
	cust := CustomerService.Get(customerID)
	if cust == nil || cust.Status == enums.StatusDeleted {
		return errorsx.InvalidParam("客户不存在")
	}
	conv := s.Get(conversationID)
	if conv == nil {
		return errorsx.InvalidParam("会话不存在")
	}
	if conv.Status == enums.IMConversationStatusClosed {
		return errorsx.InvalidParam("已关闭的会话无法关联客户")
	}
	if !s.canLinkConversationCustomer(conv, operator) {
		return errorsx.Forbidden("无权限关联该会话")
	}

	extID := strings.TrimSpace(conv.ExternalID)
	extSrc := strings.TrimSpace(string(conv.ExternalSource))

	err := sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		current := repositories.ConversationRepository.Get(ctx.Tx, conversationID)
		if current == nil {
			return errorsx.InvalidParam("会话不存在")
		}
		if extID != "" && extSrc != "" {
			existing := repositories.CustomerIdentityRepository.GetBy(ctx.Tx, enums.ExternalSource(extSrc), extID)
			if existing != nil {
				if existing.CustomerID != customerID {
					return errorsx.BusinessError(1, "该访客身份已绑定其他客户，无法关联到当前选择")
				}
			} else {
				idRow := &models.CustomerIdentity{
					CustomerID:     customerID,
					ExternalSource: enums.ExternalSource(extSrc),
					ExternalID:     extID,
					Status:         enums.StatusOk,
					AuditFields:    utils.BuildAuditFields(operator),
				}
				if err := repositories.CustomerIdentityRepository.Create(ctx.Tx, idRow); err != nil {
					return err
				}
			}
		}
		now := time.Now()
		return repositories.ConversationRepository.Updates(ctx.Tx, conversationID, map[string]any{
			"customer_id":      customerID,
			"update_user_id":   operator.UserID,
			"update_user_name": operator.Username,
			"updated_at":       now,
		})
	})
	if err != nil {
		return err
	}
	if updated := s.Get(conversationID); updated != nil {
		WsService.PublishConversationChanged(updated, enums.IMRealtimeEventConversationUpdated)
	}
	return nil
}

func (s *conversationService) canLinkConversationCustomer(conv *models.Conversation, operator *dto.AuthPrincipal) bool {
	if conv == nil || operator == nil {
		return false
	}
	if s.isAdmin(operator) {
		return true
	}
	switch conv.Status {
	case enums.IMConversationStatusPending:
		return true
	case enums.IMConversationStatusActive:
		return conv.CurrentAssigneeID == 0 || conv.CurrentAssigneeID == operator.UserID
	default:
		return false
	}
}

func (s *conversationService) buildDefaultSubject(externalInfo openidentity.ExternalInfo) string {
	if strs.IsNotBlank(externalInfo.ExternalName) {
		return externalInfo.ExternalName
	}
	return fmt.Sprintf("访客%s", hashUUID(externalInfo.ExternalID))
}

func hashUUID(uuid string) string {
	if uuid == "" {
		return "unknown"
	}

	h := md5.Sum([]byte(uuid))
	return hex.EncodeToString(h[:])[:8]
}
