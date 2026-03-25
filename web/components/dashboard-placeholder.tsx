import { ArrowUpRightIcon, Clock3Icon } from "lucide-react"

import { Button } from "@/components/ui/button"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"

type DashboardPlaceholderProps = {
  eyebrow: string
  title: string
  description: string
  nextSteps: string[]
}

export function DashboardPlaceholder({
  eyebrow,
  title,
  description,
  nextSteps,
}: DashboardPlaceholderProps) {
  return (
    <div className="flex flex-1 flex-col gap-6 p-4 pt-4 lg:p-6 lg:pt-6">
      <Card className="border-dashed">
        <CardHeader className="gap-3">
          <span className="text-xs font-medium tracking-[0.24em] uppercase text-muted-foreground">
            {eyebrow}
          </span>
          <CardTitle className="text-3xl">{title}</CardTitle>
          <CardDescription className="max-w-2xl text-sm leading-6">
            {description}
          </CardDescription>
        </CardHeader>
        <CardContent className="grid gap-4 md:grid-cols-[1.2fr_0.8fr]">
          <div className="rounded-2xl border bg-muted/40 p-5">
            <p className="text-sm font-medium">建议下一步</p>
            <div className="mt-4 grid gap-3">
              {nextSteps.map((item) => (
                <div
                  key={item}
                  className="flex items-start gap-3 rounded-xl bg-background p-3"
                >
                  <Clock3Icon className="mt-0.5 size-4 text-muted-foreground" />
                  <p className="text-sm">{item}</p>
                </div>
              ))}
            </div>
          </div>
          <div className="flex flex-col justify-between rounded-2xl border bg-background p-5">
            <div>
              <p className="text-sm font-medium">状态</p>
              <p className="mt-2 text-sm leading-6 text-muted-foreground">
                当前模块已完成页面骨架，可直接接入真实 API、表单弹窗与列表查询。
              </p>
            </div>
            <Button variant="outline" className="mt-6 justify-between">
              进入下一阶段开发
              <ArrowUpRightIcon />
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
