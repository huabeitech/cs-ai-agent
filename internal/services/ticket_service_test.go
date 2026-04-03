package services_test

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"cs-agent/internal/bootstrap"
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/config"
	"cs-agent/internal/pkg/dto"
	"cs-agent/internal/pkg/dto/request"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/repositories"
	"cs-agent/internal/services"

	"github.com/mlogclub/simple/sqls"
)

func TestCreateTicketSetsTicketNoAndDeadlines(t *testing.T) {
	setupTicketTestDB(t)
	operator := &dto.AuthPrincipal{UserID: 1, Username: "admin"}

	first, err := services.TicketService.CreateTicket(createTestTicketRequest("ticket-1"), operator)
	if err != nil {
		t.Fatalf("CreateTicket() first error = %v", err)
	}
	second, err := services.TicketService.CreateTicket(createTestTicketRequest("ticket-2"), operator)
	if err != nil {
		t.Fatalf("CreateTicket() second error = %v", err)
	}

	if first.TicketNo == "" || second.TicketNo == "" {
		t.Fatalf("expected ticket numbers to be generated, got %q and %q", first.TicketNo, second.TicketNo)
	}
	if first.TicketNo == second.TicketNo {
		t.Fatalf("expected distinct ticket numbers, got %q", first.TicketNo)
	}
	if !strings.HasPrefix(first.TicketNo, "TK") {
		t.Fatalf("expected ticket number prefix TK, got %q", first.TicketNo)
	}

	detail := services.TicketService.Get(first.ID)
	if detail == nil {
		t.Fatalf("expected ticket to exist")
	}
	if detail.NextReplyDeadlineAt == nil {
		t.Fatalf("expected next reply deadline to be populated")
	}
	if detail.ResolveDeadlineAt == nil {
		t.Fatalf("expected resolve deadline to be populated")
	}

	slaList := services.TicketSLARecordService.Find(sqls.NewCnd().Eq("ticket_id", first.ID))
	if len(slaList) != 2 {
		t.Fatalf("expected 2 SLA records, got %d", len(slaList))
	}
}

func TestAddInternalNoteAllowsMentionSameUserAcrossTickets(t *testing.T) {
	setupTicketTestDB(t)
	operator := &dto.AuthPrincipal{UserID: 1, Username: "admin"}
	mentionedUserID := createTestUser(t, "mentioned")

	first, err := services.TicketService.CreateTicket(createTestTicketRequest("note-ticket-1"), operator)
	if err != nil {
		t.Fatalf("CreateTicket() first error = %v", err)
	}
	second, err := services.TicketService.CreateTicket(createTestTicketRequest("note-ticket-2"), operator)
	if err != nil {
		t.Fatalf("CreateTicket() second error = %v", err)
	}

	payload := fmt.Sprintf(`{"mentionUserIds":[%d]}`, mentionedUserID)
	if _, err := services.TicketService.AddInternalNote(requestInternalNote(first.ID, payload), operator); err != nil {
		t.Fatalf("AddInternalNote() first error = %v", err)
	}
	if _, err := services.TicketService.AddInternalNote(requestInternalNote(second.ID, payload), operator); err != nil {
		t.Fatalf("AddInternalNote() second error = %v", err)
	}

	mentions := services.TicketMentionService.Find(sqls.NewCnd().Eq("mentioned_user_id", mentionedUserID).Asc("id"))
	if len(mentions) != 2 {
		t.Fatalf("expected 2 mention records, got %d", len(mentions))
	}
}

func TestBatchChangeStatusRollsBackOnFailure(t *testing.T) {
	setupTicketTestDB(t)
	operator := &dto.AuthPrincipal{UserID: 1, Username: "admin"}

	first, err := services.TicketService.CreateTicket(createTestTicketRequest("batch-ticket"), operator)
	if err != nil {
		t.Fatalf("CreateTicket() error = %v", err)
	}

	err = services.TicketService.BatchChangeStatus(request.BatchChangeTicketStatusRequest{
		TicketIDs: []int64{first.ID, 999999},
		Status:    string(enums.TicketStatusOpen),
		Reason:    "batch open",
	}, operator)
	if err == nil {
		t.Fatalf("expected batch change status to fail")
	}

	current := services.TicketService.Get(first.ID)
	if current == nil {
		t.Fatalf("expected ticket to exist")
	}
	if current.Status != enums.TicketStatusNew {
		t.Fatalf("expected ticket status rollback to new, got %s", current.Status)
	}
}

func setupTicketTestDB(t *testing.T) {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "ticket-test.db")
	db, err := bootstrap.InitDB(config.DBConfig{
		Type: "sqlite",
		DSN:  dbPath,
	})
	if err != nil {
		t.Fatalf("InitDB() error = %v", err)
	}
	t.Cleanup(func() {
		sqlDB, err := db.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	})
	if err := bootstrap.InitMigrations(); err != nil {
		t.Fatalf("InitMigrations() error = %v", err)
	}
}

func createTestTicketRequest(title string) request.CreateTicketRequest {
	return request.CreateTicketRequest{
		Title:    title,
		Priority: int(enums.TicketPriorityNormal),
		Severity: int(enums.TicketSeverityMinor),
	}
}

func requestInternalNote(ticketID int64, payload string) request.InternalNoteRequest {
	return request.InternalNoteRequest{
		TicketID:    ticketID,
		ContentType: "text",
		Content:     "note",
		Payload:     payload,
	}
}

func createTestUser(t *testing.T, prefix string) int64 {
	t.Helper()
	return createTestUserWithID(t, 0, prefix)
}

func createTestUserWithID(t *testing.T, id int64, prefix string) int64 {
	t.Helper()
	now := time.Now()
	username := fmt.Sprintf("%s_%d", prefix, now.UnixNano())
	user := &models.User{
		ID:       id,
		Username: username,
		Nickname: prefix,
		Status:   enums.StatusOk,
		AuditFields: models.AuditFields{
			CreatedAt:      now,
			CreateUserID:   1,
			CreateUserName: "admin",
			UpdatedAt:      now,
			UpdateUserID:   1,
			UpdateUserName: "admin",
		},
	}
	if err := repositories.UserRepository.Create(sqls.DB(), user); err != nil {
		t.Fatalf("create user error = %v", err)
	}
	return user.ID
}
