# 永续与交割合约完成清单

## 1. 目标

本文用于逐项完成并验收永续合约、交割合约。秒合约和现货不属于本文范围。

完成状态统一为：

- `[ ]` 未开始；
- `[~]` 已有主体实现，但未满足完成标准；
- `[x]` 已完成代码、迁移、自动化测试和验收；
- `[!]` 存在生产门禁。

任何涉及真实资金、仓位和交割终态的项目，不能仅凭接口可调用或手工成功一次标记为完成。

## 2. 当前结论

| 模块 | 当前状态 | 生产结论 |
| --- | --- | --- |
| 合约下单与撮合 | `[x]` P0-02 基础矩阵及 P0-07 故障门禁已完成 | 仍受 P0-05 生产价格源和 P0-06 强平告警门禁约束 |
| 逐仓仓位投影 | `[x]` 永续/交割线性、反向及并发故障矩阵已实跑 | P0-07 已完成；自动强平生产门禁仍保持关闭 |
| 全仓 | `[x]` 下单、部分成交/撤单、平仓、账户风险、正/负权益强平、专项对账、模式切换及后台处置均已完成 | 技术验收完成；生产仍由双安全开关和外部审批门禁保护 |
| 永续资金费 | `[x]` 隔离环境完整资金与故障矩阵已实跑 | 生产任务仍受 P1-01 价格源及 P0-06 告警门禁约束 |
| 交割结算 | `[x]` 隔离环境完整生命周期与故障矩阵已实跑 | 自动归档仍受生产 DELIVERY 价格源门禁约束 |
| 自动强平 | `[x]` P1-02 逐仓及 P2-01 全仓矩阵已完成 | 生产总开关保持关闭，待生产告警和资金权限确认 |
| 保险基金与 ADL | `[x]` 逐仓及全仓资金、排序和恢复矩阵已完成 | 生产总开关保持关闭，待生产账户和资金权限确认 |
| Price Engine | `[~]` 技术链路已运行 | 当前单一行情配置仅用于测试 |
| 跨服务对账 | `[~]` 订单、仓位、资金费、交割、逐仓及全仓强平专项闭环已实跑 | 生产告警平台接入仍是门禁 |

## 3. 实施顺序

### P0-01 定时任务接入

- [x] 初始化 `trade.ProcessOrderMatching`；
- [x] 初始化 `trade.ProcessPositions`；
- [x] 初始化 `trade.ProcessContractSettlements`；
- [x] 初始化 `trade.ProcessTradeEvents`；
- [x] 使用幂等 migration 支持已有环境升级；
- [x] 验证任务能发布到 Trade Task Subscriber（全新隔离库初始化四个任务，System Job Log 持续记录四类任务成功；Trade 路由及未知 Action 拒绝由单测覆盖）；
- [x] 验证多实例任务锁不会重复执行同一批次（真实扩容为两个 Trade 容器并同时调用同一 `ProcessPositions`：首轮一个实例 86ms 返回 OK，另一个 16ms 返回“已有同步任务正在执行”；随后连续 10 轮均严格每轮 1 个 OK、1 个锁拒绝，执行权在实例间正常切换；验收后恢复单实例）。

完成标准：

1. 新库执行 `system.sql` 后自动存在四个任务；
2. 旧库执行 migration 后不重复插入；
3. System Job Log 能看到任务持续成功；
4. Trade 服务重启后未完成事件、仓位和结算能够恢复。

### P0-02 合约基础交易矩阵

- [x] 永续 U 本位线性（真实完成独立开仓、加仓、部分减仓、NET 反向开仓、限价 Maker/Taker、市价、IOC、Reduce Only 与三类条件源；独立生命周期中 LONG 1@100 加仓 1@110 后数量 2、均价 105、保证金 21，再于 120 部分减仓 1，释放保证金 10.5、实现盈利 15、剩余 1 张保证金 10.5；订单、Fill、History、Reservation、Instruction、Asset Flow 和 Event 均终态）；
- [x] 永续币本位反向（真实完成 100 张开多、50 张加仓、60 张部分减仓和剩余 90 张全平；反向调和均价、BTC 保证金、正负已实现盈亏、Reservation/Freeze、Fill/History/Asset Flow 均已核对）；
- [x] 交割 U 本位线性（真实 App `PlaceOrder`→Asset 冻结→Matcher→Fill→LONG/SHORT Position→保证金/手续费→余量释放完整通过；另以 IOC BUY 2 张对 SELL 1 张验证成交 1、撤销 1、只消费成交部分并释放剩余 10.404 USDT；此前到期保证金返还、盈亏、费用、关仓及归档也已通过）；
- [x] 交割币本位反向（真实 App 下单 100 张×100 USD，BUY 51000 立即以 Maker 50000 成交；修复反向 BUY 按较高委托价低估 BTC 预占后，订单风险价使用最低可执行 Maker 价，精确冻结 0.0204 BTC、形成 0.02 BTC 保证金并扣 0.0004 BTC Taker 费；双方 Fill/Position/History/Instruction/Flow/Event 全部终态；此前 50000→55000 到期盈亏、费用和归档也已通过）；
- [x] LONG、SHORT、NET（反向 LONG 完整生命周期通过；线性 NET BUY 真实先关闭 2 张 SHORT、再开 1 张 LONG，Fill 已实现盈亏 20、两条 Position History 与三个资金步骤均唯一）；
- [x] 开仓、加仓、部分减仓、全部平仓、反向开仓（反向 LONG 100→150→90→0 完整实跑；线性 NET 2 张 SHORT 被 3 张 BUY 反向为 1 张 LONG，开仓余量仅计提 9 USDT 保证金）；
- [x] 限价、市价、条件单、Reduce Only（限价、IOC、市价单、Reduce Only 及 LAST/MARK/INDEX 三类条件源均完成真实撮合或无流动性撤单与 Asset 联调；MARK/INDEX 仅接受 30 秒内对应类型的已确认快照，触发时间、实际价格和来源均进入审计；FOK/Post Only 输入及撮合边界由自动化测试覆盖）；
- [x] 部分成交、全部成交、价格改善（真实限价全成和价格改善已通过：Maker/Taker 成交价 100、手续费 0.1/0.2、买卖双方余量 0.102/0.1 均真实解冻且冻结归零；IOC 2 张真实成交 1 张并撤销余量 1 张，只消费成交部分保证金 10 和手续费 0.2，剩余预占 10.404 全额释放）；
- [x] 撤单与成交并发（隔离实库对目标订单持行锁，同时放入 Matcher 与 Cancel RPC 后释放：本次撮合获胜，双方订单各成交 1、撤销 0；晚到撤单返回成功但无撤单日志和资金变更，`filled_qty + canceled_qty <= qty` 违规数为 0，全部冻结、事件、指令和对账均终态）；
- [x] 重复 Fill 和重复事件（相同 `POSITION_FILL_REQUIRED` 真实双重放后 Position/History/Instruction/Asset Flow 保持唯一；成交入口新增 `fill_no` 与 `match_no + order_id` 双键及完整业务身份校验，完全一致重投零写入，改数量或换成交号重投拒绝；实库两种重复 INSERT 均被对应唯一键拒绝，事实数量不变）。

每个用例必须核对：

```text
Order
Fill
Position
Position History
Reservation
Settlement Instruction
Asset Flow
Outbox / Inbox
```

完成标准：

1. 数量、金额、保证金、手续费和盈亏使用 Decimal；
2. 线性与反向公式符合设计；
3. 重放不会重复扣款、入账或更新仓位；
4. 订单终态只在 Asset 和 Position 投影完成后出现；
5. 成交量加撤销量不超过委托总量。

### P0-03 永续资金费

- [x] 正资金费率：多头付款、空头收款（方向公式单测及线性多头真实资产扣款通过）；
- [x] 负资金费率：空头付款、多头收款（方向公式单测及线性多头真实资产入账通过）；
- [x] 零资金费率（真实批次无 Asset 指令并直接完成 Settlement、Batch 和仓位时间投影）；
- [x] 多空金额不能完全抵消时由平台差额账户承接（正负费率真实批次均由平台账户反向承接，用户与平台合计净额为零）；
- [x] 平台差额账户缺失时批次整体拒绝（真实缺失时 Batch/Settlement/Instruction=0、用户余额和仓位不变并输出明确错误；补齐账户后同周期自动完成且用户 -1/平台 +1）；
- [x] 用户余额不足（真实应付 15000、可用 99；第 20 次失败后用户 Instruction=Manual、Settlement=Manual、Batch=Manual，平台步骤保持 Pending，双方余额/Position/Flow 均不变）；
- [x] 同一周期重复执行（高频任务在同一分钟多次执行，实库只保留一个 Batch、每个 Position/平台差额各一条 Instruction 和 Asset Flow）；
- [x] 历史周期补跑（真实预置三分钟前 Completed Batch 后，任务按 `T+1/T+2/T+3` 逐分钟创建并完成三个唯一批次；Settlement/History/last_funding_time 顺序连续，无跳期或重复）；
- [x] 资金费与成交仓位投影并发（Settlement 固化 qty=1/version=1 后模拟加仓到 Position qty=2/version=2；恢复仅按旧事实扣 1，Position 保持 qty=2 并在 version=3 追加资金费，不发生陈旧覆盖）；
- [x] Asset 成功、Trade 超时（先以正式 Instruction ID/BizNo 真实扣款 1 并保持 Trade Pending，再放开 Trade 重试；Asset 返回既有结果，用户业务号 Flow 始终仅 1 条，本地 Settlement/Position/Batch 正常补齐）；
- [x] Trade 本地提交前后进程退出（Trade 停机时预置 Pending Batch/Settlement/两步 Instruction，启动后仅凭 MySQL 持久事实自动完成；用户/平台各唯一 Flow，Position 时间推进且 Batch 完成）；
- [x] `last_funding_time` 和仓位版本核对（正、负、零三类真实批次均写入结算时点且版本从 1 增至 2）；
- [x] Funding Batch 资金守恒（正费率用户 -1/平台 +1，负费率用户 +1/平台 -1，零费率双方 0）；
- [x] Funding Settlement 与 Asset 流水对账（非零 Settlement 各关联唯一 Asset Flow，Batch 对账完成后进入终态 3）。

完成标准：

```text
所有用户资金费 + 平台差额 = 0
```

并且每个 Settlement 均能关联唯一 Asset 流水，失败可自动重试或进入人工处理。

### P0-04 交割合约生命周期

- [x] `open_cutoff_time` 后进入 `CLOSE_ONLY`（真实定时任务推进 Symbol 和 Delivery Batch，生命周期版本单调递增）；
- [x] `matching_stop_time` 后停止撮合（真实定时任务推进 `MATCHING_STOPPED`，Pending 与 Part Filled 活动订单均先撤销后归档）；
- [x] 撤销活动委托（真实 Pending、Part Filled、Trigger Waiting 均被分页撤销，并满足成交量+撤销量=委托量）；
- [x] 释放资金和可平仓数量预占（真实全额预占释放 10、部分成交仅释放剩余 6，均与 Asset Freeze/Flow 对账；Reduce Only 的 Reserved Close Qty 0.5 释放后仓位 avail/frozen 从 0.5/0.5 恢复为 1/0）；
- [x] 未完成订单清零后锁定交割价（Pending、Part Filled、Trigger Waiting、Canceling、Settlement Pending 均有真实门禁证据；订单终态和预占/成交结算完成前 Batch 只到 MATCHING_STOPPED，价格保持 0）；
- [x] 创建唯一 Delivery Batch（真实高频任务重复执行后，租户、Symbol、交割时间仅一个 Batch）；
- [x] 逐仓计算盈亏和交割手续费（U 本位线性与 BTC 本位反向均真实验收；反向 100 张×100 USD、50000→55000 的 PnL=0.018181818182 BTC、Fee=0.001818181818181818 BTC）；
- [x] 先返还 Trade 已持有的仓位保证金，再扣亏损/费用，最后发放盈利（真实指令按 step 1/2/3 执行并绑定唯一 Asset Flow）；
- [x] 释放仓位保证金（成功终态 Position/Isolated/Maintenance Margin 全部清零，10 USDT 保证金返还已对账）；
- [x] 关闭仓位并写 Position History（真实仓位状态 5、数量归零，唯一 Settlement History 与最后一个 Asset 指令同链路完成）；
- [x] Delivery Settlement 与 Asset 流水对账（三条真实指令均绑定唯一 Asset Flow 和 `reconciled_at`）；
- [x] Batch 对账完成后进入终态（Settlement、Position、Instruction 和流水全部完成后才进入 Completed）；
- [x] Symbol 最终归档（Batch 完成后进入 Archived=8，并成功投递唯一 `CONTRACT_SETTLED` 事件）。

必须覆盖：

- [x] 无持仓到期（真实 Batch `total_positions=0`、Settlement/Instruction 均为 0，仍完成、归档并投递唯一事件）；
- [x] 存在未完成订单（五种计数状态均已覆盖：三种活动态先撤销，Canceling 等待资金释放，Settlement Pending 等待 Fill/Order 收尾；全部清零后才锁价）；
- [x] 部分成交订单到期（订单 1 中已成交 0.4、撤销 0.6，Trade 预占消耗 4/释放 6，Asset Freeze 关闭且唯一 Flow=6；随后唯一 Batch 无仓位归档）；
- [x] 交割价缺失或超出窗口（真实缺价、窗口外报价、窗口内但缺 formula_version 均停在 PRICE_LOCKING 并持久化原因，Settlement/归档为 0；补入合法最终价后自动恢复；非单一最终价拒绝另有单测）；
- [x] 零盈亏、零手续费、零 Asset 步骤（真实零盈亏/零费用仅生成保证金返还；另一个 position_margin=0 场景 Instruction=0，仍在同事务关闭仓位、写唯一 History 并完成归档）；
- [x] 用户亏损扣款失败（真实亏损 200、可用 108.9；第 20 次失败后 Instruction=Manual、Settlement=Manual、Batch=Manual，Position/余额/History/Flow 均未变化且无归档事件）；
- [x] Asset 执行到一半服务重启（第一步保证金返还成功后真实停止/启动 Asset，再放行费用与盈利步骤；三条稳定业务号各唯一 Flow，最终 Position/Settlement/Batch/Event 全部恢复完成）；
- [x] 同一批次重复执行（高频任务多次执行后 Batch、Settlement、History、归档 Event 各一条，三类 Instruction/Asset Flow 各自唯一）。

### P0-05 交割价格

- [~] Price Engine 输出 `DELIVERY` 快照（隔离四源公式已真实出价并发布，仍待生产独立市场源配置验收）；
- [x] 固化窗口内完整输入集合（真实 DELIVERY 审计固化 all=4、accepted=3、rejected=1，每项绑定不可变 Snapshot ID）；
- [~] 支持多样本中位数或加权平均（中位数真实四源出价为 100；加权平均已有自动化测试，仍待生产参数回放）；
- [x] 固化异常剔除摘要（200 BPS 下来源 150 被剔除，缺第 4 源时明确 `accepted=2 required=3 rejected=1` 且输出为 0；补入来源 99 后审计仍保留被剔除输入）；
- [x] 公式版本、目标时点和来源快照可审计（真实输出固化公式 `acceptance-delivery-3source-v1`、毫秒目标时点和四个输入 Snapshot ID）；
- [x] 使用原输入集合确定性重放（离线工具读取实库不可变审计重算为 100；仅把审计输出改成 101 后以 output mismatch 拒绝）；
- [x] Trade 不对最终 `DELIVERY` 快照二次定价（真实 Batch 锁定价格与最终快照价格 110 完全一致）；
- [x] 配置的算法名称与实际算法一致（真实 Batch 固化配置算法 `FINAL_SNAPSHOT_V1` 与公式版本 `acceptance-delivery-v1`）。

完成标准：

1. 同一公式版本、目标时点和输入集合得到相同价格；
2. 被撤销快照不会继续用于新批次；
3. 输入不足时阻断交割并告警，不使用无审计的最新价兜底。

### P0-06 自动对账和告警

- [x] Order 与 Fill 对账（真实全量循环游标持续完成；按交易对 `price_scale` 归一比较均价后，反向合约约 10^-15 的除法尾差不再误报，历史两条 issue 已自动 RESOLVED；当前验收 Order/Fill 数量、金额、终态无 OPEN 差异）；
- [x] Fill 与 Position History 对账（真实发现缺失投影并累计 964 次，补齐不可变 History 后三条 issue 均自动 RESOLVED；当前无 OPEN）；
- [x] Reservation 与 Asset Freeze 对账（全额/部分释放均绑定唯一 Freeze/Flow；曾真实发现缺 Flow 并在恢复执行后自动 RESOLVED，当前无 OPEN）；
- [x] Position 保证金与合约钱包对账（真实发现五条无托管流水夹具差异并持续累计，修正资金事实且经过 60 秒稳定窗口后全部自动 RESOLVED；当前无 OPEN）；
- [x] Funding Settlement 与 Asset Flow 对账（正/负、失败、进程恢复、Asset 先成功等真实批次均以稳定业务号绑定唯一 Flow，Batch 仅在 reconciled_at 后完成）；
- [x] Delivery Settlement 与 Asset Flow 对账（线性/反向、全额/部分释放和中途重启均绑定唯一 Flow，Batch 仅在全部指令对账后归档）；
- [x] 强平、保险基金和 ADL 对账（全额强平、保险全额/部分/不足/冲正、ADL 两阶段恢复、父 Saga 和部分强平终态均已实跑；专项扫描核对仓位、历史、完成事件、保险承接、ADL 数量及 Asset 流水，隔离验收无 OPEN 差异）；
- [x] 差异持久化（真实 issue 以稳定 key 跨周期累计至 30～794 次，修复事实后由扫描任务自动写 status=RESOLVED/resolved_at/reason；非强平 OPEN=0）；
- [~] 差异告警（数据库行锁内固化 `last_alert_at`：首次、恢复后重开、内容变化立即告警，未变化每 30 分钟提醒；真实 `LIQUIDATION:1` 在 67 秒内继续扫描 13 次但新容器仅输出 1 条；仍待接入生产告警平台、值班组和升级规则）；
- [x] 带操作人、原因和请求号的人工处理入口（真实 Admin RPC 验证租户元数据覆盖请求租户、缺操作人拒绝、OPEN 忽略成功、重复忽略拒绝；状态=IGNORED、操作人 `990001`、原因、解决时间和唯一 `TRE` 审计事件全部落库，隔离夹具随后精确清理）。

完成标准：

1. 对账任务可重复执行；
2. 对账结果有业务日期和稳定唯一键；
3. 差异不会被后一次任务静默覆盖；
4. 修复动作具备完整审计。

### P0-07 故障注入

- [x] Redis 超时（除停机期间拒绝无锁执行、恢复后四类 Job 自动续跑外，已在 `ProcessTradeEvents` 持锁且阻塞于 Settlement 指令行锁时停止 Redis；续租失败取消业务上下文，等待中的 `ClaimLease` 返回 `context canceled`，指令保持 Pending、retry=0 且无错误落库；清理隔离记录并恢复 Redis/Trade/System 后任务再次返回 200）；
- [x] Kafka 发布超时和重复投递（隔离 Kafka 停机/恢复已修复并验证 Subscriber 以 1～30 秒退避持续重连；永续线性、永续反向、交割线性、交割反向各选一个真实已完成 `POSITION_FILL_REQUIRED` 事件并额外投递两次，消费组 lag 回到 0，四组 Order/Fill/Position/History/Instruction/Asset Flow/Inbox 指纹逐字节不变；撤单撮合并发和重复 Fill 双唯一键注入也已通过）；
- [x] Asset RPC 超时（真实停止 Asset 后 Delivery 指令持久失败并退避 7 次，恢复后不重建 Batch 即自动续跑；另有余额不足第 20 次进入 Manual 的真实边界）；
- [x] Asset 成功但 Trade 未确认（正式 Funding BizNo 先直连 Asset 成功且 Trade 保持 Pending，重试后原 Flow 仍唯一并补齐本地状态）；
- [x] MySQL 死锁（除通用双事务注入外，已用真实 `Position` 与 `SettlementInstruction` 两个事实反向加锁产生 1213；失败侧回滚后由统一重试器第 2 次成功，最终 Position `realized_pnl=2/version=2`、Instruction `retry_count=2`，无丢失或重复；仅 1213/1205 最多重试 3 次，业务错误不重试）；
- [x] Trade 事务提交后立即退出（真实停止 Trade 后保留 Pending Funding Batch/Settlement/Instruction，启动后仅凭 MySQL 自动恢复；此前 Fill/Order 终态恢复与 Kafka 重放也保持唯一）；
- [x] Position 乐观锁冲突（真实持有 Position 行锁并在事务内模拟成交把 qty 1→2、margin 10→20、version 1→2；旧 version=1 的标记价 Worker 等待 5.37 秒后 CAS 返回 false，成交字段和旧 Mark 均未被覆盖；随后 version=2 的 CAS 成功且只更新 Mark/Risk 派生字段，qty/margin 保持 2/20、version 推进至 3）；
- [x] Worker 租约过期（隔离 Pending Settlement 指令由两个容器真实竞争，首轮严格一方 Claim 成功；模型时间推进 60.001 秒后另一 Worker 成功重领并写新 lease，旧 lease 完成影响 0 行、新 lease 影响 1 行，证明过期恢复与 fencing 生效）；
- [x] 多实例并发领取（Trade 任务锁双实例连续 10 轮均保持每轮一个执行、一个锁拒绝；Settlement 指令再由两个独立容器以相同毫秒值并发调用真实 `ClaimLease`，严格一个 `claimed=true`、一个 `false`，不存在重复领取）；
- [x] 服务重启恢复（真实停止/启动 Asset 于三步 Delivery Saga 中间、停止/启动 Trade 于 Funding Saga 执行前，均无需人工重置恢复；指令、Flow、History、Batch、Event 唯一）；
- [x] Price Engine 输入暂时缺失（隔离公式在来源不存在时连续领取 220 次仍保持目标快照数为 0；告警包含 formula、authority、kind、category、market、symbol、target 和底层原因，30 秒窗口仅输出 1 条；发现 `Mode: pro` 会过滤 Info 后将节流告警提升为 Error，补入来源快照后无需重启自动恢复产出且 Outbox 全部成功）；
- [x] Snapshot Outbox 大量积压（隔离环境先注入 50,000 条完整 Redis+Kafka 消息验证大表健康查询与 ETA，再保留 11,536 条跑至终态；32 Worker 净排空 `30.93/s`，持久化切换 64 Worker 后为 `61.87/s`，`processing` 分别稳定在 32/64 仅表示瞬时在途；保留样本全部 Success、retry=0、Redis/Kafka 检查点无缺失、错误数 0，测试表和缓存已清理）。

完成标准：不存在重复资金变动、重复仓位投影、永久中间态或不可定位的资金差异。

### P1-01 Price Engine 生产配置

- [~] INDEX 使用至少三个独立市场来源（新公式强制三种独立来源且运行时按 Snapshot ID 去重，待生产源配置）；
- [~] MARK 使用指数价及溢价/基差（`INDEX_BASIS` 代码、协议、迁移、管理端和单测完成，待生产 v2 配置回放）；
- [~] 配置偏离剔除（INDEX/DELIVERY 支持异常剔除；MARK 基差支持对称 BPS 限幅，待参数验收）；
- [~] 配置平滑、上下限和版本（上一 MARK 加权平滑、基差上下限及不可变公式版本已完成，待生产参数回放）；
- [~] 完成生产历史回放（离线确定性回放工具及篡改检测单测完成，待导出生产时间段执行）；
- [x] Snapshot Outbox 回到实时水位（Claim/发布已流水线并发、最终检查点合并、健康日志新增净排空速度和 ETA；64 Worker 部署后完成 50,000 条容量验收，当前实库 `total=0/open=0`，近 10 分钟无 unhealthy、Worker 失败或慢查询日志）；
- [~] 完成容量、备份和灾备恢复演练（50,000 条 Outbox 容量验收完成；隔离环境完成 28.23 MiB、140 表一致性备份与恢复，逐表差异为 0，并固化恢复手册。生产异地备份、Binlog 时间点恢复、可用区切换及正式 RPO/RTO 仍待生产演练）。

当前 BTCUSDT MARK 和 INDEX 使用相同单一行情，适合技术测试，不适合作为真实资金费来源。

### P1-02 强平、保险基金和 ADL

- [x] 标记价格驱动风险重算（真实 `price-engine/MARK` 110 快照驱动仓位风险投影；重复读取同一投影不再产生版本写放大）；
- [x] 撤销活动订单（普通开仓单释放 10 USDT Asset 冻结；Reduce-Only 单释放 0.25 Position 可平数量；两者均在接管前撤销）；
- [x] 部分强平（真实风险档从第二档降至第一档：逐仓 LONG 10@100、MARK 96、保证金 100 时风险率 1.6；部分减仓 5 后剩余 5、保证金 50、未实现盈亏 -20、风险率 0.16、状态恢复 NORMAL；用户残值 25.2、平台手续费 4.8，各唯一流水）；
- [x] 全部强平（隔离线性永续仓位 1→0，Position/History/Liquidation/Completed Event 各唯一；同仓位重放无新增事实）；
- [x] 强平手续费（触发权益 5 中手续费 0.95 原子计入 Asset `FEE_REVENUE`，用户残值 4.05；穿仓时手续费被限制为 0，不进入保险基金缺口）；
- [x] 保险基金全额赔付（真实 Asset RPC：请求 30、承接 30、余额 100→70；同业务号重放不重复扣款）；
- [x] 保险基金部分赔付（真实 Asset RPC：请求 50、仅承接可用余额 20、剩余缺口 30、余额 20→0）；
- [x] 保险基金不足（真实 Asset RPC：余额为 0 时请求 10、承接 0、缺口 10，不生成零金额流水）；
- [x] 保险基金冲正（30 全额返还 70→100；同冲正号重放不重复入账，不同冲正号被拒绝）；
- [x] ADL 候选排序（按盈利率 × 有效杠杆生成持久化优先级；仅逐仓候选可参与，优先级相同时按 Position ID 稳定排序）；
- [x] ADL 数量上限（执行总量同时受候选数量、被接管剩余数量和剩余缺口限制）；
- [x] ADL Asset/Position 两阶段恢复（真实从 PREPARED/PROCESSING 及 ASSET_DONE 两个故障点恢复，Asset 流水保持唯一）；
- [x] Liquidation 父 Saga 恢复（ADL 子执行完成后父记录恢复为 COMPLETED，关闭破产仓位并生成唯一完成事件）；
- [x] 日终对账与差异告警（强平专项全量循环游标、稳定差异键、自动恢复和结构化节流告警已完成；部分强平终态真实扫描完成且无 OPEN 差异。生产告警平台接入仍作为 P0-06 上线门禁）；

P1-02 应用与隔离运行验收已完成；生产 `AutomaticLiquidation.Enabled` 仍必须保持 `false`，直到生产价格源、告警平台、保险基金账户及资金权限完成审批和演练。

### P2-01 全仓

- [x] 账户级权益快照（按租户、用户和保证金币种聚合 Asset 钱包、全仓仓位保证金与未实现盈亏；迁移、版本化模型、任务投影及 Admin RPC 审计字段已完成）；
- [x] 同结算资产仓位共享保证金（全仓开仓时由 Trade 在用户级分布式锁内同步读取 Asset 钱包、同结算资产全部开放仓位、活动订单和新鲜 MARK 风险投影；完整成交、部分成交后撤单、关闸平仓及资金释放均已真实验收）；
- [x] 未成交订单保证金占用（真实 Pending 全仓订单的 30 USDT Reservation 与 Asset Frozen 均进入快照，`order_margin=30`）；
- [x] 账户级维持保证金（真实两个仓位分别为 5/6，账户快照聚合为 11）；
- [x] 账户级风险率（真实 Wallet 100 + Position Margin 30 + UPNL -20 = Equity 110，Risk=11/110=0.1；非正权益使用可存储上限值并有单测）；
- [x] 账户级强平（已按租户、用户、保证金币种建立父 Saga 和仓位明细，支持先撤同风险单元全部活动订单、释放 Reservation、行锁复核账户风险、一次接管全部全仓仓位、净额结算用户钱包、独立幂等结算手续费及失败退避恢复；负权益增加 `INSURANCE_FUND`、`ADL` 持久化阶段，使用资产级默认保险配置 `symbol_id=0`，按冻结亏损分摊剩余缺口并固化合成破产价。ADL 子执行以负的账户明细 ID 作为内部命名空间，与逐仓强平正 ID 隔离；父、明细、保险赔付、ADL 执行和 Asset 流水均进入专项对账。正权益双仓、保险全额、保险部分后 ADL、重启幂等、后台详情及门禁保护重试均已验收）；
- [x] Asset 与 Mark Price 持续投影（`ProcessPositions` 每轮从 Asset RPC 读取钱包版本，并在 MARK 更新后聚合跨仓位风险；风险单元现为开放仓位、活动订单和历史快照的并集，可投影仅有挂单的账户并清零已消失单元的旧快照。真实双标的 MARK 与 Asset 版本并发验收后收敛到同一不可变版本哈希，重复执行无写放大）；
- [x] 全仓与逐仓切换规则（用户全局模式锁、标的模式锁与全仓账户锁顺序固定；存在不兼容开放仓位或活动订单时拒绝切换；仓位持久化 `position_mode`，不再以实际 LONG/SHORT 方向猜测单向/双向模式；开仓受双门禁保护，Reduce Only 平仓不受门禁影响但必须匹配仓位模式）；

全仓技术链路已经完成，但生产默认仍不开放新增风险：`CrossMarginTrading.Enabled=false` 且 `AutomaticLiquidation.Enabled=false`，只有两者同时开启才接受增加全仓敞口的订单。Reduce Only 平仓始终保留；关闭自动强平开关只阻止创建新的账户强平 Saga，已完成 Asset 副作用的 Saga 仍允许幂等恢复，避免遗留中间态。

## 4. 上线门禁

永续资金费上线必须同时满足：

- P0-01、P0-02、P0-03、P0-06、P0-07 完成；
- Price Engine 使用经过验收的 MARK、INDEX 和 FUNDING；
- Snapshot Outbox 无持续积压；
- 自动对账连续运行且无未处理差异。

交割上线必须同时满足：

- P0-01、P0-02、P0-04、P0-05、P0-06、P0-07 完成；
- 交割价格可确定性重放；
- Delivery Batch 必须等待 Asset 对账后才能归档。

自动强平、保险基金和 ADL 上线必须额外完成 P1-02。

## 5. 实施记录

| 日期 | 项目 | 结果 | 证据 |
| --- | --- | --- | --- |
| 2026-07-28 | 建立实施与验收清单 | 已完成 | 本文档 |
| 2026-07-28 | P0-01 定时任务数据库接入 | 代码完成，待部署验证 | `services/system/system.sql`、`20260728_add_trade_contract_jobs.sql` |
| 2026-07-28 | P0-03 资金费公式矩阵 | 单元测试完成，端到端验收待继续 | 正/负/零费率、线性/反向、方向及平台差额守恒测试 |
| 2026-07-28 | P0-03 资金费快照审计 | 已修复 | Trade 行情投影保留 Price Engine `formula_version` |
| 2026-07-28 | P0-03 历史周期持仓事实与顺序补跑 | 代码完成，待迁移后端到端验收 | Position History 增加业务时间、版本、结算币种；按结算时点重建仓位；从最近已完成 Batch 逐周期补跑，前序未完成时阻断后序 |
| 2026-07-28 | P0-04 交割资金步骤顺序 | 已修复，端到端验收待继续 | 新批次保证金→亏损/费用→盈利；旧批次保持兼容恢复 |
| 2026-07-28 | P0-03/P0-04 Saga 查询容量 | 已修复，待生产 EXPLAIN | 增加业务待处理与批次步骤覆盖索引 |
| 2026-07-28 | P0-06 Settlement/Asset 流水自动对账 | 代码完成，待部署联调 | 按指令号核对身份、币种、金额、操作类型；固化 Asset 流水号；稳定差异键累计、错误告警、自动恢复；Funding/Delivery Batch 对账前不得完成 |
| 2026-07-28 | P0-06 Order/Fill、Fill/Position、Reservation/Freeze、Position/Margin 对账 | 代码及单测完成，待迁移部署验收 | 全量循环游标避免只扫热点数据；稳定延迟后核对成交聚合、仓位投影、冻结守恒和保证金托管守恒；差异使用稳定业务键持久化并自动恢复 |
| 2026-07-28 | P0-06 强平/保险基金/ADL 对账 | 代码及单测完成，待部署联调 | 核对强平终态、破产仓位清零、唯一历史/完成事件、ADL 数量与执行终态、ADL 用户资产流水；新增 Asset 保险赔付只读接口，按强平号核对平台账户赔付的币种、强平 ID、请求额、承接额、ADL 接续缺口及冲正状态 |
| 2026-07-28 | P0-07 故障注入基线 | 自动化不变量测试及验收手册已建立，待隔离环境逐项执行 | 覆盖 Asset 超时持久重试、退避上限、人工处理、过期租约 fencing、旧保证金指令拦截、Price Engine 缺源阻断；`perpetual-delivery-fault-injection-runbook.md` 固化 11 类注入步骤、SQL 证据和通过条件 |
| 2026-07-28 | P0-07 Position 并发覆盖防护 | 代码完成，待并发注入验收 | 标记价扫描不再用无条件全行 Update 覆盖仓位；仅 CAS 更新 mark/risk 派生字段，version 冲突时放弃陈旧结果并由下轮重算 |
| 2026-07-28 | P1-02 自动强平生产安全门禁 | 已修复并有单测，待配置部署验证 | 新增 `AutomaticLiquidation.Enabled`，默认 false；未验收时触发风险阈值只生成 `MANUAL_REVIEW` 事实，不执行新保险赔付/ADL；已经开始的资金 Saga 仍允许恢复，避免永久中间态 |
| 2026-07-28 | P0-02 衍生品计算矩阵 | 单元测试完成，端到端矩阵待执行 | 覆盖永续/交割 × 线性/反向、LONG/SHORT/NET 方向、NET 反向开仓只按开仓余量计提保证金 |
| 2026-07-28 | P0-03 资金费输入事实绑定 | 已修复并有单测，待历史回放 | 组合资金费快照的内容哈希现在绑定 MARK、INDEX、FUNDING 三个输入快照 ID；补齐品类、市场、交易对和 Price Engine 权威方；Batch 使用不可变快照实际公式版本，不再硬编码 `premium-v1` |
| 2026-07-28 | 成交保证金扣款时点修正 | 已修复，待端到端与存量数据验收 | 不再在撮合阶段按整笔 Fill 预生成保证金扣款；改为仓位投影后仅按实际开仓数量生成并绑定 Position；兼容修复未执行的旧 `position_id=0` 指令，纯平仓旧指令删除，已执行冲突拒绝自动修正 |
| 2026-07-28 | P0-06 对账异常人工处理入口 | 通过 | Trade/Admin API 提供租户隔离的异常列表及忽略接口；真实 RPC 以请求租户 123、可信元数据租户 900199 查询时只返回 900199 夹具；缺操作人返回 1001，操作人 990001 成功忽略，重复忽略返回 1001。数据库固化状态 3、原因、resolved_at，并生成唯一成功 `CONTRACT_RECONCILIATION_ISSUE_IGNORED` 事件；验收记录、事件、缓存和临时客户端均已精确清理 |
| 2026-07-28 | P0-05 最终交割价边界与审计 | 代码完成，待 Price Engine DELIVERY 配置联调 | Trade 仅接受单个已确认最终 DELIVERY 快照，不再聚合二次定价；Batch 固化快照ID、公式版本、目标时点、配置算法及原始摘要 |
| 2026-07-28 | P0-05 Price Engine 输入审计与确定性重放 | 代码及单测完成，待生产多源配置回放 | `raw_payload` 固化完整/采用/剔除输入；快照 ID 由输出维度、价格和审计事实确定性生成；补充 DELIVERY 多源配置说明；`services/itick go test ./internal/priceengine ./internal/tasks ./internal/logic/itick` 通过 |
| 2026-07-28 | P0-05 交割价有效输入门槛 | 代码及管理端完成，待迁移部署与生产三源验收 | 公式新增 `min_input_count`；DELIVERY 运行时强制至少 3 个偏差过滤后的有效来源，不足时拒绝出价；创建/详情页可配置和查看；补充偏差剔除后仅剩 2 源不发布快照的故障测试 |
| 2026-07-28 | P0-03 资金费能力审计 | 代码与公式单测完成，故障及生产验收未完成 | 正/负/零费率、线性/反向计价、平台差额守恒、批次唯一、指令幂等、租约重试、`last_funding_time` 防重、Position History 及 Asset Flow 对账链路均已落地；未将尚未执行的数据库并发、进程退出和真实 Asset 故障场景标为完成 |
| 2026-07-28 | P0-04 Delivery Batch 持久化状态机 | 代码与单测完成，待定时任务和数据库联调 | 修复此前仅在交割时创建 Batch、无法审计 `CLOSE_ONLY/MATCHING_STOPPED/PRICE_LOCKING` 的缺口；状态单调推进，锁价错误持久化并可重试，已有生命周期 Batch 可继续原子升级为 Settling；同步 Task/Admin/App/Trade 四个触发入口，Trade 全量测试通过 |
| 2026-07-28 | P0-01/P0-07 任务锁续租 fencing | 代码和故障单测完成，待双实例 Redis/Kafka 验收 | 修复续租失败后旧 Worker 仍无锁运行的并发窗口；所有七类 Trade 定时任务现在使用可取消任务上下文，续租失败立即取消 DB/RPC 调用并向 Kafka 返回错误；Subscriber 路由表覆盖全部计划任务，且单测证明仅成功/已被另一实例领取时确认，空响应和失败均触发重试 |
| 2026-07-28 | P0-02 合约基础矩阵代码审计 | 主体代码和纯逻辑测试具备，待数据库/Asset 组合验收 | 四种 ContractType/ValueType 盈亏矩阵、LONG/SHORT/NET、NET 反向开仓余量、订单类型/TIF/触发、部分成交、Maker 价格、尘埃处理及对账约束已有自动化证据；未用纯逻辑单测替代跨表和真实资产流水验收 |
| 2026-07-28 | P1-01 INDEX_BASIS 生产标记价 | 代码、迁移、管理端及单测完成，待生产 v2 配置与历史回放 | 新算法按 `INDEX × (1 + bounded basis)` 计算 MARK，并可用上一时点 MARK 加权平滑；强制 INDEX/独立 FINAL_QUOTE/可选上一 MARK 的有序同 Symbol 输入、正基差上限和全部输入门槛；上一 MARK 查询固定为 `target-1ms` 防止自引用；正/负/零基差、上下限和平滑均有测试，原始/采用基差及未平滑价进入不可变审计 |
| 2026-07-28 | P1-01 三源 INDEX 与离线回放 | 代码和单测完成，待生产源配置及历史数据执行 | 新 INDEX 公式强制 MEDIAN/WEIGHTED_MEAN、至少三种独立 FINAL_QUOTE 来源和三个有效输入；引擎按 Snapshot ID 去重阻断伪多源；审计新增输出价，`cmd/price-replay` 可脱离当前配置/行情重算并拒绝输出篡改或重复输入 |
| 2026-07-28 | P1-01 Snapshot Outbox 排空优化 | 代码和单测完成，待部署容量验收 | 候选 Claim 与 Redis/Kafka 发布改为 Worker 流水线并发；Kafka 检查点与 Success 状态合并为原子更新，减少每条成功记录一次 SQL；健康日志输出 `drain_per_sec` 与 `eta_seconds`，明确 `processing` 仅为瞬时在途数 |
| 2026-07-28 | P0-04 交割完成归档 | 代码完成，待端到端验收 | Batch 仅在资金流水对账完成后进入 COMPLETED，随后进入 ARCHIVED 并发送稳定幂等号的 `CONTRACT_SETTLED` 事件 |
| 2026-07-28 | P0-07 MySQL 事务并发恢复 | 通过 | 合约核心事务统一处理 MySQL 1213/1205：最多 3 次、10ms/20ms 退避、响应上下文取消，普通业务错误只执行 1 次。真实 MySQL 中用关闭的 Position ID=991814 与成功 Instruction ID=197 反向加锁，明确产生 1213，失败 Worker attempts=2、另一方 attempts=1；最终 Position `realized_pnl=2/version=2`、Instruction `retry_count=2`，证明回滚重试无丢失或重复。新增显式 DSN 启用的真实事实集成测试，隔离记录执行后均为 0 |
| 2026-07-28 | P0-07 Kafka/进程恢复审计 | 代码链路闭环，待外部故障注入 | Trade Event 的事务 Outbox、领取租约、失败退避/死信、Inbox 唯一防重和 claimant fencing 已核对；撮合提交后的 Fill 与 Outbox 同事务，后续 Position/Settlement 由持久事件恢复；未把真实 Kafka 阻断、重复投递和进程 kill 提前标完成 |
| 2026-07-28 | 隔离 Compose 运行时基线 | 通过 | 全新数据库初始化完成（36 个 migration）；Trade/Asset/Itick/System、MySQL/Redis/Kafka/Etcd 健康；四个 Trade Job 持续成功；Batch、Instruction、History、Outbox、Inbox 关键唯一索引实库存在 |
| 2026-07-28 | P0-07 真实 MySQL/Redis 注入 | 部分通过 | MySQL 双事务实际触发 1213，失败事务回滚且重试后两测试行均精确累计两次；Redis 双 Worker 验证非持有者不能续租或释放，持有者释放后新 Worker 才能领取；集成测试均为环境变量显式启用 |
| 2026-07-28 | P0-07 Kafka 停机与重放 | 发现缺陷后修复并复测通过 | 首次注入发现 Broker 断开会令 Trade Event Subscriber 永久退出；新增 Event/Task Subscriber 持续重连及退避单测；修复镜像中再次停机/恢复无需重启 Trade，Outbox 从失败/领取态恢复成功，Inbox 唯一；相同事件双重放后 Outbox/Inbox 仍各一条 |
| 2026-07-28 | P0-07 Redis 停机与持锁续租失败 | 通过 | Redis 停机期间 Trade 明确拒绝取得任务锁并向 Kafka 返回失败，没有无锁执行；恢复后四类任务无需重启自动恢复。另以隔离 Settlement 指令 ID=196 的数据库行锁阻塞 `ProcessTradeEvents`，确认任务锁 owner=`1785237937719533925` 且 TTL 正常后停止 Redis；续租失败使业务上下文取消，阻塞的 `ClaimLease` 返回 `context canceled`，指令仍为 status=1/retry=0、无错误落库。随后停止 Trade、释放行锁、精确删除测试记录与缓存，Redis/Trade/System 均恢复 healthy，任务调用返回 200 |
| 2026-07-28 | P0-02 线性永续真实开仓与幂等重放 | 通过（该子场景） | 隔离租户 `900101` 的 `POSITION_FILL_REQUIRED` 经真实 Outbox→Kafka→Inbox→Position→Asset 链路完成：Position=1、History=1、10 USDT 保证金指令=1、Asset Flow=1、冻结 10→0、余额 100→90、Fill/Order 均为终态 3；同一 Kafka 事件额外重放两次后上述数量及金额完全不变 |
| 2026-07-28 | Trade Event Payload 版本默认值 | 已修复并验收 | 发现业务事件未显式赋值时生成模型会写入 `payload_version=0` 并违反 `chk_trade_event`；自定义模型入口统一归一为版本 1，保留显式更高版本，新增单测；修复镜像中 `POSITION_UPDATED`、`CONTRACT_FILL_ASSET_SETTLED`、`ORDER_SETTLED` 均以版本 1 成功投递 |
| 2026-07-28 | 成交结算终态恢复 | 已修复并验收 | 发现 Asset 指令先成功、Position 事件后确认时 Fill 会永久停在 PROCESSING；恢复任务新增“指令全部成功但 Fill 未收尾”扫描，重新核对 Position 事件后原子完成 Fill/Order 并发布结算事件；真实记录由 Fill=2/Order=11 自动恢复为 Fill=3/Order=3 |
| 2026-07-28 | P0-03 正/负/零资金费真实批次 | 通过 | 三个隔离线性永续标的分别锁定 +1%、-1%、0；正费率多头扣 1/平台入 1，负费率多头入 1/平台扣 1，零费率不生成 Asset 指令；三个 Batch/Settlement 均完成，非零指令绑定四条唯一 Asset Flow，仓位 `last_funding_time` 和 version 正确推进，用户与平台合计资金严格守恒 |
| 2026-07-28 | P0-03 INDEX 来源规范化 | 已修复并验收 | 真实任务发现完整三段来源被重复拼成 `crypto:BA:crypto:BA:BTCUSDT`；改为按三段/两段/单段显式继承维度并补单测，修复镜像成功读取 MARK、INDEX、FUNDING 不可变档案并完成三类批次 |
| 2026-07-28 | P0-03 平台差额账户缺失与恢复 | 通过 | 独立租户缺 USDT 差额账户时事务整体回滚，Batch/Settlement/Instruction=0，用户 100 与 Position last_funding_time=0 均不变，任务明确报错；补齐差额账户后复用原快照自动创建同周期批次，用户 100→99、平台 0→1，两条唯一 Flow 对账且 Position 时间/版本推进 |
| 2026-07-28 | P0-03 用户余额不足人工处理 | 通过 | 正费率长仓应付 15000、用户可用 99；Asset 无 Flow 且连续拒绝，第 20 次后用户 Instruction=5、Settlement=4、Batch=5，平台差额 Instruction 仍为 Pending，用户 99/平台 1 均不变，Position last_funding_time=0/version=1 |
| 2026-07-28 | P0-03 历史周期顺序补跑 | 通过 | 以 `T-3m` Completed Batch 为锚并提供 `T-2m/T-1m/T` 不可变快照；任务按创建时间 1785230408050→1785230410097→1785230411018 串行完成三个唯一 Batch，三个零费率 Settlement/History 的 business_time 严格递增，Position last_funding_time 最终等于 T |
| 2026-07-28 | P0-03 Trade 进程退出恢复 | 通过 | 真实停止 Trade 后写入 Pending Funding Batch、Settlement、用户 step1 与平台 step2，停机时余额保持 99/1；启动 Trade 后无需行情重算或人工重置即恢复，用户 99→98、平台 1→2，两条唯一 Flow 对账，Position last_funding_time 推进且 Batch=Completed |
| 2026-07-28 | P0-03 资金费与成交投影并发 | 通过 | Funding Settlement 先固化 qty=1/version=1/fee=-1，随后当前 Position 已加仓为 qty=2/version=2；放行 Saga 后仅扣旧事实 1，Position History 的 before/after qty 均为 2、version 2→3、realized_pnl_delta=-1，用户/平台 97/3 守恒且两条 Flow 唯一 |
| 2026-07-28 | P0-03 Asset 成功但 Trade 未确认 | 通过 | 使用正式 Instruction ID=93/BizNo 先直连 Asset 扣 1，确认用户余额 97→96、Flow=1，而 Trade Instruction/Settlement 仍 Pending；放开重试后同 BizNo 未二次扣款，用户 Flow 仍 1，平台 Flow=1，Position realized_pnl -1→-2、Batch/Settlement/Instruction 全部完成 |
| 2026-07-28 | P0-04 U 本位线性交割全链路 | 通过（该产品场景） | 隔离交割合约经 CLOSE_ONLY→MATCHING_STOPPED→PRICE_LOCKING→SETTLING→COMPLETED→ARCHIVED；最终快照 110 原样锁定，仓位保证金返还 10、手续费扣 1.1、盈利入账 10 按 step 1/2/3 执行，余额 90→108.9；Position 关闭并清零、Settlement/History 唯一、三条 Asset Flow 全部对账、`CONTRACT_SETTLED` 成功投递 |
| 2026-07-28 | P0-04 无持仓到期 | 通过 | 隔离交割合约到期时无 Position，唯一 Batch 以 `total_positions=0/settled_positions=0` 完成并归档；Settlement=0、Instruction=0，仍成功投递一个版本 1 的 `CONTRACT_SETTLED` |
| 2026-07-28 | P0-04 交割价缺失与恢复 | 通过（缺价子场景） | 到期标的无 DELIVERY 快照时 Symbol 仅进入 CLOSE_ONLY，Batch 停在 PRICE_LOCKING=3 并记录 `no valid market quote`，Settlement=0、归档事件=0；补入同交割时点最终快照后，无需人工重置即自动清除错误、锁价 100、完成并归档 |
| 2026-07-28 | P0-04 活动订单撤销与全额预占释放 | 通过（Pending 子场景） | 到期 Pending 订单先变为 CANCELED，撤销量 1；释放指令=10、Reservation=RELEASED、Asset Freeze=UNFROZEN、唯一 Flow=10、余额 98.9/冻结 10 恢复为可用 108.9/冻结 0；其后无持仓 Batch 才锁价并归档，订单与交割事件均成功投递 |
| 2026-07-28 | P0-04 部分成交订单剩余预占释放 | 通过 | 数量 1 的订单保留已成交 0.4 并撤销 0.6；Reservation 保留 consumed=4、仅 released=6，Asset Freeze 从 used=4/remain=6 变为 CLOSED、唯一释放 Flow=6，用户余额恢复且无冻结；唯一 Batch 随后以零存量仓位归档 |
| 2026-07-28 | P0-04 Trigger Waiting 与可平数量释放 | 通过 | 到期 Reduce Only 触发订单撤销 0.5，Contract Order 的 reserved_close_qty 0.5→0，Position 的 avail/frozen 0.5/0.5→1/0；缺少 DELIVERY 快照时 Batch 停在 PRICE_LOCKING 且 Settlement=0，证明撤单释放和锁价结算解耦 |
| 2026-07-28 | P0-04 零盈亏/零手续费/零 Asset 步骤 | 通过 | 第一场景 settlement_price=open_price 且 fee=0，只生成保证金返还 10，不生成费用或盈亏指令；第二场景 position_margin/pnl/fee 均为 0，Instruction=0，仍关闭 Position、写唯一结算 History、完成 Settlement/Batch 并成功投递事件 |
| 2026-07-28 | P0-04 用户余额不足人工处理 | 通过 | 线性交割亏损 200、用户可用 108.9；真实 Asset 连续拒绝且无 Flow，第 20 次失败后 Instruction=5、Settlement=4、Batch=7 并记录错误，Position 保持 DELIVERING qty=1、History=0、余额不变、归档事件=0 |
| 2026-07-28 | P0-04 Asset 中途重启恢复 | 通过 | 三步交割先在 Asset 停机时持久失败，恢复后保证金返还 10 成功；在费用/盈利执行前再次真实停止并启动 Asset，随后费用 1.1、盈利 10 自动续跑；三条指令/Flow 各唯一，余额 98.9→117.8，Position/Settlement/Batch/History/Event 均正确终态 |
| 2026-07-28 | P0-04 BTC 本位反向交割 | 通过 | 100 张×100 USD 的 LONG 从 50000 交割至 55000：锁价 55000，PnL=0.018181818182 BTC，Fee=0.001818181818181818 BTC；保证金 0.02、费用、盈利按 step 1/2/3 各形成唯一 BTC Flow，钱包 0.98→1.016363636363818182，Position/History/Settlement/Batch/Event 全部终态 |
| 2026-07-28 | P0-04 交割价窗口与审计元数据门禁 | 通过 | 30 秒窗口外快照被视为无有效报价；窗口内但 formula_version 为空时 Batch 持久化 `final DELIVERY snapshot has no formula version`；两阶段均保持 PRICE_LOCKING、Settlement/Event=0，随后补入时点内合法快照 101 后无需重置即清错、锁价并唯一归档 |
| 2026-07-28 | P0-04 Canceling 锁价门禁 | 通过 | 释放指令故意延后时 Order=9、Batch=2 且 settlement_price=0，Settlement/Event=0；放开后 Reservation=6、Freeze=UNFROZEN、唯一 Flow=10、Order=CANCELED，随后 Batch 才锁价 100 并归档 |
| 2026-07-28 | P0-04 Settlement Pending 锁价门禁 | 通过 | Fill 结算指令延后时 Order=11、Fill=2、Batch=2 且 price=0，Delivery Settlement/Event=0；资金步骤完成后恢复任务自动令 Fill/Order=3，才生成零金额交割 Settlement、关闭 Position、写 History 并归档；Position/Fill/Order/Contract 四类事件均成功 |
| 2026-07-28 | Reservation 终态计算与存量修复 | 已修复并验收 | 发现 MySQL 单表 UPDATE 左到右求值会令状态判断重复累加本次金额，导致全额消耗停在 PARTIAL、全额释放停在 RELEASING；状态表达式改为先计算再累加，新增幂等迁移修复存量记录，实库历史记录分别恢复为 CONSUMED=4、RELEASED=6 |
| 2026-07-28 | P0-06 核心跨服务对账实跑 | 通过（强平专项除外） | ORDER_FILL/FILL_POSITION_HISTORY/RESERVATION_ASSET_FREEZE/POSITION_MARGIN 四类循环游标持续完成；真实缺 History、缺 Flow、保证金托管差异跨周期累计，修正底层事实后经过稳定窗口自动 RESOLVED；Funding/Delivery 指令均以唯一 Flow 和 reconciled_at 门禁 Batch，当前非强平 OPEN issue=0 |
| 2026-07-28 | P0-02 BTC 本位反向永续仓位生命周期 | 通过 | 100 张×100 USD 于 50000 开多，60000 加仓 50 张后调和均价为 52941.176470588858；55000 减仓 60 张实现 +0.0042424242426 BTC 并释放 0.0113333333333333 BTC 保证金；50000 全平剩余 90 张实现 -0.0099999999999 BTC 并释放 0.017 BTC，最终 Position 数量/保证金/风险字段归零，Order/Fill/History/Instruction/Flow 全部终态 |
| 2026-07-28 | P0-02 线性 NET 反向开仓 | 通过 | 隔离 SHORT 2 张、均价 100、保证金 20；NET BUY 3 张@90 先全平 SHORT 并释放保证金 20、实现盈利 20，再仅按余量开 LONG 1 张并消费预占保证金 9；Fill realized_pnl=20，双 Position History、三条 Asset Flow、Reservation/Freeze、Outbox/Inbox 均唯一，钱包 80→111 |
| 2026-07-28 | Asset 业务流水号长度边界 | 发现缺陷后修复并验收 | 减仓生成的 `MARGIN_RELEASE` 业务号令 Asset `flow_no` 超过 VARCHAR(64)，真实指令失败；公共单号生成器对超长值保留前缀并附加稳定 SHA-256 摘要，短值不变。失败指令原地重试后两条 Flow 长度均不超过 64、业务号仍完整、无重复扣款/入账 |
| 2026-07-28 | Fill 已实现盈亏投影 | 发现缺陷后修复并验收 | 仓位历史和 Asset 盈亏正确但 `t_trade_fill.realized_pnl` 未回写；仓位投影现按同 Fill 的 History 汇总回写，首次执行和补偿重放共用同一路径。新全平 Fill 自动写入 -0.0099999999999，旧部分减仓补偿事件回填 +0.0042424242426，History/Flow 数量保持 1/2 不变 |
| 2026-07-28 | Order/Fill 均价精度对账 | 已修复并完成镜像复验 | MySQL 反向调和/加权除法产生约 10^-15 的尾差，被严格 Decimal Equal 误报；对账改为按交易对 `price_scale` 归一比较，保留数量、金额、费用的严格相等，单测同时覆盖尾差放行与一个最小价格单位的真实差异告警；修复镜像运行后历史 `ORDER_FILL:991101/991102` 均自动 RESOLVED，三个验收租户 OPEN issue=0 |
| 2026-07-28 | 衍生品仓位与资金跨事件顺序 | 发现竞态后修复并验收 | 真实撮合曾出现 `FILL_CREATED` 先扣手续费并释放余量、`POSITION_FILL_REQUIRED` 后创建保证金指令，导致保证金无法再从已关闭冻结记录扣除；衍生品 Fill 现在先幂等完成仓位投影并刷新指令，再严格按最小未完成步骤执行资金指令，恢复扫描同样受顺序门禁保护；修复镜像中保证金、手续费、余量释放全部一次成功且 retry=0 |
| 2026-07-28 | P0-02 线性永续限价价格改善 | 通过（全成子场景） | 隔离租户 `900106` 的老卖单 100 与新买单 101 真实撮合为 100；卖方 Maker 费 0.1、买方 Taker 费 0.2，各形成 1 张 100 均价、10 USDT 保证金的 SHORT/LONG；Reservation 分别消费 10.1/10.2 并释放 0.1/0.102，Asset Freeze 均 CLOSED、钱包 frozen=0，Order/Fill=终态 3，全部事件成功且无对账 OPEN issue |
| 2026-07-28 | P0-02 线性永续 IOC 部分成交 | 通过 | 隔离租户 `900108` 的 IOC BUY 2 张@101 对在簿 SELL 1 张@100，真实成交 1 张并撤销余量 1 张；Taker Order 为 CANCELED 且 `filled_qty=1/canceled_qty=1`，原因 `IOC residual`；只消费保证金 10 和手续费 0.2，剩余预占 10.404 全额解冻，Maker 侧也释放 0.1，双方 Position/Fill/Reservation/Instruction/Flow/Event 均正确终态 |
| 2026-07-28 | P0-02 线性永续市价单 | 通过 | 隔离租户 `900110` 的 MARKET BUY 委托价为 0，对在簿 LIMIT SELL 100 真实以 100 成交；市价方 Taker 费 0.2、限价方 Maker 费 0.1，双方各形成 1 张均价 100、保证金 10 的仓位；买方预占 10.2 全部消费，卖方消费 10.1 并释放 0.1，钱包冻结均归零，Order/Fill/Event 全部终态 |
| 2026-07-28 | P0-02 线性永续 Reduce Only | 通过 | 隔离租户 `900112` 预置 LONG 1 张、可平冻结 1、仓位保证金 10；Reduce Only SELL 成交后仅生成 LONG 的关闭历史，数量/可平/冻结/保证金全部归零并返还保证金 10，手续费 0.2 从预占消费，钱包终态 99.8；未创建 SHORT 仓位，对手方正常开 LONG，双方 Order/Fill/Instruction/Flow/Event 全部终态 |
| 2026-07-28 | 条件单最新成交价扫描 | 发现缺陷后修复并验收 | `FindLastPrice` 曾把 `decimal.Decimal` 直接交给 go-zero 严格 ORM，单列查询必然报 `not matching destination to scan`，使所有条件单停在等待态；改为带 `db:"price"` 的单字段结果结构。修复镜像中条件 BUY 由最新价 101 触发并以 Maker 价 100 成交，触发、Fill、Position、Asset 和结算事件全部终态 |
| 2026-07-28 | 条件单触发时间审计 | 发现缺口后修复并验收 | 触发流程原先只写 `biz_ext.triggeredAt`，专用 `t_trade_order.triggered_at` 保持 0；现同步写入专用列并有状态转换单测。无对手条件 MARKET 单真实触发后两处时间一致且非零，随后以 `market order has no executable liquidity` 自动撤销，预占 10.2 全额解冻、钱包冻结归零 |
| 2026-07-28 | 条件单 MARK/INDEX 价格源 | 发现语义缺口后修复并验收 | 触发任务原先忽略 `trigger_type` 并统一读取最新 Fill；现按 LAST/MARK/INDEX 分流，MARK/INDEX 只读取 30 秒内、已确认、对应 `snapshot_kind` 的租户优先快照，缺源保持等待且不兜底。真实 MARK=101、INDEX=102 分别触发两张条件单，审计来源为 `mark_price/index_price`，触发时间非零；无流动性撤单后各自预占 10.2 全额释放、钱包冻结归零 |
| 2026-07-28 | P0-02 线性永续独立加仓与部分减仓 | 通过 | 隔离租户 `900122` 预置完整托管事实的 LONG 1@100/保证金 10；于 110 加仓 1 后数量 2、加权均价 105、保证金 21、维持保证金 1.1；随后 Reduce Only 于 120 减仓 1，Fill/History 实现盈利 15，按比例释放保证金 10.5，剩余 LONG 1@105/保证金 10.5，钱包最终 103.84；两轮 Maker/Taker 费用、Reservation/Freeze、Instruction/Flow/Event 均唯一且 retry=0 |
| 2026-07-28 | P0-02 撤单与成交并发 | 通过 | 隔离租户 `900126` 对 Taker BUY 订单持有外部 `FOR UPDATE` 行锁，启动 Matcher 并同时发起 Cancel RPC 后释放；撮合获得锁并令双方 Order=`FILLED_AND_SETTLED`、各 Fill=1，晚到撤单不新增 Cancel Log。双方 Position/History 各 1，Reservation/Freeze 精确消费 10.2/10.1，钱包 frozen=0；Event=13、Instruction=4、Asset Flow=4 均唯一终态，数量超额和 OPEN 对账差异均为 0 |
| 2026-07-28 | P0-02 重复 Fill 双键幂等 | 发现校验缺口后修复并验收 | 成交入口调整为先锁订单并补齐业务身份，再同时检查 `(tenant_id, fill_no)` 与 `(tenant_id, match_no, order_id)`；完全相同重投直接返回旧 Fill，订单与 Fill 均零写入，任何价格/数量/金额/费用或业务身份冲突均拒绝。单测覆盖终态 Fill 的相同重投、篡改数量和换成交号；实库相同 `fill_no` 与相同 `match_no+order_id` 两种 INSERT 分别命中 `uk_tenant_fill_no`、`uk_tenant_match_order`，之后 Fill/History/Instruction/Flow 仍为 2/2/4/4，未完成项及 OPEN issue=0 |
| 2026-07-28 | P0-02 交割 U 本位下单到仓位 | 通过 | 隔离租户 `900128` 通过真实 App RPC 提交 SELL 1@100 与 BUY 1@101，Asset 分别冻结 10.2/10.302，Matcher 以 Maker 100 成交；形成 SHORT/LONG 各 1、保证金各 10、Maker/Taker 费 0.1/0.2，余量 0.1/0.102 释放后钱包 frozen=0。独立租户 `900130` 再以 IOC BUY 2@101 对 SELL 1@100，订单精确 `filled=1/canceled=1`、仅消费 10.2 并释放 10.404；两场 Order/Fill/Position/History/Reservation/Instruction/Flow/Outbox/Inbox 均终态且 OPEN=0 |
| 2026-07-28 | P0-02 交割币本位下单到仓位 | 发现缺陷后修复并验收 | 首次真实 App BUY 100@51000 吃 SELL 100@50000 时发现反向合约按较高委托价只冻结约 0.02 BTC，低于成交所需 0.0204，导致 Fill 重试。修复为反向 BUY 限价单扫描已在簿可成交卖单并用最低 Maker 价计算预占，非交叉挂单仍用自身价格；单测覆盖休眠挂单、多个交叉价、Market/非交叉卖单。新租户 `900134` 买单风险价=50000、Reservation=0.0204，成交形成 LONG/SHORT 各 100 张、Entry=50000、Margin=0.02、Maintenance=0.001，Taker/Maker 费 0.0004/0.0002，全部终态且 retry=0、OPEN=0；原失败 Saga 补齐正确冻结事实后也自动恢复 |
| 2026-07-28 | P0-01 双实例任务锁 | 通过 | 将验收环境临时扩为 `wklive-trade-rpc-1/2`，两个实例同时直连调用 tenant=0 的 `ProcessPositions`。首轮返回码严格为 200/2004、耗时 86ms/16ms；后续 10 轮共得到 10 次 OK 与 10 次 `SyncTaskAlreadyRunning`，且每一轮恰好各一，两个实例均曾获得执行权。未出现双执行或锁残留，随后缩回单实例且健康 |
| 2026-07-28 | P0-07 Settlement 双 Worker CAS 与租约 fencing | 通过 | 插入无业务外键的隔离 Pending 指令 ID=195，两个 Trade 容器以同一 `now=1785237508453` 并发调用真实模型 `ClaimLease`，结果严格为 false/true；把模型时间推进 60001ms 后另一 Worker 成功以 lease `1785237568454` 重领。事务内按旧 lease 完成影响 0 行、按新 lease 完成影响 1 行，随后回滚；隔离指令精确删除、缓存清理并恢复单实例 |
| 2026-07-28 | P0-07 Position 标记价/成交并发 | 通过 | 隔离 Position 初始 qty=1/margin=10/version=1；事务持行锁后先写入成交态 qty=2/margin=20/version=2，同时启动旧 version=1 的真实 `UpdateMarkRiskCAS`。成交提交后旧 Worker 的 SQL 等待 5367ms 并影响 0 行，Position 保持成交事实及原 Mark；用 version=2 再执行则影响 1 行，只把 mark=111、maintenance=1.11、unrealized=22 等派生字段写入并推进 version=3，qty/avail/margin 仍为 2/2/20。测试记录随后精确删除 |
| 2026-07-28 | P0-07 Price Engine 缺输入 | 发现可观测性缺陷后修复并通过 | 部署新引擎后插入独立公式 `ACCEPT-MISSING-INPUT-MARK-v1`，组件指向不存在的 `missing-feed/FINAL_QUOTE/crypto/NOFEED/MISSINGUSDT`。引擎 run_version 连续推进至 220、last_target_time 始终回滚为 0，目标快照严格为 0；首次发现生产模式过滤 Info 导致容器无诊断日志，修复为 30 秒节流的 Error 后日志完整输出公式及六个组件维度，25 秒窗口仅 1 条。补入 123.45 来源后自动连续产出 13 个确定性快照且 Outbox 全为 Success；随即停用并精确清除 1 个公式、14 个快照、13 个 Outbox 与 2 个 Redis 缓存键 |
| 2026-07-28 | P0-07 Snapshot Outbox 容量 | 通过 | 停止 Itick 后一次性注入 50,000 条带唯一前缀的完整 Snapshot+Quote 消息，旧化 create_times 后启动即输出 pending=50000、oldest age 和 ETA，期间无慢查询/Worker 错误。32 Worker 的 30 秒样本为 `30.93/s`、processing=32；将源码及 Etcd 配置持久化为 64 后，同负载样本为 `61.87/s`、processing=64，证实吞吐约翻倍且 processing 只是实时在途数。为控制时长裁剪未处理尾部后保留 11,536 条完整排空，其中 9,616 条由 64 Worker 在 158.116 秒完成；最终全部 status=3、retry=0、Redis/Kafka checkpoint 非零、last_error 为空。精确删除 11,536 条测试 Outbox 和两个 Redis Key；Kafka 测试消息由隔离 Topic 保留策略自然回收 |
| 2026-07-28 | P0-07 Kafka 四方向重复投递 | 通过 | 从真实终态事实中选择永续 U 本位 Fill=990101、永续币本位 Fill=991100、交割 U 本位 Fill=991711、交割币本位 Fill=991717；重放前分别对 Order、Fill、当前 Position、Position History、Settlement Instruction、Asset Flow、Inbox 生成指纹 `85ee…/4edb…/0d6b…/600a…`。向实际 Topic `trade.domain-events` 对每个原 `POSITION_FILL_REQUIRED` 再投递两次，共 8 条，`trade-realtime` 消费组 offset=295、lag=0；重放后四个指纹完全一致，History 均 1、Instruction/Flow 分别 1/1/2/2、Inbox 均 1，无重复仓位或资金事实 |
| 2026-07-28 | P0-07 故障门禁总验收 | 通过 | Redis 获取锁/持锁续租失败、Kafka 停机恢复及四方向重放、Asset RPC 超时与成功未确认、真实合约事实 MySQL 1213、Trade 提交后退出、Position CAS、Worker 租约、双实例领取、服务重启、Price Engine 缺输入、Snapshot Outbox 5 万积压全部取得运行态证据；隔离数据均已清理，核心服务全部 healthy，Outbox OPEN=0 |
| 2026-07-28 | P0-05 DELIVERY 三源门槛与确定性回放 | 技术验收通过，生产源待配置 | 激活中位数公式时先提供 100/101/150 三个独立来源；200 BPS 剔除 150 后仅剩 2 源，run_version 持续推进但 DELIVERY=0，节流日志明确 `accepted=2 required=3 rejected=1`。刷新来源并补第 4 源 99 后无需重启产出 3 个价格均为 100 的快照，raw_payload 每条均为 all=4/accepted=3/rejected=1、公式版本一致，Outbox 全 Success。`cmd/price-replay` 对 ID=77 的实库审计输出 `replay verified: price=100`；将审计副本 output_price 改 101 后明确拒绝。隔离的 1 个公式、10 个快照、3 个 Outbox、2 个 Redis Key 已清理 |
| 2026-07-28 | P0-06 对账告警节流 | 发现刷屏后修复并通过 | 原 `LIQUIDATION:1` 每 3～5 秒重复输出；新增 `last_alert_at` 并在数据库行锁事务内同时累计发现次数和预占告警窗口，多实例只允许首次、重开、内容变化或 30 分钟提醒输出。部署后 67 秒内 occurrence_count 从 2522 增至 2535，last_seen 持续变化而 last_alert_at 固定，新容器同 key 严格只有 1 条日志 |
| 2026-07-28 | P0-06 人工忽略真实 RPC | 发现生成模型参数遗漏后修复并通过 | 首次真实调用发现新增 `last_alert_at` 后生成模型 Update 有 20 个占位符但仅 19 个参数；补齐 Insert/Update 字段映射并重新部署。随后租户隔离列表、缺操作人拒绝、有效忽略、重复忽略拒绝全部通过；操作人、原因、解决时间与后台来源审计事件完整落库并成功投递 |
| 2026-07-28 | P0-04 缺价负向夹具收口 | 通过 | 四个已证明“缺最终 DELIVERY 快照时失败关闭”的 `ACCEPT-DLV-LINEAR`、`ACCEPT-DLV-LINEAR-IOC`、`ACCEPT-BTC-USD-DLV-RPC/FIX` 批次均保持 settlement_price=0、Settlement/Instruction=0，并精确转入 MANUAL_REVIEW=7；错误原因追加验收处置说明、模型缓存键清除。收口后非终态 Delivery Batch=0，30 秒窗口 DELIVERY 重试日志=0，不伪造行情、不继续自动结算 |
| 2026-07-28 | MARK 缺源按标的合并与退避 | 发现验收持仓刷屏后修复并通过 | 32 个开放 Position 分属 21 个租户/Symbol；修复前每个仓位都调用归档并输出错误，修复后同一扫描共享一次 MARK 查询结果，失败键按 tenant/symbol/source 退避 30 秒。新镜像首轮严格每个 Symbol 1 条，共 21 条；三仓位 Symbol=991800 的连续重试间隔为 34.197 秒和 30.074 秒，仓位任务仍持续执行且无无审计价格兜底 |
| 2026-07-28 | P1-02 保险基金承接与冲正矩阵 | 发现审计缺口后修复并通过 | 隔离平台账户经真实 Asset RPC 完成全额 30/30、部分 20/50、余额不足 0/10、同号幂等、30 冲正及不同冲正号拒绝；账户、Cover、Idempotent、Platform Flow 数量和金额守恒。验收发现 Proto 已有保险基金枚举但字符串映射缺失，导致 `biz_type/scene_type` 为空；补齐双向映射、单测及精确历史回填，部署镜像 `bc6b4b15…` 后重建事实，4 条幂等记录和 3 条资金流水审计维度全部非空且各唯一 |
| 2026-07-28 | P1-02 全额强平与手续费 | 发现资金缺口后修复并通过 | 原实现只在 Trade 记录手续费且把手续费加入穿仓缺口；新增 Asset `CreditPlatformRevenue`，以平台流水唯一键原子幂等计入 `FEE_REVENUE`，同号重放不重复、改金额拒绝。强平改用冻结的 Trigger Qty/Mark/Equity 计算，只从正权益收取手续费，保险/ADL 仅承接真实负权益；LIQUIDATING 仓位不再被后续 MARK 扫描改写。隔离 LONG 1@100、MARK 95、保证金 10 实跑：权益 5、手续费 0.95、用户残值 4.05、仓位 1→0，Liquidation/History/Event/User Flow/Platform Flow 各 1；重放数量不变。Trade 镜像 `b633c5a1…`、Asset 镜像 `2669ca2b…`，Etcd 生产门禁始终为 false；夹具和临时程序已精确清理 |
| 2026-07-28 | P1-02 强平前活动订单撤销 | 发现 Reduce-Only 遗留风险后修复并通过 | 原实现只撤增加风险的订单，全部接管后 Reduce-Only 订单及可平数量预占会遗留；全部接管现先撤销该风险单元所有活动订单。隔离实跑同时放置普通 BUY 1（Asset 冻结 10）和 Reduce-Only SELL 0.25（Position frozen_qty 0.25）：普通单经唯一释放指令/Asset Flow 解冻 10，Reservation=RELEASED、Freeze=UNFROZEN；Reduce-Only 的 `reserved_close_qty` 与 Position frozen_qty 均归零；两单均 CANCELED、原因 `risk liquidation`、事件各唯一，随后仓位才 1→0。重放后 Order/Instruction/Liquidation/History/Event/Asset Flow 数量保持 2/1/1/1/3/2；Trade 镜像 `c1e3a2d2…`，夹具和缓存已清理 |
| 2026-07-28 | P1-02 标记价格驱动风险重算 | 发现写放大后修复并通过 | 隔离租户真实写入 `price-engine/MARK` 权威快照并调用 Trade `ProcessPositions`：线性逐仓 LONG 2@100、保证金 40、MARK 110 后，快照 ID 固定为 `accept-risk-mark-110-v2`，未实现盈亏=20、风险档维持保证金=3.4、破产价=80、强平价=81.1224489795918367、风险率=0.0566666667，仓位保持 NORMAL 且 Liquidation=0；Trade 同时落下唯一不可变 Market Snapshot。实跑发现旧实现对同一投影每次扫描都递增 Position version，增加无效写入和 CAS 冲突；新增完整风险投影相等判断后，同一 MARK 连续重放版本稳定为 11。单测通过，Trade 镜像 `582a8c27…` |
| 2026-07-28 | P1-02 ADL 与父 Liquidation Saga | 连续发现边界缺陷后修复并通过 | 补齐基于“盈利率 × 有效杠杆”的持久化 ADL 优先级，并排除尚未具备账户级权益的全仓候选；将 Asset 成功独立持久化为 ASSET_DONE，再单独提交 Position 投影。首次实跑在 Asset 已入账后发现新 Execution 未回填自增 ID，修复后从真实 PREPARED/PROCESSING 故障点幂等恢复，40 USDT 流水始终唯一；随后在 ASSET_DONE 恢复零仓位时发现正数量前置条件，补齐零仓风险清零并再次恢复成功。最终无故障矩阵中，被接管 LONG 2 的缺口 15 由两个逐仓 SHORT 各减 1、各缓释 7.5；更高优先级 Position=99920603（ID 更大）先执行，证明排序生效；超高 rank 的全仓候选保持原样。候选资产分别入账 40/50，Execution/Instruction/Asset Flow=2/2/2，部分减仓候选的 Qty 2→1、Margin 40→20、UPnL 75→37.5 及风险字段同步重算；父 Liquidation 的 trigger/liquidated/adl qty 均为 2，破产仓位及所有风险字段清零，History=3、Completed Event=1。重放数量不变；Trade 恢复后事件已投递、两条指令均取得唯一 Asset Flow 并完成对账。最终镜像 `9e1abc44…`，生产门禁仍为 false |
| 2026-07-28 | P1-02 部分强平与专项对账 | 发现终态恢复缺口后修复并通过 | 按下一风险档名义价值上限和数量步长计算减仓量，仅在减仓后权益高于维持保证金、风险率低于 1 且 MARK 脱离强平线时采用部分强平，否则回退全部接管。隔离 LONG 10@100、MARK 96 从第二档减至第一档：Qty 10→5、Margin 100→50、UPnL -40→-20、Realized PnL=-20、Maintenance=4.8、Risk=0.16、Liquidation Price=90.9090909090909091，仓位恢复 NORMAL；用户残值 25.2、平台 `FEE_REVENUE` 4.8，Liquidation=PARTIAL_RECOVERED、History/Event/User Flow/Platform Flow 各 1，保险与 ADL 均为 0。重启及安全仓位重放后 Position version=7、五类事实数量均不变。修复 PARTIAL_RECOVERED 不再被恢复查询当作活动态，并使专项对账能够自动解决执行中遗留差异；真实全量循环游标完成、OPEN issue=0。最终 Trade 镜像 `acc582db…`，生产门禁保持 false |
| 2026-07-28 | P2-01 全仓账户级风险投影 | 第一阶段通过，交易门禁保持关闭 | 扩展 `t_contract_margin_snapshot` 固化账户维持保证金、权益、可用保证金、风险率、仓位数及 Asset 版本；`ProcessPositions` 按租户/用户/保证金币种聚合开放全仓仓位和未结订单 Reservation，经 Asset Admin RPC 读取合约钱包，输入未变化时不递增快照版本。隔离租户 910208 真实预置同一 USDT 风险单元的两个全仓仓位及一个 30 USDT Pending 订单：Wallet=100/Available=70/Frozen=30、Position Margin=30、Order Margin=30、UPnL=-20、Maintenance=11，得到 Equity=110、Available Margin=50、Risk=0.1、Position Count=2、Asset Version=5；重复任务后 snapshot version 保持 1，单仓 liquidation/bankruptcy/risk/adl 字段继续为 0。新增字段通过正式 Admin RPC 返回；另修复配置假启用入口，真实 RPC 分别以“账户级强平未启用”和“存在仓位或活动订单”拒绝 CROSS 配置及模式切换。最终 Trade 镜像 `dad9722e…`；`CROSS` 下单拒绝未移除 |
| 2026-07-28 | P2-01 全仓正权益账户强平 | 第二阶段通过，负权益路径仍关闭 | 新增 `t_contract_account_liquidation` 父 Saga 和明细表；扫描账户风险率不低于 1 的全仓快照，自动强平开启时先撤同保证金币种的活动订单并释放 Reservation，再在事务内锁定风险单元、重算风险并接管全部仓位。用户正权益按账户净额生成唯一 Asset 指令，手续费使用独立持久化指令、指数退避和平台流水幂等入账；关闭总开关后，已经进入资金/关仓阶段的 Saga 仍可恢复。隔离租户 910212 双仓实跑：用户钱包 100→118、平台 `FEE_REVENUE` 0→2、两仓和两明细均关闭、History=2；平台账户缺失时手续费指令持续退避且用户流水不重复，补齐账户并将开关恢复 false 后父 Saga 自动完成。迁移版本 42，最终 Trade 镜像 `4a89ad36…`，生产开关为 false。负权益账户当前明确转人工，待账户级保险基金/ADL和后台处置闭环 |
| 2026-07-28 | P2-01 全仓账户强平专项对账 | 通过 | 新增 `CROSS_ACCOUNT_LIQUIDATION` 全量循环扫描，按父 Saga 校验明细数及终态、仓位清零、唯一强平历史、唯一完成事件、父子金额守恒、用户净额 Asset 指令和平台手续费指令；活动 Saga 延后，人工状态形成稳定差异。隔离租户 910213 故意缺完成事件后，稳定键 `CROSS_ACCOUNT_LIQUIDATION:99121301` 连续累计发现；补齐事件后同一差异自动转 RESOLVED，父/明细/仓位/历史/事件/指令全部一致。Trade 镜像 `ba49f5b7…`，生产自动强平门禁仍为 false，验收夹具已清零 |
| 2026-07-28 | P2-01 全仓账户强平后台处置 | 通过 | 新增 Trade/Admin API 的账户强平列表、详情和人工重试，详情同时返回父 Saga、接管仓位明细和结算指令；管理端增加独立菜单、详情按钮、明细/指令弹窗及中英文状态。重试只接受 `MANUAL_REVIEW`，要求可信后台操作人和原因，只重置失败资金指令并恢复正确 Saga 阶段，不能修改金额或直接标记完成；总开关关闭时强制拒绝。隔离租户 910214 真实 RPC 验收：列表/详情均返回预期人工记录，重试返回 `automatic liquidation gate is disabled` 且事实未变。菜单 1195～1197 及角色授权已由幂等迁移落库；最终 Trade 镜像 `9d218f65…`，夹具和临时客户端已清零 |
| 2026-07-29 | P2-01 全仓负权益保险基金与 ADL | 通过 | 父 Saga 新增穿仓缺口、保险承接、ADL 缓释及数量检查点，明细固化缺口分摊、合成破产价和 ADL 结果；保险使用 `symbol_id=0` 的保证金币种默认账户，余额不足时仅从同租户、同标的、反方向的盈利逐仓仓位按既有 ADL 排序接管。隔离租户 910215：Wallet 10 + Margin 10 + UPNL -30，缺口 10 全由保险承接，用户钱包 10→0、基金 100→90；租户 910216：Wallet 20 + Margin 20 + UPNL -60，缺口 20 由保险 5 + ADL 15 覆盖，合成破产价 77.5，两个 SHORT 候选各接管 1、各缓释 7.5、各入账 32.5。两父批次、两明细、四仓位、两保险 Cover、两 ADL Execution、四 Asset 指令/流水均唯一成功；重启后数量不变且指令全部取得唯一 Flow 并完成对账。最终 Trade 镜像 `f3edb89d…` 健康，迁移校验和已登记，Etcd 门禁已恢复并复核为 false |
| 2026-07-29 | P2-01 全仓成交资金闭环与生产双门禁 | 通过 | 新增 `CrossMarginTrading.Enabled`，增加全仓风险敞口同时要求该开关和 `AutomaticLiquidation.Enabled` 为 true；标准 YAML 与最终 Etcd 均为 false。开仓在用户级续租锁内同步校验未完成账户强平、既有全仓订单、Asset 可用余额、跨仓位保证金/未实现盈亏/维持保证金及 30 秒内不可变 MARK。隔离租户 910217 完成双边全仓开仓并形成各 10 USDT 保证金、0.2 手续费；910219 完成买方 2 张中成交 1、撤销 1，精确消费/释放各 10.2；模式不匹配和无新鲜 MARK 均零写入拒绝。关闸后四个用户仍以 Reduce Only 全部平仓，仓位数量/保证金/冻结归零，钱包均为 99.6；重启前后 Order/Fill/History/Instruction/Flow 数量和金额完全不变 |
| 2026-07-29 | P2-01 仓位模式持久化与切换并发 | 发现缺陷后修复并通过 | 真实追加单向 NET 开仓时发现成交投影会把仓位实际方向保存为 LONG/SHORT，旧切换逻辑据此误判为双向模式。新增 migration `20260729_add_contract_position_mode.sql`，为当前仓位持久化 `position_mode`；另以 `20260729_fix_contract_position_mode_backfill.sql` 按最近一条可关联订单的 History 回填，避免资金费/交割/强平 History 无 `ref_order_id` 时误留默认值。订单/Fill 的 NET 映射为 ONE_WAY，显式 LONG/SHORT 映射为 HEDGE，开仓复用关闭仓位时允许受控改模，活动仓位及平仓必须严格匹配。协议、App/Admin API 同步返回该字段。实库四个既有单向全仓仓位均回填为 1；全库最终 ONE_WAY=15、HEDGE=26、非法值=0、最近订单历史不一致=0。修复后追加开仓越过模式校验并正确在过期 MARK 门禁处拒绝，证明不再误判 |
| 2026-07-29 | P0-03 资金费按标的故障隔离 | 发现缺陷后修复并通过 | 历史验收标的的非法 `funding_rate_source=acceptance` 曾令 tenant=0 的全局 `ProcessContractSettlements` 整批失败并由任务消息反复重试。现在权威输入缺失或配置非法只阻断该租户/标的的 Funding Batch，按租户、标的、结算时点 30 秒节流告警并继续扫描后续标的；新镜像中只出现 `skip funding batch` 诊断，不再出现全局 task failed |
| 2026-07-29 | P2-01 双标的 MARK/Asset 并发投影 | 通过 | 隔离租户 910221 的同一 USDT 全仓风险单元包含 LONG 1@100 和 SHORT 2@200。两标的 MARK 更新为 120/170 时，仓位 UPNL 分别为 20/60、维持保证金为 6/17；并发把 Asset 钱包从 100/version1 更新为 120/version2 后，账户快照收敛为 Position Margin=50、UPNL=80、Maintenance=23、Equity=250、Available Margin=200、Risk=0.092、Position Count=2、Asset Version=2。`source_event_no` 与 `tenant/user/USDT/assetVersion2/positionVersionSum8/orderVersionSum0` 的 SHA-256 前 24 字节完全一致；重放后 Position version=4、Snapshot version=4 均不增加。夹具、行情、Trade 快照、缓存和临时客户端已清理，System 恢复健康 |
| 2026-07-29 | P1-01 Snapshot Outbox 实时水位 | 通过 | Itick Etcd 配置为 WorkerCount=64、BatchSize=512、Idle=100ms；此前 50,000 条容量矩阵完成后，当前实库 total=0/open=0，近 10 分钟无 `snapshot outbox unhealthy`、Worker 失败或 Outbox 慢查询日志 |
| 2026-07-29 | 备份与灾备恢复演练 | 隔离环境通过，生产演练待执行 | 源库 28.23 MiB、140 张表；静默点 `--single-transaction` 备份 11,975,203 字节/0.50 秒，SHA-256=`1cdd5acc…46470d4`，隔离恢复耗时 2.68 秒。140 张表逐表精确 COUNT 差异=0，46 条迁移、40 个订单、33 个成交、41 个仓位、109 条指令、216 个事件均一致；临时恢复库和文件已销毁，System 恢复 healthy。恢复顺序、事实边界和通过标准见 `perpetual-delivery-disaster-recovery-runbook.md` |
| 2026-07-29 | 本轮全仓收口回归 | 通过 | `services/trade go test ./...` 全部通过；helpers/admin/app/task/tasks/models 的 `go test -race` 全部通过；Admin API 全量测试通过，App API 业务包编译通过（根包网络测试需监听 `:7777`）。迁移由 db-init 正式执行并登记；最终 Trade 镜像 `440d4984…` 健康，两个生产安全开关均为 false |
| 2026-07-29 | P1-01 历史窗口批量确定性回放 | 工具与自动化测试完成，待生产数据执行 | `cmd/price-replay` 从单条扩展为单文件/多文件、JSON 数组及 JSONL 批量回放；`--interval-ms` 按公式版本拒绝目标时点未对齐、重复和断档，`--json` 输出记录数、公式数、时点/价格范围、最少有效输入与剔除数。单条回放同时新增完整输入与采用/剔除分区、`min_input_count` 防篡改校验；Price Engine 与命令测试通过。该能力不替代第 6 节要求的真实三源交割窗口数据 |
| 2026-07-28 | 本轮静态回归 | 通过 | 当前工作树重新执行 `services/trade`、`services/asset`、`services/itick`、`services/system` 的 `go test ./...` 全部通过；Trade 的 `internal/logic/task` 与 `models` 额外通过 `go test -race`，MARK 退避和对账告警判断均有单测；协议向后兼容编译通过 |

## 6. 当前外部依赖与不可代填项

| 门禁 | 当前事实 | 需要提供或确认 |
| --- | --- | --- |
| P0-05/P1-01 生产价格源 | `t_itick_price_formula` 当前无生产公式；归档中只有 `price-engine` 生成的验收快照，没有可持续的交易所原始 `FINAL_QUOTE`；BTCUSDT 也只有 16:32、16:39 两组历史 MARK/INDEX/FUNDING | 至少三个真正独立的现货/指数来源标识、市场代码、Symbol 映射、接入凭据和数据许可 |
| 生产 DELIVERY 参数 | 三源门槛、偏差剔除和离线重放技术验收已通过，但不能用虚构来源替代生产市场 | 确认算法（MEDIAN 或 WEIGHTED_MEAN）、每源权重、最大偏差 BPS、锁价窗口和公式版本命名 |
| 生产历史回放 | 当前没有覆盖交割窗口的三个原始来源历史，无法形成有意义的生产回放报告 | 提供选定合约、至少一个完整交割窗口的原始快照导出，或授权接入相应历史数据源 |
| P0-06 告警渠道 | 应用已输出结构化差异和 Outbox/Price Engine 健康日志；同一对账差异已按数据库事实节流为首次/变化/重开/30 分钟提醒，隔离环境仍无生产告警平台 | 确认告警平台、规则阈值、值班组、升级链路和通知渠道 |
| P1-02/P2-01 自动强平资金权限 | 逐仓及全仓正/负权益隔离矩阵已完成，但生产总开关仍保持关闭；租户 900102 的 `LIQUIDATION:1` 是保留的人工处理夹具 | 明确生产保险基金和 `FEE_REVENUE` 账户、资金权限、审批责任人、启用窗口和回滚方案后，方可申请开启 |
| 生产备份与灾备 | 隔离环境已完成全库一致性备份、140 表恢复核对和恢复手册；本地容量与恢复流程可执行 | 确认生产 RPO/RTO、备份加密及异地保留策略，并安排 Binlog 时间点恢复、节点/可用区故障切换和回切演练 |
