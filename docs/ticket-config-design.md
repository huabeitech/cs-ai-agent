# 工单配置中心设计

本文档用于定义工单配置中心的产品与技术方案，目标是在当前工单模型已经具备 `categoryId`、`resolutionCode`、`TicketSLARecord` 等基础字段的前提下，补齐工单配置能力，让工单真正从“可存储”升级为“可管理、可运营、可演进”的正式业务对象。

本文档重点覆盖：

- 工单分类配置
- 工单解决码配置
- 工单 SLA 配置
- 配置中心与工单主流程的关系
- 后端模型、接口与前端页面建议
- 第一阶段落地范围与验收标准

## 1. 背景

### 1.1 当前现状

当前工单模块已经具备以下基础能力：

- 工单主表 `Ticket`
- 工单评论 `TicketComment`
- 工单事件日志 `TicketEventLog`
- 工单关注人 `TicketWatcher`
- 工单 SLA 记录 `TicketSLARecord`
- 工单状态流转、指派、回复、关闭、重开
- 会话转工单

从底层字段看，当前模型已经预留了：

- `categoryId`
- `type`
- `priority`
- `severity`
- `resolutionCode`
- `customFieldsJson`

但从产品落地看，当前仍存在明显缺口：

- `categoryId` 只是字段，没有真正的分类配置中心
- `resolutionCode` 只是字段，没有可维护的解决码体系
- SLA 目标值目前由代码内固定逻辑生成，不是配置驱动
- 工单表单和详情页没有充分使用这些结构化字段
- 后续自动分派、统计报表、规则配置缺少统一配置基础

### 1.2 为什么要先做配置中心

工单系统如果没有配置中心，会出现以下问题：

- 字段虽然存在，但实际无法稳定使用
- 客服创建工单时只能填自由文本，导致数据不一致
- 主管无法按分类、解决码、SLA 做队列管理和报表分析
- 后续自动化规则、AI 分类建议、批量处理都缺少标准字段基础
- 每次新增工单结构字段都需要改代码，无法沉淀运营能力

因此，工单配置中心不是“后台补页面”，而是工单工作台和运营体系的底座。

## 2. 设计目标

本期配置中心的目标如下：

- 让工单拥有可维护的分类体系
- 让工单解决结果拥有标准解决码
- 让工单 SLA 目标值由后台配置驱动
- 让新建工单、编辑工单、状态流转直接使用这些配置
- 为后续批量操作、自动分派、SLA 运营、报表和 AI 建议打基础

## 3. 范围定义

### 3.1 本期包含

- 工单分类配置
- 工单解决码配置
- 工单 SLA 配置
- 配置项在工单创建、编辑、解决流程中的接入
- 配置项在列表页和详情页中的展示接入

### 3.2 本期不包含

- 动态表单引擎
- 自定义字段设计器
- 通用规则引擎
- 客户端字段配置
- 多租户隔离配置
- 非工单业务对象的统一配置中心

## 4. 配置对象设计

## 4.1 TicketCategory

### 4.1.1 定位

`TicketCategory` 用于定义工单分类，用来承载问题归属的业务语义。分类主要用于：

- 客服建单时选择问题类别
- 队列筛选与主管视图
- 自动分派和后续规则触发
- 报表统计和问题分布分析
- AI 分类建议落点

### 4.1.2 字段建议

- `id`
- `name`
- `code`
- `parentId`
- `sortNo`
- `status`
- `remark`
- `AuditFields`

### 4.1.3 产品规则

- 分类名称必填
- 分类编码必填，系统内唯一
- 第一阶段支持两级分类即可，不建议一开始开放无限层级
- 停用分类后：
  - 历史工单仍保留原分类
  - 新建工单不可再选
- 若工单已使用某分类，不允许物理删除，建议逻辑停用

### 4.1.4 使用场景

第一阶段典型分类示例：

- 售前咨询
- 售后问题
- 退款/投诉
- 技术支持
- 账号权限
- 订单履约

## 4.2 TicketResolutionCode

### 4.2.1 定位

`TicketResolutionCode` 用于定义工单“如何被解决”这一结果维度，用来承载结案后的标准化原因。

解决码主要用于：

- 解决工单时记录结案结果
- 后续做结案质量分析
- 区分“客户已知悉”“产品修复”“配置调整”“误报”等不同解决方式
- 为运营复盘和知识沉淀提供结构化依据

### 4.2.2 字段建议

- `id`
- `name`
- `code`
- `sortNo`
- `status`
- `remark`
- `AuditFields`

### 4.2.3 产品规则

- 解决码名称必填
- 解决码编码必填，系统内唯一
- 停用后不可用于新工单解决
- 历史数据保留
- 第一阶段先只在“改为 resolved”时使用
- 是否在“关闭 closed”时也强制要求解决码，第一期建议不强制

### 4.2.4 典型解决码示例

- `answered_by_agent` 客服答复解决
- `customer_confirmed` 客户确认已解决
- `bug_fixed` 产品修复
- `config_updated` 配置调整
- `duplicate_ticket` 重复工单
- `known_issue` 已知问题说明
- `cannot_reproduce` 无法复现

## 4.3 TicketSLAConfig

### 4.3.1 定位

`TicketSLAConfig` 用于定义工单 SLA 目标值。第一阶段不做复杂策略引擎，先按优先级配置首响和解决时长。

### 4.3.2 字段建议

- `id`
- `name`
- `priority`
- `firstResponseMinutes`
- `resolutionMinutes`
- `status`
- `remark`
- `AuditFields`

### 4.3.3 产品规则

- 每个优先级最多只有一条启用中的 SLA 配置
- `firstResponseMinutes` 必须大于 0
- `resolutionMinutes` 必须大于 0
- 停用配置后，不影响历史工单已有 SLA 记录
- 新工单创建时读取当前启用配置
- 如果某优先级未配置 SLA：
  - 建议第一阶段使用系统默认值兜底
  - 同时在后台页面给予明确提示

### 4.3.4 默认 SLA 示例

- 低优先级：首响 120 分钟，解决 1440 分钟
- 普通优先级：首响 60 分钟，解决 720 分钟
- 高优先级：首响 30 分钟，解决 240 分钟
- 紧急优先级：首响 10 分钟，解决 60 分钟

## 5. 与工单主流程的关系

## 5.1 新建工单

新建工单时：

- 可选择工单分类
- 可选择优先级和严重度
- 系统根据优先级自动初始化 SLA
- 若分类未来需要驱动默认团队、默认表单、默认优先级，第一阶段先预留，不立即实现

## 5.2 编辑工单

编辑工单时：

- 可修改分类
- 可修改类型
- 不直接重算已有 SLA 记录
- 若后续业务要求“改优先级后重算 SLA”，建议另行定义规则，不在本期默认启用

## 5.3 工单解决

工单改为 `resolved` 时：

- 可填写 `resolutionCode`
- 可填写 `resolutionSummary`
- 解决码来自配置中心
- 未填写解决码是否允许提交：
  - 第一阶段建议“允许不填，但推荐填写”
  - 若你们更强调运营分析，也可以改为“必填”

## 5.4 工单关闭

工单改为 `closed` 时：

- 必须填写关闭原因 `closeReason`
- 本期不强制要求解决码
- 若工单已被 resolved 再 close，可保留原解决码

## 6. 页面设计

## 6.1 分类管理页

建议路径：

- `web/app/(console)/ticket-categories/page.tsx`

页面功能：

- 分类列表
- 创建分类
- 编辑分类
- 启停分类
- 支持父级分类选择
- 支持排序

建议列表列：

- 名称
- 编码
- 父级分类
- 状态
- 排序
- 更新时间
- 操作

## 6.2 解决码管理页

建议路径：

- `web/app/(console)/ticket-resolution-codes/page.tsx`

页面功能：

- 解决码列表
- 创建解决码
- 编辑解决码
- 启停解决码
- 排序

建议列表列：

- 名称
- 编码
- 状态
- 排序
- 更新时间
- 操作

## 6.3 SLA 配置页

建议路径：

- `web/app/(console)/ticket-sla-configs/page.tsx`

页面功能：

- SLA 配置列表
- 新增 SLA 配置
- 编辑 SLA 配置
- 按优先级查看首响和解决时长
- 启停配置

建议列表列：

- 名称
- 优先级
- 首响时长
- 解决时长
- 状态
- 更新时间
- 操作

## 7. 工单页面接入建议

## 7.1 新建/编辑工单弹窗接入

当前工单编辑弹窗应补充：

- 分类字段
- 类型字段
- 必要时预留自定义字段区域

第一阶段不建议一开始就开放复杂动态字段，只需要把分类接进来。

## 7.2 工单状态变更弹窗接入

当前状态变更弹窗在改为 `resolved` 时应补充：

- 解决码下拉
- 解决摘要输入

改为 `closed` 时保留关闭原因输入。

## 7.3 工单列表接入

列表页建议增加：

- 分类列
- 分类筛选
- 解决码可先不放列表列，但为后续报表预留

## 7.4 工单详情页接入

详情页建议展示：

- 分类
- 类型
- 解决码
- 解决摘要
- SLA 信息

## 8. 后端设计

## 8.1 模型建议

建议新增以下模型：

- `TicketCategory`
- `TicketResolutionCode`
- `TicketSLAConfig`

命名和字段风格遵循项目现有规范，统一放入 `internal/models/models.go`。

## 8.2 分层建议

- `models`：定义配置实体
- `repositories`：CRUD 和列表查询
- `services`：配置校验、状态约束、启停逻辑
- `builders`：转换 response DTO
- `controllers`：接口暴露

## 8.3 接口建议

### 分类

- `ANY /api/console/ticket-category/list`
- `POST /api/console/ticket-category/create`
- `POST /api/console/ticket-category/update`
- `POST /api/console/ticket-category/delete`

### 解决码

- `ANY /api/console/ticket-resolution-code/list`
- `POST /api/console/ticket-resolution-code/create`
- `POST /api/console/ticket-resolution-code/update`
- `POST /api/console/ticket-resolution-code/delete`

### SLA 配置

- `ANY /api/console/ticket-sla-config/list`
- `POST /api/console/ticket-sla-config/create`
- `POST /api/console/ticket-sla-config/update`
- `POST /api/console/ticket-sla-config/delete`

### 给工单表单用的选项接口

如果前端希望更轻量，也可增加统一 options 接口：

- `GET /api/console/ticket/options`

返回内容建议包含：

- 分类列表
- 解决码列表
- 优先级枚举
- 严重度枚举

## 8.4 服务接入点

### CreateTicket

创建工单时：

- 校验分类是否存在且启用
- 根据优先级读取 SLA 配置并初始化 `TicketSLARecord`
- 替换当前写死的 `ticketSLATargetMinutes(...)`

### UpdateTicket

更新工单时：

- 校验分类是否合法
- 不默认重建 SLA

### ChangeStatus

改为 `resolved` 时：

- 校验解决码是否合法
- 写入 `resolutionCode`
- 写入 `resolutionSummary`

## 9. 前端设计

## 9.1 API 封装建议

建议新增：

- `web/lib/api/ticket-config.ts`

建议方法：

- `fetchTicketCategories`
- `createTicketCategory`
- `updateTicketCategory`
- `deleteTicketCategory`
- `fetchTicketResolutionCodes`
- `createTicketResolutionCode`
- `updateTicketResolutionCode`
- `deleteTicketResolutionCode`
- `fetchTicketSLAConfigs`
- `createTicketSLAConfig`
- `updateTicketSLAConfig`
- `deleteTicketSLAConfig`
- `fetchTicketOptions`

## 9.2 页面建议

建议页面目录：

```text
web/app/(console)/
├── ticket-categories/
│   └── page.tsx
├── ticket-resolution-codes/
│   └── page.tsx
└── ticket-sla-configs/
    └── page.tsx
```

## 9.3 表单建议

- 分类选择优先使用项目已有下拉封装
- 解决码选择同样走 `option-combobox`
- SLA 配置页面优先做简单表单，不做拖拽或复杂矩阵编辑

## 10. 数据兼容与迁移

## 10.1 历史工单

历史工单可能存在：

- `categoryId = 0`
- `resolutionCode = ""`
- SLA 记录由旧逻辑生成

兼容策略建议：

- 历史数据允许为空
- 新建数据按新规则执行
- 旧工单详情页若无分类/解决码，前端显示为空，不做强制修复

## 10.2 SLA 迁移

当前 SLA 初始化来自代码逻辑。改为配置驱动后：

- 新工单读取 `TicketSLAConfig`
- 历史工单不回刷
- 若担心配置缺失，可保留代码默认值作为兜底

## 11. 权限建议

建议新增权限点：

- 查看工单分类
- 管理工单分类
- 查看解决码
- 管理解决码
- 查看 SLA 配置
- 管理 SLA 配置

第一阶段可按管理员或主管角色开放。

## 12. 第一阶段验收标准

满足以下条件即可视为配置中心第一阶段完成：

1. 后台可维护工单分类
2. 后台可维护解决码
3. 后台可维护 SLA 配置
4. 工单创建时可选择分类
5. 工单解决时可选择解决码
6. SLA 初始化来自后台配置，而非纯硬编码
7. 列表和详情页能展示分类、解决码、SLA 信息
8. 历史工单不因新配置上线而异常

## 13. 后续演进方向

配置中心第一阶段完成后，可继续演进：

- 分类绑定默认团队
- 分类绑定默认优先级
- 分类绑定动态字段模板
- 分类驱动自动分派规则
- 解决码驱动知识沉淀和复盘统计
- 更复杂的 SLA 策略
  - 按来源
  - 按团队
  - 按分类
  - 按客户等级

## 14. 总结

工单配置中心的价值不在于“再多几张后台配置页”，而在于把工单从一个带状态的文本记录，升级成一个有标准结构、可分析、可自动化、可规模化运转的正式服务对象。

第一阶段只做分类、解决码、SLA 三类配置，范围足够聚焦，同时又能直接带动工单表单、详情、队列、SLA 运营和后续规则能力的落地，是当前最值得优先推进的一步。
