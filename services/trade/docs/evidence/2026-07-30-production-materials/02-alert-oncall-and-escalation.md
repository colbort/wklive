# 值班、升级与告警演练材料

## 已核实事实

- 生产候选通道为 Kafka `admin.notifications` 和 Admin WebSocket；
- Snapshot Outbox、Price Engine、合约对账三类 firing/resolved 已真实投递；
- 系统管理员 `sys_user.id=1/admin`、租户所有者 `sys_user.id=2/test` 均启用；
- WebSocket 已验证鉴权、Origin、权限及租户隔离；
- 当前没有真实排班、人工确认回执或未确认升级执行记录；
- 外部 Webhook 按既定决定不作为当前方案。

## 候选升级规则（尚未批准）

1. T+0：向当前值班管理员 WebSocket 投递 firing；
2. T+5 分钟未确认：同事件升级为高优先级并再次通知；
3. T+10 分钟未确认：升级至系统超级管理员和目标租户所有者；
4. T+15 分钟未确认：保持自动强平、全仓和交割合约发布门禁关闭；
5. 恢复后投递 resolved，并把确认人、确认时间和处置结果归档。

以上是待审批候选值，不能直接填写为已经执行的生产事实。当前代码已具备通知和去重，
但未形成持久化人工确认及升级回执。

## 通过条件

1. 明确真实值班团队、成员、排班入口和最终负责人；
2. 批准确认超时和二/三级升级规则；
3. 在获批窗口分别触发 Outbox、价格缺源和对账差异；
4. 保存平台接收、值班确认、升级及恢复回执；
5. 更新告警报告及 SHA-256；
6. 填写 `ALERT_ONCALL_TEAM`、`ALERT_ESCALATION_POLICY` 和
   `ALERT_TEST_PRODUCTION_APPROVAL_REF`。

