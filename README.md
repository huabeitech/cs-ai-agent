# 贝壳AI客服

> An AI-agent-first customer support system that unifies live chat, knowledge retrieval, ticketing, and seamless human handoff.

`贝壳AI客服`是一个以 **AI Agent 为核心** 的智能客服系统，面向需要同时处理在线咨询、知识库问答、人工接管和工单流转的团队。

它不是一个简单的聊天机器人，而是一套围绕客服场景设计的完整系统：

- AI Agent 先接待，优先处理常见问题与标准流程
- 知识库检索驱动回答，并通过 Answerability Gate 防止超出知识库范围时乱答
- 知识库无法支撑回答时返回兜底提示，并建议用户联系人工客服
- 后台可管理会话、工单、客服组、知识库、AI 配置、Skills 和 MCP
- 提供管理后台、客服工作台与客户侧 Web 接入入口

## 核心能力

- AI-first 客服流程：AI Agent 优先接待，支持自动回复、兜底与人工协同
- 在线会话系统：支持访客会话、会话分配、转接、关闭、未读状态与实时消息
- 知识库 RAG：支持知识库、文档、切片、检索日志与检索质量分析
- 工单系统：支持会话转工单、工单分类、状态流转与处理闭环
- 客服组织管理：支持客服档案、客服组、排班与分配能力
- AI 扩展能力：支持 Skills、MCP 调试与外部能力接入
- 统一前端工程：管理后台、客服工作台、客户侧 Web 接入页与嵌入式 SDK 统一放在 `web` 目录

## AI Agent 工作流程

```mermaid
flowchart TD
    A[用户发起咨询<br/>Web 客服入口 / Open API] --> B[创建或匹配会话]
    B --> C[客户发送消息]
    C --> D[触发 AI Reply Runtime]
    D --> E[加载会话历史 / AI 配置]
    E --> F[按绑定知识库执行检索]
    F --> G{知识片段是否足以回答?}
    G -- 否 --> Z[返回知识库兜底提示<br/>并建议联系人工客服]
    G -- 是 --> H[准备 Skills / MCP Tools]
    H --> I[将可信知识上下文交给 Agent]
    I --> J{直接回复?}
    J -- 是 --> K[LLM 基于知识生成回复并返回用户]
    K --> L{问题是否结束?}
    L -- 否 --> C
    L -- 是 --> M[结束会话或沉淀数据]
    J -- 否 --> N{是否调用 Graph / MCP Tool?}
    N -- 是 --> O[执行 Skill / Graph / MCP Tool]
    O --> P{需要用户确认?}
    P -- 否 --> I
    P -- 是 --> Q[向用户发起确认]
    Q --> R{用户确认结果}
    R -- 取消 --> K
    R -- 确认转人工 --> S[会话转人工并进入待接入池]
    S --> T[按客服组/排班自动分配或人工分配]
    T --> U[客服工作台接管]
    U --> V{是否需要工单跟踪?}
    V -- 是 --> W[人工处理中创建/关联工单]
    V -- 否 --> X[人工继续处理]
    W --> X
    X --> Y[问题解决并关闭]
    R -- 确认建单 --> AA[从当前会话创建工单]
    AA --> I
    N -- 否 --> K
```

## 核心业务流程

### 1. 会话处理流程

```mermaid
flowchart TD
    A[用户进入 Web 客服入口 / Open IM] --> B[创建或匹配会话]
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
    G --> H[Answerability Gate 判定知识是否足以回答]
    H --> I{可被知识片段支撑?}
    I -- 是 --> J[将可信知识片段注入运行时上下文]
    J --> K[Agent 基于知识生成回复]
    I -- 否 --> L[返回知识库 fallback 文案<br/>并建议联系人工客服]
    K --> M[记录检索日志 / Answerability Trace / 运行日志]
    L --> M
```

### 5. AI Reply Runtime 流程

```mermaid
flowchart TD
    A[客户消息进入运行时] --> B[装载会话历史]
    B --> C[加载 AI Config]
    C --> D[按绑定知识库检索]
    D --> E{Answerability Gate 是否通过?}
    E -- 否 --> F[写回 fallback 回复<br/>并建议联系人工客服]
    E -- 是 --> G[加载可用工具并尝试命中 Skill]
    G --> H[构造带可信知识的 Agent 输入消息]
    H --> I[Agent 执行]
    I --> J{输出类型}
    J -- 直接回复 --> K[写回 AI 消息]
    J -- Tool 调用 --> L[执行 Graph / MCP Tool]
    L --> M{是否触发确认中断?}
    M -- 否 --> I
    M -- 是 --> N[保存 checkpoint / pending interrupt]
    N --> O[等待用户回复确认或取消]
    O --> P[恢复执行 Resume]
    P --> I
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
├── web/                    # unified Next.js frontend
│   ├── app/dashboard/      # admin dashboard
│   ├── app/kefu/           # customer service entry and chat pages
│   ├── components/         # React components
│   ├── lib/                # API client, SDK source and utilities
│   ├── public/sdk/         # built embeddable SDK assets
│   └── scripts/            # frontend build scripts
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

开发环境默认入口：

- 管理后台：`http://localhost:3000/dashboard`
- 客服工作台：`http://localhost:3000/dashboard/conversations`
- 客户侧 Web 接入示例：`http://localhost:3000/kefu`
- 客户侧聊天页：`http://localhost:3000/kefu/chat`

生产构建时，前端统一由 `web` 工程构建，静态产物输出到 `web/out`，后端会从 `web/out` 提供静态资源。

## 常用命令

```bash
make run          # 同时启动后端和前端
make run-server   # 启动后端
make run-web      # 启动前端
make build        # 构建前端静态资源并打包当前平台
make build-web    # 构建 web 应用和嵌入式 SDK
make test         # 运行 Go 测试
make tidy         # go mod tidy
make generator    # 执行代码生成
make enums        # 生成前端枚举
make migration    # 执行 migration
```

## 系统视角

- 管理后台：负责 AI Agent、知识库、客服组、工单与运营配置
- 客服工作台：负责接管会话、处理消息与人工服务
- 客户侧 Web 接入：通过 `/kefu`、`/kefu/chat` 和 `web/public/sdk` 中的嵌入式脚本承接用户咨询入口

这使得`贝壳AI客服`可以同时覆盖：

- AI 接待
- 人工协同
- 知识驱动回答
- 工单追踪闭环

## 开源定位

`贝壳AI客服`适合作为以下方向的开源基础项目：

- AI 客服系统
- AI Helpdesk / AI Support Platform
- RAG 可回答性判定 + Human Handoff 的落地样板
- 面向企业场景的 AI Agent 应用框架

如果你在寻找一个 **以 AI Agent 为中心，而不是仅仅把 LLM 嵌进聊天框** 的客服系统，这个项目就是为此设计的。
