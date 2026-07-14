# iTick K 线历史同步与增量更新实现流程

> 本文以 `services/itick` 当前代码为准，说明历史数据同步、实时 `1m` 聚合、每 5 分钟校正及高周期派生的实际调用链。未实现或仍有风险的部分单独列在文末。

## 1. 总体数据链路

```text
                                ┌─ iTick 单产品 /{category}/kline
管理后台手动同步指定产品/周期 ────┤
                                └─ 向历史方向分页 → MongoDB upsert

任务中心每 5 分钟触发 SyncKlines
        ↓
Redis/MySQL 加载活跃产品
        ↓
iTick 批量 /{category}/klines，最近 30 根 1m
        ↓
校验已闭合 K 线 → MongoDB upsert
        ↓
DerivedAggregator 重算受影响的高周期 → MongoDB + Redis

iTick WebSocket tick
        ↓
TickAggregator 按事件时间聚合 1m
        ↓ 每个自然分钟结束，最多约 5 秒检查延迟
BatchWriter 批量 upsert MongoDB
        ↓ 写入成功回调
DerivedAggregator 重算受影响的高周期 → MongoDB + Redis

iTick WebSocket kline@N
        ↓
只覆盖 Redis 实时 K 线缓存 → App WS 每 5 秒读取并推送
```

核心原则：

1. MongoDB 保存历史查询数据，写入使用 `(market, symbol, ts)` 幂等 upsert。
2. Redis 保存实时行情快照和正在变化的 K 线，不作为永久历史真源。
3. 本地 Tick 只直接生成 `1m`；其他周期从已落库低周期确定性重算。
4. iTick REST 最近窗口数据用于补齐、覆盖本地 Tick 聚合的临时结果。
5. 历史全量同步和每 5 分钟校正是两个独立入口，不能混用。

## 2. 活跃产品与订阅来源

### 2.1 MySQL 真源

活跃产品由 `TItickProductModel.FindActivePage` 查询，条件是：

```sql
t_itick_product.enabled = 1
AND EXISTS (
  SELECT 1
  FROM t_itick_tenant_product
  WHERE product_id = t_itick_product.id
    AND enabled = 1
)
```

因此多个租户使用相同产品时只采集一次；只要仍有一个启用租户使用，产品就不会退出订阅集合。

### 2.2 Redis 运行时索引

`ItickManager.rebuildActiveProducts` 从 MySQL 全量重建：

| Key | 类型 | 当前用途 |
| --- | --- | --- |
| `itick:v1:active_products` | Set | 活跃产品 ID |
| `itick:v1:product:{id}` | Hash | `category_code/market/symbol/exchange` |

使用临时 Set 构建后 `RENAME`，避免重建过程中读到半套 active set。Redis 丢失时可从 MySQL 恢复。

### 2.3 WebSocket 订阅集合

每个活跃产品生成以下订阅：

```text
depth, tick, quote,
kline@1, kline@2, kline@3, kline@4,
kline@5, kline@8, kline@9, kline@10
```

内部先使用 `1m/5m/15m/30m/1h/1d/1w/1mo`，发送前由 `IntervalToStream` 转换成 iTick 的 `kline@N`。

订阅按 `categoryCode` 分给对应的 `ItickWsClient`，保存在进程内存 `desiredSubs`。它是完整快照替换，不是逐条追加。连接认证成功、订阅变化或重连后，`syncDesiredSubscriptions` 负责退订旧组合、订阅新组合。

服务启动、创建/更新/批量更新租户产品时会刷新订阅集合。

## 3. MongoDB 存储与查询

MongoDB collection 按品类和周期拆分：

```text
{category}_kline_{interval}
```

例如：

```text
crypto_kline_1m
crypto_kline_5m
forex_kline_1h
```

`CoinKlineModel.UpsertBySymbolTs` 和 `BulkUpsertBySymbolTs` 使用：

```text
market + symbol + ts
```

作为匹配条件，相同时间桶重复同步只会覆盖，不会新增重复文档。

K 线同时保存 `source/sourcePriority/revision/isClosed/confirmed/actualCount/expectedCount`。MongoDB 更新管道执行 `rest(300) > derived(200) > realtime(100)` 的来源优先级和版本比较，低优先级或旧版本不能覆盖已存在的高优先级新数据。

App 的 `GetKline` 只读取 MongoDB：

```text
KType → interval → Factory.New(category, interval)
      → FindBeforeTsByMarketSymbol
```

该接口不会在查询时调用 iTick REST。当前形成中的 K 线由 App 的订阅流另外从 Redis 获取，是否与历史列表合并由客户端处理。

## 4. 管理后台历史 K 线同步

### 4.1 入口与职责

入口：

```text
SyncProductKlineHistoryLogic.SyncProductKlineHistory
```

用途：管理员为一个指定产品、指定周期主动拉取历史 K 线。它不是每 5 分钟任务，也不会扫描全部活跃产品。

请求确定：

- `categoryCode`
- `market`
- `symbol`
- `kType`
- `endTs`，为空时使用当前时间之后 1 毫秒

### 4.2 使用的 iTick 接口

历史同步使用单产品接口：

```text
GET /{category}/kline
```

关键参数：

```text
region={market}
code={symbol}
kType={指定周期}
et={当前分页结束时间}
limit=500
```

该接口支持 `et`，因此能够从指定结束时间向历史方向分页。这里不能换成批量 `/klines`，因为批量接口只适合获取最近窗口，没有历史起止范围参数。

### 4.3 向后分页算法

`FetchProductHistory` 调用 `syncBackwardRange`：

```text
et = 请求 endTs 或当前时间 + 1

while true:
    请求最多 500 根
    空结果                     → 完成
    过滤掉 ts > 初始 endTs     → 防止越界
    当前页 bulk upsert MongoDB
    返回数量 < 500             → 完成
    minTs <= 0                 → 完成
    minTs >= 当前 et           → 防止游标不前进，完成
    et = minTs - 1             → 继续向过去请求
```

接口没有对管理员暴露 `limit/maxPages`。同步会一直向过去翻页，直到 iTick 返回短页、空页或游标无法继续推进。

返回的 `syncedCount` 是本次写入列表的累计数量；因为底层是 upsert，它不等同于 MongoDB 实际新增文档数。

### 4.4 当前历史同步边界

1. 历史同步只处理管理员指定的周期。
2. 如果管理员同步的是 `1m`，每页成功落库后会通过 `DerivedWorker` 等待受影响高周期重算完成；失败会终止历史同步并返回错误。
3. 管理员直接同步其他周期时只 upsert 指定周期，不再反向影响低周期或其他周期。
4. 当前历史同步没有更新 `t_itick_kline_sync_progress`，进度表主要由 5 分钟校正流程推进。
5. 单次请求可能持续很久；当前没有保留期边界、取消任务记录或断点续传水位。

## 5. Tick 实时生成 1m

### 5.1 启动与输入

服务启动时：

```text
NewTickAggregator(Writer)
MarketDataCache.SetTickHandler(tickAggregator.Add)
tickAggregator.Start()
```

iTick WS 的 Tick 先写 Redis 最新 Tick 缓存，再同步调用 Tick handler。

### 5.2 事件时间切桶

```text
bucketTs = floor(tick.ts / 60_000) * 60_000
```

每个 `(category, market, symbol, bucketTs)` 保存一个内存桶：

```text
open     = 事件时间最早 Tick 的价格
high     = 最高 Tick 价格
low      = 最低 Tick 价格
close    = 事件时间最晚 Tick 的价格
volume   = Tick volume 累加
turnover = Tick turnover 累加
```

当前校验规则：

- 时间不能超过当前时间 30 秒；
- 只接受最近 10 分钟内的 Tick；
- 价格必须大于 0；
- volume/turnover 不能为负；
- 使用 `product + ts + price + volume + turnover` 指纹短期去重。

### 5.3 每分钟闭合

聚合器每 5 秒扫描一次：

```text
bucketTs + 60_000 <= now
```

满足条件的 `1m` 被视为闭合，送入 `BatchWriter`。同一产品已经闭合的时间桶记录在 `finalized`，之后到达的迟到 Tick 不再重新打开该桶。

示例：

```text
12:01 → 写 [12:00, 12:01)
12:02 → 写 [12:01, 12:02)
12:03 → 写 [12:02, 12:03)
12:04 → 写 [12:03, 12:04)
12:05 → 写 [12:04, 12:05)
```

实际触发可能比整分钟晚 0～5 秒，之后还需等待 BatchWriter 的 flush interval。

### 5.4 BatchWriter

BatchWriter 按：

```text
category + market + interval
```

分组缓冲，达到 batch size 或 flush interval 后执行 MongoDB bulk upsert。`1m` 写入成功后调用 `flushHandler`，将重算请求快速放入独立 `DerivedWorker` 队列，避免 MongoDB 高周期查询阻塞 BatchWriter：

```text
DerivedWorker.Enqueue → DerivedAggregator.Rebuild
```

因此只有 MongoDB `1m` 写入成功，才向上重算高周期。队列满时记录错误；BatchWriter 不会等待异步高周期重算完成。

### 5.5 当前 Tick 聚合风险

iTick 股票 Tick 的 `v` 按当日累计成交量处理：第一条或累计值重置时只建立基线，后续使用 `current.v - previous.v` 作为分钟增量；乱序 Tick 不推进基线。股票 Tick 没有可靠 `tu` 时，临时 turnover 使用 `deltaVolume × lastPrice` 估算，最终仍以 5 分钟 REST K 线校正为准。其他品类目前仍按 Tick 自带 volume/turnover 为单次增量直接累加，需要继续根据各品类官方字段语义核对。

当前分钟桶和股票累计量基线在每次 Tick 更新时同步保存到 Redis。服务启动时扫描并恢复未过期状态；分钟桶成功进入 BatchWriter 后才删除 Redis building key。没有 Tick 的分钟不会凭空生成 K 线，仍由 REST 校准补齐。

## 6. 高周期派生

### 6.1 触发入口

`DerivedAggregator.Rebuild` 有两个入口：

1. Tick 生成的 `1m` 被 BatchWriter 成功写入 MongoDB后；
2. 每 5 分钟 REST 最近窗口 `1m` 被成功 upsert 后。

两条路径共用同一套重算逻辑，保证 REST 修正能够向所有上层周期传播。

### 6.2 周期层级

```text
1m ──→ 5m
   ├─→ 15m
   ├─→ 30m
   └─→ 1h ──→ 1d ──→ 1w
                    └─→ 1mo
```

持久化和实时派生支持：

```text
1m, 5m, 15m, 30m, 1h, 1d, 1w, 1mo
```

`1y` 已定义 Proto 枚举和查询入口；App 查询 `1y` 时从 MongoDB `1d` 数据按市场日历年度边界动态聚合，不订阅 iTick WS `kline@11`，也不单独落 `1y` collection。

### 6.3 重算方法

重算先按产品和目标周期合并受影响桶，用一次 `FindRangeByMarketSymbol` 读取覆盖这些桶的完整源区间，在内存分桶后通过一次 MongoDB BulkWrite 写入，避免逐桶小查询：

```text
open     = 第一根源 K 线 open
high     = max(源 K 线 high)
low      = min(源 K 线 low)
close    = 最后一根源 K 线 close
volume   = sum(源 K 线 volume)
turnover = sum(源 K 线 turnover)
```

不是在旧高周期结果上做增量加减，因此底层 `1m` 被 REST 覆盖后可以确定性重算。派生结果记录实际和预期源 K 线数量；只有目标桶已闭合、数量完整且所有源 K 线均为 confirmed 时，才标记 `confirmed=true`。

结果执行两次写入：

1. upsert 对应周期 MongoDB collection；
2. 覆盖 Redis 对应周期实时缓存，供 App WS 推送。

### 6.4 市场时间边界

- `5m/15m/30m/1h`：固定宽度切桶；
- `1d/1w/1mo/1y`：通过 `t_itick_market_calendar` 的 IANA 时区、交易日偏移和周起始日计算；
- `t_itick_market_session` 保存交易时段，`t_itick_market_holiday` 保存休市和特殊交易日，为缺口扫描提供判断基础；
- 未配置市场日历时安全回退 UTC、周一为周起始日。

Resolver 按 `category + market + exchange` 查找定义并缓存；目前派生 K 线模型未保存 exchange，因此派生链使用 `category + market` 的默认日历记录，配置时必须为各市场提供 exchange 为空的默认行。

## 7. 每 5 分钟增量校正

### 7.1 任务入口

任务中心向 iTick 服务发送：

```text
ActionItickSyncKlines
```

`tasks/subscriber.go` 调用 `SyncKlinesLogic.SyncKlines`。代码本身不创建本地 ticker；是否严格每 5 分钟执行由外部任务中心的调度配置决定。

### 7.2 分布式互斥与任务记录

`SyncKlinesLogic`：

1. 校验 iTick API URL 和 Token；
2. 根据 `apiURL + token` 生成 Redis lock key；
3. 获取 30 秒分布式锁；
4. 新增 `t_itick_sync_task`，类型为 `reconcile_klines`；
5. 启动最多 10 分钟的后台 goroutine；
6. worker 每 10 秒续租，TTL 30 秒；
7. 完成、失败或 panic 时更新任务状态并释放锁。

如果已有任务持锁，本次请求返回“任务正在运行”，不会重复启动。

### 7.3 加载和分组产品

优先从：

```text
SMEMBERS itick:v1:active_products
```

读取产品 ID，再批量查询 MySQL 产品信息。Redis 报错或集合为空时，使用 `FindActivePage` 从 MySQL 分页扫描。

过滤无效或不支持 K 线的品类后，按以下维度分组：

```text
category + market + exchange
```

每组最多 10 个产品调用一次批量接口。

### 7.4 iTick 最近窗口请求

使用批量接口：

```text
GET /{category}/klines
```

固定参数：

```text
kType=1     // 1m
limit={ReconcileWindowBars，默认 30}
codes=最多 10 个产品
region={market}
exchange={exchange，可选}
```

请求限流器当前配置约为 400 次/分钟、burst=1；HTTP 超时 20 秒。

批量接口没有 `start/end/et`，因此这里不做历史分页，只用于有限最近窗口校正。

### 7.5 校验、落库和传播

每个返回项只接受：

```text
ts > 0
ts <= 当前最后一个已闭合 1m
ts % 60_000 == 0
low <= open/close <= high
high >= low
volume >= 0
turnover >= 0
```

通过校验的数据：

```text
bulk upsert 1m MongoDB
        ↓
DerivedAggregator.Rebuild
        ↓
重算受影响的 5m/15m/30m/1h/1d/1w/1mo
        ↓
更新高周期 MongoDB 和 Redis
        ↓
更新 t_itick_kline_sync_progress
```

进度记录当前更新：

- `latest_ts`
- `oldest_ts`
- `last_reconcile_ts`
- mode=`reconcile`
- 本次覆盖数量说明

单个产品缺少返回、MongoDB 写入、派生重算或进度更新失败时，会记录该产品错误并继续处理同批其他产品；批次结束后返回产品级汇总错误。其他批次也会继续处理。最终只要存在失败，整个同步任务状态标记失败。

### 7.6 12:00～12:05 示例

```text
12:01  Tick 聚合并写入 12:00 这根 1m，高周期随之更新
12:02  写入 12:01 这根 1m，高周期再次更新
12:03  写入 12:02 这根 1m
12:04  写入 12:03 这根 1m
12:05  写入 12:04 这根 1m
12:05 任务执行后，iTick 最近 30 根 1m 覆盖 [12:00,12:05) 等最近窗口
       所有受影响高周期再次从 MongoDB 源数据完整重算
```

实际生产建议在整 5 分钟后延迟 20～60 秒调度，避免 iTick 刚闭合的 K 线尚未稳定。

## 8. iTick WebSocket K 线的作用

iTick WS 的 `kline@1...kline@10` 当前只写 Redis：

```text
itick:v1:kline:{category}:{market}:{symbol}:{interval}
```

它不会直接写 MongoDB，也不会触发 `DerivedAggregator`。作用是为 App 提供正在形成中的、比本地每分钟派生更实时的 K 线。

App `SubscribeStream` 每 5 秒从 Redis `MGET`，当前服务端会重复推送命中的缓存数据，即使内容没有变化；去重或覆盖展示由客户端处理。

当前 iTick WS Kline 与本地 `DerivedAggregator` 会写同一个 Redis Key。写入通过 Lua 原子比较来源优先级和 revision：活跃 iTick WS 优先于 derived，并刷新版本元数据 TTL；TTL 使用系统 `ITICK_CONFIG.wsKlineStaleSeconds`（默认 30 秒），超时后 derived 可以接管，WS 恢复后重新以高优先级覆盖。

## 9. Redis 行情缓存

当前实际 Key：

| 数据 | Key | TTL |
| --- | --- | --- |
| Quote | `itick:v1:quote:{category}:{market}:{symbol}` | 30 分钟 |
| Tick | `itick:v1:tick:{category}:{market}:{symbol}` | 30 分钟 |
| Depth | `itick:v1:depth:{category}:{market}:{symbol}` | 5 分钟 |
| Kline | `itick:v1:kline:{category}:{market}:{symbol}:{interval}` | 24 小时 |

另外保存：

- `itick:v1:kline:building:{category}:{market}:{symbol}:{ts}`：未闭合 1m 桶，TTL 使用 `buildingBucketTtlMinutes`；
- `itick:v1:kline:baseline:{productKey}`：股票累计成交量基线；
- Kline 的 `:meta` Hash：来源优先级和 revision，TTL 使用 `wsKlineStaleSeconds`。

当前没有实现 Tick Redis Stream、dirty ZSet 或产品 refcount Hash。

## 10. 故障恢复与幂等性

已经具备：

- Redis active set 可从 MySQL 全量重建；
- WS 重连后根据内存 `desiredSubs` 恢复订阅；
- MongoDB 使用 upsert，可重复执行；
- 每 5 分钟任务使用 Redis 分布式锁；
- 可配置最近窗口 REST 校准可以修复短时间断线、漏 Tick、迟到 Tick；
- REST 修正后高周期重新从低周期完整计算。
- Tick 当前分钟桶和股票累计量基线可从 Redis 恢复；
- 派生任务区分高优先级实时/校准队列与低优先级历史队列；
- 高周期按产品和周期批量查询、批量写入。
- WS 在首次连接之后再次完成鉴权和恢复订阅时，按 category 立即触发 reconnect repair；同一 category 同时只运行一个恢复任务；
- reconnect repair 读取每个活跃产品 MongoDB 最新 `1m` 作为断点，通过单产品 `/kline` 的 `et` 参数从当前最后闭合分钟向后分页，直到碰到该断点，因此停机时间超过最近校准窗口也能补齐；
- 断点回补完成后再执行一次该 category 的批量最近窗口校准，矫正断线边界附近的临时 K 线；没有任何本地历史的产品不会在重连时意外启动全历史同步，仍由管理后台初始化。
- `GapRepairService` 按 `gapScanIntervalMinutes` 周期扫描活跃产品的 MongoDB `1m`；每个产品每轮最多读取 2000 根，并将分页游标保存到 `itick:v1:kline:gap_scan:{productId}`，多轮后覆盖完整历史；
- scanner 跳过最近校准窗口，并使用市场时区、Session、周末和 Holiday 过滤非交易分钟；非 crypto 市场没有日历或 Session 时保守跳过，避免把夜间休市误判为缺口；
- 缺口任务持久化到 Redis ZSet `itick:v1:kline:repair:queue` 和 Hash `itick:v1:kline:repair:jobs`，多实例通过 Lua 原子领取；
- repair worker 每批最多处理 `repairBatchSize` 个任务，使用单产品 `/kline?et=` 定点向后分页，只写缺口范围，并通过历史低优先级派生队列传播到高周期；
- repair 失败按分钟级指数退避，最大 12 小时；成功任务保存 30 天完成标记，避免休市边界或供应商确实无数据时反复请求。

仍需完善：

- 历史同步断点续传和目标保留期；
- 多实例聚合 owner 切换期间的显式恢复；

## 11. 当前实现限制与下一步

按优先级建议：

1. **核对非股票 Tick volume 语义**：确认加密货币、外汇、指数和期货是单次量还是累计量。
2. **交易所精度**：把 exchange 传播到 K 线元数据或产品上下文，避免只能使用市场默认日历。
3. **配置热更新**：当前 `ITICK_CONFIG` 在服务启动时读取，修改后需要重启服务；任务中心的执行频率仍由外部调度配置控制。
4. **可观测性**：补充 repair queue 长度、最老任务等待时间、扫描游标进度和永久失败告警。

## 12. 关键代码位置

| 职责 | 文件 |
| --- | --- |
| App MongoDB Kline 查询 | `internal/logic/getklinelogic.go` |
| 管理后台历史同步入口 | `internal/logic/syncproductklinehistorylogic.go` |
| 每 5 分钟任务入口 | `internal/logic/syncklineslogic.go` |
| 历史分页与最近窗口 REST | `internal/market/kline/sync_worker.go` |
| Tick → 1m | `internal/market/kline/tick_aggregator.go` |
| 高周期派生 | `internal/market/kline/derived_aggregator.go` |
| 历史缺口扫描与修复队列 | `internal/market/kline/gap_repair.go` |
| MongoDB 批量写入 | `internal/pkg/klinewriter/batch_writer.go` |
| MongoDB Kline Model | `models/coinklinemodel.go` |
| 周期映射 | `internal/pkg/utils/kline_intervals.go` |
| WS 订阅与消息处理 | `internal/market/client/itick_ws_client.go` |
| 活跃产品和订阅构建 | `internal/market/client/itick_manager.go` |
| Redis 行情缓存 | `internal/market/cache/market_data_cache.go` |
| 任务消息消费 | `internal/tasks/subscriber.go` |
