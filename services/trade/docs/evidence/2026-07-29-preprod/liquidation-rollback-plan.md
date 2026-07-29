# 永续与交割合约自动强平启用与回滚方案（预生产）

## 1. 目标与当前结论

- 环境：本机完整 Docker Compose 预生产验收环境
- 租户：`tenant_id=1`
- 结算币种：USDT
- 永续合约：`symbol_id=2`，`BTCUSDT`，状态启用
- 交割合约：当前无启用产品；原 `symbol_id=4`、`BTCUSDT` 空壳产品已停用
- 当前结论：**禁止启用自动强平和全仓新增风险敞口**

当前配置不满足真实资金启用条件：

1. `INSURANCE_FUND` 账户 `id=1` 可用余额为 0；
2. 没有启用的 USDT `FEE_REVENUE` 平台账户；
3. 没有 `symbol_id=0`、USDT 的默认保险基金配置；
4. 原交割合约 `symbol_id=4` 缺少 `t_trade_symbol_contract` 参数事实，已在确认
   Order/Position/Batch/Reservation/Instruction 均为 0 后通过 Admin API 停用；
5. Price Engine 仍为单源 MARK/INDEX，且不存在 DELIVERY 公式；
6. 外部告警平台、值班组及升级策略未配置。

## 2. 当前安全配置

- `AutomaticLiquidation.Enabled=false`
- `CrossMarginTrading.Enabled=false`
- Reduce Only 平仓能力保留
- 未完成 Settlement Instruction：0
- OPEN 合约对账差异：0
- 不健康 Snapshot Outbox：0

在所有启用前置条件满足并取得审批前，这两个开关必须保持 `false`。

## 3. 启用前置条件

### 3.1 价格与合约

- INDEX 使用三个以上真实独立市场来源；
- MARK 使用已回放的 `INDEX_BASIS` 公式，并固化偏离限幅、平滑权重及版本；
- FUNDING 与 MARK/INDEX 版本对应；
- DELIVERY 使用三个以上独立来源，公式版本与交割合约
  `settlement_price_algorithm` 一致；
- 新建交割产品或重新启用 `symbol_id=4` 前，补齐并复核交割时间、停止撮合时间、
  锁价窗口和交割价来源；
- 完整生产历史窗口回放通过，篡改、断档、重复目标时点和来源不足均能阻断。

### 3.2 账户与资金权限

- 建立并启用 USDT `FEE_REVENUE` 平台账户；
- 为 USDT `INSURANCE_FUND` 账户充值至经风险委员会批准的最低水位；
- 建立 `symbol_id=0` 的 USDT 默认保险配置，明确 `fund_user_id` 和 ADL 策略；
- 记录账户 ID、余额快照、操作权限、复核人和审批编号；
- 验证保险基金全额、部分承接、余额不足、冲正和 ADL 两阶段恢复。

### 3.3 告警与运维

- Snapshot Outbox、Price Engine 缺源和合约对账差异接入外部告警平台；
- 配置值班组、一级通知渠道、未确认升级链路和最终责任人；
- 完成通知回执测试；
- 确认备份、PITR、异地副本、可用区切换及正式 RPO/RTO。

## 4. 启用顺序

即使所有预检通过，也不得由 `contract-readiness` 自动开启开关。必须按发布单执行：

1. 冻结目标租户、合约和账户配置；
2. 保存 Etcd Trade 配置快照并计算 SHA-256；
3. 保存价格公式、合约参数、平台账户和保险配置快照；
4. 运行 `./deploy.sh contract-readiness`，要求输出无 `FAIL` 且最终为 `READY`；
5. 人工复核 MARK 新鲜度、Outbox、未完成 Settlement、OPEN 对账和保险基金水位；
6. 先开启 `AutomaticLiquidation.Enabled`；
7. 观察一个经审批的窗口，不立即开启全仓新增风险；
8. 观察期通过后，按独立审批开启 `CrossMarginTrading.Enabled`；
9. Reduce Only 始终可用。

## 5. 自动回滚阈值

出现任一条件立即关闭新增风险：

- MARK/INDEX/FUNDING/DELIVERY 超过配置新鲜度窗口；
- 三源价格有效输入少于最小数量；
- Snapshot Outbox 出现 Failed、Manual 或最老开放记录超过 60 秒；
- 任一 Settlement Instruction 进入 Manual；
- 出现未批准的 OPEN 对账差异；
- 保险基金低于批准水位或账户权限异常；
- 强平、保险承接或 ADL 出现重复资金流水、重复仓位历史或永久中间态；
- 外部告警投递失败或值班链路不可用。

## 6. 人工回滚步骤

1. 将 `CrossMarginTrading.Enabled` 设置为 `false`，停止增加全仓风险敞口；
2. 将 `AutomaticLiquidation.Enabled` 设置为 `false`，停止创建新的自动强平 Saga；
3. 保留 Reduce Only 平仓；
4. 不回滚已经发生 Asset 副作用的 Saga，继续按幂等事实恢复至终态；
5. 保存 Etcd、价格公式、合约、平台账户、未完成指令、Outbox 和对账快照；
6. 核对 Order、Fill、Position、History、Reservation、Instruction、Asset Flow、
   Outbox/Inbox；
7. 执行全量合约对账，OPEN 差异必须为 0 或进入带操作人、原因和请求号的人工处置；
8. 复核保险基金和手续费账户余额、冻结及全部资金流水；
9. 回滚后再次确认两个开关均为 `false`；
10. 形成事故单、影响范围、恢复结果和重新启用审批。

## 7. 观察指标

- 新鲜 MARK/INDEX/FUNDING/DELIVERY；
- Trade 四类核心任务成功率；
- Snapshot Outbox Pending/Processing/Failed/Manual 及最老开放时间；
- 未完成 Settlement Instruction；
- OPEN 对账差异；
- 保险基金、手续费收入和 ADL 缺口变化；
- 强平父子 Saga 的阶段、重试和人工态；
- 外部告警接收、通知和升级延迟。

## 8. 审批栏

- 计划启用窗口：待定
- 发布单：待定
- 操作人：待定
- 复核人：待定
- 资金权限审批人：待定
- 风险负责人：待定
- 值班负责人：待定
- 结论：当前配置不具备启用条件，门禁保持关闭
