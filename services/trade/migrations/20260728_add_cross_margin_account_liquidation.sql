-- Durable account-level liquidation boundary for one
-- (tenant,user,margin_asset) cross-margin risk unit.

CREATE TABLE IF NOT EXISTS `t_contract_account_liquidation` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT NOT NULL COMMENT '租户ID',
  `liquidation_no` VARCHAR(64) NOT NULL COMMENT '账户强平批次号',
  `user_id` BIGINT NOT NULL COMMENT '用户ID',
  `margin_asset` VARCHAR(32) NOT NULL COMMENT '全仓风险单元保证金币种',
  `margin_snapshot_id` BIGINT NOT NULL COMMENT '触发风险快照ID',
  `margin_snapshot_version` BIGINT NOT NULL COMMENT '触发风险快照版本',
  `asset_version` BIGINT NOT NULL DEFAULT 0 COMMENT '触发时Asset账户版本',
  `wallet_balance` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT '触发时Asset账户余额',
  `position_margin` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT '触发时仓位保证金合计',
  `maintenance_margin` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT '触发时维持保证金合计',
  `account_equity` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT '触发时账户权益，可为负',
  `risk_rate` DECIMAL(20,10) NOT NULL DEFAULT 0 COMMENT '触发时账户风险率',
  `gross_settlement` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT '仓位保证金与未实现盈亏的接管净额',
  `liquidation_fee` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT '账户强平手续费合计',
  `user_credit` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT '应返还用户可用余额',
  `user_debit` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT '应从用户可用余额扣除',
  `insurance_fund_amount` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT '保险基金承接金额',
  `position_count` INT NOT NULL DEFAULT 0 COMMENT '接管仓位数',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1待接管 2资金结算中 3待关仓 4已完成 5人工处理',
  `reason` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '触发、恢复或人工处理原因',
  `started_at` BIGINT NOT NULL DEFAULT 0 COMMENT '开始时间',
  `completed_at` BIGINT NOT NULL DEFAULT 0 COMMENT '完成时间',
  `version` BIGINT NOT NULL DEFAULT 0 COMMENT '并发版本号',
  `create_times` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
  `update_times` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tenant_account_liquidation_no` (`tenant_id`,`liquidation_no`),
  KEY `idx_cross_account_active` (`tenant_id`,`user_id`,`margin_asset`,`status`,`update_times`),
  CONSTRAINT `chk_cross_account_liquidation` CHECK (
    `margin_snapshot_id` > 0 AND `margin_snapshot_version` > 0 AND
    `asset_version` >= 0 AND `wallet_balance` >= 0 AND
    `position_margin` >= 0 AND `maintenance_margin` >= 0 AND
    `risk_rate` >= 0 AND `liquidation_fee` >= 0 AND
    `user_credit` >= 0 AND `user_debit` >= 0 AND
    `insurance_fund_amount` >= 0 AND `position_count` >= 0 AND
    `status` BETWEEN 1 AND 5 AND `version` >= 0
  )
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='全仓账户级强平父Saga';

CREATE TABLE IF NOT EXISTS `t_contract_account_liquidation_item` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT NOT NULL COMMENT '租户ID',
  `account_liquidation_id` BIGINT NOT NULL COMMENT '账户强平批次ID',
  `liquidation_no` VARCHAR(64) NOT NULL COMMENT '账户强平批次号',
  `position_id` BIGINT NOT NULL COMMENT '接管仓位ID',
  `position_version` BIGINT NOT NULL COMMENT '接管后的仓位版本',
  `symbol_id` BIGINT NOT NULL COMMENT '交易标的ID',
  `position_side` TINYINT NOT NULL COMMENT '仓位方向',
  `trigger_qty` DECIMAL(36,18) NOT NULL COMMENT '接管数量',
  `trigger_mark_price` DECIMAL(36,18) NOT NULL COMMENT '接管标记价格',
  `trigger_snapshot_id` VARCHAR(64) NOT NULL COMMENT '接管标记价格快照',
  `position_margin` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT '释放仓位保证金',
  `maintenance_margin` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT '接管维持保证金',
  `realized_pnl` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT '按接管标记价实现盈亏',
  `liquidation_fee` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT '分摊强平手续费',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1已锁定 2已关仓',
  `create_times` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
  `update_times` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_cross_account_position` (`tenant_id`,`account_liquidation_id`,`position_id`),
  KEY `idx_cross_account_items` (`tenant_id`,`liquidation_no`,`status`,`id`),
  CONSTRAINT `chk_cross_account_liquidation_item` CHECK (
    `position_version` > 0 AND `position_side` IN (1,2,3) AND
    `trigger_qty` > 0 AND `trigger_mark_price` > 0 AND
    `position_margin` >= 0 AND `maintenance_margin` >= 0 AND
    `liquidation_fee` >= 0 AND `status` IN (1,2)
  )
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='全仓账户级强平仓位明细';
