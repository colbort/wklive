# 生产灾备材料

## 已核实事实

- 当前 Deploy MySQL 开启 ROW Binlog，保留期 30 天，GTID 未启用；
- 本机隔离演练已完成全库一致性备份、恢复和逐表计数；
- 本机演练没有完成异地对象存储、加密密钥托管、Binlog PITR、可用区切换和回切；
- 当前 `DR_RPO_MINUTES=0`、`DR_RTO_MINUTES=0`，未获生产批准；
- 现有预生产报告不能作为生产灾备审批。

## 必须执行的生产演练

1. 配置受 KMS 或等效密钥系统保护的加密备份；
2. 备份写入与主机/主可用区故障域隔离的存储；
3. 记录全量备份 SHA-256、Binlog 起止位点及保留策略；
4. 在隔离实例恢复全量备份，并应用 Binlog 到指定毫秒前的可恢复位置；
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

