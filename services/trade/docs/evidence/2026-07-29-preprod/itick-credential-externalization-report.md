# iTick 凭据外置化验收报告

## 1. 结论

2026-07-29 将 iTick Token 从受版本控制的
`services/market/etc/market.yaml` 移出。YAML 现在只保留
`__ITICK_TOKEN__` 占位符；Deploy 使用 Docker Secret
`/run/secrets/market_token`，由 `config-seed` 在写入 Etcd 前注入。

本次改动只改进凭据保管和运行时注入方式，不代表行情凭据、数据许可或供应商生产审批
已经通过；对应 readiness 字段继续保持 `false`。

## 2. 安全边界

- 本机 Secret 文件为 `deploy/secrets/market_token`；
- `deploy/secrets/` 已由 Git 忽略；
- Secret 文件权限为 `0600`；
- 受版本控制文件中已找不到迁移前的 Token 值；
- Compose 声明只引用 Secret 文件，不把 Token 放入环境变量；
- `seed-etcd.sh` 在 Secret 缺失、不可读、为空或包含不允许字符时立即退出；
- 日志、报告和校验和文件均不记录 Token 内容或 Token 哈希。

## 3. 静态验收

- `sh -n deploy/seed-etcd.sh`：PASS；
- `docker compose -f deploy/docker-compose.yml config -q`：PASS；
- 缺失 Secret 负向测试：退出码 1，提示
  `ITICK_TOKEN secret file is missing or unreadable`；
- Git ignore、`0600` 权限和受控文件无原值检查：PASS；
- `git diff --check`：PASS。

关键文件 SHA-256：

- `services/market/etc/market.yaml`：
  `b9d14ad2f73c8263ad6e4a6baf91caedaf245e2911452329f5abab0a52ae2e6e`
- `deploy/seed-etcd.sh`：
  `dc9c1e9f75a6b35f6eb93144f7ad1474d7a1d20282b9d8cd8b3468fff5282f7e`
- `deploy/docker-compose.yml`：
  `1db7ec568010036597803ed2cdd391619304d83d5ae4eda8e10e24bd32af6d1d`

## 4. 运行验收

执行 `./deploy.sh config`：

- Docker 磁盘预检通过；
- `config-seed` 镜像构建成功；
- 17 个 Etcd 配置键全部写入成功；
- Etcd 中 iTick 配置不含占位符且 Token 字段非空；
- 重启 `market-rpc` 后 Healthcheck 为 `healthy`；
- 最近两分钟持续写入 `market-ws FINAL_QUOTE` 以及
  `price-engine INDEX/MARK/FUNDING`；
- 最近三分钟未出现 panic、fatal、鉴权失败或 Price Engine 执行失败日志。
- 完整只读生产门禁仍为 33 PASS / 14 FAIL / exit 1，两个生产风险开关保持关闭；
  14 个真实外部前置条件没有因本次凭据迁移被错误放行。

验收查询时，`market-ws FINAL_QUOTE` 最近两分钟 105 条，
INDEX/MARK/FUNDING 各 120 条，证明 Secret 注入后的实时行情和价格引擎链路持续工作。
