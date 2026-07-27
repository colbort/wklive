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

已经记录的迁移文件不允许修改；校验和发生变化时初始化会失败，必须新增迁移文件。
对于没有 `schema_migrations` 的既有数据库，初始化器会将仓库当前迁移记录为基线，
不会盲目重复执行可能已经落库的 `ALTER TABLE`。

## 配置初始化

`config-seed` 会读取项目中各服务的 `etc/*.yaml`，把本机地址转换为 Compose
服务名后写入 Etcd。修改 YAML 后可重新执行：

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
覆盖；当前脚本为避免配置转义错误，只接受字母、数字、点、下划线和连字符。默认值仅
适合本地环境，生产部署必须换成随机密钥并使用独立的密钥管理方案。

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
