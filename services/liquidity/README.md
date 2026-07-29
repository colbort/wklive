# liquidity.rpc

流动性服务，负责现货、永续合约和交割合约的内部做市、外部订单路由、库存风控、对冲与对账。

当前阶段仅包含：

- 可启动的 RPC/健康检查骨架；
- `trade.rpc`、`asset.rpc`、`market.rpc` 依赖配置；
- MySQL 8.0 数据库设计；
- SQL 设计说明。

`proto/liquidity` 已定义 Admin、Internal、Task 三组 RPC 契约。当前阶段不包含 RPC 业务实现、数据库 model、做市任务和外部交易所适配器。待 SQL 与 proto 审核通过后再生成服务端业务代码，避免模型反复变更。

## 启动配置

- etcd 服务配置键：`/wklive/liquidity-rpc/config`
- etcd 注册键：`liquidity.rpc`
- 默认端口：`8091`

## 数据单位

- 时间：Unix 毫秒；
- 价格、数量、名义金额：交易自然单位，不使用资产系统的分单位；
- 外部 API 密钥：只保存 `credential_ref`，不在数据库保存明文。
