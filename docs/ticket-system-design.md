# 客服工单系统设计方案

本文档用于定义本项目的客服工单能力设计，目标是在现有 `Conversation / Message / Customer / AgentTeam` 基础上，补齐“异步问题处理、跨人协作、SLA 管控、过程留痕、运营统计”的正式工单能力。

本文档优先解决以下问题：

1. 工单在本系统中的定位是什么
2. 工单与现有会话系统应该如何配合
3. 第一阶段应该落哪些表、接口和页面
4. 后续如何逐步扩展到自动化、AI 辅助和运营能力

## 1. 背景与目标

### 1.1 当前现状

当前系统已经具备以下基础能力：

- IM 会话与消息体系
- 客户与客户联系人体系
- 客服、客服组、会话分配与转接能力
- 标签、快捷回复、知识库、AI Agent 能力

这些能力已经足够支撑“实时接待”，但还不足以支撑“正式处理”。典型缺口如下：

- 缺少正式的问题承载对象，无法把聊天问题沉淀成可跟踪事项
- 缺少跨班次、跨客服、跨团队交接的标准载体
- 缺少状态、责任人、截止时间、SLA、升级、重开等处理机制
- 缺少内部备注、外部回复、事件留痕、工单报表

### 1.2 产品目标

工单模块的核心目标是让系统同时具备两类能力：

- `Conversation` 负责实时沟通
- `Ticket` 负责正式处理闭环

具体目标：

- 支持会话问题一键转工单
- 支持手动建单、表单建单、外部系统建单
- 支持分类、优先级、指派、转派、协作、升级
- 支持首次响应和解决时效的 SLA 管控
- 支持内部备注与客户可见回复分离
- 支持事件留痕、统计分析和后续 AI 辅助

### 1.3 非目标

当前阶段不追求以下能力：

- 完整 ITSM 套件能力
- 复杂审批流引擎
- CMDB、变更管理、问题管理等重型模块
- 过早拆分成大量高度可配置的低代码模型

本阶段目标是先做“客服工单”，不是直接做“通用流程平台”。

## 2. 总体定位与业务边界

### 2.1 工单与会话的关系

建议采用“会话 + 工单”双轨模型。

职责划分如下：

- `Conversation`：负责实时沟通、即时消息、接待过程
- `Ticket`：负责正式处理、异步跟进、协作流转、SLA、闭环

强约束：

- 会话不是工单
- 工单也不是聊天备注
- 工单应当独立建模，不应把工单流程字段直接塞进 `Conversation`

### 2.2 推荐关系模型

- 一个 `Customer` 可以有多张工单
- 一条 `Conversation` 可以关联 0 到多张工单
- 一张 `Ticket` 可以关联一条来源会话
- 一张 `Ticket` 可以有多条评论和多条事件日志
- 一张 `Ticket` 可以有多个关注人、关联工单和 SLA 记录

### 2.3 典型使用场景

- 客服在会话中发现问题无法即时解决，一键转工单
- 客户通过门户或表单提交售后、投诉、技术问题
- 工单在团队内分配、转派、升级、协作
- 客服通过工单给客户异步回复处理进度
- 工单进入待客户反馈、待内部处理、已解决、已关闭
- 系统按 SLA 做预警和升级提醒

## 3. 设计原则

### 3.1 结构化优先

工单必须是结构化对象，而不是一段自由文本。最少要有：

- 标题
- 描述
- 分类
- 优先级
- 状态
- 当前处理人
- 当前处理团队
- 来源
- 客户

### 3.2 过程留痕

工单系统内所有关键动作都要有事件日志：

- 创建
- 更新
- 指派
- 转派
- 改状态
- 回复
- 内部备注
- 关闭
- 重开
- 超时
- 升级

### 3.3 协作与客户视角分离

工单内的信息至少分两类：

- 对客户可见的公开回复
- 仅内部可见的内部备注

不能把两类内容混在一起。

### 3.4 面向现有系统演进

设计必须复用当前已有能力，而不是再造一套平行系统：

- 客户复用 `Customer`
- 团队复用 `AgentTeam`
- 处理人复用 `User / AgentProfile`
- 标签优先复用现有标签体系
- 关联会话复用 `Conversation`

## 4. 角色与权限

### 4.1 角色

- 一线客服：创建、回复、转派、关闭工单
- 组长/主管：分配、升级、批量处理、监控 SLA
- 运营管理员：配置分类、字段、SLA、规则
- 客户：提交工单、查看进度、补充信息
- 系统/AI：自动摘要、自动分类建议、自动规则触发

### 4.2 权限建议

建议拆分以下权限点：

- 查看工单
- 创建工单
- 编辑工单
- 回复客户
- 添加内部备注
- 指派/转派工单
- 关闭工单
- 重开工单
- 批量处理工单
- 配置分类/字段/SLA/规则

数据权限建议支持：

- 仅本人负责
- 本团队
- 我关注的
- 我创建的
- 全部工单

## 5. 核心业务对象

### 5.1 Ticket

`Ticket` 是正式工单主档，用于承载一个待处理问题。

建议字段：

- `id`
- `ticketNo`
- `title`
- `description`
- `source`
- `channel`
- `customerId`
- `conversationId`
- `categoryId`
- `type`
- `priority`
- `severity`
- `status`
- `currentTeamId`
- `currentAssigneeId`
- `pendingReason`
- `closeReason`
- `resolutionCode`
- `resolutionSummary`
- `firstResponseAt`
- `resolvedAt`
- `closedAt`
- `dueAt`
- `nextReplyDeadlineAt`
- `resolveDeadlineAt`
- `reopenedCount`
- `customFieldsJson`
- `extraJson`

### 5.2 TicketComment

用于承载工单沟通内容。

评论类型建议：

- `public_reply`：对客户可见
- `internal_note`：仅内部可见
- `system_log`：系统生成内容

### 5.3 TicketEventLog

用于记录工单生命周期内的关键动作与字段变更。

建议记录：

- 事件类型
- 操作人
- 旧值
- 新值
- 补充说明
- 扩展 payload

### 5.4 TicketWatcher

用于记录关注工单的内部成员，支持关注列表和通知。

### 5.5 TicketSLARecord

用于记录每条 SLA 的目标值、计时状态和超时信息。

### 5.6 TicketRelation

用于记录工单之间的关系：

- `duplicate`
- `related`
- `parent`
- `child`

## 6. 状态机设计

### 6.1 推荐状态

第一版建议使用以下状态：

- `new`
- `open`
- `pending_customer`
- `pending_internal`
- `resolved`
- `closed`
- `cancelled`

### 6.2 状态语义

- `new`：新建待受理
- `open`：处理中
- `pending_customer`：等待客户补充信息
- `pending_internal`：等待内部团队处理
- `resolved`：已给出处理结果，待最终确认
- `closed`：正式关闭
- `cancelled`：作废或取消

### 6.3 推荐流转

- `new -> open`
- `open -> pending_customer`
- `open -> pending_internal`
- `pending_customer -> open`
- `pending_internal -> open`
- `open -> resolved`
- `resolved -> closed`
- `resolved -> open`
- `new/open/pending_* -> cancelled`

### 6.4 流转约束

- 关闭时必须填写关闭原因
- 重开时必须记录重开原因
- `resolved/closed/cancelled` 默认不允许普通编辑
- 状态流转必须写入事件日志

## 7. 优先级、严重度与分类

### 7.1 优先级

建议优先级：

- `low`
- `normal`
- `high`
- `urgent`

### 7.2 严重度

建议严重度：

- `minor`
- `major`
- `critical`

说明：

- 优先级表示处理顺序
- 严重度表示问题影响范围

### 7.3 分类

第一版建议先提供固定分类：

- 售前咨询
- 订单问题
- 退款售后
- 物流异常
- 账户问题
- 投诉建议
- 技术支持
- 其他

后续再扩展成可配置分类体系和表单体系。

## 8. 工单来源设计

建议统一支持以下创建来源：

- 会话转工单
- 后台手动建单
- 客户门户/表单建单
- 外部 API 建单
- 自动规则建单

其中第一阶段最关键的是“会话转工单”。

### 8.1 会话转工单

在会话页提供一键创建工单入口。

默认带入：

- 客户
- 关联会话
- 会话标题
- 最近消息摘要
- 当前标签
- 当前处理客服

建议支持 AI 辅助生成：

- 工单标题建议
- 问题摘要建议
- 分类建议
- 优先级建议

创建成功后建议支持：

- 回写会话系统消息
- 在会话右侧展示关联工单卡片
- 跳转工单详情

## 9. SLA 设计

### 9.1 第一版 SLA 范围

建议第一版仅做两条核心 SLA：

- 首次响应 SLA
- 最终解决 SLA

### 9.2 示例规则

- 普通工单：首次响应 30 分钟，解决 24 小时
- 高优工单：首次响应 10 分钟，解决 4 小时
- 紧急工单：首次响应 5 分钟，解决 2 小时

### 9.3 计时规则

- `new/open` 开始或继续计时
- `pending_customer` 暂停解决 SLA
- `resolved/closed/cancelled` 停止计时

### 9.4 页面展示

界面应明确展示：

- 当前生效的 SLA 目标
- 剩余时长
- 是否即将超时
- 是否已超时
- 超时多久

## 10. 自动化与规则

第一阶段只需要预留能力，不需要做复杂引擎。

后续建议支持三类规则：

- 创建时规则
- 更新时规则
- 定时规则

典型场景：

- 某来源自动分配到指定团队
- 命中特定分类自动提升优先级
- VIP 客户自动套用更严格 SLA
- 超过一定时间未处理自动升级
- 待客户回复超过一定时间自动提醒或关闭

## 11. 与现有系统的联动设计

### 11.1 与 Conversation 联动

建议打通以下动作：

- 会话一键转工单
- 工单详情查看来源会话摘要
- 工单状态变更可回写会话事件
- 工单关闭后可触发满意度邀请

### 11.2 与 Customer 联动

- 工单直接关联客户
- 工单详情展示客户主档和联系方式
- 客户维度支持查看历史工单列表

### 11.3 与 AI / KnowledgeBase 联动

后续建议支持：

- AI 生成工单标题和摘要
- AI 推荐分类和优先级
- 推荐相似历史工单
- 推荐知识库回复
- 根据处理阶段生成客服建议话术

## 12. 后端数据模型建议

## 12.1 核心表清单

MVP 第一阶段建议先落以下 6 张核心表：

- `tickets`
- `ticket_comments`
- `ticket_watchers`
- `ticket_event_logs`
- `ticket_sla_records`
- `ticket_relations`

### 12.2 tickets 表建议字段

```text
id bigint pk
ticket_no varchar(64) unique
title varchar(255)
description text
source varchar(50)
channel varchar(50)
customer_id bigint
conversation_id bigint
category_id bigint
type varchar(50)
priority int
severity int
status varchar(50)
current_team_id bigint
current_assignee_id bigint
pending_reason varchar(255)
close_reason varchar(255)
resolution_code varchar(100)
resolution_summary text
first_response_at datetime null
resolved_at datetime null
closed_at datetime null
due_at datetime null
next_reply_deadline_at datetime null
resolve_deadline_at datetime null
reopened_count int
custom_fields_json text
extra_json text
create_user_id bigint
create_user_name varchar(100)
update_user_id bigint
update_user_name varchar(100)
created_at datetime
updated_at datetime
status_flag int
```

说明：

- `ticket_no` 建议独立生成，例如 `TK202604020001`
- `custom_fields_json` 用于第一版快速支持扩展字段
- 时间字段全部使用兼容 `SQLite/MySQL` 的 `datetime`

### 12.3 ticket_comments 表建议字段

```text
id bigint pk
ticket_id bigint index
comment_type varchar(50) index
author_type varchar(30) index
author_id bigint index
content text
content_type varchar(30)
attachments_json text
created_at datetime
```

### 12.4 ticket_event_logs 表建议字段

```text
id bigint pk
ticket_id bigint index
event_type varchar(50) index
operator_type varchar(30) index
operator_id bigint index
old_value text
new_value text
content text
payload text
created_at datetime
```

### 12.5 ticket_watchers 表建议字段

```text
id bigint pk
ticket_id bigint index
user_id bigint index
created_at datetime
```

### 12.6 ticket_sla_records 表建议字段

```text
id bigint pk
ticket_id bigint index
sla_type varchar(50) index
target_minutes int
status varchar(30) index
started_at datetime null
paused_at datetime null
stopped_at datetime null
breached_at datetime null
elapsed_minutes int
created_at datetime
updated_at datetime
```

### 12.7 ticket_relations 表建议字段

```text
id bigint pk
ticket_id bigint index
related_ticket_id bigint index
relation_type varchar(30) index
created_at datetime
```

## 13. 后端分层设计

按当前项目分层规范，建议拆分如下：

### 13.1 models

- `Ticket`
- `TicketComment`
- `TicketWatcher`
- `TicketEventLog`
- `TicketSLARecord`
- `TicketRelation`

### 13.2 repositories

- `ticket_repository.go`
- `ticket_comment_repository.go`
- `ticket_watcher_repository.go`
- `ticket_event_log_repository.go`
- `ticket_sla_record_repository.go`
- `ticket_relation_repository.go`

职责：

- 只负责 CRUD、分页、查询、锁与统计
- 不负责业务流转与事务编排

### 13.3 services

建议新增：

- `ticket_service.go`
- `ticket_comment_service.go`
- `ticket_sla_service.go`
- `ticket_no_service.go`

职责：

- 创建工单
- 指派/转派
- 状态流转
- 回复与备注
- 关闭与重开
- SLA 驱动
- 会话转工单编排

### 13.4 builders

- `ticket_builder.go`
- `ticket_comment_builder.go`

职责：

- `Model -> ResponseDTO` 纯映射

### 13.5 controllers

- `internal/controllers/console/ticket_controller.go`

职责：

- 参数解析
- 权限校验
- 调 service
- 调 builders
- 返回统一 `JsonResult`

## 14. 后端接口设计

统一挂载到：

- `/api/console/ticket`

### 14.1 工单列表

- `AnyList()` -> `ANY /api/console/ticket/list`

建议查询参数：

- `keyword`
- `status`
- `priority`
- `categoryId`
- `currentAssigneeId`
- `currentTeamId`
- `customerId`
- `conversationId`
- `source`
- `tagId`
- `mine`
- `watching`
- `slaStatus`
- `createdAtStart`
- `createdAtEnd`

### 14.2 工单详情

- `GetBy(id int64)` -> `GET /api/console/ticket/{id}`

建议返回：

- 工单主档
- 评论列表摘要
- 关注人
- SLA 状态
- 关联客户摘要
- 关联会话摘要

### 14.3 创建工单

- `PostCreate()` -> `POST /api/console/ticket/create`

### 14.4 更新工单

- `PostUpdate()` -> `POST /api/console/ticket/update`

### 14.5 指派工单

- `PostAssign()` -> `POST /api/console/ticket/assign`

### 14.6 转派工单

- `PostTransfer()` -> `POST /api/console/ticket/transfer`

### 14.7 状态流转

- `PostChange_status()` -> `POST /api/console/ticket/change_status`

### 14.8 公开回复

- `PostReply()` -> `POST /api/console/ticket/reply`

### 14.9 内部备注

- `PostInternal_note()` -> `POST /api/console/ticket/internal_note`

### 14.10 关闭工单

- `PostClose()` -> `POST /api/console/ticket/close`

### 14.11 重开工单

- `PostReopen()` -> `POST /api/console/ticket/reopen`

### 14.12 关注/取消关注

- `PostWatch()` -> `POST /api/console/ticket/watch`
- `PostUnwatch()` -> `POST /api/console/ticket/unwatch`

### 14.13 评论列表

- `AnyComment_list()` -> `ANY /api/console/ticket/comment_list`

### 14.14 事件列表

- `AnyEvent_list()` -> `ANY /api/console/ticket/event_list`

### 14.15 会话转工单

- `PostCreate_from_conversation()` -> `POST /api/console/ticket/create_from_conversation`

建议保留独立接口，便于后续接入 AI 摘要和会话回写。

## 15. 核心服务规则

### 15.1 CreateTicket

职责：

- 校验客户和会话是否存在
- 生成工单号
- 创建工单主档
- 初始化 SLA 记录
- 写入事件日志
- 必要时回写会话事件

### 15.2 AssignTicket

职责：

- 校验团队和处理人是否合法
- 更新当前处理人和团队
- 写入事件日志
- 发送通知

### 15.3 ChangeTicketStatus

职责：

- 校验状态流转合法性
- 更新状态和时间字段
- 驱动 SLA 暂停/恢复/停止
- 写入事件日志

### 15.4 ReplyTicket

职责：

- 新增 `public_reply`
- 首次回复时回填 `firstResponseAt`
- 必要时将 `pending_customer` 拉回 `open`
- 写入事件日志

### 15.5 AddInternalNote

职责：

- 新增 `internal_note`
- 不影响客户可见内容
- 写入事件日志

### 15.6 CloseTicket / ReopenTicket

职责：

- 校验关闭或重开原因
- 更新 `closedAt / resolvedAt / reopenedCount`
- 停止或恢复 SLA
- 写入事件日志

## 16. 前端信息架构

### 16.1 导航位置

建议在工作台增加独立导航：

- `工作台 -> 工单`

### 16.2 页面清单

建议新增：

- `web/app/workspace/tickets/page.tsx`
- `web/app/workspace/tickets/[id]/page.tsx`
- `web/app/workspace/tickets/_components/edit.tsx`
- `web/lib/api/ticket.ts`

### 16.3 工单列表页

列表页建议包含：

- 顶部筛选栏
- 工单表格
- 批量操作区
- 可保存视图

推荐筛选项：

- 关键词
- 状态
- 优先级
- 分类
- 团队
- 处理人
- 来源
- SLA 状态
- 时间范围

### 16.4 工单详情页

建议采用双栏或三栏布局。

主区域：

- 标题与状态条
- 回复区
- 评论时间线
- 内部备注时间线
- 事件记录

侧栏：

- 基础信息
- 客户信息
- 指派信息
- SLA 卡片
- 标签
- 关联会话
- 关联工单
- 关注人

### 16.5 会话页联动

建议在现有会话工作台增加：

- `创建工单`
- `查看关联工单`
- `工单状态摘要`

推荐入口位置：

- 会话头部操作区
- 右侧 `conversation-info-panel`

## 17. 前端组件建议

建议拆分如下业务组件：

- `ticket-filter-bar.tsx`
- `ticket-list-table.tsx`
- `ticket-status-badge.tsx`
- `ticket-priority-badge.tsx`
- `ticket-edit-dialog.tsx`
- `ticket-detail-header.tsx`
- `ticket-comment-timeline.tsx`
- `ticket-reply-editor.tsx`
- `ticket-side-panel.tsx`
- `ticket-sla-card.tsx`
- `ticket-related-conversation-card.tsx`

## 18. MVP 范围与迭代计划

### 18.1 第一阶段：可用版

目标：先跑通“会话转工单 + 正式处理闭环”。

范围：

- 工单主档
- 评论
- 事件日志
- 列表/详情
- 新建/编辑
- 指派/转派
- 状态流转
- 公开回复/内部备注
- 会话转工单
- 基础 SLA

### 18.2 第二阶段：运营版

目标：让主管和运营可配置、可监控。

范围：

- 分类和表单配置
- 保存视图
- 批量操作
- 关注人
- 超时预警
- 基础报表
- 自动化规则基础版

### 18.3 第三阶段：增强版

目标：做出系统差异化。

范围：

- AI 自动摘要与分类建议
- 相似工单推荐
- 知识库回复推荐
- 自动建单规则
- 客户门户提交与查询
- 满意度调查
- 合并/拆分/重复单治理

## 19. 建议的实现优先级

如果按工程落地顺序推进，建议优先完成以下事项：

1. 定义工单相关枚举和模型
2. 落 `tickets / ticket_comments / ticket_event_logs / ticket_sla_records`
3. 完成 `ticket_service` 与 `ticket_controller`
4. 先做工单列表页、详情页、创建弹窗
5. 在会话页接入“转工单”入口
6. 再扩展关注、关联工单、报表和自动化

## 20. 结论

本系统最适合的工单方案不是把工单做成“会话的一个状态”，而是做成与会话协同的正式处理对象：

- `Conversation` 承接实时接待
- `Ticket` 承接正式处理闭环
- 两者通过客户、标签、会话回链、事件日志和 AI 辅助能力打通

这样既能保持当前 IM 架构的简洁，也能逐步补齐成熟客服系统应有的处理、协作、SLA 与运营能力。
