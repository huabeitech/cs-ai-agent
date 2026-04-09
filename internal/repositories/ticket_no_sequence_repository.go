package repositories

import (
	"cs-agent/internal/models"
	"time"

	"gorm.io/gorm"
)

var TicketNoSequenceRepository = newTicketNoSequenceRepository()

func newTicketNoSequenceRepository() *ticketNoSequenceRepository {
	return &ticketNoSequenceRepository{}
}

type ticketNoSequenceRepository struct{}

func (r *ticketNoSequenceRepository) GetByDateKey(db *gorm.DB, dateKey string) *models.TicketNoSequence {
	ret := &models.TicketNoSequence{}
	if err := db.Take(ret, "date_key = ?", dateKey).Error; err != nil {
		return nil
	}
	return ret
}

func (r *ticketNoSequenceRepository) Create(db *gorm.DB, t *models.TicketNoSequence) error {
	return db.Create(t).Error
}

func (r *ticketNoSequenceRepository) UpdateNextSeq(db *gorm.DB, id int64, currentSeq, nextSeq int64, updatedAt time.Time) (bool, error) {
	result := db.Model(&models.TicketNoSequence{}).
		Where("id = ? AND next_seq = ?", id, currentSeq).
		Updates(map[string]any{
			"next_seq":   nextSeq,
			"updated_at": updatedAt,
		})
	return result.RowsAffected == 1, result.Error
}
