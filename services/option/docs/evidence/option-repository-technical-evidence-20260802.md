# Option 仓库技术证据（2026-08-02）

EVIDENCE_SCOPE: REPOSITORY_ONLY

EVIDENCE_STATUS: PASSED_NOT_RELEASE_CANDIDATE

## 1. 使用限制

本报告只证明当前共享工作区的仓库实现和隔离真实 Asset RPC 门禁。它不是预生产或生产证据，不能填写
真实 tenant、release image、Prometheus/Alertmanager 通知、容器编排、财务日终、业务参数或人员签字。

当前 Git HEAD 为 `adbab9b72885a4fcd692124a15421fab8e423b03`，但执行时工作区有166条未提交状态，
因此不存在可部署的不可变 release commit/image。必须先由负责人审查并形成干净发布提交，再在该提交上
重跑门禁；本报告不能作为发布身份。

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
| `make gen-model` | PASS | DDL 变更后执行；保险流水生成字段注释与 schema 一致 |
| `GOCACHE=/private/tmp/wklive-go-build-cache go test ./...`（Option） | PASS | 全 Option 包 |
| `GOCACHE=/private/tmp/wklive-go-build-cache go vet ./...`（Option） | PASS | 全 Option 包 |
| `GOCACHE=/private/tmp/wklive-go-build-cache go test -race ./models ./internal/logic/task` | PASS | 本轮改动相关模型与任务竞态检查 |
| `go test ./... && go vet ./...`（Admin API） | PASS | 管理 API 回归 |
| `npx prettier --check ... && npm run type-check`（Admin UI） | PASS | 中英文运营标签与类型检查 |
| `./monitoring/option-production-readiness.sh --repository-only` | PASS | 仓库项通过；本机缺少`promtool`按repository mode跳过 |
| `services/option/acceptance/run-p0-asset-rpc-e2e.sh` | PASS | Docker隔离库、Redis、真实Asset gRPC、迁移双执行和故障注入 |
| `git diff --check` | PASS | 当前差异无空白错误 |

## 4. 正式隔离链路摘要

- 主 Option/Asset 集成测试：`113.879s`。
- 资金指令：9277条；9270条成功且已对账；7条冻结前合法取消；加权终态9284。
- 501空头指派：503条资金指令全部成功对账。
- 5000空头指派：5002条资金指令全部成功对账；Asset RPC阶段`2m54.988s`。
- 501现金到期多头/502持仓：1004条资金指令全部成功对账。
- 501实物交割单元：2004条资金指令全部成功对账。
- 独立行权/实物工作进程在Asset提交后被`SIGKILL`，30秒租约自然到期后唯一接管并保持流水唯一。
- 最后交易边界：合约PAUSED、订单CANCELED、原因`CONTRACT_LAST_TRADE_ENDED`，提前行权/结算均为0。

## 5. 保险基金流水专项

- 新行原始`amount`只接受正业务绝对金额；1/3类型按`+ABS`、2/4类型按`-ABS`读取。
- 真实缺口赔付原始金额15、方向归一金额-15。
- 汇总输出：`raw_inflow=28.49 raw_outflow=54.10 signed_net=-25.61 nonpositive_rows=0`。
- 负数新增、既有流水UPDATE和DELETE均由MySQL拒绝；迁移连续执行两次成功。
- 历史流水未改写；生产仍须逐笔关联`asset_flow_no`并由财务/清算签署。

## 6. 明确未证明

- 干净、不可变的 release commit 和 Option/Asset 镜像 digest。
- 目标预生产存量备份升级、生产配置、权限和数据库独立安全审计。
- AMD64 Beanstalkd容量、24小时长稳、真实Pod编排墙钟强杀/RTO。
- 生产Prometheus抓取、`promtool`、Alertmanager电话/IM/案件及恢复通知。
- 真实财务日终、保险历史盘点、兜底额度、组合模型独立验证和业务/法务参数。
- 生产公开行情/CDN外部探针、复杂订单批准容量及真实用户通知。

以上任一项不得从本报告推断为通过。完整剩余项见
`docs/option-current-status-and-production-blockers.md`。

## 7. 完整性

同目录 `.sha256` 文件固定本报告及相关实现、迁移、门禁和运营文档。执行：

```bash
cd <repository-root>
shasum -a 256 -c services/option/docs/evidence/option-repository-technical-evidence-20260802.sha256
```

任一哈希变化后，本报告状态自动失效，必须审查变化、重跑相应门禁并生成新日期/版本证据；禁止只更新哈希。
