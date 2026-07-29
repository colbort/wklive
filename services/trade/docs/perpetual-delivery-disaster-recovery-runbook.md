# 永续与交割合约备份及灾备恢复手册

## 1. 目标与事实边界

本手册用于恢复永续、交割合约的可审计交易事实，并验证恢复后不会重复扣款、重复投影仓位或跳过未完成 Saga。

事实源按以下顺序处理：

1. MySQL：订单、成交、仓位、仓位历史、预占、结算指令、Outbox/Inbox、资金费、交割、强平及对账事实；
2. Asset：钱包、冻结、幂等记录和资金流水，仍由 Asset 服务及其数据库负责恢复；
3. Kafka：用于恢复尚未进入 Inbox 的领域事件，不能代替 MySQL 事实；
4. Redis：只保存缓存和分布式锁，不作为资金或仓位事实源；恢复时允许从 MySQL 重建；
5. iTick/Price Engine：恢复不可变行情快照和公式版本后，Trade 才能继续资金费、交割和风险扫描。

生产演练前必须明确 RPO、RTO、备份保留周期、加密方式、异地存储位置和审批责任人。

## 2. 备份前检查

```sql
SELECT COUNT(*) AS open_instructions
FROM t_trade_settlement_instruction
WHERE status IN (1,2,4,5);

SELECT event_status, COUNT(*)
FROM t_biz_trade_event
GROUP BY event_status;

SELECT status, COUNT(*)
FROM t_itick_snapshot_outbox
GROUP BY status;

SELECT COUNT(*) AS open_reconciliation_issues
FROM t_contract_reconciliation_issue
WHERE status = 1;
```

在线备份使用 `--single-transaction`。若演练要求逐表与当前源库做精确行数对比，应短暂停止 System 调度器，避免备份完成后 `sys_job_log` 等表继续增长；生产热备本身不要求停写，但验证必须使用备份时点的清单或 Binlog 位点，不能拿持续变化的源库做事后等值比较。

## 3. 一致性备份

以下命令仅示意本地 Compose 环境。生产必须把密码放入受管密钥，不得写入脚本或日志。

```bash
docker exec wklive-mysql-1 sh -lc \
  'mysqldump -uroot -p"$MYSQL_ROOT_PASSWORD" \
    --single-transaction \
    --routines --events --triggers \
    --hex-blob --set-gtid-purged=OFF \
    wklive' > /secure-backup/wklive.sql

shasum -a 256 /secure-backup/wklive.sql
```

备份记录至少包含：

- 开始和结束时间；
- 文件大小与 SHA-256；
- MySQL GTID/Binlog 位点；
- `schema_migrations` 数量和最新版本；
- 核心表行数；
- 执行人、审批单和存储位置。

## 4. 隔离恢复验证

恢复目标必须是确认不存在的一次性数据库，禁止覆盖当前 `wklive`：

```sql
SELECT COUNT(*)
FROM information_schema.schemata
WHERE schema_name = 'wklive_dr_verify';
```

结果为 0 后才允许创建和恢复：

```bash
docker exec wklive-mysql-1 sh -lc \
  'mysql -uroot -p"$MYSQL_ROOT_PASSWORD" \
    -e "CREATE DATABASE wklive_dr_verify CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci"'

docker exec -i wklive-mysql-1 sh -lc \
  'mysql -uroot -p"$MYSQL_ROOT_PASSWORD" wklive_dr_verify' \
  < /secure-backup/wklive.sql
```

必须逐表比较源快照清单与恢复库的精确 `COUNT(*)`，并至少单独核对：

```text
schema_migrations
t_trade_order
t_trade_fill
t_contract_position
t_contract_position_history
t_trade_asset_reservation
t_trade_settlement_instruction
t_biz_trade_event
t_trade_event_inbox
t_contract_funding_batch
t_contract_funding_settlement
t_contract_delivery_batch
t_contract_delivery_settlement
t_contract_liquidation
t_contract_account_liquidation
t_contract_reconciliation_issue
t_itick_authoritative_snapshot
t_itick_snapshot_outbox
```

任何缺表、迁移数不一致、核心事实行数不一致或约束创建失败都判定恢复失败。

## 5. 服务恢复顺序

1. 恢复 MySQL、Asset 数据库及不可变行情归档；
2. 校验 `schema_migrations`，禁止应用程序连接到缺迁移的库；
3. 启动 Redis；不要恢复过期任务锁，缓存允许重新生成；
4. 启动 Kafka 并确认领域事件 Topic、消费组 Offset 和保留窗口；
5. 启动 Asset、iTick、Trade；
6. 最后启动 System 调度器；
7. 观察 Trade 的四类核心任务：
   `ProcessOrderMatching`、`ProcessPositions`、`ProcessContractSettlements`、`ProcessTradeEvents`；
8. 等待未完成 Settlement Instruction、Outbox 和 Inbox 收敛；
9. 运行全量对账，OPEN 差异必须为 0 或形成经审批的人工处置记录；
10. 自动强平和全仓新增敞口开关保持关闭，直到价格源、账户、告警和资金权限复核完成。

## 6. 恢复通过标准

- 所有迁移存在且校验和一致；
- Order、Fill、Position、History、Reservation、Instruction、Asset Flow、Outbox/Inbox 可相互关联；
- 未完成 Saga 能继续执行，重放不增加重复资金或仓位事实；
- Funding/Delivery Batch 只能在 Asset 流水对账完成后进入终态；
- Snapshot Outbox 回到实时水位；
- 全量对账无未处理差异；
- Trade、Asset、iTick、System 和基础设施全部健康；
- 达到已审批的 RPO/RTO。

## 7. 2026-07-29 隔离演练记录

- 源库：28.23 MiB、140 张基础表；
- 静默点备份：11,975,203 字节，耗时 0.50 秒；
- SHA-256：`1cdd5acc4bb513b2cde353f15feb73ee894e7b5306db53fae3f8e246c46470d4`；
- 隔离恢复耗时：2.68 秒；
- 逐表精确行数：检查 140 张表，差异 0；
- 核心事实：
  `schema_migrations=46`、`orders=40`、`fills=33`、
  `positions=41`、`instructions=109`、`events=216`，源库与恢复库一致；
- 第一次热备后直接与持续写入的源库比较时，仅 `sys_job_log` 多出 97 条；交易事实全部一致。该证据说明验证清单必须绑定备份时点，随后通过静默点演练取得 140/140 一致结果；
- 一次性数据库及临时备份文件已销毁，System 服务恢复健康。

该记录证明本地备份、恢复和事实核验流程可执行；生产异地备份、Binlog 时间点恢复、节点/可用区故障切换及正式 RPO/RTO 仍需生产环境演练。
