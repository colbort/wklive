# Option Scope 2 日终资金守恒数据契约

## 1. 目标与边界

本契约定义 `t_option_reconciliation_run.scope=2` 的唯一合格生产口径。Scope 2 必须证明
Asset Option 用户钱包、Option 业务子账及相关 Asset 平台账户在同一 UTC 截止点可复算，不能用
`t_option_account` 与实时钱包相等、人工勾选或“没有开放差异案件”代替。

自动任务在 UTC 00 点后的首个调度窗口关闭前一业务日。业务区间采用半开区间
`[00:00:00.000, 次日00:00:00.000)`；所有 Asset 时间字段按毫秒比较，运行表时间按 Unix 秒记录。

## 2. 权威数据与截止点

| 范围 | 权威余额 | 权威流水 | Option 交叉账 |
| --- | --- | --- | --- |
| 用户 Option 钱包 | `t_user_asset(wallet_type=5)` | `t_asset_flow(wallet_type=5)` | `t_option_asset_instruction`、`t_option_bill` |
| 平台账户 | `t_asset_platform_account` | `t_asset_platform_flow` | 费用、保险、兜底业务事实 |

单次运行必须在数据库一致性读视图内取得以下高水位并写入 `snapshot_ref/evidence_ref`：

- 查询开始时的 Unix 毫秒；
- 用户流水最大 `id`；
- 平台流水最大 `id`；
- Option 资金指令最大 `id`；
- Option bill 最大 `id`。

重跑必须沿用同一业务日期，但追加新的 `attempt_no`。每个 attempt 使用自己的高水位并保留差异；
不得覆盖旧运行或旧明细。

## 3. 用户钱包逐币复算

每个 `tenant_id + user_id + coin` 先复算，再按 `tenant_id + coin` 汇总：

```text
expected_close = opening
               + external_net
               + option_net
               + manual_net
difference     = actual_close - expected_close
```

分类只允许：

| 分类 | Asset 条件 | 方向 |
| --- | --- | --- |
| `option_net` | `biz_type='option'` | 使用 `after_total_amount-before_total_amount` |
| `manual_net` | `scene_type IN ('manual_add','manual_sub')` | 使用实际总额差，不信任名称方向 |
| `external_net` | `biz_type IN ('payment','transfer')` | 使用实际总额差 |

冻结、解冻、锁定和解锁只能改变总额内部构成，总额变化必须为 0。其他业务类型、空业务号、空流水号、
未知操作方向均标记 `INCOMPLETE`，不得归入“其他”后输出零差额。

`actual_close` 由一致性读视图中的当前余额减去业务日结束后、快照高水位以内的总额净变化反推。
这一口径允许任务在 UTC 00:05 执行时排除新业务日已发生的流水。当天第一条流水的
`before_total_amount` 是独立期初；当天无流水的钱包期初等于反推日终。

## 4. 流水链完整性门禁

任一条件失败，该维度状态必须为 `INCOMPLETE`，运行不得写成功：

1. 每条流水满足 `before_total = before_available + before_frozen + before_locked` 及对应 after 恒等式。
2. 同一钱包按 `(create_times,id)` 排序，后一条 before 四项等于前一条 after 四项。
3. `op_type` 与总额变化一致：增加/转入为 `+change_amount`；扣减/扣冻结/扣锁定/转出为
   `-change_amount`；冻结/解冻/锁定/解锁总额变化为 0。
4. 当前余额满足总额恒等式，且通过截止后流水反推得到的日终与当天最后一条 after 一致。
5. 当天涉及的 `biz_type='option'` Asset 流水必须且只能关联一条成功 Option 资金指令，
   `asset_flow_no`、用户、币种及动作金额一致。
6. 成功 Option 指令不得缺少 `asset_flow_no`；同一 Asset 流水不得被多个指令引用。
7. Asset Option 流水与 `t_option_bill` 的业务号/实际流水引用必须可追踪；延迟写账只能输出
   `INCOMPLETE`，不能静默忽略。

## 5. 平台账户复算

Scope 2 覆盖 `FEE_REVENUE`、`INSURANCE_FUND`、`OPTION_BACKSTOP`。每个
`account_type + coin` 使用 `t_asset_platform_flow.before_available/after_available` 复算，采用与
用户钱包相同的截止后反推及连续性检查。

`scene_type='platform_manual_adjust'` 计入 `manual_net`，其他流水计入 `option_net`；不允许空
`biz_type/scene_type/biz_no`。`FEE_REVENUE` 若与其他产品共用，运行必须保留完整账户流水，Option
子账只做交叉核对，不得从平台实际余额中推测拆分值。

保险赔付历史正负号未获财务/清算批准前，以 Asset 平台流水的 `op_type` 和前后余额为权威；
`t_option_insurance_fund_flow.amount` 只用于发现符号差异，不能用于改写 Asset 结果。

## 6. 不可变结果

每个 Scope 2 attempt 先写一条 `t_option_reconciliation_run`，再写
`t_option_reconciliation_run_detail`：

| `dimension_type` | `dimension_key` | 含义 |
| ---: | --- | --- |
| 1 | `coin` | 用户钱包逐币汇总 |
| 2 | `account_type:coin` | 平台账户逐账户逐币 |
| 3 | `coin` | Option 指令/bill 与 Asset 流水交叉账 |

数据库 CHECK 强制 `expected_closing` 和 `difference_amount` 公式；触发器禁止明细 UPDATE/DELETE。
只有所有明细 `status=1`、差额为 0、完整性异常为 0 时，运行状态才允许为成功。

差异案件稳定键为 `DAILY:{business_date}:{dimension_type}:{dimension_key}:2`。差异恢复后保留旧案并
转 `RESOLVED`；数据不完整与真实金额差异必须在详情中分开计数。

## 7. 验收标准

- 健康用户、跨日后置流水、无当日流水、18 位小数均精确得到零差额。
- 缺流水、流水链断裂、非法业务分类、余额恒等式错误、Option 指令单边缺失、重复关联分别产生
  `INCOMPLETE`，不得生成成功心跳。
- 用户钱包差一最小单位 `0.000000000000000001` 时精确产生差异。
- 平台账户人工调账与业务流水分栏正确；共享费用账户不丢弃非 Option 流水。
- 同日重跑 attempt 递增；运行及明细 UPDATE/DELETE 均被数据库拒绝。
- Scope 2 最近成功超过36小时或不存在时，生产门禁和 OPT-A015 保持失败关闭。
