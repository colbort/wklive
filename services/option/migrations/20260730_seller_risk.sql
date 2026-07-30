ALTER TABLE `t_option_contract`
  ADD COLUMN `seller_margin_mode` TINYINT NOT NULL DEFAULT 1 COMMENT '卖方保证金模式：1关闭 2逐仓 3组合' AFTER `fee_account_id`,
  ADD COLUMN `initial_margin_rate` DECIMAL(20,10) NOT NULL DEFAULT 0 COMMENT '卖方初始保证金率' AFTER `seller_margin_mode`,
  ADD COLUMN `maintenance_margin_rate` DECIMAL(20,10) NOT NULL DEFAULT 0 COMMENT '卖方维持保证金率' AFTER `initial_margin_rate`,
  ADD COLUMN `min_margin_rate` DECIMAL(20,10) NOT NULL DEFAULT 0 COMMENT '卖方最低保证金率' AFTER `maintenance_margin_rate`,
  ADD COLUMN `liquidation_fee_rate` DECIMAL(20,10) NOT NULL DEFAULT 0 COMMENT '强平手续费率' AFTER `min_margin_rate`,
  ADD COLUMN `insurance_user_id` BIGINT NOT NULL DEFAULT 0 COMMENT '保险基金用户ID' AFTER `liquidation_fee_rate`,
  ADD COLUMN `insurance_account_id` BIGINT NOT NULL DEFAULT 0 COMMENT '保险基金Option账户ID' AFTER `insurance_user_id`;

CREATE TABLE IF NOT EXISTS `t_option_margin_lot` (
  `id` BIGINT NOT NULL AUTO_INCREMENT, `tenant_id` BIGINT NOT NULL DEFAULT 0,
  `user_id` BIGINT NOT NULL DEFAULT 0, `account_id` BIGINT NOT NULL DEFAULT 0,
  `contract_id` BIGINT NOT NULL DEFAULT 0, `position_id` BIGINT NOT NULL DEFAULT 0,
  `order_id` BIGINT NOT NULL DEFAULT 0, `trade_id` BIGINT NOT NULL DEFAULT 0,
  `freeze_biz_no` VARCHAR(96) NOT NULL DEFAULT '',
  `quantity` DECIMAL(32,16) NOT NULL DEFAULT 0,
  `remaining_quantity` DECIMAL(32,16) NOT NULL DEFAULT 0,
  `initial_margin` DECIMAL(32,16) NOT NULL DEFAULT 0,
  `remaining_margin` DECIMAL(32,16) NOT NULL DEFAULT 0,
  `pending_margin` DECIMAL(32,16) NOT NULL DEFAULT 0,
  `status` TINYINT NOT NULL DEFAULT 1, `create_times` BIGINT NOT NULL DEFAULT 0,
  `update_times` BIGINT NOT NULL DEFAULT 0, PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tenant_trade` (`tenant_id`, `trade_id`),
  KEY `idx_margin_lot_position` (`tenant_id`, `position_id`, `status`, `id`),
  KEY `idx_margin_lot_order` (`tenant_id`, `order_id`, `status`, `id`),
  CONSTRAINT `chk_option_margin_lot` CHECK (`quantity` > 0 AND `remaining_quantity` >= 0 AND `remaining_quantity` <= `quantity` AND `initial_margin` > 0 AND `remaining_margin` >= 0 AND `pending_margin` >= 0 AND `pending_margin` <= `remaining_margin` AND `status` IN (1,2,3,4,5,6))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='期权卖方保证金批次';

CREATE TABLE IF NOT EXISTS `t_option_margin_lot_application` (
  `id` BIGINT NOT NULL AUTO_INCREMENT, `tenant_id` BIGINT NOT NULL DEFAULT 0,
  `instruction_id` BIGINT NOT NULL DEFAULT 0, `margin_lot_id` BIGINT NOT NULL DEFAULT 0,
  `action` TINYINT NOT NULL DEFAULT 0, `amount` DECIMAL(32,16) NOT NULL DEFAULT 0,
  `create_times` BIGINT NOT NULL DEFAULT 0, PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tenant_instruction` (`tenant_id`, `instruction_id`),
  KEY `idx_application_margin_lot` (`tenant_id`, `margin_lot_id`),
  CONSTRAINT `chk_margin_lot_application` CHECK (`action` IN (2,3) AND `amount` > 0)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='保证金批次资产指令应用幂等记录';

CREATE TABLE IF NOT EXISTS `t_option_risk_account` (
  `id` BIGINT NOT NULL AUTO_INCREMENT, `tenant_id` BIGINT NOT NULL DEFAULT 0,
  `user_id` BIGINT NOT NULL DEFAULT 0, `account_id` BIGINT NOT NULL DEFAULT 0,
  `settle_coin` VARCHAR(16) NOT NULL DEFAULT '',
  `equity` DECIMAL(32,16) NOT NULL DEFAULT 0, `position_margin` DECIMAL(32,16) NOT NULL DEFAULT 0,
  `maintenance_margin` DECIMAL(32,16) NOT NULL DEFAULT 0, `unrealized_pnl` DECIMAL(32,16) NOT NULL DEFAULT 0,
  `risk_rate` DECIMAL(20,10) NOT NULL DEFAULT 0, `status` TINYINT NOT NULL DEFAULT 1,
  `last_calc_time` BIGINT NOT NULL DEFAULT 0, `create_times` BIGINT NOT NULL DEFAULT 0,
  `update_times` BIGINT NOT NULL DEFAULT 0, PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tenant_user_account_coin` (`tenant_id`, `user_id`, `account_id`, `settle_coin`),
  KEY `idx_risk_account_status` (`tenant_id`, `status`, `id`),
  CONSTRAINT `chk_option_risk_account` CHECK (`position_margin` >= 0 AND `maintenance_margin` >= 0 AND `risk_rate` >= 0 AND `status` IN (1,2,3,4,5))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='期权卖方风险账户';

CREATE TABLE IF NOT EXISTS `t_option_liquidation` (
  `id` BIGINT NOT NULL AUTO_INCREMENT, `tenant_id` BIGINT NOT NULL DEFAULT 0,
  `liquidation_no` VARCHAR(64) NOT NULL DEFAULT '', `user_id` BIGINT NOT NULL DEFAULT 0,
  `account_id` BIGINT NOT NULL DEFAULT 0, `contract_id` BIGINT NOT NULL DEFAULT 0,
  `position_id` BIGINT NOT NULL DEFAULT 0, `quantity` DECIMAL(32,16) NOT NULL DEFAULT 0,
  `mark_price` DECIMAL(32,16) NOT NULL DEFAULT 0, `maintenance_margin` DECIMAL(32,16) NOT NULL DEFAULT 0,
  `equity` DECIMAL(32,16) NOT NULL DEFAULT 0, `deficit_amount` DECIMAL(32,16) NOT NULL DEFAULT 0,
  `liquidation_fee` DECIMAL(32,16) NOT NULL DEFAULT 0, `status` TINYINT NOT NULL DEFAULT 1,
  `retry_count` INT NOT NULL DEFAULT 0, `last_error_msg` VARCHAR(500) NOT NULL DEFAULT '',
  `collateral_amount` DECIMAL(32,16) NOT NULL DEFAULT 0,
  `insurance_fund_amount` DECIMAL(32,16) NOT NULL DEFAULT 0,
  `remaining_deficit` DECIMAL(32,16) NOT NULL DEFAULT 0,
  `takeover_position_id` BIGINT NOT NULL DEFAULT 0,
  `completed_at` BIGINT NOT NULL DEFAULT 0,
  `insurance_attempt` INT NOT NULL DEFAULT 0,
  `create_times` BIGINT NOT NULL DEFAULT 0, `update_times` BIGINT NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`), UNIQUE KEY `uk_tenant_liquidation_no` (`tenant_id`, `liquidation_no`),
  KEY `idx_liquidation_status` (`tenant_id`, `status`, `id`),
  KEY `idx_liquidation_position` (`tenant_id`, `position_id`, `id`),
  CONSTRAINT `chk_option_liquidation` CHECK (`quantity` > 0 AND `deficit_amount` >= 0 AND `liquidation_fee` >= 0 AND `status` IN (1,2,3,4,5,6) AND `retry_count` >= 0 AND `insurance_attempt` >= 0)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='期权强平记录';

CREATE TABLE IF NOT EXISTS `t_option_insurance_fund_flow` (
  `id` BIGINT NOT NULL AUTO_INCREMENT, `tenant_id` BIGINT NOT NULL DEFAULT 0,
  `flow_no` VARCHAR(64) NOT NULL DEFAULT '', `contract_id` BIGINT NOT NULL DEFAULT 0,
  `liquidation_id` BIGINT NOT NULL DEFAULT 0, `flow_type` TINYINT NOT NULL DEFAULT 0,
  `coin` VARCHAR(16) NOT NULL DEFAULT '', `amount` DECIMAL(32,16) NOT NULL DEFAULT 0,
  `asset_flow_no` VARCHAR(64) NOT NULL DEFAULT '', `create_times` BIGINT NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`), UNIQUE KEY `uk_tenant_flow_no` (`tenant_id`, `flow_no`),
  KEY `idx_insurance_liquidation` (`tenant_id`, `liquidation_id`),
  CONSTRAINT `chk_option_insurance_fund_flow` CHECK (`flow_type` IN (1,2,3,4) AND `amount` <> 0)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='期权保险基金流水';
