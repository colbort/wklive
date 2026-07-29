# 三独立交易所价格公式运行与回放报告

## 1. 结论

2026-07-29 已在完整 Deploy 环境直接接入并运行三个独立现货交易所公开行情：

| Authority | provider_code | 市场 | 官方公开接口 | 用途 |
| --- | --- | --- | --- | --- |
| `binance-public` | `BINANCE` | `BINANCE` | `https://api.binance.com/api/v3/aggTrades` | INDEX、DELIVERY |
| `okx-public` | `OKX` | `OKX` | `https://www.okx.com/api/v5/market/ticker` | INDEX、DELIVERY |
| `bybit-public` | `BYBIT` | `BYBIT` | `https://api.bybit.com/v5/market/recent-trade` | INDEX、DELIVERY |
| `binance-futures-public` | `BINANCE` | `BINANCE_PERP` | `https://fapi.binance.com/fapi/v1/aggTrades` | MARK 永续基差 |

三个现货来源的 `provider_code` 互不相同；Binance 现货和永续正确归为同一
`BINANCE` 供应商，不能重复计数。所有接口为无需 API Key 的公开市场数据端点，
配置不保存凭据。公开可访问不等于数据许可已获得生产批准，因此
`PRICE_SOURCE_LICENSE_APPROVED` 继续保持 `false`。

## 2. 接入边界

- 仅允许预定义的 HTTPS 官方域名和固定路径，不接受自定义主机、用户信息、Query 或
  Fragment；
- 每个响应校验交易对、正价格、源时间、新鲜度和未来时钟偏差；
- 默认每秒轮询，单请求超时 3 秒，源时间最大年龄 30 秒；
- 同一上游成交重复返回时由不可变 Snapshot ID 和模型事务幂等去重；
- 错误首次输出，未恢复时每 30 秒提醒，恢复时输出恢复事件；
- Authority 注册通过 migration 52 落库，业务运行中的数据库操作仍全部由 models
  执行。

## 3. 不可变公式

全部公式通过 Admin API 创建并激活，没有直接修改公式表：

| 输出 | 公式 | 版本 | 算法与参数 |
| --- | --- | --- | --- |
| INDEX | `BTCUSDT-INDEX-production-v1` | `production-index-v1` | 三源 MEDIAN，最少 3，偏差 200 BPS |
| MARK | `BTCUSDT-MARK-production-v2` | `production-mark-v2` | INDEX_BASIS，Binance 永续，±200 BPS，当前值 1/前值 4 |
| FUNDING | `BTCUSDT-FUNDING-production-v1` | `production-funding-v1` | `(MARK-INDEX)/INDEX` |
| DELIVERY | `BTCUSDT-DELIVERY-production-v1` | `delivery-v1` | 三源 MEDIAN，最少 3，偏差 200 BPS，窗口 30 秒 |

旧单源 v1 公式和首次平滑/引导版本均保持不可变历史状态，当前每种输出只有上述一个
激活版本。

## 4. 运行中发现并修复的问题

### 4.1 Binance REST 响应差异

Binance REST `aggTrades` 实际响应可不带 WebSocket 消息中的 `s` 字段。解析器现采用
“字段存在则校验，字段缺省则由已白名单化的请求参数绑定交易对”，并保留错误交易对
负向测试。

### 4.2 MARK 平滑新鲜度递归老化

首次三输入 MARK 在运行约 30 秒后停止。根因是上一 MARK 仅用于平滑，却被递归继承为
新 MARK 的源时间，最终必然越过 30 秒回看窗口。修复后：

- 当前 INDEX 和永续报价决定新 MARK 的 `source_timestamp`；
- 上一 MARK 只参与平滑数值；
- 回归测试显式使用 20 秒前的上一 MARK，验证新输出源时间仍取当前市场输入；
- 运行环境先创建两输入引导版本，再切换到不可变平滑 v2。

修复后四个激活公式连续运行超过原 30 秒边界，均保持目标时点约 1 秒延迟。

## 5. 60 秒确定性回放

- 窗口：2026-07-29 19:53:38 ～ 19:54:37 HKT；
- 周期：1,000 ms；
- 总记录：240；
- 公式：4；
- 每公式：严格连续 60 条；
- 重复目标时点：0；
- 断档：0；
- 回放结果：PASS。

| 公式 | 记录 | 最少接受输入 | 剔除 | 输出范围 |
| --- | ---: | ---: | ---: | --- |
| DELIVERY | 60 | 3 | 0 | 64462 ～ 64465.23 |
| FUNDING | 60 | 2 | 0 | -0.000514179769193 ～ -0.0004752110655716 |
| INDEX | 60 | 3 | 0 | 64462 ～ 64465.23 |
| MARK | 60 | 3 | 0 | 64431.366944291126704 ～ 64434.4715594267245199 |

证据：

- `three-source-price-audits.jsonl`：
  `810c8e46f1f3a5c7c3e392f1faa63674364932d3b40a681142fe821dc238ec48`
- `three-source-price-replay.json`：
  `5fe39796e8244e7f00adc439c1e78629933e52be3806eda77ed2032415bee143`

## 6. 构建与运行

- iTick 全量 Go 测试：PASS；
- migration 数：52；
- 运行镜像：`sha256:9c7a0f85980ebc5f44a46280eefe0a4b4f48d8d2a44e3352bba9a3ed9b6062b0`；
- 运行二进制 SHA-256：
  `8a8421963e540cec7f1eb4e86bc779dc26d28f24d91ef1a291788b9d6d9c2b82`；
- iTick Healthcheck：Healthy；
- 四个来源及四种公式输出持续新鲜；
- Snapshot Outbox、对账和未完成结算继续由只读门禁检查。

标准 Docker 构建首次成功；修复版重建时 Go 模块代理对两个既有依赖返回
`unexpected EOF`。为避免无限等待，使用已运行的标准 iTick 镜像为基础，仅覆盖本机
交叉编译且测试通过的静态 Linux 二进制。源码和运行二进制对应；网络恢复后可重新执行
标准 `deploy.sh up market-rpc` 得到同等镜像。

## 7. 尚未通过的生产事实

- 三家交易所数据许可及法务审批；
- 公开端点无需凭据，但仍需安全负责人确认生产接入方式；
- 公式的 200 BPS、1:4 平滑和 MEDIAN 参数仍需风控签批；
- 保险基金注资、生产告警责任人、正式 RPO/RTO 和异地灾备；
- 交割产品继续保持 Disabled，双风险开关继续保持 `false`。

本报告证明真实三独立交易所运行链路与确定性回放技术通过，不替代上述生产审批。
