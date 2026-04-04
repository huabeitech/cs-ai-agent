# 工单系统整改与实施计划

本文档基于当前仓库内工单模块实现现状形成，目标不是重复已有设计，而是回答三个更直接的问题：

- 当前工单功能离“可长期稳定使用”的成熟系统还差什么
- 下一阶段应该先修什么，后做什么
- 每个阶段后端、前端、测试、数据变更分别要做什么

本文档作为以下文档的执行补充：

- [ticket-system-design.md](/Users/gaoyoubo/code/gaoyoubo/cs-agent/docs/ticket-system-design.md)
- [ticket-backend-design.md](/Users/gaoyoubo/code/gaoyoubo/cs-agent/docs/ticket-backend-design.md)
- [ticket-frontend-design.md](/Users/gaoyoubo/code/gaoyoubo/cs-agent/docs/ticket-frontend-design.md)
- [ticket-workbench-redesign.md](/Users/gaoyoubo/code/gaoyoubo/cs-agent/docs/ticket-workbench-redesign.md)
- [ticket-iteration-plan.md](/Users/gaoyoubo/code/gaoyoubo/cs-agent/docs/ticket-iteration-plan.md)

## 1. 当前状态判断

当前工单模块已经具备以下能力：

- 工单主表、评论、事件、关注人、协作人、提及、SLA、关联工单等基础模型
- 创建工单、会话转工单
- 列表、详情
- 指派、批量指派
- 状态流转、批量改状态
- 回复客户、内部备注
- 关闭、重开
- 风险概览页和配置页雏形

但从系统成熟度看，当前状态仍属于：

- 功能面已接近 MVP 完整
- 可演示、可试用
- 但距离“高频生产可用”还有明显差距

主要原因不在于“页面还不够多”，而在于以下底层问题尚未补齐：

1. SLA 主链路不闭环
2. 工单号生成方案不可靠
3. 提及索引设计存在冲突风险
4. 事务边界不够严格
5. 列表和详情存在明显 N+1 查询
6. 批量操作不是原子执行
7. 风险页存在前端本地筛选导致的数据遗漏
8. 缺少工单核心链路测试

因此，接下来的工作不能直接跳到 AI、报表、自动化，而应先做“稳定化整改”，再进入“运营化建设”。

## 2. 对标成熟工单系统后的能力目标

结合 Zendesk、Jira Service Management、Freshdesk 这类成熟方案，可以抽象出我们下一阶段需要补齐的 6 类核心能力：

### 2.1 队列化工作台

不是普通后台列表，而是以“谁现在该处理什么”为中心的队列。

需要具备：

- 固定视图
- 保存视图
- 多条件筛选
- SLA 风险排序
- 批量动作
- 主管干预入口

### 2.2 结构化工单入口

不是一张自由文本记录，而是标准化服务对象。

需要具备：

- 分类
- 解决码
- SLA 配置
- 请求类型
- 结构化字段
- 来源渠道标准化

### 2.3 SLA 运营闭环

不是“有个 SLA 表”，而是：

- 可计算
- 可排序
- 可预警
- 可升级
- 可报表

### 2.4 自动化与快捷动作

成熟系统普遍有：

- 规则触发
- 宏 / 快捷回复
- 批量动作
- 提醒 / 催办 / 升级

### 2.5 客户入口与知识联动

真正成熟的工单系统通常不只服务坐席，还服务客户和自助入口：

- 门户提交工单
- 工单进度查询
- 知识库建议
- 相似问题分流

### 2.6 分析与治理

成熟系统最终都会沉淀到：

- 首响达成率
- 解决时长
- 队列积压
- 重开率
- 分类分布
- 人员负载

## 3. 整改总原则

### 3.1 先修可靠性，再加复杂能力

必须先修：

- deadline 不准确
- 事务不严谨
- 索引错误
- 并发唯一键风险
- 缺测试

之后再推进：

- 宏
- 自动化
- 报表
- AI

### 3.2 以现有分层为基础，不推翻现有结构

当前项目的 `models -> repositories -> services -> controllers -> builders` 分层是正确方向，后续整改应继续遵循，不建议另起一套。

### 3.3 每个阶段都要可独立上线

每个阶段结束时都应满足：

- 有明确交付
- 有可验收结果
- 不依赖下阶段才能生效

## 4. 阶段规划总览

建议分为 5 个阶段推进：

1. Phase 0：稳定性整改
2. Phase 1：SLA 与队列后端重构
3. Phase 2：工作台增强
4. Phase 3：自动化与运营能力
5. Phase 4：客户闭环与 AI 辅助

建议时间：

- Phase 0：1 周
- Phase 1：1-2 周
- Phase 2：2 周
- Phase 3：2 周
- Phase 4：2-3 周

总计约 8-10 周，可按团队资源拆分并行。

## 5. Phase 0：稳定性整改

### 5.1 目标

把当前工单系统从“功能存在”修到“主链路可信”。

### 5.2 范围

本期只做底层可靠性，不扩新业务范围。

### 5.3 后端任务

#### A. 修复 ticketNo 生成

当前问题：

- 仅靠时间字符串 + 毫秒尾数
- 高并发存在冲突风险

建议方案：

- 新增 `ticket_no_sequence` 风格的系统配置或独立号段表
- 生成规则改为：日期前缀 + 自增序号
- 生成逻辑必须在事务内完成
- 确保 SQLite / MySQL 都可用

推荐实现：

- 新增 `internal/services/ticket_no_service.go`
- 对外暴露 `Next(tx *gorm.DB) (string, error)`
- `CreateTicket` 内统一调用

#### B. 修复 SLA 主表字段回写

当前问题：

- `TicketSLARecord` 已创建
- 但 `ticket.resolve_deadline_at` / `ticket.next_reply_deadline_at` 未形成稳定回写
- 风险页和列表排序依赖主表字段，导致数据不可信

建议方案：

- 建单时初始化 SLA 后，同步更新工单主表 deadline 字段
- 状态流转时同步调整主表 deadline
- 回复、暂停、恢复、完成时同步刷新主表聚合字段

建议新增 service 方法：

- `refreshTicketSLAFields(tx *gorm.DB, ticketID int64) error`

写操作完成后统一调用，避免散落更新。

#### C. 修复 TicketMention 唯一索引

当前问题：

- `mentioned_user_id` 独立唯一，设计错误

建议方案：

- 唯一键改为：
  - `ticket_id + comment_id + mentioned_user_id`

并通过 `AutoMigrate` 生效后补充一次幂等索引修正 migration。

#### D. 修复事务边界

当前问题：

- 事务内仍通过 service 使用 `sqls.DB()` 查询
- 事务读写没有统一走 `ctx.Tx`

建议整改：

- repository 统一提供 `db *gorm.DB` 参数的读方法
- 事务内所有存在一致性要求的查询全部改走 `ctx.Tx`
- service 内禁止“事务里再调默认 DB service 方法”

重点整改范围：

- `AddInternalNote`
- `completeFirstResponseSLA`
- `pauseResolutionSLA`
- `resumeResolutionSLA`
- `completeResolutionSLA`
- 批量操作链路

#### E. 批量接口语义调整

当前问题：

- 逐条 for-loop 执行
- 中间失败会半成功

建议方案：

- 第一阶段至少提供两种策略之一：
  - 全量原子事务
  - 明细结果返回 `{successIds, failedItems}`

建议优先：

- 指派和状态批量操作支持“全量成功或整体失败”

#### F. 补 service 测试

至少覆盖：

- 建单
- 会话转工单
- 状态流转
- 关闭 / 重开
- SLA 暂停 / 恢复 / 完成
- 内部备注提及
- 批量操作

测试目录建议：

- `internal/services/ticket_service_test.go`
- `internal/services/ticket_relation_service_test.go`

### 5.4 前端任务

本期前端只做必要适配：

- 风险提示文案改为基于真实后端字段
- 批量接口若改返回结构，前端同步处理
- 在列表和详情页去掉对错误 SLA 数据的误导性展示

### 5.5 验收标准

1. 高并发建单不再出现 ticketNo 冲突
2. 列表/风险页中的 SLA deadline 与记录表一致
3. 同一用户可在不同工单、不同备注中被反复提及
4. 批量操作语义明确
5. 核心 service 测试通过

## 6. Phase 1：SLA 与队列后端重构

### 6.1 目标

让队列和风险能力真正建立在后端准确查询之上。

### 6.2 后端任务

#### A. 列表聚合查询改造

当前问题：

- builder 内大量逐条查询
- 列表接口存在 N+1

建议方案：

- service 层提供聚合查询结果结构
- 预加载以下信息：
  - 分类名
  - 解决码名
  - 当前处理人名
  - 当前团队名
  - customer 摘要
  - watchedByMe
  - SLA 摘要

建议新增：

- `TicketListAggregate`
- `FindTicketPageAggregate(...)`

builder 只做映射，不再自行查数据库。

#### B. 风险查询后端化

当前问题：

- 风险页从第一页 200 条里前端再筛
- 数据量大时会漏

建议增加专用接口：

- `GET /api/console/ticket/risk_overview`
- `GET /api/console/ticket/risk_list`

`risk_list` 支持类型：

- overdue
- high_risk
- unassigned
- pending_internal
- pending_customer_stale
- active_without_deadline

#### C. 增强排序能力

列表建议支持：

- `updated_at desc` 默认
- `priority desc`
- `created_at desc`
- `resolve_deadline_at asc`
- `sla_risk desc`

#### D. 数据权限第一版

建议先做 4 个范围：

- `mine`
- `my_team`
- `watching`
- `all`

权限收敛在 service 层，不放在前端。

### 6.3 前端任务

- 列表页改用后端返回的聚合字段
- 风险页改用 `risk_list`
- 支持按 SLA 风险排序
- 快捷视图和筛选项与后端语义对齐

### 6.4 验收标准

1. 列表接口不再明显 N+1
2. 风险页数据完整，不依赖前端本地筛选
3. 列表支持按 SLA 风险排序
4. 数据权限开始生效

## 7. Phase 2：工作台增强

### 7.1 目标

让工单列表页和详情页成为真正高频处理工作台。

### 7.2 后端任务

#### A. 保存视图

新增模型：

- `TicketSavedView`

建议字段：

- `id`
- `name`
- `owner_user_id`
- `scope`
- `filters_json`
- `sort_by`
- `is_default`
- `sort_no`

接口：

- `/api/console/ticket-saved-view/list`
- `/create`
- `/update`
- `/delete`

#### B. 宏 / 快捷动作第一版

新增模型：

- `TicketMacro`

建议字段：

- 名称
- 作用范围
- 回复模板
- 内部备注模板
- 默认状态
- 默认指派团队 / 处理人
- 是否共享

#### C. 详情聚合增强

详情接口增加：

- 创建人信息
- 来源会话摘要
- 最近消息摘要
- 关联工单进度
- 最近一次状态变更时间
- 首响剩余 / 解决剩余时间

### 7.3 前端任务

#### A. 列表页

- 固定视图 + 保存视图
- 表格列可裁剪
- 批量操作条优化
- 行级快捷动作轻量化
- 支持“处理下一张”工作流

#### B. 详情页

- 双栏布局进一步组件化
- 时间线可按“评论/事件/全部”切换
- 侧栏突出 SLA 和来源上下文
- 宏应用入口
- 关联工单快捷创建

### 7.4 验收标准

1. 客服可通过保存视图管理个人队列
2. 主管可通过风险视图和待分配视图进行干预
3. 常见处理动作可通过宏减少重复操作
4. 详情页可支持连续处理

## 8. Phase 3：自动化与运营能力

### 8.1 目标

让工单系统从“人工处理工具”升级成“半自动运营系统”。

### 8.2 后端任务

#### A. 自动化规则第一版

新增模型：

- `TicketAutomationRule`

第一阶段不做复杂 DSL，先做固定触发器：

- 建单后
- 指派后
- 状态变更后
- 超过首响时限
- 超过解决时限
- 待客户反馈超过 N 小时
- 未分配超过 N 分钟

支持动作：

- 自动加关注人
- 自动指派团队
- 自动改状态
- 自动写内部备注
- 自动追加标签 / 分类
- 发送通知

#### B. 主管运营视图

增加统计接口：

- SLA 达成率
- 当前队列积压
- 人员负载
- 超时趋势
- 分类分布
- 重开率

### 8.3 前端任务

- 自动化规则配置页
- 运营概览页
- 团队负载面板
- 超时队列处理建议

### 8.4 验收标准

1. 可配置基础自动化规则
2. 主管可以查看 SLA 和积压趋势
3. 超时工单具备自动提醒或升级能力

## 9. Phase 4：客户闭环与 AI 辅助

### 9.1 目标

让工单不只停留在内部处理，还能与外部客户和 AI 能力形成闭环。

### 9.2 客户侧能力

建议按轻量级顺序推进：

#### A. 外部工单入口

- 客户提交工单
- 客户查看自己工单
- 客户查看公开回复

#### B. 会话与工单双向联动增强

- 工单详情中查看来源会话摘要
- 会话页查看已关联工单
- 会话和工单共享上下文摘要

### 9.3 AI 能力

建议分三步做：

#### Step 1：AI 摘要

- 会话转工单时自动生成标题和描述建议
- 长工单自动生成处理摘要

#### Step 2：AI 推荐

- 推荐分类
- 推荐解决码
- 推荐处理团队
- 推荐相似工单
- 推荐知识库文章

#### Step 3：AI 草稿

- 回复草稿
- 内部备注草稿
- 升级理由草稿

AI 第一阶段只做建议，不直接自动执行。

### 9.4 验收标准

1. 客户可通过统一入口提交和查看工单
2. 客服可看到 AI 摘要和推荐
3. AI 不直接改状态，只辅助决策

## 10. 建议的数据模型增量

建议后续新增以下模型：

- `TicketSavedView`
- `TicketMacro`
- `TicketAutomationRule`
- `TicketAutomationLog`
- `TicketPortalAccess` 或客户侧查询凭据模型
- `TicketTag` 与 `TicketTagRelation`（如后续要做更强运营筛选）

不建议现在立刻新增：

- 通用动态表单引擎
- 超复杂审批流引擎
- 全功能 ITSM CMDB 体系

## 11. 推荐目录落地

后端新增建议：

```text
internal/
├── builders/
│   ├── ticket_saved_view_builder.go
│   ├── ticket_macro_builder.go
│   └── ticket_risk_builder.go
├── controllers/
│   └── console/
│       ├── ticket_saved_view_controller.go
│       ├── ticket_macro_controller.go
│       └── ticket_automation_rule_controller.go
├── repositories/
│   ├── ticket_saved_view_repository.go
│   ├── ticket_macro_repository.go
│   ├── ticket_automation_rule_repository.go
│   └── ticket_automation_log_repository.go
├── services/
│   ├── ticket_no_service.go
│   ├── ticket_macro_service.go
│   ├── ticket_saved_view_service.go
│   ├── ticket_automation_service.go
│   └── ticket_risk_service.go
```

前端新增建议：

```text
web/app/(console)/
├── ticket-macros/
├── ticket-views/
├── ticket-automation-rules/
└── ticket-reports/
```

API service 新增建议：

```text
web/lib/api/
├── ticket-macro.ts
├── ticket-view.ts
├── ticket-report.ts
└── ticket-automation.ts
```

## 12. 近期最优先任务清单

如果只看接下来 2 周，建议严格按下面顺序推进：

### P0

- 修 ticketNo
- 修 TicketMention 索引
- 修 SLA deadline 回写
- 修事务边界
- 补 ticket service 测试

### P1

- 改造列表聚合查询，去掉 N+1
- 风险页后端化
- 批量操作语义明确化

### P2

- 保存视图
- 宏 / 快捷回复
- 数据权限第一版

不建议在 P0/P1 未完成前直接推进：

- AI 自动回复
- 复杂自动化规则
- 客户门户
- 报表中心大而全版本

## 13. 人力建议

如果按 2-3 人小团队推进，建议分工如下：

### 后端负责人

- 负责 Phase 0 和 Phase 1
- 重点守住：
  - deadline
  - 事务
  - 查询聚合
  - 测试

### 前端负责人

- 负责工作台与风险页
- 重点守住：
  - 队列体验
  - 详情处理效率
  - 筛选和保存视图

### 兼职产品 / 运营接口人

- 负责定义：
  - 分类体系
  - 解决码体系
  - SLA 规则
  - 宏模板
  - 自动化规则优先级

## 14. 最终判断

当前工单系统最值得肯定的地方是：

- 方向是对的
- 模型基础不差
- 前后端雏形已经齐了

当前最需要提升的地方是：

- 不要再把重点放在“继续补页面”
- 要先把底层可靠性、SLA 主链路、批量语义和查询性能补齐

只要 Phase 0 和 Phase 1 做扎实，这套工单系统就会从“功能模块”升级为“可作为主工作台的业务系统”；后面的宏、自动化、报表和 AI 才真正有价值。
