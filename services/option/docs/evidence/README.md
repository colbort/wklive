# Option 证据目录

本目录保存可由仓库工具复核、但不能替代目标环境或人员签署的技术证据。

- `REPOSITORY_ONLY`：只证明当前工作区代码和隔离测试结果。
- `PREPRODUCTION`：必须来自目标预生产部署，包含不可变 release/image、真实配置、人员、时间线和外部系统证据。
- `PRODUCTION`：必须关联正式变更单、生产只读证据和六方签署。

仓库证据不得复制后改名为预生产证据。每份证据必须明确状态、工作区是否干净、命令、结果、未覆盖范围和
SHA-256 清单；清单校验失败时证据立即失效。

当前证据：

- `option-repository-technical-evidence-20260802.md`
- `option-repository-technical-evidence-20260802.sha256`
- `../option-p0-007-repository-acceptance.md`（平台兜底仓库验收）
- `../templates/option-platform-backstop-e2e.md`（目标环境待填模板，不是生产证据）
- `../templates/option-production-evidence-report.md`（无专项模板时的目标环境报告底稿，默认`DRAFT`）
- `../templates/option-contract-launch-bundle.md`（本次发布完整合约集合及逐表哈希，默认`DRAFT`）
- `../../monitoring/option-launch-bundle-verify.sh`（逐行校验合约数量、唯一身份、六方审批、完整63项检查表和每份哈希）
- `../../monitoring/option-launch-bundle-verify-selftest.sh`（可重复执行一条正常及五条反向门禁样例）
- `../templates/option-contract-set-reconciliation.md`（审批合集、目标环境导出和最终公告/客户端集合三方对账，默认`DRAFT`）
- `../option-operations-input-checklist.md`（运营真实输入、责任方与专项证据入口，默认`DRAFT`）
- `../option-external-input-handoff.md`（外部责任方首批最小资料、完整合约CSV字段和收到后的执行顺序）
- `../../monitoring/option-external-contract-intake-verify.sh`及自测（完整合约CSV的22列schema、唯一身份、
  枚举、币种、正数经济参数和五时间顺序）
- `../../monitoring/option-operations-input-verify.sh`及自测（31项固定材料逐条文件/SHA/终态、逐合约输入、
  完成项、严重问题和由底层记录反算的总数守恒）
- `../../monitoring/option-evidence-finalization-verify.sh`及自测（拒绝报告表格占位符、空单元格、
  未决选择、OPEN/DRAFT行和未勾选验收项）
- `../../monitoring/option-readiness-attestation-verify.sh`及自测（生产声明全部键精确一次、无未知键、
  无缺键和占位路径）
- `../../etc/option.yaml`与`../../internal/logic/helpers/product_scope.go`（可选产品范围零值关闭及入口门禁）
- `../../monitoring/option-product-scope-verify.sh`及自测（发布声明与渲染配置逐项一致、唯一且无缺失）
- `../templates/option-mmp-readiness.md`与`option-american-exercise-readiness.md`（条件启用专项目标环境模板）

生产门禁不再只接受“文件存在+SHA-256”：归档副本还必须包含与证据类型匹配的机器可读
`APPROVED`状态，并通过统一终态校验；把状态行改为`APPROVED`但仍保留表格占位符、空值、未决选择、
OPEN/DRAFT行或未勾选验收项同样失败。仓库中的模板均保持`DRAFT`，不能直接作为生产证据。
生产声明文件本身也以`option-production-readiness.env.example`为键集合schema；重复键、拼错后形成的
未知键、缺键、`export`等非规范赋值及仓库示例占位路径都会在读取前失败，避免不同工具取不同值。

上市合集不能只靠人工查看。生产门禁会执行`option-launch-bundle-verify.sh`，逐份打开归档检查表并验证
租户、release、contract ID/code、六类审批引用、模板正文、63项勾选和SHA-256；随后把审批合集与
目标环境只读导出、最终公告/客户端导出做严格等集比较。声明数量不一致、重复、漏项、改写、未勾选或
三方多/少任一合约均拒绝。

生产功能声明也不能只停留在证据文件。门禁要求提交目标环境Option渲染YAML和SHA-256，并把卖方、组合
保证金、实物、复杂单、公开行情、MMP和美式行权逐项与`ProductScope`比较；没有独立运行时入口的
Greeks依赖扩展保持`false`。重复/缺失的环境声明、YAML段或键一律拒绝。
