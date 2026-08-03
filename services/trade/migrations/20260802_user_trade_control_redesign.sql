ALTER TABLE `t_risk_user_trade_limit`
  ADD COLUMN `contract_type` TINYINT NOT NULL DEFAULT 0 COMMENT '合约类型：0不适用/全部 1永续 2交割' AFTER `product_type`,
  ADD COLUMN `control_mode` TINYINT NOT NULL DEFAULT 1 COMMENT '控制模式：1正常 2只平仓 3仅减仓 4禁用' AFTER `contract_type`,
  ADD COLUMN `version` BIGINT NOT NULL DEFAULT 1 COMMENT '乐观锁版本' AFTER `remark`,
  DROP INDEX `uk_tenant_user_product_type`,
  ADD UNIQUE KEY `uk_tenant_user_product_contract` (`tenant_id`, `user_id`, `product_type`, `contract_type`);

ALTER TABLE `t_risk_user_trade_limit`
  DROP CHECK `chk_user_trade_limit_flags`,
  ADD CONSTRAINT `chk_user_trade_limit_flags` CHECK (`product_type` IN (1, 2, 3) AND `contract_type` IN (0, 1, 2) AND (`product_type` = 2 OR `contract_type` = 0) AND `control_mode` IN (1, 2, 3, 4) AND `can_open` IN (0, 1) AND `can_close` IN (0, 1) AND `can_cancel` IN (0, 1) AND `can_trigger_order` IN (0, 1) AND `can_api_trade` IN (0, 1) AND `trade_enabled` IN (1, 2) AND `only_reduce_only` IN (1, 2) AND `enabled` IN (1, 2));

ALTER TABLE `t_risk_user_symbol_limit`
  ADD COLUMN `control_mode` TINYINT NOT NULL DEFAULT 1 COMMENT '控制模式：1正常 2只平仓 3仅减仓 4禁用' AFTER `symbol_id`,
  ADD COLUMN `version` BIGINT NOT NULL DEFAULT 1 COMMENT '乐观锁版本' AFTER `remark`;

ALTER TABLE `t_risk_user_symbol_limit`
  DROP CHECK `chk_user_symbol_limit_values`,
  ADD CONSTRAINT `chk_user_symbol_limit_values` CHECK (`control_mode` IN (1, 2, 3, 4) AND `max_position_qty` >= 0 AND `max_position_notional` >= 0 AND `max_open_orders` >= 0 AND `max_order_qty` >= 0 AND `max_order_notional` >= 0 AND `min_order_qty` >= 0 AND `min_order_notional` >= 0 AND `max_long_position_qty` >= 0 AND `max_short_position_qty` >= 0 AND `price_deviation_rate` >= 0 AND `enabled` IN (1, 2));

CREATE TABLE IF NOT EXISTS `t_trade_user_control_audit` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT NOT NULL DEFAULT 0 COMMENT '租户ID',
  `control_id` BIGINT NOT NULL DEFAULT 0 COMMENT '控制记录ID',
  `user_id` BIGINT NOT NULL COMMENT '用户ID',
  `change_type` TINYINT NOT NULL COMMENT '变更类型：1创建 2更新 3解除 4自动失效 5迁移',
  `before_json` JSON NULL COMMENT '修改前快照',
  `after_json` JSON NULL COMMENT '修改后快照',
  `operator_id` BIGINT NOT NULL DEFAULT 0 COMMENT '操作人ID',
  `source` TINYINT NOT NULL DEFAULT 3 COMMENT '来源：1系统 2用户 3后台管理 4任务',
  `reason` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '操作原因',
  `request_id` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '请求ID',
  `create_times` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间，毫秒时间戳',
  PRIMARY KEY (`id`),
  KEY `idx_tenant_user_id` (`tenant_id`, `user_id`, `id`),
  KEY `idx_control_id` (`control_id`, `id`),
  CONSTRAINT `chk_user_control_audit_type` CHECK (`change_type` IN (1, 2, 3, 4, 5))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户交易控制变更审计';

INSERT INTO `t_risk_user_trade_limit`
(`tenant_id`, `user_id`, `product_type`, `contract_type`, `control_mode`, `can_open`, `can_close`, `can_cancel`, `can_trigger_order`, `can_api_trade`, `trade_enabled`, `only_reduce_only`, `max_open_order_count`, `max_order_count_per_day`, `max_cancel_count_per_day`, `max_open_notional`, `max_position_notional`, `risk_level`, `operator_id`, `source`, `enabled`, `effective_start_time`, `effective_end_time`, `remark`, `version`, `create_times`, `update_times`)
SELECT `tenant_id`, `user_id`, `product_type`, 0, 4, 0, 0, 1, 0, 0, 2, 2, 0, 0, 0, 0, 0, 0, 0, 1, 1, 0, 0, 'migrated from t_trade_user_config', 1, `create_times`, `update_times`
FROM `t_trade_user_config`
WHERE `symbol_id` = 0 AND `trade_enabled` = 2
ON DUPLICATE KEY UPDATE
  `control_mode` = 4,
  `trade_enabled` = 2,
  `update_times` = GREATEST(`t_risk_user_trade_limit`.`update_times`, VALUES(`update_times`));

INSERT INTO `t_risk_user_symbol_limit`
(`tenant_id`, `user_id`, `symbol_id`, `control_mode`, `max_position_qty`, `max_position_notional`, `max_open_orders`, `max_order_qty`, `max_order_notional`, `min_order_qty`, `min_order_notional`, `max_long_position_qty`, `max_short_position_qty`, `price_deviation_rate`, `operator_id`, `source`, `enabled`, `effective_start_time`, `effective_end_time`, `remark`, `version`, `create_times`, `update_times`)
SELECT `tenant_id`, `user_id`, `symbol_id`, 4, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 0, 0, 'migrated from t_trade_user_config', 1, `create_times`, `update_times`
FROM `t_trade_user_config`
WHERE `symbol_id` > 0 AND `trade_enabled` = 2
ON DUPLICATE KEY UPDATE
  `control_mode` = 4,
  `update_times` = GREATEST(`t_risk_user_symbol_limit`.`update_times`, VALUES(`update_times`));
