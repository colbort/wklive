# Option Prometheus 接入

Option RPC 在内部端口 `9105` 的 `/metrics` 暴露 Prometheus 指标；业务 gRPC 端口仍为
`8085`。生产环境只应允许监控网络访问 `9105`，不得直接映射到公网。

## 文件

- `prometheus.example.yml`：最小抓取配置，目标为 Compose 服务 `option-rpc:9105`。
- `alertmanager.example.yml`：按 critical/warning/info 路由到 SEV-1/2/3 的示例；包含
  `REPLACE_BEFORE_DEPLOY`，禁止原样部署。
- `option-alert-rules.yml`：`OPT-A030` 的组合隔离、扣款屏障及采集失败规则。
- `option-operations-alert-rules.yml`：资产、对账、风险、事件、行权、结算、强平、
  实物交割、穿仓、行情新鲜度、风险扫描、交易控制、熔断、kill switch、STP、限额、
  异常成交、MMP、正式结算价、组合风险版本、公司行动、合约系列、公开链配对和 OI
  水位、管理员受治理字段越权修改审计、已上报日终守恒差额及钱包镜像日终成功/差异/失败
  、完整资金守恒心跳缺失、保证金币种证据异常、Asset 重复冻结业务键、组合清算证据/连续失效及保险接管库存规则（当前共57条）。
- `../migrations/20260731_zr_option_operations_monitoring_indexes.sql`：既有14项时间窗口与状态聚合索引；
  该已记录迁移的校验和不可修改。
- `../migrations/20260731_zs_option_time_sensitive_monitoring_indexes.sql`：近到期行权、kill switch
  释放失败和实物逾期的3项增量索引；现有环境必须继续执行此迁移，可重复运行。
- `../migrations/20260801_option_portfolio_liquidation_monitoring.sql`：组合清算重复未终态、证据异常和
  最近三次快照失效查询的覆盖索引；可重复运行。
- `../migrations/20260731_zt_option_daily_reconciliation_run.sql`：不可变日终运行记录、钱包镜像
  零差额成功心跳和后续完整资金守恒证据容器。
- `../../asset/migrations/20260801_option_freeze_idempotency_evidence.sql`：Option 冻结业务键覆盖
  索引、唯一历史证据幂等账回填；重复键保留供 `asset_freeze_duplicate`/OPT-A033 人工处置。
- `option-production-readiness.env.example`：生产范围、审批和不可变证据清单。
- `option-production-readiness.sh`：fail-closed 上线门禁；`--repository-only` 只校验仓库事实，
  生产模式还会验证证据哈希、功能开关、Prometheus/Alertmanager 配置和条件审批。

## 部署步骤

1. 将两份 `*-alert-rules.yml` 挂载到 Prometheus `/etc/prometheus/rules/`。
2. 合并 `prometheus.example.yml` 的 scrape job；若使用 Kubernetes，应改成内部
   Service/Pod discovery，但保持 `/metrics` 和端口 `9105`。
3. 执行 `promtool check config` 和 `promtool check rules`。
4. 将 `severity=critical` 路由到 SEV-1 电话/值班系统，将 `severity=warning`
   路由到 SEV-2 值班系统，将 `severity=info` 路由到 SEV-3 安全审计队列。
   联系人和凭证必须在部署环境配置，不写入仓库。
5. 在预生产分别注入普通簿隔离、扣款屏障、资产积压、交易控制拒单窗口、熔断后残单、
   kill switch 残单、MMP 撤单失败、结算价缺失/证据异常、组合模型版本落后、公司行动
   超期/失败、系列谱系不一致和 OI 不平衡，保存 Prometheus 查询结果、Alertmanager
   通知、确认人、恢复时间及连续三个正常窗口。
6. 复制 `option-production-readiness.env.example` 到仓库外安全路径，填写真实报告及 SHA-256；
   执行 `./option-production-readiness.sh /secure/path/option-production-readiness.env`。任一 `FAIL`
   都必须保持交易门禁关闭，不得通过改脚本或填虚假 `true` 绕过。

仓库提交前可执行：

```sh
./services/option/monitoring/option-production-readiness.sh --repository-only
```

生产模式要求执行机安装 `promtool` 和 `amtool`。生产配置必须是渲染后的最终文件，并作为证据
计算 SHA-256；示例配置、空 receiver 或 `REPLACE_BEFORE_DEPLOY` 不能作为上线证据。执行机还必须
能访问 `OPTION_METRICS_URL`：门禁会实时确认关键指标族、最近采样成功且成功时间不超过45秒，
不能只用 `OPTION_PRODUCTION_METRICS_TARGET_VERIFIED=true` 自我声明放行。

## 指标口径

| 指标 | 类型 | 含义 |
| --- | --- | --- |
| `wklive_option_combo_isolation_violation_total` | Counter | 影子单被普通盘口/单腿撮合防线拒绝 |
| `wklive_option_combo_debit_barrier_violation_total` | Counter | 持仓事件事务内二次校验发现成交组或扣款不完整 |
| `wklive_option_combo_debit_barrier_stale_events` | Gauge | 超过60秒且仍被跨腿扣款屏障阻止的持仓事件 |
| `wklive_option_combo_observability_query_failure_total` | Counter | 运行时水位 SQL 失败 |
| `wklive_option_operations_count` | Gauge | 按租户与有限类别汇总的当前异常/积压数量 |
| `wklive_option_operations_oldest_timestamp_seconds` | Gauge | 对应类别最早异常的 Unix 秒 |
| `wklive_option_operations_amount` | Gauge | 按租户、类别、币种/标的汇总的保险原始流水代数和（非余额）、兜底/缺口金额及保险接管数量、标记价值、绝对 Delta |
| `wklive_option_operations_sample_success` | Gauge | 最近一次15秒节流采样是否成功 |
| `wklive_option_operations_last_success_timestamp_seconds` | Gauge | 最近成功采样 Unix 秒 |
| `wklive_option_operations_sample_failure_total` | Counter | 采样失败次数 |
| `wklive_option_risk_scan_groups` | Gauge | 最近一次完成扫描的租户钱包组数 |
| `wklive_option_risk_scan_failed_groups` | Gauge | 最近一次完成扫描的失败钱包组数 |
| `wklive_option_risk_scan_failure_ratio` | Gauge | 最近一次完成扫描的失败钱包比例 |
| `wklive_option_risk_scan_last_completed_timestamp_seconds` | Gauge | 每租户最近完成扫描 Unix 秒 |
| `wklive_option_risk_scan_execution_failure_total` | Counter | 收集/重置阶段的扫描执行失败 |
| `wklive_option_admin_rejected_mutation_total` | Counter | 管理员修改受治理合约状态/经济字段被应用拒绝；仅有限对象与原因标签 |

新增的时间敏感类别包括 `exercise_near_expiry`、`kill_switch_release_failure` 和
`physical_delivery_overdue`；事件水位的连续增长直接由一分钟 Gauge 差值持续三分钟判定。

`tenant_scope=all` 表示任务按所有租户运行；正整数表示指定租户。指标不使用用户、订单号或
成交组号标签，避免无界基数。具体业务身份通过运营工作台和日志下钻。

通用水位的 `tenant_id` 是实际租户 ID；`category` 只能取代码中固定枚举。采样失败会保留
上一次成功序列、将 `sample_success` 置0并增加失败 Counter；恢复后的旧序列会显式置0。
交易控制窗口只输出受影响租户和有限类别：价格带为一分钟内单合约超过20笔，STP 为五分钟内
单用户至少5笔或单租户超过50笔，限额为五分钟内单用户/合约超过20笔。用户、合约和订单明细
保留在不可变审计表及运营工作台，不作为 Prometheus 标签。

交易态合约的标的与标记价按30秒检查缺失、陈旧和未来时间；Greeks 按合约已批准的
`greeks_max_age_seconds` 检查，0、缺失、陈旧和未来时间均报警，不擅用标的/标记价阈值。
风险扫描失败率来自每次实际扫描的钱包
总数与失败数，不用存量 `RESTRICTED` 账户反推。

治理类固定类别如下：

- `settlement_price_overdue`、`settlement_price_invalid`；
- `portfolio_risk_config_missing`、`portfolio_risk_version_mismatch`；
- `corporate_action_due`、`corporate_action_exception`；
- `contract_series_review_stale`、`contract_series_invariant_issue`；
- `public_chain_pair_issue`、`open_interest_imbalance`。

这些类别只使用数据库事实并保持低基数。`OPT-A025` 的“下单与扫描在同一请求内解析同一参数
版本”以及 `OPT-A029` 的统计/盘口抽样、跨租户、CDN TTL/SLA 仍需预生产或外部探针验证，
不能由上述数据库水位替代。

管理员拒绝指标不使用管理员 ID、合约 ID 或请求内容作为标签；这些高基数身份只写结构化应用
日志。应用层任意一次拒绝进入 SEV-3，五分钟超过3次升 SEV-2。直接 SQL 被触发器拒绝时，同一
事务内写入审计表也会回滚，因此必须同时接入 MySQL audit/general log 或等价数据库安全审计，
不能把应用 Counter 当成数据库越权监控的全集。

`daily_conservation_issue` 只表示 Asset/财务日终程序已把非零差额写入
`t_option_reconciliation_issue(check_type=3,status=1)`；它不能证明当天任务执行过。
生产必须另有零差额成功心跳和缺失运行告警，并使用 Asset 同一一致性快照，记录模板见
`docs/templates/option-daily-fund-reconciliation.md`。

`option.ProcessDailyReconciliation` 每日使用一条 MySQL 一致性查询比较 Asset 的
`wallet_type=5` 钱包和 `t_option_account` 镜像，检查总额、可用、冻结、锁定、缺行和遗留
`account_id`。每次尝试追加写入不可变 `t_option_reconciliation_run(scope=1)`；零差额也保存
成功行，差异按 `ACCOUNT_MIRROR:{coin}` 建案。超过36小时无成功、最近一次差异或执行失败均按
OPT-A031 触发 SEV-1。该心跳只证明钱包镜像一致，不替代 scope=2 的完整期初/期末资金守恒。

保险金额类别当前只能作为原始流水交叉证据：表定义要求入金为正、出金为负，但缺口赔付仍按
正数写入。财务/清算批准代码和历史迁移前，禁止把该指标解释为保险基金净变化或 Asset 余额。
