# 公开期权链与盘口上线准备记录

OPTION_PUBLIC_MARKET_READINESS_STATUS: DRAFT

> 每个租户/市场首次公开或重大口径变更均填写。占位项未批准时只能内部联调。

## 1. 范围与责任人

| 项目 | 内容 |
| --- | --- |
| 租户/法律实体 | `[TENANT / ENTITY]` |
| 标的与结算币 | `[UNDERLYING] / [SETTLE_COIN]` |
| 适用到期 | `[EXPIRIES]` |
| 产品负责人 | `[NAME]` |
| Option 技术负责人 | `[NAME]` |
| 行情运营负责人 | `[NAME]` |
| 风控/合规复核人 | `[NAME]` |
| 计划公开时间（含时区） | `[TIME / IANA_TZ]` |

## 2. 已实现且不得改写的事实口径

- 期权链：精确 `underlying_symbol + expire_time`，公开 `TRADING`，可内部查询 `PAUSED`。
- 盘口：Option `PENDING/PART_FILLED` 正余量限价委托，同价聚合；不含用户/订单身份。
- 24h：服务端 Unix 秒滚动窗口内 Option 成交数量、成交额和笔数，每笔一次。
- OI：已落仓 `HOLDING` 多空单边值；不平衡显示较大侧并令 `oi_balanced=false`。
- Market 仅提供标的价、标记价、IV/Greeks 及各自时间，不提供 Option 委托深度。

## 3. 待业务批准参数

仓库已经确定的硬边界如下，不需要运营重新选择：单次链最多500个合约；盘口请求最多100档；
只公开 `TRADING`（内部可显式查 `PAUSED`）；当前接口为一致性快照 REST；24h 窗口为服务端
`[generated_at-86400,generated_at]`。在缓存策略正式批准前，仓库建议网关使用 `no-store`，
该建议不等于生产配置已经生效。

| 参数 | 批准值 | 依据/证据 | 审批人 |
| --- | --- | --- | --- |
| 对外字段及小数展示 | `[VALUE]` | `[REF]` | `[NAME]` |
| 最大盘口档数（≤100） | `[VALUE]` | `[REF]` | `[NAME]` |
| 链/盘口请求限流 | `[VALUE]` | `[LOAD TEST]` | `[NAME]` |
| 服务端 SLA/P95/P99 | `[VALUE]` | `[LOAD TEST]` | `[NAME]` |
| 客户端刷新间隔 | `[VALUE]` | `[REF]` | `[NAME]` |
| CDN/网关缓存 TTL | `[VALUE/NO CACHE]` | `[REF]` | `[NAME]` |
| 超时/陈旧降级阈值 | `[VALUE]` | `[REF]` | `[NAME]` |
| 历史/暂停合约可见规则 | `[VALUE]` | `[PRODUCT RULE]` | `[NAME]` |
| 用户协议/数据免责声明 | `[LINK/VERSION]` | `[LEGAL APPROVAL]` | `[NAME]` |

## 4. 上线验收

- [x] CHAIN-001～CHAIN-006 仓库级范围通过；证据为
  `docs/option-p2-006-public-market-repository-acceptance.md`。
- [x] 500合约、501整体拒绝、100档、101档拒绝、16路并发和批量查询结构已验证。
- [ ] 使用生产等价拓扑证明目标并发达到批准 SLA，无数据库热点，并附 P95/P99 报告。
- [ ] Call/Put 缺腿、OI 不平衡、陈旧行情和无盘口的页面/客服文案已演练。
- [x] 仓库层租户过滤及跨租户行情/成交/持仓/订单污染隔离已验证。
- [ ] 网关身份注入、限流、缓存键和缓存失效已在生产等价链路验证。
- [x] `generated_at`、三类行情时间、持仓水位、来源和 OI 算法已进入 RPC/REST DTO。
- [ ] OPT-A029 指标、通知对象、值班手册和关闭证据路径已接入。
- [x] 仓库文档已明确首版为一致性快照 REST，不对外声称具备增量 L2 sequence/replay。

### 4.1 已填仓库证据（2026-08-02）

| 项目 | 结果 | 证据 |
| --- | --- | --- |
| 公开索引迁移双执行 | PASS | 正式 Option/Asset RPC 脚本 |
| 500/501链边界 | PASS | 250个配对行权价；第501条整体拒绝 |
| 100×2/101档边界 | PASS | 201笔合法静态委托；101档请求拒绝 |
| 一致性快照 | PASS | 24读取者×6次，前/后态完整，撕裂0 |
| 仓库租户隔离 | PASS | 四类跨租户事实污染不进入主租户响应 |
| 24h/OI质量 | PASS | 数量9/额90/3笔；OI 8/7不平衡、4/4平衡 |
| 生产网关/CDN/限流/SLA/通知 | PENDING | 必须由预生产/生产责任人补证据并签字 |

## 5. 签字

| 角色 | 姓名 | 结论 | 时间 | 证据/工单 |
| --- | --- | --- | --- | --- |
| 产品 |  |  |  |  |
| 行情运营 |  |  |  |  |
| Option 技术 |  |  |  |  |
| 风控 |  |  |  |  |
| 合规/法务 |  |  |  |  |
