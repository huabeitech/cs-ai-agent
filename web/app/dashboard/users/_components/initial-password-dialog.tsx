"use client"

import { useState } from "react"
import { CopyIcon } from "lucide-react"
import { toast } from "sonner"

import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"

type InitialPasswordDialogProps = {
  open: boolean
  username: string
  password: string
  onOpenChange: (open: boolean) => void
}

export function InitialPasswordDialog({
  open,
  username,
  password,
  onOpenChange,
}: InitialPasswordDialogProps) {
  const [copying, setCopying] = useState(false)

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
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>用户已创建</DialogTitle>
          <DialogDescription>
            {username || "-"} 的初始密码已生成，仅在此展示一次，请及时复制并安全传达。
          </DialogDescription>
        </DialogHeader>
        <div className="rounded-xl border bg-muted/40 p-4">
          <div className="text-xs text-muted-foreground">初始密码</div>
          <div className="mt-2 break-all font-mono text-base">{password}</div>
        </div>
        <DialogFooter>
          <Button
            type="button"
            variant="outline"
            onClick={() => void handleCopy()}
            disabled={copying || !password}
          >
            <CopyIcon />
            {copying ? "复制中..." : "复制密码"}
          </Button>
          <Button type="button" onClick={() => onOpenChange(false)}>
            关闭
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
