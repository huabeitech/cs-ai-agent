(function () {
  var DEFAULT_CONFIG = {
    position: "right",
    themeColor: "#0f6cbd",
    width: "380px",
    externalSource: "web_chat",
  };

  var state = window.__CS_AGENT_WIDGET_STATE__;
  if (!state) {
    state = {
      button: null,
      frame: null,
      frameLoaded: false,
      frameReady: false,
      initSent: false,
      isOpen: false,
      isMaximized: false,
      frameHideTimer: null,
      frameDestroyTimer: null,
      config: null,
      frameUrl: null,
      animationDuration: 260,
    };
    window.__CS_AGENT_WIDGET_STATE__ = state;
  }

  function normalizeConfig(config) {
    var merged = {};
    var key;
    for (key in DEFAULT_CONFIG) {
      if (Object.prototype.hasOwnProperty.call(DEFAULT_CONFIG, key)) {
        merged[key] = DEFAULT_CONFIG[key];
      }
    }
    config = config || {};
    for (key in config) {
      if (Object.prototype.hasOwnProperty.call(config, key)) {
        merged[key] = config[key];
      }
    }
    merged.baseUrl = String(merged.baseUrl || window.location.origin).replace(/\/$/, "");
    merged.apiBaseUrl = String(merged.apiBaseUrl || merged.baseUrl).replace(/\/$/, "");
    merged.channelId = String(merged.channelId || "");
    merged.externalSource = String(merged.externalSource || "web_chat");
    return merged;
  }

  function resolveWidgetBaseUrl(config) {
    var currentScript = document.currentScript;
    if (currentScript && currentScript.src) {
      return currentScript.src.replace(/\/sdk\/cs-agent-widget\.js(?:\?.*)?$/, "");
    }
    return String(config.widgetBaseUrl || config.baseUrl || window.location.origin).replace(/\/$/, "");
  }

  function createFrameUrl(config) {
    var widgetBaseUrl = resolveWidgetBaseUrl(config);
    var frameUrl = new URL(widgetBaseUrl + "/kefu/chat/");
    frameUrl.searchParams.set("channelId", config.channelId);
    frameUrl.searchParams.set("baseUrl", config.baseUrl);
    if (config.apiBaseUrl) frameUrl.searchParams.set("apiBaseUrl", config.apiBaseUrl);
    if (config.externalSource) frameUrl.searchParams.set("externalSource", config.externalSource);
    if (config.title) frameUrl.searchParams.set("title", config.title);
    if (config.subtitle) frameUrl.searchParams.set("subtitle", config.subtitle);
    if (config.position) frameUrl.searchParams.set("position", config.position);
    if (config.themeColor) frameUrl.searchParams.set("themeColor", config.themeColor);
    if (config.width) frameUrl.searchParams.set("width", config.width);
    if (config.subject) frameUrl.searchParams.set("subject", config.subject);
    return frameUrl;
  }

  function clearFrameTimers() {
    if (state.frameHideTimer) {
      window.clearTimeout(state.frameHideTimer);
      state.frameHideTimer = null;
    }
    if (state.frameDestroyTimer) {
      window.clearTimeout(state.frameDestroyTimer);
      state.frameDestroyTimer = null;
    }
  }

  function applyFrameLayout() {
    var frame = state.frame;
    var config = state.config;
    if (!frame || !config) {
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

    if (state.isMaximized) {
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

  function postToFrame(message) {
    if (!state.frame || !state.frame.contentWindow || !state.frameUrl) {
      return;
    }
    try {
      state.frame.contentWindow.postMessage(message, state.frameUrl.origin);
    } catch (error) {
      console.error("[cs-agent-widget] postMessage failed", error);
    }
  }

  function flushFrameState() {
    if (!state.frame || !state.frameLoaded || !state.frameReady) {
      return;
    }

    if (!state.initSent) {
      state.initSent = true;
      postToFrame({
        type: "cs-agent:init",
        payload: state.config,
      });
    }

    postToFrame({ type: state.isOpen ? "cs-agent:open" : "cs-agent:minimize" });
    postToFrame({
      type: "cs-agent:maximized",
      payload: { isMaximized: state.isMaximized },
    });
  }

  function syncFrameVisibility() {
    var frame = state.frame;
    if (!frame) {
      return;
    }
    clearFrameTimers();
    applyFrameLayout();
    frame.style.display = "block";

    if (state.isOpen) {
      frame.style.visibility = "visible";
      frame.style.pointerEvents = "auto";
      state.frameHideTimer = window.setTimeout(function () {
        if (!state.frame) {
          return;
        }
        state.frame.style.opacity = "1";
        state.frame.style.transform = "translate3d(0, 0, 0) scale(1)";
      }, 16);
      flushFrameState();
      return;
    }

    frame.style.pointerEvents = "none";
    frame.style.opacity = "0";
    frame.style.transform = state.isMaximized
      ? "translate3d(0, 10px, 0) scale(0.985)"
      : "translate3d(0, 16px, 0) scale(0.96)";
    state.frameHideTimer = window.setTimeout(function () {
      if (!state.frame || state.isOpen) {
        return;
      }
      state.frame.style.visibility = "hidden";
    }, state.animationDuration);
    flushFrameState();
  }

  function destroyFrame() {
    if (!state.frame) {
      return;
    }
    clearFrameTimers();
    state.frame.style.pointerEvents = "none";
    state.frame.style.opacity = "0";
    state.frame.style.transform = "translate3d(0, 18px, 0) scale(0.94)";
    state.frame.style.visibility = "hidden";
    state.frameDestroyTimer = window.setTimeout(function () {
      if (!state.frame) {
        return;
      }
      if (state.frame.parentNode) {
        state.frame.parentNode.removeChild(state.frame);
      }
      state.frame = null;
      state.frameLoaded = false;
      state.frameReady = false;
      state.initSent = false;
      state.isOpen = false;
      state.isMaximized = false;
      clearFrameTimers();
    }, state.animationDuration);
  }

  function createFrame() {
    if (state.frame) {
      return state.frame;
    }

    state.frame = document.createElement("iframe");
    state.frame.dataset.csAgentWidget = "frame";
    state.frame.title = state.config.title || "在线客服";
    state.frame.src = state.frameUrl.toString();
    applyFrameLayout();
    state.frame.style.display = "block";
    state.frame.style.visibility = "hidden";
    state.frame.style.pointerEvents = "none";
    state.frame.style.opacity = "0";
    state.frame.style.transform = "translate3d(0, 18px, 0) scale(0.96)";
    state.frame.addEventListener("load", function () {
      state.frameLoaded = true;
      syncFrameVisibility();
    });

    document.body.appendChild(state.frame);
    return state.frame;
  }

  function handleWindowMessage(event) {
    if (!state.frame || event.source !== state.frame.contentWindow) {
      return;
    }

    var data = event.data || {};
    if (data.type === "cs-agent:ready") {
      state.frameReady = true;
      flushFrameState();
      return;
    }

    if (data.type === "cs-agent:request-minimize") {
      state.isOpen = false;
      syncFrameVisibility();
      return;
    }

    if (data.type === "cs-agent:request-close") {
      destroyFrame();
      return;
    }

    if (data.type === "cs-agent:request-toggle-maximize") {
      state.isMaximized = !state.isMaximized;
      syncFrameVisibility();
    }
  }

  function createLauncher() {
    if (state.button) {
      return state.button;
    }

    var config = state.config;
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

    button.addEventListener("click", function () {
      if (!state.frame) {
        createFrame();
      }
      state.isOpen = !state.isOpen;
      syncFrameVisibility();
    });

    document.body.appendChild(button);
    state.button = button;
    return button;
  }

  function mount(config) {
    state.config = normalizeConfig(config || window.CSAgentConfig || {});
    if (!state.config.channelId || !state.config.baseUrl) {
      console.error("[cs-agent-widget] channelId and baseUrl are required");
      return;
    }

    state.frameUrl = createFrameUrl(state.config);
    createLauncher();
  }

  function destroy() {
    clearFrameTimers();
    if (state.frame && state.frame.parentNode) {
      state.frame.parentNode.removeChild(state.frame);
    }
    if (state.button && state.button.parentNode) {
      state.button.parentNode.removeChild(state.button);
    }
    state.button = null;
    state.frame = null;
    state.frameLoaded = false;
    state.frameReady = false;
    state.initSent = false;
    state.isOpen = false;
    state.isMaximized = false;
  }

  window.CSAgentWidget = {
    mount: mount,
    destroy: destroy,
    open: function () {
      if (!state.frame) {
        createFrame();
      }
      state.isOpen = true;
      syncFrameVisibility();
    },
    close: function () {
      state.isOpen = false;
      syncFrameVisibility();
    },
  };

  if (!state.listenerBound) {
    window.addEventListener("message", handleWindowMessage);
    state.listenerBound = true;
  }

  if (window.CSAgentConfig) {
    mount(window.CSAgentConfig);
  }
})();

