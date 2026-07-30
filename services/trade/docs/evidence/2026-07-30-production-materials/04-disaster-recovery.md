# 生产灾备材料

## 已核实事实

- 当前 Deploy MySQL 开启 ROW Binlog，保留期 30 天，GTID 未启用；
- 本机隔离演练已完成全库一致性备份、恢复和逐表计数；
- 已通过两个自动清理的隔离临时库完成 ROW Binlog 精确位点 PITR 冒烟；
- 已通过两个自动清理的隔离临时库完成 CMS `AES-256-GCM` 加密、模拟远端回读、
  认证解密恢复及密文篡改拒绝烟测；
- 已对当前完整 `wklive` 业务库执行不上传的单事务导出、压缩、CMS
  `AES-256-GCM` 加密、本地回读、认证解密、压缩包哈希核对和篡改拒绝；原始 SQL
  2,106,899,349 bytes，全部临时文件退出后清理；
- 已进一步将当前全库备份恢复到无网络、无端口、宿主临时绑定目录的 MySQL 8.4
  临时容器；源库与恢复库基础表均为 144，迁移均为 66，恢复端 144 张表全部
  `CHECK TABLE=OK`，共 4,077,768 行，容器和临时数据目录全部清理；
- 已在当前完整业务库上进一步完成不上传的真实 ROW Binlog 精确位点恢复：
  加密全量快照位点为 `binlog.000014:751926031`，5,115,329-byte Binlog 尾段应用
  到停止位点 `754376595`；恢复点外事务位于 `754378816` 且未进入恢复库。144 张
  业务表加一张本次临时证据表共 145 张全部 `CHECK TABLE=OK`，源探针、恢复容器和
  数据目录全部清理；
- 生产备份命令只接受 `s3://`，`file://` 负向执行明确失败；
- 管理后台 `OBJECT_STORAGE` 已选择 S3-compatible MinIO 类型，当前配置指向
  Vultr Object Storage 新加坡 endpoint；生产备份预检已从系统配置自动识别 endpoint
  和 bucket，且未在日志输出访问密钥；
- 已创建与公开附件桶隔离的 `exv-wklive-dr-backup` 私有桶并启用版本控制，重复只读
  检查确认不存在公开 Bucket Policy；
- 已准备口令加密私钥对应的临时受控收件人证书，但该证书不是正式 KMS/HSM；
- 用户本次明确决定不上传，因此未执行真实加密上传/回读，也未证明该 bucket 与最终
  生产主机处于独立故障域；
- 本机演练没有完成正式加密密钥托管、异地实例恢复、可用区切换和回切；
- 当前 `DR_RPO_MINUTES=0`、`DR_RTO_MINUTES=0`，未获生产批准；
- 现有预生产报告不能作为生产灾备审批。
- 已创建 `dr_operator`、`production_reviewer` 和 `production_approver`
  作为执行、复核和审批系统身份；
- `dr_operator` 在管理后台只有任务及审计读取权限；主机、异地存储和 KMS 权限仍必须
  由基础设施侧单独授予。

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

## 2026-07-30 CMS AES-256-GCM 加密备份烟测

最终无告警烟测完成：

- 匹配版本 `mysqldump` 独立成功并生成非空备份；
- CMS `AES-256-GCM` 加密及密文算法检查通过；
- 模拟远端上传/回读 SHA-256 一致；
- 认证解密并恢复两条隔离事实；
- 篡改密文被 GCM 认证拒绝；
- 退出后两个临时数据库数量为 0；
- 生产本地 `file://` 目标明确拒绝。

详细输出、首次客户端兼容问题及修复见
[`encrypted-backup-smoke-report.md`](encrypted-backup-smoke-report.md)。该烟测不填写
`DR_BACKUP_ENCRYPTION` 或 `DR_OFFSITE_LOCATION`。后台 S3-compatible 存储只能作为
生产异地存储候选，仍须完成真实上传/回读、权限与保留策略核验，并提供 KMS/HSM。

专用桶、版本控制、私有策略、临时证书和未外发边界见
[`object-storage-readiness-report.md`](object-storage-readiness-report.md)。

当前完整业务库的不上传加密往返输出见
[`full-database-local-encryption-report.md`](full-database-local-encryption-report.md)。

当前完整业务库的隔离全库恢复输出见
[`full-database-local-restore-report.md`](full-database-local-restore-report.md)。

当前完整业务库的加密全量快照与真实 Binlog 精确停止位点恢复输出见
[`full-database-local-pitr-restore-report.md`](full-database-local-pitr-restore-report.md)。

## 必须执行的生产演练

1. 配置受 KMS 或等效密钥系统保护的加密备份；
2. 备份写入与主机/主可用区故障域隔离的存储；
3. 记录全量备份 SHA-256、Binlog 起止位点及保留策略；
4. 在正式异地隔离实例恢复异地全量备份，并应用 Binlog 到指定毫秒前的可恢复位置；
   当前本地全库恢复及真实 Binlog 精确位点已通过，异地故障域复演仍待执行；
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
