# Trade 用户交易控制重构与验收记录

状态：已完成（2026-08-02）  
适用范围：现货、秒合约、永续合约、交割合约  
涉及模块：`services/trade`、`services/asset`、`services/system`、`proto/trade`、`admin-api`、`admin-ui`

## 1. 结论

原“用户交易限制、用户交易对限制、用户交易配置、用户杠杆配置”四个后台入口不适合作为四套并行业务。现已收敛为一个“用户交易控制”菜单和页面：

- 产品级控制：控制某用户的现货、秒合约、全部衍生品、永续或交割交易。
- 交易对级覆盖：控制某用户在指定交易对上的数量、金额、持仓和价格限制。
- 用户杠杆：仍是用户侧交易偏好，不再作为后台独立菜单；后台继续维护交易对允许杠杆和风险档位。

完整链路已经闭环：

`Admin UI -> Admin API -> Trade Admin RPC -> MySQL -> App/Task/Internal 下单及撤单 -> 风控日志/配置审计`

因此，“Trade 服务业务成熟”指订单、成交、持仓、结算等主链路已经具备；本次补齐的是原后台用户级控制的产品设计和执行闭环，两者并不矛盾。

## 2. 改造前问题

1. `t_trade_user_config.trade_enabled` 与 `t_risk_user_trade_limit.trade_enabled` 职责重复。
2. `can_open` 曾以 `BUY` 代替开仓判断，无法正确表达衍生品开平仓。
3. `can_close`、`can_cancel`、`only_reduce_only`、挂单数、持仓及价格偏离等字段没有完整进入运行时。
4. App、Task、Internal 存在多份相似逻辑，入口行为可能漂移。
5. 后台仅有精确主键式表单，缺少统一列表、分页、失效和审计。
6. 操作人 ID 可由页面填写，不可信。
7. 用户杠杆属于用户交易偏好，不应成为后台强制配置入口。

## 3. 已落地设计

### 3.1 数据模型

后台以“用户交易控制”为统一业务概念，底层保留两层规范化存储：

- `t_risk_user_trade_limit`：产品级控制，唯一键为 `tenant_id + user_id + product_type + contract_type`。
- `t_risk_user_symbol_limit`：交易对级覆盖，唯一键为 `tenant_id + user_id + symbol_id`。
- `t_trade_user_control_audit`：创建、更新、手动解除、自动失效和迁移审计；`scope_type` 区分产品级与交易对级记录，避免相同主键碰撞。

两张控制表均包含 `control_mode`、`version`、启停状态和生效区间。`version` 用于乐观锁，更新、解除和自动失效不会静默覆盖并发修改。

`t_trade_user_config` 已退出新运行时读取路径。旧表暂留一个兼容发布周期，避免回滚时丢失旧版本依赖；它不是新功能的数据源。

### 3.2 产品范围

| 产品 | product_type | contract_type |
| --- | ---: | ---: |
| 现货 | 1 | 0 |
| 秒合约 | 3 | 0 |
| 全部衍生品默认值 | 2 | 0 |
| 永续合约 | 2 | 1 |
| 交割合约 | 2 | 2 |

衍生品按“具体 `contract_type` -> `contract_type=0` 默认值”查找；交易对级控制可覆盖产品级控制模式及数值限制。

### 3.3 控制模式

| 模式 | 运行时语义 |
| --- | --- |
| `NORMAL` | 正常交易，继续执行各数量、金额和权限限制 |
| `CLOSE_ONLY` | 只允许降低当前风险敞口 |
| `REDUCE_ONLY` | 衍生品要求显式 ReduceOnly 且确实降低敞口 |
| `DISABLED` | 拒绝新订单；是否可撤单仍由 `can_cancel` 决定 |

- 现货：BUY 增加敞口，SELL 降低敞口；资产服务仍负责余额和冻结校验。
- 秒合约：新订单都会建立新敞口，只平/只减/禁用模式均拒绝新订单。
- 永续、交割：调用方显式传递 `exposure_increasing`，结合 `side`、`position_side` 和 `is_reduce_only` 判断，禁止用 BUY/SELL 冒充开平仓。

### 3.4 统一运行时风控

App、Task、Internal 三类入口均调用同一套 `EvaluateAndLogUserOrderRisk`，当前执行：

- 控制模式、交易开关、生效区间和启停状态。
- 开仓、平仓、撤单、条件单及 API 下单权限。
- 产品级最大活动挂单数、每日下单数、每日撤单数和最大新增名义价值。
- 交易对级最小/最大数量、最小/最大名义价值、最大活动挂单数及价格偏离率。
- 现货最大持有数量和最大持有名义价值；实时读取现货钱包基础资产总额。
- 永续、交割的总持仓量、多头量、空头量、交易对持仓名义价值及产品持仓名义价值。
- 单笔撤单和全部撤单走同一 `can_cancel` 与每日撤单数校验。
- 每次通过或拒绝均写入 `t_risk_order_check_log`，快照包含命中的产品级、交易对级控制及最终模式。

数值限制为 `0` 时表示不限。时间区间采用 `[effective_start_time, effective_end_time)`；过期记录会被任务自动禁用。

## 4. Admin RPC、API 与页面

### 4.1 RPC

复用并增强已有写入/详情 RPC：

- `GetUserTradeLimit`、`SetUserTradeLimit`：产品级详情及保存。
- `GetUserSymbolLimit`、`SetUserSymbolLimit`：交易对级详情及保存。

新增运营 RPC：

- `ListUserTradeControls`：产品级、交易对级或合并游标分页查询。
- `DisableUserTradeControl`：带 `expected_version` 和原因的逻辑解除。
- `ListUserTradeControlAudits`：独立游标分页查询变更历史。

管理员 ID 和来源由可信 Metadata 注入，页面不能伪造。Admin API 与 Trade RPC 均执行租户边界校验。

### 4.2 Admin API

- `GET /trade/user-trade-controls`
- `POST /trade/user-trade-controls/disable`
- `GET /trade/user-trade-control-audits`
- 原产品级、交易对级详情和保存路由继续作为统一页面的编辑接口。

### 4.3 Admin UI

“交易管理 > 用户交易控制”作为菜单分组，下面拆分为两个独立页面，并采用与其他后台列表一致的标准 CRUD 结构：

- “用户产品级控制”：独立维护现货、秒合约、全部衍生品、永续及交割控制。
- “用户交易对级控制”：独立维护指定交易对的覆盖控制。
- 两个页面分别使用“查询条件 + 单张控制记录表 + 游标分页”，不再使用页内 Tab，也不共用组合工作台页面。
- 两个页面分别提供新增和编辑弹窗；记录标识在编辑时不可修改。
- 删除采用逻辑删除（停用）并强制填写原因，已删除记录可通过状态筛选查询，避免破坏审计链。
- 配置审计和风控校验日志分别在弹窗中独立分页查看。
- 总后台管理员不选租户时可查看全部租户，也可选择任意租户筛选和操作。
- 租户管理员的租户由登录态锁定，只能查询和操作自身租户。
- 产品级和交易对级表单各自维护，不再共用一个充满页面分支判断的工作台组件。
- 控制列表、配置审计和风控检查日志分别使用独立游标分页。
- 页面表格占满内容区，不保留无意义卡片 header。

已删除旧页面：

- `user-trade-limit.vue`
- `user-symbol-limit.vue`
- `user-trade-config.vue`
- `user-leverage-config.vue`

## 5. 审计、失效与权限

- 保存产品级或交易对级控制时，控制记录和审计记录在同一事务提交。
- 手动解除使用行锁和乐观版本校验，写 `change_type=3`。
- 定时任务和延迟队列复用同一失效函数，写 `change_type=4`、`source=TASK`。
- 审计记录包含控制范围、修改前后 JSON、操作人、来源、原因、请求 ID 和时间。
- 新增列表、解除和审计权限菜单，并从旧“用户交易限制”权限继承角色授权。
- 系统权限缓存键加入 `perms_ver`；菜单迁移提升版本后，在线管理员无需手工清 Redis 即可获得新权限。

## 6. 数据库与生成代码

已完成：

1. Trade 控制表字段、约束、旧禁用数据迁移及审计表迁移。
2. 审计 `scope_type` 的增量迁移与历史数据回填。
3. System 菜单合并、新权限及权限版本迁移。
4. 更新 `trade.sql` 和根目录 `init.sql`。
5. DDL 更新后执行 `make gen-model`，生成模型包含新增字段；未手改 `*_gen.go`。
6. 更新 Proto 后重新生成 Trade 客户端、服务端及 Admin API 代码。

## 7. 实施清单

- [x] M1：DDL、迁移、`make gen-model`、ServiceContext 接线。
- [x] M2：统一控制领域服务和订单意图判断。
- [x] M3：App、Task、Internal 下单入口接入统一服务。
- [x] M4：撤单、条件单、API 来源、挂单数、现货/衍生品持仓和价格偏离校验。
- [x] M5：Admin RPC、租户边界、配置审计和自动失效。
- [x] M6：Admin API 与单页面 UI。
- [x] M7：菜单合并、旧页面和后台用户杠杆入口下线。
- [x] M8：旧配置迁移与兼容保留。
- [x] M9：单元、RPC、数据库、权限、租户隔离和前端构建验收。

## 8. 验收记录

### 8.1 自动化检查

| 检查项 | 结果 |
| --- | --- |
| `services/trade: go test ./...` | 通过 |
| `admin-api` Trade handler/logic/types 测试 | 通过 |
| `services/system` model/admin/server 测试 | 通过 |
| `admin-ui: npm run type-check` | 通过 |
| `admin-ui: npm run build` | 通过；仅有既存的 chunk size 提示 |
| Go/前端格式与差异检查 | 通过 |

Trade 单元测试覆盖：

- 以显式敞口意图替代 BUY=开仓，并验证 BUY 平仓可通过。
- 衍生品 ReduceOnly 必须显式声明。
- 产品级通道权限和交易对级禁用覆盖。
- 限价价格偏离、生效时间区间。
- 现货最大持有数量和持有名义价值计算。

### 8.2 数据库和真实接口验收

在本地 Docker 完整依赖环境执行迁移并重建 Trade RPC、System RPC、Admin API，结果：

- Trade 与 System 新迁移均成功登记，数据库约束和菜单角色映射生效。
- 产品级创建、列表、更新、解除成功，解除后版本递增且状态禁用。
- 交易对级创建和列表成功，合并分页能够返回两类范围。
- 审计分页返回创建、更新、解除记录，`scope_type` 正确。
- 短时到期控制由任务自动禁用，版本递增，审计来源为 TASK。
- 使用旧版本解除会被乐观锁拒绝。
- 新权限迁移后在线总后台账号可直接访问列表，不再因旧 Redis 权限缓存返回 403。
- 验收创建的临时用户、产品控制、交易对控制及审计记录已按精确 ID 清理，可恢复性不适用（仅为测试数据）。

## 9. 产品验收矩阵

| 维度 | 覆盖方式 |
| --- | --- |
| 现货、秒合约、永续、交割 | 产品/合约类型数据库约束、统一入口与 Trade 全量测试 |
| 正常、只平、只减、禁用 | 统一模式判断单元测试及真实配置接口验收 |
| 普通、ReduceOnly、条件单、API 单、撤单 | 统一字段校验与 App/Task/Internal 接线测试 |
| 产品默认与交易对覆盖 | 有效控制查找及覆盖测试 |
| 数量、金额、挂单、持仓、价格偏离 | 运行时校验和专项单元测试 |
| 未生效、已生效、已过期、手动解除 | 时间区间单测、延迟到期及解除真实接口验收 |
| 总后台与租户管理员边界 | Admin API、Trade RPC 双层租户校验及公共租户隔离测试 |
| 不重启即时生效 | 运行时每次读取有效控制；真实接口修改后立即查询和失效验收 |
| 风控日志与配置审计 | 下单决策日志代码路径、配置及自动失效真实审计验收 |

## 10. 发布与后续兼容清理

发布顺序：数据库向前兼容迁移 -> Trade RPC -> System RPC -> Admin API -> Admin UI -> 菜单迁移 -> 验收。

当前没有阻塞上线的遗留项。下一兼容版本确认无需回滚旧服务后，可单独删除 `t_trade_user_config` 和已不再展示的旧 RPC/API 写入口；这属于兼容清理，不影响本次统一控制功能。

回滚时保留新增字段和审计数据，只回退调用路径与菜单，不执行破坏性数据库回滚。
