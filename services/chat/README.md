# Chat RPC Service

客服系统 RPC 服务，提供用户端客服会话、后台坐席处理、内部系统消息三个入口。

## 整体架构

客服系统按职责可以拆成 4 个端:

- 访客端: 网站浮窗、H5、App 内客服入口，负责发起咨询、发送消息、查看历史、提交评价。
- 客服工作台: 客服上线后接待用户、回复消息、转接会话、结束会话。
- 管理后台: 管理客服、分组、权限、快捷回复、会话记录、统计报表。
- 服务端: RPC 核心业务、WebSocket 实时连接、MySQL 状态数据、MongoDB 消息数据、Redis 在线状态和等待队列。

当前 `services/chat` 只承载客服核心 RPC 能力；WebSocket 网关、HTTP API、管理后台页面可以通过调用本服务完成业务落库和状态流转。

推荐项目边界:

- `chat-admin`: 客服后台前端项目，根据登录账号身份展示客服商户后台或客服坐席工作台。
- `chat-api`: `chat-admin` 的接口层/BFF，负责登录态解析、身份聚合、权限判断、调用 `services/system` 和 `services/chat`。
- `services/system`: 只管理客服商户主档 `sys_chat_merchant`。
- `services/chat`: 提供客服 RPC 业务能力，管理 `t_chat_user`、坐席、会话、工单等客服业务数据。

## 服务划分

- `ChatApp`: 用户端接口，负责创建/查询会话、发送用户消息、拉取消息、标记已读、关闭会话。
- `ChatAdmin`: 后台接口，负责坐席管理、会话分页、会话分配/转接、客服回复、后台已读、关闭会话。
- `ChatInternal`: 内部接口，供其他业务服务创建系统会话、发送系统消息、查询用户未关闭会话。

对应 proto 文件在 `../../proto/chat`:

- `chat_app.proto`
- `chat_admin.proto`
- `chat_internal.proto`
- `model.proto`
- `enum.proto`

## 目录说明

- `chat.go`: RPC 服务启动入口，从 etcd 加载公共配置和 chat 配置。
- `chat.sql`: 客服系统表结构。
- `internal/server`: go-zero 生成的 gRPC server 适配层。
- `internal/logic`: 业务逻辑实现入口。
- `internal/svc`: 服务依赖上下文。
- `chatapp`, `chatadmin`, `chatinternal`: go-zero 生成的 RPC client 包。

## 存储设计

当前采用 **MySQL + MongoDB** 混合存储:

- MySQL: 存坐席、会话、分配记录、未读数、最后一条消息编号和摘要等需要事务和筛选的状态数据。
- MongoDB: 存聊天消息正文、附件信息、卡片 payload 等高频写入和结构较灵活的数据。
- Redis: 存在线状态、WebSocket 连接关系、等待队列和短期会话映射。

`chat.sql` 当前只包含 MySQL 状态表:

- `t_chat_user`: 客服商户账号和客服坐席登录账号，保存账号、密码、昵称、头像、启禁用等身份资料。
- `t_chat_agent`: 客服坐席服务资料，保存 `chat_user_id`、分组、坐席编号、在线状态、最大接待数、当前接待数等接待状态。
- `t_chat_session`: 客服会话。
- `t_chat_assignment`: 会话分配/转接记录。
- `t_chat_quick_reply`: 快捷回复。
- `t_chat_category`: 问题分类配置。
- `t_chat_satisfaction`: 会话满意度评价。
- `t_chat_read_cursor`: 多端独立已读游标。
- `t_chat_group`: 客服分组和按组分配。
- `t_chat_work_order`: 工单和离线留言。

MongoDB 消息集合建议命名为 `chat_message`，字段与 `proto/chat/model.proto` 的 `ChatMessage` 保持一致:

```json
{
  "message_no": "M202606200001",
  "session_no": "S202606200001",
  "merchant_id": 1,
  "user_id": 10001,
  "agent_id": 20001,
  "sender_type": 1,
  "sender_id": 10001,
  "sender_name": "用户昵称",
  "message_type": 1,
  "content": "你好，我想咨询订单",
  "media_url": "",
  "media_name": "",
  "media_mime": "",
  "media_size": 0,
  "status": 1,
  "payload": {},
  "read_time": 0,
  "create_times": 1781928000000,
  "update_times": 1781928000000
}
```

推荐索引:

- `message_no` 唯一索引，用于幂等和精确查询。
- `session_no + create_times` 普通索引，用于拉取会话历史消息。
- `merchant_id + user_id + create_times` 普通索引，用于用户侧消息查询。
- `merchant_id + agent_id + create_times` 普通索引，用于客服侧消息查询。

后续可按业务阶段继续增加:

- `t_chat_bot_reply_rule`: 关键词自动回复规则。
- `t_chat_blacklist`: 恶意访客/用户黑名单。

## 生成命令

在 `services/chat` 目录执行:

```bash
make gen
```

该命令会根据 `proto/chat/*.proto` 生成 RPC 代码，并执行 import 修复、`gofmt` 和 `go mod tidy`。

生成 MySQL model:

```bash
make gen-model
```

生成 MongoDB 消息 model 时，可以单独使用 goctl 的 Mongo model 能力，或参考 `services/itick/models` 里的 Mongo model 写法手动维护。消息模型建议放在 `services/chat/models` 下，和 MySQL model 分文件管理。

## 启动

```bash
go run chat.go \
  -etcd 127.0.0.1:2379 \
  -common /wklive/common/config \
  -config /wklive/chat-rpc/config
```

服务会从 etcd 合并公共配置和客服服务配置，然后注册 `ChatAdmin`、`ChatApp`、`ChatInternal` 三组 gRPC 服务。

## 实现顺序建议

1. 先用 `chat.sql` 生成 MySQL models，并在 `internal/svc` 注入 MySQL model。
2. 增加 MongoDB 连接和 `chat_message` model，用于消息写入和历史消息查询。
3. 实现 `OpenChatSession`、`SendUserMessage`、`SendAgentMessage`，打通会话和消息主链路。
4. 实现自动分配/手动转接，维护 `t_chat_agent.current_session_count` 和 `t_chat_assignment`。
5. 实现未读数、已读时间、会话最后消息摘要更新。
6. 补充后台分页查询和内部系统消息入口。

## 重点设计

### 核心流程

用户发起咨询:

```text
用户打开客服入口
  -> 创建或识别访客/登录用户
  -> 创建或复用未关闭会话
  -> 自动分配在线坐席
  -> WebSocket 双向聊天
```

客服接待:

```text
新会话进入等待队列
  -> 系统分配给在线且有容量的坐席
  -> 客服工作台收到提醒
  -> 客服回复消息
  -> 用户确认解决
  -> 客服结束会话
  -> 用户评价
```

离线留言:

```text
没有可用坐席
  -> 会话保持等待或引导用户留言
  -> 生成工单/离线留言
  -> 后台客服后续处理
```

### 会话生命周期

会话建议按以下状态流转:

```text
WAITING -> SERVING -> PENDING_USER / PENDING_AGENT -> CLOSED
```

- 用户首次发起会话时，如果存在未关闭会话，优先复用原会话。
- 没有可用坐席时，会话保持 `WAITING`，等待后台手动分配或定时任务自动分配。
- 坐席回复后可将会话置为 `PENDING_USER`，用户回复后可置为 `PENDING_AGENT`。
- 会话关闭后不再允许继续发送普通消息，如需继续咨询应创建新会话或重新打开。

### 坐席分配

分配坐席时需要同时考虑:

- 坐席账号 `t_chat_user.enabled` 是否启用，坐席服务状态 `t_chat_agent.status` 是否在线。
- `current_session_count` 是否小于 `max_session_count`。
- 用户会话优先级，高优先级先分配。
- 手动转接时写入 `t_chat_assignment`，保留转接来源、目标坐席、操作人和原因。

更新坐席接待数和会话归属建议放在同一事务里，避免并发分配导致超量接待。

前期自动分配规则保持简单:

```text
找在线坐席
  -> 排除账号禁用、离线、忙碌坐席
  -> 过滤 current_session_count >= max_session_count 的坐席
  -> 按 current_session_count 升序
  -> 分配给最空闲坐席
```

后期可以升级为按客服分组、用户来源、VIP 等级、语言、商品分类、技能标签分配。

### 消息写入

消息正文写入 MongoDB，会话状态写入 MySQL。发送消息时建议按以下顺序保证最终一致:

1. 校验 MySQL 中的 `t_chat_session` 是否存在、未关闭、归属正确。
2. 生成 `message_no`，以 `message_no` 作为 MongoDB 幂等键写入 `chat_message`。
3. 更新 MySQL `t_chat_session.last_message_no`、`last_message`、`last_sender_type`、`last_message_time`。
4. 更新用户侧或坐席侧未读数。
5. 必要时更新会话状态，例如用户发送消息后置为 `PENDING_AGENT`。
6. MySQL 状态更新成功后再投递 WebSocket/消息队列事件。

消息编号 `message_no` 和会话编号 `session_no` 建议由统一 ID 生成器生成，避免依赖自增 ID 暴露内部规模。

如果 MongoDB 写入成功但 MySQL 状态更新失败，不要立即推送消息；应通过 `message_no` 重试更新会话状态，或将该消息标记为待补偿状态后由任务修复。

用户发消息的主链路:

```text
WebSocket/HTTP 收到用户消息
  -> 校验 session_no 是否有效且属于该用户
  -> 生成 message_no
  -> 保存 MongoDB chat_message
  -> 更新 MySQL t_chat_session 最后一条消息和未读数
  -> 推送给当前坐席
  -> 坐席不在线时累计未读
```

客服发消息的主链路:

```text
WebSocket/HTTP 收到客服消息
  -> 校验坐席是否有权限回复该会话
  -> 生成 message_no
  -> 保存 MongoDB chat_message
  -> 更新 MySQL t_chat_session 最后一条消息和未读数
  -> 推送给用户
  -> 用户不在线时累计未读
```

### 未读和已读

- 用户发送消息时增加 `agent_unread_count`。
- 坐席发送消息时增加 `user_unread_count`。
- 用户已读清空或递减 `user_unread_count`。
- 坐席已读清空或递减 `agent_unread_count`。
- `last_message_id` 可作为已读游标，后续如果要做多端同步，可以单独扩展会话读游标表。

### 幂等和并发

客服系统容易出现重复点击、网络重试和多端同时操作，建议重点保护:

- 发送消息接口支持客户端传入幂等键，避免重复消息。
- 分配会话时校验会话当前坐席和状态，防止重复分配覆盖。
- 关闭会话要允许重复调用，已关闭时直接返回当前会话。
- 关键更新使用事务或乐观锁，尤其是坐席接待数和会话状态。

### 实时推送

RPC 只负责核心业务写入，实时消息推送建议由独立网关或 WebSocket 服务承担:

- 写入消息成功后投递事件，例如 `chat.message.created`。
- 会话分配、关闭、已读也投递事件。
- App 和后台分别订阅自己需要的用户/坐席频道。
- 推送失败不影响 RPC 主事务，可通过消息队列重试。

推荐 WebSocket 事件类型:

- `chat_message`: 新聊天消息。
- `session_assigned`: 会话已分配坐席。
- `typing`: 用户或坐席正在输入。
- `message_read`: 消息已读。
- `session_closed`: 会话已关闭。

Redis 可用于维护在线状态、连接关系和等待队列:

- `chat:agent:online:{agent_id}`: 坐席在线状态，心跳续期。
- `chat:ws:user:{user_id}`: 用户连接信息。
- `chat:ws:agent:{agent_id}`: 坐席连接信息。
- `chat:session:agent:{session_no}`: 会话当前坐席。
- `chat:agent:sessions:{agent_id}`: 坐席当前会话集合。
- `chat:queue:{merchant_id}:{group_id}`: 等待分配的会话队列。

### 权限边界

- `ChatApp` 必须校验 `merchant_id`、`user_id` 和 `session_no` 归属。
- `ChatAdmin` 必须校验坐席所属客服商户，以及是否有接管/转接权限。
- `ChatInternal` 只允许可信内部服务调用，避免被外部直接创建系统消息。
- 后台查询接口要始终带客服商户过滤，避免跨客服商户数据泄露。

### 商户后台配置

客服商户后台的基础能力应该对每个商户保持一致，不在 `services/system` 为每个商户维护独立客服后台配置。

推荐边界:

- `services/system`: 只维护客服商户主档 `sys_chat_merchant`。
- `services/chat`: 维护商户自己的客服用户、坐席资料、分组、快捷回复、会话、工单。
- `chat-admin`: 使用同一套前端功能，根据 `t_chat_user.user_type` 区分客服商户和客服坐席。
- 商户之间只做数据隔离，所有 chat 表必须按 `merchant_id` 过滤。
- 如果后续要做套餐差异，可在 `sys_chat_merchant` 增加套餐/能力字段，或由 `chat-api` 做能力判断，不改变 chat 的核心业务表结构。

客服后台登录账号使用 `t_chat_user`，客服坐席资料使用 `t_chat_agent.chat_user_id` 绑定 `t_chat_user.id`:

```text
sys_chat_merchant.id -> chat.*.merchant_id
t_chat_user.id       -> t_chat_agent.chat_user_id
t_chat_group.id      -> t_chat_agent.group_id / t_chat_session.group_id
```

如果单独建设 `chat-admin` 项目，可以使用同一个登录入口，根据登录账号身份切换后台能力:

```text
系统管理员
  -> merchant_id = 0
  -> 管理所有商户、查看全局客服数据

客服商户
  -> merchant_id > 0
  -> 管理本商户的坐席、分组、快捷回复、工单、会话记录

客服坐席
  -> merchant_id > 0
  -> 只进入客服工作台，接待和处理分配给自己的会话
```

客服坐席由客服商户管理，本质上是同一客服商户下的 `t_chat_user`。只有当该账号需要接待用户时，才在 `t_chat_agent` 中创建坐席资料。

`chat-api` 建议负责聚合以下身份信息返回给 `chat-admin`:

```text
merchant_id
user_id
user_type
role_codes
perms
chat_user_id
is_chat_merchant
is_chat_agent
chat_agent_id
chat_group_id
```

身份聚合来源:

- `services/system`: 客服商户主档。
- `services/chat`: 登录账号、坐席资料、客服分组、商户自己的快捷回复、工单和会话数据。

客服商户管理坐席时，`chat-api` 直接在 `services/chat` 创建或更新 `t_chat_user`；如果该用户需要接待，再创建或更新 `t_chat_agent` 业务资料。

商户独有资料建议放在 `services/chat` 的业务表中，例如:

- 坐席资料: `t_chat_agent`
- 客服分组: `t_chat_group`
- 快捷回复: `t_chat_quick_reply`
- 问题分类: `t_chat_category`
- 工单/离线留言: `t_chat_work_order`

`services/system` 只需要管理客服商户主体 `sys_chat_merchant`；不要管理客服账号、会话、消息、工单、快捷回复等客服业务数据。

`services/system` 创建、更新、删除 `sys_chat_merchant` 时，需要调用 `ChatInternal.SyncChatMerchantUser` 同步 `t_chat_user` 中的商户主账号记录:

```text
create sys_chat_merchant
  -> SyncChatMerchantUser(UPSERT)
  -> upsert t_chat_user(user_type = MERCHANT, is_owner = YES)

update sys_chat_merchant
  -> SyncChatMerchantUser(UPSERT)
  -> update t_chat_user nickname/enabled/contact fields

delete/disable sys_chat_merchant
  -> SyncChatMerchantUser(DELETE)
  -> disable t_chat_user and merchant-scoped chat data entry
```

推荐判断规则:

- 是否客服商户: `t_chat_user.user_type = 1` 或 `t_chat_user.is_owner = 1`。
- 是否客服坐席: `t_chat_user.user_type = 2` 且存在 `t_chat_agent.merchant_id = t_chat_user.merchant_id AND t_chat_agent.chat_user_id = t_chat_user.id`。

后台身份建议分为:

- 超级管理员: 管理所有客服、所有会话和所有统计。
- 客服商户: 管理本商户的用户、坐席、分组、快捷回复、工单和会话记录。
- 客服坐席: 只看自己的会话、回复用户、创建工单。

### 功能优先级

不要一开始把机器人、工单、AI、复杂分配全部做完。建议分 4 个版本推进:

- V1 最小可用: 用户发消息、客服回复、WebSocket 实时通信、消息入库、会话列表、客服上线/离线。
- V2 工作台完善: 快捷回复、图片上传、文件上传、未读消息、会话结束、用户评价。
- V3 管理后台: 客服账号、客服分组、聊天记录查询、统计报表、离线留言。
- V4 高级能力: 工单系统、自动回复机器人、转接客服、多商户隔离、多语言客服、AI 客服。

最关键的落地顺序:

1. WebSocket 一对一聊天跑通。
2. 消息保存到 MongoDB。
3. 客服端会话列表。
4. 访客端聊天窗口。
5. 客服上线/离线。
6. 自动分配客服。
7. 历史消息。
8. 图片和文件。
9. 后台统计。
10. 机器人和工单。

### 后续可扩展

- 增加消息敏感词、图片审核、附件有效期和对象存储签名。
- 增加关键词自动回复，前期只做规则命中，用户点击转人工后再创建人工会话。

## 验证

```bash
go test ./...
```
