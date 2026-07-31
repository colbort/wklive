# 公开期权链与盘口上线准备记录

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

- [ ] CHAIN-001～CHAIN-006 全部通过，保存请求、响应、SQL 快照和时延报告。
- [ ] 500合约链、100档盘口及目标并发达到批准 SLA，无 N+1 和数据库热点。
- [ ] Call/Put 缺腿、OI 不平衡、陈旧行情和无盘口的页面/客服文案已演练。
- [ ] 同租户/跨租户权限、限流、缓存键和缓存失效已验证。
- [ ] `generated_at`、三类行情时间、持仓水位、来源和 OI 算法均在客户端可见或可诊断。
- [ ] OPT-A029 指标、通知对象、值班手册和关闭证据路径已接入。
- [ ] 已明确首版为一致性快照 REST，不对外声称具备增量 L2 sequence/replay。

## 5. 签字

| 角色 | 姓名 | 结论 | 时间 | 证据/工单 |
| --- | --- | --- | --- | --- |
| 产品 |  |  |  |  |
| 行情运营 |  |  |  |  |
| Option 技术 |  |  |  |  |
| 风控 |  |  |  |  |
| 合规/法务 |  |  |  |  |
