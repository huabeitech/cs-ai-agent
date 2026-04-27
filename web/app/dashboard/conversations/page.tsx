"use client";

import {
  ArrowRightLeftIcon,
  ChevronLeft,
  ChevronRight,
  ChevronsUpDown,
  CircleUserRoundIcon,
  CircleXIcon,
  FilePlus2Icon,
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
  DropdownMenuRadioGroup,
  DropdownMenuRadioItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  ResizableHandle,
  ResizablePanel,
  ResizablePanelGroup,
} from "@/components/ui/resizable";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Sheet, SheetContent } from "@/components/ui/sheet";
import { useAgentConversationRealtime } from "@/hooks/use-agent-conversation-realtime";
import {
  agentConversationFilterOptions,
  agentConversationSelectors,
  type AgentConversationFilterKey,
  useAgentConversationsStore,
} from "@/lib/stores/agent-conversations";
import { CreateTicketFromConversationDialog } from "../tickets/_components/create-ticket-from-conversation-dialog";
import { ChatPanel } from "./_components/chat-panel";
import { ConversationInfoPanel } from "./_components/conversation-info-panel";
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
  const [createTicketOpen, setCreateTicketOpen] = useState(false);
  const sidebarPanelRef = useRef<PanelImperativeHandle | null>(null);
  const infoPanelRef = useRef<PanelImperativeHandle | null>(null);
  const filterContainerRef = useRef<HTMLDivElement | null>(null);
  const filterMeasureRef = useRef<HTMLDivElement | null>(null);
  const [showFilterDropdown, setShowFilterDropdown] = useState(false);

  useEffect(() => {
    const container = filterContainerRef.current;
    const measure = filterMeasureRef.current;
    if (!container || !measure) {
      return;
    }

    const updateFilterMode = () => {
      setShowFilterDropdown(measure.scrollWidth > container.clientWidth);
    };

    updateFilterMode();

    const observer = new ResizeObserver(() => {
      updateFilterMode();
    });

    observer.observe(container);
    observer.observe(measure);

    return () => {
      observer.disconnect();
    };
  }, []);

  const currentFilterOption =
    agentConversationFilterOptions.find((opt) => opt.value === conversationFilter) ??
    agentConversationFilterOptions[0];
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
    <div className="flex h-full min-h-0 flex-1 flex-col bg-inherit">
      <div className="flex h-12.5 shrink-0 items-start justify-between gap-2 border-b border-border p-2">
        <div ref={filterContainerRef} className="relative min-w-0 flex-1">
          {showFilterDropdown ? (
            <DropdownMenu>
              <DropdownMenuTrigger
                render={
                  <Button
                    variant="outline"
                    className="h-8.5 w-full min-w-0 justify-between gap-2 px-3 text-xs sm:text-sm"
                  />
                }
              >
                <span className="truncate">{currentFilterOption?.label ?? "筛选状态"}</span>
                <ChevronsUpDown className="size-4 shrink-0 text-muted-foreground" />
              </DropdownMenuTrigger>
              <DropdownMenuContent align="start" className="w-44 min-w-44">
                <DropdownMenuRadioGroup
                  value={conversationFilter}
                  onValueChange={(value) =>
                    setConversationFilter(value as AgentConversationFilterKey)
                  }
                >
                  {agentConversationFilterOptions.map((opt) => (
                    <DropdownMenuRadioItem key={opt.value} value={opt.value}>
                      {opt.label}
                    </DropdownMenuRadioItem>
                  ))}
                </DropdownMenuRadioGroup>
              </DropdownMenuContent>
            </DropdownMenu>
          ) : (
            <Tabs
              value={conversationFilter}
              onValueChange={(value) =>
                setConversationFilter(value as AgentConversationFilterKey)
              }
              className="min-w-0 flex-1 gap-0"
            >
              <TabsList
                className="w-full min-w-0 justify-start"
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
          )}
          <div
            ref={filterMeasureRef}
            className="pointer-events-none absolute whitespace-nowrap opacity-0"
            aria-hidden="true"
          >
            <div className="inline-flex">
              {agentConversationFilterOptions.map((opt) => (
                <span
                  key={opt.value}
                  className="shrink-0 px-2.5 text-xs sm:text-sm"
                >
                  {opt.label}
                </span>
              ))}
            </div>
          </div>
        </div>
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
    <div className="flex h-full min-h-0 w-full flex-1 flex-col overflow-hidden bg-background text-foreground">
      <div className="flex h-12.5 shrink-0 items-center justify-between gap-3 border-b border-border px-3 py-1">
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
                <div className="flex items-center gap-2">
                  <p className="min-w-0 truncate font-medium leading-tight">
                    {conversation.customerName || `客户 #${conversation.customerId || conversation.id}`}
                  </p>
                  <span
                    className={`inline-flex shrink-0 items-center gap-1 rounded-full border px-2 text-[11px] ${
                      conversation.customerOnline
                        ? "border-emerald-200 bg-emerald-50 text-emerald-700"
                        : "border-slate-200 bg-slate-100 text-slate-600"
                    }`}
                  >
                    <span
                      className={`size-1.5 rounded-full ${
                        conversation.customerOnline
                          ? "bg-emerald-500"
                          : "bg-slate-400"
                      }`}
                    />
                    {conversation.customerOnline ? "用户在线" : "用户离线"}
                  </span>
                </div>
                <p className="mt-0.5 truncate text-xs text-muted-foreground sm:text-sm">
                  <span>渠道 #{conversation.channelId || "-"}</span>
                  {conversation.customerId ? (
                    <>
                      <span className="text-muted-foreground/60"> / </span>
                      <span>已关联客户</span>
                    </>
                  ) : null}
                </p>
              </div>
            </>
          ) : (
            <div className="min-w-0">
              <p className="truncate font-medium text-[14px] leading-tight">会话工作台</p>
              <p className="mt-0.5 truncate text-[14px] text-muted-foreground sm:text-[14px] lg:hidden">
                打开菜单选择会话
              </p>
              <p className="mt-0.5 hidden truncate text-[12px] text-muted-foreground lg:block">
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
            aria-label="会话信息"
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
                onClick={() => setCreateTicketOpen(true)}
                disabled={!conversation}
              >
                <FilePlus2Icon />
                转工单
              </DropdownMenuItem>
              <DropdownMenuItem
                onClick={() => setTransferOpen(true)}
                disabled={!conversation || conversation.status !== 3}
              >
                <ArrowRightLeftIcon />
                转接会话
              </DropdownMenuItem>
              <DropdownMenuItem
                onClick={() => setCloseOpen(true)}
                disabled={!conversation || conversation.status === 4}
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
            aria-label={infoPanelCollapsed ? "展开会话信息" : "收起会话信息"}
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
    <div className="flex h-[calc(100dvh-var(--header-height))] min-h-0 w-full min-w-0 flex-col overflow-hidden lg:h-full">
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
        className={`fixed top-12 bottom-0 left-0 z-40 flex w-[min(22rem,calc(100vw-0.75rem))] max-w-[min(22rem,calc(100vw-0.75rem))] flex-col overflow-hidden border-r border-border bg-card text-card-foreground shadow-lg transition-transform duration-300 ease-out will-change-transform touch-manipulation overscroll-contain supports-[padding:max(0px)]:pb-[env(safe-area-inset-bottom)] lg:hidden ${
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
            <div className="flex h-full min-h-0 flex-col overflow-hidden bg-card text-card-foreground">
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
            defaultSize="500px"
            minSize="20%"
            maxSize="40%"
            collapsedSize="0%"
            collapsible
            onResize={(panelSize: { asPercentage: number }) => {
              setInfoPanelCollapsed(panelSize.asPercentage <= 1);
            }}
            className="min-h-0"
          >
            <ConversationInfoPanel conversation={conversation} className="h-full" />
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
      <CreateTicketFromConversationDialog
        open={createTicketOpen}
        onOpenChange={setCreateTicketOpen}
        conversation={
          conversation
            ? {
                id: conversation.id,
                customerName: conversation.customerName,
                customerId: conversation.customerId ?? 0,
                lastMessageSummary: conversation.lastMessageSummary,
                currentAssigneeId: conversation.currentAssigneeId,
              }
            : null
        }
        onSuccess={() => {
          setCreateTicketOpen(false);
        }}
      />

      <Sheet open={mobileCustomerSheetOpen} onOpenChange={setMobileCustomerSheetOpen}>
        <SheetContent
          side="right"
          className="flex w-full flex-col gap-0 border-l p-0 sm:max-w-md"
          showCloseButton
        >
          <ConversationInfoPanel
            conversation={conversation}
            variant="embedded"
            className="min-h-0 flex-1"
          />
        </SheetContent>
      </Sheet>
    </div>
  );
}
