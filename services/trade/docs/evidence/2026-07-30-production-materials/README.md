# 永续与交割合约五组生产材料

## 当前结论

本目录按 2026-07-30 当前 Deploy 环境的真实事实整理，不把模板、建议值或本地测试
结果标记为生产审批。

| 组 | 材料 | 当前结论 |
| --- | --- | --- |
| 1 | 行情许可与历史回放 | 三源历史回放及 MARK v3 断流自恢复通过；缺供应商商业使用书面授权、v3 正式回放审批 |
| 2 | 值班、升级与告警演练 | 三类事件技术验收通过，值班/复核/审批系统账号已创建；缺账号背后的真实排班及正式审批 |
| 3 | 保险基金与强平发布 | 操作/复核/审批账号已创建；审批水位为 0、账户余额为 0，未注资 |
| 4 | 生产灾备 | 职责账号已创建，当前完整业务库不上传加密往返、144 表隔离恢复及真实 Binlog 精确停止位点全库 PITR、CMS AES-256-GCM 烟测、后台配置读取和专用私有版本化桶通过；用户本次决定不上传，仍缺上传/远端回读、KMS、异地实例切换/回切和正式 RPO/RTO |
| 5 | 交割合约启用 | 操作/复核/审批账号已创建；`BTCUSDT-20260925` 停用态独立技术预检通过，8 类交易/结算事实均为 0，仍未获准启用 |

五组材料只有在各文件“通过条件”全部取得客观证据后，才能填写
`deploy/production-readiness.env` 的生产审批字段。

2026-07-30 激活 `production-mark-v3` 后重新执行完整只读
`contract-readiness`：75 PASS、14 FAIL，运行服务、四类行情公式、实时输出、永续产品、
Outbox、对账和 Settlement 均通过；14 项仍全部属于外部审批/生产演练、实际注资以及
最终交割启用门禁。

六个职责账号现由数据库 readiness model 核对启用状态、唯一角色和写权限上限；将值班
账号错填为超级管理员的负向终检准确从 14 FAIL 增至 15 FAIL，恢复后回到 14 FAIL。

## 文件

- [01-market-data-license-and-replay.md](01-market-data-license-and-replay.md)
- [mark-v3-price-replay-report.md](mark-v3-price-replay-report.md)
- [02-alert-oncall-and-escalation.md](02-alert-oncall-and-escalation.md)
- [03-fund-and-liquidation-release.md](03-fund-and-liquidation-release.md)
- [04-disaster-recovery.md](04-disaster-recovery.md)
- [05-delivery-enablement.md](05-delivery-enablement.md)
- [production-responsibility-accounts.md](production-responsibility-accounts.md)：六个职责账号、权限核验和四组责任映射
- [encrypted-backup-smoke-report.md](encrypted-backup-smoke-report.md)：加密备份、回读、恢复、篡改拒绝和本地目标负向验收
- [full-database-local-encryption-report.md](full-database-local-encryption-report.md)：当前完整业务库不上传的导出、压缩、加密、解密哈希和篡改拒绝验收
- [full-database-local-restore-report.md](full-database-local-restore-report.md)：当前完整业务库恢复到无网络临时 MySQL 并逐表计数、检查和清理验收
- [full-database-local-pitr-restore-report.md](full-database-local-pitr-restore-report.md)：当前完整业务库加密快照、真实 Binlog 尾段、精确停止位点和越界拒绝验收
- [object-storage-readiness-report.md](object-storage-readiness-report.md)：系统配置读取、专用私有版本化桶、临时证书及未外发边界
- [delivery-preflight-report.md](delivery-preflight-report.md)：交割产品停用态配置、行情、风险档位和零历史事实只读验收
- [production-approval-input.env.example](production-approval-input.env.example)：五组负责人统一回填表
- [SHA256SUMS](SHA256SUMS)：材料完整性校验

## 下一步

复制统一回填表，由法务/行情、运维、资金/风控、基础设施和发布负责人分别填写真实
审批值。当前可以继续自动执行的只有只读检查；以下操作必须等待对应真实材料：

- 未取得三家行情供应商书面授权，不得将许可门禁改为通过；
- 未给出批准水位和资金来源，不得向保险基金调账；
- 未完成生产灾备演练，不得把预生产恢复报告标记为生产报告；
- 前四组没有全部通过，不得启用 `BTCUSDT-20260925`。
