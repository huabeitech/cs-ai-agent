package services_test

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"cs-agent/internal/bootstrap"
	"cs-agent/internal/events"
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/config"
	"cs-agent/internal/pkg/dto"
	"cs-agent/internal/pkg/dto/request"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/pkg/eventbus"
	"cs-agent/internal/repositories"
	"cs-agent/internal/services"

	"github.com/mlogclub/simple/sqls"
)

func TestTicketLightweightStatuses(t *testing.T) {
	if !enums.IsValidTicketStatus(string(enums.TicketStatusPending)) {
		t.Fatalf("pending should be valid")
	}
	if !enums.IsValidTicketStatus(string(enums.TicketStatusInProgress)) {
		t.Fatalf("in_progress should be valid")
	}
	if !enums.IsValidTicketStatus(string(enums.TicketStatusDone)) {
		t.Fatalf("done should be valid")
	}
	for _, status := range []string{"new", "open", "pending_customer", "pending_internal", "resolved", "closed", "cancelled"} {
		if enums.IsValidTicketStatus(status) {
			t.Fatalf("legacy status %s should be invalid", status)
		}
	}
}

func TestTicketProgressModelExists(t *testing.T) {
	item := models.TicketProgress{
		TicketID: 12,
		Content:  "已电话联系客户确认问题仍存在",
		AuthorID: 7,
	}
	if item.TicketID != 12 || item.AuthorID != 7 || item.Content == "" {
		t.Fatalf("unexpected progress model: %+v", item)
	}
}

func TestTicketServiceCreateTicketSetsPendingStatusAndTicketNo(t *testing.T) {
	setupTicketTestDB(t)
	operator := createTestOperator(t, "creator")
	customerID := createTestCustomer(t, "create-customer")
	tagID := createTestTag(t, "create-tag")

	created, err := services.TicketService.CreateTicket(request.CreateTicketRequest{
		Title:             "create ticket",
		Description:       "create ticket description",
		CustomerID:        customerID,
		TagIDs:            []int64{tagID},
		CurrentAssigneeID: operator.UserID,
	}, operator)
	if err != nil {
		t.Fatalf("CreateTicket() error = %v", err)
	}
	if created.TicketNo == "" || !strings.HasPrefix(created.TicketNo, "TK") {
		t.Fatalf("expected generated ticket number, got %q", created.TicketNo)
	}
	if created.Status != enums.TicketStatusPending {
		t.Fatalf("expected pending status, got %s", created.Status)
	}
	if created.Source != enums.TicketSourceManual {
		t.Fatalf("expected manual source, got %s", created.Source)
	}

	progresses := services.TicketProgressService.Find(sqls.NewCnd().Eq("ticket_id", created.ID))
	if len(progresses) != 1 {
		t.Fatalf("expected initial progress, got %d", len(progresses))
	}
	if progresses[0].Content != "创建工单" || progresses[0].AuthorID != operator.UserID {
		t.Fatalf("unexpected initial progress: %+v", progresses[0])
	}

	tags := services.TicketService.GetTags(created.ID)
	if len(tags) != 1 || tags[0].ID != tagID {
		t.Fatalf("expected ticket tag %d, got %+v", tagID, tags)
	}
}

func TestTicketServiceCreateTicketPublishesTicketCreatedEvent(t *testing.T) {
	setupTicketTestDB(t)
	operator := createTestOperator(t, "event-creator")
	eventsCh := make(chan events.TicketCreatedEvent, 1)
	_, unsubscribe := eventbus.Subscribe(func(ctx context.Context, event events.TicketCreatedEvent) error {
		eventsCh <- event
		return nil
	})
	defer unsubscribe()

	created, err := services.TicketService.CreateTicket(createTestTicketRequest("event-ticket"), operator)
	if err != nil {
		t.Fatalf("CreateTicket() error = %v", err)
	}

	select {
	case event := <-eventsCh:
		if event.TicketID != created.ID {
			t.Fatalf("expected ticket id %d, got %d", created.ID, event.TicketID)
		}
		if event.OperatorID != operator.UserID {
			t.Fatalf("expected operator id %d, got %d", operator.UserID, event.OperatorID)
		}
	case <-time.After(time.Second):
		t.Fatalf("expected ticket created event")
	}
}

func TestTicketServiceChangeStatusSetsHandledAt(t *testing.T) {
	setupTicketTestDB(t)
	operator := createTestOperator(t, "status-operator")
	ticket, err := services.TicketService.CreateTicket(createTestTicketRequest("status-ticket"), operator)
	if err != nil {
		t.Fatalf("CreateTicket() error = %v", err)
	}

	if err := services.TicketService.ChangeStatus(request.ChangeTicketStatusRequest{
		TicketID: ticket.ID,
		Status:   string(enums.TicketStatusInProgress),
	}, operator); err != nil {
		t.Fatalf("ChangeStatus() in_progress error = %v", err)
	}
	inProgress := services.TicketService.Get(ticket.ID)
	if inProgress == nil {
		t.Fatalf("expected ticket to exist")
	}
	if inProgress.Status != enums.TicketStatusInProgress {
		t.Fatalf("expected in_progress status, got %s", inProgress.Status)
	}
	if inProgress.HandledAt != nil {
		t.Fatalf("expected handled_at to remain nil before done")
	}

	if err := services.TicketService.ChangeStatus(request.ChangeTicketStatusRequest{
		TicketID: ticket.ID,
		Status:   string(enums.TicketStatusDone),
	}, operator); err != nil {
		t.Fatalf("ChangeStatus() done error = %v", err)
	}
	done := services.TicketService.Get(ticket.ID)
	if done == nil || done.HandledAt == nil {
		t.Fatalf("expected handled_at to be set after done, got %+v", done)
	}

	if err := services.TicketService.ChangeStatus(request.ChangeTicketStatusRequest{
		TicketID: ticket.ID,
		Status:   string(enums.TicketStatusPending),
	}, operator); err != nil {
		t.Fatalf("ChangeStatus() pending error = %v", err)
	}
	pending := services.TicketService.Get(ticket.ID)
	if pending == nil || pending.HandledAt != nil {
		t.Fatalf("expected handled_at to be cleared away from done, got %+v", pending)
	}
}

func TestTicketServiceAddProgressStoresContentAndAuthor(t *testing.T) {
	setupTicketTestDB(t)
	operator := createTestOperator(t, "progress-operator")
	ticket, err := services.TicketService.CreateTicket(createTestTicketRequest("progress-ticket"), operator)
	if err != nil {
		t.Fatalf("CreateTicket() error = %v", err)
	}

	progress, err := services.TicketService.AddProgress(request.CreateTicketProgressRequest{
		TicketID: ticket.ID,
		Content:  "客户已确认问题复现路径",
	}, operator)
	if err != nil {
		t.Fatalf("AddProgress() error = %v", err)
	}
	if progress.ID <= 0 {
		t.Fatalf("expected progress id")
	}
	if progress.Content != "客户已确认问题复现路径" || progress.AuthorID != operator.UserID {
		t.Fatalf("unexpected progress: %+v", progress)
	}
}

func TestTicketServiceAssignTicketRequiresTargetUser(t *testing.T) {
	setupTicketTestDB(t)
	operator := createTestOperator(t, "assign-operator")
	ticket, err := services.TicketService.CreateTicket(createTestTicketRequest("assign-ticket"), operator)
	if err != nil {
		t.Fatalf("CreateTicket() error = %v", err)
	}

	err = services.TicketService.AssignTicket(request.AssignTicketRequest{
		TicketID: ticket.ID,
		ToUserID: 0,
		Reason:   "invalid assignment",
	}, operator)
	if err == nil {
		t.Fatalf("expected AssignTicket() to reject empty target user")
	}
}

func TestTicketServiceAssignTicketRejectsDisabledUser(t *testing.T) {
	setupTicketTestDB(t)
	operator := createTestOperator(t, "assign-disabled-operator")
	disabledUserID := createTestUserWithStatus(t, "assign-disabled-user", enums.StatusDisabled)
	ticket, err := services.TicketService.CreateTicket(createTestTicketRequest("assign-disabled-ticket"), operator)
	if err != nil {
		t.Fatalf("CreateTicket() error = %v", err)
	}

	err = services.TicketService.AssignTicket(request.AssignTicketRequest{
		TicketID: ticket.ID,
		ToUserID: disabledUserID,
		Reason:   "disabled assignment",
	}, operator)
	if err == nil {
		t.Fatalf("expected AssignTicket() to reject disabled target user")
	}
}

func TestTicketServiceCreateTicketRejectsMismatchedCustomerConversation(t *testing.T) {
	setupTicketTestDB(t)
	operator := createTestOperator(t, "mismatch-operator")
	customerID := createTestCustomer(t, "mismatch-customer")
	otherCustomerID := createTestCustomer(t, "mismatch-other-customer")
	conversationID := createTestConversation(t, otherCustomerID, "mismatch-conversation")

	_, err := services.TicketService.CreateTicket(request.CreateTicketRequest{
		Title:          "mismatch ticket",
		Description:    "mismatch ticket description",
		CustomerID:     customerID,
		ConversationID: conversationID,
	}, operator)
	if err == nil {
		t.Fatalf("expected CreateTicket() to reject mismatched customer and conversation")
	}
}

func TestTicketServiceSummaryCountsStaleTickets(t *testing.T) {
	setupTicketTestDB(t)
	operator := createTestOperator(t, "summary-operator")
	mine, err := services.TicketService.CreateTicket(request.CreateTicketRequest{
		Title:             "mine stale ticket",
		Description:       "mine stale description",
		CurrentAssigneeID: operator.UserID,
	}, operator)
	if err != nil {
		t.Fatalf("CreateTicket() mine error = %v", err)
	}
	if _, err := services.TicketService.CreateTicket(createTestTicketRequest("unassigned ticket"), operator); err != nil {
		t.Fatalf("CreateTicket() unassigned error = %v", err)
	}
	staleUpdatedAt := time.Now().Add(-48 * time.Hour)
	if err := repositories.TicketRepository.Updates(sqls.DB(), mine.ID, map[string]any{
		"updated_at": staleUpdatedAt,
	}); err != nil {
		t.Fatalf("update stale ticket error = %v", err)
	}

	summary := services.TicketService.GetSummary(operator, 24)
	if summary.All != 2 {
		t.Fatalf("expected all count 2, got %d", summary.All)
	}
	if summary.Pending != 2 {
		t.Fatalf("expected pending count 2, got %d", summary.Pending)
	}
	if summary.Mine != 1 {
		t.Fatalf("expected mine count 1, got %d", summary.Mine)
	}
	if summary.Unassigned != 1 {
		t.Fatalf("expected unassigned count 1, got %d", summary.Unassigned)
	}
	if summary.Stale != 1 {
		t.Fatalf("expected stale count 1, got %d", summary.Stale)
	}
}

func TestTicketServiceFindPageAggregateFiltersStaleTickets(t *testing.T) {
	setupTicketTestDB(t)
	operator := createTestOperator(t, "stale-list-operator")
	staleOpen, err := services.TicketService.CreateTicket(createTestTicketRequest("stale open ticket"), operator)
	if err != nil {
		t.Fatalf("CreateTicket() stale open error = %v", err)
	}
	staleDone, err := services.TicketService.CreateTicket(createTestTicketRequest("stale done ticket"), operator)
	if err != nil {
		t.Fatalf("CreateTicket() stale done error = %v", err)
	}
	freshOpen, err := services.TicketService.CreateTicket(createTestTicketRequest("fresh open ticket"), operator)
	if err != nil {
		t.Fatalf("CreateTicket() fresh open error = %v", err)
	}

	if err := services.TicketService.ChangeStatus(request.ChangeTicketStatusRequest{
		TicketID: staleDone.ID,
		Status:   string(enums.TicketStatusDone),
	}, operator); err != nil {
		t.Fatalf("ChangeStatus() stale done error = %v", err)
	}
	staleUpdatedAt := time.Now().Add(-48 * time.Hour)
	for _, ticketID := range []int64{staleOpen.ID, staleDone.ID} {
		if err := repositories.TicketRepository.Updates(sqls.DB(), ticketID, map[string]any{
			"updated_at": staleUpdatedAt,
		}); err != nil {
			t.Fatalf("update stale ticket %d error = %v", ticketID, err)
		}
	}

	aggregate, err := services.TicketService.FindPageAggregateByCnd(
		sqls.NewCnd().
			NotEq("status", enums.TicketStatusDone).
			Where("updated_at < ?", time.Now().Add(-24*time.Hour)).
			Page(1, 10),
		operator.UserID,
	)
	if err != nil {
		t.Fatalf("FindPageAggregateByCnd() error = %v", err)
	}
	if len(aggregate.List) != 1 {
		t.Fatalf("expected 1 stale non-done ticket, got %d: %+v", len(aggregate.List), aggregate.List)
	}
	if aggregate.List[0].ID != staleOpen.ID {
		t.Fatalf("expected stale open ticket %d, got %d", staleOpen.ID, aggregate.List[0].ID)
	}
	if aggregate.List[0].ID == freshOpen.ID || aggregate.List[0].ID == staleDone.ID {
		t.Fatalf("stale list included fresh or done ticket: %+v", aggregate.List[0])
	}
}

func TestTicketServiceFindPageAggregateEnrichesLookups(t *testing.T) {
	setupTicketTestDB(t)
	operator := createTestOperator(t, "aggregate-operator")
	assignee := createTestOperator(t, "aggregate-assignee")
	customerID := createTestCustomer(t, "aggregate-customer")
	tagID := createTestTag(t, "aggregate-tag")

	ticket, err := services.TicketService.CreateTicket(request.CreateTicketRequest{
		Title:             "aggregate ticket",
		Description:       "aggregate description",
		CustomerID:        customerID,
		TagIDs:            []int64{tagID},
		CurrentAssigneeID: assignee.UserID,
	}, operator)
	if err != nil {
		t.Fatalf("CreateTicket() error = %v", err)
	}

	aggregate, err := services.TicketService.FindPageAggregateByCnd(sqls.NewCnd().Eq("id", ticket.ID).Page(1, 10), operator.UserID)
	if err != nil {
		t.Fatalf("FindPageAggregateByCnd() error = %v", err)
	}
	if len(aggregate.List) != 1 {
		t.Fatalf("expected 1 ticket, got %d", len(aggregate.List))
	}
	if len(aggregate.TagsByTicketID[ticket.ID]) != 1 || aggregate.TagsByTicketID[ticket.ID][0].ID != tagID {
		t.Fatalf("expected tag lookup to be populated")
	}
	if aggregate.Customers[customerID] == nil {
		t.Fatalf("expected customer lookup to be populated")
	}
	if aggregate.Users[assignee.UserID] == nil {
		t.Fatalf("expected assignee lookup to be populated")
	}
}

func TestTicketServiceTicketNoNextConcurrent(t *testing.T) {
	setupTicketTestDBWithMaxOpenConns(t, 8)

	const count = 50
	results := make(chan string, count)
	errs := make(chan error, count)
	var wg sync.WaitGroup

	for i := 0; i < count; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := sqls.WithTransaction(func(ctx *sqls.TxContext) error {
				ticketNo, err := services.TicketNoService.Next(ctx.Tx, time.Now())
				if err != nil {
					return err
				}
				results <- ticketNo
				return nil
			})
			if err != nil {
				errs <- err
			}
		}()
	}

	wg.Wait()
	close(results)
	close(errs)

	for err := range errs {
		if err != nil {
			t.Fatalf("TicketNoService.Next() concurrent error = %v", err)
		}
	}

	seen := make(map[string]struct{}, count)
	for ticketNo := range results {
		if _, ok := seen[ticketNo]; ok {
			t.Fatalf("duplicate ticket number generated: %s", ticketNo)
		}
		seen[ticketNo] = struct{}{}
	}
	if len(seen) != count {
		t.Fatalf("expected %d unique ticket numbers, got %d", count, len(seen))
	}
}

func TestTicketServiceCreateTicketConcurrentAllocatesUniqueTicketNos(t *testing.T) {
	setupTicketTestDBWithMaxOpenConns(t, 8)
	operator := createTestOperator(t, "concurrent-create-operator")

	const count = 50
	results := make(chan string, count)
	errs := make(chan error, count)
	var wg sync.WaitGroup

	for i := 0; i < count; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			ticket, err := services.TicketService.CreateTicket(request.CreateTicketRequest{
				Title:       fmt.Sprintf("concurrent ticket %d", index),
				Description: fmt.Sprintf("concurrent ticket %d description", index),
			}, operator)
			if err != nil {
				errs <- err
				return
			}
			results <- ticket.TicketNo
		}(i)
	}

	wg.Wait()
	close(results)
	close(errs)

	for err := range errs {
		if err != nil {
			t.Fatalf("CreateTicket() concurrent error = %v", err)
		}
	}

	seen := make(map[string]struct{}, count)
	for ticketNo := range results {
		if ticketNo == "" {
			t.Fatalf("expected non-empty ticket number")
		}
		if _, ok := seen[ticketNo]; ok {
			t.Fatalf("duplicate ticket number generated: %s", ticketNo)
		}
		seen[ticketNo] = struct{}{}
	}
	if len(seen) != count {
		t.Fatalf("expected %d unique ticket numbers, got %d", count, len(seen))
	}
}

func setupTicketTestDB(t *testing.T) {
	setupTicketTestDBWithMaxOpenConns(t, 0)
}

func setupTicketTestDBWithMaxOpenConns(t *testing.T, maxOpenConns int) {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "ticket-test.db")
	db, err := bootstrap.InitDB(config.DBConfig{
		Type:         "sqlite",
		DSN:          "file:" + dbPath + "?_busy_timeout=5000",
		MaxIdleConns: 1,
		MaxOpenConns: maxOpenConns,
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
		Title:       title,
		Description: title + " description",
	}
}

func createTestOperator(t *testing.T, prefix string) *dto.AuthPrincipal {
	t.Helper()
	userID := createTestUser(t, prefix)
	return &dto.AuthPrincipal{UserID: userID, Username: prefix}
}

func createTestUser(t *testing.T, prefix string) int64 {
	return createTestUserWithStatus(t, prefix, enums.StatusOk)
}

func createTestUserWithStatus(t *testing.T, prefix string, status enums.Status) int64 {
	t.Helper()
	now := time.Now()
	username := fmt.Sprintf("%s_%d", prefix, now.UnixNano())
	user := &models.User{
		Username: username,
		Nickname: prefix,
		Status:   status,
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

func createTestConversation(t *testing.T, customerID int64, prefix string) int64 {
	t.Helper()

	now := time.Now()
	item := &models.Conversation{
		CustomerID:    customerID,
		CustomerName:  prefix,
		Status:        enums.IMConversationStatusActive,
		ServiceMode:   enums.IMConversationServiceModeAIOnly,
		LastMessageAt: now,
		LastActiveAt:  now,
		AuditFields:   models.AuditFields{CreatedAt: now, UpdatedAt: now},
	}
	if err := repositories.ConversationRepository.Create(sqls.DB(), item); err != nil {
		t.Fatalf("create conversation error = %v", err)
	}
	return item.ID
}

func createTestCustomer(t *testing.T, prefix string) int64 {
	t.Helper()

	now := time.Now()
	item := &models.Customer{
		Name:   fmt.Sprintf("%s-%d", prefix, now.UnixNano()),
		Status: enums.StatusOk,
		AuditFields: models.AuditFields{
			CreatedAt:      now,
			CreateUserID:   1,
			CreateUserName: "admin",
			UpdatedAt:      now,
			UpdateUserID:   1,
			UpdateUserName: "admin",
		},
	}
	if err := repositories.CustomerRepository.Create(sqls.DB(), item); err != nil {
		t.Fatalf("create customer error = %v", err)
	}
	return item.ID
}

func createTestTag(t *testing.T, prefix string) int64 {
	t.Helper()

	now := time.Now()
	item := &models.Tag{
		Name:   fmt.Sprintf("%s-%d", prefix, now.UnixNano()),
		Status: enums.StatusOk,
		SortNo: 1,
		AuditFields: models.AuditFields{
			CreatedAt:      now,
			CreateUserID:   1,
			CreateUserName: "admin",
			UpdatedAt:      now,
			UpdateUserID:   1,
			UpdateUserName: "admin",
		},
	}
	if err := repositories.TagRepository.Create(sqls.DB(), item); err != nil {
		t.Fatalf("create tag error = %v", err)
	}
	return item.ID
}
