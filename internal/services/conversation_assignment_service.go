package services

import (
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/dto"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/repositories"
	"strings"
	"time"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
)

var ConversationAssignmentService = newConversationAssignmentService()

func newConversationAssignmentService() *conversationAssignmentService {
	return &conversationAssignmentService{}
}

type conversationAssignmentService struct {
}

func (s *conversationAssignmentService) Get(id int64) *models.ConversationAssignment {
	return repositories.ConversationAssignmentRepository.Get(sqls.DB(), id)
}

func (s *conversationAssignmentService) Take(where ...interface{}) *models.ConversationAssignment {
	return repositories.ConversationAssignmentRepository.Take(sqls.DB(), where...)
}

func (s *conversationAssignmentService) Find(cnd *sqls.Cnd) []models.ConversationAssignment {
	return repositories.ConversationAssignmentRepository.Find(sqls.DB(), cnd)
}

func (s *conversationAssignmentService) FindOne(cnd *sqls.Cnd) *models.ConversationAssignment {
	return repositories.ConversationAssignmentRepository.FindOne(sqls.DB(), cnd)
}

func (s *conversationAssignmentService) FindPageByParams(params *params.QueryParams) (list []models.ConversationAssignment, paging *sqls.Paging) {
	return repositories.ConversationAssignmentRepository.FindPageByParams(sqls.DB(), params)
}

func (s *conversationAssignmentService) FindPageByCnd(cnd *sqls.Cnd) (list []models.ConversationAssignment, paging *sqls.Paging) {
	return repositories.ConversationAssignmentRepository.FindPageByCnd(sqls.DB(), cnd)
}

func (s *conversationAssignmentService) Count(cnd *sqls.Cnd) int64 {
	return repositories.ConversationAssignmentRepository.Count(sqls.DB(), cnd)
}

func (s *conversationAssignmentService) FinishActiveAssignments(ctx *sqls.TxContext, conversationID int64, finishedAt time.Time) error {
	return ctx.Tx.Model(&models.ConversationAssignment{}).
		Where("conversation_id = ? AND status = ?", conversationID, enums.IMAssignmentStatusActive).
		Updates(map[string]any{
			"status":      enums.IMAssignmentStatusInactive,
			"finished_at": finishedAt,
		}).Error
}

func (s *conversationAssignmentService) CreateAssignment(ctx *sqls.TxContext, conversationID, fromUserID, toUserID int64, assignType enums.IMAssignmentType, reason string, operator *dto.AuthPrincipal, now time.Time) error {
	assignment := &models.ConversationAssignment{
		ConversationID: conversationID,
		FromUserID:     fromUserID,
		ToUserID:       toUserID,
		AssignType:     strings.TrimSpace(string(assignType)),
		Reason:         strings.TrimSpace(reason),
		Status:         enums.IMAssignmentStatusActive,
		CreatedAt:      now,
	}
	if operator != nil {
		assignment.OperatorID = operator.UserID
	}
	return ctx.Tx.Create(assignment).Error
}
