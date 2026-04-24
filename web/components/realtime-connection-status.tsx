"use client";

import { cn } from "@/lib/utils";

export type RealtimeConnectionStatusValue =
  | "connecting"
  | "connected"
  | "disconnected";

type RealtimeConnectionStatusProps = {
  status: RealtimeConnectionStatusValue;
  compact?: boolean;
};

const statusText: Record<RealtimeConnectionStatusValue, string> = {
  connecting: "实时通道连接中",
  connected: "实时通道已连接",
  disconnected: "实时通道已断开",
};

const compactStatusText: Record<RealtimeConnectionStatusValue, string> = {
  connecting: "平台实时：连接中",
  connected: "平台实时：在线",
  disconnected: "平台实时：已断开",
};

export function RealtimeConnectionStatus({
  status,
  compact = false,
}: RealtimeConnectionStatusProps) {
  const toneClass =
    status === "connected"
      ? "border-emerald-200/80 bg-emerald-50 text-emerald-700"
      : status === "connecting"
        ? "border-amber-200/80 bg-amber-50 text-amber-700"
        : "border-slate-200/80 bg-slate-100 text-slate-600";

  return (
    <div
      className={cn(
        compact
          ? "inline-flex items-center gap-1.5 rounded-full border px-2 py-0.5 text-[10px] font-medium tracking-[0.01em]"
          : "inline-flex items-center gap-2 rounded-full border px-2.5 py-1 text-[11px] font-medium tracking-[0.02em]",
        toneClass,
      )}
    >
      <span
        className={cn(
          "inline-block size-2 rounded-full",
          status === "connected"
            ? "bg-emerald-500 shadow-[0_0_0_4px_rgba(16,185,129,0.14)]"
            : status === "connecting"
              ? "bg-amber-500 shadow-[0_0_0_4px_rgba(245,158,11,0.16)]"
              : "bg-slate-400 shadow-[0_0_0_4px_rgba(148,163,184,0.14)]",
        )}
      />
      <span>
        {compact
          ? compactStatusText[status]
          : statusText[status]}
      </span>
    </div>
  );
}
