# 当前全库本地加密往返验收

日期：2026-07-30  
命令：`./deploy.sh contract-dr-backup-local-verify`  
源库：`wklive`  
操作性质：单事务只读导出；不上传、不创建恢复库、不修改业务表。  
结论：`PASS`。

## 验收目标

在用户明确决定不上传的边界内，验证当前完整业务库能够：

1. 使用与 MySQL 8.4 匹配的 `mysqldump` 生成一致性全量备份；
2. 完成 gzip 压缩，并在压缩成功后立即删除临时 SQL 明文；
3. 使用 CMS `AES-256-GCM` 和 RSA-3072 收件人证书完成信封加密；
4. 本地复制回读密文并核对 SHA-256；
5. 使用与证书匹配的口令加密私钥完成认证解密；
6. 解密后压缩 SQL 的 SHA-256 与加密前严格一致；
7. 拒绝被篡改的 GCM 密文；
8. 退出后清理明文、压缩包、密文、回读文件和解密文件。

私钥与口令文件仅位于 Git 忽略的 `deploy/secrets/`，脚本只接收文件路径，未把口令、
私钥或数据库凭据写入输出、命令参数或本报告。所用临时证书不是正式 KMS/HSM。

## 实跑输出

```text
DR_ENCRYPTED_BACKUP_RESULT=PASS
DR_BACKUP_MODE=local-verify
DR_BACKUP_ENCRYPTION=CMS-AES-256-GCM
DR_BACKUP_KEY_ID=wklive-dr-interim-cert-883f34296159e59d
DR_BACKUP_DESTINATION_URI=NOT_UPLOADED
DR_BACKUP_OBJECT=NOT_UPLOADED
DR_BACKUP_MANIFEST=NOT_UPLOADED
DR_BACKUP_UPLOAD_PERFORMED=false
DR_BACKUP_RAW_SQL_BYTES=2106899349
DR_BACKUP_CIPHER_BYTES=251435534
DR_BACKUP_CIPHER_SHA256=4e94256859587e4162560e181112ffb85483cc3da115d408641183a0e57290b5
DR_BACKUP_COMPRESSED_SQL_SHA256=5148e1dae508454b029af51d59d326ae2cd7f6f9083cd1ce6da06e92e3e77f33
DR_BACKUP_READBACK_SHA256=4e94256859587e4162560e181112ffb85483cc3da115d408641183a0e57290b5
DR_BACKUP_RESTORED_COMPRESSED_SQL_SHA256=5148e1dae508454b029af51d59d326ae2cd7f6f9083cd1ce6da06e92e3e77f33
DR_BACKUP_DECRYPT_VERIFIED=true
DR_BACKUP_TAMPER_REJECTED=true
DR_BACKUP_RESTORED_FACT_COUNT=0
DR_BACKUP_DURATION_SECONDS=65
```

说明：

- 原始 SQL 为 1.962 GiB；
- 加密文件为 239.79 MiB；
- 密文回读哈希与原密文完全一致；
- 解密后的压缩 SQL 哈希与加密前完全一致；
- `DR_BACKUP_RESTORED_FACT_COUNT=0` 是因为该模式按设计不创建恢复库，并非恢复失败；
- `DR_BACKUP_UPLOAD_PERFORMED=false`、三个远端字段均为 `NOT_UPLOADED`；
- 完成后临时目录中不存在 `wklive-dr-backup.*` 或
  `wklive-dr-offsite-smoke.*` 遗留。

## 回归结果

修改后的临时库烟测同时再次通过：

- 两条隔离事实恢复完成；
- 密文和回读 SHA-256 一致；
- 解密后压缩包 SHA-256 一致；
- 篡改密文被拒绝；
- 两个临时数据库和临时目录自动清理。

## 生产边界

本验收增强了当前全库导出和加密往返证据，但不证明以下生产事实：

- 外部对象存储上传与完整回读；
- 正式 KMS/HSM 私钥托管、授权和轮换；
- 隔离实例全库恢复；
- 全库 Binlog/PITR；
- 节点或可用区切换与回切；
- 获批 RPO/RTO 和正式生产审批。

因此 P1-01 灾备项仍保持 `[~]`，完整生产门禁仍不得放行。
