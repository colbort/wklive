# iTick Snapshot Outbox 容量治理方案

## 1. 背景

`t_itick_snapshot_outbox` 用于可靠发布权威行情快照。权威快照与 Outbox 记录在同一个 MySQL 事务内写入，后台 Worker 再将消息发布到 Redis 和 Option 服务。

当前状态定义如下：

| 状态 | 含义 | 处理策略 |
| --- | --- | --- |
| `1` | pending | 保留并等待处理 |
| `2` | processing | 保留；超时后允许重新认领 |
| `3` | success | 短暂保留后自动删除 |
| `4` | failed | 保留并自动重试 |
| `5` | manual | 保留并等待人工处理，后续可迁移到 dead-letter 表 |

该表是可靠投递队列，不是权威行情的永久档案。永久档案由 `t_itick_authoritative_snapshot` 保存。因此，Outbox 需要实时写入以保证事务一致性，但成功记录不需要长期留存。

## 2. 容量估算

假设有 100 个产品，每个产品每秒产生一条 Outbox：

```text
每秒：100 条
每小时：360,000 条
每天：8,640,000 条
三天：25,920,000 条
```

由于表中包含 JSON `payload`，实际压力不仅来自行数，还包括数据文件、二级索引、undo/redo、binlog、备份和主从复制。

即使只保留 1 天，数据规模仍然过大，因此不采用按天数长期保留全部成功记录的方案。

## 3. 推荐方案：单张有界活动表

保持 `t_itick_snapshot_outbox` 为单张活动表，只长期保存尚未完成的任务：

1. 所有权威快照仍实时写入 Outbox。
2. Worker 完成 Redis 和 Option 发布后将记录置为 `status=3`。
3. 成功记录保留 10～30 分钟，随后分批删除。
4. `pending`、`processing`、`failed` 记录不得由清理任务删除。
5. `manual` 记录保留供人工处理；数量较大时迁移到独立 dead-letter 表。
6. 成功率、处理耗时和失败次数写入监控系统，不依赖成功记录进行长期统计。

以成功记录保留 30 分钟估算，正常情况下活动数据约为：

```text
100 产品 × 60 秒 × 30 分钟 = 180,000 条
```

未完成和异常任务会在此基础上增加，但可通过监控及时发现。

## 4. 自动清理设计

清理任务由 iTick 服务自身启动，采用与现有 Snapshot Outbox Worker 类似的后台任务方式。应用任务比 MySQL Event 更容易配置、测试、监控和限速。

建议默认参数：

| 参数 | 建议值 |
| --- | --- |
| 成功记录保留时间 | 30 分钟 |
| 调度间隔 | 1 分钟 |
| 单批删除数量 | 5,000 条 |
| 单轮最大批次 | 10 批 |
| 批次间隔 | 100 毫秒 |

清理 SQL：

```sql
DELETE FROM t_itick_snapshot_outbox
WHERE status = 3
  AND update_times < ?
ORDER BY id
LIMIT 5000;
```

建议增加清理索引：

```sql
ALTER TABLE t_itick_snapshot_outbox
ADD INDEX idx_outbox_cleanup (status, update_times, id);
```

每轮清理需要设置最大批次数，防止历史积压导致清理任务长时间占用数据库。若某批影响行数小于批大小，可提前结束本轮任务。

需要记录以下监控指标：

- 各状态记录数。
- 最老未完成记录的等待时间。
- 每轮清理行数、耗时和错误数。
- Outbox 写入速率和成功消费速率。
- `status=5` 数量。
- 数据库复制延迟、锁等待和磁盘使用率。

## 5. 幂等要求

当前表使用 `snapshot_id` 唯一索引防止相同快照重复进入 Outbox。成功记录删除后，相同 `snapshot_id` 理论上可能再次进入并重新投递。

实施自动删除前必须保证：

1. Redis 发布以 `snapshot_id` 或版本号保证重复写入安全。
2. 永久档案作为生产端去重依据：`t_itick_authoritative_snapshot` 已存在相同 `snapshot_id` 时，不再创建新的 Outbox 记录。
3. Option 的 `SyncMarketQuote` 当前会追加行情快照，不能将其视为天然幂等；生产端去重不可省略。消费端后续仍建议增加 `snapshot_id` 去重作为纵深保护。
4. 若生产端和消费端都不能保证幂等，应增加独立的轻量去重机制，而不是永久保存包含 JSON payload 的成功 Outbox。

可选的过渡方案是成功后清空 `payload`，只保留 `snapshot_id` 和状态字段。但这仍会导致行数持续增长，仅适合作为短期措施。

## 6. 为什么不推荐按小时分表

不采用 `t_itick_snapshot_outbox_YYYYMMDDHH` 形式的小时物理分表，原因如下：

- 每天产生 24 张表，每年产生 8,760 张表。
- Worker 必须同时扫描当前小时和历史小时的重试任务。
- 跨小时仍未完成的记录容易漏处理或误删除。
- 后台列表、状态统计和人工重试需要跨表查询。
- `snapshot_id` 无法由 MySQL 在多张表间维持全局唯一。
- 建表、路由、切换和删除逻辑会显著增加运维复杂度。

小时分表只是改变数据存放位置，不能解决 Outbox 成功数据没有生命周期的问题。

## 7. 分区或归档备选方案

只有在小批量删除经压测确认无法承受时，才考虑以下备选方案。

### 7.1 MySQL RANGE 分区

可按时间进行日级分区，通过 `DROP PARTITION` 快速清理历史数据，不建议按小时建立大量分区。

需要注意：MySQL 要求唯一索引包含所有分区字段。若按 `create_times` 分区，当前唯一索引：

```sql
UNIQUE KEY uk_snapshot_outbox (snapshot_id)
```

需要调整为包含 `create_times`，这会失去 `snapshot_id` 的全表唯一约束。因此必须先完成消费端幂等或独立去重设计。

### 7.2 活动表与成功归档表

如果业务确实要求短期查询全部成功投递记录，可采用：

```text
t_itick_snapshot_outbox
  只保存未完成任务

t_itick_snapshot_outbox_success_YYYYMMDDHH
  保存成功日志，到期直接 DROP TABLE
```

Worker 始终只扫描活动表。成功归档应异步执行，不能阻塞行情处理链路。由于权威快照已有永久档案，默认不启用该方案。

## 8. 历史存量处理

已有海量数据不能通过一个大事务一次删除，建议按以下顺序处理：

1. 确认数据库磁盘空间足够创建清理索引。
2. 增加 `(status, update_times, id)` 索引。
3. 在低峰期按每批 2,000～5,000 条删除历史 `status=3` 数据。
4. 批次之间限速，并观察复制延迟、磁盘 IO、锁等待和 binlog 增长。
5. 存量下降到目标规模后，启动应用内常态化清理任务。
6. 清理完成后检查表空间；是否执行 `OPTIMIZE TABLE` 需根据磁盘空间和维护窗口单独评估，不能直接在线执行。

## 9. 与权威快照永久档案的边界

清理 Outbox 只能控制投递队列表的规模。如果普通实时 tick 都写入 `t_itick_authoritative_snapshot`，永久档案仍会以每天约 864 万条的速度增长。

后续需要单独确认权威快照的产生规则：

- 普通实时行情应主要通过 Redis 或消息系统流转。
- 只有结算、行权、审计等确需长期追溯的时间点才生成永久权威快照。
- Outbox 仅负责这些权威快照的可靠异步发布。

该问题属于权威档案的数据生命周期设计，不应通过 Outbox 分表规避。

## 10. 实施顺序

建议分阶段上线：

1. 确认 Redis 发布幂等，并上线基于永久档案的生产端去重；Option 后续补充消费端去重。
2. 添加清理索引和监控指标。
3. 低峰期分批清理历史成功记录。
4. 上线应用内自动清理任务，初始保留时间设为 30 分钟。
5. 观察至少一个完整业务周期，再决定是否缩短到 10 分钟。
6. 单独评审 `t_itick_authoritative_snapshot` 的生成频率和永久保留策略。

## 11. 最终决策

当前推荐决策如下：

- 保留 Outbox 的实时事务写入机制。
- 不采用小时物理分表作为默认方案。
- Outbox 使用单张有界活动表。
- 成功记录默认保留 30 分钟并分批自动删除。
- 未完成、失败和人工处理记录不自动删除。
- 在启用清理前完成消费端幂等确认。
- 对永久权威快照的增长问题另行治理。

## 12. Redis 权威快照缓存改造

原有 Redis v1/v2 结构为每条快照创建一个 String Key，并在 Sorted Set 中再次保存 `snapshot_id`。活跃产品的索引 TTL 会被持续刷新，历史成员无法自然过期，因此不再继续使用该结构。

新 v3 结构如下：

```text
market:authoritative:v3:latest:<authority>:<kind>:<category>:<market>:<symbol>
  每个产品只保存最新完整快照

market:authoritative:v3:hot:<authority>:<kind>:<category>:<market>:<symbol>
  Sorted Set 直接保存短期快照 JSON，默认窗口 30 分钟
```

新结构不再创建 `market:authoritative:v1:<snapshot_id>` 独立 Key。每次发布时按 `source_timestamp` 清理热点窗口之前的成员；不活跃产品的热点索引整体 TTL 为热点窗口的两倍，latest Key TTL 为 7 天。

历史读取采用两级策略：

1. Trade 结算先查询 Redis v3 热点窗口。
2. Redis 未命中时，通过 `ItickInternal/GetAuthoritativeSnapshot` 查询 `t_itick_authoritative_snapshot`。
3. MySQL 查询排除 `t_itick_snapshot_revocation` 中已经撤销的快照。

iTick 启动重建只读取每个 `authority + snapshot_kind + product` 的最新快照，不再扫描并发布完整永久档案。重启后的热点历史逐步由实时行情填充，在此期间历史查询通过 MySQL 回源保证可用。

旧缓存使用 `SCAN + UNLINK` 限速清理，并且默认关闭：

```yaml
AuthoritativeCache:
  HotWindowMinutes: 30
  LegacyCleanupEnabled: false
  LegacyCleanupScanCount: 500
  LegacyCleanupIntervalSeconds: 1
```

只有 Common、iTick 和 Trade 全部升级到 v3，且 MySQL 回源验证完成后，才可将 `LegacyCleanupEnabled` 改为 `true`。清理任务保留撤销 tombstone，不使用阻塞式 `KEYS` 或 `DEL`。
