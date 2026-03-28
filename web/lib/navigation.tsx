import {
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
  UsersIcon
} from "lucide-react";

export const dashboardNavSections = [
  {
    title: "工作台",
    items: [
      {
        title: "总览",
        url: "/dashboard",
        icon: <LayoutDashboardIcon />,
      },
      {
        title: "会话管理",
        url: "/dashboard/conversations",
        icon: <BotMessageSquareIcon />,
      },
      {
        title: "快捷回复",
        url: "/dashboard/quick-replies",
        icon: <MessageSquareMoreIcon />,
      },
      {
        title: "会话标签",
        url: "/dashboard/tags",
        icon: <TagsIcon />,
      },
      {
        title: "公司管理",
        url: "/dashboard/companies",
        icon: <Building2Icon />,
      },
      {
        title: "客户管理",
        url: "/dashboard/customers",
        icon: <UsersIcon />,
      },
    ],
  },
  {
    title: "客服管理",
    items: [
      {
        title: "客服档案",
        url: "/dashboard/agents",
        icon: <UserCogIcon />,
      },
      {
        title: "客服组排班",
        url: "/dashboard/agent-team-schedules",
        icon: <CalendarClockIcon />,
      },
      {
        title: "接入站点",
        url: "/dashboard/widget-sites",
        icon: <GlobeIcon />,
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
      },
      {
        title: "AI配置",
        url: "/dashboard/ai-configs",
        icon: <BrainCircuitIcon />,
      },
      {
        title: "AI Agent",
        url: "/dashboard/ai-agents",
        icon: <MessageSquareMoreIcon />,
      },
      {
        title: "Skills",
        url: "/dashboard/skill-definition",
        icon: <MessageSquareCodeIcon />,
      },
      {
        title: "MCP调试",
        url: "/dashboard/mcp",
        icon: <MessageSquareCodeIcon />,
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
      },
      {
        title: "角色管理",
        url: "/dashboard/roles",
        icon: <ShieldCheckIcon />,
      },
      {
        title: "权限管理",
        url: "/dashboard/permissions",
        icon: <KeyRoundIcon />,
      },
    ],
  },
] as const;

export const dashboardSecondaryNav = [
  {
    title: "系统设置",
    url: "/dashboard/settings",
    icon: <Settings2Icon />,
  },
  {
    title: "帮助中心",
    url: "/dashboard/help",
    icon: <LifeBuoyIcon />,
  },
] as const;

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
