# CMS AES-256-GCM 加密备份烟测

## 目标与边界

本烟测验证 MySQL dump、压缩、公钥信封加密、模拟远端投递、完整回读、认证解密、
隔离恢复和篡改拒绝链路。它只操作两个固定临时数据库，不读取 `wklive` 业务表，
不把同机临时目录冒充异地对象存储，也不把一次性证书冒充生产 KMS/HSM。

执行命令：

```bash
cd deploy
./deploy.sh contract-dr-backup-smoke
```

## 实现约束

- 数据库：`wklive_dr_backup_probe`、`wklive_dr_backup_restore`；
- Dump：MySQL 8.4 容器内匹配版本客户端、`--single-transaction`、
  routines/events/triggers、hex blob、Binlog source data；
- 压缩：gzip；
- 加密：CMS `AES-256-GCM`，RSA-3072 临时收件人证书；
- 投递：同机一次性目录，仅用于烟测；
- 校验：密文 SHA-256、CMS 算法、认证解密、恢复事实和篡改拒绝；
- 清理：临时数据库、明文、密文、私钥、证书和模拟远端目录。

## 首次执行发现并修复的问题

首次执行使用宿主机 MySQL 9.3 `mysqldump --routines` 连接 MySQL 8.4，客户端尝试查询
服务端不存在的 `INFORMATION_SCHEMA.LIBRARIES`。同时 dump 通过管道直接进入 gzip，
POSIX shell 只取得末端 gzip 的退出码，导致上游错误没有阻止命令输出 PASS。

修复后：

1. 当前 Compose 默认使用 MySQL 8.4 容器内匹配版本 `mysqldump`；
2. 外部数据库要求显式指定兼容客户端；
3. dump 先写入独立临时文件并检查命令退出码和非空，再执行 gzip；
4. `schema_migrations` 先检查表是否存在，再单独计数，避免 CASE 子查询被提前求值。

含客户端错误的首次结果不作为通过证据。

## 最终无告警结果

```text
DR_ENCRYPTED_BACKUP_RESULT=PASS
DR_BACKUP_MODE=smoke
DR_BACKUP_ENCRYPTION=CMS-AES-256-GCM
DR_BACKUP_KEY_ID=ephemeral-smoke-key
DR_BACKUP_CIPHER_BYTES=1481
DR_BACKUP_CIPHER_SHA256=95a21e8a1c075d999d4229681b12a1b86105b120d3633ba2bfd2b94f1635b510
DR_BACKUP_READBACK_SHA256=95a21e8a1c075d999d4229681b12a1b86105b120d3633ba2bfd2b94f1635b510
DR_BACKUP_DECRYPT_VERIFIED=true
DR_BACKUP_TAMPER_REJECTED=true
DR_BACKUP_RESTORED_FACT_COUNT=2
DR_BACKUP_DURATION_SECONDS=1
```

退出后数据库查询：

```text
SELECT COUNT(*)
FROM information_schema.schemata
WHERE schema_name IN (
  'wklive_dr_backup_probe',
  'wklive_dr_backup_restore'
);

0
```

## 负向验收

生产模式配置本地目标：

```text
DR_BACKUP_DESTINATION_URI=file:///private/tmp/not-offsite
```

命令退出码为 1，并明确输出：

```text
production DR backup refuses local file destinations; use an offsite s3:// URI
```

篡改负向测试在回读密文中间位置覆盖一个字节；只有 OpenSSL CMS GCM 认证解密返回非零
才设置 `DR_BACKUP_TAMPER_REJECTED=true`。

## 当前结论

加密备份工具链、客户端兼容、上游错误传播、密文回读、认证解密、恢复事实及篡改拒绝
均通过。以下生产事实仍未取得：

- 真实异地 S3 Bucket/Region；
- KMS/HSM Key ID、收件人证书和私钥托管/轮换策略；
- 生产全库对象及不可变保留策略；
- 隔离环境全库认证解密、PITR、节点/可用区切换与回切；
- 获批 RPO/RTO 和正式审批编号。
