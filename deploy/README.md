# 部署目录

部署文件按职责分为公共层、环境层和运维层：

```text
deploy/
├── common/                 # 公共 Compose、Dockerfile、初始化程序和脚本
├── environments/
│   ├── dev/                # 开发环境 Compose、配置、环境模板和说明
│   └── preprod/            # 预发布 Compose、配置、门禁和说明
├── operations/
│   ├── beanstalk/          # 队列持久化与故障恢复门禁
│   ├── contract/           # 合约生产门禁和交割预检
│   └── dr/                 # 备份与 PITR 演练
├── deploy.sh               # 开发环境兼容入口
└── preprod.sh              # 预发布兼容入口
```

公共层不能包含开发默认端口或某个环境的密钥。开发与预发布必须使用独立 Compose 项目、环境文件和
数据卷，禁止将开发 `.env` 或数据卷用于预发布。

数据库初始化也按环境隔离：开发环境只读挂载仓库，可直接验证 `services` 中各 RPC 的
Schema 和迁移 SQL；预发布环境不挂载源码，只运行带不可变 `RELEASE_TAG` 的 `db-init`
镜像中已经固化并生成 SHA-256 清单的 SQL 发布包。

预发布业务初始化使用独立 Profile，目前只创建一个预发布租户以及 BTCUSDT、ETHUSDT
对应的 Market、Trade、Option 和 Staking 基线资料；风险产品默认禁用或待复核。

## 开发环境

完整说明见 [`environments/dev/README.md`](environments/dev/README.md)。根目录兼容命令继续可用：

```bash
cd deploy
cp environments/dev/.env.example environments/dev/.env
chmod 600 environments/dev/.env
./deploy.sh compose-config
./deploy.sh up
```

也可以直接执行 `environments/dev/run.sh`。若新目录尚未创建 `.env`，脚本会临时兼容原来的
`deploy/.env`。

## 预发布环境

完整说明见 [`environments/preprod/README.md`](environments/preprod/README.md)：

```bash
cd deploy
cp environments/preprod/.env.example environments/preprod/.env
chmod 600 environments/preprod/.env
./preprod.sh validate
./preprod.sh up
```

## 路径约束

所有入口都显式指定 `--project-directory deploy`，因此 Compose 中的相对路径统一相对于 `deploy/`
解析。不要从环境目录直接拼装单个 Compose 文件；必须同时加载：

- `common/compose.base.yml`；
- 对应环境的 `environments/<env>/compose.yml`。

根目录兼容脚本暂时保留，供现有自动化和操作习惯平滑迁移；新增自动化应直接使用环境目录入口。
