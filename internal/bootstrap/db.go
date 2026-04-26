package bootstrap

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"cs-agent/internal/pkg/config"

	"github.com/mlogclub/simple/sqls"
	"gorm.io/driver/mysql"

	// "gorm.io/driver/sqlite" // Sqlite driver based on CGO
	"github.com/glebarez/sqlite" // Pure go SQLite driver, checkout https://github.com/glebarez/sqlite for details
	"gorm.io/gorm"

	// "gorm.io/gorm/logger"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

func InitDB(cfg config.DBConfig) (*gorm.DB, error) {
	var dialector gorm.Dialector
	switch cfg.Type {
	case "sqlite":
		if err := ensureSQLiteDir(cfg.DSN); err != nil {
			return nil, err
		}
		dialector = sqlite.Open(cfg.DSN)
	case "mysql":
		dialector = mysql.Open(cfg.DSN)
	default:
		return nil, fmt.Errorf("unsupported db type: %s", cfg.Type)
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			logger.Config{
				SlowThreshold:             time.Second,
				LogLevel:                  logger.Warn,
				IgnoreRecordNotFoundError: true,
				Colorful:                  true,
			},
		),
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "t_",
			SingularTable: true,
		},
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	if cfg.MaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	}
	if cfg.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	if cfg.ConnMaxIdleTimeSeconds > 0 {
		sqlDB.SetConnMaxIdleTime(time.Duration(cfg.ConnMaxIdleTimeSeconds) * time.Second)
	}
	if cfg.ConnMaxLifetimeSeconds > 0 {
		sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetimeSeconds) * time.Second)
	}

	sqls.SetDB(db)
	return db, nil
}

func ensureSQLiteDir(dsn string) error {
	dbPath := sqliteFilePath(dsn)
	if dbPath == "" {
		return nil
	}
	dir := filepath.Dir(dbPath)
	if dir == "." || dir == "" {
		return nil
	}
	return os.MkdirAll(dir, 0o755)
}

func sqliteFilePath(dsn string) string {
	if dsn == "" {
		return ""
	}

	path := dsn
	if after, ok := strings.CutPrefix(path, "file:"); ok {
		path = after
	}
	if idx := strings.Index(path, "?"); idx >= 0 {
		path = path[:idx]
	}

	normalized := strings.TrimSpace(path)
	if normalized == "" || normalized == ":memory:" || strings.Contains(normalized, "mode=memory") {
		return ""
	}
	return normalized
}
