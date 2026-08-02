# Option 发布合约集合对账证明

OPTION_CONTRACT_SET_RECONCILIATION_STATUS: DRAFT
OPTION_CONTRACT_SET_TENANT_ID: 0
OPTION_CONTRACT_SET_RELEASE_COMMIT: DRAFT
OPTION_CONTRACT_SET_COUNT: 0
OPTION_CONTRACT_SET_WINDOW_START_MS: 0
OPTION_CONTRACT_SET_WINDOW_END_MS: 0
OPTION_TARGET_CONTRACT_SET_SOURCE: DRAFT
OPTION_TARGET_CONTRACT_SET_SOURCE_SHA256: DRAFT
OPTION_PUBLICATION_CONTRACT_SET_SOURCE: DRAFT
OPTION_PUBLICATION_CONTRACT_SET_SOURCE_SHA256: DRAFT
OPTION_CONTRACT_SET_REVIEW_REF: DRAFT

> 每次发布使用一份。用与最终发布相同的租户和 release，从目标环境 Admin RPC/DB 只读
> 导出“本次计划上市”合约，再从最终公告/客户端配置导出对外合约集。两份原始导出均必须
> 归档并固定SHA-256；本文件不能代替原始导出。

## 1. 数据来源

| 字段 | 值 |
| --- | --- |
| tenant_id / 法律实体 | 待填写 |
| release commit / image digest | 待填写 |
| 本次上市窗口`[start_ms,end_ms)` UTC毫秒 | 待填写 |
| 目标环境合约导出方式、UTC时间、文件及SHA-256 | 待填写 |
| 最终公告/客户端配置导出方式、UTC时间、文件及SHA-256 | 待填写 |
| 独立复核工单/签名 | 待填写 |

## 2. 原始导出文件格式

`OPTION_TARGET_CONTRACT_SET_SOURCE`和`OPTION_PUBLICATION_CONTRACT_SET_SOURCE`必须是门禁机器可读的
绝对路径。两份文件都不带表头。

目标环境只读导出每个合约一行：

`<contract_id>|<contract_code>|<status>|<list_time_ms>`

`status`必须为`1`（`PENDING`），`list_time_ms`必须落在已批准的`[start_ms,end_ms)`。
建议由发布工作站使用只读账号执行等价查询，查询不要加`status=1`过滤，让门禁能够发现提前
进入交易态的合约：

```sql
SELECT CONCAT(id, '|', contract_code, '|', status, '|', list_time)
FROM t_option_contract
WHERE tenant_id = :tenant_id
  AND is_deleted = 2
  AND list_time >= :window_start_ms
  AND list_time < :window_end_ms
ORDER BY id;
```

最终公告/客户端配置导出每个合约一行：

`<contract_id>|<contract_code>`

`contract_id`必须为正整数，代码非空且不含空白；ID和代码在各自文件中均唯一。不得手工从
导出中删行，不得把Markdown表格或示例行当成原始数据。

## 3. 验收与签署

- [ ] 上市审批合集、目标环境导出和最终公告/客户端集合的数量相同。
- [ ] 三份集合的`contract_id|contract_code`排序后逐行完全相同，不存在重复、额外或缺失。
- [ ] 目标环境记录仍处于本次批准的上市前状态，没有未经批准提前进入`TRADING`。
- [ ] 两份原始导出的租户、release、时间和SHA-256可重算，导出后无手工删行。
- [ ] 独立复核人未执行原始导出，并已复核门禁的完整PASS输出。

| 角色 | 姓名/ID | 结论 | UTC时间 | 工单/签名 |
| --- | --- | --- | --- | --- |
| 技术/发布执行人 | 待填写 | 通过/拒绝 | 待填写 | 待填写 |
| 产品/运营复核人 | 待填写 | 通过/拒绝 | 待填写 | 待填写 |

只有三份集合完全相同、原始导出哈希可重算且独立复核通过时，归档副本才允许改为
`OPTION_CONTRACT_SET_RECONCILIATION_STATUS: APPROVED`。任何集合不一致时必须保持`DRAFT`或改为`REJECTED`。
