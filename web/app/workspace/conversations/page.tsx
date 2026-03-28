"use client";

import {
  ArrowRightLeftIcon,
  ChevronLeft,
  ChevronRight,
  CircleUserRoundIcon,
  CircleXIcon,
  Menu,
  MoreHorizontalIcon,
  X,
} from "lucide-react";
import { useEffect, useRef, useState } from "react";
import { toast } from "sonner";
import type { PanelImperativeHandle } from "react-resizable-panels";

import { ConversationCloseDialog } from "@/components/conversation-actions/close-dialog";
import { ConversationTransferDialog } from "@/components/conversation-actions/transfer-dialog";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  ResizableHandle,
  ResizablePanel,
  ResizablePanelGroup,
} from "@/components/ui/resizable";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet";
import { useAgentConversationRealtime } from "@/hooks/use-agent-conversation-realtime";
import {
  agentConversationFilterOptions,
  agentConversationSelectors,
  type AgentConversationFilterKey,
  useAgentConversationsStore,
} from "@/lib/stores/agent-conversations";
import { ChatPanel } from "./_components/chat-panel";
import { CustomerInfoPanel } from "./_components/customer-info-panel";
import { ConversationList } from "./_components/conversation-list";

export default function ConversationsPage() {
  const conversation = useAgentConversationsStore(
    agentConversationSelectors.selectedConversation,
  );
  const conversationFilter = useAgentConversationsStore(
    (state) => state.conversationFilter,
  );
  const setConversationFilter = useAgentConversationsStore(
    (state) => state.setConversationFilter,
  );
  const loadConversations = useAgentConversationsStore(
    (state) => state.loadConversations,
  );
  const loadMessages = useAgentConversationsStore(
    (state) => state.loadMessages,
  );
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false);
  const [infoPanelCollapsed, setInfoPanelCollapsed] = useState(false);
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);
  const [mobileCustomerSheetOpen, setMobileCustomerSheetOpen] = useState(false);
  const [transferOpen, setTransferOpen] = useState(false);
  const [closeOpen, setCloseOpen] = useState(false);
  const sidebarPanelRef = useRef<PanelImperativeHandle | null>(null);
  const infoPanelRef = useRef<PanelImperativeHandle | null>(null);
  useEffect(() => {
    void loadConversations().catch((error) => {
      toast.error(error instanceof Error ? error.message : "加载会话列表失败");
    });
  }, [loadConversations, conversationFilter]);

  async function handleConversationChanged(conversationId: number) {
    await loadConversations();
    await loadMessages(conversationId, {
      forceLoading: false,
      reset: false,
    });
  }

  useAgentConversationRealtime();

  const handleSidebarToggle = () => {
    const panel = sidebarPanelRef.current;
    if (!panel) {
      setSidebarCollapsed((current) => !current);
      return;
    }

    if (panel.isCollapsed()) {
      panel.expand();
      setSidebarCollapsed(false);
      return;
    }

    panel.collapse();
    setSidebarCollapsed(true);
  };

  const handleInfoPanelToggle = () => {
    const panel = infoPanelRef.current;
    if (!panel) {
      setInfoPanelCollapsed((current) => !current);
      return;
    }

    if (panel.isCollapsed()) {
      panel.expand();
      setInfoPanelCollapsed(false);
      return;
    }

    panel.collapse();
    setInfoPanelCollapsed(true);
  };

  const renderConversationSidebar = (opts?: { onListAfterSelect?: () => void }) => (
    <div className="flex h-full min-h-0 flex-1 flex-col">
      <div className="flex shrink-0 items-start justify-between gap-2 border-b p-2 h-12.5">
        <Tabs
          value={conversationFilter}
          onValueChange={(value) =>
            setConversationFilter(value as AgentConversationFilterKey)
          }
          className="min-w-0 flex-1 gap-0"
        >
          <TabsList
            className="w-full min-w-0 justify-start "
          >
            {agentConversationFilterOptions.map((opt) => (
              <TabsTrigger
                key={opt.value}
                value={opt.value}
                className="shrink-0 px-2.5 text-xs sm:text-sm"
              >
                {opt.label}
              </TabsTrigger>
            ))}
          </TabsList>
        </Tabs>
        <Button
          variant="ghost"
          size="icon"
          className="mt-0.5 shrink-0 lg:hidden"
          onClick={() => setMobileMenuOpen(false)}
        >
          <X className="size-4" />
        </Button>
      </div>
      <ConversationList onAfterSelect={opts?.onListAfterSelect} />
    </div>
  );

  const workspaceContent = (
    <div className="flex h-full min-h-0 w-full flex-1 flex-col overflow-hidden">
      <div className="shrink-0 flex items-center justify-between gap-3 border-b px-3 py-1 h-12.5">
        <div className="flex min-w-0 items-center gap-2 sm:gap-3">
          <Button
            variant="ghost"
            size="icon"
            className="lg:hidden"
            onClick={() => setMobileMenuOpen(true)}
          >
            <Menu className="size-4" />
          </Button>
          <Button
            variant="ghost"
            size="icon"
            className="hidden lg:flex"
            onClick={handleSidebarToggle}
          >
            {sidebarCollapsed ? (
              <ChevronRight className="size-4" />
            ) : (
              <ChevronLeft className="size-4" />
            )}
          </Button>
          {conversation ? (
            <>
              <Avatar className="size-8 shrink-0 lg:size-9">
                <AvatarImage src="" />
                <AvatarFallback>客</AvatarFallback>
              </Avatar>
              <div className="min-w-0">
                <p className="truncate font-medium leading-tight">
                  {conversation.subject}
                </p>
                <p className="mt-0.5 truncate text-xs text-muted-foreground sm:text-sm">
                  <span>{conversation.externalSource}</span>
                  <span className="text-muted-foreground/60"> / </span>
                  <span>{conversation.externalId}</span>
                </p>
              </div>
            </>
          ) : (
            <div className="min-w-0">
              <p className="truncate font-medium leading-tight">会话工作台</p>
              <p className="mt-0.5 truncate text-xs text-muted-foreground sm:text-sm lg:hidden">
                打开菜单选择会话
              </p>
              <p className="mt-0.5 hidden truncate text-sm text-muted-foreground lg:block">
                请选择左侧会话开始处理消息
              </p>
            </div>
          )}
        </div>
        <div className="flex shrink-0 items-center gap-0.5 sm:gap-1">
          <Button
            variant="ghost"
            size="icon"
            className="lg:hidden"
            disabled={!conversation}
            aria-label="客户信息"
            onClick={() => setMobileCustomerSheetOpen(true)}
          >
            <CircleUserRoundIcon className="size-4" />
          </Button>
          <DropdownMenu>
            <DropdownMenuTrigger
              render={
                <Button variant="ghost" size="icon" disabled={!conversation} />
              }
            >
              <MoreHorizontalIcon className="size-4" />
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" className="w-44 min-w-44">
              <DropdownMenuItem
                onClick={() => setTransferOpen(true)}
                disabled={!conversation || conversation.status !== 2}
              >
                <ArrowRightLeftIcon />
                转接会话
              </DropdownMenuItem>
              <DropdownMenuItem
                onClick={() => setCloseOpen(true)}
                disabled={!conversation || conversation.status === 3}
              >
                <CircleXIcon />
                关闭会话
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
          <Button
            variant="ghost"
            size="icon"
            className="hidden lg:flex"
            onClick={handleInfoPanelToggle}
            aria-label={infoPanelCollapsed ? "展开客户信息" : "收起客户信息"}
          >
            {infoPanelCollapsed ? (
              <ChevronLeft className="size-4" />
            ) : (
              <ChevronRight className="size-4" />
            )}
          </Button>
        </div>
      </div>
      <div className="flex min-h-0 w-full flex-1 overflow-hidden">
        <ChatPanel />
      </div>
    </div>
  );

  return (
    <div className="flex h-full min-h-0 w-full min-w-0 flex-col overflow-hidden">
      {/* H5 无左侧导航：顶栏 h-12、left-0；lg 起有 w-14 侧栏，与 layout 一致 */}
      {mobileMenuOpen && (
        <button
          type="button"
          aria-label="关闭会话列表"
          className="fixed top-12 right-0 bottom-0 left-0 z-30 bg-black/50 lg:hidden"
          onClick={() => setMobileMenuOpen(false)}
        />
      )}
      <div
        className={`fixed top-12 bottom-0 left-0 z-40 flex w-[min(22rem,calc(100vw-0.75rem))] max-w-[min(22rem,calc(100vw-0.75rem))] flex-col overflow-hidden border-r bg-background shadow-lg transition-transform duration-300 ease-out will-change-transform touch-manipulation overscroll-contain supports-[padding:max(0px)]:pb-[env(safe-area-inset-bottom)] lg:hidden ${
          mobileMenuOpen ? "translate-x-0" : "-translate-x-full pointer-events-none"
        }`}
        aria-hidden={!mobileMenuOpen}
      >
        {renderConversationSidebar({
          onListAfterSelect: () => setMobileMenuOpen(false),
        })}
      </div>

      <div className="flex min-h-0 min-w-0 w-full flex-1 flex-col overflow-hidden lg:hidden">
        {workspaceContent}
      </div>
      <div className="hidden min-h-0 w-full flex-1 overflow-hidden lg:flex">
        <ResizablePanelGroup orientation="horizontal">
          <ResizablePanel
            panelRef={sidebarPanelRef}
            defaultSize="20%"
            minSize="10%"
            maxSize="40%"
            collapsedSize="0%"
            collapsible
            onResize={(panelSize: { asPercentage: number }) => {
              setSidebarCollapsed(panelSize.asPercentage <= 1);
            }}
            className="min-h-0"
          >
            <div className="flex h-full min-h-0 flex-col overflow-hidden bg-white">
              {renderConversationSidebar()}
            </div>
          </ResizablePanel>
          <ResizableHandle withHandle />
          <ResizablePanel defaultSize="50%" minSize="32%" className="min-h-0">
            <div className="flex h-full min-h-0 flex-col overflow-hidden">
              {workspaceContent}
            </div>
          </ResizablePanel>
          <ResizableHandle withHandle />
          <ResizablePanel
            panelRef={infoPanelRef}
            defaultSize="20%"
            minSize="20%"
            maxSize="40%"
            collapsedSize="0%"
            collapsible
            onResize={(panelSize: { asPercentage: number }) => {
              setInfoPanelCollapsed(panelSize.asPercentage <= 1);
            }}
            className="min-h-0"
          >
            <CustomerInfoPanel conversation={conversation} className="h-full" />
          </ResizablePanel>
        </ResizablePanelGroup>
      </div>
      <ConversationTransferDialog
        open={transferOpen}
        mode="transfer"
        conversationId={conversation?.id ?? null}
        onOpenChange={setTransferOpen}
        onSuccess={async () => {
          setTransferOpen(false);
          if (conversation?.id) {
            await handleConversationChanged(conversation.id);
          }
        }}
      />
      <ConversationCloseDialog
        open={closeOpen}
        conversationId={conversation?.id ?? null}
        onOpenChange={setCloseOpen}
        onSuccess={async () => {
          setCloseOpen(false);
          if (conversation?.id) {
            await handleConversationChanged(conversation.id);
          }
        }}
      />

      <Sheet open={mobileCustomerSheetOpen} onOpenChange={setMobileCustomerSheetOpen}>
        <SheetContent
          side="right"
          className="flex w-full flex-col gap-0 border-l p-0 sm:max-w-md"
          showCloseButton
        >
          <SheetHeader className="shrink-0 space-y-1 border-b px-4 py-3 text-left">
            <SheetTitle>客户信息</SheetTitle>
            <SheetDescription>当前会话关联的客户与会话属性</SheetDescription>
          </SheetHeader>
          <CustomerInfoPanel
            conversation={conversation}
            variant="embedded"
            className="min-h-0 flex-1"
          />
        </SheetContent>
      </Sheet>
    </div>
  );
}
