# 永续与交割合约预生产价格回放报告

## 1. 环境与维度

- 环境：本机完整 Docker Compose 预生产验收环境
- 品类：`crypto`
- 市场：`BA`
- 价格 Symbol：`BTCUSDT`
- 租户：`tenant_id=1`
- 永续 Symbol：`BTCUSDT`（`symbol_id=2`）
- 交割 Symbol：`BTCUSDT`（`symbol_id=4`）
- 回放工具：`services/market/cmd/price-replay`
- 期望计算周期：1,000 ms

当前运行配置只有 `itick-ws` 单源参与 MARK/INDEX，不具备生产三独立来源条件；
`itick-rest` 已注册但未参与公式，也不存在 DELIVERY 运行公式。本报告将“真实运行窗口
确定性回放”和“三源/DELIVERY/INDEX_BASIS 参数测试”分开记录，不把测试来源冒充
生产独立市场源。

## 2. 真实运行窗口确定性回放

- 回放时间：2026-07-29 17:11:10 至 17:13:09 +08:00
- 原始审计记录：360
- 公式数量：3
- 相邻目标时点：严格连续，无重复、无断档
- 回放结果：PASS

| 公式 | 版本 | 记录数 | 最少有效输入 | 剔除数 | 输出范围 |
| --- | --- | ---: | ---: | ---: | --- |
| `BTCUSDT-INDEX-v1` | v1 | 120 | 1 | 0 | 64,446.1 ～ 64,461.42 |
| `BTCUSDT-MARK-v1` | v1 | 120 | 1 | 0 | 64,446.1 ～ 64,461.42 |
| `BTCUSDT-FUNDING-v1` | v1 | 120 | 2 | 0 | 0 ～ 0 |

证据：

- 原始审计：`price-audits.jsonl`
- 原始审计 SHA-256：
  `414fc52677785dc6928efc9e3adc2d272d52201d60722b9f584d3f4df4524471`
- JSON 回放报告：`price-replay.json`
- JSON 回放报告 SHA-256：
  `54475dde9a1dbbb6db20dad7717fc4c8a9c17535b8068905f6847f9189c57fde`

## 3. 生产参数候选技术测试

执行 `internal/priceengine` 与 `internal/logic/task` 的定向测试，结果全部 PASS：

- 加权平均：`100×1 + 110×3` 得到 `107.5`；
- 中位数：`3,1,2` 得到 `2`；
- DELIVERY 偏离剔除保留完整输入审计；
- DELIVERY 剔除后少于三个有效输入时禁止发布；
- DELIVERY Snapshot ID 由不可变审计确定性生成；
- INDEX/DELIVERY 重复 Snapshot ID 不计为独立来源；
- INDEX 公式必须使用三个不同市场来源；
- `INDEX_BASIS` MARK 对正负基差执行对称 BPS 限幅；
- MARK 基差 10%、限幅 200 BPS 时输出 102；
- MARK 以前值权重 4 平滑后输出 100.4；
- 无风险限幅的 `INDEX_BASIS` 配置被拒绝；
- 回放能拒绝输出篡改、输入分区篡改、重复目标时点和时间断档；
- DELIVERY 新公式最少有效输入数强制不低于 3。

测试日志：

- 路径：`price-parameter-tests.txt`
- SHA-256：
  `4e97d845fa82825e5ec8707851627a25ec2187e1256fc96a6fa82ba9ac093e99`

## 4. 参数建议

以下仅作为进入真实来源回放前的预生产候选，不是生产审批值：

| 参数 | 候选值 | 说明 |
| --- | --- | --- |
| INDEX 算法 | MEDIAN | 至少三个真实独立来源 |
| INDEX 最少输入 | 3 | 剔除后仍必须满足 |
| MARK 算法 | INDEX_BASIS | 指数价加永续基差 |
| MARK 基差限幅 | 200 BPS | 必须结合生产历史分布复核 |
| MARK 平滑 | 当前值 1、前值 4 | 仅技术测试通过 |
| DELIVERY 算法 | MEDIAN | 交割窗口内三源以上 |
| DELIVERY 最少输入 | 3 | 不允许单源兜底 |
| DELIVERY 回看窗口 | 30,000 ms | 必须与合约锁价窗口一致 |
| 公式周期 | 1,000 ms | 回放工具已验证连续性 |

## 5. 结论

- 真实单源运行窗口的确定性回放：PASS；
- 三源约束、加权/中位数、偏离剔除、INDEX_BASIS、平滑与篡改检测技术测试：PASS；
- 真实三独立来源运行配置：未完成；
- DELIVERY 实际运行窗口回放：未完成；
- 生产数据许可、凭据接入和参数审批：未完成。

因此本报告支持“预生产技术链路通过”，不能支持“生产 Price Engine 门禁通过”。

- 执行人：Codex（按用户授权执行）
- 复核人：待项目负责人签署
- 生产审批编号：待定
