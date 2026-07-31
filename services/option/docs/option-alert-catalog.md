# Option 告警目录与首发阈值

## 1. 使用原则

本目录给出首发门禁阈值。部署环境可收紧，不得在没有风险审批的情况下放宽。表中“待接入”表示代码或 SQL 已有判定依据，但监控采集、通知路由仍需基础设施团队接入；上线检查时必须关闭所有“待接入”。

每条告警必须携带 `tenant_id`、`contract_id`、`user_id`（适用时）、业务号、当前值、阈值、首次发生时间和最近成功水位。恢复需连续三个检测窗口正常，资金类告警还须完成对账。

## 2. 告警清单

| 编号 | 级别 | 检测条件/首发阈值 | 自动保护 | 通知 | 当前接入 |
| --- | --- | --- | --- | --- | --- |
| OPT-A001 | SEV-2 | 交易中合约标的时间戳距当前超过 30 秒，或来自未来 | 拒绝行权及依赖标的价的风险增加 | 值班、行情、风控 | 代码保护、租户指标和规则已实现；生产抓取/通知待接入 |
| OPT-A002 | SEV-2 | 标记价时间戳超过 30 秒，缺失或来自未来 | 卖方/组合保证金准入失败，风险账户转 `RESTRICTED` | 值班、行情、风控 | 代码保护、租户指标和规则已实现；生产抓取/通知待接入 |
| OPT-A003 | SEV-3 | Greeks 时间戳超过产品阈值 | 禁止依赖 Greeks 的新功能；当前基础模型不使用时只告警 | 值班、风控模型 | 缺失/未来指标和规则已实现；合约级批准阈值、生产抓取/通知待接入 |
| OPT-A004 | SEV-2 | 风险账户 `last_calc_time` 超过两个调度周期 | 账户禁止增加风险 | 值班、风控 | 状态保护、租户指标和规则已实现；生产抓取/通知待接入 |
| OPT-A005 | SEV-2 | 一个周期内风险扫描失败账户数大于 0；失败比例超过 5% 升 SEV-1 | 失败钱包转 `RESTRICTED`，其他钱包继续 | 值班、风控、Option 技术 | 失败隔离、实际扫描数/失败数/比例/执行失败指标和规则已实现；生产抓取/通知待接入 |
| OPT-A006 | SEV-2 | `PENDING/PROCESSING/FAILED` 资产指令超过 60 秒；`MANUAL_REVIEW` 任意一笔 | 停止相关业务号后续步骤；超龄处理中指令按原身份恢复重试 | 值班、清算、Asset | 水位/明细/恢复、租户指标和规则已实现；生产抓取/通知待接入 |
| OPT-A007 | SEV-1 | 同一 `instruction_no/biz_no` 出现方向、币种、金额不一致 | 暂停相关合约资金处理 | 技术负责人、清算、合规 | 开放对账租户指标和规则已实现；生产抓取/通知及字段冲突演练待接入 |
| OPT-A008 | SEV-2 | outbox/inbox 最老未完成记录超过 60 秒或连续 3 分钟增长 | 保护依赖状态；禁止到期推进 | 值班、Option 技术 | 数量/最早时间指标、超龄和连续增长规则已实现；生产抓取/通知待接入 |
| OPT-A009 | SEV-2 | 待处理主动行权超过 60 秒；到期前 30 分钟仍非 0 | 到期流程等待；必要时关闭主动行权入口 | 值班、运营、风控 | 超龄及到期前30分钟动态租户指标和规则已实现；生产抓取/通知待接入 |
| OPT-A010 | SEV-2 | 到期窗口结束后 60 秒仍未锁定结算价 | 保持等待，不允许用当前价替代 | 值班、行情、清算 | 租户级逾期合约指标、规则和查询索引已实现；生产抓取/通知待接入 |
| OPT-A011 | SEV-1 | 已发布结算价样本数低于合约最小值、来源不符或窗口外 | 阻止结算批次 | 技术负责人、清算、合规 | 结算证据校验、审批工作台、租户级异常快照指标和规则已实现；生产抓取/通知待接入 |
| OPT-A012 | SEV-2 | 结算批次超过 5 分钟未完成或存在失败指令 | 阻止批次完成和合约归档 | 值班、清算 | 水位/恢复、租户指标和规则已实现；生产抓取/通知待接入 |
| OPT-A013 | SEV-2 | 强平 `PENDING/EXECUTING` 超过 60 秒 | 账户保持风险限制 | 值班、风控、Asset | 水位/恢复、租户指标和规则已实现；生产抓取/通知待接入 |
| OPT-A014 | SEV-1 | 强平缺口未由保险/兜底完整覆盖，或兜底额度超限 | 暂停新增裸卖风险 | 管理层、风控、财务 | 未解决缺口分币种指标和规则已实现；额度配置、生产抓取/通知待接入 |
| OPT-A015 | SEV-1 | 日终完整资金守恒差额非 0，或 scope=2 成功心跳缺失 | 暂停次日相关产品开放 | 技术、清算、财务、合规 | `check_type=3` 开放差额的独立租户指标、规则和记录模板已实现；scope=1 钱包镜像自动任务/心跳已实现，Asset 期初期末、平台账户和全流水 scope=2 生产者仍待实现 |
| OPT-A016 | SEV-3；5分钟超过3次升 SEV-2 | 管理员修改已上市/系列合约状态或经济字段被拒绝 | 无；保留安全审计 | 值班、安全、风控 | 有限标签 Counter、结构化应用日志和两级规则已实现；直接 SQL 尝试仍由数据库审计日志/安全平台采集 |
| OPT-A017 | SEV-2 | 交易态合约任一强制控制参数为 0，或 5 分钟内出现 `CONTROL_NOT_CONFIGURED` 拒单 | 新单拒绝；禁止恢复交易 | 值班、风控、Option 技术 | 数据库/准入保护、租户指标和规则已实现；生产抓取/通知待接入 |
| OPT-A018 | SEV-2 | 单合约 1 分钟内价格带拒单超过 20 笔或超过请求的 10% | 保持价格带；人工核对标记价和参数 | 值班、行情、风控 | 绝对笔数窗口指标和规则已实现；请求比例分母、生产抓取/通知待接入 |
| OPT-A019 | SEV-1 | 触发 `CIRCUIT_BREAKER`，或暂停后仍存在活动订单超过 30 秒 | 合约立即 `PAUSED` 并批量撤单 | 值班、行情、风控、运营 | 暂停/撤单、熔断/残单租户指标和规则已实现；生产抓取/通知待接入 |
| OPT-A020 | SEV-2 | kill switch 启用后 30 秒仍存在该用户活动订单，或撤单/资金释放失败 | 新单持续拒绝，重复执行幂等撤单 | 值班、风控、清算 | 残单及关联释放指令失败/人工态租户指标和规则已实现；生产抓取/通知待接入 |
| OPT-A021 | SEV-3 | 单用户 5 分钟内 STP 达 5 次；单租户 5 分钟超过 50 次升 SEV-2 | cancel-taker，不产生成交 | 风控、合规、Option 技术 | 用户/租户窗口指标和规则已实现；生产分级路由/通知待接入 |
| OPT-A022 | SEV-2 | 单用户/合约限额拒单 5 分钟超过 20 次 | 保持原子拒单；核对是否异常程序化请求 | 风控、运营、Option 技术 | 用户合约窗口指标和规则已实现；生产抓取/通知待接入 |
| OPT-A023 | SEV-1 | 疑似错单或异常成交，或现金更正处于 `EXECUTING` 超过 60 秒/进入 `MANUAL_REVIEW` | 创建案件即暂停合约并撤活动单；先扣后入账，失败不提前付款 | 技术负责人、风控、合规、清算、运营 | 案件/资金保护、待复核/超龄/人工租户指标和规则已实现；生产抓取/通知待接入 |
| OPT-A024 | SEV-2 | MMP 触发后 10 秒仍有同组活动报价，`last_error_msg` 非空，或暂存 `DISABLED` 超过 60 秒 | 继续拒绝该组新 MMP 单；按原订单/业务号幂等撤单 | 值班、做市运营、风控、Option/Asset | 状态保护、错误审计、租户指标和规则已实现；生产抓取/通知待接入 |
| OPT-A025 | SEV-2；版本不一致升 SEV-1 | 存在组合持仓但无当前有效已批准参数；风险账户配置版本落后当前版本两个扫描周期；下单与扫描解析版本不一致 | 拒绝组合卖单，风险账户转 `RESTRICTED`；不得回退默认参数 | 值班、风控模型、Option 技术 | 缺配置和扫描快照版本异常的租户级指标、规则及索引已实现；准入/限制和快照字段已实现；订单未持久化参数版本，准入与扫描同请求版本一致性仍需 E2E 证据 |
| OPT-A026 | SEV-2；已有扣款或逾期升 SEV-1 | 实物交割单元进入 `CURE_REQUIRED/MANUAL_REVIEW/DEFAULTED`，或补资截止后仍未完成 | 阻止该单元入账；健康单元继续；逾期停止自动重试 | 值班、清算、风控、运营、Option/Asset | 普通异常 warning、逾期/违约 critical 指标和规则已拆分；已有扣款需从指令明细下钻；生产抓取/通知待接入 |
| OPT-A027 | SEV-2；生效后未完成或人工态升 SEV-1 | 公司行动到达 `effective_time` 后仍未完成、任一映射连续失败、进入 `MANUAL_REVIEW`，或源/后继合约提前恢复交易 | 源合约保持停牌，后继合约禁止上市，相关钱包风险账户转 `RESTRICTED` | 值班、市场运营、风控、清算、合规、Option/Asset | 到期未完成与执行/映射/合约状态异常已拆分为租户级指标和 SEV-1 规则；生产抓取/通知路由待接入 |
| OPT-A028 | SEV-3；谱系不一致升 SEV-1 | 合约系列待复核超过24小时；已生成数量不等于预计数量；谱系指向缺失合约；上市批准前合约已非 PENDING | 不自动上市；数量/谱系不一致时冻结该系列全部合约的上市审批 | 值班、产品运营、风控、Option 技术 | 24小时复核积压与数量/谱系/状态不变量已拆分为租户级指标和分级规则；生产抓取/通知待接入 |
| OPT-A029 | SEV-2；跨租户/错误状态泄漏升 SEV-1 | 公开链 Call/Put 缺腿；多空 `HOLDING` OI 不等；统计/盘口与事实表抽样不一致；响应超过批准 TTL/SLA | 标记行情延迟或暂时下线对应链/盘口；保持交易事实不变；跨租户泄漏立即关闭公开入口 | 值班、行情运营、风控、Option 技术、安全 | Call/Put 配对与多空 OI 不平衡的租户级事实表指标和规则已实现；统计/盘口抽样、跨租户/CDN TTL/SLA 探针及生产通知待接入 |
| OPT-A030 | SEV-2；残腿/普通簿污染/资金重复升 SEV-1 | 组合父单 `FUNDING/CANCELING` 超过60秒或任意 `MANUAL_REVIEW`；父/腿/影子进度不一致；组合成交组缺腿；任一 `combo_order_id>0` 影子单进入普通盘口/单腿撮合；跨腿扣款屏障失效 | 阻止父单激活或持仓事件；halt/kill switch/强撤按父单全腿取消；人工态停止自动绕过 | 值班、Option 技术、风控、清算、Asset、市场运营 | 数据库水位、普通簿失败关闭、持仓事务二次屏障、Prometheus 指标和规则已实现；生产抓取及通知路由待部署验收 |
| OPT-A031 | SEV-1 | 最近钱包镜像核对有差异或执行失败；超过36小时没有成功心跳 | 保持真实资金门禁关闭；按原租户范围补跑 | 值班、Option 技术、清算、Asset | 单条一致性 SQL、不可变 attempt、`ACCOUNT_MIRROR:{coin}` 案件、成功/差异/失败/缺失指标及规则已实现；生产调度、抓取和通知待验收 |

## 3. 最低 SQL 检测口径

监控实现可优化，但不得改变以下业务口径：

```sql
-- 超龄资产指令
SELECT tenant_id, status, COUNT(*) AS cnt, MIN(create_times) AS oldest
FROM t_option_asset_instruction
WHERE status IN (1, 4, 5)
GROUP BY tenant_id, status;

-- 风险扫描水位和保守限制账户
SELECT tenant_id, settle_coin, status, COUNT(*) AS cnt, MIN(last_calc_time) AS oldest
FROM t_option_risk_account
GROUP BY tenant_id, settle_coin, status;

-- 主动行权积压
SELECT tenant_id, contract_id, COUNT(*) AS cnt, MIN(exercise_time) AS oldest
FROM t_option_exercise
WHERE status = 1
GROUP BY tenant_id, contract_id;

-- 最近交易控制拒绝/保护事件
SELECT tenant_id, contract_id, user_id, event_type, reason,
       COUNT(*) AS cnt, MIN(create_times) AS first_time, MAX(create_times) AS last_time
FROM t_option_trading_control_event
WHERE create_times >= UNIX_TIMESTAMP() - 300
GROUP BY tenant_id, contract_id, user_id, event_type, reason;

-- 组合持仓结算币缺少当前有效批准参数
SELECT DISTINCT p.tenant_id, p.user_id, c.settle_coin
FROM t_option_position p
JOIN t_option_contract c ON c.id = p.contract_id AND c.tenant_id = p.tenant_id
LEFT JOIN t_option_portfolio_risk_config cfg
  ON cfg.tenant_id = p.tenant_id AND cfg.settle_coin = c.settle_coin
 AND cfg.status IN (2,4)
 AND cfg.effective_from <= UNIX_TIMESTAMP()
 AND (cfg.effective_until = 0 OR cfg.effective_until > UNIX_TIMESTAMP())
WHERE p.status = 1 AND c.seller_margin_mode = 2 AND cfg.id IS NULL;

-- 实物交割补资、人工复核和违约单元
SELECT tenant_id, contract_id, status, COUNT(*) AS cnt,
       MIN(cure_deadline) AS earliest_deadline,
       MIN(update_times) AS oldest_update,
       MAX(last_error_msg) AS sample_error
FROM t_option_physical_delivery_unit
WHERE status IN (3,4,6)
GROUP BY tenant_id, contract_id, status;

-- 已启用 kill switch 后仍存在的活动订单
SELECT c.tenant_id, c.user_id, COUNT(o.id) AS active_orders,
       MIN(o.create_times) AS oldest_order
FROM t_option_user_trading_control c
JOIN t_option_order o
  ON o.tenant_id = c.tenant_id AND o.user_id = c.user_id
 AND o.status IN (1, 2, 7)
WHERE c.kill_switch = 1
GROUP BY c.tenant_id, c.user_id;

-- 交易态合约控制参数缺失；正常结果必须为 0 行
SELECT id, tenant_id, contract_code
FROM t_option_contract
WHERE status = 2
  AND (max_user_long_qty <= 0 OR max_user_short_qty <= 0
    OR max_open_interest <= 0 OR order_price_band_ratio <= 0
    OR circuit_breaker_ratio <= 0);

-- 异常成交现金更正积压或人工处理
SELECT tenant_id, contract_id, status, COUNT(*) AS cnt,
       MIN(update_times) AS oldest, MAX(last_error_msg) AS sample_error
FROM t_option_trade_correction
WHERE status IN (3,5)
GROUP BY tenant_id, contract_id, status;

-- MMP 触发/配置暂存及同组残余活动报价
SELECT c.tenant_id, c.user_id, c.contract_id, c.group_code, c.status,
       c.triggered_at, c.update_times, c.last_error_msg, COUNT(o.id) AS active_orders
FROM t_option_mmp_config c
LEFT JOIN t_option_order o
  ON o.tenant_id = c.tenant_id AND o.user_id = c.user_id
 AND o.contract_id = c.contract_id AND o.mmp = 1
 AND o.mmp_group = c.group_code AND o.status IN (1,2,7)
WHERE c.status IN (2,3)
GROUP BY c.id, c.tenant_id, c.user_id, c.contract_id, c.group_code,
         c.status, c.triggered_at, c.update_times, c.last_error_msg;

-- 已到生效时间但未完成、失败或人工处理的公司行动及持仓进度
SELECT a.tenant_id, a.id AS action_id, a.event_no, a.underlying_symbol,
       a.status, a.effective_time, a.last_error_msg,
       SUM(m.position_total) AS position_total,
       SUM(m.position_completed) AS position_completed,
       SUM(m.position_failed) AS position_failed,
       MAX(m.retry_count) AS max_retry_count
FROM t_option_corporate_action a
JOIN t_option_corporate_action_contract m
  ON m.tenant_id=a.tenant_id AND m.action_id=a.id
WHERE a.effective_time<=UNIX_TIMESTAMP() AND a.status IN (2,4,6,7)
GROUP BY a.tenant_id, a.id, a.event_no, a.underlying_symbol,
         a.status, a.effective_time, a.last_error_msg;

-- 超龄系列草案或生成数量/状态/谱系不一致；后半部分正常结果必须为0行
SELECT tenant_id, id AS series_id, series_code, version, status,
       expected_contract_count, generated_contract_count, create_times
FROM t_option_contract_series
WHERE (status=1 AND create_times<UNIX_TIMESTAMP()-86400)
   OR (status=2 AND generated_contract_count<>expected_contract_count);

SELECT s.tenant_id, s.id AS series_id, s.series_code, s.version,
       s.expected_contract_count,
       COUNT(d.id) AS detail_count,
       SUM(CASE WHEN c.id IS NULL OR (s.launch_status<>2 AND c.status<>1)
                THEN 1 ELSE 0 END) AS invalid_contracts
FROM t_option_contract_series s
LEFT JOIN t_option_contract_series_detail d
  ON d.tenant_id=s.tenant_id AND d.series_id=s.id
LEFT JOIN t_option_contract c
  ON c.tenant_id=d.tenant_id AND c.id=d.contract_id
WHERE s.status=2
GROUP BY s.tenant_id,s.id,s.series_code,s.version,s.expected_contract_count
HAVING detail_count<>s.expected_contract_count OR invalid_contracts<>0;

-- 公开交易链 Call/Put 缺腿或重复；正常标准化系列结果必须为0行
SELECT tenant_id, underlying_symbol, expire_time, strike_price,
       SUM(option_type=1) AS call_count, SUM(option_type=2) AS put_count
FROM t_option_contract
WHERE status=2 AND is_deleted=2
GROUP BY tenant_id, underlying_symbol, expire_time, strike_price
HAVING call_count<>1 OR put_count<>1;

-- 已落仓多空单边 OI 不平衡；正常结果必须为0行
SELECT tenant_id, contract_id,
       SUM(CASE WHEN side=1 THEN position_qty ELSE 0 END) AS long_oi,
       SUM(CASE WHEN side=2 THEN position_qty ELSE 0 END) AS short_oi,
       MAX(update_times) AS position_as_of
FROM t_option_position
WHERE status=1 AND position_qty>0
GROUP BY tenant_id, contract_id
HAVING long_oi<>short_oi;

-- 超龄或人工处理的组合父单
SELECT tenant_id,status,COUNT(*) AS cnt,MIN(update_times) AS oldest
FROM t_option_combo_order
WHERE (status IN (1,5) AND update_times<UNIX_TIMESTAMP()-60) OR status=8
GROUP BY tenant_id,status;

-- 组合父单、腿和影子子单数量/身份不一致；正常结果必须为0行
SELECT p.tenant_id,p.id AS combo_order_id,p.combo_no,
       COUNT(l.id) AS leg_count,
       SUM(l.filled_qty<>p.filled_qty*l.ratio
           OR l.unfilled_qty<>p.unfilled_qty*l.ratio
           OR o.id IS NULL OR o.combo_order_id<>p.id
           OR o.combo_leg_no<>l.leg_no) AS bad_legs
FROM t_option_combo_order p
LEFT JOIN t_option_combo_order_leg l
  ON l.tenant_id=p.tenant_id AND l.combo_order_id=p.id
LEFT JOIN t_option_order o
  ON o.tenant_id=l.tenant_id AND o.id=l.child_order_id
GROUP BY p.tenant_id,p.id,p.combo_no,p.filled_qty,p.unfilled_qty
HAVING leg_count NOT BETWEEN 2 AND 4 OR bad_legs<>0;

-- 组合成交组腿号不完整或交易腿数与策略腿数不一致；正常结果必须为0行
SELECT t.tenant_id,t.combo_match_no,COUNT(*) AS trade_legs,
       COUNT(DISTINCT t.combo_leg_no) AS distinct_legs,
       MIN(t.combo_leg_no) AS min_leg,MAX(t.combo_leg_no) AS max_leg
FROM t_option_trade t
WHERE t.combo_match_no<>''
GROUP BY t.tenant_id,t.combo_match_no
HAVING trade_legs<>distinct_legs OR min_leg<>1 OR max_leg<>trade_legs
   OR trade_legs NOT BETWEEN 2 AND 4;
```

所有时间字段当前为 Unix 秒。采集器必须明确单位，禁止以毫秒解释。

Option 任务进程每15秒至多执行一次固定组数的租户聚合查询，并在内部
`:9105/metrics` 暴露通用运营指标：

- `wklive_option_operations_count{tenant_id,category}`：异常/积压数量；
- `wklive_option_operations_oldest_timestamp_seconds{tenant_id,category}`：最早异常 Unix 秒；
- `wklive_option_operations_amount{tenant_id,category,coin}`：保险原始流水代数和（非余额）、
  兜底负债和未解决缺口金额；保险缺口赔付符号修复获批前，不得把该保险类别解释为净变化；
- `wklive_option_operations_sample_success`、`wklive_option_operations_last_success_timestamp_seconds`
  和 `wklive_option_operations_sample_failure_total{stage}`：采样健康度。

采样失败会保留最后一次成功业务序列，同时把 `sample_success` 置0；成功恢复时对上次存在、本次消失的
标签显式写0。不得把 SQL 失败或序列消失解释为业务恢复。通用规则见
`monitoring/option-operations-alert-rules.yml`，覆盖 OPT-A001～A009/A012～A015/
A017～A029/A031（OPT-A003 仅缺失/未来时间）。交易控制/MMP 只输出租户和固定类别：阈值聚合在数据库按
用户/合约/组计算，具体身份从不可变审计与工作台下钻，禁止作为 Prometheus 标签。
价格带请求比例分母仍缺全请求事实；保险流水符号冲突须经财务/清算批准后迁移。
既有时间窗口查询索引由 `20260731_zr_option_operations_monitoring_indexes.sql` 安装；
近到期行权、kill switch 释放失败和实物逾期索引由不可覆盖的增量迁移
`20260731_zs_option_time_sensitive_monitoring_indexes.sql` 安装。
日终钱包镜像 attempt 和心跳表由 `20260731_zt_option_daily_reconciliation_run.sql` 安装；
System 每小时按 UTC 业务日补跑，出现首个成功后当日跳过，显式 tenant 调用仍可生成新 attempt。

后台 `GET /admin/option/operations/overview` 已按租户返回
`comboStaleCount`、`comboManualReviewCount`、`oldestComboExceptionTime`、
`comboInvariantIssueCount` 和 `comboIncompleteMatchGroupCount`。`comboStaleSeconds`
默认60秒，只允许10至300秒；该接口和运营工作台用于人工值班水位，不替代外部监控采集与通知。
“影子单进入普通簿”和“跨腿扣款屏障失效”必须由 matcher/worker 运行时计数器或探针检测，
禁止仅凭表中存在活动影子单判为污染。

Option RPC 已在内部 `:9105/metrics` 暴露以下 `OPT-A030` 指标，Compose 只 `expose` 到后端网络：

- `wklive_option_combo_isolation_violation_total`：普通盘口/单腿 matcher 拒绝影子单；
- `wklive_option_combo_debit_barrier_violation_total`：持仓事务二次校验拒绝缺腿或扣款不完整成交组；
- `wklive_option_combo_debit_barrier_stale_events`：超过60秒仍被扣款屏障阻止的持仓事件；
- `wklive_option_combo_observability_query_failure_total`：运行时水位查询失败。

抓取示例、告警表达式和部署检查见 `monitoring/`。指标标签只使用有限的租户作用域、路径和查询名，
不得加入 UID、订单号或 `combo_match_no`。生产仍须配置真实 Prometheus target、Alertmanager
接收人并完成通知演练，仓库不保存电话、群组或 webhook 凭证。

## 4. 通知与关闭证据

- SEV-1：电话/值班系统即时通知，5 分钟内建立事件指挥。
- SEV-2：值班系统即时通知，10 分钟内确认负责人。
- SEV-3：工作群和工单，30 分钟内确认。
- 关闭证据至少包括告警曲线、最后成功水位、相关业务行、Asset 流水/余额、重试结果和审批记录。
- 静默只能绑定具体租户/合约、明确截止时间和审批单，不允许全局无限期静默。
