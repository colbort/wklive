# OPT-P1-004 仓库验收记录

验收日期：2026-08-02  
环境：当前工作区、隔离 MySQL 8.4、独立 Redis、当前工作区构建的真实 Asset gRPC  
结论：仓库实现与真实 RPC 门禁通过；生产代表性回测、独立模型验证签字和预生产复演仍未完成。

## 实现范围

- 组合风险参数草案在创建和批准时都必须仍为未来生效；默认计划时间为当前时间后300秒。
- 回滚版本持久化不可变`source_config_id`，与批准时的直接前任`supersedes_id`分开记录。
- 数据库校验复制来源同租户、同结算币、版本更早、状态可作为来源，且全部受管参数逐字段一致。
- 管理RPC、Admin API和管理端列表暴露版本来源谱系。
- 首次组合准入通过`INSERT IGNORE`创建风险钱包行时清理唯一身份负缓存，保证随后扫描不会重复插入。

涉及迁移：

- `20260730_zh_option_portfolio_risk_governance.sql`
- `20260731_zv_option_order_portfolio_config.sql`
- `20260802_option_portfolio_risk_config_lineage.sql`

本项新增DDL。已执行`make gen-model`，生成模型增加`SourceConfigId`及对应INSERT/UPDATE参数；
自定义模型文件未被生成器覆盖。随后执行`make sync-proto`，共享protobuf类型同步成功。

## 自动化证据

正式命令：

```sh
services/option/acceptance/run-p0-asset-rpc-e2e.sh
```

脚本从空数据库建表，三项组合风险治理迁移各连续执行两次。P1-004关键输出：

```text
portfolio_version_governance= configs=4 approved=1 superseded=2 rejected=1 rollback_source=1 closed_intervals=2 versioned_orders=3 distinct_order_versions=3 final_risk_v3=3
instructions=9227 success=9223 canceled=4 reconciled=9223 weighted_terminal=9231
P0 Option/Asset RPC acceptance passed
```

真实场景按墙钟顺序完成V1→V2→复制V1参数回滚V3：

| 阶段 | 订单准入版本 | 同阶段风险扫描 | 最终风险快照 |
| --- | --- | --- | --- |
| V1 | V1 | V1 | 后续切换为V3 |
| V2 | V2 | V2 | 后续切换为V3 |
| 回滚V3 | V3，`source_config_id=V1` | V3 | V3 |

拒绝路径包括：追溯生效草案、已经到达生效时刻的审批、伪造来源参数、改写已落库来源谱系。
两个被替代版本的`effective_until`均精确等于继任版本`effective_from`，没有重叠或空窗。

其他已通过检查：

```text
make gen-model                                      PASS
make sync-proto                                     PASS
go test ./internal/logic/admin ./internal/logic/helpers ./internal/logic/task ./models  PASS
go test ./... (admin-api，含回环httptest)           PASS
npm run type-check (admin-ui)                       PASS
```

## 验收中发现并修复的问题

第一次真实运行发现：新用户组合准入先读取不到风险钱包并形成负缓存，随后事务用无缓存
`INSERT IGNORE`创建数据库行，但未清除唯一身份缓存；紧随其后的风险扫描重复INSERT并触发唯一键冲突。
模型现改为通过带唯一身份缓存键的`ExecCtx`创建行，数据库写入后失效负缓存。保留原首次准入场景
重跑后，三个钱包连续扫描和完整9227条资金指令回归均通过。

## 运营交接与仍需补齐的外部证据

以下资料仓库不能代替业务、风控和独立验证人员签署，组合保证金保持`VERIFYING`且不得生产开放：

- [ ] 按`templates/option-portfolio-risk-validation-record.md`归档生产代表性历史回测、极端压力、
  数据范围/缺失处理、结果摘要和文件哈希。
- [ ] 独立模型验证人员完成理论、实现一致性、限制条件和模型风险签字。
- [ ] 在预生产使用真实账户、目标NTP/时区、正式通知链路复演未来V1→V2→回滚，保存订单、
  风险快照、告警、操作人、审批人和恢复时间。
- [ ] 将模型验证报告和预生产版本切换报告放入只读证据库，填写
  `OPTION_MODEL_VALIDATION_REPORT(_SHA256)`与
  `OPTION_PORTFOLIO_VERSION_SWITCH_E2E_REPORT(_SHA256)`，再执行production readiness。
- [ ] 连续三个风险扫描周期版本一致，且无开放SEV-1/SEV-2告警后，由风控、运营、技术共同签署。

不得把本文件当作生产模型验证报告或独立审批签字；它只证明当前仓库实现和隔离真实RPC场景。
