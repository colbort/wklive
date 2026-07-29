# 永续与交割合约生产验收证据规范

## 1. 用途

本规范定义 `deploy.sh contract-readiness` 所需的四份生产证据。证据必须来自目标生产
环境或获批的同等预生产环境，测试夹具、本地日志和空白模板不能作为通过依据。

声明文件只保存审批结论、业务维度、证据绝对路径和 SHA-256，不得保存行情源密码、
API Key、访问令牌或数据库凭据。

## 2. 历史价格回放报告

报告至少包含：

- 环境、品类、市场、价格 Symbol、永续 Symbol、交割 Symbol；
- 三个以上真正独立来源的 Authority、市场和 Symbol 映射；
- Authority 的顺序必须与声明的 DELIVERY 权重顺序一一对应，公式不得包含声明外来源；
- 数据许可或使用批准编号，以及凭据接入批准编号；
- 回放开始/结束时间，必须覆盖至少一个完整交割锁价窗口；
- INDEX、MARK、FUNDING、DELIVERY 的公式编号、不可变版本、算法和参数；
- DELIVERY 回看窗口与交割合约 `settlement_window_seconds` 的一致性，以及公式版本与
  合约 `settlement_price_algorithm` 的一致性；
- 输入记录数、目标时点数、最少有效来源数、剔除数、断档数和篡改校验结果；
- 最大/平均来源延迟、最大偏差、输出价格范围；
- `price-replay --json` 的原始输出或其不可变归档位置；
- 执行人、复核人、时间和审批编号。

回放必须以非零退出码处理重复输入、时间断档、目标时间未对齐、输出篡改或有效来源
不足；不得人工修改回放结果后重新计算哈希。

## 3. 告警投递测试报告

报告至少包含：

- 告警平台、规则 ID、日志或指标查询条件；
- Snapshot Outbox、Price Engine 缺源、合约对账差异三类测试事件；
- 每类事件的产生时间、平台接收时间、通知时间和端到端延迟；
- 值班组、一级通知渠道、未确认升级链路和最终责任人；
- 首次、内容变化、恢复后重开及 30 分钟提醒的去重结果；
- 恢复通知或告警自动关闭结果；
- 平台事件 ID、通知回执或不可变归档位置；
- 执行人、值班负责人、时间和审批编号。

仅证明应用输出了 Error 日志不算告警投递通过；必须取得告警平台及通知渠道的接收
证据。

## 4. 自动强平启用与回滚方案

方案至少包含：

- 目标租户、结算币种、永续和交割合约；
- `INSURANCE_FUND`、`FEE_REVENUE` 账户 ID、状态、余额复核和资金权限审批编号；
- 合约默认保险配置、ADL 策略和负责人；
- 启用窗口、发布单、操作人、复核人和审批人；
- 启用前配置快照及其 SHA-256；
- 先验证价格、告警、账户和水位，再启用
  `AutomaticLiquidation.Enabled`，最后按审批启用
  `CrossMarginTrading.Enabled` 的顺序；
- 启用后观察项：新鲜 MARK、任务成功率、未完成 Settlement、OPEN 对账差异、
  保险基金余额变化和告警投递；
- 自动回滚阈值及人工回滚步骤；
- 回滚后双开关为 `false`、只允许 Reduce Only、进行全量对账并继续恢复已产生 Saga
  的验证步骤；
- 演练结果、执行人、复核人和审批编号。

`contract-readiness` 全部通过也不能自动打开开关；真实启用必须按本方案执行。

## 5. 生产灾备演练报告

演练按照
[永续与交割合约备份及灾备恢复手册](perpetual-delivery-disaster-recovery-runbook.md)
执行，报告至少包含：

- 已审批 RPO、RTO、保留周期、加密方式和异地位置；
- 备份开始/结束时间、大小、SHA-256、GTID/Binlog 位点；
- Binlog 时间点恢复的目标时点、实际丢失窗口和实际恢复耗时；
- 节点或可用区故障切换、回切步骤和耗时；
- `schema_migrations` 及全部核心事实表的时点行数对比；
- 未完成 Saga 恢复、Outbox 排空、全量对账和重复资金/仓位检查结果；
- Trade、Asset、iTick、System 及基础设施最终健康状态；
- 演练人、复核人、时间、问题单和审批编号。

只完成本地 `mysqldump` 恢复不能替代生产 Binlog、异地存储和故障切换演练。

## 6. 生成声明及证据哈希

```bash
cd deploy
cp production-readiness.env.example production-readiness.env

shasum -a 256 /absolute/path/to/price-replay-report.json
shasum -a 256 /absolute/path/to/alert-delivery-test.md
shasum -a 256 /absolute/path/to/liquidation-rollback-plan.md
shasum -a 256 /absolute/path/to/dr-exercise-report.md

./deploy.sh contract-readiness
```

Linux 可用 `sha256sum`。将输出的 64 位哈希填入对应 `*_SHA256` 字段。修改证据后必须
重新审批并重新生成哈希。

通过标准：

- 命令输出不存在 `FAIL`；
- 最终输出 `READY` 且退出码为 0；
- 预检期间两个生产安全开关仍为 `false`；
- 输出和四份证据随发布单归档。
