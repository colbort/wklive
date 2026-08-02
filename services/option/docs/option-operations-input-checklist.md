# Option 运营资料收集与发布输入清单

OPTION_OPERATIONS_INPUT_STATUS: DRAFT

更新时间：2026-08-02

> 用于把产品、运营、风控、清算、财务和合规必须提供的真实输入收敛到一个入口。仓库已经预填系统
> 硬边界，但不能代填目标市场规则、真实联系人、资本、法律结论或人员签名。禁止在本文件保存密码、
> Token、私人电话或其他应进入安全值班系统的秘密。

## 1. 发布身份与责任

| 字段 | 责任方 | 值/证据 | 状态 |
| --- | --- | --- | --- |
| tenant_id / 法律实体 / 司法辖区 | 产品/合规 | 待填写 | OPEN |
| 目标市场 / 官方规则版本 | 市场运营/合规 | 待填写 | OPEN |
| release commit / 全部image digest | 技术/SRE | 待填写 | OPEN |
| 变更单 / 计划窗口 / 回滚窗口 | 技术/SRE/运营 | 待填写 | OPEN |
| 产品、技术、风控、清算财务、运营、合规负责人 | 各责任方 | 待填写；联系方式存安全系统 | OPEN |
| 本次合约总数 / 系列ID版本 | 产品/运营 | 待填写 | OPEN |
| 合约上市合集及SHA-256 | 运营/技术 | `templates/option-contract-launch-bundle.md` | OPEN |
| 目标环境与公告/客户端合约集导出、SHA及三方对账 | 技术/产品运营 | `templates/option-contract-set-reconciliation.md` | OPEN |
| 上市合集逐合约校验输出 | 技术/运营 | `monitoring/option-launch-bundle-verify.sh`；归档完整PASS输出 | OPEN |
| Option渲染YAML / SHA-256 / `ProductScope`逐项布尔值 | 技术/SRE、产品/风控 | 与本次发布门禁声明完全一致；Greeks依赖扩展固定false；归档`option-product-scope-verify.sh`完整PASS输出 | OPEN |

## 2. 仓库已确定且运营不得改写的边界

- 所有新合约只能以`PENDING`创建；系列批准不等于逐合约上市批准。
- 生命周期必须满足`list < last_trade <= exercise_cutoff <= expire <= deliver`，五个时间分别保存和公告。
- 标的、标记价和Greeks使用独立快照时间；未批准的新鲜度阈值不能伪装为健康默认值。
- 资金以Asset为事实源；Option状态不是资金成功证据。冻结、扣减、释放和入账使用原业务身份幂等恢复。
- 保险流水使用正绝对金额，1/3为流入、2/4为流出；历史流水不因新语义被批量重写。
- 平台兜底默认关闭，只允许`DISABLED/PREFUNDED/CREDIT_FLOOR`，无限负余额不得批准。
- 组合单首版不支持legging、拍卖、多maker聚合FOK或组合保证金抵扣；MMP不等于做市义务。可选能力
  `ProductScope`零值关闭，文档声明不能替代运行时门禁。
- RFQ、大宗和正式做市义务为`DEFERRED`且无入口，运营与销售不得宣称已支持。

## 3. 每个合约必须提供的产品参数

以下每一行都必须进入单合约`option-contract-launch-checklist.md`和最终公告；“沿用默认”不是有效值。

| 参数组 | 必填字段 | 责任方 | 权威依据/签署 |
| --- | --- | --- | --- |
| 身份 | contract_code、标的、Call/Put、欧/美式、现金/实物、报价/结算/标的币 | 产品 | 待填写 |
| 经济参数 | strike、contract_unit、multiplier、price_tick、qty_step、min/max_qty | 产品/风控 | 待填写 |
| 五时间 | list、last_trade、exercise_cutoff、expire、deliver的UTC秒、本地带时区文本、IANA时区 | 产品/市场运营 | 待填写 |
| 费用 | maker、taker、行权、强平费率及每个归集账户 | 产品/财务 | 待填写 |
| 卖方担保 | 保证金模式、初始/维持/最低比率、margin/collateral coin | 风控/清算 | 待填写 |
| 交易控制 | 价格带、熔断、用户多空限额、OI上限、订单数量上限 | 风控/市场运营 | 待填写 |
| 结算 | 权威来源、算法、正式窗口、最小样本数、精度、异常/更正规则 | 行情/清算/风控 | 待填写 |
| 行权 | AUTO阈值、DNE/相反指令、费用后正收益规则、通知截止 | 产品/清算 | 待填写 |
| 行情 | 标的/标记价来源、Greeks阈值、陈旧降级和展示精度 | 行情/风控 | 待填写 |
| 条件功能 | 卖方、MMP、组合保证金、实物、复杂订单、公开行情、美式提前行权是否开放；Greeks扩展固定false | 产品/风控 | 待填写；未批一律false；与渲染YAML逐项一致 |

## 4. 目标市场与日历资料

| 资料 | 最低要求 | 责任方 | 证据入口 |
| --- | --- | --- | --- |
| 年度日历 | 官方来源、原始文件哈希、IANA时区、节假日、提前闭市、补班、DST | 市场运营/合规 | `templates/trading-calendar-annual-review.md` |
| 日历版本 | 周会话、UTC例外、生效/失效、四眼批准 | 市场运营/技术 | `templates/trading-calendar-approval.md` |
| 临时休市 | 原halt、撤单/资金释放、通知、恢复条件和三方批准 | 风控/运营/技术 | `templates/trading-halt-record.md` |
| 公司行动 | 官方公告、版本、调整公式、零碎/税费、后继合约和用户通知 | 市场运营/法务/风控 | `templates/corporate-action-case.md` |
| 系列政策 | 到期周期、节假日调整、strike带宽/步长、命名、参考价源 | 产品/市场运营/风控 | `templates/contract-series-approval.md` |

## 5. 资金、风险和卖方资料

| 资料 | 必填输入 | 责任方 | 证据入口 |
| --- | --- | --- | --- |
| 保险账本 | 正式语义批准、历史逐笔Asset桥接、下游消费者、影子复算、异常案件 | 财务/清算/合规 | `templates/option-insurance-fund-ledger-production-approval.md` |
| 日终守恒 | 同一截止点、逐币期初/流量/期末、差异案件、36小时成功心跳 | 财务/清算/技术 | `templates/option-daily-fund-reconciliation.md` |
| 平台兜底 | 模式、资本来源、policy ID/version、单笔/日累计/余额底线、偿还口径 | 管理层/财务/风控 | `templates/option-platform-backstop-policy-approval.md` |
| 保险库存退出 | 逐合约数量/金额/日限额、滑点、深度、资金来源、四眼角色 | 风控/清算/保险资金 | `templates/option-insurance-inventory-exit-approval.md` |
| 组合保证金 | 代表性数据、情景参数、集中度/流动性附加、独立验证、回滚版本 | 风控/独立验证/清算 | `templates/option-portfolio-risk-validation-record.md` |
| MMP | 用户/合约/group全集、数量/笔数/损失阈值、窗口、冷静期、恢复与通知 | 产品/风控/市场运营/SRE | `templates/option-mmp-readiness.md` |
| 风险参数变更 | 原/新版本、未来生效、影响账户、回滚触发和三个监控窗口 | 风控/技术/运营 | `templates/risk-parameter-change.md` |
| 实物违约 | 补资期限、通知、退款/继续/罚则/拍卖/承接或明确禁用、司法辖区 | 产品/法务/清算 | `templates/option-physical-delivery-default-policy-approval.md` |

## 6. 行情、容量、通知与客服资料

| 资料 | 必填输入 | 责任方 | 证据入口 |
| --- | --- | --- | --- |
| 公共行情 | 字段、展示精度、缓存TTL、限流、P95/P99、降级、免责声明 | 行情/产品/SRE/合规 | `templates/public-market-readiness.md` |
| 复杂订单 | 策略范围、父单上限、交易时段、费用、SLA、披露 | 产品/风控/市场运营 | `templates/complex-order-readiness.md` |
| 美式提前行权 | 美式合约全集、边界/FIFO/资金/故障、逐合约控制记录、通知与法律披露 | 产品/风控/清算/运营/合规 | `templates/option-american-exercise-readiness.md` |
| Beanstalk | ARM64/AMD64原生结果、批准峰值、WAL、重连、RTO、24h长稳 | SRE/容量负责人 | `templates/option-beanstalk-capacity-rto-report.md` |
| 编排故障 | 真实实例、自然租约、双实例竞争、资金唯一、告警和RTO | SRE/Option/Asset | `templates/option-orchestrator-takeover-report.md` |
| 监控通知 | 生产target、receiver、电话/IM/案件、恢复和失败升级 | SRE/值班/业务 | `templates/option-alert-delivery-test.md` |
| 用户材料 | 合约说明、风险披露、公告、FAQ、客服话术、语言和触达批次 | 产品/运营/客服/合规 | `option-product-operations-pack.md` |
| 值班资料 | 主备角色、时区、升级链、最近通知验证；敏感联系方式存安全系统 | SRE/运营 | 通用生产证据`ONCALL_ROSTER` |

## 7. 固定材料机器记录

以下31条记录与第1、4、5、6节固定资料一一对应，必须全部保留且身份唯一。归档时把`OPEN`改为
`APPROVED`或`NOT_APPLICABLE`，并填写绝对证据路径与SHA-256；不适用也必须引用批准该结论的终态记录，
不能留空。允许多项引用同一份批准材料，但每条记录都会独立验文件、哈希和终态占位符。

格式：`OPTION_OPERATIONS_MATERIAL: <material_id>|<APPROVED或NOT_APPLICABLE>|<absolute_evidence_path>|<sha256>`

```text
OPTION_OPERATIONS_MATERIAL: tenant_legal_jurisdiction|OPEN||
OPTION_OPERATIONS_MATERIAL: target_market_rules|OPEN||
OPTION_OPERATIONS_MATERIAL: release_identity|OPEN||
OPTION_OPERATIONS_MATERIAL: change_window|OPEN||
OPTION_OPERATIONS_MATERIAL: accountable_roles|OPEN||
OPTION_OPERATIONS_MATERIAL: contract_series_scope|OPEN||
OPTION_OPERATIONS_MATERIAL: launch_bundle|OPEN||
OPTION_OPERATIONS_MATERIAL: contract_set_reconciliation|OPEN||
OPTION_OPERATIONS_MATERIAL: launch_bundle_verification|OPEN||
OPTION_OPERATIONS_MATERIAL: product_scope_runtime|OPEN||
OPTION_OPERATIONS_MATERIAL: annual_calendar|OPEN||
OPTION_OPERATIONS_MATERIAL: calendar_version|OPEN||
OPTION_OPERATIONS_MATERIAL: temporary_halt|OPEN||
OPTION_OPERATIONS_MATERIAL: corporate_action|OPEN||
OPTION_OPERATIONS_MATERIAL: series_policy|OPEN||
OPTION_OPERATIONS_MATERIAL: insurance_ledger|OPEN||
OPTION_OPERATIONS_MATERIAL: daily_conservation|OPEN||
OPTION_OPERATIONS_MATERIAL: platform_backstop|OPEN||
OPTION_OPERATIONS_MATERIAL: insurance_inventory_exit|OPEN||
OPTION_OPERATIONS_MATERIAL: portfolio_margin|OPEN||
OPTION_OPERATIONS_MATERIAL: mmp|OPEN||
OPTION_OPERATIONS_MATERIAL: risk_parameter_change|OPEN||
OPTION_OPERATIONS_MATERIAL: physical_default|OPEN||
OPTION_OPERATIONS_MATERIAL: public_market|OPEN||
OPTION_OPERATIONS_MATERIAL: complex_orders|OPEN||
OPTION_OPERATIONS_MATERIAL: american_exercise|OPEN||
OPTION_OPERATIONS_MATERIAL: beanstalk_capacity|OPEN||
OPTION_OPERATIONS_MATERIAL: orchestrator_takeover|OPEN||
OPTION_OPERATIONS_MATERIAL: alert_delivery|OPEN||
OPTION_OPERATIONS_MATERIAL: user_materials|OPEN||
OPTION_OPERATIONS_MATERIAL: oncall_roster|OPEN||
```

## 8. 资料完成判定

- [ ] 本次所有合约都在`option-contract-launch-bundle.md`，逐合约检查表为`APPROVED`且
  `option-launch-bundle-verify.sh`对数量、身份、六方审批、63项正文/勾选及哈希全部PASS。
- [ ] 最终公告、配置、系列生成明细和上市合集的合约集合完全一致。
- [ ] `option-contract-set-reconciliation.md`绑定两份原始导出及SHA，审批、目标环境和公告/客户端三方集合严格相同。
- [ ] 所有适用专项模板已从`DRAFT/OPEN`进入批准或关闭终态；不适用项写明依据而不是留空。
- [ ] 最终发布材料不包含`待填写`、方括号占位符、空链接、空联系人或未确定时区。
- [ ] 未批准能力在配置、API、权限、公告、客服和销售口径中均保持关闭。
- [ ] 预生产报告绑定相同release/image，原始附件和报告SHA-256全部固定；所有生产报告通过
  `option-evidence-finalization-verify.sh`，无表格占位符、空单元格、未决状态或未勾选项；readiness声明
  通过`option-readiness-attestation-verify.sh`，全部键唯一、完整且无拼写/占位路径。
- [ ] 产品、技术、风控、清算/财务、运营和合规/法务六方完成签署。

## 9. 状态与移交

资料总数按`31 + 10 × 本次合约总数`计算：31是第7节逐条哈希绑定的发布身份、市场/日历、资金风险和运行运营四组固定资料，
每个合约另有第3节的10组参数。归档时必须把公式替换为整数；校验器从已批准上市合集读取合约数量，禁止
人工少算、抽样或让汇总与合约全集脱节。

| 汇总项 | 数量 |
| --- | ---: |
| 必填资料总数 | 31 + 10 × 本次合约总数；归档时必须改为整数 |
| 已批准 | 待填写 |
| 不适用且有依据 | 待填写 |
| OPEN/REJECTED | 待填写 |
| 未关闭SEV-1/2或资金差异 | 待填写；必须为0 |

只有全部必填资料批准、非适用项有依据、未关闭严重问题为0且六方签署完成时，归档副本才允许改为
`OPTION_OPERATIONS_INPUT_STATUS: APPROVED`。校验器会从31条固定材料记录和上市合集反算汇总，人工填写的
“已批准/不适用”数字与底层记录不一致即失败。本清单批准不替代生产readiness；两者必须同时通过。
