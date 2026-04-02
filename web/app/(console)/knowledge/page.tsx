"use client"

import { useState } from "react"
import { KnowledgeBaseList } from "./_components/knowledge-base-list"
import { DebugPanel } from "./_components/debug-panel"
import { DocumentList, type DocumentListActionState } from "./_components/document-list"
import { RetrieveLogList } from "./_components/retrieve-log-list"
import { Button } from "@/components/ui/button"
import {
  Bug,
  LayoutGridIcon,
  LayoutListIcon,
  PanelLeftCloseIcon,
  PanelLeftOpenIcon,
  PlusIcon,
  RefreshCwIcon,
  WrenchIcon,
} from "lucide-react"
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import type { KnowledgeBase } from "@/lib/api/admin"

export default function DashboardKnowledgeDocumentsPage() {
  const [selectedKnowledgeBase, setSelectedKnowledgeBase] = useState<KnowledgeBase | null>(null)
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false)
  const [debugPanelOpen, setDebugPanelOpen] = useState(false)
  const [activeTab, setActiveTab] = useState("documents")
  const [documentActionState, setDocumentActionState] = useState<DocumentListActionState | null>(null)

  return (
    <div className="flex h-[calc(100vh-4rem)]">
      <div
        className={`shrink-0 overflow-hidden transition-[width] duration-200 ${
          sidebarCollapsed ? "w-0" : "w-80"
        }`}
      >
        <KnowledgeBaseList
          selectedKnowledgeBaseId={selectedKnowledgeBase?.id ?? null}
          onSelectKnowledgeBase={setSelectedKnowledgeBase}
        />
      </div>
      <div className="relative shrink-0 bg-background">
        <Button
          variant="outline"
          size="icon"
          className="absolute top-4 left-1/2 z-10 size-7 -translate-x-1/2 rounded-full shadow-sm"
          onClick={() => setSidebarCollapsed((value) => !value)}
          aria-label={sidebarCollapsed ? "展开知识库列表" : "折叠知识库列表"}
        >
          {sidebarCollapsed ? (
            <PanelLeftOpenIcon className="size-3.5" />
          ) : (
            <PanelLeftCloseIcon className="size-3.5" />
          )}
        </Button>
      </div>
      <div className="min-w-0 min-h-0 flex-1">
        <Tabs value={activeTab} onValueChange={setActiveTab} className="h-full min-h-0 gap-0">
          <div className="border-b px-6 py-4">
            <div className="flex items-center gap-2">
              <TabsList>
                <TabsTrigger value="documents">文档</TabsTrigger>
                <TabsTrigger value="retrieveLogs">检索日志</TabsTrigger>
              </TabsList>
              {activeTab === "documents" && documentActionState ? (
                <div className="ml-auto flex items-center gap-1">
                  <Button
                    variant="ghost"
                    size="icon"
                    className="size-7"
                    onClick={documentActionState.onRefresh}
                    disabled={documentActionState.loading}
                    aria-label="刷新文档"
                  >
                    <RefreshCwIcon className={documentActionState.loading ? "size-4 animate-spin" : "size-4"} />
                  </Button>
                  <Button
                    variant={documentActionState.viewMode === "list" ? "secondary" : "ghost"}
                    size="icon"
                    className="size-7"
                    onClick={() => documentActionState.onChangeViewMode("list")}
                    aria-label="列表布局"
                  >
                    <LayoutListIcon className="size-4" />
                  </Button>
                  <Button
                    variant={documentActionState.viewMode === "grid" ? "secondary" : "ghost"}
                    size="icon"
                    className="size-7"
                    onClick={() => documentActionState.onChangeViewMode("grid")}
                    aria-label="网格布局"
                  >
                    <LayoutGridIcon className="size-4" />
                  </Button>
                  <Button
                    variant="ghost"
                    size="icon"
                    className="size-7"
                    onClick={() => setDebugPanelOpen(true)}
                    aria-label="打开调试面板"
                  >
                    <Bug className="size-4" />
                  </Button>
                  <Button
                    variant="ghost"
                    size="icon"
                    className="size-7"
                    onClick={documentActionState.onCreate}
                    aria-label="新增文档"
                  >
                    <PlusIcon className="size-4" />
                  </Button>
                </div>
              ) : null}
            </div>
          </div>
          <TabsContent value="documents" className="min-h-0 flex-1">
            <DocumentList 
              knowledgeBaseId={selectedKnowledgeBase?.id ?? null}
              onActionStateChange={setDocumentActionState}
            />
          </TabsContent>
          <TabsContent value="retrieveLogs" className="min-h-0 flex-1">
            <RetrieveLogList
              knowledgeBaseId={selectedKnowledgeBase?.id ?? null}
            />
          </TabsContent>
        </Tabs>
      </div>
      <Sheet open={debugPanelOpen} onOpenChange={setDebugPanelOpen}>
        <Button
          variant="outline"
          size="icon"
          className="absolute bottom-4 right-4 z-10"
          onClick={() => setDebugPanelOpen(true)}
        >
          <WrenchIcon className="size-4" />
        </Button>
        <SheetContent side="right" className="min-w-170">
          <SheetHeader>
            <SheetTitle>RAG 调试</SheetTitle>
          </SheetHeader>
          <DebugPanel knowledgeBaseId={selectedKnowledgeBase?.id ?? null} />
        </SheetContent>
      </Sheet>
    </div>
  )
}
