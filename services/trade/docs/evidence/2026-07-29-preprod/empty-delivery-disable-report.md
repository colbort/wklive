# 空壳交割产品停用报告

## 1. 原因

预生产终检发现 `tenant_id=1`、`symbol_id=4` 的 `BTCUSDT` 交割产品状态为启用，
但没有对应的 `t_trade_symbol_contract` 参数事实。只读核对结果：

- Order：0；
- Position：0；
- Delivery Batch：0；
- Reservation：0；
- Settlement Instruction：0；
- 合约配置行：0。

交割时间、锁价来源、算法、窗口和公式版本均没有权威来源，不能凭猜测补齐。

## 2. 变更

通过现有 Admin API `/admin/trade/symbols/update` 执行停用：

1. 第一次请求完整保留旧字段并设置 `status=2`；
2. Admin API 因旧数据 `tradingStartTime == listingTime`，以
   `开始交易时间必须晚于上线时间` 拒绝，数据库未发生状态变更；
3. 第二次请求仅把 `tradingStartTime` 从 `1784736000` 规范化为
   `1784736001`，其余交易参数保持不变；
4. API 返回成功，重新读取详情确认 `status=2`；
5. Remark 追加 `[disabled: missing contract configuration]`。

## 3. 结果

- 产品状态：Disabled；
- Order：0；
- Position：0；
- 合约配置行：0；
- 未创建价格、交割批次或资金事实；
- 重新启用前必须先补齐完整合约参数和经过验收的 DELIVERY 公式。

## 4. 证据

- `before.json`：变更前 Admin API 详情；
- `request.json`：成功停用请求；
- `response-invalid-legacy-time.json`：第一次请求的现行校验拒绝；
- `response.json`：成功响应；
- `after.json`：变更后 Admin API 详情。

- 执行人：Codex（按用户授权执行）
- 环境：当前完整 Deploy 预生产环境
- 日期：2026-07-29
