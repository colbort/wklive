# 永续与交割合约生产上线行动单

## 1. 当前结论

2026-07-30 当前完整 Deploy 环境已经完成预生产技术验收。只读
`contract-readiness` 最终结果：

- PASS：60；
- FAIL：17；
- 结论：`NOT READY`；
- `AutomaticLiquidation.Enabled=false`；
- `CrossMarginTrading.Enabled=false`。

本行动单把 17 个失败项转换为外部负责人可直接交付的材料和配置。任何 API Key、
密码、访问令牌或数据库凭据不得写入本文或 `production-readiness.env`。

预生产证据位于
[`evidence/2026-07-29-preprod`](evidence/2026-07-29-preprod)，完整哈希见
[`SHA256SUMS`](evidence/2026-07-29-preprod/SHA256SUMS)。

2026-07-30 按五组整理的当前事实、官方许可核对和执行前置条件见
[`evidence/2026-07-30-production-materials`](evidence/2026-07-30-production-materials)。

## 2. 17 个失败项

| # | 当前 FAIL | 负责人 | 必须交付 | readiness 字段或运行事实 |
| ---: | --- | --- | --- | --- |
| 1 | 行情数据许可未批准 | 法务/行情负责人 | Binance、OKX、Bybit 的数据使用许可或合同编号 | `PRICE_SOURCE_LICENSE_APPROVED=true` |
| 2 | 历史回放缺少生产审批引用 | 行情/风控负责人 | 执行、复核、审批完整的发布单或工单引用 | `HISTORICAL_REPLAY_PRODUCTION_APPROVAL_REF` |
| 3 | 生产值班组为空 | 运维负责人 | 值班组名称、最终责任人和排班入口 | `ALERT_ONCALL_TEAM` |
| 4 | 告警升级策略为空 | 运维负责人 | 一级渠道、未确认超时、二/三级升级规则 | `ALERT_ESCALATION_POLICY` |
| 5 | 告警演练缺少生产审批引用 | 运维负责人 | 含平台接收、通知、升级、恢复回执的审批归档 | `ALERT_TEST_PRODUCTION_APPROVAL_REF` |
| 6 | 保险基金审批水位未声明 | 资金/风控负责人 | 经审批的最低可用余额及审批编号 | `INSURANCE_FUND_MIN_AVAILABLE` 为正数 `DECIMAL(36,18)` |
| 7 | 保险和手续费账户权限未批准 | 资金/风控负责人 | 账户权限、最低水位和操作范围审批 | `FUND_ACCOUNT_PERMISSION_APPROVED=true` |
| 8 | 资金账户审批人为空 | 资金负责人 | 保险基金与手续费账户最终审批人 | `FUND_ACCOUNT_APPROVER` |
| 9 | 自动强平启用窗口为空 | 发布/风控负责人 | 发布单、启用窗口、操作人、复核人 | `LIQUIDATION_ENABLE_WINDOW` |
| 10 | 强平回滚方案缺少生产审批引用 | 发布/风控负责人 | 完成操作、复核、资金和风险签批的发布单引用 | `LIQUIDATION_ROLLBACK_PRODUCTION_APPROVAL_REF` |
| 11 | 正式 RPO 未批准 | 基础设施负责人 | 正整数分钟及审批编号 | `DR_RPO_MINUTES` |
| 12 | 正式 RTO 未批准 | 基础设施负责人 | 正整数分钟及审批编号 | `DR_RTO_MINUTES` |
| 13 | 备份加密未声明 | 安全/基础设施负责人 | 算法、密钥托管方式、轮换规则 | `DR_BACKUP_ENCRYPTION` |
| 14 | 异地位置未声明 | 基础设施负责人 | 异地 Region/机房/对象存储位置和保留周期 | `DR_OFFSITE_LOCATION` |
| 15 | 灾备演练缺少生产审批引用 | 基础设施负责人 | PITR、切换、回切和事实核对的正式审批归档 | `DR_EXERCISE_PRODUCTION_APPROVAL_REF` |
| 16 | 保险基金未注资 | 资金/风控负责人 | 经审批的最低水位和实际入账证据 | 数据库模型核对启用且余额达到审批水位的 `INSURANCE_FUND` |
| 17 | 未来交割合约未启用 | 发布/风控负责人 | 在其余门禁全部通过后的独立启用审批 | 数据库模型核对 `BTCUSDT-20260925` 为启用且交割时点仍在未来 |

## 3. 已完成的技术配置与最终开闸条件

### 3.1 Authority Registry 与实时来源

- Authority 已具备 `GET/POST /admin/market/authorities` 管理入口，价格公式页面从
  注册表动态加载，不再使用硬编码列表；
- 至少三个 Authority 状态启用；
- 每个 Authority 的 `allowed_kinds` 包含 `FINAL_QUOTE`；
- 每个 Authority 必须填写稳定的独立供应商标识 `provider_code`；同一供应商的
  WebSocket、REST 或其他传输通道必须填写相同值；
- 三个 Authority 的 `provider_code` 必须互不相同，不能把同一供应商的不同通道当成
  独立市场来源；公式创建和生产就绪模型都会拒绝这种配置；
- 三个来源在声明的 category、market、symbol 上持续产生新鲜快照；
- `price-engine` 必须允许发布 INDEX、MARK、FUNDING、DELIVERY。

当前已接入 `binance-public/BINANCE`、`okx-public/OKX`、
`bybit-public/BYBIT` 三个独立公开现货来源，以及用于 MARK 的
`binance-futures-public/BINANCE`。这些端点无需 API 凭据；
`PRICE_SOURCE_ACCESS_MODE=PUBLIC_NO_CREDENTIALS` 只有在数据库模型确认所有声明
Authority 均为启用的 `PUBLIC_REST` 时才通过。公开访问不替代第 1 项数据许可审批。

### 3.2 Price Engine

必须通过管理入口创建不可变的新版本，不直接修改历史公式：

- INDEX：MEDIAN 或 WEIGHTED_MEAN，组件集合与声明的 Authority 集合完全一致，
  `min_input_count >= 3`；
- MARK：`INDEX_BASIS`，使用 INDEX 与永续市场价，明确正负基差限幅和平滑权重；
- FUNDING：使用验收后的 MARK 和 INDEX；
- DELIVERY：MEDIAN 或 WEIGHTED_MEAN，至少三个 FINAL_QUOTE 来源，算法、权重、
  最大偏差、回看窗口和版本与声明完全一致；
- INDEX/MARK/FUNDING/DELIVERY 四类输出均在回看窗口内持续新鲜；
- 历史回放覆盖至少一个完整交割锁价窗口。

当前四个生产候选公式已经激活：三源 MEDIAN INDEX、INDEX_BASIS MARK v3、FUNDING
和三源 MEDIAN DELIVERY `delivery-v1`。60 秒窗口共 240 条不可变审计，四类公式
各 60 条、1 秒严格连续、断档为 0，确定性回放通过。readiness model 还会逐项核对
来源市场映射、INDEX 算法/版本/权重、MARK 永续来源/基差/1:4 平滑、FUNDING
版本，以及四类公式的窗口和执行周期，参数漂移不能通过终检。

### 3.3 永续和交割产品

- 保留一个启用且参数完整的生产永续产品；
- 原 `symbol_id=4` 空壳交割产品已经安全停用，不得直接重新启用；
- 已通过 Admin API 创建 `symbol_id=7 / BTCUSDT-20260925` 停用态新产品，并补齐
  合约、逐仓杠杆和风险档位；当前采用 2026-09-25 16:00
  Asia/Hong_Kong 交割假设，生产发布单须复核后方可启用；
- `delivery_time` 必须在未来；
- `settlement_window_seconds * 1000 == DELIVERY_LOCK_WINDOW_MS`；
- `settlement_price_algorithm == DELIVERY_FORMULA_VERSION`；
- settlement price source 指向经过验收的 DELIVERY 快照；
- 上线前确认 Order、Position、Reservation 和未完成 Instruction 的边界状态。

### 3.4 保险、手续费与默认配置

- 当前预生产事实：`tenant=1/USDT` 的 `INSURANCE_FUND` 和
  `FEE_REVENUE` 平台账户均已启用，`symbol_id=0/USDT` 默认保险配置已启用且
  ADL 关闭；两个平台账户可用余额均为 0，因此这只是结构配置完成，不代表资金门禁
  通过；
- 同租户、同结算币种存在启用的 `INSURANCE_FUND` 平台账户；
- 保险基金可用余额大于经审批的最低水位；
- 在 readiness 声明中填写同一审批水位
  `INSURANCE_FUND_MIN_AVAILABLE`，终检按数据库实际余额逐币种核对；
- 存在启用的 `FEE_REVENUE` 平台账户；
- 存在 `symbol_id=0`、同结算币种的默认保险配置；
- 明确 `fund_user_id`、钱包类型、ADL 是否启用及负责人；
- 通过 Asset/Trade 现有管理或模型入口配置，禁止在业务 Logic 或临时 shell 中直接
  写资金 SQL；
- 配置后再次验证保险全额、部分承接、余额不足、冲正和 ADL 恢复。

## 4. 必须提交的四份生产证据

| 证据 | 最低内容 | readiness 字段 |
| --- | --- | --- |
| 历史价格回放报告 | 三源映射、完整交割窗口、四类公式、断档/剔除/延迟、原始 JSON 输出、执行/复核/审批 | `HISTORICAL_REPLAY_REPORT`、`*_SHA256`、`HISTORICAL_REPLAY_PRODUCTION_APPROVAL_REF` |
| 告警投递测试报告 | Outbox、Price Engine 缺源、对账差异三类事件；平台接收、通知、升级和恢复回执 | `ALERT_TEST_REPORT`、`*_SHA256`、`ALERT_TEST_PRODUCTION_APPROVAL_REF` |
| 强平启用与回滚方案 | 账户、水位、启用顺序、观察项、自动阈值、人工回滚、执行/复核/审批 | `LIQUIDATION_ROLLBACK_PLAN`、`*_SHA256`、`LIQUIDATION_ROLLBACK_PRODUCTION_APPROVAL_REF` |
| 生产灾备报告 | 异地备份、GTID/Binlog、PITR、可用区切换/回切、事实表核对、RPO/RTO | `DR_EXERCISE_REPORT`、`*_SHA256`、`DR_EXERCISE_PRODUCTION_APPROVAL_REF` |

详细格式见
[`perpetual-delivery-production-evidence-guide.md`](perpetual-delivery-production-evidence-guide.md)。

## 5. 材料到齐后的执行顺序

1. 复制 `deploy/production-readiness.env.example` 为 Git 忽略的
   `deploy/production-readiness.env`；
2. 填写审批结论、业务维度、证据绝对路径和 SHA-256，不填写密钥；
3. 通过 Authority Registry 管理入口注册来源，再配置公式、产品、平台账户和默认
   保险配置；不得为通过门禁创建虚假来源；
4. 保持两个生产风险开关为 `false`；
5. 执行完整生产历史回放和三类告警投递测试；
6. 执行生产 PITR、可用区切换和回切演练；
7. 运行：

   ```bash
   cd deploy
   ./deploy.sh contract-readiness
   ```

8. 输出必须无 `FAIL`、最终为 `READY`、退出码为 0；
9. 将声明、四份证据、SHA-256 和终检输出随发布单归档；
10. `READY` 也不自动开闸，仍按批准窗口先开启自动强平，观察通过后再单独审批全仓。

## 6. 责任人回填

| 角色 | 姓名/团队 | 审批或工单编号 | 状态 |
| --- | --- | --- | --- |
| 行情负责人 |  |  | 未提供 |
| 风控负责人 |  |  | 未提供 |
| 资金负责人 |  |  | 未提供 |
| 安全负责人 |  |  | 未提供 |
| 运维值班负责人 |  |  | 未提供 |
| 基础设施负责人 |  |  | 未提供 |
| 发布负责人 |  |  | 未提供 |
