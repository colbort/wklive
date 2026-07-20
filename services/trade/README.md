币币交易、秒合约、U本位、币本位

交易对规则
下单校验
订单流水
撮合接入
用户交易权限
手续费
风控检查
账户变动通知

## Model 生成

执行 `make gen-model` 生成 MySQL model。项目根目录的 `goctl.yaml` 会把
`DECIMAL`、`DEC`、`FIXED`、`NUMERIC` 映射为 `decimal.Decimal`，可空列映射为
`decimal.NullDecimal`，禁止交易金额、价格、数量和费率退化为 `float64`。

该映射依赖 goctl v1.9.2 的 experimental model converter；Makefile 会校验版本。
