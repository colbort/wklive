# 全仓库 Admin RPC 与 Admin UI 审计及整改计划

更新时间：2026-08-02

基线提交：`f68b736c840556dafcb10ea8cf8e66ffceb7176a`

状态：`CORE_AND_DATABASE_REMEDIATED / UI_E2E_PENDING`

## 1. 范围与结论

本次检查覆盖所有定义了`service Admin`的RPC服务，以及它们实际使用的三套后台：

- 中央后台：`admin-api`、`admin-ui`；
- 客服后台：`chat-admin-api`、`chat-admin-ui`；
- 流动性后台：`liquidity-admin-api`、`liquidity-admin-ui`。

10个Admin RPC服务共401个RPC。审计基线中RPC Server层401/401存在显式实现，管理API只有380/401能够
追踪到真实RPC调用。整改后已达到Server 401/401、管理API真实调用401/401，且加入
`scripts/verify-admin-rpc-coverage.sh`作为持续门禁。原21个RPC缺口已清零。

Chat的`Platform`服务另有5个中央平台管理RPC，已经由中央后台使用，不计入401个`Admin` RPC总数。
System的4个通知持久化/升级RPC由`NotificationStore`内部调用，不需要公开REST页面，但仍纳入已对接。

## 2. RPC全链路基线

| 服务 | Admin RPC | Server实现 | 审计基线管理API | 整改后管理API |
| --- | ---: | ---: | ---: | --- |
| Asset | 24 | 24 | 23 | 24 |
| Chat | 45 | 45 | 44 | 45 |
| Liquidity | 34 | 34 | 19 | 34 |
| Market | 32 | 32 | 32 | 32 |
| Option | 71 | 71 | 71 | 71 |
| Payment | 47 | 47 | 47 | 47 |
| Staking | 11 | 11 | 11 | 11 |
| System | 61 | 61 | 61 | 61 |
| Trade | 53 | 53 | 49 | 53 |
| User | 23 | 23 | 23 | 23 |
| **合计** | **401** | **401** | **380** | **401** |

Liquidity缺少的15个RPC：

1. `UpdateProvider`
2. `GetProviderDetail`
3. `CancelAllQuoteOrders`
4. `GetQuoteCycleList`
5. `GetExternalFillList`
6. `CancelExternalOrder`
7. `CreateManualHedge`
8. `CancelHedgeTask`
9. `RetryHedgeTask`
10. `GetInventorySnapshotList`
11. `GetLatestInventory`
12. `ResolveRiskEvent`
13. `RunReconcile`
14. `GetReconcileDetailList`
15. `ResolveReconcileDifference`

Trade的4个空代理：`SetSymbolSession`、`GetMarginSnapshotListAdmin`、`SetContractUserConfig`、
`GetContractUserConfig`。

额外HTTP风险：Payment的`DeletePayProduct`、`DeleteTenantPayAccount`、
`DeleteTenantPayChannel`、`DeleteTenantPayChannelRule`，System的`SysUserDetail`，Chat的
`UploadProfileAvatar`生成逻辑。Payment四个删除操作没有对应RPC，现已显式失败关闭；System详情已接通。
Chat头像实际由自定义multipart handler完成文件校验、存储并调用`UpdateProfile`，未使用的生成逻辑已改为
明确报错，避免将来被误调用时返回零值成功。

## 3. Admin UI现状

### 3.1 中央Admin UI

- `npm run build`通过，Vue/TypeScript生产编译正常；新增Asset保险赔付查询API客户端。
- 没有单元测试、组件测试或浏览器E2E脚本，当前只能证明“能构建”，不能证明按钮和数据链路可用。
- Trade保证金快照等4个接口已接回真实RPC。
- 生产构建出现大包警告：`element-plus`约1.01MB（minified），超过500KB；首屏仍有优化空间。
- 已按路由懒加载业务页面，但Element Plus整体导入导致公共包偏大。

#### 3.1.1 管理员类型与租户数据边界

中央Admin UI只采用两种数据权限语义：

1. **总后台管理员**（`userType=1`）：列表未选择租户时可查看全部租户数据；选择租户后按租户筛选。
   新增、修改、审批及风控处置等写操作必须落到一个明确租户，跨租户列表必须展示`tenantId`列。
2. **租户管理员**（租户主账号`userType=2`、租户管理员`userType=3`）：租户条件由登录资料强制绑定，
   页面不提供租户切换，所有查询和操作只能作用于自己的租户。

前端公共请求拦截器会覆盖租户管理员请求中已有的`tenantId`，`TenantSelect`也会同步锁定登录租户；
这只是交互和误操作防线，不能替代RPC鉴权。RPC读接口应使用统一读范围解析：总后台的`tenantId=0`
表示全部租户，租户管理员的`tenantId=0`自动收口为登录租户，显式请求其他租户则拒绝。写接口继续使用
严格写范围校验。

本轮已将Option风控中心的风险账户、强平记录、交易控制审计、MMP、异常成交更正、组合保证金配置和
保险库存退出列表接入服务端读范围解析，并修复风险账户、强平记录原先缺少服务端租户收口的问题。
其余Admin RPC的“接口可达性401/401”不等同于“租户隔离已验证”，仍需按读/写语义进行专项自动审计。

### 3.2 Chat Admin UI

- 使用Node 20执行`npm ci`和`npm run build`均通过。
- 坐席输入框已节流发送typing事件，API调用真实`SendAgentTyping`；服务端维护正式/游客会话路由状态。
- 头像上传确认由自定义multipart handler实现，并非真实空接口。
- WebSocket仍有系统通知、坐席离开、转接、送达和已读等7类TODO事件；需区分当前产品范围与真实缺口。

### 3.3 Liquidity Admin UI

- 使用Node 20执行`npm ci`和`npm run build`均通过。
- 34个Admin RPC已经全部具有管理API入口。
- “发起对账”“人工对冲”“详情”按钮已接入真实请求；对冲支持创建/取消/重试，风险支持处置，
  对账支持发起和差异明细查看。
- 提供方缺编辑/详情，订单缺外部撤单和成交视图，对冲缺创建/取消/重试，风险缺处置，
  对账缺发起/差异明细/关闭，库存没有页面。

## 4. 整改优先级

### P0：消除假成功和资金/交易运营断链

1. 修复Trade 4个空逻辑，使用统一代理调用真实RPC。
2. 修复System用户详情空逻辑。
3. Chat typing必须调用真实`SendAgentTyping`，不得本地伪造空事件。
4. Payment 4个无RPC删除接口改为明确失败关闭；在正式删除RPC、审计和引用检查落地前不允许成功。
5. Liquidity先接通人工对冲、对冲取消/重试、外部订单撤销、风险处置、发起对账和差异处置。

### P1：补齐全部管理RPC可达性

1. 补Asset保险赔付证据查询REST与后台查看入口。
2. 补齐Liquidity剩余15个RPC的API类型、路由、逻辑和UI入口。
3. 清除全部`todo: add your logic`脚手架空实现；业务范围外接口必须明确返回“不支持”，不能零值返回。
4. 建立可重复的Admin RPC覆盖校验器：比较Proto Admin方法、Server方法、管理API真实调用及允许的内部调用。

### P2：UI质量与维护性

1. 三套Admin UI必须具备`type-check`、生产`build`和静态交互冒烟测试。
2. 所有危险动作统一确认框、原因必填、进行中禁用、成功刷新、失败保留上下文。
3. 列表统一加载、空态、错误态、分页和时间/金额格式；详情按钮必须有真实行为或隐藏。
4. 中央Admin UI优化Element Plus加载方式和chunk策略，构建不再出现大于批准阈值的公共包警告。
5. 对RPC已有但UI未开放的接口，文档必须标明“内部使用/暂不开放”及理由，不能默认为遗漏或完成。

## 5. 验收标准

### RPC/API

- 401个Admin RPC均满足以下之一：存在可追踪的管理API调用；或有已审查的内部调用声明和自动校验白名单。
- 10个Admin Server的方法集合与Proto完全相等，无缺失、无多余。
- `rg -i 'todo: add your logic'`在三套管理API业务逻辑中结果为0。
- 所有REST路由必须返回真实业务响应或明确错误；不得以`nil, nil`/裸`return`制造成功空响应。
- Admin API、Chat Admin API、Liquidity Admin API均通过`go test ./...`和`go vet ./...`。

### UI

- 三套UI依赖锁文件可复现安装，`npm run type-check`和`npm run build`全部通过。
- 页面使用的每个API路径都能在对应API路由中找到；危险按钮有确认和原因，展示按钮有实际事件处理。
- Trade保证金快照、Chat typing、Liquidity人工对冲/风险/对账等本次P0链路具备正向和失败态自动测试。
- 未安装依赖、缺测试环境或缺后端凭证必须明确标为`BLOCKED`，不得写成PASS。

### UI优化

- 首屏和公共chunk大小有构建证据；超过阈值必须有拆包或批准说明。
- 请求进行中防重复提交；错误信息可见且不泄露密钥、凭证或完整敏感payload。
- 关键表格在空数据、服务错误和超时下仍有明确状态，不显示伪造成功。

## 6. 执行顺序与状态

| 编号 | 整改项 | 状态 |
| --- | --- | --- |
| ADM-P0-001 | Trade 4个空代理 | DONE |
| ADM-P0-002 | System用户详情空代理 | DONE |
| ADM-P0-003 | Chat typing真实RPC | DONE |
| ADM-P0-004 | Payment 4个删除接口失败关闭 | DONE |
| ADM-P0-005 | Liquidity处置类P0接口与UI | DONE |
| ADM-P1-001 | Asset保险赔付证据查询 | DONE |
| ADM-P1-002 | Liquidity 15个Admin RPC完整覆盖 | DONE |
| ADM-P1-003 | Admin RPC覆盖与空逻辑自动校验器 | DONE |
| ADM-P1-004 | Admin新增接口菜单、RBAC权限与数据库迁移合并 | DONE |
| ADM-P2-001 | 三套UI类型检查和生产构建 | DONE |
| ADM-P2-002 | 危险操作、错误态、空态和分页一致性 | PARTIAL |
| ADM-P2-003 | 公共包拆分 | PARTIAL |
| ADM-SEC-001 | 中央Admin UI两级管理员租户边界 | DONE |
| ADM-SEC-002 | 全部Admin RPC读写租户隔离专项审计 | PARTIAL（Option风控中心已完成） |

## 7. 整改后测试证据

| 检查 | 命令 | 结果 |
| --- | --- | --- |
| Admin RPC覆盖 | `./scripts/verify-admin-rpc-coverage.sh` | PASS：401/401/401 |
| 中央管理API | `GOCACHE=/private/tmp/wklive-go-build-cache go test ./...` | PASS |
| Chat管理API | `GOCACHE=/private/tmp/wklive-go-build-cache go test ./...` | PASS |
| Liquidity管理API | `go test ./...` | PASS |
| 10个RPC服务 | 各服务执行`go test ./internal/...` | PASS；同时修复Liquidity OKX测试夹具和Payment DECIMAL边界校验 |
| 三套管理UI | `./scripts/verify-admin-ui.sh`（Node >= 20） | PASS；包含Chat typing与Liquidity对账假数据组件测试 |
| 数据库完整升级 | `./deploy/deploy.sh database` | PASS：119个迁移校验完成，连续执行两次均成功 |
| Liquidity RBAC菜单 | 查询`sys_menu/sys_role_menu` | PASS：34条接口规则、35条做市菜单全部授予`liquidity_admin` |
| Asset保险覆盖权限 | 查询菜单ID 759及角色授权 | PASS：接口已登记，申请员、复核员、超级管理员均获只读权限 |
| Option风控租户边界 | `go test ./utils`、`go test ./internal/logic/admin ./models`、中央UI `type-check/build` | PASS：总后台支持全租户列表，租户管理员强制绑定自己的租户 |

依赖安装时发现：`liquidity-admin-ui`有3个npm审计漏洞（1 moderate、2 high），`chat-admin-ui`有4个
（1 moderate、3 high）。本次未直接执行`npm audit fix --force`，因为它可能升级主版本并引入破坏性变化；
需要单独做依赖升级回归。

## 8. 尚未完成或需要真实环境验证

1. Chat typing和Liquidity对账差异已经补充假数据组件测试；中央Admin UI及全量页面仍缺浏览器E2E，当前PASS
   不应等同于所有页面的人工验收。
2. Chat WebSocket仍有系统通知、坐席离开、转接、送达、已读等7类产品范围TODO，需要按业务范围逐项实现。
3. Chat和Liquidity已把应用、Vue、Axios与Element Plus拆成可缓存chunk，主业务chunk明显缩小；但Element Plus
   公共chunk仍约922KB（minified），仍超过500KB。进一步优化需要改为组件级按需引入。
4. Liquidity新运营动作已经具备确认/原因/防重复提交的主要流程；全站危险动作一致性和错误态仍需继续梳理。

## 9. 数据库与菜单合并说明

本轮不仅补代码路由，也同步处理了已有库升级和全新安装：

1. `init.sql`已合并Liquidity新增15条接口权限、Asset保险覆盖权限，以及此前只存在于迁移文件中的
   Option交易日历、公司行动、合约系列、组合单、保险库存退出和平台兜底政策菜单。
2. 新增幂等数据迁移`20260802_zz_admin_extension_permissions.sql`；做市后台采用fail-closed RBAC，未登记
   菜单的接口会返回405，因此34条受保护路由必须与34条`sys_menu`接口规则一一对应。
3. `db-init`重复数据迁移清单已覆盖7月30日至8月2日新增的菜单、角色和任务，已有库执行升级不会再漏菜单。
4. 已恢复被修改的历史Asset迁移checksum，并把兼容逻辑移到新的baseline-safe reconciliation migration。
5. `db-init`现可解析MySQL CLI的`DELIMITER`脚本，触发器和存储过程迁移可以通过Go驱动执行；相关解析器有单元测试。

整改完成后，本表状态、实际测试命令、结果和仍需真实环境验证的边界必须同步更新。
