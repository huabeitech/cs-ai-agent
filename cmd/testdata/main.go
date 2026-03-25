package main

import (
	"cs-agent/cmd/testdata/agentteam"
	"cs-agent/cmd/testdata/aiagent"
	"cs-agent/cmd/testdata/aiconfig"
	"cs-agent/cmd/testdata/kb"
	"cs-agent/cmd/testdata/widgetsite"
	"cs-agent/internal/bootstrap"
	"cs-agent/internal/pkg/config"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/mlogclub/simple/sqls"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func main() {
	if err := run(); err != nil {
		slog.Error("init testdata failed", "error", err)
		os.Exit(1)
	}
}

func run() error {
	configPath := flag.String("config", "config/config.yaml", "path to config file")
	autoConfirm := flag.Bool("yes", false, "skip confirmation prompt")
	flag.Parse()

	if err := confirmDestructiveAction(*autoConfirm); err != nil {
		return err
	}

	cfg, err := config.Load(*configPath)
	if err != nil {
		return fmt.Errorf("load config failed: %w", err)
	}

	db, err := bootstrap.InitDB(cfg.DB)
	if err != nil {
		return fmt.Errorf("init db failed: %w", err)
	}
	db = withSilentSQLLogger(db)
	sqls.SetDB(db)
	slog.Info("connected database success")

	droppedTableCount, err := resetAllTables(db, cfg.DB.Type)
	if err != nil {
		return fmt.Errorf("reset all tables failed: %w", err)
	}
	slog.Info("reset all tables success", slog.Int("droppedTableCount", droppedTableCount))

	if err := bootstrap.InitMigrations(); err != nil {
		return fmt.Errorf("run migrations failed: %w", err)
	}
	slog.Info("run migrations success")

	aiConfigResult, err := aiconfig.Init()
	if err != nil {
		return fmt.Errorf("init ai config failed: %w", err)
	}
	slog.Info("ai config init success", slog.Bool("skipped", aiConfigResult.Skipped),
		slog.String("filePath", aiConfigResult.FilePath),
		slog.Int("created", aiConfigResult.Created),
		slog.Int("updated", aiConfigResult.Updated))

	kbResult, err := kb.Init()
	if err != nil {
		return fmt.Errorf("init knowledge base failed: %w", err)
	}
	slog.Info("knowledge base init success", slog.Int("created", kbResult.CreatedDocuments))

	agentTeamResult, err := agentteam.Init()
	if err != nil {
		return fmt.Errorf("init agent team failed: %w", err)
	}
	slog.Info("agent team init success", slog.Bool("teamCreated", agentTeamResult.TeamCreated),
		slog.Int("usersCreated", agentTeamResult.UsersCreated),
		slog.Int("profilesCreated", agentTeamResult.ProfilesCreated),
		slog.Int("updatesApplied", agentTeamResult.UpdatesApplied),
	)

	aiAgentResult, err := aiagent.Init()
	if err != nil {
		return fmt.Errorf("init ai agent failed: %w", err)
	}
	slog.Info("ai agent init success", slog.Int("created", aiAgentResult.Created), slog.Int("updated", aiAgentResult.Updated))

	widgetSiteResult, err := widgetsite.Init()
	if err != nil {
		return fmt.Errorf("init widget site failed: %w", err)
	}
	slog.Info("widget site init success", slog.Int("created", widgetSiteResult.Created), slog.Int("updated", widgetSiteResult.Updated))

	slog.Info("testdata initialization completed")
	return nil
}

func withSilentSQLLogger(db *gorm.DB) *gorm.DB {
	if db == nil {
		return nil
	}
	return db.Session(&gorm.Session{Logger: db.Logger.LogMode(gormlogger.Silent)})
}

func confirmDestructiveAction(autoConfirm bool) error {
	if autoConfirm {
		return nil
	}

	fmt.Println("警告：该操作会清空当前数据库中的所有表和数据。")
	fmt.Print("请输入 INIT 继续，输入其他任意内容取消：")

	var input string
	if _, err := fmt.Scanln(&input); err != nil {
		return fmt.Errorf("read confirmation failed: %w", err)
	}

	if strings.TrimSpace(input) != "INIT" {
		return fmt.Errorf("initialization cancelled")
	}
	return nil
}

func resetAllTables(db *gorm.DB, dbType string) (int, error) {
	tables, err := db.Migrator().GetTables()
	if err != nil {
		return 0, err
	}

	filterSystemTables := func(tables []string) []string {
		ret := make([]string, 0, len(tables))
		for _, table := range tables {
			if strings.HasPrefix(table, "sqlite_") {
				continue
			}
			ret = append(ret, table)
		}
		return ret
	}

	filtered := filterSystemTables(tables)
	if len(filtered) == 0 {
		return 0, nil
	}

	err = withForeignKeyChecksDisabled(db, dbType, func() error {
		items := make([]any, 0, len(filtered))
		for _, table := range filtered {
			items = append(items, table)
		}
		return db.Migrator().DropTable(items...)
	})
	if err != nil {
		return 0, err
	}

	return len(filtered), nil
}

func withForeignKeyChecksDisabled(db *gorm.DB, dbType string, fn func() error) error {
	var disableSQL string
	var enableSQL string

	switch dbType {
	case "sqlite":
		disableSQL = "PRAGMA foreign_keys = OFF"
		enableSQL = "PRAGMA foreign_keys = ON"
	case "mysql":
		disableSQL = "SET FOREIGN_KEY_CHECKS = 0"
		enableSQL = "SET FOREIGN_KEY_CHECKS = 1"
	default:
		return fn()
	}

	if err := db.Exec(disableSQL).Error; err != nil {
		return err
	}
	defer func() {
		_ = db.Exec(enableSQL).Error
	}()

	return fn()
}
