# 价格公式严格生产门禁验收报告

## 1. 目标

防止仅凭输出类型和算法编号将错误的 INDEX、MARK、FUNDING 或 DELIVERY 公式判定为
生产候选。所有查询均由 `deploy/dbinit/models` 执行，本次验收只读，不修改数据库。

## 2. 正向配置

- 现货 Authority：`binance-public`、`okx-public`、`bybit-public`；
- 来源市场：`BINANCE`、`OKX`、`BYBIT`；
- INDEX：MEDIAN、权重 `1,1,1`、200 BPS、版本
  `production-index-v1`；
- MARK：永续来源 `binance-futures-public/BINANCE_PERP`、200 BPS、当前/前值
  权重 `1:4`、版本 `production-mark-v2`；
- FUNDING：版本 `production-funding-v1`；
- DELIVERY：MEDIAN、权重 `1,1,1`、200 BPS、30 秒窗口、版本
  `delivery-v1`；
- 四类公式执行周期：1,000 ms。

加强后的完整门禁结果为 `60 PASS / 12 FAIL`。INDEX、MARK、FUNDING、DELIVERY
公式匹配数均为 1，三个来源和四类输出均新鲜。12 个失败项仍全部属于数据许可、
值班/升级、资金审批与注资、自动强平窗口、正式灾备和最终交割产品启用。

## 3. 负向验证

### 3.1 MARK 前值权重漂移

只把声明的 MARK 前值权重从 4 改为 5：

```text
READINESS_DB_INDEX_FORMULAS=1
READINESS_DB_MARK_FORMULAS=0
READINESS_DB_FUNDING_FORMULAS=1
READINESS_DB_DELIVERY_FORMULAS=1
```

结果符合预期：只有 MARK 不能通过。

### 3.2 来源市场映射漂移

只把 OKX 市场从 `OKX` 改为 `WRONG_OKX`：

```text
READINESS_DB_INDEX_FORMULAS=0
READINESS_DB_MARK_FORMULAS=1
READINESS_DB_FUNDING_FORMULAS=1
READINESS_DB_DELIVERY_FORMULAS=0
```

结果符合预期：使用三现货来源的 INDEX 和 DELIVERY 均不能通过，MARK 与 FUNDING
自身的结构匹配不受错误声明污染。

## 4. 自动化与运行检查

- `deploy/dbinit go test ./...`：PASS；
- `deploy/contract-readiness.sh` shell 语法：PASS；
- 完整 Deploy 门禁：`60 PASS / 12 FAIL`，按预期退出 1；
- 最近十分钟 iTick 无 Price Engine、外部来源、Outbox、panic 或 fatal 异常；
- `AutomaticLiquidation.Enabled=false`；
- `CrossMarginTrading.Enabled=false`。

## 5. 结论

价格公式生产门禁现在逐项验证来源身份与市场、算法、不可变版本、权重、偏差/基差、
回看窗口、执行周期和组件集合。参数漂移、来源市场漂移或错误平滑权重不能通过。
