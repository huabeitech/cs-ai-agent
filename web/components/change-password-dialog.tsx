"use client";

import { useEffect } from "react";
import { zodResolver } from "@hookform/resolvers/zod";
import { Resolver, useForm } from "react-hook-form";
import { z } from "zod/v4";
import { toast } from "sonner";

import { changeSelfPassword } from "@/lib/api/admin";
import { Button } from "@/components/ui/button";
import {
  Field,
  FieldContent,
  FieldError,
  FieldLabel,
} from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { ProjectDialog } from "@/components/project-dialog";

const changePasswordSchema = z
  .object({
    password: z.string().trim().min(1, "新密码不能为空"),
    confirmPassword: z.string().trim().min(1, "确认密码不能为空"),
  })
  .refine((data) => data.password === data.confirmPassword, {
    path: ["confirmPassword"],
    message: "两次输入的密码不一致",
  });

type ChangePasswordForm = z.infer<typeof changePasswordSchema>;

const changePasswordResolver = zodResolver(
  changePasswordSchema as never,
) as Resolver<
  z.input<typeof changePasswordSchema>,
  undefined,
  z.output<typeof changePasswordSchema>
>;

const emptyForm: ChangePasswordForm = {
  password: "",
  confirmPassword: "",
};

type ChangePasswordDialogProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSuccess: () => Promise<void>;
};

export function ChangePasswordDialog({
  open,
  onOpenChange,
  onSuccess,
}: ChangePasswordDialogProps) {
  const form = useForm<
    z.input<typeof changePasswordSchema>,
    undefined,
    z.output<typeof changePasswordSchema>
  >({
    resolver: changePasswordResolver,
    defaultValues: emptyForm,
  });
  const {
    handleSubmit,
    register,
    reset,
    formState: { errors, isSubmitting },
  } = form;

  useEffect(() => {
    if (open) {
      reset(emptyForm);
    }
  }, [open, reset]);

  async function onSubmit(values: ChangePasswordForm) {
    try {
      await changeSelfPassword(values.password.trim());
      toast.success("密码已修改，请重新登录");
      onOpenChange(false);
      await onSuccess();
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "修改密码失败");
    }
  }

  return (
    <ProjectDialog
      open={open}
      onOpenChange={onOpenChange}
      title="修改密码"
      description="修改当前登录账号的密码，提交后需要重新登录。"
      size="sm"
      allowFullscreen
      footer={
        <>
          <Button
            type="button"
            variant="outline"
            onClick={() => onOpenChange(false)}
            disabled={isSubmitting}
          >
            取消
          </Button>
          <Button type="submit" form="change-password-form" disabled={isSubmitting}>
            {isSubmitting ? "提交中..." : "确认修改"}
          </Button>
        </>
      }
    >
      <form id="change-password-form" onSubmit={handleSubmit(onSubmit)}>
        <div className="space-y-4 px-6 py-4">
          <Field data-invalid={!!errors.password}>
            <FieldLabel htmlFor="change-password-password">新密码</FieldLabel>
            <FieldContent>
              <Input
                id="change-password-password"
                type="password"
                placeholder="请输入新密码"
                autoComplete="new-password"
                aria-invalid={!!errors.password}
                {...register("password")}
              />
              <FieldError errors={[errors.password]} />
            </FieldContent>
          </Field>
          <Field data-invalid={!!errors.confirmPassword}>
            <FieldLabel htmlFor="change-password-confirm">确认密码</FieldLabel>
            <FieldContent>
              <Input
                id="change-password-confirm"
                type="password"
                placeholder="请再次输入新密码"
                autoComplete="new-password"
                aria-invalid={!!errors.confirmPassword}
                {...register("confirmPassword")}
              />
              <FieldError errors={[errors.confirmPassword]} />
            </FieldContent>
          </Field>
        </div>
      </form>
    </ProjectDialog>
  );
}
