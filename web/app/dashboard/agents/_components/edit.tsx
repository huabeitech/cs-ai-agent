"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { CheckIcon, ChevronsUpDownIcon } from "lucide-react";
import { useCallback, useEffect, useState } from "react";
import { Controller, Resolver, useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod/v4";

import { ImageInput } from "@/components/image-input";
import { ProjectDialog } from "@/components/project-dialog";
import { Button } from "@/components/ui/button";
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "@/components/ui/command";
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
  PopoverTrigger,
} from "@/components/ui/popover";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Switch } from "@/components/ui/switch";
import { Textarea } from "@/components/ui/textarea";
import {
  fetchAgentProfile,
  fetchUsersAll,
  type AdminAgentProfile,
  type AdminUser,
  type CreateAdminAgentProfilePayload
} from "@/lib/api/admin";
import {
  ServiceStatus,
  ServiceStatusLabels,
} from "@/lib/generated/enums";
import { getEnumLabel, getEnumOptions } from "@/lib/enums";

type AgentEditDialogProps = {
  open: boolean;
  saving: boolean;
  itemId: number | null;
  defaultTeamId: number | null;
  onOpenChange: (open: boolean) => void;
  onSubmit: (payload: CreateAdminAgentProfilePayload) => Promise<void>;
};

const serviceStatusOptions = getEnumOptions(ServiceStatusLabels);
const emptyForm: EditForm = {
  userId: "",
  teamId: "",
  agentCode: "",
  displayName: "",
  avatar: "",
  serviceStatus: String(ServiceStatus.Idle) as "0" | "1",
  maxConcurrentCount: "0",
  priorityLevel: "0",
  autoAssignEnabled: true,
  receiveOfflineMessage: false,
  remark: "",
};

const editFormSchema = z.object({
  userId: z.string().trim().min(1, "请选择关联用户"),
  teamId: z.string().trim().min(1, "请选择所属客服组"),
  agentCode: z.string().trim().min(1, "客服工号不能为空"),
  displayName: z.string().trim().min(1, "展示名不能为空"),
  avatar: z.string().trim(),
  serviceStatus: z.enum(["0", "1"], {
    message: "请选择客服状态",
  }),
  maxConcurrentCount: z
    .string()
    .trim()
    .regex(/^\d+$/, "最大并发必须是大于等于 0 的整数"),
  priorityLevel: z
    .string()
    .trim()
    .regex(/^-?\d+$/, "优先级必须是整数"),
  autoAssignEnabled: z.boolean(),
  receiveOfflineMessage: z.boolean(),
  remark: z.string().trim(),
});

type EditForm = z.infer<typeof editFormSchema>;
const editFormResolver = zodResolver(editFormSchema as never) as Resolver<
  z.input<typeof editFormSchema>,
  undefined,
  z.output<typeof editFormSchema>
>;

function getStatusLabel(value: string) {
  return getEnumLabel(
    ServiceStatusLabels,
    Number(value) as ServiceStatus,
  );
}

function buildForm(item: AdminAgentProfile | null): EditForm {
  if (!item) {
    return emptyForm;
  }
  return {
    userId: String(item.userId),
    teamId: String(item.teamId),
    agentCode: item.agentCode,
    displayName: item.displayName,
    avatar: item.avatar || "",
    serviceStatus: String(item.serviceStatus) as EditForm["serviceStatus"],
    maxConcurrentCount: String(item.maxConcurrentCount),
    priorityLevel: String(item.priorityLevel),
    autoAssignEnabled: item.autoAssignEnabled,
    receiveOfflineMessage: item.receiveOfflineMessage,
    remark: item.remark || "",
  };
}

function buildFormWithDefaultTeam(
  item: AdminAgentProfile | null,
  defaultTeamId: number | null,
): EditForm {
  const form = buildForm(item);
  if (!item && defaultTeamId) {
    return {
      ...form,
      teamId: String(defaultTeamId),
    };
  }
  return form;
}

function buildPayload(form: EditForm): CreateAdminAgentProfilePayload {
  return {
    userId: Number(form.userId),
    teamId: Number(form.teamId),
    agentCode: form.agentCode.trim(),
    displayName: form.displayName.trim(),
    avatar: form.avatar.trim(),
    serviceStatus: Number(form.serviceStatus),
    maxConcurrentCount: Number(form.maxConcurrentCount),
    priorityLevel: Number(form.priorityLevel),
    autoAssignEnabled: form.autoAssignEnabled,
    receiveOfflineMessage: form.receiveOfflineMessage,
    remark: form.remark.trim(),
  };
}

export function EditDialog({
  open,
  saving,
  itemId,
  defaultTeamId,
  onOpenChange,
  onSubmit,
}: AgentEditDialogProps) {
  if (!open) {
    return null;
  }

  return (
    <AgentEditDialogBody
      key={itemId ? `edit-${itemId}` : "create"}
      open={open}
      itemId={itemId}
      defaultTeamId={defaultTeamId}
      saving={saving}
      onOpenChange={onOpenChange}
      onSubmit={onSubmit}
    />
  );
}

type AgentEditDialogBodyProps = {
  open: boolean;
  saving: boolean;
  itemId: number | null;
  defaultTeamId: number | null;
  onOpenChange: (open: boolean) => void;
  onSubmit: (payload: CreateAdminAgentProfilePayload) => Promise<void>;
};

function AgentEditDialogBody({
  open,
  saving,
  itemId,
  defaultTeamId,
  onOpenChange,
  onSubmit,
}: AgentEditDialogBodyProps) {
  const [users, setUsers] = useState<AdminUser[]>([]);
  const [userSelectOpen, setUserSelectOpen] = useState(false);
  const [loading, setLoading] = useState(false);
  const userOptions = users.map((user) => ({
    value: String(user.id),
    label: `${user.nickname || user.username} (${user.username})`,
  }));
  const loadOptions = useCallback(async () => {
    try {
      const [usersData] = await Promise.all([
        fetchUsersAll(),
      ]);
      setUsers(usersData);
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "加载选项失败");
    }
  }, []);
  const form = useForm<
    z.input<typeof editFormSchema>,
    undefined,
    z.output<typeof editFormSchema>
  >({
    resolver: editFormResolver,
    defaultValues: buildFormWithDefaultTeam(null, defaultTeamId),
  });
  const {
    control,
    handleSubmit,
    reset,
    register,
    formState: { errors },
  } = form;

  useEffect(() => {
    async function loadDetail() {
      if (!itemId) {
        reset(buildFormWithDefaultTeam(null, defaultTeamId));
        return;
      }
      setLoading(true);
      try {
        const data = await fetchAgentProfile(itemId);
        reset(buildForm(data));
      } catch (error) {
        toast.error(error instanceof Error ? error.message : "加载客服档案详情失败");
      } finally {
        setLoading(false);
      }
    }
    void loadDetail();
  }, [itemId, defaultTeamId, reset]);

  useEffect(() => {
    if (open) {
      void loadOptions();
    }
  }, [loadOptions, open]);

  async function onFormSubmit(values: EditForm) {
    await onSubmit(buildPayload(values));
  }

  const formId = "agent-edit-form";

  return (
    <ProjectDialog
      open={open}
      onOpenChange={onOpenChange}
      title={itemId ? "编辑客服档案" : "新建客服档案"}
      size="lg"
      footer={
        <>
          <Button
            type="button"
            variant="outline"
            onClick={() => onOpenChange(false)}
            disabled={saving}
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
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <Field data-invalid={!!errors.userId}>
              <FieldLabel>关联用户</FieldLabel>
              <FieldContent>
                <Controller
                  control={control}
                  name="userId"
                  render={({ field }) => (
                    <Popover
                      open={userSelectOpen}
                      onOpenChange={setUserSelectOpen}
                    >
                      <PopoverTrigger
                        render={
                          <Button
                            variant="outline"
                            role="combobox"
                            aria-expanded={userSelectOpen}
                            className="w-full justify-between font-normal"
                          />
                        }
                      >
                        <span className="truncate">
                          {userOptions.find(
                            (option) => option.value === field.value,
                          )?.label ?? "请选择用户"}
                        </span>
                        <ChevronsUpDownIcon className="ml-2 size-4 shrink-0 opacity-50" />
                      </PopoverTrigger>
                      <PopoverContent
                        className="w-[var(--radix-popper-anchor-width)] p-0"
                        align="start"
                      >
                        <Command>
                          <CommandInput placeholder="搜索用户..." />
                          <CommandList>
                            <CommandEmpty>没有匹配的用户</CommandEmpty>
                            <CommandGroup>
                              {userOptions.map((option) => (
                                <CommandItem
                                  key={option.value}
                                  value={option.label}
                                  onSelect={() => {
                                    field.onChange(option.value);
                                    setUserSelectOpen(false);
                                  }}
                                >
                                  <CheckIcon
                                    className={`mr-2 size-4 ${
                                      field.value === option.value
                                        ? "opacity-100"
                                        : "opacity-0"
                                    }`}
                                  />
                                  {option.label}
                                </CommandItem>
                              ))}
                            </CommandGroup>
                          </CommandList>
                        </Command>
                      </PopoverContent>
                    </Popover>
                  )}
                />
                <FieldError errors={[errors.userId]} />
              </FieldContent>
            </Field>
            <Field data-invalid={!!errors.displayName}>
              <FieldLabel htmlFor="agent-display-name">展示名</FieldLabel>
              <FieldContent>
                <Input
                  id="agent-display-name"
                  placeholder="请输入展示名"
                  {...register("displayName")}
                />
                <FieldError errors={[errors.displayName]} />
              </FieldContent>
            </Field>
          </div>

          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <Field data-invalid={!!errors.agentCode}>
              <FieldLabel htmlFor="agent-code">客服工号</FieldLabel>
              <FieldContent>
                <Input
                  id="agent-code"
                  placeholder="例如：A1001"
                  {...register("agentCode")}
                />
                <FieldError errors={[errors.agentCode]} />
              </FieldContent>
            </Field>

            <Field className="min-h-32">
              <FieldLabel>头像</FieldLabel>
              <FieldContent>
                <Controller
                  control={control}
                  name="avatar"
                  render={({ field }) => (
                    <ImageInput
                      value={field.value}
                      onChange={field.onChange}
                      disabled={saving}
                      prefix="avatar"
                      placeholder="上传头像"
                      className="size-16 rounded-full"
                    />
                  )}
                />
              </FieldContent>
            </Field>
          </div>

          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <Field data-invalid={!!errors.serviceStatus}>
              <FieldLabel>客服状态</FieldLabel>
              <FieldContent>
                <Controller
                  control={control}
                  name="serviceStatus"
                  render={({ field }) => (
                    <Select
                      value={field.value}
                      onValueChange={field.onChange}
                      modal={false}
                    >
                      <SelectTrigger className="w-full">
                        <SelectValue>{getStatusLabel(field.value)}</SelectValue>
                      </SelectTrigger>
                      <SelectContent>
                        {serviceStatusOptions.map((option) => (
                          <SelectItem
                            key={option.value}
                            value={String(option.value)}
                          >
                            {option.label}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  )}
                />
                <FieldError errors={[errors.serviceStatus]} />
              </FieldContent>
            </Field>
          </div>

          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <Field data-invalid={!!errors.maxConcurrentCount}>
              <FieldLabel htmlFor="agent-max-concurrent-count">
                最大并发
              </FieldLabel>
              <FieldContent>
                <Input
                  id="agent-max-concurrent-count"
                  type="number"
                  min={0}
                  {...register("maxConcurrentCount")}
                />
                <FieldError errors={[errors.maxConcurrentCount]} />
              </FieldContent>
            </Field>
            <Field data-invalid={!!errors.priorityLevel}>
              <FieldLabel htmlFor="agent-priority-level">优先级</FieldLabel>
              <FieldContent>
                <Input
                  id="agent-priority-level"
                  type="number"
                  step={1}
                  {...register("priorityLevel")}
                />
                <FieldError errors={[errors.priorityLevel]} />
              </FieldContent>
            </Field>
          </div>

          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <Field>
              <FieldLabel>参与自动分配</FieldLabel>
              <FieldContent>
                <Controller
                  control={control}
                  name="autoAssignEnabled"
                  render={({ field }) => (
                    <Switch
                      checked={field.value}
                      onCheckedChange={field.onChange}
                    />
                  )}
                />
              </FieldContent>
            </Field>
            <Field>
              <FieldLabel>离线接收消息</FieldLabel>
              <FieldContent>
                <Controller
                  control={control}
                  name="receiveOfflineMessage"
                  render={({ field }) => (
                    <Switch
                      checked={field.value}
                      onCheckedChange={field.onChange}
                    />
                  )}
                />
              </FieldContent>
            </Field>
          </div>

          <Field>
            <FieldLabel htmlFor="agent-remark">备注</FieldLabel>
            <FieldContent>
              <Textarea
                id="agent-remark"
                rows={4}
                placeholder="请输入备注"
                {...register("remark")}
              />
            </FieldContent>
          </Field>
        </form>
      )}
    </ProjectDialog>
  );
}
