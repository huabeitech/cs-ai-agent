"use client"

import Image from "next/image"
import { useRouter, useSearchParams } from "next/navigation"
import { startTransition, useEffect, useState } from "react"
import { toast } from "sonner"

import { useAuth } from "@/components/auth-provider"
import { loginWithPassword } from "@/lib/api/auth"
import { cn } from "@/lib/utils"
import { Button } from "@/components/ui/button"
import {
  Field,
  FieldGroup,
  FieldLabel,
} from "@/components/ui/field"
import { Input } from "@/components/ui/input"

function detectWxWorkEnvironment() {
  if (typeof navigator === "undefined") {
    return false
  }
  const userAgent = navigator.userAgent.toLowerCase()
  return userAgent.includes("wxwork")
}

export function LoginForm({
  className,
  ...props
}: React.ComponentProps<"form">) {
  const router = useRouter()
  const searchParams = useSearchParams()
  const { session } = useAuth()
  const [isPending, setIsPending] = useState(false)
  const [isWxWorkEnv, setIsWxWorkEnv] = useState(false)
  const nextPath = searchParams.get("next")
  const wxworkError = searchParams.get("wxworkError")
  const redirectPath =
    nextPath && nextPath.startsWith("/") ? nextPath : "/dashboard"

  useEffect(() => {
    if (session) {
      router.replace(redirectPath)
    }
  }, [redirectPath, router, session])

  useEffect(() => {
    if (wxworkError) {
      toast.error(wxworkError)
    }
  }, [wxworkError])

  useEffect(() => {
    setIsWxWorkEnv(detectWxWorkEnvironment())
  }, [])

  async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()
    const formData = new FormData(event.currentTarget)
    const username = formData.get("username")?.toString().trim() ?? ""
    const password = formData.get("password")?.toString() ?? ""

    setIsPending(true)

    try {
      await loginWithPassword({ username, password })
      toast.success("登录成功，正在进入系统")
      startTransition(() => {
        router.push(redirectPath)
      })
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "登录失败")
    } finally {
      setIsPending(false)
    }
  }

  return (
    <form
      className={cn("flex flex-col gap-6", className)}
      onSubmit={handleSubmit}
      {...props}
    >
      <FieldGroup>
        <div className="flex flex-col gap-2 text-center">
          <span className="mx-auto inline-flex rounded-full border border-amber-300/60 bg-amber-50 px-3 py-1 text-[11px] font-medium tracking-[0.22em] text-amber-900 uppercase">
            贝壳AI
          </span>
          <h1 className="text-3xl font-semibold tracking-tight">欢迎使用贝壳 AI 客服平台</h1>
        </div>
        <Field>
          <FieldLabel htmlFor="username">用户名</FieldLabel>
          <Input
            id="username"
            name="username"
            placeholder="admin"
            autoComplete="username"
            required
          />
        </Field>
        <Field>
          <div className="flex items-center">
            <FieldLabel htmlFor="password">密码</FieldLabel>
            {/* <span className="ml-auto text-xs text-muted-foreground">
              演示环境接受任意非空密码
            </span> */}
          </div>
          <Input
            id="password"
            name="password"
            type="password"
            autoComplete="current-password"
            required
          />
        </Field>
        <Field>
          <Button type="submit" disabled={isPending}>
            {isPending ? "登录中..." : "登录"}
          </Button>
        </Field>
        <Field>
          <Button
            type="button"
            variant="outline"
            className="gap-2"
            onClick={() => {
              const path = isWxWorkEnv ? "/api/auth/wxwork_login" : "/api/auth/wxwork_qr_login"
              window.location.href = `${path}?next=${encodeURIComponent(redirectPath)}`
            }}
          >
            <Image src="/images/wxwork.svg" alt="" width={16} height={16} className="size-4 shrink-0" />
            企业微信登录
          </Button>
        </Field>
      </FieldGroup>
    </form>
  )
}
