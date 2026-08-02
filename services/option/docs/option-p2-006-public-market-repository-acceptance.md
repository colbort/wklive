# OPT-P2-006 公开期权链与盘口仓库级验收记录

- 验收日期：2026-08-02
- 范围：`ListOptionChain`、`GetOrderBook`、24h Option 成交统计、已落仓 OI、租户隔离、容量和一致性快照
- 结论：`CHAIN-001～CHAIN-006` 的仓库级范围通过；这不是生产行情发布批准。

## 1. 本轮审计发现和修复

1. 公开链和盘口原先使用数据库默认事务隔离级别。若生产数据库改为 `READ COMMITTED`，同一响应的
   合约、盘口、成交序号、行情和持仓可能来自不同快照。现在两个接口统一使用显式只读
   `REPEATABLE READ` 事务，响应一致性不再依赖服务器默认值。
2. 原盘口查询只用 `price>0` 间接识别限价委托。异常遗留的 `IOC/FOK/MARKET` 若处于活动状态且
   带正价格，可能进入公开深度。现在数据库查询同时强制 `order_type IN (LIMIT,POST_ONLY)`、
   `status IN (PENDING,PART_FILLED)`、正余量和 `combo_order_id=0`。
3. 正式 RPC 门禁此前依赖 `option.sql` 已有索引，但未连续安装公开行情索引迁移。现在
   `20260731_zo_option_public_market.sql` 已加入正式脚本并连续执行两次。

## 2. 验收矩阵

| 编号 | 仓库级证据 | 结果 |
| --- | --- | --- |
| CHAIN-001 | 同一精确到期构造完整一对、Call 缺 Put 和另一到期；返回2个行权价，缺腿为 `nil`，另一到期不进入 | PASS |
| CHAIN-002 | 同一到期混合 `PENDING/TRADING/PAUSED/EXPIRED`；默认只返回 `TRADING`，显式仅允许 `PAUSED`，请求 `PENDING` 被拒绝 | PASS |
| CHAIN-003 | 100档买、100档卖和同价多用户聚合；`LIMIT/POST_ONLY` 计入，`IOC/FOK/MARKET/FUNDING/CANCELED`、零余量和组合影子单排除 | PASS |
| CHAIN-004 | 对显式窗口 `[start,end]` 写入边界前、边界、边界后、结束和结束后成交；只统计3笔，数量9、成交额90 | PASS |
| CHAIN-005 | Call 多8/空7返回 OI=8 且 `oi_balanced=false`；Put 多4/空4返回 OI=4；已关闭持仓和另一租户999张不计入 | PASS |
| CHAIN-006 | 500合约形成250个完整行权价；第501条加入后整体拒绝且不截断。100档成功、101档请求拒绝；16路最大链+盘口并发读取通过 | PASS |

## 3. 一致性、隔离和容量证据

- 并发快照场景在一个写事务内交替“活动买档/无成交”和“撤单/成交序号10000”，24个读取者各读取
  6次，共144个公开盘口响应。每个响应只能看到完整的前态或后态，撕裂组合为0。
- 主租户与隔离租户使用相同标的、到期，并在行情、成交、持仓和订单事实中故意复用主租户
  `contract_id`。主租户响应没有读到另一租户的价格999、成交999、持仓999或盘口999；主租户直接
  查询另一租户真实合约 ID 返回未找到。
- 正式数据库汇总：
  `capacity_contracts=501 paired_strikes=250 overflow_single_leg=1 bid_levels=100 ask_levels=100`
  `resting_orders=201 cross_tenant_contracts=2 sparse_oi_imbalanced=1`。
- 最大合法请求在单次响应中只执行批量合约、行情、成交和持仓查询，不按合约循环查询；盘口在数据库
  按价格聚合。仓库测试记录单次及16路并发耗时，但生产 SLA 必须用批准的拓扑和峰值另行签署。

## 4. 正式门禁结果

执行：

```sh
cd services/option
bash acceptance/run-p0-asset-rpc-e2e.sh
```

结果：

- `TestP0AssetRPCEndToEnd`：PASS，`111.861s`，其中包含本专项真实 MySQL/Redis 场景。
- Option/Asset 总门禁：9240条指令，9235条成功并对账，5条合法取消，加权终态9245。
- 公开行情索引迁移连续执行两次，最终脚本输出 `P0 Option/Asset RPC acceptance passed`。
- `go test ./internal/logic/app ./models ./internal/logic/task`：PASS。

本轮没有新增或修改表字段/触发器，仅使用既有公开行情索引迁移，因此不需要重新执行
`make gen-model`。以后任何 P2-006 DDL 变更仍必须执行并记录该命令。

## 5. 生产前剩余验收

以下项目依赖生产等价基础设施或业务批准，仓库不能代签：

1. 网关身份与租户注入、公开/私有访问边界、限流、WAF、缓存键和 CDN TTL；批准前建议
   `no-store`，不得缓存跨租户响应。
2. 在批准并发、合约数量、订单数量及写入流量下运行持续压测，签署 P95/P99、错误率、数据库 CPU、
   慢查询、连接池和副本延迟；本轮16路不是生产容量承诺。
3. Prometheus/Alertmanager 实例实际加载 OPT-A029，验证缺腿和 OI 不平衡的持续窗口、恢复、IM/电话/
   值班案件路由和真实接收人。
4. 产品、行情运营、合规/法务批准公开字段、24h滚动口径、暂停/历史合约可见性、刷新频率、降级文案、
   数据许可和用户协议。
5. 如产品要求增量 WebSocket L2，必须另建撮合事件序号、快照恢复和重放协议；当前 REST
   `last_match_sequence` 只表示最后成交，不代表订单簿版本。

运营签署入口：`docs/templates/public-market-readiness.md`；故障处理入口：
`docs/option-operations-runbook.md` 第6.18节。
