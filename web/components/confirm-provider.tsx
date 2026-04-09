"use client"

import {
  createContext,
  useCallback,
  useContext,
  useRef,
  useState,
  type ReactNode,
} from "react"

import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"

type ConfirmOptions = {
  title?: ReactNode
  description?: ReactNode
  confirmText?: string
  cancelText?: string
  variant?: "default" | "destructive"
}

type ConfirmContextValue = {
  confirm: (options: ConfirmOptions) => Promise<boolean>
}

type ConfirmState = ConfirmOptions & {
  open: boolean
}

const ConfirmContext = createContext<ConfirmContextValue | null>(null)

const defaultState: ConfirmState = {
  open: false,
  title: "请确认操作",
  description: "确认后将继续执行当前操作。",
  confirmText: "确认",
  cancelText: "取消",
  variant: "default",
}

export function ConfirmProvider({ children }: { children: ReactNode }) {
  const [state, setState] = useState<ConfirmState>(defaultState)
  const resolverRef = useRef<((value: boolean) => void) | null>(null)

  const close = useCallback((result: boolean) => {
    resolverRef.current?.(result)
    resolverRef.current = null
    setState((current) => ({ ...current, open: false }))
  }, [])

  const confirm = useCallback((options: ConfirmOptions) => {
    if (resolverRef.current) {
      resolverRef.current(false)
    }

    setState({
      open: true,
      title: options.title ?? defaultState.title,
      description: options.description ?? defaultState.description,
      confirmText: options.confirmText ?? defaultState.confirmText,
      cancelText: options.cancelText ?? defaultState.cancelText,
      variant: options.variant ?? defaultState.variant,
    })

    return new Promise<boolean>((resolve) => {
      resolverRef.current = resolve
    })
  }, [])

  return (
    <ConfirmContext.Provider value={{ confirm }}>
      {children}
      <Dialog
        open={state.open}
        onOpenChange={(open) => {
          if (!open) {
            close(false)
          }
        }}
      >
        <DialogContent className="sm:max-w-md" showCloseButton>
          <DialogHeader>
            <DialogTitle>{state.title}</DialogTitle>
            <DialogDescription>{state.description}</DialogDescription>
          </DialogHeader>
          <DialogFooter className="p-2">
            <Button
              type="button"
              variant="outline"
              onClick={() => close(false)}
            >
              {state.cancelText}
            </Button>
            <Button
              type="button"
              variant={state.variant}
              onClick={() => close(true)}
            >
              {state.confirmText}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </ConfirmContext.Provider>
  )
}

export function useConfirm() {
  const ctx = useContext(ConfirmContext)
  if (!ctx) {
    throw new Error("useConfirm must be used within ConfirmProvider")
  }
  return ctx.confirm
}

export type { ConfirmOptions }
