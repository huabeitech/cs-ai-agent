(function () {
  if (window.__CS_AGENT_WIDGET_LOADED__) {
    return;
  }
  window.__CS_AGENT_WIDGET_LOADED__ = true;

  var config = window.CSAgentConfig || {};
  var baseUrl = String(config.baseUrl || "").replace(/\/$/, "");
  if (!config.appId || !baseUrl) {
    console.error("[cs-agent-widget] appId and baseUrl are required");
    return;
  }

  function resolveWidgetBaseUrl() {
    var currentScript = document.currentScript;
    if (currentScript && currentScript.src) {
      return currentScript.src.replace(/\/sdk\/cs-agent-widget\.js(?:\?.*)?$/, "");
    }
    if (/\/widget$/.test(baseUrl)) {
      return baseUrl;
    }
    return baseUrl + "/widget";
  }

  var button = document.createElement("button");
  button.type = "button";
  button.dataset.csAgentWidget = "launcher";
  button.setAttribute("aria-label", config.title || "在线客服");
  button.textContent = config.title || "在线客服";
  button.style.position = "fixed";
  button.style.bottom = "24px";
  button.style.right = config.position === "left" ? "" : "24px";
  button.style.left = config.position === "left" ? "24px" : "";
  button.style.zIndex = "2147483000";
  button.style.border = "0";
  button.style.borderRadius = "999px";
  button.style.padding = "14px 18px";
  button.style.background = config.themeColor || "#0f6cbd";
  button.style.color = "#fff";
  button.style.font = "600 14px/1 sans-serif";
  button.style.boxShadow = "0 18px 40px rgba(15, 35, 65, 0.24)";
  button.style.cursor = "pointer";

  var widgetBaseUrl = resolveWidgetBaseUrl();
  var frameUrl = new URL(widgetBaseUrl + "/frame/");
  frameUrl.searchParams.set("appId", config.appId);
  frameUrl.searchParams.set("baseUrl", baseUrl);
  if (config.apiBaseUrl) frameUrl.searchParams.set("apiBaseUrl", config.apiBaseUrl);
  if (config.title) frameUrl.searchParams.set("title", config.title);
  if (config.position) frameUrl.searchParams.set("position", config.position);
  if (config.themeColor) frameUrl.searchParams.set("themeColor", config.themeColor);
  if (config.width) frameUrl.searchParams.set("width", config.width);
  var frame = null;
  var frameLoaded = false;
  var frameReady = false;
  var initSent = false;
  var isOpen = false;
  var isMaximized = false;
  var frameHideTimer = null;
  var frameDestroyTimer = null;
  var animationDuration = 260;

  function clearFrameTimers() {
    if (frameHideTimer) {
      window.clearTimeout(frameHideTimer);
      frameHideTimer = null;
    }
    if (frameDestroyTimer) {
      window.clearTimeout(frameDestroyTimer);
      frameDestroyTimer = null;
    }
  }

  function applyFrameLayout() {
    if (!frame) {
      return;
    }

    frame.style.position = "fixed";
    frame.style.border = "0";
    frame.style.overflow = "hidden";
    frame.style.background = "#fff";
    frame.style.zIndex = "2147483000";
    frame.style.boxShadow = "0 28px 80px rgba(15, 35, 65, 0.28)";
    frame.style.willChange = "top,right,bottom,left,width,height,opacity,transform,border-radius";
    frame.style.transition =
      "top 260ms cubic-bezier(0.22, 1, 0.36, 1), right 260ms cubic-bezier(0.22, 1, 0.36, 1), bottom 260ms cubic-bezier(0.22, 1, 0.36, 1), left 260ms cubic-bezier(0.22, 1, 0.36, 1), width 260ms cubic-bezier(0.22, 1, 0.36, 1), height 260ms cubic-bezier(0.22, 1, 0.36, 1), opacity 220ms ease, transform 260ms cubic-bezier(0.22, 1, 0.36, 1), border-radius 260ms cubic-bezier(0.22, 1, 0.36, 1), box-shadow 260ms ease";
    frame.style.transformOrigin =
      config.position === "left" ? "left bottom" : "right bottom";

    if (isMaximized) {
      frame.style.top = "20px";
      frame.style.right = "20px";
      frame.style.bottom = "20px";
      frame.style.left = "20px";
      frame.style.width = "calc(100vw - 40px)";
      frame.style.maxWidth = "none";
      frame.style.height = "calc(100vh - 40px)";
      frame.style.borderRadius = "24px";
      return;
    }

    frame.style.top = "";
    frame.style.bottom = "88px";
    frame.style.right = config.position === "left" ? "" : "24px";
    frame.style.left = config.position === "left" ? "24px" : "";
    frame.style.width = config.width || "380px";
    frame.style.maxWidth = "calc(100vw - 24px)";
    frame.style.height = "min(760px, calc(100vh - 112px))";
    frame.style.borderRadius = "28px";
  }

  function flushFrameState() {
    if (!frame || !frameLoaded || !frameReady) {
      return;
    }

    if (!initSent) {
      initSent = true;
      postToFrame({
        type: "cs-agent:init",
        payload: config,
      });
    }

    postToFrame({ type: isOpen ? "cs-agent:open" : "cs-agent:minimize" });
    postToFrame({
      type: "cs-agent:maximized",
      payload: { isMaximized: isMaximized },
    });
  }

  function postToFrame(message) {
    if (!frame || !frame.contentWindow) {
      return;
    }
    try {
      frame.contentWindow.postMessage(message, frameUrl.origin);
    } catch (error) {
      console.error("[cs-agent-widget] postMessage failed", error);
    }
  }

  function syncFrameVisibility() {
    if (!frame) {
      return;
    }
    clearFrameTimers();
    applyFrameLayout();
    frame.style.display = "block";

    if (isOpen) {
      frame.style.visibility = "visible";
      frame.style.pointerEvents = "auto";
      frameHideTimer = window.setTimeout(function () {
        if (!frame) {
          return;
        }
        frame.style.opacity = "1";
        frame.style.transform = "translate3d(0, 0, 0) scale(1)";
      }, 16);
      flushFrameState();
      return;
    }

    frame.style.pointerEvents = "none";
    frame.style.opacity = "0";
    frame.style.transform = isMaximized
      ? "translate3d(0, 10px, 0) scale(0.985)"
      : "translate3d(0, 16px, 0) scale(0.96)";
    frameHideTimer = window.setTimeout(function () {
      if (!frame || isOpen) {
        return;
      }
      frame.style.visibility = "hidden";
    }, animationDuration);
    flushFrameState();
  }

  function handleWindowMessage(event) {
    if (!frame || event.source !== frame.contentWindow) {
      return;
    }

    var data = event.data || {};
    if (data.type === "cs-agent:ready") {
      frameReady = true;
      flushFrameState();
      return;
    }

    if (data.type === "cs-agent:request-minimize") {
      isOpen = false;
      syncFrameVisibility();
      return;
    }

    if (data.type === "cs-agent:request-close") {
      destroyFrame();
      return;
    }

    if (data.type === "cs-agent:request-toggle-maximize") {
      isMaximized = !isMaximized;
      syncFrameVisibility();
    }
  }

  function destroyFrame() {
    if (!frame) {
      return;
    }
    clearFrameTimers();
    frame.style.pointerEvents = "none";
    frame.style.opacity = "0";
    frame.style.transform = "translate3d(0, 18px, 0) scale(0.94)";
    frame.style.visibility = "hidden";
    frameDestroyTimer = window.setTimeout(function () {
      if (!frame) {
        return;
      }
      if (frame.parentNode) {
        frame.parentNode.removeChild(frame);
      }
      frame = null;
      frameLoaded = false;
      frameReady = false;
      initSent = false;
      isOpen = false;
      isMaximized = false;
      clearFrameTimers();
    }, animationDuration);
  }

  function createFrame() {
    if (frame) {
      return frame;
    }

    frame = document.createElement("iframe");
    frame.dataset.csAgentWidget = "frame";
    frame.title = config.title || "在线客服";
    frame.src = frameUrl.toString();
    applyFrameLayout();
    frame.style.display = "block";
    frame.style.visibility = "hidden";
    frame.style.pointerEvents = "none";
    frame.style.opacity = "0";
    frame.style.transform = "translate3d(0, 18px, 0) scale(0.96)";
    frame.addEventListener("load", function () {
      frameLoaded = true;
      syncFrameVisibility();
    });

    document.body.appendChild(frame);
    return frame;
  }

  button.addEventListener("click", function () {
    if (!frame) {
      createFrame();
    }
    isOpen = !isOpen;
    syncFrameVisibility();
  });

  window.addEventListener("message", handleWindowMessage);
  document.body.appendChild(button);
})();
