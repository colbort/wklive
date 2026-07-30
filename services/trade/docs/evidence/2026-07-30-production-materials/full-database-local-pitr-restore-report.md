# 当前全库本地 PITR 隔离恢复验收

日期：2026-07-30  
命令：`./deploy.sh contract-dr-backup-local-pitr-restore-verify`  
源库：`wklive`  
恢复环境：临时 `mysql:8.4` 容器，`--network none`，无端口映射。  
结论：`PASS`。

## 验收边界

- 在源库创建本次运行唯一的临时 PITR 探针表，并写入全量快照基线事实；
- `mysqldump --single-transaction --source-data=2` 固化一致性快照及精确
  Binlog 起点；
- 全量快照完成后写入恢复点事实并记录停止位点，再写入恢复点外事实；
- 从源 MySQL 远端读取快照起点至停止位点的 ROW Binlog，重写到隔离恢复库；
- Binlog 尾段截取完成后立即删除源库探针表，不等待整个恢复结束；
- 全量备份执行 CMS `AES-256-GCM` 加密、本地回读、认证解密、哈希校验及
  密文篡改拒绝；
- `DR_BACKUP_UPLOAD_PERFORMED=false`，没有调用对象存储；
- 隔离 MySQL 不联网、不映射端口、不使用生产数据库 Volume；
- 全量恢复后应用 Binlog 尾段，严格要求基线事实和恢复点事实各存在一条，
  恢复点外事实为零；
- 对恢复端全部基础表精确计数并执行 `CHECK TABLE`；
- PASS 前删除临时恢复容器及宿主数据目录。

## 实跑输出

```text
DR_ENCRYPTED_BACKUP_RESULT=PASS
DR_BACKUP_MODE=local-pitr-restore-verify
DR_BACKUP_ENCRYPTION=CMS-AES-256-GCM
DR_BACKUP_KEY_ID=wklive-dr-interim-cert-883f34296159e59d
DR_BACKUP_DESTINATION_URI=NOT_UPLOADED
DR_BACKUP_OBJECT=NOT_UPLOADED
DR_BACKUP_MANIFEST=NOT_UPLOADED
DR_BACKUP_UPLOAD_PERFORMED=false
DR_BACKUP_RAW_SQL_BYTES=2128866819
DR_BACKUP_CIPHER_BYTES=253773681
DR_BACKUP_CIPHER_SHA256=b5bb5e102bd677ac9e2a7d31822f50253cda1004cda5f4586ad3cfad8dfaf6a2
DR_BACKUP_COMPRESSED_SQL_SHA256=41e066c18362e606d356c15c239c5828e6c147785d9f92461792eb45197ddb8b
DR_BACKUP_READBACK_SHA256=b5bb5e102bd677ac9e2a7d31822f50253cda1004cda5f4586ad3cfad8dfaf6a2
DR_BACKUP_RESTORED_COMPRESSED_SQL_SHA256=41e066c18362e606d356c15c239c5828e6c147785d9f92461792eb45197ddb8b
DR_BACKUP_DECRYPT_VERIFIED=true
DR_BACKUP_TAMPER_REJECTED=true
DR_BACKUP_FULL_RESTORE_VERIFIED=true
DR_BACKUP_RESTORE_CLEANUP_VERIFIED=true
DR_BACKUP_RESTORED_TABLE_COUNT=145
DR_BACKUP_SOURCE_TABLE_COUNT=145
DR_BACKUP_RESTORED_CHECK_OK_COUNT=145
DR_BACKUP_RESTORED_TOTAL_ROWS=4099281
DR_BACKUP_RESTORED_COUNTS_SHA256=ec92fb6a8dfb3142cfb64c60fe64b06aea4f26e31dbb29694f17f22f107f0a33
DR_BACKUP_RESTORED_SCHEMA_MIGRATIONS=66
DR_BACKUP_RESTORED_TRADE_ORDERS=28773
DR_BACKUP_RESTORED_TRADE_FILLS=28
DR_BACKUP_RESTORED_CONTRACT_POSITIONS=0
DR_BACKUP_RESTORED_POSITION_HISTORY=0
DR_BACKUP_RESTORED_SETTLEMENT_INSTRUCTIONS=28839
DR_BACKUP_RESTORED_TRADE_EVENTS=57619
DR_BACKUP_RESTORED_TRADE_EVENT_INBOX=57619
DR_BACKUP_RESTORED_FUNDING_BATCHES=4
DR_BACKUP_RESTORED_DELIVERY_BATCHES=1
DR_BACKUP_RESTORED_RECONCILIATION_ISSUES=4
DR_BACKUP_RESTORED_AUTHORITATIVE_SNAPSHOTS=2609136
DR_BACKUP_RESTORED_SNAPSHOT_OUTBOX=30427
DR_BACKUP_PITR_SNAPSHOT_FILE=binlog.000014
DR_BACKUP_PITR_SNAPSHOT_POSITION=751926031
DR_BACKUP_PITR_STOP_FILE=binlog.000014
DR_BACKUP_PITR_STOP_POSITION=754376595
DR_BACKUP_PITR_END_FILE=binlog.000014
DR_BACKUP_PITR_END_POSITION=754378816
DR_BACKUP_PITR_BINLOG_FILE_COUNT=1
DR_BACKUP_PITR_TAIL_BYTES=5115329
DR_BACKUP_PITR_TAIL_SHA256=19a90d53dd2cd831356377a94a3307dccba28ab87ca1ec6b6cbe8474f9c86cc1
DR_BACKUP_PITR_BASELINE_VERIFIED=true
DR_BACKUP_PITR_RECOVERY_POINT_VERIFIED=true
DR_BACKUP_PITR_BOUNDARY_VERIFIED=true
DR_BACKUP_PITR_SOURCE_CLEANUP_VERIFIED=true
DR_BACKUP_PITR_BUSINESS_TABLE_COUNT=144
DR_BACKUP_DURATION_SECONDS=234
```

145 张恢复表由当前 144 张业务基础表和一张仅存在于本次恢复边界内的临时 PITR
证据表组成。恢复点外事务位置 `754378816` 严格晚于停止位置 `754376595`，恢复库
只包含基线事实和恢复点事实，未包含恢复点外事实。

## 独立清理复核

脚本退出后重新检查源库及本机运行环境：

```text
pitr_probe_tables=0
base_tables=144
schema_migrations=66
restore_containers=0
temporary_backup_directories=0
docker_available_kb=4227056
```

完整门禁重新构建 `db-init` 后磁盘一度低于保护线；确认该镜像未被任何容器引用后，
删除可由 `Dockerfile.db-init` 重建的 `wklive-db-init:latest`，最终 Docker 可用空间
回到 4,227,056 KiB，高于 4 GiB 保护线。没有删除容器卷或数据库数据。源库表数
回到 144，未遗留探针表。

## 完整门禁回归

PITR 验收和源库清理后重新执行 `./deploy.sh contract-readiness`：

- 75 PASS / 14 FAIL，最终仍为 `NOT READY`；
- MySQL、Etcd、Market、Trade、Asset、System 均健康；
- INDEX、MARK、FUNDING、DELIVERY 四类输出新鲜；
- Snapshot Outbox 健康；
- OPEN 对账差异和未完成 Settlement Instruction 均为零；
- `AutomaticLiquidation=false`、`CrossMarginTrading=false`；
- 14 项失败仍是行情许可/正式审批、保险基金水位与资金权限、强平窗口、
  正式 RPO/RTO、KMS 加密声明、异地位置、灾备审批和交割产品启用，没有用本地
  PITR 结果替代任何生产事实。

## 客户端兼容性修正

首次执行在截取 Binlog 前发现服务端镜像未包含 `mysqlbinlog`，命令以非零状态退出，
退出清理钩子随后经独立检查确认探针表、恢复容器和临时目录均为零。实现改为在任何
源库变更前要求主机存在 `mysqlbinlog`，随后使用当前已安装的 MySQL 9.3 客户端读取
MySQL 8.4 ROW Binlog；最终全库恢复、停止位点和越界拒绝全部通过。

## 结论和剩余门禁

本项证明当前完整 `wklive` 可以由加密全量快照和真实 ROW Binlog 尾段恢复到指定
停止位点，并且不会越过恢复边界。它关闭本机“全库备份 + Binlog PITR”技术缺口，
但不证明：

- 真实异地对象上传和远端回读；
- 正式 KMS/HSM 私钥托管、解封和轮换；
- 独立故障域或可用区实例恢复；
- 服务流量切换、业务复核及回切；
- 获批的生产 RPO/RTO 和灾备演练审批。

因此 P1-01 仍保持 `[~]`，生产门禁不得放行。
