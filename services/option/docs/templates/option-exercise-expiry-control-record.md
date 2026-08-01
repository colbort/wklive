# Option 行权截止与到期清算控制记录

状态：`DRAFT / READY / EXECUTING / COMPLETED / EXCEPTION`

用途：每个现金结算合约在主动行权截止和到期清算时单独填写。仓库自动化结果只证明代码基线，
不能替代本记录的预生产/生产事实、用户通知和签字。

## 1. 合约与责任人

| 项目 | 记录 |
| --- | --- |
| 租户 / 合约 ID / 合约代码 | `[TENANT]` / `[CONTRACT_ID]` / `[CONTRACT_CODE]` |
| Call/Put / 欧式或美式 / 现金结算 | `[TYPE]` / `[STYLE]` / `CASH` |
| 标的 / 结算币 / 乘数 / 行权价 | `[UNDERLYING]` / `[COIN]` / `[MULTIPLIER]` / `[STRIKE]` |
| 产品负责人 / 风控 / 技术值班 | `[NAME]` / `[NAME]` / `[NAME]` |
| 行情负责人 / 清算 / 运营 | `[NAME]` / `[NAME]` / `[NAME]` |
| 事件群 / 电话 / 证据目录 | `[CHANNEL]` / `[PHONE]` / `[IMMUTABLE_PATH]` |

## 2. 已审批参数

| 参数 | 批准值 | 数据库事实 | 复核 |
| --- | --- | --- | --- |
| 上市时间（UTC秒与本地时区） | `[VALUE]` | `[VALUE]` | `[ ]` |
| 最后交易时间 | `[VALUE]` | `[VALUE]` | `[ ]` |
| 主动行权/指令截止时间 | `[VALUE]` | `[VALUE]` | `[ ]` |
| 到期时间 / 交割时间 | `[VALUE]` / `[VALUE]` | `[VALUE]` / `[VALUE]` | `[ ]` |
| AUTO 阈值 / 行权费率 | `[VALUE]` / `[VALUE]` | `[VALUE]` / `[VALUE]` | `[ ]` |
| 结算来源 / 窗口 / 算法 / 最少样本 | `[VALUE]` | `[VALUE]` | `[ ]` |
| 手续费归集用户/账户 | `[USER]` / `[ACCOUNT]` | `[USER]` / `[ACCOUNT]` | `[ ]` |
| 用户公告 / 风险披露 / 触达批次 | `[LINK/ID]` | `[RECEIPT]` | `[ ]` |

所有时间必须同时记录 UTC 秒值、带时区文本和来源。不得写“收盘前”“当天”等相对描述。

## 3. 系统不变量（预填，禁止业务覆盖）

- 新主动行权和新到期指令只允许已上市且截止前的 `TRADING/PAUSED` 合约。
- 暂停交易不剥夺行权权；`PENDING/EXPIRED/SETTLED/OFFLINE` 不接受新提交。
- 同一客户端幂等键和相同经济请求返回原记录；同键不同请求拒绝。
- 指令只能追加版本并允许 `ACTIVE → SUPERSEDED`；经济字段、状态回退和删除均禁止。
- AUTO 需达到每单位内在价值阈值且扣费后净收益为正；DNE 不行权；相反行权不能绕过正净收益。
- 到期空头只按实际行权多头数量，以 `(position.create_times, position.id)` FIFO 承担。
- 每个资金执行域先完成低步骤再执行高步骤；任何前置扣款未成功时禁止多头或手续费入账。
- `空头 gross 扣款 = 多头净入账 + 行权费`，逐币用户资产加平台资产必须守恒。
- Option 状态不是资金成功证据；必须同时核对 Asset 指令、唯一流水、余额和对账状态。

## 4. T-24h 至截止前检查

| 时间 | 检查 | 结果/数量 | 证据 |
| --- | --- | --- | --- |
| T-24h | 持仓、活动订单、指令版本、公告触达 | `[PASS/FAIL]` | `[LINK/HASH]` |
| T-2h | Asset RPC、数据库、Redis、任务、行情源和审批人 | `[PASS/FAIL]` | `[LINK/HASH]` |
| T-30m | 主动行权积压、异常账户、AUTO/DNE/相反指令冲突 | `[PASS/FAIL]` | `[LINK/HASH]` |
| T-5m | 服务时钟、独立 NTP 偏差、配置冻结 | `[OFFSET]` | `[LINK/HASH]` |
| 截止前边界 | TRADING/PAUSED 成功，其余状态拒绝 | `[RESULT]` | `[LINK/HASH]` |

若时钟偏差超出批准阈值、行情证据不足、资金/行权积压非零或审批人缺席，状态改为 `EXCEPTION`，
停止新增风险并升级；不得人工延长截止时间或修改历史指令。

## 5. 截止快照

| 指标 | 数量/值 |
| --- | --- |
| 持仓总数 / 可行权多头数量 | `[N]` / `[QTY]` |
| ACTIVE AUTO / DNE / 相反指令 | `[N]` / `[N]` / `[N]` |
| SUPERSEDED 历史版本 | `[N]` |
| 截止后新提交拒绝数 / 幂等重放数 | `[N]` / `[N]` |
| 待处理主动行权 / 异常指派 | `[N]` / `[N]` |
| 服务 UTC / NTP UTC / 偏差 | `[TIME]` / `[TIME]` / `[OFFSET]` |

快照 SQL、导出文件、日志和哈希：`[EVIDENCE]`

## 6. 美式提前行权抽样（不适用填 N/A）

| 项目 | 事实 |
| --- | --- |
| exercise ID/no / client key / 数量 | `[VALUE]` |
| 同键并发次数 / 唯一行权数 / 唯一冻结数 | `[N]` / `[N]` / `[N]` |
| FIFO 空头顺序与分配数量 | `[LIST]` |
| 指派前后持仓数量 / 维持保证金 | `[BEFORE]` → `[AFTER]` |
| margin lot 剩余数量/金额/冻结业务号 | `[VALUE]` |
| 资金指令数 / 成功 / 对账 / 唯一流水 | `[N]` / `[N]` / `[N]` / `[N]` |
| 空头扣款 / 多头净入 / 手续费 / 差额 | `[D]` / `[C]` / `[F]` / `[D-C-F]` |
| 重放前后行权/指派/指令/流水数量 | `[BEFORE]` / `[AFTER]` |

## 7. 到期 AUTO/DNE 与 FIFO 清算

| 项目 | 事实 |
| --- | --- |
| 最终结算价 ID/版本/来源样本/确认人 | `[VALUE]` |
| 价内多头总量 / AUTO 实际行权量 | `[QTY]` / `[QTY]` |
| DNE 数量 / 阈值放弃数量 / 相反行权数量 | `[QTY]` / `[QTY]` / `[QTY]` |
| FIFO 空头承担量 / 未承担剩余量 | `[QTY]` / `[QTY]` |
| settlement / batch 状态 | `[STATUS]` / `[STATUS]` |
| 资金指令数 / 成功 / 对账 / 唯一流水 | `[N]` / `[N]` / `[N]` / `[N]` |
| batch total debit / credit / 差额 | `[D]` / `[C]` / `[D-C]` |
| DNE 用户余额变化 | `[0 EXPECTED / ACTUAL]` |
| 重放前后结算/明细/指令/流水数量 | `[BEFORE]` / `[AFTER]` |

## 8. 异常与恢复

| 时间 | 业务号/指令号 | 原状态/重试次数/错误 | 处置 | 操作人/审批 | 恢复证据 |
| --- | --- | --- | --- | --- | --- |
| `[TIME]` | `[NO]` | `[STATE/COUNT/ERROR]` | `[ACTION]` | `[IDS]` | `[LINK/HASH]` |

恢复规则：只能沿用原指令号和原幂等键；余额不足补资使用独立业务号和
`cash-settlement-topup-recovery-approval.md`；禁止改表、换号重复扣款或在前置失败时人工入账。

## 9. 完成判定与签署

- [ ] 所有行权、指派、结算、资金指令和对账均进入允许终态。
- [ ] AUTO/DNE/相反指令结果与最新 ACTIVE 版本一致。
- [ ] FIFO 数量完整，无多付、少付、重复指派或截断。
- [ ] 逐币 gross 借贷、手续费和总资产守恒，冻结余额与剩余 margin lot 一致。
- [ ] 重放未新增记录或流水；异常均有事件、责任人和关闭时间。
- [ ] 用户完成通知已发出并保存送达证据。
- [ ] 日终对账和未关闭问题已归档。

| 角色 | 姓名/ID | 决定 | 时间 | 签名/证据 |
| --- | --- | --- | --- | --- |
| 技术 | `[VALUE]` | `[APPROVE/REJECT]` | `[TIME]` | `[HASH]` |
| 行情 | `[VALUE]` | `[APPROVE/REJECT]` | `[TIME]` | `[HASH]` |
| 风控 | `[VALUE]` | `[APPROVE/REJECT]` | `[TIME]` | `[HASH]` |
| 清算/财务 | `[VALUE]` | `[APPROVE/REJECT]` | `[TIME]` | `[HASH]` |
| 运营 | `[VALUE]` | `[APPROVE/REJECT]` | `[TIME]` | `[HASH]` |
| 合规/最终批准人 | `[VALUE]` | `[APPROVE/REJECT]` | `[TIME]` | `[HASH]` |

任一必填项缺失、差额非零或签署为 REJECT 时，本合约不得被认定为完成生产行权/到期验收。
