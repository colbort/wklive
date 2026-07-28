# Price Engine 价格公式配置说明

## 1. 文档目的

本文说明 iTick Price Engine 中价格公式的配置方式、字段含义、公式依赖关系和生产使用要求。

当前 Price Engine 支持生成以下权威快照：

- `MARK`：标记价格；
- `INDEX`：指数价格；
- `FUNDING`：资金费率；
- `DELIVERY`：交割价格。

价格公式输出统一使用 `price-engine` 作为权威来源。原始行情通常来自 `itick-ws` 或 `itick-rest` 的 `FINAL_QUOTE` 快照。

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
| `itick-ws` | `FINAL_QUOTE` |
| `itick-rest` | `FINAL_QUOTE` |
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
      "authority": "itick-ws",
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
      "authority": "itick-ws",
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

### 4.3 FUNDING 资金费率

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

## 5. 当前配置评价

当前配置可以用于验证以下技术链路：

- 原始行情永久归档；
- MARK、INDEX 和 FUNDING 公式调度；
- Price Engine 权威快照生成；
- Snapshot Outbox 发布；
- Trade 服务读取资金费快照。

但当前 MARK 和 INDEX 都只使用同一个 `itick-ws / FINAL_QUOTE` 输入。单成分情况下，加权平均与中位数结果相同，因此：

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
FROM t_itick_authoritative_snapshot
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
FROM t_itick_price_formula
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
