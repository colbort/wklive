# Deploy Docker 磁盘预检与 Mongo 保护验收报告

## 1. 触发原因

- 日期：2026-07-29；
- 环境：Deploy 独立完整环境；
- 在构建并切换 Market/Admin 镜像时，Docker 虚拟盘写满；
- Mongo WiredTiger 日志写入返回 `No space left on device`，容器以 133 退出；
- Market/Admin 镜像本身已构建成功，故障发生在 Compose 依赖切换阶段。

处置时只执行 `docker image prune --force`，清理 2.135 GB 未被容器引用的悬空
镜像层。没有删除容器 Volume、Mongo 数据文件或 MySQL 数据。Mongo 随后从最后
检查点完成恢复并重新达到 Healthy。

## 2. 脚本保护

`deploy/deploy.sh` 新增：

- `./deploy.sh disk-check` 只读检查；
- 所有会触发构建的 `build`、`up`、`config/seed`、`data/merge-data`、
  `database/db-upgrade`、`db-init` 和 `contract-readiness` 命令，在每次构建前
  检查 Docker 虚拟盘；
- 默认要求可用空间不少于 `4194304 KiB`（4 GiB），可通过正整数
  `DOCKER_MIN_FREE_KB` 提高；
- 阈值无效或空间不足时在构建前返回非零，不启动构建；
- 不自动删除数据库 Volume；
- `up/start` 完成或失败后，只清理带
  `com.docker.compose.project=wklive` 标签且已经悬空的镜像；
- 首次冷部署没有运行中 Compose 容器时明确提示跳过，避免为检查空间额外拉取镜像。

## 3. 验收结果

| 检查 | 结果 |
| --- | --- |
| `sh -n deploy/deploy.sh` | PASS |
| 当前阈值正向检查 | `available=4968288KiB required=4194304KiB`，exit 0 |
| 不可满足阈值负向检查 | `required=999999999KiB`，构建前拒绝，exit 1 |
| 无效阈值保护 | 代码只接受正整数，0、空值和非数字拒绝 |
| `./deploy.sh start mongo` | Mongo 保持 Healthy |
| 项目级镜像清理 | 命令成功，只匹配 Wklive 悬空镜像 |
| 数据卷 | 未删除；MySQL/Mongo 数据仍可读取 |

验收后 Docker 状态：

- Images：47，Active：46；
- Containers：51，Active：23；
- Local Volumes：25，Active：25；
- Mongo、Market、Admin API、Trade、Asset、System 均为 Healthy。

## 4. 结论

构建不会再在低于保护阈值时继续挤压运行中数据库。脚本只阻止危险构建并回收本项目
已经悬空的镜像，不扩大到容器、数据卷或其他项目。当前 Docker 虚拟盘仍只有约
4.74 GiB 可用，已经接近默认阈值；后续完整构建前应先扩容 Docker Desktop 虚拟盘
或在人工确认后清理不再需要的其他项目容器/镜像，不能删除 Wklive 数据卷。
