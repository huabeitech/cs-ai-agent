import { DashboardPlaceholder } from "@/components/dashboard-placeholder"

export default function DashboardSettingsPage() {
  return (
    <DashboardPlaceholder
      eyebrow="Settings"
      title="系统设置骨架"
      description="系统设置页将管理认证参数、上传配置、基础信息与运行策略。"
      nextSteps={[
        "补充系统配置读取与保存接口。",
        "按安全、存储、登录策略拆分设置区块。",
        "增加敏感配置的脱敏显示和二次确认。",
      ]}
    />
  )
}
