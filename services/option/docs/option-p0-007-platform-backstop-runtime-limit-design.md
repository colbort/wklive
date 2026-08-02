# OPT-P0-007 平台兜底运行时硬额度技术设计

更新时间：2026-08-02

DESIGN_STATUS: IMPLEMENTED_AND_REPOSITORY_VERIFIED / PRODUCTION_APPROVAL_PENDING

## 1. 目标和非目标

目标是在不预判业务最终选择的前提下，让Asset原生支持三种逐租户、逐结算币政策：

- `DISABLED`：拒绝所有新平台兜底请求，Option转人工。
- `PREFUNDED`：只能使用`OPTION_BACKSTOP`平台账户已充值可用余额，余额底线为0。
- `CREDIT_FLOOR`：允许余额下降到经四眼批准的负数信用底线，不得越过。

三种模式共用单请求上限、UTC日累计上限、有限有效期、不可变版本、独立复核和真实Asset业务号幂等。
本设计不决定资本来源、会计科目、具体额度或哪种模式上线；这些值仍由
`templates/option-platform-backstop-policy-approval.md`批准。

## 2. 已消除的风险与剩余边界

原实现的`CoverPlatformBackstopDeficit`曾调用`SubAvailableAllowNegative`，没有政策、单笔、UTC日累计或
余额底线，20个并发清算可把平台账户扣到任意负数。正式实现已移除该生产路径，Asset现在以账户行锁、
批准政策、UTC日累计和原子余额下限作为最终资金边界；旧实例、其他调用方和并发请求都不能绕过。

Option侧`PlatformBackstop.Enabled=false`仍是第一层失败关闭。仓库通过不代表可在生产打开：真实资本来源、
模式/额度、告警接收器、日终口径、目标环境演练和六方签署未完成时，运行开关必须保持关闭。

## 3. 数据模型

### 3.1 `t_asset_backstop_policy`

每个`(tenant_id, coin)`拥有单调递增版本：

| 字段 | 规则 |
| --- | --- |
| `request_no` | 创建幂等键，租户内唯一 |
| `version` | 同租户/币种单调递增 |
| `mode` | 1 DISABLED、2 PREFUNDED、3 CREDIT_FLOOR |
| `per_request_limit` | 模式2/3必须为正；模式1为0 |
| `daily_limit` | 模式2/3必须为正；模式1为0 |
| `balance_floor` | DISABLED/PREFUNDED为0；CREDIT_FLOOR必须小于0 |
| `effective_from/until` | 毫秒UTC；必须是有限且`until > from`的窗口 |
| `status` | 1 DRAFT、2 APPROVED、3 REJECTED |
| `created_by/reviewed_by` | 正ID且批准/拒绝时必须不同 |
| `reason/evidence_ref/review_reason` | 非空、长度受限的治理证据 |

同一时刻若多个已批准窗口重叠，运行时选`version`最大的版本。这样新版本可在不修改旧证据的情况下立即
禁用或替换政策；历史版本保持APPROVED且可重算。配置缺失、只有DRAFT、已过期或字段非法均拒绝。

### 3.2 `t_asset_backstop_usage_daily`

以`(tenant_id, coin, usage_day)`唯一，`usage_day`为服务端UTC `YYYYMMDD`。记录累计成功赔付、最后政策ID
和更新时间。版本切换不清零当日用量；新政策以已有累计与新上限比较，避免通过切版本放大额度。

### 3.3 `t_asset_backstop_cover`

新增不可变证据：`policy_id/policy_version/policy_mode`、`daily_used_before/after`、`balance_floor`和
`balance_before/after`。重放返回原赔付事实，不重新读取当前政策、不重复占用日额度。

## 4. 管理状态机

1. `CreatePlatformBackstopPolicy`只创建DRAFT；使用已认证管理员ID，租户范围必须允许写入。
2. 同一`request_no`同参数重放返回原记录，不同参数拒绝。
3. `ReviewPlatformBackstopPolicy`只允许另一管理员把DRAFT原子改为APPROVED或REJECTED。
4. 经济字段、创建人、版本和已复核记录禁止更新；所有记录禁止删除。
5. 紧急关闭也创建并批准新的DISABLED版本，不直接改旧行或平台余额。
6. 查询接口返回最新/指定政策，生产readiness绑定APPROVED政策ID、版本、证据和渲染开关。

数据库触发器执行最终状态机和不可变约束；数据库直SQL仍必须由事务外安全审计留证。

## 5. 赔付事务顺序

同一个MySQL事务内按固定顺序：

1. `PrepareAssetIdempotent`；若已完成，读取原`backstop_cover`并验证请求参数后返回。
2. 锁`OPTION_BACKSTOP`平台账户，串行化同租户/币种资金边界。
3. 锁定当前时刻最高版本的APPROVED政策；缺失或DISABLED立即拒绝。
4. 校验`requested <= per_request_limit`。
5. 读取/创建并锁UTC日用量，校验`used + requested <= daily_limit`。
6. 计算`after = available - requested`，校验`after >= balance_floor`；PREFUNDED因此不能为负。
7. 原子扣减平台账户、增加日用量、写平台流水和带政策快照的cover。
8. 完成Asset幂等记录后提交。

任何一步失败整个事务回滚：账户、日用量、流水、cover和幂等状态均不产生部分成功。账户行锁保证20路
并发看到串行的余额和日累计；唯一键处理首次日用量插入竞态。

政策缺失/无效、单笔/日累计/余额底线超限和幂等参数变化属于可预期业务拒绝，统一返回gRPC
`FailedPrecondition`。Asset的sqlx连接把明确的业务状态码视为“可接受的事务回滚”，避免连续超限请求
错误打开MySQL熔断；`Internal/Unavailable/DeadlineExceeded`及普通数据库错误不在白名单，仍按依赖故障处理。

## 6. 时间、切换与恢复

- UTC日来自Asset服务端时间，不接受Option传入日期；`usage_day`只在成功新赔付时增加。
- 时钟回拨时不会减少已有当日累计；NTP异常由生产监控阻断政策批准和卖方增险。
- 政策在请求事务锁定时解析；提交后即使新版本生效，原cover仍绑定实际使用版本。
- Asset提交后响应丢失使用原`liquidation_no`重放，读取原cover，不重复扣余额或日额度。
- 充值只增加账户余额，不降低当日累计；不能通过补资重复使用日额度。
- 首版没有自动“偿还/释放日额度”；未来若引入冲正，必须以独立幂等业务和不可变逆向流水实现。

## 7. API与权限

Asset Admin RPC新增创建、独立复核、查询政策；Admin API只做代理并沿用租户范围元数据。System新增互斥
权限：政策创建、政策复核、政策查询。充值权限与政策审批权限分离，创建人不能复核自己的记录。

生产开关仍在Option：只有Option开关true且Asset存在当前APPROVED非DISABLED政策才会赔付；两层任一关闭
均失败关闭。Asset不会信任readiness布尔值或Option配置作为额度事实。

## 8. 验收映射

| 模板用例 | 自动化证据 |
| --- | --- |
| BST-001 | 缺政策、DRAFT、REJECTED、过期、DISABLED、租户/币种不匹配零副作用 |
| BST-002/003 | 单请求上限等号成功，超过最小单位拒绝 |
| BST-004/005/006 | UTC日累计和余额/授信底线等号成功；20并发总成功额不越界 |
| BST-007 | 提交后响应丢失和重放只占用一次，并返回原政策/余额/日用量快照 |
| BST-008 | UTC午夜纯函数边界、同日版本切换和充值不清零原日用量 |
| BST-009 | 迁移双执行；直插批准、自审、改删历史、伪造用量/cover全部拒绝 |
| BST-010/011 | 事务错误零部分成功；Option关闭时不上市/不恢复并转人工 |
| BST-012 | 跨租户/跨币隔离；目标环境日终、告警送达和生产签署另行归档 |

仓库验收至少包括模型/触发器迁移双执行、`make gen-model`、`make gen`、Go单测/vet/race、隔离MySQL与
真实Asset RPC的边界/20并发/响应丢失测试，以及Option repository readiness和证据哈希更新。

## 9. 上线和回滚

1. 先部署DDL和默认无APPROVED政策的Asset版本；此时所有平台兜底失败关闭。
2. 部署Option默认关闭版本并验证上市/恢复/人工处置门禁。
3. 在预生产创建、复核明确模式的有限期政策，执行BST-001～BST-012。
4. 六方批准后，先小租户/小额度开启Option开关；监控余额、当日使用和拒绝。
5. 回滚只需批准新的DISABLED版本并关闭Option开关；禁止删除政策、cover、用量或流水。

代码和真实Asset仓库验收已完成；生产仍必须保持`PREPROD_BLOCKED`，直到审批模板、目标环境BST报告、
告警/日终证据及六方签署全部完成。

## 10. 当前实现台账

逐项通过证据以`option-p0-007-repository-acceptance.md`为准。截至2026-08-02：

- 已完成：Option默认失败关闭；三模式不可变政策与四眼复核；Asset事务内单笔/UTC日/余额硬边界；
  cover政策快照；Admin RPC/API/UI；创建/复核互斥RBAC；严格触发器；生成回环；单测/vet/type-check；
  隔离MySQL双迁移/旁路负测；真实Asset RPC边界、20并发、版本切换/补资不重置、响应丢失重放及
  Option强平端到端。
- 仓库明确口径：首版充值、偿还或未来冲正不会释放当日额度；若要释放，必须另立不可变逆向业务，
  不能修改既有用量、cover或流水。
- 仍需外部完成：逐租户/逐币真实参数和资本凭证、生产角色实名绑定、目标环境告警/日终/故障演练、
  BST报告哈希与六方签署。完成前`PlatformBackstop.Enabled=false`。
