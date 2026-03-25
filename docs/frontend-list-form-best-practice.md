# 前端后台列表/表单最佳实践

本文档用于约束本项目后台类 CRUD 页面在前端的实现方式，供后续 AI Agent 直接复用。基线案例来源于 `web/app/dashboard/quick-replies`。

## 1. 适用范围

- `web/app/dashboard/*` 下的后台列表页
- 典型场景：列表查询、筛选、新建、编辑、删除、启用/禁用
- 页面形态：`page.tsx` 承载列表，`_components/edit.tsx` 承载弹窗表单，弹窗使用组件 `web/components/project-dialog.tsx`

## 2. 总体分层

后台 CRUD 页面统一采用“两层结构”：

1. `page.tsx`：负责列表、筛选、异步状态、toast、弹窗开关、接口调用
2. `_components/edit.tsx`：负责表单渲染、字段校验、默认值回填、payload 转换

强制要求：

- 页面层可以调用 `@/lib/api/*`
- 表单层不直接请求接口，只通过 `onSubmit` 向上抛出 payload
- 创建和编辑优先复用同一个弹窗组件，不要拆成两个几乎重复的表单

推荐示例：

- [page.tsx](web/app/dashboard/quick-replies/page.tsx)
- [edit.tsx](web/app/dashboard/quick-replies/_components/edit.tsx)

## 3. 列表页规范

### 3.1 状态组织

列表页通常包含以下 state：

- 筛选条件：如 `keyword`、`statusFilter`、`groupName`
- 列表加载状态：如 `loading`
- 提交状态：如 `saving`
- 行级操作状态：如 `actionLoadingId`
- 弹窗状态：如 `dialogOpen`
- 当前编辑项：如 `editingItem`
- 列表结果：如 `result`

约束：

- 列表加载和写操作状态要分离，不要复用一个 `loading`
- 行级操作优先使用 `actionLoadingId`，不要把整页锁死
- `editingItem === null` 表示创建态，有值表示编辑态

### 3.2 数据加载

列表加载必须收敛到单一函数，例如 `loadQuickReplies()`。

该函数应统一被以下场景复用：

- 首屏加载
- 手动点击刷新
- 创建成功后刷新
- 编辑成功后刷新
- 删除成功后刷新
- 启用/禁用成功后刷新

要求：

- 使用 `useCallback` 包装加载函数
- `useEffect` 中仅调用该函数，不要重复拼装请求参数
- 筛选参数在加载函数内部统一 `trim` 和转换

推荐模式：

```tsx
const loadList = useCallback(async () => {
  setLoading(true)
  try {
    const data = await fetchList({
      keyword: keyword.trim() || undefined,
      status: statusFilter === "all" ? undefined : statusFilter,
    })
    setResult(data)
  } catch (error) {
    toast.error(error instanceof Error ? error.message : "加载失败")
  } finally {
    setLoading(false)
  }
}, [keyword, statusFilter])

useEffect(() => {
  void loadList()
}, [loadList])
```

### 3.3 筛选区

筛选区建议放在列表上方，包含：

- 关键字输入框
- 枚举状态下拉框
- 刷新按钮
- 新建按钮

要求：

- 输入框保持受控
- 下拉框状态值统一使用字符串
- “全部” 之类的筛选项统一使用显式值，如 `"all"`
- 刷新按钮直接复用列表加载函数

### 3.4 表格区

列表展示优先使用现有 `shadcn/ui` 的 `Table` 组件，不重复封装基础表格组件。

要求：

- 列表为空且非加载中时，必须展示空状态行
- 操作列固定放右侧
- 文本较长内容允许摘要展示，如 `line-clamp-2`
- 状态展示优先使用 `Badge`

### 3.5 行级操作

行级操作统一由 `page.tsx` 负责，不放在表单组件内。

典型操作：

- 编辑
- 启用/禁用
- 删除

要求：

- 操作执行时设置 `actionLoadingId`
- 成功后使用 `toast.success`
- 成功后统一回调 `loadList()`
- 失败后使用 `toast.error`

### 3.6 分页

分页使用组件 `web/components/list-pagination.tsx`

## 4. 表单弹窗规范

### 4.1 技术选型

后台表单统一使用：

- `react-hook-form`
- `zod`
- `shadcn/ui` 的 `Field` 体系

字段组件组合优先使用：

- `Field`
- `FieldLabel`
- `FieldContent`
- `FieldError`

说明：

- 原生输入类组件使用 `register`
- `Select`、`Switch` 等非原生受控组件使用 `Controller`

### 4.2 表单文件结构

表单组件内至少拆出以下结构：

1. `schema`
2. `FormValues` 类型
3. `emptyForm`
4. `buildForm(item)`
5. `buildPayload(form)`

职责要求：

- `schema`：定义字段合法性
- `emptyForm`：定义创建态默认值
- `buildForm(item)`：将接口对象映射为表单值
- `buildPayload(form)`：将表单值映射为接口 payload

禁止：

- 在 JSX 内直接写复杂转换逻辑
- 在 `onSubmit` 里混合回填、校验、转换三种职责

### 4.3 校验规则

字段校验统一交给 `zod`，字段错误优先显示在字段下方。

要求：

- 必填字段统一用 `trim().min(1, "...")`
- 枚举字段使用 `z.enum(...)`
- 数字输入若以字符串形式接收，先校验字符串，再在 `buildPayload` 中转换
- 不要只依赖 `toast.error` 做字段校验提示

推荐模式：

```tsx
const schema = z.object({
  title: z.string().trim().min(1, "标题不能为空"),
  status: z.enum(["0", "1"], { message: "请选择状态" }),
  sortNo: z.string().trim().regex(/^\d+$/, "排序值必须是大于等于 0 的整数"),
})
```

### 4.4 默认值与回填

编辑弹窗打开时，必须能正确回填当前记录；创建弹窗打开时，必须恢复到空表单。

要求：

- `useForm({ defaultValues: buildForm(item) })`
- 当 `item` 变化时调用 `reset(buildForm(item))`
- `status` 这类枚举值在表单层统一使用字符串，不要混用数字和字符串

### 4.5 提交

表单提交只负责将合法表单值传给外层 `onSubmit`，不直接请求接口。

要求：

- `form` 标签使用 `onSubmit={handleSubmit(...)}`
- 提交按钮显式写 `type="submit"`
- 取消按钮显式写 `type="button"`
- `saving` 状态下禁用按钮，避免重复提交

## 5. 异常与提示规范

- 字段校验错误：显示在 `FieldError`
- 接口请求错误：显示 `toast.error`
- 写操作成功：显示 `toast.success`
- 列表加载失败：由页面层统一提示，不在表单层处理

禁止：

- 同一个字段错误同时在字段下方和 toast 重复提示
- 表单组件内部直接吞掉接口异常

## 6. API 调用规范

API 调用统一放在 `@/lib/api/*` 服务层，不在页面或组件中写裸 `fetch`。

要求：

- 列表：`fetchXxx`
- 创建：`createXxx`
- 更新：`updateXxx`
- 删除：`deleteXxx`

页面层负责编排这些 API 的调用顺序和刷新时机。

## 7. Quick Replies 基线总结

`web/app/dashboard/quick-replies` 可以作为标准后台列表/表单页面参考：

- [page.tsx](web/app/dashboard/quick-replies/page.tsx)：列表、筛选、行操作、弹窗状态、提交后刷新
- [edit.tsx](web/app/dashboard/quick-replies/_components/edit.tsx)：`react-hook-form + zod + Field` 表单、默认值回填、payload 转换

后续 AI Agent 在实现相似页面时，应优先复用该结构，而不是重新设计页面分层。

## 8. AI Agent 执行要求

当 AI Agent 需要新增或修改后台列表/表单页面时，默认按以下顺序执行：

1. 先参考本文档确定页面分层
2. 再参考 `quick-replies` 的实现方式确定状态组织
3. 新增表单时优先复用 `react-hook-form + zod + Field`
4. 所有写操作成功后，回到页面层统一刷新列表
5. 改动后至少执行 `cd web && pnpm typecheck`

如需偏离本文档，必须在变更说明中明确说明原因。
