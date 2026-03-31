"use client";

import { memo, useCallback, useEffect, useLayoutEffect, useRef, useState } from "react";
import { toast } from "sonner";

import { ConversationTransferDialog } from "@/components/conversation-actions/transfer-dialog";
import { ImMessageEditor } from "@/components/im-message-editor";
import { ImMessageHTML } from "@/components/im-message-html";
import { useImageLightbox } from "@/components/image-lightbox";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  ResizableHandle,
  ResizablePanel,
  ResizablePanelGroup,
} from "@/components/ui/resizable";
import { useIsLgUp } from "@/hooks/use-lg-media";
import {
  assignAgentConversation,
  type AgentMessage,
} from "@/lib/api/agent";
import { readSession } from "@/lib/auth";
import { renderIMMessageHTML } from "@/lib/im-message";
import {
  agentConversationSelectors,
  useAgentConversationsStore,
  type AgentConversationFilterKey,
} from "@/lib/stores/agent-conversations";
import { formatDateTime } from "@/lib/utils";

const EMPTY_AGENT_MESSAGES: AgentMessage[] = [];

export function ChatPanel() {
  const conversation = useAgentConversationsStore(
    agentConversationSelectors.selectedConversation,
  );
  const messages =
    useAgentConversationsStore((state) => state.messages) ??
    EMPTY_AGENT_MESSAGES;
  const loading = useAgentConversationsStore((state) => state.messagesLoading);
  const sending = useAgentConversationsStore((state) => state.sending);
  const uploadingAsset = useAgentConversationsStore(
    (state) => state.uploadingAsset,
  );
  const sendMessage = useAgentConversationsStore((state) => state.sendMessage);
  const uploadImage = useAgentConversationsStore((state) => state.uploadImage);
  const sendAttachment = useAgentConversationsStore((state) => state.sendAttachment);
  const markSelectedConversationRead = useAgentConversationsStore(
    (state) => state.markSelectedConversationRead,
  );
  const loadConversations = useAgentConversationsStore((state) => state.loadConversations);
  const loadMessages = useAgentConversationsStore((state) => state.loadMessages);
  const loadOlderMessages = useAgentConversationsStore(
    (state) => state.loadOlderMessages,
  );
  const messagesHasMore = useAgentConversationsStore(
    (state) => state.messagesHasMore,
  );
  const messagesLoadingMore = useAgentConversationsStore(
    (state) => state.messagesLoadingMore,
  );
  const conversationFilter = useAgentConversationsStore((state) => state.conversationFilter);
  const setConversationFilter = useAgentConversationsStore(
    (state) => state.setConversationFilter,
  );
  const messagesContainerRef = useRef<HTMLDivElement>(null);
  const messagesContentRef = useRef<HTMLDivElement>(null);
  const scrollBottomRafRef = useRef<number | null>(null);
  const shouldStickToBottomRef = useRef(true);
  const prependScrollAnchorRef = useRef<{ height: number; top: number } | null>(
    null,
  );
  const [claiming, setClaiming] = useState(false);
  const [claimDialogOpen, setClaimDialogOpen] = useState(false);
  const [transferDialogOpen, setTransferDialogOpen] = useState(false);
  const isLgUp = useIsLgUp();
  const isClosedConversation = conversation?.status === 3;
  const isPendingConversation = conversation?.status === 1;
  const showMessageEditor = !isClosedConversation && !isPendingConversation;

  const switchToMyActiveIfNeeded = () => {
    if (conversationFilter !== "pending") {
      return;
    }
    setConversationFilter("active" satisfies AgentConversationFilterKey);
  };

  const getViewport = useCallback(
    () => messagesContainerRef.current,
    [],
  );

  const isNearBottom = useCallback(
    (element: HTMLElement, threshold = 80) =>
      element.scrollHeight - element.scrollTop - element.clientHeight <=
      threshold,
    [],
  );

  const scrollToBottom = useCallback(() => {
    const viewport = getViewport();
    if (!viewport) {
      return;
    }
    viewport.scrollTop = viewport.scrollHeight;
  }, [getViewport]);

  /**
   * 与 widget 消息列表一致：在单条调度链内多帧滚底直到 scrollHeight 稳定，
   * 避免多段滚底叠加导致滚动条抖动。
   */
  const scheduleScrollToBottom = useCallback(
    (attempts = 4) => {
      if (scrollBottomRafRef.current !== null) {
        cancelAnimationFrame(scrollBottomRafRef.current);
      }
      const run = (remaining: number, previousHeight = -1) => {
        scrollBottomRafRef.current = requestAnimationFrame(() => {
          const viewport = getViewport();
          if (!viewport) {
            scrollBottomRafRef.current = null;
            return;
          }
          const currentHeight = viewport.scrollHeight;
          scrollToBottom();
          if (remaining > 1 && currentHeight !== previousHeight) {
            run(remaining - 1, currentHeight);
            return;
          }
          scrollBottomRafRef.current = null;
        });
      };
      run(attempts);
    },
    [getViewport, scrollToBottom],
  );

  const handleImageSettled = useCallback(() => {
    if (!shouldStickToBottomRef.current) {
      return;
    }
    scheduleScrollToBottom();
  }, [scheduleScrollToBottom]);

  const maybeMarkConversationRead = useCallback(() => {
    const viewport = getViewport();
    if (!viewport || !conversation || loading) {
      return;
    }
    if (
      typeof document !== "undefined" &&
      document.visibilityState !== "visible"
    ) {
      return;
    }
    if (!isNearBottom(viewport)) {
      return;
    }
    void markSelectedConversationRead().catch((error) => {
      toast.error(error instanceof Error ? error.message : "设置已读失败");
    });
  }, [
    conversation,
    getViewport,
    isNearBottom,
    loading,
    markSelectedConversationRead,
  ]);

  useEffect(() => {
    const viewport = getViewport();
    if (!viewport) {
      return;
    }

    const handleScroll = () => {
      shouldStickToBottomRef.current = isNearBottom(viewport);
      if (shouldStickToBottomRef.current) {
        maybeMarkConversationRead();
      }
    };

    handleScroll();
    viewport.addEventListener("scroll", handleScroll);
    return () => {
      viewport.removeEventListener("scroll", handleScroll);
    };
  }, [conversation?.id, getViewport, isNearBottom, maybeMarkConversationRead]);

  useLayoutEffect(() => {
    shouldStickToBottomRef.current = true;
    scheduleScrollToBottom();
    return () => {
      if (scrollBottomRafRef.current !== null) {
        cancelAnimationFrame(scrollBottomRafRef.current);
        scrollBottomRafRef.current = null;
      }
    };
  }, [conversation?.id, scheduleScrollToBottom]);

  useLayoutEffect(() => {
    const viewport = getViewport();
    if (!viewport) {
      return;
    }
    const anchor = prependScrollAnchorRef.current;
    if (anchor) {
      prependScrollAnchorRef.current = null;
      const nextHeight = viewport.scrollHeight;
      viewport.scrollTop = nextHeight - anchor.height + anchor.top;
      return;
    }
    if (shouldStickToBottomRef.current) {
      scheduleScrollToBottom();
    }
  }, [messages, getViewport, scheduleScrollToBottom]);

  useEffect(() => {
    const content = messagesContentRef.current;
    if (!content) {
      return;
    }

    const observer = new ResizeObserver(() => {
      if (!shouldStickToBottomRef.current) {
        return;
      }
      scheduleScrollToBottom();
    });

    observer.observe(content);
    return () => {
      observer.disconnect();
    };
  }, [conversation?.id, scheduleScrollToBottom]);

  useEffect(() => {
    maybeMarkConversationRead();
  }, [maybeMarkConversationRead, messages.length]);

  useEffect(() => {
    const handleVisibilityChange = () => {
      if (document.visibilityState === "visible") {
        maybeMarkConversationRead();
      }
    };
    const handleFocus = () => {
      maybeMarkConversationRead();
    };

    document.addEventListener("visibilitychange", handleVisibilityChange);
    window.addEventListener("focus", handleFocus);
    return () => {
      document.removeEventListener("visibilitychange", handleVisibilityChange);
      window.removeEventListener("focus", handleFocus);
    };
  }, [maybeMarkConversationRead]);

  const handleLoadOlder = async () => {
    const viewport = getViewport();
    if (!viewport || messagesLoadingMore || !messagesHasMore) {
      return;
    }
    prependScrollAnchorRef.current = {
      height: viewport.scrollHeight,
      top: viewport.scrollTop,
    };
    try {
      await loadOlderMessages();
    } catch (error) {
      prependScrollAnchorRef.current = null;
      toast.error(error instanceof Error ? error.message : "加载历史消息失败");
    }
  };

  const handleSend = async (html: string) => {
    if (!conversation || sending || isClosedConversation) return;
    try {
      shouldStickToBottomRef.current = true;
      await sendMessage(html);
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "发送消息失败");
    }
  };

  const handleClaim = async () => {
    if (!conversation || claiming) return;
    const session = readSession();
    if (!session?.user?.id) {
      toast.error("未登录或登录已过期");
      return;
    }

    setClaiming(true);
    try {
      await assignAgentConversation(
        conversation.id,
        session.user.id,
        "认领会话",
      );

      switchToMyActiveIfNeeded();
      setClaimDialogOpen(false);
      toast.success("认领成功");
      await reloadConversationData(conversation.id);
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "认领会话失败");
    } finally {
      setClaiming(false);
    }
  };

  const reloadConversationData = async (conversationId: number) => {
    await loadConversations();
    await loadMessages(conversationId, { forceLoading: true, reset: true });
  };

  if (!conversation) {
    return (
      <div className="mt-10 flex flex-1 items-center justify-center px-4">
        <div className="text-center text-muted-foreground">
          <p className="text-lg">暂无会话</p>
          <p className="mt-1 text-sm lg:hidden">点击左上角菜单打开列表并选择会话</p>
          <p className="mt-1 hidden text-sm lg:block">请从左侧选择会话开始聊天</p>
        </div>
      </div>
    );
  }

  const messagesScroll = (
    <div
      ref={messagesContainerRef}
      className="h-full min-h-0 flex-1 overflow-y-auto p-4 cs-agent-scrollbar"
    >
      <div ref={messagesContentRef} className="flex flex-col">
        {!loading && messages.length > 0 && messagesHasMore ? (
          <div className="mb-4 flex justify-center">
            <Button
              type="button"
              variant="outline"
              size="sm"
              disabled={messagesLoadingMore}
              onClick={() => void handleLoadOlder()}
            >
              {messagesLoadingMore ? "加载中…" : "加载更早的消息"}
            </Button>
          </div>
        ) : null}
        {loading ? (
          <div className="py-8 text-center text-sm text-muted-foreground">
            加载中...
          </div>
        ) : messages.length > 0 ? (
          messages.map((message) => (
            <MessageItem
              key={message.id}
              message={message}
              onImageSettled={handleImageSettled}
            />
          ))
        ) : (
          <div className="py-8 text-center text-sm text-muted-foreground">
            暂无消息
          </div>
        )}
      </div>
    </div>
  );

  const bottomPanel = (
    <div
      className={
        showMessageEditor
          ? "h-full overflow-auto border-t border-border bg-background"
          : "shrink-0 overflow-auto border-t border-border bg-background"
      }
    >
      {isClosedConversation ? (
        <div className="bg-amber-500/10 px-4 py-3 text-sm text-amber-950 dark:bg-amber-500/15 dark:text-amber-100">
          当前会话已关闭
        </div>
      ) : isPendingConversation ? (
        <div className="bg-blue-500/10 px-4 py-3 dark:bg-blue-500/15">
          <div className="flex items-center gap-2">
            <Button
              onClick={() => setClaimDialogOpen(true)}
              disabled={claiming}
              size="sm"
            >
              {claiming ? "认领中..." : "认领"}
            </Button>
          </div>
        </div>
      ) : (
        <div className="flex h-full min-h-0 flex-col">
          <div className="min-h-0 flex-1">
            <ImMessageEditor
              disabled={!conversation || sending}
              uploadingAsset={uploadingAsset}
              onSend={handleSend}
              onUploadImage={async (file) => {
                shouldStickToBottomRef.current = true;
                const uploaded = await uploadImage(file);
                return uploaded
                  ? { url: uploaded.url, filename: uploaded.filename }
                  : null;
              }}
              onSendAttachment={async (file) => {
                shouldStickToBottomRef.current = true;
                try {
                  await sendAttachment(file);
                } catch (error) {
                  toast.error(error instanceof Error ? error.message : "发送附件失败");
                }
              }}
            />
          </div>
        </div>
      )}
    </div>
  );

  return (
    <div className="flex h-full min-h-0 flex-1 flex-col overflow-hidden">
      {showMessageEditor ? (
        isLgUp ? (
          <ResizablePanelGroup
            orientation="vertical"
            className="flex min-h-0 flex-1 flex-col"
          >
            <ResizablePanel defaultSize="72%" minSize="35%" className="min-h-0">
              {messagesScroll}
            </ResizablePanel>
            <ResizableHandle withHandle />
            <ResizablePanel defaultSize="28%" minSize="18%" maxSize="55%" className="min-h-0">
              {bottomPanel}
            </ResizablePanel>
          </ResizablePanelGroup>
        ) : (
          <div className="flex min-h-0 flex-1 flex-col overflow-hidden">
            <div className="min-h-0 flex-1">{messagesScroll}</div>
            <div className="shrink-0 pb-[env(safe-area-inset-bottom)]">{bottomPanel}</div>
          </div>
        )
      ) : (
        <>
          <div className="min-h-0 flex-1">{messagesScroll}</div>
          <div className="shrink-0 pb-[env(safe-area-inset-bottom)] lg:pb-0">{bottomPanel}</div>
        </>
      )}
      <Dialog
        open={claimDialogOpen}
        onOpenChange={(open) => {
          if (claiming) {
            return;
          }
          setClaimDialogOpen(open);
        }}
      >
        <DialogContent className="max-w-md" showCloseButton={false}>
          <DialogHeader>
            <DialogTitle>确认认领会话</DialogTitle>
            <DialogDescription>
              {conversation
                ? `确认认领“${conversation.subject}”吗？认领后会话会进入我的列表。`
                : "确认认领当前会话吗？"}
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              disabled={claiming}
              onClick={() => setClaimDialogOpen(false)}
            >
              取消
            </Button>
            <Button
              type="button"
              disabled={claiming}
              onClick={() => void handleClaim()}
            >
              {claiming ? "认领中..." : "确认认领"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
      <ConversationTransferDialog
        open={transferDialogOpen}
        mode="transfer"
        conversationId={conversation.id}
        onOpenChange={setTransferDialogOpen}
        onSuccess={async () => {
          await reloadConversationData(conversation.id);
        }}
      />
    </div>
  );
}

type MessageItemProps = {
  message: AgentMessage;
  onImageSettled: () => void;
};

const MessageItem = memo(
  function MessageItem({ message, onImageSettled }: MessageItemProps) {
    const { open: openImageLightbox } = useImageLightbox();
    const isCustomer = message.senderType === "customer";
    const isAi = message.senderType === "ai";
    const isAgentSide = message.senderType === "agent" || isAi;
    const senderName = isCustomer
      ? message.senderName || "客户"
      : isAi
        ? "AI"
        : message.senderName || "客服";
    const agentAvatarSrc =
      isAgentSide && !isAi && message.senderAvatar?.trim()
        ? message.senderAvatar.trim()
        : undefined;
    const avatarFallback = isAi ? "AI" : senderName.charAt(0);
    const htmlContent = buildMessageHTML(message);
    const bubbleClassName = isAi
      ? "border border-primary/15 bg-primary/5 text-foreground shadow-sm"
      : isAgentSide
        ? "bg-emerald-600 text-white shadow-sm"
        : "border border-border/70 bg-muted/60 text-foreground shadow-sm";
    const htmlClassName = isAi
      ? "[&_a]:text-foreground [&_a]:underline [&_img]:rounded-md"
      : isAgentSide
        ? "[&_p]:text-white [&_a]:text-white [&_a]:underline [&_img]:rounded-md"
        : "[&_a]:text-foreground [&_a]:underline [&_img]:rounded-md";
    const avatarClassName = isAi
      ? "border border-primary/20 bg-primary/10 text-xs text-foreground"
      : isAgentSide
        ? "bg-emerald-600 text-xs text-white"
        : "border border-border/70 bg-muted/60 text-xs text-foreground";

    return (
      <div
        className={`mb-4 flex items-start gap-2 ${
          isAgentSide ? "justify-end" : "justify-start"
        }`}
      >
        {isAgentSide ? (
          <>
            <div className="flex max-w-[70%] flex-col items-end">
              <div className="mb-1 text-xs text-muted-foreground">
                {senderName}
              </div>
              <div className={`w-fit rounded-2xl px-3 py-2 text-left ${bubbleClassName}`}>
                <ImMessageHTML
                  html={htmlContent}
                  className={htmlClassName}
                  onImageSettled={onImageSettled}
                  onImageClick={openImageLightbox}
                />
              </div>
              <div className="mt-1 flex items-center gap-2 text-xs text-muted-foreground">
                <span>{formatDateTime(message.sentAt || "")}</span>
                {message.sendStatus === 2 && (
                  <span>{message.customerRead ? "客户已读" : "客户未读"}</span>
                )}
              </div>
            </div>
            <Avatar className="size-8 shrink-0">
              <AvatarImage src={agentAvatarSrc ?? ""} />
              <AvatarFallback className={avatarClassName}>
                {avatarFallback}
              </AvatarFallback>
            </Avatar>
          </>
        ) : (
          <>
            <Avatar className="size-8 shrink-0">
              <AvatarImage src="" />
              <AvatarFallback className={avatarClassName}>
                客
              </AvatarFallback>
            </Avatar>
            <div className="max-w-[70%]">
              <div className="mb-1 text-xs text-muted-foreground">
                {senderName}
              </div>
              <div className={`w-fit rounded-2xl px-3 py-2 ${bubbleClassName}`}>
                <ImMessageHTML
                  html={htmlContent}
                  className={htmlClassName}
                  onImageSettled={onImageSettled}
                  onImageClick={openImageLightbox}
                />
              </div>
              <div className="mt-1 flex items-center gap-2 text-xs text-muted-foreground">
                <span>{formatDateTime(message.sentAt || "")}</span>
              </div>
            </div>
          </>
        )}
      </div>
    );
  },
  (prevProps, nextProps) =>
    prevProps.message === nextProps.message &&
    prevProps.onImageSettled === nextProps.onImageSettled,
);

function buildMessageHTML(message: {
  messageType: string;
  content: string;
  payload?: string;
}) {
  return renderIMMessageHTML(message);
}
