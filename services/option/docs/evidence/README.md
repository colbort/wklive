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

生产门禁不再只接受“文件存在+SHA-256”：归档副本还必须包含与证据类型匹配的机器可读
`APPROVED`状态。仓库中的模板均保持`DRAFT`，不能直接作为生产证据。
