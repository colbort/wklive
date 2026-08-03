# 当前系统 DTM 引入必要性与 gRPC 跨服务事务评估

> 文档状态：`ASSESSMENT_COMPLETE / GLOBAL_ADOPTION_NOT_RECOMMENDED`  
> 评估日期：2026-08-03  
> 适用范围：`admin-api`、`services/{asset,trade,option,staking,payment,liquidity,user,market}`、相关数据库与消息任务  
> 核心结论：当前系统不应全量引入 DTM，也不应将 DTM 当作通用 gRPC 调度框架。现有核心资金链路继续使用“业务操作单/资金指令 + Asset 幂等 + Outbox/Inbox + 自动重试 + 对账 + 人工处理”。只有新建、低频、执行时间短、跨多个写服务且具备明确补偿动作的流程，才进入 DTM 小范围试点评估。

## 1. 评估目标

本次评估回答以下问题：

1. 当前系统是否有必要引入 DTM；
2. DTM 是否能够简化各 RPC 服务之间的 gRPC 调度；
3. DTM 与现有幂等、Outbox/Inbox、业务 Saga、资金指令及对账机制是否重复；
4. 哪些业务适合使用 DTM，哪些业务明确禁止接入；
5. 如果后续试点，需要满足哪些设计、开发和验收条件。

本次为只读架构审计，没有修改现有业务代码、数据库和部署配置。

## 2. 决策摘要

| 决策项 | 结论 |
| --- | --- |
| 全系统引入 DTM | 不建议 |
| 用 DTM 替代 go-zero/gRPC 调度 | 不适用 |
| 用 DTM 替代现有资金指令和业务操作单 | 不建议 |
| 用 DTM 替代 Outbox/Kafka | 不建议 |
| Trade/Option 核心交易结算链路接入 | 禁止直接改造 |
| Staking/Payment 已有资金链路接入 | 暂不改造 |
| Admin API 查询及单 RPC 转发接入 | 不适用 |
| 新增的短周期、可补偿、多服务写流程 | 可评估 |
| Liquidity 内部做市账户开通 | 可作为候选，但需先补齐补偿接口 |

DTM 可以减少一部分分布式事务的通用编排代码，但不能减少 gRPC 调用数量，也不会替代服务发现、负载均衡、超时、熔断、限流、租户上下文、链路追踪或查询聚合。

## 3. 当前 RPC 拓扑

### 3.1 业务服务依赖

| 调用方 | 下游 RPC | 主要语义 |
| --- | --- | --- |
| Trade | Asset、Market | 写资产冻结/结算，读取行情 |
| Option | Asset | 写保证金、权利金、交割和强平资金 |
| Staking | Asset | 写锁仓、解锁、奖励和手续费 |
| Payment | Asset | 写充值到账等资产结果 |
| Liquidity | Trade、Market、User、Asset | 查询交易对/行情，创建内部用户，初始化资金 |
| User | System | 系统配置或基础能力调用 |
| Admin API | 各业务 Admin RPC | 查询和单业务操作转发 |

Trade 的 RPC 客户端和事件基础设施集中在 `services/trade/internal/svc/servicecontext.go`；Liquidity 同时依赖四个业务 RPC，见 `services/liquidity/internal/svc/servicecontext.go`。

### 3.2 Admin API 不属于分布式事务热点

Admin API 虽然持有多个 RPC 客户端，但现有处理器基本是单个管理请求转发到单个业务 RPC。查询列表、详情、配置选项和只读聚合不需要分布式事务。

DTM 不应被加入以下链路：

- Admin UI 查询列表或详情；
- Admin API 到单个业务服务的写操作；
- 多个只读 RPC 的查询聚合；
- Market 行情获取、快照读取或 WebSocket 推送。

## 4. 当前一致性基础设施

### 4.1 Asset 幂等中心

`services/asset/asset.sql` 中的 `t_asset_idempotent` 使用租户、业务类型、场景和业务号建立唯一约束。Trade、Option、Staking、Payment、Liquidity 调用 Asset 时已经传递确定性 `biz_no`。

结论：DTM 即使接入，Asset 分支仍然必须保留业务幂等；DTM 不能替代资产流水和业务号审计。

### 4.2 Staking 可恢复资金操作单

`services/staking/staking.sql` 中的 `t_stake_operation` 已包含：

- `operation_no` 和 `request_no` 唯一约束；
- 本金、奖励、手续费三个独立步骤状态；
- 处理中、成功、可重试失败和人工处理状态；
- 重试次数、下次重试时间、错误信息和版本号。

`services/staking/internal/logic/helpers/fund_operation.go` 按步骤调用 Asset，并在每一步成功后持久化状态。失败进入延迟重试，重复请求依靠确定性业务号和 Asset 幂等收敛。

结论：Staking 已实现符合业务审计要求的持久 Saga，不建议为了使用 DTM 重写。

### 4.3 Trade 事件、预占和结算指令

`services/trade/trade.sql` 已包含：

- `t_trade_event_inbox`：事件消费幂等；
- `t_trade_asset_reservation`：订单资金预占状态镜像；
- `t_trade_settlement_instruction`：发送给 Asset 的幂等结算指令；
- 重试次数、租约/执行状态、人工处理和对账字段；
- 合约对账问题及游标。

`services/trade/internal/logic/task/contract_saga.go` 对资金指令执行、失败退避、重试上限和人工处理进行了显式控制。

结论：Trade 的领域状态需要后台可见、可对账、可手工重试。DTM 的全局事务记录不能替代这些交易领域表。

### 4.4 Option Outbox/Inbox 与资产指令

`services/option/option.sql` 已包含：

- `t_option_outbox`；
- `t_option_inbox`；
- `t_option_asset_instruction`；
- 行权、交割、强平及结算相关对账状态。

`services/option/internal/logic/task/processassetinstructionslogic.go` 根据指令动作调用 Asset，记录实际流水号、失败重试、人工处理和对账结果。

结论：期权行权、到期、实物交割、强平等流程持续时间长、步骤多，且需要明确业务中间态，不适合直接收口为 DTM 短事务。

### 4.5 Payment 可靠事件

`services/payment/payment.sql` 中的 `t_pay_outbox` 已提供事件唯一号、状态、重试次数、下次重试时间和错误信息。

结论：DTM 事务消息不能直接替代现有 Kafka/消息订阅体系，也不能替代支付渠道状态、回调流水和资金对账。

## 5. DTM 能够简化的内容

根据 DTM 官方文档，DTM 支持 Saga、TCC、XA、事务消息和 Workflow，并支持 gRPC 及 go-zero：

- go-zero 支持：<https://en.dtm.pub/ref/gozero.html>
- Workflow：<https://en.dtm.pub/practice/workflow>
- 子事务屏障：<https://en.dtm.pub/practice/barrier.html>
- DTM 架构：<https://en.dtm.pub/practice/arch.html>

DTM 可以统一处理：

1. 全局事务 ID 和分支 ID；
2. Saga 正向步骤和反向补偿的调用顺序；
3. TCC Try、Confirm、Cancel 调度；
4. DTM/业务进程崩溃后的分支恢复；
5. 重复调用、空补偿和防悬挂；
6. Workflow 分支结果记录和重放；
7. 部分通用指数退避和事务状态维护。

这些能力可以减少新业务自行实现通用事务协调器的代码量。

## 6. DTM 不能简化的内容

DTM 不负责且不能替代：

1. go-zero `zrpc` 客户端创建；
2. etcd/Kubernetes 服务注册与发现；
3. gRPC 连接池、负载均衡、超时、熔断和限流；
4. Metadata、登录身份和租户上下文传播；
5. OpenTelemetry/日志链路追踪；
6. 多个只读接口的并行查询和结果聚合；
7. Kafka 事件总线和消费者拓扑；
8. 业务账单、资金流水、结算指令和审计记录；
9. 日对账、差异处置和人工修复；
10. 业务补偿接口本身。

DTM 可以发起或重试 gRPC 分支调用，但普通 gRPC 调用仍然存在。服务仍需定义 Proto、实现 RPC、处理业务错误、保证幂等并提供补偿动作。

## 7. 引入 DTM 的新增成本

全量接入至少会增加：

- DTM 服务集群；
- DTM 共享高可用存储；
- 每个参与服务的事务回调接口；
- DTM SDK、go-zero Driver 和拦截器；
- GID、Branch ID 和错误语义传播；
- Branch Barrier 数据表及升级脚本；
- DTM 事务监控、积压告警、备份和恢复；
- DTM 宕机、存储异常、重复回调和补偿失败 Runbook；
- 原有业务操作单与 DTM 全局事务之间的映射和排障能力。

官方架构依赖多个 DTM 实例和共享高可用数据库。分支在极端情况下仍可能重复执行，因此参与接口仍须幂等。

### 7.1 与当前重试语义的冲突

Trade、Option、Staking 当前均支持“达到重试上限后进入人工处理”。DTM Workflow 官方文档说明：Workflow 必须幂等，目前不支持“达到重试上限后回滚”和部分超时回滚语义，未知错误会继续重试。

因此，即使采用 DTM，当前系统仍要额外保留：

- 重试次数和人工处理状态；
- 管理员安全重试；
- 冲正和差异处置；
- 资金对账和领域审计。

这也是当前核心链路不适合直接替换的主要原因。

## 8. 场景适用矩阵

| 场景 | 特征 | DTM 判断 | 推荐方案 |
| --- | --- | --- | --- |
| Admin 查询/列表/详情 | 只读或单 RPC | 不适用 | 普通 gRPC，必要时并行聚合 |
| Trade 下单、撮合、成交结算 | 高并发、资金预占、异步事件、需审计 | 不建议 | 现有指令、Inbox/Outbox、对账 |
| 永续资金费、交割、强平、ADL | 长流程、批次化、可人工恢复 | 不建议 | 业务批次和结算指令 |
| Option 行权、到期、实物交割、强平 | 长流程、多方资金、业务中间态 | 不建议 | Option 指令、事件箱和对账 |
| Staking 申购、收益、赎回 | 已有持久操作单和确定性业务号 | 不建议改造 | 现有资金操作单 |
| Payment 渠道回调、充值到账 | 外部系统、异步回调、需渠道对账 | 不建议 | 支付流水、Outbox、Asset 幂等 |
| 行情读取和推送 | 高频只读 | 禁止 | Market RPC/缓存/消息推送 |
| 新增短周期多服务写操作 | 可逆、秒级、明确补偿 | 可评估 | Saga 或 TCC |
| Liquidity 内部账户开通 | User + Asset + Liquidity 多写 | 候选 | 先补补偿，再试点 Saga/TCC |

## 9. 候选试点：Liquidity 内部做市账户开通

当前 `ProvisionInternalProvider` 的执行顺序为：

1. 从 Trade 查询交易对；
2. 在 User 创建内部做市用户；
3. 在 Asset 初始化基础币余额；
4. 在 Asset 初始化计价币余额；
5. 在 Liquidity 创建 Provider。

代码位置：`services/liquidity/internal/logic/admin/provisioninternalproviderlogic.go`。

当前已对两个 Asset 入账步骤使用确定性 `biz_no`，具备部分幂等基础；但仍可能出现用户已创建、部分币种已入账或 Provider 未创建的中间状态。

### 9.1 试点前置条件

在决定使用 DTM 前，必须先完成：

1. User 创建内部用户接口支持确定性请求号；
2. 明确“撤销用户”是删除、禁用还是保留待恢复；
3. Asset 提供与初始化入账对应的安全冲正接口；
4. Provider 创建具备租户内唯一幂等键；
5. 每个补偿接口允许重复调用和空补偿；
6. 明确部分资金已被使用时是否还能补偿；
7. 增加开通操作单，记录 GID、业务状态和人工处理结果；
8. 明确 DTM 不可用时该管理操作是失败关闭还是进入排队。

如果这些条件无法满足，使用 DTM 也不能保证业务可补偿，应继续采用持久开通操作单和恢复任务。

### 9.2 模式选择

- 优先考虑 Saga：适合低频管理操作，正向步骤和补偿步骤清晰；
- 需要先冻结资源、最终确认时考虑 TCC；
- 不建议 XA：跨服务且涉及共享资金热点，不应持有长数据库事务和行锁；
- 不建议把该流程包装成不可观测的动态 RPC 串，仍需保存业务开通单。

## 10. DTM 试点准入条件

只有同时满足以下条件，业务才能进入 DTM 设计评审：

1. 至少跨两个独立写服务；
2. 无法用单服务本地事务解决；
3. 正常完成时间为秒级或短分钟级；
4. 每个步骤都有确定性幂等键；
5. 每个已提交步骤都有明确、合法、可重复的补偿操作；
6. 不属于撮合、行情、行权、交割、强平等长流程；
7. 不依赖“重试 N 次后自动回滚”语义；
8. 业务允许 DTM 协调器成为关键依赖；
9. 已设计监控、告警、人工恢复和对账；
10. 经过故障注入证明比现有操作单方案更简单且风险更低。

任一条件不满足，默认不使用 DTM。

## 11. 试点验收标准

如果未来批准试点，必须完成以下测试：

### 11.1 正常流程

- 全部正向分支只执行一次业务效果；
- DTM 全局状态、业务操作单和各服务状态一致；
- 请求重放返回同一结果，不重复增减资产；
- 日志可由业务号定位到 GID 和全部 Branch ID。

### 11.2 故障注入

- 每个正向分支调用前宕机；
- 分支数据库提交后、RPC 返回前断网；
- 每个补偿分支调用前后宕机；
- DTM 服务重启；
- DTM 存储短暂不可用；
- 业务服务多实例并发收到重复回调；
- gRPC 超时但下游实际成功；
- 补偿接口连续失败；
- DTM 恢复后积压事务重放。

### 11.3 一致性与运维

- 无重复资产流水；
- 无负余额或无来源增发；
- 无永久卡在处理中且后台不可见的事务；
- 补偿失败能够告警并进入人工处理；
- DTM 全局状态与业务操作单可自动核对；
- DTM 完全不可用时不影响 Trade、Option、Staking、Payment 现有主链路；
- 提供停用 DTM 试点并切回原业务编排的回滚方案。

## 12. 推荐执行顺序

| 编号 | 工作项 | 当前状态 |
| --- | --- | --- |
| DTM-ADR-001 | 当前 RPC 拓扑与一致性机制审计 | DONE |
| DTM-ADR-002 | 全量引入必要性决策 | DONE：不建议 |
| DTM-ADR-003 | 明确禁止接入的核心链路 | DONE |
| DTM-PILOT-001 | Liquidity 开通流程补偿能力设计 | NOT_STARTED |
| DTM-PILOT-002 | 对比 DTM Saga 与现有持久操作单方案 | NOT_STARTED |
| DTM-PILOT-003 | 部署隔离的 DTM 测试环境 | NOT_APPROVED |
| DTM-PILOT-004 | 故障注入、对账与回滚验收 | NOT_APPROVED |

`NOT_APPROVED` 表示当前仅完成架构评估，不代表已经决定引入或部署 DTM。

## 13. 最终架构决策

1. 当前生产架构不增加 DTM 依赖；
2. Trade、Option、Staking、Payment 继续使用现有业务一致性方案；
3. Admin API 和普通 gRPC 调用不接入 DTM；
4. 不以 DTM 替代 Kafka、Outbox/Inbox、资金指令、对账或审计表；
5. 后续新业务按第 10 节准入条件单独评审；
6. Liquidity 内部做市账户开通仅作为候选，不在补偿接口和验收方案完成前实施；
7. 如果试点结果不能明显减少代码和故障恢复复杂度，应终止引入并保留现有架构。

本决策的重点不是拒绝 DTM，而是避免在已经具备成熟领域恢复机制的资金系统中重复建设。DTM 适合解决新的、边界清晰的跨服务短事务，不适合作为整个系统的通用 RPC 调度层。
