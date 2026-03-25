"use client"

import { useState } from "react"
import { CopyIcon } from "lucide-react"
import { toast } from "sonner"

import { type AdminUser } from "@/lib/api/admin"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"

type ResetPasswordDialogsProps = {
  open: boolean
  saving: boolean
  item: AdminUser | null
  password: string
  onOpenChange: (open: boolean) => void
  onConfirm: () => Promise<void>
}

export function ResetPasswordDialogs({
  open,
  saving,
  item,
  password,
  onOpenChange,
  onConfirm,
}: ResetPasswordDialogsProps) {
  const [copying, setCopying] = useState(false)
  const showingResult = password.trim().length > 0

  async function handleCopy() {
    if (!password || copying) {
      return
    }

    setCopying(true)
    try {
      await navigator.clipboard.writeText(password)
      toast.success("密码已复制")
    } catch {
      toast.error("复制失败，请手动复制")
    } finally {
      setCopying(false)
    }
  }

  return (
    <>
      <Dialog open={open && !showingResult} onOpenChange={onOpenChange}>
        <DialogContent showCloseButton={!saving}>
          <DialogHeader>
            <DialogTitle>确认重置密码</DialogTitle>
            <DialogDescription>
              确认后将为 {item?.username || "-"} 生成新的随机密码，并使该用户当前登录会话失效。
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button onClick={() => void onConfirm()} disabled={saving}>
              {saving ? "重置中..." : "确认重置"}
            </Button>
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
              disabled={saving}
            >
              取消
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
      <Dialog open={open && showingResult} onOpenChange={onOpenChange}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>重置密码成功</DialogTitle>
            <DialogDescription>
              {item?.username || "-"} 的新密码已生成，请及时复制并安全传达。
            </DialogDescription>
          </DialogHeader>
          <div className="rounded-xl border bg-muted/40 p-4">
            <div className="text-xs text-muted-foreground">新密码</div>
            <div className="mt-2 break-all font-mono text-base">{password}</div>
          </div>
          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => void handleCopy()} disabled={copying}>
              <CopyIcon />
              {copying ? "复制中..." : "复制密码"}
            </Button>
            <Button type="button" onClick={() => onOpenChange(false)}>
              关闭
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  )
}
