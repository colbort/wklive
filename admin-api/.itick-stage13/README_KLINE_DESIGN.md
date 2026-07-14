# iTick K 线历史与增量更新设计（待确认）

> 目标：只维护租户实际使用的产品，在服务重启、Redis 丢失、WebSocket 断线、消息乱序或定时任务重复执行时，仍能自动恢复并最终得到完整、一致的 K 线数据。

## 1. 设计结论

1. MySQL 是“租户正在使用哪些产品”的唯一真源；Redis 是可重建的运行时索引，不能成为唯一依据。
2. 实时 K 线使用 **tick** 生成，不使用 quote 生成：quote 是最新行情快照，成交量通常是累计值，且采样可能遗漏区间高低点；tick 更适合形成 OHLCV。
3. 本地只直接聚合 `1m`；`5m/15m/30m/1h/1d/1w/1mo/1y` 从已落库的低周期 K 线逐级或直接聚合。这样可减少计算量，并使高周期能够随 1m 修正而确定性重算。
4. 本地实时聚合结果是“临时/低延迟数据”；iTick REST 批量 K 线是“校准数据”。每 5 分钟拉取最近窗口，采用幂等 upsert 覆盖并重算受影响的高周期。
5. 历史回补和 5 分钟增量校准分成两个独立任务，使用不同队列、并发和限流，避免首次历史同步拖慢实时补洞。
6. 所有时间统一保存 UTC 毫秒时间戳；日、周、月、年 K 线必须按产品所属市场的时区、交易时段和交易日历切桶，不能固定按 UTC 的 24 小时、7 天或 30 天计算。

## 2. 数据职责

### 2.1 MySQL

- `t_itick_product`：产品主数据。
- `t_itick_tenant_product`：租户产品关系，是活跃产品集合的来源。
- `t_itick_kline_sync_progress`：每个产品、周期的历史回补和增量校准水位。
- 建议新增 `t_itick_product_change_outbox`：租户产品变更事件。租户关系和 outbox 在同一事务提交，消费者异步更新 Redis，解决“数据库成功但 Redis 更新失败”的双写不一致。

活跃产品建议定义为：

```text
t_itick_product.enabled = 1
AND EXISTS (
  SELECT 1 FROM t_itick_tenant_product
  WHERE product_id = t_itick_product.id AND enabled = 1
)
```

`app_visible` 只控制展示，不建议控制行情采集；否则仅隐藏产品就会中断行情。若业务明确隐藏即停采，再将它加入判断条件。

建议补索引：

```sql
ALTER TABLE t_itick_tenant_product
  ADD KEY idx_product_enabled_tenant (product_id, enabled, tenant_id);
```

### 2.2 Redis

Redis key 建议带版本前缀，方便以后平滑迁移：

| Key | 类型 | 用途 |
| --- | --- | --- |
| `itick:v1:active_products` | Set | 活跃 `product_id` 去重集合 |
| `itick:v1:product:{id}` | Hash | category/market/symbol/timezone/session 等订阅元数据 |
| `itick:v1:product_refcount` | Hash | `product_id -> 启用租户数`，用于增删判断 |
| `itick:v1:quote:{product_id}` | Hash/String | 最新 quote，设置 TTL |
| `itick:v1:depth:{product_id}` | String | 最新完整 depth 快照，设置短 TTL |
| `itick:v1:tick:last:{product_id}` | String | 最新 tick，设置 TTL |
| `itick:v1:ticks:{product_id}` | Stream | 用于 1m 聚合和故障恢复的短期 tick 流，按长度/时间裁剪 |
| `itick:v1:kline:building:{product_id}:1m` | Hash | 当前未闭合 1m K 线 |
| `itick:v1:kline:dirty` | ZSet | 需要重算的产品/周期/时间桶 |

quote/depth/tick 都是实时缓存，不作为永久历史真源。TTL 建议：quote/tick-last 10～30 分钟，depth 1～5 分钟；tick Stream 根据容量保留 1～6 小时。实际值需根据产品数、tick 频率和 Redis 内存压测确定，不能无限保留。

Redis 集合必须支持全量重建：服务启动时以及周期性对账任务从 MySQL 扫描活跃产品，在临时 key 构建完成后 `RENAME` 原子替换 active set/refcount，修复漏事件、人工改库和 Redis 数据丢失。

## 3. 租户产品变更流程

### 3.1 新增或由禁用改为启用

1. 一个 MySQL 事务内写 `t_itick_tenant_product` 和 outbox。
2. outbox 消费者查询该产品当前启用租户数（不要只依赖传入事件推算）。
3. 使用 Lua 原子更新 refcount、active set 和产品元数据。
4. 当计数从 `0 -> N` 时发布 `product.activate`：
   - 建立 quote、tick、depth 的上游订阅；
   - 创建各周期同步进度；
   - 将最近窗口校准和历史回补任务入队。
5. 重复事件必须幂等，不能重复增加引用计数。

### 3.2 删除或由启用改为禁用

1. 一个 MySQL 事务内更新关系和 outbox。
2. 消费者重新查询数据库中的启用租户数。
3. 仍有其他租户使用时只更新 refcount，不移除 active set、不取消订阅。
4. 计数为 0 时移出 active set，发布 `product.deactivate` 并取消上游订阅。
5. 不立即删除 MongoDB 历史 K 线和同步进度；保留数据能在产品重新启用时快速追平。历史数据保留期限由独立的数据生命周期策略控制。

现有批量更新接口还未实际写入数据。实现时应先计算变更前后集合差异，再在一个事务内批量 upsert/delete 与写 outbox，避免逐条操作产生中间状态。

## 4. 实时行情与 K 线生成

### 4.1 WebSocket 接收

只订阅 `active_products` 中产品。收到消息后先规范化唯一身份：

```text
product_id <-> category_code + market + normalized_symbol
```

- quote：覆盖写最新快照；可继续推送给期权等下游。
- depth：如果上游是增量深度，必须维护 sequence 并在断档时重新获取快照；不能把增量包当成完整盘口覆盖。
- tick：按 `(product_id, upstream_trade_id)` 去重；如果上游没有 trade id，使用稳定字段生成短期去重键。写入短期 Redis Stream，再交给 1m 聚合器。
- 所有数据校验时间戳、价格、成交量；明显未来时间、非法负值进入错误流并报警。

### 4.2 1m 聚合

以交易所事件时间而非服务接收时间切桶：

```text
bucket_start = floor(event_ts / 60_000) * 60_000
open  = 桶内事件时间最早 tick 的价格
high  = max(price)
low   = min(price)
close = 桶内事件时间最晚 tick 的价格
volume/turnover = 桶内增量求和
```

需注意：如果 tick 的 `volume` 是当日累计量，必须先按相邻 tick 做差；出现跨交易日、负差或重置时重新建立基线，不能直接累加累计值。此字段语义需要用 iTick 实际报文验证后再实现。

每个自然分钟结束后，聚合器在下一次 5 秒检查内将刚闭合的 `1m` K 线写入 MongoDB，不等待额外迟到窗口。因此在 `12:01/12:02/.../12:05` 会依次落库 `[12:00,12:01)/[12:01,12:02)/.../[12:04,12:05)`。迟到、漏失或上游修订的数据不再修改已经闭合的内存桶，而由每 5 分钟 REST 最近 30 根 `1m` 批量 upsert 统一补齐和校正。写入唯一键为 `(market, symbol, ts)`，任务重试不会产生重复数据。

没有 tick 不代表缺数据：非交易时段不生成 K 线；交易时段内是否用上一收盘价补零成交量 K 线，应按 iTick REST 的口径决定。默认不凭空补 K 线，交给 REST 校准确认。

### 4.3 高周期聚合

- `5m/15m/30m/1h`：从已确认的 1m 聚合。
- `1d`：从 1m 或 1h 按市场交易日聚合。
- `1w/1mo/1y`：从已确认的 1d 按市场日历聚合。
- 聚合规则：首根 open、最高 high、最低 low、末根 close、volume/turnover 求和。
- 只要底层 K 线被 REST 修正，就将所有包含它的上层桶写入 dirty ZSet，由重算 worker 幂等覆盖。
- 当前代码只支持到 `1mo`，`1y` 需要新增周期定义、存储集合、查询枚举与聚合逻辑。iTick 若没有年度 REST 周期，则由日线确定性生成。

MongoDB 文档建议增加：`source`（realtime/rest/derived）、`confirmed`、`revision`、`updatedAt`。前端查询允许返回正在形成的 K 线，但应能区分 `closed/confirmed`。

## 5. REST 历史回补与每 5 分钟校准

### 5.1 五分钟校准任务

建议在整 5 分钟结束后延迟 20～60 秒执行，避免上游 K 线尚未最终结算。任务流程：

1. 从 Redis 活跃集合取产品；Redis 不可用时降级从 MySQL 查询。
2. 按 category/market 和 iTick 批量接口上限分组，请求最近滑动窗口，不逐产品逐周期无界调用。
3. 重点校准 `1m` 最近 15～30 分钟，覆盖断线、迟到和上游修订；如果批量接口支持多周期，可低频抽检上游高周期。
4. REST 返回先校验时间桶、排序、重复项和 OHLC 合法性，再 bulk upsert。
5. 对发生新增或数值变化的 1m 标记高周期 dirty 桶并重算。
6. 只有数据落库成功后才推进 progress 水位；部分产品失败只重试失败项，不能让整批成功项回滚或水位误进。

为避免所有实例重复执行，调度器使用 Redis leader lock；具体 job 使用稳定幂等键，例如：

```text
reconcile:{product_id}:1m:{window_end}
```

任务必须有指数退避、随机抖动、429/5xx 分类重试、超时和熔断；全局限流要低于 iTick 套餐上限并预留实时接口余量。

### 5.2 历史回补任务

产品首次激活后先回补近期数据，保证用户立即可查，再后台向过去分页：

1. `1m`：按业务保留期回补，而不是默认无限追溯。
2. `1d`：可直接拉更长历史，快速满足长期图表。
3. 其他周期优先由底层数据生成；如需与供应商完全一致，可追加低优先级 REST 校验。
4. 每页成功落库后更新 `oldest_ts`；到达配置的历史起点或上游明确返回无数据才设 `full_synced=1`。

当前逻辑以“返回数量小于 limit”判断全量完成不够稳健：节假日、停牌、接口分页行为都可能导致短页。应同时记录目标历史边界，并使用连续空页/明确结束条件确认。

## 6. 完整性判定与补洞

`latest_ts` 只能表示见过的最大时间，不能证明中间连续。建议保留现有 `contiguous_ts`，并补充：

- `target_from_ts`：配置要求的历史起点。
- `last_reconcile_window_end`：最近确认完成的校准窗口。
- `consecutive_failures`、`next_retry_at`：失败调度。
- `last_error_code`：区分限流、网络、参数和无数据。

每天运行低优先级 gap scanner：按市场日历计算“本应存在”的时间桶，与 MongoDB 实际桶比较。缺口进入 repair queue，由 REST 定点补洞；补洞后再次验证才推进 `contiguous_ts`。停牌、节假日和非交易时段不算缺口。

建议建立两级校验：

- 结构校验：唯一键、时间对齐、`low <= open/close <= high`、非负量。
- 一致性校验：抽样比较本地聚合与 iTick 高周期 K 线，差异超过价格精度/成交量容忍度即报警并重算。

## 7. 并发、幂等和故障恢复

- 活跃产品变更：MySQL transaction + outbox；Redis 更新使用 Lua 原子脚本。
- 行情消费：按 `product_id` 一致性分区，同一产品在同一时刻只由一个聚合 owner 处理。
- MongoDB：唯一索引 + upsert；REST 数据优先级高于 realtime，derived 数据可被底层修正触发覆盖。
- Redis 丢失：从 MySQL 重建 active set；从 MongoDB 最后一根 K 线和 REST 最近窗口恢复聚合状态。
- WebSocket 重连：重新加载活跃集合并恢复订阅，立即为断线时间段提交 REST repair job。
- 服务滚动发布：租约续期和 leader lock 必须有 TTL；旧 owner 失效后新 owner 可接管。
- 产品停用：停止新增数据，不删除历史；未完成队列任务执行前再次检查产品是否活跃。

## 8. 查询策略

查询 K 线时：

1. 已闭合历史从 MongoDB 读取。
2. 当前形成中的桶从 Redis 合并到结果末尾。
3. 查询层按唯一键去重，以 Redis 当前桶覆盖同时间的未确认 MongoDB 数据。
4. 返回 `confirmed`/`isClosed`，避免客户端把临时 K 线当最终值。
5. Redis 不可用时仍可返回 MongoDB 已闭合数据，接口降级但不失败。

## 9. 监控与告警

至少暴露以下指标：

- 活跃产品数、各 topic 实际订阅数、Redis refcount 与 DB 实际数量差异。
- WebSocket 连接/重连次数、最后消息时间、乱序/重复/非法 tick 数。
- 各任务队列长度、处理耗时、成功率、重试次数、iTick 429/5xx。
- 每产品/周期最后确认时间、最大延迟、缺口数、REST 与本地聚合差异数。
- Redis Stream 长度/内存、MongoDB bulk upsert 耗时和失败数。

关键告警：交易时段内产品长时间无 tick、校准连续失败、gap 持续未修复、active set 与 DB 不一致、任务堆积超过一个校准周期。

## 10. 推荐实施顺序

### 第一阶段：活跃产品与可靠订阅

- 完成批量租户产品写入。
- 增加 outbox、active set/refcount、启动全量重建和周期对账。
- 订阅范围从全部产品改为活跃产品，支持 0->1 订阅和 1->0 退订。

### 第二阶段：REST 完整性链路

- 将历史回补和 5 分钟增量校准拆开。
- 增量任务使用批量接口、滑动窗口、幂等 job、失败项重试。
- 修正 progress 推进规则，加入市场日历 gap scanner。

### 第三阶段：实时聚合

- Redis 保存 quote/depth/tick，tick Stream 限长。
- 实现 tick -> 1m、迟到乱序处理、当前桶查询合并。
- REST 修正 1m 后触发高周期 dirty 重算。

### 第四阶段：高周期与年度 K 线

- 按市场时区/交易日历生成 `5m` 到 `1y`。
- 增加 source/confirmed/revision 字段与一致性抽检。
- 压测 Redis 内存、MongoDB 写入、iTick 配额和大规模重连恢复。

## 11. 实现前需要确认的业务项

1. `enabled=1` 是否就是“租户正在使用”；`app_visible=2` 是否仍需持续采集？本文默认仍采集。
2. 各品类的 1m 历史保留期分别是多少；是否真的需要无限历史？
3. iTick tick 中 `volume` 是单笔量还是当日累计量，是否提供稳定 trade id/sequence。
4. iTick 所谓“批量 K 线接口”的请求上限、频率、是否支持多产品及多周期一次请求。
5. 股票/期货是否包含盘前盘后、夜盘；日线使用交易所时区还是供应商固定口径。
6. 无成交分钟是否需要以前收补 OHLC、volume=0；建议以 iTick REST 返回口径为准。
7. MongoDB 历史数据的保留策略，以及产品停用后保留多久。

以上业务项确认后，再确定表结构迁移、Redis key 最终格式、任务拆分和代码改造清单。
