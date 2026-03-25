# 后端前端枚举 AST 生成规范

本文档定义本项目中“后端常量生成前端枚举”的统一规范。目标是保证：

- 后端是唯一事实源
- 前端枚举、值、文案完全由后端生成
- `cmd/enums/generator.go` 可以稳定通过 AST 识别并生成

## 1. 适用范围

- 后端枚举定义目录：`internal/pkg/enums`
- 前端生成文件：`web/enums/generated/enums.ts`
- 生成命令：`make enums`

只有 `internal/pkg/enums/*.go` 中符合规范的枚举定义才会被 AST 生成器识别。

## 2. 总体原则

- 前后端共用的枚举只允许在后端定义一次
- 前端禁止手写同语义的业务枚举
- 前端命名、枚举值、标签文案全部以后端为准
- 枚举定义修改后，必须重新执行 `make enums`

## 3. 生成规则

AST 生成器当前遵循以下规则：

1. 扫描 `internal/pkg/enums/*.go`
2. 收集 `type`、`const`、`xxxLabelMap`
3. 只有存在对应 `LabelMap` 的常量组才生成前端枚举
4. 前端枚举名优先使用后端类型名
5. 如果是非强类型枚举，则优先根据 `xxxLabelMap` 变量名推导枚举名
6. 前端枚举成员名通过“常量名去掉枚举名前缀”得到

因此，命名必须稳定、统一、可推导。

## 4. 强制推荐写法

### 4.1 优先使用强类型枚举

必须优先使用 typed const，不要新增无类型的 `int/string` 常量组。

推荐写法：

```go
type Status int

const (
	StatusOk       Status = 0
	StatusDisabled Status = 1
	StatusDeleted  Status = 2
)

var statusLabelMap = map[Status]string{
	StatusOk:       "启用",
	StatusDisabled: "禁用",
	StatusDeleted:  "已删除",
}
```

字符串枚举同样遵循该规则：

```go
type AIProvider string

const (
	AIProviderOpenAI AIProvider = "openai"
)

var aiProviderLabelMap = map[AIProvider]string{
	AIProviderOpenAI: "OpenAI",
}
```

## 5. 命名规范

### 5.1 类型名

- 类型名就是前端最终枚举名
- 必须使用清晰、完整、稳定的 PascalCase 命名

示例：

- `Status`
- `IMConversationStatus`
- `KnowledgeDocumentStatus`

### 5.2 常量名

- 同一枚举组中的常量名必须以类型名为前缀
- 常量后缀必须表达明确语义

示例：

- `StatusOk`
- `StatusDisabled`
- `IMConversationStatusPending`
- `KnowledgeDocumentStatusPublished`

对应前端会生成：

- `Status.Ok`
- `Status.Disabled`
- `IMConversationStatus.Pending`
- `KnowledgeDocumentStatus.Published`

### 5.3 LabelMap 名称

- 标签映射变量必须命名为 `xxxLabelMap`
- 命名必须与该枚举组语义一致
- 不允许一个 `LabelMap` 混合多个枚举组

推荐示例：

- `statusLabelMap`
- `imConversationStatusLabelMap`
- `knowledgeDocumentStatusLabelMap`

## 6. LabelMap 规范

### 6.1 必须直接使用枚举常量作为 key

正确：

```go
var statusLabelMap = map[Status]string{
	StatusOk: "启用",
}
```

错误：

```go
var statusLabelMap = map[Status]string{
	0: "启用",
}
```

### 6.2 LabelMap 必须完整覆盖该枚举组

- 枚举组中需要暴露给前端的每个常量都必须出现在对应 `LabelMap` 中
- 不允许缺项
- 不允许混入其他枚举组的 key

### 6.3 标签文案以后端为准

- 前端禁止再定义一套独立文案
- 如需改文案，必须修改后端 `LabelMap`，然后重新生成

## 7. 不推荐和禁止的写法

### 7.1 不推荐：非强类型枚举

虽然生成器兼容部分 untyped 场景，但新代码禁止继续新增：

```go
const (
	KnowledgeAnswerStatusNormal = 1
	KnowledgeAnswerStatusNoAnswer = 2
)

var knowledgeAnswerStatusLabelMap = map[int]string{
	KnowledgeAnswerStatusNormal: "正常",
	KnowledgeAnswerStatusNoAnswer: "无答案",
}
```

应逐步收敛为：

```go
type KnowledgeAnswerStatus int
```

### 7.2 禁止：常量前缀与类型名不一致

错误示例：

```go
type Status int

const (
	CommonStatusOk Status = 0
)
```

### 7.3 禁止：没有 LabelMap 却期望前端生成

没有 `LabelMap` 的常量组不会生成到前端。

### 7.4 禁止：前端手写同语义枚举

禁止在以下位置重复定义业务枚举：

- `web/lib/*`
- `web/app/*`
- `web/components/*`

前端统一直接使用：

```ts
import { Status, StatusLabels } from "@/enums/generated/enums"
```

## 8. 新增或修改枚举的标准流程

1. 在 `internal/pkg/enums` 中新增或修改枚举类型
2. 按规范定义常量和 `LabelMap`
3. 执行 `make enums`
4. 如有前端使用方，改为引用 `@/enums/generated/enums`
5. 执行 `cd web && pnpm typecheck`

## 9. 前端使用约定

- 业务枚举统一从 `@/enums/generated/enums` 导入
- 枚举工具函数统一从 `@/lib/enums` 导入
- `web/lib/enums.ts` 只允许放通用工具函数，不允许再放业务枚举定义

推荐写法：

```ts
import { Status, StatusLabels } from "@/enums/generated/enums"
import { getEnumLabel, getEnumOptions } from "@/lib/enums"
```

## 10. 现阶段执行要求

- 新增枚举必须遵循本文档
- 修改已有枚举时，如条件允许，优先顺手收敛到 typed const
- 任何会影响 AST 稳定识别的命名调整，都必须同步验证 `make enums`
