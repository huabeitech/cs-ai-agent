"use client";

import { useSyncExternalStore } from "react";

/** 与 Tailwind 默认 `lg` 断点一致（1024px） */
const LG_MEDIA_QUERY = "(min-width: 1024px)";

function subscribe(onStoreChange: () => void) {
  const mq = window.matchMedia(LG_MEDIA_QUERY);
  mq.addEventListener("change", onStoreChange);
  return () => mq.removeEventListener("change", onStoreChange);
}

function getSnapshot() {
  return window.matchMedia(LG_MEDIA_QUERY).matches;
}

function getServerSnapshot() {
  return false;
}

/** 仅客户端在 lg 及以上为 true；SSR 恒为 false，与 mobile-first 一致 */
export function useIsLgUp() {
  return useSyncExternalStore(subscribe, getSnapshot, getServerSnapshot);
}
