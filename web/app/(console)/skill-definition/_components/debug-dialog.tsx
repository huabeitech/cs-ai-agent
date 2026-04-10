"use client"

import { zodResolver } from "@hookform/resolvers/zod"
import { useEffect, useMemo, useState } from "react"
import { Resolver, useForm } from "react-hook-form"
import { z } from "zod/v4"
import { LoaderCircleIcon, PlayIcon } from "lucide-react"
import { toast } from "sonner"

import { ProjectDialog } from "@/components/project-dialog"
import { OptionCombobox } from "@/components/option-combobox"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import {
  Field,
  FieldContent,
  FieldError,
  FieldLabel,
} from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import { Textarea } from "@/components/ui/textarea"
import {
  debugRunSkillDefinition,
  fetchAIAgentsAll,
  type AIAgent,
  type SkillDebugRunPayload,
  type SkillDebugRunResult,
} from "@/lib/api/admin"

type DebugDialogProps = {
  open: boolean
  skillCode: string
  skillName: string
  onOpenChange: (open: boolean) => void
}

const debugFormSchema = z.object({
  aiAgentId: z.string().trim().min(1, "请选择 AI Agent"),
  conversationId: z.string().trim(),
  userMessage: z.string().trim().min(1, "请输入用户消息"),
})

type DebugForm = z.infer<typeof debugFormSchema>

const debugFormResolver = zodResolver(debugFormSchema as never) as Resolver<
  z.input<typeof debugFormSchema>,
  undefined,
  z.output<typeof debugFormSchema>
>

const emptyForm: DebugForm = {
  aiAgentId: "",
  conversationId: "",
  userMessage: "",
}

function ResultBlock({
  title,
  value,
  emptyText = "暂无数据",
}: {
  title: string
  value?: string
  emptyText?: string
}) {
  return (
    <Card>
      <CardHeader className="pb-3">
        <CardTitle className="text-sm">{title}</CardTitle>
      </CardHeader>
      <CardContent>
        {value ? (
          <pre className="overflow-x-auto whitespace-pre-wrap break-words rounded-lg bg-muted/50 p-3 text-xs leading-5">
            {value}
          </pre>
        ) : (
          <div className="text-sm text-muted-foreground">{emptyText}</div>
        )}
      </CardContent>
    </Card>
  )
}

export function DebugDialog({
  open,
  skillCode,
  skillName,
  onOpenChange,
}: DebugDialogProps) {
  if (!open) {
    return null
  }

  return (
    <DebugDialogBody
      key={skillCode}
      open={open}
      skillCode={skillCode}
      skillName={skillName}
      onOpenChange={onOpenChange}
    />
  )
}

function DebugDialogBody({
  open,
  skillCode,
  skillName,
  onOpenChange,
}: DebugDialogProps) {
  const formId = `skill-debug-form-${skillCode}`
  const [running, setRunning] = useState(false)
  const [aiAgents, setAiAgents] = useState<AIAgent[]>([])
  const [result, setResult] = useState<SkillDebugRunResult | null>(null)
  const form = useForm<
    z.input<typeof debugFormSchema>,
    undefined,
    z.output<typeof debugFormSchema>
  >({
    resolver: debugFormResolver,
    defaultValues: emptyForm,
  })

  const {
    handleSubmit,
    reset,
    register,
    setValue,
    watch,
    formState: { errors },
  } = form

  const selectedAgentId = watch("aiAgentId")

  useEffect(() => {
    async function loadAIAgents() {
      try {
        const data = await fetchAIAgentsAll({ status: 1 })
        setAiAgents(data)
      } catch (error) {
        console.error("Failed to load AI agents:", error)
      }
    }

    void loadAIAgents()
  }, [])

  useEffect(() => {
    if (!open) {
      return
    }
    reset(emptyForm)
    setResult(null)
  }, [open, reset])

  useEffect(() => {
    if (!open || aiAgents.length === 0 || selectedAgentId) {
      return
    }
    setValue("aiAgentId", String(aiAgents[0].id), { shouldValidate: true })
  }, [aiAgents, open, selectedAgentId, setValue])

  const aiAgentOptions = useMemo(
    () =>
      aiAgents.map((item) => ({
        value: String(item.id),
        label: item.name,
      })),
    [aiAgents],
  )

  const selectedAgent = useMemo(
    () => aiAgents.find((item) => String(item.id) === selectedAgentId) ?? null,
    [aiAgents, selectedAgentId],
  )

  async function onSubmit(values: DebugForm) {
    const payload: SkillDebugRunPayload = {
      aiAgentId: Number(values.aiAgentId),
      skillCode,
      userMessage: values.userMessage.trim(),
    }
    const conversationId = Number(values.conversationId)
    if (conversationId > 0) {
      payload.conversationId = conversationId
    }

    setRunning(true)
    try {
      const data = await debugRunSkillDefinition(payload)
      setResult(data)
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "Skill 调试失败")
      setResult(null)
    } finally {
      setRunning(false)
    }
  }

  return (
    <ProjectDialog
      open={open}
      onOpenChange={onOpenChange}
      title={`调试 Skill · ${skillName || skillCode}`}
      description="强制指定当前 Skill，直接查看 route、tools、graph、HITL 和回复结果。"
      size="xl"
      allowFullscreen
      footer={
        <>
          <Button
            type="button"
            variant="outline"
            disabled={running}
            onClick={() => onOpenChange(false)}
          >
            关闭
          </Button>
          <Button type="submit" form={formId} disabled={running}>
            {running ? <LoaderCircleIcon className="animate-spin" /> : <PlayIcon />}
            {running ? "调试中..." : "开始调试"}
          </Button>
        </>
      }
    >
      <div className="space-y-6">
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm">调试输入</CardTitle>
          </CardHeader>
          <CardContent>
            <form id={formId} onSubmit={handleSubmit(onSubmit)} className="space-y-4">
              <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
                <Field data-invalid={!!errors.aiAgentId}>
                  <FieldLabel>AI Agent</FieldLabel>
                  <FieldContent>
                    <OptionCombobox
                      value={selectedAgentId}
                      options={aiAgentOptions}
                      placeholder="选择 AI Agent"
                      searchPlaceholder="搜索 AI Agent"
                      emptyText="未找到 AI Agent"
                      onChange={(value) =>
                        setValue("aiAgentId", value, { shouldValidate: true })
                      }
                    />
                    <FieldError errors={[errors.aiAgentId]} />
                  </FieldContent>
                </Field>
                <Field data-invalid={!!errors.conversationId}>
                  <FieldLabel htmlFor="skill-debug-conversation-id">
                    Conversation ID
                  </FieldLabel>
                  <FieldContent>
                    <Input
                      id="skill-debug-conversation-id"
                      type="number"
                      min={0}
                      placeholder="可选，填已有会话 ID 以复用上下文"
                      aria-invalid={!!errors.conversationId}
                      {...register("conversationId")}
                    />
                    <FieldError errors={[errors.conversationId]} />
                  </FieldContent>
                </Field>
              </div>
              <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
                <Field>
                  <FieldLabel>Skill</FieldLabel>
                  <FieldContent>
                    <Input value={skillCode} disabled />
                  </FieldContent>
                </Field>
                <Field>
                  <FieldLabel>命中 Agent</FieldLabel>
                  <FieldContent>
                    <Input
                      value={selectedAgent?.name || "未选择"}
                      disabled
                    />
                  </FieldContent>
                </Field>
              </div>
              <Field data-invalid={!!errors.userMessage}>
                <FieldLabel htmlFor="skill-debug-user-message">用户消息</FieldLabel>
                <FieldContent>
                  <Textarea
                    id="skill-debug-user-message"
                    rows={5}
                    placeholder="输入一段用户消息，调试当前 Skill 的路由、工具和回复。"
                    aria-invalid={!!errors.userMessage}
                    {...register("userMessage")}
                  />
                  <FieldError errors={[errors.userMessage]} />
                </FieldContent>
              </Field>
            </form>
          </CardContent>
        </Card>

        <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
          <Card>
            <CardHeader className="pb-3">
              <CardTitle className="text-sm">调试摘要</CardTitle>
            </CardHeader>
            <CardContent className="space-y-3 text-sm">
              <div className="flex flex-wrap gap-2">
                <Badge variant="outline">{result?.skillCode || skillCode}</Badge>
                {result?.graphToolCode ? (
                  <Badge variant="secondary">{result.graphToolCode}</Badge>
                ) : null}
                {result?.interruptType ? (
                  <Badge variant="secondary">{result.interruptType}</Badge>
                ) : null}
                {result?.interrupted ? (
                  <Badge>已中断</Badge>
                ) : (
                  <Badge variant="outline">未中断</Badge>
                )}
              </div>
              <div className="rounded-lg bg-muted/50 p-3">
                <div className="text-xs text-muted-foreground">Skill 名称</div>
                <div className="mt-1 font-medium">{result?.skillName || skillName}</div>
              </div>
              <div className="rounded-lg bg-muted/50 p-3">
                <div className="text-xs text-muted-foreground">Plan Reason</div>
                <div className="mt-1 whitespace-pre-wrap break-words">
                  {result?.planReason || "暂无"}
                </div>
              </div>
              <div className="rounded-lg bg-muted/50 p-3">
                <div className="text-xs text-muted-foreground">Reply</div>
                <div className="mt-1 whitespace-pre-wrap break-words">
                  {result?.replyText || "暂无"}
                </div>
              </div>
              <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
                <div className="rounded-lg bg-muted/50 p-3">
                  <div className="text-xs text-muted-foreground">Checkpoint</div>
                  <div className="mt-1 break-all">
                    {result?.checkPointId || "暂无"}
                  </div>
                </div>
                <div className="rounded-lg bg-muted/50 p-3">
                  <div className="text-xs text-muted-foreground">错误信息</div>
                  <div className="mt-1 whitespace-pre-wrap break-words">
                    {result?.errorMessage || "暂无"}
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="pb-3">
              <CardTitle className="text-sm">工具视图</CardTitle>
            </CardHeader>
            <CardContent className="space-y-3 text-sm">
              <div className="rounded-lg bg-muted/50 p-3">
                <div className="text-xs text-muted-foreground">Skill Allowed Tools</div>
                <div className="mt-2 flex flex-wrap gap-2">
                  {(result?.skillAllowedToolCodes ?? []).length > 0 ? (
                    result?.skillAllowedToolCodes.map((toolCode) => (
                      <Badge key={toolCode} variant="outline">
                        {toolCode}
                      </Badge>
                    ))
                  ) : (
                    <span className="text-muted-foreground">暂无</span>
                  )}
                </div>
              </div>
              <div className="rounded-lg bg-muted/50 p-3">
                <div className="text-xs text-muted-foreground">Candidate Tools</div>
                <div className="mt-2 flex flex-wrap gap-2">
                  {(result?.toolCodes ?? []).length > 0 ? (
                    result?.toolCodes.map((toolCode) => (
                      <Badge key={toolCode} variant="outline">
                        {toolCode}
                      </Badge>
                    ))
                  ) : (
                    <span className="text-muted-foreground">暂无</span>
                  )}
                </div>
              </div>
              <div className="rounded-lg bg-muted/50 p-3">
                <div className="text-xs text-muted-foreground">Invoked Tools</div>
                <div className="mt-2 flex flex-wrap gap-2">
                  {(result?.invokedToolCodes ?? []).length > 0 ? (
                    result?.invokedToolCodes.map((toolCode) => (
                      <Badge key={toolCode} variant="secondary">
                        {toolCode}
                      </Badge>
                    ))
                  ) : (
                    <span className="text-muted-foreground">暂无</span>
                  )}
                </div>
              </div>
            </CardContent>
          </Card>
        </div>

        <div className="grid grid-cols-1 gap-4">
          <ResultBlock title="Skill Route Trace" value={result?.skillRouteTrace} />
          <ResultBlock title="Tool Search Trace" value={result?.toolSearchTrace} />
          <ResultBlock title="Graph Tool Trace" value={result?.graphToolTrace} />
          <ResultBlock title="Trace Data" value={result?.traceData} />
        </div>
      </div>
    </ProjectDialog>
  )
}
