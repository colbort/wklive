# 组合/价差单上线准备与审批记录

> 首版是独立策略簿，不包含单腿簿隐含撮合、拍卖或组合保证金抵扣。仓库已预填当前系统硬门禁；
> “待审批”项未取得业务签字时只能内部联调。

## 1. 市场规则

| 项目 | 批准值 | 依据/证据 | 审批人 |
| --- | --- | --- | --- |
| 允许策略/合约范围 | 系统值：不同现金期权、同标的/到期/币种/单位/乘数、OPEN、至少一买一卖、逐仓卖方保证金；待审批 | `validateComboOrder` |  |
| 最大腿数（首版≤4） | 系统硬上限：4 | protobuf/API/数据库 CHECK |  |
| 最大比例（首版≤8） | 系统硬上限：8，且全腿 GCD=1 | 规范化校验/数据库 CHECK |  |
| 最大父单数量 | 系统值：每腿 `父数量×比例` 同时满足该合约 min/max/step；是否另设全局上限待审批 | 合约风控参数/容量报告待补 |  |
| 净价最小变动 | 系统值：无独立净价 tick；净价必须精确等于满足各自 tick 的腿价有符号加权和；待产品确认 | 精确十进制校验 |  |
| 腿限价 | 系统值：成交同时满足净价；taker 买腿不高于腿限价、卖腿不低于腿限价 | 毛额资金安全门禁 |  |
| 费用/返佣口径 | 系统值：每条成交腿沿用合约 maker/taker 费率，无组合附加费；待财务/产品审批 | `optionTradeFees` |  |
| 策略簿交易时段 | 系统值：每条腿均须命中其已批准日历且无活动 halt；市场时段待审批 | 交易日历运行门禁 |  |
| FOK 流动性 | 系统值：仅最优单个 maker，禁止跨 maker 聚合；待产品披露确认 | 首版撮合边界 |  |
| 单腿 legging | `DISABLED` | 首版明确边界 |  |
| 组合保证金抵扣 | `DISABLED`；首版只接受逐仓毛额冻结 | P2-007 独立于 P1-004 |  |

## 2. 首版能力确认

- [x] 仅2～4条不同现金期权开仓腿，同标的/到期/币种/单位/乘数，逐仓卖方保证金。
- [x] 比例为1～8且最大公约数为1，至少一买一卖。
- [x] 净价等于精确腿价加权和；正为净支出，负为净收入；另有逐腿限价保护。
- [x] 仅 LIMIT/FOK；首版 FOK 只用一个最优 maker；影子单不进入单腿簿；非 MMP。
- [x] 全部资金毛额预冻结，不使用跨腿抵扣；所有腿冻结成功后才激活父单。
- [x] 一次策略成交的全部腿、outbox 和资金指令单事务提交。
- [x] 组合成交统一记录 `comboMatchNo/comboLegNo`；全组买方扣冻结完成前不运行持仓事件。

## 3. 当前接口与状态

- 创建：`POST /app/option/combo-orders`
- 撤销：`POST /app/option/combo-orders/cancel`
- 详情：`GET /app/option/combo-orders/detail`
- 分页：`GET /app/option/combo-orders`
- 管理列表：`GET /admin/option/combo-orders`
- 管理下钻：`GET /admin/option/combo-orders/detail`（腿、影子单、成交组、资产指令及总数）
- 管理整组强撤：`POST /admin/option/combo-orders/force-cancel`（必须填写原因）
- 父状态：`FUNDING → ACTIVE → PART_FILLED/FILLED`，撤销经 `CANCELING → CANCELED`；
  资产指令不可自动恢复时转 `MANUAL_REVIEW`。
- 普通委托列表、普通 matcher、公开盘口和单腿撤单均不能消费影子子单。

## 4. 上线验收

- [x] 规范化/边界/影子隔离/状态进度 Go 单测，Option/App API 全量测试、vet 和 App Web 类型检查。
- [x] MySQL 8.4 全量 schema、迁移双执行、关联 CHECK、不可变/禁止删除触发器和合法状态迁移。
- [x] 数据库验证同一组合成交组任一买方扣款未完成时 runnable 持仓事件为0，全完成后为全腿。
- [x] 管理父单列表、下钻、100条安全截断、父腿 DTO 关联、整组强撤接口和三项独立权限已落地；
  权限迁移双执行及 Admin API/UI 检查通过。
- [ ] COMBO-001～COMBO-009 预生产场景全部通过并归档 RPC、消息、Asset 流水和故障注入证据。
- [ ] 目标并发下无死锁、残腿、重复撮合序号、冻结泄漏或单腿簿污染。
- [ ] halt、kill switch、到期、进程重启和 Asset 响应丢失演练通过。
- [ ] 复杂簿行情、净价方向、部分成交、取消和错误原因的页面/客服文案已批准。
- [ ] 明确向用户披露首版不 legging、不拍卖、不含股票腿、不提供组合保证金抵扣。

## 5. 监控与人工查询

```sql
-- 超龄/人工组合父单
SELECT tenant_id,status,COUNT(*) AS cnt,MIN(update_times) AS oldest
FROM t_option_combo_order
WHERE (status IN (1,5) AND update_times<UNIX_TIMESTAMP()-60) OR status=8
GROUP BY tenant_id,status;

-- 父单、腿和影子单进度不一致；正常结果为0行
SELECT p.tenant_id,p.id,p.combo_no,p.status,p.filled_qty,p.unfilled_qty,
       COUNT(l.id) AS leg_count,
       SUM(l.filled_qty<>p.filled_qty*l.ratio
           OR l.unfilled_qty<>p.unfilled_qty*l.ratio
           OR o.id IS NULL OR o.combo_order_id<>p.id OR o.combo_leg_no<>l.leg_no) AS bad_legs
FROM t_option_combo_order p
LEFT JOIN t_option_combo_order_leg l
  ON l.tenant_id=p.tenant_id AND l.combo_order_id=p.id
LEFT JOIN t_option_order o
  ON o.tenant_id=l.tenant_id AND o.id=l.child_order_id
GROUP BY p.tenant_id,p.id,p.combo_no,p.status,p.filled_qty,p.unfilled_qty
HAVING leg_count NOT BETWEEN 2 AND 4 OR bad_legs<>0;
```

## 6. 签字

| 角色 | 姓名 | 结论 | 时间 | 证据/工单 |
| --- | --- | --- | --- | --- |
| 产品 |  |  |  |  |
| Option 技术 |  |  |  |  |
| 风控 |  |  |  |  |
| 清算/财务 |  |  |  |  |
| 合规/法务 |  |  |  |  |
