# 保险基金与强平发布材料

## 已核实事实

- 租户：`tenant_id=1`；
- 结算币种：USDT；
- `INSURANCE_FUND`：账户 ID 1，启用，可用余额 0，冻结余额 0；
- `FEE_REVENUE`：账户 ID 2，启用，可用余额 0，冻结余额 0；
- 默认保险配置：`symbol_id=0/USDT`，启用，ADL 关闭；
- 自动强平和全仓新增风险开关均为 `false`；
- 管理入口 `POST /admin/asset/platform-accounts/adjust` 支持唯一
  `requestNo` 的幂等保险基金调账；
- 当前没有经资金/风控审批的最低水位或注资金额。

## 注资约束

未取得最低水位、资金来源、审批人和审批编号前，不执行调账。不得把任意 1 USDT 或测试
金额当作保险基金门禁通过证据。

正式请求必须包含：

```json
{
  "tenantId": 1,
  "accountType": "INSURANCE_FUND",
  "coin": "USDT",
  "requestNo": "由发布单生成的唯一编号",
  "direction": 1,
  "amount": "经审批的注资金额",
  "remark": "包含审批编号和资金来源"
}
```

## 通过条件

1. 风控批准 `INSURANCE_FUND_MIN_AVAILABLE`；
2. 资金负责人批准账户权限、资金来源、注资金额和操作人；
3. 发布负责人批准强平启用窗口及回滚方案；
4. 使用管理接口完成幂等注资，并核对平台流水和账户版本；
5. 重跑保险全额、部分承接、余额不足、冲正及 ADL 恢复验收；
6. 更新回滚方案及 SHA-256；
7. 填写 `FUND_ACCOUNT_PERMISSION_APPROVED=true`、`FUND_ACCOUNT_APPROVER`、
   `LIQUIDATION_ENABLE_WINDOW` 和
   `LIQUIDATION_ROLLBACK_PRODUCTION_APPROVAL_REF`。

