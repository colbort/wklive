# Option 合约上市检查表

OPTION_LAUNCH_CHECKLIST_STATUS: DRAFT

只有全部适用项完成、证据哈希固定并取得六方签字后，才可把状态改为 `APPROVED`；签署后的副本不得原地改写。

合约代码：________　租户：________　计划开放时间：________　负责人：________

## 1. 产品与法律

- [ ] 合约说明书包含标的、Call/Put、欧式/美式、现金/实物、行权价、乘数、报价币、结算币。
- [ ] 最后交易、到期、主动行权截止、结算窗口、交割时间及时区明确。
- [ ] 结算价来源、算法、窗口、最小样本数、精度和异常处置已批准。
- [ ] 自动行权阈值、DNE/相反指令规则已说明；未实现能力未在产品文案中承诺。
- [ ] 卖方义务、保证金、强平、保险/平台兜底规则和最大损失披露完成。
- [ ] 对外材料未把保险基金/平台兜底描述为用户余额、存款保险、收益保证、卖方损失上限或必然垫资承诺。
- [ ] 实物交割资产不足规则、补资截止和追偿方案已批准（仅实物）。
- [ ] 实物交割补资通知、逾期违约、退款/继续交割/拍卖/平台承接等最终处置已由
  产品、风控、清算/财务、运营/客服、技术和合规/法务共同批准；
  `templates/option-physical-delivery-default-policy-approval.md`签署副本为`APPROVED`，
  所有批准路径已经实现并通过专项预生产验收；不得只依赖仓库安全默认值。

## 2. 参数四眼复核

- [ ] 若由系列生成，`series_id/version/payload_hash`、参考价来源/时间/证据、显式 expiry、
  strike band 和预计数量已使用 `docs/templates/contract-series-approval.md` 独立复核。
- [ ] 系列生成明细对每个 expiry/strike 都只有一对 Call/Put，实际数量等于预计数量；
  生成状态全部为 `PENDING`，无代码碰撞、额外模板合约或半套批次。
- [ ] `contract_code`、标的、币种映射正确且不可混淆。
- [ ] `strike_price`、`multiplier`、`price_tick`、`qty_step`、最小/最大订单量正确。
- [ ] maker/taker/行权费率及费用账户正确。
- [ ] 保证金模式、初始/维持/最低保证金率正确。
- [ ] 强平费率、保险账户、兜底策略和额度正确。
- [ ] 结算配置为 `authoritative-market`、`MEDIAN`，窗口和最小样本数符合审批。
- [ ] `exercise_cutoff_time` 早于或等于到期且晚于上市；`auto_exercise_threshold` 非负，并与公告精度一致。
- [ ] `max_user_long_qty`、`max_user_short_qty`、`max_open_interest` 均为经风控批准的正数；0 不表示无限。
- [ ] `order_price_band_ratio`、`circuit_breaker_ratio` 均为 `(0,1]` 内经审批值，并与行情波动和订单 tick 压测结果一致。
- [ ] `greeks_max_age_seconds` 为正数，已按
  `templates/option-market-freshness-approval.md` 完成产品、风控模型、市场运营和技术复核；不得复用未经证明的30秒值。
- [ ] 新建状态只能是 `PENDING`；后台参数更新不能直接修改状态。上市时标的价和标记价为正、
  各自快照不超过30秒，且生命周期日志显示全部自动门禁通过。
- [ ] 所有 Unix 时间已从产品时区交叉换算并由第二人复核。
- [ ] 未完成能力的功能开关保持关闭。

## 3. 技术门禁

- [ ] `monitoring/option-production-readiness.sh --repository-only` 通过；生产模式使用仓库外
  attestation 执行后为 `READY`，输出、attestation、生产配置和全部证据 SHA-256 已归档。
- [ ] 发布身份固定为已审查的 commit 与镜像 digest；`monitoring/option-release-scope.sh --release-clean`
  在该 commit 通过，工作区无未审查改动。`REPOSITORY_ONLY / PASSED_NOT_RELEASE_CANDIDATE` 证据不得冒充发布候选或预生产证据。
- [ ] `option-remediation-plan.md` 中该产品适用依赖项至少为 `REPOSITORY_PASSED / PREPROD_BLOCKED`，
  且 `option-current-status-and-production-blockers.md` 中对应外部阻断已有真实预生产证据和责任方签署；不得把仓库状态直接改写成生产 `DONE`。
- [ ] 当前发布 commit 已执行 `make gen-model` 和 `make gen`；生成文件与 schema/proto 一致，执行后 `git diff --check` 通过且没有未审查生成差异。
- [ ] 迁移已在预生产重复执行并核对结构、索引、约束和历史回填。
- [ ] 批量上币已执行 SER-001～SER-010；尤其并发复核、响应丢失重试、单条冲突整批回滚、
  500 条边界和目标市场节假日/DST 时间复核均有证据。
- [ ] 新库也已执行 `20260730_z_option_settlement_price_approval.sql` 和
  `20260730_zd_option_exercise_governance.sql`、`20260730_ze_option_trading_controls.sql`、
  `20260730_zf_option_trade_correction.sql`、`20260730_zg_option_mmp.sql`、
  `20260730_zh_option_portfolio_risk_governance.sql`、
  `20260730_zi_option_physical_delivery_default.sql`、
  `20260731_zu_option_greeks_freshness.sql`、
  `20260731_zv_option_order_portfolio_config.sql`、
  `20260731_zw_option_assignment_pagination.sql`、
  `20260731_zx_option_margin_coin_evidence.sql`、
  `20260731_zy_option_settlement_price_evidence.sql`、
  `20260801_option_portfolio_liquidation.sql`、
  `20260801_option_portfolio_liquidation_monitoring.sql`、
  `20260802_option_last_trade_time.sql`、
  `20260802_option_insurance_inventory_exit.sql`、
  `20260802_option_insurance_fund_flow_semantics.sql`；
  `option.sql` 不包含这些迁移创建的不可变经济字段/控制事件触发器。
- [ ] 18 项 Option 监控索引和 61 条 Option 告警规则均存在；生产 Prometheus 已加载规则，
  SEV-1/2/3 的通知、恢复和失败升级使用 `templates/option-alert-delivery-test.md` 留证。
- [ ] Option、Asset、Market、Redis、消息队列和数据库容量/健康正常。
- [ ] 标的、标记价、Greeks 三类时间戳分别可见，时间同步正常；Greeks 在批准阈值边界内显示新鲜，超过一秒显示过期并触发 OPT-A003。
- [ ] 价格带、用户多空/OI 限额、跨账户 STP、kill switch、熔断及恢复审计已验收。
- [ ] 如开放 MMP，P1-003 必须已达到仓库通过状态，且对应预生产证据和生产签署已批准；已验收报价组、阈值、触发撤单、冷静期、
  人工恢复、权限和审计；否则客户端/做市商权限必须保持关闭。
- [ ] 如开放组合保证金，P1-004 必须已达到仓库通过状态，且对应预生产证据和生产签署已批准；当前租户和结算币存在有效已批准版本，
  回测报告、独立验证、压力结果和回滚演练证据与版本 ID 一致；边界测试中新组合卖单保存的
  准入配置 ID/版本与风险账户快照一致且不可改写；否则组合卖方入口保持关闭。
- [ ] 如开放实物交割，P1-007 必须已达到仓库通过状态，且对应预生产证据和生产签署已批准；配对交割单元、先收后付、补资截止、
  人工复核、两币种守恒、通知和最终经济处置均已验收；未获批准或批准但尚未实现的退款、罚款、
  拍卖、平台承接、放弃行权等路径必须无入口，否则实物合约不得进入交易态。
- [ ] 异常成交现金更正的暂停、撤单、借贷守恒、四眼复核、先扣后入账、人工重试和不可变审计已验收；不得以人工改表替代。
- [ ] 风险扫描、强平、资金指令、行权和结算任务具备指标与告警。
- [ ] 如开放卖方交易，财务/清算已批准保险流水“正绝对金额+类型方向”语义；新流水金额为正，
  类型 1/3 计流入、2/4 计流出，流水不可修改/删除；历史数据、缺失 `asset_flow_no` 和所有下游
  `SUM(amount)` 消费者已逐项盘点并与 Asset 同截止点复算；
  `templates/option-insurance-fund-ledger-production-approval.md` 的签署归档副本为 `APPROVED`，
  未通过批准不得批量改写历史。
- [ ] 回滚/前滚步骤经过演练，且不会删除已发生资金流水。

## 4. 场景验收

- [ ] 下单、撤单、部分成交、IOC/FOK、同用户跨账户自成交保护通过。
- [ ] 限额边界和 20 单并发准入、待落仓 outbox 敞口桥接、价格带上下边界通过。
- [ ] kill switch 与并发下单、三种活动订单撤销、重复启用、管理员带理由解除通过。
- [ ] 标记价跳变等于/高于阈值时暂停和批量撤单，恢复交易双签审计通过。
- [ ] MMP 数量/笔数/损失阈值边界、同组全撤、跨组隔离、非 MMP 隔离、冷静期和人工恢复通过。
- [ ] 异常成交案件创建、申请人自审拒绝、不平分录拒绝、Asset 失败不提前入账、完成后净额为 0 通过。
- [ ] 多头、空头、价差和跨账户风险权益/保证金通过。
- [ ] 组合保证金版本切换前后新卖单保存准确配置 ID/版本；缺省、错配和事后改写被数据库拒绝，
  迁移前 `0/0` 历史单仍可正常撮合/撤单但标记为历史证据不可用。
- [ ] 行情陈旧、未来时间戳、Asset 超时和消息重放故障注入通过。
- [ ] 主动行权幂等、步长、净收益、并发指派通过（仅美式）。
- [ ] AUTO/DNE/相反指令、截止边界、阈值边界、历史版本不可篡改和重启重放通过（仅现金结算）。
- [ ] 1/500/501/5000 空头 FIFO 指派及未指派订单不受影响通过（仅美式）。
- [ ] 到期窗口取价、样本不足、确定性重算、结算资产守恒通过。
- [ ] 已确认自动结算价在使用点按不可变快照 ID 复算一致；错误中位数、窗口、缺失/重复证据
  均被数据库和生命周期双重阻止；人工更正四眼、版本保留且不被自动算法监控误报。
- [ ] 实物 Call/Put 的交割币种、资产不足和人工处置通过（仅实物）。
- [ ] 日终对账与事故恢复演练通过。

## 5. 运营准备

- [ ] 基于 `option-product-operations-pack.md` 完成公告、风险披露、FAQ、用户通知和客服话术的产品/合规审批。
- [ ] 最终发布副本不含方括号占位符、`待填写`、空联系人、空链接或未确定时区；最后交易、
  指令截止、到期和结算/交割时间分别列出，且与已批准合约记录逐项一致。
- [ ] 值班表、联系人、通知群、升级路径和权限完成并实测。
- [ ] 结算日行情、技术、风控、运营、清算双人审批人员在线。
- [ ] 告警无“待接入”，仪表盘和恢复入口可用。
- [ ] 保险与平台兜底余额达到审批下限。

## 6. 签字

| 角色 | 姓名 | 结论 | 时间 | 证据链接 |
| --- | --- | --- | --- | --- |
| 产品 |  | 通过/拒绝 |  |  |
| 技术 |  | 通过/拒绝 |  |  |
| 风控 |  | 通过/拒绝 |  |  |
| 清算/财务 |  | 通过/拒绝 |  |  |
| 运营 |  | 通过/拒绝 |  |  |
| 合规 |  | 通过/拒绝 |  |  |

任一必选项未勾选或任一角色拒绝，合约不得切换为交易状态。
未签署副本的 `OPTION_LAUNCH_CHECKLIST_STATUS` 必须保持 `DRAFT`；签署归档副本才允许标记为 `APPROVED`。
