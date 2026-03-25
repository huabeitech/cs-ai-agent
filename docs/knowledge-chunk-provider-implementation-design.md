# 知识库 Chunk Provider 工程落地设计

本文档基于前一份 provider 方案文档，进一步细化到 Go 工程落地层，目标是明确：

1. provider 接口怎么定义
2. 代码目录怎么组织
3. 配置项怎么设计
4. 索引服务如何接入 provider
5. 第一阶段建议先实现哪些内容

适用范围：

- 当前 `Golang` 后端
- 当前知识库文档索引链路
- 当前 `KnowledgeChunk` / `KnowledgeDocument` / `KnowledgeIndexService`

## 1. 总体目标

当前系统的分块逻辑集中在：

- [internal/services/knowledge_index_service.go](internal/services/knowledge_index_service.go)

其主要问题是：

- 分块逻辑写死在 service 里
- 无法根据文档类型切换策略
- 不利于持续演进与实验

因此建议将分块逻辑从 service 中抽离，变成 provider 模式：

- service 负责调度
- provider 负责分块
- registry 负责选择 provider

## 2. 推荐目录结构

建议新增目录：

```text
internal/knowledge/
├── chunk/
│   ├── provider.go
│   ├── registry.go
│   ├── types.go
│   ├── options.go
│   ├── utils.go
│   ├── fixed_provider.go
│   ├── structured_provider.go
│   ├── faq_provider.go
│   └── semantic_provider.go
```

设计原则：

- `types.go` 放公共输入输出结构
- `provider.go` 放接口定义
- `registry.go` 放 provider 注册和选择逻辑
- 各 provider 独立文件实现

这样可以避免把知识库分块逻辑继续堆在 `services` 里。

## 3. 接口设计

建议定义统一接口：

```go
type Provider interface {
	Name() string
	Supports(contentType string) bool
	Chunk(ctx context.Context, req *ChunkRequest) ([]ChunkResult, error)
}
```

### 3.1 接口说明

#### `Name()`

返回 provider 唯一名称，例如：

- `fixed`
- `structured`
- `faq`
- `semantic`

#### `Supports(contentType string)`

用于声明 provider 是否支持某种内容类型。

例如：

- `fixed` 支持所有文本内容
- `structured` 支持 `html` / `markdown`
- `faq` 也可以支持 `html` / `markdown`

#### `Chunk(...)`

执行实际分块逻辑，返回标准 chunk 列表。

## 4. 输入输出结构设计

建议定义：

```go
type ChunkRequest struct {
	KnowledgeBaseID int64
	DocumentID      int64
	DocumentTitle   string
	ContentType     string
	Content         string
	Options         ChunkOptions
}

type ChunkOptions struct {
	Provider       string
	TargetTokens   int
	MaxTokens      int
	OverlapTokens  int
	EnableFallback bool
}

type ChunkResult struct {
	ChunkNo     int
	Title       string
	Content     string
	ChunkType   string
	SectionPath string
	CharCount   int
	TokenCount  int
	Metadata    map[string]any
}
```

### 4.1 设计原则

`ChunkRequest` 必须足够完整，避免 provider 再去读数据库。

`ChunkResult` 必须标准化，后续：

- 存数据库
- 生成 embedding
- 写入向量 payload

都依赖它。

## 5. Registry 设计

建议定义 registry：

```go
type Registry struct {
	providers map[string]Provider
}
```

核心方法建议包括：

```go
func NewRegistry() *Registry
func (r *Registry) Register(p Provider)
func (r *Registry) Get(name string) Provider
func (r *Registry) MustGet(name string) Provider
func (r *Registry) Resolve(name string, contentType string) Provider
```

### 5.1 `Resolve` 的职责

`Resolve` 负责做最终 provider 选择。

推荐规则：

1. 如果指定了 provider 且存在，优先用指定 provider
2. 如果指定 provider 不支持当前 contentType，则回退默认 provider
3. 如果未指定 provider，则用默认 provider
4. 如果默认 provider 不可用，则回退 `fixed`

## 6. Provider 选择策略

建议支持两级配置：

### 6.1 知识库级别

每个知识库可以配置自己的默认 chunk provider：

- `structured`
- `faq`
- `fixed`
- `semantic`

建议未来给 `KnowledgeBase` 增加字段：

- `chunk_provider`
- `chunk_options`

### 6.2 文档级别

文档级可以覆盖知识库默认配置。

比如：

- 某个知识库默认 `structured`
- 其中 FAQ 文档单独指定 `faq`

第一阶段如果不想改太多表结构，可以先只支持知识库级配置。

## 7. 与当前 KnowledgeIndexService 的衔接

当前 `IndexDocument()` 的核心流程是：

1. 读取文档
2. 提纯文本
3. 调 `chunkText`
4. 为每个 chunk 生成 embedding
5. 写向量库
6. 写 chunk 表

建议改造成：

1. 读取文档
2. 读取知识库 chunk 配置
3. 通过 registry 选择 provider
4. 调 provider 输出 `[]ChunkResult`
5. 为 chunk 生成 embedding
6. 写向量库 payload
7. 写 `KnowledgeChunk`

也就是说，`KnowledgeIndexService` 不再负责“怎么切”，只负责“调谁切”。

## 8. 数据库存储建议

当前 `KnowledgeChunk` 已有字段：

- `Title`
- `ContentHash`
- `CharCount`
- `TokenCount`

但目前没有被真正利用。

建议第一阶段新增字段：

1. `ChunkType`
2. `SectionPath`

可以考虑放在：

- `internal/models/models.go`
- 通过 `AutoMigrate` 更新

这样 chunk 数据才能真正支持：

- 类型区分
- section 合并
- 旧版本隔离

## 9. 向量 Payload 建议

建议统一写入这些 payload 字段：

```go
map[string]any{
	"knowledge_base_id": knowledgeBase.ID,
	"document_id":       document.ID,
	"chunk_no":          chunk.ChunkNo,
	"title":             chunk.Title,
	"chunk_type":        chunk.ChunkType,
	"section_path":      chunk.SectionPath,
	"content":           chunk.Content,
	"status":            chunk.Status,
}
```

这样后续可以支持：

- 更细的过滤
- 文档版本隔离
- 邻接块回溯
- 上下文增强

## 10. 第一阶段建议实现的 Provider

从工程投入产出比考虑，建议第一阶段只实现：

### 10.1 `fixedProvider`

用途：

- 兼容当前逻辑
- 作为回退 provider
- 作为 A/B 对照组

### 10.2 `structuredProvider`

用途：

- 作为系统默认 provider
- 覆盖大部分 HTML / Markdown 知识库

### 10.3 `faqProvider`

用途：

- 单独优化客服 FAQ 型知识

不建议第一阶段就落地 `semanticProvider`，但建议先把接口和预留结构设计好。

## 11. 每个 Provider 的工程实现建议

## 11.1 `fixedProvider`

建议做法：

- 输入纯文本
- 按 token 近似值切分
- 支持 overlap
- chunkType 统一标为 `text`

它是最简单的一版，不需要结构解析。

## 11.2 `structuredProvider`

建议做法：

1. 先将 HTML / Markdown 解析成 block 列表
2. block 类型包括：
   - `heading`
   - `paragraph`
   - `list`
   - `table`
   - `code`
3. 按 heading 切 section
4. section 内再按 token 控制切 chunk

输出时：

- `Title` 取当前 section 标题
- `SectionPath` 记录标题链
- `ChunkType` 根据 block 类型决定

## 11.3 `faqProvider`

建议做法：

1. 先识别 FAQ 问答对
2. 一个问答对尽量对应一个 chunk
3. 文本组织为：

```text
问题：xxx
回答：xxx
```

4. chunkType 标为 `faq`

如果答案过长，再在答案内部二次切分，但问题必须保留。

## 11.4 `semanticProvider`

建议做法：

第一阶段只定义接口和空实现占位，不急着投入生产。

后续正式实现时再考虑：

- 句子切分
- 句向量相似度
- 语义断点识别

## 12. 配置建议

建议在知识库配置层增加：

```go
type KnowledgeChunkConfig struct {
	Provider      string `json:"provider"`
	TargetTokens  int    `json:"targetTokens"`
	MaxTokens     int    `json:"maxTokens"`
	OverlapTokens int    `json:"overlapTokens"`
}
```

默认建议：

```text
provider = structured
targetTokens = 300
maxTokens = 400
overlapTokens = 40
```

FAQ 文档建议：

```text
provider = faq
targetTokens = 250
maxTokens = 350
overlapTokens = 20
```

## 13. 推荐的接入步骤

建议按下面顺序落地：

### 第一步

新增 chunk provider 目录和公共接口

### 第二步

将当前固定切分迁移为 `fixedProvider`

### 第三步

实现 `structuredProvider`

### 第四步

在 `KnowledgeIndexService` 中改为通过 registry 调用 provider

### 第五步

补充 `KnowledgeChunk` 元数据存储

### 第六步

支持知识库级 provider 配置

### 第七步

实现 `faqProvider`

## 14. 最终建议

当前项目最务实的工程路线是：

1. 抽 provider 接口
2. 保留 `fixedProvider`
3. 尽快实现 `structuredProvider`
4. 再补 `faqProvider`
5. `semanticProvider` 暂时预留

这条路线的优点是：

- 与现有代码兼容
- 便于逐步替换
- 能快速改善当前 RAG 召回质量
- 不会在第一阶段引入过多复杂度

## 15. 下一步建议

如果继续往实现推进，下一步最适合做的是：

1. 先把接口、registry、`fixedProvider` 搭起来
2. 再实现 `structuredProvider`
3. 最后改造 `KnowledgeIndexService` 接入这套 provider

这样改造风险最低，也最容易逐步验证效果。
