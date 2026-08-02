# Option 保险基金账本生产语义审批与历史盘点记录

OPTION_INSURANCE_LEDGER_APPROVAL_STATUS: DRAFT

> 本记录是开放 Option 卖方交易和启用保险基金生产口径的前置证据。仓库验收只证明新代码契约，
> 不代表财务/清算已经批准历史解释。全部门禁、差异处置和签署完成前，状态必须保持 `DRAFT`；
> 签署归档副本才可改为 `APPROVED`，并须固定文件及附件 SHA-256，之后不得原地改写。

## 1. 审批身份与范围

| 字段 | 填写值 |
| --- | --- |
| 工单号/记录版本 | 待填写 |
| tenant_id / 法律实体 | 待填写 |
| 币种与数据时间范围 | 待填写 |
| 一致性截止时间（UTC） | 待填写 |
| Option 数据库备份/快照 ID | 待填写 |
| Asset 数据库快照 ID | 待填写 |
| 发布 commit / 镜像 digest | 待填写 |
| `20260802_option_insurance_fund_flow_semantics.sql` SHA-256 | 待填写 |
| 查询包/导出文件及 SHA-256 | 待填写 |
| 执行人 / 独立复核人 | 待填写 |

Option 与 Asset 必须使用同一稳定截止点。若快照范围、时区、租户或币种不一致，本记录只能标记
`INCOMPLETE`，不得把差额为零当作通过。

## 2. 待批准的数据契约（仓库已实现）

| `flow_type` | 业务含义 | 新行原始 `amount` | 归一方向 |
| ---: | --- | ---: | ---: |
| 1 | 强平费 | `ABS(amount) > 0` | `+ABS(amount)` |
| 2 | 缺口赔付 | `ABS(amount) > 0` | `-ABS(amount)` |
| 3 | 人工注资 | `ABS(amount) > 0` | `+ABS(amount)` |
| 4 | 人工提取 | `ABS(amount) > 0` | `-ABS(amount)` |

统一读取表达式：

```sql
CASE WHEN flow_type IN (2,4) THEN -ABS(amount) ELSE ABS(amount) END
```

必须逐项确认：

- [ ] `amount` 是业务绝对金额，不再以正负号承载方向；新行必须大于0。
- [ ] 方向只由 `flow_type` 决定；类型1/3为流入，2/4为流出。
- [ ] 归一净变动不是 Asset 平台账户余额，不能替代期初、期末或权威 Asset 流水。
- [ ] 新流水必须有非空 `asset_flow_no`；类型1/2还必须关联合约和强平记录。
- [ ] 保险流水不可 `UPDATE` 或 `DELETE`；更正只能追加可审计的新业务事实。
- [ ] 历史正数、负数和异常行不批量改写；读取时使用 `ABS` 兼容，异常逐案处理。
- [ ] 当前运行时类型1/2的 `asset_flow_no` 是 Asset 业务身份，按
  `t_asset_platform_flow.biz_no` 及对应保险/强平事实核对；类型3/4启用前须由财务/清算批准
  Asset 场景、业务号和账户映射，不能仅凭 Option 行人工入账。

审批结论：接受/拒绝/有条件接受：待填写。若有条件接受，条件、到期时间和关闭证据：待填写。

## 3. 历史基线盘点

先保存完整逐行导出，再按 `tenant_id + coin + flow_type` 运行以下只读统计；查询结果和导出文件均须
记录 SHA-256。不得在盘点阶段执行修复 SQL。

```sql
SELECT tenant_id, coin, flow_type,
       COUNT(*) AS row_count,
       SUM(amount > 0) AS positive_rows,
       SUM(amount < 0) AS negative_rows,
       SUM(amount = 0) AS zero_rows,
       SUM(asset_flow_no = '') AS missing_asset_flow_no_rows,
       SUM(amount) AS raw_algebraic_sum,
       SUM(ABS(amount)) AS raw_magnitude_sum,
       SUM(CASE WHEN flow_type IN (2,4)
                THEN -ABS(amount) ELSE ABS(amount) END) AS normalized_net
FROM t_option_insurance_fund_flow
WHERE tenant_id = ? AND create_times < ?
GROUP BY tenant_id, coin, flow_type
ORDER BY tenant_id, coin, flow_type;
```

重复业务身份必须单独导出：

```sql
SELECT tenant_id, asset_flow_no, COUNT(*) AS row_count,
       GROUP_CONCAT(id ORDER BY id) AS option_flow_ids
FROM t_option_insurance_fund_flow
WHERE tenant_id = ? AND create_times < ? AND asset_flow_no <> ''
GROUP BY tenant_id, asset_flow_no
HAVING COUNT(*) <> 1;
```

| coin/type | 总行 | 正数 | 负数 | 零数 | 缺身份 | 重复身份 | 原始代数和 | 归一净变动 | 证据文件 |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | --- |
| 待填写 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 待填写 |

验收要求：新语义生效后的新增行 `zero_rows=0`、`negative_rows=0`、缺失身份=0；历史异常允许被识别，
但必须逐行进入第6节差异案件，不得通过删除或改符号使统计“变绿”。

## 4. Option 与 Asset 逐笔核对

从 Asset 导出同截止点的 `INSURANCE_FUND` 平台账户、`t_asset_platform_flow`、
`t_asset_insurance_cover` 及相关强平/费用业务事实。现有类型1/2至少按下列键核对：

| Option 字段 | Asset/业务事实 | 验收 |
| --- | --- | --- |
| `tenant_id` | 平台账户及流水 `tenant_id` | 完全一致 |
| `coin` | 平台账户及流水 `coin` | 大小写归一后完全一致 |
| `asset_flow_no` | `t_asset_platform_flow.biz_no` | 一对一；不得缺失或复用 |
| `flow_type=1` | 保险基金强平费入账场景 | Asset `op_type=1`，金额相等 |
| `flow_type=2` | `INSURANCE_FUND_COVER` / 保险赔付 | Asset `op_type=2`，金额相等，强平身份一致 |
| `amount` | Asset 流水 `amount` | 绝对金额和精度一致，不按展示精度舍入 |
| `contract_id/liquidation_id` | Option 强平和 Asset cover 事实 | 业务链可追溯且唯一 |

| coin/type | Option行 | Asset行 | 一对一 | 未命中 | 重复 | 金额错 | 币种/方向错 | 证据文件 |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | --- |
| 待填写 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 待填写 |

通过标准：所有适用新行一对一；未命中、重复、金额错和币种/方向错均为0。历史异常必须有负责人、
根因、会计处置和关闭证据，不允许用总额相抵代替逐笔解释。

## 5. 逐币影子复算与余额桥接

对每个币种同时给出四种结果，禁止只给一个总和：

| coin | 旧原始 `SUM(amount)` | 类型归一净变动 | Asset权威净变动 | Asset期初 | Asset期末 | `期末-期初-净变动` | 解释/证据 |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | --- |
| 待填写 | 0 | 0 | 0 | 0 | 0 | 0 | 待填写 |

通过标准：Asset 期初 + 截止点内权威净变动 = Asset 期末；类型归一值与其覆盖的 Asset 保险业务净变动
相等。旧原始代数和只用于展示旧报表影响，不作为正确性标准。任何差额非0均阻断卖方生产放行。

## 6. 异常与差异案件

| case_no | 行ID/业务号 | 类型 | 金额/币种 | 异常 | 用户/财务影响 | 临时控制 | 根因 | 处置 | 负责人/期限 | 状态 |
| --- | --- | ---: | --- | --- | --- | --- | --- | --- | --- | --- |
| 待填写 | 待填写 | 0 | 待填写 | 待填写 | 待填写 | 待填写 | 待填写 | 待填写 | 待填写 | OPEN |

允许的处置只有：保留历史并修正读取/报表、追加经批准的 Asset/业务更正事实、或明确记录不影响余额的
数据质量例外。禁止修改/删除原保险流水、复用业务号、伪造 Asset 流水或批量把历史金额取反。

## 7. 下游消费者盘点

| 消费者/仓库 | 查询或字段 | 当前口径 | 修改/确认 | 负责人 | 测试与发布证据 | 结论 |
| --- | --- | --- | --- | --- | --- | --- |
| Option运营工作台 | 保险净变动 | 类型+ABS | 已由仓库实现 | Option | 待填写 | 待填写 |
| Prometheus指标/告警 | 保险净变动 | 类型+ABS | 已由仓库实现 | Option/SRE | 待填写 | 待填写 |
| 日终对账 | 平台账户及保险事实 | Asset权威余额+逐笔桥接 | 待预生产执行 | 财务/清算 | 待填写 | 待填写 |
| 数据仓库/ETL | 待填写 | 待填写 | 待填写 | 待填写 | 待填写 | 待填写 |
| 财务报表/导出 | 待填写 | 待填写 | 待填写 | 待填写 | 待填写 | 待填写 |
| 告警/离线作业/临时SQL | 待填写 | 待填写 | 待填写 | 待填写 | 待填写 | 待填写 |

必须对全部代码仓、BI、ETL、导出、手工报表和告警搜索 `t_option_insurance_fund_flow`、`amount`、
`SUM(amount)`；每个命中都有所有者和结论。不能以“主仓库已修复”推断外部消费者安全。

## 8. 预生产迁移、负向与回滚演练

- [ ] 从生产同结构备份恢复到隔离预生产，记录恢复校验和。
- [ ] 迁移连续执行两次成功；表、触发器和权限与发布 commit 一致。
- [ ] 四种类型正金额按已批准场景写入，方向归一符合第2节；未批准类型保持无写入口。
- [ ] 零数、负数、未知类型、缺少身份、类型1/2缺强平身份均被拒绝。
- [ ] 已存在流水的 `UPDATE`、`DELETE` 被数据库拒绝并进入数据库安全审计。
- [ ] 真实 Asset RPC 赔付原始金额为正、归一为负、业务身份唯一；响应丢失重放不重复扣款。
- [ ] 同截止点逐笔核对、逐币影子复算和日终守恒全部通过。
- [ ] OPT-A015及相关保险库存告警完成触发、通知、案件、恢复和失败升级。
- [ ] 回滚应用读取版本后，新行仍保持正绝对金额；不得删除触发器、恢复带符号写入或改写历史。

## 9. 最终门禁与签署

只有以下条件全部满足才可把状态改为 `APPROVED`：

- [ ] 第2节财务语义无保留批准；类型3/4未开放时有显式“不适用/关闭”结论。
- [ ] 第3～5节完整，所有新行异常计数和逐笔差异为0。
- [ ] 第6节无未关闭的资金正确性或财务影响案件。
- [ ] 第7节所有消费者有所有者、版本、测试和发布证据。
- [ ] 第8节预生产与通知演练通过，附件哈希固定。
- [ ] `option-production-readiness.sh` 生产模式为 `READY`，生产签署单为 `APPROVED`。

| 角色 | 姓名 | 结论 | 时间（UTC） | 审批/证据 |
| --- | --- | --- | --- | --- |
| 财务 | 待填写 | 通过/拒绝 | 待填写 | 待填写 |
| 清算 | 待填写 | 通过/拒绝 | 待填写 | 待填写 |
| 风控 | 待填写 | 通过/拒绝 | 待填写 | 待填写 |
| Option技术 | 待填写 | 通过/拒绝 | 待填写 | 待填写 |
| Asset技术 | 待填写 | 通过/拒绝 | 待填写 | 待填写 |
| 合规/审计 | 待填写 | 通过/拒绝 | 待填写 | 待填写 |

任一角色拒绝、任何证据哈希不一致或任何适用项未完成时，卖方入口和保险相关生产开关必须保持关闭。
