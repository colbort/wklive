# Option P2-001 交易日历仓库验收记录

## 1. 结论

P2-001 的仓库实现和真实 Asset RPC 专项已通过，设计符合标准化期权的基本交易时段治理要求：

- 合约绑定不可变日历代码，日历以更高版本和精确生效时刻演进；
- IANA 时区、左闭右开周会话、跨午夜、`CLOSED/OPEN` UTC 例外和闭市优先级有确定语义；
- 创建与复核分离，已批准版本、例外和 halt 身份受数据库不可变约束；
- 下单在初始校验和合约行锁内两次准入，缺少/歧义/闭市/halt 均 fail closed；
- 紧急 halt 先拒绝新单，再幂等撤销活动单；资金释放完成前禁止恢复；
- 日历和 halt 只管交易准入，不停止风险、强平、行权、到期和结算义务。

本结论是“仓库验收通过”，不是“生产上线批准”。目标市场、官方年度日历、预生产多实例/网络并发、故障/通知演练和角色签署未完成，状态仍为 `VERIFYING`。

## 2. 验收范围

| 范围 | 验收内容 | 结果 |
| --- | --- | --- |
| 时间算法 | 纽约 DST、开/闭边界、多会话、跨午夜、24×7、例外优先级、非法时区/重叠定义 | PASS |
| 迁移治理 | `20260731_zk_option_trading_calendar.sql` 在正式隔离验收库连续安装两次 | PASS |
| 未来版本 | 通过真实管理创建/复核逻辑将 V2 批准为未来闭市版本 | PASS |
| 边界准入 | V1 在切换前一秒开市；V2 在切换时刻精确接管并拒绝新单 | PASS |
| 旧单资金 | 切换前 GTC 单真实冻结，切换后撤销并完整释放 | PASS |
| 人工 halt | 同时存在一笔 `PENDING` 和一笔 `FUNDING` 订单时暂停、撤单、拒绝新单 | PASS |
| 恢复屏障 | `CANCELING` 资金释放前恢复失败；释放后恢复并通过新单/撤单回归 | PASS |
| 下单/halt竞态 | 20轮受控竞态窗口交替给公开平仓单或人工halt 150ms先发优势，验证真实MySQL合约行锁的两种合法序列化 | PASS（10受理后撤销/10零副作用拒绝） |
| Asset响应丢失 | halt解冻已提交但响应丢失，恢复沿用原指令和原Asset流水 | PASS |
| 不可变证据 | 改写/删除已批准日历，以及改写/删除 halt 均被 MySQL 触发器拒绝 | PASS |

## 3. 真实 RPC 证据

### 3.1 V1 → V2 未来切换

- V1 为 `SUPERSEDED`，`effective_until` 精确等于 V2 `effective_from`；V2 为 `APPROVED`。
- `switchAt-1` 唯一解析 V1 且开市；`switchAt` 唯一解析 V2，拒绝原因为 `CALENDAR_CLOSED_EXCEPTION`。
- 切换前仅有1笔真实订单；切换时新客户端订单号产生0订单和0客户端键。
- 旧单冻结/释放共2条指令，2条均 `SUCCESS/MATCHED`，2条唯一 Asset 流水。
- 用户钱包最终为总额100、可用100、冻结0。

### 3.2 人工 halt 与释放屏障

- halt 记录的 `cancel_total/cancel_success/cancel_failed = 2/2/0`。
- 暂停后新单拒绝；冻结成功的订单处于 `CANCELING` 时，恢复接口 fail closed。
- 注入一次“Asset 已提交解冻、Option 丢失响应”：原释放指令进入 `FAILED`、订单保持
  `CANCELING`，Asset 仅有1条释放流水；此时再次恢复仍 fail closed。
- 管理员以 `ASSET_INSTRUCTION_MANUAL_RETRY` 审计事件沿用原指令恢复；最终指令
  `SUCCESS/MATCHED`，释放流水仍严格为1，未发生二次解冻。
- 资金任务完成后两笔原订单均为 `CANCELED`，钱包为总额200、可用200、冻结0。
- 恢复后一笔新单受理、冻结并可公开撤销，最终钱包仍为200/200/0。
- 共3笔订单且全部取消；5条指令中4条成功且对账、1条冻结前合法取消，4条唯一 Asset 流水。
- `CONTRACT_TRADING_HALTED` 和 `CONTRACT_TRADING_RESUMED` 事件各1条，终态无活动 halt。

### 3.3 公开下单与 halt 并发竞态

- 共20个独立合约；每轮由公开 `PlaceOrder` 只减仓平仓单与管理 `HaltContractTrading`
  从同一并发屏障起跑并竞争同一合约行锁，交替给其中一方150ms先发优势，确保两种合法序列化均被覆盖。
- 只接受两种终态：订单先提交时，halt必须同步撤销并记录1/1/0；halt先提交时，
  下单必须零订单、零客户端键、零资产指令拒绝。
- 两种终态均要求原多头仓位的可用/冻结为1/0，无
  `FUNDING/PENDING/PART_FILLED/CANCELING/EXPIRING` 订单，且使用原halt可正常恢复。
- 正式结果为10轮订单先提交并全部被撤销、10轮halt先提交且新单全部零副作用拒绝；
  合计20个halt均恢复，10个客户端键仅对应已撤销订单，资产指令为0。
- 该场景故意使用零资产指令的平仓冻结，因此不改变全套资金指令门禁。

新增两个场景合计7条指令：6条成功对账，1条冻结前合法取消。最新完整真实 RPC 门禁为：

```text
instructions=9239 success=9234 canceled=5 reconciled=9234
success + 2*canceled = 9244
```

## 4. 验收发现与修复

| 发现 | 风险 | 修复 |
| --- | --- | --- |
| 验收脚本原未安装 `zk` 日历治理迁移 | 表存在但不可变触发器未被验证，形成假阳性 | 正式脚本现连续安装两次该迁移 |
| 部分旧测试数据直接插入 `APPROVED` 日历或无日历 `TRADING` 合约 | 绕过四眼与准入治理，迁移真实安装后暴露 | 公用fixture改为 `DRAFT`→会话→独立复核的合法序列 |
| 恢复逻辑原只查订单簿活动状态 | 订单已进入 `CANCELING`、Asset仍未释放时可提前恢复 | 新增 `HasUnsafeContractResumeOrders`，显式拦截 `FUNDING/PENDING/PART_FILLED/CANCELING/EXPIRING` |

## 5. 工程验证

| 检查 | 结果 |
| --- | --- |
| `go test ./models ./internal/logic/admin ./internal/logic/task` | PASS |
| `services/option/acceptance/run-p0-asset-rpc-e2e.sh` | PASS，主真实RPC场景47.685秒 |
| `zk` 迁移连续安装两次 | PASS |
| 获批日历/halt 改写与删除 | 全部被数据库拒绝 |

关键自动化入口：

- `internal/logic/task/p0_trading_calendar_rpc_integration_test.go`
- `internal/logic/task/p0_asset_rpc_integration_test.go`
- `acceptance/run-p0-asset-rpc-e2e.sh`

## 6. 生产前剩余阻断项

| 编号 | 阻断项 | 通过标准 | 资料责任 |
| --- | --- | --- | --- |
| CAL-PRE-001 | 多实例/网络下的并发新单/halt竞态 | 仓库20轮行锁基线已通过；预生产跨实例仍必须证明halt返回前的新单要么已序列化，要么零副作用拒绝，且无漏撤单 | 技术/风控 |
| CAL-PRE-002 | Asset 调用前失败、提交后丢响应 | 仓库已通过halt解冻提交后丢响应、原指令人工恢复和唯一流水；预生产仍须覆盖真实网络及多实例，恢复前一律拒绝 | 技术/清算 |
| CAL-PRE-003 | 任务/容器在撤单与释放中断 | 重启后幂等续跑，原halt计数可追溯，达到批准RTO | 技术/运维 |
| CAL-PRE-004 | 权限与四眼 | 创建、复核、halt、恢复的生产角色最小权限；跨租户和同人复核拒绝 | 安全/运营 |
| CAL-PRE-005 | 告警、通知与熔断恢复 | 残单/释放失败立即告警，用户/做市商/值班通知到达，三个健康窗口后恢复 | 运维/运营/风控 |
| CAL-PRE-006 | 目标市场年度日历 | 官方来源、文件哈希、当地/UTC对照、DST/提前收市/补班清单和业务签字完整 | 业务/市场运营/法务 |
| CAL-PRE-007 | 非交易任务续跑 | 闭市和halt期间风险、强平、行权、到期和结算在真实调度下均不停止 | 技术/风控/清算 |

## 7. 已准备的运营资料

- `docs/templates/trading-calendar-approval.md`：日历版本、官方来源、证据哈希、切换前后与四眼审批。
- `docs/templates/trading-calendar-annual-review.md`：年度例外、T-30/T-7/T-1/T+1、并发/故障/通知复演和缺失资料升级。
- `docs/templates/trading-halt-record.md`：暂停、逐单资金释放、不安全订单水位、恢复屏障和通知证据。
- `docs/option-operations-runbook.md` 6.26：日历切换、临时休市和恢复的操作流程与禁止操作。

无法仓库内预填的资料是目标交易场所、官方年度日历URL/原始文件、适用司法辖区、用户公告时限、批准峰值/RTO/SLA和实名责任人。这些必须由业务、法务、风控、运营和运维提供，不应由开发猜测。
