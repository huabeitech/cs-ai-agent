package services

import (
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/dto"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/pkg/errorsx"
	"cs-agent/internal/repositories"
	"time"

	"github.com/mlogclub/simple/sqls"
	"gorm.io/gorm"
)

var TicketTagService = newTicketTagService()

func newTicketTagService() *ticketTagService {
	return &ticketTagService{}
}

type ticketTagService struct{}

func (s *ticketTagService) Get(id int64) *models.TicketTag {
	return repositories.TicketTagRepository.Get(sqls.DB(), id)
}

func (s *ticketTagService) Take(where ...interface{}) *models.TicketTag {
	return repositories.TicketTagRepository.Take(sqls.DB(), where...)
}

func (s *ticketTagService) Find(cnd *sqls.Cnd) []models.TicketTag {
	return repositories.TicketTagRepository.Find(sqls.DB(), cnd)
}

func (s *ticketTagService) Create(db *gorm.DB, item *models.TicketTag) error {
	return repositories.TicketTagRepository.Create(db, item)
}

func (s *ticketTagService) DeleteByTicketID(db *gorm.DB, ticketID int64) error {
	return repositories.TicketTagRepository.DeleteByTicketID(db, ticketID)
}

func (s *ticketTagService) NormalizeTagIDs(tagIDs []int64) []int64 {
	if len(tagIDs) == 0 {
		return nil
	}
	seen := make(map[int64]struct{}, len(tagIDs))
	result := make([]int64, 0, len(tagIDs))
	for _, tagID := range tagIDs {
		if tagID <= 0 {
			continue
		}
		if _, ok := seen[tagID]; ok {
			continue
		}
		seen[tagID] = struct{}{}
		result = append(result, tagID)
	}
	return result
}

func (s *ticketTagService) ValidateTagIDs(tagIDs []int64) ([]int64, error) {
	normalized := s.NormalizeTagIDs(tagIDs)
	if len(normalized) == 0 {
		return nil, nil
	}
	tags := repositories.TagRepository.Find(sqls.DB(), sqls.NewCnd().In("id", normalized))
	if len(tags) != len(normalized) {
		return nil, errorsx.InvalidParam("存在无效工单标签")
	}
	for i := range tags {
		if tags[i].Status != enums.StatusOk {
			return nil, errorsx.InvalidParam("存在未启用的工单标签")
		}
	}
	return normalized, nil
}

func (s *ticketTagService) ReplaceTicketTags(db *gorm.DB, ticketID int64, tagIDs []int64, operator *dto.AuthPrincipal) error {
	if err := s.DeleteByTicketID(db, ticketID); err != nil {
		return err
	}
	if len(tagIDs) == 0 {
		return nil
	}
	now := time.Now()
	for _, tagID := range tagIDs {
		if err := s.Create(db, &models.TicketTag{
			TicketID: ticketID,
			TagID:    tagID,
			AuditFields: models.AuditFields{
				CreatedAt:      now,
				CreateUserID:   operator.UserID,
				CreateUserName: operator.Username,
				UpdatedAt:      now,
				UpdateUserID:   operator.UserID,
				UpdateUserName: operator.Username,
			},
		}); err != nil {
			return err
		}
	}
	return nil
}
