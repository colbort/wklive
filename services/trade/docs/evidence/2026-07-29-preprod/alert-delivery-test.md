# 永续与交割合约预生产告警投递测试报告

## 1. 演练范围

- 环境：本机完整 Docker Compose 预生产验收环境
- 开始时间：2026-07-29T09:23:25Z
- 结束时间：2026-07-29T09:26:56Z
- 故障：停止 Kafka 105 秒，使 Snapshot Outbox 真实积压并跨过 60 秒健康阈值
- 恢复：重新启动 Kafka，观察 Outbox 自动重试和排空
- 本地告警载体：iTick 容器结构化 Error 日志
- 外部告警平台：未配置
- 值班组、通知渠道和升级策略：未配置

## 2. 结果

| 项目 | 结果 |
| --- | ---: |
| 峰值 Pending | 177 |
| 峰值 Processing | 64 |
| 峰值不健康记录 | 1105 |
| `snapshot outbox unhealthy` 日志数 | 4 |
| `snapshot outbox worker failed` 日志数 | 0 |
| Kafka 恢复后排空时间 | 103 秒 |
| Pending/Failed/Manual/Unhealthy 最终归零 | true |

原始水位样本：`alert-outbox-samples.tsv`

原始 iTick 日志：`alert-itick.log`

- 水位样本 SHA-256：
  `1f02f46f2683cf6a65f081387084877743dc7661f3391e72dd5593a970efdafc`
- iTick 日志 SHA-256：
  `bf03f928787e540dc4d91d83d65c9529d278e0f554a909b2ebe79e6960442e25`

## 3. 结论

Snapshot Outbox 的阈值告警、积压水位、故障恢复和自动排空已在预生产环境真实执行。
本轮只验证应用日志和本地容器日志采集，不具备外部告警平台、通知回执、值班确认及
未确认升级链路，因此不能作为生产告警投递门禁的最终通过证据。

- 执行人：Codex（按用户授权执行）
- 值班负责人：待生产值班组签署
- 审批编号：预生产技术验收，不适用
