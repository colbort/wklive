# Outbox / Inbox 可靠性检查与整改记录

更新时间：2026-08-08

## 1. 目标与统一约定

跨 RPC 的异步事件统一使用 `outbox` / `inbox` 术语：

- Outbox：与业务数据在同一数据库事务中落库，后台任务按“至少一次”语义投递。
- Inbox：消费者按事件号去重，并记录处理状态，防止重复消息重复执行业务。
- 表名采用 `t_<domain>_<purpose>_outbox` / `t_<domain>_<purpose>_inbox`。
- Go 持久化模型采用 `T<Domain><Purpose>Outbox` / `T<Domain><Purpose>Inbox`。
- 管理端仍可显示“交易事件”“支付事件”等业务名称，不要求暴露技术术语。

本方案不保留“表中没有租约字段时继续运行”的运行时兼容逻辑。部署前必须先执行对应迁移。

## 2. 并发处理基线

每一条可重试异步记录必须满足以下约束：

1. 领取使用单条条件更新（CAS），不能只靠“先查询、后普通更新”。
2. 领取时写入唯一 `claimed_by` 和 `claimed_at`。
3. 成功、失败和阶段检查点都必须校验 `claimed_by`，旧进程不得覆盖新进程结果。
4. 进程异常后，只允许在租约过期后被其他实例接管。
5. 成功或失败释放租约时清空 `claimed_by` / `claimed_at`。
6. 外部 RPC 和消息投递按“至少一次”处理，下游仍必须使用稳定业务号保证幂等。
7. 多张本地表的最终状态变更必须在同一数据库事务内完成。

## 3. 检查结果

| 服务 | 持久化对象 | 整改结果 | 栅栏方式 |
| --- | --- | --- | --- |
| Payment | `t_pay_outbox` | 已补齐多实例安全领取；Outbox 成功与充值单入账状态改为同一事务 | `claimed_by` + `claimed_at` |
| Market | `t_itick_snapshot_outbox` | Redis/Kafka 分阶段检查点、成功和失败均校验领取者；每次领取使用唯一令牌 | `claimed_by` + `claimed_at` |
| Option | `t_option_outbox` | 已补齐领取者校验和过期接管 | `claimed_by` + `claimed_at` |
| Option | `t_option_asset_instruction` | 作为等价 Outbox 整改；采用“领取、成功检查点、业务完成、释放”流程 | `claimed_by` + `claimed_at` |
| Liquidity | `t_liquidity_event_outbox` | 原发布任务未实现，现已接入 Kafka 发布、重试与租约校验；当前业务代码尚无生产方写入该表 | `claimed_by` + `claimed_at` |
| Trade | `t_trade_event_outbox` | 原 `t_biz_trade_event` 已按职责改名；每次投递生成唯一领取令牌 | `claimed_by` + `claimed_at` |
| Trade | `t_trade_event_inbox` | 已具备消费者幂等和处理租约 | `update_times` 租约令牌 |
| Trade | `t_trade_settlement_instruction` | 保留业务指令命名；最终事务使用更新时间戳校验当前处理权 | 时间戳栅栏 |
| Staking | `t_stake_operation` | 保留业务操作命名；步骤检查点、失败和最终事务增加版本校验 | `version` 乐观锁 |

## 4. Trade 命名整改

`t_biz_trade_event` 的实际职责是可靠投递发件箱，旧名称无法直接表达用途，现统一为：

- 表：`t_trade_event_outbox`
- Go 模型：`TTradeEventOutbox`
- ServiceContext：`TradeEventOutboxModel`
- 配对消费表继续使用：`t_trade_event_inbox`

RPC DTO `BizTradeEvent` 暂时保留，因为它描述的是管理端展示的业务事件，不属于数据库持久化命名。

## 5. 数据库迁移

部署顺序必须为“先迁移、后滚动服务”：

1. Payment：`payment/migrations/20260808_payment_outbox_claim_lease.sql`
2. Market：`market/migrations/20260808_snapshot_outbox_claim_lease.sql`
3. Option：`option/migrations/20260808_option_outbox_claim_lease.sql`
4. Option：`option/migrations/20260808_option_asset_instruction_claim_lease.sql`
5. Liquidity：`liquidity/migrations/20260808_liquidity_outbox_claim_lease.sql`
6. Trade：`trade/migrations/20260808_rename_trade_event_outbox.sql`

Trade 服务上线前必须确认代码与表重命名在同一次发布窗口完成，避免新代码访问旧表名。

## 6. 语义边界

Outbox / Inbox 解决的是本地事务、可靠重试、并发领取和消费幂等，不提供跨服务“恰好一次”事务。Asset 等外部 RPC 仍可能在“远端成功、本地提交前崩溃”时被再次调用，因此请求必须持续使用稳定的 `biz_no` / `instruction_no` 作为幂等键。

当前实现不需要引入 DTM 才能保证这些异步链路的可靠性。若将来出现必须同步原子提交的跨服务短事务，应对具体业务单独评估，不能用 DTM 替代 Outbox / Inbox。

## 7. 验收项

- 同一条记录被多个实例同时扫描时，仅一个实例领取成功。
- 旧实例租约失效后，其成功/失败更新返回未命中，不能覆盖新实例。
- 租约过期后新实例可以恢复处理。
- 外部调用成功、本地最终事务失败时，记录可重试且不会重复入账。
- 手工重试会清空旧租约。
- 各服务模型、任务单元测试和全量 `go test ./...` 通过。
- `git diff --check` 通过，SQL 主结构与生成模型字段一致。
