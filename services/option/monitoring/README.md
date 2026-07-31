# Option Prometheus 接入

Option RPC 在内部端口 `9105` 的 `/metrics` 暴露 Prometheus 指标；业务 gRPC 端口仍为
`8085`。生产环境只应允许监控网络访问 `9105`，不得直接映射到公网。

## 文件

- `prometheus.example.yml`：最小抓取配置，目标为 Compose 服务 `option-rpc:9105`。
- `option-alert-rules.yml`：`OPT-A030` 的组合隔离、扣款屏障及采集失败规则。
- `option-operations-alert-rules.yml`：资产、对账、风险、事件、行权、结算、强平、
  实物交割、穿仓、行情新鲜度、风险扫描、交易控制、熔断、kill switch、STP、限额、
  异常成交和 MMP 水位规则。
- `../migrations/20260731_zr_option_operations_monitoring_indexes.sql`：时间窗口与状态聚合索引；
  现有环境部署指标前必须执行，可重复运行。

## 部署步骤

1. 将两份 `*-alert-rules.yml` 挂载到 Prometheus `/etc/prometheus/rules/`。
2. 合并 `prometheus.example.yml` 的 scrape job；若使用 Kubernetes，应改成内部
   Service/Pod discovery，但保持 `/metrics` 和端口 `9105`。
3. 执行 `promtool check config` 和 `promtool check rules`。
4. 将 `severity=critical` 路由到 SEV-1 电话/值班系统，将 `severity=warning`
   路由到 SEV-2 值班系统。联系人和凭证必须在部署环境配置，不写入仓库。
5. 在预生产分别注入普通簿隔离、扣款屏障、资产积压、交易控制拒单窗口、熔断后残单、
   kill switch 残单和 MMP 撤单失败，保存 Prometheus 查询结果、Alertmanager 通知、
   确认人、恢复时间及连续三个正常窗口。

## 指标口径

| 指标 | 类型 | 含义 |
| --- | --- | --- |
| `wklive_option_combo_isolation_violation_total` | Counter | 影子单被普通盘口/单腿撮合防线拒绝 |
| `wklive_option_combo_debit_barrier_violation_total` | Counter | 持仓事件事务内二次校验发现成交组或扣款不完整 |
| `wklive_option_combo_debit_barrier_stale_events` | Gauge | 超过60秒且仍被跨腿扣款屏障阻止的持仓事件 |
| `wklive_option_combo_observability_query_failure_total` | Counter | 运行时水位 SQL 失败 |
| `wklive_option_operations_count` | Gauge | 按租户与有限类别汇总的当前异常/积压数量 |
| `wklive_option_operations_oldest_timestamp_seconds` | Gauge | 对应类别最早异常的 Unix 秒 |
| `wklive_option_operations_amount` | Gauge | 按租户、类别、币种汇总的保险/兜底/缺口金额 |
| `wklive_option_operations_sample_success` | Gauge | 最近一次15秒节流采样是否成功 |
| `wklive_option_operations_last_success_timestamp_seconds` | Gauge | 最近成功采样 Unix 秒 |
| `wklive_option_operations_sample_failure_total` | Counter | 采样失败次数 |
| `wklive_option_risk_scan_groups` | Gauge | 最近一次完成扫描的租户钱包组数 |
| `wklive_option_risk_scan_failed_groups` | Gauge | 最近一次完成扫描的失败钱包组数 |
| `wklive_option_risk_scan_failure_ratio` | Gauge | 最近一次完成扫描的失败钱包比例 |
| `wklive_option_risk_scan_last_completed_timestamp_seconds` | Gauge | 每租户最近完成扫描 Unix 秒 |
| `wklive_option_risk_scan_execution_failure_total` | Counter | 收集/重置阶段的扫描执行失败 |

`tenant_scope=all` 表示任务按所有租户运行；正整数表示指定租户。指标不使用用户、订单号或
成交组号标签，避免无界基数。具体业务身份通过运营工作台和日志下钻。

通用水位的 `tenant_id` 是实际租户 ID；`category` 只能取代码中固定枚举。采样失败会保留
上一次成功序列、将 `sample_success` 置0并增加失败 Counter；恢复后的旧序列会显式置0。
交易控制窗口只输出受影响租户和有限类别：价格带为一分钟内单合约超过20笔，STP 为五分钟内
单用户至少5笔或单租户超过50笔，限额为五分钟内单用户/合约超过20笔。用户、合约和订单明细
保留在不可变审计表及运营工作台，不作为 Prometheus 标签。

交易态合约的标的与标记价按30秒检查缺失、陈旧和未来时间；Greeks 当前只检查缺失/未来，
陈旧阈值必须由产品配置后补齐，不能擅用标的/标记价阈值。风险扫描失败率来自每次实际扫描的钱包
总数与失败数，不用存量 `RESTRICTED` 账户反推。
