# Option 外部输入最小移交包

HANDOFF_STATUS: WAITING_FOR_EXTERNAL_INPUT

更新时间：2026-08-02

## 1. 用途

仓库实现、验收标准、运营模板和失败关闭门禁已经完成。本文件只说明外部责任方需要先交付什么，收到后
技术侧才能继续生成真实合约材料、执行目标环境验收并关闭生产阻断。它不是批准文件，不能把示例、测试
tenant、本机容器结果或仓库模板改名后当成生产输入。

禁止在本文件或聊天中发送密码、Token、私钥、私人电话、数据库凭证或未脱敏客户数据。敏感值只进入
批准的秘密管理、值班和工单系统；这里保存其非敏感引用编号。

## 2. 第一批必须提供的最小资料

| 序号 | 输入 | 最低完整内容 | 责任方 | 收到后的自动动作 |
| ---: | --- | --- | --- | --- |
| 1 | 产品主体 | `tenant_id`、法律实体、司法辖区、目标用户地区 | 产品/合规 | 绑定全部检查表、租户隔离和签署身份 |
| 2 | 目标市场 | 市场名称、官方规则版本/链接、IANA时区、年度日历原始文件及SHA-256 | 市场运营/合规 | 生成日历审批、年度复核和五时间校验输入 |
| 3 | 完整合约集合 | 本次全部`contract_code`，不得给代表性样本；逐合约Call/Put、欧/美式、现金/实物、标的及三币种 | 产品/运营 | 生成上市合集、逐合约检查表和三方集合对账基线 |
| 4 | 逐合约经济参数 | strike、unit、multiplier、tick、qty step、min/max qty、费用、保证金、价格带、持仓/OI上限 | 产品/风控/财务 | 填充逐合约10组参数并运行数量/步长/币种校验 |
| 5 | 五个时间 | list、last trade、exercise cutoff、expire、deliver的UTC毫秒、本地带时区文本和IANA时区 | 产品/市场运营 | 校验`list < last_trade <= cutoff <= expire <= deliver`并生成公告时间表 |
| 6 | 结算与行权规则 | 正式结算价来源/窗口/样本/精度/更正，AUTO/DNE阈值、通知截止和费用后正收益规则 | 行情/清算/风控 | 填充结算价、行权到期和通知验收材料 |
| 7 | 产品能力范围 | 卖方、平台兜底、组合保证金、实物、复杂单、公开行情、MMP、美式提前行权、保险库存退出逐项`true/false` | 产品/风控/清算 | 生成readiness声明，与渲染YAML逐项比对；未批项保持false |
| 8 | 发布与环境身份 | 目标环境、计划release commit/全部image digest、变更单、窗口、回滚窗口、metrics URL及编排器类型 | 技术/SRE | 固定不可变release并准备迁移、监控、故障和容量执行包 |
| 9 | 资金与兜底决策 | 保险语义批准；平台兜底选择DISABLED/PREFUNDED/CREDIT_FLOOR，资本主体和逐币硬额度 | 财务/清算/风控/管理层 | 填充保险、日终和平台政策；未批准时卖方/兜底失败关闭 |
| 10 | 角色与工单引用 | 产品、技术、风控、清算财务、运营、合规法务的角色姓名/工单引用；联系方式留在安全系统 | 各责任方 | 生成六方签署框架和31项固定材料记录，不代签 |

第一批可以分两次交付：先给1～7以冻结产品和合约范围，再给8～10启动目标环境验收。缺少1～7时不得
先造合约或猜市场规则；缺少8～10时可以生成DRAFT材料，但不能改为APPROVED。

## 3. 合约集合的最小交换格式

优先提供UTF-8 CSV；一行一个合约，时间均为Unix毫秒且另附本地带时区文本。至少保留以下列：

```text
contract_code,underlying,call_put,exercise_style,settlement_type,quote_coin,settle_coin,underlying_coin,strike,contract_unit,multiplier,price_tick,qty_step,min_qty,max_qty,list_time_ms,last_trade_time_ms,exercise_cutoff_time_ms,expire_time_ms,deliver_time_ms,iana_timezone,series_id
```

要求：

- `contract_code`唯一且不能为空；CSV必须覆盖本次发布全集。
- 币种必须使用系统规范代码，不使用“同上”“默认”或空值。
- 无系列时`series_id`明确写`NOT_APPLICABLE`并附批准依据，不能留空。
- CSV原文件计算SHA-256后归档；后续任何修改产生新版本和新哈希。
- 公告/客户端集合和目标环境只读导出最终必须与该批准集合严格相等。

收到CSV后先执行：

```sh
services/option/monitoring/option-external-contract-intake-verify.sh /secure/path/contracts.csv
```

校验器拒绝表头漂移、空字段、占位值、重复代码、非法枚举/币种、非正经济参数、`min_qty > max_qty`、
非毫秒时间、五时间乱序和无效IANA时区。CSV只允许不含逗号的简单字段；需要备注或带逗号文本时另附
哈希文档，不在该交换文件中使用带引号的复杂CSV单元格。

## 4. 能力范围的回复格式

必须逐项给出小写`true`或`false`，不能写“暂定”“默认开启”或只给一个总开关：

```text
OPTION_SELLER_TRADING_ENABLED=false
OPTION_PLATFORM_BACKSTOP_ENABLED=false
OPTION_PORTFOLIO_MARGIN_ENABLED=false
OPTION_PHYSICAL_DELIVERY_ENABLED=false
OPTION_COMPLEX_ORDERS_ENABLED=false
OPTION_PUBLIC_MARKET_ENABLED=false
OPTION_GREEKS_DEPENDENT_FEATURES_ENABLED=false
OPTION_MMP_ENABLED=false
OPTION_AMERICAN_EXERCISE_ENABLED=false
OPTION_INSURANCE_INVENTORY_EXIT_ENABLED=false
```

Greeks依赖扩展当前没有独立运行时入口，只能为`false`。组合保证金、实物、复杂单、MMP和美式提前行权
依赖卖方交易批准；平台兜底还依赖准确Asset政策ID/版本和BST-001～012。

## 5. 收到资料后的执行顺序

1. 将输入复制到仓库外受控工作目录，固定原文件SHA-256和工单引用。
2. 生成本次全部逐合约`option-contract-launch-checklist.md`及`option-contract-launch-bundle.md`，运行
   `option-launch-bundle-verify.sh`。
3. 生成目标环境和公告/客户端两份只读集合导出，填写`option-contract-set-reconciliation.md`并做等集校验。
4. 填写运营输入总表的31条`OPTION_OPERATIONS_MATERIAL`记录；每条绑定绝对文件路径、SHA和终态，运行
   `option-operations-input-verify.sh`，汇总由底层记录反算。
5. 渲染Option YAML，填写readiness声明，依次运行`option-readiness-attestation-verify.sh`、
   `option-product-scope-verify.sh`和`option-evidence-finalization-verify.sh`。
6. 在相同release/image上执行迁移、真实Asset、监控通知、编排故障、Beanstalk容量、数据库审计、日终和
   条件能力专项验收；差异为0且所有事件关闭后再完成六方签署。
7. 最后执行生产模式`option-production-readiness.sh`；零失败也只允许在批准窗口按租户、合约和单项能力渐进开放。

## 6. 完成判定

外部移交只有同时满足以下条件才算可执行：

- 输入覆盖发布全集，字段无空值、占位符、未确定时区或代表性抽样。
- 原始文件、审批、渲染配置和目标环境证据均有不可变版本及SHA-256。
- 经济/法律决策由对应责任方明确批准，技术人员没有代填资本、规则或签名。
- 未批准能力在运行时、API、权限、公告、客服和销售口径中全部保持关闭。
- 所有材料使用同一个tenant、release和合约集合；任何不一致先停止发布并重新对账。

收到第2节资料后，仓库侧才有新的可执行工作；在此之前状态保持
`REPOSITORY_PASSED / PREPROD_BLOCKED`。
