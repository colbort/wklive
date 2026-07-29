# 平台资金账户与默认保险配置补齐报告

## 1. 执行范围

- 环境：Deploy 独立完整环境；
- 租户：`tenant_id=1`；
- 结算资产：`USDT`；
- 执行日期：2026-07-29；
- 目标：通过既有管理接口补齐 `FEE_REVENUE` 平台账户和
  `symbol_id=0` 默认保险配置；
- 安全边界：不划转资金，不开启 ADL、自动强平或全仓新增风险。

## 2. 执行前事实

- `INSURANCE_FUND` 平台账户：存在，`id=1`、状态启用、可用余额为 0；
- `FEE_REVENUE` 平台账户：不存在；
- `tenant_id=1 / symbol_id=0 / USDT` 默认保险配置：不存在；
- `AutomaticLiquidation.Enabled=false`；
- `CrossMarginTrading.Enabled=false`。

首次通过 Admin API 设置时发现两个真实兼容问题：

1. Asset 管理接口只允许 `INSURANCE_FUND`，拒绝模型和业务已经使用的
   `FEE_REVENUE`；
2. 迁入旧库的 `t_contract_insurance_fund_account` 仍保留已从 Proto
   `reserved` 的 `fund_user_id`、`wallet_type` 必填列，导致当前模型插入失败。

## 3. 修复

- Asset 的平台账户设置、查询入口统一允许模型声明的
  `INSURANCE_FUND`、`FUNDING_DIFFERENCE`、`FEE_REVENUE`；
- 余额人工调整仍只允许 `INSURANCE_FUND`，未扩大资金操作权限；
- 新增 baseline-safe 兼容迁移
  `20260729_allow_platform_backed_insurance_config.sql`：
  - 保留旧字段和已有值；
  - 只为废弃字段增加兼容默认值；
  - 保险配置约束收敛到当前仍有效的 `adl_enabled`、`status`；
- 使用 `deploy.sh db-upgrade` 执行，最终迁移数为 49；
- Asset RPC 镜像：
  `sha256:e6658b82728b7ce0ecfeceee6d1f4b6c8fdeec3aae72d7a190ed47e82e7bd880`。

## 4. 管理接口执行结果

### 4.1 手续费收入账户

- 接口：`POST /admin/asset/platform-accounts`；
- 结果：成功；
- 事实：`id=2`、`tenant_id=1`、`account_type=FEE_REVENUE`、
  `coin=USDT`、`status=1`；
- 可用余额：0；
- 冻结余额：0。

### 4.2 默认保险配置

- 接口：`POST /admin/trade/insurance-fund/accounts`；
- 结果：成功；
- 事实：`id=1`、`tenant_id=1`、`symbol_id=0`、
  `settle_asset=USDT`、`status=1`；
- `adl_enabled=2`（关闭）。

设置后分别通过对应 GET 管理接口读回，字段与请求一致。

## 5. 验证

- `services/asset go test ./...`：通过；
- `services/trade go test ./...`：通过；
- `deploy/dbinit go test ./...`：通过；
- Asset RPC：Healthy；
- Admin API：Healthy；
- 只读 `contract-readiness`：
  - PASS：33；
  - FAIL：14；
  - 结论：`NOT READY`；
  - Snapshot Outbox、合约对账、Settlement 水位通过；
  - 两个生产风险开关仍为 false。

FAIL 数量未减少是因为当前声明仍缺真实三源、资金权限审批、值班升级、正式
RPO/RTO 等外部材料，且保险基金尚未注资、生产 DELIVERY 公式和交割合约尚未启用。
本次没有用本地测试值冒充生产审批。

## 6. 剩余动作

- 由资金负责人确认保险基金最低水位和注资金额后，通过
  `POST /admin/asset/platform-accounts/adjust` 使用唯一 `requestNo` 注资；
- 提供三个真实独立行情源后创建 INDEX、MARK、FUNDING、DELIVERY 新版本；
- 确认交割日期和发布窗口后配置并启用新的 BTCUSDT 交割合约；
- 完成生产告警、资金权限和灾备审批后重新运行只读门禁。
