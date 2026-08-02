# Option 仓库技术证据（2026-08-02）

EVIDENCE_SCOPE: REPOSITORY_ONLY

EVIDENCE_STATUS: PASSED_NOT_RELEASE_CANDIDATE

PLATFORM_BACKSTOP_SCOPE_STATUS: REPOSITORY_PASSED_PRODUCTION_APPROVAL_PENDING

## 1. 使用限制

本报告只证明当前共享工作区的仓库实现和隔离真实 Asset RPC 门禁。它不是预生产或生产证据，不能填写
真实 tenant、release image、Prometheus/Alertmanager 通知、容器编排、财务日终、业务参数或人员签字。

正式隔离门禁执行时的Git HEAD为`adbab9b72885a4fcd692124a15421fab8e423b03`且工作区有166条
未提交状态；其后外部检查点提交和本轮平台兜底复核又改变了工作区。当前报告因此不存在可部署的不可变
release commit/image。必须先由负责人审查并形成干净发布提交，再在该提交上重跑门禁；本报告不能作为
发布身份，但可证明当前工作区的平台兜底硬额度仓库实现和隔离验收。

## 2. 执行环境

| 字段 | 实际值 |
| --- | --- |
| 执行日期 | 2026-08-02 |
| 宿主 | Darwin arm64 |
| Go | `go1.26.4 darwin/arm64` |
| goctl | `1.9.2 darwin/arm64` |
| Node.js | `v20.20.2` |
| Docker Server | `29.6.2` |
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
| `GOCACHE=/private/tmp/wklive-go-build-cache go test -race ./models ./internal/logic/task` | PASS | 本轮改动相关模型与任务竞态检查 |
| `go test ./... && go vet ./...`（Admin API） | PASS | 管理 API 回归 |
| `npx prettier --check ... && npm run type-check`（Admin UI） | PASS | 中英文运营标签与类型检查 |
| `./monitoring/option-production-readiness.sh --repository-only` | PASS | 仓库项通过；本机缺少`promtool`按repository mode跳过 |
| `services/option/acceptance/run-p0-asset-rpc-e2e.sh` | PASS | Docker隔离库、Redis、真实Asset gRPC、迁移双执行和故障注入 |
| `git diff --check` | PASS | 当前差异无空白错误 |
| `go test ./internal/logic/task ./internal/logic/admin`（平台兜底增量） | PASS | 默认关闭闸门、任务与管理逻辑编译/单测通过 |
| `./acceptance/run-platform-backstop-schema-acceptance.sh` | PASS | Asset迁移双执行，3表/9触发器/8快照列，6类直SQL旁路均拒绝 |
| `./acceptance/run-platform-backstop-rbac-acceptance.sh` | PASS | System迁移双执行，权限ID唯一且申请/复核角色写权限互斥 |
| `./acceptance/run-platform-backstop-rpc-acceptance.sh` | PASS | 真实Asset RPC边界、20并发、重放、版本切换、补资及Option穿仓响应丢失 |
| `./monitoring/option-production-readiness.sh --repository-only`（P0-007实现后） | PASS | 校验硬额度实现、管理材料和证据清单；仍不代表生产批准 |
| `./monitoring/option-release-scope.sh --scope-only` | PASS | `changed=104 modified=64 added=0 untracked=40`；Option与P0-007不可拆分Asset/Admin/System依赖均在精确范围内，无删除/重命名/冲突 |

## 4. 正式隔离链路摘要

- 主 Option/Asset 集成测试：`113.879s`。
- 资金指令：9277条；9270条成功且已对账；7条冻结前合法取消；加权终态9284。
- 501空头指派：503条资金指令全部成功对账。
- 5000空头指派：5002条资金指令全部成功对账；Asset RPC阶段`2m54.988s`。
- 501现金到期多头/502持仓：1004条资金指令全部成功对账。
- 501实物交割单元：2004条资金指令全部成功对账。
- 独立行权/实物工作进程在Asset提交后被`SIGKILL`，30秒租约自然到期后唯一接管并保持流水唯一。
- 最后交易边界：合约PAUSED、订单CANCELED、原因`CONTRACT_LAST_TRADE_ENDED`，提前行权/结算均为0。
- 主门禁中较早的穿仓响应丢失用例形成于旧平台账户语义，只保留为历史链路基线；本轮另以正式Asset政策
  和硬额度RPC门禁重新证明业务号重放、资金底线与日累计，不再把旧结果解释为额度证据。

## 5. 保险基金流水专项

- 新行原始`amount`只接受正业务绝对金额；1/3类型按`+ABS`、2/4类型按`-ABS`读取。
- 真实缺口赔付原始金额15、方向归一金额-15。
- 汇总输出：`raw_inflow=28.49 raw_outflow=54.10 signed_net=-25.61 nonpositive_rows=0`。
- 负数新增、既有流水UPDATE和DELETE均由MySQL拒绝；迁移连续执行两次成功。
- 历史流水未改写；生产仍须逐笔关联`asset_flow_no`并由财务/清算签署。

## 6. 平台兜底仓库验收

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

## 7. 明确未证明

- 干净、不可变的 release commit 和 Option/Asset 镜像 digest。
- 目标预生产存量备份升级、生产配置、权限和数据库独立安全审计。
- AMD64 Beanstalkd容量、24小时长稳、真实Pod编排墙钟强杀/RTO。
- 生产Prometheus抓取、`promtool`、Alertmanager电话/IM/案件及恢复通知。
- 真实财务日终、保险历史盘点、平台兜底资本来源/实际政策参数/目标环境运行证据、组合模型独立验证和
  业务/法务参数。
- 生产公开行情/CDN外部探针、复杂订单批准容量及真实用户通知。

以上任一项不得从本报告推断为通过。完整剩余项见
`docs/option-current-status-and-production-blockers.md`。

## 8. 完整性

同目录 `.sha256` 文件固定本报告及相关实现、迁移、门禁和运营文档。执行：

```bash
cd <repository-root>
shasum -a 256 -c services/option/docs/evidence/option-repository-technical-evidence-20260802.sha256
```

任一哈希变化后，本报告状态自动失效，必须审查变化、重跑相应门禁并生成新日期/版本证据；禁止只更新哈希。
