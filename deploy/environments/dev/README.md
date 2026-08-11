# Docker Compose 部署

该编排用于单机开发和集成测试，包含：

- 10 个 RPC 服务
- 6 个 API 服务
- Admin UI、App Web、App Mobile Web、Chat Admin UI、Chat UI、Liquidity Admin UI
- Etcd、MySQL、Redis、MongoDB、Kafka
- 两个 Beanstalkd 延迟队列实例
- Kafka Topic、Etcd 配置、MySQL Schema、迁移和后台账号初始化

开发环境由 `common/compose.base.yml` 和 `environments/dev/compose.yml` 组合，统一通过根目录
`deploy.sh` 或本目录 `run.sh` 操作。预发布说明见
[`../preprod/README.md`](../preprod/README.md)，不得直接把开发环境 `.env` 或数据卷用于预发布。

## 启动

```bash
cd deploy
cp environments/dev/.env.example environments/dev/.env
# 修改 .env 中两个后台的初始密码
mkdir -p secrets
umask 077
printf '%s\n' '从密钥系统取得的 iTick Token' > secrets/itick_token
./deploy.sh compose-config
./deploy.sh up
./deploy.sh ps
```

首次构建需要下载 Go 模块和容器镜像，耗时会比后续启动长。

开发环境的 6 个 Web UI 也会由 Compose 构建并运行，默认访问地址为：

| 页面 | 地址 |
| --- | --- |
| Admin UI | `http://localhost:18080` |
| App Web | `http://localhost:18081` |
| App Mobile Web | `http://localhost:18084` |
| Chat Admin UI | `http://localhost:18082` |
| Chat UI | `http://localhost:18085` |
| Liquidity Admin UI | `http://localhost:18083` |

开发环境的 Compose `backend` 网络默认使用 `172.20.0.0/16`，`chat-api` 的
`ChatTokenIPWhitelist` 已允许该网段，因此 `app-api` 通过容器网络请求客服 token
时不会依赖动态容器 IP。若该网段与宿主机已有 Docker 网络冲突，可在 `.env` 中修改
`DEV_BACKEND_SUBNET`，配置种子会自动将同一网段写入客服 IP 白名单。

Chat API 和 Chat Admin API 共享 Docker volume `chat-uploads`，并统一挂载到
`/app/chat_uploads`，客服图片不会因为上传服务不同而只存在于某一个容器中。

执行 `./deploy.sh build` 会同时构建这些前端镜像；执行 `./deploy.sh start` 会在镜像变化后重新创建对应容器。`App Mobile Web` 是移动端 Vue 页面在浏览器中的构建，Android/iOS 原生包仍使用 Capacitor 命令构建。

两个 Beanstalkd 实例由 `common/docker/Dockerfile.beanstalkd` 构建：基础 Alpine 多架构清单使用固定
digest，Beanstalkd 固定为 `1.13-r0`，Compose 不设置主机专用 `platform`。因此 Apple Silicon
构建 `linux/arm64`，AMD64 Linux 构建 `linux/amd64`，不会再使用 `schickling/beanstalkd:latest`
的 AMD64 模拟。两个实例分别使用独立 WAL volume，并以 `-f 0` 在确认任务前同步写盘。启动后执行
以下门禁，验证协议健康、实际镜像架构和 WAL 挂载：

```bash
./deploy.sh beanstalk-readiness
```

只检查仓库配置、不要求容器运行时使用：

```bash
./beanstalk-readiness.sh --repository-only
```

在隔离临时容器/volume 中执行重建恢复烟测（不连接运行中的业务队列）：

```bash
./deploy.sh beanstalk-restart-smoke
```

烟测写入唯一任务，删除整个临时容器，再使用同一临时 WAL volume 重建并取回原任务，结束后只删除
名称以 `wklive-beanstalk-restart-smoke-<pid>` 标识的临时容器和 volume。

执行积压、连接中断、`SIGKILL` 和 WAL 恢复容量门禁（同样不连接主/备业务队列）：

```bash
./deploy.sh beanstalk-resilience-smoke
```

该门禁默认批量写入 1000 个唯一任务，要求写入速率至少 50 jobs/s；它持有一条阻塞连接后强杀
临时容器，要求客户端在 10 秒内观察到断开，并使用同一临时 WAL volume 重建，要求 15 秒内恢复协议、
积压计数和全部任务内容。所有容器、volume、tube 和临时文件都以当前进程号隔离并自动清理。
仓库默认值仅是开发/CI 烟测下限，不能代替生产容量结论。预生产应按已审批峰值提高任务量、收紧 RTO，
保存命令、宿主架构、镜像 digest 和完整输出，例如：

```bash
BEANSTALK_RESILIENCE_JOBS=50000 \
BEANSTALK_RESILIENCE_MIN_PUT_RATE=500 \
BEANSTALK_RESILIENCE_MAX_RECOVERY_SECONDS=10 \
./deploy.sh beanstalk-resilience-smoke
```

可配置项还包括 `BEANSTALK_RESILIENCE_IMAGE` 和
`BEANSTALK_RESILIENCE_MAX_DISCONNECT_SECONDS`。任务量上限为 50000，参数必须是正整数。

项目使用 Etcd client `3.7.x`，部署固定使用 Etcd `3.7.1`。不要连接旧的
Etcd `3.5.x`，否则 go-zero RPC resolver 会报告 `old cluster version`。

查看全部日志：

```bash
./deploy.sh logs
```

查看指定服务：

```bash
./deploy.sh logs liquidity-rpc
```

## 对外端口

| 服务 | 端口 |
| --- | ---: |
| payment-api | 3333 |
| app-api | 5555 |
| chat-admin-api | 6666 |
| chat-api | 7777 |
| admin-api | 8888 |
| liquidity-admin-api | 9999 |
| Etcd | 2379 |
| MySQL | 3306 |
| Redis | 6379 |
| MongoDB | 27017 |
| Kafka | 9092 |
| Beanstalkd | 11300、11301 |

RPC 端口 `8080-8089` 只在 Compose 内部网络开放，通过 Etcd 服务发现调用。

## Kafka Topic

`kafka-init` 会在 Kafka 健康后幂等创建业务 Topic 和对应的死信 Topic。单机默认使用
1 个分区和 1 个副本；可通过 `.env` 的 `KAFKA_PARTITIONS` 调整分区数。单节点部署
不能把副本数设置为大于 1。

Consumer Group 由服务首次消费时自动创建，不需要单独初始化。Topic 定义位于
`deploy/kafka-init.sh`。

## MySQL 初始化

开发环境会将仓库根目录只读挂载为 `/workspace`，因此修改各 RPC 的 Schema 或迁移 SQL
后，重新执行 `./deploy.sh db-init` 即可直接验证，无须先构建发布镜像。这个源码直连行为
仅用于本地开发和测试，预发布环境不会继承该挂载。

`db-init` 会在 MySQL 健康后执行：

1. 空库加载 `services/<service>/<service>.sql` 主 Schema。
2. 创建 `schema_migrations` 迁移记录表。
3. 首次接管当前 Schema 时记录已有迁移基线；之后只执行新增迁移。
4. 从仓库根目录 `init.sql` 幂等补充综合后台基础菜单。
5. 再执行 `services/system/migrations` 中菜单、角色、任务的幂等数据迁移；
   迁移中的新增、修正和删除优先于旧的基础菜单。
6. 创建或更新两个后台管理员，并生成 bcrypt 密码。
7. 给管理员绑定对应 `app_scope` 的管理角色和全部菜单权限。
8. 配置了 `secrets/production-operators.env` 时，幂等创建合约生产职责账号并绑定
   最小权限角色。

后台账号由 `.env` 配置：

```dotenv
ADMIN_USERNAME=admin
ADMIN_PASSWORD=replace-admin-password
LIQUIDITY_ADMIN_USERNAME=liquidityadmin
LIQUIDITY_ADMIN_PASSWORD=replace-liquidity-password
```

密码至少需要 12 个字符。修改密码后重新执行以下命令即可更新初始化管理员密码：

```bash
./deploy.sh db-init
```

### 合约生产职责账号

永续与交割合约上线使用六个职责分离账号：

| 账号默认名 | 角色 | 写权限 |
| --- | --- | --- |
| `contract_oncall` | 合约生产值班 | 无；接收行情、Outbox、对账告警并只读处置事实 |
| `insurance_operator` | 保险基金操作员 | 保险基金账户配置与幂等调账 |
| `dr_operator` | 灾备操作员 | 无后台写权限；主机、存储、KMS 权限在基础设施侧配置 |
| `delivery_operator` | 交割发布操作员 | 合约交易对配置 |
| `production_reviewer` | 生产发布复核员 | 无 |
| `production_approver` | 生产发布审批员 | 无 |

初始化方式：

```bash
mkdir -p secrets
cp common/config/production-operators.env.example secrets/production-operators.env
chmod 600 secrets/production-operators.env
# 为六个账号分别设置至少 20 个字符的唯一随机密码
./deploy.sh db-init
```

`deploy/secrets/` 已被 Git 忽略。初始化器只在
`PRODUCTION_OPERATOR_SEED_ENABLED=true` 时读取这些账号；角色定义和权限收敛位于
`services/system/migrations/20260730_add_contract_production_roles.sql`。重复执行会
更新昵称、密码和角色权限，不会给复核或审批账号增加写权限。

这些账号用于系统鉴权、操作日志和责任分离，不能替代真实人员排班、资金批准、异地
存储/KMS 配置或正式发布单。

`./deploy.sh build` 和 `./deploy.sh up` 会逐个构建应用镜像，避免 Compose/Bake
全量并行编译耗尽 Docker 资源；传入服务名时只构建指定服务。已经构建好镜像时，
可用 `./deploy.sh start` 直接启动整套 Compose 环境。

存在运行中的 Compose 容器时，所有会触发镜像构建的命令都会在每次构建前检查
Docker 虚拟盘，默认至少保留 4 GiB，防止构建过程挤占 MySQL/Mongo 持久化写入空间：

```bash
./deploy.sh disk-check
# 如宿主机容量规划要求更高，可按 KiB 调大，但不能设为 0
DOCKER_MIN_FREE_KB=8388608 ./deploy.sh disk-check
# 只调整合约生产门禁检查；默认 2 GiB，不影响构建/部署的 4 GiB 门槛
DOCKER_READINESS_MIN_FREE_KB=3145728 ./deploy.sh contract-readiness
```

空间不足时命令会在构建前失败，不会自动删除数据库卷。`up` 结束后只清理带
`com.docker.compose.project=wklive` 标签且已经悬空的旧镜像，不删除在用镜像、
容器或 Volume。首次冷部署尚无运行容器时会明确跳过虚拟盘预检。

已经记录的迁移文件不允许修改；校验和发生变化时初始化会失败，必须新增迁移文件。
对于没有 `schema_migrations` 的既有数据库，初始化器会将仓库当前迁移记录为基线，
不会盲目重复执行可能已经落库的 `ALTER TABLE`。

## 永续与交割合约生产门禁检查

生产强平和全仓开仓前，先复制只包含审批结果、证据路径及证据 SHA-256 的声明模板。
该文件不能保存行情源密码、API Key 或访问令牌：

```bash
cp operations/contract/production-readiness.env.example \
  operations/contract/production-readiness.env
# 填写生产来源、参数、审批人和演练报告的绝对路径
# 此命令使用独立的 2 GiB 磁盘保留门槛；构建/部署仍为 4 GiB
./deploy.sh contract-readiness
# 必要时可按 KiB 覆盖：DOCKER_READINESS_MIN_FREE_KB=3145728 ./deploy.sh contract-readiness
```

`INSURANCE_FUND_MIN_AVAILABLE` 必须填写经资金/风控审批的正数最低水位，
格式为 `DECIMAL(36,18)`；数据库中同租户、同币种的启用保险基金账户可用余额
必须达到该水位。任意非零余额不再视为资金门禁通过。

四份生产证据除文件和 SHA-256 外，还必须填写各自的
`*_PRODUCTION_APPROVAL_REF`，指向正式发布单、工单或审批归档。当前预生产报告中的
“待签署”或“不适用”不能作为生产审批引用。

如果声明的全部行情 Authority 都是无需凭据的公开 REST 端点，可设置
`PRICE_SOURCE_ACCESS_MODE=PUBLIC_NO_CREDENTIALS`。该模式不会仅凭声明放行：
readiness model 必须从 Authority Registry 确认每个来源都已启用、类型为
`PUBLIC_REST` 且供应商相互独立。行情数据许可仍由
`PRICE_SOURCE_LICENSE_APPROVED` 单独控制。

也可以显式指定另一份声明文件：

```bash
./deploy.sh contract-readiness /secure/path/production-readiness.env
```

检查为只读操作，覆盖三源 INDEX/DELIVERY、INDEX_BASIS MARK、FUNDING 公式、
实时 FINAL_QUOTE 与引擎输出新鲜度、永续/交割产品、保险基金和手续费平台账户、
历史回放、告警投递、资金权限、灾备演练、Outbox、对账与结算水位，以及 Etcd
中的两个生产安全开关。预检期间
`AutomaticLiquidation.Enabled` 和 `CrossMarginTrading.Enabled` 必须仍为
`false`；全部通过也不会自动打开开关，启用仍须走已批准的发布和回滚流程。
数据库只读检查由 `deploy/common/dbinit/models` 中的 readiness model 执行，shell 入口不
包含业务 SQL。
六个合约生产职责账号也由该模型直接核对 `sys_user`、`sys_role`、
`sys_user_role`、`sys_role_menu` 和 `sys_menu`：每个账号必须启用且只绑定声明角色；
值班账号必须具备三类告警读取权限，保险基金操作员只能有 3 个指定写菜单，交割操作员
只能有 1 个合约配置写菜单，灾备、复核和审批账号不得拥有后台写权限。仅在声明文件中
填写任意用户名不能通过。
公式检查不仅核对输出类型：还会核对来源 Authority 与市场一一映射、INDEX 算法/
版本/权重、MARK 永续来源/版本/基差上限/平滑权重、FUNDING 版本，以及所有公式的
回看窗口和执行周期。
四份证据的必填内容和通过标准见
[`perpetual-delivery-production-evidence-guide.md`](../services/trade/docs/perpetual-delivery-production-evidence-guide.md)。

## Option 生产门禁检查

Option 使用独立的 fail-closed 证据门禁，但通过统一部署入口执行。仓库提交阶段先验证规则、
固定标签和监控索引：

```bash
./deploy.sh option-readiness --repository-only
```

生产发布前复制声明模板到仓库外安全路径，填入真实报告、最终 Prometheus/Alertmanager 配置、
功能范围、审批引用和每个文件的 SHA-256：

```bash
cp ../services/option/monitoring/option-production-readiness.env.example \
  /secure/path/option-production-readiness.env
./deploy.sh option-readiness /secure/path/option-production-readiness.env
```

门禁固定检查 49 条通用规则、4 条组合规则、severity/catalog 完整性、17 项监控索引、日终钱包
镜像运行迁移/System 调度和既有
`20260731_zr` 迁移校验和；生产模式还要求 `promtool`/`amtool`、真实告警送达、Asset E2E、
故障注入、容量、日终对账、数据库审计和值班证据。执行机必须能访问声明的
`OPTION_METRICS_URL`；门禁会校验关键指标族、最近采样成功且成功时间不超过45秒，并要求生产
租户 scope=1 钱包镜像核对在36小时内成功，不能只靠
`OPTION_PRODUCTION_METRICS_TARGET_VERIFIED=true` 放行。卖方、组合保证金、实物交割、组合订单、
公开行情和 Greeks 依赖功能按声明范围增加专项证据，未批准功能必须显式为 `false`。

任一失败均输出 `NOT READY` 并返回非零，不会自动启用任何 Option 功能。证据填写和签署说明见
[`option-production-readiness-signoff.md`](../services/option/docs/templates/option-production-readiness-signoff.md)。

### 交割合约启用前技术验收

在取得最终产品启用审批前，可对停用态交割合约执行独立的只读技术验收：

```bash
./deploy.sh contract-delivery-preflight
# 也可显式指定另一份生产声明
./deploy.sh contract-delivery-preflight /secure/path/production-readiness.env
```

该命令要求交割合约仍为停用安全姿态，核对未来交割时间、停开仓/停撮合顺序、逐仓
杠杆、默认杠杆、连续风险档位、DELIVERY 公式及新鲜快照，并确认该新产品没有 Order、
Fill、Position、Position History、Reservation、Settlement Instruction、Delivery
Batch 或 Delivery Settlement 历史事实。所有数据库查询位于
`deploy/common/dbinit/models/deliverypreflight.go`。

通过时输出 `DELIVERY_TECHNICAL_PREFLIGHT=PASS`，同时固定输出
`DELIVERY_PRODUCTION_ENABLE_ALLOWED=false`。这只证明停用态技术配置完整，不能替代
前四组生产门禁、独立产品发布审批或启用后的完整 `contract-readiness`。

### ROW Binlog PITR 冒烟

在正式全库 PITR 演练前，可先验证当前 MySQL 与客户端能否按 ROW Binlog 精确位点
恢复：

```bash
./deploy.sh contract-dr-pitr-smoke
```

命令只使用 `wklive_dr_pitr_probe` 和 `wklive_dr_pitr_restore` 两个临时数据库，
存在同名数据库时拒绝覆盖，并明确禁止使用 `wklive`。恢复点前后各写入一条测试事实，
恢复库必须严格只有恢复点前事实才通过；重放会话不写 Binlog，结束后自动删除两个
临时库。宿主机需要安装兼容的 `mysql` 和 `mysqlbinlog` 客户端，也可分别通过
`MYSQL_CLIENT_BIN`、`MYSQLBINLOG_CLIENT_BIN` 指定路径。

该命令不读取业务表，也不能替代加密异地全量备份、全库 PITR、可用区切换/回切和
正式 RPO/RTO 审批。完整步骤见
`services/trade/docs/perpetual-delivery-disaster-recovery-runbook.md`。

### 加密异地备份

先用两个固定临时数据库验证完整加密链路：

```bash
./deploy.sh contract-dr-backup-smoke
```

烟测使用临时 RSA-3072 收件人证书执行 CMS `AES-256-GCM` 信封加密，模拟远端复制后
回读密文，核对 SHA-256，再解密恢复到隔离数据库并校验事实；同时篡改密文中间字节，
只有认证解密明确拒绝才通过。命令结束后自动删除两个临时数据库、明文、密文、临时
私钥和模拟远端目录。

在明确不上传的情况下，可对当前完整业务库执行本地加密往返校验：

```bash
./deploy.sh contract-dr-backup-local-verify
```

该模式使用 `--single-transaction` 只读导出当前业务库，压缩完成后立即删除临时 SQL
明文，再执行 CMS `AES-256-GCM` 加密、本地复制回读、解密压缩包 SHA-256 核对和密文
篡改拒绝。它要求在 Git 忽略的安全配置中提供与收件人证书匹配的加密私钥及独立口令
文件；所有中间文件均位于权限为 `0700` 的临时目录并在退出时清理。该模式不会连接
对象存储、不会创建恢复库，固定输出 `DR_BACKUP_UPLOAD_PERFORMED=false`。

需要进一步验证当前全库可恢复时，执行：

```bash
./deploy.sh contract-dr-backup-local-restore-verify
```

该模式在完成同一套加密往返后，启动 `--network none`、不映射端口的临时
`mysql:8.4` 容器，并把临时数据目录绑定到宿主受保护目录，不使用生产 MySQL Volume。
恢复完成后逐张基础表执行精确 `COUNT(*)` 和 `CHECK TABLE`；表数、迁移数必须与源库
一致。通过后删除临时容器和数据目录。默认至少要求 12 GiB 宿主临时空间，可通过
`DR_BACKUP_RESTORE_MIN_FREE_KB` 调高，但不能设为 0。

需要在同一隔离环境进一步验证全量快照和真实 ROW Binlog 的精确停止位点时，执行：

```bash
./deploy.sh contract-dr-backup-local-pitr-restore-verify
```

该模式仍不上传。它创建本次运行唯一的源库临时探针表，在全量快照前、恢复点内和
恢复点外分别写入证据事实，从 `mysqldump --source-data=2` 提取快照位点并下载截至
恢复点的 Binlog 尾段。源探针在尾段截取后立即删除；隔离全库恢复后应用尾段，只有
快照事实和恢复点事实存在、恢复点外事实不存在，且全部表计数及 `CHECK TABLE` 通过
才输出 PASS。命令要求主机安装兼容的 `mysqlbinlog`，退出时清理探针、临时容器和
数据目录。

生产执行只接受 `s3://` 目标，明确拒绝 `file://`，防止把同机目录冒充异地备份：

```bash
mkdir -p secrets
cp operations/dr/dr-backup.env.example secrets/dr-backup.env
chmod 600 secrets/dr-backup.env
# 对象存储默认读取管理后台 OBJECT_STORAGE.minio；
# 配置独立私有备份桶、收件人证书、KMS/HSM Key ID、保留期和责任账号
DR_BACKUP_ENV_FILE="$PWD/secrets/dr-backup.env" ./deploy.sh contract-dr-storage-init
DR_BACKUP_ENV_FILE="$PWD/secrets/dr-backup.env" ./deploy.sh contract-dr-backup
```

备份机只需要公开收件人证书；恢复私钥应由 KMS/HSM 或隔离恢复环境托管。命令执行
`--single-transaction` 全量 dump，先检查 `mysqldump` 的独立退出码，再压缩、加密、
上传、回读并核对密文；清单记录 Binlog 前后位点、密文/压缩 SQL 哈希、Key ID、
操作/复核账号及保留期。当前 Compose 可使用容器内与 MySQL 8.4 匹配的客户端；外部
数据库必须指定兼容的 `mysqldump`。`contract-dr-storage-init` 明确拒绝复用系统附件桶，
创建独立私有备份桶并启用版本控制；生产备份每次运行前再次验证桶存在、版本控制启用且
没有公开 bucket policy。对象上传和回读由 Deploy 内置客户端完成，不要求宿主机安装
AWS CLI，数据库配置读取位于 `deploy/common/dbinit/models`。

如灾备必须使用独立于业务附件存储的账号，可设置
`DR_BACKUP_USE_SYSTEM_OBJECT_STORAGE=false`，再在 Git 忽略的 secret 文件中配置独立
S3-compatible endpoint、bucket 和访问凭据。

本机烟测、全库本地加密往返和隔离全库恢复都不能使 `DR_BACKUP_ENCRYPTION`、
`DR_OFFSITE_LOCATION` 或生产灾备审批门禁通过。生产仍须在真实异地对象存储执行，
并在受管密钥环境完成全库 PITR、切换和回切。

### 合并已有 MySQL 的初始化数据

MySQL 已经安装并且 `wklive` Schema 已存在时，可只合并初始化数据，不启动 Compose
中的 MySQL，也不执行建表和结构迁移：

```bash
./deploy.sh data
```

`data` 会幂等执行基础菜单、菜单/角色/定时任务数据迁移，并创建或更新两个后台管理员；
`merge-data` 是同义命令。容器默认通过 `host.docker.internal:3306` 连接宿主机，
连接信息可在 `deploy/.env` 中设置：

```dotenv
MYSQL_EXTERNAL_HOST=host.docker.internal
MYSQL_PORT=3306
MYSQL_USER=root
MYSQL_PASSWORD=your-password
MYSQL_DATABASE=wklive
```

也可以使用 `MYSQL_DSN` 提供完整连接串；设置后它优先于以上字段。该命令要求目标
数据库已经包含业务表；空库请先使用 `./deploy.sh db-init` 完成完整初始化。

需要同时把已有数据库结构升级到仓库当前版本时，使用：

```bash
./deploy.sh database
```

`db-upgrade` 是同义命令。该命令不会启动 Compose 内的 MySQL；它会连接上述已有
MySQL，建立迁移历史、执行明确标记为可安全补齐的结构迁移，然后幂等合并初始化数据。
首次接管没有 `schema_migrations` 的老库时，历史迁移只登记为基线，不会重放；带有
`dbinit:baseline-safe` 标记的结构收敛迁移仍会实际执行，因此不会出现“迁移已登记但
字段仍缺失”的假升级。

## 配置初始化

`config-seed` 会读取项目中各服务的 `etc/*.yaml`，把本机地址转换为 Compose
服务名，并从 Docker Secret `/run/secrets/itick_token` 注入 iTick Token 后写入
Etcd。仓库中的 YAML 只保留 `__ITICK_TOKEN__` 占位符，真实 Token 必须放在 Git
忽略且权限为 `0600` 的 `deploy/secrets/itick_token`，不得提交到仓库。修改 YAML
或轮换 Token 后可重新执行：

```bash
./deploy.sh config
./deploy.sh restart
```

`config` 只运行配置导入，不会启动 Etcd、数据库和 Kafka 等依赖；`seed` 是兼容旧用法
的别名。执行前必须保证 seed 容器可以访问 Etcd。默认通过
`http://host.docker.internal:2379` 连接宿主机上的 Etcd。连接其他 Etcd 时，在
`deploy/.env` 中设置容器可访问的地址：

```dotenv
ETCD_ENDPOINT=http://192.0.2.10:2379
```

仅检查和渲染 Compose 配置时使用 `./deploy.sh compose-config`。

开发配置位于 `deploy/environments/dev/config/common.yaml`。MySQL、MongoDB 和 JWT 密钥由
`environments/dev/.env`
覆盖，iTick Token 由 Docker Secret 文件注入；当前脚本为避免配置转义错误，只接受
字母、数字、点、下划线和连字符。默认值仅适合本地环境，生产部署必须换成随机密钥并
使用独立的密钥管理方案。

删除 MySQL 数据卷会触发全量重新初始化；该操作会清空数据库，只能在确认不需要现有
数据后执行。

## 停止

保留数据卷：

```bash
./deploy.sh down
```

同时删除数据卷：

```bash
./deploy.sh down -v
```
