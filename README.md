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
    B --> C[进入 AI Agent]
    C --> D{是否命中 Skill / 工具流程?}
    D -- 是 --> E[执行 Skill / MCP Tool]
    D -- 否 --> F[进入知识库检索]
    F --> G[召回相关知识片段]
    G --> H[LLM 生成回答]
    E --> H
    H --> I{回答是否可信且可直接返回?}
    I -- 是 --> J[返回给用户]
    J --> K{问题是否结束?}
    K -- 否 --> C
    K -- 是 --> L[结束会话或沉淀数据]
    I -- 否 --> M[触发转人工策略]
    M --> N[进入待接入池 / 指定客服组]
    N --> O[客服工作台接管]
    O --> P{是否需要工单跟踪?}
    P -- 是 --> Q[创建或关联工单]
    P -- 否 --> R[人工继续处理]
    Q --> R
    R --> S[问题解决并关闭]
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
├── web/                    # admin console (Next.js)
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
