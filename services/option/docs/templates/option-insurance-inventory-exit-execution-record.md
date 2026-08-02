# Option 保险接管库存退出执行记录

OPTION_INSURANCE_INVENTORY_EXIT_EXECUTION_STATUS: DRAFT

> 每个退出申请单独归档。该记录用于证明四眼复核、唯一订单、真实资金和剩余库存闭环，不替代批准的限额版本。

## 1. 身份与审批

| 字段 | 记录值 |
| --- | --- |
| tenant_id |  |
| exit_id / request_no |  |
| 原 liquidation_no |  |
| 接管 position_id / margin_lot_id |  |
| contract_id / contract_code |  |
| 保险 user_id / account_id |  |
| 审批单版本 / SHA-256 |  |
| requested_by / 时间 |  |
| reviewed_by / 时间 / 意见 |  |
| submitted_by / 时间 |  |

- [ ] 创建人、复核人不同，四眼复核成立。
- [ ] create/review/execute权限分别属于获批角色。
- [ ] 申请状态只按 PENDING_REVIEW → APPROVED/REJECTED → SUBMITTED 迁移。

## 2. 执行前证据

| 检查 | 记录值 |
| --- | --- |
| 现金结算 / 空头 / 保险账户 |  |
| 合约状态 / 到期时间 |  |
| position_qty / available_qty / frozen_qty |  |
| margin_amount / margin lot余量 |  |
| mark_price / mark_snapshot_time |  |
| order_price_band_ratio / price_tick / qty_step |  |
| 申请 quantity / IOC limit_price |  |
| 保险钱包 total / available / frozen |  |
| 订单簿深度、预估滑点和证据链接 |  |
| 单次/单日/单合约批准限额余量 |  |
| 运行时五项硬限额 / 配置SHA-256 |  |
| UTC日已提交数量 / 未决预留数量 |  |

- [ ] 标记价不超过30秒，数量和价格符合步长与价格带。
- [ ] 额外可用权利金足够，未挪用接管保证金。
- [ ] 同一 position 没有其他 PENDING_REVIEW 或 APPROVED 申请。
- [ ] 单次数量、最坏权利金、UTC日预算、标记价偏离和可成交深度均满足运行时硬限额。

## 3. 订单与幂等

| 字段 | 记录值 |
| --- | --- |
| client_order_id（必须为 INS-EXIT-&lt;exit_id&gt;） |  |
| order_id / order_no |  |
| source / side / effect / reduce_only / type | ADMIN / BUY / CLOSE / YES / IOC |
| qty / limit_price |  |
| filled_qty / unfilled_qty / avg_price |  |
| order status / cancel reason |  |
| 并发或重试次数 / 返回的唯一order_id |  |

- [ ] 重复执行只存在1个订单、1个client key。
- [ ] 部分成交时未成交余量已撤销，仓位 frozen_qty 为0。
- [ ] 响应丢失或重试沿用原 exit_id/client_order_id。

## 4. 资金、持仓与守恒

| 对账项 | 执行前 | 执行后 | 差额/说明 |
| --- | ---: | ---: | --- |
| 保险仓 position_qty |  |  |  |
| 保险仓 margin_amount |  |  |  |
| margin lot remaining_quantity |  |  |  |
| margin lot remaining_margin |  |  |  |
| 保险钱包 total / available / frozen |  |  |  |
| 对手方钱包 total / available / frozen |  |  |  |
| 权利金冻结/消费/余量释放 |  |  |  |
| 成交对手方入账 |  |  |  |
| 保证金释放 |  |  |  |
| 手续费 |  |  |  |

- [ ] 每条Option资金指令均为SUCCESS/MATCHED，并按
  `action=1 ? target_biz_no : instruction_no`找到唯一Asset流水。
- [ ] 退出预算入账有独立业务号和唯一Asset流水。
- [ ] 实际减仓数量等于filled_qty，保证金仅按实际成交量释放。
- [ ] 订单、成交、仓位、margin lot、钱包和Asset流水数量/金额闭合。

## 5. 监控与恢复

| 字段 | 记录值 |
| --- | --- |
| 执行后库存数量 / 标记价值 / 绝对Delta |  |
| 24h / 8h / 1h到期水位 |  |
| OPT-A035/A036/A037状态 |  |
| 连续三个正常采样窗口 |  |
| 是否恢复相关卖方风险 / 批准人 |  |
| last_error_msg（如有）及修复记录 |  |

最终签署：执行人 ____；复核人 ____；清算 ____；风控 ____；保险资金 ____；日期 ____。
