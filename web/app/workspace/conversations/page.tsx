"use client";

import {
  ArrowRightLeftIcon,
  ChevronLeft,
  ChevronRight,
  CircleXIcon,
  Menu,
  MoreHorizontalIcon,
  SearchIcon,
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
import { Input } from "@/components/ui/input";
import {
  ResizableHandle,
  ResizablePanel,
  ResizablePanelGroup,
} from "@/components/ui/resizable";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { useAgentConversationRealtime } from "@/hooks/use-agent-conversation-realtime";
import {
  agentConversationFilterOptions,
  agentConversationSelectors,
  useAgentConversationsStore,
} from "@/lib/stores/agent-conversations";
import { ChatPanel } from "./_components/chat-panel";
import { ConversationList } from "./_components/conversation-list";

export default function ConversationsPage() {
  const conversation = useAgentConversationsStore(
    agentConversationSelectors.selectedConversation,
  );
  const searchKeyword = useAgentConversationsStore(
    (state) => state.searchKeyword,
  );
  const setSearchKeyword = useAgentConversationsStore(
    (state) => state.setSearchKeyword,
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
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);
  const [transferOpen, setTransferOpen] = useState(false);
  const [closeOpen, setCloseOpen] = useState(false);
  const sidebarPanelRef = useRef<PanelImperativeHandle | null>(null);
  const currentFilterLabel =
    agentConversationFilterOptions.find((opt) => opt.value === conversationFilter)
      ?.label ?? "筛选";

  useEffect(() => {
    void loadConversations().catch((error) => {
      toast.error(error instanceof Error ? error.message : "加载会话列表失败");
    });
  }, [loadConversations, searchKeyword, conversationFilter]);

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

  const sidebarContent = (
    <div className="flex h-full flex-col border-r lg:border-r-0">
      <div className="flex items-center justify-between gap-2 border-b p-2.5">
        <div className="flex flex-1 gap-2">
          <div className="relative flex-1">
            <SearchIcon className="absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
            <Input
              placeholder="搜索会话..."
              className="pl-9"
              value={searchKeyword}
              onChange={(e) => setSearchKeyword(e.target.value)}
            />
          </div>
          <div>
            <Select
              value={conversationFilter}
              onValueChange={(value) =>
                setConversationFilter(value as typeof conversationFilter)
              }
            >
              <SelectTrigger className="w-[116px]">
                <SelectValue>{currentFilterLabel}</SelectValue>
              </SelectTrigger>
              <SelectContent>
                {agentConversationFilterOptions.map((opt) => (
                  <SelectItem key={opt.value} value={opt.value}>
                    {opt.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
        </div>
        <Button
          variant="ghost"
          size="icon"
          className="ml-2 lg:hidden"
          onClick={() => setMobileMenuOpen(false)}
        >
          <X className="size-4" />
        </Button>
      </div>
      <ConversationList />
    </div>
  );

  const workspaceContent = (
    <div className="flex h-full min-h-0 w-full flex-1 flex-col overflow-hidden">
      <div className="shrink-0 flex items-center justify-between gap-3 border-b px-3 py-1">
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
              <Avatar className="size-9 shrink-0">
                <AvatarImage src="" />
                <AvatarFallback>客</AvatarFallback>
              </Avatar>
              <div className="min-w-0">
                <p className="truncate font-medium">{conversation.subject}</p>
                <p className="truncate text-sm text-muted-foreground">
                  <span>{conversation.channelType}</span> / <span>{conversation.externalUserId}</span>
                </p>
              </div>
            </>
          ) : (
            <div className="min-w-0">
              <p className="truncate font-medium">会话工作台</p>
              <p className="truncate text-sm text-muted-foreground">
                请选择左侧会话开始处理消息
              </p>
            </div>
          )}
        </div>
        <div className="flex shrink-0 items-center gap-1">
          <DropdownMenu>
            <DropdownMenuTrigger
              render={
                <Button
                  variant="ghost"
                  size="icon"
                  disabled={!conversation}
                />
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
        </div>
      </div>
      <div className="flex min-h-0 w-full flex-1 overflow-hidden">
        <ChatPanel />
      </div>
    </div>
  );

  return (
    <div className="flex h-full min-h-0 w-full overflow-hidden">
      <div
        className={`fixed inset-y-0 left-0 z-50 flex w-80 transform bg-white transition-transform duration-300 lg:hidden ${
          mobileMenuOpen ? "translate-x-0" : "-translate-x-full"
        }`}
      >
        {sidebarContent}
      </div>

      {mobileMenuOpen && (
        <div
          className="fixed inset-0 z-40 bg-black/50 lg:hidden"
          onClick={() => setMobileMenuOpen(false)}
        />
      )}

      <div className="flex min-h-0 w-full flex-1 overflow-hidden lg:hidden">
        {workspaceContent}
      </div>
      <div className="hidden min-h-0 w-full flex-1 overflow-hidden lg:flex">
        <ResizablePanelGroup orientation="horizontal">
          <ResizablePanel
            panelRef={sidebarPanelRef}
            defaultSize="28%"
            minSize="20%"
            collapsedSize="0%"
            collapsible
            onResize={(panelSize: { asPercentage: number }) => {
              setSidebarCollapsed(panelSize.asPercentage <= 1);
            }}
            className="min-h-0"
          >
            <div className="flex h-full min-h-0 flex-col overflow-hidden bg-white">
              {sidebarContent}
            </div>
          </ResizablePanel>
          <ResizableHandle withHandle />
          <ResizablePanel defaultSize="72%" minSize="40%" className="min-h-0">
            <div className="flex h-full min-h-0 flex-col overflow-hidden">
              {workspaceContent}
            </div>
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
    </div>
  );
}
