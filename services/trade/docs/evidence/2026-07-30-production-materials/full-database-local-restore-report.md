# 当前全库隔离恢复验收

日期：2026-07-30  
命令：`./deploy.sh contract-dr-backup-local-restore-verify`  
源库：`wklive`  
恢复环境：临时 `mysql:8.4` 容器，`--network none`，无端口映射。  
结论：`PASS`。

## 隔离和安全边界

- 使用 `mysqldump --single-transaction` 只读导出当前业务库；
- 压缩成功后立即删除临时 SQL 明文；
- CMS `AES-256-GCM` 加密、回读、认证解密及篡改拒绝全部通过；
- `DR_BACKUP_UPLOAD_PERFORMED=false`，不调用对象存储；
- 恢复 MySQL 使用本地已有镜像，不允许拉取新镜像；
- 临时 MySQL 无网络、无端口映射；
- 数据目录绑定到权限受控的宿主临时目录，不使用 `wklive_mysql-data`；
- 恢复完成后逐张基础表精确计数并执行 `CHECK TABLE`；
- PASS 输出前删除临时容器和整个临时数据目录；
- 独立复核未发现带 `wklive.dr.restore=true` 标签的容器或
  `wklive-dr-backup.*` 临时目录遗留。

## 实跑输出

```text
DR_ENCRYPTED_BACKUP_RESULT=PASS
DR_BACKUP_MODE=local-restore-verify
DR_BACKUP_ENCRYPTION=CMS-AES-256-GCM
DR_BACKUP_KEY_ID=wklive-dr-interim-cert-883f34296159e59d
DR_BACKUP_DESTINATION_URI=NOT_UPLOADED
DR_BACKUP_OBJECT=NOT_UPLOADED
DR_BACKUP_MANIFEST=NOT_UPLOADED
DR_BACKUP_UPLOAD_PERFORMED=false
DR_BACKUP_RAW_SQL_BYTES=2115925360
DR_BACKUP_CIPHER_BYTES=252401427
DR_BACKUP_CIPHER_SHA256=027ede6fa275877afaed82388ea234aebe370e02150e8d01f44099805b65d553
DR_BACKUP_COMPRESSED_SQL_SHA256=13dcebd100fcea65111f333adf7b8ee286a7a24359c7a2aeadede38b92362871
DR_BACKUP_READBACK_SHA256=027ede6fa275877afaed82388ea234aebe370e02150e8d01f44099805b65d553
DR_BACKUP_RESTORED_COMPRESSED_SQL_SHA256=13dcebd100fcea65111f333adf7b8ee286a7a24359c7a2aeadede38b92362871
DR_BACKUP_DECRYPT_VERIFIED=true
DR_BACKUP_TAMPER_REJECTED=true
DR_BACKUP_FULL_RESTORE_VERIFIED=true
DR_BACKUP_RESTORE_CLEANUP_VERIFIED=true
DR_BACKUP_RESTORED_TABLE_COUNT=144
DR_BACKUP_RESTORED_CHECK_OK_COUNT=144
DR_BACKUP_RESTORED_TOTAL_ROWS=4077768
DR_BACKUP_RESTORED_COUNTS_SHA256=914879896410c6558426cf5c6fa6ed771141eeb38e3b76e5a17885a8bf4401bc
DR_BACKUP_RESTORED_SCHEMA_MIGRATIONS=66
DR_BACKUP_RESTORED_TRADE_ORDERS=28471
DR_BACKUP_RESTORED_TRADE_FILLS=28
DR_BACKUP_RESTORED_CONTRACT_POSITIONS=0
DR_BACKUP_RESTORED_POSITION_HISTORY=0
DR_BACKUP_RESTORED_SETTLEMENT_INSTRUCTIONS=28537
DR_BACKUP_RESTORED_TRADE_EVENTS=57015
DR_BACKUP_RESTORED_TRADE_EVENT_INBOX=57014
DR_BACKUP_RESTORED_FUNDING_BATCHES=4
DR_BACKUP_RESTORED_DELIVERY_BATCHES=1
DR_BACKUP_RESTORED_RECONCILIATION_ISSUES=4
DR_BACKUP_RESTORED_AUTHORITATIVE_SNAPSHOTS=2594372
DR_BACKUP_RESTORED_SNAPSHOT_OUTBOX=30530
DR_BACKUP_DURATION_SECONDS=264
```

脚本执行时已要求恢复基础表数等于源库基础表数；运行结束后再次只读核对当前源库：

```text
schema_migrations=66
base_tables=144
```

## 结果解释

- 原始 SQL 为 1.971 GiB；
- 密文为 240.71 MiB；
- 恢复端 144 张基础表全部完成计数，144 次 `CHECK TABLE` 全部返回 `OK`；
- 恢复端总行数为 4,077,768；
- 逐表计数按表名排序后的 SHA-256 指纹为
  `914879896410c6558426cf5c6fa6ed771141eeb38e3b76e5a17885a8bf4401bc`；
- 当前服务保持在线，因此没有把恢复结束后的变化源库行数冒充同一快照逐表对比；
  恢复一致性边界由单事务 dump 固化，恢复命令无 SQL 错误且所有表通过结构检查。

## 尚未完成的生产事实

本项证明当前完整业务库可以在隔离 MySQL 8.4 中恢复，但不证明：

- 外部对象存储上传和远端回读；
- 正式 KMS/HSM 私钥托管与轮换；
- 异地实例恢复；
- 本次单独运行没有应用 Binlog；后续全库精确位点验收见
  [`full-database-local-pitr-restore-report.md`](full-database-local-pitr-restore-report.md)；
- 节点/可用区切换及回切；
- 正式 RPO/RTO 和生产审批。

因此 P1-01 仍保持 `[~]`，生产门禁不得放行。
