"use client"

import { useRouter, useSearchParams } from "next/navigation"
import { useEffect, useRef } from "react"
import { toast } from "sonner"

import { exchangeWxWorkTicket } from "@/lib/api/auth"

export default function WxWorkLoginCallbackPage() {
  const router = useRouter()
  const searchParams = useSearchParams()
  const ranRef = useRef(false)

  useEffect(() => {
    if (ranRef.current) {
      return
    }
    ranRef.current = true

    const ticket = searchParams.get("ticket")?.trim() ?? ""
    const next = searchParams.get("next")
    const nextPath = next && next.startsWith("/") ? next : "/dashboard"

    if (!ticket) {
      toast.error("企业微信登录票据不存在")
      router.replace("/login")
      return
    }

    void exchangeWxWorkTicket(ticket)
      .then(() => {
        toast.success("登录成功，正在进入系统")
        router.replace(nextPath)
      })
      .catch((error) => {
        toast.error(error instanceof Error ? error.message : "企业微信登录失败")
        router.replace("/login")
      })
  }, [router, searchParams])

  return (
    <div className="flex min-h-svh items-center justify-center bg-[linear-gradient(145deg,#fff7ed_0%,#ffffff_32%,#ecfeff_100%)] px-6">
      <div className="w-full max-w-md rounded-[28px] border border-white/70 bg-white/90 p-8 text-center shadow-[0_24px_80px_rgba(15,23,42,0.08)] backdrop-blur">
        <h1 className="text-2xl font-semibold tracking-tight">企业微信登录中</h1>
        <p className="mt-3 text-sm text-muted-foreground">
          正在校验登录票据并进入系统，请稍候。
        </p>
      </div>
    </div>
  )
}
