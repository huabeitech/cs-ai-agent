package services_test

import (
	"strings"
	"testing"
	"time"

	"cs-agent/internal/models"
	"cs-agent/internal/pkg/dto"
	"cs-agent/internal/pkg/dto/request"
	"cs-agent/internal/pkg/dto/response"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/services"

	"github.com/glebarez/sqlite"
	"github.com/mlogclub/simple/sqls"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func TestAgentTeamScheduleServiceFindCalendarSchedulesReturnsIntersectingSchedules(t *testing.T) {
	db := setupAgentTeamScheduleTestDB(t)
	createAgentTeamScheduleTestData(t, db)

	list, err := services.AgentTeamScheduleService.FindCalendarSchedules(request.AgentTeamScheduleCalendarRequest{
		StartAt: "2026-04-27 00:00:00",
		EndAt:   "2026-05-04 00:00:00",
	})
	if err != nil {
		t.Fatalf("FindCalendarSchedules() error = %v", err)
	}

	if len(list) != 3 {
		t.Fatalf("expected 3 intersecting schedules, got %d: %+v", len(list), list)
	}
	gotIDs := make([]int64, 0, len(list))
	for _, item := range list {
		gotIDs = append(gotIDs, item.ID)
	}
	wantIDs := []int64{1, 2, 3}
	for i, want := range wantIDs {
		if gotIDs[i] != want {
			t.Fatalf("expected ids %v, got %v", wantIDs, gotIDs)
		}
	}
}

func TestAgentTeamScheduleServiceFindCalendarSchedulesFiltersTeamID(t *testing.T) {
	db := setupAgentTeamScheduleTestDB(t)
	createAgentTeamScheduleTestData(t, db)

	list, err := services.AgentTeamScheduleService.FindCalendarSchedules(request.AgentTeamScheduleCalendarRequest{
		StartAt: "2026-04-27 00:00:00",
		EndAt:   "2026-05-04 00:00:00",
		TeamID:  2,
	})
	if err != nil {
		t.Fatalf("FindCalendarSchedules() error = %v", err)
	}

	if len(list) != 1 {
		t.Fatalf("expected 1 schedule for team 2, got %d: %+v", len(list), list)
	}
	if list[0].ID != 3 || list[0].TeamID != 2 {
		t.Fatalf("unexpected schedule: %+v", list[0])
	}
}

func TestAgentTeamScheduleServiceFindCalendarSchedulesValidatesTimeRange(t *testing.T) {
	setupAgentTeamScheduleTestDB(t)

	_, err := services.AgentTeamScheduleService.FindCalendarSchedules(request.AgentTeamScheduleCalendarRequest{
		StartAt: "2026-05-04 00:00:00",
		EndAt:   "2026-04-27 00:00:00",
	})
	if err == nil {
		t.Fatalf("expected invalid time range to fail")
	}
}

func TestAgentTeamScheduleServiceCreateRejectsCrossDaySchedule(t *testing.T) {
	setupAgentTeamScheduleTestDB(t)
	createAgentTeamScheduleTestTeams(t, sqls.DB())

	tomorrow := time.Now().AddDate(0, 0, 1)
	_, err := services.AgentTeamScheduleService.CreateAgentTeamSchedule(request.CreateAgentTeamScheduleRequest{
		TeamID:  1,
		StartAt: formatTestDateTime(tomorrow, "22:00:00"),
		EndAt:   formatTestDateTime(tomorrow.AddDate(0, 0, 1), "08:00:00"),
	}, testOperator())
	if err == nil {
		t.Fatalf("expected cross-day schedule to fail")
	}
	if !strings.Contains(err.Error(), "不能跨天") {
		t.Fatalf("expected cross-day error, got %v", err)
	}
}

func TestAgentTeamScheduleServiceCreateRejectsHistoricalScheduleByDay(t *testing.T) {
	setupAgentTeamScheduleTestDB(t)
	createAgentTeamScheduleTestTeams(t, sqls.DB())

	yesterday := time.Now().AddDate(0, 0, -1)
	_, err := services.AgentTeamScheduleService.CreateAgentTeamSchedule(request.CreateAgentTeamScheduleRequest{
		TeamID:  1,
		StartAt: formatTestDateTime(yesterday, "09:00:00"),
		EndAt:   formatTestDateTime(yesterday, "18:00:00"),
	}, testOperator())
	if err == nil {
		t.Fatalf("expected historical schedule to fail")
	}
	if !strings.Contains(err.Error(), "历史日期") {
		t.Fatalf("expected historical date error, got %v", err)
	}
}

func TestAgentTeamScheduleServiceCreateAllowsTodayEarlierThanCurrentTime(t *testing.T) {
	setupAgentTeamScheduleTestDB(t)
	createAgentTeamScheduleTestTeams(t, sqls.DB())

	today := time.Now()
	item, err := services.AgentTeamScheduleService.CreateAgentTeamSchedule(request.CreateAgentTeamScheduleRequest{
		TeamID:  1,
		StartAt: formatTestDateTime(today, "00:00:00"),
		EndAt:   formatTestDateTime(today, "01:00:00"),
	}, testOperator())
	if err != nil {
		t.Fatalf("expected today's schedule to pass, got %v", err)
	}
	if item == nil || item.ID == 0 {
		t.Fatalf("expected created schedule, got %+v", item)
	}
}

func TestAgentTeamScheduleServiceUpdateRejectsCrossDaySchedule(t *testing.T) {
	db := setupAgentTeamScheduleTestDB(t)
	createAgentTeamScheduleTestTeams(t, db)
	existingID := createFutureAgentTeamSchedule(t, db)
	tomorrow := time.Now().AddDate(0, 0, 1)

	err := services.AgentTeamScheduleService.UpdateAgentTeamSchedule(request.UpdateAgentTeamScheduleRequest{
		ID: existingID,
		CreateAgentTeamScheduleRequest: request.CreateAgentTeamScheduleRequest{
			TeamID:  1,
			StartAt: formatTestDateTime(tomorrow, "22:00:00"),
			EndAt:   formatTestDateTime(tomorrow.AddDate(0, 0, 1), "08:00:00"),
		},
	}, testOperator())
	if err == nil {
		t.Fatalf("expected cross-day update to fail")
	}
	if !strings.Contains(err.Error(), "不能跨天") {
		t.Fatalf("expected cross-day error, got %v", err)
	}
}

func TestAgentTeamScheduleServiceUpdateRejectsHistoricalScheduleByDay(t *testing.T) {
	db := setupAgentTeamScheduleTestDB(t)
	createAgentTeamScheduleTestTeams(t, db)
	existingID := createFutureAgentTeamSchedule(t, db)
	yesterday := time.Now().AddDate(0, 0, -1)

	err := services.AgentTeamScheduleService.UpdateAgentTeamSchedule(request.UpdateAgentTeamScheduleRequest{
		ID: existingID,
		CreateAgentTeamScheduleRequest: request.CreateAgentTeamScheduleRequest{
			TeamID:  1,
			StartAt: formatTestDateTime(yesterday, "09:00:00"),
			EndAt:   formatTestDateTime(yesterday, "18:00:00"),
		},
	}, testOperator())
	if err == nil {
		t.Fatalf("expected historical update to fail")
	}
	if !strings.Contains(err.Error(), "历史日期") {
		t.Fatalf("expected historical date error, got %v", err)
	}
}

func TestAgentTeamScheduleServiceBatchPreviewExpandsSharedRule(t *testing.T) {
	setupAgentTeamScheduleTestDB(t)
	createAgentTeamScheduleTestTeams(t, sqls.DB())
	nextMonday := nextTestWeekday(time.Monday)
	nextWednesday := nextMonday.AddDate(0, 0, 2)

	preview, err := services.AgentTeamScheduleService.BatchPreview(request.AgentTeamScheduleBatchRequest{
		TeamIDs:   []int64{1, 2},
		StartDate: nextMonday.Format(time.DateOnly),
		EndDate:   nextMonday.AddDate(0, 0, 6).Format(time.DateOnly),
		Weekdays:  []int{1, 3},
		StartTime: "09:00",
		EndTime:   "18:00",
		Remark:    "工作日白班",
	}, testOperator())
	if err != nil {
		t.Fatalf("BatchPreview() error = %v", err)
	}
	if preview.Total != 4 || len(preview.Items) != 4 {
		t.Fatalf("expected 4 preview items, got total=%d len=%d", preview.Total, len(preview.Items))
	}
	if preview.Conflict {
		t.Fatalf("expected no conflict, got %+v", preview.Items)
	}
	teamNames := make(map[int64]string)
	for _, item := range preview.Items {
		if item.TeamName == "" {
			t.Fatalf("expected all preview items to have team names: %+v", preview.Items)
		}
		teamNames[item.TeamID] = item.TeamName
	}
	if teamNames[1] == "" || teamNames[2] == "" {
		t.Fatalf("expected team ids 1 and 2 with names, got %v", teamNames)
	}
	type previewKey struct {
		teamID int64
		date   string
	}
	itemsByKey := make(map[previewKey]response.AgentTeamScheduleBatchPreviewItem)
	for _, item := range preview.Items {
		itemsByKey[previewKey{teamID: item.TeamID, date: item.Date}] = item
	}
	expected := []struct {
		teamID  int64
		date    time.Time
		weekday int
	}{
		{teamID: 1, date: nextMonday, weekday: 1},
		{teamID: 1, date: nextWednesday, weekday: 3},
		{teamID: 2, date: nextMonday, weekday: 1},
		{teamID: 2, date: nextWednesday, weekday: 3},
	}
	for _, want := range expected {
		wantDate := want.date.Format(time.DateOnly)
		item, ok := itemsByKey[previewKey{teamID: want.teamID, date: wantDate}]
		if !ok {
			t.Fatalf("expected preview item for teamID=%d date=%s, got %+v", want.teamID, wantDate, preview.Items)
		}
		if item.Weekday != want.weekday ||
			item.StartAt != formatTestDateTime(want.date, "09:00:00") ||
			item.EndAt != formatTestDateTime(want.date, "18:00:00") ||
			item.Remark != "工作日白班" {
			t.Fatalf("unexpected preview item for teamID=%d date=%s: %+v", want.teamID, wantDate, item)
		}
	}
}

func TestAgentTeamScheduleServiceBatchPreviewRejectsHistoricalDate(t *testing.T) {
	setupAgentTeamScheduleTestDB(t)
	createAgentTeamScheduleTestTeams(t, sqls.DB())
	yesterday := time.Now().AddDate(0, 0, -1)

	_, err := services.AgentTeamScheduleService.BatchPreview(request.AgentTeamScheduleBatchRequest{
		TeamIDs:   []int64{1},
		StartDate: yesterday.Format(time.DateOnly),
		EndDate:   yesterday.Format(time.DateOnly),
		Weekdays:  []int{weekdayForRequest(yesterday)},
		StartTime: "09:00",
		EndTime:   "18:00",
	}, testOperator())
	if err == nil {
		t.Fatalf("expected historical batch preview to fail")
	}
	if !strings.Contains(err.Error(), "历史日期") {
		t.Fatalf("expected historical date error, got %v", err)
	}
}

func TestAgentTeamScheduleServiceBatchPreviewRejectsInvalidTimeRange(t *testing.T) {
	setupAgentTeamScheduleTestDB(t)
	createAgentTeamScheduleTestTeams(t, sqls.DB())
	tomorrow := time.Now().AddDate(0, 0, 1)

	_, err := services.AgentTeamScheduleService.BatchPreview(request.AgentTeamScheduleBatchRequest{
		TeamIDs:   []int64{1},
		StartDate: tomorrow.Format(time.DateOnly),
		EndDate:   tomorrow.Format(time.DateOnly),
		Weekdays:  []int{weekdayForRequest(tomorrow)},
		StartTime: "18:00",
		EndTime:   "09:00",
	}, testOperator())
	if err == nil {
		t.Fatalf("expected invalid time range to fail")
	}
	if !strings.Contains(err.Error(), "结束时间必须晚于开始时间") {
		t.Fatalf("expected invalid time range error, got %v", err)
	}
}

func TestAgentTeamScheduleServiceBatchPreviewRejectsOverLimit(t *testing.T) {
	setupAgentTeamScheduleTestDB(t)
	createAgentTeamScheduleTestTeams(t, sqls.DB())
	today := time.Now()

	_, err := services.AgentTeamScheduleService.BatchPreview(request.AgentTeamScheduleBatchRequest{
		TeamIDs:   []int64{1, 2},
		StartDate: today.Format(time.DateOnly),
		EndDate:   today.AddDate(0, 0, 260).Format(time.DateOnly),
		Weekdays:  []int{1, 2, 3, 4, 5, 6, 7},
		StartTime: "09:00",
		EndTime:   "18:00",
	}, testOperator())
	if err == nil {
		t.Fatalf("expected over-limit preview to fail")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Fatalf("expected 500 limit error, got %v", err)
	}
}

func TestAgentTeamScheduleServiceBatchPreviewMarksConflicts(t *testing.T) {
	db := setupAgentTeamScheduleTestDB(t)
	createAgentTeamScheduleTestTeams(t, db)
	targetDay := time.Now().AddDate(0, 0, 2)
	existing := models.AgentTeamSchedule{
		TeamID:  1,
		StartAt: parseTestDateTime(t, formatTestDateTime(targetDay, "10:00:00")),
		EndAt:   parseTestDateTime(t, formatTestDateTime(targetDay, "12:00:00")),
		Status:  enums.StatusOk,
	}
	if err := db.Create(&existing).Error; err != nil {
		t.Fatalf("create existing schedule error = %v", err)
	}

	preview, err := services.AgentTeamScheduleService.BatchPreview(request.AgentTeamScheduleBatchRequest{
		TeamIDs:   []int64{1, 2},
		StartDate: targetDay.Format(time.DateOnly),
		EndDate:   targetDay.Format(time.DateOnly),
		Weekdays:  []int{weekdayForRequest(targetDay)},
		StartTime: "09:00",
		EndTime:   "18:00",
	}, testOperator())
	if err != nil {
		t.Fatalf("BatchPreview() error = %v", err)
	}
	if !preview.Conflict {
		t.Fatalf("expected preview conflict, got %+v", preview)
	}
	itemsByTeamID := make(map[int64]response.AgentTeamScheduleBatchPreviewItem)
	for _, item := range preview.Items {
		itemsByTeamID[item.TeamID] = item
	}
	team1Item, ok := itemsByTeamID[1]
	if !ok {
		t.Fatalf("expected team 1 preview item, got %+v", preview.Items)
	}
	team2Item, ok := itemsByTeamID[2]
	if !ok {
		t.Fatalf("expected team 2 preview item, got %+v", preview.Items)
	}
	if !team1Item.Conflict || team1Item.ConflictReason == "" {
		t.Fatalf("expected team 1 preview item to be marked as conflict: %+v", team1Item)
	}
	if team2Item.Conflict {
		t.Fatalf("expected team 2 preview item to have no conflict: %+v", team2Item)
	}
}

func TestAgentTeamScheduleServiceBatchPreviewIgnoresDisabledOverlappingSchedule(t *testing.T) {
	db := setupAgentTeamScheduleTestDB(t)
	createAgentTeamScheduleTestTeams(t, db)
	targetDay := time.Now().AddDate(0, 0, 2)
	existing := models.AgentTeamSchedule{
		TeamID:  1,
		StartAt: parseTestDateTime(t, formatTestDateTime(targetDay, "10:00:00")),
		EndAt:   parseTestDateTime(t, formatTestDateTime(targetDay, "12:00:00")),
		Status:  enums.StatusDisabled,
	}
	if err := db.Create(&existing).Error; err != nil {
		t.Fatalf("create existing schedule error = %v", err)
	}

	preview, err := services.AgentTeamScheduleService.BatchPreview(request.AgentTeamScheduleBatchRequest{
		TeamIDs:   []int64{1},
		StartDate: targetDay.Format(time.DateOnly),
		EndDate:   targetDay.Format(time.DateOnly),
		Weekdays:  []int{weekdayForRequest(targetDay)},
		StartTime: "09:00",
		EndTime:   "18:00",
	}, testOperator())
	if err != nil {
		t.Fatalf("BatchPreview() error = %v", err)
	}
	if preview.Conflict {
		t.Fatalf("expected disabled overlapping schedule to be ignored, got %+v", preview)
	}
	if len(preview.Items) != 1 || preview.Items[0].Conflict {
		t.Fatalf("expected one non-conflicting preview item, got %+v", preview.Items)
	}
}

func TestAgentTeamScheduleServiceBatchGenerateCreatesAllSchedules(t *testing.T) {
	db := setupAgentTeamScheduleTestDB(t)
	createAgentTeamScheduleTestTeams(t, db)
	nextMonday := nextTestWeekday(time.Monday)

	result, err := services.AgentTeamScheduleService.BatchGenerate(request.AgentTeamScheduleBatchRequest{
		TeamIDs:   []int64{1, 2},
		StartDate: nextMonday.Format(time.DateOnly),
		EndDate:   nextMonday.AddDate(0, 0, 2).Format(time.DateOnly),
		Weekdays:  []int{1, 3},
		StartTime: "09:00",
		EndTime:   "18:00",
		Remark:    "批量生成",
	}, testOperator())
	if err != nil {
		t.Fatalf("BatchGenerate() error = %v", err)
	}
	if result.Created != 4 {
		t.Fatalf("expected 4 created schedules, got %d", result.Created)
	}
	var schedules []models.AgentTeamSchedule
	if err := db.Where("remark = ?", "批量生成").
		Order("team_id ASC, start_at ASC").
		Find(&schedules).Error; err != nil {
		t.Fatalf("query generated schedules error = %v", err)
	}
	if len(schedules) != 4 {
		t.Fatalf("expected 4 stored schedules, got %d: %+v", len(schedules), schedules)
	}
	expected := []struct {
		teamID  int64
		startAt string
		endAt   string
	}{
		{teamID: 1, startAt: formatTestDateTime(nextMonday, "09:00:00"), endAt: formatTestDateTime(nextMonday, "18:00:00")},
		{teamID: 1, startAt: formatTestDateTime(nextMonday.AddDate(0, 0, 2), "09:00:00"), endAt: formatTestDateTime(nextMonday.AddDate(0, 0, 2), "18:00:00")},
		{teamID: 2, startAt: formatTestDateTime(nextMonday, "09:00:00"), endAt: formatTestDateTime(nextMonday, "18:00:00")},
		{teamID: 2, startAt: formatTestDateTime(nextMonday.AddDate(0, 0, 2), "09:00:00"), endAt: formatTestDateTime(nextMonday.AddDate(0, 0, 2), "18:00:00")},
	}
	for i, want := range expected {
		got := schedules[i]
		if got.TeamID != want.teamID ||
			got.StartAt.Format(time.DateTime) != want.startAt ||
			got.EndAt.Format(time.DateTime) != want.endAt {
			t.Fatalf("unexpected schedule at index %d: got teamID=%d startAt=%s endAt=%s, want teamID=%d startAt=%s endAt=%s",
				i,
				got.TeamID,
				got.StartAt.Format(time.DateTime),
				got.EndAt.Format(time.DateTime),
				want.teamID,
				want.startAt,
				want.endAt,
			)
		}
	}
}

func TestAgentTeamScheduleServiceBatchGenerateRejectsConflictsWithoutPartialCreate(t *testing.T) {
	db := setupAgentTeamScheduleTestDB(t)
	createAgentTeamScheduleTestTeams(t, db)
	targetDay := time.Now().AddDate(0, 0, 2)
	existing := models.AgentTeamSchedule{
		TeamID:  1,
		StartAt: parseTestDateTime(t, formatTestDateTime(targetDay, "10:00:00")),
		EndAt:   parseTestDateTime(t, formatTestDateTime(targetDay, "12:00:00")),
		Status:  enums.StatusOk,
	}
	if err := db.Create(&existing).Error; err != nil {
		t.Fatalf("create existing schedule error = %v", err)
	}

	_, err := services.AgentTeamScheduleService.BatchGenerate(request.AgentTeamScheduleBatchRequest{
		TeamIDs:   []int64{1, 2},
		StartDate: targetDay.Format(time.DateOnly),
		EndDate:   targetDay.Format(time.DateOnly),
		Weekdays:  []int{weekdayForRequest(targetDay)},
		StartTime: "09:00",
		EndTime:   "18:00",
		Remark:    "不应创建",
	}, testOperator())
	if err == nil {
		t.Fatalf("expected conflict batch generate to fail")
	}
	var count int64
	db.Model(&models.AgentTeamSchedule{}).Where("remark = ?", "不应创建").Count(&count)
	if count != 0 {
		t.Fatalf("expected no partial creates, got %d", count)
	}
}

func setupAgentTeamScheduleTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dbName := strings.NewReplacer("/", "_", " ", "_").Replace(t.Name())
	db, err := gorm.Open(sqlite.Open("file:"+dbName+"?mode=memory&cache=shared"), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "t_",
			SingularTable: true,
		},
	})
	if err != nil {
		t.Fatalf("open sqlite error = %v", err)
	}
	t.Cleanup(func() {
		sqlDB, err := db.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	})
	if err := db.AutoMigrate(&models.AgentTeam{}, &models.AgentTeamSchedule{}); err != nil {
		t.Fatalf("auto migrate error = %v", err)
	}
	sqls.SetDB(db)
	return db
}

func createAgentTeamScheduleTestData(t *testing.T, db *gorm.DB) {
	t.Helper()

	createAgentTeamScheduleTestTeams(t, db)

	parse := func(value string) time.Time {
		t.Helper()
		ret, err := time.ParseInLocation(time.DateTime, value, time.Local)
		if err != nil {
			t.Fatalf("parse time %q error = %v", value, err)
		}
		return ret
	}
	schedules := []models.AgentTeamSchedule{
		{ID: 1, TeamID: 1, StartAt: parse("2026-04-26 20:00:00"), EndAt: parse("2026-04-27 10:00:00"), Status: enums.StatusOk},
		{ID: 2, TeamID: 1, StartAt: parse("2026-04-28 09:00:00"), EndAt: parse("2026-04-28 18:00:00"), Status: enums.StatusOk},
		{ID: 3, TeamID: 2, StartAt: parse("2026-05-03 20:00:00"), EndAt: parse("2026-05-04 08:00:00"), Status: enums.StatusOk},
		{ID: 4, TeamID: 1, StartAt: parse("2026-04-20 09:00:00"), EndAt: parse("2026-04-20 18:00:00"), Status: enums.StatusOk},
		{ID: 5, TeamID: 2, StartAt: parse("2026-05-04 09:00:00"), EndAt: parse("2026-05-04 18:00:00"), Status: enums.StatusOk},
	}
	if err := db.Create(&schedules).Error; err != nil {
		t.Fatalf("create schedules error = %v", err)
	}
}

func createAgentTeamScheduleTestTeams(t *testing.T, db *gorm.DB) {
	t.Helper()
	teams := []models.AgentTeam{
		{ID: 1, Name: "售前组", Status: enums.StatusOk},
		{ID: 2, Name: "售后组", Status: enums.StatusOk},
	}
	if err := db.Create(&teams).Error; err != nil {
		t.Fatalf("create teams error = %v", err)
	}
}

func formatTestDateTime(date time.Time, clock string) string {
	return date.Format(time.DateOnly) + " " + clock
}

func nextTestWeekday(target time.Weekday) time.Time {
	ret := startOfTestDay(time.Now()).AddDate(0, 0, 1)
	for ret.Weekday() != target {
		ret = ret.AddDate(0, 0, 1)
	}
	return ret
}

func startOfTestDay(value time.Time) time.Time {
	year, month, day := value.In(time.Local).Date()
	return time.Date(year, month, day, 0, 0, 0, 0, time.Local)
}

func weekdayForRequest(value time.Time) int {
	if value.Weekday() == time.Sunday {
		return 7
	}
	return int(value.Weekday())
}

func createFutureAgentTeamSchedule(t *testing.T, db *gorm.DB) int64 {
	t.Helper()
	tomorrow := time.Now().AddDate(0, 0, 1)
	item := models.AgentTeamSchedule{
		TeamID:  1,
		StartAt: parseTestDateTime(t, formatTestDateTime(tomorrow, "09:00:00")),
		EndAt:   parseTestDateTime(t, formatTestDateTime(tomorrow, "18:00:00")),
		Status:  enums.StatusOk,
	}
	if err := db.Create(&item).Error; err != nil {
		t.Fatalf("create future schedule error = %v", err)
	}
	return item.ID
}

func parseTestDateTime(t *testing.T, value string) time.Time {
	t.Helper()
	ret, err := time.ParseInLocation(time.DateTime, value, time.Local)
	if err != nil {
		t.Fatalf("parse time %q error = %v", value, err)
	}
	return ret
}

func testOperator() *dto.AuthPrincipal {
	return &dto.AuthPrincipal{UserID: 1, Username: "tester", Status: enums.StatusOk}
}
