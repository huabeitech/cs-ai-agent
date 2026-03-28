package services

import (
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/dto"
	"cs-agent/internal/pkg/dto/request"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/pkg/errorsx"
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

// GetByReader 查询某会话下指定读者类型的已读游标。
// 客服侧传 operator；IM 客户侧传 external（按 ExternalID 匹配 external_reader_id）。
func (s *conversationReadStateService) GetByReader(conversationID int64, readerType enums.IMSenderType, operator *dto.AuthPrincipal, external *request.ExternalInfo) *models.ConversationReadState {
	cnd := sqls.NewCnd().
		Eq("conversation_id", conversationID).
		Eq("reader_type", readerType)
	switch readerType {
	case enums.IMSenderTypeCustomer:
		if external == nil || strings.TrimSpace(external.ExternalID) == "" {
			return nil
		}
		cnd = cnd.Eq("external_reader_id", strings.TrimSpace(external.ExternalID)).Eq("reader_id", int64(0))
	case enums.IMSenderTypeAgent:
		if operator == nil {
			return nil
		}
		cnd = cnd.Eq("reader_id", operator.UserID).Eq("external_reader_id", "")
	default:
		return nil
	}
	return s.FindOne(cnd)
}

func (s *conversationReadStateService) GetConversationReadStates(conversationID int64) (agentState, customerState *models.ConversationReadState) {
	list := s.Find(sqls.NewCnd().Eq("conversation_id", conversationID))
	return s.pickConversationReadStates(list)
}

func (s *conversationReadStateService) GetConversationReadStatesTx(ctx *sqls.TxContext, conversationID int64) (agentState, customerState *models.ConversationReadState, err error) {
	var list []models.ConversationReadState
	err = ctx.Tx.Where("conversation_id = ?", conversationID).Find(&list).Error
	if err != nil {
		return nil, nil, err
	}
	agentState, customerState = s.pickConversationReadStates(list)
	return agentState, customerState, nil
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

// MarkReadTx 更新或创建已读游标。客服传 operator；IM 客户传 external（operator 可为 nil）。
func (s *conversationReadStateService) MarkReadTx(ctx *sqls.TxContext, conversation *models.Conversation, readerType enums.IMSenderType, operator *dto.AuthPrincipal, external *request.ExternalInfo, message *models.Message, now time.Time) (*models.ConversationReadState, error) {
	if ctx == nil || conversation == nil || message == nil {
		return nil, nil
	}
	if readerType != enums.IMSenderTypeAgent && readerType != enums.IMSenderTypeCustomer {
		return nil, errorsx.InvalidParam("不支持的已读操作类型")
	}

	var readerID int64
	var externalReaderID string
	var auditUserID int64
	var auditUserName string

	switch readerType {
	case enums.IMSenderTypeAgent:
		if operator == nil {
			return nil, errorsx.Unauthorized("未登录或登录已过期")
		}
		readerID = operator.UserID
		externalReaderID = ""
		auditUserID = operator.UserID
		auditUserName = operator.Username
	case enums.IMSenderTypeCustomer:
		if external == nil || strings.TrimSpace(external.ExternalID) == "" {
			return nil, errorsx.Unauthorized("外部用户标识不能为空")
		}
		readerID = 0
		externalReaderID = strings.TrimSpace(external.ExternalID)
		auditUserID = 0
		auditUserName = strings.TrimSpace(external.ExternalName)
		if auditUserName == "" {
			auditUserName = externalReaderID
		}
	}

	item := &models.ConversationReadState{}
	err := ctx.Tx.Where("conversation_id = ? AND reader_type = ? AND reader_id = ? AND external_reader_id = ?",
		conversation.ID, readerType, readerID, externalReaderID,
	).First(item).Error
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			return nil, err
		}
		item = &models.ConversationReadState{
			ConversationID:    conversation.ID,
			ReaderType:        readerType,
			ReaderID:          readerID,
			ExternalReaderID:  externalReaderID,
			LastReadMessageID: message.ID,
			LastReadSeqNo:     message.SeqNo,
			LastReadAt:        &now,
			AuditFields: models.AuditFields{
				CreatedAt:      now,
				CreateUserID:   auditUserID,
				CreateUserName: auditUserName,
				UpdatedAt:      now,
				UpdateUserID:   auditUserID,
				UpdateUserName: auditUserName,
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

	updates := map[string]any{
		"last_read_message_id": message.ID,
		"last_read_seq_no":     message.SeqNo,
		"last_read_at":         now,
		"update_user_id":       auditUserID,
		"update_user_name":     auditUserName,
		"updated_at":           now,
	}
	if err := repositories.ConversationReadStateRepository.Updates(ctx.Tx, item.ID, updates); err != nil {
		return nil, err
	}

	item.LastReadMessageID = message.ID
	item.LastReadSeqNo = message.SeqNo
	item.LastReadAt = &now
	item.UpdatedAt = now
	item.UpdateUserID = auditUserID
	item.UpdateUserName = auditUserName
	return item, nil
}

func (s *conversationReadStateService) CountUnreadMessagesTx(ctx *sqls.TxContext, conversationID, lastReadSeqNo int64, senderTypes ...enums.IMSenderType) (int64, error) {
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
		Where("conversation_id = ? AND seq_no > ?", conversationID, lastReadSeqNo)
	if len(normalizedSenderTypes) == 1 {
		query = query.Where("sender_type = ?", normalizedSenderTypes[0])
	} else {
		query = query.Where("sender_type IN ?", normalizedSenderTypes)
	}
	err := query.Count(&count).Error
	return count, err
}
