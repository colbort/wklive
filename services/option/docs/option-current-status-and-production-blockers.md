# Option 当前状态与生产阻断清单

更新时间：2026-08-02

## 1. 当前结论

- 一般零售期权仓库内P0/P1、P2-001～P2-007以及补充项`OPT-P0-007`已完成实现、迁移、自动化和隔离真实Asset RPC验收。
- 当前统一状态为`REPOSITORY_PASSED / PREPROD_BLOCKED`，不是生产`DONE`；全部生产参数、真实环境演练和签署仍按本清单阻断。
- P2-008 RFQ/大宗/做市义务保持 `DEFERRED`，无交易入口，不是首版一般零售期权阻断项。
- 本轮正式门禁：9277条资金指令，9270条成功且已对账，7条冻结前合法取消，加权终态9284。
- Asset平台兜底已实现逐租户/逐币三模式、单笔/UTC日累计/余额底线、四眼不可变政策和真实RPC并发边界；原无限负余额方法已移除。生产仍默认关闭，直到真实资金模式、资本和额度获批并完成目标环境告警/日终及六方签署。

## 2. 必须完成的生产阻断

| 阻断域 | 当前缺口 | 验收标准 | 责任方/证据 |
| --- | --- | --- | --- |
| 保险基金账务 | 正绝对金额+类型方向尚未获生产批准；历史及下游未盘点 | 按`templates/option-insurance-fund-ledger-production-approval.md`完成备份、逐笔`asset_flow_no`与Asset核对、逐币旧/新/Asset影子复算、下游消费者盘点、负向/回滚演练和六方签字 | 财务、清算、风控、Option/Asset技术、合规；保险账本生产审批记录及日终记录 |
| 日终资金守恒 | 仓库scope=1/2已实现，未在真实生产数据/调度/通知闭环签署 | 同一截止点逐币期初+净流量=期末；不完整即失败；差额案件、恢复和36小时心跳告警实测；技术/清算/财务/合规四签 | Asset、财务、清算、技术、合规 |
| 卖方与平台兜底 | 仓库硬额度及20并发/重放已通过；保险库存限额、平台资金模式/真实额度/资本、告警日终和六方签署未批准 | `PlatformBackstop.Enabled=false`保持失败关闭；按审批模板选择预注资、批准授信或永久人工，绑定已批准Asset政策ID/版本，完成目标环境BST-001～BST-012、日终/告警/故障报告和六方签署；首版补资/偿还不释放UTC日额度 | 风控、财务、清算、管理层、Option/Asset技术、合规；`option-p0-007-repository-acceptance.md`仅为仓库基线 |
| 组合保证金模型 | 缺生产代表性回测和独立模型验证 | 数据集、缺失处理、极端压力、限制条件、实现一致性和模型风险报告与生效版本哈希一致；未来切换/回滚复演通过 | 风控、独立验证人、清算 |
| 实物交割违约 | 最终经济处置、补资期、罚则和司法辖区未批准 | 按`templates/option-physical-delivery-default-policy-approval.md`逐场景批准补资边界、通知、会计分录、退款/继续交割/罚则/拍卖/平台承接或明确禁止路径；所有批准路径先实现并完成真实通知、容器故障、两币种守恒与六方签署；任何单元不垫资、不跨屏障 | 产品、风控、清算/财务、运营/客服、技术、合规/法务 |
| 目标市场规则 | 日历、到期/行权时间、公司行动、系列生成政策仍需权威来源 | 每个合约/市场保存来源、时区、批准人、版本、通知和演练；五时间满足系统不变量 | 产品、运营、法务/合规、风控 |
| 产品公告与用户运营 | 系统能力边界、公告/FAQ/通知/客服模板已补齐，但具体合约参数、法律文本、联系人、触达计划和发布时间未批准 | 最终发布副本无占位符/`待填写`/空联系人/空链接/未确定时区；最后交易、指令截止、到期和结算时间逐项匹配合约；上市检查表与生产签署单均为`APPROVED`并固定哈希 | 产品、运营、客服、风控、清算/财务、合规/法务；产品运营包及上市检查表 |
| 生产监控通知 | 示例配置具备，真实Prometheus/Alertmanager/值班案件未签署 | target可达；15秒采样与45秒新鲜度；61条规则加载；SEV-1/2/3通知送达、恢复关闭和失败升级留证 | SRE、值班、业务负责人 |
| 多实例/容器故障 | 仓库独立进程SIGKILL已过，真实编排器和墙钟RTO未签署 | 不手工清租约；容器ID、编排事件、真实墙钟、唯一接管、无重复资金和告警恢复均留证 | SRE、Option、Asset |
| Beanstalkd生产容量 | ARM64仓库WAL/强杀已过；AMD64与批准峰值未签署 | 原生架构镜像；独立WAL；批准峰值写入/积压；SIGKILL后RTO内恢复且任务不丢；旧连接重建 | SRE、容量负责人 |
| 对外行情/复杂订单 | 仓库一致性与原子性已过，真实缓存/CDN/容量/SLA未批准 | 公网探针、租户隔离、缓存TTL/降级/限流、复杂单真实并发与消息故障报告通过 | 行情、SRE、产品、风控 |
| 数据库安全审计 | 应用拒绝与不可变触发器已过，生产直SQL审计未签署 | 按`templates/option-database-security-audit-acceptance.md`证明12类直接SQL/DDL旁路被拒绝、数据不变、事务外事件持久、人员/会话/工单可追踪，合法操作不误报，采集失败/旁路成功告警和恢复闭环；六方签署并固定哈希 | DBA、安全/SOC、Option/Asset技术、风控/清算、合规/审计 |

## 3. 放行顺序

1. 冻结目标产品范围、结算方式、市场日历、五时间和卖方开关；不批准的高级能力保持关闭。
2. 财务/清算批准保险语义、平台兜底资金模型/真实硬额度和日终口径，再准备历史盘点与影子复算；生产政策、目标环境报告和六方签署未完成前不得开启平台兜底。
3. 风控完成组合模型独立验证、保险库存退出限额和目标合约参数审批。
4. 在预生产部署真实Prometheus/Alertmanager、消息、数据库和容器编排，按生产配置执行完整门禁。
5. 归档告警送达、故障接管、日终、容量、通知和权限隔离证据；任何差额或不完整记录未关闭不得继续。
6. 六方签署单全部完成后，才可按租户、合约和能力逐项开启；禁止用一个总开关一次开放全部能力。

## 4. 明确不应继续盲目开发的事项

- 未给目标市场规则前，不自行猜测节假日、公司行动、行权截止或交割违约经济处置。
- 未给生产数据和独立验证人前，不伪造组合保证金回测/模型签字。
- 未获财务/清算批准前，不批量重写历史保险流水。
- RFQ/大宗/做市义务保持无入口；MMP只是报价保护，不等于做市义务系统。
- 自动Delta对冲和跨币种折算只有在产品/风控给出币种、汇率源、时效、滑点、额度和失败处置后再立项。

## 5. 证据入口

- 总实施台账：`docs/option-remediation-plan.md`
- 设计审查：`docs/option-design-review.md`
- 生产验收矩阵：`docs/option-acceptance-test-plan.md`
- 生产门禁：`monitoring/option-production-readiness.sh`
- 仓库技术证据与哈希：`docs/evidence/option-repository-technical-evidence-20260802.md`
- 发布候选变更清单：`docs/evidence/option-release-candidate-change-inventory-20260802.md`
- 保险账本：`docs/option-p1-008-insurance-fund-ledger-repository-acceptance.md`
- 保险账本生产语义审批与历史盘点：`docs/templates/option-insurance-fund-ledger-production-approval.md`
- 平台兜底资金模型、硬额度与工程验收：`docs/templates/option-platform-backstop-policy-approval.md`
- 生产阻断与材料覆盖矩阵：`docs/option-production-blocker-evidence-matrix.md`
- 完成度审计、剩余事项和逐项验收：`docs/option-completion-audit.md`
- 产品公告、FAQ、通知与客服底稿：`docs/option-product-operations-pack.md`
- 单合约上市门禁：`docs/option-contract-launch-checklist.md`
- 实物交割补资与违约经济处置政策：`docs/templates/option-physical-delivery-default-policy-approval.md`
- 数据库直接SQL与旁路安全审计：`docs/templates/option-database-security-audit-acceptance.md`
- 真实编排器故障注入与唯一接管：`docs/templates/option-orchestrator-takeover-report.md`
- Beanstalkd双架构容量、WAL与RTO：`docs/templates/option-beanstalk-capacity-rto-report.md`
- 生产签署单：`docs/templates/option-production-readiness-signoff.md`
- 预生产证据包：`docs/templates/option-preproduction-evidence-pack.md`
- 无专项模板时的生产证据底稿：`docs/templates/option-production-evidence-report.md`
