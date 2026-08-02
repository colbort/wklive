# Option 美式提前行权目标环境验收与审批记录

OPTION_AMERICAN_EXERCISE_READINESS_STATUS: DRAFT

允许状态：`DRAFT / APPROVED / REJECTED`。本记录只批准现金结算美式期权的提前行权；组合保证金美式、
实物美式及任何未列合约都不在范围内。仓库基线不能替代目标环境、正式规则、通知和签署。

## 1. 身份与合约全集

| 字段 | 必填值 |
| --- | --- |
| tenant_id / 法律实体 / 司法辖区 | `[VALUE]` |
| Option / Asset release或image digest | `[VALUE]` / `[VALUE]` |
| 渲染配置路径及SHA-256 | `[PATH]` / `[SHA256]` |
| 美式合约完整导出及SHA-256 | `[PATH]` / `[SHA256]` |
| 上市合集 / 公告客户端导出 / 三方对账 | `[PATHS_AND_HASHES]` |
| 变更单 / 上线窗口 / 回滚窗口 | `[VALUE]` |

| contract_id/code | Call/Put | strike/multiplier | list/last_trade/cutoff/expire/deliver | AUTO阈值/费率 | 结算价规则 | 单合约控制记录SHA |
| --- | --- | --- | --- | --- | --- | --- |
| `[VALUE]` | `[VALUE]` | `[VALUE]` | `[FIVE_TIMES]` | `[VALUE]` | `[VALUE]` | `[SHA256]` |

每个合约必须链接已完成的`option-exercise-expiry-control-record.md`；合约全集不得抽样。提前行权能力依赖
卖方交易已批准，卖方组合保证金模式保持不支持美式。

## 2. 目标环境验收

| 用例 | 最低场景 | 预期 | 实际/原始证据SHA | 结论 |
| --- | --- | --- | --- | --- |
| AMER-PRE-001 | 价内、价外、零内在价值、费用后净值为0/正一最小单位 | 仅费用后正收益可提交 | `[VALUE]` | 通过/拒绝 |
| AMER-PRE-002 | 截止前1毫秒、截止精确时刻、截止后1毫秒 | 仅截止前接受，边界使用同一UTC毫秒事实 | `[VALUE]` | 通过/拒绝 |
| AMER-PRE-003 | TRADING、PAUSED、PENDING、EXPIRED、SETTLED及公司行动迁移中 | 仅允许状态且无迁移阻断时接受 | `[VALUE]` | 通过/拒绝 |
| AMER-PRE-004 | 20路同client key、同键不同经济请求、不同键替换 | 唯一记录/指令；冲突拒绝；版本链不可变 | `[VALUE]` | 通过/拒绝 |
| AMER-PRE-005 | 1/500/501/5000空头FIFO及部分数量 | `(create_times,id)`稳定完整，无截断或超分配 | `[VALUE]` | 通过/拒绝 |
| AMER-PRE-006 | 指派与空头平仓单并发，双方交替获得行锁 | 只有合法串行化结果，无重复释放或负仓 | `[VALUE]` | 通过/拒绝 |
| AMER-PRE-007 | Asset扣款提交后响应丢失、Option SIGKILL、租约自然到期、双实例接管 | 原业务号唯一恢复，高步骤不越过扣款屏障 | `[VALUE]` | 通过/拒绝 |
| AMER-PRE-008 | Call/Put、不同账户、部分持仓、手续费和margin lot比例减少 | 数量、保证金、费用和逐币总资产守恒 | `[VALUE]` | 通过/拒绝 |
| AMER-PRE-009 | 到期AUTO/DNE/相反指令与已提前行权数量并存 | 只处理剩余数量，最新ACTIVE指令生效，无重复指派 | `[VALUE]` | 通过/拒绝 |
| AMER-PRE-010 | T-24h/T-2h/T-30m/T-5m、截止/完成通知和异常升级 | 用户、清算、值班通知在批准SLA内送达 | `[VALUE]` | 通过/拒绝 |

## 3. 资金与数量守恒汇总

| 指标 | 必须满足 | 实际 | 证据SHA |
| --- | --- | --- | --- |
| 已验收合约数 | 等于美式合约批准全集 | `[N]` | `[SHA256]` |
| 行权量 / FIFO指派量 / 多头减少 / 空头减少 | 四者相等 | `[VALUES]` | `[SHA256]` |
| 空头gross扣款 / 多头净入 / 手续费 | `gross = net + fee` | `[VALUES]` | `[SHA256]` |
| Option指令 / 唯一Asset流水 / 已对账 | 一一对应且全部终态 | `[N] / [N] / [N]` | `[SHA256]` |
| 冻结、margin lot和持仓差异 | 0 | `[VALUE]` | `[SHA256]` |
| 未关闭SEV-1/2、资金或数量差异 | 0 | `[N]` | `[SHA256]` |

## 4. 签署

| 角色 | 姓名/ID | 结论 | UTC时间 | 工单/证据 |
| --- | --- | --- | --- | --- |
| 产品/运营 | `[VALUE]` | 通过/拒绝 | `[TIME]` | `[VALUE]` |
| Option技术 | `[VALUE]` | 通过/拒绝 | `[TIME]` | `[VALUE]` |
| Asset/SRE | `[VALUE]` | 通过/拒绝 | `[TIME]` | `[VALUE]` |
| 风控 | `[VALUE]` | 通过/拒绝 | `[TIME]` | `[VALUE]` |
| 清算/财务 | `[VALUE]` | 通过/拒绝 | `[TIME]` | `[VALUE]` |
| 合规/法务 | `[VALUE]` | 通过/拒绝 | `[TIME]` | `[VALUE]` |

只有AMER-PRE-001～AMER-PRE-010全部通过、逐合约控制记录完整、资金数量守恒、通知送达且六方签署后，
归档副本才允许改为`OPTION_AMERICAN_EXERCISE_READINESS_STATUS: APPROVED`。
