package bootstrap

import (
	"cs-agent/internal/migration"
	"cs-agent/internal/models"

	"github.com/mlogclub/simple/sqls"
)

func InitMigrations() error {
	if err := sqls.DB().AutoMigrate(models.Models...); err != nil {
		return err
	}
	return migration.Migrate()
}
