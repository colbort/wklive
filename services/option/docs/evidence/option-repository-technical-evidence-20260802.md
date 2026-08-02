# Option 仓库技术证据（2026-08-02）

EVIDENCE_SCOPE: REPOSITORY_ONLY

EVIDENCE_STATUS: PASSED_NOT_RELEASE_CANDIDATE

TEST_DATA_CLASSIFICATION: SYNTHETIC_ISOLATED

PLATFORM_BACKSTOP_SCOPE_STATUS: REPOSITORY_PASSED_PRODUCTION_APPROVAL_PENDING

## 1. 使用限制

本报告只证明当前共享工作区的仓库实现和隔离真实 Asset RPC 门禁。它不是预生产或生产证据，不能填写
真实 tenant、release image、Prometheus/Alertmanager 通知、容器编排、财务日终、业务参数或人员签字。

正式隔离门禁执行时的Git HEAD为`adbab9b72885a4fcd692124a15421fab8e423b03`且工作区有166条
未提交状态；其后代码和证据先后由外部提交，当前基础HEAD为
`7d81ffcb43b2dc4d1adab70dca3548e4ad8ad191`。本轮又在该提交上补强完整合约上市合集、运营输入总表和
生产证据状态门禁，并增加逐合约数量/身份/正文/审批/哈希、三方集合对账校验器及可重复自测；进一步
补齐默认关闭的可选产品运行时门禁、无歧义渲染配置比对、运营输入机械核对及MMP/美式条件证据；并新增
通用生产证据终态校验，拒绝仅改`APPROVED`状态但保留占位符、空表格、未决状态或未勾选项；同时以
readiness示例为schema拒绝生产声明的重复、未知、缺失键及占位路径；外部首批输入另收敛为最小移交包，
完整合约CSV以独立校验器检查22列schema、唯一性和经济/时间边界，不再要求责任方从全部技术模板反推字段。
当前仍有52项未提交Option范围变化。因此本报告仍不
存在最终可部署的不可变release/image组合。必须审查并提交本轮材料，
在干净提交上重跑门禁；本报告不能作为发布身份。

## 2. 执行环境

| 字段 | 实际值 |
| --- | --- |
| 执行日期 | 2026-08-02 |
| 本轮假数据复验窗口 | 2026-08-02 19:06～19:18 HKT |
| 宿主 | Darwin arm64 |
| Go | `go1.26.4 darwin/arm64` |
| goctl | `1.9.2 darwin/arm64` |
| Node.js | `v20.20.2` |
| Docker Server | `29.6.2` |
| 当前本机 Beanstalk 镜像 | `sha256:ffb07f529a45a56bf456c1920d31297638623ecb4894ac2bb639b99a66eb634e`，`linux/arm64` |
| 正式门禁 MySQL | `mysql:8.4`，linux/arm64，image `sha256:5e7e005a680e75d935984d3d9390990d2a709b3ed67e92708e9e6747f1f754c9` |
| 隔离测试 tenant_id | `996031`（仅测试） |

## 3. 已执行命令与结果

| 命令 | 结果 | 说明 |
| --- | --- | --- |
| `make gen-model`（Asset） | PASS | 平台政策/日用量/cover DDL后实际执行；手写模型扩展保留 |
| `GOCACHE=/private/tmp/wklive-go-build-cache make gen`（Asset） | PASS | Asset API、共享proto和生成服务完成同步 |
| `GOCACHE=/private/tmp/wklive-go-build-cache go test ./... && go vet ./...`（Asset） | PASS | 全Asset包；政策边界、UTC日界、管理逻辑通过 |
| `GOCACHE=/private/tmp/wklive-go-build-cache go test ./...`（Option） | PASS | 全 Option 包 |
| `GOCACHE=/private/tmp/wklive-go-build-cache go vet ./...`（Option） | PASS | 全 Option 包 |
| `GOCACHE=/private/tmp/wklive-go-build-cache go test -count=1 -race ./models ./internal/logic/helpers ./internal/logic/app ./internal/logic/admin ./internal/logic/task` | PASS | 非缓存竞态复验覆盖模型、产品范围helper、App/Admin入口与任务状态机 |
| `go test ./... && go vet ./...`（Admin API） | PASS | 管理 API 回归 |
| `npx prettier --check ... && npm run type-check`（Admin UI） | PASS | 中英文运营标签与类型检查 |
| `./monitoring/option-production-readiness.sh --repository-only` | PASS | 仓库项通过；本机缺少`promtool`按repository mode跳过 |
| `deploy/beanstalk-readiness.sh`（当前本机Compose运行态） | PASS | 主备协议健康；镜像架构均匹配arm64宿主，容器内均为`aarch64`，各自WAL目录均为独立Docker volume；仅证明本机警告已消失，不替代目标AMD64/ARM64容量与RTO签署 |
| `services/option/acceptance/run-p0-asset-rpc-e2e.sh`（19:06～19:18 HKT假数据复验） | PASS | 合成tenant/合约/账户，Docker隔离库、隔离Redis/etcd key、真实Asset gRPC、迁移双执行和故障注入；结束自动清理 |
| `git diff --check` | PASS | 当前差异无空白错误 |
| `go test ./internal/logic/task ./internal/logic/admin`（平台兜底增量） | PASS | 默认关闭闸门、任务与管理逻辑编译/单测通过 |
| `./acceptance/run-platform-backstop-schema-acceptance.sh` | PASS | Asset迁移双执行，3表/9触发器/8快照列，6类直SQL旁路均拒绝 |
| `./acceptance/run-platform-backstop-rbac-acceptance.sh` | PASS | System迁移双执行，权限ID唯一且申请/复核角色写权限互斥 |
| `./acceptance/run-platform-backstop-rpc-acceptance.sh` | PASS | 真实Asset RPC边界、20并发、重放、版本切换、补资及Option穿仓响应丢失 |
| `./monitoring/option-production-readiness.sh --repository-only`（P0-007实现后） | PASS | 校验硬额度实现、管理材料和证据清单；仍不代表生产批准 |
| `./monitoring/option-release-scope.sh --scope-only` | PASS | `7d81ffcb`后当前为`changed=52 modified=32 added=0 untracked=20`；均为Option实现、测试、运营材料/门禁，无删除、重命名、冲突或范围外路径 |
| `option-launch-bundle-verify-selftest.sh` | PASS | 可重复的一条正常样例通过；声明2/实际1、63项未勾选、正文改写、公告集错配、目标合约提前进入TRADING五条反向样例均拒绝；逐合约身份、六方审批、上市窗口、原始导出与哈希同步校验 |
| `option-product-scope-verify-selftest.sh` | PASS | 正常声明/渲染配置通过；值不匹配、YAML重复键、YAML缺键、环境重复声明、未实现Greeks伪开启，以及组合保证金/实物/复杂单/MMP/美式脱离卖方交易五类依赖均拒绝 |
| `option-operations-input-verify-selftest.sh` | PASS | 全批准及有证据不适用两类正常输入通过；未勾选、OPEN、占位符、固定材料缺失/重复/状态未决/哈希不符/证据未终态、汇总不平、上市合约数不匹配、合集未批准/记录数不符和严重问题非零十三类反向样例均拒绝 |
| `option-evidence-finalization-verify-selftest.sh` | PASS | 正常终态报告通过；DRAFT状态、表格/正文方括号占位符、未决选择、空单元格、OPEN行、正文/表格未勾选项和部署占位符十类反向样例均拒绝 |
| `option-readiness-attestation-verify-selftest.sh` | PASS | 完整唯一键集合通过；重复键、缺键、未知键、非规范赋值和占位路径五类反向样例均拒绝 |
| `option-external-contract-intake-verify-selftest.sh` | PASS | 两合约完整CSV通过；表头漂移、重复身份、占位值、非法枚举、数量上下界、五时间、时区及空合集八类反向样例均拒绝 |
| 运营材料状态与上市集合门禁 | PASS | 所有专项模板具有DRAFT/OPEN/DEFERRED机器状态；生产必须提供完整合约上市合集、逐表SHA-256及已批准运营输入总表；31项固定材料逐条绑定终态文件与哈希，批准/不适用汇总由记录反算，不能只改汇总或抽样一个合约 |
| `GOCACHE=/private/tmp/wklive-go-build-cache go test ./... && go vet ./...`（Option，`7d81ffcb`后复核） | PASS | 外部提交后重新验证104个RPC对应104个Server方法；产品范围补强后再次全量通过 |

## 4. 可选产品运行时范围

- `ProductScope`零值失败关闭卖方开仓、组合保证金、实物交割、复杂订单、公开期权链/盘口、MMP和美式
  提前行权；Option示例配置全部显式为false。
- PENDING→TRADING和PAUSED→TRADING都复核合约能力；普通订单只阻止新增风险并保留CLOSE；组合单在
  请求和资金全部成功后的激活点二次检查，关闭时进入释放流程；公开行情与美式行权入口独立拒绝。
- 生产门禁必须取得Option渲染YAML及SHA-256，并把七个运行时开关逐项与发布声明比较；声明、
  `ProductScope`段和键都必须唯一且完整。没有独立功能
  入口的Greeks依赖扩展不能伪装成运行时能力，当前必须保持false。
- 组合保证金、实物、复杂单、MMP和美式提前行权启用时必须同时开启并批准卖方交易；MMP与美式还分别
  提交保留MMP-PRE-001～008、AMER-PRE-001～010及六方签署的专项哈希报告。
- 本项没有DDL、Proto或模型变化，因此无需执行`make gen-model`或协议生成；新增helper单测验证零值关闭、
  组合保证金/实物优先级、CLOSE保留及MMP关闭。

## 5. 合成数据隔离链路摘要

- 数据分类为`SYNTHETIC_ISOLATED`：业务数据均为假tenant、假合约、假账户和假资金，但调用真实Option逻辑、
  真实Asset gRPC、真实MySQL/Redis及正式迁移；可证明功能、事务、幂等、并发与故障恢复，不证明真实市场、
  资本、法务参数或生产签署。
- 主 Option/Asset 集成测试：`115.718s`。
- 资金指令：9277条；9270条成功且已对账；7条冻结前合法取消；加权终态9284。
- 501空头指派：503条资金指令全部成功对账。
- 5000空头指派：5002条资金指令全部成功对账；Asset RPC阶段`2m50.817s`。
- 501现金到期多头/502持仓：1004条资金指令全部成功对账；Asset RPC阶段`53.654s`。
- 501实物交割单元：2004条资金指令全部成功对账；Asset RPC阶段`1m43.268s`。
- 独立行权/实物工作进程在Asset提交后被`SIGKILL`，30秒租约自然到期后唯一接管并保持流水唯一。
- 最后交易边界：合约PAUSED、订单CANCELED、原因`CONTRACT_LAST_TRADE_ENDED`，提前行权/结算均为0。
- 主门禁中较早的穿仓响应丢失用例形成于旧平台账户语义，只保留为历史链路基线；本轮另以正式Asset政策
  和硬额度RPC门禁重新证明业务号重放、资金底线与日累计，不再把旧结果解释为额度证据。

## 6. 保险基金流水专项

- 新行原始`amount`只接受正业务绝对金额；1/3类型按`+ABS`、2/4类型按`-ABS`读取。
- 真实缺口赔付原始金额15、方向归一金额-15。
- 汇总输出：`raw_inflow=28.49 raw_outflow=54.10 signed_net=-25.61 nonpositive_rows=0`。
- 负数新增、既有流水UPDATE和DELETE均由MySQL拒绝；迁移连续执行两次成功。
- 历史流水未改写；生产仍须逐笔关联`asset_flow_no`并由财务/清算签署。

## 7. 平台兜底仓库验收

- Asset已移除无限负余额方法，并实现逐租户/逐币`DISABLED/PREFUNDED/CREDIT_FLOOR`不可变版本、异人
  复核、单请求/UTC日累计/余额底线及事务内政策和资金快照。
- 真实RPC边界验证包括缺失/草稿/拒绝/过期/关闭政策、最小单位越界、20并发恰好10成功/10拒绝、
  响应丢失重放、同日版本切换和补资不清零日用量；汇总为cover/平台流水各15、总额63。
- 正式重跑曾发现连续业务拒绝会被sqlx误计为数据库故障；现已将明确业务状态码列为可接受回滚，并以
  单测及同一真实RPC序列证明连续拒绝后正常请求不被熔断。系统/依赖错误仍不在白名单。
- Admin RPC/API/UI和System RBAC已接通；申请/复核角色写权限互斥，数据库触发器阻止直接伪造、改写或
  删除政策、日用量和cover证据。
- Option仍以`PlatformBackstop.Enabled=false`默认失败关闭；无有效政策时不上市、不恢复，已有清算缺口
  转人工。仓库`OPT-P0-007`已通过，真实资金模型、额度、人员、目标环境BST/告警/日终及六方签署仍阻断生产。

## 8. 明确未证明

- 干净、不可变的 release commit 和 Option/Asset 镜像 digest。
- 目标预生产存量备份升级、生产配置、权限和数据库独立安全审计。
- AMD64 Beanstalkd容量、24小时长稳、真实Pod编排墙钟强杀/RTO。
- 生产Prometheus抓取、`promtool`、Alertmanager电话/IM/案件及恢复通知。
- 真实财务日终、保险历史盘点、平台兜底资本来源/实际政策参数/目标环境运行证据、组合模型独立验证和
  业务/法务参数。
- 生产公开行情/CDN外部探针、复杂订单批准容量及真实用户通知。

以上任一项不得从本报告推断为通过。完整剩余项见
`docs/option-current-status-and-production-blockers.md`。

## 9. 完整性

同目录 `.sha256` 文件固定本报告及相关实现、迁移、门禁和运营文档。执行：

```bash
cd <repository-root>
shasum -a 256 -c services/option/docs/evidence/option-repository-technical-evidence-20260802.sha256
```

任一哈希变化后，本报告状态自动失效，必须审查变化、重跑相应门禁并生成新日期/版本证据；禁止只更新哈希。
