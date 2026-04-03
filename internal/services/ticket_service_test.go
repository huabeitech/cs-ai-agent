package services

import (
	"fmt"
	"testing"
	"time"

	"cs-agent/internal/models"
	"cs-agent/internal/pkg/dto"
	"cs-agent/internal/pkg/dto/request"
	"cs-agent/internal/pkg/enums"

	"github.com/mlogclub/simple/sqls"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

func TestCreateTicketSetsDeadlinesAndUniqueTicketNos(t *testing.T) {
	operator := setupTicketServiceTestDB(t)
	teamID, assigneeID := seedTicketAgent(t)

	first, err := TicketService.CreateTicket(buildCreateTicketRequest(teamID, assigneeID), operator)
	if err != nil {
		t.Fatalf("CreateTicket() error = %v", err)
	}
	second, err := TicketService.CreateTicket(buildCreateTicketRequest(teamID, assigneeID), operator)
	if err != nil {
		t.Fatalf("CreateTicket() second error = %v", err)
	}
	if first.TicketNo == second.TicketNo {
		t.Fatalf("expected unique ticket numbers, got %q", first.TicketNo)
	}
	if first.NextReplyDeadlineAt == nil {
		t.Fatalf("expected next reply deadline to be initialized")
	}
	if first.ResolveDeadlineAt == nil {
		t.Fatalf("expected resolve deadline to be initialized")
	}
	if first.Status != enums.TicketStatusOpen {
		t.Fatalf("expected ticket status open after assigned create, got %s", first.Status)
	}
}

func TestReplyTicketCompletesFirstResponseSLA(t *testing.T) {
	operator := setupTicketServiceTestDB(t)
	teamID, assigneeID := seedTicketAgent(t)
	ticket, err := TicketService.CreateTicket(buildCreateTicketRequest(teamID, assigneeID), operator)
	if err != nil {
		t.Fatalf("CreateTicket() error = %v", err)
	}

	if _, err := TicketService.ReplyTicket(requestReply(ticket.ID), operator); err != nil {
		t.Fatalf("ReplyTicket() error = %v", err)
	}

	updated := TicketService.Get(ticket.ID)
	if updated == nil {
		t.Fatalf("expected ticket to exist after reply")
	}
	if updated.FirstResponseAt == nil {
		t.Fatalf("expected first response time to be set")
	}
	if updated.NextReplyDeadlineAt != nil {
		t.Fatalf("expected next reply deadline to be cleared after first response")
	}
	firstResponseSLA := TicketSLARecordService.FindOne(
		sqls.NewCnd().
			Eq("ticket_id", ticket.ID).
			Eq("sla_type", enums.TicketSLATypeFirstResponse),
	)
	if firstResponseSLA == nil {
		t.Fatalf("expected first response SLA record")
	}
	if firstResponseSLA.Status != enums.TicketSLAStatusCompleted {
		t.Fatalf("expected first response SLA completed, got %s", firstResponseSLA.Status)
	}
}

func TestChangeStatusPauseAndResumeResolutionDeadline(t *testing.T) {
	operator := setupTicketServiceTestDB(t)
	teamID, assigneeID := seedTicketAgent(t)
	ticket, err := TicketService.CreateTicket(buildCreateTicketRequest(teamID, assigneeID), operator)
	if err != nil {
		t.Fatalf("CreateTicket() error = %v", err)
	}
	originalDeadline := ticket.ResolveDeadlineAt
	if originalDeadline == nil {
		t.Fatalf("expected initial resolve deadline")
	}

	if err := TicketService.ChangeStatus(dtoChangeStatus(ticket.ID, string(enums.TicketStatusPendingCustomer)), operator); err != nil {
		t.Fatalf("ChangeStatus(pending_customer) error = %v", err)
	}
	paused := TicketService.Get(ticket.ID)
	if paused == nil {
		t.Fatalf("expected paused ticket to exist")
	}
	if paused.ResolveDeadlineAt != nil {
		t.Fatalf("expected resolve deadline to be cleared while pending customer")
	}

	if err := TicketService.ChangeStatus(dtoChangeStatus(ticket.ID, string(enums.TicketStatusOpen)), operator); err != nil {
		t.Fatalf("ChangeStatus(open) error = %v", err)
	}
	resumed := TicketService.Get(ticket.ID)
	if resumed == nil {
		t.Fatalf("expected resumed ticket to exist")
	}
	if resumed.ResolveDeadlineAt == nil {
		t.Fatalf("expected resolve deadline to be recalculated on resume")
	}
	if !resumed.ResolveDeadlineAt.After(*originalDeadline) {
		t.Fatalf("expected resumed resolve deadline to move forward, original=%v resumed=%v", originalDeadline, resumed.ResolveDeadlineAt)
	}
}

func setupTicketServiceTestDB(t *testing.T) *dto.AuthPrincipal {
	t.Helper()

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "t_",
			SingularTable: true,
		},
	})
	if err != nil {
		t.Fatalf("gorm.Open() error = %v", err)
	}
	if err := db.AutoMigrate(models.Models...); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}
	sqls.SetDB(db)
	ticketNoSequence = 0
	return &dto.AuthPrincipal{
		UserID:   1,
		Username: "tester",
		Status:   enums.StatusOk,
	}
}

func seedTicketAgent(t *testing.T) (int64, int64) {
	t.Helper()

	now := time.Now()
	user := &models.User{
		Username: "agent_user",
		Nickname: "Agent",
		Status:   enums.StatusOk,
		AuditFields: models.AuditFields{
			CreatedAt:      now,
			CreateUserID:   1,
			CreateUserName: "tester",
			UpdatedAt:      now,
			UpdateUserID:   1,
			UpdateUserName: "tester",
		},
	}
	if err := sqls.DB().Create(user).Error; err != nil {
		t.Fatalf("create user error = %v", err)
	}
	team := &models.AgentTeam{
		Name:        "Support",
		Status:      enums.StatusOk,
		Description: "support team",
		AuditFields: user.AuditFields,
	}
	if err := sqls.DB().Create(team).Error; err != nil {
		t.Fatalf("create team error = %v", err)
	}
	profile := &models.AgentProfile{
		UserID:             user.ID,
		TeamID:             team.ID,
		AgentCode:          "A001",
		DisplayName:        "Agent One",
		ServiceStatus:      enums.ServiceStatusIdle,
		MaxConcurrentCount: 10,
		AutoAssignEnabled:  true,
		Status:             enums.StatusOk,
		AuditFields:        user.AuditFields,
	}
	if err := sqls.DB().Create(profile).Error; err != nil {
		t.Fatalf("create agent profile error = %v", err)
	}
	return team.ID, user.ID
}

func buildCreateTicketRequest(teamID, assigneeID int64) request.CreateTicketRequest {
	return request.CreateTicketRequest{
		Title:             "Ticket title",
		Description:       "Ticket description",
		Priority:          int(enums.TicketPriorityHigh),
		Severity:          int(enums.TicketSeverityMajor),
		CurrentTeamID:     teamID,
		CurrentAssigneeID: assigneeID,
	}
}

func requestReply(ticketID int64) request.ReplyTicketRequest {
	return request.ReplyTicketRequest{
		TicketID: ticketID,
		Content:  "reply content",
	}
}

func dtoChangeStatus(ticketID int64, status string) request.ChangeTicketStatusRequest {
	return request.ChangeTicketStatusRequest{
		TicketID:      ticketID,
		Status:        status,
		PendingReason: "waiting for customer",
		Reason:        "test transition",
	}
}
