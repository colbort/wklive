# Docker Compose 部署

该编排用于单机开发和集成测试，包含：

- 10 个 RPC 服务
- 6 个 API 服务
- Etcd、MySQL、Redis、MongoDB、Kafka
- 两个 Beanstalkd 延迟队列实例
- Kafka Topic、Etcd 配置、MySQL Schema、迁移和后台账号初始化

## 启动

```bash
cd deploy
cp .env.example .env
# 修改 .env 中两个后台的初始密码
mkdir -p secrets
umask 077
printf '%s\n' '从密钥系统取得的 iTick Token' > secrets/itick_token
./deploy.sh compose-config
./deploy.sh up
./deploy.sh ps
```

首次构建需要下载 Go 模块和容器镜像，耗时会比后续启动长。

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

`db-init` 会在 MySQL 健康后执行：

1. 空库加载 `services/<service>/<service>.sql` 主 Schema。
2. 创建 `schema_migrations` 迁移记录表。
3. 首次接管当前 Schema 时记录已有迁移基线；之后只执行新增迁移。
4. 从仓库根目录 `init.sql` 幂等补充综合后台基础菜单。
5. 再执行 `services/system/migrations` 中菜单、角色、任务的幂等数据迁移；
   迁移中的新增、修正和删除优先于旧的基础菜单。
6. 创建或更新两个后台管理员，并生成 bcrypt 密码。
7. 给管理员绑定对应 `app_scope` 的管理角色和全部菜单权限。

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

`./deploy.sh build` 和 `./deploy.sh up` 会逐个构建应用镜像，避免 Compose/Bake
全量并行编译耗尽 Docker 资源；传入服务名时只构建指定服务。已经构建好镜像时，
可用 `./deploy.sh start` 直接启动整套 Compose 环境。

存在运行中的 Compose 容器时，所有会触发镜像构建的命令都会在每次构建前检查
Docker 虚拟盘，默认至少保留 4 GiB，防止构建过程挤占 MySQL/Mongo 持久化写入空间：

```bash
./deploy.sh disk-check
# 如宿主机容量规划要求更高，可按 KiB 调大，但不能设为 0
DOCKER_MIN_FREE_KB=8388608 ./deploy.sh disk-check
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
cp production-readiness.env.example production-readiness.env
# 填写生产来源、参数、审批人和演练报告的绝对路径
./deploy.sh contract-readiness
```

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
数据库只读检查由 `deploy/dbinit/models` 中的 readiness model 执行，shell 入口不
包含业务 SQL。
公式检查不仅核对输出类型：还会核对来源 Authority 与市场一一映射、INDEX 算法/
版本/权重、MARK 永续来源/版本/基差上限/平滑权重、FUNDING 版本，以及所有公式的
回看窗口和执行周期。
四份证据的必填内容和通过标准见
[`perpetual-delivery-production-evidence-guide.md`](../services/trade/docs/perpetual-delivery-production-evidence-guide.md)。

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

公共配置位于 `deploy/config/common.yaml`。MySQL、MongoDB 和 JWT 密钥由 `.env`
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
