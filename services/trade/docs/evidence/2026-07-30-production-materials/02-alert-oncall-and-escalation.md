# 值班、升级与告警演练材料

## 已核实事实

- 生产候选通道为 Kafka `admin.notifications` 和 Admin WebSocket；
- Snapshot Outbox、Price Engine、合约对账三类 firing/resolved 已真实投递；
- 系统管理员 `sys_user.id=1/admin`、租户所有者 `sys_user.id=2/test` 均启用；
- WebSocket 已验证鉴权、Origin、固定子协议、权限及租户隔离；
- 告警事件已持久化到 `sys_admin_notification_incident`，支持 firing、确认、恢复、
  重开、未确认升级和发送失败回退；
- 2026-07-30 已以真实系统管理员完成三类事件的 L1 未确认升级、人工确认回执和恢复；
- 外部 Webhook 按既定决定不作为当前方案。

## 运行态验收

验收程序：`admin-api/cmd/admin-notification-acceptance`。程序使用实际后台登录、Kafka
Topic、System RPC、MySQL 和同一个 Admin WebSocket；不调用测试桩。为避免等待五分钟，
验收事件的业务发生时间统一回拨十分钟，使其在落库时已经超过确认期限；表中“发布”和
“观察”时间均为验收程序的真实本地时钟。

| 事件 | Alert Key | 发布→首次 WS | 发布→L1 | 确认人 | 恢复 |
| --- | --- | ---: | ---: | --- | --- |
| Snapshot Outbox | `acceptance-admin-ws-1785398785308` | 1,014 ms | 10,152 ms | `1/admin` | 已收到 |
| Price Engine 缺源 | `acceptance-admin-ws-1785398805467` | 1,009 ms | 4,985 ms | `1/admin` | 已收到 |
| 合约对账差异 | `acceptance-admin-ws-1785398820403` | 1,018 ms | 5,049 ms | `1/admin` | 已收到 |

三条结果共同满足：

- `missingOriginRejected=true`；
- `queryTokenRejected=true`；
- `fixedSubprotocolSelected=true`；
- `escalationLevel=1`；
- 确认回执 `ack_result.ok=true`；
- 数据库最终 `status=3`（Resolved），并保留 `escalation_level=1`、
  `acknowledged_by=1`、确认原因和恢复时间；
- Admin UI 通过同一个 WebSocket 的“确认收到”按钮发送确认，不依赖 Webhook。

部署配置当前为：首次确认超时 5 分钟、未确认每 5 分钟升级一次、最高 L3、扫描周期
15 秒。三类运行态验收只证明 L1 技术闭环；L2/L3 的真实人员升级目标仍须随排班审批
确定。

## 候选升级规则（技术配置已落地，业务尚未批准）

1. T+0：向当前值班管理员 WebSocket 投递 firing；
2. T+5 分钟未确认：同事件升级为 L1 高优先级并再次通知；
3. T+10 分钟未确认：同事件升级为 L2；
4. T+15 分钟未确认：同事件升级为 L3，并保持自动强平、全仓和交割合约发布门禁关闭；
5. 恢复后投递 resolved，并把确认人、确认时间和处置结果归档。

以上人员责任和时限仍是待审批候选值，不能直接填写为生产值班事实。技术实现与 L1
运行态验收已完成，但没有真实排班就不能证明 7×24 人员响应。

## 通过条件

1. 明确真实值班团队、成员、排班入口和最终负责人；
2. 批准确认超时和二/三级升级规则；
3. 在获批窗口由当班人员复核现有三类技术演练，补录人员响应和 L2/L3 责任目标；
4. 保存正式值班确认、升级及恢复回执；
5. 更新本报告及 SHA-256；
6. 填写 `ALERT_ONCALL_TEAM`、`ALERT_ESCALATION_POLICY` 和
   `ALERT_TEST_PRODUCTION_APPROVAL_REF`。
