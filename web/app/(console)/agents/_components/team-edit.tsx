"use client";

import { useCallback, useEffect, useState } from "react";
import { zodResolver } from "@hookform/resolvers/zod";
import { CheckIcon, ChevronsUpDownIcon } from "lucide-react";
import { Controller, Resolver, useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod/v4";

import {
  type AdminAgentTeam,
  type CreateAdminAgentTeamPayload,
  fetchAgentTeam,
  fetchUsersAll,
  type AdminUser,
} from "@/lib/api/admin";
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
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
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
import { Textarea } from "@/components/ui/textarea";
import { Status, StatusLabels } from "@/lib/generated/enums";
import { getEnumOptions } from "@/lib/enums";

type TeamEditDialogProps = {
  open: boolean;
  saving: boolean;
  itemId: number | null;
  onOpenChange: (open: boolean) => void;
  onSubmit: (payload: CreateAdminAgentTeamPayload) => Promise<void>;
};

const statusOptions = getEnumOptions(StatusLabels)
  .filter((option) => option.value !== Status.Deleted)
  .map((option) => ({
    value: String(option.value),
    label: option.label,
  }));

const emptyForm: EditForm = {
  name: "",
  leaderUserId: "0",
  status: String(Status.Ok),
  description: "",
  remark: "",
};

const editFormSchema = z.object({
  name: z.string().trim().min(1, "客服组名称不能为空"),
  leaderUserId: z.string().trim().regex(/^\d+$/, "组长用户不合法"),
  status: z.enum([String(Status.Ok), String(Status.Disabled)], {
    message: "请选择状态",
  }),
  description: z.string().trim(),
  remark: z.string().trim(),
});

type EditForm = z.infer<typeof editFormSchema>;
const editFormResolver = zodResolver(editFormSchema as never) as Resolver<
  z.input<typeof editFormSchema>,
  undefined,
  z.output<typeof editFormSchema>
>;

function buildForm(item: AdminAgentTeam | null): EditForm {
  if (!item) {
    return emptyForm;
  }
  return {
    name: item.name,
    leaderUserId: String(item.leaderUserId),
    status: String(item.status),
    description: item.description || "",
    remark: item.remark || "",
  };
}

function buildPayload(form: EditForm): CreateAdminAgentTeamPayload {
  return {
    name: form.name.trim(),
    leaderUserId: Number(form.leaderUserId),
    status: Number(form.status),
    description: form.description.trim(),
    remark: form.remark.trim(),
  };
}

export function EditDialog({
  open,
  saving,
  itemId,
  onOpenChange,
  onSubmit,
}: TeamEditDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      {open ? (
        <TeamEditDialogBody
          key={itemId ? `edit-${itemId}` : "create"}
          itemId={itemId}
          saving={saving}
          onOpenChange={onOpenChange}
          onSubmit={onSubmit}
        />
      ) : null}
    </Dialog>
  );
}

type TeamEditDialogBodyProps = Omit<TeamEditDialogProps, "open">;

function TeamEditDialogBody({
  saving,
  itemId,
  onOpenChange,
  onSubmit,
}: TeamEditDialogBodyProps) {
  const [users, setUsers] = useState<AdminUser[]>([]);
  const [userSelectOpen, setUserSelectOpen] = useState(false);
  const [loading, setLoading] = useState(false);
  const userOptions = users.map((user) => ({
    value: String(user.id),
    label: `${user.nickname || user.username} (${user.username})`,
  }));
  const loadUsers = useCallback(async () => {
    try {
      const data = await fetchUsersAll();
      setUsers(data);
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "加载用户选项失败");
    }
  }, []);
  const form = useForm<
    z.input<typeof editFormSchema>,
    undefined,
    z.output<typeof editFormSchema>
  >({
    resolver: editFormResolver,
    defaultValues: emptyForm,
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
        reset(emptyForm);
        return;
      }
      setLoading(true);
      try {
        const data = await fetchAgentTeam(itemId);
        reset(buildForm(data));
      } catch (error) {
        toast.error(error instanceof Error ? error.message : "加载客服组详情失败");
      } finally {
        setLoading(false);
      }
    }
    void loadDetail();
  }, [itemId, reset]);

  useEffect(() => {
    void loadUsers();
  }, [loadUsers]);

  async function onFormSubmit(values: EditForm) {
    await onSubmit(buildPayload(values));
  }

  return (
    <DialogContent className="max-w-xl gap-0 p-0 sm:max-w-xl">
      <DialogHeader className="px-6 pt-6">
        <DialogTitle>{itemId ? "编辑" : "新建"}</DialogTitle>
      </DialogHeader>
      {loading ? (
        <div className="flex items-center justify-center py-12">
          <div className="text-muted-foreground">加载中...</div>
        </div>
      ) : (
        <form onSubmit={handleSubmit(onFormSubmit)}>
          <div className="space-y-4 p-6">
            <Field data-invalid={!!errors.name}>
              <FieldLabel htmlFor="agent-team-name">客服组名称</FieldLabel>
              <FieldContent>
                <Input
                  id="agent-team-name"
                  placeholder="请输入客服组名称"
                  {...register("name")}
                />
                <FieldError errors={[errors.name]} />
              </FieldContent>
            </Field>
            <Field data-invalid={!!errors.leaderUserId}>
              <FieldLabel>组长</FieldLabel>
              <FieldContent>
                <Controller
                  control={control}
                  name="leaderUserId"
                  render={({ field }) => (
                    <Popover open={userSelectOpen} onOpenChange={setUserSelectOpen}>
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
                          {field.value === "0"
                            ? "暂不设置"
                            : userOptions.find((option) => option.value === field.value)?.label ?? "请选择组长"}
                        </span>
                        <ChevronsUpDownIcon className="ml-2 size-4 shrink-0 opacity-50" />
                      </PopoverTrigger>
                      <PopoverContent className="w-[var(--radix-popper-anchor-width)] p-0" align="start">
                        <Command>
                          <CommandInput placeholder="搜索用户..." />
                          <CommandList>
                            <CommandEmpty>没有匹配的用户</CommandEmpty>
                            <CommandGroup>
                              <CommandItem
                                value="暂不设置"
                                onSelect={() => {
                                  field.onChange("0");
                                  setUserSelectOpen(false);
                                }}
                              >
                                <CheckIcon
                                  className={`mr-2 size-4 ${field.value === "0" ? "opacity-100" : "opacity-0"}`}
                                />
                                暂不设置
                              </CommandItem>
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
                                      field.value === option.value ? "opacity-100" : "opacity-0"
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
                <FieldError errors={[errors.leaderUserId]} />
              </FieldContent>
            </Field>
            <Field data-invalid={!!errors.status}>
              <FieldLabel>状态</FieldLabel>
              <FieldContent>
                <Controller
                  control={control}
                  name="status"
                  render={({ field }) => (
                    <Select
                      value={field.value}
                      onValueChange={field.onChange}
                      modal={false}
                    >
                      <SelectTrigger className="w-full">
                        <SelectValue>
                          {statusOptions.find(
                            (item) => item.value === field.value,
                          )?.label ?? "请选择状态"}
                        </SelectValue>
                      </SelectTrigger>
                      <SelectContent>
                        {statusOptions.map((option) => (
                          <SelectItem key={option.value} value={option.value}>
                            {option.label}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  )}
                />
                <FieldError errors={[errors.status]} />
              </FieldContent>
            </Field>
            <Field>
              <FieldLabel htmlFor="agent-team-description">职责说明</FieldLabel>
              <FieldContent>
                <Input
                  id="agent-team-description"
                  placeholder="例如：负责售前咨询与线索转化"
                  {...register("description")}
                />
              </FieldContent>
            </Field>
            <Field>
              <FieldLabel htmlFor="agent-team-remark">备注</FieldLabel>
              <FieldContent>
                <Textarea
                  id="agent-team-remark"
                  rows={4}
                  placeholder="请输入备注"
                  {...register("remark")}
                />
              </FieldContent>
            </Field>
          </div>
          <DialogFooter className="mx-0 mb-0 px-6 py-4">
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
              disabled={saving}
            >
              取消
            </Button>
            <Button type="submit" disabled={saving || loading}>
              {saving ? "保存中..." : "保存"}
            </Button>
          </DialogFooter>
        </form>
      )}
    </DialogContent>
  );
}
