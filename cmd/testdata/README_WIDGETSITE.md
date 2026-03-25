# WidgetSite 初始化说明

## 初始化流程

`cmd/testdata/main.go` 中的数据初始化按以下顺序执行：

1. **删除所有现有表** - 清空数据库
2. **运行 migrations** - 创建表结构
3. **初始化 AI 配置** (`aiconfig.Init()`)
4. **初始化知识库** (`kb.Init()`)
5. **初始化 AI Agent** (`aiagent.Init()`)
6. **初始化 WidgetSite** (`widgetsite.Init()`)

## WidgetSite 说明

WidgetSite 代表一个嵌入式客服小组件的站点配置，用于客户将客服小组件集成到自己的网站或应用中。

### 配置字段

| 字段 | 说明 | 示例值 |
|-----|------|-------|
| `name` | 站点名称 | `官网客服` |
| `appId` | 站点唯一标识（接入 token） | `widget_official_site` |
| `aiAgentId` | 绑定的 AI Agent ID | 自动获取启用的 Agent |
| `status` | 站点状态 | `0`（启用） / `1`（禁用） |
| `remark` | 备注 | `Local testdata seed` |

### 默认测试 WidgetSite 配置

系统初始化时会创建两个预设站点：

#### 1. 官网客服
- **名称**：官网客服
- **AppID**：`widget_official_site`
- **绑定 Agent**：自动获取第一个启用的 AI Agent
- **用途**：官方网站集成的客服小组件

#### 2. APP 客服
- **名称**：APP 客服
- **AppID**：`widget_mobile_app`
- **绑定 Agent**：自动获取第一个启用的 AI Agent
- **用途**：移动应用集成的客服小组件

## 使用步骤

### 初始化测试数据

```bash
# 执行初始化（自动模式）
go run cmd/testdata/main.go -yes

# 或交互模式，输入 INIT 确认
go run cmd/testdata/main.go
```

### 验证初始化结果

```bash
# 查看数据库中的 WidgetSite
mysql> SELECT id, name, app_id, status FROM t_widget_site;

# 示例输出：
# | id | name      | app_id                | status |
# |----|-----------|----------------------|--------|
# | 1  | 官网客服   | widget_official_site | 0      |
# | 2  | APP客服    | widget_mobile_app    | 0      |
```

### 查看初始化日志

```bash
# 输出示例（包含 WidgetSite 相关信息）：
# {"level":"info","msg":"testdata initialization completed",...,"widgetSiteCreated":2,"widgetSiteUpdated":0}
```

## 扩展和自定义

### 添加更多 WidgetSite

在 `cmd/testdata/widgetsite/init.go` 的 `buildSeedItems()` 函数中添加更多配置：

```go
func buildSeedItems(aiAgentID int64) []models.WidgetSite {
	now := time.Now()
	return []models.WidgetSite{
		{
			Name: "官网客服",
			AppID: "widget_official_site",
			// ... 现有配置
		},
		{
			Name:      "帮助中心客服",
			AppID:     "widget_help_center",
			AIAgentID: aiAgentID,
			Status:    enums.StatusOk,
			Remark:    "Local testdata seed",
			AuditFields: models.AuditFields{
				CreatedAt:      now,
				CreateUserID:   0,
				CreateUserName: "System",
				UpdatedAt:      now,
				UpdateUserID:   0,
				UpdateUserName: "System",
			},
		},
	}
}
```

### 修改默认配置

直接编辑 `buildSeedItems()` 函数中的 WidgetSite 定义，重新执行初始化即可更新。

### 指定特定的 AI Agent

修改 `getDefaultAIAgentID()` 函数，改为查询特定 Agent：

```go
// 示例：获取名称为特定值的 Agent
func getSpecificAIAgentID(name string) int64 {
	aiAgent := repositories.AIAgentRepository.Take(
		sqls.DB(),
		"name = ? AND status = ?",
		name,
		enums.StatusOk,
	)
	if aiAgent != nil {
		return aiAgent.ID
	}
	return 0
}
```

## 常见问题

**Q: 为什么 WidgetSite 初始化失败？**
A: 检查以下几点：
- AI Agent 是否初始化成功（需要至少一个启用的 Agent）
- AppID 是否重复（AppID 是唯一约束）
- 查看错误日志中的具体错误信息

**Q: 如何在初始化时使用不同的 Agent？**
A: 修改 `buildSeedItems()` 中的 `aiAgentID`，或修改 `getDefaultAIAgentID()` 函数以获取特定 Agent。

**Q: WidgetSite 初始化后无法访问？**
A: 确保站点状态为 `0`（启用）。如需禁用，将 `Status` 改为 `1`。

**Q: 如何更新现有的 WidgetSite 配置？**
A: 修改 `buildSeedItems()` 中的配置，重新执行 `go run cmd/testdata/main.go -yes`，系统会自动更新（基于 AppID 匹配）。
