# 贝壳AI客服

> An AI-agent-first customer support system that unifies live chat, knowledge retrieval, ticketing, and seamless human handoff.

`贝壳AI客服`是一个以 **AI Agent 为核心** 的智能客服系统，面向需要同时处理在线咨询、知识库问答、人工接管和工单流转的团队。

它不是一个简单的聊天机器人，而是一套围绕客服场景设计的完整系统：

- AI Agent 先接待，优先处理常见问题与标准流程
- 知识库检索驱动回答，支持 RAG 场景
- 在低置信度、无答案或命中规则时转人工
- 后台可管理会话、工单、客服组、知识库、AI 配置、Skills 和 MCP
- 提供管理后台与客服工作台两套视角

## 核心能力

- AI-first 客服流程：AI Agent 优先接待，支持自动回复、兜底与人工协同
- 在线会话系统：支持访客会话、会话分配、转接、关闭、未读状态与实时消息
- 知识库 RAG：支持知识库、文档、切片、检索日志与检索质量分析
- 工单系统：支持会话转工单、工单分类、状态流转与处理闭环
- 客服组织管理：支持客服档案、客服组、排班与分配能力
- AI 扩展能力：支持 Skills、MCP 调试与外部能力接入
- 双工作区体验：管理后台负责配置和运营，客服工作台负责处理会话

## AI Agent 工作流程

```mermaid
flowchart TD
    A[用户发起咨询<br/>Web Widget / Open API] --> B[创建或匹配会话]
    B --> C[客户发送消息]
    C --> D[触发 AI Reply Runtime]
    D --> E[加载会话历史 / AI 配置 / 可用工具]
    E --> F[按需选择 Skill]
    F --> G[按需注入知识库检索结果]
    G --> H[Agent 运行并决定下一步]
    H --> I{直接回复?}
    I -- 是 --> J[LLM 生成回复并返回用户]
    J --> K{问题是否结束?}
    K -- 否 --> C
    K -- 是 --> L[结束会话或沉淀数据]
    I -- 否 --> M{是否调用 Graph / MCP Tool?}
    M -- 是 --> N[执行 Skill / Graph / MCP Tool]
    N --> O{需要用户确认?}
    O -- 否 --> H
    O -- 是 --> P[向用户发起确认]
    P --> Q{用户确认结果}
    Q -- 取消 --> J
    Q -- 确认转人工 --> R[会话转人工并进入待接入池]
    R --> S[按客服组/排班自动分配或人工分配]
    S --> T[客服工作台接管]
    T --> U{是否需要工单跟踪?}
    U -- 是 --> V[人工处理中创建/关联工单]
    U -- 否 --> W[人工继续处理]
    V --> W
    W --> X[问题解决并关闭]
    Q -- 确认建单 --> Y[从当前会话创建工单]
    Y --> H
    M -- 否 --> J
```

## 核心业务流程

### 1. 会话处理流程

```mermaid
flowchart TD
    A[用户进入 Web Widget / Open IM] --> B[创建或匹配会话]
    B --> C[客户发送消息]
    C --> D[写入 message / 更新 conversation]
    D --> E{当前是否允许 AI 回复?}
    E -- 是 --> F[异步触发 AI Reply]
    E -- 否 --> G[等待人工处理]
    F --> H[AI 回复消息写回会话]
    H --> I{用户是否继续追问?}
    I -- 是 --> C
    I -- 否 --> J[会话关闭或保持待处理]
    G --> K[客服接管 / 回复 / 转接 / 关闭]
    K --> J
```

### 2. 人工接管与分配流程

```mermaid
flowchart TD
    A[AI 判断需人工介入] --> B[发起转人工确认]
    B --> C{用户是否确认?}
    C -- 否 --> D[继续 AI 对话]
    C -- 是 --> E[会话状态置为 pending]
    E --> F[记录 handoffAt / handoffReason]
    F --> G[按 AI Agent 绑定客服组尝试自动分配]
    G --> H{是否分配成功?}
    H -- 是 --> I[进入客服工作台 Active 会话]
    H -- 否 --> J[留在待接入池]
    J --> K[主管或客服手动分配]
    K --> I
    I --> L[客服处理、转接或关闭]
```

### 3. 会话转工单流程

```mermaid
flowchart TD
    A[会话中出现投诉 / 售后 / 报障诉求] --> B{由 AI 还是人工发起?}
    B -- AI --> C[Graph Tool 整理工单草稿]
    C --> D[发起建单确认]
    D --> E{用户是否确认?}
    E -- 否 --> F[继续对话或补充信息]
    E -- 是 --> G[从当前会话创建工单]
    B -- 人工 --> H[客服工作台发起从会话建单]
    H --> G
    G --> I[写入 ticket / event log]
    I --> J[回写会话事件]
    J --> K[进入工单指派与状态流转]
    K --> L[工单处理完成并关闭]
```

### 4. 知识库处理流程

```mermaid
flowchart TD
    A[后台创建知识库] --> B[新增文档 / FAQ]
    B --> C[文档清洗与切片]
    C --> D[写入向量索引]
    D --> E[AI Agent 绑定知识库]
    E --> F[用户提问触发 AI Runtime]
    F --> G[按知识库配置执行检索]
    G --> H{是否命中有效片段?}
    H -- 是 --> I[将知识片段注入运行时上下文]
    I --> J[Agent 基于知识生成回复]
    H -- 否 --> K[走知识库 fallback 文案或升级策略]
    J --> L[记录检索日志 / 运行日志]
    K --> L
```

### 5. AI Reply Runtime 流程

```mermaid
flowchart TD
    A[客户消息进入运行时] --> B[装载会话历史]
    B --> C[加载 AI Config / 可用工具]
    C --> D[尝试命中 Skill]
    D --> E[按需注入知识库检索结果]
    E --> F[构造 Agent 输入消息]
    F --> G[Agent 执行]
    G --> H{输出类型}
    H -- 直接回复 --> I[写回 AI 消息]
    H -- Tool 调用 --> J[执行 Graph / MCP Tool]
    J --> K{是否触发确认中断?}
    K -- 否 --> G
    K -- 是 --> L[保存 checkpoint / pending interrupt]
    L --> M[等待用户回复确认或取消]
    M --> N[恢复执行 Resume]
    N --> G
```

## 适用场景

- 官网在线客服
- SaaS 产品支持
- AI + 人工混合接待
- 企业内部服务台或运营支持台
- 需要知识库问答与人工协同的客服团队

## 技术栈

- Backend: Golang
- Frontend: Next.js 16 + React 19 + shadcn/ui + Tailwind CSS
- Database: SQLite / MySQL
- Vector DB: Qdrant
- AI: OpenAI-compatible LLM / Embedding + RAG + SKILLS + MCP

## 项目结构

```text
.
├── cmd/                    # server / migration / generator
├── internal/
│   ├── controllers/        # API controllers
│   ├── services/           # business services
│   ├── repositories/       # data access
│   ├── models/             # GORM models
│   ├── migration/          # data migrations
│   └── ai/                 # LLM / RAG / MCP related logic
├── dashboard/              # dashboard (Next.js)
├── widget/                 # embeddable customer chat widget
├── config/                 # config files
└── docs/                   # project docs
```

## 快速开始

### 1. 环境要求

- Go `1.26+`
- Node.js `20+`
- `pnpm`
- Qdrant

### 2. 准备配置

复制示例配置：

```bash
cp config/config.example.yaml config/config.yaml
```

默认配置使用：

- SQLite：`data/app.db`
- Backend：`http://127.0.0.1:8083`
- Qdrant gRPC：`127.0.0.1:6334`

### 3. 启动 Qdrant

如果你本地还没有 Qdrant，可以用 Docker 快速启动：

```bash
docker run -p 6333:6333 -p 6334:6334 qdrant/qdrant
```

### 4. 安装前端依赖

```bash
cd web
pnpm install
cd ..
```

### 5. 启动项目

同时启动后端和前端：

```bash
make run
```

或分别启动：

```bash
make run-server
make run-web
```

## 常用命令

```bash
make run          # 同时启动后端和前端
make run-server   # 启动后端
make run-web      # 启动前端
make build        # 构建后端二进制
make test         # 运行 Go 测试
make tidy         # go mod tidy
make generator    # 执行代码生成
make enums        # 生成前端枚举
make migration    # 执行 migration
```

## 系统视角

- 管理后台：负责 AI Agent、知识库、客服组、工单与运营配置
- 客服工作台：负责接管会话、处理消息与人工服务
- Widget：负责承接用户侧咨询入口

这使得`贝壳AI客服`可以同时覆盖：

- AI 接待
- 人工协同
- 知识驱动回答
- 工单追踪闭环

## 开源定位

`贝壳AI客服`适合作为以下方向的开源基础项目：

- AI 客服系统
- AI Helpdesk / AI Support Platform
- RAG + Human Handoff 的落地样板
- 面向企业场景的 AI Agent 应用框架

如果你在寻找一个 **以 AI Agent 为中心，而不是仅仅把 LLM 嵌进聊天框** 的客服系统，这个项目就是为此设计的。
