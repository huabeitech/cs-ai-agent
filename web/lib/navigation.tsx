import {
  ActivitySquareIcon,
  BotMessageSquareIcon,
  BrainCircuitIcon,
  Building2Icon,
  CalendarClockIcon,
  ChartColumnIncreasingIcon,
  FileTextIcon,
  GlobeIcon,
  KeyRoundIcon,
  LayoutDashboardIcon,
  MessageSquareCodeIcon,
  MessageSquareMoreIcon,
  Settings2Icon,
  ShieldCheckIcon,
  TagsIcon,
  UserCogIcon,
  UsersIcon,
} from "lucide-react";
import type { ReactNode } from "react";

/** 与后端 internal/pkg/constants/auth.go RoleCodeSuperAdmin 一致 */
export const DASHBOARD_ROLE_SUPER_ADMIN = "super_admin";

export type DashboardNavMenuItem = {
  title: string;
  url: string;
  icon: ReactNode;
};

export type DashboardNavItemConfig = DashboardNavMenuItem & {
  /**
   * 与后端 Permission.Code 一致；缺省表示任意已登录管理员可见
   * （对应控制台接口尚未 RequirePermission 的模块）
   */
  requiredPermission?: string;
};

export type DashboardNavSectionConfig = {
  title: string;
  items: DashboardNavItemConfig[];
};

function navItemVisible(
  item: DashboardNavItemConfig,
  superAdmin: boolean,
  permissionSet: Set<string>,
): boolean {
  if (superAdmin) {
    return true;
  }
  if (!item.requiredPermission) {
    return true;
  }
  return permissionSet.has(item.requiredPermission);
}

export function filterDashboardNavForSession(
  permissions: readonly string[] | undefined,
  roles: readonly string[] | undefined,
): { title: string; items: DashboardNavMenuItem[] }[] {
  const superAdmin = roles?.includes(DASHBOARD_ROLE_SUPER_ADMIN) ?? false;
  const permissionSet = new Set(permissions ?? []);
  return dashboardNavSections
    .map((section) => ({
      title: section.title,
      items: section.items
        .filter((item) => navItemVisible(item, superAdmin, permissionSet))
        .map(({ title, url, icon }) => ({ title, url, icon })),
    }))
    .filter((section) => section.items.length > 0);
}

export function filterDashboardSecondaryNavForSession(
  permissions: readonly string[] | undefined,
  roles: readonly string[] | undefined,
): DashboardNavMenuItem[] {
  const superAdmin = roles?.includes(DASHBOARD_ROLE_SUPER_ADMIN) ?? false;
  const permissionSet = new Set(permissions ?? []);
  return dashboardSecondaryNav
    .filter((item) => navItemVisible(item, superAdmin, permissionSet))
    .map(({ title, url, icon }) => ({ title, url, icon }));
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
        url: "/dashboard",
        icon: <LayoutDashboardIcon />,
      },
      {
        title: "会话",
        url: "/dashboard/conversations",
        icon: <BotMessageSquareIcon />,
        requiredPermission: "conversation.view",
      },
      {
        title: "工单",
        url: "/dashboard/tickets",
        icon: <FileTextIcon />,
        requiredPermission: "ticket.view",
      },
      {
        title: "会话监控",
        url: "/dashboard/conversation-monitor",
        icon: <BotMessageSquareIcon />,
        requiredPermission: "conversation.view",
      },
      {
        title: "SLA风险",
        url: "/dashboard/ticket-risk",
        icon: <ChartColumnIncreasingIcon />,
        requiredPermission: "ticket.view",
      },
      {
        title: "客户管理",
        url: "/dashboard/customers",
        icon: <UsersIcon />,
        requiredPermission: "customer.view",
      },
      {
        title: "公司管理",
        url: "/dashboard/companies",
        icon: <Building2Icon />,
        requiredPermission: "company.view",
      },
    ],
  },
  {
    title: "客服配置",
    items: [
      {
        title: "分类标签",
        url: "/dashboard/tags",
        icon: <TagsIcon />,
        requiredPermission: "tag.view",
      },
      {
        title: "快捷回复",
        url: "/dashboard/quick-replies",
        icon: <MessageSquareMoreIcon />,
        requiredPermission: "quickReply.view",
      },
      {
        title: "工单优先级",
        url: "/dashboard/ticket-priorities",
        icon: <Settings2Icon />,
        requiredPermission: "ticketPriorityConfig.view",
      },
      {
        title: "工单解决码",
        url: "/dashboard/ticket-resolution-codes",
        icon: <KeyRoundIcon />,
        requiredPermission: "ticketResolutionCode.view",
      },
      {
        title: "客服档案",
        url: "/dashboard/agents",
        icon: <UserCogIcon />,
        requiredPermission: "agent.view",
      },
      {
        title: "客服组排班",
        url: "/dashboard/agent-team-schedules",
        icon: <CalendarClockIcon />,
        requiredPermission: "agentTeamSchedule.view",
      },
      {
        title: "接入渠道",
        url: "/dashboard/channels",
        icon: <GlobeIcon />,
        requiredPermission: "channel.view",
      },
    ],
  },
  {
    title: "知识与AI",
    items: [
      {
        title: "知识库",
        url: "/dashboard/knowledge",
        icon: <FileTextIcon />,
        requiredPermission: "knowledgeBase.view",
      },
      {
        title: "AI配置",
        url: "/dashboard/ai-configs",
        icon: <BrainCircuitIcon />,
        requiredPermission: "aiConfig.view",
      },
      {
        title: "AI Agent",
        url: "/dashboard/ai-agents",
        icon: <MessageSquareMoreIcon />,
        requiredPermission: "aiAgent.view",
      },
      {
        title: "Skills",
        url: "/dashboard/skill-definition",
        icon: <MessageSquareCodeIcon />,
        requiredPermission: "skillDefinition.view",
      },
      {
        title: "MCP调试",
        url: "/dashboard/mcp",
        icon: <MessageSquareCodeIcon />,
        requiredPermission: "mcp.view",
      },
      {
        title: "Agent日志",
        url: "/dashboard/agent-run-logs",
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
        url: "/dashboard/users",
        icon: <UsersIcon />,
        requiredPermission: "user.view",
      },
      {
        title: "角色管理",
        url: "/dashboard/roles",
        icon: <ShieldCheckIcon />,
        requiredPermission: "role.view",
      },
      {
        title: "权限管理",
        url: "/dashboard/permissions",
        icon: <KeyRoundIcon />,
        requiredPermission: "permission.view",
      },
    ],
  },
];

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
];

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
] as const;

export function getPageTitle(pathname: string): string {
  let matchedTitle = "后台总览";
  let longestMatch = 0;

  for (const section of dashboardNavSections) {
    for (const item of section.items) {
      if (pathname === item.url || pathname.startsWith(item.url + "/")) {
        const matchLength = item.url.length;
        if (matchLength > longestMatch) {
          longestMatch = matchLength;
          matchedTitle = item.title;
        }
      }
    }
  }

  for (const item of dashboardSecondaryNav) {
    if (pathname === item.url || pathname.startsWith(item.url + "/")) {
      const matchLength = item.url.length;
      if (matchLength > longestMatch) {
        longestMatch = matchLength;
        matchedTitle = item.title;
      }
    }
  }

  return matchedTitle;
}
