# Option 发布候选变更清单（2026-08-02）

INVENTORY_STATUS: DYNAMIC_USE_SCOPE_SCRIPT

## 1. 快照结论

原172项Option范围变化已在审计期间由外部提交为
`c73eb294e8ace983fc07f851b39f3a097334904d`；范围脚本分类修正随后进入
`7a9ae4d9e8671fd0c2a55e03086ae4722cd687b1`。平台兜底、证据门禁和运营材料又由外部提交为
`7d81ffcb43b2dc4d1adab70dca3548e4ad8ad191`，当前同时位于`master`和`origin/master`。
Codex未执行这些提交，也未改写索引。

仓库在审计期间可能继续生成检查点提交，因此本文件不把某个脏文件数量声明为永久事实。每次评审必须以
`git rev-parse HEAD`和范围脚本输出为准。`7d81ffcb`提交后，本轮继续补强完整合约上市合集、运营输入总表
及其生产门禁；随后增加逐合约合集校验器，生产门禁不再以“出现一个哈希”推断全集完成。
随后又补齐审批、目标环境和公告/客户端三方合约集合对账，并把原先只存在于验收文件的产品能力声明
实现为默认关闭的Option `ProductScope`运行时门禁；生产证据又增加统一终态校验，不能只改批准状态行而
保留空表格或未决字段；readiness声明也按完整schema拒绝重复、未知、缺失键和占位路径。
外部责任方首批输入又收敛为一份最小移交包，并提供完整合约CSV机器校验器，明确能力开关和收到资料后的自动执行顺序。
2026-08-02合成数据复验及权威计时去漂移后，当前脚本输出
`changed=52 modified=32 added=0 untracked=20`并为`SCOPE_OK`。
已提交的平台兜底范围明确包含P0-007不可拆分的Asset schema/Proto/服务/模型、Admin API/UI和System
RBAC依赖；脚本只放行这些精确文件或平台兜底专用命名模式，不会把整个Asset或Admin目录泛化为Option
范围。当前52项均在`services/option`内，没有删除、重命名或冲突；形成下一提交后仍须以
`--release-clean`证明最终工作区为0变化。

可重复检查：

```bash
services/option/monitoring/option-release-scope.sh --scope-only
```

形成最终后续提交后必须返回：

```bash
services/option/monitoring/option-release-scope.sh --release-clean
```

## 2. 初始集成提交内容分布

| 模块 | 已暂存修改 | 已暂存新增 | 发布作用 |
| --- | ---: | ---: | --- |
| `services/option` | 98 | 46 | 核心运行时、模型、迁移、测试、门禁和文档 |
| `proto/option` | 7 | 0 | RPC枚举、字段、接口及生成代码 |
| `admin-api` | 3 | 8 | Option管理路由、类型及保险库存退出代理 |
| `admin-ui` | 8 | 0 | 合约/系列/风险工作台及运营标签 |
| `services/system` | 0 | 1 | 保险库存退出权限迁移 |
| `init.sql` | 1 | 0 | 空库聚合schema同步 |

上表是初始Option集成提交分布，不是当前脏工作区计数；P0-007跨服务依赖以范围脚本当前输出和第3节为准。

## 3. 必须作为同一发布范围审查的依赖组

1. Schema/迁移/模型：`option.sql`、`schema/90_constraints.sql`、5个20260802迁移、生成模型及
   `init.sql`必须一致；DDL变更后已经执行`make gen-model`，后续任何DDL调整都必须重新执行。
2. Proto/服务：`proto/option/*.proto`、`*.pb.go`、Option client/server、Admin API类型和UI请求字段必须
   同版本发布。禁止只挑选proto或只挑选调用端。
3. 生命周期：独立`last_trade_time`、合约系列五时间、延迟队列、下单/撮合/恢复/到期屏障和管理UI
   是一个不变量组，不能拆成可产生旧时间语义的中间发布。
4. 风险/清算：组合版本谱系、订单准入版本、部分清算、保险库存退出和资金流水方向依赖同一套迁移、
   Asset业务号和真实RPC门禁。
5. 平台兜底：Option运行时开关、上市/恢复/清算门禁、Asset硬额度实现、管理RPC/API/UI、System RBAC、
   政策审批、BST-001～BST-012、渲染配置和生产readiness必须作为一个依赖组；仓库实现与隔离验收已完成，
   真实资本、参数、目标环境证据和签署未完成前禁止开启。
6. 可选产品范围：`ProductScope`配置、上市/恢复、普通/组合订单、组合资金后激活、公开行情、美式行权、
   readiness声明和渲染配置哈希是一个依赖组；未批准能力不得只在文档中写false而运行时放开。
7. 运营：指标SQL、告警、readiness、Admin UI标签、运营输入31项固定材料逐条哈希/完成项/汇总反算、
   证据终态校验、
   readiness声明schema/唯一性校验、
   对账/签署模板和证据哈希
   必须在最终代码审查后重生；
   不能先固定旧哈希再修改实现。

## 4. 生成产物与命令

按最终源文件执行，执行后再次审查差异：

```bash
cd services/option
make gen
make gen-model

cd ../../admin-api
make gen

cd ../admin-ui
PATH=/Users/sky/.nvm/versions/node/v20.20.2/bin:$PATH npx prettier --write \
  src/i18n/locales/zh-CN.ts src/i18n/locales/en-US.ts
PATH=/Users/sky/.nvm/versions/node/v20.20.2/bin:$PATH npm run type-check
```

`make gen`和`make gen-model`不是可选备注。生成后如果出现源协议/DDL未解释的广泛漂移，必须停止并审查
生成器版本（当前要求goctl 1.9.2），不能直接全部纳入提交。

## 5. 建议审查/提交顺序

1. 先冻结工作区，不再混入其他业务改动；审查`c73eb294..HEAD`完整提交链以及范围脚本报告的当前diff。
   “已经提交”不等于已经获得生产审批。
2. 审查DDL、迁移可重复性、历史兼容和触发器，再审查proto的向后兼容性。
3. 审查核心资金/持仓/生命周期逻辑及其测试；对每个数据库字段找到写入方、读取方和验收用例。
4. 审查Admin API/UI/权限，确认未批准能力默认关闭且租户隔离不被绕过。
5. 最后审查运营文档和证据，重新计算SHA-256；文档不得引用已变化的旧测试结果。
6. 可以使用逻辑提交便于复核，但合并/部署必须使用完整集成提交；不得单独cherry-pick生成代码、迁移或UI。

## 6. 提交前强制门禁

- Option：`go test ./...`、`go vet ./...`、相关包`go test -race`。
- Admin API：`go test ./...`、`go vet ./...`。
- Admin UI：Prettier check和`npm run type-check`。
- `make gen`、`make gen-model`后工作区只有经解释的生成差异。
- `option-production-readiness.sh --repository-only`通过，SHA-256清单全部OK。
- `run-p0-asset-rpc-e2e.sh`在最终集成提交上重新通过。
- `option-release-scope.sh --release-clean`在提交后通过；记录最终commit和全部镜像digest。

## 7. 明确排除和禁止操作

- 当前未发现Option范围外路径；若脚本后续报告新路径，必须单独归属，不能自动并入Option发布。
- 不擅自`git reset`、丢弃、覆盖或提交现有用户改动。
- 不把新增测试、迁移、权限或验收文档遗漏在最终审查之外。
- 不以检查点提交加脏工作区哈希替代最终release commit；发布身份只能取最终干净HEAD和镜像digest。
- 不在未完成财务/清算、模型、SRE、产品和法务签署时开启条件能力。
