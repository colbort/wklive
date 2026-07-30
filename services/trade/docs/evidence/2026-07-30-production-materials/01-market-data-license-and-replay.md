# 行情许可与历史回放材料

## 已核实事实

- 当前来源：`binance-public`、`okx-public`、`bybit-public`；
- 三个 Authority 均为独立 `provider_code`，通过公开 REST 获取 `BTCUSDT`；
- 三源 INDEX、INDEX_BASIS MARK、FUNDING、DELIVERY 每秒运行；
- 当前 MARK 不可变版本为 `production-mark-v3`：INDEX 和独立永续报价是两个必需
  实时输入，上一 MARK 仅作为可选平滑状态；
- 2026-07-29 已完成 60 秒、240 条不可变审计和确定性回放；
- 公开端点不需要 API Key，但“不需要凭据”不等于“取得商业数据许可”。

## 2026-07-30 MARK 断流恢复补充验收

上游短暂停止超过 30 秒后，旧 v2 因上一 MARK 超出回看窗口而无法重新产生 MARK，
FUNDING 随之停止。已通过 Admin API 创建并激活不可变
`BTCUSDT-MARK-production-v3 / production-mark-v3`，未修改历史 v2 记录。

- v3 `algorithm=INDEX_BASIS`、`max_lookback_ms=30000`、
  `max_deviation_bps=200`、`interval_ms=1000`；
- 三个声明成分仍为 INDEX 权重 1、Binance 永续 FINAL_QUOTE 权重 1、上一 MARK
  权重 4；
- `min_input_count=2` 仅表示 INDEX 与独立永续报价两个必需实时输入；上一 MARK 是
  可选平滑状态，不参与实时法定输入门槛；
- v3 首条不可变审计的 `accepted_inputs=2` 且
  `smoothing_bootstrap=true`，输出等于当期未平滑标记价；
- 下一秒审计的 `accepted_inputs=3`，上一 MARK 权重 4，自动恢复 1:4 平滑；
- INDEX、MARK、FUNDING、DELIVERY 随后均恢复逐秒新鲜输出。

本补充只证明断流自恢复技术门禁，不替代 v3 正式历史回放审批，也不改变本文件所列
行情商业许可门禁。

## 官方条款核对

- [Binance Developer Documentation](https://developers.binance.com/en/docs/introduction)
  说明 API 可用于市场数据、自动化和内部服务，同时要求生产上线前确认产品具体要求；
  当前没有 Binance 针对本项目的书面商业许可或审批引用。
- [OKX API Agreement](https://www.okx.com/en-sg/help/okx-api-agreement)
  将公开端点数据纳入相同市场数据限制，并要求商业产品、价格源或商业利用取得书面授权。
- [Bybit API Terms](https://www.bybit.com/common-static/compliance/legal/BYBIT/df1923006718fbba8ba70d7d762b9866.pdf)
  禁止重新包装、转售和商业利用 API/Service Data，除非另有有效授权。

## 当前门禁

`PRICE_SOURCE_LICENSE_APPROVED=false` 必须保持不变。仅有公开 URL、API 文档或本报告
不能替代三家供应商书面许可及法务审批。

## 通过条件

1. 提供 Binance、OKX、Bybit 分别覆盖本项目用途的书面许可、合同或供应商批准编号；
2. 法务/行情负责人给出审批人、审批时间和不可变归档引用；
3. 把审批信息补入现有回放报告并重新计算 SHA-256；
4. 填写 `HISTORICAL_REPLAY_PRODUCTION_APPROVAL_REF`；
5. 才可将 `PRICE_SOURCE_LICENSE_APPROVED` 改为 `true`。

如果无法取得其中任一家授权，必须先替换为具备相应用途许可的独立来源，重新运行三源
窗口和四类公式回放，不能只删除本段风险说明。
