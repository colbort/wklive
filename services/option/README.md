# Option 服务

本服务覆盖 Call/Put 合约、权利金订单与撮合、持仓、风险账户、行权、到期结算、行情/Greeks、
组合保证金、实物交割和运营治理。

当前生产状态以文档台账为准，不能根据接口或表已存在推断功能已经可上线：

- 设计评审：[`docs/option-design-review.md`](docs/option-design-review.md)
- 整改与完成证据：[`docs/option-remediation-plan.md`](docs/option-remediation-plan.md)
- 生产验收矩阵：[`docs/option-acceptance-test-plan.md`](docs/option-acceptance-test-plan.md)
- 当前状态与生产阻断：[`docs/option-current-status-and-production-blockers.md`](docs/option-current-status-and-production-blockers.md)
- 当前仓库技术证据：[`docs/evidence/option-repository-technical-evidence-20260802.md`](docs/evidence/option-repository-technical-evidence-20260802.md)
- 发布候选变更清单：[`docs/evidence/option-release-candidate-change-inventory-20260802.md`](docs/evidence/option-release-candidate-change-inventory-20260802.md)
- 运行与应急：[`docs/option-operations-runbook.md`](docs/option-operations-runbook.md)
- 告警目录：[`docs/option-alert-catalog.md`](docs/option-alert-catalog.md)
- 合约上市门禁：[`docs/option-contract-launch-checklist.md`](docs/option-contract-launch-checklist.md)
- 独立生命周期与 AUTO/DNE 仓库验收：[`docs/option-p2-002-p2-003-lifecycle-repository-acceptance.md`](docs/option-p2-002-p2-003-lifecycle-repository-acceptance.md)
- 保险基金流水语义仓库验收：[`docs/option-p1-008-insurance-fund-ledger-repository-acceptance.md`](docs/option-p1-008-insurance-fund-ledger-repository-acceptance.md)
- 复杂订单仓库验收：[`docs/option-p2-007-complex-order-repository-acceptance.md`](docs/option-p2-007-complex-order-repository-acceptance.md)

仓库级检查：

```sh
bash services/option/monitoring/option-production-readiness.sh --repository-only
services/option/acceptance/run-p0-asset-rpc-e2e.sh
```

第二条命令启动隔离环境的真实 Asset gRPC，验证用户钱包范围、净清算权益、冻结响应丢失重放、
余额镜像、现金结算故障恢复、20并发美式提前行权/FIFO指派、501/5000容量、三类进程强杀接管、
上市/最后交易/行权截止/到期/交割五个独立边界、混合AUTO/DNE到期、20路合约系列创建和500条原子生成/上市门禁，以及500/501期权链、
100/101档盘口、24h/OI、并发一致性快照和跨租户公开行情隔离；另覆盖复杂订单50路同键幂等、
真实Asset冻结/扣款响应丢失、第二腿事务回滚、激活后控制重检、FOK/STP和后台整组强撤；覆盖范围及资源清理规则见
[`acceptance/README.md`](acceptance/README.md)。它不替代生产同构消息/容器、多系列批准峰值、
生产通知和六方签署。行权/到期运营证据使用
[`docs/templates/option-exercise-expiry-control-record.md`](docs/templates/option-exercise-expiry-control-record.md)。
