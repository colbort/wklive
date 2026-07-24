# liquidity-admin-ui

交易所内部做市和外部流动性管理前端。

## 本地运行

```bash
npm install
npm run dev
```

默认地址为 `http://localhost:5180`，请求前缀为 `/liquidity/admin`，开发代理目标为
`http://127.0.0.1:8890`。可通过 `.env.development` 修改。

## 功能入口

- 运行总览
- 流动性提供方（内部做市、外部渠道）
- 交易对策略（现货、交割、永续）
- 内部报价订单与外部路由订单
- 对冲任务
- 风险事件
- 外部对账

接口封装集中在 `src/api/liquidity.ts`。当前路径按 REST 管理 API 设计，
后续 `liquidity-admin-api` 应负责认证、权限、租户上下文以及 RPC 到 HTTP 的转换。
