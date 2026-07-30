# 永续与交割合约五组生产材料

## 当前结论

本目录按 2026-07-30 当前 Deploy 环境的真实事实整理，不把模板、建议值或本地测试
结果标记为生产审批。

| 组 | 材料 | 当前结论 |
| --- | --- | --- |
| 1 | 行情许可与历史回放 | 三源历史回放及 MARK v3 断流自恢复通过；缺供应商商业使用书面授权、v3 正式回放审批 |
| 2 | 值班、升级与告警演练 | 三类事件已完成持久化、L1 升级、系统管理员确认和恢复；缺真实值班排班、L2/L3 人员责任及正式审批 |
| 3 | 保险基金与强平发布 | 账户和回滚技术方案存在；审批水位为 0、账户余额为 0，未注资 |
| 4 | 生产灾备 | 单机全库恢复及 ROW Binlog 精确位点冒烟通过；缺加密异地备份、全库 PITR、切换/回切和正式 RPO/RTO |
| 5 | 交割合约启用 | `BTCUSDT-20260925` 参数完整且无存量事实；状态仍为停用 |

五组材料只有在各文件“通过条件”全部取得客观证据后，才能填写
`deploy/production-readiness.env` 的生产审批字段。

2026-07-30 激活 `production-mark-v3` 后重新执行完整只读
`contract-readiness`：60 PASS、17 FAIL，运行服务、四类行情公式、实时输出、永续产品、
Outbox、对账和 Settlement 均通过；17 项仍全部属于外部审批/生产演练、实际注资以及
最终交割启用门禁。

## 文件

- [01-market-data-license-and-replay.md](01-market-data-license-and-replay.md)
- [mark-v3-price-replay-report.md](mark-v3-price-replay-report.md)
- [02-alert-oncall-and-escalation.md](02-alert-oncall-and-escalation.md)
- [03-fund-and-liquidation-release.md](03-fund-and-liquidation-release.md)
- [04-disaster-recovery.md](04-disaster-recovery.md)
- [05-delivery-enablement.md](05-delivery-enablement.md)
- [production-approval-input.env.example](production-approval-input.env.example)：五组负责人统一回填表
- [SHA256SUMS](SHA256SUMS)：材料完整性校验

## 下一步

复制统一回填表，由法务/行情、运维、资金/风控、基础设施和发布负责人分别填写真实
审批值。当前可以继续自动执行的只有只读检查；以下操作必须等待对应真实材料：

- 未取得三家行情供应商书面授权，不得将许可门禁改为通过；
- 未给出批准水位和资金来源，不得向保险基金调账；
- 未完成生产灾备演练，不得把预生产恢复报告标记为生产报告；
- 前四组没有全部通过，不得启用 `BTCUSDT-20260925`。
