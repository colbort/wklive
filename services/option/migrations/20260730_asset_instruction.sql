ALTER TABLE `t_option_order`
  MODIFY COLUMN `status` TINYINT NOT NULL DEFAULT 0
  COMMENT '状态：0未知 1待撮合 2部分成交 3完全成交 4已撤单 5拒单 6已过期 7资金冻结中 8撤单资金处理中 9到期资金处理中';

CREATE TABLE IF NOT EXISTS `t_option_asset_instruction` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT NOT NULL DEFAULT 0 COMMENT '租户ID',
  `instruction_no` VARCHAR(96) NOT NULL DEFAULT '' COMMENT '资产指令号/幂等键',
  `biz_no` VARCHAR(96) NOT NULL DEFAULT '' COMMENT '业务单号',
  `order_id` BIGINT NOT NULL DEFAULT 0 COMMENT '订单ID',
  `trade_id` BIGINT NOT NULL DEFAULT 0 COMMENT '成交ID',
  `position_id` BIGINT NOT NULL DEFAULT 0 COMMENT '持仓ID',
  `user_id` BIGINT NOT NULL DEFAULT 0 COMMENT '用户ID',
  `account_id` BIGINT NOT NULL DEFAULT 0 COMMENT 'Option账户ID',
  `action` TINYINT NOT NULL DEFAULT 0 COMMENT '动作：1冻结 2扣冻结 3释放冻结 4可用入账 5可用扣减',
  `target_biz_no` VARCHAR(96) NOT NULL DEFAULT '' COMMENT '扣减或释放关联的原冻结业务号',
  `coin` VARCHAR(16) NOT NULL DEFAULT '' COMMENT '币种',
  `amount` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '金额，必须为正数',
  `step_no` INT NOT NULL DEFAULT 1 COMMENT '同一业务内执行顺序',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1待执行 2执行中 3成功 4失败 5人工处理 6未执行取消',
  `retry_count` INT NOT NULL DEFAULT 0 COMMENT '重试次数',
  `next_retry_at` BIGINT NOT NULL DEFAULT 0 COMMENT '下次重试时间',
  `last_error_msg` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '最后错误',
  `asset_flow_no` VARCHAR(64) NOT NULL DEFAULT '' COMMENT 'Asset实际流水号',
  `reconciliation_status` TINYINT NOT NULL DEFAULT 1 COMMENT '对账状态：1待对账 2一致 3不一致',
  `reconciled_at` BIGINT NOT NULL DEFAULT 0 COMMENT '对账完成时间',
  `create_times` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
  `update_times` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tenant_instruction_no` (`tenant_id`, `instruction_no`),
  KEY `idx_instruction_retry` (`tenant_id`, `status`, `next_retry_at`, `id`),
  KEY `idx_instruction_biz_step` (`tenant_id`, `biz_no`, `step_no`, `status`, `id`),
  KEY `idx_instruction_order` (`tenant_id`, `order_id`),
  KEY `idx_instruction_trade` (`tenant_id`, `trade_id`),
  CONSTRAINT `chk_option_asset_instruction` CHECK (
    `action` IN (1,2,3,4,5)
    AND `amount` > 0
    AND `step_no` > 0
    AND `status` IN (1,2,3,4,5,6)
    AND `retry_count` >= 0
    AND `reconciliation_status` IN (1,2,3)
  )
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Option发送给Asset的幂等资金指令';

CREATE TABLE IF NOT EXISTS `t_option_settlement_price` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT NOT NULL DEFAULT 0 COMMENT '租户ID',
  `contract_id` BIGINT NOT NULL DEFAULT 0 COMMENT '期权合约ID',
  `price_source` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '权威价格来源',
  `window_start` BIGINT NOT NULL DEFAULT 0 COMMENT '取价窗口开始时间',
  `window_end` BIGINT NOT NULL DEFAULT 0 COMMENT '取价窗口结束时间',
  `sample_count` BIGINT NOT NULL DEFAULT 0 COMMENT '有效样本数',
  `calculation_method` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '计算方法',
  `delivery_price` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '最终结算价',
  `source_snapshot_ids` TEXT NOT NULL COMMENT '原始快照依据',
  `version` BIGINT NOT NULL DEFAULT 1 COMMENT '结算价版本',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1等待价格 2已确认 3已拒绝',
  `confirmed_by` BIGINT NOT NULL DEFAULT 0 COMMENT '确认人，0为系统',
  `confirmed_at` BIGINT NOT NULL DEFAULT 0 COMMENT '确认时间',
  `create_times` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
  `update_times` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_settlement_price_contract` (`tenant_id`, `contract_id`),
  KEY `idx_settlement_price_status` (`tenant_id`, `status`, `id`),
  CONSTRAINT `chk_option_settlement_price` CHECK (
    `status` IN (1,2,3)
    AND `version` > 0
    AND (`status` <> 2 OR (`delivery_price` > 0 AND `sample_count` > 0 AND `confirmed_at` > 0))
  )
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='期权不可变结算价快照';

CREATE TABLE IF NOT EXISTS `t_option_settlement_batch` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `tenant_id` BIGINT NOT NULL DEFAULT 0,
  `batch_no` VARCHAR(96) NOT NULL DEFAULT '',
  `contract_id` BIGINT NOT NULL DEFAULT 0,
  `settlement_price_id` BIGINT NOT NULL DEFAULT 0,
  `total_credit` DECIMAL(32,16) NOT NULL DEFAULT 0,
  `total_debit` DECIMAL(32,16) NOT NULL DEFAULT 0,
  `instruction_count` BIGINT NOT NULL DEFAULT 0,
  `success_count` BIGINT NOT NULL DEFAULT 0,
  `status` TINYINT NOT NULL DEFAULT 1,
  `last_error_msg` VARCHAR(500) NOT NULL DEFAULT '',
  `create_times` BIGINT NOT NULL DEFAULT 0,
  `update_times` BIGINT NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_settlement_batch_no` (`tenant_id`, `batch_no`),
  UNIQUE KEY `uk_settlement_batch_contract` (`tenant_id`, `contract_id`),
  KEY `idx_settlement_batch_status` (`tenant_id`, `status`, `id`),
  CONSTRAINT `chk_option_settlement_batch` CHECK (
    `status` IN (1,2,3,4,5,6,7,8)
    AND `total_credit` >= 0 AND `total_debit` >= 0
    AND `instruction_count` >= 0 AND `success_count` >= 0
    AND `success_count` <= `instruction_count`
  )
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='期权结算批次';

CREATE TABLE IF NOT EXISTS `t_option_settlement_detail` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `tenant_id` BIGINT NOT NULL DEFAULT 0,
  `batch_id` BIGINT NOT NULL DEFAULT 0,
  `batch_no` VARCHAR(96) NOT NULL DEFAULT '',
  `contract_id` BIGINT NOT NULL DEFAULT 0,
  `position_id` BIGINT NOT NULL DEFAULT 0,
  `user_id` BIGINT NOT NULL DEFAULT 0,
  `account_id` BIGINT NOT NULL DEFAULT 0,
  `side` TINYINT NOT NULL DEFAULT 0,
  `quantity` DECIMAL(32,16) NOT NULL DEFAULT 0,
  `payoff` DECIMAL(32,16) NOT NULL DEFAULT 0,
  `direction` TINYINT NOT NULL DEFAULT 0,
  `instruction_no` VARCHAR(96) NOT NULL DEFAULT '',
  `create_times` BIGINT NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_settlement_detail_position` (`tenant_id`, `batch_id`, `position_id`),
  KEY `idx_settlement_detail_user` (`tenant_id`, `user_id`, `account_id`, `id`),
  CONSTRAINT `chk_option_settlement_detail` CHECK (
    `side` IN (1,2) AND `quantity` >= 0 AND `payoff` >= 0 AND `direction` IN (1,2,3)
  )
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='期权逐持仓结算明细';

CREATE TABLE IF NOT EXISTS `t_option_reconciliation_issue` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `tenant_id` BIGINT NOT NULL DEFAULT 0,
  `issue_key` VARCHAR(160) NOT NULL DEFAULT '',
  `check_type` TINYINT NOT NULL DEFAULT 0,
  `biz_no` VARCHAR(96) NOT NULL DEFAULT '',
  `instruction_id` BIGINT NOT NULL DEFAULT 0,
  `expected_value` VARCHAR(500) NOT NULL DEFAULT '',
  `actual_value` VARCHAR(500) NOT NULL DEFAULT '',
  `detail` VARCHAR(1000) NOT NULL DEFAULT '',
  `status` TINYINT NOT NULL DEFAULT 1,
  `occurrence_count` BIGINT NOT NULL DEFAULT 1,
  `resolved_at` BIGINT NOT NULL DEFAULT 0,
  `create_times` BIGINT NOT NULL DEFAULT 0,
  `update_times` BIGINT NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_option_reconciliation_issue` (`tenant_id`, `issue_key`),
  KEY `idx_option_reconciliation_status` (`tenant_id`, `status`, `id`),
  KEY `idx_option_reconciliation_instruction` (`tenant_id`, `instruction_id`),
  CONSTRAINT `chk_option_reconciliation_issue` CHECK (
    `check_type` IN (1,2,3) AND `status` IN (1,2,3) AND `occurrence_count` > 0
  )
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Option与Asset对账差异';
