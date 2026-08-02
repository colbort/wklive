# Option 生产阻断与证据材料覆盖矩阵

更新时间：2026-08-02

`MATRIX_STATUS: REPOSITORY_MATERIALS_READY / PRODUCTION_BLOCKED`

## 1. 状态定义

- `MATERIAL_READY`：仓库已提供可执行模板、系统事实和验收标准；真实参数、环境数据和签署仍未完成。
- `PARTIAL_MATERIAL_GAP`：已有通用清单或脚本，但缺少能独立归档、签署和被生产门禁验证的专项记录。
- `EXTERNAL_EXECUTION`：材料完整，但必须由目标环境或业务角色执行，仓库不能代填。
- `DEFERRED`：不属于首版一般零售期权范围，保持无入口，不作为本轮材料缺口。

“材料完成”不等于“生产阻断关闭”。只有实际归档副本为 `APPROVED`、附件哈希匹配且生产 readiness
通过，才能在发布签署单中关闭对应阻断。

## 2. 阻断—材料—验收映射

| 生产阻断 | 仓库技术基线 | 主要验收证据 | 仓库材料/入口 | 材料状态 | 仍需外部提供/执行 |
| --- | --- | --- | --- | --- | --- |
| 保险基金账务 | 正绝对金额、类型定向、不可改删、真实赔付归一 | 历史逐笔Asset桥接、逐币影子复算、下游盘点、财务语义批准 | `templates/option-insurance-fund-ledger-production-approval.md`、`option-p1-008-insurance-fund-ledger-repository-acceptance.md` | MATERIAL_READY / EXTERNAL_EXECUTION | 生产快照、财务/清算结论、异常案件和六方签署 |
| 日终资金守恒 | scope=1/2、不可变运行/明细、成功心跳和差异案件 | 同截止点逐币期初+净流量=期末、差异恢复、36小时心跳和四方签署 | `templates/option-daily-fund-reconciliation.md`、`option-daily-conservation-contract.md` | MATERIAL_READY / EXTERNAL_EXECUTION | 真实Asset快照、生产调度、通知和财务日终签署 |
| 卖方保险库存与平台兜底额度 | 五项保险退出硬限额、四眼只减仓IOC；平台兜底三模式、事务硬额度、不可变政策/cover和默认关闭闸门 | 仓库边界/20并发/重放/旁路/RBAC已通过；生产需三选一模式、真实资本额度、目标环境告警日终和批准退出 | `templates/option-insurance-inventory-exit-approval.md`、`templates/option-platform-backstop-policy-approval.md`、`option-p0-007-platform-backstop-runtime-limit-design.md`、`option-p0-007-repository-acceptance.md` | REPOSITORY_PASSED / EXTERNAL_EXECUTION | 选择并批准模式；资本、真实参数、实名角色、目标环境BST/告警/日终和六方签署 |
| 组合保证金模型 | 版本化参数、未来生效/回滚、订单准入谱系、情景模型 | 代表性回测、压力样本、独立模型验证、版本切换和回滚 | `templates/option-portfolio-risk-validation-record.md` | MATERIAL_READY / EXTERNAL_EXECUTION | 数据集、独立验证人、报告与预生产复演 |
| 实物交割违约 | 单元隔离、先收后付、补资/违约态、原指令恢复 | 补资边界、通知、逐场景最终处置、逐币会计、司法辖区和六方签署 | `templates/option-physical-delivery-default-policy-approval.md`、`option-physical-delivery-default-record.md` | MATERIAL_READY / EXTERNAL_EXECUTION | 产品/法律政策；任何获批但未实现路径须另立工程验收 |
| 目标市场规则 | 版本日历、五时间、公司行动V1、系列生成治理 | 官方来源、IANA时区、年度例外、到期/行权政策、公司行动和系列规则 | `templates/trading-calendar-approval.md`、`trading-calendar-annual-review.md`、`corporate-action-case.md`、`contract-series-approval.md` | MATERIAL_READY / EXTERNAL_EXECUTION | 目标市场、官方来源、司法辖区和业务/法务签署 |
| 产品公告与用户运营 | 系统能力边界、通知/FAQ/客服底稿 | 具体合约参数、法律文本、触达计划、联系人、无占位符发布副本 | `option-product-operations-pack.md`、`option-contract-launch-checklist.md` | MATERIAL_READY / EXTERNAL_EXECUTION | 产品参数、法律文本、渠道/人员和发布时间 |
| 生产监控与通知 | 61条规则、18项索引、指标采样/新鲜度门禁 | Prometheus加载、Alertmanager SEV-1/2/3送达、案件/恢复/失败升级 | `templates/option-alert-delivery-test.md`、`option-preproduction-evidence-pack.md` | MATERIAL_READY / EXTERNAL_EXECUTION | 生产target、receiver、值班系统和通知回执 |
| 多实例/容器故障 | 独立进程SIGKILL、租约自然失效、唯一接管 | 真实Pod/容器ID、编排事件、墙钟RTO、重复/乱序消息和资金唯一 | `templates/option-orchestrator-takeover-report.md`、预生产证据包、运行手册 | MATERIAL_READY / EXTERNAL_EXECUTION | 目标编排器执行ORC-001～ORC-010、真实墙钟/实例/资金/告警证据和五方签署 |
| Beanstalkd原生架构与容量 | 多架构固定镜像、独立WAL、ARM64 1000任务强杀恢复 | ARM64/AMD64原生、批准峰值、旧连接断开、WAL完整恢复、RTO | `templates/option-beanstalk-capacity-rto-report.md`、`deploy/beanstalk-readiness.sh`、`beanstalk-resilience-smoke.sh` | MATERIAL_READY / EXTERNAL_EXECUTION | AMD64及目标ARM64执行BS-001～BS-010、24h长稳、批准阈值和四方签署 |
| 对外行情与复杂订单 | 公开链/盘口/OI一致性、组合单原子资金屏障 | CDN/缓存/限流/SLA外部探针，组合热点/消息故障/容器恢复 | `templates/public-market-readiness.md`、`complex-order-readiness.md` | MATERIAL_READY / EXTERNAL_EXECUTION | 目标网关/CDN、批准SLA、真实容量和通知 |
| 数据库安全审计 | 应用拒绝、不可变触发器和安全运行手册 | 直接SQL/DDL拒绝、事务外事件、身份链、误报控制、采集失败/旁路告警 | `templates/option-database-security-audit-acceptance.md` | MATERIAL_READY / EXTERNAL_EXECUTION | MySQL审计/堡垒机/SOC配置、12+5场景执行和六方签署 |
| 无专项模板的生产报告 | 门禁变量、SHA-256和状态校验已落地 | 迁移、真实Asset E2E、通用故障、容量、值班、日历预生产、模型版本切换、保险方向核对 | `templates/option-production-evidence-report.md`，每种证据单独复制 | MATERIAL_READY / EXTERNAL_EXECUTION | 真实release、环境、原始附件、批准边界、执行/复核人和签署；必须为`OPTION_EVIDENCE_STATUS: APPROVED` |
| RFQ/大宗/做市义务 | 无生产入口；MMP不冒充义务 | 获批机构规则后重新立项 | `templates/institutional-market-readiness.md` | DEFERRED | 产品、资格、监管和市场规则未批准 |

## 3. 仓库下一步材料优先级

按生产风险和可执行性排序：

1. `OPT-P0-007`：仓库原子额度、边界、20并发、响应丢失、旁路和管理权限已完成；业务必须依据审批
   模板选择预注资、硬授信或永久人工，补真实资本/额度、目标环境BST、日终桥接、SEV-1送达和六方签署。
2. `ORCHESTRATOR_TAKEOVER_REPORT`：模板已补；目标环境执行Option/Asset/队列/数据库真实容器故障、
   原租约、竞争接管、消息重复/乱序、资金唯一、通知和RTO时间线。
3. `BEANSTALK_CAPACITY_REPORT`：模板已补；目标ARM64/AMD64执行镜像身份、写入峰值、WAL积压、
   SIGKILL、旧连接断开、新连接重建、完整消费、24h长稳和批准RTO签署。

上述专项材料和仓库实现已经补齐；目标环境执行和签署仍保持`PRODUCTION_BLOCKED`。

## 4. 材料完整性验收标准

- 每个非`DEFERRED`阻断都有唯一主要证据记录、责任人、执行环境、版本、时间范围和SHA-256。
- 模板明确区分仓库预填事实和外部必填值；未执行副本默认`DRAFT`，不得预填`APPROVED`。
- 资金/风险材料包含守恒、幂等、失败屏障、回滚和差异案件；运行材料包含时间线、实例身份、RTO和通知。
- 所有生产门禁引用的报告除文件存在和哈希外，还校验机器可读批准状态。
- 未批准能力保持功能开关、API、权限、公告和销售口径关闭，不以人工流程绕过。

当前结论：仓库能完成的专项材料及平台兜底工程已经形成，其余项目等待目标环境执行和业务签署；
Option整体为`REPOSITORY_PASSED / PREPROD_BLOCKED`。材料完整和仓库通过绝不等于生产就绪。
