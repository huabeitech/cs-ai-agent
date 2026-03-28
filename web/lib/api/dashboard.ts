import { request } from "@/lib/api/client"

export type DashboardRange = "today" | "7d" | "30d"

export type DashboardStatusDistributionItem = {
  status: number
  label: string
  count: number
}

export type DashboardTrendItem = {
  date: string
  newCount: number
  closedCount: number
}

export type DashboardTeamLoad = {
  teamId: number
  teamName: string
  totalAgents: number
  onlineAgents: number
  busyAgents: number
  offlineAgents: number
  waitingConversations: number
  processingConversations: number
  maxConcurrentCapacity: number
  loadRate: number
  hasScheduleNow: boolean
}

export type DashboardAlert = {
  id: string
  level: "info" | "warning" | "error"
  title: string
  description: string
  count: number
  link: string
}

export type DashboardQuickLink = {
  title: string
  description: string
  link: string
}

export type DashboardOverview = {
  range: DashboardRange
  generatedAt: string
  summary: {
    todayNewConversations: number
    processingConversations: number
    pendingDispatchConversations: number
    onlineAgents: number
    aiServiceRate: number
  }
  conversationStats: {
    statusDistribution: DashboardStatusDistributionItem[]
    trend: DashboardTrendItem[]
  }
  agentStats: {
    onlineAgents: number
    busyAgents: number
    offlineAgents: number
    teamLoads: DashboardTeamLoad[]
  }
  aiStats: {
    enabledAiAgents: number
    enabledWidgetSites: number
    todayKnowledgeRetrieves: number
    todayKnowledgeRetrieveFailCount: number
    todayKnowledgeRetrieveFailRate: number
    todaySkillRunFailCount: number
    todayAiHandoffCount: number
  }
  alerts: DashboardAlert[]
  quickLinks: DashboardQuickLink[]
}

export function fetchDashboardOverview(range: DashboardRange) {
  return request<DashboardOverview>(`/api/console/dashboard/overview?range=${range}`)
}
