# 合约生产职责账号

## 已创建账号

2026-07-30 已通过 `deploy.sh db-init` 在管理后台幂等创建六个系统账号，并完成真实登录
验证。账号属于 `tenant_id=0`、`app_scope=1`，密码使用 bcrypt 保存，初始明文只存在
Git 忽略且权限为 `0600` 的 `deploy/secrets/production-operators.env`。

| 系统账号 | 角色 | 用途 | 后台写权限 |
| --- | --- | --- | --- |
| `contract_oncall` | `contract_oncall` | 合约生产值班、三类 Admin WS 告警接收 | 无 |
| `insurance_operator` | `insurance_fund_operator` | 执行已审批的保险基金账户配置和幂等调账 | 保险基金 3 个入口 |
| `dr_operator` | `disaster_recovery_operator` | 灾备演练执行身份及审计 | 无 |
| `delivery_operator` | `delivery_release_operator` | 在批准窗口执行交割合约配置 | 合约交易对配置 1 个入口 |
| `production_reviewer` | `production_reviewer` | 复核行情、资金、交割、对账和审计事实 | 无 |
| `production_approver` | `production_approver` | 最终系统审批身份 | 无 |

数据库权限核验结果：

```text
contract_oncall       menu=14 write=0
insurance_operator    menu=10 write=3
dr_operator            menu=6  write=0
delivery_operator      menu=10 write=1
production_reviewer    menu=22 write=0
production_approver    menu=22 write=0
```

六个账号均已通过 `/admin/system/auth/login` 验证，返回 `code=200` 且签发令牌。核验过程
没有输出密码或令牌。

## Readiness 门禁验收

`contract-readiness` 已增加数据库模型校验，不再只判断声明字段是否非空：

- 逐账号核对启用状态、`tenant_id=0`、`app_scope=1` 和唯一角色绑定；
- 值班账号必须同时具有 Price Engine、Snapshot Outbox 和合约对账三类读取权限；
- 保险基金操作员的后台写权限必须严格等于菜单 1182、1185、1186；
- 交割操作员的后台写权限必须严格等于菜单 1015；
- 灾备、复核和审批账号的后台写权限必须为 0；
- 任一账号额外绑定其他角色也会失败。

正向终检六项全部 PASS。负向验收把
`CONTRACT_ONCALL_ACCOUNT` 临时从 `contract_oncall` 改为具有全量权限的 `admin`，
其余声明和数据库不变；门禁准确新增：

```text
FAIL  contract on-call account has its exact role and read-only alert permissions
NOT READY: 15 prerequisite(s) failed
```

声明恢复为 `contract_oncall` 后重新终检，账号门禁恢复 PASS，整体仍保持原有
14 个外部前置条件失败。

## 四组责任映射

| 生产材料 | 执行/值班 | 复核 | 审批 |
| --- | --- | --- | --- |
| 值班、升级与告警 | `contract_oncall` | `production_reviewer` | `production_approver` |
| 保险基金与强平 | `insurance_operator` | `production_reviewer` | `production_approver` |
| 生产灾备 | `dr_operator` | `production_reviewer` | `production_approver` |
| 交割合约启用 | `delivery_operator` | `production_reviewer` | `production_approver` |

告警候选升级链为：T+5 分钟 L1 `contract_oncall`、T+10 分钟 L2
`production_reviewer`、T+15 分钟 L3 `production_approver`。

## 仍需外部事实

系统账号解决的是鉴权、最小权限、操作日志和职责分离，不会自动产生以下客观事实：

- 保险基金最低水位、真实注资金额和资金来源；
- 获批 RPO/RTO、加密异地存储、KMS Key ID/别名和密钥托管；
- 交割启用时间窗口及正式发布单；
- 每个系统账号背后的真实人员、排班和不可变审批编号。

上述事实到齐前，保险基金不调账、交割合约不启用、自动强平和全仓开关继续保持关闭。
