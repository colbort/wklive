# Option 发布候选变更清单（2026-08-02）

INVENTORY_STATUS: REVIEW_REQUIRED_DIRTY_WORKTREE

## 1. 快照结论

加入本清单和范围检查脚本后，工作区预计有172个精确文件状态：117个已跟踪修改、55个未跟踪文件。
当前没有删除、重命名或冲突状态；全部路径位于明确的Option发布范围，但尚未逐文件完成人工代码审查，
因此不能直接提交、构建发布镜像或把仓库证据升级为release candidate。

可重复检查：

```bash
services/option/monitoring/option-release-scope.sh --scope-only
```

形成发布提交后必须返回：

```bash
services/option/monitoring/option-release-scope.sh --release-clean
```

## 2. 模块分布

| 模块 | 已跟踪修改 | 未跟踪 | 发布作用 |
| --- | ---: | ---: | --- |
| `services/option` | 98 | 46 | 核心运行时、模型、迁移、测试、门禁和文档 |
| `proto/option` | 7 | 0 | RPC枚举、字段、接口及生成代码 |
| `admin-api` | 3 | 8 | Option管理路由、类型及保险库存退出代理 |
| `admin-ui` | 8 | 0 | 合约/系列/风险工作台及运营标签 |
| `services/system` | 0 | 1 | 保险库存退出权限迁移 |
| `init.sql` | 1 | 0 | 空库聚合schema同步 |

## 3. 必须作为同一发布范围审查的依赖组

1. Schema/迁移/模型：`option.sql`、`schema/90_constraints.sql`、5个20260802迁移、生成模型及
   `init.sql`必须一致；DDL变更后已经执行`make gen-model`，后续任何DDL调整都必须重新执行。
2. Proto/服务：`proto/option/*.proto`、`*.pb.go`、Option client/server、Admin API类型和UI请求字段必须
   同版本发布。禁止只挑选proto或只挑选调用端。
3. 生命周期：独立`last_trade_time`、合约系列五时间、延迟队列、下单/撮合/恢复/到期屏障和管理UI
   是一个不变量组，不能拆成可产生旧时间语义的中间发布。
4. 风险/清算：组合版本谱系、订单准入版本、部分清算、保险库存退出和资金流水方向依赖同一套迁移、
   Asset业务号和真实RPC门禁。
5. 运营：指标SQL、告警、readiness、Admin UI标签、对账/签署模板和证据哈希必须在最终代码审查后重生；
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

1. 先冻结工作区，不再混入其他业务改动；导出完整diff和未跟踪文件清单。
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
- 不把未跟踪测试、迁移、权限或验收文档遗漏在提交之外。
- 不以当前HEAD加脏工作区哈希替代最终release commit。
- 不在未完成财务/清算、模型、SRE、产品和法务签署时开启条件能力。
