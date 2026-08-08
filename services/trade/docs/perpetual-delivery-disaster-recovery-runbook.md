# 永续与交割合约备份及灾备恢复手册

## 1. 目标与事实边界

本手册用于恢复永续、交割合约的可审计交易事实，并验证恢复后不会重复扣款、重复投影仓位或跳过未完成 Saga。

事实源按以下顺序处理：

1. MySQL：订单、成交、仓位、仓位历史、预占、结算指令、Outbox/Inbox、资金费、交割、强平及对账事实；
2. Asset：钱包、冻结、幂等记录和资金流水，仍由 Asset 服务及其数据库负责恢复；
3. Kafka：用于恢复尚未进入 Inbox 的领域事件，不能代替 MySQL 事实；
4. Redis：只保存缓存和分布式锁，不作为资金或仓位事实源；恢复时允许从 MySQL 重建；
5. iTick/Price Engine：恢复不可变行情快照和公式版本后，Trade 才能继续资金费、交割和风险扫描。

生产演练前必须明确 RPO、RTO、备份保留周期、加密方式、异地存储位置和审批责任人。

## 2. 备份前检查

```sql
SELECT COUNT(*) AS open_instructions
FROM t_trade_settlement_instruction
WHERE status IN (1,2,4,5);

SELECT event_status, COUNT(*)
FROM t_trade_event_outbox
GROUP BY event_status;

SELECT status, COUNT(*)
FROM t_itick_snapshot_outbox
GROUP BY status;

SELECT COUNT(*) AS open_reconciliation_issues
FROM t_contract_reconciliation_issue
WHERE status = 1;
```

在线备份使用 `--single-transaction`。若演练要求逐表与当前源库做精确行数对比，应短暂停止 System 调度器，避免备份完成后 `sys_job_log` 等表继续增长；生产热备本身不要求停写，但验证必须使用备份时点的清单或 Binlog 位点，不能拿持续变化的源库做事后等值比较。

## 3. 一致性备份

以下命令仅示意本地 Compose 环境。生产必须把密码放入受管密钥，不得写入脚本或日志。

```bash
docker exec wklive-mysql-1 sh -lc \
  'mysqldump -uroot -p"$MYSQL_ROOT_PASSWORD" \
    --single-transaction \
    --routines --events --triggers \
    --hex-blob --set-gtid-purged=OFF \
    wklive' > /secure-backup/wklive.sql

shasum -a 256 /secure-backup/wklive.sql
```

备份记录至少包含：

- 开始和结束时间；
- 文件大小与 SHA-256；
- MySQL GTID/Binlog 位点；
- `schema_migrations` 数量和最新版本；
- 核心表行数；
- 执行人、审批单和存储位置。

## 4. 隔离恢复验证

恢复目标必须是确认不存在的一次性数据库，禁止覆盖当前 `wklive`：

```sql
SELECT COUNT(*)
FROM information_schema.schemata
WHERE schema_name = 'wklive_dr_verify';
```

结果为 0 后才允许创建和恢复：

```bash
docker exec wklive-mysql-1 sh -lc \
  'mysql -uroot -p"$MYSQL_ROOT_PASSWORD" \
    -e "CREATE DATABASE wklive_dr_verify CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci"'

docker exec -i wklive-mysql-1 sh -lc \
  'mysql -uroot -p"$MYSQL_ROOT_PASSWORD" wklive_dr_verify' \
  < /secure-backup/wklive.sql
```

必须逐表比较源快照清单与恢复库的精确 `COUNT(*)`，并至少单独核对：

```text
schema_migrations
t_trade_order
t_trade_fill
t_contract_position
t_contract_position_history
t_trade_asset_reservation
t_trade_settlement_instruction
t_trade_event_outbox
t_trade_event_inbox
t_contract_funding_batch
t_contract_funding_settlement
t_contract_delivery_batch
t_contract_delivery_settlement
t_contract_liquidation
t_contract_account_liquidation
t_contract_reconciliation_issue
t_itick_authoritative_snapshot
t_itick_snapshot_outbox
```

任何缺表、迁移数不一致、核心事实行数不一致或约束创建失败都判定恢复失败。

### 4.1 Binlog 精确位点冒烟

Deploy 提供只操作两个固定隔离临时库的 ROW Binlog 精确位点冒烟命令：

```bash
cd deploy
./deploy.sh contract-dr-pitr-smoke
```

命令先确认 `wklive_dr_pitr_probe` 和 `wklive_dr_pitr_restore` 均不存在，再建立同构
测试表：恢复点前写入事实 1，记录 Binlog 文件及停止位点，恢复点后再写入事实 2。
随后从远端 Binlog 流按起止位点重放到恢复库，重放会话关闭 Binlog，避免恢复事实
再次写入源 Binlog。只有源库为 2 条、恢复库严格只有事实 1、事实 2 未越过停止位点
时才输出 `DR_PITR_SMOKE_RESULT=PASS`。无论成功失败，命令都只清理这两个临时库，
不读取或修改 `wklive` 业务表。

该命令验证 ROW Binlog、客户端兼容性、位点截断和库名重写链路，不能替代全库恢复、
异地存储、节点/可用区切换或生产 RPO/RTO 演练。

### 4.2 加密备份与远端回读烟测

Deploy 提供隔离加密备份烟测：

```bash
cd deploy
./deploy.sh contract-dr-backup-smoke
```

命令只创建 `wklive_dr_backup_probe` 和 `wklive_dr_backup_restore`，并在临时目录生成
一次性 RSA-3072 收件人证书。源事实经过兼容版本 `mysqldump`、gzip、CMS
`AES-256-GCM` 信封加密，再复制到同机隔离目录模拟对象存储；随后回读密文、核对
SHA-256、认证解密、恢复数据库并精确核对两条事实。命令还会篡改密文，只有 GCM
认证拒绝篡改文件才通过。所有临时数据库、明文、密文、私钥和目录在退出时清理。

不上传的当前全库加密往返验收：

```bash
cd deploy
./deploy.sh contract-dr-backup-local-verify
```

该模式读取 `DR_BACKUP_SOURCE_DB` 的一致性快照，压缩后立即删除 SQL 明文，并使用
`DR_BACKUP_RECIPIENT_CERT`、`DR_BACKUP_RECIPIENT_KEY` 和受保护的
`DR_BACKUP_RECIPIENT_KEY_PASSPHRASE_FILE` 完成加密、复制回读、认证解密、压缩包
哈希核对和篡改拒绝。它不调用对象存储、不创建恢复数据库，固定输出
`DR_BACKUP_UPLOAD_PERFORMED=false`。该结果只能证明当前全库可被完整导出和加密往返，
不能替代隔离全库恢复、PITR 或异地存储。

当前全库隔离恢复验收：

```bash
cd deploy
./deploy.sh contract-dr-backup-local-restore-verify
```

该模式使用本地已有 `mysql:8.4` 镜像启动无网络、无端口映射的临时容器，数据目录绑定
到宿主受保护临时目录，不使用生产 MySQL Volume。解密后的全库备份恢复成功后，对每张
基础表执行精确行数计数和 `CHECK TABLE`，并要求基础表总数及 `schema_migrations`
数量与源库一致。容器和数据目录必须在输出 PASS 前删除；默认要求至少 12 GiB 可用
临时空间。

由于源系统保持在线，该模式不把恢复库行数与恢复完成后的持续变化源库逐表行数直接
比较；一致性边界由 `mysqldump --single-transaction` 固化，恢复端保存逐表计数指纹。
该验收不替代生产 Binlog/PITR、异地存储或可用区切换。

当前全库本地精确位点 PITR 验收：

```bash
cd deploy
./deploy.sh contract-dr-backup-local-pitr-restore-verify
```

该模式在全库隔离恢复基础上，创建本次运行唯一的源库探针表，以
`mysqldump --source-data=2` 固化快照 Binlog 位点，并在快照结束后分别写入恢复点内
和恢复点外事实。命令截取截至恢复点的真实 ROW Binlog 尾段，随后立即删除源探针；
隔离恢复端先导入加密全量快照，再应用重写到恢复库的 Binlog。只有快照事实与恢复点
事实存在、恢复点外事实不存在，且所有基础表计数及 `CHECK TABLE` 均通过才输出
PASS。该模式固定 `UPLOAD_PERFORMED=false`，不替代异地实例、流量切换/回切或正式
RPO/RTO。

生产命令：

```bash
DR_BACKUP_ENV_FILE=/secure/path/dr-backup.env \
  ./deploy.sh contract-dr-backup
```

生产模式具有以下硬约束：

- 默认由 Deploy 内置对象存储客户端通过 models 读取管理后台系统配置
  `OBJECT_STORAGE.minio` 的 S3-compatible endpoint 与访问密钥；灾备配置只声明
  `DR_BACKUP_BUCKET_NAME` 独立私有桶和 `DR_BACKUP_OBJECT_PREFIX`；
- 首次运行 `contract-dr-storage-init` 创建独立备份桶并启用版本控制，明确拒绝复用
  系统附件桶；每次生产备份前重新验证桶存在、版本控制启用且没有公开 bucket policy；
- 上传和回读不要求宿主机安装 AWS CLI，也不会把系统对象存储密钥导出到日志或灾备
  配置文件；
- 可设置 `DR_BACKUP_USE_SYSTEM_OBJECT_STORAGE=false` 使用独立灾备存储账户；
  无论配置来源如何，目标必须为 `s3://bucket/prefix`，`file://` 明确失败；
- 使用 KMS/HSM 管理身份对应的公开 X.509 收件人证书，备份机不需要恢复私钥；
- 必须声明 Key ID、`dr_operator`、`production_reviewer` 和保留天数；
- 全量 dump 必须独立成功后才允许压缩，不能用管道掩盖 `mysqldump` 失败；
- 上传后必须完整回读密文并核对 SHA-256；
- 清单必须记录 Binlog 前后位点、密文和压缩 SQL 哈希、责任账号、Key ID、保留期与
  远端对象 URI。

S3 回读只证明对象传输完整。生产演练还必须在隔离恢复环境通过 KMS/HSM 解封私钥，
完成认证解密、全库恢复、Binlog/PITR、事实核对及切换/回切。

## 5. 服务恢复顺序

1. 恢复 MySQL、Asset 数据库及不可变行情归档；
2. 校验 `schema_migrations`，禁止应用程序连接到缺迁移的库；
3. 启动 Redis；不要恢复过期任务锁，缓存允许重新生成；
4. 启动 Kafka 并确认领域事件 Topic、消费组 Offset 和保留窗口；
5. 启动 Asset、iTick、Trade；
6. 最后启动 System 调度器；
7. 观察 Trade 的四类核心任务：
   `ProcessOrderMatching`、`ProcessPositions`、`ProcessContractSettlements`、`ProcessTradeEvents`；
8. 等待未完成 Settlement Instruction、Outbox 和 Inbox 收敛；
9. 运行全量对账，OPEN 差异必须为 0 或形成经审批的人工处置记录；
10. 自动强平和全仓新增敞口开关保持关闭，直到价格源、账户、告警和资金权限复核完成。

## 6. 恢复通过标准

- 所有迁移存在且校验和一致；
- Order、Fill、Position、History、Reservation、Instruction、Asset Flow、Outbox/Inbox 可相互关联；
- 未完成 Saga 能继续执行，重放不增加重复资金或仓位事实；
- Funding/Delivery Batch 只能在 Asset 流水对账完成后进入终态；
- Snapshot Outbox 回到实时水位；
- 全量对账无未处理差异；
- Trade、Asset、iTick、System 和基础设施全部健康；
- 达到已审批的 RPO/RTO。

## 7. 2026-07-29 隔离演练记录

- 源库：28.23 MiB、140 张基础表；
- 静默点备份：11,975,203 字节，耗时 0.50 秒；
- SHA-256：`1cdd5acc4bb513b2cde353f15feb73ee894e7b5306db53fae3f8e246c46470d4`；
- 隔离恢复耗时：2.68 秒；
- 逐表精确行数：检查 140 张表，差异 0；
- 核心事实：
  `schema_migrations=46`、`orders=40`、`fills=33`、
  `positions=41`、`instructions=109`、`events=216`，源库与恢复库一致；
- 第一次热备后直接与持续写入的源库比较时，仅 `sys_job_log` 多出 97 条；交易事实全部一致。该证据说明验证清单必须绑定备份时点，随后通过静默点演练取得 140/140 一致结果；
- 一次性数据库及临时备份文件已销毁，System 服务恢复健康。

该记录证明本地备份、恢复和事实核验流程可执行；生产异地备份、Binlog 时间点恢复、节点/可用区故障切换及正式 RPO/RTO 仍需生产环境演练。

## 8. 2026-07-30 加密备份烟测记录

- CMS `AES-256-GCM` 加密：PASS；
- 模拟远端回读密文 SHA-256 一致：PASS；
- 认证解密及两条事实隔离恢复：PASS；
- 密文篡改拒绝：PASS；
- 生产 `file://` 目标拒绝：PASS；
- 两个固定临时数据库清理后数量：0；
- 完整结果及首次发现的客户端兼容问题修复记录见
  [`encrypted-backup-smoke-report.md`](evidence/2026-07-30-production-materials/encrypted-backup-smoke-report.md)。

该记录关闭加密备份工具链和失败传播缺口，但不把同机模拟目录、临时证书或测试私钥
声明为生产异地存储/KMS。

## 9. 2026-07-30 当前全库本地 PITR 记录

- 加密全量 SQL：2,128,866,819 bytes；
- 快照位点：`binlog.000014:751926031`；
- 恢复停止位点：`binlog.000014:754376595`；
- 恢复点外事务：`binlog.000014:754378816`，未进入恢复库；
- Binlog 尾段：5,115,329 bytes，SHA-256
  `19a90d53dd2cd831356377a94a3307dccba28ab87ca1ec6b6cbe8474f9c86cc1`；
- 144 张业务表加一张本次证据表，145 次 `CHECK TABLE` 全部 OK；
- 恢复端总行数：4,099,281；
- 基线事实、恢复点事实、越界拒绝：全部 PASS；
- 源探针、临时容器及宿主数据目录：全部清理；
- 上传：`false`；
- 总耗时：234 秒。

完整输出及首次客户端路径失败后的清理复核见
[`full-database-local-pitr-restore-report.md`](evidence/2026-07-30-production-materials/full-database-local-pitr-restore-report.md)。
该记录关闭当前业务库的本机全量快照与真实 Binlog 精确位点恢复缺口，但不替代异地
对象回读、KMS/HSM、独立故障域切换/回切或正式 RPO/RTO 审批。
