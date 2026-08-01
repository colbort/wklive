# Option 服务

本服务覆盖 Call/Put 合约、权利金订单与撮合、持仓、风险账户、行权、到期结算、行情/Greeks、
组合保证金、实物交割和运营治理。

当前生产状态以文档台账为准，不能根据接口或表已存在推断功能已经可上线：

- 设计评审：[`docs/option-design-review.md`](docs/option-design-review.md)
- 整改与完成证据：[`docs/option-remediation-plan.md`](docs/option-remediation-plan.md)
- 生产验收矩阵：[`docs/option-acceptance-test-plan.md`](docs/option-acceptance-test-plan.md)
- 运行与应急：[`docs/option-operations-runbook.md`](docs/option-operations-runbook.md)
- 告警目录：[`docs/option-alert-catalog.md`](docs/option-alert-catalog.md)
- 合约上市门禁：[`docs/option-contract-launch-checklist.md`](docs/option-contract-launch-checklist.md)

仓库级检查：

```sh
bash services/option/monitoring/option-production-readiness.sh --repository-only
services/option/acceptance/run-p0-asset-rpc-e2e.sh
```

第二条命令启动隔离环境的真实 Asset gRPC，验证用户钱包范围、净清算权益、冻结响应丢失重放、
余额镜像、现金结算故障恢复、20并发美式提前行权/FIFO指派以及混合AUTO/DNE到期；覆盖范围及
资源清理规则见 [`acceptance/README.md`](acceptance/README.md)。它不替代预生产消息、501/5000
空头真实容量、进程强杀、生产通知和六方签署。行权/到期运营证据使用
[`docs/templates/option-exercise-expiry-control-record.md`](docs/templates/option-exercise-expiry-control-record.md)。
