# AI Agent 初始化说明

## 初始化流程

`cmd/testdata/main.go` 中的数据初始化按以下顺序执行：

1. **删除所有现有表** - 清空数据库
2. **运行 migrations** - 创建表结构
3. **初始化 AI 配置** (`aiconfig.Init()`)
   - 种子文件：`cmd/testdata/aiconfig/ai_config.local.yaml`
   - 包含 LLM、Embedding、Rerank 三种模型配置
4. **初始化知识库** (`kb.Init()`)
   - 从 HTML 文件读取知识库数据
   - 创建知识库和文档
5. **初始化 AI Agent** (`aiagent.Init()`)
   - 依赖于 AI Config（需要启用的 LLM 配置）
   - 依赖于 Knowledge Base（需要已创建的知识库）
   - 创建测试 AI 凯服 Agent

## AI Agent 配置

### 种子数据字段说明

| 字段 | 说明 | 示例值 |
|-----|------|-------|
| `name` | Agent 唯一标识 | `测试AI客服` |
| `description` | Agent 描述 | `本地测试 AI 客服 Agent` |
| `status` | 启用状态 | `0`（启用） / `1`（禁用） |
| `aiConfigID` | 绑定的 AI 配置 ID | 自动获取启用的 LLM 配置 |
| `serviceMode` | 服务模式 | `3`（AI 优先）/ `1`（仅 AI）/ `2`（仅人工） |
| `systemPrompt` | 系统提示词 | 详见下方 |
| `welcomeMessage` | 欢迎语 | `您好，欢迎咨询！` |
| `teamIDs` | 可转接的客服组 ID | 多个 ID 逗号分隔（可为空） |
| `handoffMode` | 转人工模式 | `1`（待接入池）/ `2`（客服组池）/ `3`（AI托底） |
| `maxAIReplyRounds` | 单会话最大 AI 回复次数 | `5`（超过后强制转人工） |
| `fallbackMode` | 无答案兜底模式 | `1`（直接声明无答案）/ `2`（引导补充信息）/ `3`（转人工） |
| `knowledgeIDs` | 绑定的知识库 ID | 自动获取启用的知识库 |
| `sortNo` | 排序号 | `10`（用于后台展示排序） |
| `remark` | 备注 | `Local testdata seed` |

### 默认测试 Agent 配置

```yaml
名称: 测试AI客服
AI 配置: 最新启用的 LLM 配置
知识库: 最新启用的知识库（水浒传）
服务模式: AI 优先（AI 先回答，客户满意后关闭，不满意转人工）
系统提示词: 你是一个友好的客服助手，请用中文回答用户的问题。
最大回复次数: 5 轮（超过后强制转人工）
转人工模式: 进入待接入池
无答案处理: 引导补充信息或换个问法
```

## 使用步骤

### 1. 初始化测试数据

```bash
# 创建 AI Config 种子文件
cp cmd/testdata/aiconfig/ai_config.local.example.yaml cmd/testdata/aiconfig/ai_config.local.yaml

# 编辑，填入真实的 API 密钥
vim cmd/testdata/aiconfig/ai_config.local.yaml

# 执行初始化（自动模式）
go run cmd/testdata/main.go -yes

# 或交互模式，输入 INIT 确认
go run cmd/testdata/main.go
```

### 2. 验证初始化结果

```bash
# 查看数据库中的 AI Agent
mysql> SELECT id, name, status, service_mode FROM t_ai_agent;

# 查看日志输出
# 输出示例：
# {"level":"info","msg":"testdata initialization completed","droppedTables":38,"aiConfigSkipped":false,"aiConfigCreated":3,"aiConfigUpdated":0,"aiConfigFile":"cmd/testdata/aiconfig/ai_config.local.yaml","chapters":120,"createdDocuments":120,"updatedDocuments":0,"knowledgeBaseID":1,"aiAgentCreated":1,"aiAgentUpdated":0}
```

## 扩展和自定义

### 添加更多 AI Agent

在 `cmd/testdata/aiagent/init.go` 的 `buildSeedItems()` 函数中添加更多 Agent 配置：

```go
func buildSeedItems(aiConfigID, knowledgeID int64) []models.AIAgent {
	now := time.Now()
	return []models.AIAgent{
		{
			Name: "测试AI客服",
			// ... 现有配置
		},
		{
			Name: "投诉处理专员",
			Description: "专门处理投诉的 AI Agent",
			ServiceMode: enums.IMConversationServiceModeHumanOnly, // 仅人工
			SystemPrompt: "你是一个专业的投诉处理专员...",
			// ... 其他配置
		},
	}
}
```

### 修改默认 Agent 配置

直接编辑 `buildSeedItems()` 函数中的 Agent 定义，重新执行 `go run cmd/testdata/main.go -yes` 即可更新。

### 动态查询 AI 配置

如需在初始化时获取特定配置的 ID，修改 `getDefaultAIConfigID()` 和 `getDefaultKnowledgeID()` 函数：

```go
// 示例：获取特定提供商的配置
func getSpecificAIConfigID(modelType enums.AIModelType, provider enums.AIProvider) int64 {
	aiConfig := repositories.AIConfigRepository.Take(
		sqls.DB(),
		"model_type = ? AND provider = ? AND status = ?",
		string(modelType),
		string(provider),
		enums.StatusOk,
	)
	if aiConfig != nil {
		return aiConfig.ID
	}
	return 0
}
```

## 常见问题

**Q: 为什么 AI Agent 初始化失败？**
A: 检查以下几点：
- AI Config 是否初始化成功（需要至少一个启用的 LLM 配置）
- Knowledge Base 是否初始化成功（需要至少一个启用的知识库）
- 查看错误日志中的具体错误信息

**Q: 如何更改 Agent 的系统提示词？**
A: 编辑 `buildSeedItems()` 中的 `SystemPrompt` 字段，然后重新运行初始化。

**Q: 初始化时如何指定特定的 AI 配置？**
A: 修改 `getDefaultAIConfigID()` 函数，改为查询特定的配置名称或属性。
