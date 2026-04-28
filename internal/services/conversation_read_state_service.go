package services

import (
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/dto"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/pkg/errorsx"
	"cs-agent/internal/pkg/openidentity"
	"cs-agent/internal/repositories"
	"strings"
	"time"

	"github.com/mlogclub/simple/common/strs"
	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
	"gorm.io/gorm"
)

var ConversationReadStateService = newConversationReadStateService()

func newConversationReadStateService() *conversationReadStateService {
	return &conversationReadStateService{}
}

type conversationReadStateService struct {
}

// readerCursor 已读游标行的身份键（包内私有，供客服 / 客户两条路径共用）。
type readerCursor struct {
	readerType       enums.IMSenderType
	readerID         int64
	externalReaderID string
	auditUserID      int64
	auditUserName    string
}

func agentReaderCursor(operator *dto.AuthPrincipal) (readerCursor, error) {
	if operator == nil {
		return readerCursor{}, errorsx.Unauthorized("未登录或登录已过期")
	}
	return readerCursor{
		readerType:       enums.IMSenderTypeAgent,
		readerID:         operator.UserID,
		externalReaderID: "",
		auditUserID:      operator.UserID,
		auditUserName:    operator.Username,
	}, nil
}

func customerReaderCursor(external *openidentity.ExternalUser) (readerCursor, error) {
	if external == nil || strings.TrimSpace(external.ExternalID) == "" {
		return readerCursor{}, errorsx.Unauthorized("外部用户标识不能为空")
	}
	extID := strings.TrimSpace(external.ExternalID)
	name := strings.TrimSpace(external.ExternalName)
	if name == "" {
		name = extID
	}
	return readerCursor{
		readerType:       enums.IMSenderTypeCustomer,
		readerID:         0,
		externalReaderID: extID,
		auditUserID:      0,
		auditUserName:    name,
	}, nil
}

func (s *conversationReadStateService) Get(id int64) *models.ConversationReadState {
	return repositories.ConversationReadStateRepository.Get(sqls.DB(), id)
}

func (s *conversationReadStateService) Take(where ...any) *models.ConversationReadState {
	return repositories.ConversationReadStateRepository.Take(sqls.DB(), where...)
}

func (s *conversationReadStateService) Find(cnd *sqls.Cnd) []models.ConversationReadState {
	return repositories.ConversationReadStateRepository.Find(sqls.DB(), cnd)
}

func (s *conversationReadStateService) FindOne(cnd *sqls.Cnd) *models.ConversationReadState {
	return repositories.ConversationReadStateRepository.FindOne(sqls.DB(), cnd)
}

func (s *conversationReadStateService) FindPageByParams(queryParams *params.QueryParams) (list []models.ConversationReadState, paging *sqls.Paging) {
	return repositories.ConversationReadStateRepository.FindPageByParams(sqls.DB(), queryParams)
}

func (s *conversationReadStateService) FindPageByCnd(cnd *sqls.Cnd) (list []models.ConversationReadState, paging *sqls.Paging) {
	return repositories.ConversationReadStateRepository.FindPageByCnd(sqls.DB(), cnd)
}

func (s *conversationReadStateService) Count(cnd *sqls.Cnd) int64 {
	return repositories.ConversationReadStateRepository.Count(sqls.DB(), cnd)
}

func (s *conversationReadStateService) Create(item *models.ConversationReadState) error {
	return repositories.ConversationReadStateRepository.Create(sqls.DB(), item)
}

func (s *conversationReadStateService) Update(item *models.ConversationReadState) error {
	return repositories.ConversationReadStateRepository.Update(sqls.DB(), item)
}

func (s *conversationReadStateService) Updates(id int64, columns map[string]any) error {
	return repositories.ConversationReadStateRepository.Updates(sqls.DB(), id, columns)
}

func (s *conversationReadStateService) UpdateColumn(id int64, name string, value any) error {
	return repositories.ConversationReadStateRepository.UpdateColumn(sqls.DB(), id, name, value)
}

func (s *conversationReadStateService) Delete(id int64) {
	repositories.ConversationReadStateRepository.Delete(sqls.DB(), id)
}

// GetByAgentReader 查询客服侧已读游标。
func (s *conversationReadStateService) GetByAgentReader(conversationID int64, operator *dto.AuthPrincipal) *models.ConversationReadState {
	if operator == nil {
		return nil
	}
	return s.getByCursor(conversationID, readerCursor{
		readerType:       enums.IMSenderTypeAgent,
		readerID:         operator.UserID,
		externalReaderID: "",
	})
}

// GetByCustomerReader 查询 IM 客户侧已读游标（按 ExternalID）。
func (s *conversationReadStateService) GetByCustomerReader(conversationID int64, external *openidentity.ExternalUser) *models.ConversationReadState {
	if external == nil || strings.TrimSpace(external.ExternalID) == "" {
		return nil
	}
	return s.getByCursor(conversationID, readerCursor{
		readerType:       enums.IMSenderTypeCustomer,
		readerID:         0,
		externalReaderID: strings.TrimSpace(external.ExternalID),
	})
}

func (s *conversationReadStateService) getByCursor(conversationID int64, c readerCursor) *models.ConversationReadState {
	return s.FindOne(sqls.NewCnd().
		Eq("conversation_id", conversationID).
		Eq("reader_type", c.readerType).
		Eq("reader_id", c.readerID).
		Eq("external_reader_id", c.externalReaderID))
}

func (s *conversationReadStateService) GetConversationReadStates(conversationID int64) (agentState, customerState *models.ConversationReadState) {
	return s.getConversationReadStates(sqls.DB(), conversationID)
}

func (s *conversationReadStateService) getConversationReadStates(db *gorm.DB, conversationID int64) (agentState, customerState *models.ConversationReadState) {
	list := repositories.ConversationReadStateRepository.Find(db, sqls.NewCnd().Eq("conversation_id", conversationID))
	return s.pickConversationReadStates(list)
}

func (s *conversationReadStateService) pickConversationReadStates(list []models.ConversationReadState) (agentState, customerState *models.ConversationReadState) {
	for i := range list {
		item := &list[i]
		switch item.ReaderType {
		case enums.IMSenderTypeAgent:
			if agentState == nil || item.LastReadSeqNo > agentState.LastReadSeqNo {
				agentState = item
			}
		case enums.IMSenderTypeCustomer:
			if customerState == nil || item.LastReadSeqNo > customerState.LastReadSeqNo {
				customerState = item
			}
		}
	}
	return agentState, customerState
}

// MarkAgentRead 在事务内更新/创建客服已读游标。
func (s *conversationReadStateService) MarkAgentRead(ctx *sqls.TxContext, conversation *models.Conversation, operator *dto.AuthPrincipal, message *models.Message, now time.Time) (*models.ConversationReadState, error) {
	c, err := agentReaderCursor(operator)
	if err != nil {
		return nil, err
	}
	return s.markReadTxWithCursor(ctx, conversation, c, message, now)
}

// MarkCustomerRead 在事务内更新/创建 IM 客户已读游标。
func (s *conversationReadStateService) MarkCustomerRead(ctx *sqls.TxContext, conversation *models.Conversation, external *openidentity.ExternalUser, message *models.Message, now time.Time) (*models.ConversationReadState, error) {
	c, err := customerReaderCursor(external)
	if err != nil {
		return nil, err
	}
	return s.markReadTxWithCursor(ctx, conversation, c, message, now)
}

func (s *conversationReadStateService) markReadTxWithCursor(ctx *sqls.TxContext, conversation *models.Conversation, c readerCursor, message *models.Message, now time.Time) (*models.ConversationReadState, error) {
	if ctx == nil || conversation == nil || message == nil {
		return nil, nil
	}
	if c.readerType != enums.IMSenderTypeAgent && c.readerType != enums.IMSenderTypeCustomer {
		return nil, errorsx.InvalidParam("不支持的已读操作类型")
	}

	item := &models.ConversationReadState{}
	err := ctx.Tx.Where("conversation_id = ? AND reader_type = ? AND reader_id = ? AND external_reader_id = ?",
		conversation.ID, c.readerType, c.readerID, c.externalReaderID,
	).First(item).Error
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			return nil, err
		}
		item = &models.ConversationReadState{
			ConversationID:    conversation.ID,
			ReaderType:        c.readerType,
			ReaderID:          c.readerID,
			ExternalReaderID:  c.externalReaderID,
			LastReadMessageID: message.ID,
			LastReadSeqNo:     message.SeqNo,
			LastReadAt:        &now,
			AuditFields: models.AuditFields{
				CreatedAt:      now,
				CreateUserID:   c.auditUserID,
				CreateUserName: c.auditUserName,
				UpdatedAt:      now,
				UpdateUserID:   c.auditUserID,
				UpdateUserName: c.auditUserName,
			},
		}
		if err := ctx.Tx.Create(item).Error; err != nil {
			return nil, err
		}
		return item, nil
	}

	if item.LastReadSeqNo >= message.SeqNo {
		return item, nil
	}

	item.LastReadMessageID = message.ID
	item.LastReadSeqNo = message.SeqNo
	item.LastReadAt = &now
	item.UpdatedAt = now
	item.UpdateUserID = c.auditUserID
	item.UpdateUserName = c.auditUserName
	if err := repositories.ConversationReadStateRepository.Updates(ctx.Tx, item.ID, map[string]any{
		"last_read_message_id": item.LastReadMessageID,
		"last_read_seq_no":     item.LastReadSeqNo,
		"last_read_at":         item.LastReadAt,
		"updated_at":           item.UpdatedAt,
		"update_user_id":       item.UpdateUserID,
		"update_user_name":     item.UpdateUserName,
	}); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *conversationReadStateService) CountUnreadMessages(ctx *sqls.TxContext, conversationID, lastReadSeqNo int64, senderTypes ...enums.IMSenderType) (int64, error) {
	normalizedSenderTypes := make([]enums.IMSenderType, 0, len(senderTypes))
	for _, senderType := range senderTypes {
		if strs.IsBlank(string(senderType)) {
			continue
		}
		normalizedSenderTypes = append(normalizedSenderTypes, senderType)
	}
	if len(normalizedSenderTypes) == 0 {
		return 0, nil
	}
	var count int64
	query := ctx.Tx.Model(&models.Message{}).
		Where("conversation_id = ? AND seq_no > ? AND recalled_at IS NULL AND send_status <> ?", conversationID, lastReadSeqNo, int(enums.IMMessageStatusRecalled))
	if len(normalizedSenderTypes) == 1 {
		query = query.Where("sender_type = ?", normalizedSenderTypes[0])
	} else {
		query = query.Where("sender_type IN ?", normalizedSenderTypes)
	}
	err := query.Count(&count).Error
	return count, err
}
