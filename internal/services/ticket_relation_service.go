package services

import (
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/dto"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/pkg/errorsx"
	"cs-agent/internal/repositories"
	"strings"
	"time"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
)

var TicketRelationService = newTicketRelationService()

func newTicketRelationService() *ticketRelationService {
	return &ticketRelationService{}
}

type ticketRelationService struct {
}

func (s *ticketRelationService) Get(id int64) *models.TicketRelation {
	return repositories.TicketRelationRepository.Get(sqls.DB(), id)
}

func (s *ticketRelationService) Take(where ...interface{}) *models.TicketRelation {
	return repositories.TicketRelationRepository.Take(sqls.DB(), where...)
}

func (s *ticketRelationService) Find(cnd *sqls.Cnd) []models.TicketRelation {
	return repositories.TicketRelationRepository.Find(sqls.DB(), cnd)
}

func (s *ticketRelationService) FindOne(cnd *sqls.Cnd) *models.TicketRelation {
	return repositories.TicketRelationRepository.FindOne(sqls.DB(), cnd)
}

func (s *ticketRelationService) FindPageByParams(params *params.QueryParams) (list []models.TicketRelation, paging *sqls.Paging) {
	return repositories.TicketRelationRepository.FindPageByParams(sqls.DB(), params)
}

func (s *ticketRelationService) FindPageByCnd(cnd *sqls.Cnd) (list []models.TicketRelation, paging *sqls.Paging) {
	return repositories.TicketRelationRepository.FindPageByCnd(sqls.DB(), cnd)
}

func (s *ticketRelationService) Count(cnd *sqls.Cnd) int64 {
	return repositories.TicketRelationRepository.Count(sqls.DB(), cnd)
}

func (s *ticketRelationService) Create(t *models.TicketRelation) error {
	return repositories.TicketRelationRepository.Create(sqls.DB(), t)
}

func (s *ticketRelationService) Update(t *models.TicketRelation) error {
	return repositories.TicketRelationRepository.Update(sqls.DB(), t)
}

func (s *ticketRelationService) Updates(id int64, columns map[string]interface{}) error {
	return repositories.TicketRelationRepository.Updates(sqls.DB(), id, columns)
}

func (s *ticketRelationService) UpdateColumn(id int64, name string, value interface{}) error {
	return repositories.TicketRelationRepository.UpdateColumn(sqls.DB(), id, name, value)
}

func (s *ticketRelationService) Delete(id int64) {
	repositories.TicketRelationRepository.Delete(sqls.DB(), id)
}

func (s *ticketRelationService) AddRelation(ticketID, relatedTicketID int64, relationType enums.TicketRelationType, operator *dto.AuthPrincipal) error {
	if operator == nil {
		return errorsx.Unauthorized("未登录或登录已过期")
	}
	if ticketID <= 0 || relatedTicketID <= 0 {
		return errorsx.InvalidParam("工单不存在")
	}
	if ticketID == relatedTicketID {
		return errorsx.InvalidParam("不能关联自己")
	}
	if !isValidTicketRelationType(relationType) {
		return errorsx.InvalidParam("关联类型不合法")
	}
	ticket := TicketService.Get(ticketID)
	if ticket == nil {
		return errorsx.InvalidParam("工单不存在")
	}
	relatedTicket := TicketService.Get(relatedTicketID)
	if relatedTicket == nil {
		return errorsx.InvalidParam("关联工单不存在")
	}
	if repositories.TicketRelationRepository.Take(sqls.DB(), "ticket_id = ? AND related_ticket_id = ? AND relation_type = ?", ticketID, relatedTicketID, relationType) != nil {
		return errorsx.InvalidParam("该关联已存在")
	}
	now := time.Now()
	inverseType := inverseTicketRelationType(relationType)
	return sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		if err := repositories.TicketRelationRepository.Create(ctx.Tx, &models.TicketRelation{
			TicketID:        ticketID,
			RelatedTicketID: relatedTicketID,
			RelationType:    relationType,
			CreatedAt:       now,
		}); err != nil {
			return err
		}
		if repositories.TicketRelationRepository.Take(ctx.Tx, "ticket_id = ? AND related_ticket_id = ? AND relation_type = ?", relatedTicketID, ticketID, inverseType) == nil {
			if err := repositories.TicketRelationRepository.Create(ctx.Tx, &models.TicketRelation{
				TicketID:        relatedTicketID,
				RelatedTicketID: ticketID,
				RelationType:    inverseType,
				CreatedAt:       now,
			}); err != nil {
				return err
			}
		}
		if err := repositories.TicketEventLogRepository.Create(ctx.Tx, &models.TicketEventLog{
			TicketID:     ticketID,
			EventType:    enums.TicketEventTypeUpdated,
			OperatorType: enums.IMSenderTypeAgent,
			OperatorID:   operator.UserID,
			Content:      "新增关联工单",
			Payload:      strings.TrimSpace(string(relationType) + ":" + relatedTicket.TicketNo),
			CreatedAt:    now,
		}); err != nil {
			return err
		}
		return repositories.TicketEventLogRepository.Create(ctx.Tx, &models.TicketEventLog{
			TicketID:     relatedTicketID,
			EventType:    enums.TicketEventTypeUpdated,
			OperatorType: enums.IMSenderTypeAgent,
			OperatorID:   operator.UserID,
			Content:      "新增关联工单",
			Payload:      strings.TrimSpace(string(inverseType) + ":" + ticket.TicketNo),
			CreatedAt:    now,
		})
	})
}

func (s *ticketRelationService) DeleteRelation(ticketID, relationID int64, operator *dto.AuthPrincipal) error {
	if operator == nil {
		return errorsx.Unauthorized("未登录或登录已过期")
	}
	relation := s.Get(relationID)
	if relation == nil || relation.TicketID != ticketID {
		return errorsx.InvalidParam("关联关系不存在")
	}
	ticket := TicketService.Get(relation.TicketID)
	relatedTicket := TicketService.Get(relation.RelatedTicketID)
	now := time.Now()
	return sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		if err := repositories.TicketRelationRepository.DeleteByTicketRelation(ctx.Tx, relation.TicketID, relation.RelatedTicketID, string(relation.RelationType)); err != nil {
			return err
		}
		if err := repositories.TicketRelationRepository.DeleteByTicketRelation(ctx.Tx, relation.RelatedTicketID, relation.TicketID, string(inverseTicketRelationType(relation.RelationType))); err != nil {
			return err
		}
		if ticket != nil {
			if err := repositories.TicketEventLogRepository.Create(ctx.Tx, &models.TicketEventLog{
				TicketID:     ticket.ID,
				EventType:    enums.TicketEventTypeUpdated,
				OperatorType: enums.IMSenderTypeAgent,
				OperatorID:   operator.UserID,
				Content:      "移除关联工单",
				Payload:      strings.TrimSpace(string(relation.RelationType) + ":" + relationTicketNo(relatedTicket)),
				CreatedAt:    now,
			}); err != nil {
				return err
			}
		}
		if relatedTicket != nil {
			return repositories.TicketEventLogRepository.Create(ctx.Tx, &models.TicketEventLog{
				TicketID:     relatedTicket.ID,
				EventType:    enums.TicketEventTypeUpdated,
				OperatorType: enums.IMSenderTypeAgent,
				OperatorID:   operator.UserID,
				Content:      "移除关联工单",
				Payload:      strings.TrimSpace(string(inverseTicketRelationType(relation.RelationType)) + ":" + relationTicketNo(ticket)),
				CreatedAt:    now,
			})
		}
		return nil
	})
}

func isValidTicketRelationType(relationType enums.TicketRelationType) bool {
	switch relationType {
	case enums.TicketRelationTypeDuplicate, enums.TicketRelationTypeRelated, enums.TicketRelationTypeParent, enums.TicketRelationTypeChild:
		return true
	default:
		return false
	}
}

func inverseTicketRelationType(relationType enums.TicketRelationType) enums.TicketRelationType {
	switch relationType {
	case enums.TicketRelationTypeParent:
		return enums.TicketRelationTypeChild
	case enums.TicketRelationTypeChild:
		return enums.TicketRelationTypeParent
	default:
		return relationType
	}
}

func relationTicketNo(ticket *models.Ticket) string {
	if ticket == nil {
		return ""
	}
	return ticket.TicketNo
}
