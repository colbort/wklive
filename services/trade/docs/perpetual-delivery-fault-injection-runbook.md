# 永续与交割合约故障注入验收手册

## 1. 用途

本手册对应《永续与交割合约完成清单》P0-07。只允许在隔离测试环境执行。

每次演练必须记录：

- 环境、版本号和执行时间；
- `tenant_id`、`order_no`、`fill_no`、`batch_no`、`instruction_no`、`event_no`；
- 注入前、故障期间、恢复后的 SQL 结果；
- Trade、Asset、Itick、System 日志和 trace；
- 最终对账异常是否自动恢复。

未取得本手册要求的证据，不得将 P0-07 标记为完成。

## 2. 通用验收查询

以下变量均表示本次演练选定的稳定业务号，不得使用模糊条件批量修改数据。

```sql
SELECT order_no, order_status, filled_qty, canceled_qty, settlement_status
FROM t_trade_order
WHERE tenant_id = ? AND order_no = ?;

SELECT fill_no, order_no, qty, amount, fee, settlement_status, settled_at
FROM t_trade_fill
WHERE tenant_id = ? AND fill_no = ?;

SELECT id, qty, avail_qty, frozen_qty, position_margin, isolated_margin,
       status, version, update_times
FROM t_contract_position
WHERE tenant_id = ? AND id = ?;

SELECT action_key, before_version, after_version, before_qty, after_qty,
       before_position_margin, after_position_margin
FROM t_contract_position_history
WHERE tenant_id = ? AND position_id = ?
ORDER BY id;

SELECT instruction_no, action, amount, step_no, status, retry_count,
       next_retry_at, asset_flow_no, reconciled_at, last_error_msg
FROM t_trade_settlement_instruction
WHERE tenant_id = ? AND instruction_no = ?;

SELECT event_no, event_type, event_status, retry_count, claimed_by,
       claimed_at, delivered_at, last_error_msg
FROM t_biz_trade_event
WHERE tenant_id = ? AND event_no = ?;

SELECT consumer, event_no, event_type, status, retry_count, last_error_msg
FROM t_trade_event_inbox
WHERE tenant_id = ? AND event_no = ?;

SELECT issue_key, check_type, status, occurrence_count, expected_value,
       actual_value, detail, first_seen_at, last_seen_at, resolved_at
FROM t_contract_reconciliation_issue
WHERE tenant_id = ? AND biz_no = ?;
```

验收结束后必须满足：

```text
filled_qty + canceled_qty <= order.qty
position.avail_qty + position.frozen_qty <= position.qty
同一 action_key 只有一条 Position History
同一 instruction_no 只有一条 Settlement Instruction
同一 event_no/consumer 只有一条 Inbox
每个成功 Settlement Instruction 关联唯一 Asset Flow
不存在无期限停留的 PROCESSING 状态
```

## 3. 故障矩阵

### FI-01 Redis 超时

注入：

1. 仅阻断 Trade/Itick 到 Redis 的连接，不阻断 MySQL；
2. 创建一笔隔离保证金合约委托并产生 Fill；
3. 保持故障超过一次任务扫描周期后恢复 Redis。

预期：

- MySQL 中订单、Fill、Outbox/Instruction 事实不丢失；
- Redis 恢复后任务可以重新投影缓存；
- 不重复生成 Fill、Position History 或资金指令；
- Redis 不可用不能绕过 MySQL 状态机直接完成订单。

### FI-02 Kafka 发布超时与重复投递

注入：

1. 阻断 Trade Publisher 到 Kafka；
2. 产生 `POSITION_FILL_REQUIRED` 或结算完成事件；
3. 恢复 Kafka；
4. 使用相同 `event_no` 重放消息两次。

预期：

- Outbox 从 `PENDING/FAILED` 自动恢复为成功；
- 重复消息命中 `(consumer, tenant_id, event_no)` Inbox 唯一边界；
- Position History、Settlement Instruction 和 Asset Flow 均不增加第二份；
- 旧 claimant 不能确认新 claimant 的租约。

### FI-03 Asset RPC 超时

注入：

1. 在 Asset 完成业务前让 RPC 返回超时；
2. 对手续费、保证金、盈亏及交割步骤各执行一次；
3. 恢复 Asset。

预期：

- Settlement Instruction 进入 `FAILED` 并设置 `next_retry_at`；
- 重试次数递增且退避最大为 1024 秒；
- 恢复后自动成功并写入 `asset_flow_no/reconciled_at`；
- 达到 20 次后进入可定位的 `MANUAL_REVIEW`，不得永久 PROCESSING。

### FI-04 Asset 成功但 Trade 未确认

注入：

1. 让 Asset 提交资金事务；
2. 在响应到达 Trade 前中断连接或停止 Trade；
3. 重启 Trade 并等待恢复任务。

预期：

- Trade 使用同一 `instruction_no` 重试；
- Asset 幂等表返回既有结果，不发生第二次资金变化；
- Trade 最终把同一条指令标记成功；
- 对账任务关联唯一 Asset Flow。

### FI-05 MySQL 死锁

注入：

1. 两个会话以相反顺序锁定本次测试专用的 Order/Position；
2. 触发仓位投影或结算事务；
3. 释放人为锁。

预期：

- 仅 MySQL 1213/1205 在当前调用内重试，最多执行 3 次，退避为 10ms、20ms；
- 普通业务校验错误不重试；上下文取消后不再发起下一次事务；
- 被回滚事务不留下半条 Position History 或半条指令；
- 三次仍失败时返回错误，由 Inbox/Outbox 或定时恢复路径重新执行；
- 稳定幂等键阻止重复投影；
- 最终不存在订单、仓位和资金终态不一致。

证据：

- 保存两会话锁定与解锁的 SQL、Trade 错误日志及每个业务键的尝试次数；
- 对同一 `order_no`、`fill_id`、`action_key`、`instruction_no`、`event_no` 查询唯一性；
- 对比注入前后 Position、Position History、Reservation、Settlement Instruction、Trade Event 和 Asset Flow，确认没有半事务或重复事实。

### FI-06 Trade 事务提交后立即退出

注入：

1. 在撮合事务提交、发布 Outbox 之前停止 Trade；
2. 重启 Trade。

预期：

- 已提交 Fill 和 Outbox 同事务存在；
- `ProcessTradeEvents` 重新领取 Outbox；
- Position 和资金步骤最终完成；
- 不依赖原进程内存恢复。

### FI-07 Position 乐观锁冲突

注入：

1. 对同一隔离仓位并发产生两个合法 Fill；
2. 人工控制两个 Worker 读取相同版本；
3. 放行更新。

预期：

- 只有一个版本更新成功；
- 失败事件重新读取新版本并重试；
- History 的 `before_version/after_version` 连续；
- 数量、保证金和已实现盈亏没有丢失更新。

### FI-08 Worker 租约过期及多实例领取

注入：

1. Worker A 领取 Instruction 后暂停超过 60 秒；
2. Worker B 执行恢复扫描并重新领取；
3. 恢复 Worker A。

预期：

- Worker A 的旧 `update_times/claim token` 无法提交成功；
- Worker B 成为唯一有效提交者；
- Asset 即使收到重复业务号也只产生一次资金变化；
- 最终 Instruction 不停留在 PROCESSING。

### FI-09 服务重启恢复

分别在以下状态停止并重启 Trade：

- Reservation `FREEZING/RELEASING`；
- Fill `PROCESSING/FAILED`；
- Funding/Delivery Batch 处理中；
- ADL Execution `PREPARED/ASSET_DONE`；
- Liquidation 父 Saga 未完成。

预期：

- 定时任务能够发现并推进每种持久中间态；
- 子 Saga 完成后父 Saga继续完成；
- Batch 只有在 Asset 对账通过后才能完成或归档。

### FI-10 Price Engine 输入暂时缺失

注入：

1. 停止一个公式必需行情源超过 `max_lookback_ms`；
2. 保持 Price Engine 运行；
3. 恢复行情源。

预期：

- 返回带公式和组件维度的 `ErrInputUnavailable`；
- 缺失期间不生成 MARK、INDEX、FUNDING 或 DELIVERY 快照；
- 日志最多每 30 秒提示一次，不每秒刷 error；
- 输入恢复后自动恢复计算，且产生恢复日志；
- Trade 不使用无审计的最新价兜底。

### FI-11 Snapshot Outbox 大量积压

注入：

1. 暂停 Snapshot Kafka/Redis 下游；
2. 累积至少一个峰值分钟的 Outbox；
3. 恢复下游，分别使用 32 和 64 Worker 演练。

观察：

```sql
SELECT status, COUNT(*) AS row_count, MIN(create_times) AS oldest_at
FROM t_itick_snapshot_outbox
WHERE status IN (1,2,4,5)
GROUP BY status;
```

预期：

- `pending` 和最老消息年龄持续下降；`processing` 是瞬时领取量，不作为吞吐指标；
- Redis/Kafka 分阶段 checkpoint 能在重启后续跑；
- 重复发布使用稳定 `snapshot_id/event_id`；
- 积压回到实时水位后健康告警停止；
- 健康查询使用 `(status, create_times)` 索引，生产 `EXPLAIN` 不得全表扫描；
- 成功 Outbox 按保留期分批清理。

## 4. 通过判定

每个场景必须同时满足：

1. 自动化测试或测试环境实测通过；
2. 资金、仓位、订单、事件八类事实查询一致；
3. 自动对账没有 OPEN 差异，或故障期间产生的差异已经自动 RESOLVED；
4. 没有依靠直接修改终态字段完成恢复；
5. 演练证据归档并关联发布版本。

## 5. 2026-07-28 隔离环境执行记录

环境：仓库 `deploy/docker-compose.yml` 加
`deploy/docker-compose.acceptance.yml`，使用独立 Docker Volume。

已取得证据：

- 全新数据库初始化成功，记录 36 个 migration；
- `trade.ProcessOrderMatching`、`trade.ProcessPositions`、
  `trade.ProcessContractSettlements`、`trade.ProcessTradeEvents` 持续写入成功 Job Log；
- 真实 MySQL 双事务产生 1213，失败事务自动回滚并重试，最终两行测试值均为 2；
- 真实 Redis 验证非持有者无法续租、无法删除他人租约；
- 隔离 Redis 停机时 Trade 无法取得任务锁并返回失败，未绕过锁执行；恢复后四类 Job
  无需重启自动恢复连续成功；
- Trade 停机期间创建的 `ACCEPT-RESTART-20260728-1` 保持 Pending，启动后 Outbox 和 Inbox 均成功；
- 同一事件向 Kafka 重放两次后，Outbox 和 Inbox 仍各只有一条；
- Kafka 停机时 `ACCEPT-KAFKA-TIMEOUT-20260728-1` 进入 Failed 并递增重试，恢复后成功；
- 首次 Kafka 注入发现 Event Subscriber 退出后不重连；修复 Event/Task Subscriber 持续重连后，
  `ACCEPT-KAFKA-RECONNECT-20260728-2` 在不重启 Trade 的情况下自动恢复，最终 Outbox/Inbox 成功。

仍需补充后才能把对应大项标为完成：

- 使用真实 `POSITION_FILL_REQUIRED` 重放，核对 Position History、Instruction 和 Asset Flow 不增加第二份；
- 在 Position/Instruction 事务中制造 1213，而不只是测试专用行；
- 在任务已持锁且正在执行时停止 Redis，保存续租失败取消任务上下文的运行时证据；
- 在 Reservation、Funding、Delivery、ADL 父子 Saga 各中间态执行进程退出；
- 保存完整 trace、SQL 结果和发布版本号到正式验收归档。
