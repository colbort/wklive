# Authority Registry 可视化管理与生产候选参数报告

## 1. 配置决策

- 日期：2026-07-29；
- 用户确认按《永续与交割合约完成清单》的需求配置；
- 技术参数采用文档候选值；
- 未提供的真实供应商、数据许可、资金审批、值班责任人和灾备审批继续保持
  `false/空`；
- 交割产品和两个生产风险开关继续保持关闭。

已通过 Admin API 读回 `tenant=1`、Symbol 7：

- `BTCUSDT-20260925`，产品状态 Disabled；
- 线性交割合约，BTC/USDT，USDT 结算；
- 交割时点 2026-09-25 16:00 Asia/Hong_Kong；
- 15:00 停止开仓，15:30 停止撮合；
- 锁价窗口 30 秒；
- 结算公式版本 `delivery-v1`；
- 逐仓杠杆 `1/2/5/10/20/50/100`，默认 1；
- 开多/开空关闭，平多/平空保留。

发现预生产声明仍把交割 Symbol 写成 `BTCUSDT`、版本写成
`preprod-delivery-v1`，现已分别修正为 `BTCUSDT-20260925` 和
`delivery-v1`。

`deploy/production-readiness.env` 已按上述候选值生成。该文件由 Git 忽略且不保存
API Key；两个真实 iTick 通道、两项审批 false、资金/告警/灾备责任字段为空，因此
不能通过或打开生产门禁。

## 2. Admin UI

价格公式页面新增“行情来源管理”：

- 使用既有 `market:authority:list`、`market:authority:set` RBAC 权限；
- 列出 Authority、`provider_code`、`producer_type`、允许快照类型和状态；
- 支持新增来源；
- 支持修改允许类型和启停状态；
- 编辑时 Authority、Provider Code、Producer Type 禁用并提示创建后不可修改；
- 禁用前二次确认，服务端继续阻止停用被激活公式引用的来源；
- 保存后同时刷新管理列表和价格公式下拉；
- 页面仍只从 Enabled Registry 构建输出和成分下拉；
- 中英文文案已补齐。

没有增加删除入口，也没有绕过版本乐观锁或后端公式引用保护。

## 3. 验收

自动化与构建：

- `npm run type-check`：PASS；
- Prettier check：PASS；
- `npm run build`：PASS，2,070 个模块完成转换；
- Vite Node 20.20.2 运行态首页：HTTP 200；
- Vite `price-formulas.vue` 模块转换：HTTP 200；
- `git diff --check`：PASS。

生产构建产物：

- `price-formulas-Bb7g8TB8.js`：
  SHA-256 `6ebdcff2d23bc8d020c56079da6ec5dbc60ae2cf48a437f0825644dd31917f3f`；
- `price-formulas-CduPjIQy.css`：
  SHA-256 `72543122c3e0e9986d9da6d9a6fd2aa94b5f64b4805badaea20f4e396ffdcade`。

运行中 Admin API 全量 Authority 查询返回 200：

| Authority | Provider | Producer | 状态 |
| --- | --- | --- | --- |
| `market-ws` | `ITICK` | `ITICK_WS` | Enabled |
| `market-rest` | `ITICK` | `ITICK_REST` | Enabled |
| `price-engine` | `PRICE_ENGINE` | `PRICE_ENGINE` | Enabled |

默认生产候选声明复跑：

- 声明文件存在，候选 Symbol、窗口和版本检查通过；
- 核心服务、初始化器、双开关、模型检查、Outbox、对账和结算水位通过；
- 精确保持 `14 prerequisite(s) failed`、exit 1；
- 失败项仍为真实三供应商、许可/凭据审批、资金审批/注资、告警责任链和生产灾备；
- 未创建虚假来源、未注入资金、未启用交割产品或风险开关。

## 4. 结论

候选技术参数和后续来源配置入口已经就绪。真实供应商与审批资料到位后，可直接在
后台注册 Authority、创建三源公式并复跑门禁，不再需要手写 API 或 SQL。
