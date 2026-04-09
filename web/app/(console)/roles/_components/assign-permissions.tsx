"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { SearchIcon } from "lucide-react";
import { useEffect, useMemo, useState } from "react";
import { Controller, Resolver, useForm } from "react-hook-form";
import { z } from "zod/v4";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import {
  Drawer,
  DrawerContent,
  DrawerDescription,
  DrawerFooter,
  DrawerHeader,
  DrawerTitle,
} from "@/components/ui/drawer";
import {
  Field,
  FieldContent,
  FieldError,
  FieldLabel,
} from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import type { AdminPermission, AdminRole } from "@/lib/api/admin";

type AssignPermissionsDrawerProps = {
  open: boolean;
  saving: boolean;
  loading: boolean;
  item: AdminRole | null;
  permissions: AdminPermission[];
  selectedPermissionIds: number[];
  onOpenChange: (open: boolean) => void;
  onSubmit: (permissionIds: number[]) => Promise<void>;
};

const assignPermissionsSchema = z.object({
  permissionIds: z.array(z.number().int().positive()),
});

type AssignPermissionsForm = z.infer<typeof assignPermissionsSchema>;

const assignPermissionsResolver = zodResolver(
  assignPermissionsSchema as never,
) as Resolver<
  z.input<typeof assignPermissionsSchema>,
  undefined,
  z.output<typeof assignPermissionsSchema>
>;

function buildForm(selectedPermissionIds: number[]): AssignPermissionsForm {
  return {
    permissionIds: selectedPermissionIds,
  };
}

export function AssignPermissionsDrawer({
  open,
  saving,
  loading,
  item,
  permissions,
  selectedPermissionIds,
  onOpenChange,
  onSubmit,
}: AssignPermissionsDrawerProps) {
  return (
    <Drawer open={open} onOpenChange={onOpenChange} direction="right">
      {open ? (
        <AssignPermissionsDrawerBody
          key={item ? `assign-permissions-${item.id}` : "assign-permissions"}
          saving={saving}
          loading={loading}
          item={item}
          permissions={permissions}
          selectedPermissionIds={selectedPermissionIds}
          onOpenChange={onOpenChange}
          onSubmit={onSubmit}
        />
      ) : null}
    </Drawer>
  );
}

type AssignPermissionsDrawerBodyProps = {
  saving: boolean;
  loading: boolean;
  item: AdminRole | null;
  permissions: AdminPermission[];
  selectedPermissionIds: number[];
  onOpenChange: (open: boolean) => void;
  onSubmit: (permissionIds: number[]) => Promise<void>;
};

function AssignPermissionsDrawerBody({
  saving,
  loading,
  item,
  permissions,
  selectedPermissionIds,
  onOpenChange,
  onSubmit,
}: AssignPermissionsDrawerBodyProps) {
  const [keyword, setKeyword] = useState("");
  const form = useForm<
    z.input<typeof assignPermissionsSchema>,
    undefined,
    z.output<typeof assignPermissionsSchema>
  >({
    resolver: assignPermissionsResolver,
    defaultValues: buildForm(selectedPermissionIds),
  });
  const {
    control,
    handleSubmit,
    reset,
    formState: { errors },
  } = form;

  useEffect(() => {
    reset(buildForm(selectedPermissionIds));
  }, [reset, selectedPermissionIds]);

  const groupedPermissions = useMemo(() => {
    const output = keyword.trim().toLowerCase();
    const filtered = output
      ? permissions.filter((permission) =>
          `${permission.name} ${permission.code} ${permission.groupName} ${permission.apiPath}`
            .toLowerCase()
            .includes(output),
        )
      : permissions;

    const groups = new Map<string, AdminPermission[]>();
    filtered.forEach((permission) => {
      const groupName = permission.groupName || "default";
      const list = groups.get(groupName) ?? [];
      list.push(permission);
      groups.set(groupName, list);
    });
    return Array.from(groups.entries()).sort(([left], [right]) =>
      left.localeCompare(right, "zh-CN"),
    );
  }, [keyword, permissions]);

  async function onFormSubmit(values: AssignPermissionsForm) {
    await onSubmit(values.permissionIds);
  }

  return (
    <DrawerContent className="min-w-3xl">
      <DrawerHeader>
        <DrawerTitle>分配权限</DrawerTitle>
        <DrawerDescription>
          当前角色：{item?.name || "-"} {item?.code ? `(${item.code})` : ""}
        </DrawerDescription>
      </DrawerHeader>
      <form
        className="flex h-full min-h-0 flex-col"
        onSubmit={handleSubmit(onFormSubmit)}
      >
        <div className="flex-1 flex flex-col min-h-0 space-y-4 px-4 pb-4">
          <Field data-invalid={!!errors.permissionIds} className="flex-1 flex flex-col min-h-0">
            <FieldLabel>权限列表</FieldLabel>
            <FieldContent className="flex-1 min-h-0 flex flex-col">
              <div className="relative mb-2">
                <SearchIcon className="pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground" />
                <Input
                  value={keyword}
                  onChange={(event) => setKeyword(event.target.value)}
                  placeholder="搜索权限名称、编码、分组或接口"
                  className="pl-9"
                  disabled={loading}
                />
              </div>
              <Controller
                control={control}
                name="permissionIds"
                render={({ field }) => {
                  const value = field.value || [];

                  return (
                    <div className="flex-1 min-h-0 space-y-4 overflow-y-auto rounded-xl border p-3">
                      {loading ? (
                        <div className="py-8 text-center text-sm text-muted-foreground">
                          正在加载权限列表...
                        </div>
                      ) : groupedPermissions.length > 0 ? (
                        groupedPermissions.map(([groupName, list]) => (
                          <section key={groupName} className="space-y-2">
                            <div className="flex items-center gap-2">
                              <Badge variant="outline">{groupName}</Badge>
                              <span className="text-xs text-muted-foreground">
                                {list.length} 项
                              </span>
                            </div>
                            <div className="space-y-1">
                              {list.map((permission) => {
                                const checked = value.includes(permission.id);
                                return (
                                  <label
                                    key={permission.id}
                                    className="flex cursor-pointer items-start gap-3 rounded-lg border border-transparent px-3 py-2 hover:bg-muted/50"
                                  >
                                    <Checkbox
                                      checked={checked}
                                      onCheckedChange={(nextChecked) => {
                                        if (nextChecked) {
                                          field.onChange([
                                            ...value,
                                            permission.id,
                                          ]);
                                          return;
                                        }
                                        field.onChange(
                                          value.filter(
                                            (currentId) =>
                                              currentId !== permission.id,
                                          ),
                                        );
                                      }}
                                    />
                                    <div className="min-w-0 flex-1">
                                      <div className="font-medium">
                                        {permission.name}
                                      </div>
                                      <div className="text-sm text-muted-foreground">
                                        {permission.code}
                                      </div>
                                      <div className="text-xs text-muted-foreground">
                                        {permission.method || "ANY"}{" "}
                                        {permission.apiPath || "-"}
                                      </div>
                                    </div>
                                    <Badge
                                      variant={
                                        permission.status === 0
                                          ? "default"
                                          : "secondary"
                                      }
                                    >
                                      {permission.status === 1
                                        ? "禁用"
                                        : "启用"}
                                    </Badge>
                                  </label>
                                );
                              })}
                            </div>
                          </section>
                        ))
                      ) : (
                        <div className="py-8 text-center text-sm text-muted-foreground">
                          没有匹配的权限
                        </div>
                      )}
                    </div>
                  );
                }}
              />
              <FieldError errors={[errors.permissionIds]} />
            </FieldContent>
          </Field>
        </div>
        <DrawerFooter className="border-t">
          <Button type="submit" disabled={saving || loading || !item}>
            {saving ? "保存中..." : "保存权限"}
          </Button>
          <Button
            type="button"
            variant="outline"
            onClick={() => onOpenChange(false)}
            disabled={saving}
          >
            取消
          </Button>
        </DrawerFooter>
      </form>
    </DrawerContent>
  );
}
