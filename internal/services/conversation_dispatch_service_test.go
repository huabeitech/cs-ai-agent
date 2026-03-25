package services

import (
	"context"
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/constants"
	"cs-agent/internal/pkg/dto"
	"cs-agent/internal/pkg/dto/request"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/pkg/utils"
	"testing"
	"time"

	"github.com/mlogclub/simple/sqls"
)

func TestDispatchConversationAssignsLowestLoadCandidateAcrossMatchedTeams(t *testing.T) {
	setupIMServiceTestDB(t)

	team1 := seedDispatchAgentTeam(t, 1001, "售前1组", "pre_sales_1")
	team2 := seedDispatchAgentTeam(t, 1002, "售前2组", "pre_sales_2")

	seedAgentUser(t, 2001, "agent_dispatch_1")
	seedAgentUser(t, 2002, "agent_dispatch_2")

	seedEnabledAgentProfile(t, 3001, 2001, team1.ID, enums.ServiceStatusIdle, 2, 10)
	seedEnabledAgentProfile(t, 3002, 2002, team2.ID, enums.ServiceStatusIdle, 1, 5)

	now := time.Now()
	seedEnabledTeamSchedule(t, 4001, team1.ID, now.Add(-time.Hour), now.Add(time.Hour))
	seedEnabledTeamSchedule(t, 4002, team2.ID, now.Add(-time.Hour), now.Add(time.Hour))

	aiAgent := seedEnabledAIAgent(t, 5001, "调度AI")
	aiAgent.TeamIDs = utils.JoinInt64s([]int64{team1.ID, team2.ID})
	if err := sqls.DB().Model(&models.AIAgent{}).Where("id = ?", aiAgent.ID).Update("team_ids", aiAgent.TeamIDs).Error; err != nil {
		t.Fatalf("update ai agent team ids failed: %v", err)
	}

	// 给 team1 的客服先占一个处理中会话，制造更高负载。
	seedConversation(t, &models.Conversation{
		ID:                6001,
		AIAgentID:         aiAgent.ID,
		ChannelType:       enums.IMConversationChannelWebChat,
		Subject:           "已分配会话",
		Status:            enums.IMConversationStatusActive,
		ServiceMode:       enums.IMConversationServiceModeAIFirst,
		CurrentAssigneeID: 2001,
		CurrentTeamID:     team1.ID,
		AuditFields:       buildTestAuditFields(0, "system"),
	})

	pendingConversation := &models.Conversation{
		ID:            6002,
		AIAgentID:     aiAgent.ID,
		ChannelType:   enums.IMConversationChannelWebChat,
		Subject:       "待自动分配会话",
		Status:        enums.IMConversationStatusPending,
		ServiceMode:   enums.IMConversationServiceModeAIFirst,
		AuditFields:   buildTestAuditFields(0, "system"),
		HandoffReason: "用户主动要求人工",
	}
	seedConversation(t, pendingConversation)

	dispatched, err := ConversationDispatchService.DispatchConversation(pendingConversation.ID)
	if err != nil {
		t.Fatalf("dispatch conversation failed: %v", err)
	}
	if dispatched == nil {
		t.Fatalf("expected dispatched conversation, got nil")
	}
	if dispatched.Status != enums.IMConversationStatusActive {
		t.Fatalf("expected active conversation, got %d", dispatched.Status)
	}
	if dispatched.CurrentAssigneeID != 2002 {
		t.Fatalf("expected assignee 2002, got %d", dispatched.CurrentAssigneeID)
	}
	if dispatched.CurrentTeamID != team2.ID {
		t.Fatalf("expected team %d, got %d", team2.ID, dispatched.CurrentTeamID)
	}

	assignment := ConversationAssignmentService.Take("conversation_id = ? AND status = ?", pendingConversation.ID, enums.IMAssignmentStatusActive)
	if assignment == nil {
		t.Fatalf("expected active assignment record")
	}
	if assignment.ToUserID != 2002 {
		t.Fatalf("expected assignment to user 2002, got %d", assignment.ToUserID)
	}

	event := ConversationEventLogService.Take("conversation_id = ? AND event_type = ?", pendingConversation.ID, enums.IMEventTypeAssign)
	if event == nil {
		t.Fatalf("expected assign event log")
	}
}

func TestDispatchConversationKeepsPendingWhenNoScheduleMatches(t *testing.T) {
	setupIMServiceTestDB(t)

	team := seedDispatchAgentTeam(t, 1101, "售前夜班组", "pre_sales_night")
	seedAgentUser(t, 2101, "agent_pending_keep")
	seedEnabledAgentProfile(t, 3101, 2101, team.ID, enums.ServiceStatusIdle, 2, 10)

	aiAgent := seedEnabledAIAgent(t, 5101, "无排班AI")
	aiAgent.TeamIDs = utils.JoinInt64s([]int64{team.ID})
	if err := sqls.DB().Model(&models.AIAgent{}).Where("id = ?", aiAgent.ID).Update("team_ids", aiAgent.TeamIDs).Error; err != nil {
		t.Fatalf("update ai agent team ids failed: %v", err)
	}

	pendingConversation := &models.Conversation{
		ID:          6101,
		AIAgentID:   aiAgent.ID,
		ChannelType: enums.IMConversationChannelWebChat,
		Subject:     "待自动分配会话",
		Status:      enums.IMConversationStatusPending,
		ServiceMode: enums.IMConversationServiceModeAIFirst,
		AuditFields: buildTestAuditFields(0, "system"),
	}
	seedConversation(t, pendingConversation)

	dispatched, err := ConversationDispatchService.DispatchConversation(pendingConversation.ID)
	if err != nil {
		t.Fatalf("dispatch conversation failed: %v", err)
	}
	if dispatched != nil {
		t.Fatalf("expected no dispatch result, got %+v", dispatched)
	}

	stored := ConversationService.Get(pendingConversation.ID)
	if stored == nil {
		t.Fatalf("expected pending conversation")
	}
	if stored.Status != enums.IMConversationStatusPending {
		t.Fatalf("expected pending status, got %d", stored.Status)
	}
	if stored.CurrentAssigneeID != 0 {
		t.Fatalf("expected no assignee, got %d", stored.CurrentAssigneeID)
	}
}

func TestCreateConversationAutoDispatchesWithMatchedTeamSchedule(t *testing.T) {
	setupIMServiceTestDB(t)

	team := seedDispatchAgentTeam(t, 1201, "人工接待组", "human_support")
	seedAgentUser(t, 2201, "agent_create_dispatch")
	seedEnabledAgentProfile(t, 3201, 2201, team.ID, enums.ServiceStatusIdle, 3, 8)
	now := time.Now()
	seedEnabledTeamSchedule(t, 4201, team.ID, now.Add(-time.Hour), now.Add(time.Hour))

	aiAgent := seedEnabledAIAgent(t, 5201, "人工入口AI")
	aiAgent.ServiceMode = enums.IMConversationServiceModeHumanOnly
	aiAgent.TeamIDs = utils.JoinInt64s([]int64{team.ID})
	if err := sqls.DB().Model(&models.AIAgent{}).Where("id = ?", aiAgent.ID).Updates(map[string]any{
		"service_mode": aiAgent.ServiceMode,
		"team_ids":     aiAgent.TeamIDs,
	}).Error; err != nil {
		t.Fatalf("update ai agent create dispatch fields failed: %v", err)
	}

	customer := &dto.AuthPrincipal{UserID: 1301, Username: "customer_create_dispatch"}
	conversation, err := ConversationService.Create(enums.IMConversationChannelWebChat, "人工入口自动派单", aiAgent.ID, customer)
	if err != nil {
		t.Fatalf("create conversation failed: %v", err)
	}
	if conversation == nil {
		t.Fatalf("expected created conversation")
	}
	if conversation.Status != enums.IMConversationStatusActive {
		t.Fatalf("expected active status after create, got %d", conversation.Status)
	}
	if conversation.CurrentAssigneeID != 2201 {
		t.Fatalf("expected assignee 2201, got %d", conversation.CurrentAssigneeID)
	}
	if conversation.CurrentTeamID != team.ID {
		t.Fatalf("expected team %d, got %d", team.ID, conversation.CurrentTeamID)
	}
}

func TestCreateConversationKeepsPendingWhenAIAgentServiceModeIsAIFirst(t *testing.T) {
	setupIMServiceTestDB(t)

	team := seedDispatchAgentTeam(t, 1202, "AI优先人工组", "ai_first_support")
	seedAgentUser(t, 2202, "agent_ai_first")
	seedEnabledAgentProfile(t, 3202, 2202, team.ID, enums.ServiceStatusIdle, 3, 8)
	now := time.Now()
	seedEnabledTeamSchedule(t, 4202, team.ID, now.Add(-time.Hour), now.Add(time.Hour))

	aiAgent := seedEnabledAIAgent(t, 5202, "AI优先入口")
	aiAgent.ServiceMode = enums.IMConversationServiceModeAIFirst
	aiAgent.TeamIDs = utils.JoinInt64s([]int64{team.ID})
	if err := sqls.DB().Model(&models.AIAgent{}).Where("id = ?", aiAgent.ID).Updates(map[string]any{
		"service_mode": aiAgent.ServiceMode,
		"team_ids":     aiAgent.TeamIDs,
	}).Error; err != nil {
		t.Fatalf("update ai agent ai first fields failed: %v", err)
	}

	customer := &dto.AuthPrincipal{UserID: 1302, Username: "customer_ai_first"}
	conversation, err := ConversationService.Create(enums.IMConversationChannelWebChat, "AI优先入口不自动派单", aiAgent.ID, customer)
	if err != nil {
		t.Fatalf("create conversation failed: %v", err)
	}
	if conversation == nil {
		t.Fatalf("expected created conversation")
	}
	if conversation.Status != enums.IMConversationStatusPending {
		t.Fatalf("expected pending status, got %d", conversation.Status)
	}
	if conversation.CurrentAssigneeID != 0 {
		t.Fatalf("expected no assignee, got %d", conversation.CurrentAssigneeID)
	}
	if conversation.CurrentTeamID != 0 {
		t.Fatalf("expected no current team, got %d", conversation.CurrentTeamID)
	}
}

func TestTriggerReplyAutoHandoffsAndDispatchesWhenMaxAIRoundsReached(t *testing.T) {
	setupIMServiceTestDB(t)

	team := seedDispatchAgentTeam(t, 1203, "AI转人工组", "ai_handoff_team")
	seedAgentUser(t, 2203, "agent_ai_handoff")
	seedEnabledAgentProfile(t, 3203, 2203, team.ID, enums.ServiceStatusIdle, 3, 8)
	now := time.Now()
	seedEnabledTeamSchedule(t, 4203, team.ID, now.Add(-time.Hour), now.Add(time.Hour))

	aiAgent := seedEnabledAIAgent(t, 5203, "AI转人工入口")
	aiAgent.ServiceMode = enums.IMConversationServiceModeAIFirst
	aiAgent.MaxAIReplyRounds = 1
	aiAgent.TeamIDs = utils.JoinInt64s([]int64{team.ID})
	if err := sqls.DB().Model(&models.AIAgent{}).Where("id = ?", aiAgent.ID).Updates(map[string]any{
		"service_mode":        aiAgent.ServiceMode,
		"max_ai_reply_rounds": aiAgent.MaxAIReplyRounds,
		"team_ids":            aiAgent.TeamIDs,
	}).Error; err != nil {
		t.Fatalf("update ai agent handoff fields failed: %v", err)
	}

	customer := &dto.AuthPrincipal{UserID: 1303, Username: "customer_ai_handoff"}
	conversation, err := ConversationService.Create(enums.IMConversationChannelWebChat, "达到上限自动转人工", aiAgent.ID, customer)
	if err != nil {
		t.Fatalf("create conversation failed: %v", err)
	}
	if conversation == nil {
		t.Fatalf("expected created conversation")
	}
	if err := ConversationService.Updates(conversation.ID, map[string]any{
		"ai_reply_rounds": aiAgent.MaxAIReplyRounds,
	}); err != nil {
		t.Fatalf("update ai reply rounds failed: %v", err)
	}

	message, err := MessageService.SendCustomerMessage(conversation.ID, "max_round_handoff_1", enums.IMMessageTypeText, "继续追问一个问题", "", customer)
	if err != nil {
		t.Fatalf("send customer message failed: %v", err)
	}
	if err := AIReplyService.TriggerReply(context.Background(), message.ID); err != nil {
		t.Fatalf("trigger ai reply failed: %v", err)
	}

	updated := ConversationService.Get(conversation.ID)
	if updated == nil {
		t.Fatalf("expected updated conversation")
	}
	if updated.Status != enums.IMConversationStatusActive {
		t.Fatalf("expected active status after handoff dispatch, got %d", updated.Status)
	}
	if updated.CurrentAssigneeID != 2203 {
		t.Fatalf("expected assignee 2203, got %d", updated.CurrentAssigneeID)
	}
	if updated.CurrentTeamID != team.ID {
		t.Fatalf("expected team %d, got %d", team.ID, updated.CurrentTeamID)
	}
	if updated.HandoffAt == nil {
		t.Fatalf("expected handoff time")
	}
	if updated.HandoffReason != "达到AI最大回复轮次" {
		t.Fatalf("expected handoff reason, got %s", updated.HandoffReason)
	}

	handoffMessage := MessageService.Take("conversation_id = ? AND client_msg_id = ?", conversation.ID, "ai_handoff_"+utils.JoinInt64s([]int64{message.ID}))
	if handoffMessage == nil {
		t.Fatalf("expected ai handoff message")
	}

	assignment := ConversationAssignmentService.Take("conversation_id = ? AND status = ?", conversation.ID, enums.IMAssignmentStatusActive)
	if assignment == nil {
		t.Fatalf("expected active assignment")
	}
	if assignment.ToUserID != 2203 {
		t.Fatalf("expected assignment to 2203, got %d", assignment.ToUserID)
	}
}

func TestConversationServiceDispatchConversation(t *testing.T) {
	setupIMServiceTestDB(t)

	team := seedDispatchAgentTeam(t, 1204, "后台重试分配组", "console_dispatch")
	seedAgentUser(t, 2204, "agent_console_dispatch")
	seedEnabledAgentProfile(t, 3204, 2204, team.ID, enums.ServiceStatusIdle, 3, 8)
	now := time.Now()
	seedEnabledTeamSchedule(t, 4204, team.ID, now.Add(-time.Hour), now.Add(time.Hour))

	aiAgent := seedEnabledAIAgent(t, 5204, "后台重试分配AI")
	aiAgent.TeamIDs = utils.JoinInt64s([]int64{team.ID})
	if err := sqls.DB().Model(&models.AIAgent{}).Where("id = ?", aiAgent.ID).Update("team_ids", aiAgent.TeamIDs).Error; err != nil {
		t.Fatalf("update ai agent team ids failed: %v", err)
	}

	pendingConversation := &models.Conversation{
		ID:          6204,
		AIAgentID:   aiAgent.ID,
		ChannelType: enums.IMConversationChannelWebChat,
		Subject:     "后台手动重试分配",
		Status:      enums.IMConversationStatusPending,
		ServiceMode: enums.IMConversationServiceModeHumanOnly,
		AuditFields: buildTestAuditFields(0, "system"),
	}
	seedConversation(t, pendingConversation)

	admin := &dto.AuthPrincipal{UserID: 9001, Username: "admin", Roles: []string{constants.RoleCodeAdmin}}
	if err := ConversationService.DispatchConversation(pendingConversation.ID, admin); err != nil {
		t.Fatalf("dispatch conversation via conversation service failed: %v", err)
	}

	updated := ConversationService.Get(pendingConversation.ID)
	if updated == nil {
		t.Fatalf("expected updated conversation")
	}
	if updated.Status != enums.IMConversationStatusActive {
		t.Fatalf("expected active status, got %d", updated.Status)
	}
	if updated.CurrentAssigneeID != 2204 {
		t.Fatalf("expected assignee 2204, got %d", updated.CurrentAssigneeID)
	}
}

func TestUpdateAgentProfileDispatchesPendingConversation(t *testing.T) {
	setupIMServiceTestDB(t)

	team := seedDispatchAgentTeam(t, 1205, "客服上线触发组", "agent_status_dispatch")
	seedAgentUser(t, 2205, "agent_status_dispatch")
	now := time.Now()
	seedEnabledTeamSchedule(t, 4205, team.ID, now.Add(-time.Hour), now.Add(time.Hour))

	aiAgent := seedEnabledAIAgent(t, 5205, "客服上线触发AI")
	aiAgent.TeamIDs = utils.JoinInt64s([]int64{team.ID})
	if err := sqls.DB().Model(&models.AIAgent{}).Where("id = ?", aiAgent.ID).Update("team_ids", aiAgent.TeamIDs).Error; err != nil {
		t.Fatalf("update ai agent team ids failed: %v", err)
	}

	pendingConversation := &models.Conversation{
		ID:          6205,
		AIAgentID:   aiAgent.ID,
		ChannelType: enums.IMConversationChannelWebChat,
		Subject:     "客服上线后自动分配",
		Status:      enums.IMConversationStatusPending,
		ServiceMode: enums.IMConversationServiceModeHumanOnly,
		AuditFields: buildTestAuditFields(0, "system"),
	}
	seedConversation(t, pendingConversation)

	operator := &dto.AuthPrincipal{UserID: 9002, Username: "operator"}
	profile, err := AgentProfileService.CreateAgentProfile(request.CreateAgentProfileRequest{
		UserID:                2205,
		TeamID:                team.ID,
		AgentCode:             "agent_status_dispatch",
		DisplayName:           "客服上线触发",
		ServiceStatus:         enums.ServiceStatusBusy,
		MaxConcurrentCount:    3,
		PriorityLevel:         8,
		AutoAssignEnabled:     true,
		ReceiveOfflineMessage: false,
	}, operator)
	if err != nil {
		t.Fatalf("create agent profile failed: %v", err)
	}

	if err := AgentProfileService.UpdateAgentProfile(request.UpdateAgentProfileRequest{
		ID: profile.ID,
		CreateAgentProfileRequest: request.CreateAgentProfileRequest{
			UserID:                2205,
			TeamID:                team.ID,
			AgentCode:             "agent_status_dispatch",
			DisplayName:           "客服上线触发",
			ServiceStatus:         enums.ServiceStatusIdle,
			MaxConcurrentCount:    3,
			PriorityLevel:         8,
			AutoAssignEnabled:     true,
			ReceiveOfflineMessage: false,
		},
	}, operator); err != nil {
		t.Fatalf("update agent profile failed: %v", err)
	}

	updated := ConversationService.Get(pendingConversation.ID)
	if updated == nil {
		t.Fatalf("expected updated conversation")
	}
	if updated.Status != enums.IMConversationStatusActive {
		t.Fatalf("expected active status, got %d", updated.Status)
	}
	if updated.CurrentAssigneeID != 2205 {
		t.Fatalf("expected assignee 2205, got %d", updated.CurrentAssigneeID)
	}
}

func TestCreateAgentTeamScheduleDispatchesPendingConversation(t *testing.T) {
	setupIMServiceTestDB(t)

	team := seedDispatchAgentTeam(t, 1206, "排班生效触发组", "schedule_create_dispatch")
	seedAgentUser(t, 2206, "schedule_create_dispatch")
	_, err := AgentProfileService.CreateAgentProfile(request.CreateAgentProfileRequest{
		UserID:                2206,
		TeamID:                team.ID,
		AgentCode:             "schedule_create_dispatch",
		DisplayName:           "排班生效触发",
		ServiceStatus:         enums.ServiceStatusIdle,
		MaxConcurrentCount:    3,
		PriorityLevel:         8,
		AutoAssignEnabled:     true,
		ReceiveOfflineMessage: false,
	}, &dto.AuthPrincipal{UserID: 9003, Username: "operator"})
	if err != nil {
		t.Fatalf("create agent profile failed: %v", err)
	}

	aiAgent := seedEnabledAIAgent(t, 5206, "排班生效触发AI")
	aiAgent.TeamIDs = utils.JoinInt64s([]int64{team.ID})
	if err := sqls.DB().Model(&models.AIAgent{}).Where("id = ?", aiAgent.ID).Update("team_ids", aiAgent.TeamIDs).Error; err != nil {
		t.Fatalf("update ai agent team ids failed: %v", err)
	}

	pendingConversation := &models.Conversation{
		ID:          6206,
		AIAgentID:   aiAgent.ID,
		ChannelType: enums.IMConversationChannelWebChat,
		Subject:     "排班创建后自动分配",
		Status:      enums.IMConversationStatusPending,
		ServiceMode: enums.IMConversationServiceModeHumanOnly,
		AuditFields: buildTestAuditFields(0, "system"),
	}
	seedConversation(t, pendingConversation)

	now := time.Now()
	if _, err := AgentTeamScheduleService.CreateAgentTeamSchedule(request.CreateAgentTeamScheduleRequest{
		TeamID:     team.ID,
		StartAt:    now.Add(-time.Hour).Format(time.DateTime),
		EndAt:      now.Add(time.Hour).Format(time.DateTime),
		SourceType: "manual",
	}, &dto.AuthPrincipal{UserID: 9003, Username: "operator"}); err != nil {
		t.Fatalf("create agent team schedule failed: %v", err)
	}

	updated := ConversationService.Get(pendingConversation.ID)
	if updated == nil {
		t.Fatalf("expected updated conversation")
	}
	if updated.Status != enums.IMConversationStatusActive {
		t.Fatalf("expected active status, got %d", updated.Status)
	}
	if updated.CurrentAssigneeID != 2206 {
		t.Fatalf("expected assignee 2206, got %d", updated.CurrentAssigneeID)
	}
}

func TestDispatchPendingConversationsSkipsWhenAnotherScanIsRunning(t *testing.T) {
	pendingDispatchRunning.Store(true)
	defer pendingDispatchRunning.Store(false)

	dispatchedCount, err := ConversationDispatchService.DispatchPendingConversations(10)
	if err != nil {
		t.Fatalf("dispatch pending conversations failed: %v", err)
	}
	if dispatchedCount != 0 {
		t.Fatalf("expected skipped scan result 0, got %d", dispatchedCount)
	}
}

func seedDispatchAgentTeam(t *testing.T, id int64, name, _ string) *models.AgentTeam {
	t.Helper()

	item := &models.AgentTeam{
		ID:          id,
		Name:        name,
		Status:      enums.StatusOk,
		Description: name,
		AuditFields: buildTestAuditFields(0, "system"),
	}
	if err := sqls.DB().Create(item).Error; err != nil {
		t.Fatalf("seed agent team failed: %v", err)
	}
	return item
}

func seedEnabledAgentProfile(t *testing.T, id, userID, teamID int64, serviceStatus enums.ServiceStatus, maxConcurrentCount, priorityLevel int) *models.AgentProfile {
	t.Helper()

	lastStatusAt := time.Now().Add(-time.Hour)
	item := &models.AgentProfile{
		ID:                 id,
		UserID:             userID,
		TeamID:             teamID,
		AgentCode:          "code_" + utils.JoinInt64s([]int64{id}),
		DisplayName:        "客服" + utils.JoinInt64s([]int64{userID}),
		ServiceStatus:      serviceStatus,
		MaxConcurrentCount: maxConcurrentCount,
		PriorityLevel:      priorityLevel,
		AutoAssignEnabled:  true,
		Status:             enums.StatusOk,
		LastStatusAt:       &lastStatusAt,
		AuditFields:        buildTestAuditFields(0, "system"),
	}
	if err := sqls.DB().Create(item).Error; err != nil {
		t.Fatalf("seed agent profile failed: %v", err)
	}
	return item
}

func seedEnabledTeamSchedule(t *testing.T, id, teamID int64, startAt, endAt time.Time) *models.AgentTeamSchedule {
	t.Helper()

	item := &models.AgentTeamSchedule{
		ID:          id,
		TeamID:      teamID,
		StartAt:     startAt,
		EndAt:       endAt,
		SourceType:  "manual",
		Status:      enums.StatusOk,
		AuditFields: buildTestAuditFields(0, "system"),
	}
	if err := sqls.DB().Create(item).Error; err != nil {
		t.Fatalf("seed team schedule failed: %v", err)
	}
	return item
}

func seedConversation(t *testing.T, item *models.Conversation) {
	t.Helper()

	now := time.Now()
	item.CreatedAt = now
	item.UpdatedAt = now
	if err := sqls.DB().Create(item).Error; err != nil {
		t.Fatalf("seed conversation failed: %v", err)
	}
}

func buildTestAuditFields(userID int64, username string) models.AuditFields {
	now := time.Now()
	return models.AuditFields{
		CreatedAt:      now,
		CreateUserID:   userID,
		CreateUserName: username,
		UpdatedAt:      now,
		UpdateUserID:   userID,
		UpdateUserName: username,
	}
}
