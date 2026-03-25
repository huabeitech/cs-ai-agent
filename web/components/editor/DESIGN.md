# 统一编辑器设计（TipTap 单内核 + 双格式存储）

## 1. 背景与目标

当前项目存在两类编辑需求：

- `markdown`：知识文档等偏结构化内容
- `html` 富文本：所见即所得编辑、IM 消息等

现状是不同场景存在分叉实现，维护成本高，交互也不一致。  
本设计目标是在不破坏现有后端接口（`contentType + content`）前提下，统一前端编辑器体系。

### 目标

- 统一编辑内核：前端所有编辑场景尽量使用 TipTap/ProseMirror
- 兼容双格式：继续保留 `contentType = "markdown" | "html"`
- 保持可演进：后续可扩展图片上传、草稿、快捷键、插件化工具栏
- 降低迁移风险：分阶段替换，不一次性重写全部页面

### 非目标

- 不追求 markdown 与 html 的绝对无损双向转换
- 不在第一阶段实现完整协作编辑（OT/CRDT）

---

## 2. 核心结论

可以使用 TipTap 作为单一编辑内核，但不建议把 markdown/html 简化为“完全等价格式”。

- TipTap 的内部模型是 ProseMirror 文档，不是 markdown AST
- markdown 与 html 语义存在差异，复杂结构双向转换会有损
- 正确做法是：**内核统一 + 存储分型 + 转换可控**

---

## 3. 总体架构

建议目录（以 `web/components/editor` 为中心）：

```text
web/components/editor/
  index.tsx                # 统一入口组件 UnifiedEditor
  html.tsx                 # HtmlEditor（TipTap 配置）
  markdown.tsx             # MarkdownEditor（TipTap markdown 模式）
  viewer.tsx               # 统一只读渲染入口（可选）
  toolbar.tsx              # 通用工具栏（可按能力裁剪）
  schema.ts                # TipTap 扩展与能力分组
  convert.ts               # markdown/html 与 editor doc 的转换封装
  sanitize.ts              # HTML 白名单清洗（渲染前）
  types.ts                 # 统一类型定义
  DESIGN.md                # 本文档
```

---

## 4. 数据模型与接口

## 4.1 类型定义（建议）

```ts
export type EditorMode = "markdown" | "html"

export type EditorValue = {
  mode: EditorMode
  raw: string
}

export type UnifiedEditorProps = {
  value: EditorValue
  onChange: (next: EditorValue) => void
  placeholder?: string
  disabled?: boolean
  features?: {
    image?: boolean
    link?: boolean
    table?: boolean
    codeBlock?: boolean
  }
}
```

说明：

- `raw` 存放最终持久化内容（markdown 文本或 html 字符串）
- 外层业务不再关心“用什么编辑器实现”，只处理 `value/onChange`

## 4.2 现有接口兼容

与当前后端接口保持一致：

- `contentType` <- `value.mode`
- `content` <- `value.raw`

无需改后端数据结构。

---

## 5. 模式策略（重点）

## 5.1 html 模式

- 导入：`setContent(html)`
- 编辑：TipTap 常规富文本
- 导出：`editor.getHTML()`

## 5.2 markdown 模式

- 导入：`markdown -> editor doc`
- 编辑：仍使用 TipTap 内核（可配置 markdown 友好的工具栏）
- 导出：`editor doc -> markdown`

## 5.3 模式切换

当用户手动切换 `markdown/html` 时：

1. 弹出确认提示：告知可能发生格式损失
2. 用户可选：
   - 仅切换模式（保留原始内容，不做转换）
   - 执行转换（尝试 markdown/html 互转）
3. 转换失败时回退并提示错误原因

---

## 6. 转换与边界规则

## 6.1 支持稳定转换的子集

第一阶段建议仅保证以下元素稳定：

- 段落、标题（h1-h3）
- 粗体、斜体、删除线
- 无序/有序列表
- 引用
- 行内代码、代码块
- 链接
- 图片（基本属性）

## 6.2 明确有损边界

以下能力不承诺无损往返（可在 UI 上提示）：

- 复杂表格
- 自定义 HTML 属性与内联样式
- 任意嵌套块与第三方嵌入节点

---

## 7. 安全策略

所有 HTML 渲染都应先经过 sanitize，再进入 `dangerouslySetInnerHTML`。

建议白名单：

- 标签：`p`, `br`, `strong`, `em`, `del`, `blockquote`, `ul`, `ol`, `li`, `code`, `pre`, `a`, `img`, `h1`, `h2`, `h3`
- 属性：
  - `a`: `href`, `target`, `rel`
  - `img`: `src`, `alt`, `title`

安全要点：

- 禁止 `script`, `style`, `iframe` 等危险标签
- 过滤事件属性（如 `onclick`）
- 限制 `href/src` 协议（如仅 `http`, `https`, `data:image/*` 按需）

---

## 8. 组件分层建议

保持以下分层，避免页面散落编辑逻辑：

- `UnifiedEditor`：模式分发、通用 props、统一事件
- `HtmlEditor` / `MarkdownEditor`：各自实现细节
- `EditorToolbar`：按 `features` 开关按钮
- `convert.ts`：只做内容转换，不掺杂 UI
- `viewer.tsx`：只读渲染，统一 sanitize + 样式

---

## 9. 分阶段实施计划

## Phase 1（低风险统一入口）

- 在 `web/components/editor` 补齐 `index.tsx`、`html.tsx`、`markdown.tsx` 的最小实现
- 先迁移知识库文档编辑页到 `UnifiedEditor`
- 保持 IM 场景暂不动，避免一次性改动过大

验收标准：

- 业务页不再直接判断 `Textarea` vs `RichTextEditor`
- 保存结果与当前接口完全兼容

## Phase 2（收敛富文本能力）

- 抽象 TipTap schema/toolbar，沉淀为复用能力
- 迁移 IM 编辑器到统一内核配置（保留其发送快捷键与图片上传行为）

验收标准：

- 共享核心扩展与样式策略
- 场景差异通过 `features` 开关控制

## Phase 3（体验与可靠性增强）

- 引入草稿自动保存（localStorage 或服务端草稿）
- 增加快捷键、字数统计、粘贴规则统一
- 完善 sanitize 策略与回归测试

---

## 10. 测试建议

至少覆盖以下场景：

- markdown/html 各自编辑与保存
- 模式切换提示与转换失败回退
- 图片上传占位图替换成功/失败
- HTML 渲染安全（XSS 用例）
- 关键快捷键行为（Enter/Shift+Enter/Cmd+B）

---

## 11. 风险与取舍

- 风险：追求“全格式无损转换”会导致实现复杂度急剧上升
- 取舍：先定义“可稳定支持的语法子集”，其余场景用提示+降级策略处理
- 收益：统一内核后，后续功能（草稿、插件、统计、主题）可一次开发多处复用

---

## 12. 与当前项目的直接对应

建议优先替换知识库文档编辑页中的分支逻辑（markdown 文本域 vs html 富文本），统一接入 `UnifiedEditor`。  
IM 编辑器可在下一阶段迁移到同一核心配置，避免破坏现有发送交互。

