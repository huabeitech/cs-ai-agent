"use client"

import {
  ArrowUpRightIcon,
  CircleCheckIcon,
  Clock3Icon,
  FilterIcon,
} from "lucide-react"

import { formatDateTime } from "@/lib/utils"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import {
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from "@/components/ui/tabs"

type DashboardTask = {
  id: number
  module: string
  owner: string
  status: string
  progress: string
  updatedAt: string
}

export function DataTable({ data }: { data: DashboardTask[] }) {
  return (
    <Tabs
      defaultValue="modules"
      className="w-full flex-col justify-start gap-6 px-4 lg:px-6"
    >
      <div className="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
        <div>
          <CardTitle className="text-xl">模块推进看板</CardTitle>
          <CardDescription className="mt-1">
            这里是后台一期的功能骨架清单，后续可替换为真实接口列表。
          </CardDescription>
        </div>
        <div className="flex items-center gap-2">
          <Input className="w-full md:w-64" placeholder="搜索模块名称" />
          <Button variant="outline">
            <FilterIcon />
            筛选
          </Button>
        </div>
      </div>
      <TabsList className="w-fit">
        <TabsTrigger value="modules">模块列表</TabsTrigger>
        <TabsTrigger value="milestones">里程碑</TabsTrigger>
      </TabsList>
      <TabsContent value="modules" className="m-0">
        <Card>
          <CardHeader>
            <CardTitle>一期模块拆解</CardTitle>
            <CardDescription>
              覆盖账号权限、知识库、渠道接入与 Skill 能力入口。
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="overflow-hidden rounded-xl border">
              <Table>
                <TableHeader className="bg-muted/50">
                  <TableRow>
                    <TableHead>模块</TableHead>
                    <TableHead>负责人</TableHead>
                    <TableHead>状态</TableHead>
                    <TableHead>完成度</TableHead>
                    <TableHead className="text-right">更新时间</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {data.map((item) => (
                    <TableRow key={item.id}>
                      <TableCell className="font-medium">{item.module}</TableCell>
                      <TableCell>{item.owner}</TableCell>
                      <TableCell>
                        <Badge variant="outline" className="px-1.5">
                          {item.status === "已完成" ? (
                            <CircleCheckIcon className="fill-green-500 text-green-500" />
                          ) : (
                            <Clock3Icon className="text-amber-500" />
                          )}
                          {item.status}
                        </Badge>
                      </TableCell>
                      <TableCell>{item.progress}</TableCell>
                      <TableCell className="text-right text-muted-foreground">
                        {formatDateTime(item.updatedAt)}
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          </CardContent>
        </Card>
      </TabsContent>
      <TabsContent value="milestones" className="m-0">
        <Card className="border-dashed">
          <CardHeader>
            <CardTitle>下一阶段里程碑</CardTitle>
            <CardDescription>
              当前已完成 UI 基础骨架，下一步接入真实 API 和业务表单。
            </CardDescription>
          </CardHeader>
          <CardContent className="grid gap-3">
            {[
              "打通登录、用户、角色、权限列表接口。",
              "补充表单弹窗、分页查询与错误处理。",
              "接入知识库与渠道模块的实际配置能力。",
            ].map((item) => (
              <div
                key={item}
                className="flex items-center justify-between rounded-xl border px-4 py-3"
              >
                <span className="text-sm">{item}</span>
                <ArrowUpRightIcon className="size-4 text-muted-foreground" />
              </div>
            ))}
          </CardContent>
        </Card>
      </TabsContent>
    </Tabs>
  )
}
