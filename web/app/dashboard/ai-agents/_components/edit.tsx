"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import {
  ArrowDownIcon,
  ArrowUpIcon,
  InfoIcon,
  PlusIcon,
  Trash2Icon,
} from "lucide-react";
import { useEffect, useMemo, useState } from "react";
import { Controller, Resolver, useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod/v4";

import { OptionCombobox } from "@/components/option-combobox";
import { ProjectDialog } from "@/components/project-dialog";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Field,
  FieldContent,
  FieldError,
  FieldLabel,
} from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import {
  Popover,
  PopoverContent,
  PopoverDescription,
  PopoverTrigger,
} from "@/components/ui/popover";
import { Textarea } from "@/components/ui/textarea";
import {
  type AIAgent,
  type AIConfig,
  type AdminAgentTeam,
  type CreateAIAgentPayload,
  type KnowledgeBase,
  type SkillDefinition,
  fetchAIAgent,
  fetchAIConfigsAll,
  fetchAgentTeamsAll,
  fetchKnowledgeBasesAll,
  fetchSkillDefinitionsAll,
  listMCPServers,
  listMCPTools,
} from "@/lib/api/admin";
import {
  AIAgentFallbackMode,
  AIAgentFallbackModeLabels,
  AIAgentHandoffMode,
  AIAgentHandoffModeLabels,
  AIModelType,
  IMConversationServiceMode,
  IMConversationServiceModeLabels,
  Status,
} from "@/lib/generated/enums";
import { getEnumOptions } from "@/lib/enums";

type EditDialogProps = {
  open: boolean;
  saving: boolean;
  itemId: number | null;
  onOpenChange: (open: boolean) => void;
  onSubmit: (payload: CreateAIAgentPayload) => Promise<void>;
};

const schema = z.object({
  name: z.string().trim().min(1, "名称不能为空"),
  description: z.string().trim(),
  aiConfigId: z.string().trim().regex(/^\d+$/, "请选择 AI 配置"),
  serviceMode: z.string().trim().min(1, "请选择服务模式"),
  systemPrompt: z.string().trim(),
  welcomeMessage: z.string().trim(),
  replyTimeoutSeconds: z
    .number()
    .min(0, "回复超时秒数必须是大于等于 0 的整数"),
  handoffMode: z.string().trim().min(1, "请选择转人工模式"),
  maxAiReplyRounds: z
    .number()
    .min(0, "AI 最大回复次数必须是大于等于 0 的整数"),
  fallbackMode: z.string().trim().min(1, "请选择兜底模式"),
  fallbackMessage: z.string().trim(),
  remark: z.string().trim(),
});

type EditForm = z.infer<typeof schema>;

const resolver = zodResolver(schema as never) as Resolver<
  z.input<typeof schema>,
  undefined,
  z.output<typeof schema>
>;

const serviceModeOptions = getEnumOptions(IMConversationServiceModeLabels).map(
  (option) => ({
    value: String(option.value),
    label: option.label,
  }),
);

const handoffModeOptions = getEnumOptions(AIAgentHandoffModeLabels).map(
  (option) => ({
    value: String(option.value),
    label: option.label,
  }),
);

const fallbackModeOptions = getEnumOptions(AIAgentFallbackModeLabels).map(
  (option) => ({
    value: String(option.value),
    label: option.label,
  }),
);

function buildForm(item: AIAgent | null): EditForm {
  if (!item) {
    return {
      name: "",
      description: "",
      aiConfigId: "",
      serviceMode: String(IMConversationServiceMode.AIFirst),
      systemPrompt: "",
      welcomeMessage: "",
      replyTimeoutSeconds: 180,
      handoffMode: String(AIAgentHandoffMode.WaitPool),
      maxAiReplyRounds: 2,
      fallbackMode: String(AIAgentFallbackMode.GuideRephrase),
      fallbackMessage:
        "我暂时没有找到足够准确的信息。你可以补充订单号、产品名或更具体的问题，我再继续帮你查。",
      remark: "",
    };
  }
  return {
    name: item.name,
    description: item.description || "",
    aiConfigId: item.aiConfigId > 0 ? String(item.aiConfigId) : "",
    serviceMode: String(item.serviceMode),
    systemPrompt: item.systemPrompt || "",
    welcomeMessage: item.welcomeMessage || "",
    replyTimeoutSeconds: item.replyTimeoutSeconds ?? 180,
    handoffMode: String(item.handoffMode),
    maxAiReplyRounds: item.maxAiReplyRounds ?? 2,
    fallbackMode: String(item.fallbackMode),
    fallbackMessage: item.fallbackMessage || "",
    remark: item.remark || "",
  };
}

function buildPayload(
  form: EditForm,
  knowledgeIds: number[],
  teamIds: number[],
  skillIds: number[],
  directTools: CreateAIAgentPayload["directTools"],
): CreateAIAgentPayload {
  return {
    name: form.name.trim(),
    description: form.description.trim(),
    aiConfigId: Number(form.aiConfigId),
    serviceMode: Number(form.serviceMode),
    systemPrompt: form.systemPrompt.trim(),
    welcomeMessage: form.welcomeMessage.trim(),
    replyTimeoutSeconds: Number(form.replyTimeoutSeconds),
    teamIds,
    handoffMode: Number(form.handoffMode),
    maxAiReplyRounds: Number(form.maxAiReplyRounds),
    fallbackMode: Number(form.fallbackMode),
    fallbackMessage: form.fallbackMessage.trim(),
    knowledgeIds,
    skillIds,
    directTools,
    remark: form.remark.trim(),
  };
}

export function EditDialog({
  open,
  saving,
  itemId,
  onOpenChange,
  onSubmit,
}: EditDialogProps) {
  if (!open) {
    return null;
  }
  return (
    <EditDialogBody
      key={itemId ? `edit-${itemId}` : "create"}
      open={open}
      saving={saving}
      itemId={itemId}
      onOpenChange={onOpenChange}
      onSubmit={onSubmit}
    />
  );
}

function EditDialogBody({
  open,
  saving,
  itemId,
  onOpenChange,
  onSubmit,
}: EditDialogProps) {
  const formId = "ai-agent-edit-form";
  const [loading, setLoading] = useState(false);
  const form = useForm<
    z.input<typeof schema>,
    undefined,
    z.output<typeof schema>
  >({
    resolver,
    defaultValues: buildForm(null),
  });
  const {
    control,
    handleSubmit,
    register,
    reset,
    watch,
    formState: { errors },
  } = form;
  const [selectedKnowledgeIds, setSelectedKnowledgeIds] = useState<number[]>(
    [],
  );
  const [selectedTeamIds, setSelectedTeamIds] = useState<number[]>([]);
  const [selectedSkillIds, setSelectedSkillIds] = useState<number[]>([]);
  const [knowledgeToAdd, setKnowledgeToAdd] = useState("");
  const [teamToAdd, setTeamToAdd] = useState("");
  const [skillToAdd, setSkillToAdd] = useState("");
  const [directToolServerCodeToAdd, setDirectToolServerCodeToAdd] = useState("");
  const [directToolToAdd, setDirectToolToAdd] = useState("");
  const [aiConfigs, setAIConfigs] = useState<AIConfig[]>([]);
  const [knowledgeBases, setKnowledgeBases] = useState<KnowledgeBase[]>([]);
  const [agentTeams, setAgentTeams] = useState<AdminAgentTeam[]>([]);
  const [skills, setSkills] = useState<SkillDefinition[]>([]);
  const [directTools, setDirectTools] = useState<
    CreateAIAgentPayload["directTools"]
  >([]);
  const [directToolOptions, setDirectToolOptions] = useState<
    { value: string; label: string; meta: CreateAIAgentPayload["directTools"][number] }[]
  >([]);

  useEffect(() => {
    async function loadDetail() {
      if (!itemId) {
        reset(buildForm(null));
        setSelectedKnowledgeIds([]);
        setSelectedTeamIds([]);
        setSelectedSkillIds([]);
        setDirectTools([]);
        setKnowledgeToAdd("");
        setTeamToAdd("");
        setSkillToAdd("");
        setDirectToolServerCodeToAdd("");
        setDirectToolToAdd("");
        return;
      }
      setLoading(true);
      try {
        const data = await fetchAIAgent(itemId);
        reset(buildForm(data));
        setSelectedKnowledgeIds(data.knowledgeIds ?? []);
        setSelectedTeamIds((data.teams ?? []).map((team) => team.id));
        setSelectedSkillIds(data.skillIds ?? []);
        setDirectTools(data.directTools ?? []);
        setKnowledgeToAdd("");
        setTeamToAdd("");
        setSkillToAdd("");
        setDirectToolServerCodeToAdd("");
        setDirectToolToAdd("");
      } catch (error) {
        toast.error(
          error instanceof Error ? error.message : "加载 AI Agent 详情失败",
        );
      } finally {
        setLoading(false);
      }
    }
    void loadDetail();
  }, [itemId, reset]);

  useEffect(() => {
    async function loadAIConfigs() {
      try {
        const data = await fetchAIConfigsAll({
          modelType: AIModelType.LLM,
        });
        setAIConfigs(data);
      } catch (error) {
        toast.error(
          error instanceof Error ? error.message : "加载 AI 配置失败",
        );
      }
    }
    void loadAIConfigs();
  }, []);

  useEffect(() => {
    async function loadAgentTeams() {
      try {
        const data = await fetchAgentTeamsAll();
        setAgentTeams(data);
      } catch (error) {
        toast.error(error instanceof Error ? error.message : "加载客服组失败");
      }
    }
    void loadAgentTeams();
  }, []);

  useEffect(() => {
    async function loadKnowledgeBases() {
      try {
        const data = await fetchKnowledgeBasesAll({
          status: Status.Ok,
        });
        setKnowledgeBases(data);
      } catch (error) {
        toast.error(error instanceof Error ? error.message : "加载知识库失败");
      }
    }
    void loadKnowledgeBases();
  }, []);

  useEffect(() => {
    async function loadSkills() {
      try {
        const data = await fetchSkillDefinitionsAll({
          status: Status.Ok,
        });
        setSkills(data);
      } catch (error) {
        toast.error(error instanceof Error ? error.message : "加载 Skills 失败");
      }
    }
    void loadSkills();
  }, []);

  useEffect(() => {
    async function loadDirectToolOptions() {
      try {
        const servers = await listMCPServers();
        const enabledServers = servers.filter((item) => item.enabled);
        const toolGroups = await Promise.all(
          enabledServers.map(async (server) => {
            const tools = await listMCPTools(server.code);
            return tools.map((tool) => ({
              value: `${server.code}/${tool.name}`,
              label: `${server.code} / ${tool.title || tool.name}`,
              meta: {
                serverCode: server.code,
                toolName: tool.name,
                title: tool.title || tool.name,
                description: tool.description || "",
                arguments: undefined,
              },
            }));
          }),
        );
        setDirectToolOptions(toolGroups.flat());
      } catch (error) {
        toast.error(
          error instanceof Error ? error.message : "加载 Direct Tools 失败",
        );
      }
    }
    void loadDirectToolOptions();
  }, []);

  const aiConfigOptions = useMemo(
    () =>
      aiConfigs.map((item) => ({
        value: String(item.id),
        label: `${item.name} · ${item.modelName}`,
      })),
    [aiConfigs],
  );

  const teamOptions = useMemo(
    () =>
      agentTeams.map((item: AdminAgentTeam) => ({
        value: String(item.id),
        label: item.name,
      })),
    [agentTeams],
  );

  const knowledgeOptions = useMemo(
    () =>
      knowledgeBases.map((item) => ({
        value: String(item.id),
        label: item.name,
      })),
    [knowledgeBases],
  );

  const skillOptions = useMemo(
    () =>
      skills.map((item) => ({
        value: String(item.id),
        label: item.name,
      })),
    [skills],
  );

  const addableKnowledgeOptions = useMemo(
    () =>
      knowledgeOptions.filter(
        (option) => !selectedKnowledgeIds.includes(Number(option.value)),
      ),
    [knowledgeOptions, selectedKnowledgeIds],
  );

  const selectedKnowledgeOptions = useMemo(
    () =>
      selectedKnowledgeIds
        .map((id) =>
          knowledgeOptions.find((option) => Number(option.value) === id),
        )
        .filter(
          (option): option is { value: string; label: string } => !!option,
        ),
    [knowledgeOptions, selectedKnowledgeIds],
  );

  const addableSkillOptions = useMemo(
    () =>
      skillOptions.filter(
        (option) => !selectedSkillIds.includes(Number(option.value)),
      ),
    [selectedSkillIds, skillOptions],
  );

  const selectedSkillOptions = useMemo(
    () =>
      selectedSkillIds
        .map((id) => skillOptions.find((option) => Number(option.value) === id))
        .filter(
          (option): option is { value: string; label: string } => !!option,
        ),
    [selectedSkillIds, skillOptions],
  );

  const addableDirectToolOptions = useMemo(
    () =>
      directToolOptions.filter(
        (option) =>
          option.meta.serverCode === directToolServerCodeToAdd &&
          !directTools.some(
            (tool) =>
              `${tool.serverCode}/${tool.toolName}` === option.value,
          ),
      ),
    [directToolOptions, directToolServerCodeToAdd, directTools],
  );

  const directToolServerOptions = useMemo(
    () =>
      Array.from(
        new Map(
          directToolOptions.map((option) => [
            option.meta.serverCode,
            {
              value: option.meta.serverCode,
              label: option.meta.serverCode,
            },
          ]),
        ).values(),
      ),
    [directToolOptions],
  );

  const directToolsGroupedByServer = useMemo(() => {
    const groups = new Map<
      string,
      CreateAIAgentPayload["directTools"]
    >();
    for (const tool of directTools) {
      const current = groups.get(tool.serverCode) ?? [];
      current.push(tool);
      groups.set(tool.serverCode, current);
    }
    return Array.from(groups.entries());
  }, [directTools]);

  const addableTeamOptions = useMemo(
    () =>
      teamOptions.filter(
        (option) => !selectedTeamIds.includes(Number(option.value)),
      ),
    [teamOptions, selectedTeamIds],
  );

  const selectedTeamOptions = useMemo(
    () =>
      selectedTeamIds
        .map((id) => teamOptions.find((option) => Number(option.value) === id))
        .filter(
          (option): option is { value: string; label: string } => !!option,
        ),
    [selectedTeamIds, teamOptions],
  );

  const handoffMode = watch("handoffMode");

  async function onFormSubmit(values: EditForm) {
    await onSubmit(
      buildPayload(
        values,
        selectedKnowledgeIds,
        selectedTeamIds,
        selectedSkillIds,
        directTools,
      ),
    );
  }

  function handleAddTeam(value: string) {
    const id = Number(value);
    if (!Number.isFinite(id) || id <= 0 || selectedTeamIds.includes(id)) {
      return;
    }
    setSelectedTeamIds((prev) => [...prev, id]);
    setTeamToAdd("");
  }

  function handleRemoveTeam(id: number) {
    setSelectedTeamIds((prev) => prev.filter((item) => item !== id));
  }

  function handleAddKnowledge(value: string) {
    const id = Number(value);
    if (!Number.isFinite(id) || id <= 0 || selectedKnowledgeIds.includes(id)) {
      return;
    }
    setSelectedKnowledgeIds((prev) => [...prev, id]);
    setKnowledgeToAdd("");
  }

  function handleMoveKnowledge(index: number, direction: -1 | 1) {
    const targetIndex = index + direction;
    if (targetIndex < 0 || targetIndex >= selectedKnowledgeIds.length) {
      return;
    }
    setSelectedKnowledgeIds((prev) => {
      const next = [...prev];
      const current = next[index];
      next[index] = next[targetIndex];
      next[targetIndex] = current;
      return next;
    });
  }

  function handleRemoveKnowledge(id: number) {
    setSelectedKnowledgeIds((prev) => prev.filter((item) => item !== id));
  }

  function handleAddSkill(value: string) {
    const id = Number(value);
    if (!Number.isFinite(id) || id <= 0 || selectedSkillIds.includes(id)) {
      return;
    }
    setSelectedSkillIds((prev) => [...prev, id]);
    setSkillToAdd("");
  }

  function handleRemoveSkill(id: number) {
    setSelectedSkillIds((prev) => prev.filter((item) => item !== id));
  }

  function handleAddDirectTool(value: string) {
    const option = directToolOptions.find((item) => item.value === value);
    if (!option) {
      return;
    }
    setDirectTools((prev) => {
      if (
        prev.some(
          (item) =>
            item.serverCode === option.meta.serverCode &&
            item.toolName === option.meta.toolName,
        )
      ) {
        return prev;
      }
      return [...prev, option.meta];
    });
    setDirectToolServerCodeToAdd(option.meta.serverCode);
    setDirectToolToAdd("");
  }

  function handleRemoveDirectTool(value: string) {
    setDirectTools((prev) =>
      prev.filter((item) => `${item.serverCode}/${item.toolName}` !== value),
    );
  }

  return (
    <ProjectDialog
      open={open}
      onOpenChange={onOpenChange}
      title={itemId ? "编辑 AI Agent" : "新建 AI Agent"}
      size="xl"
      allowFullscreen
      defaultFullscreen
      footer={
        <>
          <Button
            type="button"
            variant="outline"
            onClick={() => onOpenChange(false)}
          >
            取消
          </Button>
          <Button type="submit" form={formId} disabled={saving || loading}>
            {saving ? "保存中..." : "保存"}
          </Button>
        </>
      }
    >
      {loading ? (
        <div className="flex items-center justify-center py-12">
          <div className="text-muted-foreground">加载中...</div>
        </div>
      ) : (
        <form
          id={formId}
          onSubmit={handleSubmit(onFormSubmit)}
          className="space-y-4"
        >
          <Field data-invalid={!!errors.name}>
            <FieldLabel htmlFor="ai-agent-name">名称</FieldLabel>
            <FieldContent>
              <Input id="ai-agent-name" {...register("name")} />
              <FieldError errors={[errors.name]} />
            </FieldContent>
          </Field>

          <Field data-invalid={!!errors.description}>
            <FieldLabel htmlFor="ai-agent-description">描述</FieldLabel>
            <FieldContent>
              <Input id="ai-agent-description" {...register("description")} />
              <FieldError errors={[errors.description]} />
            </FieldContent>
          </Field>

          <Field>
            <FieldLabel>客服组</FieldLabel>
            <FieldContent className="space-y-3">
              <OptionCombobox
                value={teamToAdd}
                options={addableTeamOptions}
                placeholder="请选择客服组"
                searchPlaceholder="搜索客服组"
                emptyText="未找到客服组"
                onChange={handleAddTeam}
              />
              <div className="flex flex-wrap gap-2">
                {selectedTeamOptions.length === 0 ? (
                  <span className="text-sm text-muted-foreground">
                    未配置客服组
                  </span>
                ) : (
                  selectedTeamOptions.map((option) => (
                    <Badge
                      key={option.value}
                      variant="secondary"
                      className="gap-1 pr-1"
                    >
                      {option.label}
                      <Button
                        type="button"
                        variant="ghost"
                        size="icon"
                        className="size-5"
                        onClick={() => handleRemoveTeam(Number(option.value))}
                        aria-label={`移除客服组 ${option.label}`}
                      >
                        <Trash2Icon className="size-3" />
                      </Button>
                    </Badge>
                  ))
                )}
              </div>
            </FieldContent>
          </Field>

          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <Field data-invalid={!!errors.serviceMode}>
              <FieldLabel>服务模式</FieldLabel>
              <FieldContent>
                <Controller
                  control={control}
                  name="serviceMode"
                  render={({ field }) => (
                    <OptionCombobox
                      value={field.value}
                      options={serviceModeOptions}
                      placeholder="请选择服务模式"
                      searchPlaceholder="搜索服务模式"
                      emptyText="未找到服务模式"
                      onChange={field.onChange}
                    />
                  )}
                />
                <FieldError errors={[errors.serviceMode]} />
              </FieldContent>
            </Field>
            <Field data-invalid={!!errors.aiConfigId}>
              <FieldLabel>AI 配置</FieldLabel>
              <FieldContent>
                <Controller
                  control={control}
                  name="aiConfigId"
                  render={({ field }) => (
                    <OptionCombobox
                      value={field.value}
                      options={aiConfigOptions}
                      placeholder="请选择 AI 配置"
                      searchPlaceholder="搜索 AI 配置"
                      emptyText="未找到 AI 配置"
                      onChange={field.onChange}
                    />
                  )}
                />
                <FieldError errors={[errors.aiConfigId]} />
              </FieldContent>
            </Field>
          </div>

          <div className="grid grid-cols-1 gap-4 sm:grid-cols-4">
            <Field data-invalid={!!errors.handoffMode}>
              <FieldLabel>转人工模式</FieldLabel>
              <FieldContent>
                <Controller
                  control={control}
                  name="handoffMode"
                  render={({ field }) => (
                    <OptionCombobox
                      value={field.value}
                      options={handoffModeOptions}
                      placeholder="请选择转人工模式"
                      searchPlaceholder="搜索转人工模式"
                      emptyText="未找到转人工模式"
                      onChange={field.onChange}
                    />
                  )}
                />
                <FieldError errors={[errors.handoffMode]} />
              </FieldContent>
            </Field>
            <Field data-invalid={!!errors.replyTimeoutSeconds}>
              <FieldLabel
                htmlFor="ai-agent-reply-timeout-seconds"
                className="flex items-center gap-1"
              >
                回复超时秒数
                <Popover>
                  <PopoverTrigger
                    render={
                      <button
                        type="button"
                        className="inline-flex items-center justify-center text-muted-foreground hover:text-foreground"
                      >
                        <InfoIcon className="size-4" />
                      </button>
                    }
                  />
                  <PopoverContent side="top" align="start" className="max-w-xs">
                    <PopoverDescription>
                      AI 自动回复的异步执行超时时间。填 0 时使用系统默认值 180 秒。
                    </PopoverDescription>
                  </PopoverContent>
                </Popover>
              </FieldLabel>
              <FieldContent>
                <Input
                  id="ai-agent-reply-timeout-seconds"
                  type="number"
                  min={0}
                  step={1}
                  {...register("replyTimeoutSeconds", { valueAsNumber: true })}
                />
                <FieldError errors={[errors.replyTimeoutSeconds]} />
              </FieldContent>
            </Field>
            <Field data-invalid={!!errors.maxAiReplyRounds}>
              <FieldLabel
                htmlFor="ai-agent-max-rounds"
                className="flex items-center gap-1"
              >
                AI 最大回复次数
                <Popover>
                  <PopoverTrigger
                    render={
                      <button
                        type="button"
                        className="inline-flex items-center justify-center text-muted-foreground hover:text-foreground"
                      >
                        <InfoIcon className="size-4" />
                      </button>
                    }
                  />
                  <PopoverContent side="top" align="start" className="max-w-xs">
                    <PopoverDescription>
                      单个会话内 AI
                      成功回复达到该次数后，下一条客户消息会自动转人工。填 0
                      表示不限制。
                    </PopoverDescription>
                  </PopoverContent>
                </Popover>
              </FieldLabel>
              <FieldContent>
                <Input
                  id="ai-agent-max-rounds"
                  type="number"
                  min={0}
                  step={1}
                  {...register("maxAiReplyRounds", { valueAsNumber: true })}
                />
                <FieldError errors={[errors.maxAiReplyRounds]} />
              </FieldContent>
            </Field>
            <Field data-invalid={!!errors.fallbackMode}>
              <FieldLabel>兜底模式</FieldLabel>
              <FieldContent>
                <Controller
                  control={control}
                  name="fallbackMode"
                  render={({ field }) => (
                    <OptionCombobox
                      value={field.value}
                      options={fallbackModeOptions}
                      placeholder="请选择兜底模式"
                      searchPlaceholder="搜索兜底模式"
                      emptyText="未找到兜底模式"
                      onChange={field.onChange}
                    />
                  )}
                />
                <FieldError errors={[errors.fallbackMode]} />
              </FieldContent>
            </Field>
          </div>

          {handoffMode === String(AIAgentHandoffMode.DefaultTeamPool) ? (
            <div className="text-xs text-muted-foreground">
              当前模式要求至少指定一个客服组，转人工后会进入这些客服组对应的待分配范围。
            </div>
          ) : null}

          <Field data-invalid={selectedKnowledgeIds.length === 0}>
            <FieldLabel>知识库</FieldLabel>
            <FieldContent className="space-y-3">
              <div className="flex items-center gap-2">
                <div className="flex-1">
                  <OptionCombobox
                    value={knowledgeToAdd}
                    options={addableKnowledgeOptions}
                    placeholder="选择并添加知识库"
                    searchPlaceholder="搜索知识库"
                    emptyText="没有可添加的知识库"
                    onChange={handleAddKnowledge}
                  />
                </div>
                <Button
                  type="button"
                  variant="outline"
                  disabled={!knowledgeToAdd}
                  onClick={() => handleAddKnowledge(knowledgeToAdd)}
                >
                  <PlusIcon />
                  添加
                </Button>
              </div>
              {selectedKnowledgeIds.length === 0 ? (
                <div className="rounded-md border border-dashed px-3 py-4 text-sm text-muted-foreground">
                  请至少选择一个知识库。
                </div>
              ) : (
                <div className="space-y-2 rounded-md border p-3">
                  {selectedKnowledgeOptions.map((option, index) => (
                    <div key={option.value} className="flex items-center gap-2">
                      <Badge
                        variant="secondary"
                        className="min-w-8 justify-center"
                      >
                        {index + 1}
                      </Badge>
                      <div className="flex-1 text-sm">{option.label}</div>
                      <Button
                        type="button"
                        variant="outline"
                        size="icon-sm"
                        disabled={index === 0}
                        onClick={() => handleMoveKnowledge(index, -1)}
                      >
                        <ArrowUpIcon />
                      </Button>
                      <Button
                        type="button"
                        variant="outline"
                        size="icon-sm"
                        disabled={index === selectedKnowledgeOptions.length - 1}
                        onClick={() => handleMoveKnowledge(index, 1)}
                      >
                        <ArrowDownIcon />
                      </Button>
                      <Button
                        type="button"
                        variant="outline"
                        size="icon-sm"
                        onClick={() =>
                          handleRemoveKnowledge(Number(option.value))
                        }
                      >
                        <Trash2Icon />
                      </Button>
                    </div>
                  ))}
                </div>
              )}
              {selectedKnowledgeIds.length === 0 ? (
                <FieldError errors={[{ message: "请至少选择一个知识库" }]} />
              ) : null}
            </FieldContent>
          </Field>

          <Field>
            <FieldLabel>Skills</FieldLabel>
            <FieldContent className="space-y-3">
              <div className="flex items-center gap-2">
                <div className="flex-1">
                  <OptionCombobox
                    value={skillToAdd}
                    options={addableSkillOptions}
                    placeholder="选择并添加 Skill"
                    searchPlaceholder="搜索 Skill"
                    emptyText="没有可添加的 Skill"
                    onChange={handleAddSkill}
                  />
                </div>
                <Button
                  type="button"
                  variant="outline"
                  disabled={!skillToAdd}
                  onClick={() => handleAddSkill(skillToAdd)}
                >
                  <PlusIcon />
                  添加
                </Button>
              </div>
              <div className="flex flex-wrap gap-2">
                {selectedSkillOptions.length === 0 ? (
                  <span className="text-sm text-muted-foreground">
                    不绑定 Skill 时，自动路由只会走知识库或转人工。
                  </span>
                ) : (
                  selectedSkillOptions.map((option) => (
                    <Badge
                      key={option.value}
                      variant="secondary"
                      className="gap-1 pr-1"
                    >
                      {option.label}
                      <Button
                        type="button"
                        variant="ghost"
                        size="icon"
                        className="size-5"
                        onClick={() => handleRemoveSkill(Number(option.value))}
                        aria-label={`移除 Skill ${option.label}`}
                      >
                        <Trash2Icon className="size-3" />
                      </Button>
                    </Badge>
                  ))
                )}
              </div>
            </FieldContent>
          </Field>

          <Field>
            <FieldLabel>Direct MCP Tools</FieldLabel>
            <FieldContent className="space-y-3">
              <div className="flex items-center gap-2">
                <div className="w-52">
                  <OptionCombobox
                    value={directToolServerCodeToAdd}
                    options={directToolServerOptions}
                    placeholder="选择 MCP Server"
                    searchPlaceholder="搜索 MCP Server"
                    emptyText="没有可用的 MCP Server"
                    onChange={(value) => {
                      setDirectToolServerCodeToAdd(value);
                      setDirectToolToAdd("");
                    }}
                  />
                </div>
                <div className="flex-1">
                  <OptionCombobox
                    value={directToolToAdd}
                    options={addableDirectToolOptions}
                    placeholder="选择该 Server 下的 Direct Tool"
                    searchPlaceholder="搜索 Direct Tool"
                    emptyText="没有可添加的 Direct Tool"
                    onChange={handleAddDirectTool}
                  />
                </div>
                <Button
                  type="button"
                  variant="outline"
                  disabled={!directToolServerCodeToAdd || !directToolToAdd}
                  onClick={() => handleAddDirectTool(directToolToAdd)}
                >
                  <PlusIcon />
                  添加
                </Button>
              </div>
              <div className="space-y-3">
                {directTools.length === 0 ? (
                  <span className="text-sm text-muted-foreground">
                    不配置 Direct Tool 时，Agent 不会直接访问 MCP，只能通过 Skill 间接调用。
                  </span>
                ) : (
                  directToolsGroupedByServer.map(([serverCode, tools]) => (
                    <div key={serverCode} className="rounded-md border p-3">
                      <div className="mb-2 text-xs font-medium text-muted-foreground">
                        {serverCode}
                      </div>
                      <div className="flex flex-wrap gap-2">
                        {tools.map((tool) => {
                          const value = `${tool.serverCode}/${tool.toolName}`;
                          return (
                            <Badge
                              key={value}
                              variant="secondary"
                              className="gap-1 pr-1"
                            >
                              {tool.title || value}
                              <Button
                                type="button"
                                variant="ghost"
                                size="icon"
                                className="size-5"
                                onClick={() => handleRemoveDirectTool(value)}
                                aria-label={`移除 Direct Tool ${value}`}
                              >
                                <Trash2Icon className="size-3" />
                              </Button>
                            </Badge>
                          );
                        })}
                      </div>
                    </div>
                  ))
                )}
              </div>
            </FieldContent>
          </Field>

          <Field data-invalid={!!errors.welcomeMessage}>
            <FieldLabel htmlFor="ai-agent-welcome-message">欢迎语</FieldLabel>
            <FieldContent>
              <Textarea
                id="ai-agent-welcome-message"
                rows={3}
                {...register("welcomeMessage")}
              />
              <FieldError errors={[errors.welcomeMessage]} />
            </FieldContent>
          </Field>

          <Field data-invalid={!!errors.systemPrompt}>
            <FieldLabel htmlFor="ai-agent-system-prompt">系统提示词</FieldLabel>
            <FieldContent>
              <Textarea
                id="ai-agent-system-prompt"
                rows={8}
                {...register("systemPrompt")}
              />
              <FieldError errors={[errors.systemPrompt]} />
            </FieldContent>
          </Field>

          <Field data-invalid={!!errors.fallbackMessage}>
            <FieldLabel htmlFor="ai-agent-fallback-message">
              兜底文案
            </FieldLabel>
            <FieldContent>
              <Textarea
                id="ai-agent-fallback-message"
                rows={3}
                {...register("fallbackMessage")}
              />
              <FieldError errors={[errors.fallbackMessage]} />
            </FieldContent>
          </Field>

          <Field data-invalid={!!errors.remark}>
            <FieldLabel htmlFor="ai-agent-remark">备注</FieldLabel>
            <FieldContent>
              <Textarea id="ai-agent-remark" rows={3} {...register("remark")} />
              <FieldError errors={[errors.remark]} />
            </FieldContent>
          </Field>
        </form>
      )}
    </ProjectDialog>
  );
}
