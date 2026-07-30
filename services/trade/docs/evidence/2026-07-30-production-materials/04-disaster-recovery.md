# 生产灾备材料

## 已核实事实

- 当前 Deploy MySQL 开启 ROW Binlog，保留期 30 天，GTID 未启用；
- 本机隔离演练已完成全库一致性备份、恢复和逐表计数；
- 已通过两个自动清理的隔离临时库完成 ROW Binlog 精确位点 PITR 冒烟；
- 本机演练没有完成异地对象存储、加密密钥托管、全库 Binlog PITR、可用区切换和回切；
- 当前 `DR_RPO_MINUTES=0`、`DR_RTO_MINUTES=0`，未获生产批准；
- 现有预生产报告不能作为生产灾备审批。

## 2026-07-30 ROW Binlog 精确位点冒烟

通过 `deploy.sh contract-dr-pitr-smoke` 执行，命令在操作前确认
`wklive_dr_pitr_probe` 和 `wklive_dr_pitr_restore` 均不存在，并明确拒绝把
`wklive` 业务库作为测试库。

第二次重复执行的结果：

```text
DR_PITR_SMOKE_RESULT=PASS
DR_PITR_BINLOG_FILE=binlog.000013
DR_PITR_START_POSITION=691892241
DR_PITR_STOP_POSITION=691894421
DR_PITR_END_POSITION=691894747
DR_PITR_SOURCE_COUNT=2
DR_PITR_RESTORED_COUNT=1
DR_PITR_RESTORED_PAYLOAD=before-recovery-point
DR_PITR_DURATION_SECONDS=0
```

源临时库在恢复点前后各写入一条事实，位点重放后的恢复库严格只包含恢复点前事实；
恢复点后的事实未越过停止位点。命令连续执行两次均通过，结束后两个临时数据库数量为
0。负向执行 `DR_PITR_SOURCE_DB=wklive` 在连接数据库前失败。

重放会话显式关闭 Binlog，避免恢复事实被再次写回源 Binlog。该证据证明当前
MySQL 8.4 ROW Binlog、MySQL 9.3 客户端、远端 Binlog 读取、库名重写及精确停止
位点链路可执行，但不把临时表冒烟冒充全库生产 PITR 或异地切换。

## 必须执行的生产演练

1. 配置受 KMS 或等效密钥系统保护的加密备份；
2. 备份写入与主机/主可用区故障域隔离的存储；
3. 记录全量备份 SHA-256、Binlog 起止位点及保留策略；
4. 在隔离实例恢复全量备份，并应用 Binlog 到指定毫秒前的可恢复位置；当前只有隔离
   临时表精确位点冒烟通过，全库生产演练仍待执行；
5. 执行节点或可用区切换、业务复核和回切；
6. 核对迁移、订单、成交、仓位、资金流水、Settlement、Outbox 和对账事实；
7. 以实际丢失窗口及恢复耗时证明获批 RPO/RTO。

## 通过条件

基础设施和安全负责人提供：

- `DR_RPO_MINUTES`；
- `DR_RTO_MINUTES`；
- `DR_BACKUP_ENCRYPTION`（算法、KMS/密钥托管与轮换）；
- `DR_OFFSITE_LOCATION`（异地 Region/机房/存储及保留期）；
- 完整演练报告、复核人、审批编号和
  `DR_EXERCISE_PRODUCTION_APPROVAL_REF`。
