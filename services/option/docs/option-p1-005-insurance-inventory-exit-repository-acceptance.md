# P1-005 保险接管库存主动退出仓库验收报告

> 验收日期：2026-08-02  
> 结论：仓库级验收通过；生产启用未批准，默认保持关闭。

## 1. 验收范围

本报告只验收现金结算期权保险账户空头库存的受控主动退出：创建申请、独立复核、只减仓 IOC 执行、实时订单结果查询、资金与持仓闭环、数据库防篡改、管理权限和生产就绪门禁。

不在本次仓库验收范围内：实物交割库存、跨币种换汇、自动做市、自动 Delta 对冲、生产真实账户及通知链路。

## 2. 设计与实现核对

| 要求 | 仓库证据 | 结果 |
| --- | --- | --- |
| 仅保险账户、现金结算、活动空头仓 | 创建和执行两阶段重新校验合约、仓位、保险账户和可用数量 | 通过 |
| 四眼审批 | `requested_by != reviewed_by`，创建/批准/拒绝事件留痕 | 通过 |
| 同仓只允许一个活动申请 | `active_key` 状态派生字段及租户内唯一索引 | 通过 |
| 确定性且幂等执行 | `request_no` 派生 client_order_id；提交事务锁申请并关联唯一 `order_id` | 通过 |
| 订单语义固定 | `ORDER_SOURCE_ADMIN + BUY + CLOSE + REDUCE_ONLY + IOC` | 通过 |
| 失败可安全重试 | 执行失败保持 APPROVED，保存 `last_error_msg`；成功后切换 SUBMITTED | 通过 |
| 结果不漂移 | 申请只保存 `order_id`；列表实时读取订单状态、`filled_qty` 和 `unfilled_qty` | 通过 |
| 资金、仓位和保证金守恒 | 复用公开订单资金、撮合、持仓事件和 FIFO margin lot 链路 | 通过 |
| 管理入口和最小权限 | Admin API/UI 已接通 create/review/execute/list 四个独立权限 | 通过 |
| 生产默认关闭 | Option 运行时 `InsuranceInventoryExit.Enabled=false`；就绪门禁 `OPTION_INSURANCE_INVENTORY_EXIT_ENABLED=false`，开启时要求审批和 E2E 报告哈希 | 通过 |
| 运行时硬限额 | 启用时五项正数配置强制单次数量、最坏权利金、UTC日预算、标记价偏离和最小可成交深度；无配置或超限 fail closed | 通过 |

## 3. 数据库与生成物

- 增量迁移：`services/option/migrations/20260802_option_insurance_inventory_exit.sql`。
- 表、唯一键、状态约束和经济字段/删除保护触发器连续执行两次。
- 从 Option DDL 目录执行 `make gen-model` 成功，生成模型包含 `ActiveKey`，自定义分页模型保留。
- `make sync-proto` 成功；RPC 模型向后兼容追加订单状态、已成交量和未成交量字段。
- System 权限迁移：`services/system/migrations/20260802_option_insurance_inventory_exit_permissions.sql`。
- 默认运行配置将功能关闭并将五项限额置0；只有启用且五项全部为正、单次不大于日上限时写路径才可运行。

## 4. 真实 Asset RPC 场景

核心场景为退出2张、订单簿仅有1张对手深度，并对同一批准申请发起20路并发执行：

- 只创建1个 ADMIN IOC 订单和1个确定性客户端订单键；
- 订单成交1张、撤销剩余1张；
- 保险仓剩1张，剩余保证金40；保险钱包总额140、可用100、冻结40；对手方钱包120；
- 退出链路形成4条资产指令和4条对应 Asset 流水，另有1条独立预算补资流水；
- 创建、批准、提交3条操作事件不可修改、不可删除；
- 申请经济字段改写和删除均被数据库拒绝。

五项运行时硬限制和 UTC 日预算拒绝场景纳入后，保险退出本身的真实 RPC 证据不变。后续再纳入 P2-001 日历专项和20轮零资金指令的下单/halt竞态后，最新全套结果为：`instructions=9239`、`success=9234`、`canceled=5`、`reconciled=9234`，加权终态 `success + 2*canceled = 9244`。本次完整复跑中，主 P0 场景耗时45.465秒；5000空头指派196.312秒；实物容量106.186秒；现金到期容量55.655秒。

## 5. 工程验证

| 命令/检查 | 结果 |
| --- | --- |
| `make gen-model`（`services/option`） | 通过 |
| `make sync-proto`（`services/option`） | 通过 |
| `go test ./...`（`services/option`） | 通过 |
| `go test ./...`（`admin-api`） | 通过 |
| Node 20.20.2 `npm run type-check`（`admin-ui`） | 通过 |
| `sh -n` 验收与就绪脚本 | 通过 |
| repository production readiness | `READY` |

验收期间发现并修复：操作事件类型字段长度不足、测试夹具模型遗漏、资产流水证据关联键错误、同仓并发创建竞争、Admin API 四个路由仍为空逻辑，以及列表未返回关联订单成交进度。

## 6. 生产放行前置条件

仓库级通过不等于生产批准。以下资料缺一不可：

1. 为具体生产合约批准不宽于仓库硬上限的单合约/单标的数量和损失预算；仓库已强制全局单次/UTC日数量、单次权利金、价格偏离和最小盘口深度，但真实值仍需签署；
2. 保险资金预算、补资和余额不足升级流程；
3. 创建、复核、执行角色名单、替班和紧急撤权记录；
4. IOC 部分成交、连续无成交、行情失效和临近到期通知送达演练；
5. 真实预生产账户 E2E 报告、资金/订单/仓位/margin lot/告警证据哈希；
6. 独立风险、清算、财务和运营签署。

完成上述条件前必须同时保持 Option 运行时 `InsuranceInventoryExit.Enabled=false` 和就绪门禁 `OPTION_INSURANCE_INVENTORY_EXIT_ENABLED=false`。跨币种退出、实物库存和自动 Delta 对冲继续作为独立设计项，不得由本功能推断为已完成。
