package response

type DashboardOverviewResponse struct {
	Range             string                        `json:"range"`
	GeneratedAt       string                        `json:"generatedAt"`
	Summary           DashboardSummaryResponse      `json:"summary"`
	ConversationStats DashboardSectionStatsResponse `json:"conversationStats"`
	AgentStats        DashboardAgentStatsResponse   `json:"agentStats"`
	AIStats           DashboardAIStatsResponse      `json:"aiStats"`
	Alerts            []DashboardAlertResponse      `json:"alerts"`
	QuickLinks        []DashboardQuickLinkResponse  `json:"quickLinks"`
}

type DashboardSummaryResponse struct {
	TodayNewConversations        int64   `json:"todayNewConversations"`
	ProcessingConversations      int64   `json:"processingConversations"`
	PendingDispatchConversations int64   `json:"pendingDispatchConversations"`
	OnlineAgents                 int64   `json:"onlineAgents"`
	AIServiceRate                float64 `json:"aiServiceRate"`
}

type DashboardSectionStatsResponse struct {
	StatusDistribution []DashboardStatusDistributionItem `json:"statusDistribution"`
	Trend              []DashboardTrendItem              `json:"trend"`
}

type DashboardStatusDistributionItem struct {
	Status int    `json:"status"`
	Label  string `json:"label"`
	Count  int64  `json:"count"`
}

type DashboardTrendItem struct {
	Date        string `json:"date"`
	NewCount    int64  `json:"newCount"`
	ClosedCount int64  `json:"closedCount"`
}

type DashboardAgentStatsResponse struct {
	OnlineAgents  int64                       `json:"onlineAgents"`
	BusyAgents    int64                       `json:"busyAgents"`
	OfflineAgents int64                       `json:"offlineAgents"`
	TeamLoads     []DashboardTeamLoadResponse `json:"teamLoads"`
}

type DashboardTeamLoadResponse struct {
	TeamID                  int64   `json:"teamId"`
	TeamName                string  `json:"teamName"`
	TotalAgents             int64   `json:"totalAgents"`
	OnlineAgents            int64   `json:"onlineAgents"`
	BusyAgents              int64   `json:"busyAgents"`
	OfflineAgents           int64   `json:"offlineAgents"`
	WaitingConversations    int64   `json:"waitingConversations"`
	ProcessingConversations int64   `json:"processingConversations"`
	MaxConcurrentCapacity   int64   `json:"maxConcurrentCapacity"`
	LoadRate                float64 `json:"loadRate"`
	HasScheduleNow          bool    `json:"hasScheduleNow"`
}

type DashboardAIStatsResponse struct {
	EnabledAIAgents                 int64   `json:"enabledAiAgents"`
	EnabledWidgetSites              int64   `json:"enabledWidgetSites"`
	TodayKnowledgeRetrieves         int64   `json:"todayKnowledgeRetrieves"`
	TodayKnowledgeRetrieveFailCount int64   `json:"todayKnowledgeRetrieveFailCount"`
	TodayKnowledgeRetrieveFailRate  float64 `json:"todayKnowledgeRetrieveFailRate"`
	TodaySkillRunFailCount          int64   `json:"todaySkillRunFailCount"`
	TodayAIHandoffCount             int64   `json:"todayAiHandoffCount"`
}

type DashboardAlertResponse struct {
	ID          string `json:"id"`
	Level       string `json:"level"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Count       int64  `json:"count"`
	Link        string `json:"link"`
}

type DashboardQuickLinkResponse struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Link        string `json:"link"`
}
