package services

import (
	"cs-agent/internal/models"
	"cs-agent/internal/repositories"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

var TicketNoService = newTicketNoService()

func newTicketNoService() *ticketNoService {
	return &ticketNoService{}
}

type ticketNoService struct{}

func (s *ticketNoService) Next(tx *gorm.DB, now time.Time) (string, error) {
	if tx == nil {
		return "", fmt.Errorf("ticket number transaction is required")
	}
	dateKey := now.Format("20060102")
	for attempt := 0; attempt < 20; attempt++ {
		current := repositories.TicketNoSequenceRepository.GetByDateKey(tx, dateKey)
		if current == nil {
			item := &models.TicketNoSequence{
				DateKey:   dateKey,
				NextSeq:   2,
				CreatedAt: now,
				UpdatedAt: now,
			}
			err := repositories.TicketNoSequenceRepository.Create(tx, item)
			if err == nil {
				return formatTicketNo(dateKey, 1), nil
			}
			if !isRetriableTicketNoError(err) {
				return "", err
			}
			time.Sleep(time.Duration(attempt+1) * 10 * time.Millisecond)
			continue
		}
		allocated := current.NextSeq
		ok, err := repositories.TicketNoSequenceRepository.UpdateNextSeq(tx, current.ID, current.NextSeq, current.NextSeq+1, now)
		if err != nil {
			if isRetriableTicketNoError(err) {
				time.Sleep(time.Duration(attempt+1) * 10 * time.Millisecond)
				continue
			}
			return "", err
		}
		if ok {
			return formatTicketNo(dateKey, allocated), nil
		}
		time.Sleep(time.Duration(attempt+1) * 10 * time.Millisecond)
	}
	return "", fmt.Errorf("failed to allocate ticket number")
}

func formatTicketNo(dateKey string, seq int64) string {
	return fmt.Sprintf("TK%s%05d", dateKey, seq)
}

func isDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "duplicate") || strings.Contains(message, "unique") || strings.Contains(message, "constraint failed")
}

func isRetriableTicketNoError(err error) bool {
	return isDuplicateKeyError(err) || isDatabaseLockedError(err)
}

func isDatabaseLockedError(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "database is locked") || strings.Contains(message, "database table is locked")
}
