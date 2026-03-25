package models

import (
	"cs-agent/internal/pkg/enums"
	"time"
)

// Models 注册所有需要迁移和代码生成的模型。
var Models = []any{
	&Migration{},
	&User{},
	&UserIdentity{},
	&Role{},
	&Permission{},
	&UserRole{},
	&RolePermission{},
	&UserPermission{},
	&LoginSession{},
	&LoginCredentialLog{},
	&Asset{},
	&Tag{},
	&Conversation{},
	&ConversationParticipant{},
	&ConversationReadState{},
	&Message{},
	&ConversationAssignment{},
	&ConversationTag{},
	&QuickReply{},
	&ConversationEventLog{},
	&AIAgent{},
	&WidgetSite{},
	&AgentProfile{},
	&AgentTeam{},
	&AgentTeamSchedule{},
	&AIConfig{},
	&TicketCategory{},
	&Ticket{},
	&TicketAssignment{},
	&TicketReply{},
	&TicketComment{},
	&TicketAttachment{},
	&TicketEventLog{},
	&KnowledgeBase{},
	&KnowledgeDocument{},
	&KnowledgeChunk{},
	&KnowledgeRetrieveLog{},
	&KnowledgeRetrieveHit{},
	&KnowledgeFeedback{},
	&SkillDefinition{},
	&SkillRunLog{},
}

type Migration struct {
	ID         int64     `gorm:"primaryKey;autoIncrement"`
	Version    int64     `gorm:"type:bigint;not null;uniqueIndex"`
	Remark     string    `gorm:"type:text"`
	Success    bool      `gorm:"not null;default:false"`
	ErrorInfo  string    `gorm:"type:text"`
	RetryCount int       `gorm:"type:int;not null;default:0"`
	CreatedAt  time.Time `gorm:"type:datetime"`
	UpdatedAt  time.Time `gorm:"type:datetime"`
}

// AuditFields 定义涉及用户操作数据的统一审计字段。
// 该结构记录数据创建与更新的时间、操作者ID和操作者名称。
type AuditFields struct {
	CreatedAt      time.Time `gorm:"type:datetime;not null;index"`          // CreatedAt 记录数据创建时间。
	CreateUserID   int64     `gorm:"type:bigint;not null;default:0;index"`  // CreateUserID 记录创建人用户ID；系统任务写0。
	CreateUserName string    `gorm:"type:varchar(100);not null;default:''"` // CreateUserName 记录创建人名称；系统任务写system。
	UpdatedAt      time.Time `gorm:"type:datetime;not null;index"`          // UpdatedAt 记录数据最近更新时间。
	UpdateUserID   int64     `gorm:"type:bigint;not null;default:0;index"`  // UpdateUserID 记录最后更新人用户ID；系统任务写0。
	UpdateUserName string    `gorm:"type:varchar(100);not null;default:''"` // UpdateUserName 记录最后更新人名称；系统任务写system。
}

// User 后台用户账号。
type User struct {
	ID           int64        `gorm:"primaryKey;autoIncrement"`
	Username     string       `gorm:"type:varchar(100);not null;uniqueIndex"`
	Nickname     string       `gorm:"type:varchar(100);not null;default:'';index"`
	Avatar       string       `gorm:"type:varchar(255);not null;default:''"`
	Mobile       *string      `gorm:"type:varchar(32);uniqueIndex"`
	Email        *string      `gorm:"type:varchar(100);uniqueIndex"`
	Password     string       `gorm:"type:varchar(255);not null;default:''"`
	PasswordSalt string       `gorm:"type:varchar(64);not null;default:''"`
	Status       enums.Status `gorm:"type:int;not null;default:0;index"`
	LastLoginAt  *time.Time   `gorm:"type:datetime"`
	LastLoginIP  string       `gorm:"type:varchar(64);not null;default:''"`
	Remark       string       `gorm:"type:text"`
	DeletedAt    *time.Time   `gorm:"type:datetime;index"`
	AuditFields
}

// UserIdentity 第三方身份绑定信息。
type UserIdentity struct {
	ID              int64        `gorm:"primaryKey;autoIncrement"`
	UserID          int64        `gorm:"type:bigint;not null;index;uniqueIndex:uk_provider_user"`
	Provider        string       `gorm:"type:varchar(50);not null;default:'';index;uniqueIndex:uk_provider_user;uniqueIndex:uk_provider_union"`
	ProviderUserID  string       `gorm:"type:varchar(128);not null;default:'';uniqueIndex:uk_provider_user"`
	ProviderUnionID *string      `gorm:"type:varchar(128);uniqueIndex:uk_provider_union"`
	ProviderCorpID  string       `gorm:"type:varchar(128);not null;default:'';index"`
	ProviderName    string       `gorm:"type:varchar(100);not null;default:''"`
	RawProfile      string       `gorm:"type:text"`
	Status          enums.Status `gorm:"type:int;not null;default:0;index"`
	LastAuthAt      *time.Time   `gorm:"type:datetime"`
	AuditFields
}

// Role 角色定义。
type Role struct {
	ID       int64        `gorm:"primaryKey;autoIncrement"`
	Name     string       `gorm:"type:varchar(100);not null;default:'';index"`
	Code     string       `gorm:"type:varchar(100);not null;uniqueIndex"`
	Status   enums.Status `gorm:"type:int;not null;default:0;index"`
	IsSystem bool         `gorm:"not null;default:false;index"`
	SortNo   int          `gorm:"type:int;not null;default:0;index"`
	Remark   string       `gorm:"type:text"`
	AuditFields
}

// Permission 权限点定义。
type Permission struct {
	ID        int64        `gorm:"primaryKey;autoIncrement"`
	Name      string       `gorm:"type:varchar(100);not null;default:''"`
	Code      string       `gorm:"type:varchar(150);not null;uniqueIndex"`
	Type      string       `gorm:"type:varchar(20);not null;default:'';index"`
	GroupName string       `gorm:"type:varchar(100);not null;default:'';index"`
	ParentID  int64        `gorm:"type:bigint;not null;default:0;index"`
	Path      string       `gorm:"type:varchar(255);not null;default:''"`
	Method    string       `gorm:"type:varchar(20);not null;default:''"`
	APIPath   string       `gorm:"column:api_path;type:varchar(255);not null;default:''"`
	SortNo    int          `gorm:"type:int;not null;default:0;index"`
	Status    enums.Status `gorm:"type:int;not null;default:0;index"`
	IsBuiltin bool         `gorm:"not null;default:true;index"`
	Remark    string       `gorm:"type:text"`
	AuditFields
}

// UserRole 用户和角色关联。
type UserRole struct {
	ID     int64 `gorm:"primaryKey;autoIncrement"`
	UserID int64 `gorm:"type:bigint;not null;index;uniqueIndex:uk_user_role"`
	RoleID int64 `gorm:"type:bigint;not null;index;uniqueIndex:uk_user_role"`
	AuditFields
}

// RolePermission 角色和权限关联。
type RolePermission struct {
	ID           int64 `gorm:"primaryKey;autoIncrement"`
	RoleID       int64 `gorm:"type:bigint;not null;index;uniqueIndex:uk_role_permission"`
	PermissionID int64 `gorm:"type:bigint;not null;index;uniqueIndex:uk_role_permission"`
	AuditFields
}

// UserPermission 用户级例外权限。
//
//	用于处理少量临时授权或拒绝授权场景。
type UserPermission struct {
	ID           int64      `gorm:"primaryKey;autoIncrement"`
	UserID       int64      `gorm:"type:bigint;not null;index;uniqueIndex:uk_user_permission"`
	PermissionID int64      `gorm:"type:bigint;not null;index;uniqueIndex:uk_user_permission"`
	Effect       int        `gorm:"type:int;not null;default:1;index"` // Effect 表示权限生效方式：1允许 -1拒绝。
	ExpiredAt    *time.Time `gorm:"type:datetime"`
	Remark       string     `gorm:"type:text"`
	AuditFields
}

// LoginSession 登录会话或刷新令牌记录。
type LoginSession struct {
	ID         int64      `gorm:"primaryKey;autoIncrement"`
	UserID     int64      `gorm:"type:bigint;not null;index"`
	TokenID    string     `gorm:"type:varchar(128);not null;uniqueIndex"`
	TokenType  string     `gorm:"type:varchar(20);not null;default:'';index"`
	ClientType string     `gorm:"type:varchar(50);not null;default:'';index"`
	ClientIP   string     `gorm:"type:varchar(64);not null;default:''"`
	UserAgent  string     `gorm:"type:varchar(255);not null;default:''"`
	ExpiredAt  time.Time  `gorm:"type:datetime;not null;index"`
	RevokedAt  *time.Time `gorm:"type:datetime;index"`
	LastSeenAt *time.Time `gorm:"type:datetime"`
	AuditFields
}

// LoginCredentialLog 登录凭证校验日志。
type LoginCredentialLog struct {
	ID        int64     `gorm:"primaryKey;autoIncrement"`
	Principal string    `gorm:"type:varchar(100);not null;default:'';index"`
	UserID    int64     `gorm:"type:bigint;not null;default:0;index"`
	Success   bool      `gorm:"not null;default:false;index"`
	ClientIP  string    `gorm:"type:varchar(64);not null;default:''"`
	UserAgent string    `gorm:"type:varchar(255);not null;default:''"`
	Reason    string    `gorm:"type:varchar(255);not null;default:''"`
	CreatedAt time.Time `gorm:"type:datetime;not null;index"`
}

// Asset 存储的文件资源，如上传的附件等。
type Asset struct {
	ID         int64               `gorm:"primaryKey;autoIncrement"`
	AssetID    string              `gorm:"type:varchar(64);not null;uniqueIndex"`
	Provider   enums.AssetProvider `gorm:"type:varchar(50);not null;default:'';index"`
	StorageKey string              `gorm:"type:varchar(255);not null;default:'';uniqueIndex:uk_storage_key"`
	Filename   string              `gorm:"type:varchar(255);not null;default:''"`
	FileSize   int64               `gorm:"type:bigint;not null;default:0"`
	MimeType   string              `gorm:"type:varchar(100);not null;default:''"`
	Status     enums.AssetStatus   `gorm:"type:int;not null;default:1;index"`
	AuditFields
}

type Tag struct {
	ID       int64        `gorm:"primaryKey;autoIncrement"`
	ParentID int64        `gorm:"type:bigint;not null;index"`
	Name     string       `gorm:"type:varchar(50);not null;"`
	Remark   string       `gorm:"type:text;"`
	SortNo   int          `gorm:"type:int;not null;default:0"`
	Status   enums.Status `gorm:"type:int;not null;default:0"`
	AuditFields
}

// Conversation 客服会话。
type Conversation struct {
	ID                  int64                           `gorm:"primaryKey;autoIncrement"`                    // ID 为会话主键。
	AIAgentID           int64                           `gorm:"type:bigint;not null;default:0;index"`        // AIAgentID 为当前会话绑定的 AI Agent ID。
	ChannelType         enums.IMConversationChannel     `gorm:"type:varchar(50);not null;default:'';index"`  // ChannelType 为会话来源渠道，如 web_chat。
	Subject             string                          `gorm:"type:varchar(255);not null;default:''"`       // Subject 为会话标题或摘要。
	Status              enums.IMConversationStatus      `gorm:"type:int;not null;default:1;index"`           // Status 为会话状态，如待接入、处理中、已关闭。
	ServiceMode         enums.IMConversationServiceMode `gorm:"type:int;not null;default:3;index"`           // ServiceMode 为服务模式，如仅AI、仅人工、AI优先人工接管。
	Priority            int                             `gorm:"type:int;not null;default:0;index"`           // Priority 为会话优先级。
	SourceUserID        int64                           `gorm:"type:bigint;not null;default:0;index"`        // SourceUserID 为发起会话的站内用户ID。
	ExternalUserID      string                          `gorm:"type:varchar(128);not null;default:'';index"` // ExternalUserID 为外部访客ID。
	CurrentAssigneeID   int64                           `gorm:"type:bigint;not null;default:0;index"`        // CurrentAssigneeID 为当前接待客服ID。
	CurrentTeamID       int64                           `gorm:"type:bigint;not null;default:0;index"`        // CurrentTeamID 为当前处理客服组ID。
	LastMessageID       int64                           `gorm:"type:bigint;not null;default:0;index"`        // LastMessageID 为最后一条消息ID。
	LastMessageAt       *time.Time                      `gorm:"type:datetime;index"`                         // LastMessageAt 为最后消息时间。
	LastMessageSummary  string                          `gorm:"type:varchar(255);not null;default:''"`       // LastMessageSummary 为最后一条消息摘要。
	CustomerUnreadCount int                             `gorm:"type:int;not null;default:0"`                 // CustomerUnreadCount 为用户侧未读数。
	AgentUnreadCount    int                             `gorm:"type:int;not null;default:0"`                 // AgentUnreadCount 为客服侧未读数。
	HandoffAt           *time.Time                      `gorm:"type:datetime;index"`                         // HandoffAt 为最近一次转人工时间。
	HandoffReason       string                          `gorm:"type:varchar(255);not null;default:''"`       // HandoffReason 为最近一次转人工原因。
	AIReplyRounds       int                             `gorm:"type:int;not null;default:0"`                 // AIReplyRounds 为当前会话内 AI 已成功回复次数。
	ClosedAt            *time.Time                      `gorm:"type:datetime;index"`                         // ClosedAt 为会话关闭时间。
	ClosedBy            int64                           `gorm:"type:bigint;not null;default:0;index"`        // ClosedBy 为关闭人用户ID，访客关闭时写0。
	CloseReason         string                          `gorm:"type:varchar(255);not null;default:''"`       // CloseReason 为关闭原因。
	AuditFields
}

// ConversationParticipant 会话参与方。
type ConversationParticipant struct {
	ID                    int64        `gorm:"primaryKey;autoIncrement"`
	ConversationID        int64        `gorm:"type:bigint;not null;index;uniqueIndex:uk_conversation_participant"`
	ParticipantType       string       `gorm:"type:varchar(30);not null;default:'';index;uniqueIndex:uk_conversation_participant"`
	ParticipantID         int64        `gorm:"type:bigint;not null;default:0;uniqueIndex:uk_conversation_participant"`
	ExternalParticipantID string       `gorm:"type:varchar(128);not null;default:''"`
	JoinedAt              *time.Time   `gorm:"type:datetime"`
	LeftAt                *time.Time   `gorm:"type:datetime"`
	Status                enums.Status `gorm:"type:int;not null;default:0;index"`
	AuditFields
}

// ConversationReadState 会话读游标。
type ConversationReadState struct {
	ID                int64              `gorm:"primaryKey;autoIncrement"`
	ConversationID    int64              `gorm:"type:bigint;not null;index;uniqueIndex:uk_conversation_reader"`
	ReaderType        enums.IMSenderType `gorm:"type:varchar(30);not null;default:'';index;uniqueIndex:uk_conversation_reader"`
	ReaderID          int64              `gorm:"type:bigint;not null;default:0;uniqueIndex:uk_conversation_reader"`
	ExternalReaderID  string             `gorm:"type:varchar(128);not null;default:'';uniqueIndex:uk_conversation_reader"`
	LastReadMessageID int64              `gorm:"type:bigint;not null;default:0;index"`
	LastReadSeqNo     int64              `gorm:"type:bigint;not null;default:0;index"`
	LastReadAt        *time.Time         `gorm:"type:datetime"`
	AuditFields
}

// Message 会话消息。
type Message struct {
	ID              int64               `gorm:"primaryKey;autoIncrement"`
	ConversationID  int64               `gorm:"type:bigint;not null;index;uniqueIndex:uk_conversation_seq;uniqueIndex:uk_conversation_client_msg"`
	ClientMsgID     string              `gorm:"type:varchar(128);not null;default:'';uniqueIndex:uk_conversation_client_msg"`
	SenderType      enums.IMSenderType  `gorm:"type:varchar(30);not null;default:'';index"`
	SenderID        int64               `gorm:"type:bigint;not null;default:0;index"`
	ReceiverType    string              `gorm:"type:varchar(30);not null;default:'';index"`
	MessageType     enums.IMMessageType `gorm:"type:varchar(30);not null;default:'';index"`
	Content         string              `gorm:"type:text"`
	Payload         string              `gorm:"type:text"`
	SeqNo           int64               `gorm:"type:bigint;not null;default:0;uniqueIndex:uk_conversation_seq"`
	SendStatus      int                 `gorm:"type:int;not null;default:2;index"`
	SentAt          *time.Time          `gorm:"type:datetime;index"`
	DeliveredAt     *time.Time          `gorm:"type:datetime"`
	ReadAt          *time.Time          `gorm:"type:datetime"`
	RecalledAt      *time.Time          `gorm:"type:datetime"`
	QuotedMessageID int64               `gorm:"type:bigint;not null;default:0;index"`
	AuditFields
}

// ConversationAssignment 会话接待关系。
type ConversationAssignment struct {
	ID             int64                    `gorm:"primaryKey;autoIncrement"`
	ConversationID int64                    `gorm:"type:bigint;not null;index"`
	FromUserID     int64                    `gorm:"type:bigint;not null;default:0;index"`
	ToUserID       int64                    `gorm:"type:bigint;not null;default:0;index"`
	AssignType     string                   `gorm:"type:varchar(30);not null;default:'';index"`
	Reason         string                   `gorm:"type:varchar(255);not null;default:''"`
	Status         enums.IMAssignmentStatus `gorm:"type:int;not null;index"`
	CreatedAt      time.Time                `gorm:"type:datetime;not null;index"`
	FinishedAt     *time.Time               `gorm:"type:datetime"`
	OperatorID     int64                    `gorm:"type:bigint;not null;default:0;index"`
}

// ConversationTag 会话标签关联
type ConversationTag struct {
	ID             int64 `gorm:"primaryKey;autoIncrement"`
	ConversationID int64 `gorm:"type:bigint;not null;index;uniqueIndex:uk_conversation_tag"`
	TagID          int64 `gorm:"type:bigint;not null;index;uniqueIndex:uk_conversation_tag"`
	AuditFields
}

// QuickReply 快捷回复。
type QuickReply struct {
	ID        int64        `gorm:"primaryKey;autoIncrement"`
	GroupName string       `gorm:"type:varchar(50);not null;default:'';index"`
	Title     string       `gorm:"type:varchar(100);not null;default:'';index"`
	Content   string       `gorm:"type:text"`
	Status    enums.Status `gorm:"type:int;not null;index"`
	SortNo    int          `gorm:"type:int;not null;index"`
	AuditFields
}

// AIAgent AI 接待实例。
type AIAgent struct {
	ID               int64                           `gorm:"primaryKey;autoIncrement"`                    // ID 为 AI Agent 主键。
	Name             string                          `gorm:"type:varchar(100);not null;default:'';index"` // Name 为 AI Agent 名称。
	Description      string                          `gorm:"type:varchar(255);not null;default:''"`       // Description 为 AI Agent 描述。
	Status           enums.Status                    `gorm:"type:int;not null;index"`                     // Status 为 AI Agent
	AIConfigID       int64                           `gorm:"type:bigint;not null;default:0;index"`        // AIConfigID 为关联的 AI 配置ID。
	ServiceMode      enums.IMConversationServiceMode `gorm:"type:int;not null;default:3;index"`           // ServiceMode 为服务模式，如仅AI、仅人工、AI优先人工接管。
	SystemPrompt     string                          `gorm:"type:text"`                                   // SystemPrompt 为该 Agent 的系统提示词。
	WelcomeMessage   string                          `gorm:"type:text"`                                   // WelcomeMessage 为该 Agent 的欢迎语或首响模板。
	TeamIDs          string                          `gorm:"type:varchar(500);not null;default:''"`       // TeamIDs 为转人工时可路由的客服组ID列表，多个之间使用逗号分隔。
	HandoffMode      enums.AIAgentHandoffMode        `gorm:"type:int;not null;default:1"`                 // HandoffMode 为转人工模式，如进入待接入池、进入默认客服组待接入池。
	MaxAIReplyRounds int                             `gorm:"type:int;not null;default:2"`                 // MaxAIReplyRounds 为单个会话允许的 AI 最大成功回复次数，超过后强制转人工。
	FallbackMode     enums.AIAgentFallbackMode       `gorm:"type:int;not null;default:2"`                 // FallbackMode 为无答案或低置信度时的兜底模式。
	KnowledgeIDs     string                          `gorm:"type:varchar(500);not null;default:''"`       // KnowledgeIDs 为绑定的知识库ID列表，按顺序表示优先级。
	SortNo           int                             `gorm:"type:int;not null;default:0;index"`           // SortNo 为后台展示排序号。
	Remark           string                          `gorm:"type:text"`                                   // Remark 为备注。
	AuditFields
}

// WidgetSite 嵌入式客服站点配置。
type WidgetSite struct {
	ID        int64        `gorm:"primaryKey;autoIncrement"`                         // ID 为站点配置主键。
	Name      string       `gorm:"type:varchar(100);not null;default:'';index"`      // Name 为站点名称。
	AppID     string       `gorm:"type:varchar(64);not null;default:'';uniqueIndex"` // AppID 为站点唯一接入标识。
	AIAgentID int64        `gorm:"type:bigint;not null;default:0;index"`             // AIAgentID 为该站点默认接入的 AI Agent ID。
	Status    enums.Status `gorm:"type:int;not null;default:0;index"`                // Status 为站点状态。
	Remark    string       `gorm:"type:text"`                                        // Remark 为备注。
	AuditFields
}

// ConversationEventLog 会话事件日志。
type ConversationEventLog struct {
	ID             int64              `gorm:"primaryKey;autoIncrement"`
	ConversationID int64              `gorm:"type:bigint;not null;index"`
	EventType      enums.IMEventType  `gorm:"type:varchar(50);not null;default:'';index"`
	OperatorType   enums.IMSenderType `gorm:"type:varchar(30);not null;default:'';index"`
	OperatorID     int64              `gorm:"type:bigint;not null;default:0;index"`
	Content        string             `gorm:"type:text"`
	Payload        string             `gorm:"type:text"`
	CreatedAt      time.Time          `gorm:"type:datetime;not null;index"`
}

// AgentProfile 客服档案。
type AgentProfile struct {
	ID                    int64               `gorm:"primaryKey;autoIncrement"`                         // ID 为客服档案主键。
	UserID                int64               `gorm:"type:bigint;not null;uniqueIndex"`                 // UserID 关联后台用户，一名用户只允许一份客服档案。
	TeamID                int64               `gorm:"type:bigint;not null;default:0;index"`             // TeamID 为客服所属客服组。
	AgentCode             string              `gorm:"type:varchar(64);not null;default:'';uniqueIndex"` // AgentCode 为客服工号，用于业务侧识别客服。
	DisplayName           string              `gorm:"type:varchar(100);not null;default:'';index"`      // DisplayName 为客服展示名，可区别于后台昵称。
	Avatar                string              `gorm:"type:varchar(1024);not null;default:''"`           // Avatar 为客服头像 URL。
	ServiceStatus         enums.ServiceStatus `gorm:"type:int;not null;default:0;index"`                // ServiceStatus 表示客服服务状态：0空闲 1忙碌。
	MaxConcurrentCount    int                 `gorm:"type:int;not null;default:0"`                      // MaxConcurrentCount 表示客服最大并发接待数。
	PriorityLevel         int                 `gorm:"type:int;not null;default:0;index"`                // PriorityLevel 表示自动分配优先级，值越大越优先。
	AutoAssignEnabled     bool                `gorm:"not null;default:true;index"`                      // AutoAssignEnabled 表示是否参与自动分配。
	ReceiveOfflineMessage bool                `gorm:"not null;default:false"`                           // ReceiveOfflineMessage 表示离线时是否仍接收离线消息或转接消息。
	LastOnlineAt          *time.Time          `gorm:"type:datetime;index"`                              // LastOnlineAt 记录最近一次在线时间。
	LastStatusAt          *time.Time          `gorm:"type:datetime;index"`                              // LastStatusAt 记录最近一次状态变更时间。
	Status                enums.Status        `gorm:"type:int;not null;default:0;index"`                // Status 表示客服档案状态
	Remark                string              `gorm:"type:text"`                                        // Remark 记录客服备注信息。
	AuditFields
}

// AgentTeam 客服组。
type AgentTeam struct {
	ID           int64        `gorm:"primaryKey;autoIncrement"`                    // ID 为客服组主键。
	Name         string       `gorm:"type:varchar(100);not null;default:'';index"` // Name 为客服组名称。
	LeaderUserID int64        `gorm:"type:bigint;not null;default:0;index"`        // LeaderUserID 为组长用户ID，0 表示暂未设置。
	Status       enums.Status `gorm:"type:int;not null;default:0;index"`           // Status 表示客服组状态
	Description  string       `gorm:"type:varchar(255);not null;default:''"`       // Description 为客服组简介，用于说明职责边界。
	Remark       string       `gorm:"type:text"`                                   // Remark 记录客服组内部备注。
	AuditFields
}

// AgentTeamSchedule 客服组排班。
type AgentTeamSchedule struct {
	ID         int64        `gorm:"primaryKey;autoIncrement"`                   // ID 为组排班主键。
	TeamID     int64        `gorm:"type:bigint;not null;index"`                 // TeamID 为被排班的客服组ID。
	StartAt    time.Time    `gorm:"type:datetime;not null;index"`               // StartAt 为班次开始时间。
	EndAt      time.Time    `gorm:"type:datetime;not null;index"`               // EndAt 为班次结束时间。
	SourceType string       `gorm:"type:varchar(30);not null;default:'';index"` // SourceType 表示排班来源，如 manual、batch_import、template_generate。
	Remark     string       `gorm:"type:varchar(255);not null;default:''"`      // Remark 记录排班备注。
	Status     enums.Status `gorm:"type:int;not null;default:0;index"`          // Status 表示组排班记录状态。
	AuditFields
}

// AIConfig AI 统一配置。
// 一条记录表示一个可直接调用的 AI 配置实例，
// 同时包含厂商接入信息、模型信息和调用参数，不再拆分 endpoint/model 两层概念。
type AIConfig struct {
	ID               int64             `gorm:"primaryKey;autoIncrement"`                    // ID 为配置主键。
	Name             string            `gorm:"type:varchar(100);not null;default:'';index"` // Name 为配置名称，用于后台识别和展示。
	Provider         enums.AIProvider  `gorm:"type:varchar(50);not null;default:'';index"`  // Provider 为供应商标识，例如 openai、azure_openai、dashscope。
	BaseURL          string            `gorm:"type:varchar(255);not null;default:''"`       // BaseURL 为模型服务基础地址，例如 https://api.openai.com/v1。
	APIKey           string            `gorm:"type:varchar(255);not null;default:''"`       // APIKey 为服务端请求模型接口所需密钥。
	ModelType        enums.AIModelType `gorm:"type:varchar(30);not null;default:'';index"`  // ModelType 为模型类型，例如 llm、embedding、rerank。
	ModelName        string            `gorm:"type:varchar(100);not null;default:'';index"` // ModelName 为实际请求时传给上游的模型名。
	Dimension        int               `gorm:"type:int;not null;default:0"`                 // Dimension 为向量维度，仅 embedding 模型通常需要填写。
	MaxContextTokens int               `gorm:"type:int;not null;default:0"`                 // MaxContextTokens 为模型支持的最大上下文 token 数。
	MaxOutputTokens  int               `gorm:"type:int;not null;default:0"`                 // MaxOutputTokens 为模型建议的最大输出 token 数。
	TimeoutMS        int               `gorm:"type:int;not null;default:30000"`             // TimeoutMS 为调用该配置的默认超时时间，单位毫秒。
	MaxRetryCount    int               `gorm:"type:int;not null;default:0"`                 // MaxRetryCount 为默认最大重试次数。
	RPMLimit         int               `gorm:"type:int;not null;default:0"`                 // RPMLimit 为每分钟请求数限制，0 表示未显式配置。
	TPMLimit         int               `gorm:"type:int;not null;default:0"`                 // TPMLimit 为每分钟 token 数限制，0 表示未显式配置。
	Status           enums.Status      `gorm:"type:int;not null;index"`                     // Status 状态；同一 modelType 仅允许一条启用记录。
	SortNo           int               `gorm:"type:int;not null;index"`                     // SortNo 为排序号，用于后台展示和人工调整顺序。
	Remark           string            `gorm:"type:text"`                                   // Remark 为备注，用于记录用途、成本、限制和切换说明等补充信息。
	AuditFields
}

// TicketCategory 工单分类。
type TicketCategory struct {
	ID          int64        `gorm:"primaryKey;autoIncrement"`                   // ID 为分类主键。
	ParentID    int64        `gorm:"type:bigint;not null;default:0;index"`       // ParentID 为父分类ID，0表示顶级分类。
	Name        string       `gorm:"type:varchar(50);not null;default:'';index"` // Name 为分类名称。
	Code        string       `gorm:"type:varchar(50);not null;default:'';index"` // Code 为分类编码。
	Description string       `gorm:"type:varchar(255);not null;default:''"`      // Description 为分类描述。
	SortNo      int          `gorm:"type:int;not null;index"`                    // SortNo 为排序号。
	Status      enums.Status `gorm:"type:int;not null;index"`                    // Status 表示分类状态：
	Remark      string       `gorm:"type:text"`                                  // Remark 记录分类备注。
	AuditFields
}

// Ticket 工单。
type Ticket struct {
	ID                 int64      `gorm:"primaryKey;autoIncrement"`                    // ID 为工单主键。
	TicketNo           string     `gorm:"type:varchar(64);not null;uniqueIndex"`       // TicketNo 为工单编号，系统生成。
	Title              string     `gorm:"type:varchar(255);not null;default:''"`       // Title 为工单标题。
	Content            string     `gorm:"type:text"`                                   // Content 为工单内容。
	ChannelType        string     `gorm:"type:varchar(50);not null;default:'';index"`  // ChannelType 为来源渠道：conversation客服会话, self_service自助, admin后台创建, api接口。
	ChannelID          string     `gorm:"type:varchar(128);not null;default:'';index"` // ChannelID 为来源渠道ID，如会话ID。
	CategoryID         int64      `gorm:"type:bigint;not null;default:0;index"`        // CategoryID 为工单分类ID。
	Priority           int        `gorm:"type:int;not null;default:0;index"`           // Priority 为优先级：0普通 1低 2中 3高 4紧急。
	Status             int        `gorm:"type:int;not null;default:1;index"`           // Status 为工单状态：1待处理 2处理中 3待确认 4已解决 5已关闭 6已取消。
	SourceUserID       int64      `gorm:"type:bigint;not null;default:0;index"`        // SourceUserID 为提交人用户ID。
	ExternalUserID     string     `gorm:"type:varchar(128);not null;default:'';index"` // ExternalUserID 为外部用户ID（访客）。
	ExternalUserName   string     `gorm:"type:varchar(100);not null;default:''"`       // ExternalUserName 为外部用户名称。
	ExternalUserEmail  string     `gorm:"type:varchar(100);not null;default:''"`       // ExternalUserEmail 为外部用户邮箱。
	ExternalUserMobile string     `gorm:"type:varchar(32);not null;default:''"`        // ExternalUserMobile 为外部用户手机。
	CurrentAssigneeID  int64      `gorm:"type:bigint;not null;default:0;index"`        // CurrentAssigneeID 为当前处理人ID。
	CurrentTeamID      int64      `gorm:"type:bigint;not null;default:0;index"`        // CurrentTeamID 为当前处理组ID。
	ConversationID     int64      `gorm:"type:bigint;not null;default:0;index"`        // ConversationID 为关联的会话ID。
	ReplyCount         int        `gorm:"type:int;not null;default:0"`                 // ReplyCount 为回复次数。
	LastReplyAt        *time.Time `gorm:"type:datetime;index"`                         // LastReplyAt 为最后回复时间。
	LastReplyUserID    int64      `gorm:"type:bigint;not null;default:0"`              // LastReplyUserID 为最后回复人ID。
	Satisfied          *int       `gorm:"type:int;index"`                              // Satisfied 为满意度评价：1满意 2不满意 3未评价。
	SatisfiedRemark    string     `gorm:"type:varchar(255);not null;default:''"`       // SatisfiedRemark 为满意度备注。
	EvaluatedAt        *time.Time `gorm:"type:datetime"`                               // EvaluatedAt 为评价时间。
	ResolvedAt         *time.Time `gorm:"type:datetime;index"`                         // ResolvedAt 为解决时间。
	ClosedAt           *time.Time `gorm:"type:datetime;index"`                         // ClosedAt 为关闭时间。
	Tags               string     `gorm:"type:varchar(500);not null;default:''"`       // Tags 为标签，多个逗号分隔。
	Remark             string     `gorm:"type:text"`                                   // Remark 为内部备注。
	AuditFields
}

// TicketAssignment 工单分配记录。
type TicketAssignment struct {
	ID         int64      `gorm:"primaryKey;autoIncrement"`                   // ID 为分配记录主键。
	TicketID   int64      `gorm:"type:bigint;not null;index"`                 // TicketID 为工单ID。
	FromUserID int64      `gorm:"type:bigint;not null;default:0;index"`       // FromUserID 为分配人ID。
	FromTeamID int64      `gorm:"type:bigint;not null;default:0;index"`       // FromTeamID 为分配组ID。
	ToUserID   int64      `gorm:"type:bigint;not null;default:0;index"`       // ToUserID 为被分配人ID。
	ToTeamID   int64      `gorm:"type:bigint;not null;default:0;index"`       // ToTeamID 为被分配组ID。
	AssignType string     `gorm:"type:varchar(30);not null;default:'';index"` // AssignType 为分配类型：create创建, transfer转接, claim领取, withdraw退回, distribute分配。
	Reason     string     `gorm:"type:varchar(255);not null;default:''"`      // Reason 为分配原因。
	Status     int        `gorm:"type:int;not null;default:1;index"`          // Status 为状态：1进行中 2已接受 3已退回 4已取消。
	CreatedAt  time.Time  `gorm:"type:datetime;not null;index"`
	AcceptedAt *time.Time `gorm:"type:datetime"`                         // AcceptedAt 为接受时间。
	FinishedAt *time.Time `gorm:"type:datetime"`                         // FinishedAt 为完成时间。
	Remark     string     `gorm:"type:varchar(255);not null;default:''"` // Remark 为备注。
	AuditFields
}

// TicketReply 工单回复。
type TicketReply struct {
	ID            int64      `gorm:"primaryKey;autoIncrement"`                   // ID 为回复主键。
	TicketID      int64      `gorm:"type:bigint;not null;index"`                 // TicketID 为工单ID。
	ParentID      int64      `gorm:"type:bigint;not null;default:0;index"`       // ParentID 为父回复ID，用于支持对话式回复。
	Content       string     `gorm:"type:text"`                                  // Content 为回复内容。
	SenderType    string     `gorm:"type:varchar(30);not null;default:'';index"` // SenderType 为发送者类型：agent客服, customer客户, system系统。
	SenderID      int64      `gorm:"type:bigint;not null;default:0;index"`       // SenderID 为发送者ID。
	SenderName    string     `gorm:"type:varchar(100);not null;default:''"`      // SenderName 为发送者名称。
	IsInternal    bool       `gorm:"not null;default:false"`                     // IsInternal 为是否内部回复（客户不可见）。
	SendStatus    int        `gorm:"type:int;not null;default:1;index"`          // SendStatus 为发送状态：1待发送 2已发送 3发送失败。
	SentAt        *time.Time `gorm:"type:datetime;index"`                        // SentAt 为发送时间。
	ReadAt        *time.Time `gorm:"type:datetime"`                              // ReadAt 为读取时间。
	AttachmentIDs string     `gorm:"type:varchar(500);not null;default:''"`      // AttachmentIDs 为附件ID列表，逗号分隔。
	Remark        string     `gorm:"type:varchar(255);not null;default:''"`      // Remark 为备注。
	AuditFields
}

// TicketComment 工单备注。
type TicketComment struct {
	ID        int64     `gorm:"primaryKey;autoIncrement"`              // ID 为备注主键。
	TicketID  int64     `gorm:"type:bigint;not null;index"`            // TicketID 为工单ID。
	Content   string    `gorm:"type:text"`                             // Content 为备注内容。
	UserID    int64     `gorm:"type:bigint;not null;default:0;index"`  // UserID 为备注人ID。
	UserName  string    `gorm:"type:varchar(100);not null;default:''"` // UserName 为备注人名称。
	CreatedAt time.Time `gorm:"type:datetime;not null;index"`
}

// TicketAttachment 工单附件。
type TicketAttachment struct {
	ID           int64  `gorm:"primaryKey;autoIncrement"`              // ID 为主键。
	TicketID     int64  `gorm:"type:bigint;not null;index"`            // TicketID 为工单ID。
	ReplyID      int64  `gorm:"type:bigint;not null;default:0;index"`  // ReplyID 为关联的回复ID。
	FileName     string `gorm:"type:varchar(255);not null;default:''"` // FileName 为文件名。
	FileSize     int64  `gorm:"type:bigint;not null;default:0"`        // FileSize 为文件大小（字节）。
	FileType     string `gorm:"type:varchar(100);not null;default:''"` // FileType 为文件类型。
	FileURL      string `gorm:"type:varchar(500);not null;default:''"` // FileURL 为文件访问URL。
	FileKey      string `gorm:"type:varchar(255);not null;default:''"` // FileKey 为存储key。
	UploadUserID int64  `gorm:"type:bigint;not null;default:0;index"`  // UploadUserID 为上传人ID。
	Status       int    `gorm:"type:int;not null;default:1;index"`     // Status 为状态：1正常 2已删除。
	AuditFields
}

// TicketEventLog 工单事件日志。
type TicketEventLog struct {
	ID           int64     `gorm:"primaryKey;autoIncrement"`                   // ID 为主键。
	TicketID     int64     `gorm:"type:bigint;not null;index"`                 // TicketID 为工单ID。
	EventType    string    `gorm:"type:varchar(50);not null;default:'';index"` // EventType 为事件类型：create创建, update更新, status_change状态变更, assign分配, reply回复, comment备注, attachment上传附件, close关闭, delete删除。
	OperatorType string    `gorm:"type:varchar(30);not null;default:'';index"` // OperatorType 为操作人类型：agent客服, customer客户, system系统。
	OperatorID   int64     `gorm:"type:bigint;not null;default:0;index"`       // OperatorID 为操作人ID。
	OperatorName string    `gorm:"type:varchar(100);not null;default:''"`      // OperatorName 为操作人名称。
	Content      string    `gorm:"type:text"`                                  // Content 为事件描述。
	OldValue     string    `gorm:"type:text"`                                  // OldValue 为旧值（JSON）。
	NewValue     string    `gorm:"type:text"`                                  // NewValue 为新值（JSON）。
	CreatedAt    time.Time `gorm:"type:datetime;not null;index"`
}

// KnowledgeBase 知识库主表。
type KnowledgeBase struct {
	ID                    int64        `gorm:"primaryKey;autoIncrement"`                       // ID 为知识库主键。
	TenantID              int64        `gorm:"type:bigint;not null;default:0;index"`           // TenantID 为租户ID。
	Name                  string       `gorm:"type:varchar(100);not null;default:'';index"`    // Name 为知识库名称。
	Description           string       `gorm:"type:text"`                                      // Description 为知识库描述。
	Status                enums.Status `gorm:"type:int;not null;index"`                        // Status 为状态
	DefaultTopK           int          `gorm:"type:int;not null;default:10"`                   // DefaultTopK 为默认召回数量。
	DefaultScoreThreshold float64      `gorm:"type:decimal(5,4);not null;default:0.5"`         // DefaultScoreThreshold 为默认相似度阈值。
	DefaultRerankLimit    int          `gorm:"type:int;not null;default:5"`                    // DefaultRerankLimit 为默认重排后保留数量。
	ChunkProvider         string       `gorm:"type:varchar(30);not null;default:'structured'"` // ChunkProvider 为知识库分块策略 provider。
	ChunkTargetTokens     int          `gorm:"type:int;not null;default:300"`                  // ChunkTargetTokens 为目标 chunk token 数。
	ChunkMaxTokens        int          `gorm:"type:int;not null;default:400"`                  // ChunkMaxTokens 为单 chunk 最大 token 数。
	ChunkOverlapTokens    int          `gorm:"type:int;not null;default:40"`                   // ChunkOverlapTokens 为相邻 chunk 重叠 token 数。
	AnswerMode            int          `gorm:"type:int;not null;default:1"`                    // AnswerMode 为回答模式：1严格知识库模式 2辅助解释模式。
	FallbackMode          int          `gorm:"type:int;not null;default:1"`                    // FallbackMode 为兜底模式：1声明无答案 2引导换问法 3转人工。
	SortNo                int          `gorm:"type:int;not null;default:0;index"`              // SortNo 为排序号，用于后台展示和知识库的人工排序管理。
	Remark                string       `gorm:"type:text"`                                      // Remark 为备注。
	AuditFields
}

// KnowledgeDocument 知识文档主表。
type KnowledgeDocument struct {
	ID              int64                              `gorm:"primaryKey;autoIncrement"`                    // ID 为文档主键。
	TenantID        int64                              `gorm:"type:bigint;not null;default:0;index"`        // TenantID 为租户ID。
	KnowledgeBaseID int64                              `gorm:"type:bigint;not null;index"`                  // KnowledgeBaseID 为所属知识库ID。
	Title           string                             `gorm:"type:varchar(255);not null;default:'';index"` // Title 为文档标题。
	ContentType     enums.KnowledgeDocumentContentType `gorm:"type:varchar(20);not null;default:'html'"`    // ContentType 为内容类型：html/markdown。
	Content         string                             `gorm:"type:text"`                                   // Content 为文档内容。
	Status          enums.Status                       `gorm:"type:int;not null;default:0;index"`           // Status 为状态
	ContentHash     string                             `gorm:"type:varchar(64);not null;default:'';index"`  // ContentHash 为内容哈希，用于变更检测。
	AuditFields
}

// KnowledgeChunk 切片元数据表。
type KnowledgeChunk struct {
	ID              int64        `gorm:"primaryKey;autoIncrement"`                    // ID 为切片主键。
	TenantID        int64        `gorm:"type:bigint;not null;default:0;index"`        // TenantID 为租户ID。
	KnowledgeBaseID int64        `gorm:"type:bigint;not null;index"`                  // KnowledgeBaseID 为知识库ID。
	DocumentID      int64        `gorm:"type:bigint;not null;index"`                  // DocumentID 为文档ID。
	ChunkNo         int          `gorm:"type:int;not null;default:0;index"`           // ChunkNo 为切片序号。
	Title           string       `gorm:"type:varchar(255);not null;default:''"`       // Title 为切片标题。
	Content         string       `gorm:"type:text"`                                   // Content 为切片内容。
	ContentHash     string       `gorm:"type:varchar(64);not null;default:'';index"`  // ContentHash 为内容哈希。
	CharCount       int          `gorm:"type:int;not null;default:0"`                 // CharCount 为字符数。
	TokenCount      int          `gorm:"type:int;not null;default:0"`                 // TokenCount 为token数。
	Status          enums.Status `gorm:"type:int;not null;default:0;index"`           // Status 为状态：1有效 2已删除。
	VectorID        string       `gorm:"type:varchar(100);not null;default:'';index"` // VectorID 为向量库中的point ID。
	CreatedAt       time.Time    `gorm:"type:datetime;not null;index"`
	UpdatedAt       time.Time    `gorm:"type:datetime;not null;index"`
}

// KnowledgeRetrieveLog 检索日志表。
type KnowledgeRetrieveLog struct {
	ID                 int64     `gorm:"primaryKey;autoIncrement"`                   // ID 为日志主键。
	TenantID           int64     `gorm:"type:bigint;not null;default:0;index"`       // TenantID 为租户ID。
	KnowledgeBaseID    int64     `gorm:"type:bigint;not null;index"`                 // KnowledgeBaseID 为知识库ID。
	Channel            string    `gorm:"type:varchar(30);not null;default:'';index"` // Channel 为渠道：im会话, agent_assist坐席辅助, api开放接口, debug调试。
	Scene              string    `gorm:"type:varchar(50);not null;default:'';index"` // Scene 为场景：first_response首响, assist辅助, qa问答。
	SessionID          string    `gorm:"type:varchar(64);not null;default:'';index"` // SessionID 为会话ID。
	ConversationID     int64     `gorm:"type:bigint;not null;default:0;index"`       // ConversationID 为会话ID。
	RequestID          string    `gorm:"type:varchar(64);not null;default:'';index"` // RequestID 为请求ID。
	Question           string    `gorm:"type:text"`                                  // Question 为原始问题。
	RewriteQuestion    string    `gorm:"type:text"`                                  // RewriteQuestion 为改写后问题。
	Answer             string    `gorm:"type:text"`                                  // Answer 为生成的答案。
	AnswerStatus       int       `gorm:"type:int;not null;default:1;index"`          // AnswerStatus 为答案状态：1正常 2无答案 3兜底 4风控拦截。
	HitCount           int       `gorm:"type:int;not null;default:0"`                // HitCount 为命中数量。
	TopScore           float64   `gorm:"type:decimal(5,4);not null;default:0"`       // TopScore 为最高相似度分数。
	ChunkProvider      string    `gorm:"type:varchar(30);not null;default:'';index"` // ChunkProvider 为分块 provider。
	ChunkTargetTokens  int       `gorm:"type:int;not null;default:0"`                // ChunkTargetTokens 为目标 token 数。
	ChunkMaxTokens     int       `gorm:"type:int;not null;default:0"`                // ChunkMaxTokens 为最大 token 数。
	ChunkOverlapTokens int       `gorm:"type:int;not null;default:0"`                // ChunkOverlapTokens 为重叠 token 数。
	RerankEnabled      bool      `gorm:"not null;default:false;index"`               // RerankEnabled 是否启用 rerank。
	RerankLimit        int       `gorm:"type:int;not null;default:0"`                // RerankLimit 为 rerank 条数。
	CitationCount      int       `gorm:"type:int;not null;default:0"`                // CitationCount 为最终引用条数。
	UsedChunkCount     int       `gorm:"type:int;not null;default:0"`                // UsedChunkCount 为进入上下文的 chunk 数。
	LatencyMs          int64     `gorm:"type:bigint;not null;default:0"`             // LatencyMs 为总耗时毫秒。
	RetrieveMs         int64     `gorm:"type:bigint;not null;default:0"`             // RetrieveMs 为检索耗时毫秒。
	GenerateMs         int64     `gorm:"type:bigint;not null;default:0"`             // GenerateMs 为生成耗时毫秒。
	PromptTokens       int       `gorm:"type:int;not null;default:0"`                // PromptTokens 为prompt token数。
	CompletionTokens   int       `gorm:"type:int;not null;default:0"`                // CompletionTokens 为completion token数。
	ModelName          string    `gorm:"type:varchar(100);not null;default:''"`      // ModelName 为使用的模型名称。
	TraceData          string    `gorm:"type:text"`                                  // TraceData 为链路追踪数据JSON。
	CreatedAt          time.Time `gorm:"type:datetime;not null;index"`
}

// KnowledgeRetrieveHit 检索命中详情表。
type KnowledgeRetrieveHit struct {
	ID            int64     `gorm:"primaryKey;autoIncrement"`              // ID 为命中记录主键。
	RetrieveLogID int64     `gorm:"type:bigint;not null;index"`            // RetrieveLogID 为检索日志ID。
	ChunkID       int64     `gorm:"type:bigint;not null;index"`            // ChunkID 为切片ID。
	DocumentID    int64     `gorm:"type:bigint;not null;index"`            // DocumentID 为文档ID。
	DocumentTitle string    `gorm:"type:varchar(255);not null;default:''"` // DocumentTitle 为文档标题。
	ChunkNo       int       `gorm:"type:int;not null;default:0"`           // ChunkNo 为切片序号。
	Title         string    `gorm:"type:varchar(255);not null;default:''"` // Title 为切片标题。
	SectionPath   string    `gorm:"type:text"`                             // SectionPath 为章节路径。
	ChunkType     string    `gorm:"type:varchar(30);not null;default:''"`  // ChunkType 为切片类型。
	Provider      string    `gorm:"type:varchar(30);not null;default:''"`  // Provider 为分块 provider。
	RankNo        int       `gorm:"type:int;not null;default:0"`           // RankNo 为排名。
	Score         float64   `gorm:"type:decimal(5,4);not null;default:0"`  // Score 为相似度分数。
	RerankScore   float64   `gorm:"type:decimal(5,4);not null;default:0"`  // RerankScore 为重排分数。
	UsedInAnswer  bool      `gorm:"not null;default:false"`                // UsedInAnswer 是否用于生成答案。
	IsCitation    bool      `gorm:"not null;default:false"`                // IsCitation 是否作为引用返回。
	Snippet       string    `gorm:"type:text"`                             // Snippet 为内容片段。
	CreatedAt     time.Time `gorm:"type:datetime;not null;index"`
}

// KnowledgeFeedback 问答反馈表。
type KnowledgeFeedback struct {
	ID             int64     `gorm:"primaryKey;autoIncrement"`              // ID 为反馈主键。
	RetrieveLogID  int64     `gorm:"type:bigint;not null;index"`            // RetrieveLogID 为检索日志ID。
	FeedbackType   int       `gorm:"type:int;not null;default:1;index"`     // FeedbackType 为反馈类型：1点赞 2点踩 3无帮助 4引用错误 5其他。
	FeedbackReason string    `gorm:"type:varchar(500);not null;default:''"` // FeedbackReason 为反馈原因。
	UserID         int64     `gorm:"type:bigint;not null;default:0;index"`  // UserID 为用户ID。
	AgentID        int64     `gorm:"type:bigint;not null;default:0;index"`  // AgentID 为坐席ID。
	Remark         string    `gorm:"type:text"`                             // Remark 为备注。
	CreatedAt      time.Time `gorm:"type:datetime;not null;index"`
}

// SkillDefinition 表示可由后台配置并参与运行时路由的 Skill 定义。
type SkillDefinition struct {
	ID          int64        `gorm:"primaryKey;autoIncrement"`                          // ID 为 Skill 主键。
	Code        string       `gorm:"type:varchar(100);not null;default:'';uniqueIndex"` // Code 为 Skill 的稳定唯一编码，供程序内部引用和路由判断使用，例如 refund_skill。
	Name        string       `gorm:"type:varchar(100);not null;default:'';index"`       // Name 为 Skill 的展示名称，用于后台列表、配置页和人工选择场景。
	Description string       `gorm:"type:varchar(255);not null;default:''"`             // Description 为 Skill 的简要说明，用于描述该 Skill 的适用场景和职责边界。
	Prompt      string       `gorm:"type:longtext"`                                     // Prompt 为 Skill 的核心提示词，在命中后注入模型上下文参与执行。
	Priority    int          `gorm:"type:int;not null;default:0;index"`                 // Priority 为 Skill 命中冲突时的优先级，数值越大优先级越高。
	Status      enums.Status `gorm:"type:int;not null;default:0;index"`                 // Status 为 Skill 当前状态，使用全局通用状态：0启用 1禁用 2删除。
	Remark      string       `gorm:"type:text"`                                         // Remark 为后台备注，用于记录配置说明、维护信息或内部协作信息。
	AuditFields
}

// SkillRunLog 表示一次 Skill 运行过程的审计日志。
type SkillRunLog struct {
	ID                int64            `gorm:"primaryKey;autoIncrement"`                    // ID 为 Skill 运行日志主键。
	ConversationID    int64            `gorm:"type:bigint;not null;default:0;index"`        // ConversationID 为关联会话ID，无会话上下文时为0。
	AIAgentID         int64            `gorm:"type:bigint;not null;default:0;index"`        // AIAgentID 为本次运行所属的 AI Agent ID。
	AIConfigID        int64            `gorm:"type:bigint;not null;default:0;index"`        // AIConfigID 为本次运行实际使用的 AI 配置ID。
	SkillDefinitionID int64            `gorm:"type:bigint;not null;default:0;index"`        // SkillDefinitionID 为最终命中的 Skill 定义ID，未命中时为0。
	SkillCode         string           `gorm:"type:varchar(100);not null;default:'';index"` // SkillCode 为最终命中的 Skill 编码，未命中时为空。
	ManualSkillCode   string           `gorm:"type:varchar(100);not null;default:'';index"` // ManualSkillCode 为本次请求显式指定的 Skill 编码。
	IntentCode        string           `gorm:"type:varchar(100);not null;default:'';index"` // IntentCode 为上游传入的意图编码。
	UserMessage       string           `gorm:"type:longtext"`                               // UserMessage 为本次请求的用户输入内容。
	Matched           bool             `gorm:"not null;default:false;index"`                // Matched 表示本次请求是否命中了 Skill。
	MatchReason       string           `gorm:"type:varchar(500);not null;default:''"`       // MatchReason 为命中或未命中的原因说明。
	FinalSelected     bool             `gorm:"not null;default:false;index"`                // FinalSelected 表示该日志记录的 Skill 是否为最终选中的执行 Skill。
	UsedModel         string           `gorm:"type:varchar(100);not null;default:''"`       // UsedModel 为本次实际调用的模型名称。
	UsedProvider      enums.AIProvider `gorm:"type:varchar(50);not null;default:''"`        // UsedProvider 为本次实际调用的模型供应商。
	ErrorMessage      string           `gorm:"type:text"`                                   // ErrorMessage 为运行过程中的错误信息。
	CreatedAt         time.Time        `gorm:"type:datetime;not null;index"`                // CreatedAt 为运行日志创建时间。
}
