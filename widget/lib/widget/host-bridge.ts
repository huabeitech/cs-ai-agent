import { setWidgetConfig, type WidgetHostConfig } from "@/lib/widget/config";

const INIT_MESSAGE_TYPE = "cs-agent:init";
const OPEN_MESSAGE_TYPE = "cs-agent:open";
const MINIMIZE_MESSAGE_TYPE = "cs-agent:minimize";
const MAXIMIZED_MESSAGE_TYPE = "cs-agent:maximized";
const READY_MESSAGE_TYPE = "cs-agent:ready";
const REQUEST_MINIMIZE_MESSAGE_TYPE = "cs-agent:request-minimize";
const REQUEST_CLOSE_MESSAGE_TYPE = "cs-agent:request-close";
const REQUEST_TOGGLE_MAXIMIZE_MESSAGE_TYPE = "cs-agent:request-toggle-maximize";

type HostBridgeOptions = {
  onOpen?: () => void;
  onMinimize?: () => void;
  onMaximizedChange?: (isMaximized: boolean) => void;
};

export function bindHostBridge(options: HostBridgeOptions = {}) {
  if (typeof window === "undefined") {
    return () => undefined;
  }

  if (window.parent && window.parent !== window) {
    window.parent.postMessage({ type: READY_MESSAGE_TYPE }, "*");
  }

  const handleMessage = (
    event: MessageEvent,
  ) => {
    const data = event.data as
      | {
          type?: string;
          payload?: WidgetHostConfig | { isMaximized?: boolean };
        }
      | undefined;
    if (!data?.type) {
      return;
    }

    if (data.type === INIT_MESSAGE_TYPE && data.payload) {
      setWidgetConfig(data.payload as WidgetHostConfig);
      return;
    }

    if (data.type === OPEN_MESSAGE_TYPE) {
      options.onOpen?.();
      return;
    }

    if (data.type === MINIMIZE_MESSAGE_TYPE) {
      options.onMinimize?.();
      return;
    }

    if (data.type === MAXIMIZED_MESSAGE_TYPE) {
      options.onMaximizedChange?.(
        Boolean((data.payload as { isMaximized?: boolean } | undefined)?.isMaximized),
      );
    }
  };

  window.addEventListener("message", handleMessage);
  return () => window.removeEventListener("message", handleMessage);
}

function postToParent(type: string) {
  if (typeof window === "undefined") {
    return;
  }
  if (window.parent && window.parent !== window) {
    window.parent.postMessage({ type }, "*");
  }
}

export function requestHostMinimize() {
  postToParent(REQUEST_MINIMIZE_MESSAGE_TYPE);
}

export function requestHostClose() {
  postToParent(REQUEST_CLOSE_MESSAGE_TYPE);
}

export function requestHostToggleMaximize() {
  postToParent(REQUEST_TOGGLE_MAXIMIZE_MESSAGE_TYPE);
}
