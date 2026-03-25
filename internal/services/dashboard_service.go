package services

import (
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/dto/response"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/repositories"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/mlogclub/simple/sqls"
	"gorm.io/gorm"
)

var DashboardService = newDashboardService()

func newDashboardService() *dashboardService {
	return &dashboardService{}
}

type dashboardService struct {
}

func (s *dashboardService) GetOverview(rangeValue string) response.DashboardOverviewResponse {
	now := time.Now()
	normalizedRange, trendDays := normalizeDashboardRange(rangeValue)
	todayStart := startOfDay(now)
	trendStart := todayStart.AddDate(0, 0, -(trendDays - 1))
	db := sqls.DB()

	conversationTodayCount := repositories.DashboardRepository.CountConversations(db, func(tx *gorm.DB) *gorm.DB {
		return tx.Where("created_at >= ?", todayStart)
	})
	processingConversationCount := repositories.DashboardRepository.CountConversations(db, func(tx *gorm.DB) *gorm.DB {
		return tx.Where("status = ?", enums.IMConversationStatusActive)
	})
	pendingConversationCount := repositories.DashboardRepository.CountConversations(db, func(tx *gorm.DB) *gorm.DB {
		return tx.Where("status = ?", enums.IMConversationStatusPending)
	})
	ticketTodayCount := repositories.DashboardRepository.CountTickets(db, func(tx *gorm.DB) *gorm.DB {
		return tx.Where("created_at >= ?", todayStart)
	})

	agentProfiles := repositories.DashboardRepository.ListEnabledAgentProfiles(db)
	agentTeams := repositories.DashboardRepository.ListEnabledAgentTeams(db)
	activeSchedules := repositories.DashboardRepository.ListActiveTeamSchedules(db, now, now)
	activeConversations := repositories.DashboardRepository.ListConversations(db, func(tx *gorm.DB) *gorm.DB {
		return tx.Where("status IN ?", []enums.IMConversationStatus{
			enums.IMConversationStatusPending,
			enums.IMConversationStatusActive,
		})
	})

	onlineAgents, busyAgents, offlineAgents, teamLoads := s.buildAgentStats(now, agentTeams, agentProfiles, activeSchedules, activeConversations)

	enabledAIAgentCount := repositories.DashboardRepository.CountAIAgents(db, func(tx *gorm.DB) *gorm.DB {
		return tx.Where("status = ?", enums.StatusOk)
	})
	enabledWidgetSiteCount := repositories.DashboardRepository.CountWidgetSites(db, func(tx *gorm.DB) *gorm.DB {
		return tx.Where("status = ?", enums.StatusOk)
	})
	knowledgeRetrieveCount := repositories.DashboardRepository.CountKnowledgeRetrieveLogs(db, func(tx *gorm.DB) *gorm.DB {
		return tx.Where("created_at >= ?", todayStart)
	})
	knowledgeRetrieveFailCount := repositories.DashboardRepository.CountKnowledgeRetrieveLogs(db, func(tx *gorm.DB) *gorm.DB {
		return tx.Where("created_at >= ? AND answer_status IN ?", todayStart, []int{2, 3, 4})
	})
	skillRunFailCount := repositories.DashboardRepository.CountSkillRunLogs(db, func(tx *gorm.DB) *gorm.DB {
		return tx.Where("created_at >= ? AND error_message <> ''", todayStart)
	})
	aiHandoffCount := repositories.DashboardRepository.CountConversations(db, func(tx *gorm.DB) *gorm.DB {
		return tx.Where("handoff_at >= ?", todayStart)
	})

	enabledAIAgents := repositories.DashboardRepository.ListAIAgents(db, func(tx *gorm.DB) *gorm.DB {
		return tx.Where("status = ?", enums.StatusOk)
	})
	alerts := s.buildAlerts(now, db, activeConversations, enabledAIAgents, agentTeams, activeSchedules)

	return response.DashboardOverviewResponse{
		Range:       normalizedRange,
		GeneratedAt: now.Format("2006-01-02 15:04:05"),
		Summary: response.DashboardSummaryResponse{
			TodayNewConversations:        conversationTodayCount,
			ProcessingConversations:      processingConversationCount,
			PendingDispatchConversations: pendingConversationCount,
			TodayNewTickets:              ticketTodayCount,
			OnlineAgents:                 onlineAgents,
			AIServiceRate:                calcAIServiceRate(activeConversations),
		},
		ConversationStats: response.DashboardSectionStatsResponse{
			StatusDistribution: buildConversationStatusDistribution(db),
			Trend:              buildConversationTrend(db, trendStart),
		},
		TicketStats: response.DashboardSectionStatsResponse{
			StatusDistribution: buildTicketStatusDistribution(db),
			Trend:              buildTicketTrend(db, trendStart),
		},
		AgentStats: response.DashboardAgentStatsResponse{
			OnlineAgents:  onlineAgents,
			BusyAgents:    busyAgents,
			OfflineAgents: offlineAgents,
			TeamLoads:     teamLoads,
		},
		AIStats: response.DashboardAIStatsResponse{
			EnabledAIAgents:                 enabledAIAgentCount,
			EnabledWidgetSites:              enabledWidgetSiteCount,
			TodayKnowledgeRetrieves:         knowledgeRetrieveCount,
			TodayKnowledgeRetrieveFailCount: knowledgeRetrieveFailCount,
			TodayKnowledgeRetrieveFailRate:  calcRate(knowledgeRetrieveFailCount, knowledgeRetrieveCount),
			TodaySkillRunFailCount:          skillRunFailCount,
			TodayAIHandoffCount:             aiHandoffCount,
		},
		Alerts:     alerts,
		QuickLinks: buildDashboardQuickLinks(),
	}
}

func (s *dashboardService) buildAgentStats(now time.Time, teams []models.AgentTeam, profiles []models.AgentProfile, schedules []models.AgentTeamSchedule, conversations []models.Conversation) (int64, int64, int64, []response.DashboardTeamLoadResponse) {
	const onlineWindow = 15 * time.Minute

	scheduledTeamIDs := make(map[int64]bool, len(schedules))
	for _, item := range schedules {
		scheduledTeamIDs[item.TeamID] = true
	}

	type teamCounter struct {
		totalAgents             int64
		onlineAgents            int64
		busyAgents              int64
		offlineAgents           int64
		waitingConversations    int64
		processingConversations int64
		maxConcurrentCapacity   int64
	}

	teamCounters := make(map[int64]*teamCounter, len(teams))
	for _, team := range teams {
		teamCounters[team.ID] = &teamCounter{}
	}

	var onlineAgents int64
	var busyAgents int64
	var offlineAgents int64

	for _, profile := range profiles {
		counter := teamCounters[profile.TeamID]
		if counter == nil {
			counter = &teamCounter{}
			teamCounters[profile.TeamID] = counter
		}
		counter.totalAgents++
		counter.maxConcurrentCapacity += int64(profile.MaxConcurrentCount)
		if profile.LastOnlineAt != nil && now.Sub(*profile.LastOnlineAt) <= onlineWindow {
			counter.onlineAgents++
			onlineAgents++
			if profile.ServiceStatus == enums.ServiceStatusBusy {
				counter.busyAgents++
				busyAgents++
			}
			continue
		}
		counter.offlineAgents++
		offlineAgents++
	}

	for _, item := range conversations {
		if item.CurrentTeamID <= 0 {
			continue
		}
		counter := teamCounters[item.CurrentTeamID]
		if counter == nil {
			counter = &teamCounter{}
			teamCounters[item.CurrentTeamID] = counter
		}
		switch item.Status {
		case enums.IMConversationStatusPending:
			counter.waitingConversations++
		case enums.IMConversationStatusActive:
			counter.processingConversations++
		}
	}

	teamLoads := make([]response.DashboardTeamLoadResponse, 0, len(teams))
	for _, team := range teams {
		counter := teamCounters[team.ID]
		if counter == nil {
			counter = &teamCounter{}
		}
		teamLoads = append(teamLoads, response.DashboardTeamLoadResponse{
			TeamID:                  team.ID,
			TeamName:                team.Name,
			TotalAgents:             counter.totalAgents,
			OnlineAgents:            counter.onlineAgents,
			BusyAgents:              counter.busyAgents,
			OfflineAgents:           counter.offlineAgents,
			WaitingConversations:    counter.waitingConversations,
			ProcessingConversations: counter.processingConversations,
			MaxConcurrentCapacity:   counter.maxConcurrentCapacity,
			LoadRate:                calcRate(counter.processingConversations, counter.maxConcurrentCapacity),
			HasScheduleNow:          scheduledTeamIDs[team.ID],
		})
	}

	sort.Slice(teamLoads, func(i, j int) bool {
		if teamLoads[i].WaitingConversations == teamLoads[j].WaitingConversations {
			if teamLoads[i].LoadRate == teamLoads[j].LoadRate {
				return teamLoads[i].TeamID < teamLoads[j].TeamID
			}
			return teamLoads[i].LoadRate > teamLoads[j].LoadRate
		}
		return teamLoads[i].WaitingConversations > teamLoads[j].WaitingConversations
	})

	return onlineAgents, busyAgents, offlineAgents, teamLoads
}

func (s *dashboardService) buildAlerts(now time.Time, db *gorm.DB, activeConversations []models.Conversation, aiAgents []models.AIAgent, teams []models.AgentTeam, schedules []models.AgentTeamSchedule) []response.DashboardAlertResponse {
	alerts := make([]response.DashboardAlertResponse, 0, 4)
	pendingTimeout := now.Add(-10 * time.Minute)
	activeTimeout := now.Add(-30 * time.Minute)

	pendingLongWaitCount := repositories.DashboardRepository.CountConversations(db, func(tx *gorm.DB) *gorm.DB {
		return tx.Where("status = ? AND created_at < ?", enums.IMConversationStatusPending, pendingTimeout)
	})
	if pendingLongWaitCount > 0 {
		alerts = append(alerts, response.DashboardAlertResponse{
			ID:          "pending-long-wait",
			Level:       "warning",
			Title:       "待接入会话堆积",
			Description: "存在超过 10 分钟仍未接入的会话，建议优先处理分配。",
			Count:       pendingLongWaitCount,
			Link:        "/dashboard/conversations",
		})
	}

	staleProcessingCount := repositories.DashboardRepository.CountConversations(db, func(tx *gorm.DB) *gorm.DB {
		return tx.Where("status = ? AND (last_message_at IS NULL OR last_message_at < ?)", enums.IMConversationStatusActive, activeTimeout)
	})
	if staleProcessingCount > 0 {
		alerts = append(alerts, response.DashboardAlertResponse{
			ID:          "stale-processing",
			Level:       "warning",
			Title:       "处理中会话长时间无响应",
			Description: "部分处理中会话已超过 30 分钟没有最新消息，需要确认跟进状态。",
			Count:       staleProcessingCount,
			Link:        "/dashboard/conversations",
		})
	}

	scheduledTeamIDs := make(map[int64]bool, len(schedules))
	for _, item := range schedules {
		scheduledTeamIDs[item.TeamID] = true
	}
	var scheduleMissingCount int64
	for _, team := range teams {
		if !scheduledTeamIDs[team.ID] {
			scheduleMissingCount++
		}
	}
	if scheduleMissingCount > 0 {
		alerts = append(alerts, response.DashboardAlertResponse{
			ID:          "team-no-schedule",
			Level:       "info",
			Title:       "客服组当前无生效排班",
			Description: "部分启用中的客服组当前没有生效排班，可能影响自动分配。",
			Count:       scheduleMissingCount,
			Link:        "/dashboard/agent-team-schedules",
		})
	}

	var aiAgentWithoutKnowledgeCount int64
	for _, item := range aiAgents {
		if strings.TrimSpace(item.KnowledgeIDs) == "" {
			aiAgentWithoutKnowledgeCount++
		}
	}
	if aiAgentWithoutKnowledgeCount > 0 {
		alerts = append(alerts, response.DashboardAlertResponse{
			ID:          "ai-no-knowledge",
			Level:       "info",
			Title:       "AI Agent 未绑定知识库",
			Description: "部分启用中的 AI Agent 尚未绑定知识库，回答质量可能不稳定。",
			Count:       aiAgentWithoutKnowledgeCount,
			Link:        "/dashboard/ai-agents",
		})
	}

	sort.Slice(alerts, func(i, j int) bool {
		if alerts[i].Count == alerts[j].Count {
			return alerts[i].ID < alerts[j].ID
		}
		return alerts[i].Count > alerts[j].Count
	})

	return alerts
}

func buildConversationStatusDistribution(db *gorm.DB) []response.DashboardStatusDistributionItem {
	statuses := []enums.IMConversationStatus{
		enums.IMConversationStatusPending,
		enums.IMConversationStatusActive,
		enums.IMConversationStatusClosed,
		enums.IMConversationStatusArchived,
	}
	ret := make([]response.DashboardStatusDistributionItem, 0, len(statuses))
	for _, status := range statuses {
		ret = append(ret, response.DashboardStatusDistributionItem{
			Status: int(status),
			Label:  labelOrDefault(enums.GetIMConversationStatusLabel(status), fmt.Sprintf("状态 %d", status)),
			Count: repositories.DashboardRepository.CountConversations(db, func(tx *gorm.DB) *gorm.DB {
				return tx.Where("status = ?", status)
			}),
		})
	}
	return ret
}

func buildTicketStatusDistribution(db *gorm.DB) []response.DashboardStatusDistributionItem {
	statuses := []enums.TicketStatus{
		enums.TicketStatusPending,
		enums.TicketStatusProcessing,
		enums.TicketStatusWaiting,
		enums.TicketStatusResolved,
		enums.TicketStatusClosed,
		enums.TicketStatusCancelled,
	}
	ret := make([]response.DashboardStatusDistributionItem, 0, len(statuses))
	for _, status := range statuses {
		ret = append(ret, response.DashboardStatusDistributionItem{
			Status: int(status),
			Label:  enums.GetTicketStatusLabel(status),
			Count: repositories.DashboardRepository.CountTickets(db, func(tx *gorm.DB) *gorm.DB {
				return tx.Where("status = ?", status)
			}),
		})
	}
	return ret
}

func buildConversationTrend(db *gorm.DB, start time.Time) []response.DashboardTrendItem {
	created := repositories.DashboardRepository.ListConversations(db, func(tx *gorm.DB) *gorm.DB {
		return tx.Select("created_at").Where("created_at >= ?", start)
	})
	closed := repositories.DashboardRepository.ListConversations(db, func(tx *gorm.DB) *gorm.DB {
		return tx.Select("closed_at").Where("closed_at IS NOT NULL AND closed_at >= ?", start)
	})
	return buildTrendItems(start, created, closed, func(item models.Conversation) *time.Time {
		return &item.CreatedAt
	}, func(item models.Conversation) *time.Time {
		return item.ClosedAt
	})
}

func buildTicketTrend(db *gorm.DB, start time.Time) []response.DashboardTrendItem {
	created := repositories.DashboardRepository.ListTickets(db, func(tx *gorm.DB) *gorm.DB {
		return tx.Select("created_at").Where("created_at >= ?", start)
	})
	closed := repositories.DashboardRepository.ListTickets(db, func(tx *gorm.DB) *gorm.DB {
		return tx.Select("closed_at").Where("closed_at IS NOT NULL AND closed_at >= ?", start)
	})
	return buildTicketTrendItems(start, created, closed)
}

func buildTrendItems(start time.Time, created []models.Conversation, closed []models.Conversation, createdAt func(models.Conversation) *time.Time, closedAt func(models.Conversation) *time.Time) []response.DashboardTrendItem {
	series := initTrendMap(start, time.Now())
	for _, item := range created {
		if ts := createdAt(item); ts != nil {
			series[ts.Format("2006-01-02")].NewCount++
		}
	}
	for _, item := range closed {
		if ts := closedAt(item); ts != nil {
			series[ts.Format("2006-01-02")].ClosedCount++
		}
	}
	return flattenTrendMap(series)
}

func buildTicketTrendItems(start time.Time, created []models.Ticket, closed []models.Ticket) []response.DashboardTrendItem {
	series := initTrendMap(start, time.Now())
	for _, item := range created {
		series[item.CreatedAt.Format("2006-01-02")].NewCount++
	}
	for _, item := range closed {
		if item.ClosedAt == nil {
			continue
		}
		series[item.ClosedAt.Format("2006-01-02")].ClosedCount++
	}
	return flattenTrendMap(series)
}

func initTrendMap(start, end time.Time) map[string]*response.DashboardTrendItem {
	series := make(map[string]*response.DashboardTrendItem)
	for current := startOfDay(start); !current.After(end); current = current.AddDate(0, 0, 1) {
		key := current.Format("2006-01-02")
		series[key] = &response.DashboardTrendItem{Date: key}
	}
	return series
}

func flattenTrendMap(series map[string]*response.DashboardTrendItem) []response.DashboardTrendItem {
	keys := make([]string, 0, len(series))
	for key := range series {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	ret := make([]response.DashboardTrendItem, 0, len(keys))
	for _, key := range keys {
		ret = append(ret, *series[key])
	}
	return ret
}

func buildDashboardQuickLinks() []response.DashboardQuickLinkResponse {
	return []response.DashboardQuickLinkResponse{
		{Title: "会话管理", Description: "查看待接入与处理中会话", Link: "/dashboard/conversations"},
		{Title: "工单列表", Description: "处理待办工单与状态流转", Link: "/dashboard/tickets"},
		{Title: "客服档案", Description: "查看客服状态与分组配置", Link: "/dashboard/agents"},
		{Title: "知识库", Description: "维护文档与查看检索日志", Link: "/dashboard/knowledge"},
		{Title: "AI Agent", Description: "配置 AI 接待策略与知识绑定", Link: "/dashboard/ai-agents"},
		{Title: "接入站点", Description: "管理小组件站点与默认 Agent", Link: "/dashboard/widget-sites"},
	}
}

func normalizeDashboardRange(value string) (string, int) {
	switch value {
	case "30d":
		return "30d", 30
	case "today":
		return "today", 1
	default:
		return "7d", 7
	}
}

func startOfDay(value time.Time) time.Time {
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, value.Location())
}

func calcRate(numerator, denominator int64) float64 {
	if denominator <= 0 {
		return 0
	}
	ratio := float64(numerator) / float64(denominator) * 100
	return float64(int(ratio*10+0.5)) / 10
}

func calcAIServiceRate(conversations []models.Conversation) float64 {
	var aiCount int64
	var total int64
	for _, item := range conversations {
		if item.Status == enums.IMConversationStatusArchived {
			continue
		}
		total++
		if item.ServiceMode == enums.IMConversationServiceModeAIOnly || item.ServiceMode == enums.IMConversationServiceModeAIFirst {
			aiCount++
		}
	}
	return calcRate(aiCount, total)
}

func labelOrDefault(value, fallback string) string {
	if strings.TrimSpace(value) != "" {
		return value
	}
	return fallback
}
