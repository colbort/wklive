# Option 发布合约上市检查表合集

OPTION_LAUNCH_CHECKLIST_STATUS: DRAFT
OPTION_LAUNCH_TENANT_ID: 0
OPTION_LAUNCH_RELEASE_COMMIT: DRAFT
OPTION_LAUNCH_CONTRACT_COUNT: 0
OPTION_LAUNCH_PRODUCT_APPROVAL_REF: DRAFT
OPTION_LAUNCH_TECH_APPROVAL_REF: DRAFT
OPTION_LAUNCH_RISK_APPROVAL_REF: DRAFT
OPTION_LAUNCH_CLEARING_APPROVAL_REF: DRAFT
OPTION_LAUNCH_OPERATIONS_APPROVAL_REF: DRAFT
OPTION_LAUNCH_COMPLIANCE_APPROVAL_REF: DRAFT

> 每次发布使用一份。单合约发布也必须列出唯一合约；多合约发布不得只提供一个代表性检查表。
> 每个明细必须引用已签署的`option-contract-launch-checklist.md`副本及SHA-256。本文件签署后不得改写。

## 1. 发布身份

| 字段 | 值 |
| --- | --- |
| tenant_id / 法律实体 | 待填写 |
| release commit / image digest | 待填写 |
| 发布变更单 | 待填写 |
| 计划窗口（UTC） | 待填写 |
| 合约总数 | 待填写 |
| 系列 ID/版本（无则不适用） | 待填写 |
| 系列审批记录及SHA-256（无则不适用） | 待填写 |

## 2. 合约明细（机器可校验）

在本节末尾每个合约写一行，不得写表头、汇总行或抽样行：

`OPTION_LAUNCH_CONTRACT: <contract_id>|<contract_code>|<absolute_checklist_path>|<checklist_sha256>|APPROVED`

约束：

- `OPTION_LAUNCH_CONTRACT_COUNT`必须是正整数，且等于本节记录行数。
- `contract_id`、`contract_code`和检查表路径分别唯一；路径必须是生产门禁可读的绝对路径。
- 每份检查表必须只声明一组`OPTION_CONTRACT_TENANT_ID`、`OPTION_CONTRACT_ID`和
  `OPTION_CONTRACT_CODE`，并与本行及发布租户完全一致。
- 合集和每份单合约检查表都必须填六类机器可读审批引用；不接受空值、`DRAFT`、`PENDING`或`REJECTED`。
- SHA-256必须是对该单合约归档副本计算的64位十六进制值，状态必须为`APPROVED`。
- 记录行不得放在Markdown表格中，不得在字段内使用`|`或换行。

模板保持记录数为0且不放示例记录，防止示例被误当成真实合约。生产副本必须填入至少一行真实记录。

## 3. 完整性检查

- [ ] 机器记录行数、`OPTION_LAUNCH_CONTRACT_COUNT`和“合约总数”三者一致，`contract_id`、`contract_code`和路径均无重复。
- [ ] 本次release/系列生成的全部合约都在表内，没有抽样、截断或额外未审批合约。
- [ ] 每个单合约检查表包含`OPTION_LAUNCH_CHECKLIST_STATUS: APPROVED`，文件非空且SHA-256匹配。
- [ ] 若来自系列，系列记录为`OPTION_CONTRACT_SERIES_APPROVAL_STATUS: APPROVED`，生成明细与本表一一对应。
- [ ] 所有合约分别核对上市、最后交易、行权截止、到期和交割五个时间，不共用模糊时间描述。
- [ ] 每个合约的结算、行情、限额、费用、保证金、保险/兜底和条件功能与签署配置一致。
- [ ] 被拒绝、撤回、延后或不在本次范围的合约没有进入渲染配置、调度消息或公告。
- [ ] 最终渲染配置、公告合约列表和本表使用相同合同集合，并分别固定SHA-256。

## 4. 签署

| 角色 | 姓名/ID | 结论 | UTC时间 | 工单/签名 |
| --- | --- | --- | --- | --- |
| 产品 | 待填写 | 通过/拒绝 | 待填写 | 待填写 |
| Option技术 | 待填写 | 通过/拒绝 | 待填写 | 待填写 |
| 风控 | 待填写 | 通过/拒绝 | 待填写 | 待填写 |
| 清算/财务 | 待填写 | 通过/拒绝 | 待填写 | 待填写 |
| 运营 | 待填写 | 通过/拒绝 | 待填写 | 待填写 |
| 合规/法务 | 待填写 | 通过/拒绝 | 待填写 | 待填写 |

只有全部合约逐项批准、集合一致、哈希匹配且六方签署通过时，归档副本才允许改为
`OPTION_LAUNCH_CHECKLIST_STATUS: APPROVED`。任何合约缺失、重复、拒绝或哈希失配时必须保持`DRAFT`
或改为`REJECTED`。
