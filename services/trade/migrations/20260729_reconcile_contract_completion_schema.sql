-- dbinit:baseline-safe
-- 已有数据库首次接管时，将永续/交割合约关键结构收敛到当前基础 Schema。
-- 所有 ADD/DROP 操作均先检查 information_schema，允许失败后安全重跑。

-- Settlement 对账字段与索引。
SET @dbinit_exists := (
  SELECT COUNT(*) FROM information_schema.columns
  WHERE table_schema = DATABASE()
    AND table_name = 't_trade_settlement_instruction'
    AND column_name = 'asset_flow_no'
);
SET @dbinit_sql := IF(
  @dbinit_exists = 0,
  'ALTER TABLE `t_trade_settlement_instruction` ADD COLUMN `asset_flow_no` VARCHAR(64) NOT NULL DEFAULT '''' COMMENT ''对账确认的Asset流水号'' AFTER `last_error_msg`',
  'SELECT 1'
);
PREPARE dbinit_stmt FROM @dbinit_sql;
EXECUTE dbinit_stmt;
DEALLOCATE PREPARE dbinit_stmt;

SET @dbinit_exists := (
  SELECT COUNT(*) FROM information_schema.columns
  WHERE table_schema = DATABASE()
    AND table_name = 't_trade_settlement_instruction'
    AND column_name = 'reconciled_at'
);
SET @dbinit_sql := IF(
  @dbinit_exists = 0,
  'ALTER TABLE `t_trade_settlement_instruction` ADD COLUMN `reconciled_at` BIGINT NOT NULL DEFAULT 0 COMMENT ''Asset流水对账完成时间'' AFTER `asset_flow_no`',
  'SELECT 1'
);
PREPARE dbinit_stmt FROM @dbinit_sql;
EXECUTE dbinit_stmt;
DEALLOCATE PREPARE dbinit_stmt;

SET @dbinit_exists := (
  SELECT COUNT(*) FROM information_schema.statistics
  WHERE table_schema = DATABASE()
    AND table_name = 't_trade_settlement_instruction'
    AND index_name = 'idx_settlement_reconcile'
);
SET @dbinit_sql := IF(
  @dbinit_exists = 0,
  'ALTER TABLE `t_trade_settlement_instruction` ADD INDEX `idx_settlement_reconcile` (`tenant_id`,`status`,`reconciled_at`,`id`)',
  'SELECT 1'
);
PREPARE dbinit_stmt FROM @dbinit_sql;
EXECUTE dbinit_stmt;
DEALLOCATE PREPARE dbinit_stmt;

-- 对账问题与循环扫描游标。
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
  `last_alert_at` BIGINT NOT NULL DEFAULT 0 COMMENT '最近一次输出或发送告警的时间',
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

SET @dbinit_exists := (
  SELECT COUNT(*) FROM information_schema.columns
  WHERE table_schema = DATABASE()
    AND table_name = 't_contract_reconciliation_issue'
    AND column_name = 'last_alert_at'
);
SET @dbinit_sql := IF(
  @dbinit_exists = 0,
  'ALTER TABLE `t_contract_reconciliation_issue` ADD COLUMN `last_alert_at` BIGINT NOT NULL DEFAULT 0 COMMENT ''最近一次输出或发送告警的时间'' AFTER `last_seen_at`',
  'SELECT 1'
);
PREPARE dbinit_stmt FROM @dbinit_sql;
EXECUTE dbinit_stmt;
DEALLOCATE PREPARE dbinit_stmt;

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

-- 全仓账户级风险投影。
SET @dbinit_exists := (
  SELECT COUNT(*) FROM information_schema.columns
  WHERE table_schema = DATABASE()
    AND table_name = 't_contract_margin_snapshot'
    AND column_name = 'maintenance_margin'
);
SET @dbinit_sql := IF(
  @dbinit_exists = 0,
  'ALTER TABLE `t_contract_margin_snapshot` ADD COLUMN `maintenance_margin` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT ''同保证金币种全仓持仓维持保证金合计'' AFTER `order_margin`',
  'SELECT 1'
);
PREPARE dbinit_stmt FROM @dbinit_sql;
EXECUTE dbinit_stmt;
DEALLOCATE PREPARE dbinit_stmt;

SET @dbinit_exists := (
  SELECT COUNT(*) FROM information_schema.columns
  WHERE table_schema = DATABASE()
    AND table_name = 't_contract_margin_snapshot'
    AND column_name = 'account_equity'
);
SET @dbinit_sql := IF(
  @dbinit_exists = 0,
  'ALTER TABLE `t_contract_margin_snapshot` ADD COLUMN `account_equity` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT ''钱包余额、全仓仓位保证金与未实现盈亏合计'' AFTER `maintenance_margin`',
  'SELECT 1'
);
PREPARE dbinit_stmt FROM @dbinit_sql;
EXECUTE dbinit_stmt;
DEALLOCATE PREPARE dbinit_stmt;

SET @dbinit_exists := (
  SELECT COUNT(*) FROM information_schema.columns
  WHERE table_schema = DATABASE()
    AND table_name = 't_contract_margin_snapshot'
    AND column_name = 'available_margin'
);
SET @dbinit_sql := IF(
  @dbinit_exists = 0,
  'ALTER TABLE `t_contract_margin_snapshot` ADD COLUMN `available_margin` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT ''Asset可用余额叠加全仓未实现盈亏后的风险可用额'' AFTER `account_equity`',
  'SELECT 1'
);
PREPARE dbinit_stmt FROM @dbinit_sql;
EXECUTE dbinit_stmt;
DEALLOCATE PREPARE dbinit_stmt;

SET @dbinit_exists := (
  SELECT COUNT(*) FROM information_schema.columns
  WHERE table_schema = DATABASE()
    AND table_name = 't_contract_margin_snapshot'
    AND column_name = 'risk_rate'
);
SET @dbinit_sql := IF(
  @dbinit_exists = 0,
  'ALTER TABLE `t_contract_margin_snapshot` ADD COLUMN `risk_rate` DECIMAL(20,10) NOT NULL DEFAULT 0 COMMENT ''账户级维持保证金/账户权益；非正权益使用上限值'' AFTER `available_margin`',
  'SELECT 1'
);
PREPARE dbinit_stmt FROM @dbinit_sql;
EXECUTE dbinit_stmt;
DEALLOCATE PREPARE dbinit_stmt;

SET @dbinit_exists := (
  SELECT COUNT(*) FROM information_schema.columns
  WHERE table_schema = DATABASE()
    AND table_name = 't_contract_margin_snapshot'
    AND column_name = 'position_count'
);
SET @dbinit_sql := IF(
  @dbinit_exists = 0,
  'ALTER TABLE `t_contract_margin_snapshot` ADD COLUMN `position_count` INT NOT NULL DEFAULT 0 COMMENT ''参与当前快照的开放全仓仓位数'' AFTER `risk_rate`',
  'SELECT 1'
);
PREPARE dbinit_stmt FROM @dbinit_sql;
EXECUTE dbinit_stmt;
DEALLOCATE PREPARE dbinit_stmt;

SET @dbinit_exists := (
  SELECT COUNT(*) FROM information_schema.columns
  WHERE table_schema = DATABASE()
    AND table_name = 't_contract_margin_snapshot'
    AND column_name = 'asset_version'
);
SET @dbinit_sql := IF(
  @dbinit_exists = 0,
  'ALTER TABLE `t_contract_margin_snapshot` ADD COLUMN `asset_version` BIGINT NOT NULL DEFAULT 0 COMMENT ''生成快照时读取的Asset账户版本'' AFTER `position_count`',
  'SELECT 1'
);
PREPARE dbinit_stmt FROM @dbinit_sql;
EXECUTE dbinit_stmt;
DEALLOCATE PREPARE dbinit_stmt;

SET @dbinit_exists := (
  SELECT COUNT(*) FROM information_schema.table_constraints
  WHERE constraint_schema = DATABASE()
    AND table_name = 't_contract_margin_snapshot'
    AND constraint_name = 'chk_cross_margin_snapshot_risk'
    AND constraint_type = 'CHECK'
);
SET @dbinit_sql := IF(
  @dbinit_exists > 0,
  'ALTER TABLE `t_contract_margin_snapshot` DROP CHECK `chk_cross_margin_snapshot_risk`',
  'SELECT 1'
);
PREPARE dbinit_stmt FROM @dbinit_sql;
EXECUTE dbinit_stmt;
DEALLOCATE PREPARE dbinit_stmt;

SET @dbinit_exists := (
  SELECT COUNT(*) FROM information_schema.table_constraints
  WHERE constraint_schema = DATABASE()
    AND table_name = 't_contract_margin_snapshot'
    AND constraint_name = 'chk_margin_snapshot_balances'
    AND constraint_type = 'CHECK'
);
SET @dbinit_sql := IF(
  @dbinit_exists > 0,
  'ALTER TABLE `t_contract_margin_snapshot` DROP CHECK `chk_margin_snapshot_balances`',
  'SELECT 1'
);
PREPARE dbinit_stmt FROM @dbinit_sql;
EXECUTE dbinit_stmt;
DEALLOCATE PREPARE dbinit_stmt;

ALTER TABLE `t_contract_margin_snapshot`
  ADD CONSTRAINT `chk_margin_snapshot_balances`
  CHECK (
    `wallet_balance` >= 0 AND `available_balance` >= 0
    AND `frozen_balance` >= 0 AND `position_margin` >= 0
    AND `order_margin` >= 0 AND `maintenance_margin` >= 0
    AND `risk_rate` >= 0 AND `position_count` >= 0
    AND `asset_version` >= 0 AND `version` >= 0
  );

-- 持仓模式；回填选择最近一条可以关联订单的 History。
SET @dbinit_exists := (
  SELECT COUNT(*) FROM information_schema.columns
  WHERE table_schema = DATABASE()
    AND table_name = 't_contract_position'
    AND column_name = 'position_mode'
);
SET @dbinit_sql := IF(
  @dbinit_exists = 0,
  'ALTER TABLE `t_contract_position` ADD COLUMN `position_mode` TINYINT NOT NULL DEFAULT 1 COMMENT ''持仓模式：1单向 2双向'' AFTER `position_side`',
  'SELECT 1'
);
PREPARE dbinit_stmt FROM @dbinit_sql;
EXECUTE dbinit_stmt;
DEALLOCATE PREPARE dbinit_stmt;

UPDATE `t_contract_position` AS p
JOIN (
  SELECT h.position_id, MAX(h.id) AS history_id
  FROM `t_contract_position_history` AS h
  JOIN `t_trade_order` AS o
    ON o.tenant_id = h.tenant_id
   AND o.id = h.ref_order_id
  GROUP BY h.position_id
) AS latest_order_history
  ON latest_order_history.position_id = p.id
JOIN `t_contract_position_history` AS h
  ON h.id = latest_order_history.history_id
JOIN `t_trade_order` AS o
  ON o.tenant_id = h.tenant_id
 AND o.id = h.ref_order_id
SET p.position_mode = CASE WHEN o.position_side = 1 THEN 1 ELSE 2 END;

SET @dbinit_exists := (
  SELECT COUNT(*) FROM information_schema.table_constraints
  WHERE constraint_schema = DATABASE()
    AND table_name = 't_contract_position'
    AND constraint_name = 'chk_position_dimensions'
    AND constraint_type = 'CHECK'
);
SET @dbinit_sql := IF(
  @dbinit_exists > 0,
  'ALTER TABLE `t_contract_position` DROP CHECK `chk_position_dimensions`',
  'SELECT 1'
);
PREPARE dbinit_stmt FROM @dbinit_sql;
EXECUTE dbinit_stmt;
DEALLOCATE PREPARE dbinit_stmt;

ALTER TABLE `t_contract_position`
  ADD CONSTRAINT `chk_position_dimensions`
  CHECK (
    `contract_type` IN (1,2)
    AND `contract_value_type` IN (1,2)
    AND `position_side` IN (1,2,3)
    AND `position_mode` IN (1,2)
    AND `margin_mode` IN (1,2)
    AND `status` BETWEEN 1 AND 6
    AND `leverage` > 0
  );

-- 全仓账户级强平父/子 Saga；新库直接创建最终结构。
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
  `deficit_amount` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT '用户余额扣尽后的账户穿仓缺口',
  `insurance_fund_amount` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT '保险基金承接金额',
  `adl_relief_amount` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT 'ADL累计缓释金额',
  `adl_qty` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT 'ADL累计接管数量',
  `position_count` INT NOT NULL DEFAULT 0 COMMENT '接管仓位数',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1待接管 2资金结算中 3待关仓 4已完成 5人工处理 6保险承接 7自动减仓',
  `reason` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '触发、恢复或人工处理原因',
  `started_at` BIGINT NOT NULL DEFAULT 0 COMMENT '开始时间',
  `completed_at` BIGINT NOT NULL DEFAULT 0 COMMENT '完成时间',
  `version` BIGINT NOT NULL DEFAULT 0 COMMENT '并发版本号',
  `create_times` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
  `update_times` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tenant_account_liquidation_no` (`tenant_id`,`liquidation_no`),
  KEY `idx_cross_account_active` (`tenant_id`,`user_id`,`margin_asset`,`status`,`update_times`),
  CONSTRAINT `chk_cross_account_liquidation`
    CHECK (
      `margin_snapshot_id` > 0 AND `margin_snapshot_version` > 0
      AND `asset_version` >= 0 AND `wallet_balance` >= 0
      AND `position_margin` >= 0 AND `maintenance_margin` >= 0
      AND `risk_rate` >= 0 AND `liquidation_fee` >= 0
      AND `user_credit` >= 0 AND `user_debit` >= 0
      AND `deficit_amount` >= 0 AND `insurance_fund_amount` >= 0
      AND `adl_relief_amount` >= 0 AND `adl_qty` >= 0
      AND `position_count` >= 0
      AND `status` IN (1,2,3,4,5,6,7) AND `version` >= 0
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
  `deficit_amount` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT '分摊到仓位的ADL目标缺口',
  `bankruptcy_price` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT '按分摊缺口冻结的合成破产价',
  `adl_relief_amount` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT '仓位ADL累计缓释金额',
  `adl_qty` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT '仓位ADL累计接管数量',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1已锁定 2已关仓',
  `create_times` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
  `update_times` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_cross_account_position` (`tenant_id`,`account_liquidation_id`,`position_id`),
  KEY `idx_cross_account_items` (`tenant_id`,`liquidation_no`,`status`,`id`),
  CONSTRAINT `chk_cross_account_liquidation_item`
    CHECK (
      `position_version` > 0 AND `position_side` IN (1,2,3)
      AND `trigger_qty` > 0 AND `trigger_mark_price` > 0
      AND `position_margin` >= 0 AND `maintenance_margin` >= 0
      AND `liquidation_fee` >= 0 AND `deficit_amount` >= 0
      AND `bankruptcy_price` >= 0 AND `adl_relief_amount` >= 0
      AND `adl_qty` >= 0 AND `status` IN (1,2)
    )
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='全仓账户级强平仓位明细';

-- 若数据库已执行过早期强平迁移，逐列补齐负权益/ADL字段。
SET @dbinit_exists := (
  SELECT COUNT(*) FROM information_schema.columns
  WHERE table_schema = DATABASE()
    AND table_name = 't_contract_account_liquidation'
    AND column_name = 'deficit_amount'
);
SET @dbinit_sql := IF(
  @dbinit_exists = 0,
  'ALTER TABLE `t_contract_account_liquidation` ADD COLUMN `deficit_amount` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT ''用户余额扣尽后的账户穿仓缺口'' AFTER `user_debit`',
  'SELECT 1'
);
PREPARE dbinit_stmt FROM @dbinit_sql;
EXECUTE dbinit_stmt;
DEALLOCATE PREPARE dbinit_stmt;

SET @dbinit_exists := (
  SELECT COUNT(*) FROM information_schema.columns
  WHERE table_schema = DATABASE()
    AND table_name = 't_contract_account_liquidation'
    AND column_name = 'adl_relief_amount'
);
SET @dbinit_sql := IF(
  @dbinit_exists = 0,
  'ALTER TABLE `t_contract_account_liquidation` ADD COLUMN `adl_relief_amount` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT ''ADL累计缓释金额'' AFTER `insurance_fund_amount`',
  'SELECT 1'
);
PREPARE dbinit_stmt FROM @dbinit_sql;
EXECUTE dbinit_stmt;
DEALLOCATE PREPARE dbinit_stmt;

SET @dbinit_exists := (
  SELECT COUNT(*) FROM information_schema.columns
  WHERE table_schema = DATABASE()
    AND table_name = 't_contract_account_liquidation'
    AND column_name = 'adl_qty'
);
SET @dbinit_sql := IF(
  @dbinit_exists = 0,
  'ALTER TABLE `t_contract_account_liquidation` ADD COLUMN `adl_qty` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT ''ADL累计接管数量'' AFTER `adl_relief_amount`',
  'SELECT 1'
);
PREPARE dbinit_stmt FROM @dbinit_sql;
EXECUTE dbinit_stmt;
DEALLOCATE PREPARE dbinit_stmt;

SET @dbinit_exists := (
  SELECT COUNT(*) FROM information_schema.columns
  WHERE table_schema = DATABASE()
    AND table_name = 't_contract_account_liquidation_item'
    AND column_name = 'deficit_amount'
);
SET @dbinit_sql := IF(
  @dbinit_exists = 0,
  'ALTER TABLE `t_contract_account_liquidation_item` ADD COLUMN `deficit_amount` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT ''分摊到仓位的ADL目标缺口'' AFTER `liquidation_fee`',
  'SELECT 1'
);
PREPARE dbinit_stmt FROM @dbinit_sql;
EXECUTE dbinit_stmt;
DEALLOCATE PREPARE dbinit_stmt;

SET @dbinit_exists := (
  SELECT COUNT(*) FROM information_schema.columns
  WHERE table_schema = DATABASE()
    AND table_name = 't_contract_account_liquidation_item'
    AND column_name = 'bankruptcy_price'
);
SET @dbinit_sql := IF(
  @dbinit_exists = 0,
  'ALTER TABLE `t_contract_account_liquidation_item` ADD COLUMN `bankruptcy_price` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT ''按分摊缺口冻结的合成破产价'' AFTER `deficit_amount`',
  'SELECT 1'
);
PREPARE dbinit_stmt FROM @dbinit_sql;
EXECUTE dbinit_stmt;
DEALLOCATE PREPARE dbinit_stmt;

SET @dbinit_exists := (
  SELECT COUNT(*) FROM information_schema.columns
  WHERE table_schema = DATABASE()
    AND table_name = 't_contract_account_liquidation_item'
    AND column_name = 'adl_relief_amount'
);
SET @dbinit_sql := IF(
  @dbinit_exists = 0,
  'ALTER TABLE `t_contract_account_liquidation_item` ADD COLUMN `adl_relief_amount` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT ''仓位ADL累计缓释金额'' AFTER `bankruptcy_price`',
  'SELECT 1'
);
PREPARE dbinit_stmt FROM @dbinit_sql;
EXECUTE dbinit_stmt;
DEALLOCATE PREPARE dbinit_stmt;

SET @dbinit_exists := (
  SELECT COUNT(*) FROM information_schema.columns
  WHERE table_schema = DATABASE()
    AND table_name = 't_contract_account_liquidation_item'
    AND column_name = 'adl_qty'
);
SET @dbinit_sql := IF(
  @dbinit_exists = 0,
  'ALTER TABLE `t_contract_account_liquidation_item` ADD COLUMN `adl_qty` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT ''仓位ADL累计接管数量'' AFTER `adl_relief_amount`',
  'SELECT 1'
);
PREPARE dbinit_stmt FROM @dbinit_sql;
EXECUTE dbinit_stmt;
DEALLOCATE PREPARE dbinit_stmt;

SET @dbinit_exists := (
  SELECT COUNT(*) FROM information_schema.table_constraints
  WHERE constraint_schema = DATABASE()
    AND table_name = 't_contract_account_liquidation'
    AND constraint_name = 'chk_cross_account_liquidation'
    AND constraint_type = 'CHECK'
);
SET @dbinit_sql := IF(
  @dbinit_exists > 0,
  'ALTER TABLE `t_contract_account_liquidation` DROP CHECK `chk_cross_account_liquidation`',
  'SELECT 1'
);
PREPARE dbinit_stmt FROM @dbinit_sql;
EXECUTE dbinit_stmt;
DEALLOCATE PREPARE dbinit_stmt;

ALTER TABLE `t_contract_account_liquidation`
  ADD CONSTRAINT `chk_cross_account_liquidation`
  CHECK (
    `margin_snapshot_id` > 0 AND `margin_snapshot_version` > 0
    AND `asset_version` >= 0 AND `wallet_balance` >= 0
    AND `position_margin` >= 0 AND `maintenance_margin` >= 0
    AND `risk_rate` >= 0 AND `liquidation_fee` >= 0
    AND `user_credit` >= 0 AND `user_debit` >= 0
    AND `deficit_amount` >= 0 AND `insurance_fund_amount` >= 0
    AND `adl_relief_amount` >= 0 AND `adl_qty` >= 0
    AND `position_count` >= 0
    AND `status` IN (1,2,3,4,5,6,7) AND `version` >= 0
  );

SET @dbinit_exists := (
  SELECT COUNT(*) FROM information_schema.table_constraints
  WHERE constraint_schema = DATABASE()
    AND table_name = 't_contract_account_liquidation_item'
    AND constraint_name = 'chk_cross_account_liquidation_item'
    AND constraint_type = 'CHECK'
);
SET @dbinit_sql := IF(
  @dbinit_exists > 0,
  'ALTER TABLE `t_contract_account_liquidation_item` DROP CHECK `chk_cross_account_liquidation_item`',
  'SELECT 1'
);
PREPARE dbinit_stmt FROM @dbinit_sql;
EXECUTE dbinit_stmt;
DEALLOCATE PREPARE dbinit_stmt;

ALTER TABLE `t_contract_account_liquidation_item`
  ADD CONSTRAINT `chk_cross_account_liquidation_item`
  CHECK (
    `position_version` > 0 AND `position_side` IN (1,2,3)
    AND `trigger_qty` > 0 AND `trigger_mark_price` > 0
    AND `position_margin` >= 0 AND `maintenance_margin` >= 0
    AND `liquidation_fee` >= 0 AND `deficit_amount` >= 0
    AND `bankruptcy_price` >= 0 AND `adl_relief_amount` >= 0
    AND `adl_qty` >= 0 AND `status` IN (1,2)
  );
