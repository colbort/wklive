# Price Engine 价格公式配置说明

## 1. 文档目的

本文说明 iTick Price Engine 中价格公式的配置方式、字段含义、公式依赖关系和生产使用要求。

当前 Price Engine 支持生成以下权威快照：

- `MARK`：标记价格；
- `INDEX`：指数价格；
- `FUNDING`：资金费率；
- `DELIVERY`：交割价格。

价格公式输出统一使用 `price-engine` 作为权威来源。原始行情通常来自 `market-ws` 或 `market-rest` 的 `FINAL_QUOTE` 快照。

## 2. 公式依赖关系

永续合约推荐使用以下计算链路：

```text
原始行情 FINAL_QUOTE
  ├─ INDEX 指数价格
  └─ MARK  标记价格
       │
       └────────┐
                ▼
          FUNDING 资金费率
```

各公式的主要用途：

| 快照类型 | 用途 |
| --- | --- |
| `INDEX` | 表示多市场公允价格，作为标记价格和资金费率的基准 |
| `MARK` | 用于未实现盈亏、风险率、强平判断和资金费率计算 |
| `FUNDING` | 用于永续合约多空双方的周期性资金费结算 |
| `DELIVERY` | 用于交割合约到期结算 |

## 3. 通用字段说明

| 字段 | 是否必填 | 说明 |
| --- | --- | --- |
| `formula_no` | 是 | 公式唯一编号，建议包含交易对、快照类型和版本 |
| `formula_version` | 是 | 业务公式版本，例如 `v1` |
| `authority` | 是 | 输出权威来源，价格公式固定为 `price-engine` |
| `snapshot_kind` | 是 | 输出类型：`MARK`、`INDEX`、`FUNDING`、`DELIVERY` |
| `category_code` | 建议填写 | 产品分类，例如数字货币使用 `crypto` |
| `market` | 建议填写 | 市场代码，例如 `BA` |
| `symbol` | 是 | 产品或交易对，例如 `BTCUSDT` |
| `algorithm` | 是 | 计算算法：加权平均、中位数或溢价率 |
| `components` | 是 | 公式依赖的原始行情或上游 Price Engine 快照 |
| `max_lookback_ms` | 是 | 允许使用的输入快照最大回看时间，单位毫秒 |
| `max_deviation_bps` | 是 | 输入价格相对中位数的最大偏离，单位 BPS |
| `interval_ms` | 是 | 公式计算周期，单位毫秒 |
| `activate` | 是 | 创建完成后是否立即激活 |

### 3.1 权威来源与快照类型约束

| 权威来源 | 允许的快照类型 |
| --- | --- |
| `market-ws` | `FINAL_QUOTE` |
| `market-rest` | `FINAL_QUOTE` |
| `price-engine` | `MARK`、`INDEX`、`FUNDING`、`DELIVERY` |

创建公式时，输出权威来源固定为 `price-engine`。公式成分可以引用原始行情，也可以引用其他 Price Engine 生成的上游快照。

### 3.2 时间参数

示例：

```text
max_lookback_ms = 30000
interval_ms     = 1000
```

表示公式每秒计算一次，并且只使用目标计算时点之前最近 30 秒内的输入快照。

如果对应窗口内不存在完全匹配的输入，计算会失败。输入匹配条件包括：

- 权威来源；
- 快照类型；
- 分类编码；
- 市场；
- 交易对；
- 来源时间。

## 4. 当前 BTCUSDT 测试配置

### 4.1 MARK 标记价格

```json
{
  "formula_no": "BTCUSDT-MARK-v1",
  "formula_version": "v1",
  "authority": "price-engine",
  "snapshot_kind": "MARK",
  "category_code": "crypto",
  "market": "BA",
  "symbol": "BTCUSDT",
  "algorithm": "WEIGHTED_MEAN",
  "components": [
    {
      "authority": "market-ws",
      "kind": "FINAL_QUOTE",
      "category_code": "crypto",
      "market": "BA",
      "symbol": "BTCUSDT",
      "weight": "1"
    }
  ],
  "max_lookback_ms": 30000,
  "max_deviation_bps": 0,
  "interval_ms": 1000
}
```

当前只有一个输入成分，权重为 `1`，因此标记价格等于该 `FINAL_QUOTE` 价格。
该配置只能用于技术联调，不能用于永续合约真实资金费、强平或交割风险计算。

生产 MARK 应创建新版本并使用 `INDEX_BASIS`：

```json
{
  "formula_no": "BTCUSDT-MARK-v2",
  "formula_version": "v2",
  "authority": "price-engine",
  "snapshot_kind": "MARK",
  "category_code": "crypto",
  "market": "BA",
  "symbol": "BTCUSDT",
  "algorithm": "INDEX_BASIS",
  "components": [
    {
      "authority": "price-engine",
      "kind": "INDEX",
      "category_code": "crypto",
      "market": "BA",
      "symbol": "BTCUSDT",
      "weight": "1"
    },
    {
      "authority": "market-ws",
      "kind": "FINAL_QUOTE",
      "category_code": "crypto",
      "market": "BA",
      "symbol": "BTCUSDT",
      "weight": "0.2"
    },
    {
      "authority": "price-engine",
      "kind": "MARK",
      "category_code": "crypto",
      "market": "BA",
      "symbol": "BTCUSDT",
      "weight": "0.8"
    }
  ],
  "max_lookback_ms": 30000,
  "max_deviation_bps": 200,
  "min_input_count": 3,
  "interval_ms": 1000
}
```

计算规则：

```text
raw_basis     = (PERPETUAL_QUOTE - INDEX) / INDEX
applied_basis = clamp(raw_basis, -max_deviation_bps/10000, +max_deviation_bps/10000)
raw_mark      = INDEX × (1 + applied_basis)
MARK          = weighted_mean(raw_mark, PREVIOUS_MARK)
```

组件顺序属于公式语义：第一个必须是同一 Price Engine 输出的 `INDEX`，第二个必须是
独立行情权威的 `FINAL_QUOTE`，可选第三个是同一输出维度的上一时点 `MARK`。第三个
组件存在时，第二、三个组件的权重分别控制当前基差价和历史 MARK 的平滑比例；引擎
强制按 `target-1ms` 查找上一 MARK，禁止读取本轮输出形成自引用。初次启用平滑版本前，
必须已有旧 MARK 历史用于启动。

系统拒绝反向顺序、缺失必需输入、跨 Symbol、自引用行情源以及未配置基差上限的公式。
每次计算的原始基差、限幅后基差和未平滑 MARK 均进入快照 `raw_payload`，用于风险事件重放。

### 4.2 INDEX 指数价格

```json
{
  "formula_no": "BTCUSDT-INDEX-v1",
  "formula_version": "v1",
  "authority": "price-engine",
  "snapshot_kind": "INDEX",
  "category_code": "crypto",
  "market": "BA",
  "symbol": "BTCUSDT",
  "algorithm": "MEDIAN",
  "components": [
    {
      "authority": "market-ws",
      "kind": "FINAL_QUOTE",
      "category_code": "crypto",
      "market": "BA",
      "symbol": "BTCUSDT",
      "weight": "1"
    }
  ],
  "max_lookback_ms": 30000,
  "max_deviation_bps": 0,
  "interval_ms": 1000
}
```

当前只有一个输入成分，因此中位数仍等于该 `FINAL_QUOTE` 价格。

该单源 INDEX 仅保留用于已有技术测试。新建 INDEX 版本必须满足：

- `MEDIAN` 或 `WEIGHTED_MEAN`；
- 至少三个组件且 `min_input_count >= 3`；
- 每个组件均为外部权威 `FINAL_QUOTE`；
- Authority、快照类型、分类、市场和 Symbol 组成的来源身份不得重复；
- 运行时相同 Snapshot ID 只计算一次，不能通过重复组件凑足数量。

### 4.3 历史确定性回放

新生成的 Price Engine 审计会在 `raw_payload` 中固化 `output_price`、完整输入、
采用/剔除输入及算法参数。工具接受单个 JSON、JSON 数组、JSONL 以及多个文件。
导出单条 `raw_payload` 为 JSON 文件后执行：

```bash
cd services/market
go run ./cmd/price-replay /path/to/evaluation-audit.json
```

生产历史窗口应导出为 JSON 数组或 JSONL，并显式校验公式计算周期。例如每秒公式：

```bash
go run ./cmd/price-replay \
  --interval-ms 1000 \
  --json \
  /path/to/delivery-window.jsonl > /path/to/replay-report.json
```

成功输出会汇总审计记录数、公式版本数、目标时点范围、价格范围、最少有效输入数和
被剔除输入总数。开启 `--interval-ms` 后，每个 `formula_no + formula_version` 必须满足：

- `target_time` 对计算周期对齐；
- 不存在重复目标时点；
- 相邻目标时点严格连续，不允许窗口缺口。

单条成功输出：

```text
replay verified: price=...
```

工具只读取不可变审计，不读取当前公式表或实时行情；记录输出被修改、采用输入出现重复
Snapshot ID、有效输入低于 `min_input_count`、完整输入与采用/剔除摘要不一致、算法参数
缺失、重算结果不一致、窗口重复或断档时都会返回非零状态。JSON 报告可直接作为生产
变更单和交割窗口验收附件。

### 4.4 FUNDING 资金费率

```json
{
  "formula_no": "BTCUSDT-FUNDING-v1",
  "formula_version": "v1",
  "authority": "price-engine",
  "snapshot_kind": "FUNDING",
  "category_code": "crypto",
  "market": "BA",
  "symbol": "BTCUSDT",
  "algorithm": "PREMIUM_RATE",
  "components": [
    {
      "authority": "price-engine",
      "kind": "MARK",
      "category_code": "crypto",
      "market": "BA",
      "symbol": "BTCUSDT",
      "weight": "1"
    },
    {
      "authority": "price-engine",
      "kind": "INDEX",
      "category_code": "crypto",
      "market": "BA",
      "symbol": "BTCUSDT",
      "weight": "1"
    }
  ],
  "max_lookback_ms": 30000,
  "max_deviation_bps": 0,
  "interval_ms": 1000
}
```

溢价率计算公式：

```text
FUNDING = (MARK - INDEX) / INDEX
```

示例：

```text
MARK  = 101000
INDEX = 100000

FUNDING = (101000 - 100000) / 100000
        = 0.01
        = 1%
```

### 4.5 DELIVERY 交割价

生产配置应使用多个相互独立的权威来源，不能复用当前单一 BA 行情测试配置：

```json
{
  "formula_no": "BTCUSDT-DELIVERY-v1",
  "formula_version": "v1",
  "authority": "price-engine",
  "snapshot_kind": "DELIVERY",
  "category_code": "crypto",
  "market": "BA",
  "symbol": "BTCUSDT",
  "algorithm": "MEDIAN",
  "components": [
    {
      "authority": "market-ws",
      "kind": "FINAL_QUOTE",
      "category_code": "crypto",
      "market": "SOURCE_A",
      "symbol": "BTCUSDT",
      "weight": "1"
    },
    {
      "authority": "market-ws",
      "kind": "FINAL_QUOTE",
      "category_code": "crypto",
      "market": "SOURCE_B",
      "symbol": "BTCUSDT",
      "weight": "1"
    },
    {
      "authority": "market-ws",
      "kind": "FINAL_QUOTE",
      "category_code": "crypto",
      "market": "SOURCE_C",
      "symbol": "BTCUSDT",
      "weight": "1"
    }
  ],
  "max_lookback_ms": 30000,
  "max_deviation_bps": 200,
  "min_input_count": 3,
  "interval_ms": 1000
}
```

`min_input_count` 表示偏差过滤后仍必须保留的有效输入数量。`DELIVERY`
无论新旧配置都强制不低于 `3`；少于 3 个有效来源时，本轮返回
`price engine input unavailable`，不会生成或发布交割快照。创建页面建议显式填写
`3`，并保证三个组件来自相互独立的行情来源。

每个输出快照的 `raw_payload` 固化目标时点、公式版本、算法、完整输入、
采用输入和被剔除输入。快照 ID 由这些事实与计算结果确定性生成，因此同一
输入集合可以重放；输入不足时计算失败，不允许以无审计的最新价兜底。

## 5. 当前配置评价

当前配置可以用于验证以下技术链路：

- 原始行情永久归档；
- MARK、INDEX 和 FUNDING 公式调度；
- Price Engine 权威快照生成；
- Snapshot Outbox 发布；
- Trade 服务读取资金费快照。

但当前 MARK 和 INDEX 都只使用同一个 `market-ws / FINAL_QUOTE` 输入。单成分情况下，加权平均与中位数结果相同，因此：

```text
MARK ≈ INDEX
FUNDING ≈ 0
```

该配置适合功能测试，不适合直接作为生产资金费率来源。

## 6. 生产环境建议

### 6.1 INDEX 指数价格

指数价格应使用多个独立市场来源：

```text
INDEX
  ├─ Binance BTCUSDT
  ├─ OKX BTC-USDT
  └─ Bybit BTCUSDT
```

建议：

- 至少使用三个相互独立的行情来源；
- 使用中位数或经过归一化的加权平均；
- 配置合理的 `max_deviation_bps` 剔除异常来源；
- 明确不同来源的交易对映射和报价币种换算规则；
- 行情源不足时停止生成权威指数价，不使用单一异常来源兜底。

### 6.2 MARK 标记价格

生产标记价格不应简单等于单一市场最新成交价。推荐基于指数价和永续合约溢价构建，并增加：

- 基差或溢价平滑；
- 价格变化上限；
- 异常输入剔除；
- 公式版本和输入快照审计；
- 历史确定性回放。

### 6.3 FUNDING 资金费率

资金费率应至少具备：

- MARK 与 INDEX 的溢价计算；
- 上限和下限；
- 平滑或时间加权；
- 固定结算周期；
- 资金费批次资金守恒；
- 用户余额不足处理；
- Asset 流水与 Trade 结算明细对账。

## 7. 创建和激活顺序

建议按依赖顺序创建并激活：

1. 确认 `FINAL_QUOTE` 持续产生；
2. 创建并激活 INDEX；
3. 确认 INDEX 快照持续生成；
4. 创建并激活 MARK；
5. 确认 MARK 快照持续生成；
6. 创建并激活 FUNDING；
7. 确认 FUNDING 快照持续生成；
8. 再启用 Trade 的生产资金费结算任务。

新公式刚激活时，如果上游快照尚未生成，可能短暂出现输入不存在。持续报错则需要核对成分的来源、类型、分类、市场、交易对和最大回看时间。

## 8. 检查 SQL

检查指定产品最近的权威快照：

```sql
SELECT
  authority,
  snapshot_kind,
  category_code,
  market,
  symbol,
  price,
  FROM_UNIXTIME(source_timestamp / 1000) AS source_time
FROM t_market_authoritative_snapshot
WHERE category_code = 'crypto'
  AND market = 'BA'
  AND symbol = 'BTCUSDT'
ORDER BY source_timestamp DESC
LIMIT 30;
```

检查已激活公式：

```sql
SELECT
  id,
  formula_no,
  authority,
  snapshot_kind,
  category_code,
  market,
  symbol,
  components,
  max_lookback_ms,
  max_deviation_bps,
  interval_ms,
  last_target_time
FROM t_market_price_formula
WHERE status = 1
ORDER BY id;
```

## 9. 上线门禁

满足以下条件前，不应启用生产资金费或自动交割：

- INDEX 使用可靠的多市场输入；
- MARK 公式通过历史回放；
- FUNDING 费率上下限和平滑策略明确；
- 输入快照缺失和异常偏离具备告警；
- 公式版本切换和撤销经过验证；
- Snapshot Outbox 无持续积压；
- Funding/Delivery 与 Asset 流水完成自动对账；
- 重复执行、超时和进程崩溃故障注入通过。
