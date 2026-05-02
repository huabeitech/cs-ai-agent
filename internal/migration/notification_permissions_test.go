package migration

import (
	"testing"
	"time"

	"cs-agent/internal/models"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func TestNotificationPermissionMigrationRegistered(t *testing.T) {
	migration, ok := migrationFuncs[7]
	if !ok {
		t.Fatalf("expected migration version 7 to be registered")
	}
	if migration.Remark != "sync notification permissions" {
		t.Fatalf("unexpected migration remark: %q", migration.Remark)
	}
}

func TestLightweightTicketPermissionMigrationRegistered(t *testing.T) {
	migration, ok := migrationFuncs[8]
	if !ok {
		t.Fatalf("expected migration version 8 to be registered")
	}
	if migration.Remark != "sync lightweight ticket permissions and reset ticket data" {
		t.Fatalf("unexpected migration remark: %q", migration.Remark)
	}
}

func TestLightweightTicketMigrationResetDeletesTicketData(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "t_",
			SingularTable: true,
		},
	})
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get sqlite db: %v", err)
	}
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	if err := db.AutoMigrate(&models.Ticket{}, &models.TicketTag{}, &models.TicketProgress{}, &models.TicketNoSequence{}); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}

	now := time.Now()
	ticket := &models.Ticket{
		TicketNo:    "TK2026050200001",
		Title:       "legacy ticket",
		Description: "legacy ticket description",
		AuditFields: models.AuditFields{CreatedAt: now, UpdatedAt: now},
	}
	if err := db.Create(ticket).Error; err != nil {
		t.Fatalf("create ticket error = %v", err)
	}
	if err := db.Create(&models.TicketTag{TicketID: ticket.ID, TagID: 1, AuditFields: models.AuditFields{CreatedAt: now, UpdatedAt: now}}).Error; err != nil {
		t.Fatalf("create ticket tag error = %v", err)
	}
	if err := db.Create(&models.TicketProgress{TicketID: ticket.ID, Content: "legacy progress", CreatedAt: now}).Error; err != nil {
		t.Fatalf("create ticket progress error = %v", err)
	}
	if err := db.Create(&models.TicketNoSequence{DateKey: "20260502", NextSeq: 2, CreatedAt: now, UpdatedAt: now}).Error; err != nil {
		t.Fatalf("create ticket no sequence error = %v", err)
	}

	if err := resetLightweightTicketData(db); err != nil {
		t.Fatalf("resetLightweightTicketData() error = %v", err)
	}

	assertTableCount(t, db, &models.Ticket{}, 0)
	assertTableCount(t, db, &models.TicketTag{}, 0)
	assertTableCount(t, db, &models.TicketProgress{}, 0)
	assertTableCount(t, db, &models.TicketNoSequence{}, 0)
}

func TestLightweightTicketMigrationDeletesObsoletePermissions(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "t_",
			SingularTable: true,
		},
	})
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get sqlite db: %v", err)
	}
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	if err := db.AutoMigrate(&models.Permission{}, &models.RolePermission{}, &models.UserPermission{}); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}

	now := time.Now()
	legacyPermission := &models.Permission{
		Name:        "回复工单",
		Code:        "ticket.reply",
		Type:        "api",
		GroupName:   "ticket",
		AuditFields: models.AuditFields{CreatedAt: now, UpdatedAt: now},
	}
	currentPermission := &models.Permission{
		Name:        "查看工单",
		Code:        "ticket.view",
		Type:        "api",
		GroupName:   "ticket",
		AuditFields: models.AuditFields{CreatedAt: now, UpdatedAt: now},
	}
	if err := db.Create(legacyPermission).Error; err != nil {
		t.Fatalf("create legacy permission error = %v", err)
	}
	if err := db.Create(currentPermission).Error; err != nil {
		t.Fatalf("create current permission error = %v", err)
	}
	if err := db.Create(&models.RolePermission{RoleID: 1, PermissionID: legacyPermission.ID, AuditFields: models.AuditFields{CreatedAt: now, UpdatedAt: now}}).Error; err != nil {
		t.Fatalf("create legacy role permission error = %v", err)
	}
	if err := db.Create(&models.UserPermission{UserID: 1, PermissionID: legacyPermission.ID, Effect: 1, AuditFields: models.AuditFields{CreatedAt: now, UpdatedAt: now}}).Error; err != nil {
		t.Fatalf("create legacy user permission error = %v", err)
	}
	if err := db.Create(&models.RolePermission{RoleID: 1, PermissionID: currentPermission.ID, AuditFields: models.AuditFields{CreatedAt: now, UpdatedAt: now}}).Error; err != nil {
		t.Fatalf("create current role permission error = %v", err)
	}

	if err := deleteObsoleteTicketPermissions(db); err != nil {
		t.Fatalf("deleteObsoleteTicketPermissions() error = %v", err)
	}

	assertPermissionCodeCount(t, db, "ticket.reply", 0)
	assertPermissionCodeCount(t, db, "ticket.view", 1)
	assertPermissionRelationCount(t, db, &models.RolePermission{}, legacyPermission.ID, 0)
	assertPermissionRelationCount(t, db, &models.UserPermission{}, legacyPermission.ID, 0)
	assertPermissionRelationCount(t, db, &models.RolePermission{}, currentPermission.ID, 1)
}

func assertTableCount(t *testing.T, db *gorm.DB, model any, expected int64) {
	t.Helper()

	var count int64
	if err := db.Model(model).Count(&count).Error; err != nil {
		t.Fatalf("count %T error = %v", model, err)
	}
	if count != expected {
		t.Fatalf("expected %T count %d, got %d", model, expected, count)
	}
}

func assertPermissionCodeCount(t *testing.T, db *gorm.DB, code string, expected int64) {
	t.Helper()

	var count int64
	if err := db.Model(&models.Permission{}).Where("code = ?", code).Count(&count).Error; err != nil {
		t.Fatalf("count permission %s error = %v", code, err)
	}
	if count != expected {
		t.Fatalf("expected permission %s count %d, got %d", code, expected, count)
	}
}

func assertPermissionRelationCount(t *testing.T, db *gorm.DB, model any, permissionID int64, expected int64) {
	t.Helper()

	var count int64
	if err := db.Model(model).Where("permission_id = ?", permissionID).Count(&count).Error; err != nil {
		t.Fatalf("count %T permission %d error = %v", model, permissionID, err)
	}
	if count != expected {
		t.Fatalf("expected %T permission %d count %d, got %d", model, permissionID, expected, count)
	}
}
