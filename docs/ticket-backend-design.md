# 客服工单后端详细设计

本文档用于在 [ticket-system-design.md](/Users/gaoyoubo/code/gaoyoubo/cs-agent/docs/ticket-system-design.md) 的基础上，继续细化工单模块的后端落地方案，重点覆盖：

- 枚举设计
- Model 设计
- Repository 设计
- Service 设计
- Controller 与接口设计
- Request / Response DTO 设计
- 事务边界与事件留痕

本文档遵循项目现有分层规范：

- `models -> repositories -> services -> controllers -> builders`
- controller 不直接调用 repository
- controller 入参使用 request DTO
- controller 出参使用 response DTO
- 写操作事务边界放在 service

## 1. 落地范围

第一阶段仅覆盖工单 MVP 的核心链路：

- 创建工单
- 会话转工单
- 工单列表
- 工单详情
- 更新工单
- 指派/转派
- 状态流转
- 公开回复
- 内部备注
- 关闭/重开
- 基础 SLA
- 事件留痕

当前阶段不覆盖：

- 客户门户工单
- 自动化规则引擎
- 动态表单配置中心
- 工单合并/拆分
- 复杂审批流

## 2. 目录建议

建议新增以下文件：

```text
internal/
├── builders/
│   ├── ticket_builder.go
│   └── ticket_comment_builder.go
├── controllers/
│   └── console/
│       └── ticket_controller.go
├── models/
│   └── models.go              // 新增 Ticket 相关 model
├── pkg/
│   ├── dto/
│   │   ├── request/
│   │   │   └── ticket_requests.go
│   │   └── response/
│   │       └── ticket_responses.go
│   └── enums/
│       └── enums.go           // 新增 Ticket 相关枚举
├── repositories/
│   ├── ticket_repository.go
│   ├── ticket_comment_repository.go
│   ├── ticket_event_log_repository.go
│   ├── ticket_sla_record_repository.go
│   ├── ticket_relation_repository.go
│   └── ticket_watcher_repository.go
└── services/
    ├── ticket_service.go
    ├── ticket_comment_service.go
    ├── ticket_event_log_service.go
    ├── ticket_sla_service.go
    └── ticket_no_service.go
```

如果第一阶段希望进一步压缩复杂度，也可以先不单独拆 `ticket_comment_service.go` 和 `ticket_event_log_service.go`，统一在 `ticket_service.go` 中处理。

## 3. 枚举设计

建议统一定义到 `internal/pkg/enums/enums.go`，并按前后端枚举约定后续生成到前端。

### 3.1 TicketStatus

```go
type TicketStatus string
```

建议值：

- `new`
- `open`
- `pending_customer`
- `pending_internal`
- `resolved`
- `closed`
- `cancelled`

建议 label：

- `new` -> `新建`
- `open` -> `处理中`
- `pending_customer` -> `待客户反馈`
- `pending_internal` -> `待内部处理`
- `resolved` -> `已解决`
- `closed` -> `已关闭`
- `cancelled` -> `已取消`

### 3.2 TicketPriority

```go
type TicketPriority int
```

建议值：

- `1` -> `低`
- `2` -> `普通`
- `3` -> `高`
- `4` -> `紧急`

### 3.3 TicketSeverity

```go
type TicketSeverity int
```

建议值：

- `1` -> `轻微`
- `2` -> `严重`
- `3` -> `致命`

### 3.4 TicketSource

```go
type TicketSource string
```

建议值：

- `manual`
- `conversation`
- `portal`
- `api`
- `rule`

### 3.5 TicketCommentType

```go
type TicketCommentType string
```

建议值：

- `public_reply`
- `internal_note`
- `system_log`

### 3.6 TicketEventType

```go
type TicketEventType string
```

建议第一版支持：

- `created`
- `updated`
- `assigned`
- `transferred`
- `status_changed`
- `replied`
- `internal_noted`
- `closed`
- `reopened`
- `sla_breached`
- `linked_conversation`

### 3.7 TicketSLAType

```go
type TicketSLAType string
```

建议值：

- `first_response`
- `resolution`

### 3.8 TicketSLAStatus

```go
type TicketSLAStatus string
```

建议值：

- `running`
- `paused`
- `completed`
- `breached`

### 3.9 TicketRelationType

```go
type TicketRelationType string
```

建议值：

- `duplicate`
- `related`
- `parent`
- `child`

## 4. Model 设计

建议按当前项目习惯继续放在 `internal/models/models.go` 中，避免一开始拆过多文件。

## 4.1 Ticket

```go
type Ticket struct {
	ID                  int64                `gorm:"primaryKey;autoIncrement"`
	TicketNo            string               `gorm:"type:varchar(64);not null;default:'';uniqueIndex"`
	Title               string               `gorm:"type:varchar(255);not null;default:'';index"`
	Description         string               `gorm:"type:text"`
	Source              enums.TicketSource   `gorm:"type:varchar(50);not null;default:'';index"`
	Channel             string               `gorm:"type:varchar(50);not null;default:'';index"`
	CustomerID          int64                `gorm:"type:bigint;not null;default:0;index"`
	ConversationID      int64                `gorm:"type:bigint;not null;default:0;index"`
	CategoryID          int64                `gorm:"type:bigint;not null;default:0;index"`
	Type                string               `gorm:"type:varchar(50);not null;default:'';index"`
	Priority            enums.TicketPriority `gorm:"type:int;not null;default:2;index"`
	Severity            enums.TicketSeverity `gorm:"type:int;not null;default:1;index"`
	Status              enums.TicketStatus   `gorm:"type:varchar(50);not null;default:'new';index"`
	CurrentTeamID       int64                `gorm:"type:bigint;not null;default:0;index"`
	CurrentAssigneeID   int64                `gorm:"type:bigint;not null;default:0;index"`
	PendingReason       string               `gorm:"type:varchar(255);not null;default:''"`
	CloseReason         string               `gorm:"type:varchar(255);not null;default:''"`
	ResolutionCode      string               `gorm:"type:varchar(100);not null;default:''"`
	ResolutionSummary   string               `gorm:"type:text"`
	FirstResponseAt     *time.Time           `gorm:"type:datetime;index"`
	ResolvedAt          *time.Time           `gorm:"type:datetime;index"`
	ClosedAt            *time.Time           `gorm:"type:datetime;index"`
	DueAt               *time.Time           `gorm:"type:datetime;index"`
	NextReplyDeadlineAt *time.Time           `gorm:"type:datetime;index"`
	ResolveDeadlineAt   *time.Time           `gorm:"type:datetime;index"`
	ReopenedCount       int                  `gorm:"type:int;not null;default:0"`
	CustomFieldsJSON    string               `gorm:"type:text"`
	ExtraJSON           string               `gorm:"type:text"`
	AuditFields
}
```

设计说明：

- `ticketNo` 用于面向客服和客户展示，不能直接暴露内部 ID
- `conversationId` 用于回链来源会话，允许为 `0`
- `CurrentTeamID` 与 `CurrentAssigneeID` 保持冗余，便于高频筛选
- `CustomFieldsJSON` 第一阶段用于快速承载扩展字段

## 4.2 TicketComment

```go
type TicketComment struct {
	ID          int64                   `gorm:"primaryKey;autoIncrement"`
	TicketID    int64                   `gorm:"type:bigint;not null;index"`
	CommentType enums.TicketCommentType `gorm:"type:varchar(50);not null;default:'';index"`
	AuthorType  enums.IMSenderType      `gorm:"type:varchar(30);not null;default:'';index"`
	AuthorID    int64                   `gorm:"type:bigint;not null;default:0;index"`
	ContentType string                  `gorm:"type:varchar(30);not null;default:''"`
	Content     string                  `gorm:"type:text"`
	Payload     string                  `gorm:"type:text"`
	CreatedAt   time.Time               `gorm:"type:datetime;not null;index"`
}
```

说明：

- `AuthorType` 可直接复用现有 `IMSenderType`
- `Payload` 可用于附件、富文本元信息等扩展

## 4.3 TicketEventLog

```go
type TicketEventLog struct {
	ID           int64                 `gorm:"primaryKey;autoIncrement"`
	TicketID     int64                 `gorm:"type:bigint;not null;index"`
	EventType    enums.TicketEventType `gorm:"type:varchar(50);not null;default:'';index"`
	OperatorType enums.IMSenderType    `gorm:"type:varchar(30);not null;default:'';index"`
	OperatorID   int64                 `gorm:"type:bigint;not null;default:0;index"`
	OldValue     string                `gorm:"type:text"`
	NewValue     string                `gorm:"type:text"`
	Content      string                `gorm:"type:text"`
	Payload      string                `gorm:"type:text"`
	CreatedAt    time.Time             `gorm:"type:datetime;not null;index"`
}
```

## 4.4 TicketWatcher

```go
type TicketWatcher struct {
	ID        int64     `gorm:"primaryKey;autoIncrement"`
	TicketID  int64     `gorm:"type:bigint;not null;index;uniqueIndex:uk_ticket_watcher"`
	UserID    int64     `gorm:"type:bigint;not null;index;uniqueIndex:uk_ticket_watcher"`
	CreatedAt time.Time `gorm:"type:datetime;not null;index"`
}
```

## 4.5 TicketSLARecord

```go
type TicketSLARecord struct {
	ID            int64                  `gorm:"primaryKey;autoIncrement"`
	TicketID      int64                  `gorm:"type:bigint;not null;index"`
	SLAType       enums.TicketSLAType    `gorm:"type:varchar(50);not null;default:'';index"`
	TargetMinutes int                    `gorm:"type:int;not null;default:0"`
	Status        enums.TicketSLAStatus  `gorm:"type:varchar(30);not null;default:'';index"`
	StartedAt     *time.Time             `gorm:"type:datetime;index"`
	PausedAt      *time.Time             `gorm:"type:datetime;index"`
	StoppedAt     *time.Time             `gorm:"type:datetime;index"`
	BreachedAt    *time.Time             `gorm:"type:datetime;index"`
	ElapsedMin    int                    `gorm:"type:int;not null;default:0"`
	CreatedAt     time.Time              `gorm:"type:datetime;not null;index"`
	UpdatedAt     time.Time              `gorm:"type:datetime;not null;index"`
}
```

## 4.6 TicketRelation

```go
type TicketRelation struct {
	ID              int64                    `gorm:"primaryKey;autoIncrement"`
	TicketID        int64                    `gorm:"type:bigint;not null;index"`
	RelatedTicketID int64                    `gorm:"type:bigint;not null;index"`
	RelationType    enums.TicketRelationType `gorm:"type:varchar(30);not null;default:'';index"`
	CreatedAt       time.Time                `gorm:"type:datetime;not null;index"`
}
```

## 5. Request DTO 设计

建议统一定义到：

- `internal/pkg/dto/request/ticket_requests.go`

## 5.1 CreateTicketRequest

```go
type CreateTicketRequest struct {
	Title             string         `json:"title"`
	Description       string         `json:"description"`
	Source            string         `json:"source"`
	Channel           string         `json:"channel"`
	CustomerID        int64          `json:"customerId"`
	ConversationID    int64          `json:"conversationId"`
	CategoryID        int64          `json:"categoryId"`
	Type              string         `json:"type"`
	Priority          int            `json:"priority"`
	Severity          int            `json:"severity"`
	CurrentTeamID     int64          `json:"currentTeamId"`
	CurrentAssigneeID int64          `json:"currentAssigneeId"`
	DueAt             string         `json:"dueAt"`
	CustomFields      map[string]any `json:"customFields"`
}
```

说明：

- `DueAt` 第一阶段可先用字符串接收，再在 service 中解析
- `CustomFields` 直接承接动态字段

## 5.2 CreateTicketFromConversationRequest

```go
type CreateTicketFromConversationRequest struct {
	ConversationID    int64          `json:"conversationId"`
	Title             string         `json:"title"`
	Description       string         `json:"description"`
	CategoryID        int64          `json:"categoryId"`
	Priority          int            `json:"priority"`
	Severity          int            `json:"severity"`
	CurrentTeamID     int64          `json:"currentTeamId"`
	CurrentAssigneeID int64          `json:"currentAssigneeId"`
	SyncToConversation bool          `json:"syncToConversation"`
	CustomFields      map[string]any `json:"customFields"`
}
```

## 5.3 UpdateTicketRequest

```go
type UpdateTicketRequest struct {
	TicketID          int64          `json:"ticketId"`
	Title             string         `json:"title"`
	Description       string         `json:"description"`
	CategoryID        int64          `json:"categoryId"`
	Type              string         `json:"type"`
	Priority          int            `json:"priority"`
	Severity          int            `json:"severity"`
	CurrentTeamID     int64          `json:"currentTeamId"`
	CurrentAssigneeID int64          `json:"currentAssigneeId"`
	DueAt             string         `json:"dueAt"`
	CustomFields      map[string]any `json:"customFields"`
}
```

## 5.4 AssignTicketRequest

```go
type AssignTicketRequest struct {
	TicketID      int64  `json:"ticketId"`
	ToUserID      int64  `json:"toUserId"`
	ToTeamID      int64  `json:"toTeamId"`
	Reason        string `json:"reason"`
}
```

## 5.5 ChangeTicketStatusRequest

```go
type ChangeTicketStatusRequest struct {
	TicketID       int64  `json:"ticketId"`
	Status         string `json:"status"`
	PendingReason  string `json:"pendingReason"`
	CloseReason    string `json:"closeReason"`
	ResolutionCode string `json:"resolutionCode"`
	Reason         string `json:"reason"`
}
```

## 5.6 ReplyTicketRequest

```go
type ReplyTicketRequest struct {
	TicketID    int64  `json:"ticketId"`
	ContentType string `json:"contentType"`
	Content     string `json:"content"`
	Payload     string `json:"payload"`
}
```

## 5.7 InternalNoteRequest

```go
type InternalNoteRequest struct {
	TicketID    int64  `json:"ticketId"`
	ContentType string `json:"contentType"`
	Content     string `json:"content"`
	Payload     string `json:"payload"`
}
```

## 5.8 CloseTicketRequest

```go
type CloseTicketRequest struct {
	TicketID    int64  `json:"ticketId"`
	CloseReason string `json:"closeReason"`
}
```

## 5.9 ReopenTicketRequest

```go
type ReopenTicketRequest struct {
	TicketID int64  `json:"ticketId"`
	Reason   string `json:"reason"`
}
```

## 6. Response DTO 设计

建议统一定义到：

- `internal/pkg/dto/response/ticket_responses.go`

## 6.1 TicketResponse

```go
type TicketResponse struct {
	ID                  int64         `json:"id"`
	TicketNo            string        `json:"ticketNo"`
	Title               string        `json:"title"`
	Description         string        `json:"description"`
	Source              string        `json:"source"`
	Channel             string        `json:"channel"`
	CustomerID          int64         `json:"customerId"`
	ConversationID      int64         `json:"conversationId"`
	CategoryID          int64         `json:"categoryId"`
	Type                string        `json:"type"`
	Priority            int           `json:"priority"`
	Severity            int           `json:"severity"`
	Status              string        `json:"status"`
	CurrentTeamID       int64         `json:"currentTeamId"`
	CurrentAssigneeID   int64         `json:"currentAssigneeId"`
	PendingReason       string        `json:"pendingReason"`
	CloseReason         string        `json:"closeReason"`
	ResolutionCode      string        `json:"resolutionCode"`
	ResolutionSummary   string        `json:"resolutionSummary"`
	FirstResponseAt     string        `json:"firstResponseAt"`
	ResolvedAt          string        `json:"resolvedAt"`
	ClosedAt            string        `json:"closedAt"`
	DueAt               string        `json:"dueAt"`
	NextReplyDeadlineAt string        `json:"nextReplyDeadlineAt"`
	ResolveDeadlineAt   string        `json:"resolveDeadlineAt"`
	ReopenedCount       int           `json:"reopenedCount"`
	CreatedAt           string        `json:"createdAt"`
	UpdatedAt           string        `json:"updatedAt"`
	Customer            *SimpleCustomerResponse     `json:"customer,omitempty"`
	Conversation        *SimpleConversationResponse `json:"conversation,omitempty"`
	CurrentAssignee     *SimpleUserResponse         `json:"currentAssignee,omitempty"`
	CurrentTeam         *SimpleTeamResponse         `json:"currentTeam,omitempty"`
	SLA                 []TicketSLAResponse         `json:"sla,omitempty"`
}
```

第一阶段如果不想在 builder 里做太多聚合，也可以先只返回主档字段，再逐步补充 `customer/conversation/currentAssignee/currentTeam` 摘要对象。

## 6.2 TicketCommentResponse

```go
type TicketCommentResponse struct {
	ID          int64  `json:"id"`
	TicketID    int64  `json:"ticketId"`
	CommentType string `json:"commentType"`
	AuthorType  string `json:"authorType"`
	AuthorID    int64  `json:"authorId"`
	ContentType string `json:"contentType"`
	Content     string `json:"content"`
	Payload     string `json:"payload"`
	CreatedAt   string `json:"createdAt"`
}
```

## 6.3 TicketEventLogResponse

```go
type TicketEventLogResponse struct {
	ID           int64  `json:"id"`
	TicketID     int64  `json:"ticketId"`
	EventType    string `json:"eventType"`
	OperatorType string `json:"operatorType"`
	OperatorID   int64  `json:"operatorId"`
	OldValue     string `json:"oldValue"`
	NewValue     string `json:"newValue"`
	Content      string `json:"content"`
	Payload      string `json:"payload"`
	CreatedAt    string `json:"createdAt"`
}
```

## 6.4 TicketSLAResponse

```go
type TicketSLAResponse struct {
	SLAType       string `json:"slaType"`
	TargetMinutes int    `json:"targetMinutes"`
	Status        string `json:"status"`
	StartedAt     string `json:"startedAt"`
	PausedAt      string `json:"pausedAt"`
	StoppedAt     string `json:"stoppedAt"`
	BreachedAt    string `json:"breachedAt"`
	ElapsedMin    int    `json:"elapsedMin"`
}
```

## 6.5 TicketDetailResponse

```go
type TicketDetailResponse struct {
	Ticket   TicketResponse           `json:"ticket"`
	Comments []TicketCommentResponse  `json:"comments"`
	Events   []TicketEventLogResponse `json:"events"`
}
```

## 7. Builder 设计

建议新增：

- `internal/builders/ticket_builder.go`
- `internal/builders/ticket_comment_builder.go`

建议方法：

```go
func BuildTicket(item *models.Ticket) response.TicketResponse
func BuildTicketList(list []models.Ticket) []response.TicketResponse
func BuildTicketComment(item *models.TicketComment) response.TicketCommentResponse
func BuildTicketCommentList(list []models.TicketComment) []response.TicketCommentResponse
func BuildTicketEventLogList(list []models.TicketEventLog) []response.TicketEventLogResponse
func BuildTicketSLAList(list []models.TicketSLARecord) []response.TicketSLAResponse
```

要求：

- builders 只做映射，不访问数据库
- 如果详情页需要聚合客户、会话、客服摘要，优先在 service 层先聚合好

## 8. Repository 设计

Repository 方法签名统一接收 `db *gorm.DB`，以适配 `sqls.DB()` 和事务 `ctx.Tx`。

## 8.1 TicketRepository

建议方法：

- `Get(db *gorm.DB, id int64) *models.Ticket`
- `TakeByTicketNo(db *gorm.DB, ticketNo string) *models.Ticket`
- `FindPageByCnd(db *gorm.DB, cnd *sqls.Cnd) ([]models.Ticket, *sqls.Paging)`
- `Create(db *gorm.DB, item *models.Ticket) error`
- `Updates(db *gorm.DB, id int64, values map[string]any) error`
- `Delete(db *gorm.DB, id int64) error`

可选补充：

- `CountByStatus(db *gorm.DB, status enums.TicketStatus) int64`
- `FindByConversationID(db *gorm.DB, conversationID int64) []models.Ticket`

## 8.2 TicketCommentRepository

建议方法：

- `Create(db *gorm.DB, item *models.TicketComment) error`
- `FindByTicketID(db *gorm.DB, ticketID int64) []models.TicketComment`
- `FindPageByTicketID(db *gorm.DB, ticketID int64, paging *sqls.Paging) ([]models.TicketComment, *sqls.Paging)`

## 8.3 TicketEventLogRepository

建议方法：

- `Create(db *gorm.DB, item *models.TicketEventLog) error`
- `FindByTicketID(db *gorm.DB, ticketID int64) []models.TicketEventLog`

## 8.4 TicketSLARecordRepository

建议方法：

- `Create(db *gorm.DB, item *models.TicketSLARecord) error`
- `FindByTicketID(db *gorm.DB, ticketID int64) []models.TicketSLARecord`
- `TakeByTicketIDAndType(db *gorm.DB, ticketID int64, slaType enums.TicketSLAType) *models.TicketSLARecord`
- `Updates(db *gorm.DB, id int64, values map[string]any) error`

## 8.5 TicketWatcherRepository

建议方法：

- `Create(db *gorm.DB, item *models.TicketWatcher) error`
- `DeleteByTicketIDAndUserID(db *gorm.DB, ticketID, userID int64) error`
- `FindByTicketID(db *gorm.DB, ticketID int64) []models.TicketWatcher`

## 9. Service 设计

建议工单业务由 `ticketService` 统一承接，避免早期拆散后逻辑分裂。

## 9.1 TicketService 建议方法

```go
type ticketService struct{}
```

建议暴露：

- `Get(id int64) *models.Ticket`
- `FindPageByCnd(cnd *sqls.Cnd) ([]models.Ticket, *sqls.Paging)`
- `GetDetail(id int64) (*dto.TicketDetailAggregate, error)`
- `CreateTicket(req request.CreateTicketRequest, operator *dto.AuthPrincipal) (*models.Ticket, error)`
- `CreateFromConversation(req request.CreateTicketFromConversationRequest, operator *dto.AuthPrincipal) (*models.Ticket, error)`
- `UpdateTicket(req request.UpdateTicketRequest, operator *dto.AuthPrincipal) error`
- `AssignTicket(req request.AssignTicketRequest, operator *dto.AuthPrincipal) error`
- `ChangeStatus(req request.ChangeTicketStatusRequest, operator *dto.AuthPrincipal) error`
- `ReplyTicket(req request.ReplyTicketRequest, operator *dto.AuthPrincipal) (*models.TicketComment, error)`
- `AddInternalNote(req request.InternalNoteRequest, operator *dto.AuthPrincipal) (*models.TicketComment, error)`
- `CloseTicket(req request.CloseTicketRequest, operator *dto.AuthPrincipal) error`
- `ReopenTicket(req request.ReopenTicketRequest, operator *dto.AuthPrincipal) error`

## 9.2 事务边界

按项目约定，以下操作必须开事务：

- 创建工单主档 + 初始化 SLA + 写事件日志
- 会话转工单 + 回写会话事件
- 状态流转 + 更新时间字段 + SLA 状态更新 + 事件日志
- 回复/内部备注 + 首次响应时间更新 + 事件日志
- 关闭/重开 + SLA 更新 + 事件日志

以下操作不一定需要事务：

- 单纯读取详情
- 单纯读取列表
- 单条简单更新且没有附带日志写入的场景

但工单模块几乎所有写操作都伴随事件日志，因此第一阶段可以统一采用事务，风格会更稳。

## 9.3 CreateTicket 流程

建议流程：

1. 校验标题、客户、会话、优先级、状态默认值
2. 如传入 `conversationId`，校验会话存在
3. 生成 `ticketNo`
4. 构建 `models.Ticket`
5. 开事务
6. 创建工单主档
7. 初始化两条 SLA 记录
8. 写入 `created` 事件
9. 若有来源会话，写入会话事件日志
10. 提交事务

## 9.4 CreateFromConversation 流程

建议流程：

1. 校验 `conversationId`
2. 加载会话和客户信息
3. 默认带入 `customerId`
4. 调用内部 `createTicket(...)`
5. 若 `SyncToConversation` 为 true，回写系统事件或系统消息

### 9.4.1 标题与描述默认策略

如果前端未显式传入标题与描述：

- 标题默认取会话 `subject`
- 描述默认取最近若干条消息摘要

后续再扩展为 AI 生成摘要。

## 9.5 UpdateTicket 流程

更新范围建议控制在非状态机字段：

- 标题
- 描述
- 分类
- 类型
- 优先级
- 严重度
- 当前团队
- 当前处理人
- 截止时间
- 自定义字段

流程：

1. 校验工单存在
2. 校验当前状态是否允许编辑
3. 组装更新字段
4. 更新主档
5. 写 `updated` 事件日志

## 9.6 AssignTicket 流程

流程：

1. 校验工单存在
2. 校验处理人和团队
3. 开事务
4. 更新 `currentTeamId/currentAssigneeId`
5. 写 `assigned` 或 `transferred` 事件
6. 提交事务

规则建议：

- 若原处理人为 0，记为 `assigned`
- 若原处理人非 0，记为 `transferred`

## 9.7 ChangeStatus 流程

流程：

1. 校验工单存在
2. 校验目标状态是否合法
3. 校验状态迁移是否允许
4. 按目标状态设置 `resolvedAt/closedAt/pendingReason/closeReason`
5. 驱动 SLA 暂停/恢复/停止
6. 写 `status_changed` 事件

规则建议：

- 改为 `pending_customer` 时必须允许填写 `pendingReason`
- 改为 `resolved` 时可要求填写 `resolutionCode`
- 改为 `closed` 时必须有 `closeReason`

## 9.8 ReplyTicket 流程

流程：

1. 校验工单存在
2. 校验内容不为空
3. 开事务
4. 插入 `TicketComment(commentType=public_reply)`
5. 若 `firstResponseAt` 为空则写入当前时间
6. 若工单当前是 `pending_customer`，自动改回 `open`
7. 更新首次响应 SLA
8. 写 `replied` 事件
9. 提交事务

## 9.9 AddInternalNote 流程

流程：

1. 校验工单存在
2. 校验内容不为空
3. 开事务
4. 插入 `TicketComment(commentType=internal_note)`
5. 写 `internal_noted` 事件
6. 提交事务

## 9.10 CloseTicket / ReopenTicket 流程

关闭：

1. 校验工单存在
2. 校验关闭原因
3. 开事务
4. 更新 `status=closed`
5. 写入 `closedAt/closeReason`
6. 停止 SLA
7. 写 `closed` 事件
8. 提交事务

重开：

1. 校验工单存在且当前可重开
2. 开事务
3. 更新 `status=open`
4. `reopenedCount + 1`
5. 恢复解决 SLA
6. 写 `reopened` 事件
7. 提交事务

## 10. SLA Service 设计

建议提供一个 `ticketSLAService`，专门承接计时规则，避免把 SLA 状态变更散落在 `ticketService` 中。

建议方法：

- `InitTicketSLAs(tx *gorm.DB, ticket *models.Ticket) error`
- `CompleteFirstResponse(tx *gorm.DB, ticketID int64, at time.Time) error`
- `PauseResolution(tx *gorm.DB, ticketID int64, at time.Time) error`
- `ResumeResolution(tx *gorm.DB, ticketID int64, at time.Time) error`
- `CompleteResolution(tx *gorm.DB, ticketID int64, at time.Time) error`
- `CloseSLAs(tx *gorm.DB, ticketID int64, at time.Time) error`

第一阶段建议先实现基础状态改写，不急于做精确的工作日历计算。

## 11. TicketNo 生成设计

建议使用独立 service：

- `ticketNoService.Next() (string, error)`

建议格式：

- `TK202604020001`

实现建议：

- 前缀固定 `TK`
- 日期使用 `YYYYMMDD`
- 序号每日递增

第一阶段可以接受用数据库查询当日最大序号后生成，只要注意并发下唯一性。

更稳妥的方案：

- 先生成时间戳 + 随机尾号
- 或引入专门序号表

如果项目当前并发量不大，第一阶段可先采用简单方案。

## 12. Controller 设计

建议新增：

- `internal/controllers/console/ticket_controller.go`

结构体：

```go
type TicketController struct {
	Ctx iris.Context
}
```

## 12.1 路由注册建议

在 `internal/bootstrap/server.go` 中按现有风格注册：

```go
m.Party("/ticket").Handle(new(console.TicketController))
```

最终路径：

- `/api/console/ticket/list`
- `/api/console/ticket/{id}`
- `/api/console/ticket/create`
- `/api/console/ticket/update`
- `/api/console/ticket/assign`
- `/api/console/ticket/change_status`
- `/api/console/ticket/reply`
- `/api/console/ticket/internal_note`
- `/api/console/ticket/close`
- `/api/console/ticket/reopen`
- `/api/console/ticket/create_from_conversation`

## 12.2 Controller 方法建议

### AnyList()

职责：

- 权限校验
- 读取分页与筛选参数
- 构建 `params.NewPagedSqlCnd(...)`
- 调 `TicketService.FindPageByCnd`
- 调 builder 返回列表 DTO

### GetBy(id int64)

职责：

- 权限校验
- 调 `TicketService.GetDetail`
- 返回详情 DTO

### PostCreate()

职责：

- 权限校验
- `params.ReadJSON`
- 调 `TicketService.CreateTicket`
- 返回创建结果

### PostCreate_from_conversation()

职责：

- 权限校验
- `params.ReadJSON`
- 调 `TicketService.CreateFromConversation`
- 返回创建结果

### PostUpdate()

职责：

- 权限校验
- 读取 JSON
- 调 `TicketService.UpdateTicket`

### PostAssign()

职责：

- 权限校验
- 调 `TicketService.AssignTicket`

### PostChange_status()

职责：

- 权限校验
- 调 `TicketService.ChangeStatus`

### PostReply()

职责：

- 权限校验
- 调 `TicketService.ReplyTicket`
- 返回 comment DTO

### PostInternal_note()

职责：

- 权限校验
- 调 `TicketService.AddInternalNote`
- 返回 comment DTO

### PostClose()

职责：

- 权限校验
- 调 `TicketService.CloseTicket`

### PostReopen()

职责：

- 权限校验
- 调 `TicketService.ReopenTicket`

## 13. 列表筛选建议

`AnyList()` 推荐支持：

- `status`
- `priority`
- `severity`
- `categoryId`
- `currentAssigneeId`
- `currentTeamId`
- `customerId`
- `conversationId`
- `source`
- `keyword`

`keyword` 建议匹配：

- `ticket_no`
- `title`
- `description`

分页排序建议：

- 默认 `.Desc("updated_at").Desc("id")`

## 14. 详情聚合建议

为了避免 controller 内拼装，建议 service 层定义一个详情聚合结构，例如：

```go
type TicketDetailAggregate struct {
	Ticket   *models.Ticket
	Comments []models.TicketComment
	Events   []models.TicketEventLog
	SLAs     []models.TicketSLARecord
}
```

后续如果需要附带客户、会话、处理人摘要，再逐步扩展：

- `Customer *models.Customer`
- `Conversation *models.Conversation`
- `CurrentAssignee *models.User`
- `CurrentTeam *models.AgentTeam`

## 15. 事件日志规范

建议统一通过内部辅助方法写日志：

```go
func (s *ticketService) logEvent(
	tx *gorm.DB,
	ticketID int64,
	eventType enums.TicketEventType,
	operator *dto.AuthPrincipal,
	oldValue string,
	newValue string,
	content string,
	payload string,
) error
```

要求：

- 所有核心写操作必须落事件
- `content` 用于人类可读摘要
- `payload` 用于结构化扩展

示例：

- 指派：`将工单指派给客服 A`
- 状态变更：`状态由处理中变更为待客户反馈`
- 关闭：`工单关闭，原因：客户确认已解决`

## 16. 错误处理建议

遵循项目统一规范：

- 不向前端透传底层 SQL 错误
- 参数问题统一转业务可读错误
- 记录不存在统一返回明确提示

建议典型错误：

- `工单不存在`
- `会话不存在`
- `客户不存在`
- `工单当前状态不允许该操作`
- `关闭原因不能为空`
- `回复内容不能为空`
- `目标处理人不存在`

## 17. 测试建议

至少补 service 层核心路径测试，优先覆盖：

- 创建工单
- 会话转工单
- 状态流转
- 公开回复触发首次响应
- 关闭和重开
- SLA 状态变化

建议：

- 以 SQLite 跑 service 测试
- 覆盖合法流转与非法流转
- 覆盖事务内多表写入成功与失败回滚

## 18. 第一阶段实现顺序建议

推荐按以下顺序实现：

1. 补枚举
2. 补 model 并注册到 `models.Models`
3. 补 generator 注册
4. 执行 `make generator`
5. 补 repositories
6. 补 `ticket_service`
7. 补 builder
8. 补 controller 与路由
9. 联调列表、详情、创建、转工单
10. 补 service 层测试

## 19. 结论

工单模块后端落地的关键不是“多几张表”，而是把以下边界立住：

- `Ticket` 作为独立正式对象
- 所有写操作通过 service 编排
- 事务内同时维护主档、评论、SLA、事件日志
- controller 只做参数、权限、响应
- builders 只做 DTO 映射

只要这个边界不乱，后续再扩展自动化、门户、AI 辅助、报表都会比较稳。
