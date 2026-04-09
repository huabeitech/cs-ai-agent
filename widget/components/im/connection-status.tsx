"use client";

import { cn } from "@/lib/utils";

type ConnectionStatusProps = {
  status: "connecting" | "connected" | "disconnected";
};

const statusText: Record<ConnectionStatusProps["status"], string> = {
  connecting: "连接中",
  connected: "在线服务",
  disconnected: "连接已断开",
};

export function ConnectionStatus({ status }: ConnectionStatusProps) {
  const toneClass =
    status === "connected"
      ? "border-emerald-200/80 bg-emerald-50 text-emerald-700"
      : status === "connecting"
        ? "border-amber-200/80 bg-amber-50 text-amber-700"
        : "border-slate-200/80 bg-slate-100 text-slate-600";

  return (
    <div
      className={cn(
        "inline-flex items-center gap-2 rounded-full border px-2.5 py-1 text-[11px] font-medium tracking-[0.02em] shadow-[0_6px_16px_rgba(15,23,42,0.06)]",
        toneClass,
      )}
    >
      <span
        className={cn(
          "cs-agent-status-dot inline-block size-2 rounded-full",
          status === "connected"
            ? "bg-emerald-500 shadow-[0_0_0_4px_rgba(16,185,129,0.14)]"
            : status === "connecting"
              ? "bg-amber-500 shadow-[0_0_0_4px_rgba(245,158,11,0.16)]"
              : "bg-slate-400 shadow-[0_0_0_4px_rgba(148,163,184,0.14)]",
        )}
      />
      <span>{statusText[status]}</span>
    </div>
  );
}
