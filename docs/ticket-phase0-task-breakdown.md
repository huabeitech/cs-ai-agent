# 工单 Phase 0 任务拆解

本文档是 [ticket-remediation-roadmap.md](/Users/gaoyoubo/code/gaoyoubo/cs-agent/docs/ticket-remediation-roadmap.md) 中 `Phase 0：稳定性整改` 的可执行拆解版本。

目标是让研发可以直接据此开工，不再需要把“整改方向”再翻译一轮。

## 1. Phase 0 目标

本阶段只解决“系统可靠性”问题，不扩新业务边界。

交付目标：

1. 工单号生成稳定
2. SLA deadline 主链路正确
3. mention 索引设计修正
4. 事务边界符合项目规范
5. 批量操作语义清晰
6. 核心 service 有回归测试

## 2. 范围界定

### 2.1 本期包含

- `ticketNo` 生成重构
- SLA 聚合字段回写
- `TicketMention` 唯一索引修复
- 事务内 DB 访问整改
- 批量指派 / 批量状态变更语义优化
- ticket service 测试补齐

### 2.2 本期不包含

- 保存视图
- 宏 / 快捷回复
- 自动化规则
- 客户门户
- AI 摘要 / AI 回复
- 大范围前端 UI 改版

## 3. 任务拆分总览

建议拆成 6 个子任务：

1. T1：ticketNo 重构
2. T2：SLA deadline 主链路修复
3. T3：TicketMention 索引修复
4. T4：事务边界整改
5. T5：批量操作语义整改
6. T6：测试补齐

建议顺序：

- 先做 T1、T3
- 再做 T2、T4
- 然后做 T5
- 最后做 T6 和联调

## 4. T1：ticketNo 重构

### 4.1 背景

当前实现：

- 代码位置：`internal/services/ticket_service.go`
- 方法：`nextTicketNo()`
- 规则：`TK + yyyyMMddHHmmss + 毫秒`

问题：

- 高并发下存在重复风险
- 依赖当前机器时间和毫秒粒度，不具备生产级稳定性

### 4.2 目标

生成规则要满足：

- 唯一
- 可读
- 可排序
- 同时兼容 SQLite 和 MySQL

### 4.3 建议方案

采用“日期前缀 + 当日递增序号”。

格式示例：

```text
TK2026040300001
TK2026040300002
```

### 4.4 推荐实现方式

方案优先级：

1. 推荐：基于 `SystemConfig` 做号段序列
2. 备选：新增独立 `TicketNoSequence` 表

为了减少模型变更，本期建议优先用 `SystemConfig`。

配置示例：

- `configKey = ticket:no:date:20260403`
- `configValue = 123`

### 4.5 后端改动点

新增文件：

- `internal/services/ticket_no_service.go`

建议方法：

```go
type ticketNoService struct{}

func (s *ticketNoService) Next(tx *gorm.DB, now time.Time) (string, error)
```

改动文件：

- `internal/services/ticket_service.go`
- `internal/services/system_config_service.go`
- `internal/repositories/system_config_repository.go`

### 4.6 实现要求

- 必须在事务内递增
- 同一天内连续编号
- 幂等错误信息明确
- 不能继续依赖 `Nanosecond()/1e6`

### 4.7 验收

1. 并发建单 100 次不出现重复
2. 工单号格式统一
3. SQLite / MySQL 均可生成

## 5. T2：SLA deadline 主链路修复

### 5.1 背景

当前系统中：

- `TicketSLARecord` 会创建
- 但 `Ticket.ResolveDeadlineAt`
- `Ticket.NextReplyDeadlineAt`

并没有形成稳定回写和刷新机制。

结果：

- 列表页超时筛选不准确
- 风险页高风险判断不准确
- summary 统计不准确

### 5.2 目标

建立统一 SLA 聚合刷新机制：

- 记录表负责原始 SLA 状态
- 主表负责聚合后的 deadline 字段

### 5.3 统一规则

#### A. `nextReplyDeadlineAt`

来源：

- first response SLA 未完成时，取 first response SLA deadline
- first response 已完成后置空

#### B. `resolveDeadlineAt`

来源：

- resolution SLA 处于 running 时，按 `startedAt + remainingMinutes` 计算
- resolution SLA paused / completed / breached 时也要有明确聚合策略

本期建议：

- running: 正常计算
- paused: 保留按当前剩余时间换算出的 deadline，或置空后二期再细化
- completed: 保留结束前 deadline 供统计，或直接保留 breach/stop 信息

为避免影响现有页面，本期建议主表中：

- `resolveDeadlineAt` 始终表示“当前活跃解决时限截止时间”
- 已完成 / 已关闭后可以保留历史值，不参与 active 筛选

### 5.4 推荐新增方法

建议新增到：

- `internal/services/ticket_service.go`

新增：

```go
func (s *ticketService) refreshTicketSLAFields(tx *gorm.DB, ticketID int64, now time.Time) error
func calcSLADeadline(record *models.TicketSLARecord, now time.Time) *time.Time
```

### 5.5 触发点

以下动作后必须刷新主表 SLA 聚合字段：

- `CreateTicket`
- `ReplyTicket`
- `ChangeStatus`
- `CloseTicket`
- `ReopenTicket`
- `ScanAndMarkBreachedSLAs`

### 5.6 后端改动文件

- `internal/services/ticket_service.go`
- `internal/repositories/ticket_sla_record_repository.go`
- `internal/repositories/ticket_repository.go`

如果需要更清晰，也可以新增：

- `internal/services/ticket_sla_service.go`

但本期为了控制范围，也可先放在 `ticket_service.go` 内完成。

### 5.7 验收

1. 新建工单后主表有正确 deadline
2. 首次回复后 `nextReplyDeadlineAt` 清空
3. 状态变更为 `pending_customer` 时，解决 SLA 正确暂停
4. 重开工单后，解决 SLA 正确恢复
5. 风险页和列表超时筛选结果一致

## 6. T3：TicketMention 索引修复

### 6.1 背景

当前模型：

- 文件：`internal/models/models.go`
- 模型：`TicketMention`
- 问题：`MentionedUserID` 被独立设置为唯一索引

这会导致：

- 同一用户无法在多个工单被提及
- 同一用户无法在同一工单不同备注中重复被提及

### 6.2 目标

把唯一约束修正为“同一条备注下同一用户不能重复插入”，而不是“全局唯一”。

### 6.3 建议修改

改为：

```go
TicketID        int64 `gorm:"type:bigint;not null;index;uniqueIndex:uk_ticket_mention"`
CommentID       int64 `gorm:"type:bigint;not null;index;uniqueIndex:uk_ticket_mention"`
MentionedUserID int64 `gorm:"type:bigint;not null;index;uniqueIndex:uk_ticket_mention"`
```

### 6.4 数据迁移

由于 `AutoMigrate` 对历史唯一索引调整不一定可靠，本期需要：

1. 通过 `AutoMigrate` 更新模型
2. 在 `internal/migration` 中补一条幂等 migration
3. migration 中显式删除旧唯一索引并创建新组合索引

### 6.5 改动文件

- `internal/models/models.go`
- `internal/migration/*` 新增 DML/DDL 修复逻辑

### 6.6 验收

1. 同一用户可在多个工单被提及
2. 同一用户可在多个 comment 被提及
3. 同一 comment 内重复 mention 仍被拦截

## 7. T4：事务边界整改

### 7.1 背景

当前主要问题：

- 事务内又通过 service 走 `sqls.DB()`
- 一致性要求强的读写没有统一使用 `ctx.Tx`

与项目规范冲突：

- 事务内所有 DB 操作都必须走 `ctx.Tx`

### 7.2 目标

把 ticket 主链路中所有事务内读写统一收敛到：

- repository 层
- `db *gorm.DB` 参数

### 7.3 重点整改方法

#### A. `AddInternalNote`

当前问题：

- 事务内使用 `TicketCollaboratorService.FindOne`
- 事务内使用 `TicketWatcherService.FindOne`
- 事务内使用 `TicketMentionService.FindOne`

整改方案：

- 改为 repository 直接用 `ctx.Tx`
- 或为 repository/service 增加带 `db` 参数的方法

#### B. SLA 相关方法

以下方法当前在事务内仍间接走默认 DB：

- `completeFirstResponseSLA`
- `pauseResolutionSLA`
- `resumeResolutionSLA`
- `completeResolutionSLA`

整改方案：

- 新增 repository 查询方法：
  - `TakeByTicketIDAndType(db *gorm.DB, ticketID int64, slaType enums.TicketSLAType) *models.TicketSLARecord`

#### C. 关联和批量相关路径复查

复查：

- `CreateTicket`
- `ChangeStatus`
- `ReopenTicket`
- `AddCollaborator`
- `AddRelation`

### 7.4 改动文件

建议涉及：

- `internal/services/ticket_service.go`
- `internal/services/ticket_relation_service.go`
- `internal/repositories/ticket_sla_record_repository.go`
- `internal/repositories/ticket_watcher_repository.go`
- `internal/repositories/ticket_collaborator_repository.go`
- `internal/repositories/ticket_mention_repository.go`

### 7.5 验收

1. 事务内不再调用默认 DB service 方法
2. 一致性敏感的读写都走 `ctx.Tx`
3. 代码 review 可明确看出事务边界

## 8. T5：批量操作语义整改

### 8.1 背景

当前实现：

- `BatchAssignTickets`
- `BatchChangeStatus`

都是逐条执行，一旦中途失败会半成功。

### 8.2 目标

让批量操作具备明确语义。

### 8.3 本期推荐策略

优先使用：

- 全量事务
- 任意一条失败则整体失败

原因：

- 与现有后台操作习惯更一致
- 前端更容易理解
- 实现复杂度低于部分成功回执

### 8.4 推荐实现

新增 service 方法：

```go
func (s *ticketService) BatchAssignTickets(req request.BatchAssignTicketRequest, operator *dto.AuthPrincipal) error
func (s *ticketService) BatchChangeStatus(req request.BatchChangeTicketStatusRequest, operator *dto.AuthPrincipal) error
```

但内部改为：

- 先预校验所有工单
- 再统一开启事务写入

注意：

- 如果每张工单都需要事件日志，必须同事务写入

### 8.5 前端影响

前端当前可以不改接口结构，只需要接受：

- 失败时整体失败
- 成功时整体成功

### 8.6 验收

1. 批量操作中途失败不会留下半完成状态
2. 事件日志与主操作保持一致
3. 前端提示语与实际结果一致

## 9. T6：测试补齐

### 9.1 目标

补足工单最关键路径的 service 测试。

### 9.2 推荐测试文件

- `internal/services/ticket_service_test.go`
- `internal/services/ticket_relation_service_test.go`

### 9.3 最低测试集

#### A. 建单

- 手动建单成功
- 会话转工单成功
- 非法分类创建失败
- 非法处理人创建失败

#### B. 工单号

- 连续创建工单号递增
- 并发创建不重复

#### C. SLA

- 创建工单生成 first response / resolution 记录
- 创建工单后主表 deadline 正确
- 首次回复后首响 SLA 完成
- `pending_customer` 后解决 SLA 暂停
- 重开后解决 SLA 恢复
- SLA 扫描后 breach 状态正确

#### D. 状态流转

- 合法流转成功
- 非法流转失败
- 有未关闭子工单时禁止关闭

#### E. mention / 协作

- 内部备注可提及多人
- 提及后自动添加 collaborator / watcher
- 同一用户多工单提及不冲突

#### F. 批量操作

- 批量指派整体成功
- 一条失败时整体回滚
- 批量状态变更整体成功
- 一条失败时整体回滚

### 9.4 测试要求

- 尽量使用现有 DB 测试初始化方式
- 测试断言不要只断接口成功，要断表状态
- 至少检查：
  - ticket 主表
  - ticket_sla_records
  - ticket_event_logs
  - ticket_watchers
  - ticket_mentions

## 10. 文件级实施清单

### 必改文件

- `internal/services/ticket_service.go`
- `internal/models/models.go`
- `internal/repositories/ticket_repository.go`
- `internal/repositories/ticket_sla_record_repository.go`

### 大概率新增文件

- `internal/services/ticket_no_service.go`
- `internal/services/ticket_service_test.go`

### 可能新增 / 调整文件

- `internal/migration/00000x_fix_ticket_mention_index.go`
- `internal/repositories/ticket_mention_repository.go`
- `internal/repositories/ticket_watcher_repository.go`
- `internal/repositories/ticket_collaborator_repository.go`
- `internal/services/ticket_relation_service_test.go`

## 11. 前端联动清单

本阶段前端不做大改，只做必要联动确认：

1. 列表页 `overdue` 和风险标签是否依赖真实 deadline
2. 风险页是否仍然从本地 200 条内筛高风险
3. 批量操作整体失败时提示是否明确

建议前端本阶段只做：

- 联调验证
- 文案修正
- 必要的小逻辑调整

## 12. 开发顺序建议

### Day 1-2

- T1 ticketNo
- T3 mention 索引

### Day 3-4

- T2 SLA deadline 主链路
- T4 事务整改

### Day 5

- T5 批量操作

### Day 6-7

- T6 测试补齐
- 联调
- 回归验证

## 13. 提交与验收建议

建议拆成 3 个 PR：

### PR1：底层模型与号段

- ticketNo
- mention 索引
- migration

### PR2：SLA 与事务整改

- deadline 回写
- SLA 聚合
- 事务收敛
- 批量操作语义

### PR3：测试与联调修正

- service tests
- 前端适配
- 文案和边界修正

这样更利于 review，也更容易回滚。

## 14. Phase 0 完成定义

满足以下条件即可认为本阶段完成：

1. 工单号生成方案替换完成
2. SLA deadline 主链路准确
3. mention 索引问题修复
4. 事务内不再混用 `sqls.DB()`
5. 批量操作语义明确并通过联调
6. 核心 ticket service 测试通过
7. Go 代码通过 `gofmt`
8. 前端至少通过 `cd web && pnpm typecheck`

## 15. 下一阶段衔接

Phase 0 完成后，下一阶段应立刻进入：

- 列表聚合去 N+1
- 风险页后端化
- 数据权限第一版

也就是 [ticket-remediation-roadmap.md](/Users/gaoyoubo/code/gaoyoubo/cs-agent/docs/ticket-remediation-roadmap.md) 中的 `Phase 1`。
