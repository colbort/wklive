# OPT-P2-007 组合/价差单仓库级验收记录

- 验收日期：2026-08-02
- 范围：2～4腿同到期现金期权组合单、规范化幂等、毛额冻结、独立策略簿、整组原子撮合、跨腿资金屏障、运营下钻与整组强撤
- 结论：`COMBO-001～COMBO-011` 的仓库级范围通过；这是受限 V1 的技术验收，不是生产开放或“完整交易所复杂订单”批准。

## 1. 产品边界与合理性结论

当前实现满足一般交易所复杂订单的核心语义：多个不同期权系列作为一个父订单提交，腿比例约分，
按组合净价和腿比例整体执行，任何一腿失败时不得留下残腿。首版主动收窄为：

- 2～4条不同的现金结算、同标的、同到期、同币种、同单位和乘数的开仓腿；比例1～8且最大公约数为1；
- 至少一买一卖，仅 `LIMIT/FOK`；FOK 只与最优单个 maker 完整成交；
- 独立策略簿，不与普通单腿簿 legging，不含股票腿、跨到期腿、拍卖、MARKET/IOC/POST_ONLY；
- 买方权利金、最高费率和卖方逐腿逐仓保证金毛额冻结，不使用组合保证金抵扣；
- 一次匹配的全部腿成交、序号、margin lot、Asset 指令、outbox 和父/子进度在一个 MySQL 事务提交。

因此，作为受限 V1，设计是合理且资金上保守的；但不能据此宣称支持多 maker 扫簿、复杂订单拍卖、
单腿簿隐含撮合、跨到期/股票腿或组合保证金等完整交易所能力。

## 2. 本轮审计发现与修复

1. 生成模型的组合幂等查询使用 go-zero 缓存。并发首个查询写入短暂负缓存后，另一请求即使已提交父单仍可能
   读到不存在。现在新增 `FindOneByTenantIdUserIdClientComboIdNoCache`，建单前、唯一键冲突后和事务错误后均
   用数据库权威查询；MySQL 1213/1205 只做最多5次的10/20/40/80毫秒有界退避。
2. 全腿冻结完成后原逻辑直接激活父单，没有重新检查冻结期间变化的 kill switch、合约/日历/halt、行情
   新鲜度、价格带及卖方当前标的价保证金。现在激活事务在锁内全部重检；失败时整组进入撤销并按原业务号释放，
   原因分别保留为 `COMBO_*_AFTER_FUNDING`。
3. 组合创建与撮合的用户、合约锁顺序存在死锁受害者风险。现在统一先锁钱包级用户控制，再按规范化合约 ID
   加锁；建单仍保留有界死锁/锁超时重试，不掩盖其他数据库错误。
4. 正式验收最初只加载 canonical schema，没有执行现有 `20260731_zp_option_combo_order.sql` 的4个不可变/
   禁止删除触发器。将迁移加入连续双执行后，发现预冻结撤单由 `FUNDING` 直接写 `CANCELED`，与数据库要求
   `FUNDING → CANCELING → CANCELED` 不一致。所有用户、控制和撮合撤单现统一调用父单状态迁移函数；无需
   Asset 释放时也在同一事务依次经过两个合法状态，需要释放时保持 `CANCELING` 直至全腿终态。
5. 正式脚本原先没有机器可读的组合结果。现在输出 `combo_acceptance=`，覆盖幂等对象基数、冻结响应丢失、
   kill/保证金漂移、FOK/STP、后台强撤、成交组、比例数量、outbox、持仓和资金指令唯一性。

## 3. 验收矩阵

| 编号 | 仓库级证据 | 结果 |
| --- | --- | --- |
| COMBO-001 | Go 校验覆盖2:2未约分、重复合约、5腿、比例9、跨标的/到期/币种/乘数、平仓/MMP腿、净价和腿 tick；MySQL CHECK 与4个不可变/禁止删除触发器连续安装两次 | PASS |
| COMBO-002 | 50路同一 `clientComboId` 并发，只生成1父单、2腿、2影子单、2冻结；全部返回同一父单；同键不同内容拒绝；无缓存权威重放和有界死锁重试生效 | PASS |
| COMBO-003 | 真实 Asset `FreezeAsset` 已提交后注入一次 `Unavailable`；父单保持 `FUNDING`，原指令重试后激活；整组撤销后钱包1000/1000/0，无重复冻结或泄漏 | PASS |
| COMBO-004 | 最优单个 maker 对2份FOK不足，父单以 `FOK_NOT_FILLED` 0腿撤销；真实足量1:2反向策略最终只产生2腿、总成交量3 | PASS |
| COMBO-005 | MySQL 临时触发器令第二腿成交写入失败；成交、outbox、资产指令及撮合序号全部为0；移除故障并重试后一次完整成交 | PASS |
| COMBO-006 | 同用户不同账户反向策略以 `SELF_TRADE_PREVENTED` 0腿撤销；影子单不进入普通盘口；单腿入口和锁后路径另有三层失败关闭测试 | PASS |
| COMBO-007 | 建单后激活前打开 kill switch，以及把标的价从100跳至1000导致卖方保证金不足；两组均全腿撤销、资金全额释放并记录精确原因 | PASS |
| COMBO-008 | 组合成交后对一个买方扣冻结注入“已提交但响应丢失”；2个 outbox 均被阻断且持仓为0；原指令恢复后2个 outbox成功、4个持仓落地、7条资金指令全部成功且唯一 | PASS |
| COMBO-009 | 真实后台上下文验证跨租户详情和强撤拒绝、空原因拒绝；授权详情返回2腿/2影子/2资金指令；整组强撤原因落库并全额释放 | PASS |
| COMBO-010 | 隔离 MySQL 异常租户精确返回超龄2、人工态1、最早异常时间800、结构异常3、缺腿成交组1；健康租户五项为0；阈值只允许10～300秒 | PASS（仓库） |
| COMBO-011 | 完整2～4腿且全部买方扣款成功才开放持仓；缺腿、重复腿号、任一扣款未完成均失败关闭；盘口/matcher隔离 Counter、屏障 Gauge、内部9105 metrics和OPT-A030规则有静态/单测证据 | PASS（仓库） |

## 4. 正式真实 RPC 与数据库证据

执行：

```sh
cd services/option
bash acceptance/run-p0-asset-rpc-e2e.sh
```

正式结果：

- `TestP0AssetRPCEndToEnd`：PASS，`116.577s`，包含 COMBO-002～COMBO-009 的真实 MySQL、Redis、
  Asset gRPC、并发和故障注入；
- `20260731_zp_option_combo_order.sql` 在隔离 MySQL 8.4 连续执行两次，状态机触发器实际参与业务验收；
- 最终组合摘要：
  `idempotent_parents=1 idempotent_legs=2 idempotent_children=2 idempotent_freezes=2`
  `freeze_loss_retried=1 kill_canceled=1 margin_canceled=1 fok_zero=1 stp_zero=1 admin_canceled=1`
  `matched_parents=2 trades=2 match_groups=1 trade_legs=2 trade_qty=3`
  `settled_outbox=2 positions=4 trade_instructions=7 trade_instruction_success=7 distinct_trade_instructions=7`；
- 总门禁：9277条 Option 资金指令，9270条成功并对账，7条冻结前合法取消，加权终态9284；
- 501/5000美式指派、501单元实物交割、501多头现金到期，以及3项独立进程强杀/自然租约接管均回归通过；
- 最终输出：`P0 Option/Asset RPC acceptance passed`。

代码门禁：

```text
go test ./...                                                        PASS
go vet ./...                                                         PASS
go test -race ./internal/logic/app ./internal/logic/task ./models   PASS
monitoring/option-production-readiness.sh --repository-only          PASS（补本文档后复跑）
```

本轮没有新增或修改任何表字段、索引、CHECK 或触发器，只把既有组合迁移纳入正式验收并修正运行状态机，
因此不需要执行 `make gen-model`。以后任何 P2-007 DDL 变更必须先执行 `make gen-model`，再检查生成模型 diff
并记录验证结果。

## 5. 资金、原子性和恢复结论

- 冻结屏障以每条影子单原业务号为幂等身份；未知结果不能当作失败释放，提交后响应丢失由原指令重试对账。
- 激活前重检能阻止冻结期间生效的 kill、闭市/halt、行情过期、价格带越界和卖方保证金漂移进入策略簿。
- 第二腿数据库错误证明整个撮合事务回滚，包括第一腿成交、序号、outbox 和资金指令；不存在“补偿式残腿”。
- 成交后的资金执行仍是异步的，但同一 `combo_match_no` 全部买方扣款成功前，所有腿的持仓事件一起关闭；
  成功后沿各合约撮合序号幂等落仓。该设计保证资金安全和最终一致，不伪称跨合约持仓是单数据库事务。
- 用户撤单、控制撤单、FOK/STP和后台强撤只操作父单状态机；不存在单腿释放或单腿管理员权限。

## 6. 生产前剩余验收

以下项目依赖生产等价基础设施或业务批准，仓库不能代签，完成前必须保持复杂订单功能关闭：

1. 在批准的策略簿容量、热点策略、并发提交/撤单/匹配和持续时长下压测，签署 P95/P99、吞吐、死锁率、
   MySQL CPU/锁等待、Redis/Asset连接池及无冻结泄漏；50路同键只证明幂等竞态，不是生产容量承诺。
2. 使用生产编排对组合专用 Option worker/消息消费者执行进程或容器强杀、重复消息、乱序和多实例接管；
   本轮已覆盖真实 Asset 提交后响应丢失及通用资金任务强杀，但未覆盖组合专用进程级故障矩阵。
3. Prometheus/Alertmanager 实例实际抓取组合指标，验证 OPT-A030/组合告警的触发、恢复、IM/电话/值班案件
   路由和真实接收人；仓库当前只证明指标、规则和静态门禁存在。
4. 外部网关验证用户身份、租户注入、三个后台权限、限流和审计导出；RPC 内部租户隔离已实测，不能替代
   生产 IAM/网关验收。
5. 产品、风控、清算/财务、运营、合规/法务批准策略范围、最大父单数量、净价展示、每腿费用、交易时段、
   部分成交/FOK语义、错误文案、行情字段和用户风险披露；明确披露不 legging、不拍卖、不含股票腿、
   不跨到期、不扫多 maker、不提供组合保证金抵扣。
6. 若需要对外复杂簿行情，另行定义公开协议、策略标识、净价方向、深度序列、快照/增量恢复和数据许可；
   当前公开单腿盘口故意排除影子单，不等于已提供 Complex Book market data。

运营签署入口：`docs/templates/complex-order-readiness.md`；故障处置入口：
`docs/option-operations-runbook.md` 的复杂订单章节。
