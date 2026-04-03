package services_test

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"cs-agent/internal/bootstrap"
	"cs-agent/internal/builders"
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

func TestAssignTicketPromotesNewTicketToOpenAndSetsTeamAssignee(t *testing.T) {
	setupTicketTestDB(t)
	operator := &dto.AuthPrincipal{UserID: 1, Username: "admin"}

	ticket, err := services.TicketService.CreateTicket(createTestTicketRequest("assign-ticket"), operator)
	if err != nil {
		t.Fatalf("CreateTicket() error = %v", err)
	}
	teamID, assigneeID := createTestAgentProfile(t, "assignee")

	if err := services.TicketService.AssignTicket(request.AssignTicketRequest{
		TicketID: ticket.ID,
		ToTeamID: teamID,
		ToUserID: assigneeID,
		Reason:   "manual assign",
	}, operator); err != nil {
		t.Fatalf("AssignTicket() error = %v", err)
	}

	current := services.TicketService.Get(ticket.ID)
	if current == nil {
		t.Fatalf("expected ticket to exist")
	}
	if current.Status != enums.TicketStatusOpen {
		t.Fatalf("expected assigned ticket status to be open, got %s", current.Status)
	}
	if current.CurrentTeamID != teamID {
		t.Fatalf("expected current team id %d, got %d", teamID, current.CurrentTeamID)
	}
	if current.CurrentAssigneeID != assigneeID {
		t.Fatalf("expected current assignee id %d, got %d", assigneeID, current.CurrentAssigneeID)
	}
}

func TestFindPageAggregateByCndBuildsWatcherAndLookupMaps(t *testing.T) {
	setupTicketTestDB(t)
	operator := &dto.AuthPrincipal{UserID: 1, Username: "admin"}
	teamID, assigneeID := createTestAgentProfile(t, "aggregate-agent")
	customerID := createTestCustomer(t, "aggregate-customer")
	categoryID := createTestTicketCategory(t, "aggregate-category")

	ticket, err := services.TicketService.CreateTicket(request.CreateTicketRequest{
		Title:             "aggregate-ticket",
		CustomerID:        customerID,
		CategoryID:        categoryID,
		Priority:          int(enums.TicketPriorityHigh),
		Severity:          int(enums.TicketSeverityMajor),
		CurrentTeamID:     teamID,
		CurrentAssigneeID: assigneeID,
	}, operator)
	if err != nil {
		t.Fatalf("CreateTicket() error = %v", err)
	}
	if err := repositories.TicketWatcherRepository.Create(sqls.DB(), &models.TicketWatcher{
		TicketID:  ticket.ID,
		UserID:    operator.UserID,
		CreatedAt: time.Now(),
	}); err != nil {
		t.Fatalf("create ticket watcher error = %v", err)
	}

	aggregate, err := services.TicketService.FindPageAggregateByCnd(
		sqls.NewCnd().Eq("id", ticket.ID).Page(1, 10),
		operator.UserID,
	)
	if err != nil {
		t.Fatalf("FindPageAggregateByCnd() error = %v", err)
	}
	if len(aggregate.List) != 1 {
		t.Fatalf("expected 1 ticket, got %d", len(aggregate.List))
	}
	if _, ok := aggregate.WatchedTicketIDs[ticket.ID]; !ok {
		t.Fatalf("expected watched ticket id to be populated")
	}
	if aggregate.Categories[categoryID] == nil {
		t.Fatalf("expected category lookup to be populated")
	}
	if aggregate.Customers[customerID] == nil {
		t.Fatalf("expected customer lookup to be populated")
	}
	if aggregate.Users[assigneeID] == nil {
		t.Fatalf("expected assignee lookup to be populated")
	}
	if aggregate.Teams[teamID] == nil {
		t.Fatalf("expected team lookup to be populated")
	}
	if len(aggregate.SLAByTicketID[ticket.ID]) != 2 {
		t.Fatalf("expected 2 sla records for ticket, got %d", len(aggregate.SLAByTicketID[ticket.ID]))
	}
}

func TestWatchTicketAffectsSummaryAndListFilter(t *testing.T) {
	setupTicketTestDB(t)
	operator := &dto.AuthPrincipal{UserID: createTestUser(t, "watch-operator"), Username: "watch-operator"}

	ticket, err := services.TicketService.CreateTicket(createTestTicketRequest("watch-ticket"), operator)
	if err != nil {
		t.Fatalf("CreateTicket() error = %v", err)
	}

	if err := services.TicketService.WatchTicket(ticket.ID, operator); err != nil {
		t.Fatalf("WatchTicket() error = %v", err)
	}

	summary := services.TicketService.GetSummary(operator)
	if summary.Watching != 1 {
		t.Fatalf("expected watching summary to be 1, got %d", summary.Watching)
	}

	aggregate, err := services.TicketService.FindPageAggregateByCnd(
		sqls.NewCnd().
			Where("id IN (SELECT ticket_id FROM t_ticket_watcher WHERE user_id = ?)", operator.UserID).
			Page(1, 10),
		operator.UserID,
	)
	if err != nil {
		t.Fatalf("FindPageAggregateByCnd() error = %v", err)
	}
	if len(aggregate.List) != 1 {
		t.Fatalf("expected 1 watched ticket, got %d", len(aggregate.List))
	}
	if aggregate.List[0].ID != ticket.ID {
		t.Fatalf("expected watched ticket id %d, got %d", ticket.ID, aggregate.List[0].ID)
	}
	if _, ok := aggregate.WatchedTicketIDs[ticket.ID]; !ok {
		t.Fatalf("expected watched ticket id to be marked in aggregate")
	}
}

func TestTicketNoServiceNextConcurrent(t *testing.T) {
	setupTicketTestDB(t)

	const count = 20
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

func TestGetRiskPageAggregateReturnsAccurateHighRiskTickets(t *testing.T) {
	setupTicketTestDB(t)
	operator := &dto.AuthPrincipal{UserID: 1, Username: "admin"}

	highRisk, err := services.TicketService.CreateTicket(createTestTicketRequest("high-risk-ticket"), operator)
	if err != nil {
		t.Fatalf("CreateTicket() highRisk error = %v", err)
	}
	safe, err := services.TicketService.CreateTicket(createTestTicketRequest("safe-ticket"), operator)
	if err != nil {
		t.Fatalf("CreateTicket() safe error = %v", err)
	}

	if err := services.TicketService.ChangeStatus(request.ChangeTicketStatusRequest{
		TicketID: highRisk.ID,
		Status:   string(enums.TicketStatusOpen),
		Reason:   "pick up high risk",
	}, operator); err != nil {
		t.Fatalf("ChangeStatus() highRisk open error = %v", err)
	}
	if err := services.TicketService.ChangeStatus(request.ChangeTicketStatusRequest{
		TicketID: safe.ID,
		Status:   string(enums.TicketStatusOpen),
		Reason:   "pick up safe",
	}, operator); err != nil {
		t.Fatalf("ChangeStatus() safe open error = %v", err)
	}

	now := time.Now()
	if err := repositories.TicketRepository.Updates(sqls.DB(), highRisk.ID, map[string]any{
		"resolve_deadline_at": now.Add(30 * time.Minute),
		"updated_at":          now,
	}); err != nil {
		t.Fatalf("update highRisk deadline error = %v", err)
	}
	if err := repositories.TicketRepository.Updates(sqls.DB(), safe.ID, map[string]any{
		"resolve_deadline_at": now.Add(6 * time.Hour),
		"updated_at":          now,
	}); err != nil {
		t.Fatalf("update safe deadline error = %v", err)
	}

	aggregate, err := services.TicketService.GetRiskPageAggregate("high_risk", 0, 60, 1, 10, operator.UserID)
	if err != nil {
		t.Fatalf("GetRiskPageAggregate() error = %v", err)
	}
	if len(aggregate.List) != 1 {
		t.Fatalf("expected 1 high risk ticket, got %d", len(aggregate.List))
	}
	if aggregate.List[0].ID != highRisk.ID {
		t.Fatalf("expected high risk ticket id %d, got %d", highRisk.ID, aggregate.List[0].ID)
	}
}

func TestBuildTicketDetailUsesAggregatedWatcherCollaboratorAndRelationLookups(t *testing.T) {
	setupTicketTestDB(t)
	operator := &dto.AuthPrincipal{UserID: 1, Username: "admin"}
	teamID, assigneeID := createTestAgentProfile(t, "detail-agent")

	parent, err := services.TicketService.CreateTicket(createTestTicketRequest("detail-parent"), operator)
	if err != nil {
		t.Fatalf("CreateTicket() parent error = %v", err)
	}
	child, err := services.TicketService.CreateTicket(request.CreateTicketRequest{
		Title:             "detail-child",
		Priority:          int(enums.TicketPriorityNormal),
		Severity:          int(enums.TicketSeverityMinor),
		CurrentTeamID:     teamID,
		CurrentAssigneeID: assigneeID,
	}, operator)
	if err != nil {
		t.Fatalf("CreateTicket() child error = %v", err)
	}
	if err := repositories.TicketWatcherRepository.Create(sqls.DB(), &models.TicketWatcher{
		TicketID:  parent.ID,
		UserID:    assigneeID,
		CreatedAt: time.Now(),
	}); err != nil {
		t.Fatalf("create watcher error = %v", err)
	}
	if err := repositories.TicketCollaboratorRepository.Create(sqls.DB(), &models.TicketCollaborator{
		TicketID:  parent.ID,
		UserID:    assigneeID,
		CreatedAt: time.Now(),
	}); err != nil {
		t.Fatalf("create collaborator error = %v", err)
	}
	if err := repositories.TicketRelationRepository.Create(sqls.DB(), &models.TicketRelation{
		TicketID:        parent.ID,
		RelatedTicketID: child.ID,
		RelationType:    enums.TicketRelationTypeChild,
		CreatedAt:       time.Now(),
	}); err != nil {
		t.Fatalf("create relation error = %v", err)
	}
	if err := repositories.TicketCommentRepository.Create(sqls.DB(), &models.TicketComment{
		TicketID:    parent.ID,
		CommentType: enums.TicketCommentTypePublicReply,
		AuthorType:  enums.IMSenderTypeAgent,
		AuthorID:    assigneeID,
		ContentType: "text",
		Content:     "reply",
		CreatedAt:   time.Now(),
	}); err != nil {
		t.Fatalf("create comment error = %v", err)
	}
	if err := repositories.TicketEventLogRepository.Create(sqls.DB(), &models.TicketEventLog{
		TicketID:     parent.ID,
		EventType:    enums.TicketEventTypeAssigned,
		OperatorType: enums.IMSenderTypeAgent,
		OperatorID:   assigneeID,
		Content:      "assigned",
		CreatedAt:    time.Now(),
	}); err != nil {
		t.Fatalf("create event error = %v", err)
	}

	aggregate, err := services.TicketService.GetDetail(parent.ID)
	if err != nil {
		t.Fatalf("GetDetail() error = %v", err)
	}
	detail := builders.BuildTicketDetail(aggregate)
	if detail == nil {
		t.Fatalf("expected ticket detail to be built")
	}
	if len(detail.Watchers) != 1 || detail.Watchers[0].UserName == "" {
		t.Fatalf("expected watcher user name to be populated")
	}
	if len(detail.Collaborators) != 1 || detail.Collaborators[0].UserName == "" || detail.Collaborators[0].TeamName == "" {
		t.Fatalf("expected collaborator user and team names to be populated")
	}
	if len(detail.RelatedTickets) != 1 {
		t.Fatalf("expected 1 related ticket, got %d", len(detail.RelatedTickets))
	}
	if detail.RelatedTickets[0].RelatedTicketNo == "" || detail.RelatedTickets[0].CurrentAssigneeName == "" || detail.RelatedTickets[0].CurrentTeamName == "" {
		t.Fatalf("expected related ticket display fields to be populated")
	}
	if len(detail.Comments) != 1 || detail.Comments[0].AuthorName == "" {
		t.Fatalf("expected comment author name to be populated")
	}
	if len(detail.Events) == 0 {
		t.Fatalf("expected events to be populated")
	}
	hasNamedEvent := false
	for i := range detail.Events {
		if detail.Events[i].OperatorName != "" {
			hasNamedEvent = true
			break
		}
	}
	if !hasNamedEvent {
		t.Fatalf("expected at least one event operator name to be populated")
	}
}

func TestCloseAndReopenTicketRefreshResolutionDeadline(t *testing.T) {
	setupTicketTestDB(t)
	operator := &dto.AuthPrincipal{UserID: 1, Username: "admin"}

	ticket, err := services.TicketService.CreateTicket(createTestTicketRequest("deadline-ticket"), operator)
	if err != nil {
		t.Fatalf("CreateTicket() error = %v", err)
	}

	original := services.TicketService.Get(ticket.ID)
	if original == nil || original.ResolveDeadlineAt == nil {
		t.Fatalf("expected initial resolve deadline to exist")
	}

	if err := services.TicketService.ChangeStatus(request.ChangeTicketStatusRequest{
		TicketID: ticket.ID,
		Status:   string(enums.TicketStatusOpen),
		Reason:   "pick up",
	}, operator); err != nil {
		t.Fatalf("ChangeStatus() open error = %v", err)
	}

	if err := services.TicketService.CloseTicket(request.CloseTicketRequest{
		TicketID:    ticket.ID,
		CloseReason: "done",
	}, operator); err != nil {
		t.Fatalf("CloseTicket() error = %v", err)
	}

	closed := services.TicketService.Get(ticket.ID)
	if closed == nil {
		t.Fatalf("expected closed ticket to exist")
	}
	if closed.ResolveDeadlineAt != nil {
		t.Fatalf("expected resolve deadline to be cleared after close")
	}

	if err := services.TicketService.ReopenTicket(request.ReopenTicketRequest{
		TicketID: ticket.ID,
		Reason:   "need follow-up",
	}, operator); err != nil {
		t.Fatalf("ReopenTicket() error = %v", err)
	}

	reopened := services.TicketService.Get(ticket.ID)
	if reopened == nil {
		t.Fatalf("expected reopened ticket to exist")
	}
	if reopened.ResolveDeadlineAt == nil {
		t.Fatalf("expected resolve deadline to be restored after reopen")
	}
}

func TestChangeStatusPendingCustomerThenOpenRefreshesSLAFields(t *testing.T) {
	setupTicketTestDB(t)
	operator := &dto.AuthPrincipal{UserID: 1, Username: "admin"}

	ticket, err := services.TicketService.CreateTicket(createTestTicketRequest("pending-ticket"), operator)
	if err != nil {
		t.Fatalf("CreateTicket() error = %v", err)
	}
	if err := services.TicketService.ChangeStatus(request.ChangeTicketStatusRequest{
		TicketID: ticket.ID,
		Status:   string(enums.TicketStatusOpen),
		Reason:   "pick up",
	}, operator); err != nil {
		t.Fatalf("ChangeStatus() open error = %v", err)
	}

	beforePending := services.TicketService.Get(ticket.ID)
	if beforePending == nil || beforePending.ResolveDeadlineAt == nil {
		t.Fatalf("expected resolve deadline before pending")
	}

	if err := services.TicketService.ChangeStatus(request.ChangeTicketStatusRequest{
		TicketID:      ticket.ID,
		Status:        string(enums.TicketStatusPendingCustomer),
		PendingReason: "waiting customer",
		Reason:        "pause for customer",
	}, operator); err != nil {
		t.Fatalf("ChangeStatus() pending error = %v", err)
	}

	pending := services.TicketService.Get(ticket.ID)
	if pending == nil {
		t.Fatalf("expected pending ticket to exist")
	}
	if pending.Status != enums.TicketStatusPendingCustomer {
		t.Fatalf("expected pending status, got %s", pending.Status)
	}
	if pending.PendingReason != "waiting customer" {
		t.Fatalf("expected pending reason to be persisted, got %q", pending.PendingReason)
	}
	if pending.ResolveDeadlineAt == nil {
		t.Fatalf("expected resolve deadline to remain calculable while pending")
	}

	if err := services.TicketService.ChangeStatus(request.ChangeTicketStatusRequest{
		TicketID: ticket.ID,
		Status:   string(enums.TicketStatusOpen),
		Reason:   "customer replied",
	}, operator); err != nil {
		t.Fatalf("ChangeStatus() reopen error = %v", err)
	}

	reopened := services.TicketService.Get(ticket.ID)
	if reopened == nil {
		t.Fatalf("expected reopened ticket to exist")
	}
	if reopened.Status != enums.TicketStatusOpen {
		t.Fatalf("expected reopened status open, got %s", reopened.Status)
	}
	if reopened.PendingReason != "" {
		t.Fatalf("expected pending reason to be cleared, got %q", reopened.PendingReason)
	}
	if reopened.ResolveDeadlineAt == nil {
		t.Fatalf("expected resolve deadline after reopening")
	}
}

func TestCloseTicketBlockedByOpenChild(t *testing.T) {
	setupTicketTestDB(t)
	operator := &dto.AuthPrincipal{UserID: 1, Username: "admin"}

	parent, err := services.TicketService.CreateTicket(createTestTicketRequest("parent-ticket"), operator)
	if err != nil {
		t.Fatalf("CreateTicket() parent error = %v", err)
	}
	child, err := services.TicketService.CreateTicket(createTestTicketRequest("child-ticket"), operator)
	if err != nil {
		t.Fatalf("CreateTicket() child error = %v", err)
	}

	if err := repositories.TicketRelationRepository.Create(sqls.DB(), &models.TicketRelation{
		TicketID:        parent.ID,
		RelatedTicketID: child.ID,
		RelationType:    enums.TicketRelationTypeChild,
		CreatedAt:       time.Now(),
	}); err != nil {
		t.Fatalf("create relation error = %v", err)
	}

	err = services.TicketService.CloseTicket(request.CloseTicketRequest{
		TicketID:    parent.ID,
		CloseReason: "done",
	}, operator)
	if err == nil {
		t.Fatalf("expected close ticket to be blocked by open child")
	}

	current := services.TicketService.Get(parent.ID)
	if current == nil {
		t.Fatalf("expected parent ticket to exist")
	}
	if current.Status != enums.TicketStatusNew {
		t.Fatalf("expected parent ticket status to remain new, got %s", current.Status)
	}
}

func TestTicketViewServiceSaveListAndDeleteOwnViews(t *testing.T) {
	setupTicketTestDB(t)
	operator := &dto.AuthPrincipal{UserID: createTestUser(t, "viewer"), Username: "viewer"}

	created, err := services.TicketViewService.Save(request.SaveTicketViewRequest{
		Name: "我的待处理",
		Filters: map[string]any{
			"quickView":    "mine",
			"statusFilter": "open",
		},
	}, operator)
	if err != nil {
		t.Fatalf("TicketViewService.Save() create error = %v", err)
	}
	if created.ID <= 0 {
		t.Fatalf("expected created ticket view id")
	}

	updated, err := services.TicketViewService.Save(request.SaveTicketViewRequest{
		ID:   created.ID,
		Name: "我的处理中",
		Filters: map[string]any{
			"quickView":    "mine",
			"statusFilter": "pending_internal",
		},
	}, operator)
	if err != nil {
		t.Fatalf("TicketViewService.Save() update error = %v", err)
	}
	if !strings.Contains(updated.FiltersJSON, "pending_internal") {
		t.Fatalf("expected updated filters json, got %s", updated.FiltersJSON)
	}

	list := services.TicketViewService.ListByUser(operator.UserID)
	if len(list) != 1 {
		t.Fatalf("expected 1 ticket view, got %d", len(list))
	}
	if list[0].Name != "我的处理中" {
		t.Fatalf("expected updated name, got %s", list[0].Name)
	}

	if err := services.TicketViewService.Delete(created.ID, operator); err != nil {
		t.Fatalf("TicketViewService.Delete() error = %v", err)
	}
	if got := services.TicketViewService.ListByUser(operator.UserID); len(got) != 0 {
		t.Fatalf("expected ticket views to be deleted, got %d", len(got))
	}
}

func TestTicketViewServiceRejectsCrossUserUpdateAndDelete(t *testing.T) {
	setupTicketTestDB(t)
	owner := &dto.AuthPrincipal{UserID: createTestUser(t, "owner"), Username: "owner"}
	other := &dto.AuthPrincipal{UserID: createTestUser(t, "other"), Username: "other"}

	created, err := services.TicketViewService.Save(request.SaveTicketViewRequest{
		Name: "owner-view",
		Filters: map[string]any{
			"quickView": "watching",
		},
	}, owner)
	if err != nil {
		t.Fatalf("TicketViewService.Save() create error = %v", err)
	}

	if _, err := services.TicketViewService.Save(request.SaveTicketViewRequest{
		ID:   created.ID,
		Name: "hijack",
		Filters: map[string]any{
			"quickView": "all",
		},
	}, other); err == nil {
		t.Fatalf("expected cross-user update to fail")
	}

	if err := services.TicketViewService.Delete(created.ID, other); err == nil {
		t.Fatalf("expected cross-user delete to fail")
	}
	if got := services.TicketViewService.ListByUser(owner.UserID); len(got) != 1 {
		t.Fatalf("expected owner view to remain, got %d", len(got))
	}
}

func setupTicketTestDB(t *testing.T) {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "ticket-test.db")
	db, err := bootstrap.InitDB(config.DBConfig{
		Type:         "sqlite",
		DSN:          "file:" + dbPath + "?_busy_timeout=5000",
		MaxIdleConns: 1,
		MaxOpenConns: 1,
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

func createTestAgentProfile(t *testing.T, prefix string) (int64, int64) {
	t.Helper()

	now := time.Now()
	userID := createTestUser(t, prefix)
	team := &models.AgentTeam{
		Name:        fmt.Sprintf("%s-team-%d", prefix, now.UnixNano()),
		Status:      enums.StatusOk,
		Description: "test team",
		AuditFields: models.AuditFields{
			CreatedAt:      now,
			CreateUserID:   1,
			CreateUserName: "admin",
			UpdatedAt:      now,
			UpdateUserID:   1,
			UpdateUserName: "admin",
		},
	}
	if err := repositories.AgentTeamRepository.Create(sqls.DB(), team); err != nil {
		t.Fatalf("create agent team error = %v", err)
	}

	profile := &models.AgentProfile{
		UserID:             userID,
		TeamID:             team.ID,
		AgentCode:          fmt.Sprintf("%s-code-%d", prefix, now.UnixNano()),
		DisplayName:        prefix,
		ServiceStatus:      enums.ServiceStatusIdle,
		MaxConcurrentCount: 5,
		AutoAssignEnabled:  true,
		Status:             enums.StatusOk,
		AuditFields: models.AuditFields{
			CreatedAt:      now,
			CreateUserID:   1,
			CreateUserName: "admin",
			UpdatedAt:      now,
			UpdateUserID:   1,
			UpdateUserName: "admin",
		},
	}
	if err := repositories.AgentProfileRepository.Create(sqls.DB(), profile); err != nil {
		t.Fatalf("create agent profile error = %v", err)
	}
	return team.ID, userID
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

func createTestTicketCategory(t *testing.T, prefix string) int64 {
	t.Helper()

	now := time.Now()
	item := &models.TicketCategory{
		Name:     fmt.Sprintf("%s-%d", prefix, now.UnixNano()),
		Status:   enums.StatusOk,
		ParentID: 0,
		SortNo:   1,
		AuditFields: models.AuditFields{
			CreatedAt:      now,
			CreateUserID:   1,
			CreateUserName: "admin",
			UpdatedAt:      now,
			UpdateUserID:   1,
			UpdateUserName: "admin",
		},
	}
	if err := repositories.TicketCategoryRepository.Create(sqls.DB(), item); err != nil {
		t.Fatalf("create ticket category error = %v", err)
	}
	return item.ID
}
