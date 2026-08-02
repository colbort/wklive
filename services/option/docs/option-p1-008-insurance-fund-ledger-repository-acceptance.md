# Option P1-008 保险基金流水语义仓库验收

## 1. 状态

- 结论：`REPOSITORY_PASSED / PREPROD_BLOCKED`
- 仓库基线：新流水统一保存正的业务绝对金额；经济方向只由 `flow_type` 决定。
- 生产限制：本次不批量改写历史流水。卖方交易或生产迁移前仍须完成财务/清算批准、生产逐笔盘点和 Asset 对账签署。

## 2. 已发现的问题

旧表注释约定“入金正、出金负”，但 `DEFICIT_COVER` 实际保存正数；运营工作台和 Prometheus 又直接
`SUM(amount)`。因此缺口赔付会被错误显示为流入。这个字段既不是 Asset 余额，也不能在旧口径下解释为
保险基金净变化。

## 3. 统一数据契约

| `flow_type` | 业务 | 新行原始 `amount` | 方向归一值 |
| ---: | --- | ---: | ---: |
| 1 | 强平费 | `ABS(amount)` | `+ABS(amount)` |
| 2 | 缺口赔付 | `ABS(amount)` | `-ABS(amount)` |
| 3 | 人工注资 | `ABS(amount)` | `+ABS(amount)` |
| 4 | 人工提取 | `ABS(amount)` | `-ABS(amount)` |

所有读取方使用：

```sql
CASE WHEN flow_type IN (2,4) THEN -ABS(amount) ELSE ABS(amount) END
```

使用 `ABS` 是有意的兼容策略：历史行即使按旧约定保存了负的支出，读取也能得到一致方向；历史经济
证据不需要先改写。方向归一净变动仍不是账户余额，实际期初、期末及权威流水以 Asset 为准。

## 4. 已实现的防线

- `option.sql` 和 `make gen-model` 产物同步为“业务绝对金额，方向由 `flow_type` 确定”。
- Go 模型写入前拒绝零数、负数、未知类型和缺失的 `asset_flow_no`；强平费/缺口赔付还必须关联合约和强平记录。
- 数据库迁移对新增行执行同样校验，并禁止流水 `UPDATE`、`DELETE`。
- 为兼容历史升级，原 CHECK 仍允许历史非零负数；新行的正数要求由 INSERT 触发器强制。这不是放宽新写入。
- 运营工作台和 Prometheus 金额指标复用同一方向表达式，避免两个读模型再次漂移。
- 正式 Asset RPC 门禁重复执行迁移，并验证真实缺口赔付原始金额 `15`、归一金额 `-15`、负数新增及改删均被拒绝。

## 5. 仓库验收标准

- 四种类型的方向单元测试全部通过，零数、负数、未知类型均拒绝。
- Go、数据库和运营聚合使用同一契约；任何直接 `SUM(amount)` 的保险净变动展示均不得通过 readiness。
- 迁移连续执行两次成功；已有行不改写。
- 真实 Asset 缺口覆盖只生成一条带 `asset_flow_no` 的 Option 流水，原始/归一金额分别为 `15/-15`。
- 流水 UPDATE、DELETE 和负数 INSERT 均由 MySQL 拒绝。
- Option、Admin API、Admin UI 回归及仓库 readiness 通过。

## 6. 生产前剩余验收

以下项目不由代码仓库验收替代，完成前保持 `PREPROD_BLOCKED`：

1. 财务和清算书面批准“原始正绝对金额 + 类型决定方向”的正式账务解释。
2. 生产备份并导出全部历史流水，按租户、币种、类型统计正数、负数、零数和缺失 `asset_flow_no`。
3. 逐笔用 `asset_flow_no` 对应 Asset 平台流水，确认金额、币种、方向、业务号和唯一性；异常不得自动修复。
4. 按币种并行输出旧原始代数和、方向归一净变动、Asset 权威净变动和期初/期末余额，差额必须解释并签字。
5. 盘点所有下游 SQL、导出、报表、告警和离线作业，禁止继续把 `SUM(amount)` 当作净变化。
6. 预生产以真实 Asset RPC、日终截止点和告警通知完成灰度复演；保存配置、日志、查询结果和审批哈希。
7. 回滚只允许回滚读取/展示版本；不得删除触发器后恢复产生带符号的新行，也不得批量反向改写历史证据。

## 7. 验收证据

- 迁移：`migrations/20260802_option_insurance_fund_flow_semantics.sql`
- 模型与测试：`models/toptioninsurancefundflowmodel.go`、`models/toptioninsurancefundflowmodel_test.go`
- 真实链路：`internal/logic/task/p0_liquidation_rpc_integration_test.go`
- 正式门禁：`acceptance/run-p0-asset-rpc-e2e.sh`、`monitoring/option-production-readiness.sh`
- 运营口径：`docs/templates/option-daily-fund-reconciliation.md`、`docs/option-daily-conservation-contract.md`
- 生产审批与历史盘点：`docs/templates/option-insurance-fund-ledger-production-approval.md`

2026-08-02 仓库正式验收结果：

- `make gen-model` 已执行，DDL 与生成模型字段注释一致。
- Option `go test ./...`、`go vet ./...`、`go test -race ./models ./internal/logic/task` 通过。
- 仓库 production readiness 中本项相关契约、读写、数据库、真实链路和文档检查全部通过；仅 repository mode 预期跳过本机缺失的 `promtool`。
- 完整 Docker + 真实 Asset RPC 门禁通过；主集成测试 `113.879s`。
- 资金指令共9277条：9270条成功且已对账，7条冻结前合法取消，加权终态9284。
- 真实缺口赔付原始金额15、方向归一金额-15；正式汇总为
  `raw_inflow=28.49 raw_outflow=54.10 signed_net=-25.61 nonpositive_rows=0`。
- 501/5000空头、501现金到期持仓、501实物交割单元等既有容量边界继续通过。

以上证明仓库实现，不替代生产财务/清算审批及逐笔历史核对。
