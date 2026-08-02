# Option 完成度审计、剩余事项与验收标准

更新时间：2026-08-02

AUDIT_STATUS: REPOSITORY_ACTIONS_COMPLETE / PREPROD_BLOCKED

## 1. 结论

`services/option`当前设计已覆盖一般标准化零售期权所需的合约、订单、撮合、资金、风险、行权、到期、
结算、行情和运营控制。P0、P1以及P2-001～P2-007的仓库实现、迁移、生成代码、自动化和隔离真实
Asset RPC证据已经完成；不能据此宣称生产`DONE`，因为真实市场参数、目标环境演练、财务/风险/法律
批准和最终签署尚未完成。

2026-08-02 19:06～19:18 HKT复验的数据分类为`SYNTHETIC_ISOLATED`：tenant、合约、账户和余额为
合成数据，Option业务逻辑、Asset gRPC、MySQL、Redis和迁移均为真实链路。该证据足以验收仓库功能、
事务、幂等、并发、资金守恒与故障接管，但不能代替真实市场规则、生产资本、目标环境或人员批准。

P2-008 RFQ、大宗和正式做市义务不属于首版一般零售期权，当前保持`DEFERRED`且无入口。MMP只是做市
保护，不等于做市义务。

本次复核关闭了一个P0配置一致性缺口：原生产门禁声明的卖方、组合保证金、实物、复杂单和公开行情等
开关并未全部进入Option运行时。现已增加默认关闭的`ProductScope`，在合约上市、普通订单、
组合单、公开行情和美式提前行权入口执行真实门禁，同时补充MMP与美式开关；生产验收必须提交渲染配置及
SHA-256，脚本逐项比较声明值和运行时值。尚无独立入口的Greeks依赖扩展禁止声明为开启。该修改不涉及
DDL、Proto或模型，无需运行`make gen-model`。独立`option-product-scope-verify.sh`还要求每项声明、
`ProductScope`段及七个运行时键恰好出现一次；正常及不匹配、重复、缺失和伪开Greeks反向样例已自动化。

## 2. 一般期权要求与当前覆盖

| 能力域 | 一般要求 | 当前结论 | 主要台账 |
| --- | --- | --- | --- |
| 合约与市场规则 | Call/Put、欧/美式、现金/实物、五时间、交易日历、系列和公司行动 | 仓库完成；真实市场规则待批准 | OPT-P0-003/005、P2-001～005 |
| 订单与撮合 | 限价/市价/IOC/FOK/Post-only、STP、MMP、撤单竞态、复杂单原子性 | 仓库完成；真实容量、网关和通知待验收 | OPT-P1-002/003、P2-007 |
| 资金与账务 | 唯一资金池、冻结/扣减/释放幂等、逐币守恒、费用、保险和兜底硬额度 | 仓库完成；保险语义、资本、额度和日终须外部批准 | OPT-P0-001/002/004/007、P1-008 |
| 风险与清算 | 净期权市值、逐仓/组合保证金、强平、保险接管、模型版本治理 | 仓库完成；独立模型验证和生产参数待签署 | OPT-P1-001/004/005 |
| 行权与到期 | 正式结算价、AUTO/DNE、FIFO指派、现金结算和实物违约屏障 | 仓库完成；正式来源、阈值、违约经济/法律政策待批准 | OPT-P0-005/006、P1-006/007、P2-002/003 |
| 行情与运营 | 期权链、盘口、24h/OI、指标告警、人工恢复、审计和发布治理 | 仓库完成；生产Prometheus、通知、DB审计和公网探针待执行 | OPT-P1-008、P2-006 |

## 3. 实施项状态

| 实施项 | 仓库状态 | 生产前仍需完成 |
| --- | --- | --- |
| OPT-P0-001～P0-007 | REPOSITORY_PASSED | 真实Asset/消息/容器复演；行情/结算/资金批准；平台兜底模式、资本、额度、BST与六方签署 |
| OPT-P1-001～P1-008 | REPOSITORY_PASSED | 生产风险参数、模型独立验证、实物违约政策、保险历史盘点、监控通知和运营签署 |
| OPT-P2-001～P2-007 | REPOSITORY_PASSED | 权威市场日历/公司行动/系列规则，网关/CDN/SLA、复杂单容量和生产编排验收 |
| OPT-P2-008 | DEFERRED | 只有机构产品、资格、监管和市场规则获批后才重新立项 |

逐项实现和测试细节以`option-remediation-plan.md`为权威台账；本文件不复制上千行运行结果，避免多处
数字漂移。

## 4. 尚未完成但不能由仓库代填的事项

| 阻断 | 必须提供的真实输入/执行 | 关闭标准 |
| --- | --- | --- |
| 产品范围与市场规则 | 目标市场、IANA时区、年度日历、五时间、结算/行权规则、系列及公司行动来源 | 四眼审批记录为`APPROVED`，配置与证据SHA匹配 |
| 保险账本与日终 | 生产备份、历史逐笔Asset桥接、逐币影子复算、差异案件、36小时成功心跳 | 财务/清算/风控/技术/合规签署，差异为0或案件全部关闭 |
| 卖方与平台兜底 | 选择DISABLED/PREFUNDED/CREDIT_FLOOR，批准资本、单笔/日累计/余额底线 | 绑定准确Asset政策ID/版本，BST-001～012、告警、日终和六方批准全部通过 |
| 模型与实物违约 | 代表性数据、独立验证人、模型版本；补资期限、最终处置和司法辖区 | 报告与运行配置版本一致；批准路径均已实现并完成两币种守恒及故障验收 |
| 生产基础设施 | 目标编排器、Prometheus/Alertmanager、MySQL审计、Beanstalk原生双架构和WAL | 容器强杀、自然租约、唯一接管、通知送达、容量/RTO和安全旁路证据全部`APPROVED` |
| 对外运营 | 最终公告、FAQ、客服、联系人、触达计划、网关/CDN/限流/SLA | 发布副本无占位符；公网探针、复杂单E2E和上线签署单通过 |

这些项目保持`PREPROD_BLOCKED`不是代码遗漏。缺少真实环境、人员或经济/法律决策时，不允许技术人员
伪造参数或签名。

## 5. 仓库已补齐的运营材料

运营真实输入、责任方和专项证据入口已汇总到`option-operations-input-checklist.md`，避免由运营人员在
多份技术台账中自行猜测字段。

| 用途 | 模板/入口 | 默认状态 |
| --- | --- | --- |
| 上市与发布 | `option-contract-launch-checklist.md`、`option-contract-launch-bundle.md`、`option-contract-set-reconciliation.md`、`option-production-readiness-signoff.md` | DRAFT |
| 市场/日历 | `option-market-freshness-approval.md`、`trading-calendar-approval.md`、`trading-calendar-annual-review.md` | DRAFT |
| 资金/保险/兜底 | `option-daily-fund-reconciliation.md`、`option-insurance-fund-ledger-production-approval.md`、`option-platform-backstop-policy-approval.md`、`option-platform-backstop-e2e.md` | DRAFT |
| 风险与交割 | `option-portfolio-risk-validation-record.md`、`option-insurance-inventory-exit-approval.md`、`option-insurance-inventory-exit-execution-record.md`、`option-physical-delivery-default-policy-approval.md` | DRAFT |
| 运行与安全 | `option-alert-delivery-test.md`、`option-orchestrator-takeover-report.md`、`option-beanstalk-capacity-rto-report.md`、`option-database-security-audit-acceptance.md` | DRAFT |
| 行情与复杂订单 | `public-market-readiness.md`、`complex-order-readiness.md` | DRAFT |
| MMP与美式提前行权 | `option-mmp-readiness.md`、`option-american-exercise-readiness.md`及逐合约`option-exercise-expiry-control-record.md` | DRAFT |
| 无专项模板的报告 | `option-production-evidence-report.md`，每种证据分别复制 | DRAFT |

生产门禁同时校验文件、SHA-256和对应机器可读`APPROVED`状态，并统一拒绝表格占位符、空单元格、
未决选择、OPEN/DRAFT行及未勾选验收项；仓库模板本身不能直接作为生产证据。
本次发布还必须提供覆盖全部合约的上市合集及无占位符运营输入总表，避免用单合约抽样或一个总批准引用
替代逐合约参数复核。`monitoring/option-launch-bundle-verify.sh`按记录数量、唯一身份、租户/release、
六方审批引用、完整63项检查正文、勾选状态和逐文件SHA-256执行强制校验，并要求审批合集、目标环境
只读导出和最终公告/客户端导出三方集合完全相同；仅在文档中放一个哈希不能通过。
`monitoring/option-operations-input-verify.sh`进一步要求31项固定材料各有唯一机器记录，逐份校验绝对路径、
SHA-256和终态内容；再校验逐合约10组输入、8项完成条件、无OPEN/REJECTED/占位值和严重问题为0。
多合约总数按`31 + 10 × 合约数`与上市合集绑定，已批准/有依据不适用数量从底层记录反算，不能手工凑平。
所有由生产readiness引用的报告还会经过`monitoring/option-evidence-finalization-verify.sh`，防止只把
状态行改为`APPROVED`或重新计算哈希而保留未完成表格。
readiness声明本身再由`monitoring/option-readiness-attestation-verify.sh`按示例schema校验全部键精确一次，
拒绝未知/缺失键、非规范赋值和占位路径后才读取，避免重复tenant、release或证据SHA产生歧义。

## 6. 逐项验收标准

每一项只有同时满足以下条件才可从`PREPROD_BLOCKED`改为`DONE`：

1. 绑定干净且不可变的release commit/image digest，生成代码与DDL/Proto一致；涉及DDL必须执行
   `make gen-model`，涉及Proto/API必须执行相应`make gen`，并把生成差异纳入同一发布身份。
   Option平台兜底还必须把Asset schema/Proto/服务/模型、Admin API/UI和System RBAC作为同一不可拆分
   发布范围；`option-release-scope.sh --scope-only`必须为`SCOPE_OK`。
2. 迁移在空库和代表性存量备份上成功，连续执行两次，触发器/索引/历史兼容和前滚/回滚决策有证据。
3. 正常、边界、并发、重放、依赖超时、提交后响应丢失、进程/容器强杀和人工恢复均使用原业务身份，
   不产生重复资金、残单、残仓或冻结泄漏。
4. Option、Asset、订单、成交、持仓、保证金、保险和逐币钱包/流水守恒；所有差异案件已关闭。
5. 真实Prometheus/Alertmanager、值班案件、IM/电话通知及恢复消息送达，容量和RTO达到批准值。
6. 产品、技术、Asset/SRE、风控、清算/财务和合规/法务按适用范围签署；文件SHA-256与门禁配置一致。
7. `monitoring/option-production-readiness.sh`生产模式零失败；随后仍只能在批准变更窗口按租户、合约和
   单项功能开关逐步开放；Option渲染YAML中的`ProductScope`与门禁声明逐项一致并固定SHA-256。

## 7. 执行顺序

1. 冻结产品范围和市场规则，填写日历、五时间、结算、行权、公司行动和系列审批。
2. 完成保险语义、日终口径、卖方资本和平台兜底模式/额度批准；未批准时保持卖方或兜底关闭。
3. 部署不可变release到预生产，执行迁移、真实Asset、编排故障、Beanstalk容量、DB审计和告警送达。
4. 完成模型、实物违约、公开行情和复杂订单专项验收，关闭全部SEV-1/2及资金差异。
5. 固定每份报告SHA-256，完成生产签署并运行生产门禁；按能力最小化开放。

外部责任方无需从全部模板中反推首批字段；`option-external-input-handoff.md`已经把产品主体、目标市场、
完整合约CSV、逐合约参数、五时间、结算/行权、能力范围、环境release、资金决策和六方角色收敛成
最小移交包。完整合约CSV先由`option-external-contract-intake-verify.sh`检查22列schema、唯一身份、枚举、
币种、正数经济参数、数量上下界、Unix毫秒和五时间顺序。收到该资料前保持失败关闭，收到后再按本节
顺序生成真实归档副本。

## 8. Beanstalk架构警告

Docker Desktop显示`AMD64`表示容器镜像架构与宿主机不一致并可能使用模拟。当前Compose已改为从
`deploy/Dockerfile.beanstalkd`构建`wklive/beanstalkd:1.13-alpine3.20`多架构镜像，未固定
`linux/amd64`，主/备各自使用独立WAL卷。验收以`deploy/deploy.sh beanstalk-readiness`实际输出和
容器`Architecture`为准；如果界面仍显示旧的`schickling/beanstalkd`，说明运行的是旧容器，需要按
正常部署流程重建，而不是忽略警告。

2026-08-02已对当前本机Compose运行态执行`deploy/beanstalk-readiness.sh`：主备均healthy，镜像
`sha256:ffb07f529a45a56bf456c1920d31297638623ecb4894ac2bb639b99a66eb634e`为`linux/arm64`，两个容器内
`uname -m`均为`aarch64`且WAL均为Docker volume。因此用户截图中的AMD64模拟警告在当前本机已关闭；
该结果仍不是目标AMD64节点、批准峰值、24小时长稳或生产RTO证据。
