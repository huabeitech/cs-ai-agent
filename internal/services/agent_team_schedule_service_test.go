package services_test

import (
	"strings"
	"testing"
	"time"

	"cs-agent/internal/models"
	"cs-agent/internal/pkg/dto"
	"cs-agent/internal/pkg/dto/request"
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
		TeamID:     1,
		StartAt:    formatTestDateTime(tomorrow, "22:00:00"),
		EndAt:      formatTestDateTime(tomorrow.AddDate(0, 0, 1), "08:00:00"),
		SourceType: "manual",
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
		TeamID:     1,
		StartAt:    formatTestDateTime(yesterday, "09:00:00"),
		EndAt:      formatTestDateTime(yesterday, "18:00:00"),
		SourceType: "manual",
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
		TeamID:     1,
		StartAt:    formatTestDateTime(today, "00:00:00"),
		EndAt:      formatTestDateTime(today, "01:00:00"),
		SourceType: "manual",
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
			TeamID:     1,
			StartAt:    formatTestDateTime(tomorrow, "22:00:00"),
			EndAt:      formatTestDateTime(tomorrow.AddDate(0, 0, 1), "08:00:00"),
			SourceType: "manual",
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
			TeamID:     1,
			StartAt:    formatTestDateTime(yesterday, "09:00:00"),
			EndAt:      formatTestDateTime(yesterday, "18:00:00"),
			SourceType: "manual",
		},
	}, testOperator())
	if err == nil {
		t.Fatalf("expected historical update to fail")
	}
	if !strings.Contains(err.Error(), "历史日期") {
		t.Fatalf("expected historical date error, got %v", err)
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
		{ID: 1, TeamID: 1, StartAt: parse("2026-04-26 20:00:00"), EndAt: parse("2026-04-27 10:00:00"), SourceType: "manual", Status: enums.StatusOk},
		{ID: 2, TeamID: 1, StartAt: parse("2026-04-28 09:00:00"), EndAt: parse("2026-04-28 18:00:00"), SourceType: "manual", Status: enums.StatusOk},
		{ID: 3, TeamID: 2, StartAt: parse("2026-05-03 20:00:00"), EndAt: parse("2026-05-04 08:00:00"), SourceType: "manual", Status: enums.StatusOk},
		{ID: 4, TeamID: 1, StartAt: parse("2026-04-20 09:00:00"), EndAt: parse("2026-04-20 18:00:00"), SourceType: "manual", Status: enums.StatusOk},
		{ID: 5, TeamID: 2, StartAt: parse("2026-05-04 09:00:00"), EndAt: parse("2026-05-04 18:00:00"), SourceType: "manual", Status: enums.StatusOk},
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

func createFutureAgentTeamSchedule(t *testing.T, db *gorm.DB) int64 {
	t.Helper()
	tomorrow := time.Now().AddDate(0, 0, 1)
	item := models.AgentTeamSchedule{
		TeamID:     1,
		StartAt:    parseTestDateTime(t, formatTestDateTime(tomorrow, "09:00:00")),
		EndAt:      parseTestDateTime(t, formatTestDateTime(tomorrow, "18:00:00")),
		SourceType: "manual",
		Status:     enums.StatusOk,
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
