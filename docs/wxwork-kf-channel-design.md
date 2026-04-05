# 企业微信客服渠道接入设计

## 1. 文档目标

本文档描述“企业微信客服”接入当前客服体系的第一期设计方案。

设计目标：

- 将企业微信客服作为一个新的客户接入渠道，接入现有客服会话体系
- 复用当前 `conversation`、`message`、`customer_identity` 等核心模型
- 客服统一在本系统后台接待，不依赖企业微信原生客服工作台进行人工接待
- 首期支持企业微信客户的 `text`、`image`、`file` 三类入站消息
- 平台侧下行发送采用 `outbox` 异步方案
- 平台状态主导会话生命周期，企业微信事件只做同步修正

已确认的业务边界：

- 第一期只支持“后台接待，企业微信只做客户入口”
- 第一阶段只支持 `text/image/file`，其余消息先降级展示
- 下行发送接受 `outbox` 异步方案
- 会话关闭以平台为准，微信 `session_status_change` 只做同步修正

## 2. 当前实现现状

当前项目中企业微信客服回调链路已经具备“接收回调”的基础能力，但尚未接入客服业务闭环。

现有入口：

- `internal/controllers/third/wechat_controller.go`
  - `GetCallback()` 用于企业微信回调 URL 校验
  - `PostCallback()` 负责读取请求体、解密回调消息、交给分发器处理
- `internal/wxwork/callback_dispatcher.go`
  - 提供回调 handler 注册和异步消费能力
- `internal/services/wx_callback_handlers/kf_msg_or_event_handler.go`
  - 已注册 `event:kf_msg_or_event` 的 handler
  - 当前逻辑仅调用 `SyncMsg()` 拉取消息并打印日志

当前缺口：

- 未持久化 `next_cursor`
- 未做消息幂等
- 未将微信客户映射到现有 `ExternalSource + ExternalID`
- 未将微信消息转成现有 `conversation/message` 数据
- 未接通人工客服下行回复到企业微信
- 未处理 `enter_session`、`msg_send_fail`、`session_status_change` 等关键事件

因此，当前状态可以认为是“回调可收到，但未真正接入客服体系”。

## 3. 总体设计思路

第一期不单独建设一套“微信客服业务域”，而是将企业微信客服视作现有客服系统中的一个新渠道。

核心设计原则：

- 平台客服体系主导，微信只是渠道层
- 现有 `conversation`、`message`、`customer_identity` 继续复用
- 企业微信渠道专有数据单独建表承载，不污染通用 IM 主表
- 微信入站消息统一转换成平台侧“客户消息”
- 平台人工回复统一转换成微信下行消息
- 所有与微信 API 交互的高失败率动作，均通过异步机制处理

一句话概括：

> 企业微信客服 = 新的 `ExternalSource` 渠道 + 渠道映射层 + 入站同步服务 + 出站投递服务

## 4. 业务边界

### 4.1 第一期要做的事情

- 企业微信客户消息进入平台后，自动映射或创建平台会话
- 平台客服可在现有后台查看和处理这些会话
- 企业微信客户发送的文本、图片、文件能够在平台侧展示
- 平台客服发送的文本消息可异步发送给企业微信客户
- 微信关键事件可同步记录到平台

### 4.2 第一期不做的事情

- 不支持企业微信原生客服工作台和本平台双端协同接待
- 不做企业微信客服坐席与平台坐席的双向实时绑定
- 不支持所有微信消息类型的完整富语义映射
- 不在第一期实现图片/文件的微信下行发送
- 不实现欢迎语、菜单、小程序卡片等高级能力
- 不自动从企业微信补齐完整客户资料画像

## 5. 现有能力复用方式

企业微信接入后，尽量复用现有开放 IM 与客服会话能力。

复用对象：

- `conversation`
  - 作为平台统一会话主表
- `message`
  - 作为平台统一消息主表
- `conversation_participant`
  - 记录客户参与方
- `conversation_read_state`
  - 复用读游标机制
- `customer_identity`
  - 复用客户第三方身份映射
- `conversation_event_log`
  - 记录渠道事件、状态修正、发送失败等
- `ConversationService`
  - 复用会话创建、分配、转接、关闭
- `MessageService`
  - 复用消息落库、未读计算、事件发布、AI 触发

企业微信客户的身份映射方式：

- `ExternalSource = wxwork_kf`
- `ExternalID = external_userid`

这意味着：

- 企业微信客户在平台中仍被视为“外部客户”
- 渠道身份与 CRM 客户的关联继续通过 `customer_identity` 维护

## 6. 渠道专有数据设计

为了避免把微信渠道的专有字段塞进通用会话表，建议新增以下表。

### 6.1 `wxwork_kf_sync_state`

用途：

- 按 `open_kfid` 保存企业微信客服消息同步游标
- 支撑 `SyncMsg` 增量拉取

建议字段：

- `id`
- `open_kfid`
- `next_cursor`
- `last_sync_at`
- `status`
- `remark`
- 审计字段

约束建议：

- `open_kfid` 唯一索引

### 6.2 `wxwork_kf_conversation`

用途：

- 维护平台 `conversation` 与企业微信客服上下文的映射关系
- 为下行发送提供 `open_kfid` 和 `external_userid`
- 记录微信侧会话状态及最近交互信息

建议字段：

- `id`
- `conversation_id`
- `open_kfid`
- `external_userid`
- `servicer_userid`
- `session_status`
- `last_wx_msg_time`
- `last_wx_msg_id`
- `raw_profile`
- `status`
- 审计字段

约束建议：

- `conversation_id` 唯一索引
- `open_kfid + external_userid` 索引

### 6.3 `wxwork_kf_message_ref`

用途：

- 做微信消息与平台消息之间的幂等映射
- 保存微信 `msgid`
- 保存原始消息 JSON 以便排障和补偿

建议字段：

- `id`
- `conversation_id`
- `message_id`
- `wx_msg_id`
- `direction`
- `origin`
- `open_kfid`
- `external_userid`
- `raw_payload`
- `send_status`
- `fail_reason`
- `status`
- 审计字段

约束建议：

- `wx_msg_id` 唯一索引
- `message_id` 索引

### 6.4 `channel_message_outbox`

用途：

- 保存平台待投递到外部渠道的消息任务
- 保证平台消息已提交后再异步调用渠道接口

建议字段：

- `id`
- `channel_type`
- `conversation_id`
- `message_id`
- `payload`
- `send_status`
- `retry_count`
- `next_retry_at`
- `last_error`
- `sent_at`
- 审计字段

约束建议：

- `channel_type + message_id` 唯一索引

## 7. 枚举与基础对象扩展

### 7.1 扩展外部来源枚举

文件：

- `internal/pkg/enums/external_identity.go`

新增：

- `ExternalSourceWxWorkKF = "wxwork_kf"`

说明：

- 该来源用于第三方企业微信客服回调
- 不应加入 `IsAllowedOpenImExternalSource`
- 因为它不属于 `/api/open/im/*` 开放 IM 前端入口

### 7.2 新增微信渠道内部枚举

建议新增文件：

- `internal/pkg/enums/wxwork_kf.go`

建议内容：

- 消息方向：`in/out`
- 出站状态：`pending/sending/sent/failed`
- 微信会话状态枚举
- 微信事件类型常量

## 8. 入站回调处理设计

### 8.1 当前回调入口保持不变

保持现有职责划分：

- controller 只负责验签、解密、转交
- callback dispatcher 只负责注册和分发
- service 负责实际业务编排

因此：

- `internal/controllers/third/wechat_controller.go` 不承担业务逻辑
- `internal/wxwork/callback_dispatcher.go` 继续作为回调总线

### 8.2 `kf_msg_or_event` 的处理方式

当前 `kf_msg_or_event_handler` 只做 `SyncMsg()` + 日志。

建议改造为：

1. 根据 `message.OpenKfID` 读取同步状态
2. 从 `wxwork_kf_sync_state.next_cursor` 开始增量拉取
3. 分页调用企业微信 `SyncMsg`
4. 每条消息交给入站业务 service 处理
5. 当前页全部处理完成后更新 `next_cursor`

建议新增 service：

- `internal/services/wxwork_kf_inbound_service.go`

该 service 负责：

- 拉取增量消息
- 消息分类
- 幂等判断
- 渠道映射
- 写入平台消息
- 更新同步状态

## 9. 入站消息落平台的业务流程

### 9.1 统一流程

企业微信同步到一条消息后，按如下流程处理：

1. 解析微信原始消息
2. 判断是否已消费过
3. 构造平台外部身份
4. 查找或创建平台会话
5. 维护渠道会话映射
6. 将微信消息转换为平台消息并落库
7. 保存微信消息映射关系

### 9.2 幂等判断

幂等的唯一键使用微信 `msgid`。

处理前步骤：

- 先查 `wxwork_kf_message_ref` 中是否已存在 `wx_msg_id`
- 已存在则直接跳过
- 不存在才继续处理

这样可以解决：

- 企业微信重复回调
- `SyncMsg` 重试导致的重复拉取
- 服务重启后的重复消费

### 9.3 构造外部身份

统一构造为：

- `ExternalSource = wxwork_kf`
- `ExternalID = external_userid`
- `ExternalName` 第一阶段可为空或回退为 `external_userid`

该对象将直接复用现有 `openidentity.ExternalInfo` 的语义。

### 9.4 会话查找与创建

优先根据以下条件查找未结束会话：

- `external_source = wxwork_kf`
- `external_id = external_userid`
- `status in (pending, active)`

若不存在，则调用现有 `ConversationService.Create(...)` 创建会话。

### 9.5 AI Agent 归属策略

企业微信客服渠道需要明确一个默认的 `AIAgentID`。

第一期建议采用最小方案：

- 配置一个全局默认 `wxwork_kf` 渠道 AI Agent

后续如需要更细粒度能力，可扩展为：

- `open_kfid -> ai_agent_id`

首期不建议在一开始就做复杂映射。

## 10. 入站消息类型映射方案

第一期仅完整支持以下三类：

- `text`
- `image`
- `file`

### 10.1 文本消息

微信消息：

- `msgtype = text`

平台映射：

- `message_type = text`
- `content = text.content`
- `payload = ""`

### 10.2 图片消息

微信消息：

- `msgtype = image`
- 关键数据为 `image.media_id`

平台映射：

- `message_type = image`
- `content = "[图片]"` 或可展示的简短摘要
- `payload` 保存原始微信 JSON 或标准化后的媒体信息

第一期说明：

- 不强制把微信图片下载并转成本平台资产
- 先以“消息可见、可追踪”为目标

### 10.3 文件消息

微信消息：

- `msgtype = file`
- 关键数据为 `file.media_id`

平台映射：

- `message_type = attachment`
- `content = "[附件]"`
- `payload` 保存原始微信 JSON 或标准化媒体信息

### 10.4 其余类型的降级处理

第一期未完整支持的消息类型，例如：

- `voice`
- `video`
- `location`
- `link`
- `business_card`
- `miniprogram`

统一降级原则：

- 不阻断主流程
- 平台可看见该消息
- 原始 JSON 保存在 `payload` 或 `wxwork_kf_message_ref.raw_payload`
- `content` 用占位文案，例如：
  - `[语音]`
  - `[视频]`
  - `[位置]`
  - `[链接]`

## 11. 微信事件处理方案

第一期仅处理关键事件。

### 11.1 `enter_session`

建议处理：

- 确保平台会话存在，不存在则创建
- 更新 `wxwork_kf_conversation`
- 写平台会话事件日志，例如“微信客户进入会话”

第一期暂不处理：

- 微信原生欢迎语
- 自动欢迎语回复

### 11.2 `session_status_change`

建议处理原则：

- 微信事件只做同步修正，不主导平台状态
- 只更新渠道映射表状态和事件日志
- 不直接覆盖平台 `conversation.status`

例如：

- `change_type = 1`
  - 记录“微信会话接入”
- `change_type = 2`
  - 记录“微信会话转接”
- `change_type = 3`
  - 记录“微信会话结束”
  - 同步更新 `wxwork_kf_conversation.session_status`
  - 不直接关闭平台会话

原因：

- 本项目已明确由平台主导会话关闭与状态流转
- 微信只是渠道层，不应反向篡改平台业务主状态

### 11.3 `msg_send_fail`

建议处理：

- 通过 `fail_msgid` 关联 `wxwork_kf_message_ref`
- 更新对应 `channel_message_outbox` 或消息映射记录
- 标记发送失败并记录失败原因
- 写入平台会话事件日志

常见失败原因应原样保存，便于排障：

- 超过 48 小时发送窗口
- 超过 5 条限制
- 用户拒收
- 客服账号已删除
- 应用关闭等

## 12. 平台下行发送设计

### 12.1 设计原则

下行发送必须与平台消息落库解耦。

原因：

- 企业微信接口调用存在网络抖动和限流风险
- 不应把第三方接口失败绑定到主业务事务
- 平台消息应优先保证落库成功

因此采用：

- 平台消息落库成功
- 写入 `channel_message_outbox`
- 异步 worker 调企业微信发送

### 12.2 改造位置

当前控制台发消息链路：

- `ConversationController.PostSend_message()`
- `MessageService.SendAgentMessage(...)`

建议改造方式：

- controller 层不做渠道分支
- `MessageService.SendAgentMessage(...)` 成功后，根据会话渠道决定是否插入 outbox

也就是说：

- 非企业微信渠道：行为不变
- 企业微信渠道：额外生成一条 outbox 任务

### 12.3 Outbox 消费流程

建议新增：

- `internal/services/channel_message_outbox_service.go`
- `internal/services/wxwork_kf_outbound_service.go`

异步消费流程：

1. 读取待发送 outbox
2. 根据 `message_id` 查询平台消息
3. 根据 `conversation_id` 查询 `wxwork_kf_conversation`
4. 组装企业微信发送请求
5. 调用 `kf.SendMsg(...)`
6. 成功则更新 outbox 状态并写消息映射
7. 失败则记录错误、安排重试或标记失败

## 13. 第一阶段下行消息能力范围

### 13.1 文本消息

第一期必须支持文本下行。

平台消息：

- `message_type = text`

微信发送结构：

- `sendmsg.Text`

映射字段：

- `touser = external_userid`
- `open_kfid = open_kfid`
- `text.content = 平台消息 content`

### 13.2 图片与文件消息

第一期不建议承诺完整支持图片/文件下行。

原因：

- 微信下行图片/文件依赖素材上传
- 当前项目尚未具备“平台资产 -> 微信 media_id”的完整能力

因此第一期建议：

- 下行文本先闭环
- 图片/文件下行暂缓到第二阶段增强

如果后续要支持，需要新增：

- 资产上传到企业微信临时素材接口
- 拿到 `media_id`
- 再调用 `kf.SendMsg`

## 14. 与现有 AI / 人工接待逻辑的关系

由于企业微信渠道接入后仍走现有 `ConversationService` 和 `MessageService`，因此现有 AI / 人工逻辑可继续复用。

具体表现：

- 企业微信客户消息进入平台后，按“客户消息”处理
- 现有 `MessageService.SendCustomerMessage(...)` 会继续触发 AI 回复链路
- 现有待接入、分配、转接、关闭流程继续生效
- 人工客服仍使用当前控制台处理会话

这意味着：

- 企业微信只是新入口
- 客服平台主工作流不需要重做

## 15. 建议的代码组织

按照项目当前分层规范，建议新增如下文件。

repositories：

- `internal/repositories/wxwork_kf_sync_state_repository.go`
- `internal/repositories/wxwork_kf_conversation_repository.go`
- `internal/repositories/wxwork_kf_message_ref_repository.go`
- `internal/repositories/channel_message_outbox_repository.go`

services：

- `internal/services/wxwork_kf_inbound_service.go`
- `internal/services/wxwork_kf_conversation_service.go`
- `internal/services/wxwork_kf_message_service.go`
- `internal/services/wxwork_kf_outbound_service.go`
- `internal/services/channel_message_outbox_service.go`

enums：

- `internal/pkg/enums/wxwork_kf.go`

说明：

- controller 不新增复杂逻辑
- 回调入口继续复用 `internal/controllers/third/wechat_controller.go`
- 业务编排全部进入 service 层

## 16. 分阶段实施建议

### 16.1 第一阶段：入站打通

目标：

- 回调可增量拉取
- `text/image/file` 入站可进入平台会话
- 幂等、cursor 持久化完成

交付结果：

- 企业微信客户发来的消息可在后台会话列表和详情页看到

### 16.2 第二阶段：文本下行闭环

目标：

- 平台客服发送文本消息，可异步发送至企业微信客户
- 失败状态可追踪

交付结果：

- 后台人工接待企业微信客户形成闭环

### 16.3 第三阶段：事件同步与可观测性

目标：

- 处理 `enter_session`
- 处理 `session_status_change`
- 处理 `msg_send_fail`
- 完善日志、重试、补偿能力

交付结果：

- 运行可观测性和排障能力增强

### 16.4 第四阶段：媒体下行增强

目标：

- 支持平台图片/文件发送到企业微信

前提：

- 补齐微信素材上传能力

## 17. 第一版建议落地范围

为了尽快上线并降低风险，建议第一版严格控制范围。

建议纳入第一版：

- `wxwork_kf` 外部渠道枚举
- `wxwork_kf_sync_state`
- `wxwork_kf_conversation`
- `wxwork_kf_message_ref`
- 入站 `text/image/file`
- 平台会话复用
- 微信消息幂等
- 下行文本 outbox
- 关键微信事件记录

建议暂缓：

- 企业微信原生坐席协同
- 双端人工接待
- 下行图片/文件
- 欢迎语和高级消息能力
- 客户资料自动补全

## 18. 风险与注意事项

### 18.1 同步 cursor 风险

若 `next_cursor` 未正确持久化，会导致：

- 重复消费
- 消息积压
- 消息错过

因此必须优先保证：

- 每页处理成功后再更新 cursor

### 18.2 幂等风险

如果不基于 `wx_msg_id` 做幂等，会导致：

- 重复写消息
- 会话未读数错误
- AI 被重复触发

### 18.3 下行窗口限制

企业微信客服下行受官方限制：

- 用户主动发消息后的 48 小时内
- 最多下发 5 条

因此必须保留失败原因并支持运维定位。

### 18.4 渠道媒体能力不对称

第一期入站支持 `image/file` 不代表下行也能立即支持。

因为：

- 入站消息可保存 `media_id`
- 下行消息需要先上传素材再发送

这部分能力应明确拆分阶段，避免第一期范围失控。

## 19. 结论

企业微信客服的第一期接入，应采用“渠道接入层 + 现有客服体系复用”的方案，而不是单独构建一套微信客服域。

推荐结论如下：

- 平台主导会话生命周期
- 企业微信只作为客户入口
- 统一复用现有会话与消息模型
- 首期优先打通入站和文本下行闭环
- 通过渠道映射表、消息映射表、outbox 和 sync_state 保证稳定性

该方案具备以下优点：

- 复用现有系统能力，开发成本可控
- 分层清晰，符合当前项目规范
- 先解决“客户能接进来、客服能回出去”的核心问题
- 为后续媒体下行、双端协同、高级事件处理预留了扩展空间
