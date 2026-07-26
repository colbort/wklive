# Option 权威行情 Kafka 消费

Option 通过 consumer group `option-market-quote-v1` 消费
`market.authoritative-snapshot.v1`，不再接受 Itick 的 `SyncMarketQuote` RPC。

## 一致性与幂等

- Itick 使用 Outbox 保证行情事件最终发布。
- Kafka 提供至少一次投递，重复事件是正常情况。
- Option 以 `(snapshot_id, contract_id)` 为唯一键逐合约 Claim。
- Claim、当前行情更新和行情快照写入处于同一个数据库事务。
- 只有 MySQL 1062 唯一键冲突被视为重复消费，其他数据库错误必须重试。

## Inbox 清理

`MarketSnapshotInboxCleanup` 控制清理任务：

- `RetentionHours` 默认 720 小时（30 天）。
- `IntervalMinutes` 默认 60 分钟。
- `BatchSize` 默认 5000，最大 10000。
- `MaxBatchesPerRun` 默认 10。

保留时间必须大于 Kafka retention、DLQ 人工重放窗口与安全余量之和。

## 作用范围

当前 `t_option_contract` 只保存 `underlying_symbol`，没有 `category_code`
和 `market` 字段。因此消费者会校验事件的 category/market，但合约匹配仍以
全局唯一的 underlying symbol 为准，并更新所有租户的对应合约。

在允许不同市场出现同名但含义不同的 symbol 前，必须先把 category/market
加入 Option 合约模型和唯一约束，再用于消费查询过滤。

## 观测

消费者每 30 秒记录累计指标：

- `success`
- `failed`
- `updated`
- `duplicates`

Kafka 组件在达到最大重试次数后写入 DLQ 并提交原消息 offset。
