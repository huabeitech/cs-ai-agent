package migration

import (
	"cs-agent/internal/models"

	"github.com/mlogclub/simple/sqls"
)

func init() {
	register(6, "backfill conversation customer name", func() error {
		db := sqls.DB()
		if !db.Migrator().HasColumn(&models.Conversation{}, "customer_name") {
			return nil
		}
		return db.Exec(`
UPDATE t_conversation
SET customer_name = (
  SELECT name FROM t_customer WHERE t_customer.id = t_conversation.customer_id
)
WHERE customer_id > 0
  AND (customer_name = '' OR customer_name IS NULL)
`).Error
	})
}
