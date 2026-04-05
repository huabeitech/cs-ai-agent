package migration

import (
	"time"

	"cs-agent/internal/models"
	"cs-agent/internal/pkg/constants"
	"cs-agent/internal/pkg/enums"

	"github.com/mlogclub/simple/sqls"
)

func init() {
	register(3, "init ticket priority configs", func() error {
		return sqls.WithTransaction(func(ctx *sqls.TxContext) error {
			count := ctx.Tx.Model(&models.TicketPriorityConfig{}).Where("status <> ?", enums.StatusDeleted).RowsAffected
			if count == 0 {
				var total int64
				if err := ctx.Tx.Model(&models.TicketPriorityConfig{}).Where("status <> ?", enums.StatusDeleted).Count(&total).Error; err != nil {
					return err
				}
				if total == 0 {
					now := time.Now()
					items := []*models.TicketPriorityConfig{
						{Name: "普通", SortNo: 10, FirstResponseMinutes: 30, ResolutionMinutes: 1440, Status: enums.StatusOk},
						{Name: "高", SortNo: 20, FirstResponseMinutes: 10, ResolutionMinutes: 240, Status: enums.StatusOk},
						{Name: "紧急", SortNo: 30, FirstResponseMinutes: 5, ResolutionMinutes: 120, Status: enums.StatusOk},
					}
					for _, item := range items {
						item.AuditFields = models.AuditFields{
							CreatedAt:      now,
							CreateUserID:   constants.SystemAuditUserID,
							CreateUserName: constants.SystemAuditUserName,
							UpdatedAt:      now,
							UpdateUserID:   constants.SystemAuditUserID,
							UpdateUserName: constants.SystemAuditUserName,
						}
						if err := ctx.Tx.Create(item).Error; err != nil {
							return err
						}
					}
				}
			}
			return nil
		})
	})
}
