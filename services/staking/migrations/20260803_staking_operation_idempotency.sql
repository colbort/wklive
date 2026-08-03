-- dbinit:baseline-safe
-- Add request idempotency, optimistic versioning and a recoverable money-operation table.

SET @dbinit_exists := (
  SELECT COUNT(*) FROM information_schema.columns
  WHERE table_schema = DATABASE()
    AND table_name = 't_stake_order'
    AND column_name = 'request_no'
);
SET @dbinit_sql := IF(
  @dbinit_exists = 0,
  'ALTER TABLE `t_stake_order` ADD COLUMN `request_no` VARCHAR(96) NOT NULL DEFAULT '''' COMMENT ''创建请求幂等号'' AFTER `order_no`',
  'SELECT 1'
);
PREPARE dbinit_stmt FROM @dbinit_sql;
EXECUTE dbinit_stmt;
DEALLOCATE PREPARE dbinit_stmt;

SET @dbinit_exists := (
  SELECT COUNT(*) FROM information_schema.columns
  WHERE table_schema = DATABASE()
    AND table_name = 't_stake_order'
    AND column_name = 'active_operation_no'
);
SET @dbinit_sql := IF(
  @dbinit_exists = 0,
  'ALTER TABLE `t_stake_order` ADD COLUMN `active_operation_no` VARCHAR(96) NOT NULL DEFAULT '''' COMMENT ''当前占用订单的资金操作号'' AFTER `request_no`',
  'SELECT 1'
);
PREPARE dbinit_stmt FROM @dbinit_sql;
EXECUTE dbinit_stmt;
DEALLOCATE PREPARE dbinit_stmt;

CREATE TABLE IF NOT EXISTS `t_stake_user_position` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT NOT NULL DEFAULT 0 COMMENT '租户ID',
  `user_id` BIGINT NOT NULL DEFAULT 0 COMMENT '用户ID',
  `product_id` BIGINT NOT NULL DEFAULT 0 COMMENT '质押产品ID',
  `staked_amount` DECIMAL(30,8) NOT NULL DEFAULT 0.00000000 COMMENT '当前在途和生效质押本金',
  `version` BIGINT NOT NULL DEFAULT 1 COMMENT '并发控制版本',
  `create_times` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间戳（毫秒）',
  `update_times` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间戳（毫秒）',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tenant_user_product` (`tenant_id`,`user_id`,`product_id`),
  KEY `idx_tenant_product` (`tenant_id`,`product_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='质押用户产品额度聚合表';

INSERT INTO `t_stake_user_position`
  (`tenant_id`,`user_id`,`product_id`,`staked_amount`,`version`,`create_times`,`update_times`)
SELECT
  `tenant_id`, `user_id`, `product_id`, COALESCE(SUM(`stake_amount`), 0), 1,
  COALESCE(MIN(`create_times`), 0), COALESCE(MAX(`update_times`), 0)
FROM `t_stake_order`
WHERE `status` IN (1,2)
GROUP BY `tenant_id`,`user_id`,`product_id`
ON DUPLICATE KEY UPDATE
  `staked_amount` = VALUES(`staked_amount`),
  `version` = `version` + 1,
  `update_times` = VALUES(`update_times`);

-- Historical rows did not carry a request number. Give each row a stable,
-- non-colliding value before adding the unique key.
UPDATE `t_stake_order`
SET `request_no` = CONCAT('LEGACY_ORDER_', `id`)
WHERE `request_no` = '';

SET @dbinit_exists := (
  SELECT COUNT(*) FROM information_schema.columns
  WHERE table_schema = DATABASE()
    AND table_name = 't_stake_order'
    AND column_name = 'version'
);
SET @dbinit_sql := IF(
  @dbinit_exists = 0,
  'ALTER TABLE `t_stake_order` ADD COLUMN `version` BIGINT NOT NULL DEFAULT 1 COMMENT ''并发控制版本'' AFTER `update_times`',
  'SELECT 1'
);
PREPARE dbinit_stmt FROM @dbinit_sql;
EXECUTE dbinit_stmt;
DEALLOCATE PREPARE dbinit_stmt;

SET @dbinit_exists := (
  SELECT COUNT(*) FROM information_schema.statistics
  WHERE table_schema = DATABASE()
    AND table_name = 't_stake_order'
    AND index_name = 'uk_tenant_user_request'
);
SET @dbinit_sql := IF(
  @dbinit_exists = 0,
  'ALTER TABLE `t_stake_order` ADD UNIQUE KEY `uk_tenant_user_request` (`tenant_id`,`user_id`,`request_no`)',
  'SELECT 1'
);
PREPARE dbinit_stmt FROM @dbinit_sql;
EXECUTE dbinit_stmt;
DEALLOCATE PREPARE dbinit_stmt;

CREATE TABLE IF NOT EXISTS `t_stake_operation` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT NOT NULL DEFAULT 0 COMMENT '租户ID',
  `user_id` BIGINT NOT NULL DEFAULT 0 COMMENT '用户ID',
  `order_id` BIGINT NOT NULL DEFAULT 0 COMMENT '质押订单ID',
  `order_no` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '质押订单号',
  `operation_no` VARCHAR(96) NOT NULL DEFAULT '' COMMENT '资金操作号',
  `request_no` VARCHAR(96) NOT NULL DEFAULT '' COMMENT '调用方幂等号',
  `operation_type` TINYINT NOT NULL DEFAULT 0 COMMENT '操作类型：1申购 2日收益 3到期收益 4到期赎回 5提前赎回 6人工收益 7人工赎回 8申购回滚',
  `principal_amount` DECIMAL(30,8) NOT NULL DEFAULT 0.00000000 COMMENT '本金金额',
  `reward_amount` DECIMAL(30,8) NOT NULL DEFAULT 0.00000000 COMMENT '收益金额',
  `fee_amount` DECIMAL(30,8) NOT NULL DEFAULT 0.00000000 COMMENT '手续费金额',
  `principal_status` TINYINT NOT NULL DEFAULT 0 COMMENT '本金步骤：0无需 1待处理 2成功',
  `reward_status` TINYINT NOT NULL DEFAULT 0 COMMENT '收益步骤：0无需 1待处理 2成功',
  `fee_status` TINYINT NOT NULL DEFAULT 0 COMMENT '手续费步骤：0无需 1待处理 2成功',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '操作状态：1待处理 2处理中 3成功 4可重试失败 5需人工处理',
  `period_end` BIGINT NOT NULL DEFAULT 0 COMMENT '收益周期结束时间（毫秒）',
  `retry_count` INT NOT NULL DEFAULT 0 COMMENT '重试次数',
  `next_retry_at` BIGINT NOT NULL DEFAULT 0 COMMENT '下次重试时间（毫秒）',
  `last_error` VARCHAR(1000) NOT NULL DEFAULT '' COMMENT '最近错误',
  `operator_user_id` BIGINT NOT NULL DEFAULT 0 COMMENT '操作人ID，用户操作为用户ID',
  `remark` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '备注',
  `version` BIGINT NOT NULL DEFAULT 1 COMMENT '并发控制版本',
  `create_times` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间戳（毫秒）',
  `update_times` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间戳（毫秒）',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tenant_operation_no` (`tenant_id`,`operation_no`),
  UNIQUE KEY `uk_tenant_user_type_request` (`tenant_id`,`user_id`,`operation_type`,`request_no`),
  KEY `idx_tenant_order_id` (`tenant_id`,`order_id`),
  KEY `idx_status_retry` (`status`,`next_retry_at`,`id`),
  KEY `idx_tenant_period` (`tenant_id`,`operation_type`,`period_end`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='质押可恢复资金操作单';

SET @dbinit_exists := (
  SELECT COUNT(*) FROM information_schema.columns
  WHERE table_schema = DATABASE()
    AND table_name = 't_stake_reward_log'
    AND column_name = 'operation_no'
);
SET @dbinit_sql := IF(
  @dbinit_exists = 0,
  'ALTER TABLE `t_stake_reward_log` ADD COLUMN `operation_no` VARCHAR(96) NOT NULL DEFAULT '''' COMMENT ''收益资金操作号'' AFTER `order_no`',
  'SELECT 1'
);
PREPARE dbinit_stmt FROM @dbinit_sql;
EXECUTE dbinit_stmt;
DEALLOCATE PREPARE dbinit_stmt;

UPDATE `t_stake_reward_log`
SET `operation_no` = CONCAT('LEGACY_REWARD_', `id`)
WHERE `operation_no` = '';

SET @dbinit_exists := (
  SELECT COUNT(*) FROM information_schema.statistics
  WHERE table_schema = DATABASE()
    AND table_name = 't_stake_reward_log'
    AND index_name = 'uk_tenant_operation_no'
);
SET @dbinit_sql := IF(
  @dbinit_exists = 0,
  'ALTER TABLE `t_stake_reward_log` ADD UNIQUE KEY `uk_tenant_operation_no` (`tenant_id`,`operation_no`)',
  'SELECT 1'
);
PREPARE dbinit_stmt FROM @dbinit_sql;
EXECUTE dbinit_stmt;
DEALLOCATE PREPARE dbinit_stmt;
