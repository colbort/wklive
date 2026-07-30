# 永续与交割合约预生产灾备演练报告

## 1. 演练范围

- 环境：本机完整 Docker Compose 预生产验收环境
- 日期：2026-07-29（Asia/Hong_Kong）
- 源数据库：`wklive`
- 恢复目标：脚本生成且预检不存在的 `wklive_dr_verify_*` 临时数据库
- 恢复目标处置：逐表核验完成后已删除
- 停写范围：`market-rpc`、`trade-rpc`、`asset-rpc`、`system-rpc`
- 操作原则：不覆盖源库，不删除持久卷，自动强平和全仓开关保持关闭

## 2. 备份与恢复证据

| 项目 | 结果 |
| --- | --- |
| 源快照计数完成时间 | 2026-07-29 17:17:28 +08:00 |
| 逻辑备份完成时间 | 2026-07-29 17:17:38 +08:00 |
| 恢复及恢复库精确计数完成时间 | 2026-07-29 17:19:49 +08:00 |
| 备份大小 | 1,388,380,350 bytes |
| 备份 SHA-256 | `8bab9a5ce3e95ebae7b40a8ee1040f849f626fd51913a4ce4f465f2e81319780` |
| 基础表数量 | 143 |
| 逐表精确行数差异 | 0 |
| `count-diff.txt` | 空文件，表示源快照与恢复库无差异 |
| 恢复及全表计数耗时 | 131 秒（从备份完成至恢复计数完成） |

备份使用 `mysqldump --single-transaction --routines --events --triggers
--hex-blob --set-gtid-purged=OFF`。恢复前确认目标数据库不存在，恢复操作仅写入新建的
临时数据库。

## 3. 核心事实表

| 表 | 源快照行数 | 恢复库行数 | 结果 |
| --- | ---: | ---: | --- |
| `schema_migrations` | 48 | 48 | PASS |
| `t_trade_order` | 20,153 | 20,153 | PASS |
| `t_trade_fill` | 28 | 28 | PASS |
| `t_contract_position` | 0 | 0 | PASS |
| `t_trade_settlement_instruction` | 20,219 | 20,219 | PASS |
| `t_biz_trade_event` | 40,379 | 40,379 | PASS |
| `t_trade_event_inbox` | 40,379 | 40,379 | PASS |
| `t_contract_funding_batch` | 1 | 1 | PASS |
| `t_contract_delivery_batch` | 1 | 1 | PASS |
| `t_contract_reconciliation_issue` | 4 | 4 | PASS |
| `t_itick_authoritative_snapshot` | 1,724,478 | 1,724,478 | PASS |
| `t_itick_snapshot_outbox` | 26,255 | 26,255 | PASS |

全部 143 张基础表均使用精确 `COUNT(*)` 比较，不使用
`information_schema.TABLES.table_rows` 的估算值。

## 4. 服务恢复与运行水位

服务按 Asset、iTick、Trade、System 顺序恢复。演练后只读复核：

- MySQL：healthy
- Etcd：healthy
- Asset RPC：healthy
- iTick RPC：healthy
- Trade RPC：healthy
- System RPC：healthy
- 未完成 Settlement Instruction：0
- OPEN 合约对账差异：0
- 不健康 Snapshot Outbox：0
- Pending Snapshot Outbox：0
- Processing Snapshot Outbox：16，均为 60 秒内实时在途
- `AutomaticLiquidation.Enabled=false`
- `CrossMarginTrading.Enabled=false`

## 5. RPO/RTO 与边界

- 本轮为停写后的事务一致性逻辑备份，源快照到恢复库的事实行数差异为 0；
- 本轮恢复及全表计数耗时 131 秒，但未把服务启动等待单独计时，因此不将其冒充已审批
  的正式 RTO；
- 单机环境未具备异地对象存储、MySQL Binlog 时间点恢复和可用区切换条件；
- 异地备份、Binlog/PITR、可用区切换、回切及正式 RPO/RTO 必须在生产基础设施复核。

## 6. 结论

预生产环境的静默备份、隔离恢复、143 张表精确比对、临时库清理、服务顺序恢复和
运行水位复核均通过。该结论证明单机预生产逻辑备份与恢复路径可执行，不代表生产异地
灾备门禁已经获批。

- 执行人：Codex（按用户授权执行）
- 复核人：待项目负责人签署
- 审批编号：预生产技术验收，不适用
