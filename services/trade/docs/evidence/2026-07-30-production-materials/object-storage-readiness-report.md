# 灾备对象存储准备报告

日期：2026-07-30  
范围：后台 S3-compatible 配置读取、专用私有备份桶、版本控制和本地加密回归。  
结论：存储基础设施准备通过；用户本次明确决定不上传，因此没有上传业务数据、测试
探针或 manifest，也不把准备检查冒充生产备份或恢复演练。

## 系统配置来源

- 管理后台系统配置键：`OBJECT_STORAGE`；
- 当前类型：`oss_type=3`；
- S3-compatible endpoint：`https://sgp1.vultrobjects.com`；
- 系统附件桶：`exv`；
- 访问密钥只由 `deploy/dbinit/models/objectstorage.go` 从数据库读取并在进程内使用，
  未写入灾备配置、命令参数、日志或证据文件；
- Deploy 内置 `object-storage` 客户端，不要求宿主机安装 AWS CLI。

当前 endpoint 实际为 Vultr Object Storage，而不是自建 MinIO。Vultr 官方兼容矩阵确认
支持 Put/Get、Bucket Policy、Lifecycle 和 Object Versioning，但不支持 Bucket
Replication：
[S3 Compatibility Matrix](https://docs.vultr.com/products/storage/object-storage/s3-compatibility-matrix)。

## 专用私有灾备桶

系统附件上传实现会对附件桶设置公开读取策略，因此灾备命令明确拒绝复用 `exv`。
2026-07-30 使用后台同一 S3-compatible 订阅创建独立桶：

```text
OBJECT_STORAGE_ENSURE_RESULT=PASS
OBJECT_STORAGE_BUCKET=exv-wklive-dr-backup
OBJECT_STORAGE_BUCKET_CREATED=true
OBJECT_STORAGE_VERSIONING=Enabled
OBJECT_STORAGE_POLICY=private
```

随后以只读命令复核：

```text
OBJECT_STORAGE_RESULT=PASS
OBJECT_STORAGE_ENDPOINT=https://sgp1.vultrobjects.com
OBJECT_STORAGE_SYSTEM_BUCKET=exv
OBJECT_STORAGE_BUCKET=exv-wklive-dr-backup
OBJECT_STORAGE_BUCKET_EXISTS=true
OBJECT_STORAGE_VERSIONING=Enabled
OBJECT_STORAGE_POLICY=private
```

`contract-dr-storage-init` 只允许创建与附件桶不同的 DNS-safe 桶，创建后强制启用版本
控制，并要求不存在公开 Bucket Policy。`contract-dr-storage-check` 和每次生产备份
都会再次执行同样的只读门禁。

## 加密密钥准备

为执行技术验证，已生成临时受控 RSA-3072 收件人证书：

- Subject：`CN=wklive-dr-recovery-interim-2026-07-30`；
- SHA-256 指纹：
  `88:3F:34:29:61:59:E5:9D:A5:53:A6:C0:88:B9:8A:C4:3F:2E:61:32:56:D8:46:F4:96:56:F9:3C:80:AE:3E:62`；
- 有效期至：2036-07-27；
- 私钥采用口令加密，私钥与口令文件权限均为 `0600`；
- 文件位于 Git 忽略的 `deploy/secrets/`。

该证书只用于推进技术链路，不是 KMS/HSM，也不能填写正式
`DR_BACKUP_ENCRYPTION` 审批字段。正式环境仍须完成独立密钥托管、轮换、恢复授权和
双人复核。

## 已执行验证

- Deploy DB model 与对象存储客户端单元测试：通过；
- 专用桶创建、私有策略和版本控制：通过；
- 只读远端桶复核：通过；
- 本地 CMS `AES-256-GCM` 加密、回读、解密恢复及篡改拒绝回归：通过；
- 当前数据库估算体积：`2935.4 MiB`；
- 宿主机临时空间预检：约 `18.95 GiB`，满足 `8 GiB` 保护值；
- 未向远端桶上传数据库、明文、密文或 manifest。

## 仍需真实证据

当前按用户决定停止在只读检查边界，不自动重试上传。只有后续明确变更该决定后，才可
继续以下项目：

1. 将一次完整加密数据库备份上传到 `exv-wklive-dr-backup/wklive/mysql/`；
2. 执行加密上传、完整回读和 SHA-256 一致性核验；
3. 使用正式 KMS/HSM 托管密钥在隔离环境完成解密、全库恢复、Binlog/PITR 和事实核对；
4. 证明对象存储与最终生产主机处于独立故障域；Vultr 当前不支持 Bucket Replication，
   如要求多区域容灾还需第二存储位置；
5. 取得正式 RPO、RTO、保留策略、操作人、复核人和审批编号。
