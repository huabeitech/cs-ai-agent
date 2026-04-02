# 客服工单前端详细设计

本文档用于在 [ticket-system-design.md](/Users/gaoyoubo/code/gaoyoubo/cs-agent/docs/ticket-system-design.md) 和 [ticket-backend-design.md](/Users/gaoyoubo/code/gaoyoubo/cs-agent/docs/ticket-backend-design.md) 的基础上，继续细化工单模块的前端实现方案。

本文档重点覆盖：

- 页面信息架构
- 列表页与详情页设计
- 会话转工单交互
- 前端 API 分层
- 组件拆分建议
- 表单与状态管理建议

本文档遵循当前项目前端约束：

- 前端目录为 `web`
- 框架为 `Next.js App Router`
- 基础组件优先使用 `shadcn/ui`
- 页面不直接散落裸 `fetch`
- API 统一收敛到 `web/lib/api/*`
- 后台列表/表单实现方式遵循 [frontend-list-form-best-practice.md](/Users/gaoyoubo/code/gaoyoubo/cs-agent/docs/frontend-list-form-best-practice.md)

## 1. 设计目标

工单前端的目标不是简单提供一组 CRUD 页面，而是为客服与主管提供一套可高频操作的处理工作台。

第一阶段要达成的用户体验目标：

- 客服能快速查看我的工单和待处理工单
- 客服能从会话页快速转工单
- 客服能在工单详情页完成回复、备注、转派、改状态、关闭、重开
- 主管能通过筛选和 SLA 信息快速发现风险工单

因此前端设计要兼顾两类使用形态：

- 列表驱动的高频操作
- 详情驱动的过程处理

## 2. 页面结构总览

建议新增以下页面与目录：

```text
web/app/workspace/tickets/
├── page.tsx
├── [id]/
│   └── page.tsx
└── _components/
    ├── ticket-filter-bar.tsx
    ├── ticket-list-table.tsx
    ├── ticket-edit-dialog.tsx
    ├── ticket-status-badge.tsx
    ├── ticket-priority-badge.tsx
    ├── ticket-detail-header.tsx
    ├── ticket-comment-timeline.tsx
    ├── ticket-reply-editor.tsx
    ├── ticket-side-panel.tsx
    ├── ticket-sla-card.tsx
    ├── ticket-related-conversation-card.tsx
    └── create-ticket-from-conversation-dialog.tsx
```

配套 API 文件建议新增：

```text
web/lib/api/ticket.ts
```

## 3. 导航与入口设计

### 3.1 主导航

建议在 `workspace` 主导航中增加：

- `工单`

如果当前工作台是会话优先，也可以先把工单放在会话旁边的一级导航中，形成：

- 会话
- 工单

### 3.2 会话页联动入口

在现有 `workspace/conversations` 中增加以下入口：

- `创建工单`
- `查看关联工单`
- `关联工单数量`

建议优先放置两个位置：

1. 会话头部操作区
2. 右侧 `conversation-info-panel`

原因：

- 头部操作区适合高频一键转单
- 信息面板适合查看上下文和已关联工单

## 4. 工单列表页设计

路径：

- `web/app/workspace/tickets/page.tsx`

列表页负责：

- 筛选
- 分页
- 批量入口
- 打开新建弹窗
- 进入详情页

按照项目现有规范，列表页应由 `page.tsx` 承载主要状态和接口调用。

## 4.1 页面布局

建议布局如下：

```text
顶部：页面标题 + 快捷视图 + 新建按钮
筛选区：关键词 / 状态 / 优先级 / 分类 / 团队 / 处理人 / SLA
主体区：工单表格
底部：分页
```

第一阶段不必先做复杂的左侧“保存视图”栏，可以先把常用视图做成顶部 tabs 或 segmented filter。

## 4.2 推荐默认视图

建议第一阶段至少提供：

- 全部工单
- 待我处理
- 我创建的
- 待客户反馈
- 已超时

实现方式建议：

- 先在前端以固定 tabs 形式实现
- 后续再升级成可配置的保存视图

## 4.3 筛选项建议

建议列表页支持：

- 关键词
- 状态
- 优先级
- 严重度
- 分类
- 当前处理团队
- 当前处理人
- 来源
- SLA 状态

关键词建议匹配：

- 工单号
- 标题
- 描述摘要

## 4.4 表格列建议

推荐默认列：

- 工单号
- 标题
- 客户
- 分类
- 优先级
- 状态
- 当前处理人
- 当前团队
- 最近更新时间
- 首次响应 SLA
- 解决 SLA
- 来源

第一阶段建议保持列数适中，不要一次塞太多维度。

## 4.5 行级操作建议

列表页行级操作建议提供：

- 查看详情
- 快速指派
- 改状态
- 关闭

第一阶段不建议在列表页放太多复杂操作，避免表格变得很重。

## 4.6 空状态设计

工单列表在无数据时建议明确区分：

- 无工单数据
- 当前筛选条件下无结果

空状态文案建议具备行动引导：

- 清空筛选
- 新建工单

## 5. 工单详情页设计

路径：

- `web/app/workspace/tickets/[id]/page.tsx`

详情页是工单处理主战场，承担：

- 查看工单完整信息
- 公开回复客户
- 记录内部备注
- 查看处理轨迹
- 修改状态、处理人、优先级、分类

## 5.1 详情页布局

建议采用双栏布局。

主区域：

- 工单头部
- 回复编辑器
- 评论/备注时间线
- 事件记录切换视图

右侧侧栏：

- 基础信息
- 客户信息
- 指派信息
- SLA 卡片
- 来源会话信息
- 关联工单

理由：

- 主区适合承载时间序列和编辑操作
- 侧栏适合承载稳定属性和摘要信息

## 5.2 详情页头部

建议展示：

- 工单号
- 标题
- 状态 badge
- 优先级 badge
- 严重度 badge
- 创建时间
- 创建人

头部操作区建议放：

- 编辑
- 指派
- 改状态
- 关闭
- 重开

## 5.3 评论时间线

评论时间线建议合并展示三类内容：

- 公开回复
- 内部备注
- 系统日志

但视觉上必须明确区分。

建议区分方式：

- `public_reply`：正常消息卡片
- `internal_note`：淡黄色或弱强调背景
- `system_log`：简洁灰色提示块

第一阶段如果担心混在一起太复杂，也可以做为两块：

- 互动记录
- 系统事件

## 5.4 回复编辑区

建议提供两个模式：

- 回复客户
- 内部备注

交互方式建议使用 tab 切换，而不是两个独立按钮。

编辑区应支持：

- 富文本或基础多行文本
- 附件
- 快捷回复
- 发送后自动刷新评论时间线

第一阶段如实现复杂度有限，可以先做：

- 多行文本
- 内部备注
- 公开回复

先不做富文本高级能力。

## 5.5 SLA 展示卡片

详情页右侧建议有独立 SLA 卡片。

展示内容：

- 首次响应 SLA
- 解决 SLA
- 当前状态
- 目标时长
- 已耗时
- 剩余时长
- 是否超时

颜色建议：

- 正常：中性
- 临近超时：黄色
- 已超时：红色
- 已完成：绿色

## 5.6 来源会话卡片

如果工单来自会话，应在侧栏显示：

- 会话标题
- 客户最近消息摘要
- 当前会话状态
- 查看会话按钮

此卡片的作用是让客服在处理工单时不需要来回切系统。

## 6. 新建/编辑工单弹窗设计

建议文件：

- `web/app/workspace/tickets/_components/ticket-edit-dialog.tsx`

用途：

- 手动新建工单
- 编辑工单基础信息

遵循项目后台表单规范：

- `page.tsx` 负责接口调用
- `ticket-edit-dialog.tsx` 负责表单
- 使用 `react-hook-form + zod + Field`

## 6.1 字段建议

第一阶段建议弹窗包含：

- 标题
- 描述
- 分类
- 优先级
- 严重度
- 当前团队
- 当前处理人
- 截止时间

如果是从会话转工单进入：

- 客户和会话来源默认显示，不允许随意修改或仅允许部分修改

## 6.2 表单结构建议

建议拆出：

- `schema`
- `FormValues`
- `emptyForm`
- `buildForm(item)`
- `buildPayload(form)`

与现有 `frontend-list-form-best-practice.md` 保持一致。

## 6.3 提交策略

创建和编辑都由父层页面提供 `onSubmit`。

表单层职责仅限：

- 渲染
- 校验
- 回填
- payload 转换

不直接请求接口。

## 7. 会话转工单弹窗设计

建议新增：

- `create-ticket-from-conversation-dialog.tsx`

这个弹窗很关键，它不是普通“新建工单”，而是“基于当前会话上下文建单”。

## 7.1 触发入口

建议在以下位置触发：

- 会话头部 “转工单” 按钮
- 会话右侧信息面板中的 “创建工单” 按钮

## 7.2 弹窗内容

建议包含以下区域：

### 基础信息

- 工单标题
- 问题描述
- 分类
- 优先级
- 严重度

### 会话上下文

- 客户名称
- 会话主题
- 最近几条消息摘要
- 当前标签

### 处理信息

- 分配团队
- 分配客服
- 是否回写会话提示

## 7.3 默认值策略

默认带入：

- 标题：会话标题
- 描述：会话最近消息摘要
- 客户：当前会话客户
- 当前处理人：会话当前客服

后续增强：

- AI 生成标题
- AI 生成问题摘要
- AI 推荐分类和优先级

## 7.4 创建成功后

建议创建成功后支持：

- toast 提示
- 刷新会话详情
- 更新右侧关联工单区域
- 提供跳转到工单详情页

## 8. 前端 API 设计

建议统一新增：

- `web/lib/api/ticket.ts`

所有工单请求通过该 service 文件收敛，不允许在页面内直接写业务 `fetch`。

## 8.1 推荐导出方法

```ts
fetchTickets(params)
fetchTicketDetail(id)
createTicket(payload)
createTicketFromConversation(payload)
updateTicket(payload)
assignTicket(payload)
changeTicketStatus(payload)
replyTicket(payload)
addTicketInternalNote(payload)
closeTicket(payload)
reopenTicket(payload)
watchTicket(payload)
unwatchTicket(payload)
fetchTicketComments(params)
fetchTicketEvents(params)
```

## 8.2 类型定义建议

建议在 `web/lib/api/ticket.ts` 中定义：

- `TicketItem`
- `TicketDetail`
- `TicketComment`
- `TicketEvent`
- `TicketSLA`
- `TicketListParams`
- `CreateTicketPayload`
- `UpdateTicketPayload`

第一阶段如果字段较多，也可以将类型拆到：

- `web/lib/api/types/ticket.ts`

但如果当前项目倾向将类型和 API 放在一起，先放同文件也可以。

## 9. 页面状态管理建议

第一阶段工单模块不建议一开始引入新的全局 store，可以优先采用页面内局部状态。

原因：

- 列表和详情的状态边界比较清晰
- 工单和会话不像消息流那样强依赖实时同步
- 先保持实现简单

建议：

- 列表页用 `useState + useEffect/useCallback`
- 详情页用局部 `loading/saving/actionLoading`
- 如果后续需要跨组件联动，再考虑引入 store

## 9.1 列表页建议状态

列表页通常包含：

- `keyword`
- `statusFilter`
- `priorityFilter`
- `categoryFilter`
- `teamFilter`
- `assigneeFilter`
- `result`
- `loading`
- `dialogOpen`
- `editingItem`
- `actionLoadingId`

## 9.2 详情页建议状态

详情页通常包含：

- `ticket`
- `comments`
- `events`
- `loading`
- `saving`
- `statusDialogOpen`
- `assignDialogOpen`
- `editDialogOpen`

## 10. 组件拆分建议

## 10.1 ticket-filter-bar.tsx

职责：

- 列表筛选 UI
- 仅负责渲染和回调
- 不直接发请求

输入参数建议：

- 当前筛选值
- `onKeywordChange`
- `onStatusChange`
- `onPriorityChange`
- `onRefresh`
- `onCreate`

## 10.2 ticket-list-table.tsx

职责：

- 展示表格
- 处理空状态
- 触发行级操作回调

不在表格内部做接口请求。

## 10.3 ticket-edit-dialog.tsx

职责：

- 创建/编辑工单表单
- 回填与校验
- `onSubmit` 向上抛 payload

## 10.4 ticket-detail-header.tsx

职责：

- 展示标题、编号、状态、优先级
- 承载编辑、指派、关闭、重开等入口

## 10.5 ticket-comment-timeline.tsx

职责：

- 按时间序列渲染回复、备注、系统记录

建议支持传入统一结构化数组，而不是在组件内部判断过多业务逻辑。

## 10.6 ticket-reply-editor.tsx

职责：

- 回复客户 / 内部备注编辑
- 切换模式
- 表单提交

## 10.7 ticket-side-panel.tsx

职责：

- 展示客户、SLA、指派、来源会话、关联工单等摘要信息

## 11. 交互细节建议

## 11.1 快速状态变更

建议在详情页头部操作区支持快速改状态，但不要直接点击即生效。

推荐方式：

- 打开轻量确认弹层
- 需要填写原因时展示输入项

例如：

- 改为 `待客户反馈` 时可填 `pendingReason`
- 关闭时必须填 `closeReason`

## 11.2 回复后刷新策略

回复成功后建议：

- 清空编辑器
- 刷新评论列表
- 刷新工单详情头部状态
- 必要时刷新 SLA 卡片

## 11.3 列表与详情联动

建议点击列表行进入详情页，而不是都走弹窗详情。

原因：

- 工单详情信息重，弹窗不适合长期处理
- 详情页更适合后续加入评论时间线和侧栏信息

## 11.4 URL 设计

建议保持：

- 列表页：`/workspace/tickets`
- 详情页：`/workspace/tickets/{id}`

如果后续要保留列表筛选状态，可以再考虑 query 参数保留筛选条件。

## 12. 类型与枚举前端同步建议

工单相关前后端共用枚举最终应走项目既有方式：

- 后端定义
- 前端通过 `make enums` 生成

前端展示时建议统一做：

- `TicketStatusLabels`
- `TicketPriorityLabels`
- `TicketSeverityLabels`

避免在页面中手写分散的文案映射。

## 13. 第一阶段实现顺序建议

推荐按以下顺序推进前端：

1. 补 `web/lib/api/ticket.ts`
2. 补工单列表页
3. 补 `ticket-edit-dialog.tsx`
4. 跑通新建工单
5. 补工单详情页
6. 跑通回复、备注、改状态、关闭、重开
7. 在会话页接入“转工单”弹窗
8. 联调关联工单展示

## 14. 与当前前端规范的对应关系

工单模块应明确遵循以下现有规范：

- 列表页由 `page.tsx` 承载列表和异步状态
- 表单弹窗由 `_components/edit.tsx` 或等价业务组件承载
- 页面层负责调用 `web/lib/api/ticket.ts`
- 表单层不直接调用接口
- 时间展示统一使用 `formatDateTime`
- 基础组件优先复用 `shadcn/ui`

如果需要下拉选择团队和客服，优先考虑复用项目内已有的下拉封装与 `option-combobox.tsx`，不要直接手写另一套选择器。

## 15. 结论

工单前端的关键不是把表单补齐，而是把以下三类体验做好：

- 列表页的高频筛选与进入详情效率
- 详情页的处理闭环效率
- 会话转工单的联动效率

只要这三类体验成型，第一阶段的工单模块就能真正被客服团队用起来。后续再逐步叠加 AI 摘要、自动化、保存视图和报表，演进路径会比较顺。
