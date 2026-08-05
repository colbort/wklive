# 交易所预发布环境部署与验收

> 状态：`IMPLEMENTED / RUNTIME_ACCEPTANCE_PENDING`  
> 适用范围：单机或单节点预发布环境，不是生产高可用部署方案  
> 编排入口：`deploy/preprod.sh`  
> Compose：`common/compose.base.yml + environments/preprod/compose.yml`

## 1. 目标

预发布环境用于在进入生产前完成：

- 与生产模式接近的 Go 服务运行配置；
- RPC、API、Admin UI 和 App Web 全链路验证；
- 数据库迁移和 Kafka Topic 初始化；
- 真实 MySQL、Redis、Mongo、Kafka、Etcd 和双 Beanstalkd 验证；
- Trade、Option、Staking、Payment、Liquidity 的故障与恢复演练；
- 菜单、RBAC、总后台/租户管理员隔离验收；
- 发布镜像、回滚步骤、告警和证据归档验证。

预发布不得直接连接生产数据库、Redis、Kafka、Etcd、对象存储或生产用户数据。

## 2. 与开发环境的区别

| 项目 | 开发环境 | 预发布环境 |
| --- | --- | --- |
| Compose 项目 | `wklive` | 默认 `wklive-preprod` |
| 数据卷 | 开发卷 | 独立预发布卷 |
| 服务 Mode | `dev/pro` 混合 | 统一渲染为 `pre` |
| 基础设施端口 | 发布到宿主机 | 不发布，只允许 Compose 内网 |
| API/Web 端口 | 直接发布 | 默认仅绑定 `127.0.0.1` |
| 密码 | 支持开发默认值 | 全部必填、最小长度和唯一性校验 |
| MySQL 运行账号 | root | 独立 `MYSQL_APP_USER`，root 仅用于迁移 |
| Redis | 无密码 | 强制密码 |
| MySQL binlog | 未强制 | ROW binlog，保留 7 天 |
| Admin 请求加密 | 默认关闭 | 强制 `REQUIRED` |
| 容器根文件系统 | 可写 | Go/Web 应用只读，临时目录使用 tmpfs |
| 容器权限 | 默认 | drop capabilities、禁止提权 |
| 镜像 | 本地开发标签 | 必须使用非 `latest` 的 `RELEASE_TAG` |
| 高风险业务开关 | 仓库默认关闭 | Readiness 再次确认关闭 |
| 前端 | 通常本机 Vite | Admin/App/客服/做市 Web 镜像 |

## 3. 当前拓扑

预发布包含：

- 10 个 RPC 服务；
- 6 个 API 服务；
- Admin UI、App Web、Chat Admin UI、Liquidity Admin UI；
- Etcd、MySQL、Redis、MongoDB、Kafka；
- 两个带 WAL 的 Beanstalkd 实例；
- 数据库迁移、管理员/RBAC 初始化、Kafka Topic 初始化和 Etcd 配置种子。

这是单节点 Compose 预发布环境，基础设施没有多副本，不能用它证明生产高可用。生产仍需独立完成
MySQL、Kafka、Redis、Etcd、对象存储、负载均衡和跨可用区设计。

## 4. 宿主机要求

- Linux AMD64/ARM64 或 Docker Desktop；
- Docker Engine 和支持 `!override` 的 Docker Compose v2；
- 建议至少 8 CPU、16 GiB 内存、100 GiB 可用磁盘；
- 正确配置 NTP，时钟偏差应小于 3 秒；
- 宿主机前置 TLS 反向代理或仅通过 SSH Tunnel 访问回环端口；
- 预发布专用 DNS、证书、告警接收器和密钥目录；
- 禁止与开发环境共用 Compose 项目名、数据卷或密钥文件。

## 5. 准备密钥

建议使用主机密钥目录，例如：

```bash
install -d -m 700 /secure/wklive/preprod

openssl genpkey -algorithm RSA \
  -pkeyopt rsa_keygen_bits:3072 \
  -out /secure/wklive/preprod/admin-api-private.pem

cp deploy/common/config/production-operators.env.example \
  /secure/wklive/preprod/production-operators.env

chmod 600 \
  /secure/wklive/preprod/admin-api-private.pem \
  /secure/wklive/preprod/production-operators.env
```

从密钥系统取得 iTick Token，不要写入 Git：

```bash
printf '%s\n' 'TOKEN_FROM_SECRET_MANAGER' \
  > /secure/wklive/preprod/itick_token
chmod 600 /secure/wklive/preprod/itick_token
```

编辑 `production-operators.env`，启用预发布职责账号并为每个账号设置不同的随机密码。文件继续使用
生产最小权限角色定义，但身份和密码必须是预发布专用，不能复用生产账号。

## 6. 准备环境文件

```bash
cd deploy
cp environments/preprod/.env.example environments/preprod/.env
chmod 600 environments/preprod/.env
```

必须修改所有 `replace-*` 值。密码只允许字母、数字、点、下划线和短横线，以避免 DSN 和配置渲染歧义。

推荐生成方式：

```bash
openssl rand -hex 24   # 48 字节安全字符，可用于普通密码
openssl rand -hex 16   # 精确 32 个 ASCII 字符，用于 ADMIN_SESSION_WRAP_KEY
```

`RELEASE_TAG` 必须是不可变版本，例如完整 Git SHA 或已冻结的发布候选标签，禁止使用 `latest`。

`ADMIN_ALLOWED_ORIGIN` 应填写 TLS 反向代理后的 Admin UI 地址。只有本机验证才允许
`http://127.0.0.1:<port>`。

## 7. 静态门禁

不启动容器即可检查仓库预发布约束：

```bash
./preprod-readiness.sh --repository-only
```

加载真实环境文件并验证密钥、文件权限、发布标签和 Compose 渲染：

```bash
./preprod.sh validate
```

验证失败时不会构建、启动或修改容器。

前端发布门禁至少执行：

```bash
npm --prefix app-packages audit --omit=dev
npm --prefix admin-ui audit --omit=dev
npm --prefix app-web audit --omit=dev
npm --prefix chat-admin-ui audit --omit=dev
npm --prefix liquidity-admin-ui audit --omit=dev

npm --prefix admin-ui run build
npm --prefix app-web run build
npm --prefix chat-admin-ui run build
npm --prefix liquidity-admin-ui run build
```

仓库已提交 4 个 UI 和 `app-packages` 的 lock 文件。共享 Axios 固定在已修复的 `1.18.x`，PostCSS
通过 override 固定在 `8.5.18` 以上；不得用 `npm audit fix --force` 绕过兼容性评估。完整 audit
仍可能报告只用于构建的开发工具项，是否放行必须结合最终镜像检查；最终 Nginx 镜像只包含 `dist`，不包含
Node.js、npm 或 `node_modules`，但 `--omit=dev` 的生产依赖审计必须为零。

当前仓库门禁和 UI 构建已完成；由于真实 `environments/preprod/.env`、预发布专用密钥和预发布主机尚未在仓库内
提供，本文状态保持 `RUNTIME_ACCEPTANCE_PENDING`。只有真实环境执行 `./preprod.sh up` 和
`./preprod.sh readiness` 全部通过后，才能改为运行时已验收。

## 8. 构建与启动

### 8.1 在预发布主机本地构建

`environments/preprod/.env` 保持：

```dotenv
PREPROD_BUILD_LOCAL=true
RELEASE_TAG=<immutable-release-tag>
```

执行：

```bash
./preprod.sh up
```

脚本逐个构建服务，使用 `RELEASE_TAG` 标记镜像，启动后自动执行运行时 Readiness。

### 8.2 从镜像仓库部署

```dotenv
PREPROD_BUILD_LOCAL=false
PREPROD_IMAGE_REGISTRY=registry.example.com/wklive
RELEASE_TAG=<immutable-release-tag>
```

```bash
docker login registry.example.com
./preprod.sh pull
./preprod.sh start
./preprod.sh readiness
```

镜像仓库流程应在 CI 中完成 Go 测试、UI type-check/build、镜像扫描、SBOM 和签名。预发布主机只按
已审批 digest 拉取；当前 Compose 使用 tag，正式流水线应额外归档各服务解析后的 image digest。

## 9. 访问入口

默认全部绑定 `127.0.0.1`：

| 服务 | 默认宿主端口 |
| --- | ---: |
| Admin UI | 18080 |
| App Web | 18081 |
| Chat Admin UI | 18082 |
| Liquidity Admin UI | 18083 |
| Payment API | 13333 |
| App API | 15555 |
| Chat Admin API | 16666 |
| Chat API | 17777 |
| Admin API | 18888 |
| Liquidity Admin API | 19999 |
| Option Metrics | 19105 |

MySQL、Redis、Mongo、Kafka、Etcd 和 Beanstalkd 不发布宿主端口。维护操作应使用：

```bash
docker compose \
  --project-directory deploy \
  --env-file deploy/environments/preprod/.env \
  -p wklive-preprod \
  -f deploy/common/compose.base.yml \
  -f deploy/environments/preprod/compose.yml \
  exec mysql mysql -uroot -p
```

不要为了临时调试修改 Compose 暴露数据库端口。需要远程访问时使用受控的 SSH、堡垒机或临时运维容器。

## 10. 运行时验收

### 数据库发布边界

预发布环境不会把代码仓库或 `services/*` 挂载进 `db-init`。构建带不可变
`RELEASE_TAG` 的 `db-init` 镜像时，会将 `init.sql`、各 RPC 主 Schema 和版本化迁移 SQL
复制到镜像内的 `/release`，同时生成 `MANIFEST.sha256`。预发布启动时只执行该发布镜像
中已经固化的 SQL，不会读取宿主机当前工作区，也不会因为开发人员临时修改 SQL 而改变
待执行内容。

因此预发布数据库变更流程为：新增迁移文件、代码评审、构建并推送同一发布版本镜像、
在预发布执行迁移和验收。已经应用的迁移仍受 `schema_migrations` 校验和保护，禁止原地
修改。开发环境则保留源码只读挂载，便于直接测试 RPC SQL。

### 预发布业务基线数据

`db-init` 的 `preprod` Profile 会幂等加载
[`data/bootstrap.sql`](data/bootstrap.sql)。租户身份由 `.env` 的
`PREPROD_TENANT_ID`、`PREPROD_TENANT_CODE` 和 `PREPROD_TENANT_NAME` 决定；首次初始化
后 ID 与 Code 不允许映射到其他租户。

当前基线严格限定为 `BTCUSDT`、`ETHUSDT`：

| 范围 | 初始化内容 | 初始安全状态 |
| --- | --- | --- |
| Tenant | 一个隔离的预发布租户 | 启用 |
| Market | crypto Category、两个 Product、租户 Category/Product、24x7 日历 | 行情可见 |
| Price Engine | 每个交易对的 INDEX、MARK、FUNDING、DELIVERY 公式 | 未启用，等待行情源和参数审批 |
| Trade | 每个交易对的现货、永续、交割、秒合约及逐仓杠杆/风险档位 | 仅现货启用；其余产品禁用 |
| Option | BTC、ETH 各一个 Call/Put 基线合约、24x7 日历草案、组合风险草案 | 合约待上市，卖方/自动行权关闭 |
| Staking | BTC、ETH 活期产品 | 禁用、APR 和额度为零 |

基线不会创建业务用户、充值、资产余额、订单、持仓、假行情、保险基金资金或生产审批
证据。需要开放合约、Option 卖方、Staking 收益或 Price Engine 公式时，必须在 Admin 中
补齐真实参数并走对应复核与门禁。扩展第三个交易对时，应同时更新基线 SQL、行情配置、
Readiness 计数及发布说明，不能直接在宿主机临时追加 SQL。

```bash
./preprod.sh readiness
```

门禁检查：

- 所有长期运行容器存在并健康；
- `db-init`、`kafka-init`、`config-seed` 成功完成；
- 基础设施没有宿主端口；
- Trade、Option 配置运行在 `pre` 模式；
- 自动强平和全仓开仓保持关闭；
- Option 卖方交易等可选能力保持关闭；
- Admin 请求加密为 `REQUIRED`；
- 业务服务使用非 root MySQL 账号。

随后按业务执行：

```bash
./preprod.sh ps
./preprod.sh logs trade-rpc
./preprod.sh logs option-rpc
./beanstalk-readiness.sh --repository-only
./preprod.sh readiness
```

Trade 和 Option 的生产门禁仍按各自文档执行。预发布 Readiness 通过只表示环境和默认安全边界正确，
不代表已经批准自动强平、全仓交易、Option 卖方交易、实物交割、MMP 或生产上线。

## 11. 测试数据原则

- 使用预发布专用租户、用户、账户、币种和交易对；
- 禁止复制生产明文用户资料、密码、支付信息或私钥；
- 需要生产分布特征时只能使用脱敏、不可逆、已审批的快照；
- 自动测试使用唯一业务号，结束后保留必要审计事实；
- 不直接删除资金流水、指令、Inbox/Outbox 或对账证据；
- 资金使用无真实价值的预发布资产，并明确与生产资产隔离。

## 12. 发布流程

1. 冻结提交并生成不可变 `RELEASE_TAG`；
2. 完成 Go/UI/数据库/镜像门禁；
3. 备份预发布数据库并记录 binlog 位点；
4. `./preprod.sh validate`；
5. `./preprod.sh up` 或 pull/start；
6. `./preprod.sh readiness`；
7. 执行 Admin、Trade、Option、Staking、Payment、Liquidity E2E；
8. 执行失败注入、重启恢复、重复请求和对账测试；
9. 归档镜像 digest、迁移版本、配置哈希、测试输出和责任人；
10. 预发布观察期结束后才能进入生产审批。

## 13. 回滚

应用镜像回滚：

```bash
./preprod.sh rollback <previous-immutable-release-tag>
```

该命令只切换应用镜像，不回滚数据库迁移。执行前必须确认新迁移对旧应用向后兼容。

数据库迁移采用前向修复原则：

- 已执行迁移不得修改或删除；
- 发现问题时新增兼容迁移；
- 只有完成正式恢复演练并批准后，才允许从备份/PITR 恢复整个隔离环境；
- 回滚前后必须重新执行 Readiness 和账实对账。

## 14. 停止与数据保护

停止容器但保留数据：

```bash
./preprod.sh stop
```

删除容器和网络但保留数据卷：

```bash
./preprod.sh down
```

预发布脚本故意不提供 `down -v` 或 destroy 命令。删除 MySQL、Kafka、Redis、Mongo、Etcd 或
Beanstalkd 数据卷属于破坏性操作，必须先确认目标项目名、完成备份并单独审批。

## 15. 尚未替代的生产能力

本预发布环境不能替代：

- 多节点数据库和自动故障转移；
- Kafka 多 Broker 和副本容灾；
- Redis Cluster/Sentinel；
- Etcd 奇数节点集群；
- 多可用区负载均衡；
- WAF、生产证书和边界防火墙；
- 独立 Prometheus、Alertmanager、日志平台和值班系统；
- 异地加密备份、KMS 和灾难恢复；
- 正式行情许可、资金准备、合规和业务审批。

这些项目必须在生产上线清单中独立验收，不能因为单节点预发布环境通过而标记完成。
