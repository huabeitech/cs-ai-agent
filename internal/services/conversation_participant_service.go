package services

import (
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/pkg/openidentity"
	"cs-agent/internal/pkg/utils"
	"cs-agent/internal/repositories"
	"time"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
)

var ConversationParticipantService = newConversationParticipantService()

func newConversationParticipantService() *conversationParticipantService {
	return &conversationParticipantService{}
}

type conversationParticipantService struct {
}

func (s *conversationParticipantService) Get(id int64) *models.ConversationParticipant {
	return repositories.ConversationParticipantRepository.Get(sqls.DB(), id)
}

func (s *conversationParticipantService) Take(where ...interface{}) *models.ConversationParticipant {
	return repositories.ConversationParticipantRepository.Take(sqls.DB(), where...)
}

func (s *conversationParticipantService) Find(cnd *sqls.Cnd) []models.ConversationParticipant {
	return repositories.ConversationParticipantRepository.Find(sqls.DB(), cnd)
}

func (s *conversationParticipantService) FindOne(cnd *sqls.Cnd) *models.ConversationParticipant {
	return repositories.ConversationParticipantRepository.FindOne(sqls.DB(), cnd)
}

func (s *conversationParticipantService) FindPageByParams(params *params.QueryParams) (list []models.ConversationParticipant, paging *sqls.Paging) {
	return repositories.ConversationParticipantRepository.FindPageByParams(sqls.DB(), params)
}

func (s *conversationParticipantService) FindPageByCnd(cnd *sqls.Cnd) (list []models.ConversationParticipant, paging *sqls.Paging) {
	return repositories.ConversationParticipantRepository.FindPageByCnd(sqls.DB(), cnd)
}

func (s *conversationParticipantService) Count(cnd *sqls.Cnd) int64 {
	return repositories.ConversationParticipantRepository.Count(sqls.DB(), cnd)
}

func (s *conversationParticipantService) Create(t *models.ConversationParticipant) error {
	return repositories.ConversationParticipantRepository.Create(sqls.DB(), t)
}

func (s *conversationParticipantService) Update(t *models.ConversationParticipant) error {
	return repositories.ConversationParticipantRepository.Update(sqls.DB(), t)
}

func (s *conversationParticipantService) Updates(id int64, columns map[string]interface{}) error {
	return repositories.ConversationParticipantRepository.Updates(sqls.DB(), id, columns)
}

func (s *conversationParticipantService) UpdateColumn(id int64, name string, value interface{}) error {
	return repositories.ConversationParticipantRepository.UpdateColumn(sqls.DB(), id, name, value)
}

func (s *conversationParticipantService) Delete(id int64) {
	repositories.ConversationParticipantRepository.Delete(sqls.DB(), id)
}

func (s *conversationParticipantService) CreateCustomerParticipant(ctx *sqls.TxContext, conversationID int64, externalUser openidentity.ExternalUser) error {
	return repositories.ConversationParticipantRepository.Create(ctx.Tx, &models.ConversationParticipant{
		ConversationID:        conversationID,
		ParticipantType:       string(enums.IMParticipantTypeCustomer),
		ParticipantID:         0,
		ExternalParticipantID: externalUser.ExternalID,
		JoinedAt:              new(time.Now()),
		Status:                enums.StatusOk,
		AuditFields:           utils.BuildAuditFields(nil),
	})
}
