# OPT-P0-007 平台兜底运行时硬额度仓库验收

更新时间：2026-08-02

`REPOSITORY_ACCEPTANCE_STATUS: PASSED / PREPROD_BLOCKED`

## 1. 结论与边界

平台兜底的仓库级无限负余额阻断已经关闭：Asset只接受当前已批准政策，在同一MySQL事务内原子执行
单请求、UTC日累计和余额/信用底线校验，并保存不可变政策、日用量和余额快照。Option仍默认关闭，
无批准政策时不允许相关合约上市/恢复，清算剩余缺口转人工。

这不是生产放行结论。真实资本来源、逐租户/逐币模式和额度、生产人员、告警/日终目标环境证据及六方
签署仍未完成，所以统一状态保持`PREPROD_BLOCKED / NOT_RELEASEABLE`。

## 2. 正式仓库实现

| 域 | 正式实现 | 结果 |
| --- | --- | --- |
| 失败关闭 | `PlatformBackstop.Enabled=false`；上市、恢复和清算三处门禁 | 默认不产生新平台兜底风险 |
| 资金政策 | `DISABLED`、`PREFUNDED`、`CREDIT_FLOOR`三模式，有限有效期和单调版本 | 缺失、草稿、拒绝、过期、关闭均拒绝 |
| 资金事务 | 锁平台账户→锁最高有效政策→锁UTC日用量→校验→扣款/用量/流水/cover/幂等提交 | 无部分成功、无无限负余额调用 |
| 错误隔离 | 业务拒绝使用`FailedPrecondition`并列入sqlx可接受错误；系统错误仍触发DB熔断 | 连续超限不会误熔断正常资金请求 |
| 幂等证据 | cover保存政策ID/版本/模式、日用量和余额before/after | 重放返回原快照，不读取新政策或重复占额 |
| 政策治理 | 创建只产DRAFT；异人复核；经济字段和终态不可变；所有证据不可删除 | 四眼和不可变由应用及触发器双重约束 |
| 管理入口 | Asset Admin RPC、Admin API、管理UI | 创建、复核、详情、分页均已接通 |
| 权限 | `platform_backstop_policy_creator`与`platform_backstop_policy_reviewer` | 创建/复核权限互斥，查询均可授予 |
| DDL | 政策、UTC日用量、cover快照及9个触发器 | 直SQL伪造政策/用量/cover被拒绝 |
| 首版冲正口径 | 充值、偿还和未来冲正不减少当日累计 | 不会通过补资或改历史重复放大日额度 |

## 3. 实现完成定义

### 3.1 Asset最终资金边界

- [x] 生产路径不再存在`SubAvailableAllowNegative`调用或方法。
- [x] 同一事务锁账户、解析最高有效批准版本、锁UTC日用量、检查三类硬边界并写完整证据。
- [x] `DISABLED`恒拒绝；`PREFUNDED`余额不低于0；`CREDIT_FLOOR`不低于批准负底线。
- [x] 缺失/DRAFT/REJECTED/未生效/过期/错租户币种均零副作用拒绝。
- [x] 使用Asset服务端UTC；同日切版本、补资和重放不清零或重复占用。
- [x] 幂等重放返回原政策、日用量和余额快照。
- [x] 政策/额度业务拒绝仍回滚事务，但不计入MySQL故障熔断；依赖错误不被白名单吞掉。

### 3.2 政策、管理和权限

- [x] 创建只产生DRAFT；同`request_no`同参数重放，改参数拒绝。
- [x] 只有另一个已认证管理员可批准/拒绝；创建人自审拒绝。
- [x] 经济字段、创建人、版本、终态、日用量和cover不可改删。
- [x] Admin RPC/API/UI覆盖创建、复核、详情和分页。
- [x] System RBAC分离创建、复核和查询；两个职责角色写权限互斥。
- [x] 紧急关闭采用新的已批准`DISABLED`版本，不修改旧记录。

### 3.3 生成和静态验收

- [x] 实际执行Asset `make gen-model`，生成器跳过手写模型扩展。
- [x] 实际执行Asset `GOCACHE=/private/tmp/wklive-go-build-cache make gen`，并同步共享proto。
- [x] 生成后重新搜索，未出现政策logic空桩或无限负余额方法。
- [x] Asset全量`go test ./...`和`go vet ./...`通过。
- [x] Option全量`go test ./...`和`go vet ./...`通过。
- [x] Admin API全量`go test ./...`和`go vet ./...`通过；沙箱端口限制场景在允许本机
  `httptest`监听后原命令通过。
- [x] Admin UI使用Node 20.20.2执行`npm run type-check`通过。

## 4. 自动化验收矩阵

| 证据ID | 已执行证据 | 结论 |
| --- | --- | --- |
| BST-001 | 连续执行无政策、DRAFT、REJECTED、过期、DISABLED真实RPC | 全部`FailedPrecondition`；零副作用且后续正常请求不熔断 |
| BST-002 | PREFUNDED单笔恰等于10 | 成功且快照、余额、用量一致 |
| BST-003 | 单笔10加`0.000000000000000001` | 原子拒绝、零副作用 |
| BST-004 | 日限额恰等于20，随后最小单位请求 | 20成功，越界拒绝 |
| BST-005 | PREFUNDED余额到0；CREDIT_FLOOR余额到-10 | 等号成功，越界拒绝 |
| BST-006 | CREDIT_FLOOR下20个并发1单位请求 | 精确10成功/10拒绝；余额-10、日用量10、10 cover/10流水 |
| BST-007 | 成功响应重放；Option注入保险和平台回补提交后响应丢失 | 原业务号返回原快照；平台23只扣一次 |
| BST-008 | UTC午夜毫秒边界单测；同日版本1→2；补资10 | UTC桶正确；版本/补资不重置20日用量 |
| BST-009 | Asset迁移连续两次；6类直SQL负向旁路 | 3表/9触发器/8快照列；全部`REJECTED` |
| BST-010 | 超限/触发器错误事务回滚及响应丢失恢复 | 无只扣款、只占用、只写cover的部分成功 |
| BST-011 | Option上市/恢复/清算门禁及单测 | 开关关闭时失败关闭并转人工 |
| BST-012 | 跨币真实RPC、逐租户并发数据、RBAC双迁移 | 无串账；生产日终/告警送达仍由目标环境归档 |

数据库汇总证据：策略9（批准7、拒绝1、草稿1），cover/平台流水各15，覆盖总额63；UTC日用量4行、
合计63。该汇总包含独立Asset边界场景40和Option穿仓场景23，隔离数据库及Redis在验收后删除。

## 5. 可重放命令

```bash
cd services/asset
make gen-model
GOCACHE=/private/tmp/wklive-go-build-cache make gen
GOCACHE=/private/tmp/wklive-go-build-cache go test ./...
GOCACHE=/private/tmp/wklive-go-build-cache go vet ./...

cd ../option
./acceptance/run-platform-backstop-schema-acceptance.sh
./acceptance/run-platform-backstop-rbac-acceptance.sh
./acceptance/run-platform-backstop-rpc-acceptance.sh
```

`run-platform-backstop-rpc-acceptance.sh`构建正式仓库Asset二进制，创建白名单隔离数据库和独立Redis，
执行真实gRPC边界/并发/重放/版本切换，再执行Option强平缺口子场景；退出时仅删除本次命名资源。

## 6. 仍需外部完成的生产材料

- 每个租户/法人和结算币的模式、单笔/日/余额底线、预警阈值和有效期。
- 资本所有者、资金来源、授信/预注资凭证、会计科目、补资SLA和失败升级。
- 生产申请员/复核员实名账号、紧急关闭工单、DB审计及SOC接收器。
- 生产70%/85%/100%告警路由、日终重算、值班回执及目标环境故障/接管报告。
- 产品、风控、财务、清算、技术、合规六方的真实结论和签名。

上述材料未齐时，`option-platform-backstop-policy-approval.md`必须保持`DRAFT`，
`OPTION_PLATFORM_BACKSTOP_ENABLED=false`，不得把仓库验收结论外推为生产批准。
