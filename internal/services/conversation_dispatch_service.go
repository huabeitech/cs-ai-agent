package services

import (
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/dto"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/pkg/utils"
	"cs-agent/internal/repositories"
	"errors"
	"log/slog"
	"math"
	"slices"
	"strings"
	"sync/atomic"
	"time"

	"github.com/mlogclub/simple/sqls"
)

var ConversationDispatchService = newConversationDispatchService()

func newConversationDispatchService() *conversationDispatchService {
	return &conversationDispatchService{}
}

type conversationDispatchService struct{}

type dispatchCandidate struct {
	profile     models.AgentProfile
	activeCount int
	loadRate    float64
}

type agentActiveConversationCount struct {
	CurrentAssigneeID int64 `gorm:"column:current_assignee_id"`
	ActiveCount       int   `gorm:"column:active_count"`
}

type dispatchPoolReport struct {
	RequestedTeamIDs    []int64
	ActiveScheduleTeams []int64
	MatchedProfiles     int
	EligibleProfiles    int
	CandidateCount      int
	Reason              string
}

var errConversationDispatchConflict = errors.New("conversation dispatch conflict")

const pendingDispatchBatchLimit = 50

var pendingDispatchRunning atomic.Bool

func (s *conversationDispatchService) DispatchConversation(conversationID int64) (*models.Conversation, error) {
	if conversationID <= 0 {
		return nil, nil
	}
	conversation := ConversationService.Get(conversationID)
	if conversation == nil {
		return nil, nil
	}
	if conversation.Status != enums.IMConversationStatusPending || conversation.CurrentAssigneeID > 0 {
		return nil, nil
	}
	aiAgent := AIAgentService.Get(conversation.AIAgentID)
	if aiAgent == nil || aiAgent.Status != enums.StatusOk {
		return nil, nil
	}
	return s.DispatchPendingConversation(conversation, aiAgent)
}

func (s *conversationDispatchService) DispatchPendingConversation(conversation *models.Conversation, aiAgent *models.AIAgent) (*models.Conversation, error) {
	if conversation == nil || aiAgent == nil {
		return nil, nil
	}
	if conversation.Status != enums.IMConversationStatusPending || conversation.CurrentAssigneeID > 0 {
		return nil, nil
	}

	teamIDs := utils.SplitInt64s(aiAgent.TeamIDs)
	if len(teamIDs) == 0 {
		slog.Debug("skip auto dispatch due to empty ai agent team ids",
			"conversation_id", conversation.ID,
			"ai_agent_id", aiAgent.ID,
		)
		return nil, nil
	}

	candidates, report, err := s.pickDispatchCandidates(teamIDs, time.Now())
	if err != nil {
		return nil, err
	}
	if len(candidates) == 0 {
		slog.Debug("no dispatch candidate available",
			"conversation_id", conversation.ID,
			"ai_agent_id", aiAgent.ID,
			"requested_team_ids", report.RequestedTeamIDs,
			"active_schedule_team_ids", report.ActiveScheduleTeams,
			"matched_profiles", report.MatchedProfiles,
			"eligible_profiles", report.EligibleProfiles,
			"reason", report.Reason,
		)
		return nil, nil
	}

	for _, candidate := range candidates {
		dispatched, err := s.tryAssignConversation(conversation.ID, candidate.profile, "自动分配")
		if err != nil {
			if errors.Is(err, errConversationDispatchConflict) {
				return nil, nil
			}
			return nil, err
		}
		if dispatched != nil {
			slog.Info("conversation auto dispatched",
				"conversation_id", dispatched.ID,
				"ai_agent_id", aiAgent.ID,
				"assignee_id", dispatched.CurrentAssigneeID,
				"team_id", dispatched.CurrentTeamID,
				"candidate_count", report.CandidateCount,
				"requested_team_ids", report.RequestedTeamIDs,
			)
			WsService.PublishConversationChanged(dispatched, enums.IMRealtimeEventConversationAssigned)
			return dispatched, nil
		}
	}
	slog.Debug("auto dispatch candidate list exhausted without assignment",
		"conversation_id", conversation.ID,
		"ai_agent_id", aiAgent.ID,
		"candidate_count", report.CandidateCount,
	)
	return nil, nil
}

func (s *conversationDispatchService) DispatchPendingConversations(limit int) (int, error) {
	if !pendingDispatchRunning.CompareAndSwap(false, true) {
		return 0, nil
	}
	defer pendingDispatchRunning.Store(false)

	if limit <= 0 {
		limit = pendingDispatchBatchLimit
	}
	conversations := ConversationService.Find(sqls.NewCnd().
		Eq("status", enums.IMConversationStatusPending).
		Eq("current_assignee_id", 0).
		Desc("id"))
	if len(conversations) == 0 {
		return 0, nil
	}

	dispatchedCount := 0
	scannedCount := 0
	for i, conversation := range conversations {
		if i >= limit {
			break
		}
		scannedCount++
		dispatched, err := s.DispatchConversation(conversation.ID)
		if err != nil {
			return dispatchedCount, err
		}
		if dispatched != nil {
			dispatchedCount++
		}
	}
	if scannedCount > 0 {
		slog.Info("pending conversation dispatch scan completed",
			"scanned_count", scannedCount,
			"dispatched_count", dispatchedCount,
			"limit", limit,
		)
	}
	return dispatchedCount, nil
}

func (s *conversationDispatchService) RunPendingDispatchLoop(interval time.Duration) {
	if interval <= 0 {
		interval = 30 * time.Second
	}
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		slog.Info("pending conversation dispatch loop started",
			"interval_seconds", int(interval/time.Second),
		)

		for {
			if _, err := s.DispatchPendingConversations(0); err != nil {
				slog.Warn("dispatch pending conversations loop failed", "error", err)
			}
			<-ticker.C
		}
	}()
}

// pickDispatchCandidates returns the eligible dispatch candidates for the given teamIDs at the given time, along with a report for debugging and analysis.
func (s *conversationDispatchService) pickDispatchCandidates(teamIDs []int64, now time.Time) ([]dispatchCandidate, dispatchPoolReport, error) {
	report := dispatchPoolReport{
		RequestedTeamIDs: append([]int64(nil), teamIDs...),
	}

	// 1. filter teams with active schedule
	activeTeamIDs := s.findActiveScheduleTeamIDs(teamIDs, now)
	report.ActiveScheduleTeams = activeTeamIDs
	if len(activeTeamIDs) == 0 {
		report.Reason = "no_active_schedule_team"
		return nil, report, nil
	}

	// 2. find agent profiles for the active teams
	profiles := AgentProfileService.GetDispatchAgents(activeTeamIDs)
	report.MatchedProfiles = len(profiles)
	if len(profiles) == 0 {
		report.Reason = "no_matched_profile"
		return nil, report, nil
	}

	enabledProfiles, enabledUserIDs, reason := s.filterEnabledDispatchProfiles(profiles)
	if reason != "" {
		report.Reason = reason
		return nil, report, nil
	}
	report.EligibleProfiles = len(enabledProfiles)

	activeCounts, err := s.findActiveConversationCountMap(enabledUserIDs)
	if err != nil {
		return nil, report, err
	}

	candidates := make([]dispatchCandidate, 0, len(enabledProfiles))
	for _, profile := range enabledProfiles {
		activeCount := activeCounts[profile.UserID]
		if profile.MaxConcurrentCount > 0 && activeCount >= profile.MaxConcurrentCount {
			continue
		}
		loadRate := float64(activeCount) / math.Max(float64(profile.MaxConcurrentCount), 1)
		candidates = append(candidates, dispatchCandidate{
			profile:     profile,
			activeCount: activeCount,
			loadRate:    loadRate,
		})
	}
	report.CandidateCount = len(candidates)
	if len(candidates) == 0 {
		report.Reason = "all_candidates_at_capacity"
		return nil, report, nil
	}

	slices.SortFunc(candidates, func(a, b dispatchCandidate) int {
		switch {
		case a.loadRate < b.loadRate:
			return -1
		case a.loadRate > b.loadRate:
			return 1
		}
		switch {
		case a.activeCount < b.activeCount:
			return -1
		case a.activeCount > b.activeCount:
			return 1
		}
		switch {
		case a.profile.PriorityLevel > b.profile.PriorityLevel:
			return -1
		case a.profile.PriorityLevel < b.profile.PriorityLevel:
			return 1
		}
		aLastStatusAt := zeroTime(a.profile.LastStatusAt)
		bLastStatusAt := zeroTime(b.profile.LastStatusAt)
		switch {
		case aLastStatusAt.Before(bLastStatusAt):
			return -1
		case aLastStatusAt.After(bLastStatusAt):
			return 1
		}
		switch {
		case a.profile.UserID < b.profile.UserID:
			return -1
		case a.profile.UserID > b.profile.UserID:
			return 1
		default:
			return 0
		}
	})
	report.Reason = "ok"
	return candidates, report, nil
}

func (s *conversationDispatchService) filterEnabledDispatchProfiles(profiles []models.AgentProfile) ([]models.AgentProfile, []int64, string) {
	userIDs := make([]int64, 0, len(profiles))
	for _, profile := range profiles {
		if profile.UserID > 0 {
			userIDs = append(userIDs, profile.UserID)
		}
	}
	if len(userIDs) == 0 {
		return nil, nil, "no_profile_with_capacity_config"
	}

	enabledUsers := UserService.Find(sqls.NewCnd().
		In("id", userIDs).
		Eq("status", enums.StatusOk))
	if len(enabledUsers) == 0 {
		return nil, nil, "no_enabled_user"
	}

	enabledUserSet := make(map[int64]struct{}, len(enabledUsers))
	for _, user := range enabledUsers {
		enabledUserSet[user.ID] = struct{}{}
	}

	enabledProfiles := make([]models.AgentProfile, 0, len(profiles))
	enabledUserIDs := make([]int64, 0, len(profiles))
	for _, profile := range profiles {
		if _, exists := enabledUserSet[profile.UserID]; !exists {
			continue
		}
		enabledProfiles = append(enabledProfiles, profile)
		enabledUserIDs = append(enabledUserIDs, profile.UserID)
	}
	if len(enabledProfiles) == 0 {
		return nil, nil, "no_profile_for_enabled_user"
	}
	return enabledProfiles, enabledUserIDs, ""
}

// findActiveScheduleTeamIDs returns the subset of teamIDs that have active schedule at the given time.
func (s *conversationDispatchService) findActiveScheduleTeamIDs(teamIDs []int64, now time.Time) []int64 {
	if len(teamIDs) == 0 {
		return nil
	}

	teams := AgentTeamService.Find(sqls.NewCnd().
		In("id", teamIDs).
		Eq("status", enums.StatusOk))
	if len(teams) == 0 {
		return nil
	}

	enabledTeamIDs := make([]int64, 0, len(teams))
	for _, team := range teams {
		enabledTeamIDs = append(enabledTeamIDs, team.ID)
	}

	schedules := AgentTeamScheduleService.Find(sqls.NewCnd().
		In("team_id", enabledTeamIDs).
		Eq("status", enums.StatusOk).
		Lte("start_at", now).
		Gt("end_at", now))

	ret := make([]int64, 0, len(schedules))
	seen := make(map[int64]struct{})
	for _, schedule := range schedules {
		if _, exists := seen[schedule.TeamID]; exists {
			continue
		}
		seen[schedule.TeamID] = struct{}{}
		ret = append(ret, schedule.TeamID)
	}
	return ret
}

func (s *conversationDispatchService) findActiveConversationCountMap(userIDs []int64) (map[int64]int, error) {
	ret := make(map[int64]int, len(userIDs))
	if len(userIDs) == 0 {
		return ret, nil
	}

	rows := make([]agentActiveConversationCount, 0)
	if err := sqls.DB().
		Model(&models.Conversation{}).
		Select("current_assignee_id, COUNT(1) AS active_count").
		Where("status = ? AND current_assignee_id IN ?", enums.IMConversationStatusActive, userIDs).
		Group("current_assignee_id").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	for _, row := range rows {
		if row.CurrentAssigneeID <= 0 {
			continue
		}
		ret[row.CurrentAssigneeID] = row.ActiveCount
	}
	return ret, nil
}

func (s *conversationDispatchService) tryAssignConversation(conversationID int64, candidate models.AgentProfile, reason string) (*models.Conversation, error) {
	now := time.Now()
	operator := systemDispatchPrincipal()

	err := sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		conversation := repositories.ConversationRepository.Get(ctx.Tx, conversationID)
		if conversation == nil {
			return errConversationDispatchConflict
		}
		if conversation.Status != enums.IMConversationStatusPending || conversation.CurrentAssigneeID > 0 {
			return errConversationDispatchConflict
		}

		if err := ConversationAssignmentService.FinishActiveAssignmentsTx(ctx, conversationID, now); err != nil {
			return err
		}
		if err := ConversationAssignmentService.CreateAssignmentTx(ctx, conversationID, conversation.CurrentAssigneeID, candidate.UserID, enums.IMAssignmentTypeAssign, reason, operator, now); err != nil {
			return err
		}

		result := ctx.Tx.Model(&models.Conversation{}).
			Where("id = ? AND status = ? AND current_assignee_id = ?", conversationID, enums.IMConversationStatusPending, 0).
			Updates(map[string]any{
				"current_assignee_id": candidate.UserID,
				"current_team_id":     candidate.TeamID,
				"status":              enums.IMConversationStatusActive,
				"update_user_id":      operator.UserID,
				"update_user_name":    operator.Username,
				"updated_at":          now,
			})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return errConversationDispatchConflict
		}

		return ConversationEventLogService.CreateEvent(ctx, conversationID, enums.IMEventTypeAssign, enums.IMSenderTypeSystem, operator.UserID, "会话已自动分配", buildDispatchEventPayload(conversation.CurrentAssigneeID, candidate.UserID, candidate.TeamID, reason), now)
	})
	if err != nil {
		return nil, err
	}
	return ConversationService.Get(conversationID), nil
}

func buildDispatchEventPayload(fromAssigneeID, toAssigneeID, toTeamID int64, reason string) string {
	return ConversationService.buildEventPayload(map[string]any{
		"fromStatus":     enums.IMConversationStatusPending,
		"toStatus":       enums.IMConversationStatusActive,
		"fromAssigneeId": fromAssigneeID,
		"toAssigneeId":   toAssigneeID,
		"toTeamId":       toTeamID,
		"reason":         strings.TrimSpace(reason),
	})
}

func systemDispatchPrincipal() *dto.AuthPrincipal {
	return &dto.AuthPrincipal{
		UserID:   0,
		Username: "system",
		Nickname: "system",
	}
}

func zeroTime(value *time.Time) time.Time {
	if value == nil {
		return time.Time{}
	}
	return *value
}
