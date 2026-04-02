import type { ReactNode } from "react"
import {
  ActivitySquareIcon,
  BotMessageSquareIcon,
  BrainCircuitIcon,
  Building2Icon,
  CalendarClockIcon,
  FileTextIcon,
  GlobeIcon,
  KeyRoundIcon,
  LayoutDashboardIcon,
  LifeBuoyIcon,
  MessageSquareCodeIcon,
  MessageSquareMoreIcon,
  Settings2Icon,
  ShieldCheckIcon,
  TagsIcon,
  UserCogIcon,
  UsersIcon,
} from "lucide-react"

/** 与后端 internal/pkg/constants/auth.go RoleCodeSuperAdmin 一致 */
export const DASHBOARD_ROLE_SUPER_ADMIN = "super_admin"

export type DashboardNavMenuItem = {
  title: string
  url: string
  icon: ReactNode
}

export type DashboardNavItemConfig = DashboardNavMenuItem & {
  /**
   * 与后端 Permission.Code 一致；缺省表示任意已登录管理员可见
   * （对应控制台接口尚未 RequirePermission 的模块，如接入站点、AI Agent）
   */
  requiredPermission?: string
}

export type DashboardNavSectionConfig = {
  title: string
  items: DashboardNavItemConfig[]
}

function navItemVisible(
  item: DashboardNavItemConfig,
  superAdmin: boolean,
  permissionSet: Set<string>
): boolean {
  if (superAdmin) {
    return true
  }
  if (!item.requiredPermission) {
    return true
  }
  return permissionSet.has(item.requiredPermission)
}

export function filterDashboardNavForSession(
  permissions: readonly string[] | undefined,
  roles: readonly string[] | undefined
): { title: string; items: DashboardNavMenuItem[] }[] {
  const superAdmin = roles?.includes(DASHBOARD_ROLE_SUPER_ADMIN) ?? false
  const permissionSet = new Set(permissions ?? [])
  return dashboardNavSections
    .map((section) => ({
      title: section.title,
      items: section.items
        .filter((item) => navItemVisible(item, superAdmin, permissionSet))
        .map(({ title, url, icon }) => ({ title, url, icon })),
    }))
    .filter((section) => section.items.length > 0)
}

export function filterDashboardSecondaryNavForSession(
  permissions: readonly string[] | undefined,
  roles: readonly string[] | undefined
): DashboardNavMenuItem[] {
  const superAdmin = roles?.includes(DASHBOARD_ROLE_SUPER_ADMIN) ?? false
  const permissionSet = new Set(permissions ?? [])
  return dashboardSecondaryNav
    .filter((item) => navItemVisible(item, superAdmin, permissionSet))
    .map(({ title, url, icon }) => ({ title, url, icon }))
}

export const dashboardNavSections: DashboardNavSectionConfig[] = [
  // {
  //   title: "总览",
  //   items: [
  //     {
  //       title: "总览",
  //       url: "/",
  //       icon: <LayoutDashboardIcon />,
  //     },
  //   ],
  // },
  {
    title: "接待中心",
    items: [
      {
        title: "总览",
        url: "/",
        icon: <LayoutDashboardIcon />,
      },
      {
        title: "会话",
        url: "/conversations",
        icon: <BotMessageSquareIcon />,
        requiredPermission: "conversation.view",
      },
      {
        title: "工单",
        url: "/tickets",
        icon: <FileTextIcon />,
        requiredPermission: "ticket.view",
      },
      {
        title: "会话监控",
        url: "/conversation-monitor",
        icon: <BotMessageSquareIcon />,
        requiredPermission: "conversation.view",
      },
      {
        title: "快捷回复",
        url: "/quick-replies",
        icon: <MessageSquareMoreIcon />,
        requiredPermission: "quickReply.view",
      },
      {
        title: "会话标签",
        url: "/tags",
        icon: <TagsIcon />,
        requiredPermission: "tag.view",
      },
      {
        title: "客户管理",
        url: "/customers",
        icon: <UsersIcon />,
        requiredPermission: "customer.view",
      },
    ],
  },
  {
    title: "客服管理",
    items: [
      {
        title: "公司管理",
        url: "/companies",
        icon: <Building2Icon />,
        requiredPermission: "company.view",
      },
      {
        title: "客服档案",
        url: "/agents",
        icon: <UserCogIcon />,
        requiredPermission: "agent.view",
      },
      {
        title: "客服组排班",
        url: "/agent-team-schedules",
        icon: <CalendarClockIcon />,
        requiredPermission: "agentTeamSchedule.view",
      },
      {
        title: "接入站点",
        url: "/widget-sites",
        icon: <GlobeIcon />,
      },
    ],
  },
  {
    title: "知识与AI",
    items: [
      {
        title: "知识库",
        url: "/knowledge",
        icon: <FileTextIcon />,
        requiredPermission: "knowledgeBase.view",
      },
      {
        title: "AI配置",
        url: "/ai-configs",
        icon: <BrainCircuitIcon />,
        requiredPermission: "aiConfig.view",
      },
      {
        title: "AI Agent",
        url: "/ai-agents",
        icon: <MessageSquareMoreIcon />,
      },
      {
        title: "Skills",
        url: "/skill-definition",
        icon: <MessageSquareCodeIcon />,
        requiredPermission: "skillDefinition.view",
      },
      {
        title: "MCP调试",
        url: "/mcp",
        icon: <MessageSquareCodeIcon />,
        requiredPermission: "mcp.view",
      },
      {
        title: "Agent日志",
        url: "/agent-run-logs",
        icon: <ActivitySquareIcon />,
        requiredPermission: "conversation.view",
      },
    ],
  },
  {
    title: "系统管理",
    items: [
      {
        title: "用户管理",
        url: "/users",
        icon: <UsersIcon />,
        requiredPermission: "user.view",
      },
      {
        title: "角色管理",
        url: "/roles",
        icon: <ShieldCheckIcon />,
        requiredPermission: "role.view",
      },
      {
        title: "权限管理",
        url: "/permissions",
        icon: <KeyRoundIcon />,
        requiredPermission: "permission.view",
      },
      {
        title: "工单分类",
        url: "/ticket-categories",
        icon: <TagsIcon />,
        requiredPermission: "ticketCategory.view",
      },
      {
        title: "解决码",
        url: "/ticket-resolution-codes",
        icon: <KeyRoundIcon />,
        requiredPermission: "ticketResolutionCode.view",
      },
      {
        title: "工单SLA",
        url: "/ticket-sla-configs",
        icon: <Settings2Icon />,
        requiredPermission: "ticketSLAConfig.view",
      },
    ],
  },
]

export const dashboardSecondaryNav: DashboardNavItemConfig[] = [
  // {
  //   title: "系统设置",
  //   url: "/settings",
  //   icon: <Settings2Icon />,
  // },
  // {
  //   title: "帮助中心",
  //   url: "/help",
  //   icon: <LifeBuoyIcon />,
  // },
]

export const dashboardQuickActions = [
  {
    title: "查看会话",
    icon: <BotMessageSquareIcon />,
  },
  {
    title: "邀请成员",
    icon: <UserCogIcon />,
  },
  {
    title: "接入机器人",
    icon: <MessageSquareCodeIcon />,
  },
] as const

export function getPageTitle(pathname: string): string {
  let matchedTitle = "后台总览"
  let longestMatch = 0

  for (const section of dashboardNavSections) {
    for (const item of section.items) {
      if (pathname === item.url || pathname.startsWith(item.url + "/")) {
        const matchLength = item.url.length
        if (matchLength > longestMatch) {
          longestMatch = matchLength
          matchedTitle = item.title
        }
      }
    }
  }

  for (const item of dashboardSecondaryNav) {
    if (pathname === item.url || pathname.startsWith(item.url + "/")) {
      const matchLength = item.url.length
      if (matchLength > longestMatch) {
        longestMatch = matchLength
        matchedTitle = item.title
      }
    }
  }

  return matchedTitle
}
