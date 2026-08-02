# OPT-P2-002 / OPT-P2-003 期权生命周期仓库级验收记录

状态：`REPOSITORY_PASSED / PREPROD_BLOCKED`

范围：最后交易、行权/指令截止、到期、交割独立时间，以及现金结算 AUTO、DNE、相反指令。
本记录证明仓库代码和隔离环境基线，不代表目标市场规则、生产行情、通知、调度或清算已经批准。

## 1. 审查结论

旧设计只有 `list_time`、`exercise_cutoff_time`、`expire_time`、`deliver_time`，其中
`expire_time` 同时承担停止交易和到期语义。这样无法表达“已停止交易、但用户仍可在截止前维护
行权指令”的标准窗口，也会让撤单、结算价锁定和 AUTO/DNE 清算发生在同一模糊边界。

现设计新增独立 `last_trade_time`，统一不变量为：

```text
list_time < last_trade_time <= exercise_cutoff_time <= expire_time <= deliver_time
```

- `last_trade_time`：停止普通单/组合单交易，合约进入 `PAUSED`，活动订单以
  `CONTRACT_LAST_TRADE_ENDED` 幂等撤销并释放资金；不得提前锁结算价或产生到期清算。
- `exercise_cutoff_time`：截止前 `TRADING/PAUSED` 合约仍可提交 AUTO、DNE、相反指令；边界及以后
  只允许查询或相同请求幂等重放。
- `expire_time`：合约才进入 `EXPIRED`，锁定获批结算价并执行 AUTO/DNE/FIFO 到期分配。
- `deliver_time`：完成现金或实物交割终态，不再与停止交易混用。

复用 `PAUSED` 是 V1 的受控选择：人工停牌和“最后交易已结束”共享状态，但恢复接口在
`now >= last_trade_time` 时故障关闭，运营端必须结合时间和 halt 记录判断原因。若未来产品需要显式
`TRADING_CLOSED` 状态，应以新枚举和迁移演进，不能覆盖历史语义。

## 2. 仓库实现

- Schema、系列 expiry、Proto、Admin API 和 Admin UI 均显式暴露 `last_trade_time/lastTradeTime`；
  服务和数据库同时校验五个时间点。迁移
  `migrations/20260802_option_last_trade_time.sql` 可重复执行，回填旧数据并为旧内部写入提供
  “缺省为行权截止”的上线兼容触发器；新 RPC/UI 必须显式提交该字段。
- DDL 变更后已执行 `make gen-model`，生成的合约及系列 expiry model 均包含 `LastTradeTime`；
  Proto 执行 `make gen`，Admin API 执行 `make gen`。以后修改字段或 DDL 仍必须重跑相应生成命令。
- 延迟队列新增独立 close-trading 动作；周期任务作为补偿扫描。两条路径共用行锁、预期时间身份和
  `closeContractTrading`，重复、迟到或旧消息不会提前到期，撤单失败会由后续扫描继续收敛。
- 普通下单、撮合、组合冻结/撮合、恢复交易、保险库存退出和资产执行前复核全部以
  `last_trade_time` 阻止新增交易；上市门禁同时要求时间顺序合法且最后交易尚未到达。
- 到期任务先确保交易已关闭，再由 `PAUSED -> EXPIRED -> SETTLED` 推进；AUTO 只执行达到阈值且
  在途撤单/到期释放通过不安全订单屏障保持 `PAUSED`，直至资金释放终态；AUTO 只执行达到阈值且
  扣费后净收益为正的数量，DNE 为零，相反指令不能绕过正净收益，空头按实际行权量 FIFO 分配。

## 3. 自动验收证据

```text
make gen-model                                                        PASS
services/option: make gen                                             PASS
admin-api: make gen                                                   PASS
services/option: go test ./...                                        PASS
admin-ui: npm run type-check（Node 20.20.2）                            PASS
services/option/acceptance/run-p0-asset-rpc-e2e.sh                    PASS
```

最近正式隔离 MySQL 8.4、Redis、真实 Asset gRPC 全套验收结果：

- 最终复跑主测试 `118.520s`；总计 9277 条资产指令，9270 条成功且已对账，7 条冻结前合法取消，
  `success + 2*canceled = 9284`。
- `P0-LAST-TRADE-INDEPENDENT` 使用五个不同时间点。在精确最后交易边界，普通活动单变为
  `CANCELED/CONTRACT_LAST_TRADE_ENDED`，合约保持 `PAUSED`，无提前行权、结算或交割；
  最后交易后、行权截止前仍可提交 DNE。
- 截止前进入但持锁跨过截止的主动行权为零落库、仓位恢复可用；等待期间切为 `EXPIRED` 同样拒绝，
  `PAUSED` 对照仍可生成一条 `ACTIVE` 指令。
- 混合 AUTO/DNE 到期只对 AUTO 多头行权一张，DNE 数量与余额变化为零；4 条真实 Asset 指令和
  流水全部成功对账，空头扣 20 = 多头净入 18 + 费用 2。
- 美式 501/5000、实物 501、现金到期 501/502，以及 Asset、行权、实物交割多进程强杀接管均通过。
- 验收脚本连续安装两次生命周期迁移；最终复跑实际输出
  `last_trade_boundary=contract_status=3 order_status=4 cancel_reason=CONTRACT_LAST_TRADE_ENDED
  distinct_times=1 premature_exercises=0 premature_settlements=0`。

## 4. 兼容、发布与回滚

- 先执行幂等迁移和回填，再发布生成代码、Option、Admin API/UI；旧内部写入在过渡期由 INSERT
  触发器补齐 `last_trade_time=exercise_cutoff_time`，不得长期依赖该兼容路径。
- 上线前查询所有合约并证明五个时间点有序；任何不合法记录、缺少系列快照或旧客户端未升级均阻断。
- 回滚应用版本时保留新字段和历史值，不删除列、不把最后交易时间覆盖回到期时间。若需业务回退，
  只能停止新系列/新合约并恢复上一版服务；已上市合约的经济时间不得原地改写。

## 5. 生产前剩余验收

- 业务、法务、风控、清算批准目标市场的五个时间点、IANA 时区、节假日调整、AUTO 阈值、费用及
  现金/实物违约处置；当前没有这些签署。
- 在生产同构多实例和 NTP 条件下验证精确边界、调度迟到、重复/乱序消息、Option/队列/数据库容器
  重启及 RTO；证明最后交易撤单和资金释放在到期前全部终态。
- 接入生产结算价来源、Prometheus/Alertmanager、值班案件和真实用户/做市商公告送达证据。
- 从同一公开下单形成的真实仓位连续走到行权、AUTO/DNE 到期及现金/实物最终处置，并完成逐币
  Asset 流水、Option 指令、钱包、费用和保证金守恒。
- 每个生产合约填写 `docs/templates/option-exercise-expiry-control-record.md` 并由技术、行情、风控、
  清算/财务、运营和合规签署。任一证据缺失时状态保持 `PREPROD_BLOCKED`。

仓库结论：独立生命周期与 AUTO/DNE 设计满足一般标准化期权的 V1 基线；生产运行、目标市场规则和
最终签署尚未完成，因此 Option 整体仍不能认定为生产就绪。
