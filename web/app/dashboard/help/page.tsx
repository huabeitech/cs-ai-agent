import { DashboardPlaceholder } from "@/components/dashboard-placeholder"

export default function DashboardHelpPage() {
  return (
    <DashboardPlaceholder
      eyebrow="Help"
      title="帮助中心骨架"
      description="帮助中心用于沉淀产品说明、权限约定、渠道接入流程与常见问题。"
      nextSteps={[
        "汇总项目开发规范与后台使用说明。",
        "补充第三方接入流程图和配置校验清单。",
        "后续可接入 markdown 或文档中心能力。",
      ]}
    />
  )
}
