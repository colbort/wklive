-- Settlement 指令与 Asset 流水自动对账。
ALTER TABLE `t_trade_settlement_instruction`
  ADD COLUMN `asset_flow_no` VARCHAR(64) NOT NULL DEFAULT ''
    COMMENT '对账确认的Asset流水号' AFTER `last_error_msg`,
  ADD COLUMN `reconciled_at` BIGINT NOT NULL DEFAULT 0
    COMMENT 'Asset流水对账完成时间' AFTER `asset_flow_no`,
  ADD INDEX `idx_settlement_reconcile`
    (`tenant_id`,`status`,`reconciled_at`,`id`);

CREATE TABLE IF NOT EXISTS `t_contract_reconciliation_issue` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT NOT NULL DEFAULT 0 COMMENT '租户ID',
  `issue_key` VARCHAR(160) NOT NULL COMMENT '稳定差异键',
  `check_type` VARCHAR(64) NOT NULL COMMENT '对账类型',
  `biz_type` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '业务类型',
  `biz_no` VARCHAR(96) NOT NULL DEFAULT '' COMMENT '业务单号',
  `instruction_id` BIGINT NOT NULL DEFAULT 0 COMMENT '结算指令ID',
  `expected_value` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '预期值摘要',
  `actual_value` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '实际值摘要',
  `detail` VARCHAR(1000) NOT NULL DEFAULT '' COMMENT '差异详情',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1待处理 2已恢复 3人工忽略',
  `occurrence_count` BIGINT NOT NULL DEFAULT 1 COMMENT '累计发现次数',
  `first_seen_at` BIGINT NOT NULL COMMENT '首次发现时间',
  `last_seen_at` BIGINT NOT NULL COMMENT '最近发现时间',
  `resolved_at` BIGINT NOT NULL DEFAULT 0 COMMENT '恢复时间',
  `operator_id` BIGINT NOT NULL DEFAULT 0 COMMENT '人工处理人',
  `resolution_reason` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '处理原因',
  `create_times` BIGINT NOT NULL COMMENT '创建时间',
  `update_times` BIGINT NOT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tenant_issue_key` (`tenant_id`,`issue_key`),
  KEY `idx_reconciliation_open` (`tenant_id`,`status`,`last_seen_at`,`id`),
  KEY `idx_reconciliation_instruction` (`tenant_id`,`instruction_id`),
  CONSTRAINT `chk_contract_reconciliation_issue`
    CHECK (`status` IN (1,2,3) AND `occurrence_count` > 0)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='合约跨服务对账差异';

CREATE TABLE IF NOT EXISTS `t_contract_reconciliation_cursor` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT NOT NULL DEFAULT 0 COMMENT '扫描租户；0表示全租户任务',
  `check_type` VARCHAR(64) NOT NULL COMMENT '对账类型',
  `last_scanned_id` BIGINT NOT NULL DEFAULT 0 COMMENT '当前轮次最后扫描的业务主键',
  `last_cycle_completed_at` BIGINT NOT NULL DEFAULT 0 COMMENT '最近完整轮次完成时间',
  `create_times` BIGINT NOT NULL COMMENT '创建时间',
  `update_times` BIGINT NOT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_reconciliation_cursor` (`tenant_id`,`check_type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='合约自动对账全量循环扫描游标';
