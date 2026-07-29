# 永续与交割合约运行告警投递验收报告

## 1. 验收范围

- 环境：本机完整 Docker Compose 预生产验收环境
- 日期：2026-07-29（Asia/Hong_Kong）
- 应用告警平台：Kafka `admin.notifications`
- 后台接收：Admin API 当前实例消费组 `admin-notifications-419117f27706`
- 后台展示通道：`/admin/ws/notifications`
- Admin API 镜像：
  `sha256:909eb263393579250611e2099b9bf7ddab82df674442b6422fa86c81c6338b5b`
- iTick 镜像：
  `sha256:f2cb7632dcc0373c8c8f73aaab8b65cefc36a9a8db3197abe2ddaed137bae131`
- Trade 镜像：
  `sha256:8ecbd11b517d2591a4d2207a0c66904a83c3c3fd5bab9687c46de74bd3132530`

本报告验证应用到 Kafka、Admin 消费组和后台 WebSocket 广播入口的技术链路。
生产值班组、短信/电话/IM 通知渠道、确认回执及未确认升级链路尚未提供，因此本报告
不能替代生产外部通知验收。

当前按项目决策只验收后台 WebSocket；Admin API 的外部告警 Webhook 配置保持
`Enabled: false`，不将未配置的外部接收方计为通过。

## 2. 告警规则

| 规则 ID | 触发条件 | 恢复条件 | 去重与重试 |
| --- | --- | --- | --- |
| `snapshot_outbox` | Failed/Manual 非零，或开放记录超过 60 秒 | Failed/Manual 清零且开放记录回到新鲜窗口 | 首次、故障类别变化和 30 分钟提醒；发送失败 30 秒重试 |
| `price_engine_input` | 活跃公式缺少权威输入 | 所有活跃公式恢复权威输入 | 公式/Authority/Kind/Category/Market/Symbol/原因组成稳定指纹；首次、内容变化和 30 分钟提醒；发送失败 30 秒重试 |
| `contract_reconciliation` | 对账扫描产生新、变化、重开或 30 分钟未恢复差异 | 由对账事实闭环；OPEN 差异继续保留 | 数据库行锁预约；Kafka 失败用 compare-and-set 释放预约，下一轮重试 |

恢复事件统一使用 `state=resolved` 和 `level=info`。所有事件均包含稳定
`alertKey`、业务来源、时间和结构化 details。

## 3. 真实事件与 Kafka 记录

| 类型 | 状态 | 应用时间 | Kafka CreateTime | Offset | 写入延迟 |
| --- | --- | --- | --- | ---: | ---: |
| Snapshot Outbox | firing | 2026-07-29 20:44:53.651 +08:00 | 20:44:53.653 | 2 | 2 ms |
| Snapshot Outbox | resolved | 2026-07-29 20:45:23.649 +08:00 | 20:45:23.650 | 3 | 1 ms |
| 合约对账通道验收 | firing | 2026-07-29 20:47:52.573 +08:00 | 20:47:52.573 | 4 | 0 ms |
| Price Engine v3 | firing | 2026-07-29 20:52:29.711 +08:00 | 20:52:29.711 | 19 | 0 ms |
| Price Engine v3 | resolved | 2026-07-29 20:52:46.706 +08:00 | 20:52:46.706 | 20 | 0 ms |

平台事件 ID：

- `snapshot_outbox:snapshot-outbox:firing:1785329093651`
- `snapshot_outbox:snapshot-outbox:resolved:1785329123649`
- `contract_reconciliation:acceptance:contract-reconciliation:20260729:firing:1785329272573`
- `price_engine_input:price-engine-input:firing:1785329549711`
- `price_engine_input:price-engine-input:resolved:1785329566706`

首次三类事件验收后的 Admin 消费组状态为 `CURRENT-OFFSET=21`、
`LOG-END-OFFSET=21`、`LAG=0`。后续 WebSocket 运行验收写入 offset 22，当前实例
消费组最终为 `CURRENT-OFFSET=23`、`LOG-END-OFFSET=23`、`LAG=0`。
这证明以上事件已被 Admin 消费；未连接真实值班客户端，因此不声明外部人员已收到
或确认。

## 4. 故障与恢复结果

### 4.1 Snapshot Outbox

1. 停止 Kafka，不删除容器或 Volume；
2. Outbox 真实出现 `failed=79`，30 秒后增至 `failed=431`；
3. Kafka 不可用时 firing 投递失败，应用严格间隔 30 秒重试；
4. Kafka 恢复后，下一次健康检查补发 firing：
   `pending=119 processing=64 failed=700 oldestOpenAge=104314`；
5. 30 秒后自动排空并发送 resolved：
   `failed=0 manual=0 oldestOpenAge=0`；
6. Kafka、iTick、Trade 和 Admin API 最终均为 Healthy。

### 4.2 Price Engine

通过正式 Admin API 创建缺少输入的
`ALERTTESTUSDT-INDEX-alert-acceptance-v3`，持续缺源期间只产生一条 firing；
通过正式状态接口撤销后产生一条 resolved。验收公式 ID 12 已撤销，不属于活跃生产
候选公式。

首次 v2 演练发现错误文本中的 `target` 每秒变化会被误判为内容变化，产生 14 条
firing（offset 5～18）。随后改为结构化 `InputUnavailableError`，告警指纹排除
`target`，保留完整 target 供日志和事件排障；v3 复测以 offset 19～20 证明修复有效。

### 4.3 合约对账

通知通道使用一次性、明确标记 `acceptanceTest=true` 的事件完成 Kafka/Admin 实际
投递，未篡改资金、订单、仓位或对账事实。业务检测路径由自动化测试覆盖：

- 成功发送时不释放数据库告警预约；
- Kafka 发送失败时按 tenant、issue key 和 reserved time 条件释放预约；
- 发送失败与释放失败同时保留错误；
- 旧任务不能清除更新后的预约。

此项证明真实通知通道和业务代码契约；生产仍应在获批演练窗口制造可恢复的真实对账
差异，补充值班通知、确认与升级回执。

### 4.4 Admin WebSocket

重新部署 Admin API 后，以初始化的真实系统管理员登录并进行浏览器等价握手：

- 旧式 `?token=...` 请求返回 `401 Unauthorized`，令牌不再进入 URL；
- 缺少 `Origin` 的请求返回 `403 Forbidden`；
- 使用允许的 `Origin`、固定子协议 `wklive-admin-notifications` 和
  `bearer.<JWT>` 子协议完成握手，返回 `101 Switching Protocols`；
- 服务端只回显固定子协议，不回显 JWT；
- 握手前实时查询用户详情和权限；连接按事件权限、租户及系统管理员身份过滤；
- 连接建立后收到 `connected` 事件，`userId=1`；
- 向 Kafka 写入明确标记 `acceptanceTest=true`、无业务数据副作用的
  `snapshot_outbox` 验收事件
  `snapshot_outbox:acceptance:admin-ws:1785331402000`；
- 同一已鉴权连接实际收到完整事件，Kafka offset 22；
- 当前实例消费组最终为 `23/23`、`LAG=0`。

Admin WebSocket 是在线后台的尽力实时通道，不提供离线值班确认、消息 ACK 或升级
回执；这些能力仍属于未完成的生产外部通知门禁。

### 4.5 通知接口解耦

告警生产逻辑已从“频道 + 任意消息”的发布器迁移为传输无关的
`common/alert.Notifier`：

- `alert.Notify` 统一校验领域告警后调用接口；
- `alert.NotifierFunc` 支持使用函数快速实现通道；
- `alert.MultiNotifier` 可组合多个实现，逐个尝试并汇总失败；
- 稳定的 `Alert.ID` 是各实现重试时的幂等键；
- 当前 Kafka/Admin 转换独立位于 `common/alert/adminnotify`；
- iTick 和 Trade 的 ServiceContext 只注入 `alert.Notifier`，任务代码不再依赖
  Kafka Channel、Admin Event 或 WebSocket；
- 后续邮件、短信或 IM 通知只需新增 `Notifier` 实现，不修改 Price Engine、
  Snapshot Outbox 或合约对账逻辑。

重新构建部署后，iTick 镜像 `f2cb7632…e131`、Trade 镜像
`8ecbd11b…2530` 均为 Healthy，近五分钟未发现
`error/panic/fatal/failed/unhealthy`。

## 5. 自动化验证

- `common/alert`：领域校验、接口调用、组合器全通道尝试及多错误汇总；
- `common/alert/adminnotify`：Admin Event 转换、恢复级别和发布失败；
- `common/alert.DeliveryTracker`：首次、失败重试、内容变化、30 分钟提醒、恢复重试；
- `services/market/internal/priceengine`：结构化缺源错误与既有计算回归；
- `services/market/internal/tasks`：Price/Outbox 事件及稳定指纹；
- `services/trade/internal/logic/task`：对账告警投递与预约释放；
- `admin-api/internal/ws`：事件权限、租户隔离、系统级告警和 Origin 策略；
- `admin-api/internal/handler/ws`：固定子协议令牌解析、查询参数令牌拒绝；
- iTick、Trade 全量 `go test ./...` 通过；
- iTick、Trade 告警相关包 `go test -race` 通过；
- Admin API `go test ./...` 全量通过；
- Admin UI `vue-tsc --noEmit` 和通知相关文件 Prettier 检查通过；
- `common`、iTick、Trade 全量 `go test ./...` 通过；
- `common/alert/...`、iTick `internal/tasks`、Trade
  `internal/logic/task` 的 `go test -race` 通过；
- 新接口及两个服务相关包 `go vet` 通过。

## 6. 结论

三类运行告警已接入同一 Kafka/Admin 通知链路，具备结构化事件、恢复事件、去重、
30 分钟提醒和同步发送失败重试。Snapshot Outbox 与 Price Engine 的真实故障/恢复
验收通过，合约对账的实际通道和检测代码契约通过；后台 WebSocket 的真实管理员
鉴权、固定子协议、Origin 门禁和端到端事件接收通过。

生产 P0-06 仍保持部分完成：缺少真实值班组、一级外部渠道、通知确认、未确认升级
策略和最终责任人，不能据此打开自动强平或全仓生产开关。

- 执行人：Codex（按用户授权执行）
- 值班负责人：待生产值班组签署
- 审批编号：预生产技术验收，不适用
