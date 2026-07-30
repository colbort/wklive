# BTCUSDT 交割合约启用前技术验收报告

日期：2026-07-30  
环境：完整 Deploy 环境  
命令：`./deploy.sh contract-delivery-preflight`  
操作性质：只读；未启用产品、未写入交易事实、未上传对象。  
结论：`DELIVERY_TECHNICAL_PREFLIGHT=PASS`；生产启用仍不允许。

## 验收边界

本次验收只证明 `tenant=1 / symbol_id=7 / BTCUSDT-20260925` 在停用态具备完整技术
配置。工具固定输出 `DELIVERY_PRODUCTION_ENABLE_ALLOWED=false`，不会修改产品状态，
也不替代行情许可、生产审批、保险基金、强平发布和灾备门禁。

数据库查询全部位于 `deploy/dbinit/models/deliverypreflight.go`，命令层只负责读取
声明参数、调用 model 和输出结果。

## 运行结果

```text
DELIVERY_TECHNICAL_PREFLIGHT=PASS
DELIVERY_PRODUCTION_ENABLE_ALLOWED=false
DELIVERY_PRODUCT_COUNT=1
DELIVERY_CONFIGURED_PRODUCT_COUNT=1
DELIVERY_SAFE_DISABLED_PRODUCT_COUNT=1
DELIVERY_SYMBOL_ID=7
DELIVERY_PRODUCT_STATUS=2
DELIVERY_SERVER_TIME_MS=1785405833968
DELIVERY_OPEN_CUTOFF_TIME_MS=1790319600000
DELIVERY_MATCHING_STOP_TIME_MS=1790321400000
DELIVERY_TIME_MS=1790323200000
DELIVERY_ISOLATED_LEVERAGE_CONFIGS=1
DELIVERY_ISOLATED_LEVERAGE_DEFAULTS=1
DELIVERY_CROSS_LEVERAGE_CONFIGS=0
DELIVERY_ENABLED_RISK_TIERS=1
DELIVERY_VALID_RISK_TIERS=1
DELIVERY_BASE_RISK_TIERS=1
DELIVERY_RISK_COVERAGE_ENDS=1
DELIVERY_FORMULAS=1
DELIVERY_CONFORMING_FORMULAS=1
DELIVERY_FRESH_SNAPSHOTS=29
DELIVERY_LATEST_SNAPSHOT_TIME_MS=1785405818935
DELIVERY_ORDERS=0
DELIVERY_FILLS=0
DELIVERY_POSITIONS=0
DELIVERY_POSITION_HISTORY=0
DELIVERY_RESERVATIONS=0
DELIVERY_SETTLEMENT_INSTRUCTIONS=0
DELIVERY_BATCHES=0
DELIVERY_SETTLEMENTS=0
DELIVERY_HISTORICAL_FACTS=0
```

## 时间与产品安全姿态

- 检查时间：2026-07-30 18:03:53 HKT；
- 停止开仓：2026-09-25 15:00:00 HKT；
- 停止撮合：2026-09-25 15:30:00 HKT；
- 交割：2026-09-25 16:00:00 HKT；
- 产品状态：2（停用）；
- 开多/开空开关：停用；
- 平多/平空开关：启用；
- 全仓：不支持；
- 逐仓：支持；
- 逐仓杠杆配置、默认杠杆及风险档位覆盖：通过。

时间顺序满足：

```text
检查时点 < 停止开仓 < 停止撮合 <= 交割
```

## 行情与事实边界

- DELIVERY 公式：`delivery-v1`；
- 公式算法、窗口、偏差、输入数量和周期与生产声明一致；
- 检查窗口内有 29 条新鲜 DELIVERY 快照；
- 最新快照时间：2026-07-30 18:03:38 HKT；
- Order、Fill、Position、Position History、Reservation、Settlement Instruction、
  Delivery Batch、Delivery Settlement 均为 0。

## 生产结论

本项停用态技术验收完成。当前仍不得将 Symbol 7 改为启用；必须先让其余生产门禁通过，
取得独立产品启用发布单，再由操作、复核、审批职责分离账号执行。启用后还必须重新运行
完整 `contract-readiness`，要求零 FAIL。
