# 客服 Agent 整体方案设计

本文档用于定义本项目“客服机器人 / 客服 Agent”能力的整体设计方案，目标是在现有 IM 会话体系上，逐步接入并串联以下三类能力：

- `RAG`：根据知识库检索结果回答用户问题
- `MCP`：根据可注册的 MCP Server / Tool 获取实时外部结果
- `Skill`：根据用户意图选择合适的技能，并按技能流程执行

本文档优先解决两个问题：

1. 当前代码现状距离“客服 Agent”还差什么
2. 接下来应该按什么结构与步骤落地，才能先跑通、再逐步增强

## 1. 设计目标

### 1.1 最终目标

构建一个面向客服场景的统一 Agent Runtime，使其能够在一轮用户消息到来后，根据上下文自动决定：

- 直接回复
- 走知识库检索回答
- 调用 MCP 工具查询实时结果
- 命中某个 Skill 并按技能流程执行
- 无法可靠处理时转人工

### 1.2 非目标

当前阶段不追求以下能力：

- 通用开放式 Autonomous Agent
- 无限步 ReAct 推理
- 任意工具自由调用
- 无约束的多轮自我规划

本项目优先做“可控、可观测、可运营”的客服 Agent，而不是通用 AI Agent 平台。

## 2. 当前现状评估

结合当前代码，现状可归纳如下：

### 2.1 已有能力

- 已有会话、消息、转人工、分配客服、WebSocket 推送等 IM 基础设施
- 已有 `AIAgent`、`AIConfig`、`KnowledgeBase`、`SkillDefinition` 等基础模型
- 已有 RAG 检索、召回、重排、知识库问答 debug 能力
- 已有 MCP client、MCP server、MCP debug 接口

### 2.2 当前不足

#### 2.2.1 自动回复链路仍然只是“RAG 回复器”

当前 `internal/services/ai_reply_service.go` 的主流程是：

1. 收到用户消息
2. 判断是否可由 AI 回复
3. 检索知识库
4. 命中则生成回答
5. 未命中则 fallback 或 handoff

这条链路没有接入 MCP，也没有接入 Skill。

#### 2.2.2 MCP 只有调试入口，没有进入对话运行时

当前 `internal/controllers/console/mcp_controller.go` 和 `internal/services/mcp_debug_service.go` 主要用于：

- 查看服务列表
- 测试连接
- 列出工具
- 手工调用工具

这套能力仍然属于“后台调试能力”，并不是“客服对话运行时能力”。

#### 2.2.3 Skill 只有表结构和骨架，没有真正执行链

当前 `internal/ai/skills` 下：

- 有 `RuntimeContext`
- 有 `ExecutionPlan`
- 有 `SkillRunLog`
- 但 `MatchSkill()` 仍未实现
- 也没有实际 Skill Executor
- 更没有参数抽取、补槽、MCP 工具编排、结果归纳

#### 2.2.4 缺少统一的 Agent 编排层

当前没有一个统一的 `AgentRuntime / Orchestrator` 来做以下事情：

- 加载本轮上下文
- 做动作决策
- 分派到 RAG / MCP / Skill
- 统一输出最终客服话术
- 统一记录运行日志

因此三类能力目前是孤立的，不具备真正“串起来运行”的基础。

## 3. 总体架构

建议引入统一的 Agent Runtime，形成如下分层：

```text
Message -> AIReplyService -> AgentRuntime
                               |
                               +-- Planner
                               +-- RAGExecutor
                               +-- MCPExecutor
                               +-- SkillExecutor
                               +-- Composer
                               +-- RunLog / Trace
```

### 3.1 各模块职责

#### 3.1.1 `AIReplyService`

定位：对话入口服务。

职责：

- 接收用户消息触发
- 做基础幂等和上下文检查
- 调用 `AgentRuntime.RunConversationTurn(...)`
- 根据结果发送 AI 消息或转人工

禁止：

- 直接堆叠 RAG / MCP / Skill 细节
- 直接承载复杂编排逻辑

#### 3.1.2 `AgentRuntime`

定位：统一编排层。

职责：

- 加载会话、消息、Agent、最近对话记录
- 生成本轮执行计划
- 执行一个受控动作
- 统一整理为最终客服回复
- 统一记录 trace / run log

这是整套方案的核心。

#### 3.1.3 `Planner`

定位：动作决策器。

职责：

- 根据当前问题和上下文，在有限动作集合中选择本轮动作
- 输出结构化 plan，而不是自由文本

本阶段只允许选择一个主动作：

- `reply`
- `rag`
- `mcp`
- `skill`
- `handoff`

#### 3.1.4 `RAGExecutor`

定位：知识库执行器。

职责：

- 进行检索、重排、上下文构建
- 生成严格受知识约束的答案
- 返回引用与检索日志

#### 3.1.5 `MCPExecutor`

定位：工具执行器。

职责：

- 选择并调用允许的 MCP Tool
- 限制调用次数
- 标准化工具输出
- 生成工具调用日志

#### 3.1.6 `SkillExecutor`

定位：技能执行器。

职责：

- 选择 Skill
- 抽取参数
- 判断是否缺少必要槽位
- 调用 MCP 或其他能力完成技能
- 产出技能结果

#### 3.1.7 `Composer`

定位：最终回复整合器。

职责：

- 将 `RAG / MCP / Skill` 的执行结果转换为统一客服话术
- 限制语气、格式、敏感表达
- 控制是否引用知识库片段或工具结果

## 4. 推荐目录拆分

建议新增 `internal/ai/agent` 目录，避免把编排逻辑继续堆在 `services` 中。

建议结构如下：

```text
internal/ai/agent/
├── runtime.go
├── planner.go
├── composer.go
├── types.go
├── context.go
├── log.go
└── executors/
    ├── rag_executor.go
    ├── mcp_executor.go
    └── skill_executor.go
```

与现有模块的关系：

- `services/ai_reply_service.go`：只保留入口职责
- `ai/rag/*`：继续承载知识检索与问答能力
- `ai/mcps/*`：继续承载 MCP 协议接入能力
- `ai/skills/*`：继续承载 Skill 模型与运行时能力

## 5. 统一主流程

建议将一轮客服 AI 回复统一为以下步骤。

### 5.1 Step 1：加载上下文

输入：

- `conversationID`
- `messageID`

加载内容：

- 当前消息
- 当前会话
- 关联的 `AIAgent`
- 关联的 `AIConfig`
- 最近若干轮消息
- 会话状态

输出建议定义为 `TurnContext`。

### 5.2 Step 2：基础规则判断

在进入 planner 前，先做确定性规则判断：

- 只有客户消息才触发
- 只有最后一条消息才继续
- 已转人工则不再 AI 回复
- 已有接待客服则不再 AI 回复
- Agent 未启用则直接退出
- 用户明确要求人工则转人工
- 达到最大 AI 回复轮次则转人工

这一层仍然建议保留在 `AIReplyService` 或 `AgentRuntime` 的前置校验中。

### 5.3 Step 3：Planner 产出执行计划

Planner 输入：

- 用户问题
- 最近对话摘要
- Agent 能力配置
- 可用知识库信息
- 可用 Skill 列表
- 可用 MCP Tool 清单

Planner 输出结构化 JSON：

```json
{
  "action": "reply|rag|mcp|skill|handoff",
  "reason": "选择原因",
  "skillCode": "",
  "serverCode": "",
  "toolName": "",
  "arguments": {},
  "needUserInput": false,
  "askUser": ""
}
```

要求：

- planner 只输出结构化动作，不直接输出最终用户答案
- planner 不允许自行执行工具
- planner 的动作空间必须受白名单约束

### 5.4 Step 4：执行主动作

按 `action` 分派：

- `reply`：直接生成简短回复
- `rag`：执行知识库问答
- `mcp`：执行工具查询
- `skill`：执行技能逻辑
- `handoff`：直接转人工

本阶段建议“一轮只执行一个主动作”，不要一开始就做多步链式动作。

### 5.5 Step 5：统一生成客服话术

执行器返回的结果，不应直接原样输出给用户，而应进入统一 `Composer`：

- 统一语气
- 统一结构
- 避免工具原始 JSON 直接暴露给用户
- 对知识库结果做引用压缩
- 对工具结果做客服式描述

### 5.6 Step 6：落库与推送

统一完成：

- 发送 AI 消息
- 记录会话事件
- 写入 Agent 运行日志
- 写入 Skill 运行日志
- 写入 MCP 调用日志
- 复用知识检索日志

## 6. 三类能力如何定位

### 6.1 RAG 的定位

适用于：

- FAQ
- 产品说明
- 流程规则
- 售后政策
- 操作步骤

不适用于：

- 实时订单状态
- 实时库存
- 用户专属数据
- 需要外部系统查询的结果

因此 RAG 本质上解决“静态知识回答”。

### 6.2 MCP 的定位

适用于：

- 实时查询
- 外部系统集成
- 可控的业务工具调用

典型例子：

- 查订单状态
- 查物流轨迹
- 查账户余额
- 查预约时间

MCP 本质上解决“工具查询与动作执行”。

### 6.3 Skill 的定位

Skill 不应只是“一段 prompt”，而应是“面向场景的受控业务流程模板”。

典型例子：

- `order_query`
- `refund_policy`
- `refund_apply`
- `delivery_track`
- `appointment_reschedule`

Skill 可以内部：

- 调用 MCP Tool
- 调用 RAG
- 调用其他受控能力

Skill 本质上解决“场景化业务编排”。

## 7. 数据模型建议

当前模型不足以表达完整 Agent 配置，建议按阶段补充。

### 7.1 `AIAgent` 扩展建议

建议增加以下能力配置字段，优先使用独立关联表而不是逗号字符串：

- `enabledSkills`
- `enabledMcpServers`
- `plannerPrompt`
- `responsePrompt`
- `routingMode`
- `maxToolCallsPerTurn`
- `maxPlannerSteps`
- `handoffScoreThreshold`

推荐新增关联表：

- `ai_agent_skill_rel`
- `ai_agent_mcp_server_rel`

原因：

- 便于后台配置
- 便于查询
- 不依赖逗号字符串解析
- 后续能扩展优先级与单独配置

### 7.2 `SkillDefinition` 扩展建议

当前 `SkillDefinition` 只有：

- `Code`
- `Name`
- `Description`
- `Prompt`

建议补充：

- `TriggerConfig`
- `ExecutionMode`
- `ExecutionConfig`
- `AllowedMCPTools`
- `RequiredSlots`
- `OutputTemplate`
- `TimeoutMS`
- `NeedConfirmation`

建议的含义：

- `TriggerConfig`：触发规则，如关键词、意图、示例问句
- `ExecutionMode`：`prompt_only`、`mcp_tool_chain`、`hybrid`
- `ExecutionConfig`：执行流程配置
- `AllowedMCPTools`：此技能允许调用的工具白名单
- `RequiredSlots`：执行前必须具备的参数

### 7.3 日志模型建议

建议新增统一的 `AgentRunLog`，用于记录每一轮 AI 执行全链路。

建议字段：

- `ConversationID`
- `MessageID`
- `AIAgentID`
- `PlanAction`
- `PlanReason`
- `PlanPayload`
- `ExecutionStatus`
- `ReplyText`
- `ErrorMessage`
- `LatencyMs`
- `CreatedAt`

同时建议新增 `MCPToolCallLog`：

- `ConversationID`
- `MessageID`
- `AIAgentID`
- `ServerCode`
- `ToolName`
- `Arguments`
- `ResultSummary`
- `IsError`
- `LatencyMs`

现有日志使用建议：

- `KnowledgeRetrieveLog`：继续记录 RAG
- `SkillRunLog`：继续记录 Skill
- `ConversationEventLog`：继续记录会话事件

## 8. MCP 运行时设计

### 8.1 不建议直接复用 Debug Service

`MCPDebugService` 的职责应保持为后台调试，不建议直接用于线上对话运行时。

建议新增：

- `internal/services/mcp_runtime_service.go`

职责：

- 根据 Agent 配置筛选可用 Server
- 根据 Skill / Agent 配置筛选可用 Tool
- 限制调用次数
- 统一错误处理
- 写工具调用日志

### 8.2 MCP 运行时安全原则

必须具备：

- Tool 白名单
- Server 白名单
- 超时控制
- 单轮最大调用次数
- 错误隔离
- 对用户不可见的内部错误日志

不建议一开始支持：

- 任意工具自动调用
- 写操作类高风险工具
- 无确认的副作用操作

第一阶段建议仅支持只读查询类工具。

## 9. Skill 运行时设计

### 9.1 Skill 的推荐抽象

Skill 是一个“场景执行模板”，不是一个自由 agent。

推荐每个 Skill 至少包含：

- 适用场景说明
- 触发规则
- 必要参数定义
- 缺参时追问文案
- 允许调用的工具
- 结果汇总方式

### 9.2 Skill 运行时流程

建议流程如下：

1. 命中 Skill
2. 解析已知槽位
3. 判断是否缺参
4. 缺参则直接向用户追问
5. 参数足够则执行工具或查询
6. 汇总结果为客服话术
7. 写 `SkillRunLog`

### 9.3 Skill 的执行模式建议

初期建议仅支持以下 2 种：

#### 9.3.1 `prompt_only`

仅基于 Skill Prompt + 上下文输出结果。

适合：

- 固定回答模板
- 简单政策说明

#### 9.3.2 `mcp_tool_chain`

由 Skill 决定调用哪个工具，并在拿到结果后生成回复。

适合：

- 订单查询
- 物流查询
- 预约查询

等到基础能力稳定后，再增加：

- `hybrid`

即先查工具，再结合知识库结果解释。

## 10. Planner 设计建议

### 10.1 为什么需要 Planner

如果没有统一 planner，后续代码容易变成：

- 先试 RAG
- 再试 MCP
- 再试 Skill
- 再试 handoff

最终演变为复杂的 if/else，难以维护。

Planner 的作用是先判断“当前最适合哪条路径”，然后只执行一条主路径。

### 10.2 Planner 的动作空间

建议限制为：

- `reply`
- `rag`
- `mcp`
- `skill`
- `handoff`

不要允许 planner 一开始就自由输出：

- 任意工具链
- 多步思维过程
- 复杂结构化计划树

先把动作空间做小，可靠性更高。

### 10.3 Planner 的输入建议

建议输入信息包括：

- 用户消息
- 最近几轮对话
- Agent 描述
- 可用知识库清单
- 可用 Skill 列表
- 可用 MCP Tool 摘要
- 转人工约束

### 10.4 Planner 的输出约束

必须：

- 输出固定 JSON
- 后端严格校验 JSON
- 非法输出按兜底处理

禁止：

- 直接把 planner 输出文本原样发给用户

## 11. 回复生成与提示词建议

建议把提示词职责分开，而不是一个大 Prompt 包办所有事情。

### 11.1 推荐分层

- `PlannerPrompt`：只负责决定动作
- `ExecutorPrompt`：只负责对应执行器内部处理
- `ComposerPrompt`：只负责生成最终客服话术

### 11.2 回复风格约束

客服场景建议统一要求：

- 简洁
- 明确
- 不编造
- 不暴露内部实现
- 不输出工具原始报错
- 资料不足时明确说明

### 11.3 RAG 场景约束

建议默认要求：

- 优先依据知识库
- 资料不足时说明不足
- 不要凭常识编造政策和金额

### 11.4 工具场景约束

建议要求：

- 工具结果转成自然语言
- 如结果不完整，明确说明“我查到的信息是...”
- 不直接暴露 JSON 或内部字段名

## 12. 可观测性与运营能力

客服 Agent 要能运营，必须先能观测。

### 12.1 必须记录的链路信息

- 本轮 action 是什么
- 为什么选这个 action
- RAG 命中了什么
- 调用了哪些工具
- Skill 是否命中
- 为什么转人工
- 最终回复内容
- 总耗时

### 12.2 建议支持的后台视角

后续后台建议提供以下调试能力：

- 查看每轮 AgentRunLog
- 查看对应的 KnowledgeRetrieveLog
- 查看 SkillRunLog
- 查看 MCPToolCallLog
- 重放 planner 输入输出

### 12.3 失败分类建议

建议把失败分为：

- `no_answer`
- `tool_error`
- `skill_not_matched`
- `planner_invalid`
- `timeout`
- `handoff`

便于运营统计与调优。

## 13. 并发与幂等建议

当前用户消息发送后会异步触发 AI 回复，后续接入多步能力后，需要加强会话级串行控制。

### 13.1 风险

- 同一会话连续发多条消息，可能并发跑多个 AI 任务
- 旧任务比新任务晚完成，可能回写过期答案
- MCP 工具可能被重复调用

### 13.2 建议

至少做以下控制：

- 每轮执行前确认 `conversation.LastMessageID == messageID`
- 任务关键节点再次确认消息未过期
- 后续可增加会话级分布式锁或串行队列

在第一阶段，至少保证“过期任务不回写消息”。

## 14. 分阶段实施方案

建议严格分阶段，不要一次性做完整 agent。

### 14.1 Phase 1：重构出 AgentRuntime，先跑通 RAG

目标：

- 保持现有 AI 自动回复能力不退化
- 把 `AIReplyService` 收敛成入口层
- 引入 `AgentRuntime`
- 引入 `RAGExecutor`
- 增加统一 `AgentRunLog`

交付结果：

- 现有 RAG 自动回复仍可用
- 主流程从“服务里写死”转成“编排层 + 执行器”

### 14.2 Phase 2：支持显式 Skill 执行

目标：

- 先不做自动命中
- 先支持 `manualSkillCode`
- 跑通 `SkillExecutor`
- 跑通 `SkillRunLog`

交付结果：

- 后台可指定一个 Skill 执行
- Skill 可先支持 `prompt_only`

### 14.3 Phase 3：接入 MCP Runtime

目标：

- 新增 `MCPRuntimeService`
- 支持 Agent / Skill 受控调用 MCP Tool
- 增加 `MCPToolCallLog`

交付结果：

- 能做“订单查询”“物流查询”类只读场景

### 14.4 Phase 4：接入 Planner

目标：

- 让模型从 `rag / mcp / skill / handoff` 中选一个动作
- 输出结构化 plan
- 严格校验 plan

交付结果：

- 客服 Agent 初步具备自动路由能力

### 14.5 Phase 5：补槽与多轮技能

目标：

- Skill 支持缺参追问
- 支持多轮补全槽位

交付结果：

- 用户没提供订单号时，Agent 能追问而不是直接失败

## 15. 第一版最小可运行范围建议

如果目标是尽快跑通，推荐第一版范围如下：

### 15.1 必做

- `AgentRuntime`
- `RAGExecutor`
- `AgentRunLog`
- `AIReplyService` 重构接入 runtime
- 运行时仍支持 fallback / handoff

### 15.2 可做

- `manualSkillCode`
- `prompt_only` Skill

### 15.3 暂不做

- 自动 Skill 匹配
- 多步 Planner
- 任意 MCP 自动工具调用
- 高风险写操作工具

这样可以在最短路径上把架构搭起来，避免继续把复杂度堆进 `AIReplyService`。

## 16. 推荐实施顺序

建议按以下顺序开发：

1. 抽出 `internal/ai/agent` 目录与类型定义
2. 将 `AIReplyService` 改为调用 `AgentRuntime`
3. 把现有 RAG 自动回复逻辑迁入 `RAGExecutor`
4. 增加 `AgentRunLog`
5. 接入 `manualSkillCode` 和 `SkillExecutor`
6. 新增 `MCPRuntimeService`
7. 做 Planner 自动路由
8. 做多轮补槽

## 17. 总结

本项目当前并不是“离客服 Agent 只差一点点配置”，而是“已经具备 RAG、MCP、Skill 的基础零件，但缺少统一编排层”。

后续的关键不是继续在 `AIReplyService` 上叠加分支，而是尽快补上：

- `AgentRuntime`
- `Planner`
- `Executor`
- `RunLog`

一旦这层建立起来，后续接入：

- 更复杂的 Skill
- 更多 MCP Tool
- 更细粒度的路由策略
- 更完善的多轮问答

都会变得可控。

现阶段最务实的路线是：

1. 先重构出统一 Runtime
2. 保持 RAG 可跑
3. 再逐步把 Skill 和 MCP 挂上去

这也是本项目从“AI 自动回复”演进到“客服 Agent”的最稳路径。
