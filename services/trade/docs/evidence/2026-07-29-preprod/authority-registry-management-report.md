# Authority Registry 管理入口、动态下拉与独立供应商门禁报告

## 1. 目标与边界

- 日期：2026-07-29；
- 环境：Deploy 独立完整环境；
- 目标：让后续真实行情源可以通过受控管理入口注册，并由价格公式页面动态选择；
- 独立性目标：Authority 表示发布身份，`provider_code` 表示真实数据供应商；同一
  供应商的 WS/REST 等不同通道不得被计算为多个独立来源；
- 边界：不创建虚假第三来源，不修改现有公式，不开启交易或生产风险开关；
- 代码约束：所有数据库查询和写入位于 `services/market/models`。

## 2. 实现

### 2.1 协议和服务

- Market Admin RPC 新增 `SetAuthorityRegistry`、`ListAuthorityRegistries`；
- Admin API 新增：
  - `GET /admin/market/authorities`；
  - `POST /admin/market/authorities`；
- 支持按 Authority、provider code、producer type、snapshot kind、状态和游标查询；
- Authority 自动规范为小写，provider code、producer type 和 snapshot kind
  规范为大写；
- 允许类型固定为 `FINAL_QUOTE`、`INDEX`、`MARK`、`FUNDING`、`DELIVERY`。

### 2.2 更新保护

- `authority`、`provider_code` 和 `producer_type` 创建后不可修改；
- 更新必须携带当前 `version`，数据库原子执行 `version=version+1`；
- 停用 Authority 或移除已有类型前，model 查询所有激活价格公式；
- 仍被输出或成分引用时拒绝更新；
- 更新成功后清除主键和 Authority 唯一键缓存。

### 2.3 管理端

- 价格公式创建页删除硬编码的
  `market-ws / market-rest / price-engine`；
- 输出 Authority 按所选输出快照类型和 `allowed_kinds` 过滤；
- 成分 Authority 和成分快照类型直接取启用注册表；
- 默认优先选择允许 `FINAL_QUOTE` 的 `market-ws`，不存在时使用第一个合法来源；
- 页面展示 `Authority (provider code / producer type)`，明确供应商和传输通道。

### 2.4 独立供应商门禁

- Registry 表新增必填 `provider_code`；
- baseline-safe migration 将 `market-ws` 和 `market-rest` 同时回填为 `ITICK`，
  `price-engine` 回填为 `PRICE_ENGINE`；
- INDEX 和 DELIVERY 公式仍要求至少三个不同 Authority，并新增至少三个不同
  `provider_code` 的服务端校验；
- `contract-readiness` 的数据库检查同时返回启用来源数和
  `COUNT(DISTINCT provider_code)`，二者都必须与声明来源数完全相等；
- 相关数据库查询全部位于 Market models 或 deploy/dbinit models，logic 和 shell
  不直接执行 SQL。

### 2.5 RBAC

- 权限迁移：
  `20260729_add_market_authority_registry_permissions.sql`；
- migration 数：51；
- 菜单 483：`GET /market/authorities` / `market:authority:list`；
- 菜单 484：`POST /market/authorities` / `market:authority:set`；
- 迁移后仅删除管理员 `userId=1` 的可重建权限缓存
  `system:user:perms:1`，未清理其他 Redis 或业务数据。

## 3. 自动化验证

- `services/market go test ./...`：通过；
- `admin-api go test ./...`：通过；
- `admin-ui npm run type-check`（Node 20.20.2）：通过；
- `prettier --check`：通过；
- `deploy/dbinit go test ./...`：通过；
- `git diff --check`：通过。

单测覆盖：

- Authority、provider code、producer type 和 allowed kinds 规范化；
- 非法 token、空类型和未知类型拒绝；
- 同一 provider 的不同 Authority 不能组成 INDEX/DELIVERY 三源；
- 激活公式引用时禁止停用；
- 只增加 allowed kind 的非破坏性更新允许；
- 存量 JSON 损坏时拒绝更新。

## 4. 运行态验收

部署镜像：

- Market RPC：
  `sha256:5513ea58e5e2b073f370c26f68931017624ae009397e373ea305001efb61aaaf`；
- Admin API：
  `sha256:b2899d87d2009289372814025f550e69bc39d856a4ff0fd574a557a07cd4d641`。

两个容器均为 Healthy。部署切换时 Docker 空间耗尽令 Mongo 以 133 退出；只清理
2.135 GB 未被容器引用的悬空镜像层后，Mongo 从最后检查点成功恢复且数据卷保留。
Market 启动首个目标时点曾等待输入一次，随后 `market-ws` 恢复至亚秒级、
INDEX/MARK/FUNDING 持续生成；Outbox 仅保留约 1～2 秒的新鲜 Pending/Processing，
Failed/Manual 为 0，恢复后的检查窗口无 unhealthy、evaluation failed、slowcall、
panic 或 fatal。

查询结果：

| Authority | provider_code | producer_type | allowed_kinds |
| --- | --- | --- | --- |
| `market-ws` | `ITICK` | `ITICK_WS` | `FINAL_QUOTE` |
| `market-rest` | `ITICK` | `ITICK_REST` | `FINAL_QUOTE` |
| `price-engine` | `PRICE_ENGINE` | `PRICE_ENGINE` | `MARK/INDEX/FUNDING/DELIVERY` |

当前 `price-engine` 允许 `MARK/INDEX/FUNDING/DELIVERY`；两个 iTick 来源只允许
`FINAL_QUOTE`，但二者同属一个 `ITICK` 供应商，与现有运行事实一致。

无副作用保护测试：

1. 请求把 `market-ws` 从 Enabled 改为 Disabled；
2. API 返回代码 `100001` 和
   `authority is referenced by active price formulas`；
3. 再次查询确认 `status=1`、`version=0`、allowed kinds 未变化。

这同时验证了运行库的激活公式引用扫描和 MySQL `JSON_TABLE` 查询。

独立供应商负向保护测试：

1. 以 `market-ws`、`market-rest` 和一个第三 Authority 请求创建三成分 INDEX；
2. 前两个 Authority 均启用且允许 `FINAL_QUOTE`，但 `provider_code` 都是 `ITICK`；
3. API 返回代码 `100001` 和
   `INDEX and DELIVERY components must use independent providers`；
4. 查询确认测试公式记录数为 0，没有产生配置副作用。

预生产声明只读复跑结果保持 `14 prerequisite(s) failed`。核心服务、初始化器、
双生产开关、模型检查、Outbox、对账和结算水位均通过；失败项仍来自真实三供应商、
凭据/许可、资金审批、告警责任链和生产灾备材料，不是代码或部署故障。

## 5. 当前结论

Authority 配置链路已完成并可直接使用；真实第三来源到位后不需要改代码或手写 SQL。
当前虽然有两个外部 Authority，但 `provider_code` 都是 `ITICK`，只计一个独立外部
供应商。WebSocket/REST 已由服务端和只读门禁共同阻止冒充两个独立来源，因此生产
三源、凭据、许可、历史回放和 DELIVERY 公式门禁仍未完成，两个生产风险开关继续
保持关闭。
