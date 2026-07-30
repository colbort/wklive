# MARK v3 与四类价格公式确定性回放

## 结论

2026-07-30 16:24:19～16:25:18 HKT，从
`t_itick_authoritative_snapshot` 只读选取当前生产候选 INDEX、MARK、FUNDING、
DELIVERY 各 60 条不可变 `raw_payload`，直接输入
`services/market/cmd/price-replay`：

- 总记录数：240；
- 公式数：4；
- 每公式：60 条；
- 期望周期：1000 ms；
- 每个公式的目标时点严格连续，断档和重复时点均为 0；
- 确定性重算：PASS；
- 四个公式的被剔除输入总数均为 0。

本报告覆盖当前 `production-mark-v3`，替代 2026-07-29 v2 报告作为当前技术回放
材料；历史 v2 报告继续保留，不作修改。

## 运行版本与参数

| 快照类型 | 公式编号 | 公式版本 | 记录 | 最少接受输入 | 输出范围 |
| --- | --- | --- | ---: | ---: | --- |
| DELIVERY | `BTCUSDT-DELIVERY-production-v1` | `delivery-v1` | 60 | 3 | 64166.08 ～ 64207 |
| FUNDING | `BTCUSDT-FUNDING-production-v1` | `production-funding-v1` | 60 | 2 | -0.0006991575632142 ～ -0.0002479797785163 |
| INDEX | `BTCUSDT-INDEX-production-v1` | `production-index-v1` | 60 | 3 | 64166.08 ～ 64207 |
| MARK | `BTCUSDT-MARK-production-v3` | `production-mark-v3` | 60 | 3 | 64142.0972915120842482 ～ 64178.5608910776309712 |

该稳定窗口内 MARK 已进入正常平滑状态，因此最少接受输入为 3。断流恢复首条
`accepted_inputs=2`、`smoothing_bootstrap=true` 的运行证据另见
[`01-market-data-license-and-replay.md`](01-market-data-license-and-replay.md)。

## 机器可读回放结果

```json
{
  "record_count": 240,
  "formula_count": 4,
  "first_target_time": 1785399859000,
  "last_target_time": 1785399918000,
  "expected_interval_ms": 1000,
  "formulas": [
    {
      "formula_no": "BTCUSDT-DELIVERY-production-v1",
      "formula_version": "delivery-v1",
      "record_count": 60,
      "first_target_time": 1785399859000,
      "last_target_time": 1785399918000,
      "minimum_output_price": "64166.08",
      "maximum_output_price": "64207",
      "minimum_accepted_input_count": 3,
      "rejected_input_count": 0
    },
    {
      "formula_no": "BTCUSDT-FUNDING-production-v1",
      "formula_version": "production-funding-v1",
      "record_count": 60,
      "first_target_time": 1785399859000,
      "last_target_time": 1785399918000,
      "minimum_output_price": "-0.0006991575632142",
      "maximum_output_price": "-0.0002479797785163",
      "minimum_accepted_input_count": 2,
      "rejected_input_count": 0
    },
    {
      "formula_no": "BTCUSDT-INDEX-production-v1",
      "formula_version": "production-index-v1",
      "record_count": 60,
      "first_target_time": 1785399859000,
      "last_target_time": 1785399918000,
      "minimum_output_price": "64166.08",
      "maximum_output_price": "64207",
      "minimum_accepted_input_count": 3,
      "rejected_input_count": 0
    },
    {
      "formula_no": "BTCUSDT-MARK-production-v3",
      "formula_version": "production-mark-v3",
      "record_count": 60,
      "first_target_time": 1785399859000,
      "last_target_time": 1785399918000,
      "minimum_output_price": "64142.0972915120842482",
      "maximum_output_price": "64178.5608910776309712",
      "minimum_accepted_input_count": 3,
      "rejected_input_count": 0
    }
  ]
}
```

## 门禁边界

这是一份当前 Deploy 环境的技术回放事实，不是外部生产批准。以下字段在取得真实材料
前仍必须保持失败：

- 三家行情供应商商业使用许可；
- 回放执行人、复核人和生产审批归档引用；
- 风控对 200 BPS、1:4 平滑及公式版本的正式签批。

