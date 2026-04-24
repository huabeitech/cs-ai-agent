"use client";

import { Loader2Icon, PlugZapIcon, WrenchIcon } from "lucide-react";
import { useEffect, useMemo, useState } from "react";
import { toast } from "sonner";

import { JsonCodeEditor } from "@/components/json-code-editor";
import { JsonViewer } from "@/components/json-viewer";
import { OptionCombobox } from "@/components/option-combobox";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Drawer,
  DrawerContent,
  DrawerDescription,
  DrawerFooter,
  DrawerHeader,
  DrawerTitle,
} from "@/components/ui/drawer";
import { Label } from "@/components/ui/label";
import {
  callMCPTool,
  listMCPServers,
  listMCPTools,
  testMCPConnection,
  type MCPConnectionResult,
  type MCPServerInfo,
  type MCPToolCallResult,
  type MCPToolInfo,
} from "@/lib/api/admin";

const defaultServerCode = "";

export default function MCPDashboardPage() {
  const [serverCode, setServerCode] = useState(defaultServerCode);
  const [servers, setServers] = useState<MCPServerInfo[]>([]);
  const [connection, setConnection] = useState<MCPConnectionResult | null>(
    null,
  );
  const [tools, setTools] = useState<MCPToolInfo[]>([]);
  const [loadingServers, setLoadingServers] = useState(true);
  const [testing, setTesting] = useState(false);
  const [loadingTools, setLoadingTools] = useState(false);
  const [callingTool, setCallingTool] = useState(false);
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [activeTool, setActiveTool] = useState<MCPToolInfo | null>(null);
  const [argumentsText, setArgumentsText] = useState("{}");
  const [toolResult, setToolResult] = useState<MCPToolCallResult | null>(null);
  const [argumentsError, setArgumentsError] = useState<string | null>(null);

  const serverOptions = useMemo(
    () =>
      servers.map((server) => ({
        value: server.code,
        label: server.enabled
          ? `${server.code} (${server.endpoint})`
          : `${server.code} (disabled)`,
      })),
    [servers],
  );

  useEffect(() => {
    async function loadServers() {
      setLoadingServers(true);
      try {
        const result = await listMCPServers();
        setServers(result);
        const firstServer = result[0];
        if (firstServer) {
          setServerCode((current) => current || firstServer.code);
        }
      } catch (error) {
        toast.error(
          error instanceof Error ? error.message : "加载 MCP 服务失败",
        );
      } finally {
        setLoadingServers(false);
      }
    }

    void loadServers();
  }, []);

  async function handleTestConnection() {
    setTesting(true);
    try {
      const result = await testMCPConnection(serverCode.trim());
      setConnection(result);
      toast.success("MCP 连接成功");
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "MCP 连接失败");
    } finally {
      setTesting(false);
    }
  }

  async function handleListTools() {
    setLoadingTools(true);
    try {
      const result = await listMCPTools(serverCode.trim());
      setTools(result);
      toast.success(`已加载 ${result.length} 个工具`);
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "加载工具失败");
    } finally {
      setLoadingTools(false);
    }
  }

  async function handleCallTool() {
    if (!activeTool) {
      toast.error("请先选择一个工具");
      return;
    }
    if (argumentsError) {
      toast.error("Arguments JSON 格式不合法");
      return;
    }

    let parsedArguments: Record<string, unknown> = {};
    try {
      parsedArguments = argumentsText.trim()
        ? (JSON.parse(argumentsText) as Record<string, unknown>)
        : {};
    } catch {
      toast.error("arguments 必须是合法 JSON");
      return;
    }

    setCallingTool(true);
    try {
      const result = await callMCPTool({
        serverCode: serverCode.trim(),
        toolName: activeTool.name,
        arguments: parsedArguments,
      });
      setToolResult(result);
      toast.success(result.isError ? "工具返回错误结果" : "工具调用成功");
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "工具调用失败");
    } finally {
      setCallingTool(false);
    }
  }

  function openToolDrawer(tool: MCPToolInfo) {
    setActiveTool(tool);
    setToolResult(null);
    if (tool.name === "lorem") {
      setArgumentsText('{\n  "wordCount": 8\n}');
    } else if (tool.name === "ping") {
      setArgumentsText('{\n  "message": "hello from dashboard"\n}');
    } else {
      setArgumentsText("{}");
    }
    setDrawerOpen(true);
  }

  return (
    <>
      <div className="space-y-6 p-6">
        {/* <div className="flex flex-col gap-2">
          <div className="flex flex-col gap-2 md:flex-row md:items-end md:justify-between">
            <div className="flex items-center gap-2">
              <Button
                variant="outline"
                onClick={() => {
                  setConnection(null)
                  setTools([])
                  setToolResult(null)
                  setActiveTool(null)
                }}
              >
                <RefreshCwIcon className="mr-2 size-4" />
                清空结果
              </Button>
            </div>
          </div>
        </div> */}

        <div className="flex flex-wrap items-center justify-between gap-4 rounded-lg border px-4 py-2">
            <div className="flex min-w-0 flex-1 items-center gap-3">
              <span className="shrink-0 text-sm font-medium">服务配置</span>
              <div className="min-w-[280px] max-w-[520px] flex-1">
                <OptionCombobox
                  value={serverCode}
                  options={serverOptions}
                  placeholder="选择一个 MCP Server"
                  searchPlaceholder="搜索 serverCode"
                  emptyText="没有可用的 MCP Server"
                  disabled={loadingServers}
                  onChange={(value) => {
                    setServerCode(value);
                    setConnection(null);
                    setTools([]);
                    setToolResult(null);
                    setActiveTool(null);
                  }}
                />
              </div>
            </div>
            <div className="flex shrink-0 flex-wrap gap-2">
              <Button
                onClick={() => void handleTestConnection()}
                disabled={testing || loadingServers || !serverCode}
              >
                {testing ? (
                  <Loader2Icon className="mr-2 size-4 animate-spin" />
                ) : (
                  <PlugZapIcon className="mr-2 size-4" />
                )}
                测试连接
              </Button>
              <Button
                variant="outline"
                onClick={() => void handleListTools()}
                disabled={loadingTools || loadingServers || !serverCode}
              >
                {loadingTools ? (
                  <Loader2Icon className="mr-2 size-4 animate-spin" />
                ) : (
                  <WrenchIcon className="mr-2 size-4" />
                )}
                列出工具
              </Button>
            </div>
            {connection ? (
              <div className="w-full rounded-lg border bg-muted/30 p-4 text-sm">
                <div className="flex flex-wrap items-center gap-2">
                  <Badge>已连接</Badge>
                  <span className="font-medium">
                    {connection.serverName || "-"}
                  </span>
                </div>
                <div className="mt-3 grid gap-2 text-muted-foreground grid-cols-5">
                  <div>serverCode: {connection.serverCode}</div>
                  <div>protocol: {connection.protocol || "-"}</div>
                  <div className="md:col-span-2 break-all">
                    endpoint: {connection.endpoint}
                  </div>
                  <div>version: {connection.version || "-"}</div>
                </div>
              </div>
            ) : null}
        </div>

        {tools.length === 0 ? (
          <div className="rounded-lg border border-dashed p-6 text-sm text-muted-foreground">
            暂无工具结果，先点击上方“列出工具”。
          </div>
        ) : (
          <div className="border rounded-lg">
            <div className="overflow-hidden">
              <div className="grid grid-cols-[minmax(0,220px)_minmax(0,1fr)_88px] gap-4 border-b bg-muted/40 px-4 py-3 text-sm font-medium text-muted-foreground">
                <div>工具名</div>
                <div>描述</div>
                <div className="text-right">操作</div>
              </div>
              {tools.map((tool) => (
                <div
                  key={tool.name}
                  className="grid grid-cols-[minmax(0,220px)_minmax(0,1fr)_88px] gap-4 border-b px-4 py-4 text-sm last:border-b-0"
                >
                  <div className="font-medium">{tool.name}</div>
                  <div className="text-muted-foreground">
                    {tool.description || "-"}
                  </div>
                  <div className="text-right">
                    <Button
                      type="button"
                      variant="outline"
                      size="sm"
                      onClick={() => openToolDrawer(tool)}
                    >
                      查看
                    </Button>
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}
      </div>

      <Drawer open={drawerOpen} direction="right" onOpenChange={setDrawerOpen}>
        <DrawerContent className="min-w-3xl">
          <DrawerHeader>
            <DrawerTitle>{activeTool?.name || "工具详情"}</DrawerTitle>
            <DrawerDescription>
              在这里查看工具详细信息，并使用当前选择的 MCP Server
              直接做一次真实调用测试。
            </DrawerDescription>
          </DrawerHeader>
          <div className="flex-1 space-y-4 overflow-y-auto px-4 pb-4">
            {activeTool ? (
              <>
                <div className="rounded-lg border p-4">
                  <div className="flex flex-wrap items-center gap-2">
                    <Badge variant="secondary">{activeTool.name}</Badge>
                    {activeTool.title ? (
                      <span className="text-sm font-medium">
                        {activeTool.title}
                      </span>
                    ) : null}
                  </div>
                  <p className="mt-3 text-sm text-muted-foreground">
                    {activeTool.description || "暂无描述"}
                  </p>
                </div>

                <div className="space-y-4 rounded-lg border p-4">
                  <div>
                    <p className="mb-2 text-xs font-medium text-muted-foreground">
                      Input Schema
                    </p>
                    <JsonViewer value={activeTool.inputSchema} />
                  </div>
                  <div>
                    <p className="mb-2 text-xs font-medium text-muted-foreground">
                      Output Schema
                    </p>
                    <JsonViewer value={activeTool.outputSchema} />
                  </div>
                </div>

                <div className="space-y-4 rounded-lg border p-4">
                  <div className="grid gap-2">
                    <Label htmlFor="tool-arguments">Arguments JSON</Label>
                    <JsonCodeEditor
                      value={argumentsText}
                      onChange={setArgumentsText}
                      onValidationChange={setArgumentsError}
                    />
                  </div>
                  <Button
                    onClick={() => void handleCallTool()}
                    disabled={
                      callingTool ||
                      loadingServers ||
                      !serverCode ||
                      !!argumentsError
                    }
                  >
                    {callingTool ? (
                      <Loader2Icon className="mr-2 size-4 animate-spin" />
                    ) : (
                      <WrenchIcon className="mr-2 size-4" />
                    )}
                    测试工具
                  </Button>
                  {toolResult ? (
                    <div className="space-y-4 rounded-lg border p-4">
                      <div className="flex flex-wrap items-center gap-2">
                        <Badge
                          variant={
                            toolResult.isError ? "destructive" : "default"
                          }
                        >
                          {toolResult.isError ? "返回错误" : "调用成功"}
                        </Badge>
                        <span className="text-sm font-medium">
                          {toolResult.toolName}
                        </span>
                      </div>
                      <div>
                        <p className="mb-2 text-xs font-medium text-muted-foreground">
                          Content
                        </p>
                        <JsonViewer value={toolResult.content} />
                      </div>
                      <div>
                        <p className="mb-2 text-xs font-medium text-muted-foreground">
                          Structured Content
                        </p>
                        <JsonViewer value={toolResult.structuredContent} />
                      </div>
                    </div>
                  ) : null}
                </div>
              </>
            ) : null}
          </div>
          <DrawerFooter>
            <Button variant="outline" onClick={() => setDrawerOpen(false)}>
              关闭
            </Button>
          </DrawerFooter>
        </DrawerContent>
      </Drawer>
    </>
  );
}
