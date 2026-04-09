# Agent Team (客服组) 初始化文档

## 概述

Agent Team 初始化模块在 `cmd/testdata/agentteam/init.go` 中，用于创建客服组和客服人员的测试数据。

## 初始化过程

该模块在 AI Agent 初始化**之前**调用，按照以下顺序初始化：

1. **客服组创建**
   - 名称：`默认客服组`
   - 组长：Bootstrap 管理员用户（username: `admin`）
   - 状态：启用
   - 描述：`Local testdata seed - default service team`

2. **客服用户创建**
   - **客服A**：
     - 用户名：`agent_a`
     - 昵称：`客服A`
     - 工号：`AGENT_A`
   - **客服B**：
     - 用户名：`agent_b`
     - 昵称：`客服B`
     - 工号：`AGENT_B`
   - 默认密码：`ChangeMe123!`（与管理员相同）

3. **客服档案创建**
   - 为每个客服用户创建 `AgentProfile` 记录
   - 关联到 `默认客服组`
   - 初始设置：
     - 服务状态：`空闲（ServiceStatusIdle）`
     - 最大并发接待数：`5`
     - 自动分配优先级：`10`
     - 开启自动分配：`true`

## 返回结果

```go
type InitResult struct {
    TeamCreated      bool  // 是否新创建了客服组
    UsersCreated     int   // 新创建的用户数
    ProfilesCreated  int   // 新创建的客服档案数
    UpdatesApplied   int   // 更新的档案数
}
```

## 幂等性

该初始化过程是幂等的：

- **客服组**：通过 `name` 字段去重。若组已存在，则只更新组长信息
- **客服用户**：通过 `username` 字段去重。若用户已存在，则只更新昵称和状态
- **客服档案**：通过 `user_id` 字段去重。若档案已存在，则更新关键信息（组ID、工号、显示名等）

## 事务支持

所有数据库操作均使用事务包装：

```go
sqls.WithTransaction(func(ctx *sqls.TxContext) error { ... })
```

确保数据一致性，任何操作失败都会导致整个事务回滚。

## 审计字段

创建和更新的所有记录都包含标准审计字段：

```go
AuditFields{
    CreatedAt:      now,
    CreateUserID:   constants.SystemAuditUserID,      // 0
    CreateUserName: constants.SystemAuditUserName,    // "System"
    UpdatedAt:      now,
    UpdateUserID:   constants.SystemAuditUserID,      // 0
    UpdateUserName: constants.SystemAuditUserName,    // "System"
}
```

## 调用顺序

在 `cmd/testdata/main.go` 中的初始化顺序：

```
1. aiconfig.Init()      // AI Configuration
2. kb.Init()            // Knowledge Base
3. agentteam.Init()     // Agent Team ← 客服组初始化
4. aiagent.Init()       // AI Agent (可使用 TeamIDs 字段)
5. widgetsite.Init()    // Widget Site
```

## 测试运行

运行所有测试数据初始化：

```bash
cd cs-agent
go run cmd/testdata/main.go -config config/config.yaml
# 或使用 -yes 标志跳过确认
go run cmd/testdata/main.go -config config/config.yaml -yes
```

初始化输出示例：

```
2026-03-25T10:30:45.123+08:00 INFO testdata initialization completed
  droppedTables=45
  aiConfigSkipped=false
  aiConfigFile=cmd/testdata/aiconfig/ai_config.local.yaml
  aiConfigCreated=3
  aiConfigUpdated=0
  knowledgeBaseID=1
  kbChaptersTotal=120
  kbDocumentsCreated=120
  kbDocumentsUpdated=0
  agentTeamCreated=true
  agentTeamUsersCreated=2
  agentTeamProfilesCreated=2
  agentTeamUpdatesApplied=0
  aiAgentCreated=1
  aiAgentUpdated=0
  widgetSiteCreated=2
  widgetSiteUpdated=0
```

## 扩展说明

### 添加更多客服

如需在初始化中添加更多客服，修改 `initUsersAndProfiles()` 函数中的 `agentUsers` 切片：

```go
agentUsers := []struct {
    username string
    nickname string
    code     string
}{
    { username: "agent_a", nickname: "客服A", code: "AGENT_A" },
    { username: "agent_b", nickname: "客服B", code: "AGENT_B" },
    { username: "agent_c", nickname: "客服C", code: "AGENT_C" }, // 新增
}
```

### 自定义客服组名称和组长

修改 `initTeam()` 函数：

```go
func initTeam(leaderUserID int64) (bool, error) {
    teamName := "你的自定义组名"  // 修改这里
    // ...
}
```

## 相关模型

- `models.AgentTeam` - 客服组
- `models.User` - 用户账号
- `models.AgentProfile` - 客服档案
- `models.AuditFields` - 审计字段
- `enums.ServiceStatus` - 服务状态（空闲/忙碌）
- `enums.Status` - 通用状态

## 相关仓库

- `repositories.AgentTeamRepository`
- `repositories.UserRepository`
- `repositories.AgentProfileRepository`
